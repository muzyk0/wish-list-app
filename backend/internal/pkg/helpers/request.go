package helpers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// BindAndValidate binds request body to the provided struct and validates it.
// Returns echo.HTTPError if binding or validation fails.
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
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return nil
}
