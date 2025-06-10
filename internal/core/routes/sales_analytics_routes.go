// File: j-ticketing/internal/core/routes/sales_analytics_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

func SetupSalesAnalyticsRoutes(app *fiber.App, salesAnalyticsHandler *handlers.SalesAnalyticsHandler, jwtService jwt.JWTService) {
	api := app.Group("/api")

	// Sales Analytics endpoints
	analytics := api.Group("/analytics")

	// Protect all analytics routes with authentication and role-based access
	analytics.Get("/totalRevenue",
		middleware.Protected(jwtService),
		middleware.HasAnyRole("ADMIN", "MEMBER", "SYSADMIN"),
		salesAnalyticsHandler.GetTotalRevenue)

	analytics.Get("/totalOrders",
		middleware.Protected(jwtService),
		middleware.HasAnyRole("ADMIN", "MEMBER", "SYSADMIN"),
		salesAnalyticsHandler.GetTotalOrders)

	analytics.Get("/avgOrderValue",
		middleware.Protected(jwtService),
		middleware.HasAnyRole("ADMIN", "MEMBER", "SYSADMIN"),
		salesAnalyticsHandler.GetAvgOrderValue)

	analytics.Get("/topSalesProduct",
		middleware.Protected(jwtService),
		middleware.HasAnyRole("ADMIN", "MEMBER", "SYSADMIN"),
		salesAnalyticsHandler.GetTopSalesProduct)
}
