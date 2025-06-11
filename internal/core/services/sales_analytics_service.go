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
	orderRepo         *repositories.OrderTicketGroupRepository
	ticketGroupRepo   *repositories.TicketGroupRepository
	customerRepo      repositories.CustomerRepository
	ticketVariantRepo *repositories.TicketVariantRepository
}

func NewSalesAnalyticsService(
	orderRepo *repositories.OrderTicketGroupRepository,
	ticketGroupRepo *repositories.TicketGroupRepository,
	customerRepo repositories.CustomerRepository,
	ticketVariantRepo *repositories.TicketVariantRepository,
) *SalesAnalyticsService {
	return &SalesAnalyticsService{
		orderRepo:         orderRepo,
		ticketGroupRepo:   ticketGroupRepo,
		customerRepo:      customerRepo,
		ticketVariantRepo: ticketVariantRepo,
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

// GetSalesByTicketGroup retrieves sales by ticket group analysis
func (s *SalesAnalyticsService) GetSalesByTicketGroup(startDate, endDate string) (*dto.SalesByTicketGroupResponse, error) {
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

	// Generate ticket group trend with percentages
	ticketGroupTrend := s.generateTicketGroupTrend(allTicketGroups, successfulOrders)

	return &dto.SalesByTicketGroupResponse{
		TicketGroupTrend: ticketGroupTrend,
	}, nil
}

// GetSalesByAgeGroup retrieves sales by age group analysis
func (s *SalesAnalyticsService) GetSalesByAgeGroup(startDate, endDate string) (*dto.SalesByAgeGroupResponse, error) {
	// Validate date range
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// Get successful orders within date range
	successfulOrders, err := s.getSuccessfulOrders(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get successful orders: %w", err)
	}

	// Generate age group trend
	ageGroupTrend := s.generateAgeGroupTrend(successfulOrders)

	return &dto.SalesByAgeGroupResponse{
		SalesByAgeGroupTrend: ageGroupTrend,
	}, nil
}

// GetSalesByPaymentMethod retrieves sales by payment method analysis
func (s *SalesAnalyticsService) GetSalesByPaymentMethod(startDate, endDate string) (*dto.SalesByPaymentMethodResponse, error) {
	// Validate date range
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// Get successful orders within date range
	successfulOrders, err := s.getSuccessfulOrders(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get successful orders: %w", err)
	}

	// Generate payment method trend
	paymentMethodTrend := s.generatePaymentMethodTrend(successfulOrders)

	return &dto.SalesByPaymentMethodResponse{
		PaymentMethodTrend: paymentMethodTrend,
	}, nil
}

// GetSalesByNationality retrieves sales by nationality analysis
func (s *SalesAnalyticsService) GetSalesByNationality(startDate, endDate string) (*dto.SalesByNationalityResponse, error) {
	// Validate date range
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// Get successful orders within date range
	successfulOrders, err := s.getSuccessfulOrders(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get successful orders: %w", err)
	}

	// Generate nationality trend
	nationalityTrend := s.generateNationalityTrend(successfulOrders)

	return &dto.SalesByNationalityResponse{
		SalesByNationalityTrend: nationalityTrend,
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

// generateTicketGroupTrend generates ticket group trend with sales percentages
func (s *SalesAnalyticsService) generateTicketGroupTrend(allTicketGroups []models.TicketGroup, successfulOrders []models.OrderTicketGroup) []dto.TicketGroupTrend {
	// Create a map to track ticket groups and their sales data
	ticketGroupSales := make(map[uint]*dto.TicketGroupTrend)
	totalOrders := len(successfulOrders)

	// Initialize all ticket groups with zero values
	for _, ticketGroup := range allTicketGroups {
		ticketGroupSales[ticketGroup.TicketGroupId] = &dto.TicketGroupTrend{
			TicketGroupName: ticketGroup.GroupNameEn,
			TotalRevenue:    0,
			TotalOrders:     0,
			SalesPercentage: 0,
		}
	}

	// Aggregate sales data from successful orders
	for _, order := range successfulOrders {
		if salesData, exists := ticketGroupSales[order.TicketGroupId]; exists {
			salesData.TotalRevenue += order.TotalAmount
			salesData.TotalOrders++
		}
	}

	// Calculate percentages and convert map to slice
	var ticketGroupTrend []dto.TicketGroupTrend
	for _, salesData := range ticketGroupSales {
		if totalOrders > 0 {
			salesData.SalesPercentage = math.Round((float64(salesData.TotalOrders)/float64(totalOrders))*100*100) / 100
		}
		ticketGroupTrend = append(ticketGroupTrend, *salesData)
	}

	return ticketGroupTrend
}

// generateAgeGroupTrend generates age group trend with sales percentages
func (s *SalesAnalyticsService) generateAgeGroupTrend(successfulOrders []models.OrderTicketGroup) []dto.AgeGroupTrend {
	ageGroupSales := make(map[string]*dto.AgeGroupTrend)
	totalOrders := len(successfulOrders)
	currentYear := time.Now().Year()

	// Initialize age groups (exclude "Unknown")
	ageGroups := []string{"0-12", "13-17", "18-35", "36-50", "51+"}
	for _, ageGroup := range ageGroups {
		ageGroupSales[ageGroup] = &dto.AgeGroupTrend{
			AgeGroup:        ageGroup,
			TotalRevenue:    0,
			TotalOrders:     0,
			SalesPercentage: 0,
		}
	}

	// Aggregate sales data by age group (only for valid Malaysian ICs)
	for _, order := range successfulOrders {
		// Only process if it's a valid Malaysian IC
		if utils.IsMalaysianIC(order.Customer.IdentificationNo) {
			age := utils.ExtractAgeFromMalaysianIC(order.Customer.IdentificationNo, currentYear)
			ageGroup := utils.CategorizeAge(age)

			// Only process if it's not "Unknown" (which means it's a valid age)
			if ageGroup != "Unknown" {
				if salesData, exists := ageGroupSales[ageGroup]; exists {
					salesData.TotalRevenue += order.TotalAmount
					salesData.TotalOrders++
				}
			}
		}
		// If not Malaysian IC or Unknown age group, we skip this order
	}

	// Calculate percentages and convert to slice
	var ageGroupTrend []dto.AgeGroupTrend
	for _, ageGroup := range ageGroups { // Use the same order as initialization
		salesData := ageGroupSales[ageGroup]
		if totalOrders > 0 {
			salesData.SalesPercentage = math.Round((float64(salesData.TotalOrders)/float64(totalOrders))*100*100) / 100
		}
		// Include all age groups, even those with 0 orders
		ageGroupTrend = append(ageGroupTrend, *salesData)
	}

	return ageGroupTrend
}

// generatePaymentMethodTrend generates payment method trend with sales percentages
func (s *SalesAnalyticsService) generatePaymentMethodTrend(successfulOrders []models.OrderTicketGroup) []dto.PaymentMethodTrend {
	paymentMethodSales := make(map[string]*dto.PaymentMethodTrend)
	totalOrders := len(successfulOrders)

	// Initialize payment methods
	paymentMethods := []string{"Credit / Debit Card", "FPX"}
	for _, method := range paymentMethods {
		paymentMethodSales[method] = &dto.PaymentMethodTrend{
			PaymentMethod:   method,
			TotalRevenue:    0,
			TotalOrders:     0,
			SalesPercentage: 0,
		}
	}

	// Aggregate sales data by payment method
	for _, order := range successfulOrders {
		var paymentMethod string

		// If bank_code and bank_name are present, it's FPX
		if order.BankCode.Valid && order.BankName.Valid &&
			order.BankCode.String != "" && order.BankName.String != "" {
			paymentMethod = "FPX"
		} else {
			paymentMethod = "Credit / Debit Card"
		}

		if salesData, exists := paymentMethodSales[paymentMethod]; exists {
			salesData.TotalRevenue += order.TotalAmount
			salesData.TotalOrders++
		}
	}

	// Calculate percentages and convert to slice
	var paymentMethodTrend []dto.PaymentMethodTrend
	for _, salesData := range paymentMethodSales {
		if totalOrders > 0 {
			salesData.SalesPercentage = math.Round((float64(salesData.TotalOrders)/float64(totalOrders))*100*100) / 100
		}
		// Only include payment methods that have orders
		if salesData.TotalOrders > 0 {
			paymentMethodTrend = append(paymentMethodTrend, *salesData)
		}
	}

	return paymentMethodTrend
}

// generateNationalityTrend generates nationality trend with sales percentages
func (s *SalesAnalyticsService) generateNationalityTrend(successfulOrders []models.OrderTicketGroup) []dto.NationalityTrend {
	nationalitySales := make(map[string]*dto.NationalityTrend)
	totalOrders := len(successfulOrders)

	// Initialize nationalities
	nationalities := []string{"Local", "International"}
	for _, nationality := range nationalities {
		nationalitySales[nationality] = &dto.NationalityTrend{
			Nationality:     nationality,
			TotalRevenue:    0,
			TotalOrders:     0,
			SalesPercentage: 0,
		}
	}

	// Aggregate sales data by nationality
	for _, order := range successfulOrders {
		nationality := utils.DetermineNationality(order.Customer.IdentificationNo)

		if salesData, exists := nationalitySales[nationality]; exists {
			salesData.TotalRevenue += order.TotalAmount
			salesData.TotalOrders++
		}
	}

	// Calculate percentages and convert to slice
	var nationalityTrend []dto.NationalityTrend
	for _, salesData := range nationalitySales {
		if totalOrders > 0 {
			salesData.SalesPercentage = math.Round((float64(salesData.TotalOrders)/float64(totalOrders))*100*100) / 100
		}
		// Only include nationalities that have orders
		if salesData.TotalOrders > 0 {
			nationalityTrend = append(nationalityTrend, *salesData)
		}
	}

	return nationalityTrend
}
