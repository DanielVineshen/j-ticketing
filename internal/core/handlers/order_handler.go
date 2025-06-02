// File: j-ticketing/internal/core/handlers/order_handler.go
package handlers

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	orderDto "j-ticketing/internal/core/dto/order"
	"j-ticketing/internal/core/dto/payment"
	services "j-ticketing/internal/core/services"
	dbModels "j-ticketing/internal/db/models"
	"j-ticketing/pkg/email"
	"j-ticketing/pkg/errors"
	"j-ticketing/pkg/jwt"
	"j-ticketing/pkg/models"
	"j-ticketing/pkg/utils"
	"log"
	"strconv"
	"strings"
)

// OrderHandler handles HTTP requests for orders
type OrderHandler struct {
	orderService        *services.OrderService
	customerService     services.CustomerService
	jwtService          jwt.JWTService
	paymentService      *services.PaymentService
	emailService        email.EmailService
	ticketGroupService  *services.TicketGroupService
	paymentConfig       payment.PaymentConfig
	pdfService          *services.PDFService
	notificationService services.NotificationService
}

// NewOrderHandler creates a new instance of OrderHandler
func NewOrderHandler(orderService *services.OrderService,
	customerService services.CustomerService,
	jwtService jwt.JWTService,
	paymentService *services.PaymentService,
	emailService email.EmailService,
	ticketGroupService *services.TicketGroupService,
	paymentConfig payment.PaymentConfig,
	pdfService *services.PDFService,
	notificationService services.NotificationService) *OrderHandler {
	return &OrderHandler{
		orderService:        orderService,
		customerService:     customerService,
		jwtService:          jwtService,
		paymentService:      paymentService,
		emailService:        emailService,
		ticketGroupService:  ticketGroupService,
		paymentConfig:       paymentConfig,
		pdfService:          pdfService,
		notificationService: notificationService,
	}
}

// GetOrderTicketGroups handles GET requests for order ticket groups
func (h *OrderHandler) GetOrderTicketGroups(c *fiber.Ctx) error {
	// Get the customer ID from the context (set by auth middleware)
	custId, ok := c.Locals("userId").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewBaseErrorResponse(
			"User not authenticated", nil,
		))
	}

	// Get the user type from the context
	userType, ok := c.Locals("userType").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewBaseErrorResponse(
			"User type not found", nil,
		))
	}

	// Get the user role from the context
	userRole, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewBaseErrorResponse(
			"User role not found", nil,
		))
	}

	var response interface{}
	var err error

	// If user is admin, allow fetching all orders or filtering by customer ID
	if userType == "admin" && (userRole == "SYSADMIN" || userRole == "ADMIN") {
		// Admin can optionally filter by customer ID
		filterCustId := c.Query("custId")
		response, err = h.orderService.GetAllOrderTicketGroups(filterCustId, "", "", "")
	} else if userType == "customer" {
		// Customer can only see their own orders
		response, err = h.orderService.GetAllOrderTicketGroups(custId, "", "", "")
	} else {
		return c.Status(fiber.StatusForbidden).JSON(models.NewBaseErrorResponse(
			"You are not authorized to view these orders", nil,
		))
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"Failed to retrieve order ticket groups: "+err.Error(), nil,
		))
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}

// GetOrderTicketGroup handles GET request for a specific order ticket group
func (h *OrderHandler) GetOrderTicketGroup(c *fiber.Ctx) error {
	// Parse the order ticket group ID from the request
	orderTicketGroupIdStr := c.Query("orderTicketGroupId")
	if orderTicketGroupIdStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Missing order ticket group ID", nil,
		))
	}

	orderTicketGroupId, err := strconv.ParseUint(orderTicketGroupIdStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid order ticket group ID", nil,
		))
	}
	// Get the order ticket group
	order, err := h.orderService.GetOrderTicketGroup(uint(orderTicketGroupId))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
			"Order ticket group not found", nil,
		))
	}

	return c.JSON(models.NewBaseSuccessResponse(order))
}

func (h *OrderHandler) GetOrderNonMemberInquiry(c *fiber.Ctx) error {
	orderNoStr := c.Query("orderNo")
	if orderNoStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Missing order number", nil,
		))
	}

	emailStr := c.Query("email")
	if emailStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Missing email", nil,
		))
	}

	// Get the order ticket group
	order, err := h.orderService.GetOrderNonMemberInquiry(orderNoStr, emailStr)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
			"Order ticket group not found", nil,
		))
	}

	return c.JSON(models.NewBaseSuccessResponse(order))
}

