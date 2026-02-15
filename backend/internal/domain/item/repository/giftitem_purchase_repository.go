//go:generate go run github.com/matryer/moq@latest -out ../service/mock_giftitem_purchase_repository_test.go -pkg service . GiftItemPurchaseRepositoryInterface

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
)

// GiftItemPurchaseRepositoryInterface defines operations for gift item purchases
type GiftItemPurchaseRepositoryInterface interface {
	// MarkAsPurchased marks a gift item as purchased
	MarkAsPurchased(ctx context.Context, giftItemID, userID pgtype.UUID, purchasedPrice pgtype.Numeric) (*models.GiftItem, error)
}

// GiftItemPurchaseRepository handles purchase-related database operations
type GiftItemPurchaseRepository struct {
	db *database.DB
}

// NewGiftItemPurchaseRepository creates a new GiftItemPurchaseRepository
func NewGiftItemPurchaseRepository(db *database.DB) GiftItemPurchaseRepositoryInterface {
	return &GiftItemPurchaseRepository{
		db: db,
	}
}

// giftItemColumnsPurchase is the standard column list for gift_items queries
const giftItemColumnsPurchase = `id, owner_id, name, description, link, image_url, price, priority,
	reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at,
	purchased_price, notes, position, archived_at, created_at, updated_at`

// MarkAsPurchased marks a gift item as purchased
func (r *GiftItemPurchaseRepository) MarkAsPurchased(ctx context.Context, giftItemID, userID pgtype.UUID, purchasedPrice pgtype.Numeric) (*models.GiftItem, error) {
	query := fmt.Sprintf(`
		UPDATE gift_items SET
			purchased_by_user_id = $2,
			purchased_at = $3,
			purchased_price = $4,
			reserved_by_user_id = NULL,
			reserved_at = NULL,
			updated_at = NOW()
		WHERE id = $1
		RETURNING %s
	`, giftItemColumnsPurchase)

	var updatedGiftItem models.GiftItem
	now := pgtype.Timestamptz{Time: time.Now(), Valid: true}
	err := r.db.QueryRowxContext(ctx, query,
		giftItemID,
		userID,
		now,
		purchasedPrice,
	).StructScan(&updatedGiftItem)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGiftItemNotFound
		}
		return nil, fmt.Errorf("failed to mark gift item as purchased: %w", err)
	}

	return &updatedGiftItem, nil
}
