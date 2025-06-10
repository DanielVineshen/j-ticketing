package service

import (
	"fmt"
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
