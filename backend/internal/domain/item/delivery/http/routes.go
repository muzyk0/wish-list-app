package http

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers item domain HTTP routes
func RegisterRoutes(e *echo.Echo, h *Handler, authMiddleware echo.MiddlewareFunc) {
	// All item routes require authentication
	items := e.Group("/api/items", authMiddleware)
	items.GET("", h.GetMyItems)
	items.POST("", h.CreateItem)
	items.GET("/stats", h.GetHomeStats)
	items.GET("/:id", h.GetItem)
	items.PUT("/:id", h.UpdateItem)
	items.DELETE("/:id", h.DeleteItem)
	items.POST("/:id/mark-purchased", h.MarkItemAsPurchased)
}
