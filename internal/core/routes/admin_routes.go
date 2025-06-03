// File: j-ticketing/internal/core/routes/admin_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// SetupAdminRoutes configures all admin-related routes
func SetupAdminRoutes(api fiber.Router, adminHandler *handlers.AdminHandler, jwtService jwt.JWTService) {
	adminGroup := api.Group("/api/admin")

	// Admin Profile Management Routes (accessible by ADMIN and SYSADMIN)
	// These are for admins managing their own profiles
	adminGroup.Get("/profile", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "MEMBER", "SYSADMIN"), adminHandler.GetAdminProfile)
	adminGroup.Put("/profile", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "MEMBER", "SYSADMIN"), adminHandler.UpdateAdminProfile)
	adminGroup.Put("/password", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "MEMBER", "SYSADMIN"), adminHandler.ChangePassword)

	// Admin Management Routes (accessible only by SYSADMIN)
	// These are for SYSADMIN managing other admin accounts
	adminGroup.Get("/management", middleware.Protected(jwtService), middleware.HasRole("SYSADMIN"), adminHandler.GetAllAdmins)
	adminGroup.Post("/management", middleware.Protected(jwtService), middleware.HasRole("SYSADMIN"), adminHandler.CreateAdmin)
	adminGroup.Put("/management", middleware.Protected(jwtService), middleware.HasRole("SYSADMIN"), adminHandler.UpdateAdminManagement)
	adminGroup.Delete("/management", middleware.Protected(jwtService), middleware.HasRole("SYSADMIN"), adminHandler.DeleteAdmin)
}
