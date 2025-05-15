// File: internal/core/models/banner.go
package models

import (
	"time"
)

// Banner represents the banner table
type Banner struct {
	BannerId       uint      `gorm:"primaryKey;column:banner_id;type:bigint unsigned AUTO_INCREMENT"`
	TicketGroupId  uint      `gorm:"column:ticket_group_id;type:bigint unsigned"`
	Placement      int       `gorm:"column:placement"`
	RedirectURL    string    `gorm:"column:redirect_url;type:varchar(255)"`
	AttachmentName string    `gorm:"column:attachment_name;type:varchar(255)"`
	AttachmentPath string    `gorm:"column:attachment_path;type:varchar(255)"`
	AttachmentSize int64     `gorm:"column:attachment_size;type:bigint"`
	ContentType    string    `gorm:"column:content_type;type:varchar(255)"`
	UniqueExt      string    `gorm:"column:unique_extension;type:varchar(255)"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`

	// Define relationship with constraint:false to prevent auto-generation of constraints
	TicketGroup TicketGroup `gorm:"foreignKey:TicketGroupId;references:TicketGroupId;constraint:false"`
}

// TableName overrides the table name
func (Banner) TableName() string {
	return "Banner"
}
