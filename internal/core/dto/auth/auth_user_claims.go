// File: j-ticketing/internal/core/dto/auth/auth_user_claims.go
package dto

// UserClaims represents the claims in the JWT token
type UserClaims struct {
	UserID   string   `json:"userId"`
	Username string   `json:"username"`
	UserType string   `json:"userType"` // "admin" or "customer"
	FullName string   `json:"fullName"`
	Role     string   `json:"role"`  // For admin: "SYSADMIN", "ADMIN"; For customer: "CUSTOMER"
	Roles    []string `json:"roles"` // Multiple roles if needed
}
