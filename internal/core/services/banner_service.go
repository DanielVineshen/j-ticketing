// File: j-ticketing/internal/core/services/banner_service.go
package service

import (
	"errors"
	"fmt"
	"io"
	dto "j-ticketing/internal/core/dto/banner"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/utils"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// BannerService handles operations related to banner management
type BannerService struct {
	bannerRepo *repositories.BannerRepository
	fileUtil   *FileUtil
}

// NewBannerService creates a new banner service
func NewBannerService(bannerRepo *repositories.BannerRepository) *BannerService {
	return &BannerService{
		bannerRepo: bannerRepo,
		fileUtil:   NewFileUtil(),
	}
}

// GetAllBanners retrieves all banners
func (s *BannerService) GetAllBanners() ([]dto.Banner, error) {
	bannerModels, err := s.bannerRepo.FindAll()
	if err != nil {
		return nil, err
	}

	var banners []dto.Banner
	for _, model := range bannerModels {
		// Convert dates from yyyyMMdd to yyyy-MM-dd format
		formattedStartDate := s.formatDateFromDB(model.ActiveStartDate)
		formattedEndDate := s.formatDateFromDB(model.ActiveEndDate)

		// Convert timestamps to Malaysia time with yyyy-MM-dd HH:mm:ss format
		formattedCreatedAt := s.formatTimestampToMalaysia(model.CreatedAt)
		formattedUpdatedAt := s.formatTimestampToMalaysia(model.UpdatedAt)

		banner := dto.Banner{
			BannerId:        model.BannerId,
			Placement:       model.Placement,
			RedirectURL:     model.RedirectURL,
			UploadedBy:      model.UploadedBy,
			ActiveEndDate:   formattedEndDate,
			ActiveStartDate: formattedStartDate,
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

// CreateBanner creates a new banner with file upload
func (s *BannerService) CreateBanner(request *dto.CreateNewBannerRequest, file *multipart.FileHeader) (*dto.Banner, error) {
	// Validate and convert dates
	formattedStartDate, formattedEndDate, err := s.validateAndFormatDates(request.ActiveStartDate, request.ActiveEndDate)
	if err != nil {
		return nil, err
	}

	// Validate and upload file
	uniqueFileName, err := s.fileUtil.UploadAttachmentFile(file)
	if err != nil {
		return nil, err
	}

	// Get the highest placement number and increment by 1
	maxPlacement, err := s.bannerRepo.GetMaxPlacement()
	if err != nil {
		// If no banners exist, start with placement 1
		maxPlacement = 0
	}

	// Create banner model
	bannerModel := &models.Banner{
		Placement:       maxPlacement + 1,
		RedirectURL:     request.RedirectURL,
		UploadedBy:      request.UploadedBy,
		ActiveEndDate:   formattedEndDate,   // yyyyMMdd format for DB
		ActiveStartDate: formattedStartDate, // yyyyMMdd format for DB
		IsActive:        request.IsActive,
		Duration:        request.Duration,
		AttachmentName:  file.Filename,
		AttachmentPath:  os.Getenv("BANNER_STORAGE_PATH"),
		AttachmentSize:  file.Size,
		ContentType:     file.Header.Get("Content-Type"),
		UniqueExt:       uniqueFileName,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save to database
	if err := s.bannerRepo.Create(bannerModel); err != nil {
		// Delete uploaded file if database save fails
		s.fileUtil.DeleteAttachmentFile(uniqueFileName)
		return nil, err
	}

	// Convert to response DTO with proper date formatting
	response := &dto.Banner{
		BannerId:        bannerModel.BannerId,
		Placement:       bannerModel.Placement,
		RedirectURL:     bannerModel.RedirectURL,
		UploadedBy:      bannerModel.UploadedBy,
		ActiveEndDate:   request.ActiveEndDate,   // Return original yyyy-MM-dd format
		ActiveStartDate: request.ActiveStartDate, // Return original yyyy-MM-dd format
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

	// Validate and convert dates
	formattedStartDate, formattedEndDate, err := s.validateAndFormatDates(request.ActiveStartDate, request.ActiveEndDate)
	if err != nil {
		return nil, err
	}

	var uniqueFileName string
	var oldUniqueFileName string

	// Handle file upload if new file is provided
	if file != nil {
		// Upload new file
		uniqueFileName, err = s.fileUtil.UploadAttachmentFile(file)
		if err != nil {
			return nil, err
		}
		oldUniqueFileName = existingBanner.UniqueExt

		// Update file-related fields
		existingBanner.AttachmentName = file.Filename
		existingBanner.AttachmentSize = file.Size
		existingBanner.ContentType = file.Header.Get("Content-Type")
		existingBanner.UniqueExt = uniqueFileName
	}

	// Update other fields
	existingBanner.RedirectURL = request.RedirectURL
	existingBanner.UploadedBy = request.UploadedBy
	existingBanner.ActiveEndDate = formattedEndDate     // yyyyMMdd format for DB
	existingBanner.ActiveStartDate = formattedStartDate // yyyyMMdd format for DB
	existingBanner.IsActive = request.IsActive
	existingBanner.Duration = request.Duration
	existingBanner.UpdatedAt = time.Now()

	// Save to database
	if err := s.bannerRepo.Update(existingBanner); err != nil {
		// Delete new file if database update fails
		if uniqueFileName != "" {
			s.fileUtil.DeleteAttachmentFile(uniqueFileName)
		}
		return nil, err
	}

	// Delete old file if new file was uploaded successfully
	if oldUniqueFileName != "" && uniqueFileName != "" {
		s.fileUtil.DeleteAttachmentFile(oldUniqueFileName)
	}

	// Convert to response DTO with proper date formatting
	response := &dto.Banner{
		BannerId:        existingBanner.BannerId,
		Placement:       existingBanner.Placement,
		RedirectURL:     existingBanner.RedirectURL,
		UploadedBy:      existingBanner.UploadedBy,
		ActiveEndDate:   request.ActiveEndDate,   // Return original yyyy-MM-dd format
		ActiveStartDate: request.ActiveStartDate, // Return original yyyy-MM-dd format
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
	s.fileUtil.DeleteAttachmentFile(existingBanner.UniqueExt)

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

// GetImageInfo retrieves information about a banner image based on its unique extension (existing method)
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

// validateAndFormatDates validates and converts date format from yyyy-MM-dd to yyyyMMdd
func (s *BannerService) validateAndFormatDates(startDateStr, endDateStr string) (string, string, error) {
	// Parse start date
	startDate, err := utils.ParseUTC(startDateStr, utils.DateOnlyFormat)
	if err != nil {
		return "", "", fmt.Errorf("invalid start date format. Expected yyyy-MM-dd, got: %s", startDateStr)
	}

	// Parse end date
	endDate, err := utils.ParseUTC(endDateStr, utils.DateOnlyFormat)
	if err != nil {
		return "", "", fmt.Errorf("invalid end date format. Expected yyyy-MM-dd, got: %s", endDateStr)
	}

	// Validate that end date is after start date
	if endDate.Before(startDate) || endDate.Equal(startDate) {
		return "", "", errors.New("end date must be after start date")
	}

	// Convert to yyyyMMdd format for database storage
	formattedStartDate := startDate.Format("20060102")
	formattedEndDate := endDate.Format("20060102")

	return formattedStartDate, formattedEndDate, nil
}

// formatDateFromDB converts date from yyyyMMdd format (DB) to yyyy-MM-dd format (API response)
func (s *BannerService) formatDateFromDB(dateStr string) string {
	if dateStr == "" || len(dateStr) != 8 {
		return dateStr // Return as-is if invalid format
	}

	// Parse yyyyMMdd format
	date, err := time.Parse("20060102", dateStr)
	if err != nil {
		return dateStr // Return as-is if parsing fails
	}

	// Convert to yyyy-MM-dd format
	return date.Format(utils.DateOnlyFormat)
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

// FileUtil handles file operations
type FileUtil struct {
	maxFileSize         int64
	allowedContentTypes []string
}

// NewFileUtil creates a new file utility
func NewFileUtil() *FileUtil {
	return &FileUtil{
		maxFileSize: 500 * 1024 * 1024, // 500MB
		allowedContentTypes: []string{
			"image/jpeg",
			"image/png",
		},
	}
}

// UploadAttachmentFile uploads a file and returns the unique filename
func (f *FileUtil) UploadAttachmentFile(file *multipart.FileHeader) (string, error) {
	// Validate file
	if err := f.validateFile(file); err != nil {
		return "", err
	}

	// Generate unique filename
	uniqueFileName := uuid.New().String() + "-" + file.Filename

	// Get storage path
	storagePath := os.Getenv("BANNER_STORAGE_PATH")
	if storagePath == "" {
		return "", errors.New("BANNER_STORAGE_PATH environment variable not set")
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(storagePath, 0755); err != nil {
		return "", fmt.Errorf("failed to create storage directory: %v", err)
	}

	// Save file
	targetPath := filepath.Join(storagePath, uniqueFileName)
	if err := saveUploadedFile(file, targetPath); err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	return uniqueFileName, nil
}

// DeleteAttachmentFile deletes a file by its unique filename
func (f *FileUtil) DeleteAttachmentFile(uniqueFileName string) {
	storagePath := os.Getenv("BANNER_STORAGE_PATH")
	if storagePath == "" {
		return
	}

	filePath := filepath.Join(storagePath, uniqueFileName)
	os.Remove(filePath) // Ignore errors as file might not exist
}

// validateFile validates the uploaded file
func (f *FileUtil) validateFile(file *multipart.FileHeader) error {
	// Check file size
	if file.Size > f.maxFileSize {
		return errors.New("file size exceeds maximum allowed size of 500MB")
	}

	// Check content type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		return errors.New("content type not specified")
	}

	isAllowed := false
	for _, allowedType := range f.allowedContentTypes {
		if contentType == allowedType {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return errors.New("file type not allowed. Only JPEG and PNG images are supported")
	}

	return nil
}

// saveUploadedFile saves the uploaded file to the specified path
func saveUploadedFile(file *multipart.FileHeader, dst string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, src)
	return err
}
