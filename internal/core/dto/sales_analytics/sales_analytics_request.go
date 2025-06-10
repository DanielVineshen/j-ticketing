// File: j-ticketing/internal/core/dto/sales_analytics/sales_analytics_request.go
package dto

import (
	"j-ticketing/pkg/validation"
)

type SalesAnalyticsRequest struct {
	StartDate string `json:"startDate" validate:"required,max=255"`
	EndDate   string `json:"endDate" validate:"required,max=255"`
}

func (r *SalesAnalyticsRequest) Validate() error {
	return validation.ValidateStruct(r)
}
