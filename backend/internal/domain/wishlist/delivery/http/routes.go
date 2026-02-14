package http

import "github.com/labstack/echo/v4"

// RegisterRoutes registers all wishlist HTTP routes
func RegisterRoutes(e *echo.Echo, h *Handler, authMiddleware echo.MiddlewareFunc) {
	// Authenticated wishlist routes
	wishlists := e.Group("/api/wishlists", authMiddleware)
	wishlists.POST("", h.CreateWishList)
	wishlists.GET("", h.GetWishListsByOwner)
	wishlists.GET("/:id", h.GetWishList)
	wishlists.PUT("/:id", h.UpdateWishList)
	wishlists.DELETE("/:id", h.DeleteWishList)

	// Public wishlist routes (no auth required)
	public := e.Group("/api/public")
	public.GET("/wishlists/:slug", h.GetWishListByPublicSlug)
	public.GET("/wishlists/:slug/gift-items", h.GetGiftItemsByPublicSlug)
}
