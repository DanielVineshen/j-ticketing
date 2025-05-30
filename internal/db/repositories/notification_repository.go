// File: j-ticketing/internal/db/repositories/notification_repository.go
package repositories

import (
	"gorm.io/gorm"
	"j-ticketing/internal/db/models"
)

// NotificationRepository handles database operations for notifications
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new notification repository
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create creates a new notification
func (r *NotificationRepository) Create(notification *models.Notification) error {
	return r.db.Create(notification).Error
}

// Update updates an existing notification
func (r *NotificationRepository) Update(notification *models.Notification) error {
	return r.db.Save(notification).Error
}

// Delete soft deletes a notification by setting IsDeleted to true
func (r *NotificationRepository) Delete(notificationId uint) error {
	return r.db.Model(&models.Notification{}).
		Where("notification_id = ?", notificationId).
		Update("is_deleted", true).Error
}

// HardDelete permanently deletes a notification from the database
func (r *NotificationRepository) HardDelete(notificationId uint) error {
	return r.db.Delete(&models.Notification{}, notificationId).Error
}

// FindByID finds a notification by ID (excluding soft deleted ones)
func (r *NotificationRepository) FindByID(notificationId uint) (*models.Notification, error) {
	var notification models.Notification
	err := r.db.Where("notification_id = ? AND is_deleted = ?", notificationId, false).
		First(&notification).Error
	if err != nil {
		return nil, err
	}
	return &notification, nil
}

// FindAll retrieves all notifications (excluding soft deleted ones)
func (r *NotificationRepository) FindAll() ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Where("is_deleted = ?", false).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

// FindByAuthorityLevel retrieves notifications by authority level
func (r *NotificationRepository) FindByAuthorityLevel(authorityLevel string) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Where("authority_level = ? AND is_deleted = ?", authorityLevel, false).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

// FindByType retrieves notifications by type
func (r *NotificationRepository) FindByType(notificationType string) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Where("type = ? AND is_deleted = ?", notificationType, false).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

// FindUnread retrieves all unread notifications (excluding soft deleted ones)
func (r *NotificationRepository) FindUnread() ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Where("is_read = ? AND is_deleted = ?", false, false).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

// FindUnreadByAuthorityLevel retrieves unread notifications by authority level
func (r *NotificationRepository) FindUnreadByAuthorityLevel(authorityLevel string) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Where("authority_level = ? AND is_read = ? AND is_deleted = ?", authorityLevel, false, false).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

// MarkAsRead marks a notification as read
func (r *NotificationRepository) MarkAsRead(notificationId uint) error {
	return r.db.Model(&models.Notification{}).
		Where("notification_id = ?", notificationId).
		Update("is_read", true).Error
}

// MarkAsUnread marks a notification as unread
func (r *NotificationRepository) MarkAsUnread(notificationId uint) error {
	return r.db.Model(&models.Notification{}).
		Where("notification_id = ?", notificationId).
		Update("is_read", false).Error
}

// MarkAllAsReadByAuthorityLevel marks all notifications as read for a specific authority level
func (r *NotificationRepository) MarkAllAsReadByAuthorityLevel(authorityLevel string) error {
	return r.db.Model(&models.Notification{}).
		Where("authority_level = ? AND is_read = ? AND is_deleted = ?", authorityLevel, false, false).
		Update("is_read", true).Error
}

// CountUnreadByAuthorityLevel counts unread notifications for a specific authority level
func (r *NotificationRepository) CountUnreadByAuthorityLevel(authorityLevel string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("authority_level = ? AND is_read = ? AND is_deleted = ?", authorityLevel, false, false).
		Count(&count).Error
	return count, err
}

// FindWithPagination retrieves notifications with pagination
func (r *NotificationRepository) FindWithPagination(offset, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Where("is_deleted = ?", false).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

// FindByAuthorityLevelWithPagination retrieves notifications by authority level with pagination
func (r *NotificationRepository) FindByAuthorityLevelWithPagination(authorityLevel string, offset, limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	err := r.db.Where("authority_level = ? AND is_deleted = ?", authorityLevel, false).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

// CountByAuthorityLevel counts total notifications for a specific authority level
func (r *NotificationRepository) CountByAuthorityLevel(authorityLevel string) (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("authority_level = ? AND is_deleted = ?", authorityLevel, false).
		Count(&count).Error
	return count, err
}

// CountAll counts total notifications (excluding soft deleted ones)
func (r *NotificationRepository) CountAll() (int64, error) {
	var count int64
	err := r.db.Model(&models.Notification{}).
		Where("is_deleted = ?", false).
		Count(&count).Error
	return count, err
}
