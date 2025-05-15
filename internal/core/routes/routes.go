// FILE: internal/core/routes/routes.go (Updated with protected routes)
package routes

import (
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// SetupRoutes configures all application routes
func SetupRoutes(app *fiber.App, ticketGroupHandler *handlers.TicketGroupHandler, jwtService jwt.JWTService) {
	// API group
	api := app.Group("/api")

	// Public routes (no authentication required)
	publicRoutes := api.Group("/public")
	publicRoutes.Get("/ticket-groups", ticketGroupHandler.GetAllTicketGroups)
	publicRoutes.Get("/ticket-groups/:id", ticketGroupHandler.GetTicketGroupByID)
	publicRoutes.Get("/ticket-groups/:id/with-banners", ticketGroupHandler.GetTicketGroupWithBanners)

	// Admin routes - protected by authentication and role-based authorization
	adminRoutes := api.Group("/admin", middleware.Protected(jwtService))

	// System admin routes
	sysAdminRoutes := adminRoutes.Group("/sysAdmin", middleware.HasRole("SYSADMIN"))
	sysAdminRoutes.Post("/ticket-groups", ticketGroupHandler.CreateTicketGroup)
	sysAdminRoutes.Put("/ticket-groups/:id", ticketGroupHandler.UpdateTicketGroup)
	sysAdminRoutes.Delete("/ticket-groups/:id", ticketGroupHandler.DeleteTicketGroup)

	// Owner routes
	ownerRoutes := adminRoutes.Group("/owner", middleware.HasRole("OWNER"))
	ownerRoutes.Get("/ticket-groups", ticketGroupHandler.GetAllTicketGroups)

	// Staff routes
	staffRoutes := adminRoutes.Group("/staff", middleware.HasRole("STAFF"))
	staffRoutes.Get("/ticket-groups", ticketGroupHandler.GetAllTicketGroups)

	// Personnel routes (accessible by both OWNER and STAFF)
	personnelRoutes := adminRoutes.Group("/personnel", middleware.HasAnyRole("OWNER", "STAFF"))
	personnelRoutes.Get("/ticket-groups/:id", ticketGroupHandler.GetTicketGroupByID)

	// Customer routes - protected by authentication but customer-specific
	customerRoutes := api.Group("/customer", middleware.Protected(jwtService), middleware.HasRole("CUSTOMER"))
	customerRoutes.Get("/ticket-groups", ticketGroupHandler.GetAllTicketGroups)
	customerRoutes.Get("/ticket-groups/:id", ticketGroupHandler.GetTicketGroupByID)
}
