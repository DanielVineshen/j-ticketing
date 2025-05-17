package routes

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"io"
	"j-ticketing/internal/core/dto/payment"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type TransactionResponse struct {
	IDTransaksi     string `json:"id_transaksi"`
	OrderNo         string `json:"order_no"`
	StatusTransaksi string `json:"status_transaksi"`
	StatusMessage   string `json:"status_message"`
	TarikhTransaksi string `json:"tarikh_transaksi"`
	KodBank         string `json:"kod_bank"`
	NamaBank        string `json:"nama_bank"`
	JpMsgToken      string `json:"jp_msg_token"`
}

// First, define the request and response structures for the Johor Zoo API
type ZooTicketItem struct {
	ItemId string `json:"ItemId"`
	Qty    int    `json:"Qty"`
}

type ZooTicketRequest struct {
	TranDate    string          `json:"TranDate"`
	ReferenceNo string          `json:"ReferenceNo"`
	Items       []ZooTicketItem `json:"Items"`
}

type ZooTicketInfo struct {
	TWBID       string `json:"TWBID"`
	ItemId      string `json:"ItemId"`
	EncryptedID string `json:"EncryptedID"`
	AdmitDate   string `json:"AdmitDate"`
	UnitPrice   string `json:"UnitPrice"`
	ItemDesc    string `json:"ItemDesc"`
	ItemDesc2   string `json:"ItemDesc2"`
	ItemDesc3   string `json:"ItemDesc3"`
}

type ZooTicketResponse struct {
	StatusCode    string          `json:"StatusCode"`
	ReceiptNumber string          `json:"ReceiptNumber"`
	Tickets       []ZooTicketInfo `json:"Tickets"`
}

