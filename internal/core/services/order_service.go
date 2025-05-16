// File: j-ticketing/internal/core/services/order_service.go
package service

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"errors"
	"fmt"
	orderDto "j-ticketing/internal/core/dto/order"
	ticketGroupDto "j-ticketing/internal/core/dto/ticket_group"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	mathrand "math/rand"
	"strings"
	"time"
)

var secureRandom = mathrand.New(mathrand.NewSource(time.Now().UnixNano()))

// OrderService handles order-related operations
type OrderService struct {
	orderTicketGroupRepo *repositories.OrderTicketGroupRepository
	orderTicketInfoRepo  *repositories.OrderTicketInfoRepository
	ticketGroupRepo      *repositories.TicketGroupRepository
	tagRepo              *repositories.TagRepository
	groupGalleryRepo     *repositories.GroupGalleryRepository
	ticketDetailRepo     *repositories.TicketDetailRepository
}

// NewOrderService creates a new order service
func NewOrderService(
	orderTicketGroupRepo *repositories.OrderTicketGroupRepository,
	orderTicketInfoRepo *repositories.OrderTicketInfoRepository,
	ticketGroupRepo *repositories.TicketGroupRepository,
	tagRepo *repositories.TagRepository,
	groupGalleryRepo *repositories.GroupGalleryRepository,
	ticketDetailRepo *repositories.TicketDetailRepository,
) *OrderService {
	return &OrderService{
		orderTicketGroupRepo: orderTicketGroupRepo,
		orderTicketInfoRepo:  orderTicketInfoRepo,
		ticketGroupRepo:      ticketGroupRepo,
		tagRepo:              tagRepo,
		groupGalleryRepo:     groupGalleryRepo,
		ticketDetailRepo:     ticketDetailRepo,
	}
}

// GetAllOrderTicketGroups retrieves all order ticket groups for a user
func (s *OrderService) GetAllOrderTicketGroups(custId string) (orderDto.OrderTicketGroupResponse, error) {
	var orders []models.OrderTicketGroup
	var err error

	// If customer ID is provided, retrieve orders for that customer only
	if custId != "" {
		orders, err = s.orderTicketGroupRepo.FindByCustomerID(custId)
	} else {
		// Otherwise, retrieve all orders
		orders, err = s.orderTicketGroupRepo.FindAll()
	}

	if err != nil {
		return orderDto.OrderTicketGroupResponse{}, err
	}

	response := orderDto.OrderTicketGroupResponse{
		OrderTicketGroups: make([]orderDto.OrderTicketGroupDTO, 0, len(orders)),
	}

	for _, order := range orders {
		orderDTO, err := s.mapOrderToDTO(&order)
		if err != nil {
			return orderDto.OrderTicketGroupResponse{}, err
		}

		response.OrderTicketGroups = append(response.OrderTicketGroups, orderDTO)
	}

	return response, nil
}

// GetOrderTicketGroup retrieves a specific order ticket group
func (s *OrderService) GetOrderTicketGroup(orderTicketGroupId uint) (*orderDto.OrderTicketGroupDTO, error) {
	// Use FindWithDetails to get the order with its relations
	order, err := s.orderTicketGroupRepo.FindWithDetails(orderTicketGroupId)
	if err != nil {
		return nil, err
	}

	orderDTO, err := s.mapOrderToDTO(order)
	if err != nil {
		return nil, err
	}

	return &orderDTO, nil
}

