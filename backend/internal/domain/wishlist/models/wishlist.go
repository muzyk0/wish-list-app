package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type WishList struct {
	ID           pgtype.UUID        `db:"id"`
	OwnerID      pgtype.UUID        `db:"owner_id"`
	Title        string             `db:"title"`
	Description  pgtype.Text        `db:"description"`
	Occasion     pgtype.Text        `db:"occasion"`
	OccasionDate pgtype.Date        `db:"occasion_date"`
	IsPublic     pgtype.Bool        `db:"is_public"`
	PublicSlug   pgtype.Text        `db:"public_slug"`
	ViewCount    pgtype.Int4        `db:"view_count"`
	CreatedAt    pgtype.Timestamptz `db:"created_at"`
	UpdatedAt    pgtype.Timestamptz `db:"updated_at"`
}

// WishListWithItemCount extends WishList with item count (from JOIN query)
type WishListWithItemCount struct {
	WishList
	ItemCount int64 `db:"item_count"`
}
