package http

import "github.com/labstack/echo/v4"

// RegisterRoutes registers all reservation HTTP routes
func RegisterRoutes(e *echo.Echo, h *Handler, authMiddleware echo.MiddlewareFunc) {
	// Authenticated reservation routes
	reservations := e.Group("/api/reservations", authMiddleware)
	reservations.POST("/wishlist/:wishlistId/item/:itemId", h.CreateReservation)
	reservations.DELETE("/wishlist/:wishlistId/item/:itemId", h.CancelReservation)
	reservations.GET("/user", h.GetUserReservations)

	// Guest reservation routes (no auth required)
	guest := e.Group("/api/guest")
	guest.GET("/reservations", h.GetGuestReservations)

	// Public reservation status (no auth required)
	public := e.Group("/api/public")
	public.GET("/reservations/list/:slug/item/:itemId", h.GetReservationStatus)
}
