// File: j-ticketing/internal/core/services/customer_service.go
package service

import (
	"database/sql"
	"fmt"
	customerDto "j-ticketing/internal/core/dto/customer"
	dto "j-ticketing/internal/core/dto/customer"
	orderDto "j-ticketing/internal/core/dto/order"
	"j-ticketing/internal/db/models"
	"j-ticketing/internal/db/repositories"
	"j-ticketing/pkg/utils"
	"time"
)

// CustomerService handles customer-related operations
//type CustomerService interface {
//	GetAllCustomers() (customerDto.DetailedCustomerResponse, error)
//	RegisterCustomer(email, password, identificationNo, fullName, contactNo string) (*models.Customer, error)
//	GetCustomerByID(id string) (*models.Customer, error)
//	UpdateCustomer(id string, req dto.UpdateCustomerRequest) (*models.Customer, error)
//	ChangePassword(id, currentPassword, newPassword string) (*models.Customer, error)
//	ListCustomers() ([]models.Customer, error)
//	GetCustomerByEmail(email string) (*models.Customer, error)
//	CreateCustomerLog(logType string, title string, message string, customer models.Customer) error
//}

type CustomerService struct {
	customerRepo    repositories.CustomerRepository
	customerLogRepo *repositories.CustomerLogRepository
}

// NewCustomerService creates a new customer service
func NewCustomerService(customerRepo repositories.CustomerRepository, customerLogRepo *repositories.CustomerLogRepository) *CustomerService {
	return &CustomerService{
		customerRepo:    customerRepo,
		customerLogRepo: customerLogRepo,
	}
}

func (s *CustomerService) GetAllCustomers() (customerDto.DetailedCustomerResponse, error) {
	customers, err := s.customerRepo.FindAll()
	if err != nil {
		return customerDto.DetailedCustomerResponse{}, err
	}

	response := customerDto.DetailedCustomerResponse{
		DetailedCustomers: make([]customerDto.DetailedCustomer, 0, len(customers)),
	}

	for _, customer := range customers {
		customerDTO, err := s.mapCustomerToDTO(&customer)
		if err != nil {
			return customerDto.DetailedCustomerResponse{}, err
		}

		response.DetailedCustomers = append(response.DetailedCustomers, customerDTO)
	}

	return response, nil
}

func (s *CustomerService) mapCustomerToDTO(customer *models.Customer) (customerDto.DetailedCustomer, error) {
	cust := customerDto.DetailedCustomer{
		CustID:            customer.CustId,
		Email:             customer.Email,
		FullName:          customer.FullName,
		IdentificationNo:  customer.IdentificationNo,
		IsDisabled:        customer.IsDisabled,
		ContactNo:         getStringFromNullString(customer.ContactNo),
		OrderTicketGroups: make([]orderDto.OrderProfileDTO, 0),
		CustomerLogs:      make([]customerDto.CustomerLog, 0),
	}

	customerLogs := customer.CustomerLogs

	for _, log := range customerLogs {
		customerLog := customerDto.CustomerLog{
			CustLogId: log.CustLogId,
			CustId:    log.CustId,
			Type:      log.Type,
			Title:     log.Title,
			Message:   log.Message,
			Date:      log.Date,
			CreatedAt: log.CreatedAt.Format(time.RFC3339),
			UpdatedAt: log.UpdatedAt.Format(time.RFC3339),
		}

		cust.CustomerLogs = append(cust.CustomerLogs, customerLog)
	}

	orderTicketGroups := customer.OrderTicketGroups

	for _, order := range orderTicketGroups {
		orderProfile := orderDto.OrderProfileDTO{
			OrderTicketGroupId: order.OrderTicketGroupId,
			TicketGroupId:      order.TicketGroupId,
			CustId:             order.CustId,
			TransactionId:      order.TransactionId,
			OrderNo:            order.OrderNo,
			TransactionStatus:  order.TransactionStatus,
			TransactionDate:    order.TransactionDate,
			MsgToken:           order.MsgToken,
			BillId:             order.BillId,
			ProductId:          order.ProductId,
			TotalAmount:        order.TotalAmount,
			BuyerName:          order.BuyerName,
			BuyerEmail:         order.BuyerEmail,
			ProductDesc:        order.ProductDesc,
			OrderTicketInfo:    make([]orderDto.OrderTicketInfoDTO, 0),
			OrderTicketLog:     make([]orderDto.OrderTicketLogDTO, 0),
			CreatedAt:          order.CreatedAt.Format(time.RFC3339),
			UpdatedAt:          order.UpdatedAt.Format(time.RFC3339),
		}

		// Use preloaded OrderTicketInfos directly (no database call)
		for _, info := range order.OrderTicketInfos {
			infoDTO := orderDto.OrderTicketInfoDTO{
				OrderTicketInfoId:  info.OrderTicketInfoId,
				OrderTicketGroupId: info.OrderTicketGroupId,
				ItemId:             info.ItemId,
				UnitPrice:          info.UnitPrice,
				ItemDesc1:          info.ItemDesc1,
				ItemDesc2:          info.ItemDesc2,
				PrintType:          info.PrintType,
				QuantityBought:     info.QuantityBought,
				EncryptedId:        info.EncryptedId,
				AdmitDate:          info.AdmitDate,
				Variant:            info.Variant,
				CreatedAt:          info.CreatedAt.Format(time.RFC3339),
				UpdatedAt:          info.UpdatedAt.Format(time.RFC3339),
			}

			if info.Twbid.Valid {
				infoDTO.Twbid = info.Twbid.String
			}

			orderProfile.OrderTicketInfo = append(orderProfile.OrderTicketInfo, infoDTO)
		}

		for _, ticket := range order.OrderTicketLogs {
			ticketDTO := orderDto.OrderTicketLogDTO{
				OrderTicketLogId:   ticket.OrderTicketLogId,
				OrderTicketGroupId: ticket.OrderTicketGroupId,
				PerformedBy:        getStringFromNullString(ticket.PerformedBy),
				Type:               ticket.Type,
				Title:              ticket.Title,
				Message:            ticket.Message,
				Date:               ticket.Date,
				CreatedAt:          ticket.CreatedAt.Format(time.RFC3339),
				UpdatedAt:          ticket.UpdatedAt.Format(time.RFC3339),
			}

			orderProfile.OrderTicketLog = append(orderProfile.OrderTicketLog, ticketDTO)
		}

		cust.OrderTicketGroups = append(cust.OrderTicketGroups, orderProfile)
	}

	return cust, nil
}

