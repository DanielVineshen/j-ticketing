// File: j-ticketing/internal/core/routes/tag_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// SetupTagRoutes configures all tag related routes
func SetupTagRoutes(app *fiber.App, tagHandler *handlers.TagHandler, jwtService jwt.JWTService) {
	// Group tag routes
	tags := app.Group("/api/tags")

	// All routes require authentication and ADMIN role
	tags.Get("/", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN"), tagHandler.GetAllTags)
	tags.Post("/", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN"), tagHandler.CreateTag)
	tags.Put("/", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN"), tagHandler.UpdateTag)
	tags.Delete("/", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN"), tagHandler.DeleteTag)
}
