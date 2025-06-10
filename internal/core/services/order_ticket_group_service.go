package service

import (
	"fmt"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/utils"
	"math"
	"time"
)

type OrderTicketGroupService struct {
	orderTicketGroupRepo *repositories.OrderTicketGroupRepository
}

// NewOrderTicketGroupService creates a new instance of OrderTicketGroupService
func NewOrderTicketGroupService(
	orderTicketGroupRepo *repositories.OrderTicketGroupRepository,
) *OrderTicketGroupService {
	return &OrderTicketGroupService{
		orderTicketGroupRepo: orderTicketGroupRepo,
	}
}

type OrderDateRangeResponse struct {
	StartDate           string           `json:"startDate"`
	EndDate             string           `json:"endDate"`
	CurrentTotalRecords int              `json:"currentTotalRecords"`
	PastTotalRecords    int              `json:"pastTotalRecords"`
	DiffFromPastPeriod  float64          `json:"diffFromPastPeriod"`
	DailyData           []OrderDailyData `json:"dailyData"`
}

type OrderDailyData struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

func (o *OrderTicketGroupService) GetTotalOnsiteVisitorsWithinRange(startDate, endDate string) (*OrderDateRangeResponse, error) {
	// Parse dates to calculate the period duration
	start, err := time.Parse(utils.DateOnlyFormat, startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid startDate format: %w", err)
	}

	end, err := time.Parse(utils.DateOnlyFormat, endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid endDate format: %w", err)
	}

	// Calculate the duration of the current period
	duration := end.Sub(start)

	// Calculate past period dates
	pastEndDate := start.AddDate(0, 0, -1)      // Day before startDate
	pastStartDate := pastEndDate.Add(-duration) // Same duration backwards

	// Get current period orders
	orders, _ := o.orderTicketGroupRepo.FindOrderWithinDateRange(startDate, endDate)

	// Get past period orders - just need the count
	pastOrders, _ := o.orderTicketGroupRepo.FindOrderWithinDateRange(
		pastStartDate.Format(utils.DateOnlyFormat),
		pastEndDate.Format(utils.DateOnlyFormat),
	)

	// Calculate totals
	currentTotalRecords := len(orders)
	pastTotalRecords := len(pastOrders)

	// Calculate percentage difference from past period
	var diffFromPastPeriod float64
	if pastTotalRecords > 0 {
		diffFromPastPeriod = ((float64(currentTotalRecords) - float64(pastTotalRecords)) / float64(pastTotalRecords)) * 100
		// Round to 2 decimal places
		diffFromPastPeriod = math.Round(diffFromPastPeriod*100) / 100
	} else if currentTotalRecords > 0 {
		// If past period has 0 records but current has records, it's 100% increase
		diffFromPastPeriod = 100.00
	} else {
		// Both periods have 0 records, no change
		diffFromPastPeriod = 0.00
	}

	// Group current orders by date
	dailyCounts := make(map[string]int)
	for _, order := range orders {
		malaysiaTime, err := utils.ToMalaysiaTime(order.CreatedAt)
		if err != nil {
			continue
		}
		dateStr := malaysiaTime.Format(utils.DateOnlyFormat)
		dailyCounts[dateStr]++
	}

	// Generate complete date range (including days with 0 records)
	dates := o.generateDateRange(startDate, endDate)
	dailyData := make([]OrderDailyData, 0)

	for _, date := range dates {
		count := dailyCounts[date] // Will be 0 if date not in map
		dailyData = append(dailyData, OrderDailyData{
			Date:  date,
			Count: count,
		})
	}

	return &OrderDateRangeResponse{
		StartDate:           startDate,
		EndDate:             endDate,
		CurrentTotalRecords: currentTotalRecords,
		PastTotalRecords:    pastTotalRecords,
		DiffFromPastPeriod:  diffFromPastPeriod,
		DailyData:           dailyData,
	}, nil
}

type NewVsReturningDateRangeResponse struct {
	StartDate                    string                    `json:"startDate"`
	EndDate                      string                    `json:"endDate"`
	CurrentNewTotalRecords       int                       `json:"currentNewTotalRecords"`
	CurrentReturningTotalRecords int                       `json:"currentReturningTotalRecords"`
	PastNewTotalRecords          int                       `json:"pastNewTotalRecords"`
	PastReturningTotalRecords    int                       `json:"pastReturningTotalRecords"`
	DiffNewFromPastPeriod        float64                   `json:"diffNewFromPastPeriod"`
	DiffReturningFromPastPeriod  float64                   `json:"diffReturningFromPastPeriod"`
	DailyData                    []NewVsReturningDailyData `json:"dailyData"`
}

type NewVsReturningDailyData struct {
	Date           string `json:"date"`
	NewCount       int    `json:"newCount"`
	ReturningCount int    `json:"returningCount"`
	Total          int    `json:"total"`
}

