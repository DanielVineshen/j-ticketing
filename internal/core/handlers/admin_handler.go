// File: j-ticketing/internal/core/handlers/admin_handler.go
package handlers

import (
	dto "j-ticketing/internal/core/dto/admin"
	service "j-ticketing/internal/core/services"
	"j-ticketing/pkg/models"
	"j-ticketing/pkg/utils"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// AdminHandler handles admin-related HTTP requests
type AdminHandler struct {
	adminService        *service.AdminServiceExtended
	notificationService service.NotificationService
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(adminService *service.AdminServiceExtended, notificationService service.NotificationService) *AdminHandler {
	return &AdminHandler{
		adminService:        adminService,
		notificationService: notificationService,
	}
}

// Profile Management Handlers (for admins managing their own profile)

// GetAdminProfile handles getting an admin's own profile (extracts adminId from JWT token)
func (h *AdminHandler) GetAdminProfile(c *fiber.Ctx) error {
	// Get admin ID from JWT token
	userID := c.Locals("userId").(string)

	admin, err := h.adminService.GetAdminByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
			"Admin profile does not exist", nil,
		))
	}

	// Create the response DTO
	response := dto.AdminProfileResponse{
		Admin: dto.AdminProfile{
			AdminID:   int(admin.AdminId),
			Username:  admin.Username,
			FullName:  admin.FullName,
			Email:     admin.Email,
			ContactNo: admin.ContactNo,
			Role:      admin.Role,
		},
	}

	return c.JSON(models.NewBaseSuccessResponse(response))
}

// UpdateAdminProfile handles updating an admin's own profile
func (h *AdminHandler) UpdateAdminProfile(c *fiber.Ctx) error {
	// Parse request
	var req dto.UpdateAdminProfileRequest
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

	// Update admin profile using adminId from request body
	admin, err := h.adminService.UpdateAdminProfile(req)
	if err != nil {
		if err.Error() == "admin not found" {
			return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
				"Admin not found", nil,
			))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"Internal server Error: "+err.Error(), nil,
		))
	}

	malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	if err != nil {
		return err
	}
	// Create notification for admin profile update
	message := admin.Username + " (" + admin.FullName + ") has updated their profile"
	err = h.notificationService.CreateNotification(
		admin.FullName,
		"admin",
		"Admin",
		"Admin update profile",
		message,
		malaysiaTime,
	)
	if err != nil {
		// Log error but don't fail the request
	}

	return c.JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

// ChangePassword handles changing an admin's password
func (h *AdminHandler) ChangePassword(c *fiber.Ctx) error {
	var req dto.ChangePasswordRequest
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

	// Change password using adminId from request body
	admin, err := h.adminService.ChangePassword(req)
	if err != nil {
		if err.Error() == "admin not found" {
			return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
				"Admin not found", nil,
			))
		}
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	if err != nil {
		return err
	}
	// Create notification for password change
	message := admin.Username + " (" + admin.FullName + ") has changed their password"
	err = h.notificationService.CreateNotification(
		admin.FullName,
		"admin",
		"Admin",
		"Admin changed password",
		message,
		malaysiaTime,
	)
	if err != nil {
		// Log error but don't fail the request
	}

	return c.JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

// Admin Management Handlers (for SYSADMIN managing other admins)

// GetAllAdmins handles getting all admins for management
func (h *AdminHandler) GetAllAdmins(c *fiber.Ctx) error {
	admins, err := h.adminService.GetAllAdmins()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"Could not get admins: "+err.Error(), nil,
		))
	}

	return c.JSON(models.NewBaseSuccessResponse(admins))
}

// CreateAdmin handles creating a new admin
func (h *AdminHandler) CreateAdmin(c *fiber.Ctx) error {
	// Parse request
	var req dto.CreateAdminRequest
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

	// Create admin
	admin, err := h.adminService.CreateAdmin(req)
	if err != nil {
		if err.Error() == "username already exists" {
			return c.Status(fiber.StatusConflict).JSON(models.NewBaseErrorResponse(
				"Username already exists", nil,
			))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"Failed to create admin: "+err.Error(), nil,
		))
	}

	malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	if err != nil {
		return err
	}
	// Create notification for admin creation
	currentUser := c.Locals("username").(string) // Assuming you have username in JWT
	message := "New admin account created: " + admin.Username + " (" + admin.FullName + ") by " + currentUser
	err = h.notificationService.CreateNotification(
		currentUser,
		"admin",
		"Admin",
		"Admin account created",
		message,
		malaysiaTime,
	)
	if err != nil {
		// Log error but don't fail the request
	}

	return c.JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

// UpdateAdminManagement handles updating an admin via management interface
func (h *AdminHandler) UpdateAdminManagement(c *fiber.Ctx) error {
	// Parse request
	var req dto.UpdateAdminManagementRequest
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

	// Update admin
	admin, err := h.adminService.UpdateAdminManagement(req)
	if err != nil {
		if err.Error() == "admin not found" {
			return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
				"Admin not found", nil,
			))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"Failed to update admin: "+err.Error(), nil,
		))
	}

	malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	if err != nil {
		return err
	}
	// Create notification for admin update
	currentUser := c.Locals("username").(string) // Assuming you have username in JWT
	message := "Admin account updated: " + admin.Username + " (" + admin.FullName + ") by " + currentUser
	err = h.notificationService.CreateNotification(
		currentUser,
		"admin",
		"Admin",
		"Admin account updated",
		message,
		malaysiaTime,
	)
	if err != nil {
		// Log error but don't fail the request
	}

	return c.JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

// DeleteAdmin handles deleting an admin
func (h *AdminHandler) DeleteAdmin(c *fiber.Ctx) error {
	// Parse request
	var req dto.DeleteAdminRequest
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

	// Delete admin
	err := h.adminService.DeleteAdmin(req)
	if err != nil {
		if err.Error() == "admin not found" {
			return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
				"Admin not found", nil,
			))
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"Failed to delete admin: "+err.Error(), nil,
		))
	}

	malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	if err != nil {
		return err
	}
	// Create notification for admin deletion
	currentUser := c.Locals("username").(string) // Assuming you have username in JWT
	message := "Admin account deleted (ID: " + strconv.Itoa(int(req.AdminID)) + ") by " + currentUser
	err = h.notificationService.CreateNotification(
		currentUser,
		"admin",
		"Admin",
		"Admin account deleted",
		message,
		malaysiaTime,
	)
	if err != nil {
		// Log error but don't fail the request
	}

	return c.JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}
