// File: j-ticketing/internal/db/models/tag.go
package models

import (
	"time"
)

// Tag represents the tag table
type Tag struct {
	TagId     uint      `gorm:"primaryKey;column:tag_id;type:bigint unsigned AUTO_INCREMENT"`
	TagName   string    `gorm:"column:tag_name;type:varchar(255);uniqueIndex;not null"`
	TagDesc   string    `gorm:"column:tag_desc;type:text;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`

	// Relationships
	TicketTags []TicketTag `gorm:"foreignKey:TagId"`
}

// TableName overrides the table name
func (Tag) TableName() string {
	return "Tag"
}
