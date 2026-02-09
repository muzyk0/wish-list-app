package reservations

import (
	"wish-list/internal/domains/reservations/handlers"
	"wish-list/internal/domains/reservations/repositories"
	"wish-list/internal/domains/reservations/services"
	oldRepositories "wish-list/internal/repositories"
	db "wish-list/internal/shared/db/models"
)

// NewReservationHandler creates a fully initialized reservation handler with all dependencies
func NewReservationHandler(database *db.DB, giftItemRepo oldRepositories.GiftItemRepositoryInterface) *handlers.ReservationHandler {
	// Initialize repository
	reservationRepo := repositories.NewReservationRepository(database)

	// Initialize service (depends on gift item repo from old location until items domain is migrated)
	reservationService := services.NewReservationService(reservationRepo, giftItemRepo)

	// Return handler
	return handlers.NewReservationHandler(reservationService)
}
