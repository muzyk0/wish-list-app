package apperrors

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConstructors(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) *AppError
		message  string
		wantCode int
	}{
		{"BadRequest", BadRequest, "bad input", http.StatusBadRequest},
		{"Unauthorized", Unauthorized, "no token", http.StatusUnauthorized},
		{"Forbidden", Forbidden, "access denied", http.StatusForbidden},
		{"NotFound", NotFound, "not found", http.StatusNotFound},
		{"Conflict", Conflict, "duplicate", http.StatusConflict},
		{"TooManyRequests", TooManyRequests, "slow down", http.StatusTooManyRequests},
		{"Internal", Internal, "oops", http.StatusInternalServerError},
		{"BadGateway", BadGateway, "upstream", http.StatusBadGateway},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn(tt.message)
			assert.Equal(t, tt.wantCode, err.Code)
			assert.Equal(t, tt.message, err.Message)
			require.NoError(t, err.Err)
			assert.Nil(t, err.Details)
		})
	}
}

func TestNew(t *testing.T) {
	err := New(http.StatusTeapot, "I'm a teapot")
	assert.Equal(t, 418, err.Code)
	assert.Equal(t, "I'm a teapot", err.Message)
}

func TestError(t *testing.T) {
	t.Run("without cause", func(t *testing.T) {
		err := NotFound("item not found")
		assert.Equal(t, "item not found", err.Error())
	})

	t.Run("with cause", func(t *testing.T) {
		cause := errors.New("sql: no rows")
		err := NotFound("item not found").Wrap(cause)
		assert.Equal(t, "item not found: sql: no rows", err.Error())
	})
}

func TestWrap(t *testing.T) {
	cause := errors.New("db connection failed")
	original := Internal("Something went wrong")
	wrapped := original.Wrap(cause)

	// Wrapped is a new instance
	assert.NotSame(t, original, wrapped)

	// Original is not mutated
	require.NoError(t, original.Err)

	// Wrapped carries the cause
	assert.Equal(t, cause, wrapped.Err)
	assert.Equal(t, original.Code, wrapped.Code)
	assert.Equal(t, original.Message, wrapped.Message)
}

func TestUnwrap(t *testing.T) {
	cause := errors.New("root cause")
	err := Internal("failed").Wrap(cause)

	// errors.Is works through Unwrap
	assert.ErrorIs(t, err, cause)
}

func TestUnwrapChain(t *testing.T) {
	sentinel := errors.New("ErrUserNotFound")
	appErr := NotFound("User not found").Wrap(sentinel)

	require.ErrorIs(t, appErr, sentinel)

	var target *AppError
	require.ErrorAs(t, appErr, &target)
	assert.Equal(t, http.StatusNotFound, target.Code)
}

func TestWithMessage(t *testing.T) {
	original := NotFound("generic not found")
	custom := original.WithMessage("Wishlist not found")

	// New instance
	assert.NotSame(t, original, custom)

	// Original unchanged
	assert.Equal(t, "generic not found", original.Message)

	// Custom has new message, same code
	assert.Equal(t, "Wishlist not found", custom.Message)
	assert.Equal(t, http.StatusNotFound, custom.Code)
}

func TestNewValidationError(t *testing.T) {
	details := map[string]string{
		"email":    "must be a valid email address",
		"password": "must be at least 8 characters long",
	}

	err := NewValidationError(details)

	assert.Equal(t, http.StatusBadRequest, err.Code)
	assert.Equal(t, "Validation failed", err.Message)
	assert.Equal(t, details, err.Details)
	assert.Contains(t, err.Details, "email")
	assert.Contains(t, err.Details, "password")
}

func TestImmutability(t *testing.T) {
	// Calling constructor twice returns different instances
	err1 := NotFound("a")
	err2 := NotFound("b")

	assert.NotSame(t, err1, err2)
	assert.NotEqual(t, err1.Message, err2.Message)
}
