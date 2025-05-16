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

// FindAll returns all banners
func (r *BannerRepository) FindAll() ([]models.Banner, error) {
	var banners []models.Banner
	result := r.db.Find(&banners)
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

// FindByTicketGroupID finds banners by ticket group ID
func (r *BannerRepository) FindByTicketGroupID(ticketGroupID uint) ([]models.Banner, error) {
	var banners []models.Banner
	result := r.db.Where("ticket_group_id = ?", ticketGroupID).Find(&banners)
	return banners, result.Error
}
