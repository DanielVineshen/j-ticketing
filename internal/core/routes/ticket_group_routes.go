// File: j-ticketing/internal/core/routes/ticket_group_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"
	"j-ticketing/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// SetupTicketGroupRoutes configures all ticket group related routes
func SetupTicketGroupRoutes(app *fiber.App, ticketGroupHandler *handlers.TicketGroupHandler, jwtService jwt.JWTService) {
	// Ticket group routes group
	ticketGroup := app.Group("/api/ticketGroups")

	// Public routes (no authentication required)
	ticketGroup.Get("/", ticketGroupHandler.GetTicketGroups)
	ticketGroup.Get("/ticketProfile", ticketGroupHandler.GetTicketProfile)
	ticketGroup.Get("/ticketVariants", ticketGroupHandler.GetTicketVariants)
}
