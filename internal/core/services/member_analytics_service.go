// File: j-ticketing/internal/core/services/member_analytics_service.go
package service

import (
	"fmt"
	dto "j-ticketing/internal/core/dto/member_analytics"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/utils"
	"math"
	"time"
)

type MemberAnalyticsService struct {
	customerRepo repositories.CustomerRepository
}

func NewMemberAnalyticsService(
	customerRepo repositories.CustomerRepository,
) *MemberAnalyticsService {
	return &MemberAnalyticsService{
		customerRepo: customerRepo,
	}
}

// GetTotalMembers retrieves total members analysis
func (s *MemberAnalyticsService) GetTotalMembers(startDate, endDate string) (*dto.TotalMembersResponse, error) {
	// Validate date range
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// Get all members (customers with password)
	allMembers, err := s.getAllMembers()
	if err != nil {
		return nil, fmt.Errorf("failed to get all members: %w", err)
	}

	// Get new members within the date range
	newMembersInPeriod, err := s.getMembersInDateRange(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get new members in period: %w", err)
	}

	// Calculate new member percentage
	var newMemberPercentage float64
	if len(allMembers) > 0 {
		newMemberPercentage = math.Round((float64(len(newMembersInPeriod))/float64(len(allMembers)))*100*100) / 100
	}

	// Generate member trend
	memberTrend := s.generateMemberTrend(allMembers, startDate, endDate)

	return &dto.TotalMembersResponse{
		NewMemberPercentage: newMemberPercentage,
		SumTotalMembers:     len(allMembers),
		MemberTrend:         memberTrend,
	}, nil
}

// GetMembersNetGrowth retrieves members net growth analysis
func (s *MemberAnalyticsService) GetMembersNetGrowth(startDate, endDate string) (*dto.MembersNetGrowthResponse, error) {
	// Validate date range
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// Get new members within the date range
	newMembersInPeriod, err := s.getMembersInDateRange(startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get new members in period: %w", err)
	}

	sumTotalNewMembers := len(newMembersInPeriod)
	sumTotalChurnedMembers := 0 // As per requirement, stick with 0
	sumTotalNetGrowth := sumTotalNewMembers - sumTotalChurnedMembers

	// Generate net growth trend
	netGrowthTrend := s.generateNetGrowthTrend(newMembersInPeriod, startDate, endDate)

	return &dto.MembersNetGrowthResponse{
		SumTotalNewMembers:     sumTotalNewMembers,
		SumTotalChurnedMembers: sumTotalChurnedMembers,
		SumTotalNetGrowth:      sumTotalNetGrowth,
		NetGrowthTrend:         netGrowthTrend,
	}, nil
}

// GetMembersByAgeGroup retrieves members by age group analysis
func (s *MemberAnalyticsService) GetMembersByAgeGroup(startDate, endDate string) (*dto.MembersByAgeGroupResponse, error) {
	// Validate date range
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// Get all members (customers with password)
	allMembers, err := s.getAllMembers()
	if err != nil {
		return nil, fmt.Errorf("failed to get all members: %w", err)
	}

	// Generate age group trend
	ageGroupTrend := s.generateAgeGroupTrend(allMembers)

	return &dto.MembersByAgeGroupResponse{
		MembersByAgeGroupTrend: ageGroupTrend,
	}, nil
}

// GetMembersByNationality retrieves members by nationality analysis
func (s *MemberAnalyticsService) GetMembersByNationality(startDate, endDate string) (*dto.MembersByNationalityResponse, error) {
	// Validate date range
	if err := s.validateDateRange(startDate, endDate); err != nil {
		return nil, err
	}

	// Get all members (customers with password)
	allMembers, err := s.getAllMembers()
	if err != nil {
		return nil, fmt.Errorf("failed to get all members: %w", err)
	}

	// Generate nationality trend
	nationalityTrend := s.generateNationalityTrend(allMembers)

	return &dto.MembersByNationalityResponse{
		MembersByNationalityTrend: nationalityTrend,
	}, nil
}

// Helper methods

