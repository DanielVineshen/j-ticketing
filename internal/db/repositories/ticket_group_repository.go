// File: j-ticketing/internal/db/repositories/ticket_group_repository.go
package repositories

import (
	"j-ticketing/internal/db/models"

	"gorm.io/gorm"
)

// TicketGroupRepository handles database operations for ticket groups
type TicketGroupRepository struct {
	Db *gorm.DB
}

// NewTicketGroupRepository creates a new ticket group repository
func NewTicketGroupRepository(db *gorm.DB) *TicketGroupRepository {
	return &TicketGroupRepository{Db: db}
}

// FindAll returns all ticket groups
func (r *TicketGroupRepository) FindAll() ([]models.TicketGroup, error) {
	var ticketGroups []models.TicketGroup
	result := r.Db.Find(&ticketGroups)
	return ticketGroups, result.Error
}

// FindByID finds a ticket group by ID
func (r *TicketGroupRepository) FindByID(id uint) (*models.TicketGroup, error) {
	var ticketGroup models.TicketGroup
	result := r.Db.First(&ticketGroup, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &ticketGroup, nil
}

// Create creates a new ticket group
func (r *TicketGroupRepository) Create(ticketGroup *models.TicketGroup) error {
	return r.Db.Create(ticketGroup).Error
}

// Update updates a ticket group
func (r *TicketGroupRepository) Update(ticketGroup *models.TicketGroup) error {
	return r.Db.Save(ticketGroup).Error
}

// Delete deletes a ticket group
func (r *TicketGroupRepository) Delete(id uint) error {
	return r.Db.Delete(&models.TicketGroup{}, id).Error
}

// FindActiveTicketGroups finds all active ticket groups
func (r *TicketGroupRepository) FindActiveTicketGroups() ([]models.TicketGroup, error) {
	var ticketGroups []models.TicketGroup
	result := r.Db.Where("is_active = ?", true).Find(&ticketGroups)
	return ticketGroups, result.Error
}

// GetContentTypeByUniqueExtension finds the content type for a group gallery by unique extension
func (r *TicketGroupRepository) GetContentTypeByUniqueExtension(uniqueExtension string) (string, error) {
	var contentType string
	result := r.Db.Model(&models.TicketGroup{}).
		Select("content_type").
		Where("unique_extension = ?", uniqueExtension).
		First(&contentType)

	return contentType, result.Error
}
