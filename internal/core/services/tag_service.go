// File: j-ticketing/internal/core/services/tag_service.go
package service

import (
	"errors"
	dto "j-ticketing/internal/core/dto/tag"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"time"

	"gorm.io/gorm"
)

type TagService struct {
	tagRepo *repositories.TagRepository
}

func NewTagService(tagRepo *repositories.TagRepository) *TagService {
	return &TagService{
		tagRepo: tagRepo,
	}
}

// GetAllTags retrieves all tags from the database
func (s *TagService) GetAllTags() (*dto.TagListResponse, error) {
	tags, err := s.tagRepo.FindAll()
	if err != nil {
		return nil, err
	}

	// Convert models to DTOs
	var tagDTOs []dto.Tag
	for _, tag := range tags {
		tagDTOs = append(tagDTOs, dto.Tag{
			TagId:   tag.TagId,
			TagName: tag.TagName,
			TagDesc: tag.TagDesc,
		})
	}

	return dto.NewTagListResponse(tagDTOs), nil
}

// CreateTag creates a new tag after validating uniqueness
func (s *TagService) CreateTag(req *dto.CreateNewTagRequest) (*dto.Tag, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Check if tag name already exists
	existingTag, err := s.tagRepo.FindByName(req.TagName)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if existingTag != nil {
		return nil, errors.New("tag name already exists")
	}

	// Create new tag model
	tag := &models.Tag{
		TagName:   req.TagName,
		TagDesc:   req.TagDesc,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	if err := s.tagRepo.Create(tag); err != nil {
		return nil, err
	}

	// Return DTO
	return &dto.Tag{
		TagId:   tag.TagId,
		TagName: tag.TagName,
		TagDesc: tag.TagDesc,
	}, nil
}

// UpdateTag updates an existing tag with uniqueness validation
func (s *TagService) UpdateTag(req *dto.UpdateTagRequest) (*dto.Tag, error) {
	// Validate request
	if err := req.Validate(); err != nil {
		return nil, err
	}

	// Find existing tag by ID
	existingTag, err := s.tagRepo.FindByID(req.TagId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tag not found")
		}
		return nil, err
	}

	// Check if tag name is being changed
	if existingTag.TagName != req.TagName {
		// Tag name is being changed, check if new name already exists
		tagWithNewName, err := s.tagRepo.FindByName(req.TagName)
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		if tagWithNewName != nil {
			return nil, errors.New("tag name already exists")
		}
	}

	// Update tag fields
	existingTag.TagName = req.TagName
	existingTag.TagDesc = req.TagDesc
	existingTag.UpdatedAt = time.Now()

	// Save to database
	if err := s.tagRepo.Update(existingTag); err != nil {
		return nil, err
	}

	// Return DTO
	return &dto.Tag{
		TagId:   existingTag.TagId,
		TagName: existingTag.TagName,
		TagDesc: existingTag.TagDesc,
	}, nil
}

// DeleteTag deletes a tag by ID
func (s *TagService) DeleteTag(req *dto.DeleteTagRequest) error {
	// Validate request
	if err := req.Validate(); err != nil {
		return err
	}

	// Check if tag exists
	_, err := s.tagRepo.FindByID(req.TagId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("tag not found")
		}
		return err
	}

	// Delete tag
	return s.tagRepo.Delete(req.TagId)
}
