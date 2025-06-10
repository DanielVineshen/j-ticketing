// File: j-ticketing/internal/core/services/sales_analytics_service.go
package service

import (
	"fmt"
	dto "j-ticketing/internal/core/dto/sales_analytics"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/utils"
	"math"
	"time"
)

type SalesAnalyticsService struct {
	orderRepo       *repositories.OrderTicketGroupRepository
	ticketGroupRepo *repositories.TicketGroupRepository
}

func NewSalesAnalyticsService(
	orderRepo *repositories.OrderTicketGroupRepository,
	ticketGroupRepo *repositories.TicketGroupRepository,
) *SalesAnalyticsService {
	return &SalesAnalyticsService{
		orderRepo:       orderRepo,
		ticketGroupRepo: ticketGroupRepo,
	}
}

// GetTotalRevenue retrieves total revenue analysis
func (s *SalesAnalyticsService) GetTotalRevenue(startDate, endDate string) (*dto.TotalRevenueResponse, error) {
	// Validate date range
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// Get successful orders within date range
	successfulOrders, err := s.getSuccessfulOrders(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get successful orders: %w", err)
	}

	// Calculate sum total revenue
	var sumTotalRevenue float64
	for _, order := range successfulOrders {
		sumTotalRevenue += order.TotalAmount
	}

	// Generate revenue trend
	revenueTrend := s.generateRevenueTrend(successfulOrders, startDate, endDate)

	return &dto.TotalRevenueResponse{
		SumTotalRevenue: sumTotalRevenue,
		RevenueTrend:    revenueTrend,
	}, nil
}

// GetTotalOrders retrieves total orders analysis
func (s *SalesAnalyticsService) GetTotalOrders(startDate, endDate string) (*dto.TotalOrdersResponse, error) {
	// Validate date range
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// Get successful orders within date range
	successfulOrders, err := s.getSuccessfulOrders(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get successful orders: %w", err)
	}

	// Calculate sum total orders
	sumTotalOrders := len(successfulOrders)

	// Generate order trend
	orderTrend := s.generateOrderTrend(successfulOrders, startDate, endDate)

	return &dto.TotalOrdersResponse{
		SumTotalOrders: sumTotalOrders,
		OrderTrend:     orderTrend,
	}, nil
}

// GetAvgOrderValue retrieves average order value analysis
func (s *SalesAnalyticsService) GetAvgOrderValue(startDate, endDate string) (*dto.AvgOrderValueResponse, error) {
	// Validate date range
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// Get successful orders within date range
	successfulOrders, err := s.getSuccessfulOrders(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get successful orders: %w", err)
	}

	// Calculate total average order value
	var totalRevenue float64
	totalOrders := len(successfulOrders)

	for _, order := range successfulOrders {
		totalRevenue += order.TotalAmount
	}

	var totalAvgOrderValue float64
	if totalOrders > 0 {
		totalAvgOrderValue = math.Round((totalRevenue/float64(totalOrders))*100) / 100
	}

	// Generate average order value trend
	avgOrderValueTrend := s.generateAvgOrderValueTrend(successfulOrders, startDate, endDate)

	return &dto.AvgOrderValueResponse{
		TotalAvgOrderValue: totalAvgOrderValue,
		AvgOrderValueTrend: avgOrderValueTrend,
	}, nil
}

// GetTopSalesProduct retrieves top sales product analysis
func (s *SalesAnalyticsService) GetTopSalesProduct(startDate, endDate string) (*dto.TopSalesProductResponse, error) {
	// Validate date range
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// Get all ticket groups
	allTicketGroups, err := s.ticketGroupRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get ticket groups: %w", err)
	}

	// Get successful orders within date range
	successfulOrders, err := s.getSuccessfulOrders(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get successful orders: %w", err)
	}

	// Generate top sales product trend
	topSaleProductTrend := s.generateTopSalesProductTrend(allTicketGroups, successfulOrders)

	// Find the top sales product (highest total quantity)
	var topSaleProduct dto.TopSaleProduct
	maxQuantity := 0

	for _, trend := range topSaleProductTrend {
		if trend.TotalQuantity > maxQuantity {
			maxQuantity = trend.TotalQuantity
			topSaleProduct = dto.TopSaleProduct{
				TicketGroupName: trend.TicketGroupName,
				SumTotalOrders:  trend.TotalQuantity,
			}
		}
	}

	return &dto.TopSalesProductResponse{
		TopSaleProduct:      topSaleProduct,
		TopSaleProductTrend: topSaleProductTrend,
	}, nil
}

// Helper methods

