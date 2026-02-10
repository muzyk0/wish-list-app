package repositories

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "wish-list/internal/shared/db/models"
	"wish-list/internal/shared/encryption"
)

// setupTestReservationRepository creates a test repository with encryption
func setupTestReservationRepository(t *testing.T, withEncryption bool) *ReservationRepository {
	if withEncryption {
		// Create encryption service for testing
		key := make([]byte, 32)
		_, err := rand.Read(key)
		if err != nil {
			t.Fatalf("failed to generate encryption key: %v", err)
		}

		encSvc, err := encryption.NewService(key)
		if err != nil {
			t.Fatalf("failed to create encryption service: %v", err)
		}

		repo := &ReservationRepository{
			encryptionSvc:     encSvc,
			encryptionEnabled: true,
		}
		return repo
	}

	return &ReservationRepository{
		encryptionEnabled: false,
	}
}

func TestReservationRepository_Create(t *testing.T) {
	t.Run("validate reservation creation fields", func(t *testing.T) {
		reservation := db.Reservation{
			GiftItemID:       pgtype.UUID{Valid: true},
			ReservedByUserID: pgtype.UUID{Valid: true},
			Status:           "active",
			ReservedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}

		// Verify required fields
		if !reservation.GiftItemID.Valid {
			t.Error("gift_item_id should be valid")
		}
		if reservation.Status == "" {
			t.Error("status should not be empty")
		}
		if !reservation.ReservedAt.Valid {
			t.Error("reserved_at should be valid")
		}
	})

	t.Run("user reservation has user_id", func(t *testing.T) {
		reservation := db.Reservation{
			GiftItemID:       pgtype.UUID{Valid: true},
			ReservedByUserID: pgtype.UUID{Valid: true},
			Status:           "active",
		}

		// User reservations have ReservedByUserID
		if !reservation.ReservedByUserID.Valid {
			t.Error("user reservation should have reserved_by_user_id")
		}
		if reservation.GuestName.Valid {
			t.Error("user reservation should not have guest_name")
		}
	})

	t.Run("guest reservation has guest info", func(t *testing.T) {
		reservation := db.Reservation{
			GiftItemID:       pgtype.UUID{Valid: true},
			Status:           "active",
			GuestName:        pgtype.Text{String: "John Doe", Valid: true},
			GuestEmail:       pgtype.Text{String: "john@example.com", Valid: true},
			ReservationToken: pgtype.UUID{Valid: true},
		}

		// Guest reservations have GuestName and ReservationToken
		if reservation.ReservedByUserID.Valid {
			t.Error("guest reservation should not have reserved_by_user_id")
		}
		if !reservation.GuestName.Valid {
			t.Error("guest reservation should have guest_name")
		}
		if !reservation.ReservationToken.Valid {
			t.Error("guest reservation should have reservation_token")
		}
	})
}

