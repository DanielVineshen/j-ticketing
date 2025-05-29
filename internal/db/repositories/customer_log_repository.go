package repositories

import (
	"gorm.io/gorm"
	"j-ticketing/internal/db/models"
)

// CustomerLogRepository handles database operations for audit logs
type CustomerLogRepository struct {
	db *gorm.DB
}

// NewCustomerLogRepository creates a new audit log repository
func NewCustomerLogRepository(db *gorm.DB) *CustomerLogRepository {
	return &CustomerLogRepository{db: db}
}

// Create creates a new audit log
func (r *CustomerLogRepository) Create(customerLog *models.CustomerLog) error {
	return r.db.Create(customerLog).Error
}
