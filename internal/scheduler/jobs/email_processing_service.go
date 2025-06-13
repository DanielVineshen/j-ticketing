// File: j-ticketing/internal/scheduler/jobs/email_processing_service.go
package jobs

import (
	"fmt"
	services "j-ticketing/internal/core/services"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/email"
	"j-ticketing/pkg/utils"
	"log"
)

// EmailProcessingService handles processing of ticket orders
type EmailProcessingService struct {
	paymentService       *services.PaymentService
	orderTicketGroupRepo *repositories.OrderTicketGroupRepository
	orderTicketInfoRepo  *repositories.OrderTicketInfoRepository
	emailService         email.EmailService
	ticketGroupService   *services.TicketGroupService
	pdfService           *services.PDFService
	orderService         *services.OrderService
}

// NewEmailProcessingService creates a new EmailProcessingService
func NewEmailProcessingService(
	paymentService *services.PaymentService,
	orderTicketGroupRepo *repositories.OrderTicketGroupRepository,
	orderTicketInfoRepo *repositories.OrderTicketInfoRepository,
	emailService email.EmailService,
	ticketGroupService *services.TicketGroupService,
	pdfService *services.PDFService,
	orderService *services.OrderService,
) *EmailProcessingService {
	return &EmailProcessingService{
		paymentService:       paymentService,
		orderTicketGroupRepo: orderTicketGroupRepo,
		orderTicketInfoRepo:  orderTicketInfoRepo,
		emailService:         emailService,
		ticketGroupService:   ticketGroupService,
		pdfService:           pdfService,
		orderService:         orderService,
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

	ticketGroup, err := s.ticketGroupService.GetTicketGroup(order.TicketGroupId)
	if err != nil {
		log.Printf("Error finding ticket group %s: %v", order.TicketGroupId, err)
	}

	var orderItems []email.OrderInfo
	var ticketInfos []email.TicketInfo
	if ticketGroup.IsTicketInternal {
		_, orderItems, ticketInfos, err = s.paymentService.GenerateInternalQRCodes(order.OrderNo)
	} else {
		// Only call the Zoo API if payment was successful
		order, orderItems, ticketInfos, err = s.paymentService.PostToZooAPI(order.OrderNo)
		if err != nil {
			log.Printf("Error posting to Johor Zoo API: %v", err)
		} else {
			err = s.orderService.CreateOrderTicketLog("order", "QR Code Assigned", "Order was assigned with qr codes for each ticket", "QR Service", order)
			if err != nil {
				return err
			}
		}
	}

	total := utils.CalculateOrderTotal(orderItems)

	var ticketGroupName string
	if order.LangChosen == "bm" {
		ticketGroupName = ticketGroup.GroupNameBm
	} else if order.LangChosen == "en" {
		ticketGroupName = ticketGroup.GroupNameEn
	} else if order.LangChosen == "cn" {
		ticketGroupName = ticketGroup.GroupNameCn
	} else {
		ticketGroupName = ""
	}
	orderOverview := email.OrderOverview{
		TicketGroup:  ticketGroupName,
		FullName:     order.BuyerName,
		PurchaseDate: order.TransactionDate,
		EntryDate:    orderItems[0].EntryDate,
		Quantity:     len(orderItems),
		OrderNumber:  order.OrderNo,
		Total:        total,
	}

	pdfBytes, pdfFilename, err := s.pdfService.GenerateTicketPDF(orderOverview, orderItems, ticketInfos, order.LangChosen)
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

	err = s.emailService.SendTicketsEmail(order.BuyerEmail, orderOverview, orderItems, ticketInfos, []email.Attachment{pdfAttachment}, order.LangChosen)
	if err != nil {
		log.Printf("Failed to send tickets email to %s: %v", order.BuyerEmail, err)
		// Continue anyway since the password has been reset
	} else {
		err = s.orderService.CreateOrderTicketLog("order", "Email Sent", "Email for the order was successfully sent out with its receipt", "Email Service", order)
		if err != nil {
			return err
		}

		order.IsEmailSent = true
		// Save the updated order
		err = s.paymentService.UpdateOrderTicketGroup(order)
		if err != nil {
			log.Printf("Failed to update order ticket group: %v", err)
		}
	}

	log.Printf("Successfully processed order")
	return nil
}
