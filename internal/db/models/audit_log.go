// File: internal/core/models/audit_log.go
package models

import (
	"time"
)

// AuditLog represents the audit_log table
type AuditLog struct {
	AuditLogId     uint      `gorm:"primaryKey;column:audit_log_id;type:bigint unsigned AUTO_INCREMENT"`
	UserId         string    `gorm:"column:user_id;type:varchar(255)"`
	PerformedBy    string    `gorm:"column:performed_by;type:varchar(255)"`
	AuthorityLevel string    `gorm:"column:authority_level;type:varchar(255)"`
	BeforeChanged  string    `gorm:"column:before_changed;type:text"`
	AfterChanged   string    `gorm:"column:after_changed;type:text"`
	LogType        string    `gorm:"column:log_type;type:varchar(255)"`
	LogTypeDesc    string    `gorm:"column:log_type_desc;type:varchar(255)"`
	LogAction      string    `gorm:"column:log_action;type:varchar(255)"`
	LogDesc        string    `gorm:"column:log_desc;type:text"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

// TableName overrides the table name
func (AuditLog) TableName() string {
	return "Audit_Log"
}