// mapOrderToDTO maps an order model to its DTO representation
func (s *OrderService) mapOrderToDTO(order *models.OrderTicketGroup) (orderDto.OrderTicketGroupDTO, error) {
	// Map the order profile
	orderProfile := orderDto.OrderProfileDTO{
		OrderTicketGroupId: order.OrderTicketGroupId,
		TicketGroupId:      order.TicketGroupId,
		CustId:             order.CustId,
		TransactionId:      order.TransactionId,
		OrderNo:            order.OrderNo,
		TransactionStatus:  order.TransactionStatus,
		TransactionDate:    order.TransactionDate,
		MsgToken:           order.MsgToken,
		BillId:             order.BillId,
		ProductId:          order.ProductId,
		TotalAmount:        order.TotalAmount,
		BuyerName:          order.BuyerName,
		BuyerEmail:         order.BuyerEmail,
		ProductDesc:        order.ProductDesc,
		OrderTicketInfo:    make([]orderDto.OrderTicketInfoDTO, 0),
		CreatedAt:          order.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          order.UpdatedAt.Format(time.RFC3339),
	}

	// Add optional fields if they exist
	if order.StatusMessage.Valid {
		orderProfile.StatusMessage = order.StatusMessage.String
	}
	if order.BankCode.Valid {
		orderProfile.BankCode = order.BankCode.String
	}
	if order.BankName.Valid {
		orderProfile.BankName = order.BankName.String
	}

	// Get order ticket info items if they're not already loaded
	var orderInfos []models.OrderTicketInfo
	if len(order.OrderTicketInfos) > 0 {
		orderInfos = order.OrderTicketInfos
	} else {
		var err error
		orderInfos, err = s.orderTicketInfoRepo.FindByOrderTicketGroupID(order.OrderTicketGroupId)
		if err != nil {
			return orderDto.OrderTicketGroupDTO{}, err
		}
	}

	// Map order ticket info items
	for _, info := range orderInfos {
		infoDTO := orderDto.OrderTicketInfoDTO{
			OrderTicketInfoId:  info.OrderTicketInfoId,
			OrderTicketGroupId: info.OrderTicketGroupId,
			ItemId:             info.ItemId,
			UnitPrice:          info.UnitPrice,
			ItemDesc1:          info.ItemDesc1,
			ItemDesc2:          info.ItemDesc2,
			PrintType:          info.PrintType,
			QuantityBought:     info.QuantityBought,
			EncryptedId:        info.EncryptedId,
			AdmitDate:          info.AdmitDate,
			Variant:            info.Variant,
			CreatedAt:          info.CreatedAt.Format(time.RFC3339),
			UpdatedAt:          info.UpdatedAt.Format(time.RFC3339),
		}

		if info.Twbid.Valid {
			infoDTO.Twbid = info.Twbid.String
		}

		orderProfile.OrderTicketInfo = append(orderProfile.OrderTicketInfo, infoDTO)
	}

	// Get the ticket group
	var ticketGroup *models.TicketGroup
	if order.TicketGroup.TicketGroupId > 0 {
		// If already loaded via preload
		ticketGroup = &order.TicketGroup
	} else {
		// Otherwise fetch it
		var err error
		ticketGroup, err = s.ticketGroupRepo.FindByID(order.TicketGroupId)
		if err != nil {
			return orderDto.OrderTicketGroupDTO{}, err
		}
	}

	// Get tags for this ticket group
	tags, err := s.tagRepo.FindByTicketGroupID(ticketGroup.TicketGroupId)
	if err != nil {
		return orderDto.OrderTicketGroupDTO{}, err
	}

	// Map tags to DTOs
	tagDTOs := make([]ticketGroupDto.TagDTO, 0, len(tags))
	for _, tag := range tags {
		tagDTOs = append(tagDTOs, ticketGroupDto.TagDTO{
			TagId:   tag.TagId,
			TagName: tag.TagName,
			TagDesc: tag.TagDesc,
		})
	}

	// Get gallery items for this ticket group
	galleries, err := s.groupGalleryRepo.FindByTicketGroupID(ticketGroup.TicketGroupId)
	if err != nil {
		return orderDto.OrderTicketGroupDTO{}, err
	}

	// Map gallery items to DTOs
	galleryDTOs := make([]ticketGroupDto.GroupGalleryDTO, 0, len(galleries))
	for _, gallery := range galleries {
		galleryDTOs = append(galleryDTOs, ticketGroupDto.GroupGalleryDTO{
			GroupGalleryId:  gallery.GroupGalleryId,
			AttachmentName:  gallery.AttachmentName,
			AttachmentPath:  gallery.AttachmentPath,
			AttachmentSize:  gallery.AttachmentSize,
			ContentType:     gallery.ContentType,
			UniqueExtension: gallery.UniqueExtension,
		})
	}

	// Get ticket details for this ticket group
	details, err := s.ticketDetailRepo.FindByTicketGroupID(ticketGroup.TicketGroupId)
	if err != nil {
		return orderDto.OrderTicketGroupDTO{}, err
	}

	// Map ticket details to DTOs
	detailDTOs := make([]ticketGroupDto.TicketDetailDTO, 0, len(details))
	for _, detail := range details {
		detailDTOs = append(detailDTOs, ticketGroupDto.TicketDetailDTO{
			TicketDetailId: detail.TicketDetailId,
			Title:          detail.Title,
			TitleIcon:      detail.TitleIcon,
			RawHtml:        detail.RawHtml,
			DisplayFlag:    detail.DisplayFlag,
		})
	}

	// Parse organiser facilities from string to string array
	var organiserFacilities []string
	if ticketGroup.OrganiserFacilities != "" {
		// Split the string based on the semicolon separator
		organiserFacilities = strings.Split(ticketGroup.OrganiserFacilities, ";")

		// Trim any whitespace from each facility
		for i, facility := range organiserFacilities {
			organiserFacilities[i] = strings.TrimSpace(facility)
		}
	} else {
		organiserFacilities = []string{} // Empty array if no facilities
	}

	// Build the ticket profile DTO
	ticketProfile := ticketGroupDto.TicketProfileDTO{
		TicketGroupId:            ticketGroup.TicketGroupId,
		GroupType:                ticketGroup.GroupType,
		GroupName:                ticketGroup.GroupName,
		GroupDesc:                ticketGroup.GroupDesc,
		OperatingHours:           ticketGroup.OperatingHours,
		PricePrefix:              ticketGroup.PricePrefix,
		PriceSuffix:              ticketGroup.PriceSuffix,
		AttachmentName:           ticketGroup.AttachmentName,
		AttachmentPath:           ticketGroup.AttachmentPath,
		AttachmentSize:           ticketGroup.AttachmentSize,
		ContentType:              ticketGroup.ContentType,
		UniqueExtension:          ticketGroup.UniqueExtension,
		IsActive:                 ticketGroup.IsActive,
		IsTicketInternal:         ticketGroup.IsTicketInternal,
		TicketIds:                ticketGroup.TicketIds.String,
		Tags:                     tagDTOs,
		GroupGallery:             galleryDTOs,
		TicketDetails:            detailDTOs,
		LocationAddress:          ticketGroup.LocationAddress,
		LocationMapEmbedUrl:      ticketGroup.LocationMapUrl,
		OrganiserName:            ticketGroup.OrganiserName,
		OrganiserAddress:         ticketGroup.OrganiserAddress,
		OrganiserDescriptionHtml: ticketGroup.OrganiserDescHtml,
		OrganiserContact:         ticketGroup.OrganiserContact,
		OrganiserEmail:           ticketGroup.OrganiserEmail,
		OrganiserWebsite:         ticketGroup.OrganiserWebsite,
		OrganiserOperatingHours:  ticketGroup.OrganiserOperatingHour,
		OrganiserFacilities:      organiserFacilities,
		CreatedAt:                ticketGroup.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                ticketGroup.UpdatedAt.Format(time.RFC3339),
	}

	// Handle optional date fields
	if ticketGroup.ActiveStartDate.Valid {
		ticketProfile.ActiveStartDate = ticketGroup.ActiveStartDate.String
	}
	if ticketGroup.ActiveEndDate.Valid {
		ticketProfile.ActiveEndDate = ticketGroup.ActiveEndDate.String
	}

	// Create the complete order ticket group DTO
	orderTicketGroupDTO := orderDto.OrderTicketGroupDTO{
		OrderProfile:  orderProfile,
		TicketProfile: ticketProfile,
	}

	return orderTicketGroupDTO, nil
}

