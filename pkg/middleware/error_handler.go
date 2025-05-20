// // FILE: pkg/middleware/error_handler.go
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	stderrors "errors"
	"fmt"
	"j-ticketing/pkg/errors"
	"j-ticketing/pkg/models"
	"log/slog"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// GlobalErrorHandler is a middleware that handles errors globally
func GlobalErrorHandler(logger *slog.Logger) fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		// Default to internal server error
		status := fiber.StatusInternalServerError
		response := &models.BaseErrorResponse{
			RespCode: errors.PROCESSING_ERROR.Code,
			RespDesc: errors.PROCESSING_ERROR.Message,
		}

		// Generate error code for 500 errors for tracking
		errorCode := generateErrorCode()
		logger.Error("Error occurred",
			"code", errorCode,
			"error", err)

		// Check for specific error types
		switch e := err.(type) {
		case *fiber.Error:
			// Handle Fiber's built-in errors
			status = e.Code
			response.RespDesc = e.Message

			// Map HTTP status codes to our error codes
			switch status {
			case fiber.StatusBadRequest:
				response.RespCode = errors.INVALID_INPUT_FORMAT.Code
			case fiber.StatusUnauthorized:
				response.RespCode = errors.USER_NOT_AUTHORIZED.Code
			case fiber.StatusForbidden:
				response.RespCode = errors.USER_NOT_PERMITTED.Code
			case fiber.StatusNotFound:
				response.RespCode = errors.FILE_NOT_FOUND.Code
			case fiber.StatusUnsupportedMediaType:
				response.RespCode = errors.UNSUPPORTED_MEDIA_TYPE_EXCEPTION.Code
			}

		case *errors.BadRequestError:
			// Handle BadRequestError
			status = fiber.StatusBadRequest
			response.RespCode = e.ErrorCode.Code
			response.RespDesc = e.ErrorCode.Message
			if e.Result != nil {
				response.Result = e.Result
			}

		case *errors.NotFoundError:
			// Handle NotFoundError
			status = fiber.StatusNotFound
			response.RespCode = e.ErrorCode.Code
			response.RespDesc = e.ErrorCode.Message
			if e.Result != nil {
				response.Result = e.Result
			}

		case *errors.ForbiddenError:
			// Handle ForbiddenError
			status = fiber.StatusForbidden
			response.RespCode = e.ErrorCode.Code
			response.RespDesc = e.ErrorCode.Message
			if e.Result != nil {
				response.Result = e.Result
			}

		case *errors.UnauthorizedError:
			// Handle UnauthorizedError
			status = fiber.StatusUnauthorized
			response.RespCode = e.ErrorCode.Code
			response.RespDesc = e.ErrorCode.Message
			if e.Result != nil {
				response.Result = e.Result
			}

		case *errors.InternalServerError:
			// Handle InternalServerError
			status = fiber.StatusInternalServerError
			response.RespCode = e.ErrorCode.Code
			response.RespDesc = e.ErrorCode.Message
			if e.Result != nil {
				response.Result = e.Result
			}
			// Add error code to response for tracking
			response.RespDesc = response.RespDesc + " - " + errorCode

		case *errors.ValidationError:
			// Handle ValidationError
			status = fiber.StatusBadRequest
			response.RespCode = e.ErrorCode.Code
			response.RespDesc = e.Error()
			if len(e.FieldErrors) > 0 {
				response.Result = e.FieldErrors
			} else if e.Result != nil {
				response.Result = e.Result
			}

		case validator.ValidationErrors:
			// Handle validator.ValidationErrors
			status = fiber.StatusBadRequest
			response.RespCode = errors.INVALID_INPUT_VALUES.Code

			// Get first validation error
			fieldErrors := make(map[string]string)
			for _, ve := range e {
				fieldName := strings.ToLower(ve.Field())
				fieldErrors[fieldName] = formatValidationError(ve)
			}

			if len(fieldErrors) > 0 {
				// Use only the first error message in the response
				for _, msg := range fieldErrors {
					response.RespDesc = msg
					break
				}
				response.Result = fieldErrors
			} else {
				response.RespDesc = errors.INVALID_INPUT_VALUES.Message
			}

		case error:
			// Check if it's a gorm.ErrRecordNotFound error
			if stderrors.Is(err, gorm.ErrRecordNotFound) {
				status = fiber.StatusNotFound
				response.RespCode = errors.ENTITY_NOT_FOUND_EXCEPTION.Code
				response.RespDesc = errors.ENTITY_NOT_FOUND_EXCEPTION.Message + " - " + errorCode
				break
			}

			// It's some other error, let it fall through to default handling

		default:
			// Handle unknown errors
			logger.Error("Unhandled error", "error", err)
			response.RespDesc = response.RespDesc + " - " + errorCode
		}

		// Return JSON response
		return c.Status(status).JSON(response)
	}
}

// ValidationErrorHandler is a middleware for handling validation errors
func ValidationErrorHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Continue to the next handler
		return c.Next()
	}
}

// formatValidationError formats validation error messages
func formatValidationError(err validator.FieldError) string {
	field := strings.ToLower(err.Field())

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("Field '%s' is required", field)
	case "email":
		return fmt.Sprintf("Field '%s' must be a valid email", field)
	case "min":
		return fmt.Sprintf("Field '%s' must be at least %s characters long", field, err.Param())
	case "max":
		return fmt.Sprintf("Field '%s' must be at most %s characters long", field, err.Param())
	default:
		return fmt.Sprintf("Field '%s' failed validation: %s", field, err.Tag())
	}
}

// generateErrorCode generates a unique error code
func generateErrorCode() string {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "unknown"
	}
	return hex.EncodeToString(bytes)
}
