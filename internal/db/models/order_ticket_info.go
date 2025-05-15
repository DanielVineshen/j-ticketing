// File: internal/core/models/order_ticket_info.go
package models

import (
	"database/sql"
	"time"
)

// OrderTicketInfo represents the order_ticket_info table
type OrderTicketInfo struct {
	OrderTicketInfoId  uint           `gorm:"primaryKey;column:order_ticket_info_id;type:bigint unsigned AUTO_INCREMENT"`
	OrderTicketGroupId uint           `gorm:"column:order_ticket_group_id;type:bigint unsigned;not null"`
	ItemId             string         `gorm:"column:item_id;type:varchar(255);not null"`
	UnitPrice          float64        `gorm:"column:unit_price;type:decimal(10,2);not null"`
	ItemDesc1          string         `gorm:"column:item_desc_1;type:varchar(255);not null"`
	ItemDesc2          string         `gorm:"column:item_desc_2;type:varchar(255);not null"`
	PrintType          string         `gorm:"column:print_type;type:varchar(255);not null"`
	QuantityBought     int            `gorm:"column:quantity_bought;not null"`
	Twbid              sql.NullString `gorm:"column:twbid;type:varchar(255);null"`
	EncryptedId        string         `gorm:"column:encrypted_id;type:varchar(255);not null"`
	AdmitDate          string         `gorm:"column:admit_date;type:varchar(255);not null"`
	Variant            string         `gorm:"column:variant;type:varchar(255);not null"`
	CreatedAt          time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt          time.Time      `gorm:"column:updated_at;not null"`

	// Relationship with constraint:false
	OrderTicketGroup OrderTicketGroup `gorm:"foreignKey:OrderTicketGroupId;references:OrderTicketGroupId;constraint:false"`
}

// TableName overrides the table name
func (OrderTicketInfo) TableName() string {
	return "Order_Ticket_Info"
}
