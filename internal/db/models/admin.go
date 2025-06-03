// File: j-ticketing/internal/db/models/admin.go
package models

import (
	"time"
)

// Admin represents the admin table
type Admin struct {
	AdminId    uint      `gorm:"primaryKey;column:admin_id;type:bigint unsigned AUTO_INCREMENT"`
	Username   string    `gorm:"column:username;type:varchar(255);uniqueIndex;not null"`
	Password   string    `gorm:"column:password;type:varchar(255);not null"`
	FullName   string    `gorm:"column:full_name;type:varchar(255);not null"`
	Role       string    `gorm:"column:role;type:varchar(255);not null"`
	Email      string    `gorm:"column:email;type:varchar(255);not null"`
	ContactNo  string    `gorm:"column:contact_no;type:varchar(255);not null"`
	IsDisabled bool      `gorm:"column:is_disabled;type:boolean;default:false;not null"`
	CreatedAt  time.Time `gorm:"column:created_at;not null"`
	UpdatedAt  time.Time `gorm:"column:updated_at;not null"`
}

// TableName overrides the table name
func (Admin) TableName() string {
	return "Admin"
}
