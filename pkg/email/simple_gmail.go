// File: j-ticketing/pkg/email/simple_gmail.go
package email

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Add golang.org/x/oauth2 and golang.org/x/oauth2/google to your go.mod file with:
// go get golang.org/x/oauth2
// go get golang.org/x/oauth2/google

// SimpleGmailSender uses a more direct approach to send Gmail
type SimpleGmailSender struct {
	Email        string
	ClientID     string
	ClientSecret string
	RefreshToken string
}

// NewSimpleGmailSender creates a new simple Gmail sender
func NewSimpleGmailSender(email, clientID, clientSecret, refreshToken string) *SimpleGmailSender {
	return &SimpleGmailSender{
		Email:        email,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RefreshToken: refreshToken,
	}
}

// SendEmail sends an email through Gmail
func (s *SimpleGmailSender) SendEmail(to []string, subject, body string) error {
	// Using the official Google OAuth2 library
	config := &oauth2.Config{
		ClientID:     s.ClientID,
		ClientSecret: s.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  "urn:ietf:wg:oauth:2.0:oob",
		Scopes:       []string{"https://mail.google.com/"},
	}

	// Get token using refresh token
	token := &oauth2.Token{
		RefreshToken: s.RefreshToken,
	}

	// Get a new token source using the refresh token
	tokenSource := config.TokenSource(context.Background(), token)

	// Get a new access token
	newToken, err := tokenSource.Token()
	if err != nil {
		return fmt.Errorf("failed to get access token: %v", err)
	}

	fmt.Printf("Successfully obtained new access token (length: %d)\n", len(newToken.AccessToken))

	// Connect to SMTP server directly
	fmt.Println("Connecting to Gmail SMTP server...")
	host := "smtp.gmail.com"
	port := "587"

	// Establish TCP connection
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer conn.Close()

	// Read server greeting
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read server greeting: %v", err)
	}
	fmt.Printf("Server: %s", buf[:n])

	// Send EHLO
	randomName := fmt.Sprintf("localhost-%d", rand.Intn(1000000))
	fmt.Printf("Client: EHLO %s\r\n", randomName)
	_, err = fmt.Fprintf(conn, "EHLO %s\r\n", randomName)
	if err != nil {
		return fmt.Errorf("failed to send EHLO: %v", err)
	}

	// Read server response to EHLO
	n, err = conn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read EHLO response: %v", err)
	}
	fmt.Printf("Server: %s", buf[:n])

	// Send STARTTLS
	fmt.Println("Client: STARTTLS")
	_, err = fmt.Fprintf(conn, "STARTTLS\r\n")
	if err != nil {
		return fmt.Errorf("failed to send STARTTLS: %v", err)
	}

	// Read server response to STARTTLS
	n, err = conn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read STARTTLS response: %v", err)
	}
	fmt.Printf("Server: %s", buf[:n])

	// Upgrade connection to TLS
	fmt.Println("Upgrading connection to TLS...")
	tlsConn := tls.Client(conn, &tls.Config{
		ServerName: host,
	})
	if err = tlsConn.Handshake(); err != nil {
		return fmt.Errorf("TLS handshake failed: %v", err)
	}

	// Send EHLO again after TLS
	fmt.Printf("Client: EHLO %s\r\n", randomName)
	_, err = fmt.Fprintf(tlsConn, "EHLO %s\r\n", randomName)
	if err != nil {
		return fmt.Errorf("failed to send EHLO after TLS: %v", err)
	}

	// Read server response to EHLO
	n, err = tlsConn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read EHLO response after TLS: %v", err)
	}
	fmt.Printf("Server: %s", buf[:n])

	// Generate XOAUTH2 string
	// The format must be: "user=<email>\x01auth=Bearer <token>\x01\x01"
	// Encoded in base64
	rawAuth := []byte(fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", s.Email, newToken.AccessToken))
	auth := base64.StdEncoding.EncodeToString(rawAuth)

	// Send AUTH XOAUTH2
	fmt.Println("Client: AUTH XOAUTH2")
	_, err = fmt.Fprintf(tlsConn, "AUTH XOAUTH2 %s\r\n", auth)
	if err != nil {
		return fmt.Errorf("failed to send AUTH XOAUTH2: %v", err)
	}

	// Read server response to AUTH
	n, err = tlsConn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read AUTH response: %v", err)
	}

	authResp := string(buf[:n])
	fmt.Printf("Server: %s", authResp)

	// Check if authentication failed
	if !strings.HasPrefix(authResp, "235") {
		return fmt.Errorf("authentication failed: %s", authResp)
	}

	fmt.Println("Authentication successful!")

	// Set up email
	fmt.Println("Setting up email...")

	// MAIL FROM
	fmt.Printf("Client: MAIL FROM:<%s>\r\n", s.Email)
	_, err = fmt.Fprintf(tlsConn, "MAIL FROM:<%s>\r\n", s.Email)
	if err != nil {
		return fmt.Errorf("failed to send MAIL FROM: %v", err)
	}

	// Read server response
	n, err = tlsConn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read MAIL FROM response: %v", err)
	}
	fmt.Printf("Server: %s", buf[:n])

	// RCPT TO for each recipient
	for _, recipient := range to {
		fmt.Printf("Client: RCPT TO:<%s>\r\n", recipient)
		_, err = fmt.Fprintf(tlsConn, "RCPT TO:<%s>\r\n", recipient)
		if err != nil {
			return fmt.Errorf("failed to send RCPT TO: %v", err)
		}

		// Read server response
		n, err = tlsConn.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to read RCPT TO response: %v", err)
		}
		fmt.Printf("Server: %s", buf[:n])
	}

	// DATA
	fmt.Println("Client: DATA")
	_, err = fmt.Fprintf(tlsConn, "DATA\r\n")
	if err != nil {
		return fmt.Errorf("failed to send DATA: %v", err)
	}

	// Read server response
	n, err = tlsConn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read DATA response: %v", err)
	}
	fmt.Printf("Server: %s", buf[:n])

	// Send email content
	date := time.Now().Format("Mon, 02 Jan 2006 15:04:05 -0700")
	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"Date: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n.\r\n", s.Email, strings.Join(to, ", "), subject, date, body)

	_, err = fmt.Fprintf(tlsConn, message)
	if err != nil {
		return fmt.Errorf("failed to send email content: %v", err)
	}

	// Read server response
	n, err = tlsConn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read email content response: %v", err)
	}
	fmt.Printf("Server: %s", buf[:n])

	// QUIT
	fmt.Println("Client: QUIT")
	_, err = fmt.Fprintf(tlsConn, "QUIT\r\n")
	if err != nil {
		return fmt.Errorf("failed to send QUIT: %v", err)
	}

	// Read server response
	n, err = tlsConn.Read(buf)
	if err != nil {
		return fmt.Errorf("failed to read QUIT response: %v", err)
	}
	fmt.Printf("Server: %s", buf[:n])

	fmt.Println("Email sent successfully!")
	return nil
}

// SendSimpleGmailEmail is a convenience function to send an email
func SendSimpleGmailEmail(
	email string,
	to []string,
	subject string,
	body string,
	clientID string,
	clientSecret string,
	refreshToken string,
) error {
	sender := NewSimpleGmailSender(email, clientID, clientSecret, refreshToken)
	return sender.SendEmail(to, subject, body)
}
