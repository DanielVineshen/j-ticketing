// FILE: internal/auth/handlers/auth_handler.go
package handlers

import (
	dto "j-ticketing/internal/core/dto/auth"
	service "j-ticketing/internal/services"
	responseModel "j-ticketing/pkg/models"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// AuthHandler handles authentication related HTTP requests
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login handles the login request
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	// Parse login request
	var req dto.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Validate the request
	if err := req.Validate(); err != nil {
		// For now, just use a simple error message
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Validation failed: " + err.Error(),
		})
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid user type",
		})
	}

	if err != nil {
		log.Printf("Login failed: %v", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": err.Error(),
		})
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

	return c.JSON(responseModel.NewBaseSuccessResponse(tokenResp))
}

// RefreshToken handles the token refresh request
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || len(authHeader) < 8 || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Refresh token is required",
		})
	}

	// Extract token from header (remove 'Bearer ' prefix)
	refreshToken := strings.TrimPrefix(authHeader, "Bearer ")

	// Refresh token
	tokenResp, err := h.authService.RefreshToken(refreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(responseModel.NewBaseSuccessResponse(tokenResp))
}

// Logout handles the logout request
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")
	if authHeader == "" || len(authHeader) < 8 || !strings.HasPrefix(authHeader, "Bearer ") {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Access token is required",
		})
	}

	// Extract token from header (remove 'Bearer ' prefix)
	accessToken := strings.TrimPrefix(authHeader, "Bearer ")

	// Get user ID from context (set by Protected middleware)
	username, ok := c.Locals("username").(string)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "User not authenticated",
		})
	}

	// Revoke token
	if err := h.authService.RevokeToken(username, accessToken); err != nil {
		log.Printf("Failed to revoke token: %v", err)
	}

	return c.JSON(responseModel.NewBaseSuccessResponse(fiber.Map{
		"message": "Successfully logged out",
	}))
}

// ValidateToken handles token validation (mostly for testing)
func (h *AuthHandler) ValidateToken(c *fiber.Ctx) error {
	// Get token from authorization header
	authHeader := c.Get("Authorization")
	if authHeader == "" || len(authHeader) < 8 || authHeader[:7] != "Bearer " {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid authorization header",
		})
	}

	// Extract token
	token := authHeader[7:]

	// Validate token
	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(claims)
}
