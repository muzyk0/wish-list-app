package server

import (
	healthhttp "wish-list/internal/domain/health/delivery/http"

	"wish-list/internal/app/database"

	"github.com/labstack/echo/v4"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// SetupRoutes registers all domain routes on the Echo instance.
// This function is the central router that calls each domain's RegisterRoutes().
// It will be completed in Phase 5 when all domains are migrated.
func SetupRoutes(e *echo.Echo, db *database.DB) {
	// Swagger documentation endpoint
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Health domain
	healthHandler := healthhttp.NewHandler(db)
	healthhttp.RegisterRoutes(e, healthHandler)

	// User domain (Phase 4C) - requires service dependencies, wired in Phase 5:
	// userhttp.RegisterRoutes(e, userHandler, authMiddleware)

	// Remaining domain route registration (Phase 4D-4I, wired in Phase 5):
	// authhttp.RegisterRoutes(...)
	// wishlisthttp.RegisterRoutes(...)
	// itemhttp.RegisterRoutes(...)
	// wishlistitemhttp.RegisterRoutes(...)
	// reservationhttp.RegisterRoutes(...)
	// storagehttp.RegisterRoutes(...)
}
