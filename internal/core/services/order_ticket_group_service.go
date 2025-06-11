package service

import (
	"fmt"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/utils"
	"math"
	"sort"
	"time"
)

type OrderTicketGroupService struct {
	orderTicketGroupRepo *repositories.OrderTicketGroupRepository
	ticketGroupRepo      *repositories.TicketGroupRepository
}

// NewOrderTicketGroupService creates a new instance of OrderTicketGroupService
func NewOrderTicketGroupService(
	orderTicketGroupRepo *repositories.OrderTicketGroupRepository,
	ticketGroupRepo *repositories.TicketGroupRepository,
) *OrderTicketGroupService {
	return &OrderTicketGroupService{
		orderTicketGroupRepo: orderTicketGroupRepo,
		ticketGroupRepo:      ticketGroupRepo,
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
	pastEndDate := start.AddDate(0, 0, -1)
	pastStartDate := pastEndDate.Add(-duration)

	// Get all orders
	allOrders, err := o.orderTicketGroupRepo.FindAllSuccessfulOrders()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve orders: %w", err)
	}

	// Filter orders by admit date and count tickets
	currentTotalRecords := 0
	pastTotalRecords := 0
	dailyCounts := make(map[string]int)

	for _, order := range allOrders {
		if len(order.OrderTicketInfos) == 0 {
			continue // Skip orders without ticket info
		}

		// Get admit date from first ticket info (all will be the same)
		admitDateStr := order.OrderTicketInfos[0].AdmitDate
		admitDate, err := time.Parse("2006-01-02", admitDateStr)
		if err != nil {
			continue // Skip invalid dates
		}

		ticketCount := len(order.OrderTicketInfos)

		// Check if admit date falls in current period
		if admitDate.After(start.Add(-time.Second)) && admitDate.Before(end.Add(24*time.Hour)) {
			currentTotalRecords += ticketCount
			dailyCounts[admitDateStr] += ticketCount
		}

		// Check if admit date falls in past period
		if admitDate.After(pastStartDate.Add(-time.Second)) && admitDate.Before(pastEndDate.Add(24*time.Hour)) {
			pastTotalRecords += ticketCount
		}
	}

	// Calculate percentage difference
	var diffFromPastPeriod float64
	if pastTotalRecords > 0 {
		diffFromPastPeriod = ((float64(currentTotalRecords) - float64(pastTotalRecords)) / float64(pastTotalRecords)) * 100
		diffFromPastPeriod = math.Round(diffFromPastPeriod*100) / 100
	} else if currentTotalRecords > 0 {
		diffFromPastPeriod = 100.00
	} else {
		diffFromPastPeriod = 0.00
	}

	// Generate complete date range for daily data
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
	// Parse dates
	start, err := time.Parse(utils.DateOnlyFormat, startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid startDate format: %w", err)
	}

	end, err := time.Parse(utils.DateOnlyFormat, endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid endDate format: %w", err)
	}

	// Calculate past period dates
	duration := end.Sub(start)
	pastEndDate := start.AddDate(0, 0, -1)
	pastStartDate := pastEndDate.Add(-duration)

	// Get all orders
	allOrders, err := o.orderTicketGroupRepo.FindAllSuccessfulOrders()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve all orders: %w", err)
	}

	// Track each customer's first admit date
	customerFirstAdmitDate := make(map[string]time.Time)

	for _, order := range allOrders {
		if len(order.OrderTicketInfos) == 0 {
			continue
		}

		custId := order.CustId
		admitDateStr := order.OrderTicketInfos[0].AdmitDate
		admitDate, err := time.Parse("2006-01-02", admitDateStr)
		if err != nil {
			continue
		}

		// Track earliest admit date for each customer
		if firstAdmitDate, exists := customerFirstAdmitDate[custId]; !exists || admitDate.Before(firstAdmitDate) {
			customerFirstAdmitDate[custId] = admitDate
		}
	}

	// Process current and past periods
	currentNewCount := 0
	currentReturningCount := 0
	pastNewCount := 0
	pastReturningCount := 0
	currentDailyCounts := make(map[string]map[string]int)

	for _, order := range allOrders {
		if len(order.OrderTicketInfos) == 0 {
			continue
		}

		custId := order.CustId
		admitDateStr := order.OrderTicketInfos[0].AdmitDate
		admitDate, err := time.Parse("2006-01-02", admitDateStr)
		if err != nil {
			continue
		}

		ticketCount := len(order.OrderTicketInfos)
		firstAdmitDate := customerFirstAdmitDate[custId]
		isNewVisitor := admitDate.Format("2006-01-02") == firstAdmitDate.Format("2006-01-02")

		// Check if admit date is in current period
		if admitDate.After(start.Add(-time.Second)) && admitDate.Before(end.Add(24*time.Hour)) {
			if currentDailyCounts[admitDateStr] == nil {
				currentDailyCounts[admitDateStr] = make(map[string]int)
			}

			if isNewVisitor {
				currentNewCount += ticketCount
				currentDailyCounts[admitDateStr]["new"] += ticketCount
			} else {
				currentReturningCount += ticketCount
				currentDailyCounts[admitDateStr]["returning"] += ticketCount
			}
		}

		// Check if admit date is in past period
		if admitDate.After(pastStartDate.Add(-time.Second)) && admitDate.Before(pastEndDate.Add(24*time.Hour)) {
			if isNewVisitor {
				pastNewCount += ticketCount
			} else {
				pastReturningCount += ticketCount
			}
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

	// Generate daily data
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

type PeakDayAnalysisResponse struct {
	StartDate    string          `json:"startDate"`
	EndDate      string          `json:"endDate"`
	PeakDay      string          `json:"peakDay"`
	PeakDayCount int             `json:"peakDayCount"`
	WeeklyData   []WeeklyDayData `json:"weeklyData"`
}

type WeeklyDayData struct {
	DayOfWeek string  `json:"dayOfWeek"`
	Count     int     `json:"count"`
	Average   float64 `json:"average"`
}

func (o *OrderTicketGroupService) GetAveragePeakDayAnalysis(startDate, endDate string) (*PeakDayAnalysisResponse, error) {
	// Parse dates
	start, err := time.Parse(utils.DateOnlyFormat, startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid startDate format: %w", err)
	}

	end, err := time.Parse(utils.DateOnlyFormat, endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid endDate format: %w", err)
	}

	// Get all orders
	allOrders, err := o.orderTicketGroupRepo.FindAllSuccessfulOrders()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve orders: %w", err)
	}

	// Group ticket counts by day of week
	dayOfWeekCounts := make(map[time.Weekday]int)
	dayOfWeekDates := make(map[time.Weekday][]string) // Track dates for each day of week for averaging

	// Initialize all days to 0
	for i := time.Sunday; i <= time.Saturday; i++ {
		dayOfWeekCounts[i] = 0
		dayOfWeekDates[i] = make([]string, 0)
	}

	// Track unique dates in the range for proper averaging calculation
	uniqueDatesInRange := make(map[string]bool)

	for _, order := range allOrders {
		if len(order.OrderTicketInfos) == 0 {
			continue
		}

		// Get admit date from first ticket info
		admitDateStr := order.OrderTicketInfos[0].AdmitDate
		admitDate, err := time.Parse("2006-01-02", admitDateStr)
		if err != nil {
			continue
		}

		// Check if admit date falls within the specified range
		if admitDate.After(start.Add(-time.Second)) && admitDate.Before(end.Add(24*time.Hour)) {
			ticketCount := len(order.OrderTicketInfos)
			dayOfWeek := admitDate.Weekday()

			dayOfWeekCounts[dayOfWeek] += ticketCount

			// Track unique dates for this day of week
			if !contains(dayOfWeekDates[dayOfWeek], admitDateStr) {
				dayOfWeekDates[dayOfWeek] = append(dayOfWeekDates[dayOfWeek], admitDateStr)
			}

			// Track all unique dates in range
			uniqueDatesInRange[admitDateStr] = true
		}
	}

	// Calculate averages and find peak day
	weeklyData := make([]WeeklyDayData, 0)
	peakDay := ""
	peakDayCount := 0

	// Convert weekday to string names
	dayNames := map[time.Weekday]string{
		time.Sunday:    "Sunday",
		time.Monday:    "Monday",
		time.Tuesday:   "Tuesday",
		time.Wednesday: "Wednesday",
		time.Thursday:  "Thursday",
		time.Friday:    "Friday",
		time.Saturday:  "Saturday",
	}

	// Calculate the number of each day of week in the date range
	dayOfWeekOccurrences := calculateDayOccurrences(start, end)

	for day := time.Sunday; day <= time.Saturday; day++ {
		dayName := dayNames[day]
		totalCount := dayOfWeekCounts[day]
		occurrences := dayOfWeekOccurrences[day]

		var average float64
		if occurrences > 0 {
			average = float64(totalCount) / float64(occurrences)
		}

		// Round to 2 decimal places
		average = math.Round(average*100) / 100

		weeklyData = append(weeklyData, WeeklyDayData{
			DayOfWeek: dayName,
			Count:     totalCount,
			Average:   average,
		})

		// Find peak day (day with highest total count)
		if totalCount > peakDayCount {
			peakDayCount = totalCount
			peakDay = dayName
		}
	}

	return &PeakDayAnalysisResponse{
		StartDate:    startDate,
		EndDate:      endDate,
		PeakDay:      peakDay,
		PeakDayCount: peakDayCount,
		WeeklyData:   weeklyData,
	}, nil
}

// Helper function to calculate how many times each day of week occurs in the date range
func calculateDayOccurrences(start, end time.Time) map[time.Weekday]int {
	occurrences := make(map[time.Weekday]int)

	// Initialize all days to 0
	for i := time.Sunday; i <= time.Saturday; i++ {
		occurrences[i] = 0
	}

	// Count occurrences of each day of week in the range
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		occurrences[d.Weekday()]++
	}

	return occurrences
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

type VisitorsByAttractionResponse struct {
	StartDate      string                  `json:"startDate"`
	EndDate        string                  `json:"endDate"`
	TotalVisitors  int                     `json:"totalVisitors"`
	AttractionData []AttractionVisitorData `json:"attractionData"`
}

type AttractionVisitorData struct {
	TicketGroupId   uint    `json:"ticketGroupId"`
	TicketGroupName string  `json:"ticketGroupName"`
	TotalVisitors   int     `json:"totalVisitors"`
	Percentage      float64 `json:"percentage"`
}

func (o *OrderTicketGroupService) GetVisitorsByAttraction(startDate, endDate string) (*VisitorsByAttractionResponse, error) {
	// Parse dates
	start, err := time.Parse(utils.DateOnlyFormat, startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid startDate format: %w", err)
	}

	end, err := time.Parse(utils.DateOnlyFormat, endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid endDate format: %w", err)
	}

	// Get all orders
	allOrders, err := o.orderTicketGroupRepo.FindAllSuccessfulOrders()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve orders: %w", err)
	}

	// Get all ticket groups to ensure we show all attractions (even with 0 visitors)
	allTicketGroups, err := o.ticketGroupRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve ticket groups: %w", err)
	}

	// Create a map to store visitor counts by ticket group
	ticketGroupVisitors := make(map[uint]int)
	ticketGroupNames := make(map[uint]string)

	// Initialize all ticket groups with 0 visitors and store their names
	for _, ticketGroup := range allTicketGroups {
		ticketGroupVisitors[ticketGroup.TicketGroupId] = 0
		// Use English name, fallback to BM if English is empty
		if ticketGroup.GroupNameEn != "" {
			ticketGroupNames[ticketGroup.TicketGroupId] = ticketGroup.GroupNameEn
		} else {
			ticketGroupNames[ticketGroup.TicketGroupId] = ticketGroup.GroupNameBm
		}
	}

	// Count visitors for each ticket group within the date range
	totalVisitors := 0

	for _, order := range allOrders {
		if len(order.OrderTicketInfos) == 0 {
			continue
		}

		// Get admit date from first ticket info
		admitDateStr := order.OrderTicketInfos[0].AdmitDate
		admitDate, err := time.Parse("2006-01-02", admitDateStr)
		if err != nil {
			continue
		}

		// Check if admit date falls within the specified range
		if admitDate.After(start.Add(-time.Second)) && admitDate.Before(end.Add(24*time.Hour)) {
			ticketCount := len(order.OrderTicketInfos)
			ticketGroupId := order.TicketGroupId

			ticketGroupVisitors[ticketGroupId] += ticketCount
			totalVisitors += ticketCount
		}
	}

	// Build response data
	attractionData := make([]AttractionVisitorData, 0)

	for ticketGroupId, visitorCount := range ticketGroupVisitors {
		// Calculate percentage
		var percentage float64
		if totalVisitors > 0 {
			percentage = (float64(visitorCount) / float64(totalVisitors)) * 100
			// Round to 2 decimal places
			percentage = math.Round(percentage*100) / 100
		}

		attractionData = append(attractionData, AttractionVisitorData{
			TicketGroupId:   ticketGroupId,
			TicketGroupName: ticketGroupNames[ticketGroupId],
			TotalVisitors:   visitorCount,
			Percentage:      percentage,
		})
	}

	// Sort by total visitors in descending order (most popular first)
	sort.Slice(attractionData, func(i, j int) bool {
		return attractionData[i].TotalVisitors > attractionData[j].TotalVisitors
	})

	return &VisitorsByAttractionResponse{
		StartDate:      startDate,
		EndDate:        endDate,
		TotalVisitors:  totalVisitors,
		AttractionData: attractionData,
	}, nil
}

