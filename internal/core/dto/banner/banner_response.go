// File: j-ticketing/internal/core/dto/banner/banner_response.go
package dto

type BannerListResponse struct {
	Banners []Banner `json:"banners"`
}

// NewBannerListResponse creates a new banner list response with guaranteed non-nil slice
func NewBannerListResponse(banners []Banner) *BannerListResponse {
	if banners == nil {
		banners = make([]Banner, 0)
	}
	return &BannerListResponse{
		Banners: banners,
	}
}

type Banner struct {
	BannerId        uint   `json:"bannerId"`
	Placement       int    `json:"placement"`
	RedirectURL     string `json:"redirectUrl"`
	UploadedBy      string `json:"uploadedBy"`
	ActiveEndDate   string `json:"activeEndDate"`   // yyyy-MM-dd format
	ActiveStartDate string `json:"activeStartDate"` // yyyy-MM-dd format
	IsActive        bool   `json:"isActive"`
	Duration        int    `json:"duration"`
	AttachmentName  string `json:"attachmentName"`
	AttachmentPath  string `json:"attachmentPath"`
	AttachmentSize  int64  `json:"attachmentSize"`
	ContentType     string `json:"contentType"`
	UniqueExt       string `json:"uniqueExtension"`
	CreatedAt       string `json:"createdAt"` // yyyy-MM-dd HH:mm:ss format (Malaysia time)
	UpdatedAt       string `json:"updatedAt"` // yyyy-MM-dd HH:mm:ss format (Malaysia time)
}