// CreateOrder creates a new order ticket group and returns the order ID
func (s *OrderService) CreateOrder(custId string, req *orderDto.CreateOrderRequest) (uint, error) {
	// Validate ticket group exists
	ticketGroup, err := s.ticketGroupRepo.FindByID(req.TicketGroupId)
	if err != nil {
		return 0, fmt.Errorf("ticket group not found: %w", err)
	}

	// Ensure ticket group is active
	if !ticketGroup.IsActive {
		return 0, errors.New("ticket group is not active")
	}

	// Parse and validate date
	orderDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return 0, fmt.Errorf("invalid date format: %w", err)
	}

	// Generate order number and transaction ID
	orderNo := generateOrderNumber()
	transactionId := generateTransactionId()

	// Calculate total amount based on tickets
	var totalAmount float64 = 0

	// Create order ticket group
	orderTicketGroup := &models.OrderTicketGroup{
		TicketGroupId:     req.TicketGroupId,
		CustId:            custId,
		TransactionId:     transactionId,
		OrderNo:           orderNo,
		TransactionStatus: "PENDING", // Initial status
		TransactionDate:   time.Now().String(),
		MsgToken:          generateMessageToken(),
		BillId:            generateBillId(),
		ProductId:         fmt.Sprintf("TG%d", req.TicketGroupId),
		TotalAmount:       totalAmount, // Will be updated after calculating tickets
		BuyerName:         req.FullName,
		BuyerEmail:        req.Email,
		ProductDesc:       ticketGroup.GroupName,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Set optional fields based on payment type
	if req.PaymentType == "fpx" {
		orderTicketGroup.BankCode = sql.NullString{
			String: req.BankCode,
			Valid:  true,
		}

		// You may need to lookup the bank name based on the code
		bankName, err := s.getBankNameByCode(req.BankCode)
		if err == nil && bankName != "" {
			orderTicketGroup.BankName = sql.NullString{
				String: bankName,
				Valid:  true,
			}
		}
	}

	// Save order ticket group to get the ID
	err = s.orderTicketGroupRepo.Create(orderTicketGroup)
	if err != nil {
		return 0, fmt.Errorf("failed to create order: %w", err)
	}

	// Process tickets
	orderTicketInfos := make([]models.OrderTicketInfo, 0, len(req.Tickets))

	for _, ticket := range req.Tickets {
		// You might need to look up ticket details from a ticket repository
		// For now, we'll create basic info

		// Calculate unit price based on ticket ID (This is a placeholder - implement actual logic)
		unitPrice := s.calculateTicketPrice(ticket.TicketId, req.TicketGroupId)

		// Update total amount
		totalAmount += unitPrice * float64(ticket.Qty)

		// Create ticket info entries for each quantity
		for i := 0; i < ticket.Qty; i++ {
			orderTicketInfo := models.OrderTicketInfo{
				OrderTicketGroupId: orderTicketGroup.OrderTicketGroupId,
				ItemId:             ticket.TicketId,
				UnitPrice:          unitPrice,
				ItemDesc1:          fmt.Sprintf("%s - %s", ticketGroup.GroupName, ticket.TicketId),
				ItemDesc2:          req.Date,
				PrintType:          "ETICKET",
				QuantityBought:     1, // Each entry represents 1 ticket
				EncryptedId:        generateEncryptedId(orderTicketGroup.OrderTicketGroupId, ticket.TicketId, i),
				AdmitDate:          orderDate.String(),
				Variant:            "STANDARD",
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			}

			orderTicketInfos = append(orderTicketInfos, orderTicketInfo)
		}
	}

	// Update total amount in order ticket group
	orderTicketGroup.TotalAmount = totalAmount
	err = s.orderTicketGroupRepo.Update(orderTicketGroup)
	if err != nil {
		return 0, fmt.Errorf("failed to update order total amount: %w", err)
	}

	// Save all ticket info entries
	err = s.orderTicketInfoRepo.BatchCreate(orderTicketInfos)
	if err != nil {
		return 0, fmt.Errorf("failed to create order tickets: %w", err)
	}

	// Return the order ID
	return orderTicketGroup.OrderTicketGroupId, nil
}

