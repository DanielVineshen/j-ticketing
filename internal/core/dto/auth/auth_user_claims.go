// FILE: internal/auth/models/auth.go
package dto

// UserClaims represents the claims in the JWT token
type UserClaims struct {
	UserID   string   `json:"userId"`
	Username string   `json:"username"`
	UserType string   `json:"userType"` // "admin" or "customer"
	Role     string   `json:"role"`     // For admin: "SYSADMIN", "OWNER", "STAFF"; For customer: "CUSTOMER"
	Roles    []string `json:"roles"`    // Multiple roles if needed
}
