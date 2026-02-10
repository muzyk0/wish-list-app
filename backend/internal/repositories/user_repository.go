package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	db "wish-list/internal/shared/db/models"
	"wish-list/internal/shared/encryption"
)

// Sentinel errors for user repository
var (
	ErrUserNotFound = errors.New("user not found")
)

//go:generate go run github.com/matryer/moq@latest -out ../services/mock_user_repository_test.go -pkg services . UserRepositoryInterface

// UserRepositoryInterface defines the interface for user database operations
type UserRepositoryInterface interface {
	Create(ctx context.Context, user db.User) (*db.User, error)
	GetByID(ctx context.Context, id pgtype.UUID) (*db.User, error)
	GetByEmail(ctx context.Context, email string) (*db.User, error)
	Update(ctx context.Context, user db.User) (*db.User, error)
	Delete(ctx context.Context, id pgtype.UUID) error
	DeleteWithExecutor(ctx context.Context, executor db.Executor, id pgtype.UUID) error
	List(ctx context.Context, limit, offset int) ([]*db.User, error)
	ListInactiveSince(ctx context.Context, since time.Time) ([]*db.User, error)
}

type UserRepository struct {
	db                *db.DB
	encryptionSvc     *encryption.Service
	encryptionEnabled bool
}

func NewUserRepository(database *db.DB) UserRepositoryInterface {
	return &UserRepository{
		db:                database,
		encryptionEnabled: false, // Encryption disabled by default for backward compatibility
	}
}

// NewUserRepositoryWithEncryption creates a new UserRepository with encryption enabled
func NewUserRepositoryWithEncryption(database *db.DB, encryptionSvc *encryption.Service) UserRepositoryInterface {
	return &UserRepository{
		db:                database,
		encryptionSvc:     encryptionSvc,
		encryptionEnabled: encryptionSvc != nil,
	}
}

// encryptUserPII encrypts PII fields in the user struct
func (r *UserRepository) encryptUserPII(ctx context.Context, user *db.User) error {
	if !r.encryptionEnabled || r.encryptionSvc == nil {
		return nil
	}

	// Encrypt email
	if user.Email != "" {
		encrypted, err := r.encryptionSvc.Encrypt(ctx, user.Email)
		if err != nil {
			return fmt.Errorf("failed to encrypt email: %w", err)
		}
		user.EncryptedEmail = pgtype.Text{String: encrypted, Valid: true}
	}

	// Encrypt first name
	if user.FirstName.Valid {
		encrypted, err := r.encryptionSvc.Encrypt(ctx, user.FirstName.String)
		if err != nil {
			return fmt.Errorf("failed to encrypt first name: %w", err)
		}
		user.EncryptedFirstName = pgtype.Text{String: encrypted, Valid: true}
	}

	// Encrypt last name
	if user.LastName.Valid {
		encrypted, err := r.encryptionSvc.Encrypt(ctx, user.LastName.String)
		if err != nil {
			return fmt.Errorf("failed to encrypt last name: %w", err)
		}
		user.EncryptedLastName = pgtype.Text{String: encrypted, Valid: true}
	}

	return nil
}

// decryptUserPII decrypts PII fields in the user struct
func (r *UserRepository) decryptUserPII(ctx context.Context, user *db.User) error {
	if !r.encryptionEnabled || r.encryptionSvc == nil {
		return nil
	}

	// Decrypt email
	if user.EncryptedEmail.Valid {
		decrypted, err := r.encryptionSvc.Decrypt(ctx, user.EncryptedEmail.String)
		if err != nil {
			return fmt.Errorf("failed to decrypt email: %w", err)
		}
		user.Email = decrypted
	}

	// Decrypt first name
	if user.EncryptedFirstName.Valid {
		decrypted, err := r.encryptionSvc.Decrypt(ctx, user.EncryptedFirstName.String)
		if err != nil {
			return fmt.Errorf("failed to decrypt first name: %w", err)
		}
		user.FirstName = pgtype.Text{String: decrypted, Valid: true}
	}

	// Decrypt last name
	if user.EncryptedLastName.Valid {
		decrypted, err := r.encryptionSvc.Decrypt(ctx, user.EncryptedLastName.String)
		if err != nil {
			return fmt.Errorf("failed to decrypt last name: %w", err)
		}
		user.LastName = pgtype.Text{String: decrypted, Valid: true}
	}

	return nil
}

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, user db.User) (*db.User, error) {
	// Encrypt PII before inserting
	if err := r.encryptUserPII(ctx, &user); err != nil {
		return nil, fmt.Errorf("failed to encrypt user PII: %w", err)
	}

	query := `
		INSERT INTO users (
			email, password_hash, first_name, last_name, avatar_url,
			encrypted_email, encrypted_first_name, encrypted_last_name
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8
		) RETURNING
			id, email, encrypted_email, first_name, encrypted_first_name,
			last_name, encrypted_last_name, avatar_url, is_verified,
			created_at, updated_at, last_login_at, deactivated_at
	`

	var createdUser db.User
	err := r.db.QueryRowxContext(ctx, query,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.AvatarUrl,
		user.EncryptedEmail,
		user.EncryptedFirstName,
		user.EncryptedLastName,
	).StructScan(&createdUser)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Decrypt PII before returning
	if err := r.decryptUserPII(ctx, &createdUser); err != nil {
		return nil, fmt.Errorf("failed to decrypt user PII: %w", err)
	}

	return &createdUser, nil
}

