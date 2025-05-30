// File: j-ticketing/internal/core/dto/banner/banner_request.go (Updated - Remove RedirectURL validation)
package dto

import (
	"j-ticketing/pkg/validation"
)

type CreateNewBannerRequest struct {
	RedirectURL     string `json:"redirectUrl"` // Removed validation - now optional
	UploadedBy      string `json:"uploadedBy" validate:"required,max=255"`
	ActiveEndDate   string `json:"activeEndDate" validate:"required,max=255"`
	ActiveStartDate string `json:"activeStartDate" validate:"required,max=255"`
	IsActive        bool   `json:"isActive"`
	Duration        int    `json:"duration" validate:"required,min=1"`
}

func (r *CreateNewBannerRequest) Validate() error {
	return validation.ValidateStruct(r)
}

type UpdateBannerRequest struct {
	BannerId        uint   `json:"bannerId" validate:"required"`
	RedirectURL     string `json:"redirectUrl"` // Removed validation - now optional
	UploadedBy      string `json:"uploadedBy" validate:"required,max=255"`
	ActiveEndDate   string `json:"activeEndDate" validate:"required,max=255"`
	ActiveStartDate string `json:"activeStartDate" validate:"required,max=255"`
	IsActive        bool   `json:"isActive"`
	Duration        int    `json:"duration" validate:"required,min=1"`
}

func (r *UpdateBannerRequest) Validate() error {
	return validation.ValidateStruct(r)
}

type DeleteBannerRequest struct {
	BannerId uint `json:"bannerId" validate:"required"`
}

func (r *DeleteBannerRequest) Validate() error {
	return validation.ValidateStruct(r)
}

type BannerPlacementUpdate struct {
	BannerId  uint `json:"bannerId" validate:"required"`
	Placement int  `json:"placement" validate:"required,min=1"`
}

type UpdateBannerPlacementsRequest struct {
	Banners []BannerPlacementUpdate `json:"banners" validate:"required,dive"`
}

func (r *UpdateBannerPlacementsRequest) Validate() error {
	return validation.ValidateStruct(r)
}
