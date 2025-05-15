// File: internal/core/models/group_gallery.go
package models

import (
	"time"
)

// GroupGallery represents the group_gallery table
type GroupGallery struct {
	GroupGalleryId  uint      `gorm:"primaryKey;column:group_gallery_id;type:bigint unsigned AUTO_INCREMENT"`
	TicketGroupId   uint      `gorm:"column:ticket_group_id;type:bigint unsigned"`
	AttachmentName  string    `gorm:"column:attachment_name;type:varchar(255)"`
	AttachmentPath  string    `gorm:"column:attachment_path;type:varchar(255)"`
	AttachmentSize  int64     `gorm:"column:attachment_size;type:bigint"`
	ContentType     string    `gorm:"column:content_type;type:varchar(255)"`
	UniqueExtension string    `gorm:"column:unique_extension;type:varchar(255)"`
	CreatedAt       time.Time `gorm:"column:created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at"`

	// Relationship with constraint:false
	TicketGroup TicketGroup `gorm:"foreignKey:TicketGroupId;references:TicketGroupId;constraint:false"`
}

// TableName overrides the table name
func (GroupGallery) TableName() string {
	return "Group_Gallery"
}
