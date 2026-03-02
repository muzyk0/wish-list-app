//go:generate go run github.com/matryer/moq@latest -out ../service/mock_wishlistitem_repository_test.go -pkg service . WishlistItemRepositoryInterface

package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"wish-list/internal/app/database"
)

// Sentinel errors for wishlist-item repository
var (
	ErrItemNotInWishlist = errors.New("item not found in wishlist")
)

// WishlistItemRepositoryInterface defines the interface for wishlist-item association operations.
// Querying items by wishlist (with pagination) is handled by GiftItemRepositoryInterface.GetByWishListPaginated.
type WishlistItemRepositoryInterface interface {
	Attach(ctx context.Context, wishlistID, itemID pgtype.UUID) error
	Detach(ctx context.Context, wishlistID, itemID pgtype.UUID) error
	IsAttached(ctx context.Context, wishlistID, itemID pgtype.UUID) (bool, error)
	GetWishlistsForItem(ctx context.Context, itemID pgtype.UUID) ([]pgtype.UUID, error)
	DetachAll(ctx context.Context, itemID pgtype.UUID) error
}

// WishlistItemRepository implements WishlistItemRepositoryInterface
type WishlistItemRepository struct {
	db *database.DB
}

// NewWishlistItemRepository creates a new WishlistItemRepository
func NewWishlistItemRepository(db *database.DB) WishlistItemRepositoryInterface {
	return &WishlistItemRepository{
		db: db,
	}
}

// Attach creates an association between wishlist and item
func (r *WishlistItemRepository) Attach(ctx context.Context, wishlistID, itemID pgtype.UUID) error {
	query := `
		INSERT INTO wishlist_items (wishlist_id, gift_item_id, added_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (wishlist_id, gift_item_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, wishlistID, itemID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to attach item to wishlist: %w", err)
	}

	return nil
}

// Detach removes an association between wishlist and item
func (r *WishlistItemRepository) Detach(ctx context.Context, wishlistID, itemID pgtype.UUID) error {
	query := `
		DELETE FROM wishlist_items
		WHERE wishlist_id = $1 AND gift_item_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, wishlistID, itemID)
	if err != nil {
		return fmt.Errorf("failed to detach item from wishlist: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrItemNotInWishlist
	}

	return nil
}

// IsAttached checks if an item is attached to a wishlist
func (r *WishlistItemRepository) IsAttached(ctx context.Context, wishlistID, itemID pgtype.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM wishlist_items
			WHERE wishlist_id = $1 AND gift_item_id = $2
		)
	`

	var exists bool
	if err := r.db.GetContext(ctx, &exists, query, wishlistID, itemID); err != nil {
		return false, fmt.Errorf("failed to check item attachment: %w", err)
	}

	return exists, nil
}

// GetWishlistsForItem retrieves all wishlist IDs that an item is attached to
func (r *WishlistItemRepository) GetWishlistsForItem(ctx context.Context, itemID pgtype.UUID) ([]pgtype.UUID, error) {
	query := `
		SELECT wishlist_id
		FROM wishlist_items
		WHERE gift_item_id = $1
		ORDER BY added_at DESC
	`

	var wishlistIDs []pgtype.UUID
	if err := r.db.SelectContext(ctx, &wishlistIDs, query, itemID); err != nil {
		return nil, fmt.Errorf("failed to get wishlists for item: %w", err)
	}

	return wishlistIDs, nil
}

// DetachAll removes item from all wishlists
func (r *WishlistItemRepository) DetachAll(ctx context.Context, itemID pgtype.UUID) error {
	query := `
		DELETE FROM wishlist_items
		WHERE gift_item_id = $1
	`

	_, err := r.db.ExecContext(ctx, query, itemID)
	if err != nil {
		return fmt.Errorf("failed to detach item from all wishlists: %w", err)
	}

	return nil
}
