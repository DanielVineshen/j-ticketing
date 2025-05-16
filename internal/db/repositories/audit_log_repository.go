// File: j-ticketing/internal/db/repositories/audit_log_repository.go
package repositories

import (
	"j-ticketing/internal/db/models"
	"time"

	"gorm.io/gorm"
)

// AuditLogRepository handles database operations for audit logs
type AuditLogRepository struct {
	db *gorm.DB
}

// NewAuditLogRepository creates a new audit log repository
func NewAuditLogRepository(db *gorm.DB) *AuditLogRepository {
	return &AuditLogRepository{db: db}
}

// Create creates a new audit log
func (r *AuditLogRepository) Create(auditLog *models.AuditLog) error {
	return r.db.Create(auditLog).Error
}

// FindByUserID finds audit logs by user ID
func (r *AuditLogRepository) FindByUserID(userID string) ([]models.AuditLog, error) {
	var auditLogs []models.AuditLog
	result := r.db.Where("user_id = ?", userID).Find(&auditLogs)
	return auditLogs, result.Error
}

// FindByDateRange finds audit logs within a date range
func (r *AuditLogRepository) FindByDateRange(startDate, endDate time.Time) ([]models.AuditLog, error) {
	var auditLogs []models.AuditLog
	result := r.db.Where("created_at BETWEEN ? AND ?", startDate, endDate).Find(&auditLogs)
	return auditLogs, result.Error
}

// FindByLogType finds audit logs by log type
func (r *AuditLogRepository) FindByLogType(logType string) ([]models.AuditLog, error) {
	var auditLogs []models.AuditLog
	result := r.db.Where("log_type = ?", logType).Find(&auditLogs)
	return auditLogs, result.Error
}
