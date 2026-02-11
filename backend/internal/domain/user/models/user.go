package models

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type User struct {
	ID                 pgtype.UUID        `db:"id" json:"id"`
	Email              string             `db:"email" json:"email"`
	EncryptedEmail     pgtype.Text        `db:"encrypted_email" json:"-"` // PII encrypted
	PasswordHash       pgtype.Text        `db:"password_hash" json:"-"`   // Never expose password hashes
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
