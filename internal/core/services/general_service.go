// File: j-ticketing/internal/core/services/general_service.go
package service

import (
	"errors"
	dto "j-ticketing/internal/core/dto/general"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/utils"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

// GeneralService handles operations related to general settings management
type GeneralService struct {
	generalRepo *repositories.GeneralRepository
	fileUtil    *utils.FileUtil
}

// NewGeneralService creates a new general service
func NewGeneralService(generalRepo *repositories.GeneralRepository) *GeneralService {
	return &GeneralService{
		generalRepo: generalRepo,
		fileUtil:    utils.NewFileUtil(),
	}
}

// GetGeneralSettings retrieves the general settings
func (s *GeneralService) GetGeneralSettings() (*dto.GeneralResponse, error) {
	generalModel, err := s.generalRepo.FindFirst()
	if err != nil {
		return nil, errors.New("general settings not found")
	}

	// Convert timestamps to Malaysia time with yyyy-MM-dd HH:mm:ss format
	formattedCreatedAt := s.formatTimestampToMalaysia(generalModel.CreatedAt)
	formattedUpdatedAt := s.formatTimestampToMalaysia(generalModel.UpdatedAt)

	response := &dto.GeneralResponse{
		GeneralId:                      generalModel.GeneralId,
		ContactEmail:                   generalModel.ContactEmail,
		TimeZone:                       generalModel.TimeZone,
		EnvKey:                         generalModel.EnvKey,
		EnvValue:                       generalModel.EnvValue,
		AttachmentName:                 generalModel.AttachmentName,
		ContentType:                    generalModel.ContentType,
		UniqueExtension:                generalModel.UniqueExtension,
		PrivacyPolicyContentEn:         generalModel.PrivacyPolicyContentEn,
		PrivacyPolicyContentBm:         generalModel.PrivacyPolicyContentBm,
		PrivacyPolicyContentCn:         generalModel.PrivacyPolicyContentCn,
		PrivacyPolicyLastUpdatedDate:   generalModel.PrivacyPolicyLastUpdatedDate,
		TermsOfPurchaseContentEn:       generalModel.TermsOfPurchaseContentEn,
		TermsOfPurchaseContentBm:       generalModel.TermsOfPurchaseContentBm,
		TermsOfPurchaseContentCn:       generalModel.TermsOfPurchaseContentCn,
		TermsOfPurchaseLastUpdatedDate: generalModel.TermsOfPurchaseLastUpdatedDate,
		TermsOfServiceContentEn:        generalModel.TermsOfServiceContentEn,
		TermsOfServiceContentBm:        generalModel.TermsOfServiceContentBm,
		TermsOfServiceContentCn:        generalModel.TermsOfServiceContentCn,
		TermsOfServiceLastUpdatedDate:  generalModel.TermsOfServiceLastUpdatedDate,
		FaqContentEn:                   generalModel.FaqContentEn,
		FaqContentBm:                   generalModel.FaqContentBm,
		FaqContentCn:                   generalModel.FaqContentCn,
		FaqLastUpdatedDate:             generalModel.FaqLastUpdatedDate,
		ContactUsContentEn:             generalModel.ContactUsContentEn,
		ContactUsContentBm:             generalModel.ContactUsContentBm,
		ContactUsContentCn:             generalModel.ContactUsContentCn,
		ContactUsLastUpdatedDate:       generalModel.ContactUsLastUpdatedDate,
		RefundPolicyContentEn:          generalModel.RefundPolicyContentEn,
		RefundPolicyContentBm:          generalModel.RefundPolicyContentBm,
		RefundPolicyContentCn:          generalModel.RefundPolicyContentCn,
		RefundPolicyLastUpdatedDate:    generalModel.RefundPolicyLastUpdatedDate,
		ZooApiBaseUrl:                  generalModel.ZooApiBaseUrl,
		ZooQrEndpoint:                  generalModel.ZooQrEndpoint,
		ZooTokenEndpoint:               generalModel.ZooTokenEndpoint,
		ZooApiUsername:                 generalModel.ZooApiUsername,
		ZooApiPassword:                 generalModel.ZooApiPassword,
		JpGatewayUrl:                   generalModel.JpGatewayUrl,
		JpPaymentEndpoint:              generalModel.JpPaymentEndpoint,
		JpRedflowEndpoint:              generalModel.JpRedflowEndpoint,
		JpBankListEndpoint:             generalModel.JpBankListEndpoint,
		JpApiKey:                       generalModel.JpApiKey,
		JpAgToken:                      generalModel.JpAgToken,
		EmailHost:                      generalModel.EmailHost,
		EmailPort:                      generalModel.EmailPort,
		EmailUsername:                  generalModel.EmailUsername,
		EmailPassword:                  generalModel.EmailPassword,
		EmailFrom:                      generalModel.EmailFrom,
		EmailUseSsl:                    generalModel.EmailUseSsl,
		EmailClientId:                  generalModel.EmailClientId,
		EmailClientSecret:              generalModel.EmailClientSecret,
		EmailRefreshToken:              generalModel.EmailRefreshToken,
		CreatedAt:                      formattedCreatedAt,
		UpdatedAt:                      formattedUpdatedAt,
	}

	return response, nil
}

// UpdateGeneralSettings updates the general settings
func (s *GeneralService) UpdateGeneralSettings(request *dto.UpdateGeneralRequest, file *multipart.FileHeader) error {
	// Find existing general settings
	existingGeneral, err := s.generalRepo.FindFirst()
	if err != nil {
		return errors.New("general settings not found")
	}

	var uniqueFileName string
	var oldUniqueFileName string

	// Handle file upload if new file is provided
	if file != nil {
		// Get storage path
		storagePath := os.Getenv("GENERAL_STORAGE_PATH")
		if storagePath == "" {
			return errors.New("GENERAL_STORAGE_PATH environment variable not set")
		}

		// Upload new file
		uniqueFileName, err = s.fileUtil.UploadAttachmentFile(file, storagePath)
		if err != nil {
			return err
		}
		oldUniqueFileName = existingGeneral.UniqueExtension

		// Update file-related fields
		existingGeneral.AttachmentName = file.Filename
		existingGeneral.AttachmentSize = file.Size
		existingGeneral.ContentType = file.Header.Get("Content-Type")
		existingGeneral.UniqueExtension = uniqueFileName
		existingGeneral.AttachmentPath = storagePath
	}

	// Update other fields
	existingGeneral.ContactEmail = request.ContactEmail
	existingGeneral.TimeZone = request.TimeZone
	existingGeneral.UpdatedAt = time.Now()

	// Save to database
	if err := s.generalRepo.Update(existingGeneral); err != nil {
		// Delete new file if database update fails
		if uniqueFileName != "" {
			storagePath := os.Getenv("GENERAL_STORAGE_PATH")
			if storagePath != "" {
				s.fileUtil.DeleteAttachmentFile(uniqueFileName, storagePath)
			}
		}
		return err
	}

	// Delete old file if new file was uploaded successfully
	if oldUniqueFileName != "" && uniqueFileName != "" {
		storagePath := os.Getenv("GENERAL_STORAGE_PATH")
		if storagePath != "" {
			s.fileUtil.DeleteAttachmentFile(oldUniqueFileName, storagePath)
		}
	}

	return nil
}

// GetImageInfo retrieves information about an attachment based on its unique extension
func (s *GeneralService) GetImageInfo(uniqueExtension string) (string, string, error) {
	// Get content type from general repository
	contentType, err := s.generalRepo.GetContentTypeByUniqueExtension(uniqueExtension)
	if err != nil {
		return "", "", err
	}

	if contentType == "" {
		return "", "", errors.New("content type not found")
	}

	// Get storage path from environment variable
	storagePath := os.Getenv("GENERAL_STORAGE_PATH")

	// Validate that the file exists
	filePath := filepath.Join(storagePath, uniqueExtension)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", "", err
	}

	return contentType, filePath, nil
}

