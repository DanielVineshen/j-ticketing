// File: j-ticketing/internal/db/models/notification.go
package models

import (
	"database/sql"
	"time"
)

// Notification represents the notification table
type Notification struct {
	NotificationId uint           `gorm:"primaryKey;column:notification_id;type:bigint unsigned AUTO_INCREMENT"`
	PerformedBy    sql.NullString `gorm:"column:performed_by;type:varchar(255);null"`
	AuthorityLevel string         `gorm:"column:authority_level;type:varchar(255);not null"`
	Type           string         `gorm:"column:type;type:varchar(255);not null"`
	Title          string         `gorm:"column:title;type:varchar(255);not null"`
	Message        sql.NullString `gorm:"column:message;type:text;not null"`
	Date           string         `gorm:"column:date;type:char(14);null"`
	IsRead         bool           `gorm:"column:is_read;type:boolean;default:false;not null"`
	IsDeleted      bool           `gorm:"column:is_deleted;type:boolean;default:false;not null"`
	CreatedAt      time.Time      `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;type:datetime;not null"`

	// Relationships defined without foreign key constraints
	// Add any relationships here if needed in the future
}

// TableName overrides the table name
func (Notification) TableName() string {
	return "Notification"
}
