// File: j-ticketing/pkg/email/email.go
package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"j-ticketing/internal/db/models"
	logger "log/slog"
	"math/rand"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

// EmailService is the interface for email operations
type EmailService interface {
	SendEmail(to []string, subject, body string) error
	SendPasswordResetEmail(to string, newPassword string) error
	SendTicketsEmail(to string, orderOverview OrderOverview, orderItems []OrderInfo, tickets []TicketInfo, attachments []Attachment, language string) error
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

	return s.sendEmailWithRetry([]string{to}, subject, body)
}

// OrderInfo represents information about the overall order
type OrderOverview struct {
	TicketGroup  string
	FullName     string
	PurchaseDate string
	EntryDate    string
	Quantity     int
	OrderNumber  string
	Total        float64
}

// OrderInfo represents information about a single order item
type OrderInfo struct {
	Description string
	Quantity    int
	Price       float64
	EntryDate   string
}

// TicketInfo represents information for a single ticket
type TicketInfo struct {
	Label   string
	Content string
}

func (s *emailService) SendTicketsEmail(to string, orderOverview OrderOverview, orderItems []OrderInfo, tickets []TicketInfo, attachments []Attachment, language string) error {
	subject, body, qrAttachments, _ := sendTicketsEmail(orderOverview, orderItems, tickets, language)

	allAttachments := append(attachments, qrAttachments...)

	return s.sendEmailWithAttachmentsWithRetry([]string{to}, subject, body, allAttachments)
}

// NewEmailService creates a new email service
func NewEmailService(generalModel *models.General) EmailService {
	var tokenManager *OAuth2TokenManager
	useOAuth := false

	// Check if OAuth2 credentials are configured
	if generalModel.EmailClientId != "" && generalModel.EmailClientSecret != "" && generalModel.EmailRefreshToken != "" {
		useOAuth = true

		// Log OAuth configuration (for debugging)
		logger.Info("Initializing OAuth2 email service")
		logger.Info("OAuth2 email username", "value", generalModel.EmailUsername)
		logger.Info("OAuth2 client ID", "length", len(generalModel.EmailClientId))
		logger.Info("OAuth2 client secret", "length", len(generalModel.EmailClientSecret))
		logger.Info("OAuth2 refresh token", "length", len(generalModel.EmailRefreshToken))

		tokenManager = NewOAuth2TokenManager(OAuth2Config{
			ClientID:     generalModel.EmailClientId,
			ClientSecret: generalModel.EmailClientSecret,
			RefreshToken: generalModel.EmailRefreshToken,
		})
	} else {
		logger.Warn("OAuth2 is NOT enabled - missing one or more required credentials")
	}

	return &emailService{
		from:         generalModel.EmailFrom,
		host:         generalModel.EmailHost,
		port:         strconv.Itoa(generalModel.EmailPort),
		username:     generalModel.EmailUsername,
		password:     generalModel.EmailPassword,
		useSSL:       generalModel.EmailUseSsl,
		useOAuth:     useOAuth,
		tokenManager: tokenManager,
	}
}

func (s *emailService) sendEmailWithRetry(to []string, subject string, body string) error {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		logger.Info("Attempting to send email",
			"attempt", attempt+1,
			"recipients", to,
			"maxRetries", maxRetries)

		err := s.SendEmail(to, subject, body)
		if err == nil {
			logger.Info("Email sent successfully",
				"attempt", attempt+1,
				"recipients", to)
			return nil // Success
		}

		lastErr = err
		logger.Info("Email sending failed, will retry",
			"attempt", attempt+1,
			"error", err.Error(),
			"nextRetryIn", (attempt+1)*2,
			"recipients", to)

		// Wait before retrying (with exponential backoff)
		time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
	}

	return fmt.Errorf("sendEmailWithRetry failed after %d attempts: %w", maxRetries, lastErr)
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
	logger.Info("Using Gmail API to send email...")

	// Pass the token manager instead of individual credentials
	return SendDirectGmailEmail(
		s.username,
		to,
		subject,
		body,
		s.tokenManager,
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

func (s *emailService) sendEmailWithAttachmentsWithRetry(to []string, subject string, body string, attachments []Attachment) error {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		logger.Info("Attempting to send email",
			"attempt", attempt+1,
			"recipients", to,
			"maxRetries", maxRetries)

		err := s.SendEmailWithAttachments(to, subject, body, attachments)
		if err == nil {
			logger.Info("Email sent successfully",
				"attempt", attempt+1,
				"recipients", to)
			return nil // Success
		}

		lastErr = err
		logger.Info("Email sending failed, will retry",
			"attempt", attempt+1,
			"error", err.Error(),
			"nextRetryIn", (attempt+1)*2,
			"recipients", to)

		// Wait before retrying (with exponential backoff)
		time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
	}

	return fmt.Errorf("sendEmailWithAttachmentsWithRetry failed after %d attempts: %w", maxRetries, lastErr)
}

