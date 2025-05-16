package service

import (
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
)

// TicketGroupService handles business logic for ticket groups
type TicketGroupService struct {
	ticketGroupRepo *repositories.TicketGroupRepository
	bannerRepo      *repositories.BannerRepository
}

// NewTicketGroupService creates a new ticket group service
func NewTicketGroupService(
	ticketGroupRepo *repositories.TicketGroupRepository,
	bannerRepo *repositories.BannerRepository,
) *TicketGroupService {
	return &TicketGroupService{
		ticketGroupRepo: ticketGroupRepo,
		bannerRepo:      bannerRepo,
	}
}

// GetAllTicketGroups returns all ticket groups
func (s *TicketGroupService) GetAllTicketGroups() ([]models.TicketGroup, error) {
	return s.ticketGroupRepo.FindAll()
}

// GetTicketGroupByID returns a ticket group by ID
func (s *TicketGroupService) GetTicketGroupByID(id uint) (*models.TicketGroup, error) {
	return s.ticketGroupRepo.FindByID(id)
}

// CreateTicketGroup creates a new ticket group
func (s *TicketGroupService) CreateTicketGroup(ticketGroup *models.TicketGroup) error {
	return s.ticketGroupRepo.Create(ticketGroup)
}

// UpdateTicketGroup updates a ticket group
func (s *TicketGroupService) UpdateTicketGroup(ticketGroup *models.TicketGroup) error {
	return s.ticketGroupRepo.Update(ticketGroup)
}

// DeleteTicketGroup deletes a ticket group
func (s *TicketGroupService) DeleteTicketGroup(id uint) error {
	return s.ticketGroupRepo.Delete(id)
}

// GetTicketGroupWithBanners returns a ticket group with its banners
func (s *TicketGroupService) GetTicketGroupWithBanners(id uint) (*models.TicketGroup, []models.Banner, error) {
	ticketGroup, err := s.ticketGroupRepo.FindByID(id)
	if err != nil {
		return nil, nil, err
	}

	banners, err := s.bannerRepo.FindByTicketGroupID(id)
	if err != nil {
		return ticketGroup, nil, err
	}

	return ticketGroup, banners, nil
}
