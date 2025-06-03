// File: j-ticketing/internal/core/dto/auth/auth_response.go
package dto

// TokenResponse represents the structure for a token response
type TokenResponse struct {
	TokenID      uint     `json:"tokenId,omitempty"`
	AccessToken  string   `json:"accessToken"`
	RefreshToken string   `json:"refreshToken"`
	TokenType    string   `json:"tokenType"`
	ExpiresIn    int64    `json:"expiresIn"`
	User         UserInfo `json:"user"`
}

type UserInfo struct {
	// Admin specific fields
	AdminID  uint   `json:"adminId,omitempty"`
	Username string `json:"username,omitempty"`
	Role     string `json:"role,omitempty"`

	// Customer specific fields
	CustID           string `json:"custId,omitempty"`
	IdentificationNo string `json:"identificationNo,omitempty"`

	// Common fields
	FullName   string `json:"fullName,omitempty"`
	Email      string `json:"email,omitempty"`
	ContactNo  string `json:"contactNo,omitempty"`
	IsDisabled bool   `json:"isDisabled,omitempty"`
}
