package service

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"j-ticketing/internal/core/dto/payment"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/config"
	"j-ticketing/pkg/email"
	"j-ticketing/pkg/utils"
	"log"
	logger "log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type PaymentService struct {
	orderTicketGroupRepo *repositories.OrderTicketGroupRepository
	orderTicketInfoRepo  *repositories.OrderTicketInfoRepository
	ticketGroupRepo      *repositories.TicketGroupRepository
	tagRepo              *repositories.TagRepository
	groupGalleryRepo     *repositories.GroupGalleryRepository
	ticketDetailRepo     *repositories.TicketDetailRepository
	paymentConfig        *payment.PaymentConfig
	cfg                  *config.Config
}

// NewPaymentService creates a new order service
func NewPaymentService(
	orderTicketGroupRepo *repositories.OrderTicketGroupRepository,
	orderTicketInfoRepo *repositories.OrderTicketInfoRepository,
	ticketGroupRepo *repositories.TicketGroupRepository,
	tagRepo *repositories.TagRepository,
	groupGalleryRepo *repositories.GroupGalleryRepository,
	ticketDetailRepo *repositories.TicketDetailRepository,
	paymentConfig *payment.PaymentConfig,
	cfg *config.Config,
) *PaymentService {
	return &PaymentService{
		orderTicketGroupRepo: orderTicketGroupRepo,
		orderTicketInfoRepo:  orderTicketInfoRepo,
		ticketGroupRepo:      ticketGroupRepo,
		tagRepo:              tagRepo,
		groupGalleryRepo:     groupGalleryRepo,
		ticketDetailRepo:     ticketDetailRepo,
		cfg:                  cfg,
	}
}

func (s *PaymentService) FindByOrderNo(orderNo string) (*models.OrderTicketGroup, error) {
	// Find the order first
	order, err := s.orderTicketGroupRepo.FindByOrderNo(orderNo)
	if err != nil {
		log.Printf("Error finding order %s: %v", orderNo, err)
		return nil, err
	}

	if order == nil {
		log.Printf("Order not found: %s", orderNo)
		return nil, fmt.Errorf("order not found: %s", orderNo)
	}

	return order, nil
}

func (s *PaymentService) UpdateOrderTicketGroup(orderTicketGroup *models.OrderTicketGroup) error {
	// Find the order first
	err := s.orderTicketGroupRepo.Update(orderTicketGroup)
	if err != nil {
		log.Printf("Error updating order: %v", err)
		return err
	}
	return nil
}

// Define the function to post to the Zoo API
func (s *PaymentService) PostToZooAPI(orderNo string) (*models.OrderTicketGroup, []email.OrderInfo, []email.TicketInfo, error) {
	// Find the order first
	orderTicketGroup, err := s.orderTicketGroupRepo.FindByOrderNo(orderNo)
	if err != nil {
		log.Printf("Error finding order %s: %v", orderNo, err)
		return nil, nil, nil, err
	}

	if orderTicketGroup == nil {
		log.Printf("Order not found: %s", orderNo)
		return nil, nil, nil, fmt.Errorf("order not found: %s", orderNo)
	}

	ticketGroupName := orderTicketGroup.TicketGroup.GroupNameBm
	fmt.Printf(ticketGroupName)

	// Get the order ticket items
	orderTickets, err := s.orderTicketInfoRepo.FindByOrderTicketGroupID(orderTicketGroup.OrderTicketGroupId)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get order tickets: %w", err)
	}

	// Build the request
	items := make([]payment.TicketItem, 0, len(orderTickets))
	for _, ticket := range orderTickets {
		items = append(items, payment.TicketItem{
			ItemId: ticket.ItemId,
			Qty:    ticket.QuantityBought,
		})
	}

	var zooTicketInfos []payment.ZooTicketInfo
	// Check if first order ticket is already assigned with qr code or not
	if orderTickets[0].EncryptedId == "" {
		logger.Info("ZooTicketInfo will now be assigned with QR Codes")
		// Get the first order ticket info to know the admit date
		admissionDate := orderTickets[0].AdmitDate

		// Create the request payload
		payload := payment.ZooTicketRequest{
			TranDate:    admissionDate,
			ReferenceNo: orderNo, // Use the order number as reference
			Items:       items,
		}

		// Convert to JSON
		jsonData, err := json.Marshal(payload)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to marshal JSON: %w", err)
		}

		// Get a fresh token from the token generation endpoint
		token, err := generateZooAPIToken(s.cfg)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to generate API token: %w", err)
		}

		// Create a new HTTP client
		client := &http.Client{
			Timeout: time.Second * 60,
		}

		// Create the request
		var value = "PostOnlinePurchase2"
		//if ticketGroupName == "Zoo Johor" {
		//	value = "PostOnlinePurchase"
		//} else {
		//	value = "PostOnlinePurchase2" // Used for botani
		//}

		req, err := http.NewRequest("POST", s.cfg.ZooAPI.ZooBaseURL+"/api/JohorZoo/"+value, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Add headers
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", "Bearer "+token)

		// Send the request
		resp, err := client.Do(req)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to send request: %w", err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				log.Printf("Error closing response body: %v", err)
			}
		}(resp.Body)

		// Read the response
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to read response: %w", err)
		}

		// Check if the response status is OK
		if resp.StatusCode != http.StatusOK {
			return nil, nil, nil, fmt.Errorf("API returned non-OK status: %d, body: %s", resp.StatusCode, string(body))
		}

		// Parse the response
		var zooResponse payment.ZooTicketResponse
		err = json.Unmarshal(body, &zooResponse)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to parse response: %w", err)
		}

		zooTicketInfos = zooResponse.Tickets

		// Check if the status code is OK
		if zooResponse.StatusCode != "OK" {
			return nil, nil, nil, fmt.Errorf("API returned status code: %s", zooResponse.StatusCode)
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
			//ticketToUpdate.EncryptedId = "STF020"
			ticketToUpdate.AdmitDate = zooTicket.AdmitDate

			// Parse unit price if needed
			if unitPrice, err := strconv.ParseFloat(zooTicket.UnitPrice, 64); err == nil {
				ticketToUpdate.UnitPrice = unitPrice
			}

			// Update the ticket in the database
			err = s.orderTicketInfoRepo.Update(ticketToUpdate)
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
	} else {
		logger.Info("ZooTicketInfo have already been assigned with QR Codes")
		for _, ticket := range orderTickets {
			zooTicketInfos = append(zooTicketInfos, payment.ZooTicketInfo{
				TWBID:       ticket.Twbid.String,
				ItemId:      ticket.ItemId,
				EncryptedID: ticket.EncryptedId,
				AdmitDate:   ticket.AdmitDate,
				UnitPrice:   fmt.Sprintf("%.2f", ticket.UnitPrice),
				ItemDesc:    ticket.ItemDesc1,
				ItemDesc2:   ticket.ItemDesc2,
				ItemDesc3:   ticket.ItemDesc3,
			})
		}
	}

	ticketInfos := ConvertZooTicketsToTicketInfo(zooTicketInfos, orderTicketGroup.LangChosen)

	orderItems := ConvertZooTicketsToOrderItems(zooTicketInfos, orderTicketGroup.LangChosen)

	return orderTicketGroup, orderItems, ticketInfos, nil
}

