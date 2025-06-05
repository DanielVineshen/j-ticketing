// File: j-ticketing/internal/core/dto/admin/admin_request.go
package dto

import (
	"j-ticketing/pkg/validation"
)

// UpdateAdminProfileRequest represents the structure for updating own admin profile
type UpdateAdminProfileRequest struct {
	FullName  string `json:"fullName" validate:"required,max=255"`
	Email     string `json:"email" validate:"required,email,max=255"`
	ContactNo string `json:"contactNo" validate:"required,max=255"`
}

// Validate validates the update admin profile request
func (r *UpdateAdminProfileRequest) Validate() error {
	return validation.ValidateStruct(r)
}

// ChangePasswordRequest represents the structure for changing password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required,min=8"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
}

// Validate validates the change password request
func (r *ChangePasswordRequest) Validate() error {
	return validation.ValidateStruct(r)
}

// CreateAdminRequest represents the structure for creating a new admin (management)
type CreateAdminRequest struct {
	Username   string `json:"username" validate:"required,max=255"`
	Password   string `json:"password" validate:"required,min=8"`
	FullName   string `json:"fullName" validate:"required,max=255"`
	Email      string `json:"email" validate:"required,email,max=255"`
	ContactNo  string `json:"contactNo" validate:"required,max=255"`
	Role       string `json:"role" validate:"required,oneof=ADMIN SYSADMIN MEMBER"`
	IsDisabled bool   `json:"isDisabled"`
}

// Validate validates the create admin request
func (r *CreateAdminRequest) Validate() error {
	return validation.ValidateStruct(r)
}

// UpdateAdminManagementRequest represents the structure for updating admin via management
type UpdateAdminManagementRequest struct {
	AdminID    uint   `json:"adminId" validate:"required"`
	FullName   string `json:"fullName" validate:"required,max=255"`
	Email      string `json:"email" validate:"required,email,max=255"`
	ContactNo  string `json:"contactNo" validate:"required,max=255"`
	Role       string `json:"role" validate:"required,oneof=ADMIN SYSADMIN MEMBER"`
	IsDisabled bool   `json:"isDisabled"`
}

// Validate validates the update admin management request
func (r *UpdateAdminManagementRequest) Validate() error {
	return validation.ValidateStruct(r)
}

// DeleteAdminRequest represents the structure for deleting an admin
type DeleteAdminRequest struct {
	AdminID uint `json:"adminId" validate:"required"`
}

// Validate validates the delete admin request
func (r *DeleteAdminRequest) Validate() error {
	return validation.ValidateStruct(r)
}
