// File: j-ticketing/internal/core/handlers/banner_handler.go
package handlers

import (
	service "j-ticketing/internal/core/services"
	"j-ticketing/pkg/errors"
	"j-ticketing/pkg/models"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

// BannerImageHandler handles requests for serving banner images
type BannerHandler struct {
	bannerService *service.BannerService
}

// NewBannerImageHandler creates a new banner image handler
func NewBannerImageHandler(bannerImageService *service.BannerService) *BannerHandler {
	return &BannerHandler{
		bannerService: bannerImageService,
	}
}

// GetBannerImage serves a banner image by its unique extension
func (h *BannerHandler) GetBannerImage(c *fiber.Ctx) error {
	uniqueExtension := c.Params("uniqueExtension")
	if uniqueExtension == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			errors.INVALID_INPUT_FORMAT.Code, "Missing uniqueExtension parameter", nil,
		))
	}

	// Get the content type and file path from the service
	contentType, filePath, err := h.bannerService.GetImageInfo(uniqueExtension)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			errors.FILE_NOT_FOUND.Code, "File not found.", nil,
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
