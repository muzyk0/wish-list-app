package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Reservation struct {
	ID                  pgtype.UUID        `db:"id"`
	WishlistID          pgtype.UUID        `db:"wishlist_id"` // Reservation is for item in specific wishlist
	GiftItemID          pgtype.UUID        `db:"gift_item_id"`
	ReservedByUserID    pgtype.UUID        `db:"reserved_by_user_id"`
	GuestName           pgtype.Text        `db:"guest_name"`
	EncryptedGuestName  pgtype.Text        `db:"encrypted_guest_name"` // PII encrypted
	GuestEmail          pgtype.Text        `db:"guest_email"`
	EncryptedGuestEmail pgtype.Text        `db:"encrypted_guest_email"` // PII encrypted
	ReservationToken    pgtype.UUID        `db:"reservation_token"`
	Status              string             `db:"status"`
	ReservedAt          pgtype.Timestamptz `db:"reserved_at"`
	ExpiresAt           pgtype.Timestamptz `db:"expires_at"`
	CanceledAt          pgtype.Timestamptz `db:"canceled_at"`
	CancelReason        pgtype.Text        `db:"cancel_reason"`
	NotificationSent    pgtype.Bool        `db:"notification_sent"`
	UpdatedAt           pgtype.Timestamptz `db:"updated_at"`
}
