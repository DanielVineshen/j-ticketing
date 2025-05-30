// FILE: internal/auth/routes/auth_routes.go
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

	// Protected routes
	auth.Post("/logout", middleware.Protected(jwtService), authHandler.Logout)
	auth.Get("/validate", middleware.Protected(jwtService), authHandler.ValidateToken)
}
