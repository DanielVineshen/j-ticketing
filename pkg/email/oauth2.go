// File: j-ticketing/pkg/email/oauth2.go
package email

import (
	"fmt"
	"strings"
)

// OAuth2Config holds the configuration for OAuth2 authentication
type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	RefreshToken string
}

// OAuth2TokenManager manages OAuth2 tokens
type OAuth2TokenManager struct {
	config OAuth2Config
}

// NewOAuth2TokenManager creates a new OAuth2 token manager
func NewOAuth2TokenManager(config OAuth2Config) *OAuth2TokenManager {
	// Validate client ID format
	clientID := config.ClientID
	if !strings.Contains(clientID, ".apps.googleusercontent.com") {
		fmt.Println("Warning: Client ID doesn't appear to be in the standard Google format (should end with .apps.googleusercontent.com)")
		fmt.Println("This might cause authentication failures with Google's OAuth2 service")
	}

	// Trim any whitespace that might have been accidentally included
	config.ClientID = strings.TrimSpace(config.ClientID)
	config.ClientSecret = strings.TrimSpace(config.ClientSecret)
	config.RefreshToken = strings.TrimSpace(config.RefreshToken)

	// Log the first few characters of each credential for debugging
	fmt.Printf("OAuth2 configuration:\n")
	fmt.Printf("- Client ID prefix: %s\n", clientID[:min(10, len(clientID))])
	if strings.Contains(clientID, ".apps.googleusercontent.com") {
		fmt.Println("- Client ID appears to be in the correct format")
	}

	return &OAuth2TokenManager{
		config: config,
	}
}

// GetToken returns a valid OAuth2 token, refreshing if necessary
func (m *OAuth2TokenManager) GetToken() (string, error) {
	// Get a fresh token each time
	return getOAuth2Token(
		m.config.ClientID,
		m.config.ClientSecret,
		m.config.RefreshToken,
	)
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
