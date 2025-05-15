// File: internal/core/models/order_ticket_group.go
package models

import (
	"time"
)

// OrderTicketGroup represents the order_ticket_group table
type OrderTicketGroup struct {
	OrderTicketGroupId uint      `gorm:"primaryKey;column:order_ticket_group_id;type:bigint unsigned AUTO_INCREMENT"`
	TicketGroupId      uint      `gorm:"column:ticket_group_id;type:bigint unsigned"`
	CustId             string    `gorm:"column:cust_id;type:varchar(255)"`
	Payment            string    `gorm:"column:payment"`
	CreatedAt          time.Time `gorm:"column:created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at"`

	// Relationships with constraint:false to prevent auto-generation of constraints
	TicketGroup      TicketGroup       `gorm:"foreignKey:TicketGroupId;references:TicketGroupId;constraint:false"`
	Customer         Customer          `gorm:"foreignKey:CustId;references:CustId;constraint:false"`
	OrderTicketInfos []OrderTicketInfo `gorm:"-"`
}

// TableName overrides the table name
func (OrderTicketGroup) TableName() string {
	return "Order_Ticket_Group"
}
