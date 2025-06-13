// File: j-ticketing/internal/core/handlers/notification_handler.go
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

func (n *NotificationHandler) UpdateNotifications(c *fiber.Ctx) error {
	// Parse request
	var req notificationDto.UpdateNotificationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request body", nil,
		))
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	notification, err := n.notificationService.Get(req.NotificationId)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
			"Notification does not exist", nil,
		))
	}

	// Only update if values are provided (not null)
	if req.IsRead != nil {
		notification.IsRead = *req.IsRead
	}

	if req.IsDeleted != nil {
		notification.IsDeleted = *req.IsDeleted
	}

	err = n.notificationService.Update(notification)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"Error updating notification record: "+err.Error(), nil,
		))
	}

	return c.JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}
