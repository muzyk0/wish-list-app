package http

import "github.com/labstack/echo/v4"

// RegisterRoutes registers all reservation HTTP routes
func RegisterRoutes(
	e *echo.Echo,
	h *Handler,
	optionalAuthMiddleware echo.MiddlewareFunc,
	authMiddleware echo.MiddlewareFunc,
) {
	// Public reservation routes — guests and authenticated users.
	// optionalAuthMiddleware sets user context when token is present; guests proceed without it.
	public := e.Group("/api/public")
	public.POST("/reservations/wishlist/:wishlistId/item/:itemId", h.CreateReservation, optionalAuthMiddleware)
	public.DELETE("/reservations/wishlist/:wishlistId/item/:itemId", h.CancelReservation, optionalAuthMiddleware)
	public.GET("/reservations/list/:slug/item/:itemId", h.GetReservationStatus)

	// Authenticated-only reservation routes (mobile / registered users).
	authenticated := e.Group("/api/reservations", authMiddleware)
	authenticated.GET("/user", h.GetUserReservations)
	authenticated.GET("/wishlist-owner", h.GetWishlistOwnerReservations)

	// Guest reservation routes — no auth required, token-based.
	guest := e.Group("/api/guest")
	guest.GET("/reservations", h.GetGuestReservations)
}
