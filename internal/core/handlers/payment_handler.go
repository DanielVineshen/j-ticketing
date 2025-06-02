package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"j-ticketing/internal/core/dto/payment"
	services "j-ticketing/internal/core/services"
	"j-ticketing/pkg/email"
	"j-ticketing/pkg/utils"
	"log"
	logger "log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type PaymentHandler struct {
	paymentService      *services.PaymentService
	paymentConfig       payment.PaymentConfig
	emailService        email.EmailService
	ticketGroupService  *services.TicketGroupService
	pdfService          *services.PDFService
	orderService        *services.OrderService
	customerService     *services.CustomerService
	notificationService *services.NotificationService
}

// NewPaymentHandler creates a new instance of PaymentHandler
func NewPaymentHandler(paymentService *services.PaymentService,
	paymentConfig payment.PaymentConfig,
	emailService email.EmailService,
	ticketGroupService *services.TicketGroupService,
	pdfService *services.PDFService,
	orderService *services.OrderService,
	customerService *services.CustomerService,
	notificationService *services.NotificationService) *PaymentHandler {
	return &PaymentHandler{
		paymentService:      paymentService,
		paymentConfig:       paymentConfig,
		emailService:        emailService,
		ticketGroupService:  ticketGroupService,
		pdfService:          pdfService,
		orderService:        orderService,
		customerService:     customerService,
		notificationService: notificationService,
	}
}

