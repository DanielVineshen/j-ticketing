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
	INVALID_INPUT_FORMAT      = ErrorCode{4000, "Invalid input format was provided."}
	INVALID_INPUT_VALUES      = ErrorCode{4001, "Invalid input values were provided."}
	INVALID_CONSTRAINT_VALUES = ErrorCode{4002, "Invalid constraint values were provided."}
	USER_NOT_AUTHORIZED       = ErrorCode{4003, "User is not authorized."}
	USER_ACCOUNT_DELETED      = ErrorCode{4004, "User account has been deleted."}
	INVALID_REFRESH_TOKEN     = ErrorCode{4005, "Refresh token is not valid."}
	USER_NOT_EXIST            = ErrorCode{4006, "User does not exist."}
	INVALID_ACCESS_TOKEN      = ErrorCode{4007, "Access token is not valid."}
	MISSING_AUTH_TOKEN_HEADER = ErrorCode{4008, "Missing authorization token header."}
	CATEGORY_TYPE_INVALID     = ErrorCode{4009, "Category type is not valid"}
	USER_ROLE_INVALID         = ErrorCode{4010, "User role is not valid."}
	USER_NOT_PERMITTED        = ErrorCode{4011, "User is not permitted to perform this action."}
	INVALID_CREDENTIALS       = ErrorCode{4012, "Invalid credentials were provided."}
	DECODING_ERROR            = ErrorCode{4013, "There was a problem decoding/decrypting the response body."}
	MAPPING_ERROR             = ErrorCode{4014, "There was a problem mapping the response body."}
	FILE_NOT_FOUND            = ErrorCode{4015, "File not found."}
	FILE_SIZE_ERROR           = ErrorCode{4016, "File size exceeds maximum limit of 5MB."}
	FILE_TYPE_ERROR           = ErrorCode{4017, "Only JPEG and PNG file types are allowed."}
	USER_LACKS_PERMISSION     = ErrorCode{4018, "User lacks the permission to perform this action."}
	USER_ACCOUNT_DISABLED     = ErrorCode{4019, "User account has been disabled."}

	// Server-side errors (5xxx range)
	PROCESSING_ERROR                 = ErrorCode{5000, "Something went wrong when processing request."}
	UNCAUGHT_EXCEPTION               = ErrorCode{5001, "An uncaught exception occurred when processing request."}
	DATA_ACCESS_EXCEPTION            = ErrorCode{5002, "Data access exception occurred."}
	ENTITY_NOT_FOUND_EXCEPTION       = ErrorCode{5003, "Entity not found exception occurred."}
	FILE_UPLOAD_ERROR                = ErrorCode{5004, "Something went wrong when uploading the file."}
	FILE_DELETE_ERROR                = ErrorCode{5005, "Something went wrong when deleting the file."}
	UNSUPPORTED_MEDIA_TYPE_EXCEPTION = ErrorCode{5006, "Unsupported media type exception occurred."}

	// Add more error codes as needed
)
