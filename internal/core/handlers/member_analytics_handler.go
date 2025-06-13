// File: j-ticketing/internal/core/handlers/member_analytics_handler.go
package handlers

import (
	dto "j-ticketing/internal/core/dto/member_analytics"
	service "j-ticketing/internal/core/services"
	"j-ticketing/pkg/errors"
	"j-ticketing/pkg/models"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type MemberAnalyticsHandler struct {
	memberAnalyticsService *service.MemberAnalyticsService
}

func NewMemberAnalyticsHandler(memberAnalyticsService *service.MemberAnalyticsService) *MemberAnalyticsHandler {
	return &MemberAnalyticsHandler{
		memberAnalyticsService: memberAnalyticsService,
	}
}

// GetTotalMembers handles GET /api/analytics/totalMembers
func (h *MemberAnalyticsHandler) GetTotalMembers(c *fiber.Ctx) error {
	// Parse and validate request
	var req dto.MemberAnalyticsRequest
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

	// Get total members from service
	response, err := h.memberAnalyticsService.GetTotalMembers(req.StartDate, req.EndDate)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve total members",
			"message": err.Error(),
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}

// GetMembersNetGrowth handles GET /api/analytics/membersNetGrowth
func (h *MemberAnalyticsHandler) GetMembersNetGrowth(c *fiber.Ctx) error {
	// Parse and validate request
	var req dto.MemberAnalyticsRequest
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

	// Get members net growth from service
	response, err := h.memberAnalyticsService.GetMembersNetGrowth(req.StartDate, req.EndDate)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve members net growth",
			"message": err.Error(),
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}

// GetMembersByAgeGroup handles GET /api/analytics/membersByAgeGroup
func (h *MemberAnalyticsHandler) GetMembersByAgeGroup(c *fiber.Ctx) error {
	// Parse and validate request
	var req dto.MemberAnalyticsRequest
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

	// Get members by age group from service
	response, err := h.memberAnalyticsService.GetMembersByAgeGroup(req.StartDate, req.EndDate)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve members by age group",
			"message": err.Error(),
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}

// GetMembersByNationality handles GET /api/analytics/membersByNationality
func (h *MemberAnalyticsHandler) GetMembersByNationality(c *fiber.Ctx) error {
	// Parse and validate request
	var req dto.MemberAnalyticsRequest
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

	// Get members by nationality from service
	response, err := h.memberAnalyticsService.GetMembersByNationality(req.StartDate, req.EndDate)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to retrieve members by nationality",
			"message": err.Error(),
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}
