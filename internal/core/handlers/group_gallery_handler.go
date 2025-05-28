// File: j-ticketing/internal/core/handlers/gallery_group_handler.go
package handlers

import (
	service "j-ticketing/internal/core/services"
	"j-ticketing/pkg/models"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GroupGalleryHandler handles requests for serving group gallery images
type GroupGalleryHandler struct {
	groupGalleryService *service.GroupGalleryService
}

// NewGroupGalleryHandler creates a new group gallery image handler
func NewGroupGalleryHandler(groupGalleryService *service.GroupGalleryService) *GroupGalleryHandler {
	return &GroupGalleryHandler{
		groupGalleryService: groupGalleryService,
	}
}

// GetGroupGalleryImage serves a group gallery image by its unique extension
func (h *GroupGalleryHandler) GetGroupGalleryImage(c *fiber.Ctx) error {
	uniqueExtension := c.Params("uniqueExtension")
	if uniqueExtension == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Missing uniqueExtension parameter", nil,
		))
	}

	// Get the content type and file path from the service
	contentType, filePath, err := h.groupGalleryService.GetImageInfo(uniqueExtension)
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