type VisitorsByAgeGroupResponse struct {
	StartDate     string         `json:"startDate"`
	EndDate       string         `json:"endDate"`
	TotalVisitors int            `json:"totalVisitors"`
	AgeGroupData  []AgeGroupData `json:"ageGroupData"`
}

type AgeGroupData struct {
	AgeGroup      string  `json:"ageGroup"`
	TotalVisitors int     `json:"totalVisitors"`
	Percentage    float64 `json:"percentage"`
}

func (o *OrderTicketGroupService) GetVisitorsByAgeGroup(startDate, endDate string) (*VisitorsByAgeGroupResponse, error) {
	// Parse dates
	start, err := time.Parse(utils.DateOnlyFormat, startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid startDate format: %w", err)
	}

	end, err := time.Parse(utils.DateOnlyFormat, endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid endDate format: %w", err)
	}

	// Get all orders
	allOrders, err := o.orderTicketGroupRepo.FindAllSuccessfulOrders()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve orders: %w", err)
	}

	// Initialize age group counters
	ageGroupCounts := map[string]int{
		"0-12":    0,
		"13-17":   0,
		"18-35":   0,
		"36-50":   0,
		"51+":     0,
		"Unknown": 0, // For cases where age cannot be determined
	}

	totalVisitors := 0
	currentYear := time.Now().Year()

	for _, order := range allOrders {
		if len(order.OrderTicketInfos) == 0 {
			continue
		}

		// Get admit date from first ticket info
		admitDateStr := order.OrderTicketInfos[0].AdmitDate
		admitDate, err := time.Parse("2006-01-02", admitDateStr)
		if err != nil {
			continue
		}

		// Check if admit date falls within the specified range
		if admitDate.After(start.Add(-time.Second)) && admitDate.Before(end.Add(24*time.Hour)) {
			ticketCount := len(order.OrderTicketInfos)

			// Get customer age from IC
			age := utils.ExtractAgeFromMalaysianIC(order.Customer.IdentificationNo, currentYear)
			ageGroup := utils.CategorizeAge(age)

			ageGroupCounts[ageGroup] += ticketCount
			totalVisitors += ticketCount
		}
	}

	// Build response data
	ageGroupData := make([]AgeGroupData, 0)

	// Define the order of age groups for consistent output
	ageGroupOrder := []string{"0-12", "13-17", "18-35", "36-50", "51+", "Unknown"}

	for _, ageGroup := range ageGroupOrder {
		visitorCount := ageGroupCounts[ageGroup]

		// Calculate percentage
		var percentage float64
		if totalVisitors > 0 {
			percentage = (float64(visitorCount) / float64(totalVisitors)) * 100
			// Round to 2 decimal places
			percentage = math.Round(percentage*100) / 100
		}

		ageGroupData = append(ageGroupData, AgeGroupData{
			AgeGroup:      ageGroup,
			TotalVisitors: visitorCount,
			Percentage:    percentage,
		})
	}

	return &VisitorsByAgeGroupResponse{
		StartDate:     startDate,
		EndDate:       endDate,
		TotalVisitors: totalVisitors,
		AgeGroupData:  ageGroupData,
	}, nil
}

