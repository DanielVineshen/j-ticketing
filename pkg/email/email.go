// File: j-ticketing/pkg/email/email_service.go
package email

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"
	"io"
	"j-ticketing/pkg/config"
	logger "log/slog"
	"math/rand"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"github.com/boombuler/barcode"
	"github.com/boombuler/barcode/qr"
)

// EmailService is the interface for email operations
type EmailService interface {
	SendEmail(to []string, subject, body string) error
	SendPasswordResetEmail(to string, newPassword string) error
	SendTicketsEmail(to string, orderOverview OrderOverview, orderItems []OrderInfo, tickets []TicketInfo, attachments []Attachment) error
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
	Quatity      string
	OrderNumber  string
}

// OrderInfo represents information about a single order item
type OrderInfo struct {
	Description string
	Quantity    string
	Price       string
	EntryDate   string
}

// TicketInfo represents information for a single ticket
type TicketInfo struct {
	Label   string
	Content string
}

// SendTicketsEmail sends an email with QR codes for tickets
func (s *emailService) SendTicketsEmail(to string, orderOverview OrderOverview, orderItems []OrderInfo, tickets []TicketInfo, attachments []Attachment) error {
	subject := "Your " + orderOverview.TicketGroup + " Tickets"

	var address string
	var contactNo string
	var email string
	if orderOverview.TicketGroup == "Zoo Johor" {
		address = "Jalan Gertak Merah, Taman Istana<br>80000 Johor Bahru, Johor<br>General Line: +07-223 0404"
		contactNo = "+07-223 0404"
		email = "zoojohor@johor.gov.my"
	} else {
		address = "Taman Botani Diraja Johor Istana Besar Johor<br>80000 Johor Bahru, Johor<br>General Line: +07-485 8101"
		contactNo = "+07-485 8101"
		email = "botani.johor@gmail.com"
	}

	// Begin building HTML content
	var contentBuilder bytes.Buffer

	contentBuilder.WriteString(`
    <div style="padding: 20px 0px">
        <h4 style="font-size:16px">Dear ` + orderOverview.FullName + `,</h4>
        <p style="font-size:14px">Thank you for your purchase. Below are your tickets for ` + orderOverview.TicketGroup + `:</p>
	</div>
    `)

	contentBuilder.WriteString(`
<div class="order-info-section">
    <table width="100%" cellpadding="0" cellspacing="0" border="0" style="margin: 25px 0 30px;">
        <tr>
            <td width="49%" valign="top">
                <div class="info-group">
                    <div class="info-label">Lead participant</div>
                    <div class="info-value">` + orderOverview.FullName + `</div>
                </div>
            </td>
            <td width="49%" valign="top">
                <div class="info-group">
                    <div class="info-label">Purchase Date</div>
                    <div class="info-value">` + orderOverview.PurchaseDate + `</div>
                </div>
            </td>
        </tr>
        <tr>
			<td width="49%" valign="top" style="padding-top: 20px;">
                <div class="info-group">
                    <div class="info-label">Entry Date</div>
                    <div class="info-value">` + orderOverview.EntryDate + `</div>
                </div>
            </td>
            <td width="49%" valign="top" style="padding-top: 20px;">
                <div class="info-group">
                    <div class="info-label">Order No.</div>
                    <div class="info-value">` + orderOverview.OrderNumber + `</div>
                </div>
            </td>
        </tr>
    </table>
</div>
`)

	contentBuilder.WriteString(`
    <div>
        <div class="redeem-title">
            <h4>Redeem Individual Units</h4>
            <p>Scan the QR codes below to redeem your units individually.</p>
        </div>
	</div>
    `)

	// Start a table for QR codes - much better support in email clients
	contentBuilder.WriteString(`
        <table cellspacing="10" cellpadding="0" border="0" align="center" style="margin: 20px auto;">
        <tr>
    `)

	// Generate QR code for each ticket and add to email content
	ticketCount := len(tickets)
	maxColumns := 3 // Maximum number of QR codes per row

	for i, ticket := range tickets {
		// Generate QR code
		qrCode, err := qr.Encode(ticket.Content, qr.M, qr.Auto)
		//qrCode, err := qr.Encode("STF020", qr.M, qr.Auto) // HARDCODED VALUE FOR NOW
		if err != nil {
			return fmt.Errorf("failed to generate QR code: %w", err)
		}

		// Scale QR code to appropriate size
		qrCode, err = barcode.Scale(qrCode, 150, 150)
		if err != nil {
			return fmt.Errorf("failed to scale QR code: %w", err)
		}

		// Encode QR code as base64 string to embed in email
		var qrBuffer bytes.Buffer
		err = png.Encode(&qrBuffer, qrCode)
		if err != nil {
			return fmt.Errorf("failed to encode QR code as PNG: %w", err)
		}

		qrBase64 := base64.StdEncoding.EncodeToString(qrBuffer.Bytes())

		// Add ticket with QR code to email content using table cell
		contentBuilder.WriteString(`
            <td align="center" valign="top" style="padding: 5px; width: 160px;">
                <img src="data:image/png;base64,` + qrBase64 + `" alt="QR Code" style="width: 150px; height: 150px; border: 1px solid #eee; padding: 5px; margin-bottom: 8px;">
                <div style="font-size: 12px; font-weight: bold; color: #333; text-align: center; word-wrap: break-word; line-height: 1.2;">` + ticket.Label + `</div>
            </td>
        `)

		// Start a new row after maxColumns items or at the end
		if (i+1)%maxColumns == 0 && i < ticketCount-1 {
			contentBuilder.WriteString(`
			</tr>
			<tr>
			`)
		}
	}

	// Close the table row and table
	contentBuilder.WriteString(`
        </tr>
        </table>
    `)

	contentBuilder.WriteString(`
		<div style="text-align: center;">
			<p>Please show these QR codes at the entrance for scanning.</p>
        	<p>We hope you enjoy your visit to ` + orderOverview.TicketGroup + `!</p>
		</div>
        
        <div class="terms-section">
            <h4>Terma dan Syarat Perkhidmatan</h4>
            <p>Berikut adalah terma dan syarat penggunaan laman web ` + orderOverview.TicketGroup + ` bagi pembelian secara dalam talian. Sekiranya anda mengakses laman web ini dan menggunakan perkhidmatan yang ditawarkan, ia merupakan pengakuan dan persetujuan bahawa anda terikat kepada terma dan syarat sebagaimana berikut :</p>
            
            <p><strong>i) Pembelian Secara Dalam Talian</strong><br>
            Pembeli hendaklah memastikan tarikh, hari, jenis tiket dan kuantiti adalah betul sebelum mengklik butang bayaran.
            
			Bagi bayaran melalui kad kredit, kad debit atau perkhidmatan perbankan internet seperti Maybank2U atau lain-lain bank, anda hendaklah memastikan anda adalah pemilik akaun dan maklum mengenai pembayaran tersebut.</p>
            
            <p><strong>ii) Pengesahan Pembelian</strong><br>
            Selepas penerimaan pembayaran, anda akan menerima resit dan tiket yang tertera QR Code melalui emel yang telah didaftarkan. Sila bawa bersama resit dan tiket tersebut semasa berkunjung ke ` + orderOverview.TicketGroup + ` bagi mengelakkan sebarang permasalahan.</p>
            
            <p><strong>iii) Polisi Bayaran Balik</strong><br>
            Perkhidmatan pembelian tiket secara dalam talian ini beroperasi atas polisi tiada bayaran balik. Kesemua bayaran yang telah diterima tidak akan dibayar balik kepada pembeli kecuali di dalam keadaan tertentu yang akan ditentukan oleh pihak pengurusan antaranya permasalahan yang tidak dapat dielakkan seperti masalah teknikal laman web/sistem atau permasalahan berkaitan sistem perbankan.

			Proses bayaran balik adalah dalam tempoh 14 hari dari tarikh masalah dikenalpasti. Bagi situasi di mana pembeli telah terlebih membuat bayaran (sekiranya ada) , bayaran balik hanya akan dilaksanakan setelah bukti pembayaran dikemukakan kepada pihak pengurusan.</p>

			<p><strong>iv) Polisi Menukar Tarikh Tiket</strong><br>
            Perkhidmatan pembelian tiket secara dalam talian ini beroperasi atas polisi penukaran tarikh adalah tidak dibenarkan. Sekiranya pengunjung tidak dapat hadir pada tarikh yang telah dijadualkan, penukaran tarikh adalah tidak dibenarkan dan tiada pulangan bayaran akan dibuat.</p>

			<p><strong>v) Had Tanggungjawab</strong><br>
            Pihak pengurusan tidak menjamin bahawa fungsi yang terdapat di dalam laman web ini tidak akan terganggu atau bebas dari sebarang kesalahan. Pihak pengurusan juga tidak akan bertanggungjawab atas sebarang kerosakan, kemusnahan, gangguan perkhidmatan, kerugian, kehilangan simpanan atau kesan sampingan yang lain ketika mengoperasikan atau kegagalan mengoperasikan laman web ini, akses tanpa kebenaran, kenyataan atau tindakan pihak ketiga di laman web ini atau perkara-perjara lain yang berkaitan dengan laman web ini.</p>
        </div>
    </div>
    `)

	// Main content with customer information
	contentBuilder.WriteString(`
    <div>
        <div class="order-summary">
            <div class="section-title">
                <h3>Order Summary</h3>
            </div>
            <table class="order-table">
                <thead>
                    <tr>
                        <th>Item</th>
                        <th>Quantity</th>
                        <th>Price</th>
                        <th>Total</th>
                    </tr>
                </thead>
                <tbody>
    `)

	// Add order items to the table
	var subtotal float64

	for _, item := range orderItems {
		// Calculate item total
		price, _ := strconv.ParseFloat(item.Price, 64)
		qty, _ := strconv.Atoi(item.Quantity)
		total := price * float64(qty)
		subtotal += total

		contentBuilder.WriteString(`
                    <tr>
                        <td>` + item.Description + `<br><span class="item-date">` + item.EntryDate + `</span></td>
                        <td>` + item.Quantity + `</td>
                        <td>MYR ` + item.Price + `</td>
                        <td>MYR ` + fmt.Sprintf("%.2f", total) + `</td>
                    </tr>
        `)
	}

	// Add total row
	contentBuilder.WriteString(`
                </tbody>
                <tfoot>
                    <tr>
                        <td colspan="3">Subtotal</td>
                        <td>MYR ` + fmt.Sprintf("%.2f", subtotal) + `</td>
                    </tr>
                    <tr>
                        <td colspan="3">Total (Inclusive GST)</td>
                        <td>MYR ` + fmt.Sprintf("%.2f", subtotal) + `</td>
                    </tr>
                </tfoot>
            </table>
        </div>
    `)

	// Complete HTML email body with nice formatting
	body := fmt.Sprintf(`
<html>
<head>
    <style>
        body { 
            font-family: Arial, sans-serif; 
            margin: 0; 
            padding: 0; 
            color: #333; 
            background-color: #f9f9f9;
        }
        .container { 
            max-width: 650px; 
            margin: 0 auto; 
            background-color: #ffffff; 
            border-radius: 5px;
            box-shadow: 0 0 10px rgba(0,0,0,0.1);
        }
        .header { 
            background-color: #D5C58A; 
            color: #000000; 
            padding: 15px; 
            border-radius: 5px 5px 0 0; 
            text-align: center; 
        }
		.logo-container {
			display: table;
			width: 100%%;
			background-color: #ffffff;
			border-bottom: 1px solid #eee;
		}
		.logo {
			display: table-cell;
			width: 120px;
			vertical-align: middle;
		}
		.company-info {
			text-align: center;
			vertical-align: middle;
			font-size: 18px;
			color: #000000;
			margin: 0px 0px 0px 21px;
			padding: 0px;
		}
        .order-summary {
            margin: 15px 0;
            background-color: #f9f9f9;
            border-radius: 5px;
            padding: 15px;
        }
        .order-table {
            width: 100%%;
            border-collapse: collapse;
            margin: 15px 0;
        }
        .order-table th {
            background-color: #f0f0f0;
            text-align: left;
            padding: 8px;
            font-size: 14px;
            border-bottom: 1px solid #ddd;
        }
        .order-table td {
            padding: 8px;
            border-bottom: 1px solid #eee;
            font-size: 14px;
        }
        .order-table tfoot td {
            font-weight: bold;
            border-top: 2px solid #ddd;
        }
        .item-date {
            font-size: 12px;
            color: #666;
        }
        .order-meta {
            display: flex;
            flex-wrap: wrap;
            margin-top: 15px;
            font-size: 14px;
        }
        .meta-item {
            flex: 1;
            min-width: 200px;
            margin-bottom: 10px;
        }
        .meta-item span:first-child {
            font-weight: bold;
            color: #666;
        }
        .status-confirmed {
            color: #28a745;
            font-weight: bold;
        }
        .section-title {
            text-align: center;
        }
        .section-title h3 {
            margin: 0;
            color: #333;
            font-size: 18px;
        }
        .main-title {
            margin-top: 30px;
        }
        .redeem-title {
            margin-bottom: 20px;
        }
        .redeem-title h4 {
            font-size: 16px;
        }
        .redeem-title p {
            margin: 0;
            font-size: 14px;
            color: #666;
        }
        .order-summary {
            background-color: #f9f9f9;
            border-radius: 5px;
            padding: 15px;
        }
        .order-table {
            width: 100%%;
            border-collapse: collapse;
        }
        .order-table th {
            background-color: #f0f0f0;
            text-align: left;
            padding: 8px;
            font-size: 14px;
            border-bottom: 1px solid #ddd;
        }
        .order-table td {
            padding: 8px;
            border-bottom: 1px solid #eee;
            font-size: 14px;
        }
        .order-table tfoot td {
            font-weight: bold;
            border-top: 2px solid #ddd;
        }
        .terms-section {
            margin-top: 30px;
            padding: 15px;
            background-color: #f9f9f9;
            border-radius: 5px;
            font-size: 12px;
        }
        .terms-section h4 {
            margin-top: 0;
            border-bottom: 1px solid #ddd;
            padding-bottom: 5px;
        }
        .footer { 
            margin-top: 20px; 
            padding: 15px 0; 
            text-align: center; 
            font-size: 12px; 
            color: #888;
            border-top: 1px solid #eee;
        }
		.order-info-section {
			background-color: #f8f8f8;
			border-radius: 5px;
			padding: 20px;
		}
		.info-label {
			color: #666;
			font-size: 14px;
			margin-bottom: 5px;
		}
		.info-value {
			font-weight: bold;
			font-size: 16px;
			color: #333;
		}
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1 style="text-transform: uppercase;font-size: 32px;">`+orderOverview.TicketGroup+`</h1>
        	<div class="company-info">
				<p>`+address+`</p>
			</div>
        </div>
        
        %s
        
        <div class="footer">
            <p>This is an automated message, please do not reply to this email.</p>
            <p>Contact us: `+contactNo+` | `+email+`</p>
            <p>&copy; %d `+orderOverview.TicketGroup+`</p>
        </div>
    </div>
</body>
</html>
`, contentBuilder.String(), time.Now().Year())

	return s.sendEmailWithAttachmentsWithRetry([]string{to}, subject, body, attachments)
}

