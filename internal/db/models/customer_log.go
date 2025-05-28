// File: j-ticketing/internal/db/models/customer_log.go
package models

import (
	"time"
)

// CustomerLog represents the customer_log table
type CustomerLog struct {
	CustLogId uint      `gorm:"primaryKey;column:cust_log_id;type:bigint unsigned AUTO_INCREMENT"`
	CustId    string    `gorm:"column:cust_id;type:varchar(255);not null"`
	Type      string    `gorm:"column:type;type:varchar(255);not null"`
	Title     string    `gorm:"column:title;type:varchar(255);not null"`
	Message   string    `gorm:"column:message;type:text;not null"`
	Date      string    `gorm:"column:date;type:char(14);not null"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null"`

	// Relationships defined without foreign key constraints
	Customer Customer `gorm:"foreignKey:CustId;references:CustId;constraint:false"`
}

// TableName overrides the table name
func (CustomerLog) TableName() string {
	return "Customer_Log"
}
