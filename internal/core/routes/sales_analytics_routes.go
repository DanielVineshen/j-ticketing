// File: j-ticketing/internal/core/routes/sales_analytics_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

func SetupSalesAnalyticsRoutes(app *fiber.App, salesAnalyticsHandler *handlers.SalesAnalyticsHandler, jwtService jwt.JWTService) {
	analytics := app.Group("/api/analytics")

	// Original 4 analytics routes
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

	analytics.Get("/salesByTicketGroup",
		middleware.Protected(jwtService),
		middleware.HasAnyRole("ADMIN", "MEMBER", "SYSADMIN"),
		salesAnalyticsHandler.GetSalesByTicketGroup)

	analytics.Get("/salesByAgeGroup",
		middleware.Protected(jwtService),
		middleware.HasAnyRole("ADMIN", "MEMBER", "SYSADMIN"),
		salesAnalyticsHandler.GetSalesByAgeGroup)

	analytics.Get("/salesByPaymentMethod",
		middleware.Protected(jwtService),
		middleware.HasAnyRole("ADMIN", "MEMBER", "SYSADMIN"),
		salesAnalyticsHandler.GetSalesByPaymentMethod)

	analytics.Get("/salesByNationality",
		middleware.Protected(jwtService),
		middleware.HasAnyRole("ADMIN", "MEMBER", "SYSADMIN"),
		salesAnalyticsHandler.GetSalesByNationality)

	analytics.Get("/salesByTicketVariant",
		middleware.Protected(jwtService),
		middleware.HasAnyRole("ADMIN", "MEMBER", "SYSADMIN"),
		salesAnalyticsHandler.GetSalesByTicketVariant)
}
