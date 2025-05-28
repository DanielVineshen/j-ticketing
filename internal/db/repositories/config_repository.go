package repositories

import (
	"gorm.io/gorm"
	"j-ticketing/internal/db/models"
)

// ConfigRepository handles database operations for audit logs
type ConfigRepository struct {
	db *gorm.DB
}

// NewConfigRepository creates a new audit log repository
func NewConfigRepository(db *gorm.DB) *ConfigRepository {
	return &ConfigRepository{db: db}
}

// Create creates a new audit log
func (r *ConfigRepository) Create(auditLog *models.Config) error {
	return r.db.Create(auditLog).Error
}
