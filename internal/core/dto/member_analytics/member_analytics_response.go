// File: j-ticketing/internal/core/dto/member_analytics/member_analytics_response.go
package dto

// TotalMembersResponse represents the response for total members API
type TotalMembersResponse struct {
	NewMemberPercentage float64       `json:"newMemberPercentage"`
	SumTotalMembers     int           `json:"sumTotalMembers"`
	MemberTrend         []MemberTrend `json:"memberTrend"`
}

type MemberTrend struct {
	Date            string `json:"date"`
	TotalNewMembers int    `json:"totalNewMembers"`
}

// MembersNetGrowthResponse represents the response for members net growth API
type MembersNetGrowthResponse struct {
	SumTotalNewMembers     int              `json:"sumTotalNewMembers"`
	SumTotalChurnedMembers int              `json:"sumTotalChurnedMembers"`
	SumTotalNetGrowth      int              `json:"sumTotalNetGrowth"`
	NetGrowthTrend         []NetGrowthTrend `json:"netGrowthTrend"`
}

type NetGrowthTrend struct {
	Date                string `json:"date"`
	TotalNewMembers     int    `json:"totalNewMembers"`
	TotalChurnedMembers int    `json:"totalChurnedMembers"`
	NetGrowth           int    `json:"netGrowth"`
}

// MembersByAgeGroupResponse represents the response for members by age group API
type MembersByAgeGroupResponse struct {
	MembersByAgeGroupTrend []AgeGroupTrend `json:"membersByAgeGroupTrend"`
}

type AgeGroupTrend struct {
	AgeGroup          string  `json:"ageGroup"`
	TotalMembers      int     `json:"totalMembers"`
	MembersPercentage float64 `json:"membersPercentage"`
}

// MembersByNationalityResponse represents the response for members by nationality API
type MembersByNationalityResponse struct {
	MembersByNationalityTrend []NationalityTrend `json:"membersByNationalityTrend"`
}

type NationalityTrend struct {
	Nationality       string  `json:"nationality"`
	TotalMembers      int     `json:"totalMembers"`
	MembersPercentage float64 `json:"membersPercentage"`
}