// validateDateRange validates the date format and ensures endDate is after startDate
func (s *MemberAnalyticsService) validateDateRange(startDate, endDate string) error {
	// Parse dates using the expected format
	start, err := time.Parse(utils.DateOnlyFormat, startDate)
	if err != nil {
		return fmt.Errorf("invalid startDate format. Expected yyyy-MM-dd, got: %s", startDate)
	}

	end, err := time.Parse(utils.DateOnlyFormat, endDate)
	if err != nil {
		return fmt.Errorf("invalid endDate format. Expected yyyy-MM-dd, got: %s", endDate)
	}

	if end.Before(start) {
		return fmt.Errorf("endDate cannot be earlier than startDate")
	}

	return nil
}

// getAllMembers retrieves all customers with password (members)
func (s *MemberAnalyticsService) getAllMembers() ([]models.Customer, error) {
	// Get all customers
	customers, err := s.customerRepo.FindAll()
	if err != nil {
		return nil, err
	}

	// Filter customers with password (members)
	var members []models.Customer
	for _, customer := range customers {
		if customer.Password.Valid && customer.Password.String != "" {
			members = append(members, customer)
		}
	}

	return members, nil
}

// getMembersInDateRange retrieves members created within the date range
func (s *MemberAnalyticsService) getMembersInDateRange(startDate, endDate string) ([]models.Customer, error) {
	// Get all members
	allMembers, err := s.getAllMembers()
	if err != nil {
		return nil, err
	}

	// Filter members within date range
	var membersInRange []models.Customer
	for _, member := range allMembers {
		if s.isWithinDateRange(member.CreatedAt, startDate, endDate) {
			membersInRange = append(membersInRange, member)
		}
	}

	return membersInRange, nil
}

// isWithinDateRange checks if creation date is within the specified range
func (s *MemberAnalyticsService) isWithinDateRange(createdAt time.Time, startDate, endDate string) bool {
	// Convert created date to Malaysia time and extract date only
	malaysiaTime, err := utils.ToMalaysiaTime(createdAt)
	if err != nil {
		return false
	}

	dateOnly := malaysiaTime.Format(utils.DateOnlyFormat)
	return dateOnly >= startDate && dateOnly <= endDate
}

// extractDateFromCreatedAt extracts date part from created at time
func (s *MemberAnalyticsService) extractDateFromCreatedAt(createdAt time.Time) string {
	malaysiaTime, err := utils.ToMalaysiaTime(createdAt)
	if err != nil {
		return ""
	}
	return malaysiaTime.Format(utils.DateOnlyFormat)
}

// generateDateRange generates all dates between startDate and endDate
func (s *MemberAnalyticsService) generateDateRange(startDate, endDate string) []string {
	start, _ := time.Parse(utils.DateOnlyFormat, startDate)
	end, _ := time.Parse(utils.DateOnlyFormat, endDate)

	var dates []string
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		dates = append(dates, d.Format(utils.DateOnlyFormat))
	}

	return dates
}

// generateMemberTrend generates member trend data showing daily new members
func (s *MemberAnalyticsService) generateMemberTrend(allMembers []models.Customer, startDate, endDate string) []dto.MemberTrend {
	// Create daily member count aggregations
	dailyNewMembers := make(map[string]int)

	// Count new members for each day in the period
	for _, member := range allMembers {
		memberDate := s.extractDateFromCreatedAt(member.CreatedAt)
		if memberDate != "" && memberDate >= startDate && memberDate <= endDate {
			dailyNewMembers[memberDate]++
		}
	}

	// Generate date range and create trend data
	dates := s.generateDateRange(startDate, endDate)
	var memberTrend []dto.MemberTrend

	for _, date := range dates {
		memberTrend = append(memberTrend, dto.MemberTrend{
			Date:            date,
			TotalNewMembers: dailyNewMembers[date], // Will be 0 if no members on that date
		})
	}

	return memberTrend
}

// generateNetGrowthTrend generates net growth trend data
func (s *MemberAnalyticsService) generateNetGrowthTrend(newMembers []models.Customer, startDate, endDate string) []dto.NetGrowthTrend {
	// Create daily member count aggregations
	dailyNewMembers := make(map[string]int)

	for _, member := range newMembers {
		date := s.extractDateFromCreatedAt(member.CreatedAt)
		if date != "" {
			dailyNewMembers[date]++
		}
	}

	// Generate date range and create trend data
	dates := s.generateDateRange(startDate, endDate)
	var netGrowthTrend []dto.NetGrowthTrend

	for _, date := range dates {
		totalNewMembers := dailyNewMembers[date]
		totalChurnedMembers := 0 // As per requirement
		netGrowth := totalNewMembers - totalChurnedMembers

		netGrowthTrend = append(netGrowthTrend, dto.NetGrowthTrend{
			Date:                date,
			TotalNewMembers:     totalNewMembers,
			TotalChurnedMembers: totalChurnedMembers,
			NetGrowth:           netGrowth,
		})
	}

	return netGrowthTrend
}

