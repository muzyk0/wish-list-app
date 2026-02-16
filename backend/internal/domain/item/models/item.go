package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type GiftItem struct {
	ID                pgtype.UUID        `db:"id"`
	OwnerID           pgtype.UUID        `db:"owner_id"` // Items belong to users, not wishlists
	Name              string             `db:"name"`
	Description       pgtype.Text        `db:"description"`
	Link              pgtype.Text        `db:"link"`
	ImageUrl          pgtype.Text        `db:"image_url"`
	Price             pgtype.Numeric     `db:"price"`
	Priority          pgtype.Int4        `db:"priority"`
	ReservedByUserID  pgtype.UUID        `db:"reserved_by_user_id"`
	ReservedAt        pgtype.Timestamptz `db:"reserved_at"`
	PurchasedByUserID pgtype.UUID        `db:"purchased_by_user_id"`
	PurchasedAt       pgtype.Timestamptz `db:"purchased_at"`
	PurchasedPrice    pgtype.Numeric     `db:"purchased_price"`
	Notes             pgtype.Text        `db:"notes"`
	Position          pgtype.Int4        `db:"position"`
	ArchivedAt        pgtype.Timestamptz `db:"archived_at"` // Soft delete
	CreatedAt         pgtype.Timestamptz `db:"created_at"`
	UpdatedAt         pgtype.Timestamptz `db:"updated_at"`
}
