// File: internal/core/models/ticket_tag.go
package models

import (
	"time"
)

// TicketTag represents the junction table between TicketGroup and Tag
type TicketTag struct {
	TicketGroupId uint      `gorm:"primaryKey;column:ticket_group_id;type:bigint unsigned;not null"`
	TagId         uint      `gorm:"primaryKey;column:tag_id;type:bigint unsigned;not null"`
	CreatedAt     time.Time `gorm:"column:created_at;not null"`
	UpdatedAt     time.Time `gorm:"column:updated_at;not null"`

	// Define relationships with constraint:false
	TicketGroup TicketGroup `gorm:"foreignKey:TicketGroupId;references:TicketGroupId;constraint:false"`
	Tag         Tag         `gorm:"foreignKey:TagId;references:TagId;constraint:false"`
}

// TableName overrides the table name
func (TicketTag) TableName() string {
	return "Ticket_Tag"
}
