// File: j-ticketing/internal/core/handlers/order_handlers.go
package handlers

import (
	orderDto "j-ticketing/internal/core/dto/order"
	services "j-ticketing/internal/core/services"
	"j-ticketing/pkg/errors"
	"j-ticketing/pkg/models"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// OrderHandler handles HTTP requests for orders
type OrderHandler struct {
	orderService *services.OrderService
}

// NewOrderHandler creates a new instance of OrderHandler
func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// GetOrderTicketGroups handles GET requests for order ticket groups
func (h *OrderHandler) GetOrderTicketGroups(c *fiber.Ctx) error {
	// Get the customer ID from the context (set by auth middleware)
	custId, ok := c.Locals("userId").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewBaseErrorResponse(
			errors.USER_NOT_AUTHORIZED.Code, "User not authenticated", nil,
		))
	}

	// Get the user type from the context
	userType, ok := c.Locals("userType").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewBaseErrorResponse(
			errors.USER_NOT_AUTHORIZED.Code, "User type not found", nil,
		))
	}

	// Get the user role from the context
	userRole, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewBaseErrorResponse(
			errors.USER_NOT_AUTHORIZED.Code, "User role not found", nil,
		))
	}

	var response interface{}
	var err error

	// If user is admin, allow fetching all orders or filtering by customer ID
	if userType == "admin" && (userRole == "SYSADMIN" || userRole == "OWNER") {
		// Admin can optionally filter by customer ID
		filterCustId := c.Query("custId")
		response, err = h.orderService.GetAllOrderTicketGroups(filterCustId)
	} else if userType == "customer" {
		// Customer can only see their own orders
		response, err = h.orderService.GetAllOrderTicketGroups(custId)
	} else {
		return c.Status(fiber.StatusForbidden).JSON(models.NewBaseErrorResponse(
			errors.USER_NOT_PERMITTED.Code, "You are not authorized to view these orders", nil,
		))
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			errors.PROCESSING_ERROR.Code, "Failed to retrieve order ticket groups: "+err.Error(), nil,
		))
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}

// GetOrderTicketGroup handles GET request for a specific order ticket group
func (h *OrderHandler) GetOrderTicketGroup(c *fiber.Ctx) error {
	// Get the customer ID from the context (set by auth middleware)
	custId, ok := c.Locals("userId").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewBaseErrorResponse(
			errors.USER_NOT_AUTHORIZED.Code, "User not authenticated", nil,
		))
	}

	// Get the user type from the context
	userType, ok := c.Locals("userType").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewBaseErrorResponse(
			errors.USER_NOT_AUTHORIZED.Code, "User type not found", nil,
		))
	}

	// Get the user role from the context
	userRole, ok := c.Locals("role").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewBaseErrorResponse(
			errors.USER_NOT_AUTHORIZED.Code, "User role not found", nil,
		))
	}

	// Parse the order ticket group ID from the request
	orderTicketGroupIdStr := c.Params("id")
	if orderTicketGroupIdStr == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			errors.INVALID_INPUT_FORMAT.Code, "Missing order ticket group ID", nil,
		))
	}

	orderTicketGroupId, err := strconv.ParseUint(orderTicketGroupIdStr, 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			errors.INVALID_INPUT_VALUES.Code, "Invalid order ticket group ID", nil,
		))
	}

	// Get the order ticket group
	order, err := h.orderService.GetOrderTicketGroup(uint(orderTicketGroupId))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
			errors.ENTITY_NOT_FOUND_EXCEPTION.Code, "Order ticket group not found", nil,
		))
	}

	// Check if the user is authorized to view this order
	if userType == "customer" && order.OrderProfile.CustId != custId {
		return c.Status(fiber.StatusForbidden).JSON(models.NewBaseErrorResponse(
			errors.USER_NOT_PERMITTED.Code, "You are not authorized to view this order", nil,
		))
	}

	// If user is admin, allow access to all orders
	if userType == "admin" && (userRole != "SYSADMIN" && userRole != "OWNER" && userRole != "STAFF") {
		return c.Status(fiber.StatusForbidden).JSON(models.NewBaseErrorResponse(
			errors.USER_NOT_PERMITTED.Code, "You are not authorized to view this order", nil,
		))
	}

	return c.JSON(models.NewBaseSuccessResponse(order))
}

// CreateOrderTicketGroup handles POST requests to create a new order
func (h *OrderHandler) CreateOrderTicketGroup(c *fiber.Ctx) error {
	// Get the customer ID from the context (set by auth middleware)
	custId, ok := c.Locals("userId").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewBaseErrorResponse(
			errors.USER_NOT_AUTHORIZED.Code, "User not authenticated", nil,
		))
	}

	// Get the user type from the context
	userType, ok := c.Locals("userType").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewBaseErrorResponse(
			errors.USER_NOT_AUTHORIZED.Code, "User type not found", nil,
		))
	}

	// Only customers can create orders
	if userType != "customer" {
		return c.Status(fiber.StatusForbidden).JSON(models.NewBaseErrorResponse(
			errors.USER_NOT_PERMITTED.Code, "Only customers can create orders", nil,
		))
	}

	// Parse request body
	var req orderDto.CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			errors.INVALID_INPUT_FORMAT.Code, "Invalid request format", nil,
		))
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			errors.INVALID_INPUT_VALUES.Code, err.Error(), nil,
		))
	}

	// Create the order
	err := h.orderService.CreateOrder(custId, &req)
	if err != nil {
		// Determine appropriate error code based on the error
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
				errors.ENTITY_NOT_FOUND_EXCEPTION.Code, err.Error(), nil,
			))
		} else if strings.Contains(err.Error(), "payment processing failed") {
			return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
				errors.PROCESSING_ERROR.Code, err.Error(), nil,
			))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
				errors.PROCESSING_ERROR.Code, "Failed to create order: "+err.Error(), nil,
			))
		}
	}

	// Return success response
	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(nil))
}
