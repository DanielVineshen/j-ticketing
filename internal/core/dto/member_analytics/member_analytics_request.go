// File: j-ticketing/internal/core/dto/member_analytics/member_analytics_request.go
package dto

import (
	"j-ticketing/pkg/validation"
)

type MemberAnalyticsRequest struct {
	StartDate string `json:"startDate" validate:"required,max=255"`
	EndDate   string `json:"endDate" validate:"required,max=255"`
}

func (r *MemberAnalyticsRequest) Validate() error {
	return validation.ValidateStruct(r)
}
