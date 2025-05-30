// File: j-ticketing/internal/core/services/banner_service.go
package service

import (
	"errors"
	"fmt"
	dto "j-ticketing/internal/core/dto/banner"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/utils"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"
)

// BannerService handles operations related to banner management
type BannerService struct {
	bannerRepo *repositories.BannerRepository
	fileUtil   *utils.FileUtil
}

// NewBannerService creates a new banner service
func NewBannerService(bannerRepo *repositories.BannerRepository) *BannerService {
	return &BannerService{
		bannerRepo: bannerRepo,
		fileUtil:   utils.NewFileUtil(), // Use the new file util from utils package
	}
}

// GetAllBanners retrieves all banners (no filtering)
func (s *BannerService) GetAllBanners() ([]dto.Banner, error) {
	bannerModels, err := s.bannerRepo.FindAll()
	if err != nil {
		// Return empty array instead of nil on error
		return make([]dto.Banner, 0), nil
	}

	// If no banners found, return empty array (not nil slice)
	if len(bannerModels) == 0 {
		return make([]dto.Banner, 0), nil
	}

	// Initialize with capacity to avoid nil slice
	banners := make([]dto.Banner, 0, len(bannerModels))

	for _, model := range bannerModels {
		// Convert timestamps to Malaysia time with yyyy-MM-dd HH:mm:ss format
		formattedCreatedAt := s.formatTimestampToMalaysia(model.CreatedAt)
		formattedUpdatedAt := s.formatTimestampToMalaysia(model.UpdatedAt)

		banner := dto.Banner{
			BannerId:        model.BannerId,
			Placement:       model.Placement,
			RedirectURL:     model.RedirectURL,
			UploadedBy:      model.UploadedBy,
			ActiveEndDate:   model.ActiveEndDate,   // Keep original yyyy-MM-dd format
			ActiveStartDate: model.ActiveStartDate, // Keep original yyyy-MM-dd format
			IsActive:        model.IsActive,
			Duration:        model.Duration,
			AttachmentName:  model.AttachmentName,
			AttachmentPath:  model.AttachmentPath,
			AttachmentSize:  model.AttachmentSize,
			ContentType:     model.ContentType,
			UniqueExt:       model.UniqueExt,
			CreatedAt:       formattedCreatedAt,
			UpdatedAt:       formattedUpdatedAt,
		}
		banners = append(banners, banner)
	}

	return banners, nil
}

// GetFilteredBanners retrieves only active banners within their active date range
func (s *BannerService) GetFilteredBanners() ([]dto.Banner, error) {
	bannerModels, err := s.bannerRepo.FindAll()
	if err != nil {
		// Return empty array instead of nil on error
		return make([]dto.Banner, 0), nil
	}

	// If no banners found, return empty array
	if len(bannerModels) == 0 {
		return make([]dto.Banner, 0), nil
	}

	// Initialize with capacity
	banners := make([]dto.Banner, 0, len(bannerModels))

	// Get current date in Malaysia timezone for consistency
	currentDate, err := s.getCurrentMalaysiaDate()
	if err != nil {
		// Fallback to UTC date if Malaysia time fails
		currentDate = time.Now().Format("2006-01-02")
	}

	for _, model := range bannerModels {
		// Check if banner is active and within the active date range
		if s.isBannerActiveAndValid(model, currentDate) {
			// Convert timestamps to Malaysia time with yyyy-MM-dd HH:mm:ss format
			formattedCreatedAt := s.formatTimestampToMalaysia(model.CreatedAt)
			formattedUpdatedAt := s.formatTimestampToMalaysia(model.UpdatedAt)

			banner := dto.Banner{
				BannerId:        model.BannerId,
				Placement:       model.Placement,
				RedirectURL:     model.RedirectURL,
				UploadedBy:      model.UploadedBy,
				ActiveEndDate:   model.ActiveEndDate,   // Keep original yyyy-MM-dd format
				ActiveStartDate: model.ActiveStartDate, // Keep original yyyy-MM-dd format
				IsActive:        model.IsActive,
				Duration:        model.Duration,
				AttachmentName:  model.AttachmentName,
				AttachmentPath:  model.AttachmentPath,
				AttachmentSize:  model.AttachmentSize,
				ContentType:     model.ContentType,
				UniqueExt:       model.UniqueExt,
				CreatedAt:       formattedCreatedAt,
				UpdatedAt:       formattedUpdatedAt,
			}
			banners = append(banners, banner)
		}
	}

	// Always return initialized slice (never nil)
	return banners, nil
}

