package helpers

import (
	"wish-list/internal/pkg/apperrors"

	"github.com/labstack/echo/v4"
)

// BindAndValidate binds request body to the provided struct and validates it.
// Returns *apperrors.AppError if binding or validation fails.
//
// Example usage in handler:
//
//	func (h *Handler) CreateItem(c echo.Context) error {
//	    var req dto.CreateItemRequest
//	    if err := helpers.BindAndValidate(c, &req); err != nil {
//	        return err  // AppError flows to centralized handler
//	    }
//	    // req is now bound and validated
//	}
func BindAndValidate(c echo.Context, req any) error {
	if err := c.Bind(req); err != nil {
		return apperrors.BadRequest("Invalid request body")
	}

	if err := c.Validate(req); err != nil {
		// Validator returns *apperrors.AppError with field details
		return err
	}

	return nil
}
