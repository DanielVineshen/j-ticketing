// File: j-ticketing/internal/core/routes/general_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// SetupGeneralRoutes configures all general settings related routes
func SetupGeneralRoutes(app *fiber.App, generalHandler *handlers.GeneralHandler, jwtService jwt.JWTService) {
	// Group general settings routes
	settings := app.Group("/api/settings")

	// Public route for serving attachments
	settings.Get("/general/attachment/:uniqueExtension", generalHandler.GetGeneralAttachment)

	// Protected routes (only SYSADMIN can access)
	settings.Get("/general", middleware.Protected(jwtService), middleware.HasAnyRole("SYSADMIN"), generalHandler.GetGeneralSettings)
	settings.Put("/general", middleware.Protected(jwtService), middleware.HasAnyRole("SYSADMIN"), generalHandler.UpdateGeneralSettings)

	// Content-specific get and update routes
	settings.Get("/privacyPolicy", generalHandler.GetPrivacyPolicy)
	settings.Put("/privacyPolicy", middleware.Protected(jwtService), middleware.HasAnyRole("SYSADMIN"), generalHandler.UpdatePrivacyPolicy)
	settings.Get("/termsOfPurchase", generalHandler.GetTermsOfPurchase)
	settings.Put("/termsOfPurchase", middleware.Protected(jwtService), middleware.HasAnyRole("SYSADMIN"), generalHandler.UpdateTermsOfPurchase)
	settings.Get("/termsOfService", generalHandler.GetTermsOfService)
	settings.Put("/termsOfService", middleware.Protected(jwtService), middleware.HasAnyRole("SYSADMIN"), generalHandler.UpdateTermsOfService)
	settings.Get("/faq", generalHandler.GetFaq)
	settings.Put("/faq", middleware.Protected(jwtService), middleware.HasAnyRole("SYSADMIN"), generalHandler.UpdateFaq)
	settings.Get("/contactUs", generalHandler.GetContactUs)
	settings.Put("/contactUs", middleware.Protected(jwtService), middleware.HasAnyRole("SYSADMIN"), generalHandler.UpdateContactUs)
	settings.Get("/refundPolicy", generalHandler.GetRefundPolicy)
	settings.Put("/refundPolicy", middleware.Protected(jwtService), middleware.HasAnyRole("SYSADMIN"), generalHandler.UpdateRefundPolicy)

	// Integrations
	settings.Put("/integration", middleware.Protected(jwtService), middleware.HasAnyRole("SYSADMIN"), generalHandler.UpdateIntegration)
}
