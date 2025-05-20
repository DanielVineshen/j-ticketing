package email

import (
	logger "log/slog"
	"strings"
	"sync"
	"time"
)

// OAuth2Config holds the configuration for OAuth2 authentication
type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	RefreshToken string
}

// OAuth2TokenManager manages OAuth2 tokens
type OAuth2TokenManager struct {
	config      OAuth2Config
	cachedToken string
	tokenExpiry time.Time
	mutex       sync.Mutex
}

// NewOAuth2TokenManager creates a new OAuth2 token manager
func NewOAuth2TokenManager(config OAuth2Config) *OAuth2TokenManager {
	// Validate client ID format
	clientID := config.ClientID
	if !strings.Contains(clientID, ".apps.googleusercontent.com") {
		logger.Info("Warning: Client ID doesn't appear to be in the standard Google format (should end with .apps.googleusercontent.com)")
		logger.Info("This might cause authentication failures with Google's OAuth2 service")
	}

	// Trim any whitespace that might have been accidentally included
	config.ClientID = strings.TrimSpace(config.ClientID)
	config.ClientSecret = strings.TrimSpace(config.ClientSecret)
	config.RefreshToken = strings.TrimSpace(config.RefreshToken)

	// Log the first few characters of each credential for debugging
	logger.Info("OAuth2 configuration")

	// Log client ID prefix safely
	prefix := ""
	if len(clientID) > 0 {
		endIndex := min(10, len(clientID))
		prefix = clientID[:endIndex]
	}
	logger.Info("Client ID information",
		"prefix", prefix,
		"correctFormat", strings.Contains(clientID, ".apps.googleusercontent.com"))

	return &OAuth2TokenManager{
		config: config,
	}
}

// GetToken returns a valid OAuth2 token, refreshing if necessary
func (m *OAuth2TokenManager) GetToken() (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if we have a valid cached token
	// Use 5 minute buffer before expiry to avoid edge cases
	if m.cachedToken != "" {
		// Use time.Until instead of t.Sub(time.Now())
		timeUntilExpiry := time.Until(m.tokenExpiry)

		logger.Info("Token state",
			"timeUntilExpiry", timeUntilExpiry.String(),
			"willUseCached", timeUntilExpiry > 5*time.Minute)

		if timeUntilExpiry > 5*time.Minute {
			logger.Info("Using cached OAuth token")
			return m.cachedToken, nil
		}
	}

	logger.Info("Requesting fresh OAuth token")

	// Get a fresh token
	token, err := getOAuth2Token(
		m.config.ClientID,
		m.config.ClientSecret,
		m.config.RefreshToken,
	)

	if err != nil {
		return "", err
	}

	// Cache the token
	// Google tokens typically last 1 hour (3600 seconds)
	m.cachedToken = token
	m.tokenExpiry = time.Now().Add(55 * time.Minute) // Refresh 5 min before expiry

	logger.Info("Successfully cached new OAuth token",
		"expiresAt", m.tokenExpiry.Format(time.RFC3339))

	return token, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
