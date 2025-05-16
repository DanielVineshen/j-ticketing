// File: j-ticketing/internal/core/handlers/ticket_group_handlers.go
package handlers

import (
	dto "j-ticketing/internal/core/dto/ticket_group"
	services "j-ticketing/internal/core/services"
	"j-ticketing/pkg/errors"
	"j-ticketing/pkg/models"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// TicketGroupHandler handles HTTP requests for ticket groups
type TicketGroupHandler struct {
	ticketGroupService *services.TicketGroupService
}

// NewTicketGroupHandler creates a new instance of TicketGroupHandler
func NewTicketGroupHandler(ticketGroupService *services.TicketGroupService) *TicketGroupHandler {
	return &TicketGroupHandler{
		ticketGroupService: ticketGroupService,
	}
}

// GetTicketGroups handles GET requests for ticket groups
func (h *TicketGroupHandler) GetTicketGroups(c *fiber.Ctx) error {
	// Check if only active ticket groups should be returned
	activeOnly := c.Query("active") == "true"

	var response dto.TicketGroupResponse
	var err error

	if activeOnly {
		response, err = h.ticketGroupService.GetActiveTicketGroups()
	} else {
		response, err = h.ticketGroupService.GetAllTicketGroups()
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}

// GetTicketProfile handles GET requests for a ticket profile
func (h *TicketGroupHandler) GetTicketProfile(c *fiber.Ctx) error {
	// Parse the ticket group ID from the request
	ticketGroupIdStr := c.Query("ticketGroupId")

	if ticketGroupIdStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			errors.INVALID_INPUT_FORMAT.Code, "Missing ticketGroupId parameter", nil,
		))
	}

	ticketGroupId, err := strconv.ParseUint(ticketGroupIdStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			errors.INVALID_INPUT_VALUES.Code, "Invalid ticketGroupId parameter", nil,
		))
	}

	// Get the ticket profile
	response, err := h.ticketGroupService.GetTicketProfile(uint(ticketGroupId))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			errors.PROCESSING_ERROR.Code, "Internal Server Error: "+err.Error(), nil,
		))
	}

	// Return the response
	return c.JSON(models.NewBaseSuccessResponse(response))
}

// GetTicketVariants handles GET requests for ticket variants
func (h *TicketGroupHandler) GetTicketVariants(c *fiber.Ctx) error {
	// Parse the query parameters
	ticketGroupIdStr := c.Query("ticketGroupId")
	date := c.Query("date")

	// Validate the parameters
	if ticketGroupIdStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			errors.INVALID_INPUT_FORMAT.Code, "Missing ticketGroupId parameter", nil,
		))
	}

	if date == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			errors.INVALID_INPUT_FORMAT.Code, "Missing date parameter", nil,
		))
	}

	// Parse the ticket group ID
	ticketGroupId, err := strconv.ParseUint(ticketGroupIdStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			errors.INVALID_INPUT_VALUES.Code, "Invalid ticketGroupId parameter", nil,
		))
	}

	// Validate the date format (YYYY-MM-DD)
	_, err = time.Parse("2006-01-02", date)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			errors.INVALID_INPUT_VALUES.Code, "Invalid date format. Required format: YYYY-MM-DD", nil,
		))
	}

	// Get the ticket variants
	response, err := h.ticketGroupService.GetTicketVariants(uint(ticketGroupId), date)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			errors.PROCESSING_ERROR.Code, "Failed to get ticket variants: "+err.Error(), nil,
		))
	}

	// Return the response
	return c.JSON(models.NewBaseSuccessResponse(response))
}
