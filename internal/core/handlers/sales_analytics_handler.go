// File: j-ticketing/internal/core/handlers/sales_analytics_handler.go
package handlers

import (
	dto "j-ticketing/internal/core/dto/sales_analytics"
	service "j-ticketing/internal/core/services"
	"j-ticketing/pkg/errors"
	"j-ticketing/pkg/models"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type SalesAnalyticsHandler struct {
	salesAnalyticsService *service.SalesAnalyticsService
}

func NewSalesAnalyticsHandler(salesAnalyticsService *service.SalesAnalyticsService) *SalesAnalyticsHandler {
	return &SalesAnalyticsHandler{
		salesAnalyticsService: salesAnalyticsService,
	}
}

// GetTotalRevenue handles GET /api/analytics/totalRevenue
func (h *SalesAnalyticsHandler) GetTotalRevenue(c *fiber.Ctx) error {
	// Parse and validate request
	var req dto.SalesAnalyticsRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request parameters",
			"message": err.Error(),
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		if validationErr, ok := err.(*errors.ValidationError); ok {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error":  "Validation failed",
				"fields": validationErr.FieldErrors,
			})
		}
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"message": err.Error(),
		})
	}

	// Get total revenue from service
	response, err := h.salesAnalyticsService.GetTotalRevenue(req.StartDate, req.EndDate)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve total revenue",
			"message": err.Error(),
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}

// GetTotalOrders handles GET /api/analytics/totalOrders
func (h *SalesAnalyticsHandler) GetTotalOrders(c *fiber.Ctx) error {
	// Parse and validate request
	var req dto.SalesAnalyticsRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request parameters",
			"message": err.Error(),
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		if validationErr, ok := err.(*errors.ValidationError); ok {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error":  "Validation failed",
				"fields": validationErr.FieldErrors,
			})
		}
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"message": err.Error(),
		})
	}

	// Get total orders from service
	response, err := h.salesAnalyticsService.GetTotalOrders(req.StartDate, req.EndDate)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve total orders",
			"message": err.Error(),
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}

// GetAvgOrderValue handles GET /api/analytics/avgOrderValue
func (h *SalesAnalyticsHandler) GetAvgOrderValue(c *fiber.Ctx) error {
	// Parse and validate request
	var req dto.SalesAnalyticsRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request parameters",
			"message": err.Error(),
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		if validationErr, ok := err.(*errors.ValidationError); ok {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error":  "Validation failed",
				"fields": validationErr.FieldErrors,
			})
		}
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"message": err.Error(),
		})
	}

	// Get average order value from service
	response, err := h.salesAnalyticsService.GetAvgOrderValue(req.StartDate, req.EndDate)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve average order value",
			"message": err.Error(),
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}

// GetTopSalesProduct handles GET /api/analytics/topSalesProduct
func (h *SalesAnalyticsHandler) GetTopSalesProduct(c *fiber.Ctx) error {
	// Parse and validate request
	var req dto.SalesAnalyticsRequest
	if err := c.QueryParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Invalid request parameters",
			"message": err.Error(),
		})
	}

	// Validate request
	if err := req.Validate(); err != nil {
		if validationErr, ok := err.(*errors.ValidationError); ok {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error":  "Validation failed",
				"fields": validationErr.FieldErrors,
			})
		}
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"message": err.Error(),
		})
	}

	// Get top sales product from service
	response, err := h.salesAnalyticsService.GetTopSalesProduct(req.StartDate, req.EndDate)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve top sales product",
			"message": err.Error(),
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}
