// File: j-ticketing/internal/core/handlers/dashboard_handler.go
package handlers

import (
	dto "j-ticketing/internal/core/dto/dashboard"
	service "j-ticketing/internal/core/services"
	"j-ticketing/pkg/errors"
	"j-ticketing/pkg/models"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct {
	dashboardService *service.DashboardService
}

func NewDashboardHandler(dashboardService *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: dashboardService,
	}
}

// GetDashboardAnalysis handles GET /api/dashboard
func (h *DashboardHandler) GetDashboardAnalysis(c *fiber.Ctx) error {
	// Parse and validate request
	var req dto.DashboardRequest
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

	// Get dashboard analysis from service
	analysisData, err := h.dashboardService.GetDashboardAnalysis(req.StartDate, req.EndDate)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve dashboard analysis",
			"message": err.Error(),
		})
	}

	// Prepare response
	response := dto.DashboardResponse{
		OrdersAnalysis:   analysisData["ordersAnalysis"],
		ProductAnalysis:  analysisData["productAnalysis"],
		CustomerAnalysis: analysisData["customerAnalysis"],
		Notifications:    analysisData["notifications"],
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}
