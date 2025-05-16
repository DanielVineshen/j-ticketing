// FILE: internal/auth/handlers/admin_handler.go
package handlers

import (
	service "j-ticketing/internal/core/services"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// AdminHandler handles admin-related HTTP requests
type AdminHandler struct {
	adminService service.AdminService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(adminService service.AdminService) *AdminHandler {
	return &AdminHandler{
		adminService: adminService,
	}
}

// CreateAdmin handles admin creation
func (h *AdminHandler) CreateAdmin(c *fiber.Ctx) error {
	// Only SYSADMIN role can create admins
	if c.Locals("role") != "SYSADMIN" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Only system administrators can create admin accounts",
		})
	}

	// Parse request
	type CreateAdminRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
		FullName string `json:"fullName"`
		Role     string `json:"role"`
	}

	var req CreateAdminRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Validate request
	if req.Username == "" || req.Password == "" || req.FullName == "" || req.Role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "All fields are required",
		})
	}

	// Validate role
	validRoles := map[string]bool{
		"SYSADMIN": true,
		"OWNER":    true,
		"STAFF":    true,
	}
	if !validRoles[req.Role] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid role. Must be one of: SYSADMIN, OWNER, STAFF",
		})
	}

	// Create admin
	admin, err := h.adminService.CreateAdmin(req.Username, req.Password, req.FullName, req.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	// Remove password from response
	admin.Password = ""

	return c.Status(fiber.StatusCreated).JSON(admin)
}

// GetAdmin handles getting an admin by ID
func (h *AdminHandler) GetAdmin(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid admin ID",
		})
	}

	admin, err := h.adminService.GetAdminByID(uint(id))
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "Admin not found",
		})
	}

	// Remove password from response
	admin.Password = ""

	return c.JSON(admin)
}

// UpdateAdmin handles updating an admin
func (h *AdminHandler) UpdateAdmin(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid admin ID",
		})
	}

	// Parse request
	type UpdateAdminRequest struct {
		FullName string `json:"fullName"`
		Role     string `json:"role"`
	}

	var req UpdateAdminRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid request body",
		})
	}

	// Validate request
	if req.FullName == "" || req.Role == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "All fields are required",
		})
	}

	// Validate role
	validRoles := map[string]bool{
		"SYSADMIN": true,
		"OWNER":    true,
		"STAFF":    true,
	}
	if !validRoles[req.Role] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid role. Must be one of: SYSADMIN, OWNER, STAFF",
		})
	}

	// Update admin
	admin, err := h.adminService.UpdateAdmin(uint(id), req.FullName, req.Role)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	// Remove password from response
	admin.Password = ""

	return c.JSON(admin)
}

// ChangePassword handles changing an admin's password
func (h *AdminHandler) ChangePassword(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid admin ID",
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
	err = h.adminService.ChangePassword(uint(id), req.CurrentPassword, req.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password changed successfully",
	})
}

// DeleteAdmin handles deleting an admin
func (h *AdminHandler) DeleteAdmin(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 32)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"message": "Invalid admin ID",
		})
	}

	// Only SYSADMIN role can delete admins
	if c.Locals("role") != "SYSADMIN" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"message": "Only system administrators can delete admin accounts",
		})
	}

	err = h.adminService.DeleteAdmin(uint(id))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"message": "Admin deleted successfully",
	})
}

// ListAdmins handles listing all admins
func (h *AdminHandler) ListAdmins(c *fiber.Ctx) error {
	admins, err := h.adminService.ListAdmins()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	return c.JSON(admins)
}
