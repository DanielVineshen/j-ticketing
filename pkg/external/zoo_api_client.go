// File: j-ticketing/pkg/external/zoo_api_client.go
package external

import (
	"encoding/json"
	"fmt"
	"io"
	logger "log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ZooAPIClient is the client for the Zoo API
type ZooAPIClient struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

// TokenResponse represents the response of the token request
type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// TicketItem represents a ticket item from the Zoo API
type TicketItem struct {
	ItemId          string  `json:"ItemId"`
	UnitPrice       float64 `json:"UnitPrice"`
	ItemDescription string  `json:"ItemDescription"`
	ItemDesc1       string  `json:"ItemDesc1"`
	ItemDesc2       string  `json:"ItemDesc2"`
	PrintType       string  `json:"PrintType"`
	Qty             int     `json:"Qty"`
}

// NewZooAPIClient creates a new Zoo API client
func NewZooAPIClient(baseURL, username, password string) *ZooAPIClient {
	// Ensure baseURL has a protocol
	if baseURL != "" && !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		baseURL = "https://" + baseURL
	}

	return &ZooAPIClient{
		baseURL:    baseURL,
		username:   username,
		password:   password,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// GetToken gets an access token from the Zoo API
func (c *ZooAPIClient) GetToken() (string, error) {
	data := url.Values{}
	data.Set("grant_type", "password")
	data.Set("UserName", c.username)
	data.Set("Password", c.password)

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/Token", c.baseURL), strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute token request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("Failed to close response body", "error", err)
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %w", err)
	}

	return tokenResp.AccessToken, nil
}

// GetTicketItems gets the ticket items for a specific date
func (c *ZooAPIClient) GetTicketItems(ticketGroupName string, date string) ([]TicketItem, error) {
	// First, get a token
	token, err := c.GetToken()
	if err != nil {
		return nil, fmt.Errorf("failed to get token: %w", err)
	}

	var value string
	if ticketGroupName == "Zoo Johor" {
		value = "GetOnlineItem"
	} else {
		value = "GetOnlineItem2" // Used for botani
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/JohorZoo/%s?TranDate=%s", c.baseURL, value, date), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create ticket items request: %w", err)
	}

	// Set the authorization header
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute ticket items request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logger.Error("Failed to close response body", "error", err)
		}
	}(resp.Body)

	// Check the response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ticket items request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode the response
	var items []TicketItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, fmt.Errorf("failed to decode ticket items response: %w", err)
	}

	return items, nil
}