func SetupPaymentRoutes(app *fiber.App, paymentConfig payment.PaymentConfig, orderTicketGroupRepo *repositories.OrderTicketGroupRepository, orderTicketInfoRepo *repositories.OrderTicketInfoRepository) {
	app.Post("/decrypt", func(c *fiber.Ctx) error {
		// Original combined payload (IV:ciphertext)
		payload := `5\/4g3kU5e3TeIHRBuODwaQ==:m o5UiCWiJedyfDIxY8IrF49tfd0qejW9Iv\/5XGKQZ7BZP4ahvwIO5zDxg0nEXL x HEsuhscS7g5t2T2Ip\/4xd5bJzmbMsHJsK29Qo224Fohzf9itYxvD8njnshKi1GcBEQNQbX1  F1VTzAskn84ARSI QWM Qepcerg59quUGL17xYGLo3hoKhUFnXFclcdCsL9iv19riJXpQ65n\/ 2ZvjXfbPv fUE4lRIYtP58qh9ABUUxSPUCNoyPp\/ CfuEVyNqvG T3fZeRB86AD1ujDtP4SAx\/cOLYrELKgqaE=`

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

		// Use the raw key directly from the PHP example
		hexKey := paymentConfig.APIKey

		// Take the first 32 characters of the key
		// This is what worked in our previous test
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

		// Try to decode as JSON
		var jsonData interface{}
		jsonErr := json.Unmarshal([]byte(result), &jsonData)

		return c.JSON(fiber.Map{
			"success":     true,
			"data":        result,
			"json":        jsonData,
			"isValidJson": jsonErr == nil,
		})
	})

	// Payment return handler (callback from payment gateway)
	app.Get("/payment/return", func(c *fiber.Ctx) error {
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

		hexKey := paymentConfig.APIKey

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
		var transactionData TransactionResponse
		jsonErr := json.Unmarshal([]byte(result), &transactionData)
		if jsonErr != nil {
			// Handle error
			fmt.Println("Error parsing JSON:", jsonErr)
			return jsonErr
		}

		// Update the order in the database
		err = UpdateOrderFromPaymentResponse(transactionData.OrderNo, transactionData, *orderTicketGroupRepo)
		if err != nil {
			log.Printf("Error updating order: %v", err)
			// Continue with the redirect even if the update fails
			// This ensures the user sees a response, and we can fix the data later if needed
		}

		// Extract and process payment status and other details
		status := transactionData.StatusTransaksi
		log.Printf("Payment status detected: %s", status)

		// Find the order first
		order, err := orderTicketGroupRepo.FindByOrderNo(transactionData.OrderNo)
		if err != nil {
			log.Printf("Error finding order %s: %v", transactionData.OrderNo, err)
			return err
		}

		if order == nil {
			log.Printf("Order not found: %s", transactionData.OrderNo)
			return fmt.Errorf("order not found: %s", transactionData.OrderNo)
		}

		// Determine the transaction status for the database
		var dbStatus string
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
			// Only call the Zoo API if payment was successful
			err = PostToZooAPI(order, transactionData.OrderNo, *orderTicketInfoRepo)
			if err != nil {
				log.Printf("Error posting to Johor Zoo API: %v", err)
				// Continue with redirect even if this fails, we can retry later
			}
		}

		successURL := paymentConfig.FrontendBaseURL + "/paymentRedirect"

		log.Printf("Redirecting to external success page: %s", successURL)

		// Build the full URL with query parameters
		redirectURL := fmt.Sprintf("%s?orderTicketGroupId=%s&transactionStatus=%s&orderNo=%s",
			successURL,
			url.QueryEscape(strconv.Itoa(int(order.OrderTicketGroupId))),
			url.QueryEscape(dbStatus),
			url.QueryEscape(transactionData.OrderNo))

		return c.Redirect(redirectURL)
	})

	// Payment process - this will redirect to the payment gateway
	app.Post("/payment/process", func(c *fiber.Ctx) error {
		randomStr, err := GenerateRandom16()
		if err != nil {
			// handle error
		}
		log.Printf("orderNo: %s", randomStr)

		token := c.FormValue("token")
		orderNo := c.FormValue("orderNo")
		billId := c.FormValue("billId")
		productId := c.FormValue("productId")
		buyerName := c.FormValue("buyerName")
		buyerEmail := c.FormValue("buyerEmail")
		totalAmount := c.FormValue("totalAmount")
		productDesc := c.FormValue("productDesc")
		msgToken := c.FormValue("msgToken")
		bankCode := c.FormValue("bankCode")

		agToken := paymentConfig.AGToken
		method := "getRedirectUrl"
		redirectUrl := paymentConfig.BaseURL + "/payment/return"

		// Calculate the jp_checksum as described
		// Concatenate the values in the required order: buyerName + agToken + orderNo + totalAmount
		concatenatedString := buyerName + agToken + orderNo + totalAmount
		log.Printf("Concatenated String for checksum: %s", concatenatedString)

		// Generate SHA-512 hash
		hasher := sha512.New()
		hasher.Write([]byte(concatenatedString))
		checksum := hex.EncodeToString(hasher.Sum(nil))
		log.Printf("Generated Checksum: %s", checksum)

		// Create form data for x-www-form-urlencoded request
		formData := url.Values{}
		formData.Set("jp_buyer_name", buyerName)
		if bankCode != "" {
			formData.Set("jp_bank_code", bankCode)
		}
		formData.Set("jp_token", token)
		formData.Set("jp_ag_token", agToken)
		formData.Set("bill_id", billId)
		formData.Set("jp_order_no", orderNo)
		if msgToken != "" {
			formData.Set("jp_msg_token", msgToken)
		}
		formData.Set("jp_total_amount", totalAmount)
		formData.Set("jp_product_id", productId)
		formData.Set("jp_product_desc", productDesc)
		formData.Set("jp_email", buyerEmail)
		formData.Set("method", method)
		formData.Set("jp_redirect_url", redirectUrl)
		formData.Set("jp_checksum", checksum)

		if bankCode != "" && msgToken != "" {
			formData.Set("jp_gateway", "2")
		} else {
			formData.Set("jp_gateway", "1963")
		}

		// Prepare the request URL - ensure it's using HTTPS
		apiURL := paymentConfig.GatewayURL

		// Log the request parameters for debugging
		log.Printf("Making request to: %s", apiURL)
		log.Printf("Payment request parameters: %v", formData)

		// Make the request directly in the browser - create an HTML form that auto-submits
		formHTML := `
			<!DOCTYPE html>
			<html>
			<head>
				<title>Redirecting to Payment Gateway</title>
				<style>
					body {
						font-family: Arial, sans-serif;
						text-align: center;
						margin-top: 50px;
					}
					.loader {
						border: 6px solid #f3f3f3;
						border-top: 6px solid #3498db;
						border-radius: 50%;
						width: 50px;
						height: 50px;
						animation: spin 2s linear infinite;
						margin: 20px auto;
					}
					@keyframes spin {
						0% { transform: rotate(0deg); }
						100% { transform: rotate(360deg); }
					}
				</style>
			</head>
			<body>
				<h2>Redirecting to Payment Gateway</h2>
				<div class="loader"></div>
				<p>Please wait, you will be redirected automatically...</p>
				
				<form id="paymentForm" action="` + apiURL + `" method="post">
			`

		// Add all the form fields
		for key, values := range formData {
			for _, value := range values {
				formHTML += `        <input type="hidden" name="` + key + `" value="` + value + `">
`
			}
		}

		formHTML += `
				</form>

				<script>
					// Auto-submit the form when the page loads
					document.addEventListener('DOMContentLoaded', function() {
						document.getElementById('paymentForm').submit();
					});
				</script>
			</body>
			</html>`

		// Return the form page
		c.Set("Content-Type", "text/html")
		return c.Status(200).SendString(formHTML)
	})

	// Payment success
	app.Get("/payment/success", func(c *fiber.Ctx) error {
		orderID := c.Query("order_id")
		transactionID := c.Query("transaction_id")

		return c.Render("success", fiber.Map{
			"Title":         "Payment Successful",
			"OrderID":       orderID,
			"TransactionID": transactionID,
		})
	})

	// Payment failure
	app.Get("/payment/failure", func(c *fiber.Ctx) error {
		errorCode := c.Query("error_code")
		errorMessage := c.Query("error_message")

		return c.Render("failure", fiber.Map{
			"Title":        "Payment Failed",
			"ErrorCode":    errorCode,
			"ErrorMessage": errorMessage,
		})
	})

	// API endpoint to generate a token
	app.Post("/api/generate-token", func(c *fiber.Ctx) error {
		// Get the API key from config
		apiKey := paymentConfig.APIKey

		// Create form data for x-www-form-urlencoded request
		formData := url.Values{}
		formData.Set("jp_ag_token", "ZOO")
		formData.Set("method", "getRedirectUrl")
		formData.Set("jp_gateway", "2")

		// Create a new HTTP client
		client := &http.Client{
			Timeout: time.Second * 30,
		}

		// Create a new request
		req, err := http.NewRequest("POST", "https://johorpay-stag.johor.gov.my/JP_gateway/redflow", strings.NewReader(formData.Encode()))
		if err != nil {
			log.Printf("Error creating request: %v", err)
			return c.JSON(fiber.Map{
				"success": false,
				"message": "Failed to create request",
			})
		}

		// Add headers
		req.Header.Add("jp-api-key", apiKey)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		// Execute the request
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error executing request: %v", err)
			return c.JSON(fiber.Map{
				"success": false,
				"message": "Failed to connect to token service",
			})
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			return c.JSON(fiber.Map{
				"success": false,
				"message": "Failed to read response",
			})
		}

		// Parse the JSON response
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			log.Printf("Error parsing JSON: %v", err)
			return c.JSON(fiber.Map{
				"success": false,
				"message": "Failed to parse response",
			})
		}

		// Check if the token generation was successful
		if success, ok := result["success"].(bool); ok && success {
			if responseMsg, ok := result["response_msg"].(map[string]interface{}); ok {
				if randKey, ok := responseMsg["rand_key"].(string); ok {
					// Return the token
					return c.JSON(fiber.Map{
						"success": true,
						"token":   randKey,
					})
				}
			}
		}

		// If we got here, something went wrong
		log.Printf("Failed to get token from response: %v", string(body))
		return c.JSON(fiber.Map{
			"success": false,
			"message": "Failed to extract token from response",
		})
	})

	// API endpoint to get bank list
	app.Post("/api/bank-list", func(c *fiber.Ctx) error {
		// Parse request body
		var request struct {
			Mode string `json:"mode"`
		}

		if err := c.BodyParser(&request); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"success": false,
				"message": "Invalid request",
			})
		}

		// Determine the mode value
		var mode string
		if request.Mode == "individual" || request.Mode == "01" {
			mode = "01"
		} else {
			mode = "02"
		}

		// Get the API key from config
		apiKey := paymentConfig.APIKey

		// Create form data for x-www-form-urlencoded request
		formData := url.Values{}
		formData.Set("jp_ag_token", "ZOO")
		formData.Set("method", "getBankList")
		formData.Set("mode", mode)

		// Create a new HTTP client
		client := &http.Client{
			Timeout: time.Second * 30,
		}

		// Create a new request
		req, err := http.NewRequest("POST", "https://johorpay-stag.johor.gov.my/JP_gateway/getBankList", strings.NewReader(formData.Encode()))
		if err != nil {
			log.Printf("Error creating request: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Failed to create request",
			})
		}

		// Add headers
		req.Header.Add("jp-api-key", apiKey)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		// Execute the request
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("Error executing request: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Failed to connect to bank list service",
			})
		}
		defer resp.Body.Close()

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Error reading response body: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Failed to read response",
			})
		}

		// Parse the JSON response
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			log.Printf("Error parsing JSON: %v", err)
			return c.Status(500).JSON(fiber.Map{
				"success": false,
				"message": "Failed to parse response",
			})
		}

		// Check if the request was successful
		if success, ok := result["success"].(bool); ok && success {
			// The response has a data field that contains a JSON string (not an object)
			// We need to parse this string into an array of bank objects
			if dataStr, ok := result["data"].(string); ok {
				var banks []map[string]interface{}
				if err := json.Unmarshal([]byte(dataStr), &banks); err != nil {
					log.Printf("Error parsing bank data: %v", err)
					return c.Status(500).JSON(fiber.Map{
						"success": false,
						"message": "Failed to parse bank data",
					})
				}

				// Return the parsed banks
				return c.JSON(fiber.Map{
					"success": true,
					"banks":   banks,
				})
			}
		}

		// If we got here, something went wrong
		log.Printf("Failed to get bank list from response: %v", string(body))
		return c.Status(500).JSON(fiber.Map{
			"success": false,
			"message": "Failed to retrieve bank list",
		})
	})
}

