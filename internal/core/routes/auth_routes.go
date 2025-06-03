// File: j-ticketing/internal/core/routes/auth_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// SetupAuthRoutes configures all authentication related routes
func SetupAuthRoutes(app *fiber.App, authHandler *handlers.AuthHandler, jwtService jwt.JWTService) {
	// Auth routes group
	auth := app.Group("/auth")

	// Public routes (no authentication required)
	auth.Post("/login", authHandler.Login)
	auth.Post("/refresh-token", authHandler.RefreshToken)

	// Customer routes
	auth.Post("/customer/create", authHandler.CreateCustomer)
	auth.Post("/customer/reset-password", authHandler.ResetCustomerPassword)

	// Admin routes
	auth.Post("/admin/reset-password", authHandler.ResetAdminPassword)

	// Protected routes
	auth.Get("/logout", middleware.Protected(jwtService), authHandler.Logout)
	auth.Get("/validate", middleware.Protected(jwtService), authHandler.ValidateToken)
}
