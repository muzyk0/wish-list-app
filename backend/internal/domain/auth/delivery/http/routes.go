package http

import (
	"github.com/labstack/echo/v4"

	"wish-list/internal/app/middleware"
)

// RegisterRoutes registers auth domain HTTP routes on the /api/auth group.
// It accepts both the auth Handler and the OAuthHandler, plus auth middleware for protected endpoints.
func RegisterRoutes(e *echo.Echo, h *Handler, oh *OAuthHandler, authMiddleware echo.MiddlewareFunc) {
	authGroup := e.Group("/api/auth")

	// Refresh endpoint - rate limited to prevent token brute force
	// Limit: 20 requests/minute per IP, burst of 30
	refreshLimiter := middleware.NewRefreshRateLimiter()
	authGroup.POST("/refresh", h.Refresh,
		middleware.AuthRateLimitMiddleware(refreshLimiter, middleware.IPIdentifier))

	// Exchange endpoint - rate limited to prevent handoff code enumeration
	// Limit: 10 requests/minute per IP, burst of 15
	exchangeLimiter := middleware.NewExchangeRateLimiter()
	authGroup.POST("/exchange", h.Exchange,
		middleware.AuthRateLimitMiddleware(exchangeLimiter, middleware.IPIdentifier))

	// OAuth endpoints with rate limiting (5 req/min)
	oauthLimiter := middleware.NewOAuthRateLimiter()
	oauthGroup := authGroup.Group("/oauth")
	oauthGroup.Use(middleware.AuthRateLimitMiddleware(oauthLimiter, middleware.IPIdentifier))
	oauthGroup.POST("/google", oh.GoogleOAuth)
	oauthGroup.POST("/facebook", oh.FacebookOAuth)

	// Protected auth endpoints (require authentication)
	authGroup.POST("/mobile-handoff", h.MobileHandoff, authMiddleware)
	authGroup.POST("/logout", h.Logout, authMiddleware)
	authGroup.POST("/change-email", h.ChangeEmail, authMiddleware)
	authGroup.POST("/change-password", h.ChangePassword, authMiddleware)
}
