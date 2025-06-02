// File: j-ticketing/internal/core/services/order_service.go
package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	orderDto "j-ticketing/internal/core/dto/order"
	"j-ticketing/internal/core/dto/payment"
	ticketGroupDto "j-ticketing/internal/core/dto/ticket_group"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/utils"
	"log"
	mathrand "math/rand"
	"net/http"
	"net/url"
	"strconv"
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
	paymentConfig        *payment.PaymentConfig
	ticketGroupService   *TicketGroupService
	orderTicketLogRepo   *repositories.OrderTicketLogRepository
	customerService      *CustomerService
}

// NewOrderService creates a new order service
func NewOrderService(
	orderTicketGroupRepo *repositories.OrderTicketGroupRepository,
	orderTicketInfoRepo *repositories.OrderTicketInfoRepository,
	ticketGroupRepo *repositories.TicketGroupRepository,
	tagRepo *repositories.TagRepository,
	groupGalleryRepo *repositories.GroupGalleryRepository,
	ticketDetailRepo *repositories.TicketDetailRepository,
	paymentConfig *payment.PaymentConfig,
	ticketGroupService *TicketGroupService,
	orderTicketLogRepo *repositories.OrderTicketLogRepository,
	customerService *CustomerService,
) *OrderService {
	return &OrderService{
		orderTicketGroupRepo: orderTicketGroupRepo,
		orderTicketInfoRepo:  orderTicketInfoRepo,
		ticketGroupRepo:      ticketGroupRepo,
		tagRepo:              tagRepo,
		groupGalleryRepo:     groupGalleryRepo,
		ticketDetailRepo:     ticketDetailRepo,
		paymentConfig:        paymentConfig,
		ticketGroupService:   ticketGroupService,
		orderTicketLogRepo:   orderTicketLogRepo,
		customerService:      customerService,
	}
}

