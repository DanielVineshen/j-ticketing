// File: j-ticketing/internal/db/models/customer.go
package models

import (
	"database/sql"
	"time"
)

// Customer represents the customer table
type Customer struct {
	CustId           string         `gorm:"primaryKey;column:cust_id;type:varchar(255)"`
	Email            string         `gorm:"column:email;type:varchar(255);uniqueIndex;not null"`
	Password         sql.NullString `gorm:"column:password;type:varchar(255);null"`
	IdentificationNo string         `gorm:"column:identification_no;type:varchar(255);not null"`
	FullName         string         `gorm:"column:full_name;type:varchar(255);not null"`
	ContactNo        sql.NullString `gorm:"column:contact_no;type:varchar(255);null"`
	IsDisabled       bool           `gorm:"column:is_disabled;type:boolean;default:false;not null"`
	CreatedAt        time.Time      `gorm:"column:created_at;not null"`
	UpdatedAt        time.Time      `gorm:"column:updated_at;not null"`

	// Relationships defined without foreign key constraints
	OrderTicketGroups []OrderTicketGroup `gorm:"-"`
}

// TableName overrides the table name
func (Customer) TableName() string {
	return "Customer"
}
