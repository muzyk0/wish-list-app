package http

import (
	"errors"

	userservice "wish-list/internal/domain/user/service"
	"wish-list/internal/pkg/apperrors"
)

// mapUserServiceError converts user service errors to AppErrors
func mapUserServiceError(err error) error {
	switch {
	case errors.Is(err, userservice.ErrUserAlreadyExists):
		return apperrors.Conflict("User with this email already exists")
	case errors.Is(err, userservice.ErrUserNotFound):
		return apperrors.NotFound("User not found")
	case errors.Is(err, userservice.ErrInvalidPassword):
		return apperrors.Unauthorized("Current password is incorrect")
	case errors.Is(err, userservice.ErrInvalidCredentials):
		return apperrors.Unauthorized("Invalid credentials")
	default:
		return apperrors.Internal("Failed to process request").Wrap(err)
	}
}
