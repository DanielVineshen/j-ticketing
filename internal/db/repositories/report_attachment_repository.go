// File: j-ticketing/internal/db/repositories/report_attachment_repository.go
package repositories

import (
	"gorm.io/gorm"
	"j-ticketing/internal/db/models"
)

// ReportAttachmentRepositoryImpl implements ReportAttachmentRepository
type ReportAttachmentRepositoryImpl struct {
	db *gorm.DB
}

// NewReportAttachmentRepository creates a new report attachment repository
func NewReportAttachmentRepository(db *gorm.DB) ReportAttachmentRepository {
	return &ReportAttachmentRepositoryImpl{db: db}
}

func (r *ReportAttachmentRepositoryImpl) Create(attachment *models.ReportAttachment) error {
	return r.db.Create(attachment).Error
}

func (r *ReportAttachmentRepositoryImpl) FindByReportID(reportId uint) ([]models.ReportAttachment, error) {
	var attachments []models.ReportAttachment
	err := r.db.Preload("Report").Where("report_id = ?", reportId).Find(&attachments).Error
	return attachments, err
}

func (r *ReportAttachmentRepositoryImpl) FindByID(id uint) (*models.ReportAttachment, error) {
	var attachment models.ReportAttachment
	err := r.db.Preload("Report").First(&attachment, id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &attachment, nil
}

func (r *ReportAttachmentRepositoryImpl) Delete(id uint) error {
	return r.db.Delete(&models.ReportAttachment{}, id).Error
}

func (r *ReportAttachmentRepositoryImpl) FindByReportAndType(reportId uint, attachmentType string) ([]models.ReportAttachment, error) {
	var attachments []models.ReportAttachment
	err := r.db.Preload("Report").Where("report_id = ? AND type = ?", reportId, attachmentType).Find(&attachments).Error
	return attachments, err
}
