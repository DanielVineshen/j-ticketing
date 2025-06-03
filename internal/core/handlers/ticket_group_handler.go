// File: j-ticketing/internal/core/handlers/ticket_group_handler.go
package handlers

import (
	"encoding/json"
	"fmt"
	dto "j-ticketing/internal/core/dto/ticket_group"
	services "j-ticketing/internal/core/services"
	"j-ticketing/pkg/models"
	"j-ticketing/pkg/validation"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
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

func (h *TicketGroupHandler) CreateTicketGroup(c *fiber.Ctx) error {
	// Parse and validate form data
	req, err := h.parseCreateTicketGroupRequest(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	// Validate the request struct
	if err := validation.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Validation failed: "+err.Error(), nil,
		))
	}

	// Additional custom validations
	if err := h.validateCustomFields(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	_, err = h.ticketGroupService.CreateTicketGroup(req, req.Attachment, req.GroupGalleries)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"Failed to create ticket group: "+err.Error(), nil,
		))
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

// parseCreateTicketGroupRequest parses the multipart form data into a structured request
func (h *TicketGroupHandler) parseCreateTicketGroupRequest(c *fiber.Ctx) (*dto.CreateTicketGroupRequest, error) {
	req := &dto.CreateTicketGroupRequest{}

	// Parse basic fields
	var err error

	// Order ticket limit
	orderTicketLimitStr := c.FormValue("orderTicketLimit")
	if orderTicketLimitStr == "" {
		return nil, fmt.Errorf("orderTicketLimit is required")
	}
	req.OrderTicketLimit, err = strconv.Atoi(orderTicketLimitStr)
	if err != nil {
		return nil, fmt.Errorf("orderTicketLimit must be a valid integer")
	}

	// Scan setting
	req.ScanSetting = c.FormValue("scanSetting")

	// Group names
	req.GroupNameBm = c.FormValue("groupNameBm")
	req.GroupNameEn = c.FormValue("groupNameEn")
	req.GroupNameCn = c.FormValue("groupNameCn")

	// Group descriptions
	req.GroupDescBm = c.FormValue("groupDescBm")
	req.GroupDescEn = c.FormValue("groupDescEn")
	req.GroupDescCn = c.FormValue("groupDescCn")

	// Optional redirection fields
	req.GroupRedirectionSpanBm = c.FormValue("groupRedirectionSpanBm")
	req.GroupRedirectionSpanEn = c.FormValue("groupRedirectionSpanEn")
	req.GroupRedirectionSpanCn = c.FormValue("groupRedirectionSpanCn")
	req.GroupRedirectionUrl = c.FormValue("groupRedirectionUrl")

	// Group slots
	req.GroupSlot1Bm = c.FormValue("groupSlot1Bm")
	req.GroupSlot1En = c.FormValue("groupSlot1En")
	req.GroupSlot1Cn = c.FormValue("groupSlot1Cn")
	req.GroupSlot2Bm = c.FormValue("groupSlot2Bm")
	req.GroupSlot2En = c.FormValue("groupSlot2En")
	req.GroupSlot2Cn = c.FormValue("groupSlot2Cn")
	req.GroupSlot3Bm = c.FormValue("groupSlot3Bm")
	req.GroupSlot3En = c.FormValue("groupSlot3En")
	req.GroupSlot3Cn = c.FormValue("groupSlot3Cn")
	req.GroupSlot4Bm = c.FormValue("groupSlot4Bm")
	req.GroupSlot4En = c.FormValue("groupSlot4En")
	req.GroupSlot4Cn = c.FormValue("groupSlot4Cn")

	// Price prefixes and suffixes
	req.PricePrefixBm = c.FormValue("pricePrefixBm")
	req.PricePrefixEn = c.FormValue("pricePrefixEn")
	req.PricePrefixCn = c.FormValue("pricePrefixCn")
	req.PriceSuffixBm = c.FormValue("priceSuffixBm")
	req.PriceSuffixEn = c.FormValue("priceSuffixEn")
	req.PriceSuffixCn = c.FormValue("priceSuffixCn")

	// Date fields
	req.ActiveStartDate = c.FormValue("activeStartDate")
	req.ActiveEndDate = c.FormValue("activeEndDate")

	// Boolean field
	isActiveStr := c.FormValue("isActive")
	if isActiveStr != "" {
		req.IsActive, err = strconv.ParseBool(isActiveStr)
		if err != nil {
			return nil, fmt.Errorf("isActive must be a valid boolean")
		}
	}

	// Location information
	req.LocationAddress = c.FormValue("locationAddress")
	req.LocationMapUrl = c.FormValue("locationMapUrl")

	// Organiser information
	req.OrganiserNameBm = c.FormValue("organiserNameBm")
	req.OrganiserNameEn = c.FormValue("organiserNameEn")
	req.OrganiserNameCn = c.FormValue("organiserNameCn")
	req.OrganiserAddress = c.FormValue("organiserAddress")
	req.OrganiserDescHtmlBm = c.FormValue("organiserDescHtmlBm")
	req.OrganiserDescHtmlEn = c.FormValue("organiserDescHtmlEn")
	req.OrganiserDescHtmlCn = c.FormValue("organiserDescHtmlCn")
	req.OrganiserContact = c.FormValue("organiserContact")
	req.OrganiserEmail = c.FormValue("organiserEmail")
	req.OrganiserWebsite = c.FormValue("organiserWebsite")
	req.OrganiserFacilitiesBm = c.FormValue("organiserFacilitiesBm")
	req.OrganiserFacilitiesEn = c.FormValue("organiserFacilitiesEn")
	req.OrganiserFacilitiesCn = c.FormValue("organiserFacilitiesCn")

	// Parse JSON arrays - helper function to handle both quoted and unquoted JSON
	parseJSONField := func(fieldValue, fieldName string) ([]byte, error) {
		if fieldValue == "" {
			return nil, fmt.Errorf("%s is required", fieldName)
		}

		// Remove any surrounding quotes if present
		trimmed := strings.TrimSpace(fieldValue)
		if strings.HasPrefix(trimmed, "\"") && strings.HasSuffix(trimmed, "\"") {
			// It's a quoted JSON string, unquote it
			var unquoted string
			if err := json.Unmarshal([]byte(trimmed), &unquoted); err != nil {
				return nil, fmt.Errorf("invalid quoted JSON format for %s: %v", fieldName, err)
			}
			return []byte(unquoted), nil
		}
		// It's direct JSON
		return []byte(trimmed), nil
	}

	// Parse ticketDetails
	ticketDetailsStr := c.FormValue("ticketDetails")
	ticketDetailsBytes, err := parseJSONField(ticketDetailsStr, "ticketDetails")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(ticketDetailsBytes, &req.TicketDetails); err != nil {
		return nil, fmt.Errorf("invalid JSON format for ticketDetails: %v", err)
	}

	// Parse ticketVariants
	ticketVariantsStr := c.FormValue("ticketVariants")
	ticketVariantsBytes, err := parseJSONField(ticketVariantsStr, "ticketVariants")
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(ticketVariantsBytes, &req.TicketVariants); err != nil {
		return nil, fmt.Errorf("invalid JSON format for ticketVariants: %v", err)
	}

	// Parse ticketTags (optional)
	ticketTagsStr := c.FormValue("ticketTags")
	if ticketTagsStr != "" {
		ticketTagsBytes, err := parseJSONField(ticketTagsStr, "ticketTags")
		if err != nil {
			// ticketTags is optional, so we don't return error if empty
			req.TicketTags = []dto.TicketTagsRequest{}
		} else {
			if err := json.Unmarshal(ticketTagsBytes, &req.TicketTags); err != nil {
				return nil, fmt.Errorf("invalid JSON format for ticketTags: %v", err)
			}
		}
	} else {
		req.TicketTags = []dto.TicketTagsRequest{}
	}

	// Handle file uploads
	attachment, err := c.FormFile("attachment")
	if err != nil {
		return nil, fmt.Errorf("attachment file is required")
	}
	req.Attachment = attachment

	// Handle multiple gallery files (optional)
	form, err := c.MultipartForm()
	if err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %v", err)
	}

	if files := form.File["groupGalleries"]; files != nil {
		req.GroupGalleries = files
	}

	return req, nil
}

