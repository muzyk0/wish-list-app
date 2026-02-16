package http

import (
	"errors"

	userservice "wish-list/internal/domain/user/service"
	"wish-list/internal/pkg/apperrors"
)

// mapAuthServiceError converts auth-related service errors to AppErrors
func mapAuthServiceError(err error) error {
	switch {
	case errors.Is(err, userservice.ErrUserNotFound):
		return apperrors.Unauthorized("User not found")
	case errors.Is(err, userservice.ErrInvalidPassword):
		return apperrors.Unauthorized("Current password is incorrect")
	case errors.Is(err, userservice.ErrUserAlreadyExists):
		return apperrors.Conflict("Email already in use")
	default:
		return apperrors.Internal("Failed to process request").Wrap(err)
	}
}
