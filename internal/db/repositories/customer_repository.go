// File: j-ticketing/internal/db/repositories/customer_repository.go
package repositories

import (
	"fmt"
	"j-ticketing/internal/db/models"
	"j-ticketing/pkg/utils"
	"time"

	"gorm.io/gorm"
)

// CustomerRepository is the interface for customer database operations
type CustomerRepository interface {
	Create(customer *models.Customer) error
	FindAll() ([]models.Customer, error)
	FindByID(id string) (*models.Customer, error)
	FindByEmail(email string) (*models.Customer, error)
	Update(customer *models.Customer) error
	Delete(id string) error
	List() ([]models.Customer, error)
	FindByDateRange(startDate, endDate string) ([]models.Customer, error)
}

type customerRepository struct {
	db *gorm.DB
}

// NewCustomerRepository creates a new customer repository
func NewCustomerRepository(db *gorm.DB) CustomerRepository {
	return &customerRepository{
		db: db,
	}
}

// Create creates a new customer
func (r *customerRepository) Create(customer *models.Customer) error {
	return r.db.Create(customer).Error
}

func (r *customerRepository) FindAll() ([]models.Customer, error) {
	var customers []models.Customer
	result := r.db.
		Preload("CustomerLogs", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("OrderTicketGroups", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("OrderTicketGroups.OrderTicketInfos", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Preload("OrderTicketGroups.OrderTicketLogs", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at DESC")
		}).
		Order("created_at DESC").
		Find(&customers)

	return customers, result.Error
}

// FindByID finds a customer by ID
func (r *customerRepository) FindByID(id string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.Where("cust_id = ?", id).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// FindByEmail finds a customer by email
func (r *customerRepository) FindByEmail(email string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.Where("email = ?", email).First(&customer).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

// Update updates a customer
func (r *customerRepository) Update(customer *models.Customer) error {
	return r.db.Save(customer).Error
}

// Delete deletes a customer
func (r *customerRepository) Delete(id string) error {
	return r.db.Delete(&models.Customer{}, "cust_id = ?", id).Error
}

// List lists all customers
func (r *customerRepository) List() ([]models.Customer, error) {
	var customers []models.Customer
	err := r.db.Find(&customers).Error
	if err != nil {
		return nil, err
	}
	return customers, nil
}

func (r *customerRepository) FindByDateRange(startDate, endDate string) ([]models.Customer, error) {
	startTime, err := time.Parse(utils.DateOnlyFormat, startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date format: %v", err)
	}

	endTime, err := time.Parse(utils.DateOnlyFormat, endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid end date format: %v", err)
	}

	// Set end time to end of day
	endTime = endTime.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	var customers []models.Customer
	result := r.db.Where("created_at >= ? AND created_at <= ?", startTime, endTime).
		Order("created_at ASC").
		Find(&customers)

	return customers, result.Error
}
