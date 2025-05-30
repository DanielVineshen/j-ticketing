// File: j-ticketing/internal/core/routes/dashboard_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

func SetupDashboardRoutes(app *fiber.App, dashboardHandler *handlers.DashboardHandler, jwtService jwt.JWTService) {
	api := app.Group("/api")

	// Dashboard endpoint
	api.Get("/dashboard", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN"), dashboardHandler.GetDashboardAnalysis)
}