// validateCustomFields performs additional custom validations
func (h *TicketGroupHandler) validateCustomFields(req *dto.CreateTicketGroupRequest) error {
	// Validate date formats
	if err := h.validateDateFormat(req.ActiveStartDate, "activeStartDate"); err != nil {
		return err
	}

	if req.ActiveEndDate != "" {
		if err := h.validateDateFormat(req.ActiveEndDate, "activeEndDate"); err != nil {
			return err
		}

		// Validate that end date is after start date
		startDate, _ := time.Parse("2006-01-02", req.ActiveStartDate)
		endDate, _ := time.Parse("2006-01-02", req.ActiveEndDate)
		if endDate.Before(startDate) {
			return fmt.Errorf("activeEndDate must be after activeStartDate")
		}
	}

	// Validate file types for attachment
	if err := h.validateFileType(req.Attachment, []string{".jpg", ".jpeg", ".png", ".pdf"}); err != nil {
		return fmt.Errorf("attachment: %v", err)
	}

	// Validate gallery files if present
	for i, gallery := range req.GroupGalleries {
		if err := h.validateFileType(gallery, []string{".jpg", ".jpeg", ".png", ".gif"}); err != nil {
			return fmt.Errorf("groupGalleries[%d]: %v", i, err)
		}
	}

	// Validate file sizes (50MB limit)
	if req.Attachment.Size > 50*1024*1024 {
		return fmt.Errorf("attachment file size must not exceed 5MB")
	}

	for i, gallery := range req.GroupGalleries {
		if gallery.Size > 50*1024*1024 {
			return fmt.Errorf("groupGalleries[%d] file size must not exceed 5MB", i)
		}
	}

	return nil
}

