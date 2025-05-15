// File: internal/core/models/order_ticket_info.go
package models

import (
	"time"
)

// OrderTicketInfo represents the order_ticket_info table
type OrderTicketInfo struct {
	OrderTicketInfoId  uint      `gorm:"primaryKey;column:order_ticket_info_id;type:bigint unsigned AUTO_INCREMENT"`
	OrderTicketGroupId uint      `gorm:"column:order_ticket_group_id;type:bigint unsigned"`
	ItemId             string    `gorm:"column:item_id;type:varchar(255)"`
	UnitPrice          float64   `gorm:"column:unit_price;type:decimal(10,2)"`
	ItemDesc1          string    `gorm:"column:item_desc_1;type:varchar(255)"`
	ItemDesc2          string    `gorm:"column:item_desc_2;type:varchar(255)"`
	PrintType          string    `gorm:"column:print_type;type:varchar(255)"`
	QuantityBought     int       `gorm:"column:quantity_bought"`
	Twbid              string    `gorm:"column:twbid;type:varchar(255)"`
	EncryptedId        string    `gorm:"column:encrypted_id;type:varchar(255)"`
	AdmitDate          string    `gorm:"column:admit_date;type:varchar(255)"`
	Variant            string    `gorm:"column:variant;type:varchar(255)"`
	CreatedAt          time.Time `gorm:"column:created_at"`
	UpdatedAt          time.Time `gorm:"column:updated_at"`

	// Relationship with constraint:false
	OrderTicketGroup OrderTicketGroup `gorm:"foreignKey:OrderTicketGroupId;references:OrderTicketGroupId;constraint:false"`
}

// TableName overrides the table name
func (OrderTicketInfo) TableName() string {
	return "Order_Ticket_Info"
}
