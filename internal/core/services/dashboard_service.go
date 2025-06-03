// File: j-ticketing/internal/core/services/dashboard_service.go
package service

import (
	"fmt"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/utils"
	"time"
)

type DashboardService struct {
	orderRepo        *repositories.OrderTicketGroupRepository
	ticketGroupRepo  *repositories.TicketGroupRepository
	customerRepo     repositories.CustomerRepository
	notificationRepo *repositories.NotificationRepository
}

func NewDashboardService(
	orderRepo *repositories.OrderTicketGroupRepository,
	ticketGroupRepo *repositories.TicketGroupRepository,
	customerRepo repositories.CustomerRepository,
	notificationRepo *repositories.NotificationRepository,
) *DashboardService {
	return &DashboardService{
		orderRepo:        orderRepo,
		ticketGroupRepo:  ticketGroupRepo,
		customerRepo:     customerRepo,
		notificationRepo: notificationRepo,
	}
}

// OrderAnalysis represents the order analysis data structure
type OrderAnalysis struct {
	TotalSumOrders int          `json:"totalSumOrders"`
	TotalSumAmount float64      `json:"totalSumAmount"`
	OrderTrends    []OrderTrend `json:"orderTrends"`
}

type OrderTrend struct {
	TrendType    string           `json:"trendType"`
	TrendResults []OrderTrendData `json:"trendResults"`
}

// Separate structs for different trend types
type OrderTrendData struct {
	Date        string   `json:"date"`
	TotalOrders *int     `json:"totalOrders,omitempty"`
	TotalAmount *float64 `json:"totalAmount,omitempty"`
}

// ProductAnalysis represents the product analysis data structure
type ProductAnalysis struct {
	TotalTicketGroup   int            `json:"totalTicketGroup"`
	TotalSumTicketSold int            `json:"totalSumTicketSold"`
	ProductTrends      []ProductTrend `json:"productTrends"`
}

type ProductTrend struct {
	TrendType    string             `json:"trendType"`
	TrendResults []ProductTrendData `json:"trendResults"`
}

type ProductTrendData struct {
	Date         string `json:"date"`
	QuantitySold int    `json:"quantitySold"`
}

// CustomerAnalysis represents the customer analysis data structure
type CustomerAnalysis struct {
	TotalCustomers int             `json:"totalCustomers"`
	TotalMembers   int             `json:"totalMembers"`
	CustomerTrend  []CustomerTrend `json:"customerTrend"`
}

type CustomerTrend struct {
	TrendType    string              `json:"trendType"`
	TrendResults []CustomerTrendData `json:"trendResults"`
}

type CustomerTrendData struct {
	Date  string `json:"date"`
	Total int    `json:"total"`
}

// NotificationAnalysis represents the notification analysis data structure
type NotificationAnalysis struct {
	TotalNotifications       int                   `json:"totalNotifications"`
	TotalUnreadNotifications int                   `json:"totalUnreadNotifications"`
	NotificationLogs         []NotificationLogData `json:"notificationLogs"`
}

type NotificationLogData struct {
	NotificationId uint   `json:"notificationId"`
	PerformedBy    string `json:"performedBy"`
	AuthorityLevel string `json:"authorityLevel"`
	Type           string `json:"type"`
	Title          string `json:"title"`
	Message        string `json:"message"`
	Date           string `json:"date"`
	IsRead         bool   `json:"isRead"`
	IsDeleted      bool   `json:"isDeleted"`
}

// GetDashboardAnalysis retrieves all dashboard analysis data
func (s *DashboardService) GetDashboardAnalysis(startDate, endDate string) (map[string]interface{}, error) {
	// Validate date format and range
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// Get all analysis data concurrently
	orderAnalysis, err := s.getOrderAnalysis(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get order analysis: %w", err)
	}

	productAnalysis, err := s.getProductAnalysis(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get product analysis: %w", err)
	}

	customerAnalysis, err := s.getCustomerAnalysis(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get customer analysis: %w", err)
	}

	notificationAnalysis, err := s.getNotificationAnalysis(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification analysis: %w", err)
	}

	// Return the actual objects, not JSON strings
	return map[string]interface{}{
		"ordersAnalysis":   orderAnalysis,
		"productAnalysis":  productAnalysis,
		"customerAnalysis": customerAnalysis,
		"notifications":    notificationAnalysis,
	}, nil
}

