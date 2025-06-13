// File: j-ticketing/internal/core/routes/report_routes.go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"
)

// SetupReportRoutes configures all report related routes
func SetupReportRoutes(app *fiber.App, reportHandler *handlers.ReportHandler, jwtService jwt.JWTService) {
	// Report routes group
	reports := app.Group("/api/reports")

	// Apply authentication and authorization middleware
	reports.Use(middleware.Protected(jwtService))
	reports.Use(middleware.HasAnyRole("ADMIN", "SYSADMIN"))

	// IMPORTANT: Specific routes MUST come before parameterized routes
	// Utility endpoints (specific routes first)
	reports.Get("/data-options", reportHandler.GetValidDataOptions) // Get valid data options for report type

	// Report generation and management (specific routes)
	reports.Post("/generate", reportHandler.GenerateReport) // Generate report file
	reports.Post("/preview", reportHandler.PreviewReport)   // Preview report data without saving

	// Attachment management (specific routes)
	reports.Get("/attachments/:attachmentId/download", reportHandler.DownloadReport) // Download specific attachment

	// CRUD operations (general routes)
	reports.Post("/", reportHandler.CreateReport) // Create new report configuration
	reports.Get("/", reportHandler.ListReports)   // List all reports (with optional filtering)

	// Parameterized routes (MUST come after specific routes)
	reports.Get("/:id", reportHandler.GetReport)                        // Get specific report by ID
	reports.Put("/:id", reportHandler.UpdateReport)                     // Update report configuration
	reports.Delete("/:id", reportHandler.DeleteReport)                  // Soft delete report
	reports.Get("/:id/attachments", reportHandler.GetReportAttachments) // Get all attachments for a report

	// Scheduled reports (for future implementation)
	// reports.Post("/:id/schedule", reportHandler.ScheduleReport)    // Schedule report generation
	// reports.Delete("/:id/schedule", reportHandler.UnscheduleReport) // Remove scheduled report
}
