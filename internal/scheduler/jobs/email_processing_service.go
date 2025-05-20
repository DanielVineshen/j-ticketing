package jobs

import (
	"fmt"
	"j-ticketing/internal/core/dto/payment"
	services "j-ticketing/internal/core/services"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/email"
	"log"
)

// EmailProcessingService handles processing of ticket orders
type EmailProcessingService struct {
	paymentService       *services.PaymentService
	orderTicketGroupRepo *repositories.OrderTicketGroupRepository
	orderTicketInfoRepo  *repositories.OrderTicketInfoRepository
	paymentConfig        payment.PaymentConfig
	emailService         email.EmailService
	ticketGroupService   *services.TicketGroupService
	pdfService           *services.PDFService
}

// NewEmailProcessingService creates a new EmailProcessingService
func NewEmailProcessingService(paymentService *services.PaymentService, orderTicketGroupRepo *repositories.OrderTicketGroupRepository, orderTicketInfoRepo *repositories.OrderTicketInfoRepository, paymentConfig payment.PaymentConfig, emailService email.EmailService, ticketGroupService *services.TicketGroupService, pdfService *services.PDFService) *EmailProcessingService {
	return &EmailProcessingService{
		paymentService:       paymentService,
		orderTicketGroupRepo: orderTicketGroupRepo,
		orderTicketInfoRepo:  orderTicketInfoRepo,
		paymentConfig:        paymentConfig,
		emailService:         emailService,
		ticketGroupService:   ticketGroupService,
		pdfService:           pdfService,
	}
}

func (s *EmailProcessingService) ProcessPendingOrders() (int, error) {
	// Retrieve orders that need processing, limit to 50 at a time to avoid overload
	orders, err := s.orderTicketGroupRepo.FindEmailPendingOrderTicketGroups()
	if err != nil {
		return 0, fmt.Errorf("error finding orders to process: %w", err)
	}

	log.Printf("Found %d orders to process", len(orders))

	processedCount := 0

	// Process each order
	for _, order := range orders {
		if err := s.processOrder(&order); err != nil {
			log.Printf("Error processing order %s: %v", order.OrderTicketGroupId, err)
		} else {
			processedCount++
		}
	}

	return processedCount, nil
}

// processOrder processes a single order
func (s *EmailProcessingService) processOrder(order *models.OrderTicketGroup) error {
	log.Printf("Processing order %v", order)

	orderTicketInfos, err := s.orderTicketInfoRepo.FindByOrderTicketGroupID(order.OrderTicketGroupId)
	if err != nil {
		return fmt.Errorf("error finding order ticker infos to process: %w", err)
	}

	if orderTicketInfos[0].EncryptedId == "" {

	}

	var orderItems []email.OrderInfo
	var ticketInfos []email.TicketInfo
	// Only call the Zoo API if payment was successful
	orderItems, ticketInfos, err = s.paymentService.PostToZooAPI(order.OrderNo)
	if err != nil {
		log.Printf("Error posting to Johor Zoo API: %v", err)
		// Continue with redirect even if this fails, we can retry later
	}

	ticketGroup, err := s.ticketGroupService.GetTicketGroup(order.TicketGroupId)
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

	pdfBytes, pdfFilename, err := s.pdfService.GenerateTicketPDF(ticketGroup.GroupName, ticketInfos)
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

	err = s.emailService.SendTicketsEmail(order.BuyerEmail, orderOverview, orderItems, ticketInfos, []email.Attachment{pdfAttachment})
	if err != nil {
		log.Printf("Failed to send tickets email to %s: %v", order.BuyerEmail, err)
		// Continue anyway since the password has been reset
	}
	order.IsEmailSent = true
	// Save the updated order
	err = s.paymentService.UpdateOrderTicketGroup(order)
	if err != nil {
		log.Printf("Failed to update order ticket group: %v", err)
	}

	log.Printf("Successfully processed order")
	return nil
}
