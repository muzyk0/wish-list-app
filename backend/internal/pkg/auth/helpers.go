package auth

import "github.com/labstack/echo/v4"

// MustGetUserID extracts user ID from context in protected routes.
// ONLY use in handlers where JWTMiddleware is already applied via routes.
// Returns empty string if user not in context (should never happen in protected routes).
//
// Example usage in handler:
//
//	func (h *Handler) CreateItem(c echo.Context) error {
//	    userID := auth.MustGetUserID(c)  // No error check needed - middleware guarantees it
//	    // ... rest of handler logic
//	}
func MustGetUserID(c echo.Context) string {
	userID, _, _, _ := GetUserFromContext(c)
	return userID
}

// MustGetUserInfo extracts all user info from context in protected routes.
// ONLY use in handlers where JWTMiddleware is already applied via routes.
// Returns empty strings if user not in context (should never happen in protected routes).
//
// Example usage in handler:
//
//	func (h *Handler) GetProfile(c echo.Context) error {
//	    userID, email, userType := auth.MustGetUserInfo(c)
//	    // ... rest of handler logic
//	}
func MustGetUserInfo(c echo.Context) (userID, email, userType string) {
	userID, email, userType, _ = GetUserFromContext(c)
	return
}
