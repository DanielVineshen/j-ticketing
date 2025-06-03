// File: j-ticketing/internal/core/handlers/auth_handlers.go
package handlers

import (
	"fmt"
	dto "j-ticketing/internal/core/dto/auth"
	service "j-ticketing/internal/core/services"
	"j-ticketing/pkg/email"
	"j-ticketing/pkg/models"
	"j-ticketing/pkg/utils"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles authentication related HTTP requests
type AuthHandler struct {
	authService         service.AuthService
	emailService        email.EmailService
	customerService     service.CustomerService
	notificationService service.NotificationService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService service.AuthService, emailService email.EmailService, customerService service.CustomerService, notificationService service.NotificationService) *AuthHandler {
	return &AuthHandler{
		authService:         authService,
		emailService:        emailService,
		customerService:     customerService,
		notificationService: notificationService,
	}
}

// Login handles the login request
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	// Parse login request
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request body", nil,
		))
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		// For now, just use a simple error message
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Validation failed: "+err.Error(), nil,
		))
	}

	// Set default user type if not provided
	if req.UserType == "" {
		req.UserType = "admin" // Default to admin if not specified
	}

	log.Printf("Login attempt: username=%s, userType=%s", req.Username, req.UserType)

	var tokenResp *dto.TokenResponse
	var err error

	// Handle login based on user type
	switch req.UserType {
	case "admin":
		tokenResp, err = h.authService.LoginAdmin(req.Username, req.Password)
	case "customer":
		tokenResp, err = h.authService.LoginCustomer(req.Username, req.Password)
	default:
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid user type", nil,
		))
	}

	if err != nil {
		log.Printf("Login failed: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	// Don't try to get userID from context here - it won't be set before authentication
	// Use the username from the request instead
	userID := req.Username

	// Save token to database
	if err := h.authService.SaveToken(
		userID,
		req.UserType,
		tokenResp.AccessToken,
		tokenResp.RefreshToken,
		c.IP(),
		c.Get("User-Agent"),
	); err != nil {
		log.Printf("Failed to save token: %v", err)
		// Log error but don't fail the request
	}

	return c.JSON(models.NewBaseSuccessResponse(tokenResp))
}

// RefreshToken handles the token refresh request
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || len(authHeader) < 8 || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Refresh token is required", nil,
		))
	}

	// Extract token from header (remove 'Bearer ' prefix)
	refreshToken := strings.TrimPrefix(authHeader, "Bearer ")

	// Refresh token
	tokenResp, err := h.authService.RefreshToken(refreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	return c.JSON(models.NewBaseSuccessResponse(tokenResp))
}

// Logout handles the logout request
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || len(authHeader) < 8 || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Access token is required", nil,
		))
	}

	// Extract token from header (remove 'Bearer ' prefix)
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")

	// Get user ID from context (set by Protected middleware)
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewBaseErrorResponse(
			"User not authenticated", nil,
		))
	}

	// Revoke token
	if err := h.authService.RevokeToken(username, accessToken); err != nil {
		log.Printf("Failed to revoke token: %v", err)
	}

	return c.JSON(models.NewBaseSuccessResponse(map[string]bool{
		"success": true,
	}))
}

// ValidateToken handles token validation (mostly for testing)
func (h *AuthHandler) ValidateToken(c *fiber.Ctx) error {
	// Get token from authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid authorization header", nil,
		))
	}

	// Extract token
	token := authHeader[7:]

	// Validate token
	isValid, err := h.authService.ValidateToken(token)
	if err != nil || !isValid {
		return c.Status(fiber.StatusUnauthorized).JSON(models.NewBaseErrorResponse(
			"Invalid or expired token", nil,
		))
	}

	return c.JSON(models.NewBaseSuccessResponse(map[string]bool{
		"valid": true,
	}))
}

// CreateCustomer handles creating a new customer
func (h *AuthHandler) CreateCustomer(c *fiber.Ctx) error {
	// Parse create customer request
	var req dto.CreateCustomerRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request format", nil,
		))
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Validation failed: "+err.Error(), nil,
		))
	}

	// Create customer
	customer, err := h.authService.CreateCustomer(&req)
	if err != nil {
		log.Printf("Failed to create customer: %v", err)

		// Check for specific errors
		if err.Error() == "email already exists" {
			return c.Status(fiber.StatusConflict).JSON(models.NewBaseErrorResponse(
				"A customer with this email already exists", nil,
			))
		}

		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"Failed to create customer: "+err.Error(), nil,
		))
	}

	err = h.customerService.CreateCustomerLog("account", "Member Joined", "Customer registered as a new member", *customer)
	if err != nil {
		return err
	}

	malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	if err != nil {
		return err
	}
	message := fmt.Sprintf("%s (%s) has created an account", customer.Email, customer.FullName)
	err = h.notificationService.CreateNotification(
		"system",
		"system",
		"Customer",
		"New customer created",
		message,
		malaysiaTime,
	)
	if err != nil {
		return err
	}

	// Create a response object that doesn't include sensitive data
	response := map[string]interface{}{
		"custId":           customer.CustId,
		"email":            customer.Email,
		"identificationNo": customer.IdentificationNo,
		"fullName":         customer.FullName,
		"contactNo":        customer.ContactNo.String,
		"isDisabled":       customer.IsDisabled,
		"createdAt":        customer.CreatedAt,
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(response))
}

// ResetCustomerPassword handles resetting a customer's password
func (h *AuthHandler) ResetCustomerPassword(c *fiber.Ctx) error {
	// Parse reset password request
	var req dto.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request format", nil,
		))
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Validation failed: "+err.Error(), nil,
		))
	}

	// Reset password
	customer, _, err := h.authService.ResetCustomerPassword(req.Email)
	if err != nil {
		log.Printf("Failed to reset password: %v", err)
		// Don't expose specific error details to the client (security measure)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"An error occurred while processing your request", nil,
		))
	}

	err = h.customerService.CreateCustomerLog("account", "Member Reset Password", "Customer reset their password", *customer)
	if err != nil {
		return err
	}

	malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	if err != nil {
		return err
	}
	message := fmt.Sprintf("%s (%s) has reset their password", customer.Email, customer.FullName)
	err = h.notificationService.CreateNotification(
		customer.FullName,
		"CUSTOMER",
		"Password reset",
		"Customer reset password email was sent",
		message,
		malaysiaTime,
	)
	if err != nil {
		return err
	}

	// Always return success (security measure)
	return c.JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

// ResetAdminPassword handles resetting a admin's password
func (h *AuthHandler) ResetAdminPassword(c *fiber.Ctx) error {
	// Parse reset password request
	var req dto.ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request format", nil,
		))
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Validation failed: "+err.Error(), nil,
		))
	}

	// Reset password
	admin, _, err := h.authService.ResetAdminPassword(req.Email)
	if err != nil {
		log.Printf("Failed to reset password: %v", err)
		// Don't expose specific error details to the client (security measure)
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"An error occurred while processing your request", nil,
		))
	}

	malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	if err != nil {
		return err
	}
	message := fmt.Sprintf("%s (%s) has reset their password", admin.Email, admin.FullName)
	err = h.notificationService.CreateNotification(
		admin.FullName,
		admin.Role,
		"Password reset",
		"Admin reset password email was sent",
		message,
		malaysiaTime,
	)
	if err != nil {
		return err
	}

	// Always return success (security measure)
	return c.JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}