// CreateOrderTicketGroup handles POST requests to create a new order
func (h *OrderHandler) CreateOrderTicketGroup(c *fiber.Ctx) error {
	// Parse request body first so we can use the data either way
	var req orderDto.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request format", nil,
		))
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	authHeader := c.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := h.jwtService.ValidateToken(token)

	// Variable to hold the customer ID
	var custId string

	// Check if token exists (user is authenticated)
	if err == nil && claims.UserID != "" {
		custId = claims.UserID

		customer, err := h.customerService.GetCustomerByID(claims.UserID)
		if err == nil {
			req.Email = customer.Email
			req.FullName = customer.FullName
			if customer.ContactNo.Valid {
				req.ContactNo = customer.ContactNo.String
			} else {
				req.ContactNo = "" // Empty string for NULL values
			}
		} else {
			return c.Status(fiber.StatusForbidden).JSON(models.NewBaseErrorResponse(
				errors.USER_NOT_AUTHORIZED.Message, nil,
			))
		}
	} else {
		// Check if any of the required fields are empty strings
		if req.Email == "" || req.IdentificationNo == "" || req.FullName == "" || req.ContactNo == "" {
			return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
				"All customer information (email, identification number, full name, and contact number) must be provided", nil,
			))
		}

		customer, err := h.customerService.GetCustomerByEmail(req.Email)
		if err != nil {
			// Customer doesn't exist, create a new one
			newCustomer, err := h.customerService.RegisterCustomer(
				req.Email,
				"",
				req.IdentificationNo,
				req.FullName,
				req.ContactNo,
			)

			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
					"Failed to create customer: "+err.Error(), nil,
				))
			}

			custId = newCustomer.CustId
		} else {
			custId = customer.CustId
		}
	}

	// Create the order using the custId we determined
	orderTicketGroup, err := h.orderService.CreateOrder(custId, &req)
	if err != nil {
		// Determine appropriate error code based on the error
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
				err.Error(), nil,
			))
		} else if strings.Contains(err.Error(), "payment processing failed") {
			return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
				err.Error(), nil,
			))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
				"Failed to create order: "+err.Error(), nil,
			))
		}
	}

	err = h.orderService.CreateOrderTicketLog("order", "Order Created", "Order was created via Online channel", "System", orderTicketGroup)
	if err != nil {
		return err
	}

	malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	if err != nil {
		return err
	}
	message := fmt.Sprintf("%s (%s) has created an order no: %s", req.Email, req.FullName, orderTicketGroup.OrderNo)
	err = h.notificationService.CreateNotification(
		req.FullName,
		"customer",
		"Order",
		"Customer create order",
		message,
		malaysiaTime,
	)
	if err != nil {
		return err
	}

	// Generate the checkout URL
	checkoutURL := h.generateCheckoutURL(orderTicketGroup.OrderTicketGroupId, req.PaymentType)

	// Return success response with redirect information
	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(map[string]interface{}{
		"redirectURL": checkoutURL,
		"orderID":     orderTicketGroup.OrderTicketGroupId,
	}))
}

