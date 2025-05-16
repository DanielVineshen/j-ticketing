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
	from     string
	host     string
	port     string
	username string
	password string
	useSSL   bool
}

// NewEmailService creates a new email service
func NewEmailService(cfg *config.Config) EmailService {
	return &emailService{
		from:     cfg.Email.From,
		host:     cfg.Email.Host,
		port:     cfg.Email.Port,
		username: cfg.Email.Username,
		password: cfg.Email.Password,
		useSSL:   cfg.Email.UseSSL,
	}
}

// SendEmail sends an email
func (s *emailService) SendEmail(to []string, subject, body string) error {
	// Construct message
	message := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=UTF-8\r\n"+
		"\r\n"+
		"%s\r\n", strings.Join(to, ","), subject, body))

	// Authentication
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	// Connect to the server
	addr := fmt.Sprintf("%s:%s", s.host, s.port)

	var err error
	if s.useSSL {
		// TLS config
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
			ServerName:         s.host,
		}

		// Connect to the SMTP server with TLS
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return fmt.Errorf("failed to connect to SMTP server: %w", err)
		}
		defer conn.Close()

		client, err := smtp.NewClient(conn, s.host)
		if err != nil {
			return fmt.Errorf("failed to create SMTP client: %w", err)
		}
		defer client.Close()

		// Auth
		if err = client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}

		// Set sender and recipients
		if err = client.Mail(s.from); err != nil {
			return fmt.Errorf("failed to set sender: %w", err)
		}

		for _, recipient := range to {
			if err = client.Rcpt(recipient); err != nil {
				return fmt.Errorf("failed to set recipient %s: %w", recipient, err)
			}
		}

		// Send the message
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

		return client.Quit()
	}

	// Use standard SMTP for non-SSL
	err = smtp.SendMail(addr, auth, s.from, to, message)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
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
