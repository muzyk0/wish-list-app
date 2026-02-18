package auth

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewRefreshTokenCookie(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "with valid token",
			value: "valid-refresh-token-123",
		},
		{
			name:  "with empty token",
			value: "",
		},
		{
			name:  "with long token",
			value: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cookie := NewRefreshTokenCookie(tt.value)

			// Check cookie properties
			assert.Equal(t, RefreshTokenCookieName, cookie.Name, "Cookie name mismatch")
			assert.Equal(t, tt.value, cookie.Value, "Cookie value mismatch")
			assert.Equal(t, "/", cookie.Path, "Cookie path should be /")
			assert.True(t, cookie.HttpOnly, "Cookie should be HttpOnly")
			assert.True(t, cookie.Secure, "Cookie should be Secure")
			assert.Equal(t, http.SameSiteNoneMode, cookie.SameSite, "Cookie should have SameSite=None")
			assert.Equal(t, RefreshTokenMaxAge, cookie.MaxAge, "Cookie MaxAge mismatch")
			assert.Equal(t, 7*24*60*60, cookie.MaxAge, "MaxAge should be 7 days in seconds")
		})
	}
}

func TestClearRefreshTokenCookie(t *testing.T) {
	t.Run("should create cookie that clears refresh token", func(t *testing.T) {
		cookie := ClearRefreshTokenCookie()

		// Check cookie properties for clearing
		assert.Equal(t, RefreshTokenCookieName, cookie.Name, "Cookie name mismatch")
		assert.Empty(t, cookie.Value, "Cookie value should be empty")
		assert.Equal(t, "/", cookie.Path, "Cookie path should be /")
		assert.True(t, cookie.HttpOnly, "Cookie should be HttpOnly")
		assert.True(t, cookie.Secure, "Cookie should be Secure")
		assert.Equal(t, http.SameSiteNoneMode, cookie.SameSite, "Cookie should have SameSite=None")
		assert.Equal(t, -1, cookie.MaxAge, "MaxAge should be -1 to delete cookie")
		assert.Equal(t, time.Unix(0, 0), cookie.Expires, "Expires should be epoch time")
	})
}

func TestRefreshTokenCookieConstants(t *testing.T) {
	t.Run("verify cookie constants", func(t *testing.T) {
		assert.Equal(t, "refreshToken", RefreshTokenCookieName)
		assert.Equal(t, 604800, RefreshTokenMaxAge, "MaxAge should be 7 days = 604800 seconds")
	})
}

func TestCookieSecuritySettings(t *testing.T) {
	t.Run("refresh token cookie should have proper security settings", func(t *testing.T) {
		cookie := NewRefreshTokenCookie("test-token")

		// Security checks
		assert.True(t, cookie.HttpOnly, "HttpOnly prevents XSS attacks")
		assert.True(t, cookie.Secure, "Secure ensures HTTPS-only transmission")
		assert.Equal(t, http.SameSiteNoneMode, cookie.SameSite, "SameSite=None allows cross-domain usage")
		assert.Equal(t, "/", cookie.Path, "Path=/ makes cookie available site-wide")
	})

	t.Run("clear cookie should maintain security settings", func(t *testing.T) {
		cookie := ClearRefreshTokenCookie()

		// Security checks (should match NewRefreshTokenCookie)
		assert.True(t, cookie.HttpOnly, "HttpOnly should be maintained")
		assert.True(t, cookie.Secure, "Secure should be maintained")
		assert.Equal(t, http.SameSiteNoneMode, cookie.SameSite, "SameSite should be maintained")
	})
}

func TestCookieExpiration(t *testing.T) {
	t.Run("refresh token cookie should expire in 7 days", func(t *testing.T) {
		cookie := NewRefreshTokenCookie("test-token")

		expectedSeconds := 7 * 24 * 60 * 60
		assert.Equal(t, expectedSeconds, cookie.MaxAge, "Cookie should expire in 7 days")
	})

	t.Run("clear cookie should expire immediately", func(t *testing.T) {
		cookie := ClearRefreshTokenCookie()

		assert.Equal(t, -1, cookie.MaxAge, "MaxAge=-1 deletes cookie immediately")
		assert.True(t, cookie.Expires.Before(time.Now()), "Expires should be in the past")
		assert.Equal(t, time.Unix(0, 0), cookie.Expires, "Expires should be epoch time")
	})
}

func TestCookieConsistency(t *testing.T) {
	t.Run("all cookie properties except Value and MaxAge should match", func(t *testing.T) {
		refreshCookie := NewRefreshTokenCookie("token")
		clearCookie := ClearRefreshTokenCookie()

		assert.Equal(t, refreshCookie.Name, clearCookie.Name, "Name should match")
		assert.Equal(t, refreshCookie.Path, clearCookie.Path, "Path should match")
		assert.Equal(t, refreshCookie.HttpOnly, clearCookie.HttpOnly, "HttpOnly should match")
		assert.Equal(t, refreshCookie.Secure, clearCookie.Secure, "Secure should match")
		assert.Equal(t, refreshCookie.SameSite, clearCookie.SameSite, "SameSite should match")

		// Value and MaxAge should differ
		assert.NotEqual(t, refreshCookie.Value, clearCookie.Value, "Value should differ")
		assert.NotEqual(t, refreshCookie.MaxAge, clearCookie.MaxAge, "MaxAge should differ")
	})
}
