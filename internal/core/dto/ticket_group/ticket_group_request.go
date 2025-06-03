package dto

import "mime/multipart"

type CreateTicketGroupRequest struct {
	// Basic settings
	OrderTicketLimit int    `json:"orderTicketLimit" validate:"required,min=1"`
	ScanSetting      string `json:"scanSetting" validate:"required,oneof=none qr_code barcode"`

	// Group names (multilingual)
	GroupNameBm string `json:"groupNameBm" validate:"required,min=1,max=255"`
	GroupNameEn string `json:"groupNameEn" validate:"required,min=1,max=255"`
	GroupNameCn string `json:"groupNameCn" validate:"required,min=1,max=255"`

	// Group descriptions (multilingual)
	GroupDescBm string `json:"groupDescBm" validate:"required,min=1"`
	GroupDescEn string `json:"groupDescEn" validate:"required,min=1"`
	GroupDescCn string `json:"groupDescCn" validate:"required,min=1"`

	// Redirection settings (optional)
	GroupRedirectionSpanBm string `json:"groupRedirectionSpanBm" validate:"omitempty,max=255"`
	GroupRedirectionSpanEn string `json:"groupRedirectionSpanEn" validate:"omitempty,max=255"`
	GroupRedirectionSpanCn string `json:"groupRedirectionSpanCn" validate:"omitempty,max=255"`
	GroupRedirectionUrl    string `json:"groupRedirectionUrl" validate:"omitempty,url,max=500"`

	// Group slots (optional, multilingual)
	GroupSlot1Bm string `json:"groupSlot1Bm" validate:"omitempty,max=255"`
	GroupSlot1En string `json:"groupSlot1En" validate:"omitempty,max=255"`
	GroupSlot1Cn string `json:"groupSlot1Cn" validate:"omitempty,max=255"`
	GroupSlot2Bm string `json:"groupSlot2Bm" validate:"omitempty,max=255"`
	GroupSlot2En string `json:"groupSlot2En" validate:"omitempty,max=255"`
	GroupSlot2Cn string `json:"groupSlot2Cn" validate:"omitempty,max=255"`
	GroupSlot3Bm string `json:"groupSlot3Bm" validate:"omitempty,max=255"`
	GroupSlot3En string `json:"groupSlot3En" validate:"omitempty,max=255"`
	GroupSlot3Cn string `json:"groupSlot3Cn" validate:"omitempty,max=255"`
	GroupSlot4Bm string `json:"groupSlot4Bm" validate:"omitempty,max=255"`
	GroupSlot4En string `json:"groupSlot4En" validate:"omitempty,max=255"`
	GroupSlot4Cn string `json:"groupSlot4Cn" validate:"omitempty,max=255"`

	// Price prefixes and suffixes (multilingual)
	PricePrefixBm string `json:"pricePrefixBm" validate:"required,min=1,max=50"`
	PricePrefixEn string `json:"pricePrefixEn" validate:"required,min=1,max=50"`
	PricePrefixCn string `json:"pricePrefixCn" validate:"required,min=1,max=50"`
	PriceSuffixBm string `json:"priceSuffixBm" validate:"required,min=1,max=50"`
	PriceSuffixEn string `json:"priceSuffixEn" validate:"required,min=1,max=50"`
	PriceSuffixCn string `json:"priceSuffixCn" validate:"required,min=1,max=50"`

	// Date settings
	ActiveStartDate string `json:"activeStartDate" validate:"required"`
	ActiveEndDate   string `json:"activeEndDate" validate:"omitempty"`
	IsActive        bool   `json:"isActive"`

	// Location information
	LocationAddress string `json:"locationAddress" validate:"required,min=1"`
	LocationMapUrl  string `json:"locationMapUrl" validate:"required,url"`

	// Organiser information (multilingual)
	OrganiserNameBm       string `json:"organiserNameBm" validate:"required,min=1,max=255"`
	OrganiserNameEn       string `json:"organiserNameEn" validate:"required,min=1,max=255"`
	OrganiserNameCn       string `json:"organiserNameCn" validate:"required,min=1,max=255"`
	OrganiserAddress      string `json:"organiserAddress" validate:"required,min=1"`
	OrganiserDescHtmlBm   string `json:"organiserDescHtmlBm" validate:"required,min=1"`
	OrganiserDescHtmlEn   string `json:"organiserDescHtmlEn" validate:"required,min=1"`
	OrganiserDescHtmlCn   string `json:"organiserDescHtmlCn" validate:"required,min=1"`
	OrganiserContact      string `json:"organiserContact" validate:"omitempty,max=50"`
	OrganiserEmail        string `json:"organiserEmail" validate:"omitempty,email,max=255"`
	OrganiserWebsite      string `json:"organiserWebsite" validate:"omitempty,url,max=500"`
	OrganiserFacilitiesBm string `json:"organiserFacilitiesBm" validate:"omitempty"`
	OrganiserFacilitiesEn string `json:"organiserFacilitiesEn" validate:"omitempty"`
	OrganiserFacilitiesCn string `json:"organiserFacilitiesCn" validate:"omitempty"`

	// Complex data (arrays)
	TicketDetails  []TicketDetailsRequest  `json:"ticketDetails" validate:"required,min=1,dive"`
	TicketVariants []TicketVariantsRequest `json:"ticketVariants" validate:"required,min=1,dive"`
	TicketTags     []TicketTagsRequest     `json:"ticketTags" validate:"omitempty,dive"`

	// Files (handled separately)
	Attachment     *multipart.FileHeader   `json:"-"`
	GroupGalleries []*multipart.FileHeader `json:"-"`
}

