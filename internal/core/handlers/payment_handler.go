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
	"log"
	"net/url"
	"strconv"
	"strings"
)

type PaymentHandler struct {
	paymentService     *services.PaymentService
	paymentConfig      payment.PaymentConfig
	emailService       email.EmailService
	ticketGroupService *services.TicketGroupService
}

// NewPaymentHandler creates a new instance of PaymentHandler
func NewPaymentHandler(paymentService *services.PaymentService, paymentConfig payment.PaymentConfig, emailService email.EmailService, ticketGroupService *services.TicketGroupService) *PaymentHandler {
	return &PaymentHandler{
		paymentService:     paymentService,
		paymentConfig:      paymentConfig,
		emailService:       emailService,
		ticketGroupService: ticketGroupService,
	}
}

func (h *PaymentHandler) PaymentReturn(c *fiber.Ctx) error {
	// Get the payload and remove surrounding quotes first
	payload := strings.Trim(c.Query("payload"), "\"")

	// Split the payload into IV and ciphertext components
	parts := strings.SplitN(payload, ":", 2)
	if len(parts) != 2 {
		fmt.Println("Invalid payload format, expected IV:ciphertext")
		return fmt.Errorf("invalid payload format: expected IV:ciphertext format")
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
		return err
	}

	// Decode ciphertext from base64
	cipherData, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		fmt.Println("Error decoding ciphertext:", err)
		return err
	}

	// Print diagnostic information
	fmt.Println("IV length:", len(iv), "bytes")
	fmt.Println("Key length:", len(key), "bytes")
	fmt.Println("Ciphertext length:", len(cipherData), "bytes")

	// Create a new AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println("Error creating cipher:", err)
		return err
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
		return jsonErr
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

	var dbStatus = order.TransactionStatus
	if order.TransactionStatus != "success" {
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

		var ticketInfos []email.TicketInfo
		var orderItems []email.OrderInfo
		// This would go after the database update but before the redirect
		if dbStatus == "success" {
			// Only call the Zoo API if payment was successful
			orderItems, ticketInfos, err = h.paymentService.PostToZooAPI(transactionData.OrderNo)
			if err != nil {
				log.Printf("Error posting to Johor Zoo API: %v", err)
				// Continue with redirect even if this fails, we can retry later
			}

			ticketGroup, err := h.ticketGroupService.GetTicketGroup(order.TicketGroupId)
			if err != nil {
				log.Printf("Error finding ticket group %s: %v", order.TicketGroupId, err)
			}

			orderOverview := email.OrderOverview{
				TicketGroup:  ticketGroup.GroupName,
				FullName:     order.BuyerName,
				PurchaseDate: order.TransactionDate,
				EntryDate:    orderItems[0].EntryDate,
				Quatity:      orderItems[0].Description,
				OrderNumber:  order.OrderNo,
			}

			err = h.emailService.SendTicketsEmail(order.BuyerEmail, orderOverview, orderItems, ticketInfos)
			if err != nil {
				log.Printf("Failed to send tickets email to %s: %v", order.BuyerEmail, err)
				// Continue anyway since the password has been reset
			}
			order.IsEmailSent = true
			// Save the updated order
			err = h.paymentService.UpdateOrderTicketGroup(order)
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
