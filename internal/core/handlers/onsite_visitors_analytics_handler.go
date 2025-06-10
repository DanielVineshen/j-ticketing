package handlers

import (
	"github.com/gofiber/fiber/v2"
	service "j-ticketing/internal/core/services"
	"j-ticketing/pkg/models"
)

// OnsiteVisitorsAnalyticsHandler handles onsiteVisitorsAnalytics-related HTTP requests
type OnsiteVisitorsAnalyticsHandler struct {
	customerService                *service.CustomerService
	onsiteVisitorsAnalyticsService *service.OnsiteVisitorsAnalyticsService
}

// NewOnsiteVisitorsAnalyticsHandler creates a new onsiteVisitorsAnalytics handler
func NewOnsiteVisitorsAnalyticsHandler(customerService *service.CustomerService, onsiteVisitorsAnalyticsService *service.OnsiteVisitorsAnalyticsService) *OnsiteVisitorsAnalyticsHandler {
	return &OnsiteVisitorsAnalyticsHandler{
		customerService:                customerService,
		onsiteVisitorsAnalyticsService: onsiteVisitorsAnalyticsService,
	}
}

func (o *OnsiteVisitorsAnalyticsHandler) GetTotalOnsiteVisitors(c *fiber.Ctx) error {

	return c.JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

func (o *OnsiteVisitorsAnalyticsHandler) GetNewVsReturningVisitors(c *fiber.Ctx) error {
	return c.JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
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
