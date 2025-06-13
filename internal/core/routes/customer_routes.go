// File: j-ticketing/internal/core/routes/customer_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// SetupCustomerRoutes configures all customer-related routes
func SetupCustomerRoutes(api fiber.Router, customerHandler *handlers.CustomerHandler, jwtService jwt.JWTService) {
	customerGroup := api.Group("/api/customer")

	//Public routes
	customerGroup.Get("/profile", customerHandler.GetCustomer)

	//Customer routes
	customerGroup.Put("/profile", middleware.Protected(jwtService), middleware.HasRole("CUSTOMER"), customerHandler.UpdateCustomer)
	customerGroup.Put("/password", middleware.Protected(jwtService), middleware.HasRole("CUSTOMER"), customerHandler.ChangePassword)

	// Admin
	customerGroup.Get("/management", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), customerHandler.GetCustomerManagement)
}
