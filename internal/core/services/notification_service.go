// File: j-ticketing/internal/core/services/notification_service.go
package service

import (
	"database/sql"
	"fmt"
	notificationDto "j-ticketing/internal/core/dto/notification"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"time"
)

// NotificationService handles operations related to serving notifications
type NotificationService struct {
	notificationRepo *repositories.NotificationRepository
}

// NewNotificationService creates a new notifications service
func NewNotificationService(notificationRepo *repositories.NotificationRepository) *NotificationService {
	return &NotificationService{
		notificationRepo: notificationRepo,
	}
}

// Create creates a new notification
func (n *NotificationService) Create(notification models.Notification) error {
	// Set timestamps
	notification.CreatedAt = time.Now()
	notification.UpdatedAt = time.Now()

	// Ensure default values
	if notification.IsRead == false && notification.IsDeleted == false {
		// These are already false by default, but making it explicit
	}

	return n.notificationRepo.Create(&notification)
}

// CreateNotification creates a new notification with provided parameters
func (n *NotificationService) CreateNotification(performedBy, authorityLevel, notificationType, title, message, date string) error {
	notification := models.Notification{
		PerformedBy:    sql.NullString{String: performedBy, Valid: performedBy != ""},
		AuthorityLevel: authorityLevel,
		Type:           notificationType,
		Title:          title,
		Message:        sql.NullString{String: message, Valid: message != ""},
		Date:           date,
		IsRead:         false,
		IsDeleted:      false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	return n.notificationRepo.Create(&notification)
}

// Update updates an existing notification
func (n *NotificationService) Update(notification models.Notification) error {
	// Update the timestamp
	notification.UpdatedAt = time.Now()

	return n.notificationRepo.Update(&notification)
}

// UpdateNotification updates a notification by ID with new values
func (n *NotificationService) UpdateNotification(notificationId uint, title, message string) error {
	// First get the existing notification
	notification, err := n.notificationRepo.FindByID(notificationId)
	if err != nil {
		return fmt.Errorf("notification not found: %w", err)
	}

	// Update fields
	notification.Title = title
	notification.Message = sql.NullString{String: message, Valid: message != ""}
	notification.UpdatedAt = time.Now()

	return n.notificationRepo.Update(notification)
}

// Delete soft deletes a notification
func (n *NotificationService) Delete(notification models.Notification) error {
	return n.notificationRepo.Delete(notification.NotificationId)
}

// DeleteByID soft deletes a notification by ID
func (n *NotificationService) DeleteByID(notificationId uint) error {
	return n.notificationRepo.Delete(notificationId)
}

// HardDelete permanently deletes a notification
func (n *NotificationService) HardDelete(notificationId uint) error {
	return n.notificationRepo.HardDelete(notificationId)
}

// GetAll retrieves all notifications (excluding soft deleted ones)
func (n *NotificationService) GetAll() ([]notificationDto.NotificationDetails, error) {
	notifications, err := n.notificationRepo.FindAll()
	if err != nil {
		return nil, err
	}

	var notificationDetailsList = make([]notificationDto.NotificationDetails, 0)

	for _, notification := range notifications {
		notificationDetail := notificationDto.NotificationDetails{
			NotificationID: notification.NotificationId,
			PerformedBy:    getStringFromNullString(notification.PerformedBy),
			AuthorityLevel: notification.AuthorityLevel,
			Type:           notification.Type,
			Title:          notification.Title,
			Message:        getStringFromNullString(notification.Message),
			Date:           notification.Date,
			IsRead:         notification.IsRead,
			IsDeleted:      notification.IsDeleted,
			CreatedAt:      notification.CreatedAt.Format(time.RFC3339),
			UpdatedAt:      notification.UpdatedAt.Format(time.RFC3339),
		}
		notificationDetailsList = append(notificationDetailsList, notificationDetail)
	}

	return notificationDetailsList, nil
}

// Get retrieves a specific notification by ID
func (n *NotificationService) Get(notificationId uint) (models.Notification, error) {
	notification, err := n.notificationRepo.FindByID(notificationId)
	if err != nil {
		return models.Notification{}, err
	}
	return *notification, nil
}

// GetByAuthorityLevel retrieves notifications by authority level
func (n *NotificationService) GetByAuthorityLevel(authorityLevel string) ([]models.Notification, error) {
	return n.notificationRepo.FindByAuthorityLevel(authorityLevel)
}

// GetByType retrieves notifications by type
func (n *NotificationService) GetByType(notificationType string) ([]models.Notification, error) {
	return n.notificationRepo.FindByType(notificationType)
}

// GetUnread retrieves all unread notifications
func (n *NotificationService) GetUnread() ([]models.Notification, error) {
	return n.notificationRepo.FindUnread()
}

// GetUnreadByAuthorityLevel retrieves unread notifications by authority level
func (n *NotificationService) GetUnreadByAuthorityLevel(authorityLevel string) ([]models.Notification, error) {
	return n.notificationRepo.FindUnreadByAuthorityLevel(authorityLevel)
}

// MarkAsRead marks a notification as read
func (n *NotificationService) MarkAsRead(notificationId uint) error {
	return n.notificationRepo.MarkAsRead(notificationId)
}

// MarkAsUnread marks a notification as unread
func (n *NotificationService) MarkAsUnread(notificationId uint) error {
	return n.notificationRepo.MarkAsUnread(notificationId)
}

// MarkAllAsReadByAuthorityLevel marks all notifications as read for a specific authority level
func (n *NotificationService) MarkAllAsReadByAuthorityLevel(authorityLevel string) error {
	return n.notificationRepo.MarkAllAsReadByAuthorityLevel(authorityLevel)
}

// GetUnreadCount gets the count of unread notifications for a specific authority level
func (n *NotificationService) GetUnreadCount(authorityLevel string) (int64, error) {
	return n.notificationRepo.CountUnreadByAuthorityLevel(authorityLevel)
}

// GetTotalCount gets the total count of notifications for a specific authority level
func (n *NotificationService) GetTotalCount(authorityLevel string) (int64, error) {
	return n.notificationRepo.CountByAuthorityLevel(authorityLevel)
}

// GetWithPagination retrieves notifications with pagination
func (n *NotificationService) GetWithPagination(page, pageSize int) ([]models.Notification, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	return n.notificationRepo.FindWithPagination(offset, pageSize)
}

// GetByAuthorityLevelWithPagination retrieves notifications by authority level with pagination
func (n *NotificationService) GetByAuthorityLevelWithPagination(authorityLevel string, page, pageSize int) ([]models.Notification, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	return n.notificationRepo.FindByAuthorityLevelWithPagination(authorityLevel, offset, pageSize)
}

// CreateSystemNotification creates a system-generated notification
func (n *NotificationService) CreateSystemNotification(authorityLevel, notificationType, title, message string) error {
	return n.CreateNotification("SYSTEM", authorityLevel, notificationType, title, message, time.Now().Format("2006-01-02 15:04:05"))
}

// CreateUserNotification creates a user-generated notification
func (n *NotificationService) CreateUserNotification(performedBy, authorityLevel, notificationType, title, message string) error {
	return n.CreateNotification(performedBy, authorityLevel, notificationType, title, message, time.Now().Format("2006-01-02 15:04:05"))
}

// BulkMarkAsRead marks multiple notifications as read
func (n *NotificationService) BulkMarkAsRead(notificationIds []uint) error {
	for _, id := range notificationIds {
		if err := n.MarkAsRead(id); err != nil {
			return fmt.Errorf("failed to mark notification %d as read: %w", id, err)
		}
	}
	return nil
}

// BulkDelete soft deletes multiple notifications
func (n *NotificationService) BulkDelete(notificationIds []uint) error {
	for _, id := range notificationIds {
		if err := n.DeleteByID(id); err != nil {
			return fmt.Errorf("failed to delete notification %d: %w", id, err)
		}
	}
	return nil
}