// CreateBanner creates a new banner with file upload
func (s *BannerService) CreateBanner(request *dto.CreateNewBannerRequest, file *multipart.FileHeader) (*dto.Banner, error) {
	// Validate dates (keep original format, just validate logic)
	if err := s.validateDateRange(request.ActiveStartDate, request.ActiveEndDate); err != nil {
		return nil, err
	}

	// Get storage path
	storagePath := os.Getenv("BANNER_STORAGE_PATH")
	if storagePath == "" {
		return nil, errors.New("BANNER_STORAGE_PATH environment variable not set")
	}

	// Validate and upload file
	uniqueFileName, err := s.fileUtil.UploadAttachmentFile(file, storagePath)
	if err != nil {
		return nil, err
	}

	// Get the highest placement number and increment by 1
	maxPlacement, err := s.bannerRepo.GetMaxPlacement()
	if err != nil {
		// If no banners exist, start with placement 1
		maxPlacement = 0
	}

	// Create banner model (keep dates in original yyyy-MM-dd format)
	bannerModel := &models.Banner{
		Placement:       maxPlacement + 1,
		RedirectURL:     request.RedirectURL,
		UploadedBy:      request.UploadedBy,
		ActiveEndDate:   request.ActiveEndDate,   // Keep original yyyy-MM-dd format
		ActiveStartDate: request.ActiveStartDate, // Keep original yyyy-MM-dd format
		IsActive:        request.IsActive,
		Duration:        request.Duration,
		AttachmentName:  file.Filename,
		AttachmentPath:  storagePath,
		AttachmentSize:  file.Size,
		ContentType:     file.Header.Get("Content-Type"),
		UniqueExt:       uniqueFileName,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save to database
	if err := s.bannerRepo.Create(bannerModel); err != nil {
		// Delete uploaded file if database save fails
		s.fileUtil.DeleteAttachmentFile(uniqueFileName, storagePath)
		return nil, err
	}

	// Convert to response DTO
	response := &dto.Banner{
		BannerId:        bannerModel.BannerId,
		Placement:       bannerModel.Placement,
		RedirectURL:     bannerModel.RedirectURL,
		UploadedBy:      bannerModel.UploadedBy,
		ActiveEndDate:   bannerModel.ActiveEndDate,
		ActiveStartDate: bannerModel.ActiveStartDate,
		IsActive:        bannerModel.IsActive,
		Duration:        bannerModel.Duration,
		AttachmentName:  bannerModel.AttachmentName,
		AttachmentPath:  bannerModel.AttachmentPath,
		AttachmentSize:  bannerModel.AttachmentSize,
		ContentType:     bannerModel.ContentType,
		UniqueExt:       bannerModel.UniqueExt,
		CreatedAt:       s.formatTimestampToMalaysia(bannerModel.CreatedAt),
		UpdatedAt:       s.formatTimestampToMalaysia(bannerModel.UpdatedAt),
	}

	return response, nil
}

// UpdateBanner updates an existing banner
func (s *BannerService) UpdateBanner(request *dto.UpdateBannerRequest, file *multipart.FileHeader) (*dto.Banner, error) {
	// Find existing banner
	existingBanner, err := s.bannerRepo.FindByID(request.BannerId)
	if err != nil {
		return nil, errors.New("banner not found")
	}

	// Validate dates (keep original format, just validate logic)
	if err := s.validateDateRange(request.ActiveStartDate, request.ActiveEndDate); err != nil {
		return nil, err
	}

	var uniqueFileName string
	var oldUniqueFileName string

	// Handle file upload if new file is provided
	if file != nil {
		// Get storage path
		storagePath := os.Getenv("BANNER_STORAGE_PATH")
		if storagePath == "" {
			return nil, errors.New("BANNER_STORAGE_PATH environment variable not set")
		}

		// Upload new file
		uniqueFileName, err = s.fileUtil.UploadAttachmentFile(file, storagePath)
		if err != nil {
			return nil, err
		}
		oldUniqueFileName = existingBanner.UniqueExt

		// Update file-related fields
		existingBanner.AttachmentName = file.Filename
		existingBanner.AttachmentSize = file.Size
		existingBanner.ContentType = file.Header.Get("Content-Type")
		existingBanner.UniqueExt = uniqueFileName

		// Update uploadedBy only when new file is provided
		existingBanner.UploadedBy = request.UploadedBy
	}

	// Update other fields (keep dates in original yyyy-MM-dd format)
	existingBanner.RedirectURL = request.RedirectURL
	// Don't update UploadedBy if no file provided - keep original
	if file == nil {
		// Keep original UploadedBy when no file is provided
		// existingBanner.UploadedBy stays unchanged
	}
	existingBanner.ActiveEndDate = request.ActiveEndDate     // Keep original yyyy-MM-dd format
	existingBanner.ActiveStartDate = request.ActiveStartDate // Keep original yyyy-MM-dd format
	existingBanner.IsActive = request.IsActive
	existingBanner.Duration = request.Duration
	existingBanner.UpdatedAt = time.Now()

	// Save to database
	if err := s.bannerRepo.Update(existingBanner); err != nil {
		// Delete new file if database update fails
		if uniqueFileName != "" {
			storagePath := os.Getenv("BANNER_STORAGE_PATH")
			if storagePath != "" {
				s.fileUtil.DeleteAttachmentFile(uniqueFileName, storagePath)
			}
		}
		return nil, err
	}

	// Delete old file if new file was uploaded successfully
	if oldUniqueFileName != "" && uniqueFileName != "" {
		storagePath := os.Getenv("BANNER_STORAGE_PATH")
		if storagePath != "" {
			s.fileUtil.DeleteAttachmentFile(oldUniqueFileName, storagePath)
		}
	}

	// Convert to response DTO
	response := &dto.Banner{
		BannerId:        existingBanner.BannerId,
		Placement:       existingBanner.Placement,
		RedirectURL:     existingBanner.RedirectURL,
		UploadedBy:      existingBanner.UploadedBy, // This will be original or updated based on file presence
		ActiveEndDate:   existingBanner.ActiveEndDate,
		ActiveStartDate: existingBanner.ActiveStartDate,
		IsActive:        existingBanner.IsActive,
		Duration:        existingBanner.Duration,
		AttachmentName:  existingBanner.AttachmentName,
		AttachmentPath:  existingBanner.AttachmentPath,
		AttachmentSize:  existingBanner.AttachmentSize,
		ContentType:     existingBanner.ContentType,
		UniqueExt:       existingBanner.UniqueExt,
		CreatedAt:       s.formatTimestampToMalaysia(existingBanner.CreatedAt),
		UpdatedAt:       s.formatTimestampToMalaysia(existingBanner.UpdatedAt),
	}

	return response, nil
}

// DeleteBanner deletes a banner by ID
func (s *BannerService) DeleteBanner(bannerId uint) error {
	// Find existing banner
	existingBanner, err := s.bannerRepo.FindByID(bannerId)
	if err != nil {
		return errors.New("banner not found")
	}

	// Delete from database
	if err := s.bannerRepo.Delete(bannerId); err != nil {
		return err
	}

	// Delete associated file
	storagePath := os.Getenv("BANNER_STORAGE_PATH")
	if storagePath != "" {
		s.fileUtil.DeleteAttachmentFile(existingBanner.UniqueExt, storagePath)
	}

	return nil
}

// UpdateBannerPlacements updates multiple banner placements
func (s *BannerService) UpdateBannerPlacements(banners []dto.BannerPlacementUpdate) error {
	for _, bannerUpdate := range banners {
		if err := s.bannerRepo.UpdatePlacement(bannerUpdate.BannerId, bannerUpdate.Placement); err != nil {
			return fmt.Errorf("failed to update placement for banner %d: %v", bannerUpdate.BannerId, err)
		}
	}
	return nil
}

// GetImageInfo retrieves information about a banner image based on its unique extension
func (s *BannerService) GetImageInfo(uniqueExtension string) (string, string, error) {
	// Get content type from banner repository
	contentType, err := s.bannerRepo.GetContentTypeByUniqueExtension(uniqueExtension)
	if err != nil {
		return "", "", err
	}

	if contentType == "" {
		return "", "", errors.New("content type not found")
	}

	// Get storage path from environment variable
	storagePath := os.Getenv("BANNER_STORAGE_PATH")

	// Validate that the file exists
	filePath := filepath.Join(storagePath, uniqueExtension)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", "", err
	}

	return contentType, filePath, nil
}

