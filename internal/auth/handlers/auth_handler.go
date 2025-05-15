// FILE: internal/auth/handlers/auth_handler.go
package handlers

import (
	"j-ticketing/internal/auth/models"
	"j-ticketing/internal/auth/service"

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
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	var tokenResp *models.TokenResponse
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
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	// Save token to database
	userID := c.Locals("userId").(string)
	if err := h.authService.SaveToken(
		userID,
		req.UserType,
		tokenResp.AccessToken,
		tokenResp.RefreshToken,
		c.IP(),
		c.Get("User-Agent"),
	); err != nil {
		// Log error but don't fail the request
		// You might want to handle this differently
	}

	return c.JSON(tokenResp)
}

// RefreshToken handles the token refresh request
func (h *AuthHandler) RefreshToken(c *fiber.Ctx) error {
	// Get refresh token from request
	refreshToken := c.FormValue("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Refresh token is required",
		})
	}

	// Refresh token
	tokenResp, err := h.authService.RefreshToken(refreshToken)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(tokenResp)
}

// Logout handles the logout request
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	// Get refresh token from request
	refreshToken := c.FormValue("refresh_token")
	if refreshToken == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Refresh token is required",
		})
	}

	// Get user ID from context (set by Protected middleware)
	userID := c.Locals("userId").(string)

	// Revoke token
	if err := h.authService.RevokeToken(userID, refreshToken); err != nil {
		// Log error but don't fail the request
		// You might want to handle this differently
	}

	return c.JSON(fiber.Map{
		"message": "Successfully logged out",
	})
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
