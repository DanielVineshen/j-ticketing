package models

// BaseErrorResponse represents the standard error response structure
type BaseErrorResponse struct {
	RespCode int         `json:"respCode"`
	RespDesc string      `json:"respDesc"`
	Result   interface{} `json:"result"`
}

// NewBaseErrorResponse creates a new BaseErrorResponse
func NewBaseErrorResponse(code int, message string, result interface{}) *BaseErrorResponse {
	return &BaseErrorResponse{
		RespCode: code,
		RespDesc: message,
		Result:   result,
	}
}

// BaseSuccessResponse represents the standard success response structure
type BaseSuccessResponse struct {
	RespCode int         `json:"respCode"`
	RespDesc string      `json:"respDesc"`
	Result   interface{} `json:"result"`
}

// NewBaseSuccessResponse creates a new BaseSuccessResponse
func NewBaseSuccessResponse(result interface{}) *BaseSuccessResponse {
	return &BaseSuccessResponse{
		RespCode: 2000,
		RespDesc: "Success",
		Result:   result,
	}
}

// GenericMessage represents a simple status message
type GenericMessage struct {
	Status bool `json:"status"`
}

// NewGenericMessage creates a new GenericMessage
func NewGenericMessage(status bool) *GenericMessage {
	return &GenericMessage{
		Status: status,
	}
}
