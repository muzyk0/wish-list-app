package validation

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

// CustomValidator wraps the validator instance
type CustomValidator struct {
	validator *validator.Validate
}

// NewValidator creates a new validator instance
func NewValidator() *CustomValidator {
	return &CustomValidator{
		validator: validator.New(),
	}
}

// Validate validates a struct
func (cv *CustomValidator) Validate(i any) error {
	if err := cv.validator.Struct(i); err != nil {
		// Convert validation errors to user-friendly messages
		return fmt.Errorf("validation failed: %s", formatValidationErrors(err))
	}
	return nil
}

// formatValidationErrors formats validation errors into a readable string
func formatValidationErrors(err error) string {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		formattedErrors := make([]string, 0, len(validationErrors))
		for _, e := range validationErrors {
			formattedErrors = append(formattedErrors, formatFieldError(e))
		}
		return strings.Join(formattedErrors, "; ")
	}
	return err.Error()
}

// formatFieldError formats a single field error
func formatFieldError(e validator.FieldError) string {
	field := strings.ToLower(e.Field())

	switch e.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "min":
		return fmt.Sprintf("%s must be at least %s characters long", field, e.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s characters long", field, e.Param())
	case "url":
		return field + " must be a valid URL"
	case "uuid":
		return field + " must be a valid UUID"
	default:
		return fmt.Sprintf("%s failed validation on %s", field, e.Tag())
	}
}

// ValidationMiddleware creates a middleware that validates request bodies
func ValidationMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Validation is handled by the custom validator via c.Validate()
			return next(c)
		}
	}
}
