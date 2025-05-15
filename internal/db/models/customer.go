// File: internal/core/models/customer.go
package models

import (
	"time"
)

// Customer represents the customer table
type Customer struct {
	CustId           string    `gorm:"primaryKey;column:cust_id;type:varchar(255)"`
	Email            string    `gorm:"column:email;type:varchar(255);uniqueIndex"`
	Password         string    `gorm:"column:password;type:varchar(255)"`
	IdentificationNo string    `gorm:"column:identification_no;type:varchar(255)"`
	FullName         string    `gorm:"column:full_name;type:varchar(255)"`
	ContactNo        string    `gorm:"column:contact_no;type:varchar(255)"`
	IsDisabled       bool      `gorm:"column:is_disabled;type:boolean;default:false"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`

	// Relationships defined without foreign key constraints
	OrderTicketGroups []OrderTicketGroup `gorm:"-"`
}

// TableName overrides the table name
func (Customer) TableName() string {
	return "Customer"
}
