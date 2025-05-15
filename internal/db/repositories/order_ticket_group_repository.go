// File: internal/db/repositories/order_ticket_group_repository.go
package repositories

import (
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
	result := r.db.Find(&orderTicketGroups)
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
	result := r.db.Where("cust_id = ?", custID).Find(&orderTicketGroups)
	return orderTicketGroups, result.Error
}

// FindWithDetails finds an order ticket group with all its details
func (r *OrderTicketGroupRepository) FindWithDetails(id uint) (*models.OrderTicketGroup, error) {
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

// Create creates a new order ticket group
func (r *OrderTicketGroupRepository) Create(orderTicketGroup *models.OrderTicketGroup) error {
	return r.db.Create(orderTicketGroup).Error
}

// Update updates an order ticket group
func (r *OrderTicketGroupRepository) Update(orderTicketGroup *models.OrderTicketGroup) error {
	return r.db.Save(orderTicketGroup).Error
}
