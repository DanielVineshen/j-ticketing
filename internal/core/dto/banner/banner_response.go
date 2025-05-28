package dto

type BannerListResponse struct {
	Banners []Banner `json:"banners"`
}

type Banner struct {
	BannerId        uint   `json:"bannerId"`
	Placement       int    `json:"placement"`
	RedirectURL     string `json:"redirectUrl"`
	UploadedBy      string `json:"uploadedBy"`
	ActiveEndDate   string `json:"activeEndDate"`
	ActiveStartDate string `json:"activeStartDate"`
	IsActive        bool   `json:"isActive"`
	Duration        int    `json:"duration"`
	AttachmentName  string `json:"attachmentName"`
	AttachmentPath  string `json:"attachmentPath"`
	AttachmentSize  int64  `json:"attachmentSize"`
	ContentType     string `json:"contentType"`
	UniqueExt       string `json:"uniqueExtension"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}
