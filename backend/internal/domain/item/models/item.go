package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type GiftItem struct {
	ID                pgtype.UUID        `db:"id" json:"id"`
	OwnerID           pgtype.UUID        `db:"owner_id" json:"owner_id"` // Items belong to users, not wishlists
	Name              string             `db:"name" json:"name"`
	Description       pgtype.Text        `db:"description" json:"description"`
	Link              pgtype.Text        `db:"link" json:"link"`
	ImageUrl          pgtype.Text        `db:"image_url" json:"image_url"`
	Price             pgtype.Numeric     `db:"price" json:"price"`
	Priority          pgtype.Int4        `db:"priority" json:"priority"`
	ReservedByUserID  pgtype.UUID        `db:"reserved_by_user_id" json:"reserved_by_user_id"`
	ReservedAt        pgtype.Timestamptz `db:"reserved_at" json:"reserved_at"`
	PurchasedByUserID pgtype.UUID        `db:"purchased_by_user_id" json:"purchased_by_user_id"`
	PurchasedAt       pgtype.Timestamptz `db:"purchased_at" json:"purchased_at"`
	PurchasedPrice    pgtype.Numeric     `db:"purchased_price" json:"purchased_price"`
	Notes             pgtype.Text        `db:"notes" json:"notes"`
	Position          pgtype.Int4        `db:"position" json:"position"`
	ArchivedAt        pgtype.Timestamptz `db:"archived_at" json:"archived_at"` // Soft delete
	CreatedAt         pgtype.Timestamptz `db:"created_at" json:"created_at"`
	UpdatedAt         pgtype.Timestamptz `db:"updated_at" json:"updated_at"`
}
