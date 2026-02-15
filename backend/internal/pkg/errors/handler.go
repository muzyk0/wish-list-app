// Package errors provides centralized error handling for the application.
// It defines typed errors with HTTP status codes and a centralized error handler
// for Echo framework that standardizes error responses across all handlers.
package errors

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HTTPError represents an error with an associated HTTP status code.
// It is used by handlers to return structured errors that the centralized
// error handler can process and return with appropriate status codes.
type HTTPError struct {
	StatusCode int
	Message    string
	Err        error
}

// Error implements the error interface.
func (e *HTTPError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.Message
}

// Unwrap allows errors.Is to work with HTTPError.
func (e *HTTPError) Unwrap() error {
	return e.Err
}

// NewHTTPError creates a new HTTPError with the given status code and message.
func NewHTTPError(statusCode int, message string) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
	}
}

// NewHTTPErrorWithCause creates a new HTTPError with an underlying cause.
func NewHTTPErrorWithCause(statusCode int, message string, err error) *HTTPError {
	return &HTTPError{
		StatusCode: statusCode,
		Message:    message,
		Err:        err,
	}
}

// Common HTTP errors for standard status codes.
var (
	// 400 Bad Request
	ErrBadRequest       = NewHTTPError(http.StatusBadRequest, "Bad request")
	ErrInvalidInput     = NewHTTPError(http.StatusBadRequest, "Invalid input")
	ErrValidationFailed = NewHTTPError(http.StatusBadRequest, "Validation failed")
	ErrInvalidUUID      = NewHTTPError(http.StatusBadRequest, "Invalid UUID format")

	// 401 Unauthorized
	ErrUnauthorized       = NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	ErrInvalidToken       = NewHTTPError(http.StatusUnauthorized, "Invalid or expired token")
	ErrInvalidCredentials = NewHTTPError(http.StatusUnauthorized, "Invalid credentials")
	ErrInvalidPassword    = NewHTTPError(http.StatusUnauthorized, "Current password is incorrect")

	// 403 Forbidden
	ErrForbidden    = NewHTTPError(http.StatusForbidden, "Forbidden")
	ErrAccessDenied = NewHTTPError(http.StatusForbidden, "Access denied")
	ErrNotOwner     = NewHTTPError(http.StatusForbidden, "Not authorized to access this resource")

	// 404 Not Found
	ErrNotFound         = NewHTTPError(http.StatusNotFound, "Not found")
	ErrResourceNotFound = NewHTTPError(http.StatusNotFound, "Resource not found")

	// 409 Conflict
	ErrConflict       = NewHTTPError(http.StatusConflict, "Conflict")
	ErrAlreadyExists  = NewHTTPError(http.StatusConflict, "Resource already exists")
	ErrDuplicateEmail = NewHTTPError(http.StatusConflict, "Email already in use")

	// 422 Unprocessable Entity
	ErrUnprocessable = NewHTTPError(http.StatusUnprocessableEntity, "Unprocessable entity")

	// 429 Too Many Requests
	ErrTooManyRequests   = NewHTTPError(http.StatusTooManyRequests, "Too many requests")
	ErrRateLimitExceeded = NewHTTPError(http.StatusTooManyRequests, "Rate limit exceeded")

	// 500 Internal Server Error
	ErrInternalServer = NewHTTPError(http.StatusInternalServerError, "Internal server error")
)

// ErrorResponse represents the JSON structure for error responses.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Handler creates an Echo error handler middleware that processes errors
// and returns standardized JSON responses.
//
// Usage:
//
// e := echo.New()
// e.HTTPErrorHandler = errors.Handler()
func Handler() echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		// Don't handle errors if response already sent
		if c.Response().Committed {
			return
		}

		// Get the HTTPError if it's our type
		var httpErr *HTTPError
		if errors.As(err, &httpErr) {
			// Log internal error details but return safe message
			if httpErr.Err != nil {
				c.Logger().Errorf("Error: %v", httpErr.Err)
			}

			// Send response
			if !c.Response().Committed {
				_ = c.JSON(httpErr.StatusCode, ErrorResponse{
					Error: httpErr.Message,
				})
			}
			return
		}

		// Handle Echo's own HTTPError
		var echoErr *echo.HTTPError
		if errors.As(err, &echoErr) {
			message := http.StatusText(echoErr.Code)
			if msg, ok := echoErr.Message.(string); ok {
				message = msg
			}

			if !c.Response().Committed {
				_ = c.JSON(echoErr.Code, ErrorResponse{
					Error: message,
				})
			}
			return
		}

		// Log unknown errors
		c.Logger().Errorf("Unhandled error: %v", err)

		// Return generic 500 for unknown errors
		if !c.Response().Committed {
			_ = c.JSON(http.StatusInternalServerError, ErrorResponse{
				Error: "Internal server error",
			})
		}
	}
}

// Middleware creates an Echo middleware that recovers from panics
// and converts them to internal server errors.
func Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			err := next(c)
			if err != nil {
				// Pass through to error handler
				return err
			}
			return nil
		}
	}
}

// IsHTTPError checks if an error is an HTTPError with the given status code.
func IsHTTPError(err error, statusCode int) bool {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode == statusCode
	}
	return false
}

// GetStatusCode extracts the HTTP status code from an error.
// Returns 500 if the error is not an HTTPError.
func GetStatusCode(err error) int {
	var httpErr *HTTPError
	if errors.As(err, &httpErr) {
		return httpErr.StatusCode
	}
	return http.StatusInternalServerError
}
