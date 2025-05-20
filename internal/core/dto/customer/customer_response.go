// FILE: internal/core/dto/customer/customer_response.go (Updated)
package dto

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
