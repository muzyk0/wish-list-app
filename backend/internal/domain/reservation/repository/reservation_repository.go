//go:generate go run github.com/matryer/moq@latest -out ../service/mock_reservation_repository_test.go -pkg service . ReservationRepositoryInterface

package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"

	"wish-list/internal/app/database"
	"wish-list/internal/domain/reservation/models"
	"wish-list/internal/pkg/encryption"
)

// Sentinel errors for reservation repository
var (
	ErrReservationNotFound = errors.New("reservation not found")
	ErrNoActiveReservation = errors.New("no active reservation found")
)

// ReservationRepositoryInterface defines the interface for reservation database operations
type ReservationRepositoryInterface interface {
	Create(ctx context.Context, reservation models.Reservation) (*models.Reservation, error)
	GetByID(ctx context.Context, id pgtype.UUID) (*models.Reservation, error)
	GetByToken(ctx context.Context, token pgtype.UUID) (*models.Reservation, error)
	GetByGiftItem(ctx context.Context, giftItemID pgtype.UUID) ([]*models.Reservation, error)
	GetActiveReservationForGiftItem(ctx context.Context, giftItemID pgtype.UUID) (*models.Reservation, error)
	GetReservationsByUser(ctx context.Context, userID pgtype.UUID, limit, offset int) ([]*models.Reservation, error)
	UpdateStatus(ctx context.Context, reservationID pgtype.UUID, status string, canceledAt pgtype.Timestamptz, cancelReason pgtype.Text) (*models.Reservation, error)
	UpdateStatusByToken(ctx context.Context, token pgtype.UUID, status string, canceledAt pgtype.Timestamptz, cancelReason pgtype.Text) (*models.Reservation, error)
	ListUserReservationsWithDetails(ctx context.Context, userID pgtype.UUID, limit, offset int) ([]ReservationDetail, error)
	ListGuestReservationsWithDetails(ctx context.Context, token pgtype.UUID) ([]ReservationDetail, error)
	CountUserReservations(ctx context.Context, userID pgtype.UUID) (int, error)
	ListWishlistOwnerReservations(ctx context.Context, ownerUserID pgtype.UUID, limit, offset int) ([]ReservationDetail, error)
	CountWishlistOwnerReservations(ctx context.Context, ownerUserID pgtype.UUID) (int, error)
	LinkGuestReservationsToUserByEmail(ctx context.Context, guestEmail string, userID pgtype.UUID) (int, error)
}

type ReservationDetail struct {
	ID                  pgtype.UUID
	GiftItemID          pgtype.UUID
	ReservedByUserID    pgtype.UUID
	GuestName           pgtype.Text
	EncryptedGuestName  pgtype.Text `db:"encrypted_guest_name"` // PII encrypted
	GuestEmail          pgtype.Text
	EncryptedGuestEmail pgtype.Text `db:"encrypted_guest_email"` // PII encrypted
	ReservationToken    pgtype.UUID
	Status              string
	ReservedAt          pgtype.Timestamptz
	ExpiresAt           pgtype.Timestamptz
	CanceledAt          pgtype.Timestamptz
	CancelReason        pgtype.Text
	NotificationSent    pgtype.Bool
	GiftItemName        pgtype.Text
	GiftItemImageURL    pgtype.Text
	GiftItemPrice       pgtype.Numeric
	WishlistID          pgtype.UUID
	WishlistTitle       pgtype.Text
	OwnerFirstName      pgtype.Text
	OwnerLastName       pgtype.Text
	// Populated only for owner-perspective queries (who reserved my items)
	ReserverFirstName pgtype.Text `db:"reserver_first_name"`
	ReserverLastName  pgtype.Text `db:"reserver_last_name"`
}

type ReservationRepository struct {
	db                *database.DB
	encryptionSvc     *encryption.Service
	encryptionEnabled bool
}

func NewReservationRepository(db *database.DB) ReservationRepositoryInterface {
	return &ReservationRepository{
		db:                db,
		encryptionEnabled: false, // Encryption disabled by default for backward compatibility
	}
}

// NewReservationRepositoryWithEncryption creates a new ReservationRepository with encryption enabled
func NewReservationRepositoryWithEncryption(db *database.DB, encryptionSvc *encryption.Service) ReservationRepositoryInterface {
	return &ReservationRepository{
		db:                db,
		encryptionSvc:     encryptionSvc,
		encryptionEnabled: encryptionSvc != nil,
	}
}