// SendEmailWithAttachments sends an email with optional attachments
func (s *emailService) SendEmailWithAttachments(to []string, subject, body string, attachments []Attachment) error {
	// If OAuth is configured and it's Gmail, use the Gmail API (most reliable method)
	if s.useOAuth && s.tokenManager != nil && strings.Contains(s.host, "gmail.com") {
		return s.sendEmailViaGmailAPIWithAttachments(to, subject, body, attachments)
	}

	// Fallback to standard SMTP if OAuth is not configured
	return s.sendEmailViaSmtpWithAttachments(to, subject, body, attachments)
}

// Attachment represents an email attachment
type Attachment struct {
	Name    string
	Content []byte
	Type    string
	CID     string
}

// sendEmailViaSmtpWithAttachments sends an email with attachments using SMTP
func (s *emailService) sendEmailViaSmtpWithAttachments(to []string, subject, body string, attachments []Attachment) error {
	// Create a unique boundary for MIME parts
	boundary := "==Boundary_" + randomString(30)

	// Build email headers
	var message bytes.Buffer
	message.WriteString("MIME-Version: 1.0\r\n")
	message.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	message.WriteString(fmt.Sprintf("From: %s\r\n", s.from))
	message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))

	// Check if we have CID attachments
	hasCIDAttachments := false
	for _, attachment := range attachments {
		if attachment.CID != "" {
			hasCIDAttachments = true
			break
		}
	}

	// Set content type based on whether there are attachments
	if hasCIDAttachments {
		// Use multipart/related for inline images with CID
		message.WriteString(fmt.Sprintf("Content-Type: multipart/related; boundary=%s\r\n\r\n", boundary))
	} else {
		// Use multipart/mixed for regular attachments
		message.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary))
	}

	// Add HTML body
	message.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	message.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	message.WriteString(body)
	message.WriteString("\r\n")

	// Add attachments if any
	for _, attachment := range attachments {
		message.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
		message.WriteString(fmt.Sprintf("Content-Type: %s\r\n", attachment.Type))

		if attachment.CID != "" {
			// Inline attachment with Content-ID
			message.WriteString(fmt.Sprintf("Content-ID: <%s>\r\n", attachment.CID))
			message.WriteString("Content-Disposition: inline\r\n")
		} else {
			// Regular attachment
			message.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", attachment.Name))
		}

		message.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")

		// Convert attachment to base64
		b64 := base64.StdEncoding.EncodeToString(attachment.Content)

		// Split the base64 string into lines of 76 characters
		lineLength := 76
		for i := 0; i < len(b64); i += lineLength {
			end := i + lineLength
			if end > len(b64) {
				end = len(b64)
			}
			message.WriteString(b64[i:end] + "\r\n")
		}
	}

	// Close the multipart message
	if len(attachments) > 0 {
		message.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))
	}

	// Connect to the server
	auth := smtp.PlainAuth("", s.username, s.password, s.host)

	// Send the email
	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	return smtp.SendMail(addr, auth, s.from, to, message.Bytes())
}

