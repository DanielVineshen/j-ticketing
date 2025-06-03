// File: j-ticketing/internal/core/dto/tag/tag_response.go
package dto

type TagListResponse struct {
	Tag []Tag `json:"tags"`
}

// NewBannerListResponse creates a new banner list response with guaranteed non-nil slice
func NewTagListResponse(tags []Tag) *TagListResponse {
	if tags == nil {
		tags = make([]Tag, 0)
	}
	return &TagListResponse{
		Tag: tags,
	}
}

type Tag struct {
	TagId   uint   `json:"tagId"`
	TagName string `json:"tagName"`
	TagDesc string `json:"tagDesc"`
}