// formatTimestampToMalaysia converts UTC timestamp to Malaysia time in yyyy-MM-dd HH:mm:ss format
func (s *GeneralService) formatTimestampToMalaysia(utcTime time.Time) string {
	// Convert to Malaysia time and format
	formattedTime, err := utils.FormatToMalaysiaTime(utcTime, utils.FullDateTimeFormat)
	if err != nil {
		// Fallback to UTC time if conversion fails
		return utcTime.Format(utils.FullDateTimeFormat)
	}
	return formattedTime
}

// UpdatePrivacyPolicy updates the privacy policy content
func (s *GeneralService) UpdatePrivacyPolicy(request *dto.UpdatePrivacyPolicyRequest) error {
	// Find existing general settings
	existingGeneral, err := s.generalRepo.FindFirst()
	if err != nil {
		return errors.New("general settings not found")
	}

	// Update privacy policy fields
	existingGeneral.PrivacyPolicyContentEn = request.PrivacyPolicyContentEn
	existingGeneral.PrivacyPolicyContentBm = request.PrivacyPolicyContentBm
	existingGeneral.PrivacyPolicyContentCn = request.PrivacyPolicyContentCn
	existingGeneral.PrivacyPolicyLastUpdatedDate = request.PrivacyPolicyLastUpdatedDate
	existingGeneral.UpdatedAt = time.Now()

	// Save to database
	return s.generalRepo.Update(existingGeneral)
}

// UpdateTermsOfPurchase updates the terms of purchase content
func (s *GeneralService) UpdateTermsOfPurchase(request *dto.UpdateTermsOfPurchaseRequest) error {
	// Find existing general settings
	existingGeneral, err := s.generalRepo.FindFirst()
	if err != nil {
		return errors.New("general settings not found")
	}

	// Update terms of purchase fields
	existingGeneral.TermsOfPurchaseContentEn = request.TermsOfPurchaseContentEn
	existingGeneral.TermsOfPurchaseContentBm = request.TermsOfPurchaseContentBm
	existingGeneral.TermsOfPurchaseContentCn = request.TermsOfPurchaseContentCn
	existingGeneral.TermsOfPurchaseLastUpdatedDate = request.TermsOfPurchaseLastUpdatedDate
	existingGeneral.UpdatedAt = time.Now()

	// Save to database
	return s.generalRepo.Update(existingGeneral)
}

