// FILE: internal/core/routes/customer_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// SetupCustomerRoutes configures all customer-related routes
func SetupCustomerRoutes(api fiber.Router, customerHandler *handlers.CustomerHandler, jwtService jwt.JWTService) {
	customerGroup := api.Group("/api")

	publicCustomerRoutes := api.Group("/api/customer")
	publicCustomerRoutes.Get("/profile", customerHandler.GetCustomer)

	// Customer routes - protected by authentication
	customerRoutes := api.Group("/api/customer", middleware.Protected(jwtService), middleware.HasRole("CUSTOMER"))

	// Profile routes
	customerRoutes.Put("/profile", customerHandler.UpdateCustomer)
	customerRoutes.Put("/password", customerHandler.ChangePassword)

	// Admin
	customerGroup.Get("/customer/management", middleware.Protected(jwtService), middleware.HasRole("CUSTOMER"), customerHandler.GetCustomerManagement)
}
