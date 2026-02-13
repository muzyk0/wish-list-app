package helpers

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

// PaginationParams holds parsed pagination parameters
type PaginationParams struct {
	Page   int
	Limit  int
	Offset int
}

// ParsePagination extracts and validates pagination parameters from query string.
// Defaults: page=1, limit=10
// Constraints: page >= 1, 1 <= limit <= 100
//
// Example usage in handler:
//
//	func (h *Handler) GetItems(c echo.Context) error {
//	    pagination := helpers.ParsePagination(c)
//	    items, err := h.service.GetItems(ctx, pagination.Limit, pagination.Offset)
//	    // ...
//	}
func ParsePagination(c echo.Context) PaginationParams {
	page := 1
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	limit := 10
	if limitStr := c.QueryParam("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 100 {
			limit = parsedLimit
		}
	}

	offset := (page - 1) * limit

	return PaginationParams{
		Page:   page,
		Limit:  limit,
		Offset: offset,
	}
}
