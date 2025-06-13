// File: j-ticketing/internal/db/repositories/report_repository.go
package repositories

import (
	"gorm.io/gorm"
	"j-ticketing/internal/db/models"
)

// ReportRepository interface
type ReportRepository interface {
	Create(report *models.Report) error
	FindByID(id uint) (*models.Report, error)
	FindAll() ([]models.Report, error)
	FindByType(reportType string) ([]models.Report, error)
	Update(report *models.Report) error
	SoftDelete(id uint) error
	FindActiveReports() ([]models.Report, error)
	FindByFrequency(frequency string) ([]models.Report, error)
}

// ReportAttachmentRepository interface
type ReportAttachmentRepository interface {
	Create(attachment *models.ReportAttachment) error
	FindByReportID(reportId uint) ([]models.ReportAttachment, error)
	FindByID(id uint) (*models.ReportAttachment, error)
	Delete(id uint) error
	FindByReportAndType(reportId uint, attachmentType string) ([]models.ReportAttachment, error)
}

// ReportRepositoryImpl implements ReportRepository
type ReportRepositoryImpl struct {
	db *gorm.DB
}

// NewReportRepository creates a new report repository
func NewReportRepository(db *gorm.DB) ReportRepository {
	return &ReportRepositoryImpl{db: db}
}

func (r *ReportRepositoryImpl) Create(report *models.Report) error {
	return r.db.Create(report).Error
}

func (r *ReportRepositoryImpl) FindByID(id uint) (*models.Report, error) {
	var report models.Report
	err := r.db.Preload("ReportAttachments").Where("report_id = ? AND is_deleted = ?", id, false).First(&report).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &report, nil
}

func (r *ReportRepositoryImpl) FindAll() ([]models.Report, error) {
	var reports []models.Report
	err := r.db.Preload("ReportAttachments").Where("is_deleted = ?", false).Find(&reports).Error
	return reports, err
}

func (r *ReportRepositoryImpl) FindByType(reportType string) ([]models.Report, error) {
	var reports []models.Report
	err := r.db.Preload("ReportAttachments").Where("type = ? AND is_deleted = ?", reportType, false).Find(&reports).Error
	return reports, err
}

func (r *ReportRepositoryImpl) Update(report *models.Report) error {
	return r.db.Save(report).Error
}

func (r *ReportRepositoryImpl) SoftDelete(id uint) error {
	return r.db.Model(&models.Report{}).Where("report_id = ?", id).Update("is_deleted", true).Error
}

func (r *ReportRepositoryImpl) FindActiveReports() ([]models.Report, error) {
	var reports []models.Report
	err := r.db.Preload("ReportAttachments").Where("is_deleted = ?", false).Find(&reports).Error
	return reports, err
}

func (r *ReportRepositoryImpl) FindByFrequency(frequency string) ([]models.Report, error) {
	var reports []models.Report
	err := r.db.Preload("ReportAttachments").Where("frequency = ? AND is_deleted = ?", frequency, false).Find(&reports).Error
	return reports, err
}
