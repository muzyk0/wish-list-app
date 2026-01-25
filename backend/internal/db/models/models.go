package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type GiftItem struct {
	ID                pgtype.UUID        `db:"id" json:"id"`
	WishlistID        pgtype.UUID        `db:"wishlist_id" json:"wishlist_id"`
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
	CreatedAt         pgtype.Timestamptz `db:"created_at" json:"created_at"`
	UpdatedAt         pgtype.Timestamptz `db:"updated_at" json:"updated_at"`
}

type Reservation struct {
	ID                  pgtype.UUID        `db:"id" json:"id"`
	GiftItemID          pgtype.UUID        `db:"gift_item_id" json:"gift_item_id"`
	ReservedByUserID    pgtype.UUID        `db:"reserved_by_user_id" json:"reserved_by_user_id"`
	GuestName           pgtype.Text        `db:"guest_name" json:"guest_name"`
	EncryptedGuestName  pgtype.Text        `db:"encrypted_guest_name" json:"-"` // PII encrypted
	GuestEmail          pgtype.Text        `db:"guest_email" json:"guest_email"`
	EncryptedGuestEmail pgtype.Text        `db:"encrypted_guest_email" json:"-"` // PII encrypted
	ReservationToken    pgtype.UUID        `db:"reservation_token" json:"reservation_token"`
	Status              string             `db:"status" json:"status"`
	ReservedAt          pgtype.Timestamptz `db:"reserved_at" json:"reserved_at"`
	ExpiresAt           pgtype.Timestamptz `db:"expires_at" json:"expires_at"`
	CanceledAt          pgtype.Timestamptz `db:"canceled_at" json:"canceled_at"`
	CancelReason        pgtype.Text        `db:"cancel_reason" json:"cancel_reason"`
	NotificationSent    pgtype.Bool        `db:"notification_sent" json:"notification_sent"`
}

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

type User struct {
	ID                 pgtype.UUID        `db:"id" json:"id"`
	Email              string             `db:"email" json:"email"`
	EncryptedEmail     pgtype.Text        `db:"encrypted_email" json:"-"` // PII encrypted
	PasswordHash       pgtype.Text        `db:"password_hash" json:"password_hash"`
	FirstName          pgtype.Text        `db:"first_name" json:"first_name"`
	EncryptedFirstName pgtype.Text        `db:"encrypted_first_name" json:"-"` // PII encrypted
	LastName           pgtype.Text        `db:"last_name" json:"last_name"`
	EncryptedLastName  pgtype.Text        `db:"encrypted_last_name" json:"-"` // PII encrypted
	AvatarUrl          pgtype.Text        `db:"avatar_url" json:"avatar_url"`
	IsVerified         pgtype.Bool        `db:"is_verified" json:"is_verified"`
	CreatedAt          pgtype.Timestamptz `db:"created_at" json:"created_at"`
	UpdatedAt          pgtype.Timestamptz `db:"updated_at" json:"updated_at"`
	LastLoginAt        pgtype.Timestamptz `db:"last_login_at" json:"last_login_at"`
	DeactivatedAt      pgtype.Timestamptz `db:"deactivated_at" json:"deactivated_at"`
}

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