// GetAllOrderTicketGroups retrieves all order ticket groups for a user
func (s *OrderService) GetAllOrderTicketGroups(custId string, orderNo string, startDate string, endDate string) (orderDto.OrderTicketGroupResponse, error) {
	var orders []models.OrderTicketGroup
	var err error

	// If customer ID is provided, retrieve orders for that customer only
	if custId != "" {
		orders, err = s.orderTicketGroupRepo.FindByCustomerID(custId)
	} else {
		orders, err = s.orderTicketGroupRepo.FindByDateRange(orderNo, startDate, endDate)
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
func (s *OrderService) GetOrderTicketGroupRaw(orderTicketGroupId uint) (*models.OrderTicketGroup, error) {
	// Use FindWithOrderTicketGroupId to get the order with its relations
	order, err := s.orderTicketGroupRepo.FindWithOrderTicketGroupId(orderTicketGroupId)
	if err != nil {
		return nil, err
	}

	return order, nil
}

// GetOrderTicketGroup retrieves a specific order ticket group
func (s *OrderService) GetOrderTicketGroup(orderTicketGroupId uint) (*orderDto.OrderTicketGroupDTO, error) {
	// Use FindWithOrderTicketGroupId to get the order with its relations
	order, err := s.orderTicketGroupRepo.FindWithOrderTicketGroupId(orderTicketGroupId)
	if err != nil {
		return nil, err
	}

	orderDTO, err := s.mapOrderToDTO(order)
	if err != nil {
		return nil, err
	}

	return &orderDTO, nil
}

func (s *OrderService) GetOrderNonMemberInquiry(orderNo string, email string) (*orderDto.OrderTicketGroupDTO, error) {
	// Use FindWithOrderTicketGroupId to get the order with its relations
	order, err := s.orderTicketGroupRepo.FindWithOrderNoAndEmail(orderNo, email)
	if err != nil {
		return nil, err
	}

	orderDTO, err := s.mapOrderToDTO(order)
	if err != nil {
		return nil, err
	}

	return &orderDTO, nil
}

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
		OrderTicketLog:     make([]orderDto.OrderTicketLogDTO, 0),
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

	// Use preloaded OrderTicketInfos directly (no database call)
	for _, info := range order.OrderTicketInfos {
		infoDTO := orderDto.OrderTicketInfoDTO{
			OrderTicketInfoId:  info.OrderTicketInfoId,
			OrderTicketGroupId: info.OrderTicketGroupId,
			ItemId:             info.ItemId,
			UnitPrice:          info.UnitPrice,
			ItemDesc1:          info.ItemDesc1,
			ItemDesc2:          info.ItemDesc2,
			ItemDesc3:          info.ItemDesc3,
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

	for _, ticket := range order.OrderTicketLogs {
		ticketDTO := orderDto.OrderTicketLogDTO{
			OrderTicketLogId:   ticket.OrderTicketLogId,
			OrderTicketGroupId: ticket.OrderTicketGroupId,
			PerformedBy:        getStringFromNullString(ticket.PerformedBy),
			Type:               ticket.Type,
			Title:              ticket.Title,
			Message:            ticket.Message,
			Date:               ticket.Date,
			CreatedAt:          ticket.CreatedAt.Format(time.RFC3339),
			UpdatedAt:          ticket.UpdatedAt.Format(time.RFC3339),
		}

		orderProfile.OrderTicketLog = append(orderProfile.OrderTicketLog, ticketDTO)
	}

	//  Use preloaded TicketGroup directly (no database call)
	ticketGroup := &order.TicketGroup

	// Validate that ticket group is loaded
	if ticketGroup.TicketGroupId == 0 {
		return orderDto.OrderTicketGroupDTO{}, fmt.Errorf("ticket group not preloaded for order %d", order.OrderTicketGroupId)
	}

	// Use preloaded tags from TicketGroup.TicketTags.Tag (no database call)
	tagDTOs := make([]ticketGroupDto.TagDTO, 0, len(ticketGroup.TicketTags))
	for _, ticketTag := range ticketGroup.TicketTags {
		if ticketTag.Tag.TagId != 0 { // Ensure tag is loaded
			tagDTOs = append(tagDTOs, ticketGroupDto.TagDTO{
				TagId:   ticketTag.Tag.TagId,
				TagName: ticketTag.Tag.TagName,
				TagDesc: ticketTag.Tag.TagDesc,
			})
		}
	}

	// Use preloaded galleries from TicketGroup.GroupGalleries (no database call)
	galleryDTOs := make([]ticketGroupDto.GroupGalleryDTO, 0, len(ticketGroup.GroupGalleries))
	for _, gallery := range ticketGroup.GroupGalleries {
		galleryDTOs = append(galleryDTOs, ticketGroupDto.GroupGalleryDTO{
			GroupGalleryId:  gallery.GroupGalleryId,
			AttachmentName:  gallery.AttachmentName,
			AttachmentPath:  gallery.AttachmentPath,
			AttachmentSize:  gallery.AttachmentSize,
			ContentType:     gallery.ContentType,
			UniqueExtension: gallery.UniqueExtension,
		})
	}

	// Use preloaded details from TicketGroup.TicketDetails (no database call)
	detailDTOs := make([]ticketGroupDto.TicketDetailDTO, 0, len(ticketGroup.TicketDetails))
	for _, detail := range ticketGroup.TicketDetails {
		detailDTOs = append(detailDTOs, ticketGroupDto.TicketDetailDTO{
			TicketDetailId: detail.TicketDetailId,
			Title:          detail.Title,
			TitleIcon:      detail.TitleIcon,
			RawHtml:        detail.RawHtml,
			DisplayFlag:    detail.DisplayFlag,
		})
	}

	ticketVariantsDTOs := make([]ticketGroupDto.TicketVariantDTO, 0, len(ticketGroup.TicketVariants))
	for _, variant := range ticketGroup.TicketVariants {
		ticketVariantsDTOs = append(ticketVariantsDTOs, ticketGroupDto.TicketVariantDTO{
			TicketVariantId: &variant.TicketVariantId,
			TicketGroupId:   &variant.TicketGroupId,
			NameBm:          variant.NameBm,
			NameEn:          variant.NameEn,
			NameCn:          variant.NameCn,
			DescBm:          variant.DescBm,
			DescEn:          variant.DescEn,
			DescCn:          variant.DescCn,
			UnitPrice:       variant.UnitPrice,
		})
	}

	// Parse organiser facilities from string to string array
	var organiserFacilitiesBm []string
	facilitiesBmStr := getStringFromNullString(ticketGroup.OrganiserFacilitiesBm)
	if facilitiesBmStr != "" {
		organiserFacilitiesBm = strings.Split(facilitiesBmStr, ";")
		for i, facility := range organiserFacilitiesBm {
			organiserFacilitiesBm[i] = strings.TrimSpace(facility)
		}
	} else {
		organiserFacilitiesBm = []string{}
	}

	var organiserFacilitiesEn []string
	facilitiesEnStr := getStringFromNullString(ticketGroup.OrganiserFacilitiesEn)
	if facilitiesEnStr != "" {
		organiserFacilitiesEn = strings.Split(facilitiesEnStr, ";")
		for i, facility := range organiserFacilitiesEn {
			organiserFacilitiesEn[i] = strings.TrimSpace(facility)
		}
	} else {
		organiserFacilitiesEn = []string{}
	}

	var organiserFacilitiesCn []string
	facilitiesCnStr := getStringFromNullString(ticketGroup.OrganiserFacilitiesCn)
	if facilitiesCnStr != "" {
		organiserFacilitiesCn = strings.Split(facilitiesCnStr, ";")
		for i, facility := range organiserFacilitiesCn {
			organiserFacilitiesCn[i] = strings.TrimSpace(facility)
		}
	} else {
		organiserFacilitiesCn = []string{}
	}

	// Build the ticket profile DTO
	ticketProfile := ticketGroupDto.TicketProfileDTO{
		TicketGroupId:              ticketGroup.TicketGroupId,
		OrderTicketLimit:           ticketGroup.OrderTicketLimit,
		ScanSetting:                ticketGroup.ScanSetting,
		GroupType:                  ticketGroup.GroupType,
		GroupNameBm:                ticketGroup.GroupNameBm,
		GroupNameEn:                ticketGroup.GroupNameEn,
		GroupNameCn:                ticketGroup.GroupNameCn,
		GroupDescBm:                ticketGroup.GroupDescBm,
		GroupDescEn:                ticketGroup.GroupDescEn,
		GroupDescCn:                ticketGroup.GroupDescCn,
		GroupRedirectionSpanBm:     nullStringToPointer(ticketGroup.GroupRedirectionSpanBm),
		GroupRedirectionSpanEn:     nullStringToPointer(ticketGroup.GroupRedirectionSpanEn),
		GroupRedirectionSpanCn:     nullStringToPointer(ticketGroup.GroupRedirectionSpanCn),
		GroupRedirectionUrl:        nullStringToPointer(ticketGroup.GroupRedirectionUrl),
		GroupSlot1Bm:               nullStringToPointer(ticketGroup.GroupSlot1Bm),
		GroupSlot1En:               nullStringToPointer(ticketGroup.GroupSlot1En),
		GroupSlot1Cn:               nullStringToPointer(ticketGroup.GroupSlot1Cn),
		GroupSlot2Bm:               nullStringToPointer(ticketGroup.GroupSlot2Bm),
		GroupSlot2En:               nullStringToPointer(ticketGroup.GroupSlot2En),
		GroupSlot2Cn:               nullStringToPointer(ticketGroup.GroupSlot2Cn),
		GroupSlot3Bm:               nullStringToPointer(ticketGroup.GroupSlot3Bm),
		GroupSlot3En:               nullStringToPointer(ticketGroup.GroupSlot3En),
		GroupSlot3Cn:               nullStringToPointer(ticketGroup.GroupSlot3Cn),
		GroupSlot4Bm:               nullStringToPointer(ticketGroup.GroupSlot4Bm),
		GroupSlot4En:               nullStringToPointer(ticketGroup.GroupSlot4En),
		GroupSlot4Cn:               nullStringToPointer(ticketGroup.GroupSlot4Cn),
		PricePrefixBm:              ticketGroup.PricePrefixBm,
		PricePrefixEn:              ticketGroup.PricePrefixEn,
		PricePrefixCn:              ticketGroup.PricePrefixCn,
		PriceSuffixBm:              ticketGroup.PriceSuffixBm,
		PriceSuffixEn:              ticketGroup.PriceSuffixEn,
		PriceSuffixCn:              ticketGroup.PriceSuffixCn,
		AttachmentName:             ticketGroup.AttachmentName,
		AttachmentPath:             ticketGroup.AttachmentPath,
		AttachmentSize:             ticketGroup.AttachmentSize,
		ContentType:                ticketGroup.ContentType,
		UniqueExtension:            ticketGroup.UniqueExtension,
		IsActive:                   ticketGroup.IsActive,
		IsTicketInternal:           ticketGroup.IsTicketInternal,
		TicketIds:                  ticketGroup.TicketIds.String,
		Tags:                       tagDTOs,
		GroupGallery:               galleryDTOs,
		TicketDetails:              detailDTOs,
		LocationAddress:            ticketGroup.LocationAddress,
		LocationMapEmbedUrl:        ticketGroup.LocationMapUrl,
		OrganiserNameBm:            ticketGroup.OrganiserNameBm,
		OrganiserNameEn:            ticketGroup.OrganiserNameEn,
		OrganiserNameCn:            ticketGroup.OrganiserNameCn,
		OrganiserAddress:           ticketGroup.OrganiserAddress,
		OrganiserDescriptionHtmlBm: ticketGroup.OrganiserDescHtmlBm,
		OrganiserDescriptionHtmlEn: ticketGroup.OrganiserDescHtmlEn,
		OrganiserDescriptionHtmlCn: ticketGroup.OrganiserDescHtmlCn,
		OrganiserContact:           nullStringToPointer(ticketGroup.OrganiserContact),
		OrganiserEmail:             nullStringToPointer(ticketGroup.OrganiserEmail),
		OrganiserWebsite:           nullStringToPointer(ticketGroup.OrganiserWebsite),
		OrganiserFacilitiesBm:      organiserFacilitiesBm,
		OrganiserFacilitiesEn:      organiserFacilitiesEn,
		OrganiserFacilitiesCn:      organiserFacilitiesCn,
		TicketVariants:             ticketVariantsDTOs,
		CreatedAt:                  ticketGroup.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                  ticketGroup.UpdatedAt.Format(time.RFC3339),
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
func (s *OrderService) CreateOrder(custId string, req *orderDto.CreateOrderRequest) (*models.OrderTicketGroup, error) {
	// Validate ticket group exists
	ticketGroup, err := s.ticketGroupRepo.FindByID(req.TicketGroupId)
	if err != nil {
		return nil, fmt.Errorf("ticket group not found: %w", err)
	}

	// Ensure ticket group is active
	if !ticketGroup.IsActive {
		return nil, errors.New("ticket group is not active")
	}

	// Parse and validate date
	orderDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Retrieve available ticket variants for validation and pricing
	ticketVariantsResponse, err := s.ticketGroupService.GetTicketVariants(req.TicketGroupId, req.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve ticket variants: %w", err)
	}

	ticketVariantMap := make(map[string]ticketGroupDto.TicketVariantDTO)

	if !ticketGroup.IsTicketInternal {
		// Create a map for easier lookup of ticket variants
		for _, variant := range ticketVariantsResponse.TicketVariants {
			ticketVariantMap[*variant.TicketId] = variant
		}
	} else {
		for _, variant := range ticketVariantsResponse.TicketVariants {
			ticketVariantMap[strconv.Itoa(int(*variant.TicketVariantId))] = variant
		}
	}

	// Validate that all requested tickets exist in available variants
	for _, ticket := range req.Tickets {
		if _, exists := ticketVariantMap[ticket.TicketId]; !exists {
			return nil, fmt.Errorf("ticket ID %s is not available for this group and date", ticket.TicketId)
		}
	}

	// Generate order number and transaction ID
	orderNo := generateOrderNumber()

	var mode = ""
	if req.Mode == "individual" {
		mode = "01"
	} else if req.Mode == "corporate" {
		mode = "02"
	}

	// Create order ticket group
	orderTicketGroup := &models.OrderTicketGroup{
		TicketGroupId:     req.TicketGroupId,
		CustId:            custId,
		TransactionId:     "",
		OrderNo:           orderNo,
		TransactionStatus: "initiate", // Initial status
		TransactionDate:   "",         // Only assigned trans date after order redirect
		MsgToken:          mode,
		BillId:            generateBillId(),
		ProductId:         fmt.Sprintf("TG%d", req.TicketGroupId),
		TotalAmount:       0, // Will be updated after calculating tickets
		BuyerName:         req.FullName,
		BuyerEmail:        req.Email,
		ProductDesc:       ticketGroup.GroupNameBm,
		IsEmailSent:       false,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Set optional fields based on payment type
	if req.PaymentType == "fpx" {
		// Validate bank code if FPX payment is selected
		if req.BankCode == "" {
			return nil, errors.New("bank code is required for FPX payment")
		}

		// Lookup the bank name based on the code
		bankName, err := s.getBankNameByCode(req.BankCode, req.Mode)
		if err != nil {
			return nil, fmt.Errorf("invalid bank code: %w", err)
		}

		// Set bank code and name
		orderTicketGroup.BankCode = sql.NullString{
			String: req.BankCode,
			Valid:  true,
		}

		orderTicketGroup.BankName = sql.NullString{
			String: bankName,
			Valid:  true,
		}
	}

	// Calculate totalAmount first by looping through tickets before creating any records
	totalAmount := 0.0 // or whatever type your totalAmount is
	orderTicketInfos := make([]models.OrderTicketInfo, 0, len(req.Tickets))

	for _, ticket := range req.Tickets {
		// Get the ticket variant details
		variant, exists := ticketVariantMap[ticket.TicketId]
		if !exists {
			// This should never happen as we've already validated all tickets
			return nil, fmt.Errorf("unexpected error: ticket ID %s not found", ticket.TicketId)
		}

		// Use the unit price from the variant
		unitPrice := variant.UnitPrice

		// Create multiple entries for tickets with quantity > 1
		for i := 0; i < ticket.Qty; i++ {
			// Update total amount for each individual ticket
			totalAmount += unitPrice

			// Create a new ticket info entry for each quantity
			orderTicketInfo := models.OrderTicketInfo{
				OrderTicketGroupId: orderTicketGroup.OrderTicketGroupId,
				ItemId:             ticket.TicketId,
				UnitPrice:          unitPrice,
				ItemDesc1:          variant.NameBm,
				ItemDesc2:          variant.NameEn,
				ItemDesc3:          variant.NameCn,
				PrintType:          *variant.PrintType, // Use print type from API
				QuantityBought:     1,                  // Fixed quantity of 1
				EncryptedId:        "",
				AdmitDate:          orderDate.Format("2006-01-02"), // Format the date consistently
				Variant:            "default",
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			}

			// Add directly to the slice
			orderTicketInfos = append(orderTicketInfos, orderTicketInfo)
		}
	}

	// Check if totalAmount == 0
	if totalAmount == 0 {
		return nil, fmt.Errorf("failed to create order: total amount for tickets must be more than 0")
	}

	// Save order ticket group to get the ID
	orderTicketGroup, err = s.orderTicketGroupRepo.Create(orderTicketGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Now that we have the OrderTicketGroupId, update all orderTicketInfos
	for i := range orderTicketInfos {
		orderTicketInfos[i].OrderTicketGroupId = orderTicketGroup.OrderTicketGroupId
	}

	// Update total amount in order ticket group
	orderTicketGroup.TotalAmount = totalAmount
	err = s.orderTicketGroupRepo.Update(orderTicketGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to update order total amount: %w", err)
	}

	// Save all ticket info entries
	err = s.orderTicketInfoRepo.BatchCreate(orderTicketInfos)
	if err != nil {
		return nil, fmt.Errorf("failed to create order tickets: %w", err)
	}

	// Return the order ID
	return orderTicketGroup, nil
}

// CreateOrder creates a new free order ticket group and returns the order ID
func (s *OrderService) CreateFreeOrder(cust *models.Customer, req *orderDto.CreateFreeOrderRequest) (*models.OrderTicketGroup, error) {
	// Validate ticket group exists
	ticketGroup, err := s.ticketGroupRepo.FindByID(req.TicketGroupId)
	if err != nil {
		return nil, fmt.Errorf("ticket group not found: %w", err)
	}

	// Ensure ticket group is active
	if !ticketGroup.IsActive {
		return nil, errors.New("ticket group is not active")
	}

	// Parse and validate date
	orderDate, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Retrieve available ticket variants for validation and pricing
	ticketVariantsResponse, err := s.ticketGroupService.GetTicketVariants(req.TicketGroupId, req.Date)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve ticket variants: %w", err)
	}

	ticketVariantMap := make(map[string]ticketGroupDto.TicketVariantDTO)

	if !ticketGroup.IsTicketInternal {
		// Create a map for easier lookup of ticket variants
		for _, variant := range ticketVariantsResponse.TicketVariants {
			ticketVariantMap[*variant.TicketId] = variant
		}
	} else {
		for _, variant := range ticketVariantsResponse.TicketVariants {
			ticketVariantMap[strconv.Itoa(int(*variant.TicketVariantId))] = variant
		}
	}

	// Validate that all requested tickets exist in available variants
	for _, ticket := range req.Tickets {
		if _, exists := ticketVariantMap[ticket.TicketId]; !exists {
			return nil, fmt.Errorf("ticket ID %s is not available for this group and date", ticket.TicketId)
		}
	}

	// Generate order number and transaction ID
	orderNo := generateOrderNumber()

	malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	if err != nil {
		// Handle the error appropriately
		// Perhaps log it, or return it from your function
		return nil, err // or handle differently based on your application's needs
	}

	// Create order ticket group
	orderTicketGroup := &models.OrderTicketGroup{
		TicketGroupId:     req.TicketGroupId,
		CustId:            cust.CustId,
		TransactionId:     "",
		OrderNo:           orderNo,
		TransactionStatus: "success",
		TransactionDate:   malaysiaTime,
		MsgToken:          "",
		BillId:            generateBillId(),
		ProductId:         fmt.Sprintf("TG%d", req.TicketGroupId),
		TotalAmount:       0,
		BuyerName:         req.FullName,
		BuyerEmail:        req.Email,
		ProductDesc:       ticketGroup.GroupNameBm,
		IsEmailSent:       false,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Calculate totalAmount first by looping through tickets before creating any records
	totalAmount := 0.0 // or whatever type your totalAmount is
	orderTicketInfos := make([]models.OrderTicketInfo, 0, len(req.Tickets))

	for _, ticket := range req.Tickets {
		// Get the ticket variant details
		variant, exists := ticketVariantMap[ticket.TicketId]
		if !exists {
			return nil, fmt.Errorf("unexpected error: ticket ID %s not found", ticket.TicketId)
		}

		// Use the unit price from the variant
		unitPrice := variant.UnitPrice

		// Calculate total amount for all tickets in this loop
		totalAmount += unitPrice * float64(ticket.Qty) // Adjust type as needed

		// Prepare orderTicketInfos without OrderTicketGroupId for now
		for i := 0; i < ticket.Qty; i++ {
			// Create a new ticket info entry for each quantity
			orderTicketInfo := models.OrderTicketInfo{
				// OrderTicketGroupId will be set after creating the order
				ItemId:         ticket.TicketId,
				UnitPrice:      unitPrice,
				ItemDesc1:      variant.NameBm,
				ItemDesc2:      variant.NameEn,
				ItemDesc3:      variant.NameCn,
				PrintType:      *variant.PrintType,
				QuantityBought: 1,
				EncryptedId:    "",
				AdmitDate:      orderDate.Format("2006-01-02"),
				Variant:        "default",
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}

			// Add to the slice
			orderTicketInfos = append(orderTicketInfos, orderTicketInfo)
		}
	}

	// Check if totalAmount > 0
	if totalAmount > 0 {
		return nil, fmt.Errorf("failed to create order: total amount for tickets exceed 0")
	}

	// If we pass the validation, save the order ticket group
	orderTicketGroup, err = s.orderTicketGroupRepo.Create(orderTicketGroup)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Now that we have the OrderTicketGroupId, update all orderTicketInfos
	for i := range orderTicketInfos {
		orderTicketInfos[i].OrderTicketGroupId = orderTicketGroup.OrderTicketGroupId
	}

	// Save all ticket info entries
	err = s.orderTicketInfoRepo.BatchCreate(orderTicketInfos)
	if err != nil {
		return nil, fmt.Errorf("failed to create order tickets: %w", err)
	}

	err = s.customerService.CreateCustomerLog("purchase", "Purchase Completed", "Ticket package purchased via Online", *cust)
	if err != nil {
		return nil, err
	}

	// Return the order
	return orderTicketGroup, nil
}

// getBankNameByCode retrieves a bank name by its code and validates if the bank is enabled
// Returns the bank name if found and enabled, or an error if not
func (s *OrderService) getBankNameByCode(bankCode, mode string) (string, error) {
	// Get the API key from config
	apiKey := s.paymentConfig.APIKey

	// Create form data for x-www-form-urlencoded request
	formData := url.Values{}
	formData.Set("jp_ag_token", "ZOO")
	formData.Set("method", "getBankList")
	if mode == "individual" {
		formData.Set("mode", "01")
	} else {
		formData.Set("mode", "02")
	}

	// Create a new HTTP client
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	// Create a new request
	req, err := http.NewRequest("POST", s.paymentConfig.GatewayURL+"/JP_gateway/getBankList", strings.NewReader(formData.Encode()))
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return "", fmt.Errorf("error creating request: %w", err)

	}

	// Add headers
	req.Header.Add("jp-api-key", apiKey)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Execute the request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error executing request: %v", err)
		return "", fmt.Errorf("error executing request: %w", err)

	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	// Parse the JSON response
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("Error parsing JSON: %v", err)
		return "", fmt.Errorf("error parsing JSON: %w", err)
	}

	// Check if the request was successful
	if success, ok := result["success"].(bool); ok && success {
		// The response has a data field that contains a JSON string (not an object)
		// We need to parse this string into an array of bank objects
		if dataStr, ok := result["data"].(string); ok {
			var banks []map[string]interface{}
			if err := json.Unmarshal([]byte(dataStr), &banks); err != nil {
				log.Printf("Error parsing bank data: %v", err)
				return "", fmt.Errorf("error parsing bank data: %w", err)
			}

			// Look for the bank with the matching code
			for _, bank := range banks {
				if value, ok := bank["value"].(string); ok && value == bankCode {
					// Check if the bank is enabled
					if enabled, ok := bank["enabled"].(float64); ok && enabled == 1 {
						// Return the bank name
						if name, ok := bank["name"].(string); ok {
							return name, nil
						}
					} else {
						return "", fmt.Errorf("bank %s is disabled", bankCode)
					}
				}
			}
		}
	}

	return "", fmt.Errorf("failed to retrieve bank list")
}

func (s *OrderService) CreateOrderTicketLog(logType string, title string, message string, performedBy string, orderTicketGroup *models.OrderTicketGroup) error {
	malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	if err != nil {
		return err
	}

	if performedBy == "" {
		performedBy = "System"
	}

	orderTicketLog := models.OrderTicketLog{
		OrderTicketGroupId: orderTicketGroup.OrderTicketGroupId,
		Type:               logType,
		Title:              title,
		Message:            message,
		Date:               malaysiaTime,
		PerformedBy:        sql.NullString{String: performedBy, Valid: true},
	}

	err = s.orderTicketLogRepo.Create(&orderTicketLog)
	if err != nil {
		return err
	}

	return nil
}

// Utility functions for generating IDs and tokens

func generateOrderNumber() string {
	// Format: ORD-YYYYMMDDHHmmss-XXXX
	timestamp := time.Now().Format("20060102150405")
	random := fmt.Sprintf("%04d", secureRandom.Intn(10000))
	return fmt.Sprintf("ORD-%s-%s", timestamp, random)
}

func generateBillId() string {
	// Format: BILL-YYYYMMDDHHmmss-XXXX
	timestamp := time.Now().Format("20060102150405")
	random := fmt.Sprintf("%04d", secureRandom.Intn(10000))
	return fmt.Sprintf("BILL-%s-%s", timestamp, random)
}
