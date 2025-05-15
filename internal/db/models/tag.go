// File: internal/core/models/tag.go
package models

import (
	"time"
)

// Tag represents the tag table
type Tag struct {
	TagId     uint      `gorm:"primaryKey;column:tag_id;type:bigint unsigned AUTO_INCREMENT"`
	TagName   string    `gorm:"column:tag_name;type:varchar(255);uniqueIndex"`
	TagDesc   string    `gorm:"column:tag_desc;type:text"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`

	// Relationships
	TicketGroups []TicketGroup `gorm:"-"`
}

// TableName overrides the table name
func (Tag) TableName() string {
	return "Tag"
}
