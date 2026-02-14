package http

import (
	"wish-list/internal/pkg/auth"

	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers storage routes on the Echo instance.
// The s3Client nil check is done at the caller level (app layer).
func RegisterRoutes(e *echo.Echo, h *Handler, tokenManager *auth.TokenManager) {
	imageUpload := e.Group("/api/images")
	imageUpload.Use(auth.JWTMiddleware(tokenManager))
	imageUpload.POST("/upload", h.UploadImage)
}
