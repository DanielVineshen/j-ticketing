// FILE: internal/auth/service/customer_service.go
package service

import (
	"database/sql"
	"fmt"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	bcryptPassword "j-ticketing/pkg/utils"
	"time"
)

// CustomerService handles customer-related operations
type CustomerService interface {
	RegisterCustomer(email, password, identificationNo, fullName, contactNo string) (*models.Customer, error)
	GetCustomerByID(id string) (*models.Customer, error)
	UpdateCustomer(id, fullName, contactNo string) (*models.Customer, error)
	ChangePassword(id, currentPassword, newPassword string) error
	DisableCustomer(id string) error
	EnableCustomer(id string) error
	ListCustomers() ([]models.Customer, error)
}

type customerService struct {
	customerRepo repositories.CustomerRepository
}

// NewCustomerService creates a new customer service
func NewCustomerService(customerRepo repositories.CustomerRepository) CustomerService {
	return &customerService{
		customerRepo: customerRepo,
	}
}

// RegisterCustomer registers a new customer
func (s *customerService) RegisterCustomer(email, password, identificationNo, fullName, contactNo string) (*models.Customer, error) {
	// Check if email already exists
	existingCustomer, err := s.customerRepo.FindByEmail(email)
	if err == nil && existingCustomer != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Hash the password
	hashedPassword, err := bcryptPassword.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Generate a unique customer ID
	custID, err := bcryptPassword.GenerateRandomToken(8)
	if err != nil {
		return nil, err
	}
	custID = fmt.Sprintf("CUST-%s", custID[:8])

	// Create the customer
	customer := &models.Customer{
		CustId:           custID,
		Email:            email,
		Password:         sql.NullString{String: hashedPassword, Valid: hashedPassword != ""},
		IdentificationNo: identificationNo,
		FullName:         fullName,
		ContactNo:        sql.NullString{String: contactNo, Valid: contactNo != ""},
		IsDisabled:       false,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	// Save to database
	err = s.customerRepo.Create(customer)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// GetCustomerByID retrieves a customer by ID
func (s *customerService) GetCustomerByID(id string) (*models.Customer, error) {
	return s.customerRepo.FindByID(id)
}

// UpdateCustomer updates a customer's information
func (s *customerService) UpdateCustomer(id, fullName, contactNo string) (*models.Customer, error) {
	customer, err := s.customerRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	customer.FullName = fullName
	customer.ContactNo = sql.NullString{
		String: contactNo,
		Valid:  contactNo != "", // Will be NULL if contactNo is empty
	}
	customer.UpdatedAt = time.Now()

	err = s.customerRepo.Update(customer)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// ChangePassword changes a customer's password
func (s *customerService) ChangePassword(id, currentPassword, newPassword string) error {
	customer, err := s.customerRepo.FindByID(id)
	if err != nil {
		return err
	}

	// Verify current password
	err = bcryptPassword.CheckPassword(currentPassword, customer.Password.String)
	if err != nil {
		return fmt.Errorf("current password is incorrect")
	}

	// Hash the new password
	hashedPassword, err := bcryptPassword.HashPassword(newPassword)
	if err != nil {
		return err
	}

	customer.Password = sql.NullString{
		String: hashedPassword,
		Valid:  hashedPassword != "", // Will be NULL if contactNo is empty
	}

	customer.UpdatedAt = time.Now()

	return s.customerRepo.Update(customer)
}

// DisableCustomer disables a customer's account
func (s *customerService) DisableCustomer(id string) error {
	customer, err := s.customerRepo.FindByID(id)
	if err != nil {
		return err
	}

	customer.IsDisabled = true
	customer.UpdatedAt = time.Now()

	return s.customerRepo.Update(customer)
}

// EnableCustomer enables a customer's account
func (s *customerService) EnableCustomer(id string) error {
	customer, err := s.customerRepo.FindByID(id)
	if err != nil {
		return err
	}

	customer.IsDisabled = false
	customer.UpdatedAt = time.Now()

	return s.customerRepo.Update(customer)
}

// ListCustomers lists all customers
func (s *customerService) ListCustomers() ([]models.Customer, error) {
	customers, err := s.customerRepo.List()
	if err != nil {
		return nil, err
	}

	// Remove password from response
	for i := range customers {
		customers[i].Password = sql.NullString{
			String: "",
			Valid:  false,
		}
	}

	return customers, nil
}
