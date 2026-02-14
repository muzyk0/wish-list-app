package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type WishList struct {
	ID           pgtype.UUID        `db:"id" json:"id"`
	OwnerID      pgtype.UUID        `db:"owner_id" json:"owner_id"`
	Title        string             `db:"title" json:"title"`
	Description  pgtype.Text        `db:"description" json:"description"`
	Occasion     pgtype.Text        `db:"occasion" json:"occasion"`
	OccasionDate pgtype.Date        `db:"occasion_date" json:"occasion_date"`
	TemplateID   string             `db:"template_id" json:"template_id"`
	IsPublic     pgtype.Bool        `db:"is_public" json:"is_public"`
	PublicSlug   pgtype.Text        `db:"public_slug" json:"public_slug"`
	ViewCount    pgtype.Int4        `db:"view_count" json:"view_count"`
	CreatedAt    pgtype.Timestamptz `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamptz `db:"updated_at" json:"updated_at"`
}

// WishListWithItemCount extends WishList with item count (from JOIN query)
type WishListWithItemCount struct {
	WishList
	ItemCount int64 `db:"item_count" json:"item_count"`
}
