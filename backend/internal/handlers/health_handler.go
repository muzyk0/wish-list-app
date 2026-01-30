package handlers

import (
	"context"
	"net/http"
	"time"

	db "wish-list/internal/db/models"

	"github.com/labstack/echo/v4"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	db *db.DB
}

// NewHealthHandler creates a new health check handler
func NewHealthHandler(database *db.DB) *HealthHandler {
	return &HealthHandler{
		db: database,
	}
}

// HealthResponse represents the response from the health check endpoint
type HealthResponse struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks,omitempty"`
	Error  string            `json:"error,omitempty"`
}

// Health checks the health of the application and its dependencies
func (h *HealthHandler) Health(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	// Check database connection
	if err := h.db.PingContext(ctx); err != nil {
		return c.JSON(http.StatusServiceUnavailable, HealthResponse{
			Status: "unhealthy",
			Error:  "database connection failed",
		})
	}

	return c.JSON(http.StatusOK, HealthResponse{
		Status: "healthy",
		Checks: map[string]string{
			"database": "ok",
		},
	})
}
