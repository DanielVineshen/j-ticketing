// File: j-ticketing/internal/core/dto/order/order_request.go
package dto

import (
	"errors"
	"j-ticketing/pkg/validation"
)

// CreateOrderRequest represents the request structure for creating a new order
type CreateOrderRequest struct {
	TicketGroupId    uint            `json:"ticketGroupId" validate:"required"`
	IdentificationNo string          `json:"identificationNo" validate:"required"`
	FullName         string          `json:"fullName" validate:"required"`
	Email            string          `json:"email" validate:"required,email"`
	ContactNo        string          `json:"contactNo" validate:"required"`
	Date             string          `json:"date" validate:"required"`
	Tickets          []TicketRequest `json:"tickets" validate:"required,dive"`
	PaymentType      string          `json:"paymentType" validate:"required,oneof=credit/debit fpx"`
	Mode             string          `json:"mode" validate:"omitempty,oneof=individual corporate"`
	BankCode         string          `json:"bankCode" validate:"omitempty"`
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

	return nil
}
