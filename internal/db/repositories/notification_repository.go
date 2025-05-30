// File: j-ticketing/internal/db/repositories/notification_repository.go
package repositories

import (
	"j-ticketing/internal/db/models"

	"gorm.io/gorm"
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

// FindAll returns all notifications
func (r *NotificationRepository) FindAll() ([]models.Notification, error) {
	var notifications []models.Notification
	result := r.db.Order("created_at DESC").Find(&notifications)
	return notifications, result.Error
}

// FindUnreadWithLimit finds unread notifications with a limit
func (r *NotificationRepository) FindUnreadWithLimit(limit int) ([]models.Notification, error) {
	var notifications []models.Notification
	result := r.db.Where("is_read = ? AND is_deleted = ?", false, false).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications)
	return notifications, result.Error
}

// CountUnread counts all unread notifications
func (r *NotificationRepository) CountUnread() (int64, error) {
	var count int64
	result := r.db.Model(&models.Notification{}).
		Where("is_read = ? AND is_deleted = ?", false, false).
		Count(&count)
	return count, result.Error
}

// FindByID finds a notification by ID
func (r *NotificationRepository) FindByID(id uint) (*models.Notification, error) {
	var notification models.Notification
	result := r.db.First(&notification, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &notification, nil
}

// Update updates a notification
func (r *NotificationRepository) Update(notification *models.Notification) error {
	return r.db.Save(notification).Error
}

// Delete deletes a notification
func (r *NotificationRepository) Delete(id uint) error {
	return r.db.Delete(&models.Notification{}, id).Error
}

// Legacy methods - keeping for backward compatibility if needed elsewhere
// You can remove these if they're not used anywhere else in your codebase

// FindByDateRangeUnread finds unread notifications within a date range
func (r *NotificationRepository) FindByDateRangeUnread(startDate, endDate string, limit int) ([]models.Notification, error) {
	var notifications []models.Notification

	query := r.db.Where("is_read = ? AND is_deleted = ?", false, false)

	// Add date range filter if provided
	if startDate != "" && endDate != "" {
		query = query.Where("date >= ? AND date <= ?", startDate, endDate)
	}

	result := query.Order("created_at DESC").Limit(limit).Find(&notifications)
	return notifications, result.Error
}

// CountByDateRangeUnread counts unread notifications within a date range
func (r *NotificationRepository) CountByDateRangeUnread(startDate, endDate string) (int64, error) {
	var count int64

	query := r.db.Model(&models.Notification{}).Where("is_read = ? AND is_deleted = ?", false, false)

	// Add date range filter if provided
	if startDate != "" && endDate != "" {
		query = query.Where("date >= ? AND date <= ?", startDate, endDate)
	}

	result := query.Count(&count)
	return count, result.Error
}
