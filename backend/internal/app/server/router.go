package server

import (
	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// SetupRoutes registers all domain routes on the Echo instance.
// This function is the central router that calls each domain's RegisterRoutes().
// It will be completed in Phase 5 when all domains are migrated.
func SetupRoutes(e *echo.Echo) {
	// Swagger documentation endpoint
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Domain route registration will be added here during Phase 5:
	// health.RegisterRoutes(...)
	// auth.RegisterRoutes(...)
	// user.RegisterRoutes(...)
	// wishlist.RegisterRoutes(...)
	// item.RegisterRoutes(...)
	// wishlist_item.RegisterRoutes(...)
	// reservation.RegisterRoutes(...)
	// storage.RegisterRoutes(...)
}
