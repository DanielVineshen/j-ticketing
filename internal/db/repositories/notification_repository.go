package repositories

import (
	"gorm.io/gorm"
	"j-ticketing/internal/db/models"
)

// NotificationRepository handles database operations for audit logs
type NotificationRepository struct {
	db *gorm.DB
}

// NewNotificationRepository creates a new audit log repository
func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

// Create creates a new audit log
func (r *NotificationRepository) Create(auditLog *models.Notification) error {
	return r.db.Create(auditLog).Error
}
