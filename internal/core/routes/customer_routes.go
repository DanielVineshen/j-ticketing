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
	// Customer routes - protected by authentication
	customerRoutes := api.Group("/api/customer")

	// Profile routes
	customerRoutes.Get("/profile", customerHandler.GetCustomer)
	customerRoutes.Put("/profile", customerHandler.UpdateCustomer, middleware.Protected(jwtService), middleware.HasRole("CUSTOMER"))
	customerRoutes.Put("/password", customerHandler.ChangePassword, middleware.Protected(jwtService), middleware.HasRole("CUSTOMER"))
}
