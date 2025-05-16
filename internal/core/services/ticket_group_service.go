// FILE: internal/services/ticket_group_service.go
package service

import (
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"strings"
	"time"
)

// TicketGroupResponse represents the response structure for ticket groups
type TicketGroupResponse struct {
	TicketGroups []TicketGroupDTO `json:"ticketGroups"`
}

// TicketGroupDTO represents the data transfer object for a ticket group
type TicketGroupDTO struct {
	TicketGroupId   uint     `json:"ticketGroupId"`
	GroupType       string   `json:"groupType"`
	GroupName       string   `json:"groupName"`
	GroupDesc       string   `json:"groupDesc"`
	OperatingHours  string   `json:"operatingHours"`
	PricePrefix     string   `json:"pricePrefix"`
	PriceSuffix     string   `json:"priceSuffix"`
	AttachmentName  string   `json:"attachmentName"`
	AttachmentPath  string   `json:"attachmentPath"`
	AttachmentSize  int64    `json:"attachmentSize"`
	ContentType     string   `json:"contentType"`
	UniqueExtension string   `json:"uniqueExtension"`
	IsActive        bool     `json:"isActive"`
	Tags            []TagDTO `json:"tags"`
}

// TagDTO represents the data transfer object for a tag
type TagDTO struct {
	TagId   uint   `json:"tagId"`
	TagName string `json:"tagName"`
	TagDesc string `json:"tagDesc"`
}

// TicketGroupService handles ticket group-related operations
type TicketGroupService struct {
	ticketGroupRepo  *repositories.TicketGroupRepository
	tagRepo          *repositories.TagRepository
	groupGalleryRepo *repositories.GroupGalleryRepository
	ticketDetailRepo *repositories.TicketDetailRepository
}

// NewTicketGroupService creates a new instance of TicketGroupService
func NewTicketGroupService(
	ticketGroupRepo *repositories.TicketGroupRepository,
	tagRepo *repositories.TagRepository,
	groupGalleryRepo *repositories.GroupGalleryRepository,
	ticketDetailRepo *repositories.TicketDetailRepository,
) *TicketGroupService {
	return &TicketGroupService{
		ticketGroupRepo:  ticketGroupRepo,
		tagRepo:          tagRepo,
		groupGalleryRepo: groupGalleryRepo,
		ticketDetailRepo: ticketDetailRepo,
	}
}

// GetAllTicketGroups retrieves all ticket groups with their associated tags
func (s *TicketGroupService) GetAllTicketGroups() (TicketGroupResponse, error) {
	// Fetch all ticket groups
	ticketGroups, err := s.ticketGroupRepo.FindAll()
	if err != nil {
		return TicketGroupResponse{}, err
	}

	return s.buildTicketGroupResponse(ticketGroups)
}

// GetActiveTicketGroups retrieves only active ticket groups with their associated tags
func (s *TicketGroupService) GetActiveTicketGroups() (TicketGroupResponse, error) {
	// Fetch active ticket groups
	ticketGroups, err := s.ticketGroupRepo.FindActiveTicketGroups()
	if err != nil {
		return TicketGroupResponse{}, err
	}

	return s.buildTicketGroupResponse(ticketGroups)
}

// GetTicketGroupById retrieves a specific ticket group by ID with its associated tags
func (s *TicketGroupService) GetTicketGroupById(id uint) (*TicketGroupDTO, error) {
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
	tagDTOs := make([]TagDTO, 0, len(tags))
	for _, tag := range tags {
		tagDTOs = append(tagDTOs, TagDTO{
			TagId:   tag.TagId,
			TagName: tag.TagName,
			TagDesc: tag.TagDesc,
		})
	}

	// Create the ticket group DTO
	ticketGroupDTO := &TicketGroupDTO{
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
func (s *TicketGroupService) buildTicketGroupResponse(ticketGroups []models.TicketGroup) (TicketGroupResponse, error) {
	// Create the response
	response := TicketGroupResponse{
		TicketGroups: make([]TicketGroupDTO, 0, len(ticketGroups)),
	}

	// Populate the response with ticket groups
	for _, ticketGroup := range ticketGroups {
		// Get tags for this ticket group
		tags, err := s.tagRepo.FindByTicketGroupID(ticketGroup.TicketGroupId)
		if err != nil {
			return TicketGroupResponse{}, err
		}

		// Map tags to DTOs
		tagDTOs := make([]TagDTO, 0, len(tags))
		for _, tag := range tags {
			tagDTOs = append(tagDTOs, TagDTO{
				TagId:   tag.TagId,
				TagName: tag.TagName,
				TagDesc: tag.TagDesc,
			})
		}

		// Create the ticket group DTO
		ticketGroupDTO := TicketGroupDTO{
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
func (s *TicketGroupService) GetTicketProfile(ticketGroupId uint) (*TicketProfileResponse, error) {
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
	tagDTOs := make([]TagDTO, 0, len(tags))
	for _, tag := range tags {
		tagDTOs = append(tagDTOs, TagDTO{
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
	profile := &TicketProfileDTO{
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
	response := &TicketProfileResponse{
		RespCode: 200,
		RespDesc: "Success",
		Result: TicketProfileResult{
			TicketProfile: *profile,
		},
	}

	return response, nil
}

// getGroupGallery retrieves gallery items for a ticket group
func (s *TicketGroupService) getGroupGallery(ticketGroupId uint) ([]GroupGalleryDTO, error) {
	// This would be implemented by calling a repository method
	// Create a GroupGalleryRepository or use this from another service
	galleries, err := s.groupGalleryRepo.FindByTicketGroupID(ticketGroupId)
	if err != nil {
		return nil, err
	}

	galleryDTOs := make([]GroupGalleryDTO, 0, len(galleries))
	for _, gallery := range galleries {
		galleryDTOs = append(galleryDTOs, GroupGalleryDTO{
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
func (s *TicketGroupService) getTicketDetails(ticketGroupId uint) ([]TicketDetailDTO, error) {
	// This would be implemented by calling a repository method
	// Create a TicketDetailRepository or use this from another service
	details, err := s.ticketDetailRepo.FindByTicketGroupID(ticketGroupId)
	if err != nil {
		return nil, err
	}

	detailDTOs := make([]TicketDetailDTO, 0, len(details))
	for _, detail := range details {
		detailDTOs = append(detailDTOs, TicketDetailDTO{
			TicketDetailId: detail.TicketDetailId,
			Title:          detail.Title,
			TitleIcon:      detail.TitleIcon,
			RawHtml:        detail.RawHtml,
			DisplayFlag:    detail.DisplayFlag,
		})
	}

	return detailDTOs, nil
}
