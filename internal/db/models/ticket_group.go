// File: internal/core/models/ticket_group.go
package models

import (
	"database/sql"
	"time"
)

// TicketGroup represents the ticket_group table
type TicketGroup struct {
	TicketGroupId          uint           `gorm:"primaryKey;column:ticket_group_id;type:bigint unsigned AUTO_INCREMENT"`
	GroupType              string         `gorm:"column:group_type;type:varchar(255);not null"`
	GroupName              string         `gorm:"column:group_name;type:varchar(255);not null"`
	GroupDesc              string         `gorm:"column:group_desc;type:varchar(255);not null"`
	OperatingHours         string         `gorm:"column:operating_hours;type:varchar(255);not null"`
	PricePrefix            string         `gorm:"column:price_prefix;type:varchar(255);not null"`
	PriceSuffix            string         `gorm:"column:price_suffix;type:varchar(255);not null"`
	AttachmentName         string         `gorm:"column:attachment_name;type:varchar(255);not null"`
	AttachmentPath         string         `gorm:"column:attachment_path;type:varchar(255);not null"`
	AttachmentSize         int64          `gorm:"column:attachment_size;type:bigint;not null"`
	ContentType            string         `gorm:"column:content_type;type:varchar(255);not null"`
	UniqueExtension        string         `gorm:"column:unique_extension;type:varchar(255);not null"`
	ActiveEndDate          sql.NullString `gorm:"column:active_end_date;type:char(8);null"`
	ActiveStartDate        sql.NullString `gorm:"column:active_start_date;type:char(8);null"`
	IsActive               bool           `gorm:"column:is_active;type:boolean;not null"`
	IsTicketInternal       string         `gorm:"column:is_ticket_internal;type:varchar(255);not null"`
	LocationAddress        string         `gorm:"column:location_address;type:text;not null"`
	LocationMapUrl         string         `gorm:"column:location_map_url;type:text;not null"`
	OrganiserName          string         `gorm:"column:organiser_name;type:varchar(255);not null"`
	OrganiserAddress       string         `gorm:"column:organiser_address;type:varchar(255);not null"`
	OrganiserDescHtml      string         `gorm:"column:organiser_desc_html;type:text;not null"`
	OrganiserContact       string         `gorm:"column:organiser_contact;type:varchar(255);not null"`
	OrganiserEmail         string         `gorm:"column:organiser_email;type:varchar(255);not null"`
	OrganiserWebsite       string         `gorm:"column:organiser_website;type:varchar(255);not null"`
	OrganiserOperatingHour string         `gorm:"column:organiser_operating_hour;type:varchar(255);not null"`
	OrganiserFacilities    string         `gorm:"column:organiser_facilities;type:text;not null"`
	TicketIds              sql.NullString `gorm:"column:ticket_ids;type:text;null"`
	CreatedAt              time.Time      `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt              time.Time      `gorm:"column:updated_at;type:datetime;not null"`

	// Relationships defined without foreign key constraints
	// These will be used for Go code navigation but won't create DB constraints
	OrderTicketGroups []OrderTicketGroup `gorm:"-"`
	GroupGallery      []GroupGallery     `gorm:"-"`
	TicketTag         []TicketTag        `gorm:"-"`
	TicketGroup       []TicketGroup      `gorm:"-"`
}

// TableName overrides the table name
func (TicketGroup) TableName() string {
	return "Ticket_Group"
}
