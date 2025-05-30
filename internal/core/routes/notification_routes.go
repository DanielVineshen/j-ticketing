package routes

import (
	"github.com/gofiber/fiber/v2"
	"j-ticketing/internal/core/handlers"
	"j-ticketing/internal/core/middleware"
	"j-ticketing/pkg/jwt"
)

// SetupNotificationRoutes configures all notification related routes
func SetupNotificationRoutes(app *fiber.App, notificationHandler *handlers.NotificationHandler, jwtService jwt.JWTService) {
	// Notification routes group
	notification := app.Group("/api")

	// Public routes (no authentication required)
	notification.Get("/notifications", middleware.Protected(jwtService), middleware.HasAnyRole("ADMIN", "MEMBER"), notificationHandler.GetAllNotifications)
}
