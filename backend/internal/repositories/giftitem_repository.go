package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "wish-list/internal/db/models"
)

// GiftItemRepositoryInterface defines the interface for gift item database operations
type GiftItemRepositoryInterface interface {
	Create(ctx context.Context, giftItem db.GiftItem) (*db.GiftItem, error)
	GetByID(ctx context.Context, id pgtype.UUID) (*db.GiftItem, error)
	GetByWishList(ctx context.Context, wishlistID pgtype.UUID) ([]*db.GiftItem, error)
	Update(ctx context.Context, giftItem db.GiftItem) (*db.GiftItem, error)
	Delete(ctx context.Context, id pgtype.UUID) error
	DeleteWithExecutor(ctx context.Context, executor db.Executor, id pgtype.UUID) error
	Reserve(ctx context.Context, giftItemID, userID pgtype.UUID) (*db.GiftItem, error)
	Unreserve(ctx context.Context, giftItemID pgtype.UUID) (*db.GiftItem, error)
	MarkAsPurchased(ctx context.Context, giftItemID, userID pgtype.UUID, purchasedPrice pgtype.Numeric) (*db.GiftItem, error)
	GetPublicWishListGiftItems(ctx context.Context, publicSlug string) ([]*db.GiftItem, error)
	ReserveIfNotReserved(ctx context.Context, giftItemID, userID pgtype.UUID) (*db.GiftItem, error)
	DeleteWithReservationNotification(ctx context.Context, giftItemID pgtype.UUID) ([]*db.Reservation, error)
}

type GiftItemRepository struct {
	db *db.DB
}

func NewGiftItemRepository(database *db.DB) *GiftItemRepository {
	return &GiftItemRepository{
		db: database,
	}
}

// Create inserts a new gift item into the database
func (r *GiftItemRepository) Create(ctx context.Context, giftItem db.GiftItem) (*db.GiftItem, error) {
	query := `
		INSERT INTO gift_items (
			wishlist_id, name, description, link, image_url, price, priority, notes, position
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING
			id, wishlist_id, name, description, link, image_url, price, priority, reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at, purchased_price, notes, position, created_at, updated_at
	`

	var createdGiftItem db.GiftItem
	err := r.db.QueryRowxContext(ctx, query,
		giftItem.WishlistID,
		giftItem.Name,
		db.TextToString(giftItem.Description),
		db.TextToString(giftItem.Link),
		db.TextToString(giftItem.ImageUrl),
		giftItem.Price,
		giftItem.Priority,
		db.TextToString(giftItem.Notes),
		giftItem.Position,
	).StructScan(&createdGiftItem)

	if err != nil {
		return nil, fmt.Errorf("failed to create gift item: %w", err)
	}

	return &createdGiftItem, nil
}

// GetByID retrieves a gift item by ID
func (r *GiftItemRepository) GetByID(ctx context.Context, id pgtype.UUID) (*db.GiftItem, error) {
	query := `
		SELECT
			id, wishlist_id, name, description, link, image_url, price, priority, reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at, purchased_price, notes, position, created_at, updated_at
		FROM gift_items
		WHERE id = $1
	`

	var giftItem db.GiftItem
	err := r.db.GetContext(ctx, &giftItem, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("gift item not found")
		}
		return nil, fmt.Errorf("failed to get gift item: %w", err)
	}

	return &giftItem, nil
}

// GetByWishList retrieves gift items by wishlist ID
func (r *GiftItemRepository) GetByWishList(ctx context.Context, wishlistID pgtype.UUID) ([]*db.GiftItem, error) {
	query := `
		SELECT
			id, wishlist_id, name, description, link, image_url, price, priority, reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at, purchased_price, notes, position, created_at, updated_at
		FROM gift_items
		WHERE wishlist_id = $1
		ORDER BY position ASC
		LIMIT 100
	`

	var giftItems []*db.GiftItem
	err := r.db.SelectContext(ctx, &giftItems, query, wishlistID)
	if err != nil {
		return nil, fmt.Errorf("failed to get gift items by wishlist: %w", err)
	}

	return giftItems, nil
}