// encryptReservationPII encrypts guest PII fields in the reservation struct
func (r *ReservationRepository) encryptReservationPII(ctx context.Context, reservation *models.Reservation) error {
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
		// Avoid persisting plaintext when encryption is enabled
		reservation.GuestName = pgtype.Text{Valid: false}
	}

	// Encrypt guest email
	if reservation.GuestEmail.Valid {
		encrypted, err := r.encryptionSvc.Encrypt(ctx, reservation.GuestEmail.String)
		if err != nil {
			return fmt.Errorf("failed to encrypt guest email: %w", err)
		}
		reservation.EncryptedGuestEmail = pgtype.Text{String: encrypted, Valid: true}
		// Avoid persisting plaintext when encryption is enabled
		reservation.GuestEmail = pgtype.Text{Valid: false}
	}

	return nil
}

// decryptReservationPII decrypts guest PII fields in the reservation struct
func (r *ReservationRepository) decryptReservationPII(ctx context.Context, reservation *models.Reservation) error {
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

// decryptReservationDetailPII decrypts guest PII fields in the reservation detail struct
func (r *ReservationRepository) decryptReservationDetailPII(ctx context.Context, detail *ReservationDetail) error {
	if !r.encryptionEnabled || r.encryptionSvc == nil {
		return nil
	}

	// Decrypt guest name
	if detail.EncryptedGuestName.Valid {
		decrypted, err := r.encryptionSvc.Decrypt(ctx, detail.EncryptedGuestName.String)
		if err != nil {
			return fmt.Errorf("failed to decrypt guest name: %w", err)
		}
		detail.GuestName = pgtype.Text{String: decrypted, Valid: true}
	}

	// Decrypt guest email
	if detail.EncryptedGuestEmail.Valid {
		decrypted, err := r.encryptionSvc.Decrypt(ctx, detail.EncryptedGuestEmail.String)
		if err != nil {
			return fmt.Errorf("failed to decrypt guest email: %w", err)
		}
		detail.GuestEmail = pgtype.Text{String: decrypted, Valid: true}
	}

	return nil
}

// Create inserts a new reservation into the database
func (r *ReservationRepository) Create(ctx context.Context, reservation models.Reservation) (*models.Reservation, error) {
	// Encrypt guest PII before inserting
	if err := r.encryptReservationPII(ctx, &reservation); err != nil {
		return nil, fmt.Errorf("failed to encrypt reservation PII: %w", err)
	}

	query := `
		INSERT INTO reservations (
			wishlist_id, gift_item_id, reserved_by_user_id, guest_name, encrypted_guest_name,
			guest_email, encrypted_guest_email, status, reserved_at, expires_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		) RETURNING
			id, wishlist_id, gift_item_id, reserved_by_user_id, guest_name, encrypted_guest_name,
			guest_email, encrypted_guest_email, reservation_token, status, reserved_at,
			expires_at, canceled_at, cancel_reason, notification_sent, updated_at
	`

	var createdReservation models.Reservation
	err := r.db.QueryRowxContext(ctx, query,
		reservation.WishlistID,
		reservation.GiftItemID,
		reservation.ReservedByUserID,
		reservation.GuestName,
		reservation.EncryptedGuestName,
		reservation.GuestEmail,
		reservation.EncryptedGuestEmail,
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
func (r *ReservationRepository) GetByID(ctx context.Context, id pgtype.UUID) (*models.Reservation, error) {
	query := `
		SELECT
			id, wishlist_id, gift_item_id, reserved_by_user_id, guest_name, encrypted_guest_name,
			guest_email, encrypted_guest_email, reservation_token, status, reserved_at,
			expires_at, canceled_at, cancel_reason, notification_sent, updated_at
		FROM reservations
		WHERE id = $1
	`

	var reservation models.Reservation
	err := r.db.GetContext(ctx, &reservation, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReservationNotFound
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
func (r *ReservationRepository) GetByToken(ctx context.Context, token pgtype.UUID) (*models.Reservation, error) {
	query := `
		SELECT
			id, wishlist_id, gift_item_id, reserved_by_user_id, guest_name, encrypted_guest_name,
			guest_email, encrypted_guest_email, reservation_token, status, reserved_at,
			expires_at, canceled_at, cancel_reason, notification_sent, updated_at
		FROM reservations
		WHERE reservation_token = $1
	`

	var reservation models.Reservation
	err := r.db.GetContext(ctx, &reservation, query, token)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReservationNotFound
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
func (r *ReservationRepository) GetByGiftItem(ctx context.Context, giftItemID pgtype.UUID) ([]*models.Reservation, error) {
	query := `
		SELECT
			id, wishlist_id, gift_item_id, reserved_by_user_id, guest_name, encrypted_guest_name,
			guest_email, encrypted_guest_email, reservation_token, status, reserved_at,
			expires_at, canceled_at, cancel_reason, notification_sent, updated_at
		FROM reservations
		WHERE gift_item_id = $1
		ORDER BY reserved_at DESC
	`

	var reservations []*models.Reservation
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
func (r *ReservationRepository) GetActiveReservationForGiftItem(ctx context.Context, giftItemID pgtype.UUID) (*models.Reservation, error) {
	query := `
		SELECT
			id, wishlist_id, gift_item_id, reserved_by_user_id, guest_name, encrypted_guest_name,
			guest_email, encrypted_guest_email, reservation_token, status, reserved_at,
			expires_at, canceled_at, cancel_reason, notification_sent, updated_at
		FROM reservations
		WHERE gift_item_id = $1 AND status = 'active'
		LIMIT 1
	`

	var reservation models.Reservation
	err := r.db.GetContext(ctx, &reservation, query, giftItemID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoActiveReservation
		}
		return nil, fmt.Errorf("failed to get active reservation for gift item: %w", err)
	}

	// Decrypt guest PII before returning
	if err := r.decryptReservationPII(ctx, &reservation); err != nil {
		return nil, fmt.Errorf("failed to decrypt reservation PII: %w", err)
	}

	return &reservation, nil
}

// GetReservationsByUser retrieves reservations made by a user
func (r *ReservationRepository) GetReservationsByUser(ctx context.Context, userID pgtype.UUID, limit, offset int) ([]*models.Reservation, error) {
	query := `
		SELECT r.id, r.wishlist_id, r.gift_item_id, r.reserved_by_user_id, r.guest_name, r.encrypted_guest_name,
			r.guest_email, r.encrypted_guest_email, r.reservation_token, r.status, r.reserved_at,
			r.expires_at, r.canceled_at, r.cancel_reason, r.notification_sent, r.updated_at
		FROM reservations r
		JOIN gift_items gi ON r.gift_item_id = gi.id
		WHERE r.reserved_by_user_id = $1 AND r.status = 'active'
		ORDER BY r.reserved_at DESC
		LIMIT $2 OFFSET $3
	`

	var reservations []*models.Reservation
	err := r.db.SelectContext(ctx, &reservations, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get reservations by user: %w", err)
	}

	// Decrypt guest PII for all reservations
	for _, reservation := range reservations {
		if err := r.decryptReservationPII(ctx, reservation); err != nil {
			return nil, fmt.Errorf("failed to decrypt reservation PII: %w", err)
		}
	}

	return reservations, nil
}

// UpdateStatus updates the status of a reservation
func (r *ReservationRepository) UpdateStatus(ctx context.Context, reservationID pgtype.UUID, status string, canceledAt pgtype.Timestamptz, cancelReason pgtype.Text) (*models.Reservation, error) {
	query := `
		UPDATE reservations SET
			status = $2,
			canceled_at = CASE WHEN $2 = 'canceled' THEN $3 ELSE canceled_at END,
			cancel_reason = CASE WHEN $2 = 'canceled' THEN $4 ELSE cancel_reason END,
			updated_at = NOW()
		WHERE id = $1
		RETURNING
			id, wishlist_id, gift_item_id, reserved_by_user_id, guest_name, encrypted_guest_name,
			guest_email, encrypted_guest_email, reservation_token, status, reserved_at,
			expires_at, canceled_at, cancel_reason, notification_sent, updated_at
	`

	var updatedReservation models.Reservation
	err := r.db.QueryRowxContext(ctx, query,
		reservationID,
		status,
		canceledAt,
		cancelReason,
	).StructScan(&updatedReservation)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReservationNotFound
		}
		return nil, fmt.Errorf("failed to update reservation status: %w", err)
	}

	// Decrypt guest PII before returning
	if err := r.decryptReservationPII(ctx, &updatedReservation); err != nil {
		return nil, fmt.Errorf("failed to decrypt reservation PII: %w", err)
	}

	return &updatedReservation, nil
}

// UpdateStatusByToken updates the status of a reservation by token
func (r *ReservationRepository) UpdateStatusByToken(ctx context.Context, token pgtype.UUID, status string, canceledAt pgtype.Timestamptz, cancelReason pgtype.Text) (*models.Reservation, error) {
	query := `
		UPDATE reservations SET
			status = $2,
			canceled_at = CASE WHEN $2 = 'canceled' THEN $3 ELSE canceled_at END,
			cancel_reason = CASE WHEN $2 = 'canceled' THEN $4 ELSE cancel_reason END,
			updated_at = NOW()
		WHERE reservation_token = $1
		RETURNING
			id, wishlist_id, gift_item_id, reserved_by_user_id, guest_name, encrypted_guest_name,
			guest_email, encrypted_guest_email, reservation_token, status, reserved_at,
			expires_at, canceled_at, cancel_reason, notification_sent, updated_at
	`

	var updatedReservation models.Reservation
	err := r.db.QueryRowxContext(ctx, query,
		token,
		status,
		canceledAt,
		cancelReason,
	).StructScan(&updatedReservation)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrReservationNotFound
		}
		return nil, fmt.Errorf("failed to update reservation status by token: %w", err)
	}

	// Decrypt guest PII before returning
	if err := r.decryptReservationPII(ctx, &updatedReservation); err != nil {
		return nil, fmt.Errorf("failed to decrypt reservation PII: %w", err)
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
			r.encrypted_guest_name,
			r.guest_email,
			r.encrypted_guest_email,
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
			w.id as wishlist_id,
			w.title as wishlist_title,
			u.first_name as owner_first_name,
			u.last_name as owner_last_name
		FROM reservations r
		JOIN gift_items gi ON r.gift_item_id = gi.id
		JOIN wishlist_items wi ON wi.gift_item_id = gi.id
		JOIN wishlists w ON wi.wishlist_id = w.id
		LEFT JOIN users u ON w.owner_id = u.id
		WHERE r.reserved_by_user_id = $1 AND r.status IN ('active', 'canceled')
		ORDER BY r.reserved_at DESC
		LIMIT $2 OFFSET $3
	`

	var reservations []ReservationDetail
	err := r.db.SelectContext(ctx, &reservations, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list user reservations with details: %w", err)
	}

	// Decrypt guest PII for all reservations
	for i := range reservations {
		if err := r.decryptReservationDetailPII(ctx, &reservations[i]); err != nil {
			return nil, fmt.Errorf("failed to decrypt reservation detail PII: %w", err)
		}
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
			r.encrypted_guest_name,
			r.guest_email,
			r.encrypted_guest_email,
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
			w.id as wishlist_id,
			w.title as wishlist_title,
			u.first_name as owner_first_name,
			u.last_name as owner_last_name
		FROM reservations r
		JOIN gift_items gi ON r.gift_item_id = gi.id
		JOIN wishlist_items wi ON wi.gift_item_id = gi.id
		JOIN wishlists w ON wi.wishlist_id = w.id
		LEFT JOIN users u ON w.owner_id = u.id
		WHERE r.reservation_token = $1 AND r.status IN ('active', 'canceled')
		ORDER BY r.reserved_at DESC
	`

	var reservations []ReservationDetail
	err := r.db.SelectContext(ctx, &reservations, query, token)
	if err != nil {
		return nil, fmt.Errorf("failed to list guest reservations with details: %w", err)
	}

	// Decrypt guest PII for all reservations
	for i := range reservations {
		if err := r.decryptReservationDetailPII(ctx, &reservations[i]); err != nil {
			return nil, fmt.Errorf("failed to decrypt reservation detail PII: %w", err)
		}
	}

	return reservations, nil
}

func (r *ReservationRepository) CountUserReservations(ctx context.Context, userID pgtype.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM reservations
		WHERE reserved_by_user_id = $1 AND status IN ('active', 'canceled')
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to count user reservations: %w", err)
	}

	return count, nil
}

// ListWishlistOwnerReservations retrieves all reservations on items belonging to the specified user's wishlists.
// This is the "owner view": who reserved my wishlist items (includes both guest and authenticated reservations).
func (r *ReservationRepository) ListWishlistOwnerReservations(ctx context.Context, ownerUserID pgtype.UUID, limit, offset int) ([]ReservationDetail, error) {
	query := `
		SELECT
			r.id,
			r.gift_item_id,
			r.reserved_by_user_id,
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
			w.id as wishlist_id,
			w.title as wishlist_title
		FROM reservations r
		JOIN gift_items gi ON r.gift_item_id = gi.id
		JOIN wishlists w ON r.wishlist_id = w.id
		WHERE w.owner_id = $1 AND r.status IN ('active', 'canceled')
		ORDER BY r.reserved_at DESC
		LIMIT $2 OFFSET $3
	`

	var reservations []ReservationDetail
	err := r.db.SelectContext(ctx, &reservations, query, ownerUserID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list wishlist owner reservations: %w", err)
	}

	return reservations, nil
}

// CountWishlistOwnerReservations counts all reservations on items belonging to the specified user's wishlists.
func (r *ReservationRepository) CountWishlistOwnerReservations(ctx context.Context, ownerUserID pgtype.UUID) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM reservations r
		JOIN gift_items gi ON r.gift_item_id = gi.id
		JOIN wishlists w ON r.wishlist_id = w.id
		WHERE w.owner_id = $1 AND r.status IN ('active', 'canceled')
	`

	var count int
	err := r.db.GetContext(ctx, &count, query, ownerUserID)
	if err != nil {
		return 0, fmt.Errorf("failed to count wishlist owner reservations: %w", err)
	}

	return count, nil
}

// LinkGuestReservationsToUserByEmail attaches active guest reservations to a user account by email.
// This supports post-registration linking so guest reservations become visible in authenticated flows.
func (r *ReservationRepository) LinkGuestReservationsToUserByEmail(ctx context.Context, guestEmail string, userID pgtype.UUID) (int, error) {
	if !userID.Valid {
		return 0, fmt.Errorf("invalid user id")
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(guestEmail))
	if normalizedEmail == "" {
		return 0, nil
	}

	// Fast path when encryption is disabled: match directly in SQL.
	if !r.encryptionEnabled || r.encryptionSvc == nil {
		query := `
			UPDATE reservations
			SET reserved_by_user_id = $1, updated_at = NOW()
			WHERE reserved_by_user_id IS NULL
			  AND status = 'active'
			  AND guest_email IS NOT NULL
			  AND LOWER(TRIM(guest_email)) = $2
		`

		result, err := r.db.ExecContext(ctx, query, userID, normalizedEmail)
		if err != nil {
			return 0, fmt.Errorf("failed to link guest reservations by email: %w", err)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return 0, fmt.Errorf("failed to get affected rows for reservation linking: %w", err)
		}

		return int(affected), nil
	}

	// Encryption-enabled path: fetch candidate rows, decrypt, compare in app layer, then update matches.
	type candidate struct {
		ID                  pgtype.UUID `db:"id"`
		GuestEmail          pgtype.Text `db:"guest_email"`
		EncryptedGuestEmail pgtype.Text `db:"encrypted_guest_email"`
	}

	var candidates []candidate
	selectQuery := `
		SELECT id, guest_email, encrypted_guest_email
		FROM reservations
		WHERE reserved_by_user_id IS NULL
		  AND status = 'active'
		  AND (guest_email IS NOT NULL OR encrypted_guest_email IS NOT NULL)
	`

	if err := r.db.SelectContext(ctx, &candidates, selectQuery); err != nil {
		return 0, fmt.Errorf("failed to load guest reservation candidates: %w", err)
	}

	var matchedIDs []pgtype.UUID
	for _, c := range candidates {
		var email string

		switch {
		case c.EncryptedGuestEmail.Valid:
			decrypted, err := r.encryptionSvc.Decrypt(ctx, c.EncryptedGuestEmail.String)
			if err != nil {
				// Skip corrupted row to keep linking best-effort.
				continue
			}
			email = decrypted
		case c.GuestEmail.Valid:
			email = c.GuestEmail.String
		default:
			continue
		}

		if strings.ToLower(strings.TrimSpace(email)) == normalizedEmail {
			matchedIDs = append(matchedIDs, c.ID)
		}
	}

	if len(matchedIDs) == 0 {
		return 0, nil
	}

	updateQuery := `
		UPDATE reservations
		SET reserved_by_user_id = $1, updated_at = NOW()
		WHERE id = $2
		  AND reserved_by_user_id IS NULL
		  AND status = 'active'
	`

	linkedCount := 0
	for _, reservationID := range matchedIDs {
		result, err := r.db.ExecContext(ctx, updateQuery, userID, reservationID)
		if err != nil {
			return 0, fmt.Errorf("failed to link reservation %s: %w", reservationID.String(), err)
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return 0, fmt.Errorf("failed to get affected rows for reservation %s: %w", reservationID.String(), err)
		}

		linkedCount += int(affected)
	}

	return linkedCount, nil
}
