package http

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers auth domain HTTP routes on the /api/auth group.
// It accepts both the auth Handler and the OAuthHandler, plus auth middleware for protected endpoints.
func RegisterRoutes(e *echo.Echo, h *Handler, oh *OAuthHandler, authMiddleware echo.MiddlewareFunc) {
	authGroup := e.Group("/api/auth")

	// Public auth endpoints
	authGroup.POST("/refresh", h.Refresh)
	authGroup.POST("/exchange", h.Exchange)

	// OAuth endpoints
	authGroup.POST("/oauth/google", oh.GoogleOAuth)
	authGroup.POST("/oauth/facebook", oh.FacebookOAuth)

	// Protected auth endpoints (require authentication)
	authGroup.POST("/mobile-handoff", h.MobileHandoff, authMiddleware)
	authGroup.POST("/logout", h.Logout, authMiddleware)
	authGroup.POST("/change-email", h.ChangeEmail, authMiddleware)
	authGroup.POST("/change-password", h.ChangePassword, authMiddleware)
}
