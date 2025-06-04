// File: j-ticketing/internal/db/repositories/general_repository.go
package repositories

import (
	"j-ticketing/internal/db/models"

	"gorm.io/gorm"
)

// GeneralRepository handles database operations for general settings
type GeneralRepository struct {
	db *gorm.DB
}

// NewGeneralRepository creates a new general repository
func NewGeneralRepository(db *gorm.DB) *GeneralRepository {
	return &GeneralRepository{db: db}
}

// FindFirst returns the first (and usually only) general settings record
func (r *GeneralRepository) FindFirst() (*models.General, error) {
	var general models.General
	result := r.db.First(&general)
	if result.Error != nil {
		return nil, result.Error
	}
	return &general, nil
}

// Update updates the general settings record
func (r *GeneralRepository) Update(general *models.General) error {
	return r.db.Save(general).Error
}

// GetContentTypeByUniqueExtension finds the content type for an attachment by unique extension
func (r *GeneralRepository) GetContentTypeByUniqueExtension(uniqueExtension string) (string, error) {
	var contentType string
	result := r.db.Model(&models.General{}).
		Select("content_type").
		Where("unique_extension = ?", uniqueExtension).
		First(&contentType)

	return contentType, result.Error
}
