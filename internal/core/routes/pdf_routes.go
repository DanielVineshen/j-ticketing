// File: j-ticketing/internal/core/routes/pdf_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"

	"github.com/gofiber/fiber/v2"
)

// SetupTicketPDFRoutes configures the route for the simple PDF generator
func SetupTicketPDFRoutes(app *fiber.App, simplePDFHandler *handlers.PDFHandler) {
	// Create a route for generating a PDF ticket
	app.Get("/pdf/receipt", simplePDFHandler.GenerateTicketPDF)
}
