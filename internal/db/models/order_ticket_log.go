// File: j-ticketing/internal/db/models/order_ticket_log.go
package models

import (
	"time"
)

// OrderTicketLog represents the order_ticket_log table
type OrderTicketLog struct {
	OrderTicketLogId   uint      `gorm:"primaryKey;column:order_ticket_log_id;type:bigint unsigned AUTO_INCREMENT"`
	OrderTicketGroupId uint      `gorm:"column:order_ticket_group_id;type:bigint unsigned;not null"`
	Type               string    `gorm:"column:type;type:varchar(255);not null"`
	Title              string    `gorm:"column:title;type:varchar(255);not null"`
	Message            string    `gorm:"column:message;type:text;not null"`
	Date               string    `gorm:"column:date;type:char(14);not null"`
	CreatedAt          time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt          time.Time `gorm:"column:updated_at;type:datetime;not null"`

	// Relationships defined without foreign key constraints
	OrderTicketGroup OrderTicketGroup `gorm:"foreignKey:OrderTicketGroupId;references:OrderTicketGroupId;constraint:false"`
}

// TableName overrides the table name
func (OrderTicketLog) TableName() string {
	return "Order_Ticket_Log"
}
