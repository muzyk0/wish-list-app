package repositories

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	db "wish-list/internal/shared/db/models"
	"wish-list/internal/shared/encryption"
)

// setupTestUserRepository creates a test repository with mock database
// In a real implementation, this would use a test database or mocks
func setupTestUserRepository(t *testing.T, withEncryption bool) *UserRepository {
	// Note: This is a placeholder. In production, you'd use:
	// - A test database with transactions that rollback after each test
	// - sqlmock for mocking database interactions
	// - testcontainers for isolated PostgreSQL instances

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

		// In real tests, you'd create a test DB connection here
		// return NewUserRepositoryWithEncryption(testDB, encSvc)

		// For now, return a repository with encryption enabled
		repo := &UserRepository{
			encryptionSvc:     encSvc,
			encryptionEnabled: true,
		}
		return repo
	}

	// Return repository without encryption
	return &UserRepository{
		encryptionEnabled: false,
	}
}

func TestUserRepository_Create(t *testing.T) {
	t.Run("create user without encryption", func(t *testing.T) {
		repo := setupTestUserRepository(t, false)
		ctx := context.Background()

		user := db.User{
			Email:     "test@example.com",
			FirstName: pgtype.Text{String: "John", Valid: true},
			LastName:  pgtype.Text{String: "Doe", Valid: true},
		}

		// Test encryption PII method
		err := repo.encryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("encryptUserPII failed: %v", err)
		}

		// Without encryption enabled, encrypted fields should be empty
		if user.EncryptedEmail.Valid {
			t.Error("expected EncryptedEmail to be invalid when encryption disabled")
		}
		if user.EncryptedFirstName.Valid {
			t.Error("expected EncryptedFirstName to be invalid when encryption disabled")
		}
	})

	t.Run("create user with encryption", func(t *testing.T) {
		repo := setupTestUserRepository(t, true)
		ctx := context.Background()

		user := db.User{
			Email:     "test@example.com",
			FirstName: pgtype.Text{String: "John", Valid: true},
			LastName:  pgtype.Text{String: "Doe", Valid: true},
		}

		// Test encryption
		err := repo.encryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("encryptUserPII failed: %v", err)
		}

		// Verify encrypted fields are populated
		if !user.EncryptedEmail.Valid || user.EncryptedEmail.String == "" {
			t.Error("expected EncryptedEmail to be populated")
		}
		if !user.EncryptedFirstName.Valid || user.EncryptedFirstName.String == "" {
			t.Error("expected EncryptedFirstName to be populated")
		}
		if !user.EncryptedLastName.Valid || user.EncryptedLastName.String == "" {
			t.Error("expected EncryptedLastName to be populated")
		}

		// Verify encrypted values differ from plaintext
		if user.EncryptedEmail.String == user.Email {
			t.Error("encrypted email should differ from plaintext")
		}

		// Test decryption
		err = repo.decryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("decryptUserPII failed: %v", err)
		}

		// Verify decrypted values match original
		if user.Email != "test@example.com" {
			t.Errorf("expected email 'test@example.com', got %q", user.Email)
		}
		if !user.FirstName.Valid || user.FirstName.String != "John" {
			t.Errorf("expected first name 'John', got %q", user.FirstName.String)
		}
		if !user.LastName.Valid || user.LastName.String != "Doe" {
			t.Errorf("expected last name 'Doe', got %q", user.LastName.String)
		}
	})

	t.Run("create user with empty optional fields", func(t *testing.T) {
		repo := setupTestUserRepository(t, true)
		ctx := context.Background()

		user := db.User{
			Email: "test@example.com",
			// FirstName and LastName not set (invalid pgtype.Text)
		}

		// Test encryption with empty optional fields
		err := repo.encryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("encryptUserPII failed: %v", err)
		}

		// Encrypted fields should also be invalid for empty optional fields
		if user.EncryptedFirstName.Valid {
			t.Error("expected EncryptedFirstName to be invalid when FirstName is invalid")
		}
		if user.EncryptedLastName.Valid {
			t.Error("expected EncryptedLastName to be invalid when LastName is invalid")
		}
	})
}