// validateDateRange validates that end date is after start date (keeping original format)
func (s *BannerService) validateDateRange(startDateStr, endDateStr string) error {
	// Parse start date
	startDate, err := utils.ParseUTC(startDateStr, utils.DateOnlyFormat)
	if err != nil {
		return fmt.Errorf("invalid start date format. Expected yyyy-MM-dd, got: %s", startDateStr)
	}

	// Parse end date
	endDate, err := utils.ParseUTC(endDateStr, utils.DateOnlyFormat)
	if err != nil {
		return fmt.Errorf("invalid end date format. Expected yyyy-MM-dd, got: %s", endDateStr)
	}

	// Validate that end date is after start date
	if endDate.Before(startDate) || endDate.Equal(startDate) {
		return errors.New("end date must be after start date")
	}

	return nil
}

// getCurrentMalaysiaDate gets the current date in Malaysia timezone
func (s *BannerService) getCurrentMalaysiaDate() (string, error) {
	malaysiaTime, err := utils.GetCurrentMalaysiaTime()
	if err != nil {
		return "", err
	}
	return malaysiaTime.Format("2006-01-02"), nil
}

// isBannerActiveAndValid checks if banner is active and within the active date range
func (s *BannerService) isBannerActiveAndValid(banner models.Banner, currentDate string) bool {
	// Check if banner is marked as active
	if !banner.IsActive {
		return false
	}

	// Parse dates for comparison
	current, err := time.Parse("2006-01-02", currentDate)
	if err != nil {
		return false
	}

	startDate, err := time.Parse("2006-01-02", banner.ActiveStartDate)
	if err != nil {
		return false
	}

	endDate, err := time.Parse("2006-01-02", banner.ActiveEndDate)
	if err != nil {
		return false
	}

	// Check if current date is within the active range (inclusive of both start and end dates)
	return (current.Equal(startDate) || current.After(startDate)) &&
		(current.Before(endDate) || current.Equal(endDate))
}

// formatTimestampToMalaysia converts UTC timestamp to Malaysia time in yyyy-MM-dd HH:mm:ss format
func (s *BannerService) formatTimestampToMalaysia(utcTime time.Time) string {
	// Convert to Malaysia time and format
	formattedTime, err := utils.FormatToMalaysiaTime(utcTime, utils.FullDateTimeFormat)
	if err != nil {
		// Fallback to UTC time if conversion fails
		return utcTime.Format(utils.FullDateTimeFormat)
	}
	return formattedTime
}
