package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	service "j-ticketing/internal/core/services"
	"j-ticketing/pkg/models"
	"j-ticketing/pkg/utils"
	"time"
)

// OnsiteVisitorsAnalyticsHandler handles onsiteVisitorsAnalytics-related HTTP requests
type OnsiteVisitorsAnalyticsHandler struct {
	orderTicketGroupService *service.OrderTicketGroupService
	customerService         *service.CustomerService
}

// NewOnsiteVisitorsAnalyticsHandler creates a new onsiteVisitorsAnalytics handler
func NewOnsiteVisitorsAnalyticsHandler(orderTicketGroupService *service.OrderTicketGroupService, customerService *service.CustomerService) *OnsiteVisitorsAnalyticsHandler {
	return &OnsiteVisitorsAnalyticsHandler{
		orderTicketGroupService: orderTicketGroupService,
		customerService:         customerService,
	}
}

func (o *OnsiteVisitorsAnalyticsHandler) GetTotalOnsiteVisitors(c *fiber.Ctx) error {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	// Validate required parameters
	if startDate == "" || endDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "startDate and endDate are required parameters",
		})
	}

	// Validate date format
	if !isValidDateFormat(startDate) || !isValidDateFormat(endDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date format. Use YYYY-MM-DD",
		})
	}

	response, err := o.orderTicketGroupService.GetTotalOnsiteVisitorsWithinRange(startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to fetch customer records: %v", err),
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}

func (o *OnsiteVisitorsAnalyticsHandler) GetNewVsReturningVisitors(c *fiber.Ctx) error {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	// Validate required parameters
	if startDate == "" || endDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "startDate and endDate are required parameters",
		})
	}

	// Validate date format
	if !isValidDateFormat(startDate) || !isValidDateFormat(endDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date format. Use YYYY-MM-DD",
		})
	}

	response, err := o.orderTicketGroupService.GetNewVsReturningVisitors(startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to fetch customer records: %v", err),
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}

func (o *OnsiteVisitorsAnalyticsHandler) GetAveragePeakDayAnalysis(c *fiber.Ctx) error {
	return c.JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

func (o *OnsiteVisitorsAnalyticsHandler) GetVisitorsByAttraction(c *fiber.Ctx) error {
	return c.JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

func (o *OnsiteVisitorsAnalyticsHandler) GetVisitorsByAgeGroup(c *fiber.Ctx) error {
	return c.JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

func (o *OnsiteVisitorsAnalyticsHandler) GetVisitorsByNationality(c *fiber.Ctx) error {
	return c.JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

// Helper function to validate date format
func isValidDateFormat(dateStr string) bool {
	_, err := time.Parse(utils.DateOnlyFormat, dateStr)
	return err == nil
}
