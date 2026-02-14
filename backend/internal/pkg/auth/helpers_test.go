package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestMustGetUserID(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(c echo.Context)
		expectedUserID string
	}{
		{
			name: "valid user in context",
			setupContext: func(c echo.Context) {
				c.Set("user_id", "user-123")
				c.Set("email", "test@example.com")
				c.Set("user_type", "user")
			},
			expectedUserID: "user-123",
		},
		{
			name: "empty context returns empty string",
			setupContext: func(c echo.Context) {
				// No user data set
			},
			expectedUserID: "",
		},
		{
			name: "partial context (only userID)",
			setupContext: func(c echo.Context) {
				c.Set("user_id", "user-456")
			},
			expectedUserID: "user-456",
		},
		{
			name: "context with nil userID",
			setupContext: func(c echo.Context) {
				c.Set("user_id", nil)
			},
			expectedUserID: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Echo context
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Setup context
			tt.setupContext(c)

			// Get user ID
			result := MustGetUserID(c)

			assert.Equal(t, tt.expectedUserID, result, "User ID mismatch")
		})
	}
}

func TestMustGetUserInfo(t *testing.T) {
	tests := []struct {
		name             string
		setupContext     func(c echo.Context)
		expectedUserID   string
		expectedEmail    string
		expectedUserType string
	}{
		{
			name: "valid user info in context",
			setupContext: func(c echo.Context) {
				c.Set("user_id", "user-123")
				c.Set("email", "test@example.com")
				c.Set("user_type", "admin")
			},
			expectedUserID:   "user-123",
			expectedEmail:    "test@example.com",
			expectedUserType: "admin",
		},
		{
			name: "empty context returns empty strings",
			setupContext: func(c echo.Context) {
				// No user data set
			},
			expectedUserID:   "",
			expectedEmail:    "",
			expectedUserType: "",
		},
		{
			name: "partial context (only userID and email, defaults to user type)",
			setupContext: func(c echo.Context) {
				c.Set("user_id", "user-456")
				c.Set("email", "user@example.com")
				// user_type not set, should default to "user"
			},
			expectedUserID:   "user-456",
			expectedEmail:    "user@example.com",
			expectedUserType: "user", // Defaults to "user" when not set
		},
		{
			name: "regular user type",
			setupContext: func(c echo.Context) {
				c.Set("user_id", "user-789")
				c.Set("email", "regular@example.com")
				c.Set("user_type", "user")
			},
			expectedUserID:   "user-789",
			expectedEmail:    "regular@example.com",
			expectedUserType: "user",
		},
		{
			name: "guest user type",
			setupContext: func(c echo.Context) {
				c.Set("user_id", "guest-123")
				c.Set("email", "guest@example.com")
				c.Set("user_type", "guest")
			},
			expectedUserID:   "guest-123",
			expectedEmail:    "guest@example.com",
			expectedUserType: "guest",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Echo context
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			// Setup context
			tt.setupContext(c)

			// Get user info
			userID, email, userType := MustGetUserInfo(c)

			assert.Equal(t, tt.expectedUserID, userID, "User ID mismatch")
			assert.Equal(t, tt.expectedEmail, email, "Email mismatch")
			assert.Equal(t, tt.expectedUserType, userType, "User type mismatch")
		})
	}
}

func TestMustGetUserIDNoPanic(t *testing.T) {
	t.Run("should not panic with empty context", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		assert.NotPanics(t, func() {
			_ = MustGetUserID(c)
		}, "MustGetUserID should not panic")
	})

	t.Run("should not panic with nil context values", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set("user_id", nil)
		c.Set("email", nil)
		c.Set("user_type", nil)

		assert.NotPanics(t, func() {
			_ = MustGetUserID(c)
		}, "MustGetUserID should not panic with nil values")
	})
}

func TestMustGetUserInfoNoPanic(t *testing.T) {
	t.Run("should not panic with empty context", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		assert.NotPanics(t, func() {
			_, _, _ = MustGetUserInfo(c)
		}, "MustGetUserInfo should not panic")
	})

	t.Run("should not panic with nil context values", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set("user_id", nil)
		c.Set("email", nil)
		c.Set("user_type", nil)

		assert.NotPanics(t, func() {
			_, _, _ = MustGetUserInfo(c)
		}, "MustGetUserInfo should not panic with nil values")
	})
}

func TestMustHelpersConsistency(t *testing.T) {
	t.Run("MustGetUserID and MustGetUserInfo should return consistent userID", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		c.Set("user_id", "user-consistent-123")
		c.Set("email", "consistent@example.com")
		c.Set("user_type", "user")

		userIDFromSingle := MustGetUserID(c)
		userIDFromMulti, email, userType := MustGetUserInfo(c)

		assert.Equal(t, userIDFromSingle, userIDFromMulti, "Both functions should return same userID")
		assert.Equal(t, "consistent@example.com", email)
		assert.Equal(t, "user", userType)
	})
}
