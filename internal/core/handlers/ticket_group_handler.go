// File: j-ticketing/internal/core/handlers/ticket_group_handler.go
package handlers

import (
	dto "j-ticketing/internal/core/dto/ticket_group"
	services "j-ticketing/internal/core/services"
	"j-ticketing/pkg/models"
	"net/http"
	"os"
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
			"Missing ticketGroupId parameter", nil,
		))
	}

	ticketGroupId, err := strconv.ParseUint(ticketGroupIdStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid ticketGroupId parameter", nil,
		))
	}

	// Get the ticket profile
	response, err := h.ticketGroupService.GetTicketProfile(uint(ticketGroupId))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Internal Server Error: "+err.Error(), nil,
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
			"Missing ticketGroupId parameter", nil,
		))
	}

	if date == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Missing date parameter", nil,
		))
	}

	// Parse the ticket group ID
	ticketGroupId, err := strconv.ParseUint(ticketGroupIdStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid ticketGroupId parameter", nil,
		))
	}

	// Validate the date format (YYYY-MM-DD)
	_, err = time.Parse("2006-01-02", date)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid date format. Required format: YYYY-MM-DD", nil,
		))
	}

	// Get the ticket variants
	response, err := h.ticketGroupService.GetTicketVariants(uint(ticketGroupId), date)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"Failed to get ticket variants: "+err.Error(), nil,
		))
	}

	// Return the response
	return c.JSON(models.NewBaseSuccessResponse(response))
}

// GetTicketGroupImage serves a ticket group image by its unique extension
func (h *TicketGroupHandler) GetTicketGroupImage(c *fiber.Ctx) error {
	uniqueExtension := c.Params("uniqueExtension")
	if uniqueExtension == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Missing uniqueExtension parameter", nil,
		))
	}

	// Get the content type and file path from the service
	contentType, filePath, err := h.ticketGroupService.GetImageInfo(uniqueExtension)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"File not found.", nil,
		))
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open file",
		})
	}
	defer file.Close()

	// Get file info for Last-Modified header
	fileInfo, err := file.Stat()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get file information",
		})
	}

	// Set response headers for proper caching and content type
	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "inline; filename=\""+uniqueExtension+"\"")
	c.Set("Cache-Control", "public, max-age=86400, must-revalidate") // 24 hours cache
	c.Set("Last-Modified", fileInfo.ModTime().Format(http.TimeFormat))
	c.Set("Expires", time.Now().Add(24*time.Hour).Format(http.TimeFormat))

	// Send the file
	return c.SendFile(filePath)
}
