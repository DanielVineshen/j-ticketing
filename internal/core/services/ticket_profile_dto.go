// File: internal/core/services/ticket_profile_dto.go
package service

// TicketProfileResponse represents the full API response
type TicketProfileResponse struct {
	RespCode int                 `json:"respCode"`
	RespDesc string              `json:"respDesc"`
	Result   TicketProfileResult `json:"result"`
}

// TicketProfileResult wraps the ticket profile
type TicketProfileResult struct {
	TicketProfile TicketProfileDTO `json:"ticketProfile"`
}

// TicketProfileDTO represents the complete ticket profile data
type TicketProfileDTO struct {
	// Outer Layer For Ticket Profile
	TicketGroupId    uint   `json:"ticketGroupId"`
	GroupType        string `json:"groupType"`
	GroupName        string `json:"groupName"`
	GroupDesc        string `json:"groupDesc"`
	OperatingHours   string `json:"operatingHours"`
	PricePrefix      string `json:"pricePrefix"`
	PriceSuffix      string `json:"priceSuffix"`
	AttachmentName   string `json:"attachmentName"`
	AttachmentPath   string `json:"attachmentPath"`
	AttachmentSize   int64  `json:"attachmentSize"`
	ContentType      string `json:"contentType"`
	UniqueExtension  string `json:"uniqueExtension"`
	IsActive         bool   `json:"isActive"`
	IsTicketInternal string `json:"isTicketInternal"`
	TicketIds        string `json:"ticketIds"`

	// Filter Tags
	Tags []TagDTO `json:"tags"`

	// Inner Layer For Ticket Profile
	GroupGallery []GroupGalleryDTO `json:"groupGallery"`

	ActiveStartDate string `json:"activeStartDate,omitempty"`
	ActiveEndDate   string `json:"activeEndDate,omitempty"`

	// Detail Tab
	TicketDetails []TicketDetailDTO `json:"ticketDetails"`

	LocationAddress     string `json:"locationAddress"`
	LocationMapEmbedUrl string `json:"locationMapEmbedUrl"`

	// Organiser Tab
	OrganiserName            string   `json:"organiserName"`
	OrganiserAddress         string   `json:"organiserAddress"`
	OrganiserDescriptionHtml string   `json:"organiserDescriptionHtml"`
	OrganiserContact         string   `json:"organiserContact"`
	OrganiserEmail           string   `json:"organiserEmail"`
	OrganiserWebsite         string   `json:"organiserWebsite"`
	OrganiserOperatingHours  string   `json:"organiserOperatingHours"`
	OrganiserFacilities      []string `json:"organiserFacilities"`

	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// GroupGalleryDTO represents a gallery item for a ticket group
type GroupGalleryDTO struct {
	GroupGalleryId  uint   `json:"groupGalleryId"`
	AttachmentName  string `json:"attachmentName"`
	AttachmentPath  string `json:"attachmentPath"`
	AttachmentSize  int64  `json:"attachmentSize"`
	ContentType     string `json:"contentType"`
	UniqueExtension string `json:"uniqueExtension"`
}

// TicketDetailDTO represents a detail item for a ticket
type TicketDetailDTO struct {
	TicketDetailId uint   `json:"ticketDetailId"`
	Title          string `json:"title"`
	TitleIcon      string `json:"titleIcon"`
	RawHtml        string `json:"rawHtml"`
	DisplayFlag    bool   `json:"displayFlag"`
}
