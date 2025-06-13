// File: j-ticketing/internal/core/routes/member_analytics_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

func SetupMemberAnalyticsRoutes(app *fiber.App, memberAnalyticsHandler *handlers.MemberAnalyticsHandler, jwtService jwt.JWTService) {
	analytics := app.Group("/api/analytics")

	// Member analytics routes
	analytics.Get("/totalMembers",
		middleware.Protected(jwtService),
		middleware.HasAnyRole("ADMIN", "MEMBER", "SYSADMIN"),
		memberAnalyticsHandler.GetTotalMembers)

	analytics.Get("/membersNetGrowth",
		middleware.Protected(jwtService),
		middleware.HasAnyRole("ADMIN", "MEMBER", "SYSADMIN"),
		memberAnalyticsHandler.GetMembersNetGrowth)

	analytics.Get("/membersByAgeGroup",
		middleware.Protected(jwtService),
		middleware.HasAnyRole("ADMIN", "MEMBER", "SYSADMIN"),
		memberAnalyticsHandler.GetMembersByAgeGroup)

	analytics.Get("/membersByNationality",
		middleware.Protected(jwtService),
		middleware.HasAnyRole("ADMIN", "MEMBER", "SYSADMIN"),
		memberAnalyticsHandler.GetMembersByNationality)
}
