// File: j-ticketing/internal/core/dto/ticket_group/ticket_group_response.go
package dto

// TicketGroupResponse represents the response structure for ticket groups
type TicketGroupResponse struct {
	TicketGroups []TicketGroupDTO `json:"ticketGroups"`
}

// TicketGroupDTO represents the data transfer object for a ticket group
type TicketGroupDTO struct {
	TicketGroupId          uint     `json:"ticketGroupId"`
	Placement              int      `json:"placement"`
	OrderTicketLimit       int      `json:"orderTicketLimit"`
	ScanSetting            string   `json:"scanSetting"`
	GroupType              string   `json:"groupType"`
	GroupNameBm            string   `json:"groupNameBm"`
	GroupNameEn            string   `json:"groupNameEn"`
	GroupNameCn            string   `json:"groupNameCn"`
	GroupDescBm            string   `json:"groupDescBm"`
	GroupDescEn            string   `json:"groupDescEn"`
	GroupDescCn            string   `json:"groupDescCn"`
	GroupRedirectionSpanBm *string  `json:"groupRedirectionSpanBm"`
	GroupRedirectionSpanEn *string  `json:"groupRedirectionSpanEn"`
	GroupRedirectionSpanCn *string  `json:"groupRedirectionSpanCn"`
	GroupRedirectionUrl    *string  `json:"groupRedirectionUrl"`
	GroupSlot1Bm           *string  `json:"groupSlot1Bm"`
	GroupSlot1En           *string  `json:"groupSlot1En"`
	GroupSlot1Cn           *string  `json:"groupSlot1Cn"`
	GroupSlot2Bm           *string  `json:"groupSlot2Bm"`
	GroupSlot2En           *string  `json:"groupSlot2En"`
	GroupSlot2Cn           *string  `json:"groupSlot2Cn"`
	GroupSlot3Bm           *string  `json:"groupSlot3Bm"`
	GroupSlot3En           *string  `json:"groupSlot3En"`
	GroupSlot3Cn           *string  `json:"groupSlot3Cn"`
	GroupSlot4Bm           *string  `json:"groupSlot4Bm"`
	GroupSlot4En           *string  `json:"groupSlot4En"`
	GroupSlot4Cn           *string  `json:"groupSlot4Cn"`
	PricePrefixBm          string   `json:"pricePrefixBm"`
	PricePrefixEn          string   `json:"pricePrefixEn"`
	PricePrefixCn          string   `json:"pricePrefixCn"`
	PriceSuffixBm          string   `json:"priceSuffixBm"`
	PriceSuffixEn          string   `json:"priceSuffixEn"`
	PriceSuffixCn          string   `json:"priceSuffixCn"`
	AttachmentName         string   `json:"attachmentName"`
	AttachmentPath         string   `json:"attachmentPath"`
	AttachmentSize         int64    `json:"attachmentSize"`
	ContentType            string   `json:"contentType"`
	UniqueExtension        string   `json:"uniqueExtension"`
	IsActive               bool     `json:"isActive"`
	Tags                   []TagDTO `json:"tags"`
}

