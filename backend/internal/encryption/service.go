package encryption

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

var (
	// ErrInvalidCiphertext is returned when decryption fails due to invalid ciphertext
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	// ErrInvalidKeySize is returned when the encryption key size is incorrect
	ErrInvalidKeySize = errors.New("encryption key must be 32 bytes for AES-256")
)

// Service provides field-level encryption for PII data using AES-256-GCM
type Service struct {
	dataKey []byte // 32-byte key for AES-256
	gcm     cipher.AEAD
}

// NewService creates a new encryption service with the provided data key
// The key must be 32 bytes for AES-256 encryption
func NewService(dataKey []byte) (*Service, error) {
	if len(dataKey) != 32 {
		return nil, ErrInvalidKeySize
	}

	block, err := aes.NewCipher(dataKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create AES cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &Service{
		dataKey: dataKey,
		gcm:     gcm,
	}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM and returns base64-encoded ciphertext
// The ciphertext includes the nonce prepended to the encrypted data
func (s *Service) Encrypt(ctx context.Context, plaintext string) (string, error) {
	if plaintext == "" {
		return "", nil
	}

	// Generate a random nonce
	nonce := make([]byte, s.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt the plaintext
	// The nonce is prepended to the ciphertext
	ciphertext := s.gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Return base64-encoded ciphertext
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts base64-encoded ciphertext using AES-256-GCM
func (s *Service) Decrypt(ctx context.Context, ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	// Decode base64
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	nonceSize := s.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", ErrInvalidCiphertext
	}

	// Extract nonce and ciphertext
	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]

	// Decrypt
	plaintext, err := s.gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", fmt.Errorf("%w", ErrInvalidCiphertext)
	}

	return string(plaintext), nil
}

// EncryptFields encrypts multiple fields in a single call
func (s *Service) EncryptFields(ctx context.Context, fields map[string]string) (map[string]string, error) {
	encrypted := make(map[string]string, len(fields))
	for key, value := range fields {
		encryptedValue, err := s.Encrypt(ctx, value)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt field %s: %w", key, err)
		}
		encrypted[key] = encryptedValue
	}
	return encrypted, nil
}

// DecryptFields decrypts multiple fields in a single call
func (s *Service) DecryptFields(ctx context.Context, fields map[string]string) (map[string]string, error) {
	decrypted := make(map[string]string, len(fields))
	for key, value := range fields {
		decryptedValue, err := s.Decrypt(ctx, value)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt field %s: %w", key, err)
		}
		decrypted[key] = decryptedValue
	}
	return decrypted, nil
}