func (h *OrderHandler) CreateFreeOrderTicketGroup(c *fiber.Ctx) error {
	// Parse request body first so we can use the data either way
	var req orderDto.CreateFreeOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request format", nil,
		))
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	authHeader := c.Get("Authorization")
	token := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := h.jwtService.ValidateToken(token)

	// Variable to hold the customer ID
	var cust dbModels.Customer

	// Check if token exists (user is authenticated)
	if err == nil && claims.UserID != "" {
		customer, err := h.customerService.GetCustomerByID(claims.UserID)
		if err == nil {
			req.Email = customer.Email
			req.FullName = customer.FullName
			if customer.ContactNo.Valid {
				req.ContactNo = customer.ContactNo.String
			} else {
				req.ContactNo = "" // Empty string for NULL values
			}
		} else {
			return c.Status(fiber.StatusForbidden).JSON(models.NewBaseErrorResponse(
				errors.USER_NOT_AUTHORIZED.Message, nil,
			))
		}

		cust = *customer
	} else {
		// Check if any of the required fields are empty strings
		if req.Email == "" || req.IdentificationNo == "" || req.FullName == "" || req.ContactNo == "" {
			return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
				"All customer information (email, identification number, full name, and contact number) must be provided", nil,
			))
		}

		customer, err := h.customerService.GetCustomerByEmail(req.Email)
		if err != nil {
			// Customer doesn't exist, create a new one
			newCustomer, err := h.customerService.RegisterCustomer(
				req.Email,
				"",
				req.IdentificationNo,
				req.FullName,
				req.ContactNo,
			)

			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
					"Failed to create customer: "+err.Error(), nil,
				))
			}

			cust = *newCustomer
		} else {
			cust = *customer
		}
	}

	// Create the order using the custId we determined
	orderTicketGroup, err := h.orderService.CreateFreeOrder(&cust, &req)
	if err != nil {
		// Determine appropriate error code based on the error
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
				err.Error(), nil,
			))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
				"Failed to create order: "+err.Error(), nil,
			))
		}
	}

	err = h.orderService.CreateOrderTicketLog("order", "Order Created", "Order was created via Online channel", "System", orderTicketGroup)
	if err != nil {
		return err
	}

	malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	if err != nil {
		return err
	}
	message := fmt.Sprintf("%s (%s) has created an order no: %s", req.Email, req.FullName, orderTicketGroup.OrderNo)
	err = h.notificationService.CreateNotification(
		req.FullName,
		"customer",
		"Order",
		"Customer create order",
		message,
		malaysiaTime,
	)
	if err != nil {
		return err
	}

	err = h.customerService.CreateCustomerLog("purchase", "Purchase Completed", "Ticket package purchased via Online", cust)
	if err != nil {
		return err
	}

	malaysiaTime, err = utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	if err != nil {
		return err
	}
	message = fmt.Sprintf("%s (%s) has completed their purchase for order no: %s", req.Email, req.FullName, orderTicketGroup.OrderNo)
	err = h.notificationService.CreateNotification(
		req.FullName,
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
	_, orderItems, ticketInfos, err := h.paymentService.PostToZooAPI(orderTicketGroup.OrderNo)
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
		message := fmt.Sprintf("%s has been assigned with qr codes for their tickets", orderTicketGroup.OrderNo)
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

	ticketGroup, err := h.ticketGroupService.GetTicketGroup(orderTicketGroup.TicketGroupId)
	if err != nil {
		log.Printf("Error finding ticket group %s: %v", orderTicketGroup.TicketGroupId, err)
	}

	total := utils.CalculateOrderTotal(orderItems)

	orderOverview := email.OrderOverview{
		TicketGroup:  ticketGroup.GroupName,
		FullName:     orderTicketGroup.BuyerName,
		PurchaseDate: orderTicketGroup.TransactionDate,
		EntryDate:    orderItems[0].EntryDate,
		Quantity:     len(orderItems),
		OrderNumber:  orderTicketGroup.OrderNo,
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

	err = h.emailService.SendTicketsEmail(orderTicketGroup.BuyerEmail, orderOverview, orderItems, ticketInfos, []email.Attachment{pdfAttachment})
	if err != nil {
		log.Printf("Failed to send tickets email to %s: %v", orderTicketGroup.BuyerEmail, err)
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
		message := fmt.Sprintf("%s has sent out the email to the customer", orderTicketGroup.OrderNo)
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

		orderTicketGroup.IsEmailSent = true
		// Save the updated order
		err = h.paymentService.UpdateOrderTicketGroup(orderTicketGroup)
	}
	if err != nil {
		return err
	}

	// Return success response with redirect information
	return c.Status(fiber.StatusOK).JSON(models.NewBaseSuccessResponse(map[string]interface{}{
		"orderTicketGroupId": orderTicketGroup.OrderTicketGroupId,
		"transactionStatus":  orderTicketGroup.TransactionStatus,
		"orderNo":            orderTicketGroup.OrderNo,
	}))

	//successURL := h.paymentConfig.FrontendBaseURL + "/paymentRedirect"
	//
	//// Build the full URL with query parameters
	//redirectURL := fmt.Sprintf("%s?orderTicketGroupId=%s&transactionStatus=%s&orderNo=%s",
	//	successURL,
	//	url.QueryEscape(strconv.Itoa(int(orderTicketGroup.OrderTicketGroupId))),
	//	url.QueryEscape(orderTicketGroup.TransactionStatus),
	//	url.QueryEscape(orderTicketGroup.OrderNo))
	//
	//log.Printf("Complete redirect URL: %s", redirectURL)
	//
	//return c.Redirect(redirectURL)
}

func (h *OrderHandler) GetOrderManagement(c *fiber.Ctx) error {
	startDate := c.Query("startDate")
	//if startDate == "" {
	//	return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
	//		"Missing start date", nil,
	//	))
	//}

	endDate := c.Query("endDate")
	//if endDate == "" {
	//	return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
	//		"Missing end date", nil,
	//	))
	//}

	orderNo := c.Query("orderNo")

	orderTicketGroups, err := h.orderService.GetAllOrderTicketGroups("", orderNo, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"Could not get order ticket groups: "+err.Error(), nil,
		))
	}

	return c.JSON(models.NewBaseSuccessResponse(orderTicketGroups))
}

// Generate the checkout URL based on order ID and payment type
func (h *OrderHandler) generateCheckoutURL(orderID uint, paymentType string) string {
	// Base checkout URL
	baseURL := "/checkout.html"

	// Add parameters for the checkout page
	return fmt.Sprintf("%s?orderID=%d&paymentType=%s", baseURL, orderID, paymentType)
}
