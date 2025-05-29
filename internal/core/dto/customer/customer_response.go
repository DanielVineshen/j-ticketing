// FILE: internal/core/dto/customer/customer_response.go (Updated)
package dto

import dto "j-ticketing/internal/core/dto/order"

// CustomerResponse represents the response structure for customer profile
type CustomerResponse struct {
	Customer Customer `json:"customerProfile"`
}

type Customer struct {
	CustID           string `json:"custId"`
	Email            string `json:"email"`
	FullName         string `json:"fullName"`
	IdentificationNo string `json:"identificationNo"`
	IsDisabled       bool   `json:"isDisabled"`
	ContactNo        string `json:"contactNo"`
}

type DetailedCustomerResponse struct {
	DetailedCustomers []DetailedCustomer `json:"customers"`
}

type DetailedCustomer struct {
	CustID            string                `json:"custId"`
	Email             string                `json:"email"`
	FullName          string                `json:"fullName"`
	IdentificationNo  string                `json:"identificationNo"`
	IsDisabled        bool                  `json:"isDisabled"`
	ContactNo         string                `json:"contactNo"`
	OrderTicketGroups []dto.OrderProfileDTO `json:"orderTicketGroup"`
	CustomerLogs      []CustomerLog         `json:"customersLogs"`
}

type CustomerLog struct {
	CustLogId uint   `json:"custLogId"`
	CustId    string `json:"custId"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	Date      string `json:"date"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}