// validateDateRange validates the date format and ensures endDate is after startDate
func (s *DashboardService) validateDateRange(startDate, endDate string) error {
	// Parse dates using the expected format
	start, err := time.Parse(utils.DateOnlyFormat, startDate)
	if err != nil {
		return fmt.Errorf("invalid startDate format. Expected yyyy-MM-dd, got: %s", startDate)
	}

	end, err := time.Parse(utils.DateOnlyFormat, endDate)
	if err != nil {
		return fmt.Errorf("invalid endDate format. Expected yyyy-MM-dd, got: %s", endDate)
	}

	if end.Before(start) {
		return fmt.Errorf("endDate cannot be earlier than startDate")
	}

	return nil
}

// getOrderAnalysis retrieves order analysis data
func (s *DashboardService) getOrderAnalysis(startDate, endDate string) (*OrderAnalysis, error) {
	// Get all orders within the date range with success status
	orders, err := s.orderRepo.FindByDateRange("", startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Filter successful orders within date range
	var successfulOrders []models.OrderTicketGroup
	for _, order := range orders {
		if order.TransactionStatus == "success" && s.isWithinDateRange(order.TransactionDate, startDate, endDate) {
			successfulOrders = append(successfulOrders, order)
		}
	}

	// Calculate totals
	totalOrders := len(successfulOrders)
	var totalAmount float64
	for _, order := range successfulOrders {
		totalAmount += order.TotalAmount
	}

	// Generate daily trends
	orderTrends := s.generateOrderTrends(successfulOrders, startDate, endDate)

	return &OrderAnalysis{
		TotalSumOrders: totalOrders,
		TotalSumAmount: totalAmount,
		OrderTrends:    orderTrends,
	}, nil
}

// getProductAnalysis retrieves product analysis data
func (s *DashboardService) getProductAnalysis(startDate, endDate string) (*ProductAnalysis, error) {
	// Get ALL ticket groups (not filtered by creation date)
	allTicketGroups, err := s.ticketGroupRepo.FindAll()
	if err != nil {
		return nil, err
	}

	// Total ticket group count is ALL ticket groups, not filtered by period
	totalTicketGroup := len(allTicketGroups)

	// Get all successful orders within period for sales calculations
	orders, err := s.orderRepo.FindByDateRange("", startDate, endDate)
	if err != nil {
		return nil, err
	}

	var successfulOrders []models.OrderTicketGroup
	for _, order := range orders {
		if order.TransactionStatus == "success" && s.isWithinDateRange(order.TransactionDate, startDate, endDate) {
			successfulOrders = append(successfulOrders, order)
		}
	}

	// Calculate total tickets sold and generate trends
	totalSumTicketSold := 0
	productTrends := make([]ProductTrend, 0)

	// Create a map to track ticket groups and their daily sales
	ticketGroupSales := make(map[uint]map[string]int) // ticketGroupId -> date -> quantity

	for _, order := range successfulOrders {
		// Initialize if not exists
		if ticketGroupSales[order.TicketGroupId] == nil {
			ticketGroupSales[order.TicketGroupId] = make(map[string]int)
		}

		// Get order date
		orderDate := s.extractDateFromTransactionDate(order.TransactionDate)

		// Sum quantities from order ticket infos
		for _, info := range order.OrderTicketInfos {
			ticketGroupSales[order.TicketGroupId][orderDate] += info.QuantityBought
			totalSumTicketSold += info.QuantityBought
		}
	}

	// Generate trends for ALL ticket groups (even those with 0 sales)
	for _, ticketGroup := range allTicketGroups {
		dailySales := ticketGroupSales[ticketGroup.TicketGroupId]
		if dailySales == nil {
			dailySales = make(map[string]int) // Empty sales data
		}

		trend := ProductTrend{
			TrendType:    ticketGroup.GroupNameEn,
			TrendResults: s.generateProductTrendData(dailySales, startDate, endDate),
		}
		productTrends = append(productTrends, trend)
	}

	return &ProductAnalysis{
		TotalTicketGroup:   totalTicketGroup,
		TotalSumTicketSold: totalSumTicketSold,
		ProductTrends:      productTrends,
	}, nil
}

// getCustomerAnalysis retrieves customer analysis data
func (s *DashboardService) getCustomerAnalysis(startDate, endDate string) (*CustomerAnalysis, error) {
	// Get all customers ordered by creation date
	allCustomers, err := s.customerRepo.FindAll()
	if err != nil {
		return nil, err
	}

	// Convert start and end dates for filtering
	startTime, _ := time.Parse(utils.DateOnlyFormat, startDate)
	endTime, _ := time.Parse(utils.DateOnlyFormat, endDate)
	endTime = endTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second) // End of day

	// Filter customers for card metrics (created within period)
	var customersInPeriod []models.Customer
	var membersInPeriod []models.Customer

	for _, customer := range allCustomers {
		// Convert UTC CreatedAt to Malaysia timezone
		malaysiaTime, err := utils.ToMalaysiaTime(customer.CreatedAt)
		if err != nil {
			continue
		}

		// Filter for card metrics (within period only)
		if malaysiaTime.After(startTime.Add(-time.Second)) && malaysiaTime.Before(endTime) {
			customersInPeriod = append(customersInPeriod, customer)

			// Check if customer is a member (has password)
			if customer.Password.Valid && customer.Password.String != "" {
				membersInPeriod = append(membersInPeriod, customer)
			}
		}
	}

	// Generate customer trends
	customerTrends := s.generateCustomerTrendsNew(allCustomers, startDate, endDate)

	return &CustomerAnalysis{
		TotalCustomers: len(customersInPeriod),
		TotalMembers:   len(membersInPeriod),
		CustomerTrend:  customerTrends,
	}, nil
}