// sendEmailViaGmailAPIWithAttachments sends an email with attachments using Gmail API
func (s *emailService) sendEmailViaGmailAPIWithAttachments(to []string, subject, body string, attachments []Attachment) error {
	logger.Info("Starting Gmail API email delivery",
		"recipients", to,
		"subject", subject,
		"attachmentCount", len(attachments))

	// Get access token using token manager
	startTokenTime := time.Now()
	accessToken, err := s.tokenManager.GetToken()
	if err != nil {
		logger.Error("Failed to get access token",
			"error", err,
			"timeSpent", time.Since(startTokenTime))
		return fmt.Errorf("failed to get access token: %w", err)
	}
	logger.Info("Obtained access token", "timeSpent", time.Since(startTokenTime))

	// Prepare the MIME message with attachments
	startMimeTime := time.Now()
	var messageBuffer bytes.Buffer
	boundary := "==Boundary_" + randomString(30)

	// Add headers
	messageBuffer.WriteString(fmt.Sprintf("From: %s\r\n", s.username))
	messageBuffer.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(to, ", ")))
	messageBuffer.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	messageBuffer.WriteString("MIME-Version: 1.0\r\n")

	// Check if we have CID attachments
	hasCIDAttachments := false
	for _, attachment := range attachments {
		if attachment.CID != "" {
			hasCIDAttachments = true
			break
		}
	}

	// If we have attachments, create a multipart message
	if len(attachments) > 0 {
		if hasCIDAttachments {
			// Use multipart/related for inline images with CID
			messageBuffer.WriteString(fmt.Sprintf("Content-Type: multipart/related; boundary=%s\r\n\r\n", boundary))
		} else {
			// Use multipart/mixed for regular attachments
			messageBuffer.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary))
		}

		messageBuffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		messageBuffer.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		messageBuffer.WriteString(body)
		messageBuffer.WriteString("\r\n")

		// Add attachments
		for _, attachment := range attachments {
			messageBuffer.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
			messageBuffer.WriteString(fmt.Sprintf("Content-Type: %s\r\n", attachment.Type))

			if attachment.CID != "" {
				// Inline attachment with Content-ID
				messageBuffer.WriteString(fmt.Sprintf("Content-ID: <%s>\r\n", attachment.CID))
				messageBuffer.WriteString("Content-Disposition: inline\r\n")
			} else {
				// Regular attachment
				messageBuffer.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", attachment.Name))
			}

			messageBuffer.WriteString("Content-Transfer-Encoding: base64\r\n\r\n")

			// Encode attachment as base64 in 76-character lines
			encoded := base64.StdEncoding.EncodeToString(attachment.Content)
			for i := 0; i < len(encoded); i += 76 {
				end := i + 76
				if end > len(encoded) {
					end = len(encoded)
				}
				messageBuffer.WriteString(encoded[i:end] + "\r\n")
			}
		}

		messageBuffer.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))
	} else {
		// Simple HTML email without attachments
		messageBuffer.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		messageBuffer.WriteString(body)
	}

	logger.Info("MIME message preparation complete",
		"timeSpent", time.Since(startMimeTime),
		"messageSize", messageBuffer.Len())

	// Base64 URL encode the message for the Gmail API
	startEncodingTime := time.Now()
	encodedMessage := base64.URLEncoding.EncodeToString(messageBuffer.Bytes())
	logger.Info("Message encoding complete",
		"timeSpent", time.Since(startEncodingTime),
		"encodedSize", len(encodedMessage))

	// Create the request body
	reqBody := fmt.Sprintf(`{
        "raw": "%s"
    }`, encodedMessage)

	// Make the request to Gmail API
	startAPIRequestTime := time.Now()
	logger.Info("Preparing Gmail API request",
		"endpoint", "https://gmail.googleapis.com/gmail/v1/users/me/messages/send")

	req, err := http.NewRequest("POST", "https://gmail.googleapis.com/gmail/v1/users/me/messages/send", strings.NewReader(reqBody))
	if err != nil {
		logger.Error("Failed to create API request", "error", err)
		return fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Content-Type", "application/json")

	// Send the request
	logger.Info("Sending request to Gmail API")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("Failed to make API request",
			"error", err,
			"timeSpent", time.Since(startAPIRequestTime))
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
		logger.Error("Failed to read API response",
			"error", err,
			"httpStatus", resp.StatusCode)
		return fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("Gmail API returned error",
			"httpStatus", resp.StatusCode,
			"response", string(responseBody),
			"timeSpent", time.Since(startAPIRequestTime))
		return fmt.Errorf("error sending email via Gmail API: %s", responseBody)
	}

	logger.Info("Email delivery completed successfully",
		"recipients", to,
		"subject", subject,
		"httpStatus", resp.StatusCode,
		"timeSpent", time.Since(startAPIRequestTime),
		"totalTime", time.Since(startMimeTime))
	return nil
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
