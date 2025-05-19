// File: j-ticketing/internal/core/routes/banner_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"

	"github.com/gofiber/fiber/v2"
)

// SetupBannerRoutes configures all banner related routes
func SetupBannerRoutes(app *fiber.App, bannerHandler *handlers.BannerHandler) {
	// Group gallery routes group
	banner := app.Group("/api/banners")

	// Public routes (no authentication required)
	banner.Get("/attachment/:uniqueExtension", bannerHandler.GetBannerImage)
}
