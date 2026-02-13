package helpers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// BindAndValidate binds request body to the provided struct and validates it.
// Returns error response if binding or validation fails.
//
// Example usage in handler:
//
//	func (h *Handler) CreateItem(c echo.Context) error {
//	    var req dto.CreateItemRequest
//	    if err := helpers.BindAndValidate(c, &req); err != nil {
//	        return err
//	    }
//	    // req is now bound and validated
//	}
func BindAndValidate(c echo.Context, req interface{}) error {
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return nil
}
