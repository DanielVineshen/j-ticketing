// File: j-ticketing/pkg/email/email_service.go
package email

import (
	"fmt"
	"j-ticketing/pkg/config"
	"net/smtp"
	"strings"
	"time"
)

// EmailService is the interface for email operations
type EmailService interface {
	SendEmail(to []string, subject, body string) error
	SendPasswordResetEmail(to string, newPassword string) error
}

type emailService struct {
	from         string
	host         string
	port         string
	username     string
	password     string
	useSSL       bool
	useOAuth     bool
	tokenManager *OAuth2TokenManager
}

// SendPasswordResetEmail sends a password reset email
func (s *emailService) SendPasswordResetEmail(to string, newPassword string) error {
	subject := "Your Password Has Been Reset"

	// HTML email body with nice formatting
	body := fmt.Sprintf(`
    <html>
    <head>
        <style>
            body { font-family: Arial, sans-serif; margin: 0; padding: 20px; color: #333; }
            .container { max-width: 600px; margin: 0 auto; background-color: #f8f9fa; padding: 20px; border-radius: 5px; }
            .header { background-color: #007bff; color: white; padding: 10px; border-radius: 5px 5px 0 0; text-align: center; }
            .content { padding: 20px; background-color: white; border-radius: 0 0 5px 5px; }
            .password { font-family: monospace; font-size: 16px; background-color: #f0f0f0; padding: 10px; border-radius: 3px; margin: 10px 0; display: inline-block; }
            .footer { margin-top: 20px; font-size: 12px; color: #666; text-align: center; }
        </style>
    </head>
    <body>
        <div class="container">
            <div class="header">
                <h2>Password Reset</h2>
            </div>
            <div class="content">
                <p>Your password has been successfully reset.</p>
                <p>Your new password is: <span class="password">%s</span></p>
                <p>Please login with this new password and change it to something more memorable.</p>
                <p>If you did not request this password reset, please contact our support team immediately.</p>
            </div>
            <div class="footer">
                <p>This is an automated message, please do not reply to this email.</p>
                <p>&copy; %d Johor Ticketing System</p>
            </div>
        </div>
    </body>
    </html>
    `, newPassword, time.Now().Year())

	return s.SendEmail([]string{to}, subject, body)
}

// NewEmailService creates a new email service
func NewEmailService(cfg *config.Config) EmailService {
	var tokenManager *OAuth2TokenManager
	useOAuth := false

	// Check if OAuth2 credentials are configured
	if cfg.Email.ClientID != "" && cfg.Email.ClientSecret != "" && cfg.Email.RefreshToken != "" {
		useOAuth = true

		// Log OAuth configuration (for debugging)
		fmt.Printf("Initializing OAuth2 email service with:\n")
		fmt.Printf("- Username: %s\n", cfg.Email.Username)
		fmt.Printf("- Client ID length: %d characters\n", len(cfg.Email.ClientID))
		fmt.Printf("- Client Secret length: %d characters\n", len(cfg.Email.ClientSecret))
		fmt.Printf("- Refresh Token length: %d characters\n", len(cfg.Email.RefreshToken))

		tokenManager = NewOAuth2TokenManager(OAuth2Config{
			ClientID:     cfg.Email.ClientID,
			ClientSecret: cfg.Email.ClientSecret,
			RefreshToken: cfg.Email.RefreshToken,
		})
	} else {
		fmt.Println("OAuth2 is NOT enabled - missing one or more required credentials")
	}

	return &emailService{
		from:         cfg.Email.From,
		host:         cfg.Email.Host,
		port:         cfg.Email.Port,
		username:     cfg.Email.Username,
		password:     cfg.Email.Password,
		useSSL:       cfg.Email.UseSSL,
		useOAuth:     useOAuth,
		tokenManager: tokenManager,
	}
}

// SendEmail sends an email
func (s *emailService) SendEmail(to []string, subject, body string) error {
	// If OAuth is configured and it's Gmail, use the Gmail API (most reliable method)
	if s.useOAuth && s.tokenManager != nil && strings.Contains(s.host, "gmail.com") {
		return s.sendEmailViaGmailAPI(to, subject, body)
	}

	// Fallback to standard SMTP if OAuth is not configured
	return s.sendEmailViaSmtp(to, subject, body)
}

// sendEmailViaGmailAPI sends an email using the Gmail API
func (s *emailService) sendEmailViaGmailAPI(to []string, subject, body string) error {
	fmt.Println("Using Gmail API to send email...")

	// Use the direct Gmail API method that we know works
	return SendDirectGmailEmail(
		s.username,
		to,
		subject,
		body,
		s.tokenManager.config.ClientID,
		s.tokenManager.config.ClientSecret,
		s.tokenManager.config.RefreshToken,
	)
}

// sendEmailViaSmtp sends an email using standard SMTP
func (s *emailService) sendEmailViaSmtp(to []string, subject, body string) error {
	// Construct message
	message := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", strings.Join(to, ","), subject, body))

	// Connect to the server
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	// Use standard SMTP authentication
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	// Send the email
	err := smtp.SendMail(addr, auth, s.from, to, message)
	if err != nil {
		return fmt.Errorf("failed to send email via SMTP: %w", err)
	}

	return nil
}

// TestOAuth2 tests if OAuth2 token acquisition works
func (s *emailService) TestOAuth2() error {
	if !s.useOAuth || s.tokenManager == nil {
		return fmt.Errorf("OAuth2 is not configured")
	}

	// Try to get a token
	_, err := s.tokenManager.GetToken()
	if err != nil {
		return fmt.Errorf("failed to acquire OAuth2 token: %w", err)
	}

	return nil
}