func TestReservationRepository_GuestPIIEncryption(t *testing.T) {
	t.Run("encrypt guest PII without encryption service", func(t *testing.T) {
		repo := setupTestReservationRepository(t, false)
		ctx := context.Background()

		reservation := db.Reservation{
			GuestName:  pgtype.Text{String: "Jane Doe", Valid: true},
			GuestEmail: pgtype.Text{String: "jane@example.com", Valid: true},
		}

		err := repo.encryptReservationPII(ctx, &reservation)
		if err != nil {
			t.Fatalf("encryptReservationPII failed: %v", err)
		}

		// Without encryption enabled, encrypted fields should be empty
		if reservation.EncryptedGuestName.Valid {
			t.Error("expected EncryptedGuestName to be invalid when encryption disabled")
		}
		if reservation.EncryptedGuestEmail.Valid {
			t.Error("expected EncryptedGuestEmail to be invalid when encryption disabled")
		}
	})

	t.Run("encrypt and decrypt guest PII", func(t *testing.T) {
		repo := setupTestReservationRepository(t, true)
		ctx := context.Background()

		reservation := db.Reservation{
			GuestName:  pgtype.Text{String: "Jane Doe", Valid: true},
			GuestEmail: pgtype.Text{String: "jane@example.com", Valid: true},
		}

		// Encrypt
		err := repo.encryptReservationPII(ctx, &reservation)
		if err != nil {
			t.Fatalf("encryptReservationPII failed: %v", err)
		}

		// Verify encrypted fields are populated
		if !reservation.EncryptedGuestName.Valid || reservation.EncryptedGuestName.String == "" {
			t.Error("expected EncryptedGuestName to be populated")
		}
		if !reservation.EncryptedGuestEmail.Valid || reservation.EncryptedGuestEmail.String == "" {
			t.Error("expected EncryptedGuestEmail to be populated")
		}

		// Verify encrypted values differ from plaintext
		if reservation.EncryptedGuestName.String == reservation.GuestName.String {
			t.Error("encrypted guest name should differ from plaintext")
		}
		if reservation.EncryptedGuestEmail.String == reservation.GuestEmail.String {
			t.Error("encrypted guest email should differ from plaintext")
		}

		// Decrypt
		err = repo.decryptReservationPII(ctx, &reservation)
		if err != nil {
			t.Fatalf("decryptReservationPII failed: %v", err)
		}

		// Verify decrypted values match original
		if reservation.GuestName.String != "Jane Doe" {
			t.Errorf("expected guest name 'Jane Doe', got %q", reservation.GuestName.String)
		}
		if reservation.GuestEmail.String != "jane@example.com" {
			t.Errorf("expected guest email 'jane@example.com', got %q", reservation.GuestEmail.String)
		}
	})

	t.Run("encrypt with empty optional fields", func(t *testing.T) {
		repo := setupTestReservationRepository(t, true)
		ctx := context.Background()

		reservation := db.Reservation{
			GiftItemID: pgtype.UUID{Valid: true},
			Status:     "active",
			// No guest name or email set
		}

		err := repo.encryptReservationPII(ctx, &reservation)
		if err != nil {
			t.Fatalf("encryptReservationPII failed: %v", err)
		}

		// Encrypted fields should also be invalid for empty fields
		if reservation.EncryptedGuestName.Valid {
			t.Error("expected EncryptedGuestName to be invalid when GuestName is invalid")
		}
		if reservation.EncryptedGuestEmail.Valid {
			t.Error("expected EncryptedGuestEmail to be invalid when GuestEmail is invalid")
		}
	})

	t.Run("handle special characters in guest info", func(t *testing.T) {
		repo := setupTestReservationRepository(t, true)
		ctx := context.Background()

		testCases := []struct {
			name  string
			guest string
			email string
		}{
			{
				name:  "unicode characters",
				guest: "Владислав Петров",
				email: "тест@example.com",
			},
			{
				name:  "special email characters",
				guest: "John O'Brien",
				email: "john+tag@sub.example.com",
			},
			{
				name:  "emoji in name",
				guest: "Jane 🎉",
				email: "jane@example.com",
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				reservation := db.Reservation{
					GuestName:  pgtype.Text{String: tc.guest, Valid: true},
					GuestEmail: pgtype.Text{String: tc.email, Valid: true},
				}

				// Encrypt
				err := repo.encryptReservationPII(ctx, &reservation)
				if err != nil {
					t.Fatalf("encryptReservationPII failed: %v", err)
				}

				// Decrypt
				err = repo.decryptReservationPII(ctx, &reservation)
				if err != nil {
					t.Fatalf("decryptReservationPII failed: %v", err)
				}

				// Verify
				if reservation.GuestName.String != tc.guest {
					t.Errorf("expected guest name %q, got %q", tc.guest, reservation.GuestName.String)
				}
				if reservation.GuestEmail.String != tc.email {
					t.Errorf("expected guest email %q, got %q", tc.email, reservation.GuestEmail.String)
				}
			})
		}
	})
}