// NewEmailService creates a new email service
func NewEmailService(cfg *config.Config) EmailService {
	var tokenManager *OAuth2TokenManager
	useOAuth := false

	// Check if OAuth2 credentials are configured
	if cfg.Email.ClientID != "" && cfg.Email.ClientSecret != "" && cfg.Email.RefreshToken != "" {
		useOAuth = true

		// Log OAuth configuration (for debugging)
		logger.Info("Initializing OAuth2 email service")
		logger.Info("OAuth2 email username", "value", cfg.Email.Username)
		logger.Info("OAuth2 client ID", "length", len(cfg.Email.ClientID))
		logger.Info("OAuth2 client secret", "length", len(cfg.Email.ClientSecret))
		logger.Info("OAuth2 refresh token", "length", len(cfg.Email.RefreshToken))

		tokenManager = NewOAuth2TokenManager(OAuth2Config{
			ClientID:     cfg.Email.ClientID,
			ClientSecret: cfg.Email.ClientSecret,
			RefreshToken: cfg.Email.RefreshToken,
		})
	} else {
		logger.Warn("OAuth2 is NOT enabled - missing one or more required credentials")
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

	// Set content type based on whether there are attachments
	if len(attachments) > 0 {
		message.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n\r\n", boundary))
		message.WriteString(fmt.Sprintf("--%s\r\n", boundary))
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
		message.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", attachment.Name))
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

	// If we have attachments, create a multipart message
	if len(attachments) > 0 {
		logger.Info("Creating multipart MIME message",
			"attachmentCount", len(attachments))

		messageBuffer.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", boundary))
		messageBuffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		messageBuffer.WriteString("Content-Type: text/html; charset=UTF-8\r\n\r\n")
		messageBuffer.WriteString(body)
		messageBuffer.WriteString("\r\n")

		// Add each attachment
		for i, attachment := range attachments {
			logger.Info("Adding attachment to email",
				"index", i+1,
				"name", attachment.Name,
				"type", attachment.Type,
				"size", len(attachment.Content))

			messageBuffer.WriteString(fmt.Sprintf("\r\n--%s\r\n", boundary))
			messageBuffer.WriteString(fmt.Sprintf("Content-Type: %s\r\n", attachment.Type))
			messageBuffer.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", attachment.Name))
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

		// Close the MIME message
		messageBuffer.WriteString(fmt.Sprintf("\r\n--%s--\r\n", boundary))
	} else {
		// Simple HTML email without attachments
		logger.Info("Creating simple HTML email without attachments")
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
