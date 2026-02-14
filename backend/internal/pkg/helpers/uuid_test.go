package helpers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestParseUUID(t *testing.T) {
	tests := []struct {
		name               string
		uuidStr            string
		expectedValid      bool
		expectedStatusCode int
		expectedError      string
	}{
		{
			name:          "valid UUID v4",
			uuidStr:       "550e8400-e29b-41d4-a716-446655440000",
			expectedValid: true,
		},
		{
			name:          "valid UUID from google/uuid",
			uuidStr:       uuid.New().String(),
			expectedValid: true,
		},
		{
			name:               "invalid UUID - too short",
			uuidStr:            "550e8400",
			expectedValid:      false,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Invalid UUID format",
		},
		{
			name:               "invalid UUID - wrong format",
			uuidStr:            "not-a-uuid",
			expectedValid:      false,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Invalid UUID format",
		},
		{
			name:               "invalid UUID - empty string",
			uuidStr:            "",
			expectedValid:      false,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Invalid UUID format",
		},
		{
			name:          "valid UUID without hyphens",
			uuidStr:       "550e8400e29b41d4a716446655440000",
			expectedValid: true,
		},
		{
			name:               "invalid UUID - wrong characters",
			uuidStr:            "550e8400-e29b-41d4-a716-44665544000g",
			expectedValid:      false,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Invalid UUID format",
		},
		{
			name:               "invalid UUID - spaces",
			uuidStr:            "550e8400-e29b-41d4-a716-446655440000 ",
			expectedValid:      false,
			expectedStatusCode: http.StatusBadRequest,
			expectedError:      "Invalid UUID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Echo context
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Parse UUID
			result, err := ParseUUID(c, tt.uuidStr)

			if tt.expectedValid {
				assert.Nil(t, err, "Expected no error for valid UUID")
				assert.True(t, result.Valid, "UUID should be valid")
				assert.NotEqual(t, [16]byte{}, result.Bytes, "UUID bytes should not be zero")
			} else {
				assert.NotNil(t, err, "Expected non-nil error for invalid UUID")
				var httpErr *echo.HTTPError
				assert.True(t, errors.As(err, &httpErr), "Error should be echo.HTTPError")
				assert.Equal(t, tt.expectedStatusCode, httpErr.Code, "Status code mismatch")
			}
		})
	}
}

func TestMustParseUUID(t *testing.T) {
	tests := []struct {
		name          string
		uuidStr       string
		expectedValid bool
	}{
		{
			name:          "valid UUID v4",
			uuidStr:       "550e8400-e29b-41d4-a716-446655440000",
			expectedValid: true,
		},
		{
			name:          "valid UUID from google/uuid",
			uuidStr:       uuid.New().String(),
			expectedValid: true,
		},
		{
			name:          "invalid UUID returns invalid (not error)",
			uuidStr:       "not-a-uuid",
			expectedValid: false,
		},
		{
			name:          "empty string returns invalid",
			uuidStr:       "",
			expectedValid: false,
		},
		{
			name:          "malformed UUID returns invalid",
			uuidStr:       "550e8400-wrong",
			expectedValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MustParseUUID(tt.uuidStr)

			assert.Equal(t, tt.expectedValid, result.Valid, "Valid flag mismatch")
			if tt.expectedValid {
				assert.NotEqual(t, [16]byte{}, result.Bytes, "Valid UUID should have non-zero bytes")
			}
		})
	}
}

func TestMustParseUUIDNoPanic(t *testing.T) {
	t.Run("should not panic on any input", func(t *testing.T) {
		testInputs := []string{
			"",
			"invalid",
			"550e8400-e29b-41d4-a716-446655440000",
			strings.Repeat("a", 1000),
			"null",
			"undefined",
		}

		for _, input := range testInputs {
			assert.NotPanics(t, func() {
				_ = MustParseUUID(input)
			}, "MustParseUUID should never panic for input: %s", input)
		}
	})
}

func TestParseUUIDConsistency(t *testing.T) {
	t.Run("ParseUUID and MustParseUUID should agree on valid UUIDs", func(t *testing.T) {
		validUUID := "550e8400-e29b-41d4-a716-446655440000"

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		result1, err := ParseUUID(c, validUUID)
		result2 := MustParseUUID(validUUID)

		assert.Nil(t, err, "ParseUUID should return nil error for valid UUID")
		assert.True(t, result1.Valid, "ParseUUID result should be valid")
		assert.True(t, result2.Valid, "MustParseUUID result should be valid")
		assert.Equal(t, result1.Bytes, result2.Bytes, "Both functions should return same UUID bytes")
	})

	t.Run("ParseUUID and MustParseUUID should both handle invalid UUIDs", func(t *testing.T) {
		invalidUUID := "not-a-uuid"

		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		result1, err := ParseUUID(c, invalidUUID)
		result2 := MustParseUUID(invalidUUID)

		// ParseUUID returns echo.HTTPError for invalid UUID
		assert.NotNil(t, err, "ParseUUID returns non-nil error for invalid UUID")
		var httpErr *echo.HTTPError
		assert.True(t, errors.As(err, &httpErr), "Error should be echo.HTTPError")
		assert.Equal(t, http.StatusBadRequest, httpErr.Code, "HTTP error should be 400")

		// Both should have invalid UUID
		assert.False(t, result1.Valid, "ParseUUID result should be invalid")
		assert.False(t, result2.Valid, "MustParseUUID result should be invalid")
	})
}
