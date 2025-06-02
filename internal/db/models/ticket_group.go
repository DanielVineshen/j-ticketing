// File: j-ticketing/internal/db/models/ticket_group.go
package models

import (
	"database/sql"
	"time"
)

// TicketGroup represents the ticket_group table
type TicketGroup struct {
	TicketGroupId    uint   `gorm:"primaryKey;column:ticket_group_id;type:bigint unsigned AUTO_INCREMENT"`
	OrderTicketLimit int    `gorm:"column:order_ticket_limit;type:int;not null;default:1"`
	ScanSetting      string `gorm:"column:scan_setting;type:varchar(255);not null;default:'none'"`
	GroupType        string `gorm:"column:group_type;type:varchar(255);not null"`

	// Multilingual Group Names
	GroupNameBm string `gorm:"column:group_name_bm;type:varchar(255);not null"`
	GroupNameEn string `gorm:"column:group_name_en;type:varchar(255);not null"`
	GroupNameCn string `gorm:"column:group_name_cn;type:varchar(255);not null"`

	// Multilingual Group Descriptions
	GroupDescBm string `gorm:"column:group_desc_bm;type:varchar(255);not null"`
	GroupDescEn string `gorm:"column:group_desc_en;type:varchar(255);not null"`
	GroupDescCn string `gorm:"column:group_desc_cn;type:varchar(255);not null"`

	// Redirection Fields
	GroupRedirectionSpanBm sql.NullString `gorm:"column:group_redirection_span_bm;type:varchar(255);null"`
	GroupRedirectionSpanEn sql.NullString `gorm:"column:group_redirection_span_en;type:varchar(255);null"`
	GroupRedirectionSpanCn sql.NullString `gorm:"column:group_redirection_span_cn;type:varchar(255);null"`
	GroupRedirectionUrl    sql.NullString `gorm:"column:group_redirection_url;type:varchar(255);null"`

	// Slot Fields (Nullable)
	GroupSlot1Bm sql.NullString `gorm:"column:group_slot_1_bm;type:varchar(255);null"`
	GroupSlot1En sql.NullString `gorm:"column:group_slot_1_en;type:varchar(255);null"`
	GroupSlot1Cn sql.NullString `gorm:"column:group_slot_1_cn;type:varchar(255);null"`
	GroupSlot2Bm sql.NullString `gorm:"column:group_slot_2_bm;type:varchar(255);null"`
	GroupSlot2En sql.NullString `gorm:"column:group_slot_2_en;type:varchar(255);null"`
	GroupSlot2Cn sql.NullString `gorm:"column:group_slot_2_cn;type:varchar(255);null"`
	GroupSlot3Bm sql.NullString `gorm:"column:group_slot_3_bm;type:varchar(255);null"`
	GroupSlot3En sql.NullString `gorm:"column:group_slot_3_en;type:varchar(255);null"`
	GroupSlot3Cn sql.NullString `gorm:"column:group_slot_3_cn;type:varchar(255);null"`
	GroupSlot4Bm sql.NullString `gorm:"column:group_slot_4_bm;type:varchar(255);null"`
	GroupSlot4En sql.NullString `gorm:"column:group_slot_4_en;type:varchar(255);null"`
	GroupSlot4Cn sql.NullString `gorm:"column:group_slot_4_cn;type:varchar(255);null"`

	// Multilingual Price Fields
	PricePrefixBm string `gorm:"column:price_prefix_bm;type:varchar(255);not null"`
	PricePrefixEn string `gorm:"column:price_prefix_en;type:varchar(255);not null"`
	PricePrefixCn string `gorm:"column:price_prefix_cn;type:varchar(255);not null"`
	PriceSuffixBm string `gorm:"column:price_suffix_bm;type:varchar(255);not null"`
	PriceSuffixEn string `gorm:"column:price_suffix_en;type:varchar(255);not null"`
	PriceSuffixCn string `gorm:"column:price_suffix_cn;type:varchar(255);not null"`

	// Attachment Fields
	AttachmentName  string `gorm:"column:attachment_name;type:varchar(255);not null"`
	AttachmentPath  string `gorm:"column:attachment_path;type:varchar(255);not null"`
	AttachmentSize  int64  `gorm:"column:attachment_size;type:bigint;not null"`
	ContentType     string `gorm:"column:content_type;type:varchar(255);not null"`
	UniqueExtension string `gorm:"column:unique_extension;type:varchar(255);not null"`

	// Date Fields
	ActiveEndDate   sql.NullString `gorm:"column:active_end_date;type:varchar(255);null"`
	ActiveStartDate sql.NullString `gorm:"column:active_start_date;type:varchar(255);null"`

	// Status Fields
	IsActive         bool `gorm:"column:is_active;type:boolean;not null"`
	IsTicketInternal bool `gorm:"column:is_ticket_internal;type:boolean;not null"`

	// Location Fields
	LocationAddress string `gorm:"column:location_address;type:text;not null"`
	LocationMapUrl  string `gorm:"column:location_map_url;type:text;not null"`

	// Multilingual Organiser Fields
	OrganiserNameBm       string         `gorm:"column:organiser_name_bm;type:varchar(255);not null"`
	OrganiserNameEn       string         `gorm:"column:organiser_name_en;type:varchar(255);not null"`
	OrganiserNameCn       string         `gorm:"column:organiser_name_cn;type:varchar(255);not null"`
	OrganiserAddress      string         `gorm:"column:organiser_address;type:varchar(255);not null"`
	OrganiserDescHtmlBm   string         `gorm:"column:organiser_desc_html_bm;type:text;not null"`
	OrganiserDescHtmlEn   string         `gorm:"column:organiser_desc_html_en;type:text;not null"`
	OrganiserDescHtmlCn   string         `gorm:"column:organiser_desc_html_cn;type:text;not null"`
	OrganiserContact      sql.NullString `gorm:"column:organiser_contact;type:varchar(255);null"`
	OrganiserEmail        sql.NullString `gorm:"column:organiser_email;type:varchar(255);null"`
	OrganiserWebsite      sql.NullString `gorm:"column:organiser_website;type:varchar(255);null"`
	OrganiserFacilitiesBm sql.NullString `gorm:"column:organiser_facilities_bm;type:text;null"`
	OrganiserFacilitiesEn sql.NullString `gorm:"column:organiser_facilities_en;type:text;null"`
	OrganiserFacilitiesCn sql.NullString `gorm:"column:organiser_facilities_cn;type:text;null"`

	// Other Fields
	TicketIds sql.NullString `gorm:"column:ticket_ids;type:text;null"`
	CreatedAt time.Time      `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt time.Time      `gorm:"column:updated_at;type:datetime;not null"`

	// Relationships defined without foreign key constraints
	// These will be used for Go code navigation but won't create DB constraints
	OrderTicketGroups []OrderTicketGroup `gorm:"foreignKey:TicketGroupId"`
	GroupGalleries    []GroupGallery     `gorm:"foreignKey:TicketGroupId"`
	TicketDetails     []TicketDetail     `gorm:"foreignKey:TicketGroupId"`
	TicketTags        []TicketTag        `gorm:"foreignKey:TicketGroupId"`
	TicketVariants    []TicketVariant    `gorm:"foreignKey:TicketGroupId"`
}

// TableName overrides the table name
func (TicketGroup) TableName() string {
	return "Ticket_Group"
}
