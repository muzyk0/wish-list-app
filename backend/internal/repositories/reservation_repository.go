package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	db "wish-list/internal/db/models"
	"wish-list/internal/encryption"
)

// ReservationRepositoryInterface defines the interface for reservation database operations
type ReservationRepositoryInterface interface {
	Create(ctx context.Context, reservation db.Reservation) (*db.Reservation, error)
	GetByID(ctx context.Context, id pgtype.UUID) (*db.Reservation, error)
	GetByToken(ctx context.Context, token pgtype.UUID) (*db.Reservation, error)
	GetByGiftItem(ctx context.Context, giftItemID pgtype.UUID) ([]*db.Reservation, error)
	GetActiveReservationForGiftItem(ctx context.Context, giftItemID pgtype.UUID) (*db.Reservation, error)
	GetReservationsByUser(ctx context.Context, userID pgtype.UUID, limit, offset int) ([]*db.Reservation, error)
	UpdateStatus(ctx context.Context, reservationID pgtype.UUID, status string, canceledAt pgtype.Timestamptz, cancelReason pgtype.Text) (*db.Reservation, error)
	UpdateStatusByToken(ctx context.Context, token pgtype.UUID, status string, canceledAt pgtype.Timestamptz, cancelReason pgtype.Text) (*db.Reservation, error)
	ListUserReservationsWithDetails(ctx context.Context, userID pgtype.UUID, limit, offset int) ([]ReservationDetail, error)
	ListGuestReservationsWithDetails(ctx context.Context, token pgtype.UUID) ([]ReservationDetail, error)
}

type ReservationDetail struct {
	ID               pgtype.UUID        `json:"id"`
	GiftItemID       pgtype.UUID        `json:"gift_item_id"`
	ReservedByUserID pgtype.UUID        `json:"reserved_by_user_id"`
	GuestName        pgtype.Text        `json:"guest_name"`
	GuestEmail       pgtype.Text        `json:"guest_email"`
	ReservationToken pgtype.UUID        `json:"reservation_token"`
	Status           string             `json:"status"`
	ReservedAt       pgtype.Timestamptz `json:"reserved_at"`
	ExpiresAt        pgtype.Timestamptz `json:"expires_at"`
	CanceledAt       pgtype.Timestamptz `json:"canceled_at"`
	CancelReason     pgtype.Text        `json:"cancel_reason"`
	NotificationSent pgtype.Bool        `json:"notification_sent"`
	GiftItemName     pgtype.Text        `json:"gift_item_name"`
	GiftItemImageURL pgtype.Text        `json:"gift_item_image_url"`
	GiftItemPrice    pgtype.Numeric     `json:"gift_item_price"`
	WishlistTitle    pgtype.Text        `json:"wishlist_title"`
	OwnerFirstName   pgtype.Text        `json:"owner_first_name"`
	OwnerLastName    pgtype.Text        `json:"owner_last_name"`
}

type ReservationRepository struct {
	db                *db.DB
	encryptionSvc     *encryption.Service
	encryptionEnabled bool
}

func NewReservationRepository(database *db.DB) *ReservationRepository {
	return &ReservationRepository{
		db:                database,
		encryptionEnabled: false, // Encryption disabled by default for backward compatibility
	}
}

// NewReservationRepositoryWithEncryption creates a new ReservationRepository with encryption enabled
func NewReservationRepositoryWithEncryption(database *db.DB, encryptionSvc *encryption.Service) *ReservationRepository {
	return &ReservationRepository{
		db:                database,
		encryptionSvc:     encryptionSvc,
		encryptionEnabled: encryptionSvc != nil,
	}
}

// encryptReservationPII encrypts guest PII fields in the reservation struct
func (r *ReservationRepository) encryptReservationPII(ctx context.Context, reservation *db.Reservation) error {
	if !r.encryptionEnabled || r.encryptionSvc == nil {
		return nil
	}

	// Encrypt guest name
	if reservation.GuestName.Valid {
		encrypted, err := r.encryptionSvc.Encrypt(ctx, reservation.GuestName.String)
		if err != nil {
			return fmt.Errorf("failed to encrypt guest name: %w", err)
		}
		reservation.EncryptedGuestName = pgtype.Text{String: encrypted, Valid: true}
	}

	// Encrypt guest email
	if reservation.GuestEmail.Valid {
		encrypted, err := r.encryptionSvc.Encrypt(ctx, reservation.GuestEmail.String)
		if err != nil {
			return fmt.Errorf("failed to encrypt guest email: %w", err)
		}
		reservation.EncryptedGuestEmail = pgtype.Text{String: encrypted, Valid: true}
	}

	return nil
}