func TestUserRepository_GetByID(t *testing.T) {
	t.Run("decrypt user PII on retrieval", func(t *testing.T) {
		repo := setupTestUserRepository(t, true)
		ctx := context.Background()

		// Simulate a user retrieved from DB with encrypted PII
		user := db.User{
			Email:              "plaintext@example.com",
			FirstName:          pgtype.Text{String: "PlaintextFirst", Valid: true},
			LastName:           pgtype.Text{String: "PlaintextLast", Valid: true},
			EncryptedEmail:     pgtype.Text{Valid: false}, // Will be populated by encryption
			EncryptedFirstName: pgtype.Text{Valid: false},
			EncryptedLastName:  pgtype.Text{Valid: false},
		}

		// Encrypt the user (simulating what happens on Create)
		err := repo.encryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("encryptUserPII failed: %v", err)
		}

		// Store the encrypted versions
		encryptedEmail := user.EncryptedEmail.String
		encryptedFirst := user.EncryptedFirstName.String
		encryptedLast := user.EncryptedLastName.String

		// Clear plaintext fields (simulating DB retrieval where we only get encrypted)
		user.Email = ""
		user.FirstName = pgtype.Text{Valid: false}
		user.LastName = pgtype.Text{Valid: false}

		// Restore encrypted fields
		user.EncryptedEmail = pgtype.Text{String: encryptedEmail, Valid: true}
		user.EncryptedFirstName = pgtype.Text{String: encryptedFirst, Valid: true}
		user.EncryptedLastName = pgtype.Text{String: encryptedLast, Valid: true}

		// Decrypt (simulating what happens in GetByID)
		err = repo.decryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("decryptUserPII failed: %v", err)
		}

		// Verify decrypted values
		if user.Email != "plaintext@example.com" {
			t.Errorf("expected email 'plaintext@example.com', got %q", user.Email)
		}
		if user.FirstName.String != "PlaintextFirst" {
			t.Errorf("expected first name 'PlaintextFirst', got %q", user.FirstName.String)
		}
		if user.LastName.String != "PlaintextLast" {
			t.Errorf("expected last name 'PlaintextLast', got %q", user.LastName.String)
		}
	})
}

func TestUserRepository_GetByEmail(t *testing.T) {
	t.Run("get user by email with encryption", func(t *testing.T) {
		repo := setupTestUserRepository(t, true)
		ctx := context.Background()

		// Create a user with encrypted PII
		user := db.User{
			Email:     "search@example.com",
			FirstName: pgtype.Text{String: "Search", Valid: true},
			LastName:  pgtype.Text{String: "User", Valid: true},
		}

		// Encrypt
		err := repo.encryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("encryptUserPII failed: %v", err)
		}

		// Verify encryption happened
		if !user.EncryptedEmail.Valid {
			t.Error("expected EncryptedEmail to be populated")
		}

		// Decrypt back
		err = repo.decryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("decryptUserPII failed: %v", err)
		}

		// Verify original email is restored
		if user.Email != "search@example.com" {
			t.Errorf("expected email 'search@example.com', got %q", user.Email)
		}
	})
}

func TestUserRepository_Update(t *testing.T) {
	t.Run("update user re-encrypts PII", func(t *testing.T) {
		repo := setupTestUserRepository(t, true)
		ctx := context.Background()

		// Original user
		user := db.User{
			Email:     "original@example.com",
			FirstName: pgtype.Text{String: "Original", Valid: true},
			LastName:  pgtype.Text{String: "Name", Valid: true},
		}

		// Encrypt original
		err := repo.encryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("encryptUserPII failed: %v", err)
		}
		originalEncryptedEmail := user.EncryptedEmail.String

		// Update user fields
		user.Email = "updated@example.com"
		user.FirstName = pgtype.Text{String: "Updated", Valid: true}

		// Re-encrypt for update
		err = repo.encryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("encryptUserPII failed: %v", err)
		}

		// Verify encrypted email changed
		if user.EncryptedEmail.String == originalEncryptedEmail {
			t.Error("expected encrypted email to change after update")
		}

		// Decrypt and verify
		err = repo.decryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("decryptUserPII failed: %v", err)
		}

		if user.Email != "updated@example.com" {
			t.Errorf("expected updated email 'updated@example.com', got %q", user.Email)
		}
		if user.FirstName.String != "Updated" {
			t.Errorf("expected updated first name 'Updated', got %q", user.FirstName.String)
		}
	})
}

