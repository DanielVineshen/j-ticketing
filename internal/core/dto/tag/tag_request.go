// File: j-ticketing/internal/core/dto/tag/tag_request.go
package dto

import (
	"j-ticketing/pkg/validation"
)

type CreateNewTagRequest struct {
	TagName string `json:"tagName" validate:"required,max=255"`
	TagDesc string `json:"tagDesc" validate:"required,max=255"`
}

func (r *CreateNewTagRequest) Validate() error {
	return validation.ValidateStruct(r)
}

type UpdateTagRequest struct {
	TagId   uint   `json:"tagId" validate:"required"`
	TagName string `json:"tagName" validate:"required,max=255"`
	TagDesc string `json:"tagDesc" validate:"required,max=255"`
}

func (r *UpdateTagRequest) Validate() error {
	return validation.ValidateStruct(r)
}

type DeleteTagRequest struct {
	TagId uint `json:"tagId" validate:"required"`
}

func (r *DeleteTagRequest) Validate() error {
	return validation.ValidateStruct(r)
}
