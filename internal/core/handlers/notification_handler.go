package handlers

import (
	"github.com/gofiber/fiber/v2"
	notificationDto "j-ticketing/internal/core/dto/notification"
	service "j-ticketing/internal/core/services"
	"j-ticketing/pkg/models"
)

// NotificationHandler handles notification-related HTTP requests
type NotificationHandler struct {
	notificationService service.NotificationService
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService service.NotificationService) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
	}
}

func (n *NotificationHandler) GetAllNotifications(c *fiber.Ctx) error {
	notificationDetails, _ := n.notificationService.GetAll()

	notificationResponse := notificationDto.NotificationResponse{
		Notifications: notificationDetails,
	}

	return c.JSON(models.NewBaseSuccessResponse(notificationResponse))
}