type VisitorsByNationalityResponse struct {
	StartDate       string            `json:"startDate"`
	EndDate         string            `json:"endDate"`
	TotalVisitors   int               `json:"totalVisitors"`
	NationalityData []NationalityData `json:"nationalityData"`
}

type NationalityData struct {
	Nationality   string  `json:"nationality"`
	TotalVisitors int     `json:"totalVisitors"`
	Percentage    float64 `json:"percentage"`
}

func (o *OrderTicketGroupService) GetVisitorsByNationality(startDate, endDate string) (*VisitorsByNationalityResponse, error) {
	// Parse dates
	start, err := time.Parse(utils.DateOnlyFormat, startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid startDate format: %w", err)
	}

	end, err := time.Parse(utils.DateOnlyFormat, endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid endDate format: %w", err)
	}

	// Get all orders
	allOrders, err := o.orderTicketGroupRepo.FindAllSuccessfulOrders()
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve orders: %w", err)
	}

	// Initialize nationality counters
	nationalityCounts := map[string]int{
		"Local":         0, // Malaysian
		"International": 0, // Non-Malaysian
	}

	totalVisitors := 0

	for _, order := range allOrders {
		if len(order.OrderTicketInfos) == 0 {
			continue
		}

		// Get admit date from first ticket info
		admitDateStr := order.OrderTicketInfos[0].AdmitDate
		admitDate, err := time.Parse("2006-01-02", admitDateStr)
		if err != nil {
			continue
		}

		// Check if admit date falls within the specified range
		if admitDate.After(start.Add(-time.Second)) && admitDate.Before(end.Add(24*time.Hour)) {
			ticketCount := len(order.OrderTicketInfos)

			// Determine nationality based on IC format
			nationality := utils.DetermineNationality(order.Customer.IdentificationNo)

			nationalityCounts[nationality] += ticketCount
			totalVisitors += ticketCount
		}
	}

	// Build response data
	nationalityData := make([]NationalityData, 0)

	// Define the order for consistent output (Local first, then International)
	nationalityOrder := []string{"Local", "International"}

	for _, nationality := range nationalityOrder {
		visitorCount := nationalityCounts[nationality]

		// Calculate percentage
		var percentage float64
		if totalVisitors > 0 {
			percentage = (float64(visitorCount) / float64(totalVisitors)) * 100
			// Round to 2 decimal places
			percentage = math.Round(percentage*100) / 100
		}

		nationalityData = append(nationalityData, NationalityData{
			Nationality:   nationality,
			TotalVisitors: visitorCount,
			Percentage:    percentage,
		})
	}

	return &VisitorsByNationalityResponse{
		StartDate:       startDate,
		EndDate:         endDate,
		TotalVisitors:   totalVisitors,
		NationalityData: nationalityData,
	}, nil
}
