package service

import "j-ticketing/internal/db/repositories"

// OnsiteVisitorsAnalyticsService handles operations related to serving onsiteVisitorsAnalytics
type OnsiteVisitorsAnalyticsService struct {
	customerRepo repositories.CustomerRepository
}

// NewOnsiteVisitorsAnalyticsService creates a onsiteVisitorsAnalytics notifications service
func NewOnsiteVisitorsAnalyticsService(customerRepo repositories.CustomerRepository) *OnsiteVisitorsAnalyticsService {
	return &OnsiteVisitorsAnalyticsService{
		customerRepo: customerRepo,
	}
}