// Update modifies an existing gift item
func (r *GiftItemRepository) Update(ctx context.Context, giftItem db.GiftItem) (*db.GiftItem, error) {
	query := `
		UPDATE gift_items SET
			name = $2,
			description = $3,
			link = $4,
			image_url = $5,
			price = $6,
			priority = $7,
			notes = $8,
			position = $9,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id, wishlist_id, name, description, link, image_url, price, priority, reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at, purchased_price, notes, position, created_at, updated_at
	`

	var updatedGiftItem db.GiftItem
	err := r.db.QueryRowxContext(ctx, query,
		giftItem.ID,
		giftItem.Name,
		db.TextToString(giftItem.Description),
		db.TextToString(giftItem.Link),
		db.TextToString(giftItem.ImageUrl),
		giftItem.Price,
		giftItem.Priority,
		db.TextToString(giftItem.Notes),
		giftItem.Position,
	).StructScan(&updatedGiftItem)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("gift item not found")
		}
		return nil, fmt.Errorf("failed to update gift item: %w", err)
	}

	return &updatedGiftItem, nil
}

// Delete removes a gift item by ID
func (r *GiftItemRepository) Delete(ctx context.Context, id pgtype.UUID) error {
	return r.DeleteWithExecutor(ctx, r.db, id)
}

// DeleteWithExecutor removes a gift item by ID using the provided executor (for transactions)
func (r *GiftItemRepository) DeleteWithExecutor(ctx context.Context, executor db.Executor, id pgtype.UUID) error {
	query := `DELETE FROM gift_items WHERE id = $1`

	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete gift item: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.New("gift item not found")
	}

	return nil
}

// Reserve marks a gift item as reserved by a user
func (r *GiftItemRepository) Reserve(ctx context.Context, giftItemID, userID pgtype.UUID) (*db.GiftItem, error) {
	query := `
		UPDATE gift_items SET
			reserved_by_user_id = $2,
			reserved_at = $3,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id, wishlist_id, name, description, link, image_url, price, priority, reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at, purchased_price, notes, position, created_at, updated_at
	`

	var updatedGiftItem db.GiftItem
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	err := r.db.QueryRowxContext(ctx, query,
		giftItemID,
		userID,
		now,
	).StructScan(&updatedGiftItem)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("gift item not found")
		}
		return nil, fmt.Errorf("failed to reserve gift item: %w", err)
	}

	return &updatedGiftItem, nil
}

// Unreserve removes reservation from a gift item
func (r *GiftItemRepository) Unreserve(ctx context.Context, giftItemID pgtype.UUID) (*db.GiftItem, error) {
	query := `
		UPDATE gift_items SET
			reserved_by_user_id = NULL,
			reserved_at = NULL,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id, wishlist_id, name, description, link, image_url, price, priority, reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at, purchased_price, notes, position, created_at, updated_at
	`

	var updatedGiftItem db.GiftItem
	err := r.db.QueryRowxContext(ctx, query, giftItemID).StructScan(&updatedGiftItem)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("gift item not found")
		}
		return nil, fmt.Errorf("failed to unreserve gift item: %w", err)
	}

	return &updatedGiftItem, nil
}

// MarkAsPurchased marks a gift item as purchased
func (r *GiftItemRepository) MarkAsPurchased(ctx context.Context, giftItemID, userID pgtype.UUID, purchasedPrice pgtype.Numeric) (*db.GiftItem, error) {
	query := `
		UPDATE gift_items SET
			purchased_by_user_id = $2,
			purchased_at = $3,
			purchased_price = $4,
			reserved_by_user_id = NULL,
			reserved_at = NULL,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id, wishlist_id, name, description, link, image_url, price, priority, reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at, purchased_price, notes, position, created_at, updated_at
	`

	var updatedGiftItem db.GiftItem
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	err := r.db.QueryRowxContext(ctx, query,
		giftItemID,
		userID,
		now,
		purchasedPrice,
	).StructScan(&updatedGiftItem)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("gift item not found")
		}
		return nil, fmt.Errorf("failed to mark gift item as purchased: %w", err)
	}

	return &updatedGiftItem, nil
}

