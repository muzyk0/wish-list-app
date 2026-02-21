package http

import (
	"errors"

	"wish-list/internal/domain/wishlist/service"
	"wish-list/internal/pkg/apperrors"
)

// mapWishlistServiceError converts wishlist service errors to AppErrors
func mapWishlistServiceError(err error) error {
	switch {
	case errors.Is(err, service.ErrWishListNotFound):
		return apperrors.NotFound("Wish list not found")
	case errors.Is(err, service.ErrWishListForbidden):
		return apperrors.Forbidden("Access denied")
	case errors.Is(err, service.ErrWishListTitleRequired):
		return apperrors.BadRequest("Title is required")
	case errors.Is(err, service.ErrSlugTaken):
		return apperrors.Conflict("This URL slug is already taken. Please choose a different one.")
	case errors.Is(err, service.ErrSlugInvalid):
		return apperrors.BadRequest("Slug must contain only lowercase letters, digits, and hyphens (e.g. my-birthday-2026)")
	default:
		return apperrors.Internal("Failed to process request").Wrap(err)
	}
}
