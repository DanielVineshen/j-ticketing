// File: j-ticketing/internal/db/repositories/gallery_group_repository.go
package repositories

import (
	"j-ticketing/internal/db/models"

	"gorm.io/gorm"
)

// GroupGalleryRepository handles database operations for group galleries
type GroupGalleryRepository struct {
	db *gorm.DB
}

// NewGroupGalleryRepository creates a new group gallery repository
func NewGroupGalleryRepository(db *gorm.DB) *GroupGalleryRepository {
	return &GroupGalleryRepository{db: db}
}

// FindAll returns all group galleries
func (r *GroupGalleryRepository) FindAll() ([]models.GroupGallery, error) {
	var groupGalleries []models.GroupGallery
	result := r.db.Find(&groupGalleries)
	return groupGalleries, result.Error
}

// FindByID finds a group gallery by ID
func (r *GroupGalleryRepository) FindByID(id uint) (*models.GroupGallery, error) {
	var groupGallery models.GroupGallery
	result := r.db.First(&groupGallery, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &groupGallery, nil
}

// FindByTicketGroupID finds group galleries by ticket group ID
func (r *GroupGalleryRepository) FindByTicketGroupID(ticketGroupID uint) ([]models.GroupGallery, error) {
	var groupGalleries []models.GroupGallery
	result := r.db.Where("ticket_group_id = ?", ticketGroupID).Find(&groupGalleries)
	return groupGalleries, result.Error
}

// Create creates a new group gallery
func (r *GroupGalleryRepository) Create(groupGallery *models.GroupGallery) error {
	return r.db.Create(groupGallery).Error
}

// Update updates a group gallery
func (r *GroupGalleryRepository) Update(groupGallery *models.GroupGallery) error {
	return r.db.Save(groupGallery).Error
}

// Delete deletes a group gallery
func (r *GroupGalleryRepository) Delete(id uint) error {
	return r.db.Delete(&models.GroupGallery{}, id).Error
}

// GetContentTypeByUniqueExtension finds the content type for a ticket group by unique extension
func (r *GroupGalleryRepository) GetContentTypeByUniqueExtension(uniqueExtension string) (string, error) {
	var contentType string
	result := r.db.Model(&models.GroupGallery{}).
		Select("content_type").
		Where("unique_extension = ?", uniqueExtension).
		First(&contentType)

	return contentType, result.Error
}