// TicketDetailRequest represents a ticket detail item
type TicketDetailsRequest struct {
	TitleBm     string `json:"titleBm" validate:"required,min=1,max=255"`
	TitleEn     string `json:"titleEn" validate:"required,min=1,max=255"`
	TitleCn     string `json:"titleCn" validate:"required,min=1,max=255"`
	TitleIcon   string `json:"titleIcon" validate:"required,max=255"`
	RawHtmlBm   string `json:"rawHtmlBm" validate:"required,min=1"`
	RawHtmlEn   string `json:"rawHtmlEn" validate:"required,min=1"`
	RawHtmlCn   string `json:"rawHtmlCn" validate:"required,min=1"`
	DisplayFlag bool   `json:"displayFlag"`
}

// TicketVariantRequest represents a ticket variant
type TicketVariantsRequest struct {
	NameBm    string  `json:"nameBm" validate:"required,min=1,max=255"`
	NameEn    string  `json:"nameEn" validate:"required,min=1,max=255"`
	NameCn    string  `json:"nameCn" validate:"required,min=1,max=255"`
	DescBm    string  `json:"descBm" validate:"required,min=1,max=255"`
	DescEn    string  `json:"descEn" validate:"required,min=1,max=255"`
	DescCn    string  `json:"descCn" validate:"required,min=1,max=255"`
	UnitPrice float64 `json:"unitPrice" validate:"gte=0"`
}

// TicketTagsRequest represents a ticket tag association
type TicketTagsRequest struct {
	TagId uint `json:"tagId" validate:"required,min=1"`
}

type UpdateTicketGroupImageRequest struct {
	TicketGroupId uint                  `json:"ticketGroupId" validate:"required,min=1"`
	Attachment    *multipart.FileHeader `json:"-"`
}

type UpdateTicketGroupBasicInfoRequest struct {
	TicketGroupId uint `json:"ticketGroupId" validate:"required,min=1"`

	// Basic settings
	OrderTicketLimit int    `json:"orderTicketLimit" validate:"required,min=1"`
	ScanSetting      string `json:"scanSetting" validate:"required,oneof=none qr_code barcode"`

	// Group names (multilingual)
	GroupNameBm string `json:"groupNameBm" validate:"required,min=1,max=255"`
	GroupNameEn string `json:"groupNameEn" validate:"required,min=1,max=255"`
	GroupNameCn string `json:"groupNameCn" validate:"required,min=1,max=255"`

	// Group descriptions (multilingual)
	GroupDescBm string `json:"groupDescBm" validate:"required,min=1"`
	GroupDescEn string `json:"groupDescEn" validate:"required,min=1"`
	GroupDescCn string `json:"groupDescCn" validate:"required,min=1"`

	// Redirection settings (optional)
	GroupRedirectionSpanBm string `json:"groupRedirectionSpanBm" validate:"omitempty,max=255"`
	GroupRedirectionSpanEn string `json:"groupRedirectionSpanEn" validate:"omitempty,max=255"`
	GroupRedirectionSpanCn string `json:"groupRedirectionSpanCn" validate:"omitempty,max=255"`
	GroupRedirectionUrl    string `json:"groupRedirectionUrl" validate:"omitempty,url,max=500"`

	// Group slots (optional, multilingual)
	GroupSlot1Bm string `json:"groupSlot1Bm" validate:"omitempty,max=255"`
	GroupSlot1En string `json:"groupSlot1En" validate:"omitempty,max=255"`
	GroupSlot1Cn string `json:"groupSlot1Cn" validate:"omitempty,max=255"`
	GroupSlot2Bm string `json:"groupSlot2Bm" validate:"omitempty,max=255"`
	GroupSlot2En string `json:"groupSlot2En" validate:"omitempty,max=255"`
	GroupSlot2Cn string `json:"groupSlot2Cn" validate:"omitempty,max=255"`
	GroupSlot3Bm string `json:"groupSlot3Bm" validate:"omitempty,max=255"`
	GroupSlot3En string `json:"groupSlot3En" validate:"omitempty,max=255"`
	GroupSlot3Cn string `json:"groupSlot3Cn" validate:"omitempty,max=255"`
	GroupSlot4Bm string `json:"groupSlot4Bm" validate:"omitempty,max=255"`
	GroupSlot4En string `json:"groupSlot4En" validate:"omitempty,max=255"`
	GroupSlot4Cn string `json:"groupSlot4Cn" validate:"omitempty,max=255"`

	// Price prefixes and suffixes (multilingual)
	PricePrefixBm string `json:"pricePrefixBm" validate:"required,min=1,max=50"`
	PricePrefixEn string `json:"pricePrefixEn" validate:"required,min=1,max=50"`
	PricePrefixCn string `json:"pricePrefixCn" validate:"required,min=1,max=50"`
	PriceSuffixBm string `json:"priceSuffixBm" validate:"required,min=1,max=50"`
	PriceSuffixEn string `json:"priceSuffixEn" validate:"required,min=1,max=50"`
	PriceSuffixCn string `json:"priceSuffixCn" validate:"required,min=1,max=50"`

	// Date settings
	ActiveStartDate string `json:"activeStartDate" validate:"required"`
	ActiveEndDate   string `json:"activeEndDate" validate:"omitempty"`
	IsActive        bool   `json:"isActive"`

	// Location information
	LocationAddress string `json:"locationAddress" validate:"required,min=1"`
	LocationMapUrl  string `json:"locationMapUrl" validate:"required,url"`

	TicketTags []TicketTagsRequest `json:"ticketTags" validate:"omitempty,dive"`
}

