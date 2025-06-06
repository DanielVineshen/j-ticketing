// File: j-ticketing/internal/core/handlers/general_handler.go
package handlers

import (
	dto "j-ticketing/internal/core/dto/general"
	service "j-ticketing/internal/core/services"
	"j-ticketing/pkg/models"
	"mime/multipart"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GeneralHandler handles requests for general settings operations
type GeneralHandler struct {
	generalService      *service.GeneralService
	notificationService service.NotificationService
}

// NewGeneralHandler creates a new general handler
func NewGeneralHandler(generalService *service.GeneralService, notificationService service.NotificationService) *GeneralHandler {
	return &GeneralHandler{
		generalService:      generalService,
		notificationService: notificationService,
	}
}

// GetGeneralSettings retrieves the general settings
func (h *GeneralHandler) GetGeneralSettings(c *fiber.Ctx) error {
	settings, err := h.generalService.GetGeneralSettings()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.NewBaseErrorResponse(
			"Failed to retrieve general settings", nil,
		))
	}

	return c.Status(fiber.StatusOK).JSON(models.NewBaseSuccessResponse(settings))
}

// UpdateGeneralSettings updates the general settings
func (h *GeneralHandler) UpdateGeneralSettings(c *fiber.Ctx) error {
	// // Get the admin info from the context (set by auth middleware)
	// adminFullName := c.Locals("fullName").(string)
	// adminRole := c.Locals("role").(string)

	// Parse the multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Failed to parse multipart form", nil,
		))
	}

	// Create request from form data
	request := &dto.UpdateGeneralRequest{
		ContactEmail: getFormValue(form.Value, "contactEmail"),
		TimeZone:     getFormValue(form.Value, "timeZone"),
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	// Get file if provided (optional for update)
	var file *multipart.FileHeader
	files := form.File["attachment"]
	if len(files) > 0 {
		file = files[0]
	}

	// Update general settings through service
	err = h.generalService.UpdateGeneralSettings(request, file)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	// // Create notification
	// malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	// if err != nil {
	// 	return err
	// }
	// message := fmt.Sprintf("%s has updated the general settings", adminFullName)
	// err = h.notificationService.CreateNotification(
	// 	adminFullName,
	// 	adminRole,
	// 	"General Settings",
	// 	"Update general settings",
	// 	message,
	// 	malaysiaTime,
	// )
	// if err != nil {
	// 	return err
	// }

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

// GetGeneralAttachment serves the general attachment from the settings record
func (h *GeneralHandler) GetGeneralAttachment(c *fiber.Ctx) error {
	// Get the content type and file path from the service (no uniqueExtension needed)
	contentType, filePath, uniqueExtension, err := h.generalService.GetImageInfo()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"File not found.", nil,
		))
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open file",
		})
	}
	defer file.Close()

	// Get file info for Last-Modified header
	fileInfo, err := file.Stat()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to get file information",
		})
	}

	// Set response headers for proper caching and content type
	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "inline; filename=\""+uniqueExtension+"\"")
	c.Set("Cache-Control", "public, max-age=86400, must-revalidate") // 24 hours cache
	c.Set("Last-Modified", fileInfo.ModTime().Format(http.TimeFormat))
	c.Set("Expires", time.Now().Add(24*time.Hour).Format(http.TimeFormat))

	// Send the file
	return c.SendFile(filePath)
}

func (h *GeneralHandler) GetPrivacyPolicy(c *fiber.Ctx) error {
	general, err := h.generalService.FindGeneralSettings()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
			"Error getting general:"+err.Error(), nil,
		))
	}

	privacyPolicy := dto.PrivacyPolicyResponse{
		PrivacyPolicyContentBm:       general.PrivacyPolicyContentBm,
		PrivacyPolicyContentEn:       general.PrivacyPolicyContentEn,
		PrivacyPolicyContentCn:       general.PrivacyPolicyContentCn,
		PrivacyPolicyLastUpdatedDate: general.PrivacyPolicyLastUpdatedDate,
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(privacyPolicy))
}

