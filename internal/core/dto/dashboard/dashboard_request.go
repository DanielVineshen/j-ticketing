// File: j-ticketing/internal/core/dto/dashboard/dashboard_request.go
package dto

import (
	"j-ticketing/pkg/validation"
)

type DashboardRequest struct {
	StartDate string `json:"startDate" validate:"required,max=255"`
	EndDate   string `json:"endDate" validate:"required,max=255"`
}

func (r *DashboardRequest) Validate() error {
	return validation.ValidateStruct(r)
}
