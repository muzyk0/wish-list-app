package auth

import (
	"net/http"
	"time"
)

const (
	// RefreshTokenCookieName is the name of the refresh token cookie
	RefreshTokenCookieName = "refreshToken"

	// RefreshTokenMaxAge is the max age of the refresh token cookie in seconds (7 days)
	RefreshTokenMaxAge = 7 * 24 * 60 * 60
)

// NewRefreshTokenCookie creates a new refresh token cookie with secure settings.
// The cookie is httpOnly, secure, and has SameSite=None to work across domains.
//
// Example usage in handler:
//
//	func (h *Handler) Login(c echo.Context) error {
//	    refreshToken, err := h.tokenManager.GenerateRefreshToken(...)
//	    if err != nil {
//	        return err
//	    }
//	    c.SetCookie(auth.NewRefreshTokenCookie(refreshToken))
//	    // ...
//	}
func NewRefreshTokenCookie(value string) *http.Cookie {
	return &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   RefreshTokenMaxAge,
	}
}

// ClearRefreshTokenCookie creates a cookie that clears the refresh token.
// Used for logout or session invalidation.
//
// Example usage in handler:
//
//	func (h *Handler) Logout(c echo.Context) error {
//	    c.SetCookie(auth.ClearRefreshTokenCookie())
//	    return c.JSON(http.StatusOK, map[string]string{"message": "Logged out"})
//	}
func ClearRefreshTokenCookie() *http.Cookie {
	return &http.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
	}
}