// Helper functions for order creation

func (s *OrderService) getBankNameByCode(bankCode string) (string, error) {
	// This would typically query a bank repository or use a lookup table
	// For simplicity, using a map
	banks := map[string]string{
		"MBBEMYKL": "Maybank",
		"CIBBMYKL": "CIMB Bank",
		"PHBMMYKL": "Public Bank",
		"RHBBMYKL": "RHB Bank",
		"HBMBMYKL": "HSBC Bank",
		// Add more banks as needed
	}

	if name, exists := banks[bankCode]; exists {
		return name, nil
	}

	return "", fmt.Errorf("bank code not found: %s", bankCode)
}

func (s *OrderService) calculateTicketPrice(ticketId string, ticketGroupId uint) float64 {
	// This would typically query a ticket price repository
	// For now, returning a placeholder price
	// Implement actual price lookup logic
	return 50.00
}

// Utility functions for generating IDs and tokens

func generateOrderNumber() string {
	// Format: ORD-YYYYMMDDHHmmss-XXXX
	timestamp := time.Now().Format("20060102150405")
	random := fmt.Sprintf("%04d", secureRandom.Intn(10000))
	return fmt.Sprintf("ORD-%s-%s", timestamp, random)
}

func generateTransactionId() string {
	// Format: TXN-YYYYMMDDHHmmss-XXXX
	timestamp := time.Now().Format("20060102150405")
	random := fmt.Sprintf("%04d", secureRandom.Intn(10000))
	return fmt.Sprintf("TXN-%s-%s", timestamp, random)
}

func generateMessageToken() string {
	// Generate a random token
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to math/rand if crypto/rand fails
		binary.BigEndian.PutUint64(b, uint64(secureRandom.Int63()))
		binary.BigEndian.PutUint64(b[8:], uint64(secureRandom.Int63()))
	}
	return fmt.Sprintf("%x", b)
}

func generateBillId() string {
	// Format: BILL-YYYYMMDDHHmmss-XXXX
	timestamp := time.Now().Format("20060102150405")
	random := fmt.Sprintf("%04d", secureRandom.Intn(10000))
	return fmt.Sprintf("BILL-%s-%s", timestamp, random)
}

func generateEncryptedId(orderGroupId uint, ticketId string, index int) string {
	// This would typically use a proper encryption algorithm
	// For simplicity, using a hash-based approach
	data := fmt.Sprintf("%d:%s:%d:%d", orderGroupId, ticketId, index, time.Now().UnixNano())
	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}
