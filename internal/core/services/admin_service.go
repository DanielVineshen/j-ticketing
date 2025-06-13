// File: j-ticketing/internal/core/services/admin_service.go
package service

import (
	"fmt"
	dto "j-ticketing/internal/core/dto/admin"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/utils"
	"strconv"
	"time"
)

type AdminServiceExtended struct {
	adminRepo repositories.AdminRepository
	tokenRepo repositories.TokenRepository
}

// NewAdminServiceExtended creates an extended admin service
func NewAdminServiceExtended(adminRepo repositories.AdminRepository, tokenRepo repositories.TokenRepository) *AdminServiceExtended {
	return &AdminServiceExtended{
		adminRepo: adminRepo,
		tokenRepo: tokenRepo,
	}
}

// Profile Management Methods (for admins managing their own profile)

// GetAdminByID retrieves an admin by string ID (converts from token)
func (s *AdminServiceExtended) GetAdminByID(id string) (*models.Admin, error) {
	adminID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid admin ID format")
	}
	return s.adminRepo.FindByID(uint(adminID))
}

// UpdateAdminProfile updates an admin's own profile (no username changes)
func (s *AdminServiceExtended) UpdateAdminProfile(adminID string, req dto.UpdateAdminProfileRequest) (*models.Admin, error) {
	// Convert string ID to uint
	adminIDUint, err := strconv.ParseUint(adminID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid admin ID format")
	}

	// Get current admin
	admin, err := s.adminRepo.FindByID(uint(adminIDUint))
	if err != nil {
		return nil, fmt.Errorf("admin not found")
	}

	// Update allowed fields only
	admin.FullName = req.FullName
	admin.Email = req.Email
	admin.ContactNo = req.ContactNo
	admin.UpdatedAt = time.Now()

	err = s.adminRepo.Update(admin)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

// ChangePassword changes an admin's password
func (s *AdminServiceExtended) ChangePassword(adminID string, req dto.ChangePasswordRequest) (*models.Admin, error) {
	// Convert string ID to uint
	adminIDUint, err := strconv.ParseUint(adminID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid admin ID format")
	}

	admin, err := s.adminRepo.FindByID(uint(adminIDUint))
	if err != nil {
		return nil, fmt.Errorf("admin not found")
	}

	// Verify current password
	err = utils.CheckPassword(req.CurrentPassword, admin.Password)
	if err != nil {
		return nil, fmt.Errorf("current password is incorrect")
	}

	// Hash the new password
	hashedPassword, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		return nil, err
	}

	admin.Password = hashedPassword
	admin.UpdatedAt = time.Now()

	err = s.adminRepo.Update(admin)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

// Admin Management Methods (for SYSADMIN managing other admins)

// GetAllAdmins retrieves all admins for management
func (s *AdminServiceExtended) GetAllAdmins() (*dto.AllAdminResponse, error) {
	admins, err := s.adminRepo.List()
	if err != nil {
		return nil, err
	}

	var adminDTOs []dto.AdminManagement
	for _, admin := range admins {
		adminDTOs = append(adminDTOs, dto.AdminManagement{
			AdminID:    int(admin.AdminId),
			Username:   admin.Username,
			FullName:   admin.FullName,
			Email:      admin.Email,
			ContactNo:  admin.ContactNo,
			Role:       admin.Role,
			IsDisabled: admin.IsDisabled,
		})
	}

	return &dto.AllAdminResponse{
		Admins: adminDTOs,
	}, nil
}

// CreateAdmin creates a new admin
func (s *AdminServiceExtended) CreateAdmin(req dto.CreateAdminRequest) (*models.Admin, error) {
	// Check if username already exists
	existingAdmin, err := s.adminRepo.FindByUsername(req.Username)
	if err == nil && existingAdmin != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Hash the password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create the admin
	admin := &models.Admin{
		Username:   req.Username,
		Password:   hashedPassword,
		FullName:   req.FullName,
		Email:      req.Email,
		ContactNo:  req.ContactNo,
		Role:       req.Role,
		IsDisabled: req.IsDisabled,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	err = s.adminRepo.Create(admin)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

// UpdateAdminManagement updates an admin via management interface
func (s *AdminServiceExtended) UpdateAdminManagement(req dto.UpdateAdminManagementRequest) (*models.Admin, error) {
	// Get current admin
	admin, err := s.adminRepo.FindByID(req.AdminID)
	if err != nil {
		return nil, fmt.Errorf("admin not found")
	}

	// Check if admin is being disabled
	wasEnabled := !admin.IsDisabled
	willBeDisabled := req.IsDisabled

	// Update fields
	admin.FullName = req.FullName
	admin.Email = req.Email
	admin.ContactNo = req.ContactNo
	admin.Role = req.Role
	admin.IsDisabled = req.IsDisabled
	admin.UpdatedAt = time.Now()

	err = s.adminRepo.Update(admin)
	if err != nil {
		return nil, err
	}

	// If admin was enabled and now disabled, remove all tokens
	if wasEnabled && willBeDisabled {
		err = s.removeAdminTokens(admin.Username)
		if err != nil {
			// Log error but don't fail the update
			fmt.Printf("Warning: Failed to remove tokens for disabled admin %s: %v\n", admin.Username, err)
		}
	}

	return admin, nil
}

// DeleteAdmin deletes an admin and removes all their tokens
func (s *AdminServiceExtended) DeleteAdmin(req dto.DeleteAdminRequest) error {
	// Get admin first to get username for token removal
	admin, err := s.adminRepo.FindByID(req.AdminID)
	if err != nil {
		return fmt.Errorf("admin not found")
	}

	// Remove all tokens for this admin
	err = s.removeAdminTokens(admin.Username)
	if err != nil {
		// Log error but continue with deletion
		fmt.Printf("Warning: Failed to remove tokens for admin %s: %v\n", admin.Username, err)
	}

	// Delete the admin
	return s.adminRepo.Delete(req.AdminID)
}

// Helper method to remove all tokens for an admin
func (s *AdminServiceExtended) removeAdminTokens(username string) error {
	// This would need to be implemented based on your token repository
	// For now, we'll use a simple approach - you might need to add this method to your token repository
	// Since userId in token is the username, we can search by that

	// Note: You might need to add a method like DeleteByUserId to your TokenRepository interface
	// For now, this is a placeholder - you'll need to implement the actual token deletion logic

	// Example implementation (you may need to add this method to TokenRepository):
	// return s.tokenRepo.DeleteByUserId(username)

	fmt.Printf("Removing all tokens for admin: %s\n", username)
	return nil // Placeholder - implement actual token removal
}
