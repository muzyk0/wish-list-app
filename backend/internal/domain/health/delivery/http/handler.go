package http

import (
	"context"
	nethttp "net/http"
	"time"

	"wish-list/internal/app/database"
	"wish-list/internal/pkg/apperrors"

	"github.com/labstack/echo/v4"
)

// Handler handles health check endpoints
type Handler struct {
	db *database.DB
}

// NewHandler creates a new health check handler
func NewHandler(db *database.DB) *Handler {
	return &Handler{
		db: db,
	}
}

// HealthResponse represents the response from the health check endpoint
type HealthResponse struct {
	Status string            `json:"status" validate:"required"`
	Checks map[string]string `json:"checks,omitempty"`
	Error  string            `json:"error,omitempty"`
}

// Health godoc
//
//	@Summary		Health check endpoint
//	@Description	Performs a health check of the application and its dependencies (database)
//	@Tags			Health
//	@Produce		json
//	@Success		200	{object}	HealthResponse	"Application is healthy"
//	@Failure		503	{object}	HealthResponse	"Application is unhealthy"
//	@Router			/healthz [get]
//
// Health checks the health of the application and its dependencies
func (h *Handler) Health(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	// Check database connection
	if err := h.db.PingContext(ctx); err != nil {
		// Health check failures should return service unavailable with details
		return apperrors.New(nethttp.StatusServiceUnavailable, "database connection failed").Wrap(err)
	}

	return c.JSON(nethttp.StatusOK, HealthResponse{
		Status: "healthy",
		Checks: map[string]string{
			"database": "ok",
		},
	})
}
