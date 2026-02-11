package repository

import (
	"context"
	"crypto/rand"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"wish-list/internal/domain/user/models"
	"wish-list/internal/pkg/encryption"
)

// setupTestUserRepository creates a test repository with mock database
func setupTestUserRepository(t *testing.T, withEncryption bool) *UserRepository {
	if withEncryption {
		key := make([]byte, 32)
		_, err := rand.Read(key)
		if err != nil {
			t.Fatalf("failed to generate encryption key: %v", err)
		}

		encSvc, err := encryption.NewService(key)
		if err != nil {
			t.Fatalf("failed to create encryption service: %v", err)
		}

		repo := &UserRepository{
			encryptionSvc:     encSvc,
			encryptionEnabled: true,
		}
		return repo
	}

	return &UserRepository{
		encryptionEnabled: false,
	}
}

func TestUserRepository_Create(t *testing.T) {
	t.Run("create user without encryption", func(t *testing.T) {
		repo := setupTestUserRepository(t, false)
		ctx := context.Background()

		user := models.User{
			Email:     "test@example.com",
			FirstName: pgtype.Text{String: "John", Valid: true},
			LastName:  pgtype.Text{String: "Doe", Valid: true},
		}

		err := repo.encryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("encryptUserPII failed: %v", err)
		}

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

		user := models.User{
			Email:     "test@example.com",
			FirstName: pgtype.Text{String: "John", Valid: true},
			LastName:  pgtype.Text{String: "Doe", Valid: true},
		}

		err := repo.encryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("encryptUserPII failed: %v", err)
		}

		if !user.EncryptedEmail.Valid || user.EncryptedEmail.String == "" {
			t.Error("expected EncryptedEmail to be populated")
		}
		if !user.EncryptedFirstName.Valid || user.EncryptedFirstName.String == "" {
			t.Error("expected EncryptedFirstName to be populated")
		}
		if !user.EncryptedLastName.Valid || user.EncryptedLastName.String == "" {
			t.Error("expected EncryptedLastName to be populated")
		}

		if user.EncryptedEmail.String == user.Email {
			t.Error("encrypted email should differ from plaintext")
		}

		err = repo.decryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("decryptUserPII failed: %v", err)
		}

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

		user := models.User{
			Email: "test@example.com",
		}

		err := repo.encryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("encryptUserPII failed: %v", err)
		}

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

		user := models.User{
			Email:              "plaintext@example.com",
			FirstName:          pgtype.Text{String: "PlaintextFirst", Valid: true},
			LastName:           pgtype.Text{String: "PlaintextLast", Valid: true},
			EncryptedEmail:     pgtype.Text{Valid: false},
			EncryptedFirstName: pgtype.Text{Valid: false},
			EncryptedLastName:  pgtype.Text{Valid: false},
		}

		err := repo.encryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("encryptUserPII failed: %v", err)
		}

		encryptedEmail := user.EncryptedEmail.String
		encryptedFirst := user.EncryptedFirstName.String
		encryptedLast := user.EncryptedLastName.String

		user.Email = ""
		user.FirstName = pgtype.Text{Valid: false}
		user.LastName = pgtype.Text{Valid: false}

		user.EncryptedEmail = pgtype.Text{String: encryptedEmail, Valid: true}
		user.EncryptedFirstName = pgtype.Text{String: encryptedFirst, Valid: true}
		user.EncryptedLastName = pgtype.Text{String: encryptedLast, Valid: true}

		err = repo.decryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("decryptUserPII failed: %v", err)
		}

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

		user := models.User{
			Email:     "search@example.com",
			FirstName: pgtype.Text{String: "Search", Valid: true},
			LastName:  pgtype.Text{String: "User", Valid: true},
		}

		err := repo.encryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("encryptUserPII failed: %v", err)
		}

		if !user.EncryptedEmail.Valid {
			t.Error("expected EncryptedEmail to be populated")
		}

		err = repo.decryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("decryptUserPII failed: %v", err)
		}

		if user.Email != "search@example.com" {
			t.Errorf("expected email 'search@example.com', got %q", user.Email)
		}
	})
}

func TestUserRepository_Update(t *testing.T) {
	t.Run("update user re-encrypts PII", func(t *testing.T) {
		repo := setupTestUserRepository(t, true)
		ctx := context.Background()

		user := models.User{
			Email:     "original@example.com",
			FirstName: pgtype.Text{String: "Original", Valid: true},
			LastName:  pgtype.Text{String: "Name", Valid: true},
		}

		err := repo.encryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("encryptUserPII failed: %v", err)
		}
		originalEncryptedEmail := user.EncryptedEmail.String

		user.Email = "updated@example.com"
		user.FirstName = pgtype.Text{String: "Updated", Valid: true}

		err = repo.encryptUserPII(ctx, &user)
		if err != nil {
			t.Fatalf("encryptUserPII failed: %v", err)
		}

		if user.EncryptedEmail.String == originalEncryptedEmail {
			t.Error("expected encrypted email to change after update")
		}

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

		users := []*models.User{
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

		for _, user := range users {
			err := repo.encryptUserPII(ctx, user)
			if err != nil {
				t.Fatalf("encryptUserPII failed: %v", err)
			}
		}

		for i, user := range users {
			if !user.EncryptedEmail.Valid {
				t.Errorf("user %d: expected EncryptedEmail to be populated", i)
			}
		}

		for _, user := range users {
			err := repo.decryptUserPII(ctx, user)
			if err != nil {
				t.Fatalf("decryptUserPII failed: %v", err)
			}
		}

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
				user := models.User{
					Email:     tc.email,
					FirstName: pgtype.Text{String: tc.first, Valid: true},
					LastName:  pgtype.Text{String: tc.last, Valid: true},
				}

				err := repo.encryptUserPII(ctx, &user)
				if err != nil {
					t.Fatalf("encryptUserPII failed: %v", err)
				}

				err = repo.decryptUserPII(ctx, &user)
				if err != nil {
					t.Fatalf("decryptUserPII failed: %v", err)
				}

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

		user := models.User{
			Email:     "test@example.com",
			FirstName: pgtype.Text{String: "Test", Valid: true},
		}

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