func (h *PaymentHandler) PaymentReturn(c *fiber.Ctx) error {
	transactionData, err := h.decipherPayload(c)
	if err != nil {
		return err
	}

	// Find the order first
	order, err := h.paymentService.FindByOrderNo(transactionData.OrderNo)
	if err != nil {
		log.Printf("Error finding order %s: %v", transactionData.OrderNo, err)
		return err
	}

	if order == nil {
		log.Printf("Order not found: %s", transactionData.OrderNo)
		return fmt.Errorf("order not found: %s", transactionData.OrderNo)
	}

	cust := order.Customer

	var dbStatus = order.TransactionStatus
	if dbStatus != "success" {
		// Update the order in the database
		err = h.paymentService.UpdateOrderFromPaymentResponse(transactionData.OrderNo, transactionData, order)
		if err != nil {
			log.Printf("Error updating order: %v", err)
			// Continue with the redirect even if the update fails
			// This ensures the user sees a response, and we can fix the data later if needed
		}

		// Extract and process payment status and other details
		status := transactionData.StatusTransaksi
		log.Printf("Payment status detected: %s", status)

		// Determine the transaction status for the database
		switch transactionData.StatusTransaksi {
		case "00":
			dbStatus = "success"
		case "AP", "09", "99":
			dbStatus = "pending"
		default:
			dbStatus = "failed"
		}

		// This would go after the database update but before the redirect
		if dbStatus == "success" {
			err = h.customerService.CreateCustomerLog("purchase", "Purchase Completed", "Ticket package purchased via Online", cust)
			if err != nil {
				return err
			}

			malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
			if err != nil {
				return err
			}
			message := fmt.Sprintf("%s (%s) has completed their purchase for order no: %s", cust.Email, cust.FullName, transactionData.OrderNo)
			err = h.notificationService.CreateNotification(
				cust.FullName,
				"customer",
				"Order",
				"Customer successful purchase order",
				message,
				malaysiaTime,
			)
			if err != nil {
				return err
			}

			// Only call the Zoo API if payment was successful
			_, orderItems, ticketInfos, err := h.paymentService.PostToZooAPI(transactionData.OrderNo)
			if err != nil {
				log.Printf("Error posting to Johor Zoo API: %v", err)
				// Continue with redirect even if this fails, we can retry later
			} else {
				err = h.orderService.CreateOrderTicketLog("order", "QR Code Assigned", "Order was assigned with qr codes for each ticket", "QR Service", order)
				if err != nil {
					return err
				}

				malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
				if err != nil {
					return err
				}
				message := fmt.Sprintf("%s has been assigned with qr codes for their tickets", transactionData.OrderNo)
				err = h.notificationService.CreateNotification(
					"system",
					"system",
					"Order",
					"Order assigned qr codes",
					message,
					malaysiaTime,
				)
				if err != nil {
					return err
				}
			}

			ticketGroup, err := h.ticketGroupService.GetTicketGroup(order.TicketGroupId)
			if err != nil {
				log.Printf("Error finding ticket group %s: %v", order.TicketGroupId, err)
			}

			total := utils.CalculateOrderTotal(orderItems)

			orderOverview := email.OrderOverview{
				TicketGroup:  ticketGroup.GroupNameBm,
				FullName:     order.BuyerName,
				PurchaseDate: order.TransactionDate,
				EntryDate:    orderItems[0].EntryDate,
				Quantity:     len(orderItems),
				OrderNumber:  order.OrderNo,
				Total:        total,
			}

			pdfBytes, pdfFilename, err := h.pdfService.GenerateTicketPDF(orderOverview, orderItems, ticketInfos)
			if err != nil {
				log.Printf("Error generating PDF: %v", err)
			}

			// Create attachment if PDF was successfully generated
			var pdfAttachment email.Attachment
			if err == nil && pdfBytes != nil {
				pdfAttachment = email.Attachment{
					Name:    pdfFilename,
					Content: pdfBytes,
					Type:    "application/pdf",
				}
			}

			err = h.emailService.SendTicketsEmail(order.BuyerEmail, orderOverview, orderItems, ticketInfos, []email.Attachment{pdfAttachment})
			if err != nil {
				log.Printf("Failed to send tickets email to %s: %v", order.BuyerEmail, err)
				// Continue anyway since the password has been reset
			} else {
				err = h.orderService.CreateOrderTicketLog("order", "Email Sent", "Email for the order was successfully sent out with its receipt", "Email Service", order)
				if err != nil {
					return err
				}

				malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
				if err != nil {
					return err
				}
				message := fmt.Sprintf("%s has sent out the email to the customer", transactionData.OrderNo)
				err = h.notificationService.CreateNotification(
					"system",
					"system",
					"Order",
					"Order email sent",
					message,
					malaysiaTime,
				)
				if err != nil {
					return err
				}

				order.IsEmailSent = true
				// Save the updated order
				err = h.paymentService.UpdateOrderTicketGroup(order)
			}
			if err != nil {
				log.Printf("Failed to update order ticket group: %v", err)
			}
		} else if dbStatus == "failed" {
			err = h.customerService.CreateCustomerLog("purchase", "Purchase Failed", "Ticket package failed via Online", cust)
			if err != nil {
				return err
			}

			malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
			if err != nil {
				return err
			}
			message := fmt.Sprintf("%s (%s) has failed to complete purchase for order no: %s", cust.Email, cust.FullName, transactionData.OrderNo)
			err = h.notificationService.CreateNotification(
				cust.FullName,
				"customer",
				"Order",
				"Customer failed purchase order",
				message,
				malaysiaTime,
			)
			if err != nil {
				return err
			}
		}
	}

	successURL := h.paymentConfig.FrontendBaseURL + "/paymentRedirect"

	log.Printf("Redirecting to external success page: %s", successURL)

	// Build the full URL with query parameters
	redirectURL := fmt.Sprintf("%s?orderTicketGroupId=%s&transactionStatus=%s&orderNo=%s",
		successURL,
		url.QueryEscape(strconv.Itoa(int(order.OrderTicketGroupId))),
		url.QueryEscape(dbStatus),
		url.QueryEscape(transactionData.OrderNo))

	log.Printf("Complete redirect URL: %s", redirectURL)

	return c.Redirect(redirectURL)
}

