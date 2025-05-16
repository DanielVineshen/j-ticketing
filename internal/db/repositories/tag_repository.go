// File: j-ticketing/internal/db/repositories/tag_repository.go
package repositories

import (
	"j-ticketing/internal/db/models"

	"gorm.io/gorm"
)

// TagRepository handles database operations for tags
type TagRepository struct {
	db *gorm.DB
}

// NewTagRepository creates a new tag repository
func NewTagRepository(db *gorm.DB) *TagRepository {
	return &TagRepository{db: db}
}

// FindAll returns all tags
func (r *TagRepository) FindAll() ([]models.Tag, error) {
	var tags []models.Tag
	result := r.db.Find(&tags)
	return tags, result.Error
}

// FindByID finds a tag by ID
func (r *TagRepository) FindByID(id uint) (*models.Tag, error) {
	var tag models.Tag
	result := r.db.First(&tag, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &tag, nil
}

// FindByName finds a tag by name
func (r *TagRepository) FindByName(name string) (*models.Tag, error) {
	var tag models.Tag
	result := r.db.Where("tag_name = ?", name).First(&tag)
	if result.Error != nil {
		return nil, result.Error
	}
	return &tag, nil
}

// Create creates a new tag
func (r *TagRepository) Create(tag *models.Tag) error {
	return r.db.Create(tag).Error
}

// Update updates a tag
func (r *TagRepository) Update(tag *models.Tag) error {
	return r.db.Save(tag).Error
}

// Delete deletes a tag
func (r *TagRepository) Delete(id uint) error {
	return r.db.Delete(&models.Tag{}, id).Error
}

// FindByTicketGroupID finds tags associated with a ticket group
func (r *TagRepository) FindByTicketGroupID(ticketGroupID uint) ([]models.Tag, error) {
	var tags []models.Tag
	result := r.db.Joins("JOIN Ticket_Tag ON Ticket_Tag.tag_id = Tag.tag_id").
		Where("Ticket_Tag.ticket_group_id = ?", ticketGroupID).
		Find(&tags)
	return tags, result.Error
}