// validateDateRange validates the date format and ensures endDate is after startDate
func (s *SalesAnalyticsService) validateDateRange(startDate, endDate string) error {
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

// getSuccessfulOrders retrieves successful orders within the date range
func (s *SalesAnalyticsService) getSuccessfulOrders(startDate, endDate string) ([]models.OrderTicketGroup, error) {
	// Get all orders within the date range
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

	return successfulOrders, nil
}

// isWithinDateRange checks if transaction date is within the specified range
func (s *SalesAnalyticsService) isWithinDateRange(transactionDate, startDate, endDate string) bool {
	// Extract date from transaction date (format: 2025-05-21 12:49:01)
	if len(transactionDate) < 10 {
		return false
	}

	dateOnly := transactionDate[:10] // Extract yyyy-MM-dd part
	return dateOnly >= startDate && dateOnly <= endDate
}

// extractDateFromTransactionDate extracts date part from transaction date
func (s *SalesAnalyticsService) extractDateFromTransactionDate(transactionDate string) string {
	if len(transactionDate) >= 10 {
		return transactionDate[:10]
	}
	return ""
}

// generateDateRange generates all dates between startDate and endDate
func (s *SalesAnalyticsService) generateDateRange(startDate, endDate string) []string {
	start, _ := time.Parse(utils.DateOnlyFormat, startDate)
	end, _ := time.Parse(utils.DateOnlyFormat, endDate)

	var dates []string
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format(utils.DateOnlyFormat))
	}

	return dates
}

// generateRevenueTrend generates revenue trend data
func (s *SalesAnalyticsService) generateRevenueTrend(orders []models.OrderTicketGroup, startDate, endDate string) []dto.RevenueTrend {
	// Create daily revenue aggregations
	dailyRevenue := make(map[string]float64)

	for _, order := range orders {
		date := s.extractDateFromTransactionDate(order.TransactionDate)
		if date != "" {
			dailyRevenue[date] += order.TotalAmount
		}
	}

	// Generate date range and create trend data
	dates := s.generateDateRange(startDate, endDate)
	var revenueTrend []dto.RevenueTrend

	for _, date := range dates {
		revenueTrend = append(revenueTrend, dto.RevenueTrend{
			Date:         date,
			TotalRevenue: dailyRevenue[date],
		})
	}

	return revenueTrend
}

// generateOrderTrend generates order trend data
func (s *SalesAnalyticsService) generateOrderTrend(orders []models.OrderTicketGroup, startDate, endDate string) []dto.OrderTrend {
	// Create daily order count aggregations
	dailyOrders := make(map[string]int)

	for _, order := range orders {
		date := s.extractDateFromTransactionDate(order.TransactionDate)
		if date != "" {
			dailyOrders[date]++
		}
	}

	// Generate date range and create trend data
	dates := s.generateDateRange(startDate, endDate)
	var orderTrend []dto.OrderTrend

	for _, date := range dates {
		orderTrend = append(orderTrend, dto.OrderTrend{
			Date:        date,
			TotalOrders: dailyOrders[date],
		})
	}

	return orderTrend
}

// generateAvgOrderValueTrend generates average order value trend data
func (s *SalesAnalyticsService) generateAvgOrderValueTrend(orders []models.OrderTicketGroup, startDate, endDate string) []dto.AvgOrderValueTrend {
	// Create daily aggregations
	dailyRevenue := make(map[string]float64)
	dailyOrders := make(map[string]int)

	for _, order := range orders {
		date := s.extractDateFromTransactionDate(order.TransactionDate)
		if date != "" {
			dailyRevenue[date] += order.TotalAmount
			dailyOrders[date]++
		}
	}

	// Generate date range and create trend data
	dates := s.generateDateRange(startDate, endDate)
	var avgOrderValueTrend []dto.AvgOrderValueTrend

	for _, date := range dates {
		var avgOrderValue float64
		if dailyOrders[date] > 0 {
			avgOrderValue = math.Round((dailyRevenue[date]/float64(dailyOrders[date]))*100) / 100
		}

		avgOrderValueTrend = append(avgOrderValueTrend, dto.AvgOrderValueTrend{
			Date:          date,
			AvgOrderValue: avgOrderValue,
		})
	}

	return avgOrderValueTrend
}

// generateTopSalesProductTrend generates top sales product trend data
func (s *SalesAnalyticsService) generateTopSalesProductTrend(allTicketGroups []models.TicketGroup, successfulOrders []models.OrderTicketGroup) []dto.TopSaleProductTrend {
	// Create a map to track ticket groups and their sales data
	ticketGroupSales := make(map[uint]*dto.TopSaleProductTrend)

	// Initialize all ticket groups with zero values
	for _, ticketGroup := range allTicketGroups {
		ticketGroupSales[ticketGroup.TicketGroupId] = &dto.TopSaleProductTrend{
			TicketGroupName: ticketGroup.GroupNameEn,
			TotalQuantity:   0,
			TotalRevenue:    0,
			TotalOrders:     0,
		}
	}

	// Aggregate sales data from successful orders
	for _, order := range successfulOrders {
		if salesData, exists := ticketGroupSales[order.TicketGroupId]; exists {
			salesData.TotalRevenue += order.TotalAmount
			salesData.TotalOrders++

			// Sum quantities from order ticket infos
			for _, info := range order.OrderTicketInfos {
				salesData.TotalQuantity += info.QuantityBought
			}
		}
	}

	// Convert map to slice
	var topSaleProductTrend []dto.TopSaleProductTrend
	for _, salesData := range ticketGroupSales {
		topSaleProductTrend = append(topSaleProductTrend, *salesData)
	}

	return topSaleProductTrend
}
