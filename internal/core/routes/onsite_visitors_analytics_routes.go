package routes

import (
	"github.com/gofiber/fiber/v2"
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"
)

// SetupOnsiteVisitorsAnalyticsRoutes configures all onsiteVisitorAnalytics related routes
func SetupOnsiteVisitorsAnalyticsRoutes(app *fiber.App, onsiteVisitorAnalyticsHandler *handlers.OnsiteVisitorsAnalyticsHandler, jwtService jwt.JWTService) {
	// OnsiteVisitorsAnalytics routes group
	onsiteVisitorAnalytics := app.Group("/api/analytics")

	onsiteVisitorAnalytics.Get("/totalOnsiteVisitors", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), onsiteVisitorAnalyticsHandler.GetTotalOnsiteVisitors)
	onsiteVisitorAnalytics.Get("/newVsReturningVisitors", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), onsiteVisitorAnalyticsHandler.GetNewVsReturningVisitors)
	onsiteVisitorAnalytics.Get("/averagePeakDayAnalysis", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), onsiteVisitorAnalyticsHandler.GetAveragePeakDayAnalysis)
	onsiteVisitorAnalytics.Get("/visitorsByAttraction", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), onsiteVisitorAnalyticsHandler.GetVisitorsByAttraction)
	onsiteVisitorAnalytics.Get("/visitorsByAgeGroup", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), onsiteVisitorAnalyticsHandler.GetVisitorsByAgeGroup)
	onsiteVisitorAnalytics.Get("/visitorsByNationality", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), onsiteVisitorAnalyticsHandler.GetVisitorsByNationality)
}