func (h *PaymentHandler) PaymentRedirect(c *fiber.Ctx) error {
	logger.Info("Payment redirect was triggered by JohorPay!")

	log.Printf("============ PAYMENT REDIRECT RECEIVED ============")
	log.Printf("Time: %s", time.Now().Format(time.RFC3339))

	// Log request method and URL
	log.Printf("Method: %s", c.Method())
	log.Printf("Path: %s", c.Path())
	log.Printf("Full URL: %s", c.BaseURL()+c.OriginalURL())

	// Log all HTTP headers
	log.Printf("------ Headers ------")
	c.Request().Header.VisitAll(func(key, value []byte) {
		log.Printf("%s: %s", string(key), string(value))
	})

	// Log all query parameters
	log.Printf("------ Query Parameters ------")
	queryParams := c.Queries()
	for k, v := range queryParams {
		log.Printf("%s: %s", k, v)
	}

	// Enhanced body logging
	log.Printf("------ Request Body ------")

	// Get raw body bytes
	body := c.Body()
	log.Printf("Body length: %d bytes", len(body))

	if len(body) > 0 {
		// Log raw body as string
		bodyStr := string(body)
		log.Printf("Raw body: %s", bodyStr)

		// Log body as hex dump for binary inspection
		log.Printf("Body hex dump:")
		for i := 0; i < len(body); i += 16 {
			end := i + 16
			if end > len(body) {
				end = len(body)
			}
			chunk := body[i:end]
			hexStr := ""
			asciiStr := ""

			for _, b := range chunk {
				hexStr += fmt.Sprintf("%02x ", b)
				if b >= 32 && b <= 126 { // Printable ASCII
					asciiStr += string(b)
				} else {
					asciiStr += "."
				}
			}

			// Pad for alignment if needed
			for j := len(chunk); j < 16; j++ {
				hexStr += "   "
			}

			log.Printf("  %04x: %s | %s", i, hexStr, asciiStr)
		}

		// Try to parse as URL-encoded form data
		contentType := c.Get("Content-Type")
		if strings.Contains(contentType, "application/x-www-form-urlencoded") {
			log.Printf("Detected form-urlencoded content type")

			// Parse as form data
			var form map[string][]string
			if err := c.QueryParser(&form); err == nil {
				log.Printf("Parsed form data:")
				for k, v := range form {
					log.Printf("  %s: %v", k, v)
				}
			} else {
				log.Printf("Failed to parse form data: %v", err)

				// Try manual parsing
				formItems := strings.Split(bodyStr, "&")
				log.Printf("Manual form parsing (%d items):", len(formItems))
				for _, item := range formItems {
					parts := strings.SplitN(item, "=", 2)
					if len(parts) == 2 {
						key, _ := url.QueryUnescape(parts[0])
						value, _ := url.QueryUnescape(parts[1])
						log.Printf("  %s: %s", key, value)
					} else if len(parts) == 1 {
						key, _ := url.QueryUnescape(parts[0])
						log.Printf("  %s: (no value)", key)
					}
				}
			}
		}

		// Try to parse as JSON
		if strings.Contains(contentType, "application/json") ||
			(bodyStr != "" && bodyStr[0] == '{' && bodyStr[len(bodyStr)-1] == '}') {
			log.Printf("Attempting to parse as JSON")
			var jsonData map[string]interface{}
			if err := json.Unmarshal(body, &jsonData); err == nil {
				jsonBytes, _ := json.MarshalIndent(jsonData, "", "  ")
				log.Printf("Parsed JSON data:\n%s", string(jsonBytes))
			} else {
				log.Printf("Failed to parse as JSON: %v", err)
			}
		}

		// Look for specific patterns in the body
		if strings.Contains(bodyStr, ":") && strings.Contains(bodyStr, "==") {
			log.Printf("Body appears to contain encoded/encrypted data (contains ':' and '==')")

			// Split and log parts separately
			parts := strings.Split(bodyStr, ":")
			log.Printf("Found %d parts separated by ':'", len(parts))
			for i, part := range parts {
				log.Printf("Part %d: %s", i+1, part)

				// Check if this part looks like Base64
				if strings.Contains(part, "==") || strings.Contains(part, "=") {
					log.Printf("Part %d appears to be Base64 encoded", i+1)
				}
			}
		}
	} else {
		log.Printf("Body is empty")
	}

	// Log any cookies
	log.Printf("------ Cookies ------")
	cookieHeader := c.Get("Cookie")
	log.Printf("Cookie header: %s", cookieHeader)

	// Log any session data
	log.Printf("------ Session/Store Data ------")

	// Log IP address and user agent
	log.Printf("------ Client Info ------")
	log.Printf("IP: %s", c.IP())
	log.Printf("User-Agent: %s", c.Get("User-Agent"))

	log.Printf("============ END PAYMENT REDIRECT LOG ============")

	transactionData, err := h.decipherPayload(c)
	if err != nil {
		return err
	}

	logger.Info("With transaction data: %v", transactionData)

	// Find the order first
	order, err := h.paymentService.FindByOrderNo(transactionData.OrderNo)
	if err != nil {
		log.Printf("Error finding order %s: %v", transactionData.OrderNo, err)
		return err
	}

	if order == nil {
		log.Printf("Order not found: %s", transactionData.OrderNo)
		return fmt.Errorf("order not found: %s", transactionData.OrderNo)
	}

	cust := order.Customer

	var dbStatus = order.TransactionStatus
	if dbStatus != "success" {
		// Update the order in the database
		err = h.paymentService.UpdateOrderFromPaymentResponse(transactionData.OrderNo, transactionData, order)
		if err != nil {
			log.Printf("Error updating order: %v", err)
			// Continue with the redirect even if the update fails
			// This ensures the user sees a response, and we can fix the data later if needed
		}

		// Extract and process payment status and other details
		status := transactionData.StatusTransaksi
		log.Printf("Payment status detected: %s", status)

		// Determine the transaction status for the database
		switch transactionData.StatusTransaksi {
		case "00":
			dbStatus = "success"
		case "AP", "09", "99":
			dbStatus = "pending"
		default:
			dbStatus = "failed"
		}

		// This would go after the database update but before the redirect
		if dbStatus == "success" {
			err = h.customerService.CreateCustomerLog("purchase", "Purchase Completed", "Ticket package purchased via Online", cust)
			if err != nil {
				return err
			}

			malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
			if err != nil {
				return err
			}
			message := fmt.Sprintf("%s (%s) has completed their purchase for order no: %s", cust.Email, cust.FullName, transactionData.OrderNo)
			err = h.notificationService.CreateNotification(
				cust.FullName,
				"customer",
				"Order",
				"Customer successful purchase order",
				message,
				malaysiaTime,
			)
			if err != nil {
				return err
			}

			// Only call the Zoo API if payment was successful
			orderTicketGroup, orderItems, ticketInfos, err := h.paymentService.PostToZooAPI(transactionData.OrderNo)
			if err != nil {
				log.Printf("Error posting to Johor Zoo API: %v", err)
				// Continue with redirect even if this fails, we can retry later
			} else {
				err = h.orderService.CreateOrderTicketLog("order", "QR Code Assigned", "Order was assigned with qr codes for each ticket", "QR Service", orderTicketGroup)
				if err != nil {
					return err
				}

				malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
				if err != nil {
					return err
				}
				message := fmt.Sprintf("%s has been assigned with qr codes for their tickets", transactionData.OrderNo)
				err = h.notificationService.CreateNotification(
					"system",
					"system",
					"Order",
					"Order assigned qr codes",
					message,
					malaysiaTime,
				)
				if err != nil {
					return err
				}
			}

			ticketGroup, err := h.ticketGroupService.GetTicketGroup(order.TicketGroupId)
			if err != nil {
				log.Printf("Error finding ticket group %s: %v", order.TicketGroupId, err)
			}

			total := utils.CalculateOrderTotal(orderItems)

			orderOverview := email.OrderOverview{
				TicketGroup:  ticketGroup.GroupNameBm,
				FullName:     order.BuyerName,
				PurchaseDate: order.TransactionDate,
				EntryDate:    orderItems[0].EntryDate,
				Quantity:     len(orderItems),
				OrderNumber:  order.OrderNo,
				Total:        total,
			}

			pdfBytes, pdfFilename, err := h.pdfService.GenerateTicketPDF(orderOverview, orderItems, ticketInfos)
			if err != nil {
				log.Printf("Error generating PDF: %v", err)
			}

			// Create attachment if PDF was successfully generated
			var pdfAttachment email.Attachment
			if err == nil && pdfBytes != nil {
				pdfAttachment = email.Attachment{
					Name:    pdfFilename,
					Content: pdfBytes,
					Type:    "application/pdf",
				}
			}

			err = h.emailService.SendTicketsEmail(order.BuyerEmail, orderOverview, orderItems, ticketInfos, []email.Attachment{pdfAttachment})
			if err != nil {
				log.Printf("Failed to send tickets email to %s: %v", order.BuyerEmail, err)
				// Continue anyway since the password has been reset
			} else {
				err = h.orderService.CreateOrderTicketLog("order", "Email Sent", "Email for the order was successfully sent out with its receipt", "Email Service", orderTicketGroup)
				if err != nil {
					return err
				}

				malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
				if err != nil {
					return err
				}
				message := fmt.Sprintf("%s has sent out the email to the customer", transactionData.OrderNo)
				err = h.notificationService.CreateNotification(
					"system",
					"system",
					"Order",
					"Order email sent",
					message,
					malaysiaTime,
				)
				if err != nil {
					return err
				}

				order.IsEmailSent = true
				// Save the updated order
				err = h.paymentService.UpdateOrderTicketGroup(order)
			}
			if err != nil {
				log.Printf("Failed to update order ticket group: %v", err)
			}
		} else if dbStatus == "failed" {
			err = h.customerService.CreateCustomerLog("purchase", "Purchase Failed", "Ticket package failed via Online", cust)
			if err != nil {
				return err
			}

			malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
			if err != nil {
				return err
			}
			message := fmt.Sprintf("%s (%s) has failed to complete purchase for order no: %s", cust.Email, cust.FullName, transactionData.OrderNo)
			err = h.notificationService.CreateNotification(
				cust.FullName,
				"customer",
				"Order",
				"Customer failed purchase order",
				message,
				malaysiaTime,
			)
			if err != nil {
				return err
			}
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": "Payment processed failed",
			})
		}
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "Payment processed successfully",
	})
}