// UpdateTermsOfService updates the terms of service content
func (s *GeneralService) UpdateTermsOfService(request *dto.UpdateTermsOfServiceRequest) error {
	// Find existing general settings
	existingGeneral, err := s.generalRepo.FindFirst()
	if err != nil {
		return errors.New("general settings not found")
	}

	// Update terms of service fields
	existingGeneral.TermsOfServiceContentEn = request.TermsOfServiceContentEn
	existingGeneral.TermsOfServiceContentBm = request.TermsOfServiceContentBm
	existingGeneral.TermsOfServiceContentCn = request.TermsOfServiceContentCn
	existingGeneral.TermsOfServiceLastUpdatedDate = request.TermsOfServiceLastUpdatedDate
	existingGeneral.UpdatedAt = time.Now()

	// Save to database
	return s.generalRepo.Update(existingGeneral)
}

// UpdateFaq updates the FAQ content
func (s *GeneralService) UpdateFaq(request *dto.UpdateFaqRequest) error {
	// Find existing general settings
	existingGeneral, err := s.generalRepo.FindFirst()
	if err != nil {
		return errors.New("general settings not found")
	}

	// Update FAQ fields
	existingGeneral.FaqContentEn = request.FaqContentEn
	existingGeneral.FaqContentBm = request.FaqContentBm
	existingGeneral.FaqContentCn = request.FaqContentCn
	existingGeneral.FaqLastUpdatedDate = request.FaqLastUpdatedDate
	existingGeneral.UpdatedAt = time.Now()

	// Save to database
	return s.generalRepo.Update(existingGeneral)
}

// UpdateContactUs updates the contact us content
func (s *GeneralService) UpdateContactUs(request *dto.UpdateContactUsRequest) error {
	// Find existing general settings
	existingGeneral, err := s.generalRepo.FindFirst()
	if err != nil {
		return errors.New("general settings not found")
	}

	// Update contact us fields
	existingGeneral.ContactUsContentEn = request.ContactUsContentEn
	existingGeneral.ContactUsContentBm = request.ContactUsContentBm
	existingGeneral.ContactUsContentCn = request.ContactUsContentCn
	existingGeneral.ContactUsLastUpdatedDate = request.ContactUsLastUpdatedDate
	existingGeneral.UpdatedAt = time.Now()

	// Save to database
	return s.generalRepo.Update(existingGeneral)
}

// UpdateRefundPolicy updates the refund policy content
func (s *GeneralService) UpdateRefundPolicy(request *dto.UpdateRefundPolicyRequest) error {
	// Find existing general settings
	existingGeneral, err := s.generalRepo.FindFirst()
	if err != nil {
		return errors.New("general settings not found")
	}

	// Update refund policy fields
	existingGeneral.RefundPolicyContentEn = request.RefundPolicyContentEn
	existingGeneral.RefundPolicyContentBm = request.RefundPolicyContentBm
	existingGeneral.RefundPolicyContentCn = request.RefundPolicyContentCn
	existingGeneral.RefundPolicyLastUpdatedDate = request.RefundPolicyLastUpdatedDate
	existingGeneral.UpdatedAt = time.Now()

	// Save to database
	return s.generalRepo.Update(existingGeneral)
}

func (s *GeneralService) UpdateIntegrationConfig(request *dto.UpdateIntegrationConfigRequest) error {
	// Find existing general settings
	existingGeneral, err := s.generalRepo.FindFirst()
	if err != nil {
		return errors.New("general settings not found")
	}

	// Update integration fields
	existingGeneral.ZooApiBaseUrl = request.ZooApiBaseUrl
	existingGeneral.ZooQrEndpoint = request.ZooQrEndpoint
	existingGeneral.ZooTokenEndpoint = request.ZooTokenEndpoint
	existingGeneral.ZooApiUsername = request.ZooApiUsername
	existingGeneral.ZooApiPassword = request.ZooApiPassword
	existingGeneral.JpGatewayUrl = request.JpGatewayUrl
	existingGeneral.JpPaymentEndpoint = request.JpPaymentEndpoint
	existingGeneral.JpRedflowEndpoint = request.JpRedflowEndpoint
	existingGeneral.JpBankListEndpoint = request.JpBankListEndpoint
	existingGeneral.JpApiKey = request.JpApiKey
	existingGeneral.JpAgToken = request.JpAgToken
	existingGeneral.EmailUsername = request.EmailUsername
	existingGeneral.EmailPassword = request.EmailPassword
	existingGeneral.EmailFrom = request.EmailFrom
	existingGeneral.EmailClientId = request.EmailClientId
	existingGeneral.EmailClientSecret = request.EmailClientSecret
	existingGeneral.EmailRefreshToken = request.EmailRefreshToken
	existingGeneral.UpdatedAt = time.Now()

	// Save to database
	return s.generalRepo.Update(existingGeneral)
}
