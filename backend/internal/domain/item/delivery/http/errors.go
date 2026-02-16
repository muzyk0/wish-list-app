package http

import (
	"errors"

	"wish-list/internal/domain/item/service"
	"wish-list/internal/pkg/apperrors"
)

// mapItemServiceError converts item service errors to AppErrors
func mapItemServiceError(err error) error {
	switch {
	case errors.Is(err, service.ErrItemNotFound):
		return apperrors.NotFound("Item not found")
	case errors.Is(err, service.ErrItemForbidden):
		return apperrors.Forbidden("Access denied")
	case errors.Is(err, service.ErrItemTitleRequired):
		return apperrors.BadRequest("Title is required")
	default:
		return apperrors.Internal("Failed to process request").Wrap(err)
	}
}
