// File: j-ticketing/internal/core/dto/auth/auth_request.go
package dto

import (
	"j-ticketing/pkg/validation"
)

// LoginRequest represents the structure for a login request
type LoginRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=6,max=100"`
	UserType string `json:"userType" validate:"required,oneof=admin customer"` // "admin" or "customer"
}

// Validate validates the login request
func (r *LoginRequest) Validate() error {
	return validation.ValidateStruct(r)
}

// CreateCustomerRequest represents the structure for creating a new customer
type CreateCustomerRequest struct {
	Email            string `json:"email" validate:"required,email,max=255"`
	Password         string `json:"password" validate:"omitempty,min=8,max=100"` // Optional - can be null
	IdentificationNo string `json:"identificationNo" validate:"required,max=255"`
	FullName         string `json:"fullName" validate:"required,max=255"`
	ContactNo        string `json:"contactNo" validate:"required,max=255"` // Optional - can be null
}

// Validate validates the create customer request
func (r *CreateCustomerRequest) Validate() error {
	return validation.ValidateStruct(r)
}

// ResetPasswordRequest represents the structure for a password reset request
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email,max=255"`
}

// Validate validates the reset password request
func (r *ResetPasswordRequest) Validate() error {
	return validation.ValidateStruct(r)
}

// PasswordChangeResult represents the structure for a password change result
type PasswordChangeResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	// The new password won't be included in the API response for security
	// It will only be sent in the email
}
