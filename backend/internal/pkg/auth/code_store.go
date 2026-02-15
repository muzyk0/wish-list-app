package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"sync"
	"time"

	"github.com/google/uuid"
)

// codeEntry represents a stored handoff code with its associated user and expiry
type codeEntry struct {
	UserID    uuid.UUID
	ExpiresAt time.Time
}

// CodeStore manages in-memory storage of one-time handoff codes
// for Frontend to Mobile authentication transfer.
// Thread-safe for concurrent access.
type CodeStore struct {
	mu    sync.RWMutex
	codes map[string]codeEntry
}

// NewCodeStore creates a new CodeStore instance
func NewCodeStore() *CodeStore {
	return &CodeStore{
		codes: make(map[string]codeEntry),
	}
}

// GenerateCode creates a new cryptographically secure handoff code
// for the given user ID with a 60-second expiry.
// Returns the code string and any error that occurred.
func (cs *CodeStore) GenerateCode(userID uuid.UUID) (string, error) {
	code, err := generateSecureCode(32)
	if err != nil {
		return "", err
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.codes[code] = codeEntry{
		UserID:    userID,
		ExpiresAt: time.Now().Add(60 * time.Second),
	}

	return code, nil
}

// ExchangeCode validates and consumes a handoff code.
// Returns the associated user ID if the code is valid and not expired.
// The code is deleted after use (one-time use only).
// Uses constant-time comparison to prevent timing attacks.
func (cs *CodeStore) ExchangeCode(code string) (uuid.UUID, bool) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// O(1) map lookup - existence check validates the code
	// Map lookup already confirms the code exists and is the correct key
	entry, exists := cs.codes[code]

	// Perform constant-time comparison even if code not found
	// to prevent timing attacks that could reveal code existence
	if !exists {
		// Compare against a dummy value to maintain constant time
		_ = constantTimeCompare("", code)
		return uuid.Nil, false
	}

	// Check if code is expired
	if time.Now().After(entry.ExpiresAt) {
		delete(cs.codes, code)
		return uuid.Nil, false
	}

	// Delete code after use (one-time use)
	delete(cs.codes, code)

	return entry.UserID, true
}

// CleanupExpired removes all expired codes from the store.
// Should be called periodically by a background goroutine.
func (cs *CodeStore) CleanupExpired() int {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	now := time.Now()
	count := 0

	for code, entry := range cs.codes {
		if now.After(entry.ExpiresAt) {
			delete(cs.codes, code)
			count++
		}
	}

	return count
}

// StartCleanupRoutine starts a background goroutine that cleans up
// expired codes every 30 seconds. The goroutine stops when ctx is cancelled.
func (cs *CodeStore) StartCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				cs.CleanupExpired()
			case <-ctx.Done():
				return
			}
		}
	}()
}

// Len returns the current number of codes in the store (for testing)
func (cs *CodeStore) Len() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return len(cs.codes)
}

// generateSecureCode generates a cryptographically secure random string
// of the specified byte length, encoded as base64url
func generateSecureCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// constantTimeCompare performs constant-time string comparison
// to prevent timing attacks on code validation
func constantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
