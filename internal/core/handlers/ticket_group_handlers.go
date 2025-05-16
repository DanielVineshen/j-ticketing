// File: j-ticketing/internal/core/handlers/ticket_group_handlers.go
package handlers

import (
	"fmt"
	dto "j-ticketing/internal/core/dto/ticket_group"
	services "j-ticketing/internal/core/services"
	"j-ticketing/pkg/models"
	"strconv"

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

	return c.JSON(response)
}

// GetTicketProfile handles GET requests for a ticket profile
func (h *TicketGroupHandler) GetTicketProfile(c *fiber.Ctx) error {
	// Parse the ticket group ID from the request
	ticketGroupIdStr := c.Query("ticketGroupId")
	fmt.Printf("ticketGroupIdStr: %s\n", ticketGroupIdStr) // Corrected printf syntax

	if ticketGroupIdStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			9999, "Bad Request: Missing ticketGroupId parameter", nil,
		))
	}

	ticketGroupId, err := strconv.ParseUint(ticketGroupIdStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"respCode": 400,
			"respDesc": "Bad Request: Invalid ticketGroupId parameter",
		})
	}

	// Get the ticket profile
	response, err := h.ticketGroupService.GetTicketProfile(uint(ticketGroupId))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"respCode": 500,
			"respDesc": "Internal Server Error: " + err.Error(),
		})
	}

	// Return the response
	return c.JSON(response)
}
