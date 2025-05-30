// File: internal/core/models/banner.go
package models

import (
	"time"
)

// Banner represents the banner table
type Banner struct {
	BannerId       uint      `gorm:"primaryKey;column:banner_id;type:bigint unsigned AUTO_INCREMENT"`
	Placement      int       `gorm:"column:placement;not null"`
	RedirectURL    string    `gorm:"column:redirect_url;type:varchar(255);not null"`
	AttachmentName string    `gorm:"column:attachment_name;type:varchar(255);not null"`
	AttachmentPath string    `gorm:"column:attachment_path;type:varchar(255);not null"`
	AttachmentSize int64     `gorm:"column:attachment_size;type:bigint;not null"`
	ContentType    string    `gorm:"column:content_type;type:varchar(255);not null"`
	UniqueExt      string    `gorm:"column:unique_extension;type:varchar(255);not null"`
	CreatedAt      time.Time `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null"`
}

// TableName overrides the table name
func (Banner) TableName() string {
	return "Banner"
}