func (o *OrderTicketGroupService) GetNewVsReturningVisitors(startDate, endDate string) (*NewVsReturningDateRangeResponse, error) {
	// Parse dates to calculate the period duration
	start, err := time.Parse(utils.DateOnlyFormat, startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid startDate format: %w", err)
	}

	end, err := time.Parse(utils.DateOnlyFormat, endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid endDate format: %w", err)
	}

	// Calculate the duration of the current period
	duration := end.Sub(start)

	// Calculate past period dates
	pastEndDate := start.AddDate(0, 0, -1)      // Day before startDate
	pastStartDate := pastEndDate.Add(-duration) // Same duration backwards

	// Get ALL orders to determine first order for each customer
	allOrders, err := o.orderTicketGroupRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all orders: %w", err)
	}

	// Create a map to track each customer's first order date
	customerFirstOrderDate := make(map[string]time.Time)

	for _, order := range allOrders {
		// Convert UTC CreatedAt to Malaysia timezone
		malaysiaTime, err := utils.ToMalaysiaTime(order.CreatedAt)
		if err != nil {
			continue
		}

		custId := order.CustId

		// Track the earliest order date for each customer
		if firstOrderDate, exists := customerFirstOrderDate[custId]; !exists || malaysiaTime.Before(firstOrderDate) {
			customerFirstOrderDate[custId] = malaysiaTime
		}
	}

	// Filter orders for current period
	currentPeriodOrders := make([]models.OrderTicketGroup, 0)
	pastPeriodOrders := make([]models.OrderTicketGroup, 0)

	for _, order := range allOrders {
		malaysiaTime, err := utils.ToMalaysiaTime(order.CreatedAt)
		if err != nil {
			continue
		}

		// Check if order falls in current period
		if malaysiaTime.After(start.Add(-time.Second)) && malaysiaTime.Before(end.Add(24*time.Hour)) {
			currentPeriodOrders = append(currentPeriodOrders, order)
		}

		// Check if order falls in past period
		if malaysiaTime.After(pastStartDate.Add(-time.Second)) && malaysiaTime.Before(pastEndDate.Add(24*time.Hour)) {
			pastPeriodOrders = append(pastPeriodOrders, order)
		}
	}

	// Process current period orders
	currentNewCount := 0
	currentReturningCount := 0
	currentDailyCounts := make(map[string]map[string]int) // date -> type -> count

	for _, order := range currentPeriodOrders {
		malaysiaTime, _ := utils.ToMalaysiaTime(order.CreatedAt)
		dateStr := malaysiaTime.Format(utils.DateOnlyFormat)

		if currentDailyCounts[dateStr] == nil {
			currentDailyCounts[dateStr] = make(map[string]int)
		}

		custId := order.CustId
		firstOrderDate := customerFirstOrderDate[custId]

		// Check if this order date is the same as the customer's first order date
		if malaysiaTime.Format(utils.DateOnlyFormat) == firstOrderDate.Format(utils.DateOnlyFormat) {
			// New customer
			currentNewCount++
			currentDailyCounts[dateStr]["new"]++
		} else {
			// Returning customer
			currentReturningCount++
			currentDailyCounts[dateStr]["returning"]++
		}
	}

	// Process past period orders
	pastNewCount := 0
	pastReturningCount := 0

	for _, order := range pastPeriodOrders {
		malaysiaTime, _ := utils.ToMalaysiaTime(order.CreatedAt)

		custId := order.CustId
		firstOrderDate := customerFirstOrderDate[custId]

		// Check if this order date is the same as the customer's first order date
		if malaysiaTime.Format(utils.DateOnlyFormat) == firstOrderDate.Format(utils.DateOnlyFormat) {
			pastNewCount++
		} else {
			pastReturningCount++
		}
	}

	// Calculate percentage differences
	var diffNewFromPastPeriod float64
	if pastNewCount > 0 {
		diffNewFromPastPeriod = ((float64(currentNewCount) - float64(pastNewCount)) / float64(pastNewCount)) * 100
		diffNewFromPastPeriod = math.Round(diffNewFromPastPeriod*100) / 100
	} else if currentNewCount > 0 {
		diffNewFromPastPeriod = 100.00
	} else {
		diffNewFromPastPeriod = 0.00
	}

	var diffReturningFromPastPeriod float64
	if pastReturningCount > 0 {
		diffReturningFromPastPeriod = ((float64(currentReturningCount) - float64(pastReturningCount)) / float64(pastReturningCount)) * 100
		diffReturningFromPastPeriod = math.Round(diffReturningFromPastPeriod*100) / 100
	} else if currentReturningCount > 0 {
		diffReturningFromPastPeriod = 100.00
	} else {
		diffReturningFromPastPeriod = 0.00
	}

	// Generate complete date range for daily data
	dates := o.generateDateRange(startDate, endDate)
	dailyData := make([]NewVsReturningDailyData, 0)

	for _, date := range dates {
		newCount := 0
		returningCount := 0

		if dailyCounts, exists := currentDailyCounts[date]; exists {
			newCount = dailyCounts["new"]
			returningCount = dailyCounts["returning"]
		}

		total := newCount + returningCount

		dailyData = append(dailyData, NewVsReturningDailyData{
			Date:           date,
			NewCount:       newCount,
			ReturningCount: returningCount,
			Total:          total,
		})
	}

	return &NewVsReturningDateRangeResponse{
		StartDate:                    startDate,
		EndDate:                      endDate,
		CurrentNewTotalRecords:       currentNewCount,
		CurrentReturningTotalRecords: currentReturningCount,
		PastNewTotalRecords:          pastNewCount,
		PastReturningTotalRecords:    pastReturningCount,
		DiffNewFromPastPeriod:        diffNewFromPastPeriod,
		DiffReturningFromPastPeriod:  diffReturningFromPastPeriod,
		DailyData:                    dailyData,
	}, nil
}

// generateDateRange generates a slice of date strings between start and end dates
func (o *OrderTicketGroupService) generateDateRange(startDate, endDate string) []string {
	start, _ := time.Parse(utils.DateOnlyFormat, startDate)
	end, _ := time.Parse(utils.DateOnlyFormat, endDate)

	var dates []string
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format(utils.DateOnlyFormat))
	}

	return dates
}
