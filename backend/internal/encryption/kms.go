package encryption

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
)

// KMSClient wraps AWS KMS operations for encryption key management
type KMSClient struct {
	client *kms.Client
	keyID  string
}

// NewKMSClient creates a new KMS client with AWS configuration
func NewKMSClient(ctx context.Context, keyID string) (*KMSClient, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	return &KMSClient{
		client: kms.NewFromConfig(cfg),
		keyID:  keyID,
	}, nil
}

// GenerateDataKey generates a new 256-bit data encryption key using KMS
// Returns the plaintext key (for immediate use) and encrypted key (for storage)
func (k *KMSClient) GenerateDataKey(ctx context.Context) (plaintextKey []byte, encryptedKey string, err error) {
	result, err := k.client.GenerateDataKey(ctx, &kms.GenerateDataKeyInput{
		KeyId:         &k.keyID,
		KeySpec:       "AES_256",
		NumberOfBytes: nil,
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to generate data key: %w", err)
	}

	encryptedKeyB64 := base64.StdEncoding.EncodeToString(result.CiphertextBlob)
	return result.Plaintext, encryptedKeyB64, nil
}

// DecryptDataKey decrypts an encrypted data key using KMS
func (k *KMSClient) DecryptDataKey(ctx context.Context, encryptedKey string) ([]byte, error) {
	encryptedKeyBytes, err := base64.StdEncoding.DecodeString(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encrypted key: %w", err)
	}

	result, err := k.client.Decrypt(ctx, &kms.DecryptInput{
		CiphertextBlob: encryptedKeyBytes,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt data key: %w", err)
	}

	return result.Plaintext, nil
}

// GetOrCreateDataKey retrieves the data key from environment or generates a new one
// For development, uses ENCRYPTION_DATA_KEY env var
// For production, should use KMS to generate and store encrypted keys
func GetOrCreateDataKey(ctx context.Context) ([]byte, error) {
	// Try to get from environment variable first (development/testing)
	if keyStr := os.Getenv("ENCRYPTION_DATA_KEY"); keyStr != "" {
		key, err := base64.StdEncoding.DecodeString(keyStr)
		if err != nil {
			return nil, fmt.Errorf("failed to decode ENCRYPTION_DATA_KEY: %w", err)
		}
		if len(key) != 32 {
			return nil, fmt.Errorf("ENCRYPTION_DATA_KEY must be 32 bytes (got %d)", len(key))
		}
		return key, nil
	}

	// For production, use KMS
	kmsKeyID := os.Getenv("KMS_KEY_ID")
	if kmsKeyID != "" {
		kmsClient, err := NewKMSClient(ctx, kmsKeyID)
		if err != nil {
			return nil, fmt.Errorf("failed to create KMS client: %w", err)
		}

		// Check if we have an encrypted data key stored
		encryptedDataKey := os.Getenv("ENCRYPTED_DATA_KEY")
		if encryptedDataKey != "" {
			// Decrypt existing key
			plaintextKey, err := kmsClient.DecryptDataKey(ctx, encryptedDataKey)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt data key: %w", err)
			}
			return plaintextKey, nil
		}

		// Generate new key if none exists
		plaintextKey, _, err := kmsClient.GenerateDataKey(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate data key: %w", err)
		}

		// Note: In production, store ENCRYPTED_DATA_KEY in secret manager
		// The encrypted key should be persisted externally, not logged
		return plaintextKey, nil
	}

	// Fallback: generate a random key (ONLY for development/testing)
	serverEnv := os.Getenv("SERVER_ENV")
	if serverEnv != "" && serverEnv != "development" {
		return nil, fmt.Errorf("no encryption key configured: set ENCRYPTION_DATA_KEY or KMS_KEY_ID for %s environment", serverEnv)
	}

	// Development-only fallback
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}
	// Do not log key material - development mode uses ephemeral key
	return key, nil
}