// UpdatePrivacyPolicy updates the privacy policy content
func (h *GeneralHandler) UpdatePrivacyPolicy(c *fiber.Ctx) error {
	// // Get the admin info from the context (set by auth middleware)
	// adminFullName := c.Locals("fullName").(string)
	// adminRole := c.Locals("role").(string)

	var request dto.UpdatePrivacyPolicyRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request body", nil,
		))
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	// Update privacy policy through service
	err := h.generalService.UpdatePrivacyPolicy(&request)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	// // Create notification
	// malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	// if err != nil {
	// 	return err
	// }
	// message := fmt.Sprintf("%s has updated the privacy policy", adminFullName)
	// err = h.notificationService.CreateNotification(
	// 	adminFullName,
	// 	adminRole,
	// 	"Privacy Policy",
	// 	"Update privacy policy",
	// 	message,
	// 	malaysiaTime,
	// )
	// if err != nil {
	// 	return err
	// }

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

func (h *GeneralHandler) GetTermsOfPurchase(c *fiber.Ctx) error {
	general, err := h.generalService.FindGeneralSettings()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
			"Error getting general:"+err.Error(), nil,
		))
	}

	termsOfPurchase := dto.TermsOfPurchaseResponse{
		TermsOfPurchaseContentEn:       general.TermsOfPurchaseContentEn,
		TermsOfPurchaseContentBm:       general.TermsOfPurchaseContentBm,
		TermsOfPurchaseContentCn:       general.TermsOfPurchaseContentCn,
		TermsOfPurchaseLastUpdatedDate: general.TermsOfPurchaseLastUpdatedDate,
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(termsOfPurchase))
}

// UpdateTermsOfPurchase updates the terms of purchase content
func (h *GeneralHandler) UpdateTermsOfPurchase(c *fiber.Ctx) error {
	// // Get the admin info from the context (set by auth middleware)
	// adminFullName := c.Locals("fullName").(string)
	// adminRole := c.Locals("role").(string)

	var request dto.UpdateTermsOfPurchaseRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request body", nil,
		))
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	// Update terms of purchase through service
	err := h.generalService.UpdateTermsOfPurchase(&request)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	// // Create notification
	// malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	// if err != nil {
	// 	return err
	// }
	// message := fmt.Sprintf("%s has updated the terms of purchase", adminFullName)
	// err = h.notificationService.CreateNotification(
	// 	adminFullName,
	// 	adminRole,
	// 	"Terms of Purchase",
	// 	"Update terms of purchase",
	// 	message,
	// 	malaysiaTime,
	// )
	// if err != nil {
	// 	return err
	// }

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

func (h *GeneralHandler) GetTermsOfService(c *fiber.Ctx) error {
	general, err := h.generalService.FindGeneralSettings()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
			"Error getting general:"+err.Error(), nil,
		))
	}

	termsOfService := dto.TermsOfServiceResponse{
		TermsOfServiceContentEn:       general.TermsOfServiceContentEn,
		TermsOfServiceContentBm:       general.TermsOfServiceContentBm,
		TermsOfServiceContentCn:       general.TermsOfServiceContentCn,
		TermsOfServiceLastUpdatedDate: general.TermsOfServiceLastUpdatedDate,
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(termsOfService))
}

// UpdateTermsOfService updates the terms of service content
func (h *GeneralHandler) UpdateTermsOfService(c *fiber.Ctx) error {
	// // Get the admin info from the context (set by auth middleware)
	// adminFullName := c.Locals("fullName").(string)
	// adminRole := c.Locals("role").(string)

	var request dto.UpdateTermsOfServiceRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request body", nil,
		))
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	// Update terms of service through service
	err := h.generalService.UpdateTermsOfService(&request)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	// // Create notification
	// malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	// if err != nil {
	// 	return err
	// }
	// message := fmt.Sprintf("%s has updated the terms of service", adminFullName)
	// err = h.notificationService.CreateNotification(
	// 	adminFullName,
	// 	adminRole,
	// 	"Terms of Service",
	// 	"Update terms of service",
	// 	message,
	// 	malaysiaTime,
	// )
	// if err != nil {
	// 	return err
	// }

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

func (h *GeneralHandler) GetFaq(c *fiber.Ctx) error {
	general, err := h.generalService.FindGeneralSettings()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
			"Error getting general:"+err.Error(), nil,
		))
	}

	faq := dto.FaqResponse{
		FaqContentEn:       general.FaqContentEn,
		FaqContentBm:       general.FaqContentBm,
		FaqContentCn:       general.FaqContentCn,
		FaqLastUpdatedDate: general.FaqLastUpdatedDate,
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(faq))
}

