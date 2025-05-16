// FILE: internal/auth/handlers/customer_handler.go
package handlers

import (
	"database/sql"
	service "j-ticketing/internal/core/services"

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

// RegisterCustomer handles customer registration
func (h *CustomerHandler) RegisterCustomer(c *fiber.Ctx) error {
	// Parse request
	type RegisterCustomerRequest struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		IdentificationNo string `json:"identificationNo"`
		FullName         string `json:"fullName"`
		ContactNo        string `json:"contactNo"`
	}

	var req RegisterCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Validate request
	if req.Email == "" || req.Password == "" || req.FullName == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Email, password, and full name are required",
		})
	}

	// Register customer
	customer, err := h.customerService.RegisterCustomer(
		req.Email,
		req.Password,
		req.IdentificationNo,
		req.FullName,
		req.ContactNo,
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	// Remove password from response
	customer.Password = sql.NullString{
		String: "",
		Valid:  false,
	}

	return c.Status(fiber.StatusCreated).JSON(customer)
}

// GetCustomer handles getting a customer by ID
func (h *CustomerHandler) GetCustomer(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Customer ID is required",
		})
	}

	// Check if requesting user is the customer or an admin
	userID := c.Locals("userId").(string)
	userRole := c.Locals("role").(string)
	if userID != id && userRole != "SYSADMIN" && userRole != "OWNER" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "You are not authorized to view this customer",
		})
	}

	customer, err := h.customerService.GetCustomerByID(id)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Customer not found",
		})
	}

	// Remove password from response
	customer.Password = sql.NullString{
		String: "",
		Valid:  false,
	}

	return c.JSON(customer)
}

// UpdateCustomer handles updating a customer
func (h *CustomerHandler) UpdateCustomer(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Customer ID is required",
		})
	}

	// Check if requesting user is the customer
	userID := c.Locals("userId").(string)
	if userID != id {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "You are not authorized to update this customer",
		})
	}

	// Parse request
	type UpdateCustomerRequest struct {
		FullName  string `json:"fullName"`
		ContactNo string `json:"contactNo"`
	}

	var req UpdateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Update customer
	customer, err := h.customerService.UpdateCustomer(id, req.FullName, req.ContactNo)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	// Remove password from response
	customer.Password = sql.NullString{
		String: "",
		Valid:  false,
	}

	return c.JSON(customer)
}

// ChangePassword handles changing a customer's password
func (h *CustomerHandler) ChangePassword(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Customer ID is required",
		})
	}

	// Check if requesting user is the customer
	userID := c.Locals("userId").(string)
	if userID != id {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "You are not authorized to change this customer's password",
		})
	}

	// Parse request
	type ChangePasswordRequest struct {
		CurrentPassword string `json:"currentPassword"`
		NewPassword     string `json:"newPassword"`
	}

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Validate request
	if req.CurrentPassword == "" || req.NewPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "All fields are required",
		})
	}

	// Change password
	err := h.customerService.ChangePassword(id, req.CurrentPassword, req.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password changed successfully",
	})
}

// DisableCustomer handles disabling a customer's account (admin only)
func (h *CustomerHandler) DisableCustomer(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Customer ID is required",
		})
	}

	// Only SYSADMIN and OWNER roles can disable customers
	userRole := c.Locals("role").(string)
	if userRole != "SYSADMIN" && userRole != "OWNER" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "You are not authorized to disable customer accounts",
		})
	}

	err := h.customerService.DisableCustomer(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Customer account disabled successfully",
	})
}

// EnableCustomer handles enabling a customer's account (admin only)
func (h *CustomerHandler) EnableCustomer(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Customer ID is required",
		})
	}

	// Only SYSADMIN and OWNER roles can enable customers
	userRole := c.Locals("role").(string)
	if userRole != "SYSADMIN" && userRole != "OWNER" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "You are not authorized to enable customer accounts",
		})
	}

	err := h.customerService.EnableCustomer(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Customer account enabled successfully",
	})
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