func (s *CustomerService) GetCustomerByEmail(email string) (*models.Customer, error) {
	// This simply calls the repository's FindByEmail method
	return s.customerRepo.FindByEmail(email)
}

// RegisterCustomer registers a new customer
func (s *CustomerService) RegisterCustomer(email, password, identificationNo, fullName, contactNo string) (*models.Customer, error) {
	// Check if email already exists
	existingCustomer, err := s.customerRepo.FindByEmail(email)
	if err == nil && existingCustomer != nil {
		return nil, fmt.Errorf("email already exists")
	}

	// Initialize password as a null string
	var passwordField sql.NullString

	// Only hash and set password if it's provided
	if password != "" {
		// Hash the password
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			return nil, err
		}

		// Set the password field with the hashed password
		passwordField = sql.NullString{
			String: hashedPassword,
			Valid:  true,
		}
	} else {
		// If no password is provided, leave it as NULL in the database
		passwordField = sql.NullString{
			String: "",
			Valid:  false,
		}
	}

	// Generate a unique customer ID
	custID, err := utils.GenerateRandomToken(8)
	if err != nil {
		return nil, err
	}
	custID = fmt.Sprintf("CUST-%s", custID[:8])

	// Create the customer
	customer := &models.Customer{
		CustId:           custID,
		Email:            email,
		Password:         passwordField,
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
func (s *CustomerService) GetCustomerByID(id string) (*models.Customer, error) {
	return s.customerRepo.FindByID(id)
}

// UpdateCustomer updates a customer's information
func (s *CustomerService) UpdateCustomer(id string, req dto.UpdateCustomerRequest) (*models.Customer, error) {
	customer, err := s.customerRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	customer.Email = req.Email
	customer.FullName = req.FullName
	customer.IdentificationNo = req.IdentificationNo

	// Handle contact number (can be null)
	customer.ContactNo = sql.NullString{
		String: req.ContactNo,
		Valid:  req.ContactNo != "",
	}

	customer.UpdatedAt = time.Now()

	err = s.customerRepo.Update(customer)
	if err != nil {
		return nil, err
	}

	return customer, nil
}

// ChangePassword changes a customer's password
func (s *CustomerService) ChangePassword(id, currentPassword, newPassword string) (*models.Customer, error) {
	customer, err := s.customerRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	// Verify current password
	err = utils.CheckPassword(currentPassword, customer.Password.String)
	if err != nil {
		return nil, fmt.Errorf("current password is incorrect")
	}

	// Hash the new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return nil, err
	}

	customer.Password = sql.NullString{
		String: hashedPassword,
		Valid:  hashedPassword != "", // Will be NULL if contactNo is empty
	}

	customer.UpdatedAt = time.Now()

	return customer, s.customerRepo.Update(customer)
}

// ListCustomers lists all customers
func (s *CustomerService) ListCustomers() ([]models.Customer, error) {
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

func (s *CustomerService) CreateCustomerLog(logType string, title string, message string, customer models.Customer) error {
	malaysiaTime, err := utils.FormatCurrentMalaysiaTime(utils.FullDateTimeFormat)
	if err != nil {
		return err
	}

	customerLog := models.CustomerLog{
		CustId:  customer.CustId,
		Type:    logType,
		Title:   title,
		Message: message,
		Date:    malaysiaTime,
	}

	err = s.customerLogRepo.Create(&customerLog)
	if err != nil {
		return err
	}

	return nil
}

func getStringFromNullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return "" // Return empty string if NULL
}
