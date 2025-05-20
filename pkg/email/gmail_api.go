// File: j-ticketing/pkg/email/gmail_api.go
package email

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	logger "log/slog"
	"net/http"
	"net/url"
	"strings"
)

// SendDirectGmailEmail is a convenience function to send an email via Gmail API
func SendDirectGmailEmail(
	from string,
	to []string,
	subject,
	body string,
	tokenManager *OAuth2TokenManager,
) error {
	// Get access token using the token manager
	accessToken, err := tokenManager.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get access token: %w", err)
	}

	// Rest of the function is the same...
	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", from, strings.Join(to, ", "), subject, body)

	// Base64 encode the message
	encodedMessage := base64.URLEncoding.EncodeToString([]byte(message))

	// Prepare the API request
	reqBody := fmt.Sprintf(`{
        "raw": "%s"
    }`, encodedMessage)

	// Make the request to Gmail API
	req, err := http.NewRequest("POST", "https://gmail.googleapis.com/gmail/v1/users/me/messages/send", strings.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("Failed to close response body", "error", err)
		}
	}(resp.Body)

	// Check the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error sending email via Gmail API: %s", responseBody)
	}

	logger.Info("Email sent successfully via Gmail API!")
	return nil
}

// getOAuth2Token gets a fresh access token using the refresh token
func getOAuth2Token(clientID, clientSecret, refreshToken string) (string, error) {
	logger.Info("Getting new access token from Google...")

	data := url.Values{}
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequest("POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("Failed to close response body", "error", err)
		}
	}(resp.Body)

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error response from Google: %s", body)
	}

	// Parse JSON response
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("access_token is empty in response")
	}

	logger.Info("Successfully obtained access token", "length", len(tokenResp.AccessToken))
	return tokenResp.AccessToken, nil
}
