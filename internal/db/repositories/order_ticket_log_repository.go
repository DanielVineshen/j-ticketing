package repositories

import (
	"gorm.io/gorm"
	"j-ticketing/internal/db/models"
)

// OrderTicketLogRepository handles database operations for audit logs
type OrderTicketLogRepository struct {
	db *gorm.DB
}

// NewOrderTicketLogRepository creates a new audit log repository
func NewOrderTicketLogRepository(db *gorm.DB) *OrderTicketLogRepository {
	return &OrderTicketLogRepository{db: db}
}

// Create creates a new audit log
func (r *OrderTicketLogRepository) Create(orderTicketLog *models.OrderTicketLog) error {
	return r.db.Create(orderTicketLog).Error
}