type DeleteTicketGroupGalleryRequest struct {
	GroupGalleryId uint `json:"groupGalleryId" validate:"required,min=1"`
}

type UpdateTicketGroupDetailsRequest struct {
	TicketGroupId uint                  `json:"ticketGroupId" validate:"required,min=1"`
	TicketDetails []UpdateTicketDetails `json:"ticketDetails" validate:"required,min=1,dive"`
}

type UpdateTicketDetails struct {
	TicketDetailId uint   `json:"ticketDetailId" validate:"required,min=1"`
	TitleBm        string `json:"titleBm" validate:"required,min=1,max=255"`
	TitleEn        string `json:"titleEn" validate:"required,min=1,max=255"`
	TitleCn        string `json:"titleCn" validate:"required,min=1,max=255"`
	TitleIcon      string `json:"titleIcon" validate:"required,max=255"`
	RawHtmlBm      string `json:"rawHtmlBm" validate:"required,min=1"`
	RawHtmlEn      string `json:"rawHtmlEn" validate:"required,min=1"`
	RawHtmlCn      string `json:"rawHtmlCn" validate:"required,min=1"`
	DisplayFlag    bool   `json:"displayFlag"`
}

type UpdateTicketGroupVariantsRequest struct {
	TicketGroupId  uint                   `json:"ticketGroupId" validate:"required,min=1"`
	TicketVariants []UpdateTicketVariants `json:"ticketVariants" validate:"required,min=1,dive"`
}

type UpdateTicketVariants struct {
	TicketVariantId *uint   `json:"ticketVariantId" validate:"omitempty"`
	NameBm          string  `json:"nameBm" validate:"required,min=1,max=255"`
	NameEn          string  `json:"nameEn" validate:"required,min=1,max=255"`
	NameCn          string  `json:"nameCn" validate:"required,min=1,max=255"`
	DescBm          string  `json:"descBm" validate:"required,min=1,max=255"`
	DescEn          string  `json:"descEn" validate:"required,min=1,max=255"`
	DescCn          string  `json:"descCn" validate:"required,min=1,max=255"`
	UnitPrice       float64 `json:"unitPrice" validate:"gte=0"`
}

type UpdateTicketGroupOrganiserInfoRequest struct {
	TicketGroupId         uint   `json:"ticketGroupId" validate:"required,min=1"`
	OrganiserNameBm       string `json:"organiserNameBm" validate:"required,min=1,max=255"`
	OrganiserNameEn       string `json:"organiserNameEn" validate:"required,min=1,max=255"`
	OrganiserNameCn       string `json:"organiserNameCn" validate:"required,min=1,max=255"`
	OrganiserAddress      string `json:"organiserAddress" validate:"required,min=1"`
	OrganiserDescHtmlBm   string `json:"organiserDescHtmlBm" validate:"required,min=1"`
	OrganiserDescHtmlEn   string `json:"organiserDescHtmlEn" validate:"required,min=1"`
	OrganiserDescHtmlCn   string `json:"organiserDescHtmlCn" validate:"required,min=1"`
	OrganiserContact      string `json:"organiserContact" validate:"omitempty,max=50"`
	OrganiserEmail        string `json:"organiserEmail" validate:"omitempty,email,max=255"`
	OrganiserWebsite      string `json:"organiserWebsite" validate:"omitempty,url,max=500"`
	OrganiserFacilitiesBm string `json:"organiserFacilitiesBm" validate:"omitempty"`
	OrganiserFacilitiesEn string `json:"organiserFacilitiesEn" validate:"omitempty"`
	OrganiserFacilitiesCn string `json:"organiserFacilitiesCn" validate:"omitempty"`
}
