// File: j-ticketing/internal/core/routes/group_gallery_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"

	"github.com/gofiber/fiber/v2"
)

// SetupGroupGalleryRoutes configures all group gallery related routes
func SetupGroupGalleryRoutes(app *fiber.App, groupGalleryHandler *handlers.GroupGalleryHandler) {
	// Group gallery routes group
	groupGallery := app.Group("/api/groupGallery")

	// Public routes (no authentication required)
	groupGallery.Get("/attachment/:uniqueExtension", groupGalleryHandler.GetGroupGalleryImage)
}
