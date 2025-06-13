// File: j-ticketing/internal/core/routes/notification_routes.go
package routes

import (
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"

	"github.com/gofiber/fiber/v2"
)

// SetupNotificationRoutes configures all notification related routes
func SetupNotificationRoutes(app *fiber.App, notificationHandler *handlers.NotificationHandler, jwtService jwt.JWTService) {
	// Notification routes group
	notification := app.Group("/api")

	// Public routes (no authentication required)
	notification.Get("/notifications", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), notificationHandler.GetAllNotifications)

	notification.Put("/notification", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "SYSADMIN", "MEMBER"), notificationHandler.UpdateNotifications)
}
