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
	default:
		return apperrors.Internal("Failed to process request").Wrap(err)
	}
}
