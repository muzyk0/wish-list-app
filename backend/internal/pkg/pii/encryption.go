// Package pii provides utilities for encrypting and decrypting Personally Identifiable Information (PII)
// using field-level encryption. This package centralizes PII handling to ensure consistent
// encryption practices across the codebase.
package pii

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"wish-list/internal/pkg/encryption"
)

// FieldEncryptor provides methods for encrypting and decrypting PII fields
type FieldEncryptor struct {
	svc     *encryption.Service
	enabled bool
}

// NewFieldEncryptor creates a new FieldEncryptor instance
func NewFieldEncryptor(svc *encryption.Service) *FieldEncryptor {
	return &FieldEncryptor{
		svc:     svc,
		enabled: svc != nil,
	}
}

// EncryptField encrypts a string field and returns a pgtype.Text suitable for database storage
func (f *FieldEncryptor) EncryptField(ctx context.Context, value string) (pgtype.Text, error) {
	if !f.enabled || f.svc == nil || value == "" {
		return pgtype.Text{Valid: false}, nil
	}

	encrypted, err := f.svc.Encrypt(ctx, value)
	if err != nil {
		return pgtype.Text{Valid: false}, fmt.Errorf("failed to encrypt field: %w", err)
	}

	return pgtype.Text{String: encrypted, Valid: true}, nil
}

// DecryptField decrypts a pgtype.Text field and returns the plaintext string
func (f *FieldEncryptor) DecryptField(ctx context.Context, encrypted pgtype.Text) (string, error) {
	if !f.enabled || f.svc == nil || !encrypted.Valid {
		return "", nil
	}

	decrypted, err := f.svc.Decrypt(ctx, encrypted.String)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt field: %w", err)
	}

	return decrypted, nil
}

// EncryptOptionalField encrypts an optional string field that may be empty
func (f *FieldEncryptor) EncryptOptionalField(ctx context.Context, value string) (pgtype.Text, error) {
	if value == "" {
		return pgtype.Text{Valid: false}, nil
	}
	return f.EncryptField(ctx, value)
}

// EncryptToString encrypts a string and returns the encrypted string (not wrapped in pgtype.Text)
func (f *FieldEncryptor) EncryptToString(ctx context.Context, value string) (string, error) {
	if !f.enabled || f.svc == nil || value == "" {
		return value, nil
	}

	encrypted, err := f.svc.Encrypt(ctx, value)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt value: %w", err)
	}

	return encrypted, nil
}

// DecryptToString decrypts an encrypted string
func (f *FieldEncryptor) DecryptToString(ctx context.Context, encrypted string) (string, error) {
	if !f.enabled || f.svc == nil || encrypted == "" {
		return encrypted, nil
	}

	decrypted, err := f.svc.Decrypt(ctx, encrypted)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt value: %w", err)
	}

	return decrypted, nil
}

// IsEnabled returns whether encryption is enabled
func (f *FieldEncryptor) IsEnabled() bool {
	return f.enabled
}
