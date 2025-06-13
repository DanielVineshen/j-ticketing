// File: j-ticketing/internal/core/dto/notification/notification_request.go
package dto

import "j-ticketing/pkg/validation"

type UpdateNotificationRequest struct {
	NotificationId uint  `json:"notificationId" validate:"required,min=1"`
	IsRead         *bool `json:"isRead"`
	IsDeleted      *bool `json:"isDeleted"`
}

// Validate validates the update customer request
func (r *UpdateNotificationRequest) Validate() error {
	return validation.ValidateStruct(r)
}
