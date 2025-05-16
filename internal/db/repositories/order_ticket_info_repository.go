// File: j-ticketing/internal/db/repositories/order_ticket_info_repository.go
package repositories

import (
	"j-ticketing/internal/db/models"

	"gorm.io/gorm"
)

// OrderTicketInfoRepository handles database operations for order ticket info
type OrderTicketInfoRepository struct {
	db *gorm.DB
}

// NewOrderTicketInfoRepository creates a new order ticket info repository
func NewOrderTicketInfoRepository(db *gorm.DB) *OrderTicketInfoRepository {
	return &OrderTicketInfoRepository{db: db}
}

// FindByOrderTicketGroupID finds order ticket infos by order ticket group ID
func (r *OrderTicketInfoRepository) FindByOrderTicketGroupID(orderTicketGroupID uint) ([]models.OrderTicketInfo, error) {
	var orderTicketInfos []models.OrderTicketInfo
	result := r.db.Where("order_ticket_group_id = ?", orderTicketGroupID).Find(&orderTicketInfos)
	return orderTicketInfos, result.Error
}

// FindByID finds an order ticket info by ID
func (r *OrderTicketInfoRepository) FindByID(id uint) (*models.OrderTicketInfo, error) {
	var orderTicketInfo models.OrderTicketInfo
	result := r.db.First(&orderTicketInfo, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &orderTicketInfo, nil
}

// Create creates a new order ticket info
func (r *OrderTicketInfoRepository) Create(orderTicketInfo *models.OrderTicketInfo) error {
	return r.db.Create(orderTicketInfo).Error
}

// BatchCreate creates multiple order ticket infos
func (r *OrderTicketInfoRepository) BatchCreate(orderTicketInfos []models.OrderTicketInfo) error {
	return r.db.Create(&orderTicketInfos).Error
}

// Update updates an order ticket info
func (r *OrderTicketInfoRepository) Update(orderTicketInfo *models.OrderTicketInfo) error {
	return r.db.Save(orderTicketInfo).Error
}
