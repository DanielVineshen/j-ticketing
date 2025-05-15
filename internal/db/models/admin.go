// File: internal/core/models/admin.go
package models

import (
	"time"
)

// Admin represents the admin table
type Admin struct {
	AdminId   uint      `gorm:"primaryKey;column:admin_id;type:bigint unsigned AUTO_INCREMENT"`
	Username  string    `gorm:"column:username;type:varchar(255);uniqueIndex;not null"`
	Password  string    `gorm:"column:password;type:varchar(255);not null"`
	FullName  string    `gorm:"column:full_name;type:varchar(255);not null"`
	Role      string    `gorm:"column:role;type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

// TableName overrides the table name
func (Admin) TableName() string {
	return "Admin"
}
