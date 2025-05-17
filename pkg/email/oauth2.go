// File: j-ticketing/pkg/email/oauth2.go
package email

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/smtp"
	"net/url"
	"strings"
	"sync"
	"time"
)

// OAuth2Token represents an OAuth2 access token
type OAuth2Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope,omitempty"`
	ExpiresAt   time.Time
}

// OAuth2Config holds the configuration for OAuth2 authentication
type OAuth2Config struct {
	ClientID     string
	ClientSecret string
	RefreshToken string
}

// OAuth2TokenManager manages OAuth2 tokens
type OAuth2TokenManager struct {
	config       OAuth2Config
	currentToken *OAuth2Token
	mutex        sync.Mutex
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

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetToken returns a valid OAuth2 token, refreshing if necessary
func (m *OAuth2TokenManager) GetToken() (string, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// If we have a token and it's still valid, return it
	if m.currentToken != nil && time.Now().Before(m.currentToken.ExpiresAt) {
		return m.currentToken.AccessToken, nil
	}

	// Otherwise, we need to refresh the token
	token, err := m.refreshToken()
	if err != nil {
		return "", fmt.Errorf("failed to refresh token: %w", err)
	}

	m.currentToken = token
	return token.AccessToken, nil
}

// refreshToken obtains a new access token using the refresh token
func (m *OAuth2TokenManager) refreshToken() (*OAuth2Token, error) {
	tokenURL := "https://oauth2.googleapis.com/token"

	data := url.Values{}
	data.Set("client_id", m.config.ClientID)
	data.Set("client_secret", m.config.ClientSecret)
	data.Set("refresh_token", m.config.RefreshToken)
	data.Set("grant_type", "refresh_token")

	// Debug logging for troubleshooting (be careful not to log sensitive data in production)
	fmt.Printf("Using Client ID: %s...[truncated]\n", m.config.ClientID[:8])

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to make token request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, body)
	}

	var token OAuth2Token
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Set expiration time (subtract 5 minutes for safety margin)
	token.ExpiresAt = time.Now().Add(time.Duration(token.ExpiresIn-300) * time.Second)

	return &token, nil
}

// GenerateXOAuth2Token generates an XOAUTH2 token for SMTP authentication
func GenerateXOAuth2Token(username, accessToken string) string {
	auth := fmt.Sprintf("user=%sauth=Bearer %s\x01", username, accessToken)
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// SMTPOAuth2Auth implements smtp.Auth interface for OAuth2
type SMTPOAuth2Auth struct {
	username   string
	tokenMgr   *OAuth2TokenManager
	xoauth2Str string
	expiry     time.Time
}

// NewSMTPOAuth2Auth creates a new SMTP OAuth2 authenticator
func NewSMTPOAuth2Auth(username string, tokenMgr *OAuth2TokenManager) *SMTPOAuth2Auth {
	return &SMTPOAuth2Auth{
		username: username,
		tokenMgr: tokenMgr,
	}
}

// Start implements smtp.Auth interface
func (a *SMTPOAuth2Auth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	// Refresh token if needed
	if a.xoauth2Str == "" || time.Now().After(a.expiry) {
		token, err := a.tokenMgr.GetToken()
		if err != nil {
			return "", nil, err
		}

		// Format: "user=<username>\x01auth=Bearer <token>\x01\x01"
		auth := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", a.username, token)
		a.xoauth2Str = base64.StdEncoding.EncodeToString([]byte(auth))

		// Set expiry for 5 minutes less than the actual token expiry
		a.expiry = time.Now().Add(55 * time.Minute)

		fmt.Printf("Generated XOAUTH2 token of length %d\n", len(a.xoauth2Str))
	}

	return "XOAUTH2", []byte(a.xoauth2Str), nil
}

// Next implements smtp.Auth interface
func (a *SMTPOAuth2Auth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		// The server sent an error; we should decode it
		resp := ""
		if len(fromServer) > 0 {
			resp = string(fromServer)
		}
		return nil, fmt.Errorf("unexpected server response: %s", resp)
	}
	return nil, nil
}
