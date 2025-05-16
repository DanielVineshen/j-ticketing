// File: j-ticketing/internal/db/models/group_gallery.go
package models

import (
	"time"
)

// GroupGallery represents the Group_Gallery table
type GroupGallery struct {
	GroupGalleryId  uint      `gorm:"primaryKey;column:group_gallery_id;type:bigint unsigned AUTO_INCREMENT"`
	TicketGroupId   uint      `gorm:"column:ticket_group_id;type:bigint unsigned;not null"`
	AttachmentName  string    `gorm:"column:attachment_name;type:varchar(255);not null"`
	AttachmentPath  string    `gorm:"column:attachment_path;type:varchar(255);not null"`
	AttachmentSize  int64     `gorm:"column:attachment_size;type:bigint;not null"`
	ContentType     string    `gorm:"column:content_type;type:varchar(255);not null"`
	UniqueExtension string    `gorm:"column:unique_extension;type:varchar(255);not null"`
	CreatedAt       time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt       time.Time `gorm:"column:updated_at;type:datetime;not null"`

	// Relationship with TicketGroup (without creating DB constraint)
	TicketGroup TicketGroup `gorm:"foreignKey:TicketGroupId;references:TicketGroupId;constraint:false"`
}

// TableName overrides the table name
func (GroupGallery) TableName() string {
	return "Group_Gallery"
}
