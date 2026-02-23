package http

import (
	"errors"

	"wish-list/internal/domain/reservation/service"
	"wish-list/internal/pkg/apperrors"
)

// mapReservationServiceError converts reservation service errors to AppErrors
func mapReservationServiceError(err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidGiftItemID):
		return apperrors.BadRequest("Invalid gift item ID")
	case errors.Is(err, service.ErrInvalidReservationWishlist):
		return apperrors.BadRequest("Invalid wishlist ID")
	case errors.Is(err, service.ErrGiftItemNotInWishlist):
		return apperrors.NotFound("Gift item not found in wishlist")
	case errors.Is(err, service.ErrGiftItemNotInPublicWishlist):
		return apperrors.NotFound("Gift item not found in public wishlist")
	case errors.Is(err, service.ErrGiftItemAlreadyReserved):
		return apperrors.Conflict("Gift item is already reserved")
	case errors.Is(err, service.ErrGuestInfoRequired):
		return apperrors.BadRequest("Guest name is required")
	case errors.Is(err, service.ErrReservationNotFound):
		return apperrors.NotFound("Reservation not found")
	case errors.Is(err, service.ErrMissingUserOrToken):
		return apperrors.BadRequest("Either user ID or reservation token must be provided")
	default:
		return apperrors.Internal("Failed to process request").Wrap(err)
	}
}