func (h *PaymentHandler) decipherPayload(c *fiber.Ctx) (payment.TransactionResponse, error) {
	// Get query parameters
	queryParams := c.Queries()
	logger.Info("queryParams from decipherPayload: %v", queryParams)

	// Get the payload and remove surrounding quotes first
	payload := strings.Trim(c.Query("payload"), "\"")

	// Split the payload into IV and ciphertext components
	parts := strings.SplitN(payload, ":", 2)
	if len(parts) != 2 {
		fmt.Println("Invalid payload format, expected IV:ciphertext")
		return payment.TransactionResponse{}, fmt.Errorf("invalid payload format: expected IV:ciphertext format")
	}

	// Clean up the components
	ivBase64 := strings.ReplaceAll(parts[0], "\\/", "/")
	ivBase64 = strings.ReplaceAll(ivBase64, " ", "+")
	cipherText := strings.ReplaceAll(parts[1], " ", "+")
	cipherText = strings.ReplaceAll(cipherText, "\\/", "/")

	log.Printf("ivBase64: %s", ivBase64)
	log.Printf("cipherText: %s", cipherText)

	hexKey := h.paymentConfig.APIKey

	// Take the first 32 characters of the key
	key := []byte(hexKey[:32])

	// Decode IV from base64
	iv, err := base64.StdEncoding.DecodeString(ivBase64)
	if err != nil {
		fmt.Println("Error decoding IV:", err)
		return payment.TransactionResponse{}, err
	}

	// Decode ciphertext from base64
	cipherData, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		fmt.Println("Error decoding ciphertext:", err)
		return payment.TransactionResponse{}, err
	}

	// Print diagnostic information
	fmt.Println("IV length:", len(iv), "bytes")
	fmt.Println("Key length:", len(key), "bytes")
	fmt.Println("Ciphertext length:", len(cipherData), "bytes")

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("Error creating cipher:", err)
		return payment.TransactionResponse{}, err
	}

	// Create decrypter in CBC mode
	mode := cipher.NewCBCDecrypter(block, iv)

	// Create a buffer for decryption
	plaintext := make([]byte, len(cipherData))

	// Decrypt
	mode.CryptBlocks(plaintext, cipherData)

	// Apply PKCS7 padding removal
	paddingLen := int(plaintext[len(plaintext)-1])
	if paddingLen > 0 && paddingLen <= aes.BlockSize {
		plaintext = plaintext[:len(plaintext)-paddingLen]
	}

	// Convert to string and print
	result := string(plaintext)
	fmt.Println("\nDecrypted result:")
	fmt.Println(result)

	// Then unmarshal the JSON into this struct
	var transactionData payment.TransactionResponse
	jsonErr := json.Unmarshal([]byte(result), &transactionData)
	if jsonErr != nil {
		// Handle error
		fmt.Println("Error parsing JSON:", jsonErr)
		return payment.TransactionResponse{}, err
	}
	return transactionData, nil
}
