package http

import (
	"errors"

	"wish-list/internal/domain/wishlist_item/service"
	"wish-list/internal/pkg/apperrors"
)

// mapWishlistItemServiceError converts wishlist_item service errors to AppErrors
func mapWishlistItemServiceError(err error) error {
	switch {
	case errors.Is(err, service.ErrWishListNotFound):
		return apperrors.NotFound("Wishlist not found")
	case errors.Is(err, service.ErrWishListForbidden):
		return apperrors.Forbidden("Access denied")
	case errors.Is(err, service.ErrItemNotFound):
		return apperrors.NotFound("Item not found")
	case errors.Is(err, service.ErrItemForbidden):
		return apperrors.Forbidden("Access denied to item")
	case errors.Is(err, service.ErrItemAlreadyAttached):
		return apperrors.Conflict("Item already attached to this wishlist")
	case errors.Is(err, service.ErrItemNotInWishlist):
		return apperrors.NotFound("Item not found in this wishlist")
	case errors.Is(err, service.ErrInvalidWishlistItemWLID):
		return apperrors.BadRequest("Invalid wishlist ID")
	case errors.Is(err, service.ErrInvalidWishlistItemID):
		return apperrors.BadRequest("Invalid item ID")
	case errors.Is(err, service.ErrInvalidWishlistItemUser):
		return apperrors.BadRequest("Invalid user ID")
	case errors.Is(err, service.ErrWishlistItemTitleRequired):
		return apperrors.BadRequest("Title is required")
	case errors.Is(err, service.ErrManualReservedNameEmpty):
		return apperrors.BadRequest("reserved_by_name is required")
	case errors.Is(err, service.ErrItemNotAvailable):
		return apperrors.Conflict("Item is already reserved or purchased")
	default:
		return apperrors.Internal("Failed to process request").Wrap(err)
	}
}
