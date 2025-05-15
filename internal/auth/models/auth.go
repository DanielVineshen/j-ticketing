// FILE: internal/auth/models/auth.go
package models

// LoginRequest represents the structure for a login request
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	UserType string `json:"userType"` // "admin" or "customer"
}

// TokenResponse represents the structure for a token response
type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	TokenType    string `json:"tokenType"`
	ExpiresIn    int64  `json:"expiresIn"` // in seconds
}

// UserClaims represents the claims in the JWT token
type UserClaims struct {
	UserID   string   `json:"userId"`
	Username string   `json:"username"`
	UserType string   `json:"userType"` // "admin" or "customer"
	Role     string   `json:"role"`     // For admin: "SYSADMIN", "OWNER", "STAFF"; For customer: "CUSTOMER"
	Roles    []string `json:"roles"`    // Multiple roles if needed
}
