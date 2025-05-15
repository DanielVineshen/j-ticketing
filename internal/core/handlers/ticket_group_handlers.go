package handlers

import (
	"j-ticketing/internal/db/models"
	"net/http"
	"strconv"

	service "j-ticketing/internal/services"

	"github.com/gofiber/fiber/v2"
)

// TicketGroupHandler handles HTTP requests for ticket groups
type TicketGroupHandler struct {
	ticketGroupService *service.TicketGroupService
}

// NewTicketGroupHandler creates a new ticket group handler
func NewTicketGroupHandler(ticketGroupService *service.TicketGroupService) *TicketGroupHandler {
	return &TicketGroupHandler{
		ticketGroupService: ticketGroupService,
	}
}

// GetAllTicketGroups handles GET /ticket-groups
func (h *TicketGroupHandler) GetAllTicketGroups(c *fiber.Ctx) error {
	ticketGroups, err := h.ticketGroupService.GetAllTicketGroups()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to get ticket groups",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Ticket groups retrieved successfully",
		"data":    ticketGroups,
	})
}

// GetTicketGroupByID handles GET /ticket-groups/:id
func (h *TicketGroupHandler) GetTicketGroupByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid ID",
			"error":   err.Error(),
		})
	}

	ticketGroup, err := h.ticketGroupService.GetTicketGroupByID(uint(id))
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"message": "Ticket group not found",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Ticket group retrieved successfully",
		"data":    ticketGroup,
	})
}

// CreateTicketGroup handles POST /ticket-groups
func (h *TicketGroupHandler) CreateTicketGroup(c *fiber.Ctx) error {
	ticketGroup := new(models.TicketGroup)
	if err := c.BodyParser(ticketGroup); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	if err := h.ticketGroupService.CreateTicketGroup(ticketGroup); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to create ticket group",
			"error":   err.Error(),
		})
	}

	return c.Status(http.StatusCreated).JSON(fiber.Map{
		"message": "Ticket group created successfully",
		"data":    ticketGroup,
	})
}

// UpdateTicketGroup handles PUT /ticket-groups/:id
func (h *TicketGroupHandler) UpdateTicketGroup(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid ID",
			"error":   err.Error(),
		})
	}

	existingTicketGroup, err := h.ticketGroupService.GetTicketGroupByID(uint(id))
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"message": "Ticket group not found",
			"error":   err.Error(),
		})
	}

	updatedTicketGroup := new(models.TicketGroup)
	if err := c.BodyParser(updatedTicketGroup); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	updatedTicketGroup.TicketGroupId = existingTicketGroup.TicketGroupId
	if err := h.ticketGroupService.UpdateTicketGroup(updatedTicketGroup); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to update ticket group",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Ticket group updated successfully",
		"data":    updatedTicketGroup,
	})
}

// DeleteTicketGroup handles DELETE /ticket-groups/:id
func (h *TicketGroupHandler) DeleteTicketGroup(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid ID",
			"error":   err.Error(),
		})
	}

	if err := h.ticketGroupService.DeleteTicketGroup(uint(id)); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"message": "Failed to delete ticket group",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Ticket group deleted successfully",
	})
}

// GetTicketGroupWithBanners handles GET /ticket-groups/:id/with-banners
func (h *TicketGroupHandler) GetTicketGroupWithBanners(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid ID",
			"error":   err.Error(),
		})
	}

	ticketGroup, banners, err := h.ticketGroupService.GetTicketGroupWithBanners(uint(id))
	if err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"message": "Failed to get ticket group with banners",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Ticket group with banners retrieved successfully",
		"data": fiber.Map{
			"ticketGroup": ticketGroup,
			"banners":     banners,
		},
	})
}
