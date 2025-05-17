// File: j-ticketing/pkg/email/gmail_xoauth2.go
package email

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"
)

// SendGmailWithXOAuth2 sends an email using Gmail's XOAUTH2 authentication
// This is a simplified, focused implementation specifically for Gmail with OAuth2
func SendGmailWithXOAuth2(
	from string,
	to []string,
	subject string,
	body string,
	clientID string,
	clientSecret string,
	refreshToken string,
) error {
	// Create the OAuth2 token manager
	tokenManager := NewOAuth2TokenManager(OAuth2Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RefreshToken: refreshToken,
	})

	// Get the access token
	fmt.Println("Getting OAuth2 access token...")
	accessToken, err := tokenManager.GetToken()
	if err != nil {
		return fmt.Errorf("failed to get OAuth2 token: %w", err)
	}
	fmt.Printf("Successfully obtained OAuth2 token (length: %d characters)\n", len(accessToken))

	// Connect to Gmail SMTP server
	fmt.Println("Connecting to Gmail SMTP server...")
	smtpServer := "smtp.gmail.com"
	smtpPort := "587" // Always use 587 for STARTTLS
	conn, err := smtp.Dial(fmt.Sprintf("%s:%s", smtpServer, smtpPort))
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	// Start TLS
	fmt.Println("Starting TLS...")
	tlsConfig := &tls.Config{
		ServerName: smtpServer,
	}
	if err = conn.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}
	fmt.Println("TLS started successfully")

	// Authenticate with XOAUTH2
	fmt.Println("Authenticating with XOAUTH2...")
	auth := createXOAUTH2Token(from, accessToken)
	if err = conn.Auth(xoauth2Auth{
		username: from,
		token:    auth,
	}); err != nil {
		return fmt.Errorf("XOAUTH2 authentication failed: %w", err)
	}
	fmt.Println("XOAUTH2 authentication successful")

	// Set sender and recipients
	fmt.Printf("Setting sender: %s\n", from)
	if err = conn.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, recipient := range to {
		fmt.Printf("Adding recipient: %s\n", recipient)
		if err = conn.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to add recipient %s: %w", recipient, err)
		}
	}

	// Send email body
	fmt.Println("Sending email...")
	wc, err := conn.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	msg := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s", from, strings.Join(to, ", "), subject, body)

	if _, err = fmt.Fprintf(wc, msg); err != nil {
		return fmt.Errorf("failed to write email: %w", err)
	}

	if err = wc.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	// Quit the SMTP session
	if err = conn.Quit(); err != nil {
		return fmt.Errorf("failed to quit SMTP session: %w", err)
	}

	fmt.Println("Email sent successfully!")
	return nil
}

// createXOAUTH2Token creates a base64-encoded XOAUTH2 token
func createXOAUTH2Token(username, accessToken string) string {
	// Must use actual binary bytes for \x01, not the literal string "\x01"
	data := []byte("user=" + username + "\x01auth=Bearer " + accessToken + "\x01\x01")
	return base64.StdEncoding.EncodeToString(data)
}

// xoauth2Auth implements smtp.Auth for XOAUTH2
type xoauth2Auth struct {
	username string
	token    string
}

// Start begins the auth exchange
func (a xoauth2Auth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "XOAUTH2", []byte(a.token), nil
}

// Next handles server challenge
func (a xoauth2Auth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		// XOAUTH2 doesn't expect a challenge, so this is likely an error
		return nil, fmt.Errorf("unexpected challenge: %s", fromServer)
	}
	return nil, nil
}
