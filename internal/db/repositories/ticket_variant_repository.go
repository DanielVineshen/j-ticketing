// File: j-ticketing/internal/db/repositories/ticket_variant_repository.go
package repositories

import (
	"j-ticketing/internal/db/models"

	"gorm.io/gorm"
)

// TicketVariantRepository handles database operations for ticket variants
type TicketVariantRepository struct {
	db *gorm.DB
}

// NewTicketVariantRepository creates a new ticket variant repository
func NewTicketVariantRepository(db *gorm.DB) *TicketVariantRepository {
	return &TicketVariantRepository{db: db}
}

// FindByTicketGroupID finds variants associated with a ticket group
func (r *TicketVariantRepository) FindByTicketGroupID(ticketGroupID uint) ([]models.TicketVariant, error) {
	var variants []models.TicketVariant
	result := r.db.Where("ticket_group_id = ?", ticketGroupID).Find(&variants)
	return variants, result.Error
}
