// File: j-ticketing/pkg/validation/validation.go
package validation

import (
	"j-ticketing/pkg/errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

// Initialize the validator
func init() {
	validate = validator.New()

	// Register custom validation tag names to use json property names in error messages
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		return name
	})
}

// ValidateStruct validates a struct based on validation tags
func ValidateStruct(s interface{}) error {
	err := validate.Struct(s)
	if err != nil {
		validationError := errors.NewValidationError("Validation failed")

		// Use validator's error details
		for _, err := range err.(validator.ValidationErrors) {
			field := err.Field()
			tag := err.Tag()

			var message string
			switch tag {
			case "required":
				message = field + " is required"
			case "email":
				message = field + " must be a valid email address"
			case "min":
				message = field + " must be at least " + err.Param() + " characters long"
			case "max":
				message = field + " must be at most " + err.Param() + " characters long"
			case "oneof":
				message = field + " must be one of [" + err.Param() + "]"
			default:
				message = field + " validation failed on tag " + tag
			}

			validationError.AddFieldError(field, message)
		}

		return validationError
	}

	return nil
}
