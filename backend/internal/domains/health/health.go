package health

import (
	"wish-list/internal/domains/health/handlers"
	db "wish-list/internal/shared/db/models"
)

// NewHealthHandler creates a new health handler for monitoring endpoints
func NewHealthHandler(database *db.DB) *handlers.HealthHandler {
	return handlers.NewHealthHandler(database)
}