// decryptReservationPII decrypts guest PII fields in the reservation struct
func (r *ReservationRepository) decryptReservationPII(ctx context.Context, reservation *db.Reservation) error {
	if !r.encryptionEnabled || r.encryptionSvc == nil {
		return nil
	}

	// Decrypt guest name
	if reservation.EncryptedGuestName.Valid {
		decrypted, err := r.encryptionSvc.Decrypt(ctx, reservation.EncryptedGuestName.String)
		if err != nil {
			return fmt.Errorf("failed to decrypt guest name: %w", err)
		}
		reservation.GuestName = pgtype.Text{String: decrypted, Valid: true}
	}

	// Decrypt guest email
	if reservation.EncryptedGuestEmail.Valid {
		decrypted, err := r.encryptionSvc.Decrypt(ctx, reservation.EncryptedGuestEmail.String)
		if err != nil {
			return fmt.Errorf("failed to decrypt guest email: %w", err)
		}
		reservation.GuestEmail = pgtype.Text{String: decrypted, Valid: true}
	}

	return nil
}

// Create inserts a new reservation into the database
func (r *ReservationRepository) Create(ctx context.Context, reservation db.Reservation) (*db.Reservation, error) {
	// Encrypt guest PII before inserting
	if err := r.encryptReservationPII(ctx, &reservation); err != nil {
		return nil, fmt.Errorf("failed to encrypt reservation PII: %w", err)
	}

	query := `
		INSERT INTO reservations (
			gift_item_id, reserved_by_user_id, guest_name, encrypted_guest_name,
			guest_email, encrypted_guest_email, status, reserved_at, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		) RETURNING
			id, gift_item_id, reserved_by_user_id, guest_name, encrypted_guest_name,
			guest_email, encrypted_guest_email, reservation_token, status, reserved_at,
			expires_at, canceled_at, cancel_reason, notification_sent
	`

	var createdReservation db.Reservation
	err := r.db.QueryRowxContext(ctx, query,
		reservation.GiftItemID,
		reservation.ReservedByUserID,
		db.TextToString(reservation.GuestName),
		db.TextToString(reservation.EncryptedGuestName),
		db.TextToString(reservation.GuestEmail),
		db.TextToString(reservation.EncryptedGuestEmail),
		reservation.Status,
		reservation.ReservedAt,
		reservation.ExpiresAt,
	).StructScan(&createdReservation)

	if err != nil {
		return nil, fmt.Errorf("failed to create reservation: %w", err)
	}

	// Decrypt guest PII before returning
	if err := r.decryptReservationPII(ctx, &createdReservation); err != nil {
		return nil, fmt.Errorf("failed to decrypt reservation PII: %w", err)
	}

	return &createdReservation, nil
}

// GetByID retrieves a reservation by ID
func (r *ReservationRepository) GetByID(ctx context.Context, id pgtype.UUID) (*db.Reservation, error) {
	query := `
		SELECT
			id, gift_item_id, reserved_by_user_id, guest_name, encrypted_guest_name,
			guest_email, encrypted_guest_email, reservation_token, status, reserved_at,
			expires_at, canceled_at, cancel_reason, notification_sent
		FROM reservations
		WHERE id = $1
	`

	var reservation db.Reservation
	err := r.db.GetContext(ctx, &reservation, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("reservation not found")
		}
		return nil, fmt.Errorf("failed to get reservation: %w", err)
	}

	// Decrypt guest PII before returning
	if err := r.decryptReservationPII(ctx, &reservation); err != nil {
		return nil, fmt.Errorf("failed to decrypt reservation PII: %w", err)
	}

	return &reservation, nil
}

