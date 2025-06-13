// File: j-ticketing/internal/db/models/report.go
package models

import "time"

// Report represents the main report configuration
type Report struct {
	ReportId    uint      `json:"reportId" gorm:"primaryKey;column:report_id"`
	Title       string    `json:"title" gorm:"column:title"`
	Type        string    `json:"type" gorm:"column:type"`                // onsite_visitors, sales, members, online_visitors
	DataOptions string    `json:"dataOptions" gorm:"column:data_options"` // semicolon-separated
	Frequency   string    `json:"frequency" gorm:"column:frequency"`      // one_time, daily, weekly, monthly, quarterly, annual
	EmailTo     string    `json:"emailTo" gorm:"column:email_to"`
	Desc        *string   `json:"desc" gorm:"column:desc"`
	IsDeleted   bool      `json:"isDeleted" gorm:"column:is_deleted;default:false"`
	CreatedAt   time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt   time.Time `json:"updatedAt" gorm:"column:updated_at"`

	// Relationships
	ReportAttachments []ReportAttachment `json:"attachments" gorm:"foreignKey:ReportId"`
}

func (Report) TableName() string {
	return "Report"
}
