// File: j-ticketing/pkg/email/gmail_oauth2.go
package email

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"
)

// GmailOAuth2 implements an OAuth2 authenticator specifically for Gmail
type GmailOAuth2 struct {
	UserEmail string
	TokenMgr  *OAuth2TokenManager
}

// NewGmailOAuth2 creates a new Gmail-specific OAuth2 authenticator
func NewGmailOAuth2(email string, tokenMgr *OAuth2TokenManager) smtp.Auth {
	return &GmailOAuth2{
		UserEmail: email,
		TokenMgr:  tokenMgr,
	}
}

// Start begins the authentication with the server
func (a *GmailOAuth2) Start(server *smtp.ServerInfo) (string, []byte, error) {
	fmt.Println("Getting access token for OAuth2 authentication...")
	token, err := a.TokenMgr.GetToken()
	if err != nil {
		return "", nil, fmt.Errorf("failed to get OAuth2 token: %w", err)
	}
	fmt.Printf("Access token received (length: %d)\n", len(token))

	// Gmail's XOAUTH2 format - this format is critical and must be exact
	// The format is: "user=<email>\x01auth=Bearer <token>\x01\x01"
	// Note: Make sure the \x01 bytes are actually binary bytes, not the string "\x01"
	xoauth2Data := []byte("user=" + a.UserEmail + "\x01auth=Bearer " + token + "\x01\x01")

	// Encode with base64
	encodedAuth := base64.StdEncoding.EncodeToString(xoauth2Data)

	fmt.Printf("Using Gmail OAuth2 with email: %s\n", a.UserEmail)
	fmt.Printf("XOAUTH2 data length: %d, encoded length: %d\n", len(xoauth2Data), len(encodedAuth))

	// Print a sample of the token for debugging (be careful with sensitive data)
	fmt.Println("XOAUTH2 token structure validation:")
	fmt.Printf("- First 10 chars of encoded token: %s...\n", encodedAuth[:10])
	fmt.Printf("- Contains 'user=': %v\n", strings.Contains(string(xoauth2Data), "user="))
	fmt.Printf("- Contains 'auth=Bearer': %v\n", strings.Contains(string(xoauth2Data), "auth=Bearer"))

	// Gmail uses "XOAUTH2" as the auth method
	return "XOAUTH2", []byte(encodedAuth), nil
}

// Next handles server challenge-response
func (a *GmailOAuth2) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		// Gmail should not request more data for XOAUTH2
		return nil, fmt.Errorf("unexpected server challenge: %s", fromServer)
	}
	return nil, nil
}

// SendEmailWithGmailOAuth2 sends an email using Gmail's OAuth2 authentication
func SendEmailWithGmailOAuth2(config *EmailConfig, tokenMgr *OAuth2TokenManager, to []string, subject, body string) error {
	// For Gmail, use the configured port or default to 587
	smtpServer := "smtp.gmail.com"
	smtpPort := config.Port

	// Log the port being used
	fmt.Printf("Connecting to Gmail SMTP server on port %s\n", smtpPort)

	// Connect to SMTP server
	client, err := smtp.Dial(fmt.Sprintf("%s:%s", smtpServer, smtpPort))
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer client.Close()

	// If using port 587, we need to use STARTTLS
	// If using port 465, we're already on an SSL connection
	if smtpPort == "587" {
		// Use STARTTLS with proper TLS configuration
		tlsConfig := &tls.Config{
			ServerName:         smtpServer,
			InsecureSkipVerify: false, // Set to true if you want to skip TLS verification (not recommended for production)
		}

		fmt.Println("Starting TLS with Gmail SMTP server...")
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("STARTTLS failed: %w", err)
		}
		fmt.Println("TLS connection established successfully")
	} else if smtpPort != "465" {
		fmt.Printf("Warning: Using non-standard port %s for Gmail SMTP. Standard ports are 587 (STARTTLS) or 465 (SSL)\n", smtpPort)
	}

	// Authenticate with OAuth2
	fmt.Println("Attempting Gmail OAuth2 authentication...")
	auth := NewGmailOAuth2(config.Username, tokenMgr)
	if err = client.Auth(auth); err != nil {
		fmt.Printf("Gmail OAuth2 authentication failed: %v\n", err)
		return fmt.Errorf("OAuth2 authentication failed: %w", err)
	}
	fmt.Println("Gmail OAuth2 authentication succeeded")

	// Set the sender and recipient
	fmt.Printf("Setting sender: %s\n", config.Username)
	if err = client.Mail(config.Username); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, recipient := range to {
		fmt.Printf("Adding recipient: %s\n", recipient)
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send the email body
	fmt.Println("Opening data connection...")
	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data connection: %w", err)
	}

	fmt.Println("Writing email content...")
	// Write the email headers and body
	_, err = fmt.Fprintf(wc, "To: %s\r\n", to[0])
	if err != nil {
		return fmt.Errorf("failed to write recipient: %w", err)
	}

	_, err = fmt.Fprintf(wc, "From: %s\r\n", config.Username)
	if err != nil {
		return fmt.Errorf("failed to write sender: %w", err)
	}

	_, err = fmt.Fprintf(wc, "Subject: %s\r\n", subject)
	if err != nil {
		return fmt.Errorf("failed to write subject: %w", err)
	}

	_, err = fmt.Fprintf(wc, "Content-Type: text/html; charset=UTF-8\r\n\r\n")
	if err != nil {
		return fmt.Errorf("failed to write content type: %w", err)
	}

	_, err = fmt.Fprintf(wc, "%s\r\n", body)
	if err != nil {
		return fmt.Errorf("failed to write body: %w", err)
	}

	fmt.Println("Closing data connection...")
	if err = wc.Close(); err != nil {
		return fmt.Errorf("failed to close data connection: %w", err)
	}

	// Send the QUIT command and close the connection
	fmt.Println("Sending QUIT command...")
	err = client.Quit()
	if err != nil {
		return fmt.Errorf("failed to quit SMTP session: %w", err)
	}

	fmt.Println("Email sent successfully using Gmail OAuth2")
	return nil
}

// EmailConfig holds email configuration
type EmailConfig struct {
	Host         string
	Port         string
	Username     string
	Password     string
	From         string
	UseSSL       bool
	ClientID     string
	ClientSecret string
	RefreshToken string
}
