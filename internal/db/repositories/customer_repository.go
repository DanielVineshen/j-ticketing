// File: j-ticketing/internal/db/repositories/customer_repository.go\
package repositories

import (
	"j-ticketing/internal/db/models"

	"gorm.io/gorm"
)

// CustomerRepository is the interface for customer database operations
type CustomerRepository interface {
	Create(customer *models.Customer) error
	FindByID(id string) (*models.Customer, error)
	FindByEmail(email string) (*models.Customer, error)
	Update(customer *models.Customer) error
	Delete(id string) error
	List() ([]models.Customer, error)
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
