// File: j-ticketing/internal/db/models/config.go
package models

import (
	"time"
)

// Config represents the config table
type Config struct {
	ConfigId  uint      `gorm:"primaryKey;column:config_id;type:bigint unsigned AUTO_INCREMENT"`
	Key       string    `gorm:"column:key;type:varchar(255);not null"`
	Value     string    `gorm:"column:value;type:varchar(255);not null"`
	CreatedAt time.Time `gorm:"column:created_at;type:datetime;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:datetime;not null"`

	// Relationships defined without foreign key constraints
	// Add any relationships here if needed in the future
}

// TableName overrides the table name
func (Config) TableName() string {
	return "Config"
}
