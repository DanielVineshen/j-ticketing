package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	dto "j-ticketing/internal/core/dto/report"
	services "j-ticketing/internal/core/services"
	"j-ticketing/pkg/models"
	"j-ticketing/pkg/validation"
	"strconv"
)

// ReportHandler handles report-related HTTP requests
type ReportHandler struct {
	reportService *services.ReportService
}

// NewReportHandler creates a new report handler
func NewReportHandler(reportService *services.ReportService) *ReportHandler {
	return &ReportHandler{
		reportService: reportService,
	}
}

// CreateReport creates a new report configuration
func (h *ReportHandler) CreateReport(c *fiber.Ctx) error {
	var req dto.CreateReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := validation.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	response, err := h.reportService.CreateReport(&req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to create report: %v", err),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(response))
}

// GetReport retrieves a report by ID
func (h *ReportHandler) GetReport(c *fiber.Ctx) error {
	reportIdStr := c.Params("id")
	reportId, err := strconv.ParseUint(reportIdStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid report ID",
		})
	}

	response, err := h.reportService.GetReport(uint(reportId))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to retrieve report: %v", err),
		})
	}

	if response == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Report not found",
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}

// ListReports retrieves all reports with optional filtering
func (h *ReportHandler) ListReports(c *fiber.Ctx) error {
	reportType := c.Query("type")
	frequency := c.Query("frequency")

	var response interface{}
	var err error

	if reportType != "" {
		response, err = h.reportService.GetReportsByType(reportType)
	} else if frequency != "" {
		response, err = h.reportService.GetReportsByFrequency(frequency)
	} else {
		response, err = h.reportService.GetAllReports()
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to retrieve reports: %v", err),
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}

// UpdateReport updates an existing report
func (h *ReportHandler) UpdateReport(c *fiber.Ctx) error {
	reportIdStr := c.Params("id")
	reportId, err := strconv.ParseUint(reportIdStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid report ID",
		})
	}

	var req dto.UpdateReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := validation.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	response, err := h.reportService.UpdateReport(uint(reportId), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to update report: %v", err),
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}

// DeleteReport soft deletes a report
func (h *ReportHandler) DeleteReport(c *fiber.Ctx) error {
	reportIdStr := c.Params("id")
	reportId, err := strconv.ParseUint(reportIdStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid report ID",
		})
	}

	err = h.reportService.DeleteReport(uint(reportId))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to delete report: %v", err),
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(fiber.Map{
		"message": "Report deleted successfully",
	}))
}

// GenerateReport generates a report based on configuration and date range
func (h *ReportHandler) GenerateReport(c *fiber.Ctx) error {
	var req dto.GenerateReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := validation.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	// Validate date format
	if !isValidDateFormat(req.StartDate) || !isValidDateFormat(req.EndDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date format. Use YYYY-MM-DD",
		})
	}

	attachment, err := h.reportService.GenerateReport(&req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to generate report: %v", err),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(attachment))
}

// DownloadReport downloads a report attachment
func (h *ReportHandler) DownloadReport(c *fiber.Ctx) error {
	attachmentIdStr := c.Params("attachmentId")
	attachmentId, err := strconv.ParseUint(attachmentIdStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid attachment ID",
		})
	}

	attachment, err := h.reportService.GetReportAttachment(uint(attachmentId))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to retrieve attachment: %v", err),
		})
	}

	if attachment == nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Attachment not found",
		})
	}

	// Set headers for file download
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", attachment.AttachmentName))
	c.Set("Content-Type", attachment.ContentType)
	c.Set("Content-Length", strconv.FormatInt(attachment.AttachmentSize, 10))

	return c.SendFile(attachment.AttachmentPath)
}

// GetReportAttachments retrieves all attachments for a report
func (h *ReportHandler) GetReportAttachments(c *fiber.Ctx) error {
	reportIdStr := c.Params("id")
	reportId, err := strconv.ParseUint(reportIdStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid report ID",
		})
	}

	attachments, err := h.reportService.GetReportAttachments(uint(reportId))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to retrieve attachments: %v", err),
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(attachments))
}

// GetValidDataOptions returns valid data options for a report type
func (h *ReportHandler) GetValidDataOptions(c *fiber.Ctx) error {
	reportType := c.Query("type")
	if reportType == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Report type is required",
		})
	}

	var validOptions []string
	switch reportType {
	case "onsite_visitors":
		validOptions = dto.ValidOnsiteVisitorsDataOptions()
	case "sales":
		// Add when implemented
		validOptions = []string{"sales", "orders", "average_order_value", "top_sales_product", "sales_by_attraction", "sales_by_age_group", "sales_by_payment_method", "sales_by_nationality"}
	case "members":
		// Add when implemented
		validOptions = []string{"new_member_rate", "new_members_vs_churn", "total_members_growth", "members_by_age_group", "members_by_nationality"}
	case "online_visitors":
		// Add when implemented
		validOptions = []string{"total_unique_visitors", "peak_time_analysis", "average_session_durations", "new_vs_returning_visitors", "visitors_by_device"}
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid report type",
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(fiber.Map{
		"reportType":   reportType,
		"validOptions": validOptions,
	}))
}

// PreviewReport generates a preview of report data without saving
func (h *ReportHandler) PreviewReport(c *fiber.Ctx) error {
	var req dto.GenerateReportRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := validation.ValidateStruct(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "Validation failed",
			"details": err.Error(),
		})
	}

	// Validate date format
	if !isValidDateFormat(req.StartDate) || !isValidDateFormat(req.EndDate) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid date format. Use YYYY-MM-DD",
		})
	}

	preview, err := h.reportService.PreviewReport(&req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to generate preview: %v", err),
		})
	}

	return c.JSON(models.NewBaseSuccessResponse(preview))
}
