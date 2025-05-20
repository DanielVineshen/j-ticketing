// File: j-ticketing/internal/core/dto/customer/customer_request.go
package dto

import (
	"j-ticketing/pkg/validation"
)

// UpdateCustomerRequest represents the structure for updating a customer profile
type UpdateCustomerRequest struct {
	Email            string `json:"email" validate:"required,email,max=255"`
	Password         string `json:"password" validate:"omitempty,min=8,max=100"` // Optional - can be null
	IdentificationNo string `json:"identificationNo" validate:"required,max=255"`
	FullName         string `json:"fullName" validate:"required,max=255"`
	ContactNo        string `json:"contactNo" validate:"required,max=255"` // Optional - can be null
}

// Validate validates the update customer request
func (r *UpdateCustomerRequest) Validate() error {
	return validation.ValidateStruct(r)
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"currentPassword" validate:"required,min=8"`
	NewPassword     string `json:"newPassword" validate:"required,min=8"`
}

// Validate validates the update customer request
func (r *ChangePasswordRequest) Validate() error {
	return validation.ValidateStruct(r)
}
