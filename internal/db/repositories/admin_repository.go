// FILE: internal/repositories/admin_repository.go (full implementation)
package repositories

import (
	"j-ticketing/internal/db/models"

	"gorm.io/gorm"
)

// AdminRepository is the interface for admin database operations
type AdminRepository interface {
	Create(admin *models.Admin) error
	FindByID(id uint) (*models.Admin, error)
	FindByUsername(username string) (*models.Admin, error)
	Update(admin *models.Admin) error
	Delete(id uint) error
	List() ([]models.Admin, error)
}

type adminRepository struct {
	db *gorm.DB
}

// NewAdminRepository creates a new admin repository
func NewAdminRepository(db *gorm.DB) AdminRepository {
	return &adminRepository{
		db: db,
	}
}

// Create creates a new admin
func (r *adminRepository) Create(admin *models.Admin) error {
	return r.db.Create(admin).Error
}

// FindByID finds an admin by ID
func (r *adminRepository) FindByID(id uint) (*models.Admin, error) {
	var admin models.Admin
	err := r.db.Where("admin_id = ?", id).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// FindByUsername finds an admin by username
func (r *adminRepository) FindByUsername(username string) (*models.Admin, error) {
	var admin models.Admin
	err := r.db.Where("username = ?", username).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// Update updates an admin
func (r *adminRepository) Update(admin *models.Admin) error {
	return r.db.Save(admin).Error
}

// Delete deletes an admin
func (r *adminRepository) Delete(id uint) error {
	return r.db.Delete(&models.Admin{}, id).Error
}

// List lists all admins
func (r *adminRepository) List() ([]models.Admin, error) {
	var admins []models.Admin
	err := r.db.Find(&admins).Error
	if err != nil {
		return nil, err
	}
	return admins, nil
}