// validateDateFormat validates YYYY-MM-DD date format
func (h *TicketGroupHandler) validateDateFormat(dateStr, fieldName string) error {
	if dateStr == "" {
		return fmt.Errorf("%s is required", fieldName)
	}

	_, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return fmt.Errorf("%s must be in YYYY-MM-DD format", fieldName)
	}

	return nil
}

// validateFileType validates file extensions
func (h *TicketGroupHandler) validateFileType(file *multipart.FileHeader, allowedTypes []string) error {
	if file == nil {
		return fmt.Errorf("file is required")
	}

	filename := strings.ToLower(file.Filename)
	for _, allowedType := range allowedTypes {
		if strings.HasSuffix(filename, allowedType) {
			return nil
		}
	}

	return fmt.Errorf("file type not allowed. Allowed types: %v", allowedTypes)
}

// UpdateTicketGroupPlacement handles PUT requests to update ticket group placements
func (h *TicketGroupHandler) UpdateTicketGroupPlacement(c *fiber.Ctx) error {
	// Parse request body
	var req dto.UpdatePlacementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request format", nil,
		))
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Validation failed: "+err.Error(), nil,
		))
	}

	// Call service to update placements
	err := h.ticketGroupService.UpdatePlacements(req.TicketGroups)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"Failed to update placements: "+err.Error(), nil,
		))
	}

	return c.JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

func (h *TicketGroupHandler) UpdateTicketGroupImage(c *fiber.Ctx) error {
	// Parse request body
	var req dto.UpdateTicketGroupImageRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request format", nil,
		))
	}

	// Validate the request struct
	if err := validation.ValidateStruct(req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Validation failed: "+err.Error(), nil,
		))
	}

	// Handle file uploads
	attachment, err := c.FormFile("attachment")
	if err != nil {
		return fmt.Errorf("attachment file is required")
	}
	req.Attachment = attachment

	// Validate file types for attachment
	if err := h.validateFileType(req.Attachment, []string{".jpg", ".jpeg", ".png", ".pdf"}); err != nil {
		return fmt.Errorf("attachment: %v", err)
	}

	// Validate file sizes (50MB limit)
	if req.Attachment.Size > 50*1024*1024 {
		return fmt.Errorf("attachment file size must not exceed 5MB")
	}

	// Call service to update image
	err = h.ticketGroupService.UpdateTicketGroupImage(req.TicketGroupId, req.Attachment)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"Failed to update image: "+err.Error(), nil,
		))
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}