// getNotificationAnalysis retrieves notification analysis data
func (s *DashboardService) getNotificationAnalysis(startDate, endDate string) (*NotificationAnalysis, error) {
	// Get total count of all notifications (read + unread)
	totalCount, err := s.notificationRepo.CountAll()
	if err != nil {
		return nil, fmt.Errorf("failed to count all notifications: %w", err)
	}

	// Get all unread notifications
	unreadNotifications, err := s.notificationRepo.FindUnread()
	if err != nil {
		return nil, fmt.Errorf("failed to get unread notifications: %w", err)
	}

	// Count total unread notifications
	totalUnreadCount := len(unreadNotifications)

	// Limit to latest 10 notifications for logs (they're already ordered by created_at DESC)
	var notificationsForLogs []models.Notification
	if len(unreadNotifications) > 10 {
		notificationsForLogs = unreadNotifications[:10]
	} else {
		notificationsForLogs = unreadNotifications
	}

	// Convert to response format
	notificationLogs := make([]NotificationLogData, 0, len(notificationsForLogs))
	for _, notification := range notificationsForLogs {
		performedBy := ""
		if notification.PerformedBy.Valid {
			performedBy = notification.PerformedBy.String
		}

		message := ""
		if notification.Message.Valid {
			message = notification.Message.String
		}

		notificationLogs = append(notificationLogs, NotificationLogData{
			NotificationId: notification.NotificationId,
			PerformedBy:    performedBy,
			AuthorityLevel: notification.AuthorityLevel,
			Type:           notification.Type,
			Title:          notification.Title,
			Message:        message,
			Date:           notification.Date,
			IsRead:         notification.IsRead,
			IsDeleted:      notification.IsDeleted,
		})
	}

	return &NotificationAnalysis{
		TotalNotifications:       int(totalCount),
		TotalUnreadNotifications: totalUnreadCount,
		NotificationLogs:         notificationLogs,
	}, nil
}

// Helper methods

func (s *DashboardService) isWithinDateRange(transactionDate, startDate, endDate string) bool {
	// Extract date from transaction date (format: 2025-05-21 12:49:01)
	if len(transactionDate) < 10 {
		return false
	}

	dateOnly := transactionDate[:10] // Extract yyyy-MM-dd part
	return s.isDateInRange(dateOnly, startDate, endDate)
}

func (s *DashboardService) isDateInRange(date, startDate, endDate string) bool {
	return date >= startDate && date <= endDate
}

