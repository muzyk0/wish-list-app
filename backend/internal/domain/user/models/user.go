package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID                 pgtype.UUID        `db:"id"`
	Email              string             `db:"email"`
	EncryptedEmail     pgtype.Text        `db:"encrypted_email"` // PII encrypted
	PasswordHash       pgtype.Text        `db:"password_hash"`   // Never expose password hashes
	FirstName          pgtype.Text        `db:"first_name"`
	EncryptedFirstName pgtype.Text        `db:"encrypted_first_name"` // PII encrypted
	LastName           pgtype.Text        `db:"last_name"`
	EncryptedLastName  pgtype.Text        `db:"encrypted_last_name"` // PII encrypted
	AvatarUrl          pgtype.Text        `db:"avatar_url"`
	IsVerified         pgtype.Bool        `db:"is_verified"`
	CreatedAt          pgtype.Timestamptz `db:"created_at"`
	UpdatedAt          pgtype.Timestamptz `db:"updated_at"`
	LastLoginAt        pgtype.Timestamptz `db:"last_login_at"`
	DeactivatedAt      pgtype.Timestamptz `db:"deactivated_at"`
}
