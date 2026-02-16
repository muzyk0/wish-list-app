package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// WishlistItem represents the many-to-many relationship between wishlists and items
type WishlistItem struct {
	WishlistID pgtype.UUID        `db:"wishlist_id"`
	GiftItemID pgtype.UUID        `db:"gift_item_id"`
	AddedAt    pgtype.Timestamptz `db:"added_at"`
}
