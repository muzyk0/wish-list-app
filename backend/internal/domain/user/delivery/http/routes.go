package http

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers user domain HTTP routes
func RegisterRoutes(e *echo.Echo, h *Handler, authMiddleware echo.MiddlewareFunc) {
	// Public auth routes
	auth := e.Group("/api/auth")
	auth.POST("/register", h.Register)
	auth.POST("/login", h.Login)

	// Protected user routes
	protected := e.Group("/api/protected", authMiddleware)
	protected.GET("/profile", h.GetProfile)
	protected.PUT("/profile", h.UpdateProfile)
	protected.DELETE("/account", h.DeleteAccount)
	protected.GET("/export-data", h.ExportUserData)
}
