// internal/core/dto/ticket_group_dto.go
package dto

type CreateTicketGroupRequest struct {
	GroupName string `json:"groupName"`
	GroupDesc string `json:"groupDesc"`
	// Other fields...
}
