package encryption

import (
	"context"
	"crypto/rand"
	"testing"
)

func TestNewService(t *testing.T) {
	t.Run("valid 32-byte key", func(t *testing.T) {
		key := make([]byte, 32)
		_, err := rand.Read(key)
		if err != nil {
			t.Fatalf("failed to generate random key: %v", err)
		}

		svc, err := NewService(key)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if svc == nil {
			t.Fatal("expected service to be non-nil")
		}
	})

	t.Run("invalid key size - too short", func(t *testing.T) {
		key := make([]byte, 16) // AES-128, not AES-256
		_, err := rand.Read(key)
		if err != nil {
			t.Fatalf("failed to generate random key: %v", err)
		}

		_, err = NewService(key)
		if err != ErrInvalidKeySize {
			t.Fatalf("expected ErrInvalidKeySize, got %v", err)
		}
	})

	t.Run("invalid key size - too long", func(t *testing.T) {
		key := make([]byte, 64)
		_, err := rand.Read(key)
		if err != nil {
			t.Fatalf("failed to generate random key: %v", err)
		}

		_, err = NewService(key)
		if err != ErrInvalidKeySize {
			t.Fatalf("expected ErrInvalidKeySize, got %v", err)
		}
	})
}

func TestEncryptDecrypt(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate random key: %v", err)
	}

	svc, err := NewService(key)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	ctx := context.Background()

	t.Run("encrypt and decrypt simple text", func(t *testing.T) {
		plaintext := "hello@example.com"

		ciphertext, err := svc.Encrypt(ctx, plaintext)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		if ciphertext == "" {
			t.Fatal("ciphertext should not be empty")
		}

		if ciphertext == plaintext {
			t.Fatal("ciphertext should not equal plaintext")
		}

		decrypted, err := svc.Decrypt(ctx, ciphertext)
		if err != nil {
			t.Fatalf("decrypt failed: %v", err)
		}

		if decrypted != plaintext {
			t.Fatalf("expected %q, got %q", plaintext, decrypted)
		}
	})

	t.Run("encrypt empty string", func(t *testing.T) {
		ciphertext, err := svc.Encrypt(ctx, "")
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		if ciphertext != "" {
			t.Fatal("ciphertext should be empty for empty plaintext")
		}
	})

	t.Run("decrypt empty string", func(t *testing.T) {
		plaintext, err := svc.Decrypt(ctx, "")
		if err != nil {
			t.Fatalf("decrypt failed: %v", err)
		}

		if plaintext != "" {
			t.Fatal("plaintext should be empty for empty ciphertext")
		}
	})

	t.Run("encrypt PII data", func(t *testing.T) {
		testCases := []struct {
			name      string
			plaintext string
		}{
			{"email", "user@example.com"},
			{"name with spaces", "John Doe"},
			{"unicode characters", "Владислав Петров"},
			{"special characters", "user+tag@example.com"},
			{"long text", "This is a very long piece of text that should still be encrypted and decrypted correctly without any issues"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				ciphertext, err := svc.Encrypt(ctx, tc.plaintext)
				if err != nil {
					t.Fatalf("encrypt failed: %v", err)
				}

				decrypted, err := svc.Decrypt(ctx, ciphertext)
				if err != nil {
					t.Fatalf("decrypt failed: %v", err)
				}

				if decrypted != tc.plaintext {
					t.Fatalf("expected %q, got %q", tc.plaintext, decrypted)
				}
			})
		}
	})

	t.Run("decrypt invalid ciphertext", func(t *testing.T) {
		_, err := svc.Decrypt(ctx, "invalid-base64!")
		if err == nil {
			t.Fatal("expected error for invalid base64")
		}
	})

	t.Run("decrypt tampered ciphertext", func(t *testing.T) {
		plaintext := "hello@example.com"
		ciphertext, err := svc.Encrypt(ctx, plaintext)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		// Tamper with the ciphertext by adding extra character
		tamperedCiphertext := ciphertext + "A"

		_, err = svc.Decrypt(ctx, tamperedCiphertext)
		if err == nil {
			t.Fatal("expected error for tampered ciphertext")
		}
	})

	t.Run("encrypt produces different ciphertext each time", func(t *testing.T) {
		plaintext := "hello@example.com"

		ciphertext1, err := svc.Encrypt(ctx, plaintext)
		if err != nil {
			t.Fatalf("encrypt 1 failed: %v", err)
		}

		ciphertext2, err := svc.Encrypt(ctx, plaintext)
		if err != nil {
			t.Fatalf("encrypt 2 failed: %v", err)
		}

		// Due to random nonce, ciphertexts should be different
		if ciphertext1 == ciphertext2 {
			t.Fatal("expected different ciphertexts for same plaintext (due to random nonce)")
		}

		// But both should decrypt to same plaintext
		decrypted1, err := svc.Decrypt(ctx, ciphertext1)
		if err != nil {
			t.Fatalf("decrypt 1 failed: %v", err)
		}

		decrypted2, err := svc.Decrypt(ctx, ciphertext2)
		if err != nil {
			t.Fatalf("decrypt 2 failed: %v", err)
		}

		if decrypted1 != plaintext || decrypted2 != plaintext {
			t.Fatalf("expected both to decrypt to %q, got %q and %q", plaintext, decrypted1, decrypted2)
		}
	})
}