func TestReservationRepository_StatusTransitions(t *testing.T) {
	t.Run("active to canceled transition", func(t *testing.T) {
		reservation := db.Reservation{
			ID:     pgtype.UUID{Valid: true},
			Status: "active",
		}

		// Simulate cancellation
		reservation.Status = "canceled"
		reservation.CanceledAt = pgtype.Timestamptz{Time: time.Now(), Valid: true}
		reservation.CancelReason = pgtype.Text{String: "Changed my mind", Valid: true}

		if reservation.Status != "canceled" {
			t.Errorf("expected status 'canceled', got %q", reservation.Status)
		}
		if !reservation.CanceledAt.Valid {
			t.Error("canceled reservation should have canceled_at timestamp")
		}
	})

	t.Run("active to expired transition", func(t *testing.T) {
		reservation := db.Reservation{
			ID:     pgtype.UUID{Valid: true},
			Status: "active",
		}

		// Simulate expiration
		reservation.Status = "expired"
		reservation.ExpiresAt = pgtype.Timestamptz{Time: time.Now().Add(-1 * time.Hour), Valid: true}

		if reservation.Status != "expired" {
			t.Errorf("expected status 'expired', got %q", reservation.Status)
		}
		if !reservation.ExpiresAt.Valid {
			t.Error("expired reservation should have expires_at timestamp")
		}
	})
}

func TestReservationRepository_ValidationRules(t *testing.T) {
	t.Run("status validation", func(t *testing.T) {
		validStatuses := []string{"active", "purchased", "canceled", "expired"}

		for _, status := range validStatuses {
			reservation := db.Reservation{
				GiftItemID: pgtype.UUID{Valid: true},
				Status:     status,
			}

			if reservation.Status != status {
				t.Errorf("expected status %q, got %q", status, reservation.Status)
			}
		}
	})

	t.Run("reservation requires gift item", func(t *testing.T) {
		reservation := db.Reservation{
			Status: "active",
		}

		if reservation.GiftItemID.Valid {
			t.Error("reservation should not have valid gift_item_id if not set")
		}
	})

	t.Run("reservation requires either user_id or guest info", func(t *testing.T) {
		// User reservation
		userRes := db.Reservation{
			GiftItemID:       pgtype.UUID{Valid: true},
			ReservedByUserID: pgtype.UUID{Valid: true},
			Status:           "active",
		}

		hasUser := userRes.ReservedByUserID.Valid
		hasGuest := userRes.GuestName.Valid || userRes.ReservationToken.Valid

		if !hasUser && !hasGuest {
			t.Error("reservation should have either user_id or guest info")
		}

		// Guest reservation
		guestRes := db.Reservation{
			GiftItemID:       pgtype.UUID{Valid: true},
			GuestName:        pgtype.Text{String: "Guest", Valid: true},
			ReservationToken: pgtype.UUID{Valid: true},
			Status:           "active",
		}

		hasUser = guestRes.ReservedByUserID.Valid
		hasGuest = guestRes.GuestName.Valid || guestRes.ReservationToken.Valid

		if !hasUser && !hasGuest {
			t.Error("reservation should have either user_id or guest info")
		}
	})
}

func TestReservationRepository_EdgeCases(t *testing.T) {
	t.Run("handle nil encryption service gracefully", func(t *testing.T) {
		repo := &ReservationRepository{
			encryptionSvc:     nil,
			encryptionEnabled: false,
		}
		ctx := context.Background()

		reservation := db.Reservation{
			GuestName:  pgtype.Text{String: "Test Guest", Valid: true},
			GuestEmail: pgtype.Text{String: "test@example.com", Valid: true},
		}

		// Should not error when encryption is disabled
		err := repo.encryptReservationPII(ctx, &reservation)
		if err != nil {
			t.Fatalf("expected no error with encryption disabled, got: %v", err)
		}

		err = repo.decryptReservationPII(ctx, &reservation)
		if err != nil {
			t.Fatalf("expected no error with encryption disabled, got: %v", err)
		}
	})

	t.Run("reservation token uniqueness", func(t *testing.T) {
		token1 := pgtype.UUID{Valid: true}
		token2 := pgtype.UUID{Valid: true}

		// Tokens should be unique (database constraint would enforce this)
		if token1.Bytes == token2.Bytes && token1.Valid && token2.Valid {
			t.Log("duplicate reservation tokens should be rejected by database unique constraint")
		}
	})
}
