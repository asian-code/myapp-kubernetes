package validation

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

// Validate validates a struct using struct tags
func Validate(s interface{}) error {
	if err := validate.Struct(s); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			return formatValidationErrors(validationErrors)
		}
		return err
	}
	return nil
}

// formatValidationErrors converts validator errors to user-friendly messages
func formatValidationErrors(errs validator.ValidationErrors) error {
	var messages []string
	for _, err := range errs {
		messages = append(messages, formatFieldError(err))
	}
	return fmt.Errorf("validation failed: %s", strings.Join(messages, "; "))
}

// formatFieldError formats a single field error
func formatFieldError(err validator.FieldError) string {
	field := err.Field()

	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, err.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of [%s]", field, err.Param())
	default:
		return fmt.Sprintf("%s failed validation (%s)", field, err.Tag())
	}
}

// MustValidate validates a struct and panics if validation fails
// Useful for configuration validation at startup
func MustValidate(s interface{}) {
	if err := Validate(s); err != nil {
		panic(fmt.Sprintf("Configuration validation failed: %v", err))
	}
}
