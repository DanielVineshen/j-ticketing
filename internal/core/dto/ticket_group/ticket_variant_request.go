// File: j-ticketing/internal/core/dto/ticket_group/ticket_variant_request.go
package dto

// TicketVariantRequest represents the request structure for ticket variants
type TicketVariantRequest struct {
	TicketGroupId uint   `json:"ticketGroupId" validate:"required"`
	Date          string `json:"date" validate:"required"`
}

// Validate validates the ticket variant request
func (r *TicketVariantRequest) Validate() error {
	return nil // We'll add validation later if needed
}

// TicketVariantResponse represents the response structure for ticket variants
type TicketVariantResponse struct {
	TicketVariants []TicketVariantDTO `json:"ticketVariants"`
}

// TicketVariantDTO represents the data transfer object for a ticket variant
type TicketVariantDTO struct {
	TicketId  string  `json:"ticketId"`
	UnitPrice float64 `json:"unitPrice"`
	ItemDesc1 string  `json:"itemDesc1"`
	ItemDesc2 string  `json:"itemDesc2"`
	ItemDesc3 string  `json:"itemDesc3"`
	PrintType string  `json:"printType"`
	Qty       int     `json:"qty"`
}
