// File: j-ticketing/internal/core/dto/order/order_request.go
package dto

import (
	"errors"
	"j-ticketing/pkg/validation"
)

// CreateOrderRequest represents the request structure for creating a new order
type CreateOrderRequest struct {
	TicketGroupId    uint            `json:"ticketGroupId" validate:"required"`
	IdentificationNo string          `json:"identificationNo"`                 // Made optional
	FullName         string          `json:"fullName"`                         // Made optional
	Email            string          `json:"email" validate:"omitempty,email"` // Made optional but must be valid if present
	ContactNo        string          `json:"contactNo"`                        // Made optional
	Date             string          `json:"date" validate:"required"`
	LangChosen       string          `json:"langChosen" validate:"required,oneof=bm en cn"`
	Tickets          []TicketRequest `json:"tickets" validate:"required,dive"`
	PaymentType      string          `json:"paymentType" validate:"required,oneof=credit/debit fpx"`
	Mode             string          `json:"mode" validate:"omitempty,oneof=individual corporate"`
	BankCode         string          `json:"bankCode" validate:"omitempty"`
}

// CreateFreeOrderRequest represents the request structure for creating a new free order
type CreateFreeOrderRequest struct {
	TicketGroupId    uint            `json:"ticketGroupId" validate:"required"`
	IdentificationNo string          `json:"identificationNo"`                 // Made optional
	FullName         string          `json:"fullName"`                         // Made optional
	Email            string          `json:"email" validate:"omitempty,email"` // Made optional but must be valid if present
	ContactNo        string          `json:"contactNo"`                        // Made optional
	Date             string          `json:"date" validate:"required"`
	LangChosen       string          `json:"langChosen" validate:"required,oneof=bm en cn"`
	Tickets          []TicketRequest `json:"tickets" validate:"required,dive"`
	AllowBypass      bool            `json:"AllowBypass" validate:"omitempty"`
}

// TicketRequest represents a ticket item in the order request
type TicketRequest struct {
	TicketId string `json:"ticketId" validate:"required"`
	Qty      int    `json:"qty" validate:"required,min=1"`
}

// Validate validates the create order request
func (r *CreateOrderRequest) Validate() error {
	// Basic validation using the validation package
	if err := validation.ValidateStruct(r); err != nil {
		return err
	}

	// Additional validation
	if r.PaymentType == "fpx" && r.Mode == "" {
		return errors.New("mode is required for FPX payment type")
	}

	if r.PaymentType == "fpx" && r.BankCode == "" {
		return errors.New("bank code is required for FPX payment type")
	}

	// Check if any of the personal information fields are present
	// If one is present, all must be present
	hasIdentificationNo := r.IdentificationNo != ""
	hasFullName := r.FullName != ""
	hasEmail := r.Email != ""
	hasContactNo := r.ContactNo != ""

	// If any field is present but not all fields are present
	if (hasIdentificationNo || hasFullName || hasEmail || hasContactNo) &&
		!(hasIdentificationNo && hasFullName && hasEmail && hasContactNo) {
		return errors.New("if any of identificationNo, fullName, email, or contactNo is provided, all must be provided")
	}

	return nil
}

func (r *CreateFreeOrderRequest) Validate() error {
	// Basic validation using the validation package
	if err := validation.ValidateStruct(r); err != nil {
		return err
	}

	// Check if any of the personal information fields are present
	// If one is present, all must be present
	hasIdentificationNo := r.IdentificationNo != ""
	hasFullName := r.FullName != ""
	hasEmail := r.Email != ""
	hasContactNo := r.ContactNo != ""

	// If any field is present but not all fields are present
	if (hasIdentificationNo || hasFullName || hasEmail || hasContactNo) &&
		!(hasIdentificationNo && hasFullName && hasEmail && hasContactNo) {
		return errors.New("if any of identificationNo, fullName, email, or contactNo is provided, all must be provided")
	}

	return nil
}