// GetByToken retrieves a reservation by token
func (r *ReservationRepository) GetByToken(ctx context.Context, token pgtype.UUID) (*db.Reservation, error) {
	query := `
		SELECT
			id, gift_item_id, reserved_by_user_id, guest_name, encrypted_guest_name,
			guest_email, encrypted_guest_email, reservation_token, status, reserved_at,
			expires_at, canceled_at, cancel_reason, notification_sent
		FROM reservations
		WHERE reservation_token = $1
	`

	var reservation db.Reservation
	err := r.db.GetContext(ctx, &reservation, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("reservation not found")
		}
		return nil, fmt.Errorf("failed to get reservation by token: %w", err)
	}

	// Decrypt guest PII before returning
	if err := r.decryptReservationPII(ctx, &reservation); err != nil {
		return nil, fmt.Errorf("failed to decrypt reservation PII: %w", err)
	}

	return &reservation, nil
}

// GetByGiftItem retrieves all reservations for a gift item
func (r *ReservationRepository) GetByGiftItem(ctx context.Context, giftItemID pgtype.UUID) ([]*db.Reservation, error) {
	query := `
		SELECT
			id, gift_item_id, reserved_by_user_id, guest_name, encrypted_guest_name,
			guest_email, encrypted_guest_email, reservation_token, status, reserved_at,
			expires_at, canceled_at, cancel_reason, notification_sent
		FROM reservations
		WHERE gift_item_id = $1
		ORDER BY reserved_at DESC
	`

	var reservations []*db.Reservation
	err := r.db.SelectContext(ctx, &reservations, query, giftItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservations by gift item: %w", err)
	}

	// Decrypt guest PII for all reservations
	for _, reservation := range reservations {
		if err := r.decryptReservationPII(ctx, reservation); err != nil {
			return nil, fmt.Errorf("failed to decrypt reservation PII: %w", err)
		}
	}

	return reservations, nil
}

// GetActiveReservationForGiftItem retrieves the active reservation for a gift item
func (r *ReservationRepository) GetActiveReservationForGiftItem(ctx context.Context, giftItemID pgtype.UUID) (*db.Reservation, error) {
	query := `
		SELECT
			id, gift_item_id, reserved_by_user_id, guest_name, guest_email, reservation_token, status, reserved_at, expires_at, canceled_at, cancel_reason, notification_sent
		FROM reservations
		WHERE gift_item_id = $1 AND status = 'active'
		LIMIT 1
	`

	var reservation db.Reservation
	err := r.db.GetContext(ctx, &reservation, query, giftItemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No active reservation
		}
		return nil, fmt.Errorf("failed to get active reservation for gift item: %w", err)
	}

	return &reservation, nil
}

// GetReservationsByUser retrieves reservations made by a user
func (r *ReservationRepository) GetReservationsByUser(ctx context.Context, userID pgtype.UUID, limit, offset int) ([]*db.Reservation, error) {
	query := `
		SELECT r.id, r.gift_item_id, r.reserved_by_user_id, r.guest_name, r.guest_email, r.reservation_token, r.status, r.reserved_at, r.expires_at, r.canceled_at, r.cancel_reason, r.notification_sent
		FROM reservations r
		JOIN gift_items gi ON r.gift_item_id = gi.id
		JOIN wishlists w ON gi.wishlist_id = w.id
		WHERE r.reserved_by_user_id = $1 AND r.status = 'active'
		ORDER BY r.reserved_at DESC
		LIMIT $2 OFFSET $3
	`

	var reservations []*db.Reservation
	err := r.db.SelectContext(ctx, &reservations, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservations by user: %w", err)
	}

	return reservations, nil
}

// UpdateStatus updates the status of a reservation
func (r *ReservationRepository) UpdateStatus(ctx context.Context, reservationID pgtype.UUID, status string, canceledAt pgtype.Timestamptz, cancelReason pgtype.Text) (*db.Reservation, error) {
	query := `
		UPDATE reservations SET
			status = $2,
			canceled_at = CASE WHEN $2 = 'cancelled' THEN $3 ELSE canceled_at END,
			cancel_reason = CASE WHEN $2 = 'cancelled' THEN $4 ELSE cancel_reason END,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id, gift_item_id, reserved_by_user_id, guest_name, guest_email, reservation_token, status, reserved_at, expires_at, canceled_at, cancel_reason, notification_sent
	`

	var updatedReservation db.Reservation
	err := r.db.QueryRowxContext(ctx, query,
		reservationID,
		status,
		canceledAt,
		db.TextToString(cancelReason),
	).StructScan(&updatedReservation)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("reservation not found")
		}
		return nil, fmt.Errorf("failed to update reservation status: %w", err)
	}

	return &updatedReservation, nil
}

