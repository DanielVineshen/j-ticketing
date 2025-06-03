package dto

import (
	"fmt"
	"j-ticketing/pkg/validation"
)

// UpdatePlacementRequest represents the request to update ticket group placements
type UpdatePlacementRequest struct {
	TicketGroups []PlacementItem `json:"ticketGroups" validate:"required,min=1,dive"`
}

// PlacementItem represents a single ticket group placement update
type PlacementItem struct {
	TicketGroupId uint `json:"ticketGroupId" validate:"required,min=1"`
	Placement     int  `json:"placement" validate:"required,min=1"`
}

// Validate validates the update placement request
func (r *UpdatePlacementRequest) Validate() error {
	if err := validation.ValidateStruct(r); err != nil {
		return err
	}

	// Additional validation: check for duplicate ticket group IDs
	seen := make(map[uint]bool)
	for _, item := range r.TicketGroups {
		if seen[item.TicketGroupId] {
			return fmt.Errorf("duplicate ticket group ID: %d", item.TicketGroupId)
		}
		seen[item.TicketGroupId] = true
	}

	// Check for duplicate placement values
	placementSeen := make(map[int]bool)
	for _, item := range r.TicketGroups {
		if placementSeen[item.Placement] {
			return fmt.Errorf("duplicate placement value: %d", item.Placement)
		}
		placementSeen[item.Placement] = true
	}

	return nil
}
