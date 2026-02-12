package db

// Type aliases pointing to canonical domain model types.
// This ensures backward compatibility during migration (Phase 5).
// All aliases are removed in Phase 6 when this file is deleted.

import (
	itemmodels "wish-list/internal/domain/item/models"
	reservationmodels "wish-list/internal/domain/reservation/models"
	usermodels "wish-list/internal/domain/user/models"
	wishlistmodels "wish-list/internal/domain/wishlist/models"
	wishlistitemmodels "wish-list/internal/domain/wishlist_item/models"

	"github.com/jackc/pgx/v5/pgtype"
)

type GiftItem = itemmodels.GiftItem
type WishList = wishlistmodels.WishList
type WishListWithItemCount = wishlistmodels.WishListWithItemCount
type User = usermodels.User
type Reservation = reservationmodels.Reservation
type WishlistItem = wishlistitemmodels.WishlistItem

// Template has no domain model (feature removed per business decision).
// Kept as regular struct until Phase 6 cleanup.
type Template struct {
	ID              string             `db:"id" json:"id"`
	Name            string             `db:"name" json:"name"`
	Description     pgtype.Text        `db:"description" json:"description"`
	PreviewImageUrl pgtype.Text        `db:"preview_image_url" json:"preview_image_url"`
	Config          []byte             `db:"config" json:"config"` // JSONB stored as bytes
	IsDefault       pgtype.Bool        `db:"is_default" json:"is_default"`
	CreatedAt       pgtype.Timestamptz `db:"created_at" json:"created_at"`
	UpdatedAt       pgtype.Timestamptz `db:"updated_at" json:"updated_at"`
}