// GenerateRandom16 generates a cryptographically secure random string of 16 characters
func GenerateRandom16() (string, error) {
	// We need 12 bytes to get 16 characters in base64
	randomBytes := make([]byte, 12)

	// Read random bytes using crypto/rand
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode to base64 and trim to exactly 16 characters
	randomString := base64.URLEncoding.EncodeToString(randomBytes)[:16]

	return randomString, nil
}

func UpdateOrderFromPaymentResponse(orderNo string, transactionData TransactionResponse,
	orderTicketGroupRepo repositories.OrderTicketGroupRepository) error {

	// Find the order first
	order, err := orderTicketGroupRepo.FindByOrderNo(orderNo)
	if err != nil {
		log.Printf("Error finding order %s: %v", orderNo, err)
		return err
	}

	if order == nil {
		log.Printf("Order not found: %s", orderNo)
		return fmt.Errorf("order not found: %s", orderNo)
	}

	// Determine the transaction status for the database
	var dbStatus string
	switch transactionData.StatusTransaksi {
	case "00":
		dbStatus = "success"
	case "AP", "09", "99":
		dbStatus = "pending"
	default:
		dbStatus = "failed"
	}

	// Update order fields
	order.TransactionId = transactionData.IDTransaksi
	order.TransactionStatus = dbStatus
	//order.TransactionDate = transactionData.TarikhTransaksi
	order.StatusMessage = sql.NullString{String: transactionData.StatusMessage, Valid: transactionData.StatusMessage != ""}
	order.UpdatedAt = time.Now()

	// Save the updated order
	err = orderTicketGroupRepo.Update(order)
	if err != nil {
		log.Printf("Error updating order: %v", err)
		return err
	}

	log.Printf("Successfully updated order %s with transaction ID %s and status %s",
		orderNo, transactionData.IDTransaksi, dbStatus)

	return nil
}

