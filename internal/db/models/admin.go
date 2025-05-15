// File: internal/core/models/admin.go
package models

import (
	"time"
)

// Admin represents the admin table
type Admin struct {
	AdminId   uint      `gorm:"primaryKey;column:admin_id;type:bigint unsigned AUTO_INCREMENT"`
	Username  string    `gorm:"column:username;type:varchar(255);uniqueIndex"`
	Password  string    `gorm:"column:password;type:varchar(255)"`
	FullName  string    `gorm:"column:full_name;type:varchar(255)"`
	Role      string    `gorm:"column:role;type:varchar(255)"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

// TableName overrides the table name
func (Admin) TableName() string {
	return "Admin"
}
