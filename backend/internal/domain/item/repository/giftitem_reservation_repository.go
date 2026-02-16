//go:generate go run github.com/matryer/moq@latest -out ../service/mock_giftitem_reservation_repository_test.go -pkg service . GiftItemReservationRepositoryInterface

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"wish-list/internal/app/database"
	"wish-list/internal/domain/item/models"
	reservationmodels "wish-list/internal/domain/reservation/models"
	"wish-list/internal/pkg/logger"
)

// GiftItemReservationRepositoryInterface defines operations for gift item reservations
type GiftItemReservationRepositoryInterface interface {
	// Reserve marks a gift item as reserved by a user
	Reserve(ctx context.Context, giftItemID, userID pgtype.UUID) (*models.GiftItem, error)
	// Unreserve removes reservation from a gift item
	Unreserve(ctx context.Context, giftItemID pgtype.UUID) (*models.GiftItem, error)
	// ReserveIfNotReserved atomically reserves a gift item if it's not already reserved
	ReserveIfNotReserved(ctx context.Context, giftItemID, userID pgtype.UUID) (*models.GiftItem, error)
	// DeleteWithReservationNotification deletes a gift item and returns active reservations
	DeleteWithReservationNotification(ctx context.Context, giftItemID pgtype.UUID) ([]*reservationmodels.Reservation, error)
}

// GiftItemReservationRepository handles reservation-related database operations
type GiftItemReservationRepository struct {
	db *database.DB
}

// NewGiftItemReservationRepository creates a new GiftItemReservationRepository
func NewGiftItemReservationRepository(db *database.DB) GiftItemReservationRepositoryInterface {
	return &GiftItemReservationRepository{
		db: db,
	}
}

// giftItemColumns is the standard column list for gift_items queries
const giftItemColumnsReservation = `id, owner_id, name, description, link, image_url, price, priority,
	reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at,
	purchased_price, notes, position, archived_at, created_at, updated_at`

// Reserve marks a gift item as reserved by a user
func (r *GiftItemReservationRepository) Reserve(ctx context.Context, giftItemID, userID pgtype.UUID) (*models.GiftItem, error) {
	query := fmt.Sprintf(`
		UPDATE gift_items SET
			reserved_by_user_id = $2,
			reserved_at = $3,
			updated_at = NOW()
		WHERE id = $1
		RETURNING %s
	`, giftItemColumnsReservation)

	var updatedGiftItem models.GiftItem
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	err := r.db.QueryRowxContext(ctx, query,
		giftItemID,
		userID,
		now,
	).StructScan(&updatedGiftItem)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGiftItemNotFound
		}
		return nil, fmt.Errorf("failed to reserve gift item: %w", err)
	}

	return &updatedGiftItem, nil
}

// Unreserve removes reservation from a gift item
func (r *GiftItemReservationRepository) Unreserve(ctx context.Context, giftItemID pgtype.UUID) (*models.GiftItem, error) {
	query := fmt.Sprintf(`
		UPDATE gift_items SET
			reserved_by_user_id = NULL,
			reserved_at = NULL,
			updated_at = NOW()
		WHERE id = $1
		RETURNING %s
	`, giftItemColumnsReservation)

	var updatedGiftItem models.GiftItem
	err := r.db.QueryRowxContext(ctx, query, giftItemID).StructScan(&updatedGiftItem)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGiftItemNotFound
		}
		return nil, fmt.Errorf("failed to unreserve gift item: %w", err)
	}

	return &updatedGiftItem, nil
}

// ReserveIfNotReserved atomically reserves a gift item if it's not already reserved
func (r *GiftItemReservationRepository) ReserveIfNotReserved(ctx context.Context, giftItemID, userID pgtype.UUID) (*models.GiftItem, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			logger.Warn("transaction rollback error", "error", rbErr)
		}
	}()

	lockQuery := `
		SELECT id, reserved_by_user_id, reserved_at
		FROM gift_items
		WHERE id = $1
		FOR UPDATE
	`

	var currentItem models.GiftItem
	err = tx.GetContext(ctx, &currentItem, lockQuery, giftItemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGiftItemNotFound
		}
		return nil, fmt.Errorf("failed to lock gift item: %w", err)
	}

	if currentItem.ReservedByUserID.Valid {
		return nil, ErrGiftItemAlreadyReserved
	}

	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	updateQuery := fmt.Sprintf(`
		UPDATE gift_items SET
			reserved_by_user_id = $2,
			reserved_at = $3,
			updated_at = NOW()
		WHERE id = $1 AND reserved_by_user_id IS NULL
		RETURNING %s
	`, giftItemColumnsReservation)

	var updatedGiftItem models.GiftItem
	err = tx.QueryRowxContext(ctx, updateQuery,
		giftItemID,
		userID,
		now,
	).StructScan(&updatedGiftItem)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGiftItemConcurrentReserve
		}
		return nil, fmt.Errorf("failed to reserve gift item: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit reservation: %w", err)
	}

	return &updatedGiftItem, nil
}

// DeleteWithReservationNotification deletes a gift item and returns active reservations
func (r *GiftItemReservationRepository) DeleteWithReservationNotification(ctx context.Context, giftItemID pgtype.UUID) ([]*reservationmodels.Reservation, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if rbErr := tx.Rollback(); rbErr != nil && !errors.Is(rbErr, sql.ErrTxDone) {
			logger.Warn("transaction rollback error", "error", rbErr)
		}
	}()

	getReservationsQuery := `
		SELECT id, wishlist_id, gift_item_id, reserved_by_user_id, guest_name, guest_email,
			reservation_token, status, reserved_at, expires_at, canceled_at,
			cancel_reason, notification_sent, updated_at
		FROM reservations
		WHERE gift_item_id = $1 AND status = 'active'
	`

	var activeReservations []*reservationmodels.Reservation
	err = tx.SelectContext(ctx, &activeReservations, getReservationsQuery, giftItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active reservations: %w", err)
	}

	deleteQuery := `DELETE FROM gift_items WHERE id = $1`
	result, err := tx.ExecContext(ctx, deleteQuery, giftItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete gift item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, ErrGiftItemNotFound
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return activeReservations, nil
}
