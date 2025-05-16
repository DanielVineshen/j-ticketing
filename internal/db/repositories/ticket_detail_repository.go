// File: internal/db/repositories/ticket_detail_repository.go
package repositories

import (
	"j-ticketing/internal/db/models"

	"gorm.io/gorm"
)

// TicketDetailRepository handles database operations for ticket details
type TicketDetailRepository struct {
	db *gorm.DB
}

// NewTicketDetailRepository creates a new ticket detail repository
func NewTicketDetailRepository(db *gorm.DB) *TicketDetailRepository {
	return &TicketDetailRepository{db: db}
}

// FindByTicketGroupID finds details associated with a ticket group
func (r *TicketDetailRepository) FindByTicketGroupID(ticketGroupID uint) ([]models.TicketDetail, error) {
	var details []models.TicketDetail
	result := r.db.Where("ticket_group_id = ?", ticketGroupID).Find(&details)
	return details, result.Error
}
