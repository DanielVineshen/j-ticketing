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
