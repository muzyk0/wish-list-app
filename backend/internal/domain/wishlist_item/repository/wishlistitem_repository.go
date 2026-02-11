//go:generate go run github.com/matryer/moq@latest -out ../service/mock_wishlistitem_repository_test.go -pkg service . WishlistItemRepositoryInterface

package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"wish-list/internal/app/database"
	db "wish-list/internal/shared/db/models"
)

// Sentinel errors for wishlist-item repository
var (
	ErrItemNotInWishlist = errors.New("item not found in wishlist")
)

// WishlistItemRepositoryInterface defines the interface for wishlist-item association operations
type WishlistItemRepositoryInterface interface {
	Attach(ctx context.Context, wishlistID, itemID pgtype.UUID) error
	Detach(ctx context.Context, wishlistID, itemID pgtype.UUID) error
	GetByWishlist(ctx context.Context, wishlistID pgtype.UUID, page, limit int) ([]*db.GiftItem, error)
	GetByWishlistCount(ctx context.Context, wishlistID pgtype.UUID) (int64, error)
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

// GetByWishlist retrieves all items in a wishlist with pagination
func (r *WishlistItemRepository) GetByWishlist(ctx context.Context, wishlistID pgtype.UUID, page, limit int) ([]*db.GiftItem, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	query := `
		SELECT
			gi.name, gi.id, gi.owner_id, gi.name, gi.description, gi.link, gi.image_url,
			gi.price, gi.priority, gi.reserved_by_user_id, gi.reserved_at,
			gi.purchased_by_user_id, gi.purchased_at, gi.purchased_price,
			gi.notes, gi.position, gi.archived_at, gi.created_at, gi.updated_at,gi.purchased_by_user_id, gi.reserved_by_user_id
		FROM gift_items gi
		INNER JOIN wishlist_items wi ON wi.gift_item_id = gi.id
		WHERE wi.wishlist_id = $1
		  AND gi.archived_at IS NULL
		ORDER BY wi.added_at DESC, gi.created_at DESC
		LIMIT $2 OFFSET $3
	`

	var items []*db.GiftItem
	if err := r.db.SelectContext(ctx, &items, query, wishlistID, limit, offset); err != nil {
		return nil, fmt.Errorf("failed to get wishlist items: %w", err)
	}

	return items, nil
}

// GetByWishlistCount returns the count of items in a wishlist
func (r *WishlistItemRepository) GetByWishlistCount(ctx context.Context, wishlistID pgtype.UUID) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM wishlist_items wi
		INNER JOIN gift_items gi ON gi.id = wi.gift_item_id
		WHERE wi.wishlist_id = $1
		  AND gi.archived_at IS NULL
	`

	var count int64
	if err := r.db.GetContext(ctx, &count, query, wishlistID); err != nil {
		return 0, fmt.Errorf("failed to count wishlist items: %w", err)
	}

	return count, nil
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
