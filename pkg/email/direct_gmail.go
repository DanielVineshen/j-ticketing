// File: j-ticketing/pkg/email/direct_gmail.go
package email

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// DirectGmailSender sends emails directly via Gmail API
type DirectGmailSender struct {
	Username     string
	ClientID     string
	ClientSecret string
	RefreshToken string
}

// NewDirectGmailSender creates a new direct Gmail sender
func NewDirectGmailSender(username, clientID, clientSecret, refreshToken string) *DirectGmailSender {
	return &DirectGmailSender{
		Username:     username,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RefreshToken: refreshToken,
	}
}

// GetAccessToken gets a fresh access token using the refresh token
func (s *DirectGmailSender) GetAccessToken() (string, error) {
	fmt.Println("Getting new access token from Google...")

	data := url.Values{}
	data.Set("client_id", s.ClientID)
	data.Set("client_secret", s.ClientSecret)
	data.Set("refresh_token", s.RefreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequest("POST", "https://oauth2.googleapis.com/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read and parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error response from Google: %s", body)
	}

	// Parse JSON response properly
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	err = json.Unmarshal(body, &tokenResp)
	if err != nil {
		// If JSON parsing fails, try the basic extraction as fallback
		responseStr := string(body)
		if !strings.Contains(responseStr, "access_token") {
			return "", fmt.Errorf("access_token not found in response: %s", responseStr)
		}
		accessToken := strings.Split(strings.Split(responseStr, "\"access_token\":\"")[1], "\"")[0]
		fmt.Printf("Successfully obtained access token (length: %d) [using fallback parser]\n", len(accessToken))
		return accessToken, nil
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("access_token is empty in response")
	}

	fmt.Printf("Successfully obtained access token (length: %d)\n", len(tokenResp.AccessToken))
	return tokenResp.AccessToken, nil
}

// SendEmail sends an email using the Gmail API directly
func (s *DirectGmailSender) SendEmail(to []string, subject, body string) error {
	// Get access token
	accessToken, err := s.GetAccessToken()
	if err != nil {
		return fmt.Errorf("failed to get access token: %v", err)
	}

	// Format the email message
	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", s.Username, strings.Join(to, ", "), subject, body)

	// Base64 encode the message
	encodedMessage := base64.URLEncoding.EncodeToString([]byte(message))

	// Prepare the API request
	reqBody := fmt.Sprintf(`{
		"raw": "%s"
	}`, encodedMessage)

	// Make the request to Gmail API
	req, err := http.NewRequest("POST", "https://gmail.googleapis.com/gmail/v1/users/me/messages/send", strings.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Check the response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error sending email via Gmail API: %s", responseBody)
	}

	fmt.Println("Email sent successfully via Gmail API!")
	return nil
}

// SendDirectGmailEmail is a convenience function to send an email via Gmail API
func SendDirectGmailEmail(
	from string,
	to []string,
	subject,
	body,
	clientID,
	clientSecret,
	refreshToken string,
) error {
	sender := NewDirectGmailSender(from, clientID, clientSecret, refreshToken)
	return sender.SendEmail(to, subject, body)
}