// generateAgeGroupTrend generates age group trend with member percentages based on valid Malaysian IC holders only
func (s *MemberAnalyticsService) generateAgeGroupTrend(allMembers []models.Customer) []dto.AgeGroupTrend {
	ageGroupMembers := make(map[string]*dto.AgeGroupTrend)
	currentYear := time.Now().Year()

	// Initialize age groups
	ageGroups := []string{"0-12", "13-17", "18-35", "36-50", "51+"}
	for _, ageGroup := range ageGroups {
		ageGroupMembers[ageGroup] = &dto.AgeGroupTrend{
			AgeGroup:          ageGroup,
			TotalMembers:      0,
			MembersPercentage: 0,
		}
	}

	// Count total members with valid Malaysian IC first
	totalMalaysianICMembers := 0
	for _, member := range allMembers {
		if utils.IsMalaysianIC(member.IdentificationNo) {
			age := utils.ExtractAgeFromMalaysianIC(member.IdentificationNo, currentYear)
			ageGroup := utils.CategorizeAge(age)

			// Only count if it's not "Unknown" (which means it's a valid age)
			if ageGroup != "Unknown" {
				totalMalaysianICMembers++
			}
		}
	}

	// Aggregate members by age group (only for valid Malaysian ICs)
	for _, member := range allMembers {
		// Only process if it's a valid Malaysian IC
		if utils.IsMalaysianIC(member.IdentificationNo) {
			age := utils.ExtractAgeFromMalaysianIC(member.IdentificationNo, currentYear)
			ageGroup := utils.CategorizeAge(age)

			// Only process if it's not "Unknown" (which means it's a valid age)
			if ageGroup != "Unknown" {
				if memberData, exists := ageGroupMembers[ageGroup]; exists {
					memberData.TotalMembers++
				}
			}
		}
		// If not Malaysian IC or Unknown age group, we skip this member
	}

	// Calculate percentages based on totalMalaysianICMembers and convert to slice
	var ageGroupTrend []dto.AgeGroupTrend
	for _, ageGroup := range ageGroups { // Use the same order as initialization
		memberData := ageGroupMembers[ageGroup]
		if totalMalaysianICMembers > 0 {
			memberData.MembersPercentage = math.Round((float64(memberData.TotalMembers)/float64(totalMalaysianICMembers))*100*100) / 100
		}
		// Include all age groups, even those with 0 members
		ageGroupTrend = append(ageGroupTrend, *memberData)
	}

	return ageGroupTrend
}

// generateNationalityTrend generates nationality trend with member percentages
func (s *MemberAnalyticsService) generateNationalityTrend(allMembers []models.Customer) []dto.NationalityTrend {
	nationalityMembers := make(map[string]*dto.NationalityTrend)
	totalMembers := len(allMembers)

	// Initialize nationalities
	nationalities := []string{"Local", "International"}
	for _, nationality := range nationalities {
		nationalityMembers[nationality] = &dto.NationalityTrend{
			Nationality:       nationality,
			TotalMembers:      0,
			MembersPercentage: 0,
		}
	}

	// Aggregate members by nationality
	for _, member := range allMembers {
		nationality := utils.DetermineNationality(member.IdentificationNo)

		if memberData, exists := nationalityMembers[nationality]; exists {
			memberData.TotalMembers++
		}
	}

	// Calculate percentages and convert to slice
	var nationalityTrend []dto.NationalityTrend
	for _, memberData := range nationalityMembers {
		if totalMembers > 0 {
			memberData.MembersPercentage = math.Round((float64(memberData.TotalMembers)/float64(totalMembers))*100*100) / 100
		}
		// Only include nationalities that have members
		if memberData.TotalMembers > 0 {
			nationalityTrend = append(nationalityTrend, *memberData)
		}
	}

	return nationalityTrend
}
