package pii

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"wish-list/internal/pkg/encryption"
)

func TestNewFieldEncryptor(t *testing.T) {
	tests := []struct {
		name    string
		svc     *encryption.Service
		wantNil bool
	}{
		{
			name:    "with encryption service",
			svc:     createTestEncryptionService(t),
			wantNil: false,
		},
		{
			name:    "without encryption service",
			svc:     nil,
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryptor := NewFieldEncryptor(tt.svc)
			if encryptor == nil {
				t.Errorf("NewFieldEncryptor() = nil, want non-nil")
			}
			if tt.svc != nil && !encryptor.enabled {
				t.Errorf("NewFieldEncryptor().enabled = false, want true")
			}
			if tt.svc == nil && encryptor.enabled {
				t.Errorf("NewFieldEncryptor().enabled = true, want false")
			}
		})
	}
}

func TestFieldEncryptor_EncryptField(t *testing.T) {
	ctx := context.Background()
	svc := createTestEncryptionService(t)
	encryptor := NewFieldEncryptor(svc)

	tests := []struct {
		name      string
		value     string
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "encrypt non-empty value",
			value:     "test@example.com",
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "empty value returns invalid",
			value:     "",
			wantValid: false,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := encryptor.EncryptField(ctx, tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("EncryptField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result.Valid != tt.wantValid {
				t.Errorf("EncryptField() Valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if tt.wantValid && result.String == "" {
				t.Errorf("EncryptField() returned empty string for valid encryption")
			}
		})
	}
}

func TestFieldEncryptor_DecryptField(t *testing.T) {
	ctx := context.Background()
	svc := createTestEncryptionService(t)
	encryptor := NewFieldEncryptor(svc)

	// First encrypt a value
	encrypted, _ := encryptor.EncryptField(ctx, "test@example.com")

	tests := []struct {
		name       string
		encrypted  pgtype.Text
		wantResult string
		wantErr    bool
	}{
		{
			name:       "decrypt valid encrypted value",
			encrypted:  encrypted,
			wantResult: "test@example.com",
			wantErr:    false,
		},
		{
			name:       "invalid text returns empty",
			encrypted:  pgtype.Text{Valid: false},
			wantResult: "",
			wantErr:    false,
		},
		{
			name:       "empty valid text returns empty",
			encrypted:  pgtype.Text{String: "", Valid: true},
			wantResult: "",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := encryptor.DecryptField(ctx, tt.encrypted)
			if (err != nil) != tt.wantErr {
				t.Errorf("DecryptField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.wantResult {
				t.Errorf("DecryptField() = %v, want %v", result, tt.wantResult)
			}
		})
	}
}

func TestFieldEncryptor_EncryptDecrypt_RoundTrip(t *testing.T) {
	ctx := context.Background()
	svc := createTestEncryptionService(t)
	encryptor := NewFieldEncryptor(svc)

	original := "sensitive-data@example.com"

	// Encrypt
	encrypted, err := encryptor.EncryptField(ctx, original)
	if err != nil {
		t.Fatalf("EncryptField() failed: %v", err)
	}
	if !encrypted.Valid {
		t.Fatal("EncryptField() returned invalid result")
	}

	// Decrypt
	decrypted, err := encryptor.DecryptField(ctx, encrypted)
	if err != nil {
		t.Fatalf("DecryptField() failed: %v", err)
	}

	if decrypted != original {
		t.Errorf("Round-trip failed: got %v, want %v", decrypted, original)
	}
}

func TestFieldEncryptor_EncryptOptionalField(t *testing.T) {
	ctx := context.Background()
	svc := createTestEncryptionService(t)
	encryptor := NewFieldEncryptor(svc)

	tests := []struct {
		name      string
		value     string
		wantValid bool
	}{
		{
			name:      "non-empty value",
			value:     "test",
			wantValid: true,
		},
		{
			name:      "empty value",
			value:     "",
			wantValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := encryptor.EncryptOptionalField(ctx, tt.value)
			if err != nil {
				t.Errorf("EncryptOptionalField() error = %v", err)
				return
			}
			if result.Valid != tt.wantValid {
				t.Errorf("EncryptOptionalField() Valid = %v, want %v", result.Valid, tt.wantValid)
			}
		})
	}
}

func TestFieldEncryptor_EncryptToString(t *testing.T) {
	ctx := context.Background()
	svc := createTestEncryptionService(t)
	encryptor := NewFieldEncryptor(svc)

	tests := []struct {
		name     string
		value    string
		wantSame bool // if true, expect same value back (when encryption disabled or empty)
	}{
		{
			name:     "encrypt value",
			value:    "test@example.com",
			wantSame: false,
		},
		{
			name:     "empty value",
			value:    "",
			wantSame: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := encryptor.EncryptToString(ctx, tt.value)
			if err != nil {
				t.Errorf("EncryptToString() error = %v", err)
				return
			}
			if tt.wantSame && result != tt.value {
				t.Errorf("EncryptToString() = %v, want %v", result, tt.value)
			}
			if !tt.wantSame && result == tt.value {
				t.Errorf("EncryptToString() returned unencrypted value")
			}
		})
	}
}

func TestFieldEncryptor_DecryptToString(t *testing.T) {
	ctx := context.Background()
	svc := createTestEncryptionService(t)
	encryptor := NewFieldEncryptor(svc)

	// Encrypt first
	original := "test@example.com"
	encrypted, _ := encryptor.EncryptToString(ctx, original)

	tests := []struct {
		name      string
		encrypted string
		want      string
	}{
		{
			name:      "decrypt valid encrypted",
			encrypted: encrypted,
			want:      original,
		},
		{
			name:      "empty string",
			encrypted: "",
			want:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := encryptor.DecryptToString(ctx, tt.encrypted)
			if err != nil {
				t.Errorf("DecryptToString() error = %v", err)
				return
			}
			if result != tt.want {
				t.Errorf("DecryptToString() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestFieldEncryptor_IsEnabled(t *testing.T) {
	tests := []struct {
		name string
		svc  *encryption.Service
		want bool
	}{
		{
			name: "with service",
			svc:  createTestEncryptionService(t),
			want: true,
		},
		{
			name: "without service",
			svc:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encryptor := NewFieldEncryptor(tt.svc)
			if got := encryptor.IsEnabled(); got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFieldEncryptor_DisabledEncryption(t *testing.T) {
	ctx := context.Background()
	encryptor := NewFieldEncryptor(nil) // Disabled

	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "encrypt with disabled encryption",
			value: "test@example.com",
		},
		{
			name:  "decrypt with disabled encryption",
			value: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When encryption is disabled, EncryptField should return invalid
			encrypted, err := encryptor.EncryptField(ctx, tt.value)
			if err != nil {
				t.Errorf("EncryptField() error = %v", err)
			}
			if encrypted.Valid {
				t.Errorf("EncryptField() Valid = true when encryption disabled, want false")
			}

			// DecryptField should return empty string without error
			decrypted, err := encryptor.DecryptField(ctx, pgtype.Text{String: tt.value, Valid: true})
			if err != nil {
				t.Errorf("DecryptField() error = %v", err)
			}
			if decrypted != "" {
				t.Errorf("DecryptField() = %v, want empty string", decrypted)
			}
		})
	}
}

// createTestEncryptionService creates an encryption service for testing
func createTestEncryptionService(t *testing.T) *encryption.Service {
	t.Helper()
	// Create a 32-byte key for AES-256
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	svc, err := encryption.NewService(key)
	if err != nil {
		t.Fatalf("Failed to create encryption service: %v", err)
	}

	return svc
}