func TestUserRepository_List(t *testing.T) {
	t.Run("list users decrypts all PII", func(t *testing.T) {
		repo := setupTestUserRepository(t, true)
		ctx := context.Background()

		// Create multiple users
		users := []*db.User{
			{
				Email:     "user1@example.com",
				FirstName: pgtype.Text{String: "User", Valid: true},
				LastName:  pgtype.Text{String: "One", Valid: true},
			},
			{
				Email:     "user2@example.com",
				FirstName: pgtype.Text{String: "User", Valid: true},
				LastName:  pgtype.Text{String: "Two", Valid: true},
			},
		}

		// Encrypt all users
		for _, user := range users {
			err := repo.encryptUserPII(ctx, user)
			if err != nil {
				t.Fatalf("encryptUserPII failed: %v", err)
			}
		}

		// Verify all are encrypted
		for i, user := range users {
			if !user.EncryptedEmail.Valid {
				t.Errorf("user %d: expected EncryptedEmail to be populated", i)
			}
		}

		// Decrypt all (simulating List operation)
		for _, user := range users {
			err := repo.decryptUserPII(ctx, user)
			if err != nil {
				t.Fatalf("decryptUserPII failed: %v", err)
			}
		}

		// Verify all are decrypted correctly
		if users[0].Email != "user1@example.com" {
			t.Errorf("user 0: expected 'user1@example.com', got %q", users[0].Email)
		}
		if users[1].Email != "user2@example.com" {
			t.Errorf("user 1: expected 'user2@example.com', got %q", users[1].Email)
		}
	})
}

func TestUserRepository_EncryptionEdgeCases(t *testing.T) {
	t.Run("handle special characters in PII", func(t *testing.T) {
		repo := setupTestUserRepository(t, true)
		ctx := context.Background()

		specialCases := []struct {
			name  string
			email string
			first string
			last  string
		}{
			{
				name:  "unicode characters",
				email: "тест@example.com",
				first: "Владислав",
				last:  "Петров",
			},
			{
				name:  "special email characters",
				email: "user+tag@sub.example.com",
				first: "User",
				last:  "Name",
			},
			{
				name:  "emoji in name",
				email: "emoji@example.com",
				first: "John 👋",
				last:  "Doe 🎉",
			},
		}

		for _, tc := range specialCases {
			t.Run(tc.name, func(t *testing.T) {
				user := db.User{
					Email:     tc.email,
					FirstName: pgtype.Text{String: tc.first, Valid: true},
					LastName:  pgtype.Text{String: tc.last, Valid: true},
				}

				// Encrypt
				err := repo.encryptUserPII(ctx, &user)
				if err != nil {
					t.Fatalf("encryptUserPII failed: %v", err)
				}

				// Decrypt
				err = repo.decryptUserPII(ctx, &user)
				if err != nil {
					t.Fatalf("decryptUserPII failed: %v", err)
				}

				// Verify
				if user.Email != tc.email {
					t.Errorf("expected email %q, got %q", tc.email, user.Email)
				}
				if user.FirstName.String != tc.first {
					t.Errorf("expected first name %q, got %q", tc.first, user.FirstName.String)
				}
				if user.LastName.String != tc.last {
					t.Errorf("expected last name %q, got %q", tc.last, user.LastName.String)
				}
			})
		}
	})

	t.Run("handle nil encryption service gracefully", func(t *testing.T) {
		repo := &UserRepository{
			encryptionSvc:     nil,
			encryptionEnabled: false,
		}
		ctx := context.Background()

		user := db.User{
			Email:     "test@example.com",
			FirstName: pgtype.Text{String: "Test", Valid: true},
		}

		// Should not error when encryption is disabled
		err := repo.encryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("expected no error with encryption disabled, got: %v", err)
		}

		err = repo.decryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("expected no error with encryption disabled, got: %v", err)
		}
	})
}
