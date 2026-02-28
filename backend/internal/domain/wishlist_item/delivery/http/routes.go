package http

import (
	"github.com/labstack/echo/v4"
)

// RegisterRoutes registers wishlist-item domain HTTP routes
func RegisterRoutes(e *echo.Echo, h *Handler, authMiddleware echo.MiddlewareFunc) {
	wishlists := e.Group("/api/wishlists", authMiddleware)
	wishlists.GET("/:id/items", h.GetWishlistItems)
	wishlists.POST("/:id/items", h.AttachItemToWishlist)
	wishlists.POST("/:id/items/new", h.CreateItemInWishlist)
	wishlists.DELETE("/:id/items/:itemId", h.DetachItemFromWishlist)
	wishlists.PATCH("/:id/items/:itemId/mark-reserved", h.MarkManualReservation)
}