// Define the function to post to the Zoo API
func PostToZooAPI(order *models.OrderTicketGroup, orderNo string, orderTicketInfoRepo repositories.OrderTicketInfoRepository) error {
	// Get the order ticket items
	orderTickets, err := orderTicketInfoRepo.FindByOrderTicketGroupID(order.OrderTicketGroupId)
	if err != nil {
		return fmt.Errorf("failed to get order tickets: %w", err)
	}

	// Build the request
	items := make([]ZooTicketItem, 0, len(orderTickets))
	for _, ticket := range orderTickets {
		items = append(items, ZooTicketItem{
			ItemId: ticket.ItemId,
			Qty:    ticket.QuantityBought,
		})
	}

	// Format the admission date from the order
	admissionDate := order.TransactionDate[:10]

	// Create the request payload
	payload := ZooTicketRequest{
		TranDate:    admissionDate,
		ReferenceNo: orderNo, // Use the order number as reference
		Items:       items,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Get a fresh token from the token generation endpoint
	token, err := generateZooAPIToken()
	if err != nil {
		return fmt.Errorf("failed to generate API token: %w", err)
	}

	// Create a new HTTP client
	client := &http.Client{
		Timeout: time.Second * 60,
	}

	// Create the request
	req, err := http.NewRequest("POST", "https://eglobal2.ddns.net/johorzooapi/api/JohorZoo/PostOnlinePurchase", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Check if the response status is OK
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned non-OK status: %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse the response
	var zooResponse ZooTicketResponse
	err = json.Unmarshal(body, &zooResponse)
	if err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if the status code is OK
	if zooResponse.StatusCode != "OK" {
		return fmt.Errorf("API returned status code: %s", zooResponse.StatusCode)
	}

	// Update each ticket with the data from the API
	// Create a map of tickets by item ID for quick lookup
	ticketByItemId := make(map[string][]*models.OrderTicketInfo)
	for i := range orderTickets {
		ticketByItemId[orderTickets[i].ItemId] = append(ticketByItemId[orderTickets[i].ItemId], &orderTickets[i])
	}

	// Update the tickets with data from the response
	for _, zooTicket := range zooResponse.Tickets {
		// Get the list of tickets for this item ID
		ticketsForItem, exists := ticketByItemId[zooTicket.ItemId]
		if !exists || len(ticketsForItem) == 0 {
			log.Printf("No matching tickets found for item ID: %s", zooTicket.ItemId)
			continue
		}

		// Get the next ticket that hasn't been updated yet
		var ticketToUpdate *models.OrderTicketInfo
		for _, t := range ticketsForItem {
			if t.EncryptedId == "" {
				ticketToUpdate = t
				break
			}
		}

		if ticketToUpdate == nil {
			log.Printf("All tickets for item ID %s have already been updated", zooTicket.ItemId)
			continue
		}

		// Update the ticket with data from the Zoo API
		ticketToUpdate.EncryptedId = zooTicket.EncryptedID
		ticketToUpdate.AdmitDate = zooTicket.AdmitDate

		// Parse unit price if needed
		if unitPrice, err := strconv.ParseFloat(zooTicket.UnitPrice, 64); err == nil {
			ticketToUpdate.UnitPrice = unitPrice
		}

		// Update the ticket in the database
		err = orderTicketInfoRepo.Update(ticketToUpdate)
		if err != nil {
			log.Printf("Failed to update ticket %s: %v", ticketToUpdate.OrderTicketInfoId, err)
			// Continue updating other tickets
		}

		// Remove this ticket from the list to ensure we don't update it again
		for i, t := range ticketsForItem {
			if t == ticketToUpdate {
				ticketsForItem = append(ticketsForItem[:i], ticketsForItem[i+1:]...)
				break
			}
		}
		ticketByItemId[zooTicket.ItemId] = ticketsForItem
	}

	return nil
}

// Function to generate a token for the Zoo API
func generateZooAPIToken() (string, error) {
	// Create form data for x-www-form-urlencoded request
	formData := url.Values{}
	formData.Set("grant_type", "password")
	formData.Set("UserName", "Tester")
	formData.Set("Password", "TestingAbc123")

	// Create a new HTTP client
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	// Create a new request
	req, err := http.NewRequest("POST", "https://eglobal2.ddns.net/johorzooapi/Token", strings.NewReader(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("error creating token request: %w", err)
	}

	// Add headers
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error executing token request: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading token response: %w", err)
	}

	// Parse the JSON response
	var tokenResponse struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return "", fmt.Errorf("error parsing token JSON: %w", err)
	}

	// Check if we got an access token
	if tokenResponse.AccessToken == "" {
		return "", fmt.Errorf("no access token in response: %s", string(body))
	}

	// Return the access token
	return tokenResponse.AccessToken, nil
}
