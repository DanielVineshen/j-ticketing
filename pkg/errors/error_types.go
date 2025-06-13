// File: j-ticketing/pkg/errors/zoo_api_client.go
package errors

// AppError is the base custom error type for application errors
type AppError struct {
	ErrorCode ErrorCode
	Result    interface{}
}

// Error implements the error interface
func (e *AppError) Error() string {
	return e.ErrorCode.Message
}

// GetErrorCode returns the error code
func (e *AppError) GetErrorCode() ErrorCode {
	return e.ErrorCode
}

// GetResult returns the result data
func (e *AppError) GetResult() interface{} {
	return e.Result
}

// NewAppError creates a new AppError
func NewAppError(errorCode ErrorCode, result interface{}) *AppError {
	return &AppError{
		ErrorCode: errorCode,
		Result:    result,
	}
}

// BadRequestError represents a bad request error
type BadRequestError struct {
	*AppError
}

// NewBadRequestError creates a new BadRequestError
func NewBadRequestError(errorCode ErrorCode) *BadRequestError {
	return &BadRequestError{
		AppError: NewAppError(errorCode, nil),
	}
}

// NewBadRequestErrorWithResult creates a new BadRequestError with result
func NewBadRequestErrorWithResult(errorCode ErrorCode, result interface{}) *BadRequestError {
	return &BadRequestError{
		AppError: NewAppError(errorCode, result),
	}
}

// NotFoundError represents a not found error
type NotFoundError struct {
	*AppError
}

// NewNotFoundError creates a new NotFoundError
func NewNotFoundError(errorCode ErrorCode) *NotFoundError {
	return &NotFoundError{
		AppError: NewAppError(errorCode, nil),
	}
}

// NewNotFoundErrorWithResult creates a new NotFoundError with result
func NewNotFoundErrorWithResult(errorCode ErrorCode, result interface{}) *NotFoundError {
	return &NotFoundError{
		AppError: NewAppError(errorCode, result),
	}
}

// ForbiddenError represents a forbidden error
type ForbiddenError struct {
	*AppError
}

// NewForbiddenError creates a new ForbiddenError
func NewForbiddenError(errorCode ErrorCode) *ForbiddenError {
	return &ForbiddenError{
		AppError: NewAppError(errorCode, nil),
	}
}

// NewForbiddenErrorWithResult creates a new ForbiddenError with result
func NewForbiddenErrorWithResult(errorCode ErrorCode, result interface{}) *ForbiddenError {
	return &ForbiddenError{
		AppError: NewAppError(errorCode, result),
	}
}

// UnauthorizedError represents an unauthorized error
type UnauthorizedError struct {
	*AppError
}

// NewUnauthorizedError creates a new UnauthorizedError
func NewUnauthorizedError(errorCode ErrorCode) *UnauthorizedError {
	return &UnauthorizedError{
		AppError: NewAppError(errorCode, nil),
	}
}

// NewUnauthorizedErrorWithResult creates a new UnauthorizedError with result
func NewUnauthorizedErrorWithResult(errorCode ErrorCode, result interface{}) *UnauthorizedError {
	return &UnauthorizedError{
		AppError: NewAppError(errorCode, result),
	}
}

// InternalServerError represents an internal server error
type InternalServerError struct {
	*AppError
}

// NewInternalServerError creates a new InternalServerError
func NewInternalServerError(errorCode ErrorCode) *InternalServerError {
	return &InternalServerError{
		AppError: NewAppError(errorCode, nil),
	}
}

// NewInternalServerErrorWithResult creates a new InternalServerError with result
func NewInternalServerErrorWithResult(errorCode ErrorCode, result interface{}) *InternalServerError {
	return &InternalServerError{
		AppError: NewAppError(errorCode, result),
	}
}

// ValidationError represents validation errors
type ValidationError struct {
	*AppError
	FieldErrors map[string]string
}

// NewValidationError creates a new ValidationError
func NewValidationError(message string) *ValidationError {
	return &ValidationError{
		AppError:    NewAppError(INVALID_INPUT_VALUES, message),
		FieldErrors: make(map[string]string),
	}
}

// AddFieldError adds a field error
func (e *ValidationError) AddFieldError(field, message string) {
	e.FieldErrors[field] = message
}

// Error overrides the AppError.Error()
func (e *ValidationError) Error() string {
	if len(e.FieldErrors) > 0 {
		for _, msg := range e.FieldErrors {
			return msg // Return first error message
		}
	}
	return e.AppError.Error()
}