// UpdateFaq updates the FAQ content
func (h *GeneralHandler) UpdateFaq(c *fiber.Ctx) error {
	// // Get the admin info from the context (set by auth middleware)
	// adminFullName := c.Locals("fullName").(string)
	// adminRole := c.Locals("role").(string)

	var request dto.UpdateFaqRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request body", nil,
		))
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	// Update FAQ through service
	err := h.generalService.UpdateFaq(&request)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	// // Create notification
	// malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	// if err != nil {
	// 	return err
	// }
	// message := fmt.Sprintf("%s has updated the FAQ", adminFullName)
	// err = h.notificationService.CreateNotification(
	// 	adminFullName,
	// 	adminRole,
	// 	"FAQ",
	// 	"Update FAQ",
	// 	message,
	// 	malaysiaTime,
	// )
	// if err != nil {
	// 	return err
	// }

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

func (h *GeneralHandler) GetContactUs(c *fiber.Ctx) error {
	general, err := h.generalService.FindGeneralSettings()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
			"Error getting general:"+err.Error(), nil,
		))
	}

	contactUs := dto.ContactUsResponse{
		ContactUsContentEn:       general.ContactUsContentEn,
		ContactUsContentBm:       general.ContactUsContentBm,
		ContactUsContentCn:       general.ContactUsContentCn,
		ContactUsLastUpdatedDate: general.ContactUsLastUpdatedDate,
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(contactUs))
}

// UpdateContactUs updates the contact us content
func (h *GeneralHandler) UpdateContactUs(c *fiber.Ctx) error {
	// // Get the admin info from the context (set by auth middleware)
	// adminFullName := c.Locals("fullName").(string)
	// adminRole := c.Locals("role").(string)

	var request dto.UpdateContactUsRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request body", nil,
		))
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	// Update contact us through service
	err := h.generalService.UpdateContactUs(&request)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	// // Create notification
	// malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	// if err != nil {
	// 	return err
	// }
	// message := fmt.Sprintf("%s has updated the contact us content", adminFullName)
	// err = h.notificationService.CreateNotification(
	// 	adminFullName,
	// 	adminRole,
	// 	"Contact Us",
	// 	"Update contact us content",
	// 	message,
	// 	malaysiaTime,
	// )
	// if err != nil {
	// 	return err
	// }

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

func (h *GeneralHandler) GetRefundPolicy(c *fiber.Ctx) error {
	general, err := h.generalService.FindGeneralSettings()
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(models.NewBaseErrorResponse(
			"Error getting general:"+err.Error(), nil,
		))
	}

	refundPolicy := dto.RefundPolicyResponse{
		RefundPolicyContentEn:       general.RefundPolicyContentEn,
		RefundPolicyContentBm:       general.RefundPolicyContentBm,
		RefundPolicyContentCn:       general.RefundPolicyContentCn,
		RefundPolicyLastUpdatedDate: general.RefundPolicyLastUpdatedDate,
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(refundPolicy))
}

// UpdateRefundPolicy updates the refund policy content
func (h *GeneralHandler) UpdateRefundPolicy(c *fiber.Ctx) error {
	// Get the admin info from the context (set by auth middleware)
	// adminFullName := c.Locals("fullName").(string)
	// adminRole := c.Locals("role").(string)

	var request dto.UpdateRefundPolicyRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request body", nil,
		))
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	// Update refund policy through service
	err := h.generalService.UpdateRefundPolicy(&request)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	// Create notification
	// malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	// if err != nil {
	// 	return err
	// }
	// message := fmt.Sprintf("%s has updated the refund policy", adminFullName)
	// err = h.notificationService.CreateNotification(
	// 	adminFullName,
	// 	adminRole,
	// 	"Refund Policy",
	// 	"Update refund policy",
	// 	message,
	// 	malaysiaTime,
	// )
	// if err != nil {
	// 	return err
	// }

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}

func (h *GeneralHandler) UpdateIntegration(c *fiber.Ctx) error {
	var request dto.UpdateIntegrationConfigRequest
	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			"Invalid request body", nil,
		))
	}

	// Validate request
	if err := request.Validate(); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(err)
	}

	// Update integration through service
	err := h.generalService.UpdateIntegrationConfig(&request)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.NewBaseErrorResponse(
			err.Error(), nil,
		))
	}

	return c.Status(fiber.StatusCreated).JSON(models.NewBaseSuccessResponse(models.NewGenericMessage(true)))
}