// GetByID retrieves a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id pgtype.UUID) (*db.User, error) {
	query := `
		SELECT
			id, email, encrypted_email, password_hash, first_name, encrypted_first_name,
			last_name, encrypted_last_name, avatar_url, is_verified,
			created_at, updated_at, last_login_at, deactivated_at
		FROM users
		WHERE id = $1
	`

	var user db.User
	err := r.db.GetContext(ctx, &user, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Decrypt PII before returning
	if err := r.decryptUserPII(ctx, &user); err != nil {
		return nil, fmt.Errorf("failed to decrypt user PII: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*db.User, error) {
	query := `
		SELECT
			id, email, encrypted_email, password_hash, first_name, encrypted_first_name,
			last_name, encrypted_last_name, avatar_url, is_verified,
			created_at, updated_at, last_login_at, deactivated_at
		FROM users
		WHERE email = $1
	`

	var user db.User
	err := r.db.GetContext(ctx, &user, query, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	// Decrypt PII before returning
	if err := r.decryptUserPII(ctx, &user); err != nil {
		return nil, fmt.Errorf("failed to decrypt user PII: %w", err)
	}

	return &user, nil
}

// Update modifies an existing user
func (r *UserRepository) Update(ctx context.Context, user db.User) (*db.User, error) {
	// Encrypt PII before updating
	if err := r.encryptUserPII(ctx, &user); err != nil {
		return nil, fmt.Errorf("failed to encrypt user PII: %w", err)
	}

	query := `
		UPDATE users SET
			email = $1,
			encrypted_email = $2,
			first_name = $3,
			encrypted_first_name = $4,
			last_name = $5,
			encrypted_last_name = $6,
			avatar_url = $7,
			updated_at = NOW()
		WHERE id = $8
		RETURNING
			id, email, encrypted_email, first_name, encrypted_first_name,
			last_name, encrypted_last_name, avatar_url, is_verified,
			created_at, updated_at, last_login_at, deactivated_at
	`

	var updatedUser db.User
	err := r.db.QueryRowxContext(ctx, query,
		user.Email,
		user.EncryptedEmail,
		user.FirstName,
		user.EncryptedFirstName,
		user.LastName,
		user.EncryptedLastName,
		user.AvatarUrl,
		user.ID,
	).StructScan(&updatedUser)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Decrypt PII before returning
	if err := r.decryptUserPII(ctx, &updatedUser); err != nil {
		return nil, fmt.Errorf("failed to decrypt user PII: %w", err)
	}

	return &updatedUser, nil
}

// Delete removes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id pgtype.UUID) error {
	return r.DeleteWithExecutor(ctx, r.db, id)
}

// DeleteWithExecutor removes a user by ID using the provided executor (for transactions)
func (r *UserRepository) DeleteWithExecutor(ctx context.Context, executor db.Executor, id pgtype.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := executor.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return ErrUserNotFound
	}

	return nil
}

// List retrieves users with pagination
func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]*db.User, error) {
	query := `
		SELECT
			id, email, encrypted_email, first_name, encrypted_first_name,
			last_name, encrypted_last_name, avatar_url, is_verified,
			created_at, updated_at, last_login_at, deactivated_at
		FROM users
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	var users []*db.User
	err := r.db.SelectContext(ctx, &users, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	// Decrypt PII for all users
	for _, user := range users {
		if err := r.decryptUserPII(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to decrypt user PII: %w", err)
		}
	}

	return users, nil
}

// ListInactiveSince retrieves users who haven't been active since the given date
func (r *UserRepository) ListInactiveSince(ctx context.Context, since time.Time) ([]*db.User, error) {
	query := `
		SELECT
			id, email, encrypted_email, first_name, encrypted_first_name,
			last_name, encrypted_last_name, avatar_url, is_verified,
			created_at, updated_at, last_login_at, deactivated_at
		FROM users
		WHERE last_login_at < $1 OR (last_login_at IS NULL AND created_at < $1)
		ORDER BY created_at DESC
	`

	var users []*db.User
	err := r.db.SelectContext(ctx, &users, query, since)
	if err != nil {
		return nil, fmt.Errorf("failed to list inactive users: %w", err)
	}

	// Decrypt PII for all users
	for _, user := range users {
		if err := r.decryptUserPII(ctx, user); err != nil {
			return nil, fmt.Errorf("failed to decrypt user PII: %w", err)
		}
	}

	return users, nil
}