func (s *DashboardService) extractDateFromTransactionDate(transactionDate string) string {
	if len(transactionDate) >= 10 {
		return transactionDate[:10]
	}
	return ""
}

func (s *DashboardService) generateOrderTrends(orders []models.OrderTicketGroup, startDate, endDate string) []OrderTrend {
	// Create daily aggregations
	dailyOrders := make(map[string]int)
	dailyAmounts := make(map[string]float64)

	for _, order := range orders {
		date := s.extractDateFromTransactionDate(order.TransactionDate)
		if date != "" {
			dailyOrders[date]++
			dailyAmounts[date] += order.TotalAmount
		}
	}

	// Generate date range
	dates := s.generateDateRange(startDate, endDate)

	// Create trends - separate data for each trend type
	orderTrendData := make([]OrderTrendData, 0)
	amountTrendData := make([]OrderTrendData, 0)

	for _, date := range dates {
		// For Total Order trend - only include totalOrders
		orderCount := dailyOrders[date]
		orderTrendData = append(orderTrendData, OrderTrendData{
			Date:        date,
			TotalOrders: &orderCount,
			TotalAmount: nil,
		})

		// For Total Amount trend - only include totalAmount
		amount := dailyAmounts[date]
		amountTrendData = append(amountTrendData, OrderTrendData{
			Date:        date,
			TotalOrders: nil,
			TotalAmount: &amount,
		})
	}

	return []OrderTrend{
		{
			TrendType:    "Total Order",
			TrendResults: orderTrendData,
		},
		{
			TrendType:    "Total Amount",
			TrendResults: amountTrendData,
		},
	}
}

func (s *DashboardService) generateProductTrendData(dailySales map[string]int, startDate, endDate string) []ProductTrendData {
	dates := s.generateDateRange(startDate, endDate)
	results := make([]ProductTrendData, 0)

	for _, date := range dates {
		results = append(results, ProductTrendData{
			Date:         date,
			QuantitySold: dailySales[date],
		})
	}

	return results
}

func (s *DashboardService) generateCustomerTrendsNew(allCustomers []models.Customer, startDate, endDate string) []CustomerTrend {
	dates := s.generateDateRange(startDate, endDate)

	// Count customers by date (for new customers calculation)
	dailyNewCustomers := make(map[string]int)

	// Calculate cumulative totals efficiently
	// First, get all customers created before start date
	startTime, _ := time.Parse(utils.DateOnlyFormat, startDate)
	customersBeforeStart := 0

	for _, customer := range allCustomers {
		// Convert UTC CreatedAt to Malaysia timezone
		malaysiaTime, err := utils.ToMalaysiaTime(customer.CreatedAt)
		if err != nil {
			continue
		}

		dateStr := malaysiaTime.Format(utils.DateOnlyFormat)

		if malaysiaTime.Before(startTime) {
			customersBeforeStart++
		} else {
			// Count for daily new customers
			dailyNewCustomers[dateStr]++
		}
	}

	// Generate trend data
	newCustomerTrendData := make([]CustomerTrendData, 0)
	totalCustomerTrendData := make([]CustomerTrendData, 0)

	runningTotal := customersBeforeStart

	for _, date := range dates {
		// New customers for this date
		newCustomers := dailyNewCustomers[date]

		// Add new customers to running total
		runningTotal += newCustomers

		newCustomerTrendData = append(newCustomerTrendData, CustomerTrendData{
			Date:  date,
			Total: newCustomers,
		})

		totalCustomerTrendData = append(totalCustomerTrendData, CustomerTrendData{
			Date:  date,
			Total: runningTotal,
		})
	}

	return []CustomerTrend{
		{
			TrendType:    "New Customers",
			TrendResults: newCustomerTrendData,
		},
		{
			TrendType:    "Total Customers",
			TrendResults: totalCustomerTrendData,
		},
	}
}

func (s *DashboardService) generateDateRange(startDate, endDate string) []string {
	start, _ := time.Parse(utils.DateOnlyFormat, startDate)
	end, _ := time.Parse(utils.DateOnlyFormat, endDate)

	var dates []string
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format(utils.DateOnlyFormat))
	}

	return dates
}
