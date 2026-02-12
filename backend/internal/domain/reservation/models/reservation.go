package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Reservation struct {
	ID                  pgtype.UUID        `db:"id" json:"id"`
	WishlistID          pgtype.UUID        `db:"wishlist_id" json:"wishlist_id"` // Reservation is for item in specific wishlist
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
	CancelReason        pgtype.Text        `db:"cancel_reason" json:"canceled_reason"`
	NotificationSent    pgtype.Bool        `db:"notification_sent" json:"notification_sent"`
	UpdatedAt           pgtype.Timestamptz `db:"updated_at" json:"updated_at"`
}
