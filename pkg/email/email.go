// File: j-ticketing/pkg/email/email_service.go
package email

import (
	"crypto/tls"
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
		if cfg.Email.ClientID == "" {
			fmt.Println("- CLIENT_ID environment variable is not set")
		}
		if cfg.Email.ClientSecret == "" {
			fmt.Println("- CLIENT_SECRET environment variable is not set")
		}
		if cfg.Email.RefreshToken == "" {
			fmt.Println("- REFRESH_TOKEN environment variable is not set")
		}
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
	// First, log the authentication method we're going to use
	if s.useOAuth {
		fmt.Println("Sending email using OAuth2 authentication")
	} else {
		fmt.Println("Sending email using password authentication")
	}

	// Construct message
	message := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", strings.Join(to, ","), subject, body))

	// Connect to the server
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	fmt.Printf("Connecting to SMTP server at %s\n", addr)

	// Use standard SMTP for non-SSL as a simplification
	if !s.useSSL {
		fmt.Println("Using standard SMTP (non-SSL)")
		var auth smtp.Auth
		if s.useOAuth && s.tokenManager != nil {
			auth = NewSMTPOAuth2Auth(s.username, s.tokenManager)
		} else {
			auth = smtp.PlainAuth("", s.username, s.password, s.host)
		}

		err := smtp.SendMail(addr, auth, s.from, to, message)
		if err != nil {
			return fmt.Errorf("failed to send email: %w", err)
		}
		return nil
	}

	// For SSL, we'll implement the more verbose approach for better debugging
	fmt.Println("Using SSL for SMTP connection")

	// TLS config
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         s.host,
	}

	// Connect to the SMTP server with TLS
	fmt.Println("Establishing TLS connection...")
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	fmt.Println("Creating SMTP client...")
	client, err := smtp.NewClient(conn, s.host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	// Auth based on configuration
	if s.useOAuth && s.tokenManager != nil {
		fmt.Println("Attempting OAuth2 authentication...")
		// Get a fresh token for debugging
		token, err := s.tokenManager.GetToken()
		if err != nil {
			fmt.Printf("Failed to get OAuth2 token: %v\n", err)
		} else {
			fmt.Printf("Successfully retrieved OAuth2 token (length: %d)\n", len(token))
		}

		// Use OAuth2 authentication
		auth := NewSMTPOAuth2Auth(s.username, s.tokenManager)
		if err = client.Auth(auth); err != nil {
			fmt.Printf("OAuth2 authentication failed: %v\n", err)

			// If OAuth fails and we have a password, try password auth as fallback
			if s.password != "" {
				fmt.Println("Falling back to password authentication")
				fallbackAuth := smtp.PlainAuth("", s.username, s.password, s.host)
				if err = client.Auth(fallbackAuth); err != nil {
					return fmt.Errorf("both OAuth2 and password authentication failed: %w", err)
				}
				fmt.Println("Password authentication succeeded as fallback")
			} else {
				return fmt.Errorf("SMTP OAuth2 authentication failed and no password fallback available: %w", err)
			}
		} else {
			fmt.Println("OAuth2 authentication succeeded")
		}
	} else {
		// Use standard authentication
		if s.password == "" {
			return fmt.Errorf("no authentication method available: OAuth2 is disabled and password is empty")
		}

		fmt.Println("Attempting password authentication...")
		auth := smtp.PlainAuth("", s.username, s.password, s.host)
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP password authentication failed: %w", err)
		}
		fmt.Println("Password authentication succeeded")
	}

	// Set sender and recipients
	fmt.Println("Setting sender and recipients...")
	if err = client.Mail(s.from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, recipient := range to {
		if err = client.Rcpt(recipient); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
		}
	}

	// Send the message
	fmt.Println("Sending message...")
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to start email data: %w", err)
	}

	_, err = writer.Write(message)
	if err != nil {
		return fmt.Errorf("failed to write email data: %w", err)
	}

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close email data writer: %w", err)
	}

	fmt.Println("Email sent successfully")
	return client.Quit()
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

	// If we're using OAuth and it's Gmail, try multiple implementations
	if s.useOAuth && s.tokenManager != nil && strings.Contains(s.host, "gmail.com") {
		// Array to collect errors
		var errors []error

		// Try the simple direct implementation first
		fmt.Println("Trying Simple Direct implementation...")
		err := SendSimpleGmailEmail(
			s.username,
			[]string{to},
			subject,
			body,
			s.tokenManager.config.ClientID,
			s.tokenManager.config.ClientSecret,
			s.tokenManager.config.RefreshToken,
		)

		if err == nil {
			return nil
		}

		errors = append(errors, fmt.Errorf("simple direct method failed: %v", err))
		fmt.Printf("Simple Direct method failed: %v\n", err)

		// Try the Direct Gmail API method
		fmt.Println("Trying Direct Gmail API method...")
		err = SendDirectGmailEmail(
			s.username,
			[]string{to},
			subject,
			body,
			s.tokenManager.config.ClientID,
			s.tokenManager.config.ClientSecret,
			s.tokenManager.config.RefreshToken,
		)

		if err == nil {
			return nil
		}

		errors = append(errors, fmt.Errorf("gmail API method failed: %v", err))
		fmt.Printf("Gmail API method failed: %v\n", err)

		// Try the XOAUTH2 method
		fmt.Println("Trying XOAUTH2 method...")
		err = SendGmailWithXOAuth2(
			s.username,
			[]string{to},
			subject,
			body,
			s.tokenManager.config.ClientID,
			s.tokenManager.config.ClientSecret,
			s.tokenManager.config.RefreshToken,
		)

		if err == nil {
			return nil
		}

		errors = append(errors, fmt.Errorf("XOAUTH2 method failed: %v", err))
		fmt.Printf("XOAUTH2 method failed: %v\n", err)

		// If all methods failed, return a combined error
		return fmt.Errorf("all Gmail sending methods failed: %v", errors)
	}

	// Otherwise, use the standard implementation
	return s.SendEmail([]string{to}, subject, body)
}
