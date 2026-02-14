package response

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// Success sends a successful JSON response
func Success(c echo.Context, status int, data any) error {
	return c.JSON(status, data)
}

// Error sends an error JSON response
func Error(c echo.Context, status int, message string) error {
	return c.JSON(status, map[string]string{"error": message})
}

// ValidationError sends a validation error JSON response
func ValidationError(c echo.Context, errors map[string]string) error {
	return c.JSON(http.StatusBadRequest, map[string]any{
		"error":   "validation failed",
		"details": errors,
	})
}
