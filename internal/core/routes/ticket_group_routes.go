// File: j-ticketing/internal/core/routes/ticket_group_routes.go
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
	ticketGroup.Get("/ticketProfile", ticketGroupHandler.GetTicketProfile)
	ticketGroup.Get("/ticketVariants", ticketGroupHandler.GetTicketVariants)
	ticketGroup.Get("/attachment/:uniqueExtension", ticketGroupHandler.GetTicketGroupImage)

	ticketGroup.Post("/", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), ticketGroupHandler.CreateTicketGroup)
	ticketGroup.Put("/placements", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), ticketGroupHandler.UpdateTicketGroupPlacement)
	ticketGroup.Put("/image", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), ticketGroupHandler.UpdateTicketGroupImage)
	ticketGroup.Put("/basicInfo", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), ticketGroupHandler.UpdateTicketGroupBasicInfo)
	ticketGroup.Post("/gallery", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), ticketGroupHandler.UploadTicketGroupGallery)
	ticketGroup.Delete("/gallery", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), ticketGroupHandler.DeleteTicketGroupGallery)
	ticketGroup.Put("/details", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), ticketGroupHandler.UpdateTicketGroupDetails)
	ticketGroup.Put("/variants", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), ticketGroupHandler.UpdateTicketGroupVariants)
	ticketGroup.Put("/organiserInfo", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), ticketGroupHandler.UpdateTicketGroupOrganiserInfo)
}
