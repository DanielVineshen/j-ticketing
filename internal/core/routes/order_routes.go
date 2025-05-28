// File: j-ticketing/internal/core/routes/order_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// SetupOrderRoutes configures all order related routes
func SetupOrderRoutes(app *fiber.App, orderHandler *handlers.OrderHandler, jwtService jwt.JWTService) {
	// Order routes group
	orderGroup := app.Group("/api")

	// Protected routes (require authentication)
	orderGroup.Get("/orderTicketGroups", middleware.Protected(jwtService), orderHandler.GetOrderTicketGroups)

	orderGroup.Get("/orderTicketGroup", orderHandler.GetOrderTicketGroup)
	orderGroup.Get("/orderNonMemberInquiry", orderHandler.GetOrderNonMemberInquiry)

	// Add create order endpoint
	orderGroup.Post("/orderTicketGroup", orderHandler.CreateOrderTicketGroup)
	orderGroup.Post("/orderTicketGroup/free", orderHandler.CreateFreeOrderTicketGroup)

	// Admin module
	orderGroup.Get("/orderTicketGroups/management", middleware.HasAnyRole("ADMIN", "MEMBER"), orderHandler.GetOrderManagement)
}
