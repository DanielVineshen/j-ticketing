// File: j-ticketing/internal/core/services/group_gallery_service.go
package service

import (
	"j-ticketing/internal/db/repositories"
	"os"
	"path/filepath"
)

// GroupGalleryService handles operations related to serving group gallery images
type GroupGalleryService struct {
	groupGalleryRepo *repositories.GroupGalleryRepository
}

// GroupGalleryService creates a new group gallery image service
func NewGroupGalleryService(groupGalleryRepo *repositories.GroupGalleryRepository) *GroupGalleryService {
	return &GroupGalleryService{
		groupGalleryRepo: groupGalleryRepo,
	}
}

// GetImageInfo retrieves information about a group gallery image based on its unique extension
func (s *GroupGalleryService) GetImageInfo(uniqueExtension string) (string, string, error) {
	// Get content type from group gallery repository
	contentType, err := s.groupGalleryRepo.GetContentTypeByUniqueExtension(uniqueExtension)
	if err != nil {
		return "", "", err
	}

	if contentType == "" {
		return "", "", err
	}

	// Get storage path from environment variable
	storagePath := os.Getenv("GROUP_GALLERY_STORAGE_PATH")

	// Validate that the file exists
	filePath := filepath.Join(storagePath, uniqueExtension)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", "", err
	}

	return contentType, filePath, nil
}