func TestEncryptDecryptFields(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("failed to generate random key: %v", err)
	}

	svc, err := NewService(key)
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	ctx := context.Background()

	t.Run("encrypt and decrypt multiple fields", func(t *testing.T) {
		fields := map[string]string{
			"email":      "user@example.com",
			"first_name": "John",
			"last_name":  "Doe",
		}

		encrypted, err := svc.EncryptFields(ctx, fields)
		if err != nil {
			t.Fatalf("encrypt fields failed: %v", err)
		}

		if len(encrypted) != len(fields) {
			t.Fatalf("expected %d encrypted fields, got %d", len(fields), len(encrypted))
		}

		// Verify all fields are encrypted (different from original)
		for key, originalValue := range fields {
			encryptedValue, exists := encrypted[key]
			if !exists {
				t.Fatalf("encrypted field %s not found", key)
			}
			if encryptedValue == originalValue {
				t.Fatalf("encrypted value for %s should differ from original", key)
			}
		}

		// Decrypt all fields
		decrypted, err := svc.DecryptFields(ctx, encrypted)
		if err != nil {
			t.Fatalf("decrypt fields failed: %v", err)
		}

		// Verify decrypted matches original
		for key, originalValue := range fields {
			decryptedValue, exists := decrypted[key]
			if !exists {
				t.Fatalf("decrypted field %s not found", key)
			}
			if decryptedValue != originalValue {
				t.Fatalf("decrypted value for %s: expected %q, got %q", key, originalValue, decryptedValue)
			}
		}
	})

	t.Run("encrypt empty fields map", func(t *testing.T) {
		fields := map[string]string{}

		encrypted, err := svc.EncryptFields(ctx, fields)
		if err != nil {
			t.Fatalf("encrypt fields failed: %v", err)
		}

		if len(encrypted) != 0 {
			t.Fatalf("expected 0 encrypted fields, got %d", len(encrypted))
		}
	})

	t.Run("decrypt empty fields map", func(t *testing.T) {
		fields := map[string]string{}

		decrypted, err := svc.DecryptFields(ctx, fields)
		if err != nil {
			t.Fatalf("decrypt fields failed: %v", err)
		}

		if len(decrypted) != 0 {
			t.Fatalf("expected 0 decrypted fields, got %d", len(decrypted))
		}
	})
}

func TestKeyRotation(t *testing.T) {
	// Simulate key rotation scenario
	t.Run("decrypt with different key fails", func(t *testing.T) {
		// Create service with first key
		key1 := make([]byte, 32)
		_, err := rand.Read(key1)
		if err != nil {
			t.Fatalf("failed to generate key1: %v", err)
		}

		svc1, err := NewService(key1)
		if err != nil {
			t.Fatalf("failed to create service1: %v", err)
		}

		// Encrypt with first key
		ctx := context.Background()
		plaintext := "sensitive-data@example.com"
		ciphertext, err := svc1.Encrypt(ctx, plaintext)
		if err != nil {
			t.Fatalf("encrypt failed: %v", err)
		}

		// Create service with different key
		key2 := make([]byte, 32)
		_, err = rand.Read(key2)
		if err != nil {
			t.Fatalf("failed to generate key2: %v", err)
		}

		svc2, err := NewService(key2)
		if err != nil {
			t.Fatalf("failed to create service2: %v", err)
		}

		// Attempt to decrypt with second key should fail
		_, err = svc2.Decrypt(ctx, ciphertext)
		if err == nil {
			t.Fatal("expected error when decrypting with different key")
		}
	})
}
