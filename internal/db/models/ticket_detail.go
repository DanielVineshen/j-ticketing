// File: internal/core/models/ticket_detail.go
package models

import (
	"time"
)

// TicketDetail represents the Ticket_Detail table
type TicketDetail struct {
	TicketDetailId uint      `gorm:"primaryKey;column:ticket_detail_id;type:bigint unsigned AUTO_INCREMENT"`
	TicketGroupId  uint      `gorm:"column:ticket_group_id;type:bigint unsigned;not null"`
	Title          string    `gorm:"column:title;type:varchar(255);not null"`
	TitleIcon      string    `gorm:"column:title_icon;type:varchar(255);not null"`
	RawHtml        string    `gorm:"column:raw_html;type:text;not null"`
	DisplayFlag    bool      `gorm:"column:display_flag;type:boolean;not null"`
	CreatedAt      time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt      time.Time `gorm:"column:updated_at;type:datetime;not null"`

	// Relationship with TicketGroup (without creating DB constraint)
	TicketGroup TicketGroup `gorm:"foreignKey:TicketGroupId;references:TicketGroupId;constraint:false"`
}

// TableName overrides the table name
func (TicketDetail) TableName() string {
	return "Ticket_Detail"
}
