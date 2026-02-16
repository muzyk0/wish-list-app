package http

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers health check routes on the Echo instance.
func RegisterRoutes(e *echo.Echo, h *Handler) {
	e.GET("/healthz", h.Health)
}
