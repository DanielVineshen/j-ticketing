// File: j-ticketing/internal/db/models/audit_log.go
package models

import (
	"database/sql"
	"time"
)

// AuditLog represents the audit_log table
type AuditLog struct {
	AuditLogId     uint           `gorm:"primaryKey;column:audit_log_id;type:bigint unsigned AUTO_INCREMENT"`
	UserId         string         `gorm:"column:user_id;type:varchar(255);not null"`
	PerformedBy    sql.NullString `gorm:"column:performed_by;type:varchar(255);null"`
	AuthorityLevel string         `gorm:"column:authority_level;type:varchar(255);not null"`
	BeforeChanged  sql.NullString `gorm:"column:before_changed;type:text;null"`
	AfterChanged   sql.NullString `gorm:"column:after_changed;type:text;null"`
	LogType        string         `gorm:"column:log_type;type:varchar(255);not null"`
	LogTypeDesc    sql.NullString `gorm:"column:log_type_desc;type:varchar(255);null"`
	LogAction      string         `gorm:"column:log_action;type:varchar(255);not null"`
	LogDesc        sql.NullString `gorm:"column:log_desc;type:text;null"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
}

// TableName overrides the table name
func (AuditLog) TableName() string {
	return "Audit_Log"
}
