// File: internal/core/models/ticket_group.go
package models

import (
	"time"
)

// TicketGroup represents the ticket_group table
type TicketGroup struct {
	TicketGroupId   uint      `gorm:"primaryKey;column:ticket_group_id;type:bigint unsigned AUTO_INCREMENT"`
	GroupName       string    `gorm:"column:group_name;type:varchar(255)"`
	GroupDesc       string    `gorm:"column:group_desc;type:varchar(255)"`
	PriceDesc       string    `gorm:"column:price_desc;type:varchar(255)"`
	OperatingDesc   string    `gorm:"column:operatin_desc;type:varchar(255)"`
	ActiveStartDate time.Time `gorm:"column:active_start_date"`
	ActiveEndDate   string    `gorm:"column:active_end_date;type:char(8)"`
	GroupType       string    `gorm:"column:group_type;type:varchar(255)"`
	AttachmentName  string    `gorm:"column:attachment_name;type:varchar(255)"`
	AttachmentPath  string    `gorm:"column:attachment_path;type:varchar(255)"`
	AttachmentSize  int64     `gorm:"column:attachment_size;type:bigint"`
	ContentType     string    `gorm:"column:content_type;type:varchar(255)"`
	UniqueExtension string    `gorm:"column:unique_extension;type:varchar(255)"`
	IsActive        bool      `gorm:"column:is_active;type:boolean;default:true"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime"`

	// Relationships defined without foreign key constraints
	// These will be used for Go code navigation but won't create DB constraints
	Banners           []Banner           `gorm:"-"`
	OrderTicketGroups []OrderTicketGroup `gorm:"-"`
}

// TableName overrides the table name
func (TicketGroup) TableName() string {
	return "Ticket_Group"
}
