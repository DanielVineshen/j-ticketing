// File: j-ticketing/internal/core/routes/banner_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// SetupBannerRoutes configures all banner related routes
func SetupBannerRoutes(app *fiber.App, bannerHandler *handlers.BannerHandler, jwtService jwt.JWTService) {
	// Group banner routes
	banner := app.Group("/api/banners")

	// Public routes (no authentication required)
	banner.Get("/", bannerHandler.GetFilteredBanners)
	banner.Get("/attachment/:uniqueExtension", bannerHandler.GetBannerImage)

	// CRUD routes (add authentication middleware as needed)
	banner.Get("/all", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN"), bannerHandler.GetAllBanners)
	banner.Post("/", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN"), bannerHandler.CreateBanner)
	banner.Put("/", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN"), bannerHandler.UpdateBanner)
	banner.Delete("/", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN"), bannerHandler.DeleteBanner)
	banner.Put("/placements", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN"), bannerHandler.UpdateBannerPlacements)
}
