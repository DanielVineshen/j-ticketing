// File: j-ticketing/internal/db/models/ticket_variant.go
package models

import (
	"time"
)

// TicketVariant represents the ticket_variant table
type TicketVariant struct {
	TicketVariantId uint      `gorm:"primaryKey;column:ticket_variant_id;type:bigint unsigned AUTO_INCREMENT"`
	TicketGroupId   uint      `gorm:"column:ticket_group_id;type:bigint unsigned;not null;index"`
	TicketId        string    `gorm:"column:ticket_id;type:varchar(255);not null"`
	NameBm          string    `gorm:"column:name_bm;type:varchar(255);not null"`
	NameEn          string    `gorm:"column:name_en;type:varchar(255);not null"`
	NameCn          string    `gorm:"column:name_cn;type:varchar(255);not null"`
	DescBm          string    `gorm:"column:desc_bm;type:text;not null"`
	DescEn          string    `gorm:"column:desc_en;type:text;not null"`
	DescCn          string    `gorm:"column:desc_cn;type:text;not null"`
	UnitPrice       float64   `gorm:"column:unit_price;type:decimal(10,2);not null"`
	CreatedAt       time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt       time.Time `gorm:"column:updated_at;type:datetime;not null"`

	// Relationships defined without foreign key constraints
	// Reference to the ticket group this variant belongs to
	TicketGroup TicketGroup `gorm:"foreignKey:TicketGroupId;references:TicketGroupId"`
}

// TableName overrides the table name
func (TicketVariant) TableName() string {
	return "Ticket_Variant"
}
