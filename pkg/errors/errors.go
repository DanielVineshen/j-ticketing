// File: j-ticketing/pkg/errors/errors.go
package errors

// ErrorCode represents various error codes and their associated messages
// used throughout the application.
type ErrorCode struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Error codes constants
var (
	// Success operation
	SUCCESS = ErrorCode{2000, "Success"}

	// Client-side errors (4xxx range)
	INVALID_INPUT_FORMAT = ErrorCode{4000, "Invalid input format was provided."}
	INVALID_INPUT_VALUES = ErrorCode{4001, "Invalid input values were provided."}
	USER_NOT_AUTHORIZED  = ErrorCode{4002, "User is not authorized."}
	NOT_FOUND            = ErrorCode{4003, "Record(s) not found."}
	USER_NOT_PERMITTED   = ErrorCode{4004, "User is not permitted to perform this action."}

	// Server-side errors (5xxx range)
	PROCESSING_ERROR                 = ErrorCode{5000, "Something went wrong when processing request."}
	ENTITY_NOT_FOUND_EXCEPTION       = ErrorCode{5003, "Entity not found exception occurred."}
	UNSUPPORTED_MEDIA_TYPE_EXCEPTION = ErrorCode{5006, "Unsupported media type exception occurred."}

	// Add more error codes as needed
)
