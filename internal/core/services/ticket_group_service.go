// File: j-ticketing/internal/core/services/ticket_group_service.go
package service

import (
	dto "j-ticketing/internal/core/dto/ticket_group"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/config"
	"j-ticketing/pkg/external"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TicketGroupService handles ticket group-related operations
type TicketGroupService struct {
	ticketGroupRepo  *repositories.TicketGroupRepository
	tagRepo          *repositories.TagRepository
	groupGalleryRepo *repositories.GroupGalleryRepository
	ticketDetailRepo *repositories.TicketDetailRepository
	zooAPIClient     *external.ZooAPIClient
	config           *config.Config
}

// NewTicketGroupService creates a new instance of TicketGroupService
func NewTicketGroupService(
	ticketGroupRepo *repositories.TicketGroupRepository,
	tagRepo *repositories.TagRepository,
	groupGalleryRepo *repositories.GroupGalleryRepository,
	ticketDetailRepo *repositories.TicketDetailRepository,
	cfg *config.Config,
) *TicketGroupService {
	zooAPIClient := external.NewZooAPIClient(
		cfg.ZooAPI.ZooBaseURL,
		cfg.ZooAPI.Username,
		cfg.ZooAPI.Password,
	)

	return &TicketGroupService{
		ticketGroupRepo:  ticketGroupRepo,
		tagRepo:          tagRepo,
		groupGalleryRepo: groupGalleryRepo,
		ticketDetailRepo: ticketDetailRepo,
		zooAPIClient:     zooAPIClient,
		config:           cfg,
	}
}

func (s *TicketGroupService) GetTicketGroup(ticketGroupId uint) (*models.TicketGroup, error) {
	ticketGroup, err := s.ticketGroupRepo.FindByID(ticketGroupId)
	if err != nil {
		log.Printf("Error finding ticket group %s: %v", ticketGroupId, err)
	}
	return ticketGroup, err
}

// GetAllTicketGroups retrieves all ticket groups with their associated tags
func (s *TicketGroupService) GetAllTicketGroups() (dto.TicketGroupResponse, error) {
	// Fetch all ticket groups
	ticketGroups, err := s.ticketGroupRepo.FindAll()
	if err != nil {
		return dto.TicketGroupResponse{}, err
	}

	return s.buildTicketGroupResponse(ticketGroups)
}

// GetActiveTicketGroups retrieves only active ticket groups with their associated tags
func (s *TicketGroupService) GetActiveTicketGroups() (dto.TicketGroupResponse, error) {
	// Fetch active ticket groups
	ticketGroups, err := s.ticketGroupRepo.FindActiveTicketGroups()
	if err != nil {
		return dto.TicketGroupResponse{}, err
	}

	return s.buildTicketGroupResponse(ticketGroups)
}

// GetTicketGroupById retrieves a specific ticket group by ID with its associated tags
func (s *TicketGroupService) GetTicketGroupById(id uint) (*dto.TicketGroupDTO, error) {
	// Fetch the ticket group
	ticketGroup, err := s.ticketGroupRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Get tags for this ticket group
	tags, err := s.tagRepo.FindByTicketGroupID(ticketGroup.TicketGroupId)
	if err != nil {
		return nil, err
	}

	// Map tags to DTOs
	tagDTOs := make([]dto.TagDTO, 0, len(tags))
	for _, tag := range tags {
		tagDTOs = append(tagDTOs, dto.TagDTO{
			TagId:   tag.TagId,
			TagName: tag.TagName,
			TagDesc: tag.TagDesc,
		})
	}

	// Create the ticket group DTO
	ticketGroupDTO := &dto.TicketGroupDTO{
		TicketGroupId:   ticketGroup.TicketGroupId,
		GroupType:       ticketGroup.GroupType,
		GroupName:       ticketGroup.GroupName,
		GroupDesc:       ticketGroup.GroupDesc,
		OperatingHours:  ticketGroup.OperatingHours,
		PricePrefix:     ticketGroup.PricePrefix,
		PriceSuffix:     ticketGroup.PriceSuffix,
		AttachmentName:  ticketGroup.AttachmentName,
		AttachmentPath:  ticketGroup.AttachmentPath,
		AttachmentSize:  ticketGroup.AttachmentSize,
		ContentType:     ticketGroup.ContentType,
		UniqueExtension: ticketGroup.UniqueExtension,
		IsActive:        ticketGroup.IsActive,
		Tags:            tagDTOs,
	}

	return ticketGroupDTO, nil
}

// buildTicketGroupResponse constructs the response with ticket groups and their tags
func (s *TicketGroupService) buildTicketGroupResponse(ticketGroups []models.TicketGroup) (dto.TicketGroupResponse, error) {
	// Create the response
	response := dto.TicketGroupResponse{
		TicketGroups: make([]dto.TicketGroupDTO, 0, len(ticketGroups)),
	}

	// Populate the response with ticket groups
	for _, ticketGroup := range ticketGroups {
		// Get tags for this ticket group
		tags, err := s.tagRepo.FindByTicketGroupID(ticketGroup.TicketGroupId)
		if err != nil {
			return dto.TicketGroupResponse{}, err
		}

		// Map tags to DTOs
		tagDTOs := make([]dto.TagDTO, 0, len(tags))
		for _, tag := range tags {
			tagDTOs = append(tagDTOs, dto.TagDTO{
				TagId:   tag.TagId,
				TagName: tag.TagName,
				TagDesc: tag.TagDesc,
			})
		}

		// Create the ticket group DTO
		ticketGroupDTO := dto.TicketGroupDTO{
			TicketGroupId:   ticketGroup.TicketGroupId,
			GroupType:       ticketGroup.GroupType,
			GroupName:       ticketGroup.GroupName,
			GroupDesc:       ticketGroup.GroupDesc,
			OperatingHours:  ticketGroup.OperatingHours,
			PricePrefix:     ticketGroup.PricePrefix,
			PriceSuffix:     ticketGroup.PriceSuffix,
			AttachmentName:  ticketGroup.AttachmentName,
			AttachmentPath:  ticketGroup.AttachmentPath,
			AttachmentSize:  ticketGroup.AttachmentSize,
			ContentType:     ticketGroup.ContentType,
			UniqueExtension: ticketGroup.UniqueExtension,
			IsActive:        ticketGroup.IsActive,
			Tags:            tagDTOs,
		}

		response.TicketGroups = append(response.TicketGroups, ticketGroupDTO)
	}

	return response, nil
}

// GetTicketProfile retrieves a complete ticket profile by ticket group ID
func (s *TicketGroupService) GetTicketProfile(ticketGroupId uint) (*dto.TicketProfileResult, error) {
	// 1. Get the ticket group
	ticketGroup, err := s.ticketGroupRepo.FindByID(ticketGroupId)
	if err != nil {
		return nil, err
	}

	// 2. Get tags for this ticket group
	tags, err := s.tagRepo.FindByTicketGroupID(ticketGroup.TicketGroupId)
	if err != nil {
		return nil, err
	}

	// 3. Map tags to DTOs
	tagDTOs := make([]dto.TagDTO, 0, len(tags))
	for _, tag := range tags {
		tagDTOs = append(tagDTOs, dto.TagDTO{
			TagId:   tag.TagId,
			TagName: tag.TagName,
			TagDesc: tag.TagDesc,
		})
	}

	// 4. Get gallery items for this ticket group
	galleryItems, err := s.getGroupGallery(ticketGroup.TicketGroupId)
	if err != nil {
		return nil, err
	}

	// 5. Get ticket details for this ticket group
	ticketDetails, err := s.getTicketDetails(ticketGroup.TicketGroupId)
	if err != nil {
		return nil, err
	}

	// 6. Parse organiser facilities from string to string array
	var organiserFacilities []string
	if ticketGroup.OrganiserFacilities != "" {
		// Split the string based on the semicolon separator
		organiserFacilities = strings.Split(ticketGroup.OrganiserFacilities, ";")

		// Optional: Trim any whitespace from each facility
		for i, facility := range organiserFacilities {
			organiserFacilities[i] = strings.TrimSpace(facility)
		}
	} else {
		organiserFacilities = []string{} // Empty array if no facilities
	}

	// 7. Create the response
	profile := &dto.TicketProfileDTO{
		TicketGroupId:            ticketGroup.TicketGroupId,
		GroupType:                ticketGroup.GroupType,
		GroupName:                ticketGroup.GroupName,
		GroupDesc:                ticketGroup.GroupDesc,
		OperatingHours:           ticketGroup.OperatingHours,
		PricePrefix:              ticketGroup.PricePrefix,
		PriceSuffix:              ticketGroup.PriceSuffix,
		AttachmentName:           ticketGroup.AttachmentName,
		AttachmentPath:           ticketGroup.AttachmentPath,
		AttachmentSize:           ticketGroup.AttachmentSize,
		ContentType:              ticketGroup.ContentType,
		UniqueExtension:          ticketGroup.UniqueExtension,
		IsActive:                 ticketGroup.IsActive,
		IsTicketInternal:         ticketGroup.IsTicketInternal,
		TicketIds:                ticketGroup.TicketIds.String,
		Tags:                     tagDTOs,
		GroupGallery:             galleryItems,
		TicketDetails:            ticketDetails,
		LocationAddress:          ticketGroup.LocationAddress,
		LocationMapEmbedUrl:      ticketGroup.LocationMapUrl,
		OrganiserName:            ticketGroup.OrganiserName,
		OrganiserAddress:         ticketGroup.OrganiserAddress,
		OrganiserDescriptionHtml: ticketGroup.OrganiserDescHtml,
		OrganiserContact:         ticketGroup.OrganiserContact,
		OrganiserEmail:           ticketGroup.OrganiserEmail,
		OrganiserWebsite:         ticketGroup.OrganiserWebsite,
		OrganiserOperatingHours:  ticketGroup.OrganiserOperatingHour,
		OrganiserFacilities:      organiserFacilities,
		CreatedAt:                ticketGroup.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                ticketGroup.UpdatedAt.Format(time.RFC3339),
	}

	// 8. Handle optional date fields
	if ticketGroup.ActiveStartDate.Valid {
		profile.ActiveStartDate = ticketGroup.ActiveStartDate.String
	}
	if ticketGroup.ActiveEndDate.Valid {
		profile.ActiveEndDate = ticketGroup.ActiveEndDate.String
	}

	// 9. Prepare the complete response
	response := &dto.TicketProfileResult{
		TicketProfile: *profile,
	}

	return response, nil
}

// getGroupGallery retrieves gallery items for a ticket group
func (s *TicketGroupService) getGroupGallery(ticketGroupId uint) ([]dto.GroupGalleryDTO, error) {
	// This would be implemented by calling a repository method
	// Create a GroupGalleryRepository or use this from another service
	galleries, err := s.groupGalleryRepo.FindByTicketGroupID(ticketGroupId)
	if err != nil {
		return nil, err
	}

	galleryDTOs := make([]dto.GroupGalleryDTO, 0, len(galleries))
	for _, gallery := range galleries {
		galleryDTOs = append(galleryDTOs, dto.GroupGalleryDTO{
			GroupGalleryId:  gallery.GroupGalleryId,
			AttachmentName:  gallery.AttachmentName,
			AttachmentPath:  gallery.AttachmentPath,
			AttachmentSize:  gallery.AttachmentSize,
			ContentType:     gallery.ContentType,
			UniqueExtension: gallery.UniqueExtension,
		})
	}

	return galleryDTOs, nil
}

// getTicketDetails retrieves ticket details for a ticket group
func (s *TicketGroupService) getTicketDetails(ticketGroupId uint) ([]dto.TicketDetailDTO, error) {
	// This would be implemented by calling a repository method
	// Create a TicketDetailRepository or use this from another service
	details, err := s.ticketDetailRepo.FindByTicketGroupID(ticketGroupId)
	if err != nil {
		return nil, err
	}

	detailDTOs := make([]dto.TicketDetailDTO, 0, len(details))
	for _, detail := range details {
		detailDTOs = append(detailDTOs, dto.TicketDetailDTO{
			TicketDetailId: detail.TicketDetailId,
			Title:          detail.Title,
			TitleIcon:      detail.TitleIcon,
			RawHtml:        detail.RawHtml,
			DisplayFlag:    detail.DisplayFlag,
		})
	}

	return detailDTOs, nil
}

// GetTicketVariants retrieves ticket variants for a specific ticket group and date
func (s *TicketGroupService) GetTicketVariants(ticketGroupId uint, date string) (*dto.TicketVariantResponse, error) {
	// First, check if the ticket group exists
	ticketGroup, err := s.ticketGroupRepo.FindByID(ticketGroupId)
	if err != nil {
		return nil, err
	}

	// Get available ticket items from the external API
	ticketItems, err := s.zooAPIClient.GetTicketItems(ticketGroup.GroupName, date)
	if err != nil {
		return nil, err
	}

	// Check if the ticket group has specific ticket IDs to filter
	var allowedTicketIDs map[string]bool
	if ticketGroup.TicketIds.Valid && ticketGroup.TicketIds.String != "" {
		// Split the comma-separated list of ticket IDs
		ticketIDsRaw := strings.Split(ticketGroup.TicketIds.String, ";")
		allowedTicketIDs = make(map[string]bool)

		for _, id := range ticketIDsRaw {
			id = strings.TrimSpace(id)
			if id != "" {
				allowedTicketIDs[id] = true
			}
		}
	}

	// Convert the ticket items to DTOs, filtering by the allowed ticket IDs if necessary
	ticketVariants := make([]dto.TicketVariantDTO, 0)
	for _, item := range ticketItems {
		// Skip if we have a filter and this item is not in the allowed list
		if allowedTicketIDs != nil && !allowedTicketIDs[item.ItemId] {
			continue
		}

		// Create the DTO
		variant := dto.TicketVariantDTO{
			TicketId:  item.ItemId,
			UnitPrice: item.UnitPrice,
			ItemDesc1: item.ItemDescription,
			ItemDesc2: item.ItemDesc1,
			ItemDesc3: item.ItemDesc2,
			PrintType: item.PrintType,
			Qty:       item.Qty,
		}

		ticketVariants = append(ticketVariants, variant)
	}

	// Create the response
	response := &dto.TicketVariantResponse{
		TicketVariants: ticketVariants,
	}

	return response, nil
}

// GetImageInfo retrieves information about a ticket group image based on its unique extension
func (s *TicketGroupService) GetImageInfo(uniqueExtension string) (string, string, error) {
	// Get content type from ticket group repository
	contentType, err := s.ticketGroupRepo.GetContentTypeByUniqueExtension(uniqueExtension)
	if err != nil {
		return "", "", err
	}

	if contentType == "" {
		return "", "", err
	}

	// Get storage path from environment variable
	storagePath := os.Getenv("TICKET_GROUP_STORAGE_PATH")

	// Validate that the file exists
	filePath := filepath.Join(storagePath, uniqueExtension)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", "", err
	}

	return contentType, filePath, nil
}