// Function to generate a token for the Zoo API
func generateZooAPIToken(cfg *config.Config) (string, error) {
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
	req, err := http.NewRequest("POST", cfg.ZooAPI.ZooBaseURL+"/Token", strings.NewReader(formData.Encode()))
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

// ConvertZooTicketsToTicketInfo converts tickets from Zoo API format to TicketInfo format for emails
func ConvertZooTicketsToTicketInfo(zooTickets []payment.ZooTicketInfo, langChosen string) []email.TicketInfo {
	ticketInfos := make([]email.TicketInfo, 0, len(zooTickets))

	for _, ticket := range zooTickets {
		var label string
		if langChosen == "bm" {
			label = ticket.ItemDesc
		} else if langChosen == "en" {
			label = ticket.ItemDesc2
		} else if langChosen == "cn" {
			label = ticket.ItemDesc3
		} else {
			label = "unknown"
		}

		// Create TicketInfo with EncryptedID as QR code content
		ticketInfo := email.TicketInfo{
			Content: ticket.EncryptedID,
			Label:   label,
		}

		ticketInfos = append(ticketInfos, ticketInfo)
	}

	return ticketInfos
}

// ConvertZooTicketsToOrderItems converts tickets from Zoo API format to OrderInfo format for emails
// with grouping by ItemId to combine identical tickets
func ConvertZooTicketsToOrderItems(zooTickets []payment.ZooTicketInfo, langChosen string) []email.OrderInfo {
	// Use a map to group tickets by ItemId
	ticketGroups := make(map[string]*struct {
		Count     int
		ItemDesc  string
		UnitPrice string
		EntryDate string
	})

	// Group tickets by ItemId and count them
	for _, ticket := range zooTickets {
		itemId := ticket.ItemId

		// If this is the first ticket with this ItemId, initialize the group
		if _, exists := ticketGroups[itemId]; !exists {
			var itemDesc string
			if langChosen == "bm" {
				itemDesc = ticket.ItemDesc
			} else if langChosen == "en" {
				itemDesc = ticket.ItemDesc2
			} else if langChosen == "cn" {
				itemDesc = ticket.ItemDesc3
			} else {
				itemDesc = "unknown"
			}

			ticketGroups[itemId] = &struct {
				Count     int
				ItemDesc  string
				UnitPrice string
				EntryDate string
			}{
				Count:     0,
				ItemDesc:  itemDesc,
				UnitPrice: ticket.UnitPrice,
				EntryDate: ticket.AdmitDate,
			}
		}

		// Increment the count for this ItemId
		ticketGroups[itemId].Count++
	}

	// Convert the grouped tickets to OrderInfo objects
	orderItems := make([]email.OrderInfo, 0, len(ticketGroups))

	for _, group := range ticketGroups {
		// Only create order items for groups with at least one ticket
		if group.Count > 0 {
			// Create an OrderInfo for this group
			orderInfo := email.OrderInfo{
				Description: group.ItemDesc,
				Quantity:    group.Count,
				Price:       utils.ParseFloat(group.UnitPrice),
				EntryDate:   group.EntryDate,
			}

			orderItems = append(orderItems, orderInfo)
		}
	}

	return orderItems
}

func (s *PaymentService) GenerateInternalQRCodes(orderNo string) (*models.OrderTicketGroup, []email.OrderInfo, []email.TicketInfo, error) {
	// Find the order first
	orderTicketGroup, err := s.orderTicketGroupRepo.FindByOrderNo(orderNo)
	if err != nil {
		log.Printf("Error finding order %s: %v", orderNo, err)
		return nil, nil, nil, err
	}

	if orderTicketGroup == nil {
		log.Printf("Order not found: %s", orderNo)
		return nil, nil, nil, fmt.Errorf("order not found: %s", orderNo)
	}

	ticketGroupName := orderTicketGroup.TicketGroup.GroupNameBm
	fmt.Printf(ticketGroupName)

	// Get the order ticket items
	orderTickets, err := s.orderTicketInfoRepo.FindByOrderTicketGroupID(orderTicketGroup.OrderTicketGroupId)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get order tickets: %w", err)
	}

	// Build the request
	items := make([]payment.TicketItem, 0, len(orderTickets))
	for _, ticket := range orderTickets {
		items = append(items, payment.TicketItem{
			ItemId: ticket.ItemId,
			Qty:    ticket.QuantityBought,
		})
	}

	var zooTicketInfos []payment.ZooTicketInfo
	// Check if first order ticket is already assigned with qr code or not
	if orderTickets[0].EncryptedId == "" {
		logger.Info("Internal Tickets will now be assigned with QR Codes")

		randomStr, _ := utils.GenerateRandomString(12)
		for _, ticket := range orderTickets {
			zooTicketInfos = append(zooTicketInfos, payment.ZooTicketInfo{
				TWBID:       "",
				ItemId:      ticket.ItemId,
				EncryptedID: randomStr,
				AdmitDate:   ticket.AdmitDate,
				UnitPrice:   fmt.Sprintf("%.2f", ticket.UnitPrice),
				ItemDesc:    ticket.ItemDesc1,
				ItemDesc2:   ticket.ItemDesc2,
				ItemDesc3:   ticket.ItemDesc3,
			})

			ticket.EncryptedId = randomStr

			// Update the ticket in the database
			err = s.orderTicketInfoRepo.Update(&ticket)
			if err != nil {
				log.Printf("Failed to update ticket %s: %v", ticket.OrderTicketInfoId, err)
				// Continue updating other tickets
			}
		}
	} else {
		logger.Info("Internal Tickets have already been assigned with QR Codes")
		for _, ticket := range orderTickets {
			zooTicketInfos = append(zooTicketInfos, payment.ZooTicketInfo{
				TWBID:       ticket.Twbid.String,
				ItemId:      ticket.ItemId,
				EncryptedID: ticket.EncryptedId,
				AdmitDate:   ticket.AdmitDate,
				UnitPrice:   fmt.Sprintf("%.2f", ticket.UnitPrice),
				ItemDesc:    ticket.ItemDesc1,
				ItemDesc2:   ticket.ItemDesc2,
				ItemDesc3:   ticket.ItemDesc3,
			})
		}
	}

	ticketInfos := ConvertZooTicketsToTicketInfo(zooTicketInfos, orderTicketGroup.LangChosen)

	orderItems := ConvertZooTicketsToOrderItems(zooTicketInfos, orderTicketGroup.LangChosen)

	return orderTicketGroup, orderItems, ticketInfos, nil
}

func (s *PaymentService) UpdateOrderFromPaymentResponse(orderNo string, transactionData payment.TransactionResponse, order *models.OrderTicketGroup) error {

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
	order.TransactionDate = transactionData.TarikhTransaksi
	order.TransactionStatus = dbStatus
	order.BankCurrentStatus = transactionData.StatusTransaksi
	order.StatusMessage = sql.NullString{String: transactionData.StatusMessage, Valid: transactionData.StatusMessage != ""}
	order.UpdatedAt = time.Now()

	// Save the updated order
	err := s.orderTicketGroupRepo.Update(order)
	if err != nil {
		log.Printf("Error updating order: %v", err)
		return err
	}

	log.Printf("Successfully updated order %s with transaction ID %s and status %s",
		orderNo, transactionData.IDTransaksi, dbStatus)

	return nil
}
