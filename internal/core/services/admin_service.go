// File: j-ticketing/internal/core/services/admin_service.go
package service

import (
	"fmt"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	bcryptPassword "j-ticketing/pkg/utils"
	"time"
)

// AdminService handles admin-related operations
type AdminService interface {
	CreateAdmin(username, password, fullName, role string) (*models.Admin, error)
	GetAdminByID(id uint) (*models.Admin, error)
	UpdateAdmin(id uint, fullName, role string) (*models.Admin, error)
	ChangePassword(id uint, currentPassword, newPassword string) error
	DeleteAdmin(id uint) error
	ListAdmins() ([]models.Admin, error)
}

type adminService struct {
	adminRepo repositories.AdminRepository
}

// NewAdminService creates a new admin service
func NewAdminService(adminRepo repositories.AdminRepository) AdminService {
	return &adminService{
		adminRepo: adminRepo,
	}
}

// CreateAdmin creates a new admin user
func (s *adminService) CreateAdmin(username, password, fullName, role string) (*models.Admin, error) {
	// Check if username already exists
	existingAdmin, err := s.adminRepo.FindByUsername(username)
	if err == nil && existingAdmin != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Hash the password
	hashedPassword, err := bcryptPassword.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create the admin user
	admin := &models.Admin{
		Username:  username,
		Password:  hashedPassword,
		FullName:  fullName,
		Role:      role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	err = s.adminRepo.Create(admin)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

// GetAdminByID retrieves an admin by ID
func (s *adminService) GetAdminByID(id uint) (*models.Admin, error) {
	return s.adminRepo.FindByID(id)
}

// UpdateAdmin updates an admin's information
func (s *adminService) UpdateAdmin(id uint, fullName, role string) (*models.Admin, error) {
	admin, err := s.adminRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	admin.FullName = fullName
	admin.Role = role
	admin.UpdatedAt = time.Now()

	err = s.adminRepo.Update(admin)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

// ChangePassword changes an admin's password
func (s *adminService) ChangePassword(id uint, currentPassword, newPassword string) error {
	admin, err := s.adminRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Verify current password
	err = bcryptPassword.CheckPassword(currentPassword, admin.Password)
	if err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash the new password
	hashedPassword, err := bcryptPassword.HashPassword(newPassword)
	if err != nil {
		return err
	}

	admin.Password = hashedPassword
	admin.UpdatedAt = time.Now()

	return s.adminRepo.Update(admin)
}

// DeleteAdmin deletes an admin
func (s *adminService) DeleteAdmin(id uint) error {
	return s.adminRepo.Delete(id)
}

// ListAdmins lists all admins
func (s *adminService) ListAdmins() ([]models.Admin, error) {
	admins, err := s.adminRepo.List()
	if err != nil {
		return nil, err
	}

	// Remove password from response
	for i := range admins {
		admins[i].Password = ""
	}

	return admins, nil
}