// TagDTO represents the data transfer object for a tag
type TagDTO struct {
	TagId   uint   `json:"tagId"`
	TagName string `json:"tagName"`
	TagDesc string `json:"tagDesc"`
}

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
	TicketGroupId          uint    `json:"ticketGroupId"`
	Placement              int     `json:"placement"`
	OrderTicketLimit       int     `json:"orderTicketLimit"`
	ScanSetting            string  `json:"scanSetting"`
	GroupType              string  `json:"groupType"`
	GroupNameBm            string  `json:"groupNameBm"`
	GroupNameEn            string  `json:"groupNameEn"`
	GroupNameCn            string  `json:"groupNameCn"`
	GroupDescBm            string  `json:"groupDescBm"`
	GroupDescEn            string  `json:"groupDescEn"`
	GroupDescCn            string  `json:"groupDescCn"`
	GroupRedirectionSpanBm *string `json:"groupRedirectionSpanBm"`
	GroupRedirectionSpanEn *string `json:"groupRedirectionSpanEn"`
	GroupRedirectionSpanCn *string `json:"groupRedirectionSpanCn"`
	GroupRedirectionUrl    *string `json:"groupRedirectionUrl"`
	GroupSlot1Bm           *string `json:"groupSlot1Bm"`
	GroupSlot1En           *string `json:"groupSlot1En"`
	GroupSlot1Cn           *string `json:"groupSlot1Cn"`
	GroupSlot2Bm           *string `json:"groupSlot2Bm"`
	GroupSlot2En           *string `json:"groupSlot2En"`
	GroupSlot2Cn           *string `json:"groupSlot2Cn"`
	GroupSlot3Bm           *string `json:"groupSlot3Bm"`
	GroupSlot3En           *string `json:"groupSlot3En"`
	GroupSlot3Cn           *string `json:"groupSlot3Cn"`
	GroupSlot4Bm           *string `json:"groupSlot4Bm"`
	GroupSlot4En           *string `json:"groupSlot4En"`
	GroupSlot4Cn           *string `json:"groupSlot4Cn"`
	PricePrefixBm          string  `json:"pricePrefixBm"`
	PricePrefixEn          string  `json:"pricePrefixEn"`
	PricePrefixCn          string  `json:"pricePrefixCn"`
	PriceSuffixBm          string  `json:"priceSuffixBm"`
	PriceSuffixEn          string  `json:"priceSuffixEn"`
	PriceSuffixCn          string  `json:"priceSuffixCn"`
	AttachmentName         string  `json:"attachmentName"`
	AttachmentPath         string  `json:"attachmentPath"`
	AttachmentSize         int64   `json:"attachmentSize"`
	ContentType            string  `json:"contentType"`
	UniqueExtension        string  `json:"uniqueExtension"`
	IsActive               bool    `json:"isActive"`
	IsTicketInternal       bool    `json:"isTicketInternal"`
	TicketIds              string  `json:"ticketIds"`

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
	OrganiserNameBm            string   `json:"organiserNameBm"`
	OrganiserNameEn            string   `json:"organiserNameEn"`
	OrganiserNameCn            string   `json:"organiserNameCn"`
	OrganiserAddress           string   `json:"organiserAddress"`
	OrganiserDescriptionHtmlBm string   `json:"organiserDescriptionHtmlBm"`
	OrganiserDescriptionHtmlEn string   `json:"organiserDescriptionHtmlEn"`
	OrganiserDescriptionHtmlCn string   `json:"organiserDescriptionHtmlCn"`
	OrganiserContact           *string  `json:"organiserContact"`
	OrganiserEmail             *string  `json:"organiserEmail"`
	OrganiserWebsite           *string  `json:"organiserWebsite"`
	OrganiserOperatingHours    *string  `json:"organiserOperatingHours"`
	OrganiserFacilitiesBm      []string `json:"organiserFacilitiesBm"`
	OrganiserFacilitiesEn      []string `json:"organiserFacilitiesEn"`
	OrganiserFacilitiesCn      []string `json:"organiserFacilitiesCn"`

	TicketVariants []TicketVariantDTO `json:"ticketVariants"`

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

type TicketVariantDTO struct {
	TicketVariantId *uint   `json:"ticketVariantId"`
	TicketGroupId   *uint   `json:"ticketGroupId"`
	TicketId        *string `json:"ticketId"`
	NameBm          string  `json:"nameBm"`
	NameEn          string  `json:"nameEn"`
	NameCn          string  `json:"nameCn"`
	DescBm          string  `json:"descBm"`
	DescEn          string  `json:"descEn"`
	DescCn          string  `json:"descCn"`
	UnitPrice       float64 `json:"unitPrice"`
	PrintType       *string `json:"printType"`
}
