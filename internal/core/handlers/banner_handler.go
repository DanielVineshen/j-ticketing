// File: j-ticketing/internal/core/handlers/banner_handler.go
package handlers

import (
	dto "j-ticketing/internal/core/dto/banner"
	service "j-ticketing/internal/core/services"
	"j-ticketing/pkg/models"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// BannerHandler handles requests for banner operations
type BannerHandler struct {
	bannerService *service.BannerService
}

// NewBannerHandler creates a new banner handler
func NewBannerHandler(bannerService *service.BannerService) *BannerHandler {
	return &BannerHandler{
		bannerService: bannerService,
	}
}

// GetAllBanners retrieves all banners
func (h *BannerHandler) GetAllBanners(c *fiber.Ctx) error {
	banners, err := h.bannerService.GetAllBanners()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"Failed to retrieve banners", nil,
		))
	}

	response := dto.BannerListResponse{
		Banners: banners,
	}

	return c.Status(fiber.StatusOK).JSON(models.NewBaseSuccessResponse(response))
}

// CreateBanner creates a new banner with file upload
func (h *BannerHandler) CreateBanner(c *fiber.Ctx) error {
	// Get the customer ID from the context (set by auth middleware)
	adminUserName, ok := c.Locals("username").(string)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewBaseErrorResponse(
			"User not authenticated", nil,
		))
	}

	// Parse the multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Failed to parse multipart form", nil,
		))
	}

	// Get the file from form
	files := form.File["attachment"]
	if len(files) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Attachment file is required", nil,
		))
	}

	file := files[0]

	// Create request from form data
	request := &dto.CreateNewBannerRequest{
		RedirectURL:     getFormValue(form.Value, "redirectUrl"),
		UploadedBy:      adminUserName,
		ActiveEndDate:   getFormValue(form.Value, "activeEndDate"),
		ActiveStartDate: getFormValue(form.Value, "activeStartDate"),
		IsActive:        getFormValue(form.Value, "isActive") == "true",
	}

	// Parse duration
	if durationStr := getFormValue(form.Value, "duration"); durationStr != "" {
		if duration, err := strconv.Atoi(durationStr); err == nil {
			request.Duration = duration
		}
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	// Create banner through service
	_, err = h.bannerService.CreateBanner(request, file)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

// UpdateBanner updates an existing banner
func (h *BannerHandler) UpdateBanner(c *fiber.Ctx) error {
	// Get the customer ID from the context (set by auth middleware)
	adminUserName, ok := c.Locals("username").(string)

	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewBaseErrorResponse(
			"User not authenticated", nil,
		))
	}

	// Parse the multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Failed to parse multipart form", nil,
		))
	}

	// Create request from form data
	request := &dto.UpdateBannerRequest{
		RedirectURL:     getFormValue(form.Value, "redirectUrl"),
		UploadedBy:      adminUserName,
		ActiveEndDate:   getFormValue(form.Value, "activeEndDate"),
		ActiveStartDate: getFormValue(form.Value, "activeStartDate"),
		IsActive:        getFormValue(form.Value, "isActive") == "true",
	}

	// Parse banner ID and duration
	if bannerIdStr := getFormValue(form.Value, "bannerId"); bannerIdStr != "" {
		if bannerId, err := strconv.ParseUint(bannerIdStr, 10, 32); err == nil {
			request.BannerId = uint(bannerId)
		}
	}

	if durationStr := getFormValue(form.Value, "duration"); durationStr != "" {
		if duration, err := strconv.Atoi(durationStr); err == nil {
			request.Duration = duration
		}
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	// Get file if provided (optional for update)
	var file *multipart.FileHeader
	files := form.File["attachment"]
	if len(files) > 0 {
		file = files[0]
	}

	// Update banner through service
	_, err = h.bannerService.UpdateBanner(request, file)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

// DeleteBanner deletes a banner by ID
func (h *BannerHandler) DeleteBanner(c *fiber.Ctx) error {
	var request dto.DeleteBannerRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request body", nil,
		))
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	// Delete banner through service
	err := h.bannerService.DeleteBanner(request.BannerId)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

// UpdateBannerPlacements updates multiple banner placements
func (h *BannerHandler) UpdateBannerPlacements(c *fiber.Ctx) error {
	var request dto.UpdateBannerPlacementsRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request body", nil,
		))
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	// Update placements through service
	err := h.bannerService.UpdateBannerPlacements(request.Banners)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

// GetBannerImage serves a banner image by its unique extension (existing method)
func (h *BannerHandler) GetBannerImage(c *fiber.Ctx) error {
	uniqueExtension := c.Params("uniqueExtension")
	if uniqueExtension == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Missing uniqueExtension parameter", nil,
		))
	}

	// Get the content type and file path from the service
	contentType, filePath, err := h.bannerService.GetImageInfo(uniqueExtension)
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

// Helper function to get form values
func getFormValue(values map[string][]string, key string) string {
	if vals, exists := values[key]; exists && len(vals) > 0 {
		return vals[0]
	}
	return ""
}
