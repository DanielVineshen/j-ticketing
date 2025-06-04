// File: j-ticketing/internal/core/services/ticket_group_service.go
package service

import (
	"database/sql"
	"errors"
	"fmt"
	"gorm.io/gorm"
	dto "j-ticketing/internal/core/dto/ticket_group"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/config"
	"j-ticketing/pkg/external"
	"j-ticketing/pkg/utils"
	"log"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// TicketGroupService handles ticket group-related operations
type TicketGroupService struct {
	ticketGroupRepo   *repositories.TicketGroupRepository
	tagRepo           *repositories.TagRepository
	groupGalleryRepo  *repositories.GroupGalleryRepository
	ticketDetailRepo  *repositories.TicketDetailRepository
	ticketVariantRepo *repositories.TicketVariantRepository
	zooAPIClient      *external.ZooAPIClient
	config            *config.Config
}

// NewTicketGroupService creates a new instance of TicketGroupService
func NewTicketGroupService(
	ticketGroupRepo *repositories.TicketGroupRepository,
	tagRepo *repositories.TagRepository,
	groupGalleryRepo *repositories.GroupGalleryRepository,
	ticketDetailRepo *repositories.TicketDetailRepository,
	ticketVariantRepo *repositories.TicketVariantRepository,
	cfg *config.Config,
) *TicketGroupService {
	zooAPIClient := external.NewZooAPIClient(
		cfg.ZooAPI.ZooBaseURL,
		cfg.ZooAPI.Username,
		cfg.ZooAPI.Password,
	)

	return &TicketGroupService{
		ticketGroupRepo:   ticketGroupRepo,
		tagRepo:           tagRepo,
		groupGalleryRepo:  groupGalleryRepo,
		ticketDetailRepo:  ticketDetailRepo,
		ticketVariantRepo: ticketVariantRepo,
		zooAPIClient:      zooAPIClient,
		config:            cfg,
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
		TicketGroupId:          ticketGroup.TicketGroupId,
		Placement:              ticketGroup.Placement,
		OrderTicketLimit:       ticketGroup.OrderTicketLimit,
		ScanSetting:            ticketGroup.ScanSetting,
		GroupType:              ticketGroup.GroupType,
		GroupNameBm:            ticketGroup.GroupNameBm,
		GroupNameEn:            ticketGroup.GroupNameEn,
		GroupNameCn:            ticketGroup.GroupNameCn,
		GroupDescBm:            ticketGroup.GroupDescBm,
		GroupDescEn:            ticketGroup.GroupDescEn,
		GroupDescCn:            ticketGroup.GroupDescCn,
		GroupRedirectionSpanBm: nullStringToPointer(ticketGroup.GroupRedirectionSpanBm),
		GroupRedirectionSpanEn: nullStringToPointer(ticketGroup.GroupRedirectionSpanEn),
		GroupRedirectionSpanCn: nullStringToPointer(ticketGroup.GroupRedirectionSpanCn),
		GroupRedirectionUrl:    nullStringToPointer(ticketGroup.GroupRedirectionUrl),
		GroupSlot1Bm:           nullStringToPointer(ticketGroup.GroupSlot1Bm),
		GroupSlot1En:           nullStringToPointer(ticketGroup.GroupSlot1En),
		GroupSlot1Cn:           nullStringToPointer(ticketGroup.GroupSlot1Cn),
		GroupSlot2Bm:           nullStringToPointer(ticketGroup.GroupSlot2Bm),
		GroupSlot2En:           nullStringToPointer(ticketGroup.GroupSlot2En),
		GroupSlot2Cn:           nullStringToPointer(ticketGroup.GroupSlot2Cn),
		GroupSlot3Bm:           nullStringToPointer(ticketGroup.GroupSlot3Bm),
		GroupSlot3En:           nullStringToPointer(ticketGroup.GroupSlot3En),
		GroupSlot3Cn:           nullStringToPointer(ticketGroup.GroupSlot3Cn),
		GroupSlot4Bm:           nullStringToPointer(ticketGroup.GroupSlot4Bm),
		GroupSlot4En:           nullStringToPointer(ticketGroup.GroupSlot4En),
		GroupSlot4Cn:           nullStringToPointer(ticketGroup.GroupSlot4Cn),
		PricePrefixBm:          ticketGroup.PricePrefixBm,
		PricePrefixEn:          ticketGroup.PricePrefixEn,
		PricePrefixCn:          ticketGroup.PricePrefixCn,
		PriceSuffixBm:          ticketGroup.PriceSuffixBm,
		PriceSuffixEn:          ticketGroup.PriceSuffixEn,
		PriceSuffixCn:          ticketGroup.PriceSuffixCn,
		AttachmentName:         ticketGroup.AttachmentName,
		AttachmentPath:         ticketGroup.AttachmentPath,
		AttachmentSize:         ticketGroup.AttachmentSize,
		ContentType:            ticketGroup.ContentType,
		UniqueExtension:        ticketGroup.UniqueExtension,
		ActiveStartDate:        nullStringToPointer(ticketGroup.ActiveStartDate),
		ActiveEndDate:          nullStringToPointer(ticketGroup.ActiveEndDate),
		IsActive:               ticketGroup.IsActive,
		Tags:                   tagDTOs,
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
			TicketGroupId:          ticketGroup.TicketGroupId,
			Placement:              ticketGroup.Placement,
			OrderTicketLimit:       ticketGroup.OrderTicketLimit,
			ScanSetting:            ticketGroup.ScanSetting,
			GroupType:              ticketGroup.GroupType,
			GroupNameBm:            ticketGroup.GroupNameBm,
			GroupNameEn:            ticketGroup.GroupNameEn,
			GroupNameCn:            ticketGroup.GroupNameCn,
			GroupDescBm:            ticketGroup.GroupDescBm,
			GroupDescEn:            ticketGroup.GroupDescEn,
			GroupDescCn:            ticketGroup.GroupDescCn,
			GroupRedirectionSpanBm: nullStringToPointer(ticketGroup.GroupRedirectionSpanBm),
			GroupRedirectionSpanEn: nullStringToPointer(ticketGroup.GroupRedirectionSpanEn),
			GroupRedirectionSpanCn: nullStringToPointer(ticketGroup.GroupRedirectionSpanCn),
			GroupRedirectionUrl:    nullStringToPointer(ticketGroup.GroupRedirectionUrl),
			GroupSlot1Bm:           nullStringToPointer(ticketGroup.GroupSlot1Bm),
			GroupSlot1En:           nullStringToPointer(ticketGroup.GroupSlot1En),
			GroupSlot1Cn:           nullStringToPointer(ticketGroup.GroupSlot1Cn),
			GroupSlot2Bm:           nullStringToPointer(ticketGroup.GroupSlot2Bm),
			GroupSlot2En:           nullStringToPointer(ticketGroup.GroupSlot2En),
			GroupSlot2Cn:           nullStringToPointer(ticketGroup.GroupSlot2Cn),
			GroupSlot3Bm:           nullStringToPointer(ticketGroup.GroupSlot3Bm),
			GroupSlot3En:           nullStringToPointer(ticketGroup.GroupSlot3En),
			GroupSlot3Cn:           nullStringToPointer(ticketGroup.GroupSlot3Cn),
			GroupSlot4Bm:           nullStringToPointer(ticketGroup.GroupSlot4Bm),
			GroupSlot4En:           nullStringToPointer(ticketGroup.GroupSlot4En),
			GroupSlot4Cn:           nullStringToPointer(ticketGroup.GroupSlot4Cn),
			PricePrefixBm:          ticketGroup.PricePrefixBm,
			PricePrefixEn:          ticketGroup.PricePrefixEn,
			PricePrefixCn:          ticketGroup.PricePrefixCn,
			PriceSuffixBm:          ticketGroup.PriceSuffixBm,
			PriceSuffixEn:          ticketGroup.PriceSuffixEn,
			PriceSuffixCn:          ticketGroup.PriceSuffixCn,
			AttachmentName:         ticketGroup.AttachmentName,
			AttachmentPath:         ticketGroup.AttachmentPath,
			AttachmentSize:         ticketGroup.AttachmentSize,
			ContentType:            ticketGroup.ContentType,
			UniqueExtension:        ticketGroup.UniqueExtension,
			ActiveStartDate:        nullStringToPointer(ticketGroup.ActiveStartDate),
			ActiveEndDate:          nullStringToPointer(ticketGroup.ActiveEndDate),
			IsActive:               ticketGroup.IsActive,
			Tags:                   tagDTOs,
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

	ticketVariants, err := s.getLocalTicketVariants(ticketGroup.TicketGroupId)
	if err != nil {
		return nil, err
	}

	// Parse organiser facilities from string to string array
	var organiserFacilitiesBm []string
	facilitiesBmStr := getStringFromNullString(ticketGroup.OrganiserFacilitiesBm)
	if facilitiesBmStr != "" {
		organiserFacilitiesBm = strings.Split(facilitiesBmStr, ";")
		for i, facility := range organiserFacilitiesBm {
			organiserFacilitiesBm[i] = strings.TrimSpace(facility)
		}
	} else {
		organiserFacilitiesBm = []string{}
	}

	var organiserFacilitiesEn []string
	facilitiesEnStr := getStringFromNullString(ticketGroup.OrganiserFacilitiesEn)
	if facilitiesEnStr != "" {
		organiserFacilitiesEn = strings.Split(facilitiesEnStr, ";")
		for i, facility := range organiserFacilitiesEn {
			organiserFacilitiesEn[i] = strings.TrimSpace(facility)
		}
	} else {
		organiserFacilitiesEn = []string{}
	}

	var organiserFacilitiesCn []string
	facilitiesCnStr := getStringFromNullString(ticketGroup.OrganiserFacilitiesCn)
	if facilitiesCnStr != "" {
		organiserFacilitiesCn = strings.Split(facilitiesCnStr, ";")
		for i, facility := range organiserFacilitiesCn {
			organiserFacilitiesCn[i] = strings.TrimSpace(facility)
		}
	} else {
		organiserFacilitiesCn = []string{}
	}

	// Build the ticket profile DTO
	profile := &dto.TicketProfileDTO{
		TicketGroupId:              ticketGroup.TicketGroupId,
		Placement:                  ticketGroup.Placement,
		OrderTicketLimit:           ticketGroup.OrderTicketLimit,
		ScanSetting:                ticketGroup.ScanSetting,
		GroupType:                  ticketGroup.GroupType,
		GroupNameBm:                ticketGroup.GroupNameBm,
		GroupNameEn:                ticketGroup.GroupNameEn,
		GroupNameCn:                ticketGroup.GroupNameCn,
		GroupDescBm:                ticketGroup.GroupDescBm,
		GroupDescEn:                ticketGroup.GroupDescEn,
		GroupDescCn:                ticketGroup.GroupDescCn,
		GroupRedirectionSpanBm:     nullStringToPointer(ticketGroup.GroupRedirectionSpanBm),
		GroupRedirectionSpanEn:     nullStringToPointer(ticketGroup.GroupRedirectionSpanEn),
		GroupRedirectionSpanCn:     nullStringToPointer(ticketGroup.GroupRedirectionSpanCn),
		GroupRedirectionUrl:        nullStringToPointer(ticketGroup.GroupRedirectionUrl),
		GroupSlot1Bm:               nullStringToPointer(ticketGroup.GroupSlot1Bm),
		GroupSlot1En:               nullStringToPointer(ticketGroup.GroupSlot1En),
		GroupSlot1Cn:               nullStringToPointer(ticketGroup.GroupSlot1Cn),
		GroupSlot2Bm:               nullStringToPointer(ticketGroup.GroupSlot2Bm),
		GroupSlot2En:               nullStringToPointer(ticketGroup.GroupSlot2En),
		GroupSlot2Cn:               nullStringToPointer(ticketGroup.GroupSlot2Cn),
		GroupSlot3Bm:               nullStringToPointer(ticketGroup.GroupSlot3Bm),
		GroupSlot3En:               nullStringToPointer(ticketGroup.GroupSlot3En),
		GroupSlot3Cn:               nullStringToPointer(ticketGroup.GroupSlot3Cn),
		GroupSlot4Bm:               nullStringToPointer(ticketGroup.GroupSlot4Bm),
		GroupSlot4En:               nullStringToPointer(ticketGroup.GroupSlot4En),
		GroupSlot4Cn:               nullStringToPointer(ticketGroup.GroupSlot4Cn),
		PricePrefixBm:              ticketGroup.PricePrefixBm,
		PricePrefixEn:              ticketGroup.PricePrefixEn,
		PricePrefixCn:              ticketGroup.PricePrefixCn,
		PriceSuffixBm:              ticketGroup.PriceSuffixBm,
		PriceSuffixEn:              ticketGroup.PriceSuffixEn,
		PriceSuffixCn:              ticketGroup.PriceSuffixCn,
		AttachmentName:             ticketGroup.AttachmentName,
		AttachmentPath:             ticketGroup.AttachmentPath,
		AttachmentSize:             ticketGroup.AttachmentSize,
		ContentType:                ticketGroup.ContentType,
		UniqueExtension:            ticketGroup.UniqueExtension,
		ActiveStartDate:            nullStringToPointer(ticketGroup.ActiveStartDate),
		ActiveEndDate:              nullStringToPointer(ticketGroup.ActiveEndDate),
		IsActive:                   ticketGroup.IsActive,
		IsTicketInternal:           ticketGroup.IsTicketInternal,
		TicketIds:                  ticketGroup.TicketIds.String,
		Tags:                       tagDTOs,
		GroupGallery:               galleryItems,
		TicketDetails:              ticketDetails,
		TicketVariants:             ticketVariants,
		LocationAddress:            ticketGroup.LocationAddress,
		LocationMapEmbedUrl:        ticketGroup.LocationMapUrl,
		OrganiserNameBm:            ticketGroup.OrganiserNameBm,
		OrganiserNameEn:            ticketGroup.OrganiserNameEn,
		OrganiserNameCn:            ticketGroup.OrganiserNameCn,
		OrganiserAddress:           ticketGroup.OrganiserAddress,
		OrganiserDescriptionHtmlBm: ticketGroup.OrganiserDescHtmlBm,
		OrganiserDescriptionHtmlEn: ticketGroup.OrganiserDescHtmlEn,
		OrganiserDescriptionHtmlCn: ticketGroup.OrganiserDescHtmlCn,
		OrganiserContact:           nullStringToPointer(ticketGroup.OrganiserContact),
		OrganiserEmail:             nullStringToPointer(ticketGroup.OrganiserEmail),
		OrganiserWebsite:           nullStringToPointer(ticketGroup.OrganiserWebsite),
		OrganiserFacilitiesBm:      organiserFacilitiesBm,
		OrganiserFacilitiesEn:      organiserFacilitiesEn,
		OrganiserFacilitiesCn:      organiserFacilitiesCn,
		CreatedAt:                  ticketGroup.CreatedAt.Format(time.RFC3339),
		UpdatedAt:                  ticketGroup.UpdatedAt.Format(time.RFC3339),
	}

	// 8. Handle optional date fields
	if ticketGroup.ActiveStartDate.Valid {
		profile.ActiveStartDate = nullStringToPointer(ticketGroup.ActiveStartDate)
	}
	if ticketGroup.ActiveEndDate.Valid {
		profile.ActiveEndDate = nullStringToPointer(ticketGroup.ActiveEndDate)
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
			TitleEn:        detail.TitleEn,
			TitleBm:        detail.TitleBm,
			TitleCn:        detail.TitleCn,
			TitleIcon:      detail.TitleIcon,
			RawHtmlBm:      detail.RawHtmlBm,
			RawHtmlEn:      detail.RawHtmlEn,
			RawHtmlCn:      detail.RawHtmlCn,
			DisplayFlag:    detail.DisplayFlag,
		})
	}

	return detailDTOs, nil
}

