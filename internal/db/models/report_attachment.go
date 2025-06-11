package models

import "time"

// ReportAttachment represents the generated report files
type ReportAttachment struct {
	ReportAttachmentId uint      `json:"reportAttachmentId" gorm:"primaryKey;column:report_attachment_id"`
	ReportId           uint      `json:"reportId" gorm:"column:report_id"`
	Type               string    `json:"type" gorm:"column:type"`
	EmailTo            string    `json:"emailTo" gorm:"column:email_to"`
	AttachmentName     string    `json:"attachmentName" gorm:"column:attachment_name"`
	AttachmentPath     string    `json:"attachmentPath" gorm:"column:attachment_path"`
	AttachmentSize     int64     `json:"attachmentSize" gorm:"column:attachment_size"`
	ContentType        string    `json:"contentType" gorm:"column:content_type"`
	UniqueExtension    string    `json:"uniqueExtension" gorm:"column:unique_extension"`
	CreatedAt          time.Time `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt          time.Time `json:"updatedAt" gorm:"column:updated_at"`

	// Relationships
	Report Report `json:"report" gorm:"foreignKey:ReportId"`
}

func (ReportAttachment) TableName() string {
	return "Report_Attachment"
}
