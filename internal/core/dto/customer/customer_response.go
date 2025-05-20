// FILE: internal/core/dto/customer/customer_response.go (Updated)
package dto

// CustomerResponse represents the response structure for customer profile
type CustomerResponse struct {
	Customer Customer `json:"customerProfile"`
}

type Customer struct {
	CustID           string `json:"custId,omitempty"`
	Email            string `json:"email,omitempty"`
	FullName         string `json:"fullName,omitempty"`
	IdentificationNo string `json:"identificationNo,omitempty"`
	IsDisabled       bool   `json:"isDisabled,omitempty"`
	ContactNo        string `json:"contactNo,omitempty"`
}