func (s *TicketGroupService) getLocalTicketVariants(ticketGroupId uint) ([]dto.TicketVariantDTO, error) {
	// This would be implemented by calling a repository method
	// Create a TicketVariantRepository or use this from another service
	variants, err := s.ticketVariantRepo.FindByTicketGroupID(ticketGroupId)
	if err != nil {
		return nil, err
	}

	variantDTOs := make([]dto.TicketVariantDTO, 0, len(variants))
	for _, variant := range variants {
		ticketIdStr := strconv.FormatUint(uint64(variant.TicketVariantId), 10)
		variantDTOs = append(variantDTOs, dto.TicketVariantDTO{
			TicketVariantId: &variant.TicketVariantId,
			TicketGroupId:   &variant.TicketGroupId,
			TicketId:        &ticketIdStr,
			NameBm:          variant.NameBm,
			NameEn:          variant.NameEn,
			NameCn:          variant.NameCn,
			DescBm:          variant.DescBm,
			DescEn:          variant.DescEn,
			DescCn:          variant.DescCn,
			UnitPrice:       variant.UnitPrice,
		})
	}

	return variantDTOs, nil
}

// GetTicketVariants retrieves ticket variants for a specific ticket group and date
func (s *TicketGroupService) GetTicketVariants(ticketGroupId uint, date string) (*dto.TicketVariantResponse, error) {
	// First, check if the ticket group exists
	ticketGroup, err := s.ticketGroupRepo.FindByID(ticketGroupId)
	if err != nil {
		return nil, err
	}

	ticketVariants := make([]dto.TicketVariantDTO, 0)
	if !ticketGroup.IsTicketInternal {
		// Get available ticket items from the external API
		ticketItems, err := s.zooAPIClient.GetTicketItems(ticketGroup.GroupNameBm, date)
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

		for _, item := range ticketItems {
			// Skip if we have a filter and this item is not in the allowed list
			if allowedTicketIDs != nil && !allowedTicketIDs[item.ItemId] {
				continue
			}

			variant := dto.TicketVariantDTO{
				TicketVariantId: nil,
				TicketGroupId:   nil,
				TicketId:        &item.ItemId,
				NameBm:          item.ItemDescription,
				NameEn:          item.ItemDesc1,
				NameCn:          item.ItemDesc2,
				DescBm:          "",
				DescEn:          "",
				DescCn:          "",
				UnitPrice:       item.UnitPrice,
				PrintType:       item.PrintType,
			}

			ticketVariants = append(ticketVariants, variant)
		}
	} else {
		ticketItems, err := s.ticketVariantRepo.FindByTicketGroupID(ticketGroup.TicketGroupId)
		if err != nil {
			return nil, err
		}

		for _, item := range ticketItems {
			ticketIdStr := strconv.FormatUint(uint64(item.TicketVariantId), 10)
			variant := dto.TicketVariantDTO{
				TicketVariantId: &item.TicketVariantId,
				TicketGroupId:   &item.TicketGroupId,
				TicketId:        &ticketIdStr, // Convert uint to string
				NameBm:          item.NameBm,
				NameEn:          item.NameEn,
				NameCn:          item.NameCn,
				DescBm:          item.DescBm,
				DescEn:          item.DescEn,
				DescCn:          item.DescCn,
				UnitPrice:       item.UnitPrice,
				PrintType:       "",
			}

			ticketVariants = append(ticketVariants, variant)
		}
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

// CreateTicketGroup creates a new ticket group with all related data
func (s *TicketGroupService) CreateTicketGroup(
	req *dto.CreateTicketGroupRequest,
	attachment *multipart.FileHeader,
	galleries []*multipart.FileHeader,
) (*models.TicketGroup, error) {
	maxPlacement, err := s.ticketGroupRepo.GetMaxPlacement()
	if err != nil {
		// If no ticket groups exist, start with placement 1
		maxPlacement = 0
	}

	// Begin transaction
	tx := s.ticketGroupRepo.Db.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Handle main attachment upload
	storagePath := os.Getenv("TICKET_GROUP_STORAGE_PATH")
	if storagePath == "" {
		tx.Rollback()
		return nil, errors.New("TICKET_GROUP_STORAGE_PATH environment variable not set")
	}

	fileUtil := utils.NewFileUtil()
	uniqueFileName, err := fileUtil.UploadAttachmentFile(attachment, storagePath)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to upload attachment: %w", err)
	}

	// 2. Create ticket group
	ticketGroup := &models.TicketGroup{
		Placement:              maxPlacement + 1,
		OrderTicketLimit:       req.OrderTicketLimit,
		ScanSetting:            req.ScanSetting,
		GroupType:              "ongoing",
		GroupNameBm:            req.GroupNameBm,
		GroupNameEn:            req.GroupNameEn,
		GroupNameCn:            req.GroupNameCn,
		GroupDescBm:            req.GroupDescBm,
		GroupDescEn:            req.GroupDescEn,
		GroupDescCn:            req.GroupDescCn,
		GroupRedirectionSpanBm: nullStringFromString(req.GroupRedirectionSpanBm),
		GroupRedirectionSpanEn: nullStringFromString(req.GroupRedirectionSpanEn),
		GroupRedirectionSpanCn: nullStringFromString(req.GroupRedirectionSpanCn),
		GroupRedirectionUrl:    nullStringFromString(req.GroupRedirectionUrl),
		GroupSlot1Bm:           nullStringFromString(req.GroupSlot1Bm),
		GroupSlot1En:           nullStringFromString(req.GroupSlot1En),
		GroupSlot1Cn:           nullStringFromString(req.GroupSlot1Cn),
		GroupSlot2Bm:           nullStringFromString(req.GroupSlot2Bm),
		GroupSlot2En:           nullStringFromString(req.GroupSlot2En),
		GroupSlot2Cn:           nullStringFromString(req.GroupSlot2Cn),
		GroupSlot3Bm:           nullStringFromString(req.GroupSlot3Bm),
		GroupSlot3En:           nullStringFromString(req.GroupSlot3En),
		GroupSlot3Cn:           nullStringFromString(req.GroupSlot3Cn),
		GroupSlot4Bm:           nullStringFromString(req.GroupSlot4Bm),
		GroupSlot4En:           nullStringFromString(req.GroupSlot4En),
		GroupSlot4Cn:           nullStringFromString(req.GroupSlot4Cn),
		PricePrefixBm:          req.PricePrefixBm,
		PricePrefixEn:          req.PricePrefixEn,
		PricePrefixCn:          req.PricePrefixCn,
		PriceSuffixBm:          req.PriceSuffixBm,
		PriceSuffixEn:          req.PriceSuffixEn,
		PriceSuffixCn:          req.PriceSuffixCn,
		AttachmentName:         attachment.Filename,
		AttachmentPath:         storagePath,
		AttachmentSize:         attachment.Size,
		ContentType:            attachment.Header.Get("Content-Type"),
		UniqueExtension:        uniqueFileName,
		ActiveStartDate:        nullStringFromString(req.ActiveStartDate),
		ActiveEndDate:          nullStringFromString(req.ActiveEndDate),
		IsActive:               req.IsActive,
		IsTicketInternal:       true, // Set based on your business logic
		LocationAddress:        req.LocationAddress,
		LocationMapUrl:         req.LocationMapUrl,
		OrganiserNameBm:        req.OrganiserNameBm,
		OrganiserNameEn:        req.OrganiserNameEn,
		OrganiserNameCn:        req.OrganiserNameCn,
		OrganiserAddress:       req.OrganiserAddress,
		OrganiserDescHtmlBm:    req.OrganiserDescHtmlBm,
		OrganiserDescHtmlEn:    req.OrganiserDescHtmlEn,
		OrganiserDescHtmlCn:    req.OrganiserDescHtmlCn,
		OrganiserContact:       nullStringFromString(req.OrganiserContact),
		OrganiserEmail:         nullStringFromString(req.OrganiserEmail),
		OrganiserWebsite:       nullStringFromString(req.OrganiserWebsite),
		OrganiserFacilitiesBm:  nullStringFromString(req.OrganiserFacilitiesBm),
		OrganiserFacilitiesEn:  nullStringFromString(req.OrganiserFacilitiesEn),
		OrganiserFacilitiesCn:  nullStringFromString(req.OrganiserFacilitiesCn),
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}

	// Save ticket group
	if err := tx.Create(ticketGroup).Error; err != nil {
		tx.Rollback()
		fileUtil.DeleteAttachmentFile(uniqueFileName, storagePath)
		return nil, fmt.Errorf("failed to create ticket group: %w", err)
	}

	// 3. Create ticket details
	for _, detail := range req.TicketDetails {
		ticketDetail := &models.TicketDetail{
			TicketGroupId: ticketGroup.TicketGroupId,
			TitleEn:       detail.TitleEn,
			TitleBm:       detail.TitleBm,
			TitleCn:       detail.TitleCn,
			TitleIcon:     detail.TitleIcon,
			RawHtmlBm:     detail.RawHtmlBm,
			RawHtmlEn:     detail.RawHtmlEn,
			RawHtmlCn:     detail.RawHtmlCn,
			DisplayFlag:   detail.DisplayFlag,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := tx.Create(ticketDetail).Error; err != nil {
			tx.Rollback()
			fileUtil.DeleteAttachmentFile(uniqueFileName, storagePath)
			return nil, fmt.Errorf("failed to create ticket detail: %w", err)
		}
	}

	// 4. Create ticket variants
	for _, variant := range req.TicketVariants {
		ticketVariant := &models.TicketVariant{
			TicketGroupId: ticketGroup.TicketGroupId,
			NameBm:        variant.NameBm,
			NameEn:        variant.NameEn,
			NameCn:        variant.NameCn,
			DescBm:        variant.DescBm,
			DescEn:        variant.DescEn,
			DescCn:        variant.DescCn,
			UnitPrice:     variant.UnitPrice,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := tx.Create(ticketVariant).Error; err != nil {
			tx.Rollback()
			fileUtil.DeleteAttachmentFile(uniqueFileName, storagePath)
			return nil, fmt.Errorf("failed to create ticket variant: %w", err)
		}
	}

	// 5. Handle gallery uploads
	galleryPath := os.Getenv("GROUP_GALLERY_STORAGE_PATH")
	if galleryPath == "" {
		galleryPath = storagePath // Use same path as main attachment if not specified
	}

	var uploadedGalleries []string
	for _, gallery := range galleries {
		galleryFileName, err := fileUtil.UploadAttachmentFile(gallery, galleryPath)
		if err != nil {
			// Rollback and clean up
			tx.Rollback()
			fileUtil.DeleteAttachmentFile(uniqueFileName, storagePath)
			for _, uploaded := range uploadedGalleries {
				fileUtil.DeleteAttachmentFile(uploaded, galleryPath)
			}
			return nil, fmt.Errorf("failed to upload gallery image: %w", err)
		}
		uploadedGalleries = append(uploadedGalleries, galleryFileName)

		// Create gallery record
		groupGallery := &models.GroupGallery{
			TicketGroupId:   ticketGroup.TicketGroupId,
			AttachmentName:  gallery.Filename,
			AttachmentPath:  galleryPath,
			AttachmentSize:  gallery.Size,
			ContentType:     gallery.Header.Get("Content-Type"),
			UniqueExtension: galleryFileName,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}
		if err := tx.Create(groupGallery).Error; err != nil {
			tx.Rollback()
			fileUtil.DeleteAttachmentFile(uniqueFileName, storagePath)
			for _, uploaded := range uploadedGalleries {
				fileUtil.DeleteAttachmentFile(uploaded, galleryPath)
			}
			return nil, fmt.Errorf("failed to create gallery record: %w", err)
		}
	}

	// 6. Create ticket tags associations
	for _, tagReq := range req.TicketTags {
		// First verify the tag exists
		tag, err := s.tagRepo.FindByID(tagReq.TagId)
		if err != nil {
			tx.Rollback()
			fileUtil.DeleteAttachmentFile(uniqueFileName, storagePath)
			for _, uploaded := range uploadedGalleries {
				fileUtil.DeleteAttachmentFile(uploaded, galleryPath)
			}
			return nil, fmt.Errorf("tag with ID %d not found: %w", tagReq.TagId, err)
		}

		// Create the ticket tag association
		ticketTag := &models.TicketTag{
			TicketGroupId: ticketGroup.TicketGroupId,
			TagId:         tag.TagId,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := tx.Create(ticketTag).Error; err != nil {
			tx.Rollback()
			fileUtil.DeleteAttachmentFile(uniqueFileName, storagePath)
			for _, uploaded := range uploadedGalleries {
				fileUtil.DeleteAttachmentFile(uploaded, galleryPath)
			}
			return nil, fmt.Errorf("failed to create ticket tag association: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		// Clean up uploaded files
		fileUtil.DeleteAttachmentFile(uniqueFileName, storagePath)
		for _, uploaded := range uploadedGalleries {
			fileUtil.DeleteAttachmentFile(uploaded, galleryPath)
		}
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return ticketGroup, nil
}

// UpdatePlacements updates the placement values for multiple ticket groups
func (s *TicketGroupService) UpdatePlacements(placements []dto.PlacementItem) error {
	// Begin transaction
	tx := s.ticketGroupRepo.Db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// First, verify all ticket groups exist
	ticketGroupIds := make([]uint, len(placements))
	for i, item := range placements {
		ticketGroupIds[i] = item.TicketGroupId
	}

	var existingGroups []models.TicketGroup
	if err := tx.Where("ticket_group_id IN ?", ticketGroupIds).Find(&existingGroups).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to find ticket groups: %w", err)
	}

	// Check if all requested ticket groups exist
	if len(existingGroups) != len(placements) {
		tx.Rollback()

		// Find which IDs don't exist
		existingMap := make(map[uint]bool)
		for _, group := range existingGroups {
			existingMap[group.TicketGroupId] = true
		}

		var missingIds []uint
		for _, item := range placements {
			if !existingMap[item.TicketGroupId] {
				missingIds = append(missingIds, item.TicketGroupId)
			}
		}

		return fmt.Errorf("ticket groups not found: %v", missingIds)
	}

	// Update each ticket group's placement
	for _, item := range placements {
		if err := tx.Model(&models.TicketGroup{}).
			Where("ticket_group_id = ?", item.TicketGroupId).
			Updates(map[string]interface{}{
				"placement":  item.Placement,
				"updated_at": time.Now(),
			}).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update placement for ticket group %d: %w", item.TicketGroupId, err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *TicketGroupService) UpdateTicketGroupImage(ticketGroupId uint, attachment *multipart.FileHeader) error {
	ticketGroup, err := s.ticketGroupRepo.FindByID(ticketGroupId)
	if err != nil {
		log.Printf("Error finding ticket group %s: %v", ticketGroupId, err)
	}

	storagePath := os.Getenv("TICKET_GROUP_STORAGE_PATH")
	if storagePath == "" {
		return errors.New("TICKET_GROUP_STORAGE_PATH environment variable not set")
	}

	fileUtil := utils.NewFileUtil()
	uniqueFileName, err := fileUtil.UploadAttachmentFile(attachment, storagePath)
	if err != nil {
		return fmt.Errorf("failed to upload attachment: %w", err)
	}

	ticketGroup.AttachmentName = attachment.Filename
	ticketGroup.AttachmentPath = storagePath
	ticketGroup.AttachmentSize = attachment.Size
	ticketGroup.ContentType = attachment.Header.Get("Content-Type")
	ticketGroup.UniqueExtension = uniqueFileName

	err = s.ticketGroupRepo.Update(ticketGroup)
	if err != nil {
		log.Printf("Error updating attachment: %v", err)
		return err
	}

	return nil
}

func (s *TicketGroupService) UpdateTicketGroupBasicInfo(basicInfo dto.UpdateTicketGroupBasicInfoRequest) error {
	// Begin transaction
	tx := s.ticketGroupRepo.Db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	ticketGroup, err := s.ticketGroupRepo.FindByID(basicInfo.TicketGroupId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("ticket group not found: %w", err)
	}

	ticketGroup.OrderTicketLimit = basicInfo.OrderTicketLimit
	ticketGroup.ScanSetting = basicInfo.ScanSetting
	ticketGroup.GroupNameBm = basicInfo.GroupNameBm
	ticketGroup.GroupNameEn = basicInfo.GroupNameEn
	ticketGroup.GroupNameCn = basicInfo.GroupNameCn
	ticketGroup.GroupDescBm = basicInfo.GroupDescBm
	ticketGroup.GroupDescEn = basicInfo.GroupDescEn
	ticketGroup.GroupDescCn = basicInfo.GroupDescCn
	ticketGroup.GroupRedirectionSpanBm = nullStringFromString(basicInfo.GroupRedirectionSpanBm)
	ticketGroup.GroupRedirectionSpanEn = nullStringFromString(basicInfo.GroupRedirectionSpanEn)
	ticketGroup.GroupRedirectionSpanCn = nullStringFromString(basicInfo.GroupRedirectionSpanCn)
	ticketGroup.GroupRedirectionUrl = nullStringFromString(basicInfo.GroupRedirectionUrl)
	ticketGroup.GroupSlot1Bm = nullStringFromString(basicInfo.GroupSlot1Bm)
	ticketGroup.GroupSlot1En = nullStringFromString(basicInfo.GroupSlot1En)
	ticketGroup.GroupSlot1Cn = nullStringFromString(basicInfo.GroupSlot1Cn)
	ticketGroup.GroupSlot2Bm = nullStringFromString(basicInfo.GroupSlot2Bm)
	ticketGroup.GroupSlot2En = nullStringFromString(basicInfo.GroupSlot2En)
	ticketGroup.GroupSlot2Cn = nullStringFromString(basicInfo.GroupSlot2Cn)
	ticketGroup.GroupSlot3Bm = nullStringFromString(basicInfo.GroupSlot3Bm)
	ticketGroup.GroupSlot3En = nullStringFromString(basicInfo.GroupSlot3En)
	ticketGroup.GroupSlot3Cn = nullStringFromString(basicInfo.GroupSlot3Cn)
	ticketGroup.GroupSlot4Bm = nullStringFromString(basicInfo.GroupSlot4Bm)
	ticketGroup.GroupSlot4En = nullStringFromString(basicInfo.GroupSlot4En)
	ticketGroup.GroupSlot4Cn = nullStringFromString(basicInfo.GroupSlot4Cn)
	ticketGroup.PricePrefixBm = basicInfo.PricePrefixBm
	ticketGroup.PricePrefixEn = basicInfo.PricePrefixEn
	ticketGroup.PricePrefixCn = basicInfo.PricePrefixCn
	ticketGroup.PriceSuffixBm = basicInfo.PriceSuffixBm
	ticketGroup.PriceSuffixEn = basicInfo.PriceSuffixEn
	ticketGroup.PriceSuffixCn = basicInfo.PriceSuffixCn
	ticketGroup.ActiveStartDate = nullStringFromString(basicInfo.ActiveStartDate)
	ticketGroup.ActiveEndDate = nullStringFromString(basicInfo.ActiveEndDate)
	ticketGroup.IsActive = basicInfo.IsActive
	ticketGroup.LocationAddress = basicInfo.LocationAddress
	ticketGroup.LocationMapUrl = basicInfo.LocationMapUrl

	if err := tx.Save(ticketGroup).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update ticket group basic info: %w", err)
	}

	if len(basicInfo.TicketTags) > 0 {
		if err := s.updateTicketTags(tx, basicInfo.TicketGroupId, basicInfo.TicketTags); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update ticket tags: %w", err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Helper method to handle ticket tags updates
func (s *TicketGroupService) updateTicketTags(tx *gorm.DB, ticketGroupId uint, newTags []dto.TicketTagsRequest) error {
	// 1. Get existing ticket tags for this ticket group
	var existingTicketTags []models.TicketTag
	if err := tx.Where("ticket_group_id = ?", ticketGroupId).Find(&existingTicketTags).Error; err != nil {
		return fmt.Errorf("failed to fetch existing ticket tags: %w", err)
	}

	// 2. Create maps for comparison
	existingTagsMap := make(map[uint]bool)
	for _, existingTag := range existingTicketTags {
		existingTagsMap[existingTag.TagId] = true
	}

	newTagsMap := make(map[uint]bool)
	var newTagIds []uint
	for _, newTag := range newTags {
		// Validate that the tag exists
		var tag models.Tag
		if err := tx.First(&tag, newTag.TagId).Error; err != nil {
			return fmt.Errorf("tag with ID %d not found: %w", newTag.TagId, err)
		}

		newTagsMap[newTag.TagId] = true
		newTagIds = append(newTagIds, newTag.TagId)
	}

	// 3. Find tags to add (new tags that don't exist in current tags)
	var tagsToAdd []uint
	for _, tagId := range newTagIds {
		if !existingTagsMap[tagId] {
			tagsToAdd = append(tagsToAdd, tagId)
		}
	}

	// 4. Find tags to remove (existing tags that are not in new tags)
	var tagsToRemove []uint
	for _, existingTag := range existingTicketTags {
		if !newTagsMap[existingTag.TagId] {
			tagsToRemove = append(tagsToRemove, existingTag.TagId)
		}
	}

	// 5. Add new tags
	for _, tagId := range tagsToAdd {
		ticketTag := models.TicketTag{
			TicketGroupId: ticketGroupId,
			TagId:         tagId,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := tx.Create(&ticketTag).Error; err != nil {
			return fmt.Errorf("failed to create ticket tag for tag ID %d: %w", tagId, err)
		}
	}

	// 6. Remove tags that are no longer needed
	if len(tagsToRemove) > 0 {
		if err := tx.Where("ticket_group_id = ? AND tag_id IN ?", ticketGroupId, tagsToRemove).
			Delete(&models.TicketTag{}).Error; err != nil {
			return fmt.Errorf("failed to remove ticket tags: %w", err)
		}
	}

	return nil
}

func (s *TicketGroupService) UploadTicketGroupGallery(ticketGroupId uint, attachment *multipart.FileHeader) error {
	_, err := s.ticketGroupRepo.FindByID(ticketGroupId)
	if err != nil {
		log.Printf("Error finding ticket group %s: %v", ticketGroupId, err)
	}

	storagePath := os.Getenv("GROUP_GALLERY_STORAGE_PATH")
	if storagePath == "" {
		return errors.New("GROUP_GALLERY_STORAGE_PATH environment variable not set")
	}

	fileUtil := utils.NewFileUtil()
	uniqueFileName, err := fileUtil.UploadAttachmentFile(attachment, storagePath)
	if err != nil {
		return fmt.Errorf("failed to upload attachment: %w", err)
	}

	groupGallery := &models.GroupGallery{
		TicketGroupId:   ticketGroupId,
		AttachmentName:  attachment.Filename,
		AttachmentPath:  storagePath,
		AttachmentSize:  attachment.Size,
		ContentType:     attachment.Header.Get("Content-Type"),
		UniqueExtension: uniqueFileName,
	}

	err = s.groupGalleryRepo.Create(groupGallery)
	if err != nil {
		log.Printf("Error creating group gallery: %v", err)
		return err
	}

	return nil
}

func (s *TicketGroupService) DeleteTicketGroupGallery(groupGalleryId uint) (*models.TicketGroup, error) {
	groupGallery, err := s.groupGalleryRepo.FindByID(groupGalleryId)
	if err != nil {
		log.Printf("Error finding group gallery %s: %v", groupGalleryId, err)
	}

	ticketGroup, err := s.ticketGroupRepo.FindByID(groupGallery.TicketGroupId)
	if err != nil {
		log.Printf("Error finding ticket group %s: %v", groupGallery.TicketGroupId, err)
	}

	storagePath := os.Getenv("GROUP_GALLERY_STORAGE_PATH")
	if storagePath == "" {
		return nil, errors.New("GROUP_GALLERY_STORAGE_PATH environment variable not set")
	}

	fileUtil := utils.NewFileUtil()
	fileUtil.DeleteAttachmentFile(groupGallery.UniqueExtension, storagePath)

	err = s.groupGalleryRepo.Delete(groupGalleryId)
	if err != nil {
		log.Printf("Error deleting group gallery: %v", err)
		return nil, err
	}

	return ticketGroup, nil
}

func (s *TicketGroupService) UpdateTicketGroupDetails(details dto.UpdateTicketGroupDetailsRequest) error {
	// Begin transaction
	tx := s.ticketGroupRepo.Db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Verify that the ticket group exists
	ticketGroup, err := s.ticketGroupRepo.FindByID(details.TicketGroupId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("ticket group not found: %w", err)
	}

	// 2. Get existing ticket details for this ticket group to build a validation map
	existingDetails, err := s.ticketDetailRepo.FindByTicketGroupID(details.TicketGroupId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to fetch existing ticket details: %w", err)
	}

	// 3. Create a map of valid ticket detail IDs for this ticket group
	validDetailIDs := make(map[uint]bool)
	for _, detail := range existingDetails {
		validDetailIDs[detail.TicketDetailId] = true
	}

	// 4. Process each ticket detail update
	for _, updateDetail := range details.TicketDetails {
		// Verify that the ticket detail ID belongs to the specified ticket group
		if !validDetailIDs[updateDetail.TicketDetailId] {
			tx.Rollback()
			return fmt.Errorf("ticket detail ID %d does not belong to ticket group ID %d",
				updateDetail.TicketDetailId, details.TicketGroupId)
		}

		// Update the ticket detail
		updateData := map[string]interface{}{
			"title_bm":     updateDetail.TitleBm,
			"title_en":     updateDetail.TitleEn,
			"title_cn":     updateDetail.TitleCn,
			"title_icon":   updateDetail.TitleIcon,
			"raw_html_bm":  updateDetail.RawHtmlBm,
			"raw_html_en":  updateDetail.RawHtmlEn,
			"raw_html_cn":  updateDetail.RawHtmlCn,
			"display_flag": updateDetail.DisplayFlag,
			"updated_at":   time.Now(),
		}

		result := tx.Model(&models.TicketDetail{}).
			Where("ticket_detail_id = ? AND ticket_group_id = ?",
				updateDetail.TicketDetailId, details.TicketGroupId).
			Updates(updateData)

		if result.Error != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update ticket detail ID %d: %w",
				updateDetail.TicketDetailId, result.Error)
		}

		if result.RowsAffected == 0 {
			tx.Rollback()
			return fmt.Errorf("ticket detail ID %d not found or does not belong to ticket group ID %d",
				updateDetail.TicketDetailId, details.TicketGroupId)
		}
	}

	// 5. Update the ticket group's updated timestamp
	ticketGroup.UpdatedAt = time.Now()
	if err := tx.Save(ticketGroup).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update ticket group timestamp: %w", err)
	}

	// 6. Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *TicketGroupService) UpdateTicketGroupVariants(variant dto.UpdateTicketGroupVariantsRequest) error {
	// Begin transaction
	tx := s.ticketGroupRepo.Db.Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. Verify that the ticket group exists
	ticketGroup, err := s.ticketGroupRepo.FindByID(variant.TicketGroupId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("ticket group not found: %w", err)
	}

	// 2. Get existing ticket variants for this ticket group
	existingVariants, err := s.ticketVariantRepo.FindByTicketGroupID(variant.TicketGroupId)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to fetch existing ticket variants: %w", err)
	}

	// 3. Create maps for easier comparison
	existingVariantsMap := make(map[uint]*models.TicketVariant)
	for i := range existingVariants {
		existingVariantsMap[existingVariants[i].TicketVariantId] = &existingVariants[i]
	}

	// 4. Track which existing variants should be kept
	var variantsToKeep []uint

	// 5. Process incoming ticket variants
	for _, newVariant := range variant.TicketVariants {
		if newVariant.TicketVariantId != nil && *newVariant.TicketVariantId > 0 {
			// This is an update to an existing variant
			existingVariantId := *newVariant.TicketVariantId

			if existingVariant, exists := existingVariantsMap[existingVariantId]; exists {
				// Verify the variant belongs to the correct ticket group
				if existingVariant.TicketGroupId != variant.TicketGroupId {
					tx.Rollback()
					return fmt.Errorf("ticket variant ID %d does not belong to ticket group ID %d",
						existingVariantId, variant.TicketGroupId)
				}

				// Update existing variant
				existingVariant.NameBm = newVariant.NameBm
				existingVariant.NameEn = newVariant.NameEn
				existingVariant.NameCn = newVariant.NameCn
				existingVariant.DescBm = newVariant.DescBm
				existingVariant.DescEn = newVariant.DescEn
				existingVariant.DescCn = newVariant.DescCn
				existingVariant.UnitPrice = newVariant.UnitPrice
				existingVariant.UpdatedAt = time.Now()

				if err := tx.Save(existingVariant).Error; err != nil {
					tx.Rollback()
					return fmt.Errorf("failed to update ticket variant ID %d: %w", existingVariantId, err)
				}

				variantsToKeep = append(variantsToKeep, existingVariantId)
			} else {
				tx.Rollback()
				return fmt.Errorf("ticket variant with ID %d not found", existingVariantId)
			}
		} else {
			// This is a new variant to be created
			newTicketVariant := &models.TicketVariant{
				TicketGroupId: variant.TicketGroupId,
				NameBm:        newVariant.NameBm,
				NameEn:        newVariant.NameEn,
				NameCn:        newVariant.NameCn,
				DescBm:        newVariant.DescBm,
				DescEn:        newVariant.DescEn,
				DescCn:        newVariant.DescCn,
				UnitPrice:     newVariant.UnitPrice,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}

			if err := tx.Create(newTicketVariant).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create new ticket variant: %w", err)
			}
		}
	}

	// 6. Delete ticket variants that are no longer needed
	variantsToKeepMap := make(map[uint]bool)
	for _, id := range variantsToKeep {
		variantsToKeepMap[id] = true
	}

	for _, existingVariant := range existingVariants {
		if !variantsToKeepMap[existingVariant.TicketVariantId] {
			// This variant should be deleted
			if err := tx.Delete(&existingVariant).Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to delete ticket variant ID %d: %w", existingVariant.TicketVariantId, err)
			}
		}
	}

	// 7. Update the ticket group's updated timestamp
	ticketGroup.UpdatedAt = time.Now()
	if err := tx.Save(ticketGroup).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update ticket group timestamp: %w", err)
	}

	// 8. Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *TicketGroupService) UpdateTicketGroupOrganiserInfo(organiserInfo dto.UpdateTicketGroupOrganiserInfoRequest) error {
	ticketGroup, err := s.ticketGroupRepo.FindByID(organiserInfo.TicketGroupId)
	if err != nil {
		log.Printf("Error finding ticket group %s: %v", organiserInfo.TicketGroupId, err)
	}

	ticketGroup.OrganiserNameBm = organiserInfo.OrganiserNameBm
	ticketGroup.OrganiserNameEn = organiserInfo.OrganiserNameEn
	ticketGroup.OrganiserNameCn = organiserInfo.OrganiserNameCn
	ticketGroup.OrganiserAddress = organiserInfo.OrganiserAddress
	ticketGroup.OrganiserDescHtmlBm = organiserInfo.OrganiserDescHtmlBm
	ticketGroup.OrganiserDescHtmlEn = organiserInfo.OrganiserDescHtmlEn
	ticketGroup.OrganiserDescHtmlCn = organiserInfo.OrganiserDescHtmlCn
	ticketGroup.OrganiserContact = nullStringFromString(organiserInfo.OrganiserContact)
	ticketGroup.OrganiserEmail = nullStringFromString(organiserInfo.OrganiserEmail)
	ticketGroup.OrganiserWebsite = nullStringFromString(organiserInfo.OrganiserWebsite)
	ticketGroup.OrganiserFacilitiesBm = nullStringFromString(organiserInfo.OrganiserFacilitiesBm)
	ticketGroup.OrganiserFacilitiesEn = nullStringFromString(organiserInfo.OrganiserFacilitiesEn)
	ticketGroup.OrganiserFacilitiesCn = nullStringFromString(organiserInfo.OrganiserFacilitiesCn)

	err = s.ticketGroupRepo.Update(ticketGroup)
	if err != nil {
		log.Printf("Error updating organiser info: %v", err)
		return err
	}

	return nil
}

// Helper function to convert string to sql.NullString
func nullStringFromString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
