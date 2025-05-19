// File: j-ticketing/internal/core/services/banner_service.go
package service

import (
	"j-ticketing/internal/db/repositories"
	"os"
	"path/filepath"
)

// BannerService handles operations related to serving banner images
type BannerService struct {
	bannerRepo *repositories.BannerRepository
}

// BannerService creates a new banner image service
func NewBannerService(bannerRepo *repositories.BannerRepository) *BannerService {
	return &BannerService{
		bannerRepo: bannerRepo,
	}
}

// GetImageInfo retrieves information about a banner image based on its unique extension
func (s *BannerService) GetImageInfo(uniqueExtension string) (string, string, error) {
	// Get content type from banner repository
	contentType, err := s.bannerRepo.GetContentTypeByUniqueExtension(uniqueExtension)
	if err != nil {
		return "", "", err
	}

	if contentType == "" {
		return "", "", err
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
