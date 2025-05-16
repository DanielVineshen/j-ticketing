// FILE: internal/core/handlers/ticket_group_handler.go
package handlers

import (
	"fmt"
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

	var response services.TicketGroupResponse
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

// GetTicketGroupById handles GET requests for a specific ticket group
func (h *TicketGroupHandler) GetTicketGroupById(c *fiber.Ctx) error {
	// Parse the ticket group ID from the URL
	idStr := c.Params("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid ticket group ID",
		})
	}

	// Get the ticket group
	ticketGroup, err := h.ticketGroupService.GetTicketGroupById(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// Create the response with a single ticket group
	response := services.TicketGroupResponse{
		TicketGroups: []services.TicketGroupDTO{*ticketGroup},
	}

	return c.JSON(response)
}

// CreateTicketGroup handles POST requests to create a new ticket group
func (h *TicketGroupHandler) CreateTicketGroup(c *fiber.Ctx) error {
	// TODO: Implement create ticket group functionality
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "Create ticket group functionality not implemented",
	})
}

// UpdateTicketGroup handles PUT requests to update a ticket group
func (h *TicketGroupHandler) UpdateTicketGroup(c *fiber.Ctx) error {
	// TODO: Implement update ticket group functionality
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "Update ticket group functionality not implemented",
	})
}

// DeleteTicketGroup handles DELETE requests to delete a ticket group
func (h *TicketGroupHandler) DeleteTicketGroup(c *fiber.Ctx) error {
	// TODO: Implement delete ticket group functionality
	return c.Status(fiber.StatusNotImplemented).JSON(fiber.Map{
		"message": "Delete ticket group functionality not implemented",
	})
}

// GetTicketProfile handles GET requests for a ticket profile
func (h *TicketGroupHandler) GetTicketProfile(c *fiber.Ctx) error {
	// Log full request information
	fmt.Printf("Full URL: %s\n", c.OriginalURL())
	fmt.Printf("All Query Parameters: %v\n", c.Queries())

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
