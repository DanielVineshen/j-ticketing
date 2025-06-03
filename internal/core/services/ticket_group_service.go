// File: j-ticketing/internal/core/services/ticket_group_service.go
package service

import (
	"database/sql"
	"errors"
	"fmt"
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
				PrintType:       &item.PrintType,
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
				PrintType:       nil,
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
		OrderTicketLimit:       req.OrderTicketLimit,
		ScanSetting:            req.ScanSetting,
		GroupType:              "event", // You might want to add this to the request
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
			Title:         detail.Title,
			TitleIcon:     detail.TitleIcon,
			RawHtml:       detail.RawHtml,
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

	// 6. Create tags (if needed - not in the request but might be needed)
	// You can add tag creation logic here if required

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

// Helper function to convert string to sql.NullString
func nullStringFromString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
