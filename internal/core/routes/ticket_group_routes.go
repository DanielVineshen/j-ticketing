// FILE: internal/ticket/routes/ticket_group_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// SetupTicketGroupRoutes configures all ticket group related routes
func SetupTicketGroupRoutes(app *fiber.App, ticketGroupHandler *handlers.TicketGroupHandler, jwtService jwt.JWTService) {
	// Ticket group routes group
	ticketGroup := app.Group("/api/ticketGroups")

	// Public routes (no authentication required)
	ticketGroup.Get("/", ticketGroupHandler.GetTicketGroups)
	ticketGroup.Get("/:id", ticketGroupHandler.GetTicketGroupById)

	// Public ticket profile endpoint
	ticketGroup.Get("/public/ticketProfile", ticketGroupHandler.GetTicketProfile)

	// Protected routes (admin operations)
	ticketGroup.Post("/", middleware.Protected(jwtService), ticketGroupHandler.CreateTicketGroup)
	ticketGroup.Put("/:id", middleware.Protected(jwtService), ticketGroupHandler.UpdateTicketGroup)
	ticketGroup.Delete("/:id", middleware.Protected(jwtService), ticketGroupHandler.DeleteTicketGroup)
}