// GetPublicWishListGiftItems retrieves gift items for a public wishlist by slug
func (r *GiftItemRepository) GetPublicWishListGiftItems(ctx context.Context, publicSlug string) ([]*db.GiftItem, error) {
	query := `
		SELECT gi.id, gi.wishlist_id, gi.name, gi.description, gi.link, gi.image_url, gi.price, gi.priority, gi.reserved_by_user_id, gi.reserved_at, gi.purchased_by_user_id, gi.purchased_at, gi.purchased_price, gi.notes, gi.position, gi.created_at, gi.updated_at
		FROM gift_items gi
		JOIN wishlists w ON gi.wishlist_id = w.id
		WHERE w.public_slug = $1 AND w.is_public = true
		ORDER BY gi.position ASC
		LIMIT 100
	`

	var giftItems []*db.GiftItem
	err := r.db.SelectContext(ctx, &giftItems, query, publicSlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get public wishlist gift items: %w", err)
	}

	return giftItems, nil
}

// ReserveIfNotReserved atomically reserves a gift item if it's not already reserved
func (r *GiftItemRepository) ReserveIfNotReserved(ctx context.Context, giftItemID, userID pgtype.UUID) (*db.GiftItem, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Lock the gift item row for update to prevent concurrent reservations
	lockQuery := `
		SELECT id, reserved_by_user_id, reserved_at
		FROM gift_items
		WHERE id = $1
		FOR UPDATE
	`

	var currentItem db.GiftItem
	err = tx.GetContext(ctx, &currentItem, lockQuery, giftItemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("gift item not found")
		}
		return nil, fmt.Errorf("failed to lock gift item: %w", err)
	}

	// Check if already reserved
	if currentItem.ReservedByUserID.Valid {
		return nil, errors.New("gift item is already reserved")
	}

	// Now update the reservation
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	updateQuery := `
		UPDATE gift_items SET
			reserved_by_user_id = $2,
			reserved_at = $3,
			updated_at = NOW()
		WHERE id = $1 AND reserved_by_user_id IS NULL
		RETURNING
			id, wishlist_id, name, description, link, image_url, price, priority, reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at, purchased_price, notes, position, created_at, updated_at
	`

	var updatedGiftItem db.GiftItem
	err = tx.QueryRowxContext(ctx, updateQuery,
		giftItemID,
		userID,
		now,
	).StructScan(&updatedGiftItem)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("gift item was reserved by another transaction")
		}
		return nil, fmt.Errorf("failed to reserve gift item: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit reservation: %w", err)
	}

	return &updatedGiftItem, nil
}

// DeleteWithReservationNotification deletes a gift item and returns any active reservations for notification purposes
func (r *GiftItemRepository) DeleteWithReservationNotification(ctx context.Context, giftItemID pgtype.UUID) ([]*db.Reservation, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// First, get any active reservations for this gift item
	getReservationsQuery := `
		SELECT id, gift_item_id, reserved_by_user_id, guest_name, guest_email, reservation_token, status, reserved_at, expires_at, canceled_at, cancel_reason, notification_sent
		FROM reservations
		WHERE gift_item_id = $1 AND status = 'active'
	`

	var activeReservations []*db.Reservation
	err = tx.SelectContext(ctx, &activeReservations, getReservationsQuery, giftItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active reservations: %w", err)
	}

	// Delete the gift item
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
		return nil, errors.New("gift item not found")
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return activeReservations, nil
}