// UpdateStatusByToken updates the status of a reservation by token
func (r *ReservationRepository) UpdateStatusByToken(ctx context.Context, token pgtype.UUID, status string, canceledAt pgtype.Timestamptz, cancelReason pgtype.Text) (*db.Reservation, error) {
	query := `
		UPDATE reservations SET
			status = $2,
			canceled_at = CASE WHEN $2 = 'cancelled' THEN $3 ELSE canceled_at END,
			cancel_reason = CASE WHEN $2 = 'cancelled' THEN $4 ELSE cancel_reason END,
			updated_at = NOW()
		WHERE reservation_token = $1
		RETURNING
			id, gift_item_id, reserved_by_user_id, guest_name, guest_email, reservation_token, status, reserved_at, expires_at, canceled_at, cancel_reason, notification_sent
	`

	var updatedReservation db.Reservation
	err := r.db.QueryRowxContext(ctx, query,
		token,
		status,
		canceledAt,
		db.TextToString(cancelReason),
	).StructScan(&updatedReservation)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("reservation not found")
		}
		return nil, fmt.Errorf("failed to update reservation status by token: %w", err)
	}

	return &updatedReservation, nil
}

// ListUserReservationsWithDetails retrieves reservations with detailed information for a user
func (r *ReservationRepository) ListUserReservationsWithDetails(ctx context.Context, userID pgtype.UUID, limit, offset int) ([]ReservationDetail, error) {
	query := `
		SELECT
			r.id,
			r.gift_item_id,
			r.reserved_by_user_id,
			r.guest_name,
			r.guest_email,
			r.reservation_token,
			r.status,
			r.reserved_at,
			r.expires_at,
			r.canceled_at,
			r.cancel_reason,
			r.notification_sent,
			gi.name as gift_item_name,
			gi.image_url as gift_item_image_url,
			gi.price as gift_item_price,
			w.title as wishlist_title,
			u.first_name as owner_first_name,
			u.last_name as owner_last_name
		FROM reservations r
		JOIN gift_items gi ON r.gift_item_id = gi.id
		JOIN wishlists w ON gi.wishlist_id = w.id
		LEFT JOIN users u ON w.owner_id = u.id
		WHERE r.reserved_by_user_id = $1 AND r.status IN ('active', 'cancelled')
		ORDER BY r.reserved_at DESC
		LIMIT $2 OFFSET $3
	`

	var reservations []ReservationDetail
	err := r.db.SelectContext(ctx, &reservations, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user reservations with details: %w", err)
	}

	return reservations, nil
}

// ListGuestReservationsWithDetails retrieves reservations with detailed information for a guest using token
func (r *ReservationRepository) ListGuestReservationsWithDetails(ctx context.Context, token pgtype.UUID) ([]ReservationDetail, error) {
	query := `
		SELECT
			r.id,
			r.gift_item_id,
			r.reserved_by_user_id,
			r.guest_name,
			r.guest_email,
			r.reservation_token,
			r.status,
			r.reserved_at,
			r.expires_at,
			r.canceled_at,
			r.cancel_reason,
			r.notification_sent,
			gi.name as gift_item_name,
			gi.image_url as gift_item_image_url,
			gi.price as gift_item_price,
			w.title as wishlist_title,
			u.first_name as owner_first_name,
			u.last_name as owner_last_name
		FROM reservations r
		JOIN gift_items gi ON r.gift_item_id = gi.id
		JOIN wishlists w ON gi.wishlist_id = w.id
		LEFT JOIN users u ON w.owner_id = u.id
		WHERE r.reservation_token = $1 AND r.status IN ('active', 'cancelled')
		ORDER BY r.reserved_at DESC
	`

	var reservations []ReservationDetail
	err := r.db.SelectContext(ctx, &reservations, query, token)
	if err != nil {
		return nil, fmt.Errorf("failed to list guest reservations with details: %w", err)
	}

	return reservations, nil
}
