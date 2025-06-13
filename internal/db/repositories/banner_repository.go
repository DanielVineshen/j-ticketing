// File: j-ticketing/internal/db/repositories/banner_repository.go
package repositories

import (
	"j-ticketing/internal/db/models"

	"gorm.io/gorm"
)

// BannerRepository handles database operations for banners
type BannerRepository struct {
	db *gorm.DB
}

// NewBannerRepository creates a new banner repository
func NewBannerRepository(db *gorm.DB) *BannerRepository {
	return &BannerRepository{db: db}
}

// FindAll returns all banners ordered by placement
func (r *BannerRepository) FindAll() ([]models.Banner, error) {
	var banners []models.Banner
	result := r.db.Order("placement ASC").Find(&banners)
	return banners, result.Error
}

// FindByID finds a banner by ID
func (r *BannerRepository) FindByID(id uint) (*models.Banner, error) {
	var banner models.Banner
	result := r.db.First(&banner, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &banner, nil
}

// Create creates a new banner
func (r *BannerRepository) Create(banner *models.Banner) error {
	return r.db.Create(banner).Error
}

// Update updates an existing banner
func (r *BannerRepository) Update(banner *models.Banner) error {
	return r.db.Save(banner).Error
}

// Delete deletes a banner by ID
func (r *BannerRepository) Delete(id uint) error {
	return r.db.Delete(&models.Banner{}, id).Error
}

// UpdatePlacement updates the placement of a specific banner
func (r *BannerRepository) UpdatePlacement(bannerId uint, placement int) error {
	return r.db.Model(&models.Banner{}).
		Where("banner_id = ?", bannerId).
		Update("placement", placement).Error
}

// GetMaxPlacement returns the highest placement number
func (r *BannerRepository) GetMaxPlacement() (int, error) {
	var maxPlacement int
	result := r.db.Model(&models.Banner{}).
		Select("COALESCE(MAX(placement), 0)").
		Scan(&maxPlacement)
	return maxPlacement, result.Error
}

// FindByTicketGroupID finds banners by ticket group ID (existing method)
func (r *BannerRepository) FindByTicketGroupID(ticketGroupID uint) ([]models.Banner, error) {
	var banners []models.Banner
	result := r.db.Where("ticket_group_id = ?", ticketGroupID).Find(&banners)
	return banners, result.Error
}

// GetContentTypeByUniqueExtension finds the content type for a banner by unique extension (existing method)
func (r *BannerRepository) GetContentTypeByUniqueExtension(uniqueExtension string) (string, error) {
	var contentType string
	result := r.db.Model(&models.Banner{}).
		Select("content_type").
		Where("unique_extension = ?", uniqueExtension).
		First(&contentType)

	return contentType, result.Error
}
