// File: j-ticketing/internal/db/repositories/order_ticket_group_repository.go
package repositories

import (
	"errors"
	"j-ticketing/internal/db/models"

	"gorm.io/gorm"
)

// OrderTicketGroupRepository handles database operations for order ticket groups
type OrderTicketGroupRepository struct {
	db *gorm.DB
}

// NewOrderTicketGroupRepository creates a new order ticket group repository
func NewOrderTicketGroupRepository(db *gorm.DB) *OrderTicketGroupRepository {
	return &OrderTicketGroupRepository{db: db}
}

// FindAll returns all order ticket groups
func (r *OrderTicketGroupRepository) FindAll() ([]models.OrderTicketGroup, error) {
	var orderTicketGroups []models.OrderTicketGroup
	result := r.db.Preload("Customer").
		Preload("OrderTicketLogs").
		Preload("TicketGroup").
		Preload("TicketGroup.TicketTags").
		Preload("TicketGroup.TicketTags.Tag").
		Preload("TicketGroup.GroupGalleries").
		Preload("TicketGroup.TicketDetails").
		Preload("OrderTicketInfos").
		Find(&orderTicketGroups)
	return orderTicketGroups, result.Error
}

// FindByID finds an order ticket group by ID
func (r *OrderTicketGroupRepository) FindByID(id uint) (*models.OrderTicketGroup, error) {
	var orderTicketGroup models.OrderTicketGroup
	result := r.db.First(&orderTicketGroup, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &orderTicketGroup, nil
}

// FindByCustomerID finds order ticket groups by customer ID
func (r *OrderTicketGroupRepository) FindByCustomerID(custID string) ([]models.OrderTicketGroup, error) {
	var orderTicketGroups []models.OrderTicketGroup
	result := r.db.Where("cust_id = ?", custID).
		Preload("OrderTicketLogs").
		Preload("TicketGroup").
		Preload("TicketGroup.TicketTags").
		Preload("TicketGroup.TicketTags.Tag").
		Preload("TicketGroup.GroupGalleries").
		Preload("TicketGroup.TicketDetails").
		Preload("OrderTicketInfos").Find(&orderTicketGroups)
	return orderTicketGroups, result.Error
}

// FindWithOrderTicketGroupId finds an order ticket group with all its details
func (r *OrderTicketGroupRepository) FindWithOrderTicketGroupId(id uint) (*models.OrderTicketGroup, error) {
	var orderTicketGroup models.OrderTicketGroup
	result := r.db.Preload("OrderTicketInfos").
		Preload("TicketGroup").
		Preload("Customer").
		First(&orderTicketGroup, id)
	if result.Error != nil {
		return nil, result.Error
	}
	return &orderTicketGroup, nil
}

// FindWithOrderTicketGroupId finds an order ticket group with all its details
func (r *OrderTicketGroupRepository) FindWithOrderNoAndEmail(orderNo string, email string) (*models.OrderTicketGroup, error) {
	var order models.OrderTicketGroup

	result := r.db.Where("order_no = ? AND buyer_email = ?", orderNo, email).First(&order)
	if result.Error != nil {
		return nil, result.Error
	}
	return &order, nil
}

func (r *OrderTicketGroupRepository) FindEmailPendingOrderTicketGroups() ([]models.OrderTicketGroup, error) {
	var orders []models.OrderTicketGroup

	result := r.db.Where("transaction_status = 'success' AND is_email_sent = 0 AND transaction_date != ''").Find(&orders)
	if result.Error != nil {
		return nil, result.Error
	}
	return orders, nil
}

// FindByOrderNo finds a order ticket group by its order number
func (r *OrderTicketGroupRepository) FindByOrderNo(orderNo string) (*models.OrderTicketGroup, error) {
	var order models.OrderTicketGroup

	result := r.db.Where("order_no = ?", orderNo).First(&order)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil // No error, but order not found
		}
		return nil, result.Error
	}

	return &order, nil
}

// Create creates a new order ticket group
func (r *OrderTicketGroupRepository) Create(orderTicketGroup *models.OrderTicketGroup) error {
	return r.db.Create(orderTicketGroup).Error
}

// Update updates an order ticket group
func (r *OrderTicketGroupRepository) Update(orderTicketGroup *models.OrderTicketGroup) error {
	return r.db.Save(orderTicketGroup).Error
}
