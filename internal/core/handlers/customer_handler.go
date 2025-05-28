// File: j-ticketing/internal/core/handlers/customer_handlers.go
package handlers

import (
	dto "j-ticketing/internal/core/dto/customer"
	service "j-ticketing/internal/core/services"
	"j-ticketing/pkg/models"

	"github.com/gofiber/fiber/v2"
)

// CustomerHandler handles customer-related HTTP requests
type CustomerHandler struct {
	customerService service.CustomerService
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(customerService service.CustomerService) *CustomerHandler {
	return &CustomerHandler{
		customerService: customerService,
	}
}

// GetCustomer handles getting a customer by custId
func (h *CustomerHandler) GetCustomer(c *fiber.Ctx) error {
	custId := c.Params("custId")
	if custId == "" {
		// Try to get from query parameter if not in path
		custId = c.Query("custId")
		if custId == "" {
			return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
				"Missing custId parameter", nil,
			))
		}
	}

	customer, err := h.customerService.GetCustomerByID(custId)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
			"Customer does not exist", nil,
		))
	}

	// Create the response DTO
	response := dto.CustomerResponse{
		Customer: dto.Customer{
			CustID:           customer.CustId,
			Email:            customer.Email,
			FullName:         customer.FullName,
			IdentificationNo: customer.IdentificationNo,
			IsDisabled:       customer.IsDisabled,
			ContactNo:        customer.ContactNo.String,
		},
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}

// UpdateCustomer handles updating a customer
func (h *CustomerHandler) UpdateCustomer(c *fiber.Ctx) error {
	// Parse request
	var req dto.UpdateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request body", nil,
		))
	}

	// Validate request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	// Get customer ID from token
	userID := c.Locals("userId").(string)

	// Update customer
	_, err := h.customerService.UpdateCustomer(userID, req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"Internal server Error: "+err.Error(), nil,
		))
	}

	return c.JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

// ChangePassword handles changing a customer's password
func (h *CustomerHandler) ChangePassword(c *fiber.Ctx) error {

	var req dto.ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request body", nil,
		))
	}

	// Get customer ID from token
	userID := c.Locals("userId").(string)

	// Change password
	err := h.customerService.ChangePassword(userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	return c.JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

// ListCustomers handles listing all customers (admin only)
func (h *CustomerHandler) ListCustomers(c *fiber.Ctx) error {
	// Only SYSADMIN and OWNER roles can list customers
	userRole := c.Locals("role").(string)
	if userRole != "SYSADMIN" && userRole != "OWNER" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "You are not authorized to list customer accounts",
		})
	}

	customers, err := h.customerService.ListCustomers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(customers)
}
