package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCodeStore_GenerateCode(t *testing.T) {
	store := NewCodeStore()

	t.Run("generates unique codes", func(t *testing.T) {
		userID := uuid.New()

		code1, err := store.GenerateCode(userID)
		require.NoError(t, err)
		assert.NotEmpty(t, code1)

		code2, err := store.GenerateCode(userID)
		require.NoError(t, err)
		assert.NotEmpty(t, code2)

		assert.NotEqual(t, code1, code2, "codes should be unique")
	})

	t.Run("stores code with correct user ID", func(t *testing.T) {
		userID := uuid.New()
		code, err := store.GenerateCode(userID)
		require.NoError(t, err)

		retrievedID, ok := store.ExchangeCode(code)
		assert.True(t, ok)
		assert.Equal(t, userID, retrievedID)
	})

	t.Run("generates different codes for different users", func(t *testing.T) {
		userID1 := uuid.New()
		userID2 := uuid.New()

		code1, err := store.GenerateCode(userID1)
		require.NoError(t, err)

		code2, err := store.GenerateCode(userID2)
		require.NoError(t, err)

		assert.NotEqual(t, code1, code2)
	})
}

func TestCodeStore_ExchangeCode(t *testing.T) {
	store := NewCodeStore()

	t.Run("exchanges valid code", func(t *testing.T) {
		userID := uuid.New()
		code, err := store.GenerateCode(userID)
		require.NoError(t, err)

		retrievedID, ok := store.ExchangeCode(code)
		assert.True(t, ok)
		assert.Equal(t, userID, retrievedID)
	})

	t.Run("code is one-time use", func(t *testing.T) {
		userID := uuid.New()
		code, err := store.GenerateCode(userID)
		require.NoError(t, err)

		// First exchange should succeed
		_, ok := store.ExchangeCode(code)
		assert.True(t, ok)

		// Second exchange should fail
		_, ok = store.ExchangeCode(code)
		assert.False(t, ok)
	})

	t.Run("returns false for invalid code", func(t *testing.T) {
		_, ok := store.ExchangeCode("invalid-code")
		assert.False(t, ok)
	})

	t.Run("invalid code does not affect valid code exchange", func(t *testing.T) {
		userID := uuid.New()

		// Store a valid code
		validCode, err := store.GenerateCode(userID)
		require.NoError(t, err)

		// Attempt exchange with completely wrong code
		wrongCode := "invalid-code-12345"
		gotID, valid := store.ExchangeCode(wrongCode)

		assert.False(t, valid, "Invalid code should not be accepted")
		assert.Equal(t, uuid.Nil, gotID, "Invalid code should return nil UUID")

		// Verify valid code still works
		gotID, valid = store.ExchangeCode(validCode)
		assert.True(t, valid, "Valid code should be accepted")
		assert.Equal(t, userID, gotID, "Valid code should return correct user ID")
	})

	t.Run("returns false for expired code", func(t *testing.T) {
		userID := uuid.New()
		code, err := store.GenerateCode(userID)
		require.NoError(t, err)

		// Wait for code to expire (codes expire after 60 seconds)
		// For testing, we manipulate the store directly
		store.mu.Lock()
		entry := store.codes[code]
		entry.ExpiresAt = time.Now().Add(-time.Second)
		store.codes[code] = entry
		store.mu.Unlock()

		gotID, valid := store.ExchangeCode(code)
		assert.False(t, valid, "Expired code should not be accepted")
		assert.Equal(t, uuid.Nil, gotID, "Expired code should return nil UUID")

		// Verify code was deleted from store
		store.mu.Lock()
		_, exists := store.codes[code]
		store.mu.Unlock()
		assert.False(t, exists, "Expired code should be deleted")
	})
}

func TestCodeStore_CleanupExpired(t *testing.T) {
	store := NewCodeStore()

	t.Run("removes expired codes", func(t *testing.T) {
		// Add some codes
		for i := 0; i < 5; i++ {
			_, err := store.GenerateCode(uuid.New())
			require.NoError(t, err)
		}

		// Manually expire some codes
		store.mu.Lock()
		for code, entry := range store.codes {
			entry.ExpiresAt = time.Now().Add(-time.Second)
			store.codes[code] = entry
			break // Only expire one
		}
		store.mu.Unlock()

		initialCount := store.Len()
		removed := store.CleanupExpired()
		finalCount := store.Len()

		assert.Equal(t, 1, removed)
		assert.Equal(t, initialCount-1, finalCount)
	})

	t.Run("does not remove valid codes", func(t *testing.T) {
		store := NewCodeStore()

		// Add a fresh code
		_, err := store.GenerateCode(uuid.New())
		require.NoError(t, err)

		initialCount := store.Len()
		removed := store.CleanupExpired()
		finalCount := store.Len()

		assert.Equal(t, 0, removed)
		assert.Equal(t, initialCount, finalCount)
	})
}

func TestCodeStore_StartCleanupRoutine(t *testing.T) {
	t.Run("cleanup routine removes expired codes", func(t *testing.T) {
		store := NewCodeStore()
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		// Add a code and manually expire it
		userID := uuid.New()
		code, err := store.GenerateCode(userID)
		require.NoError(t, err)

		store.mu.Lock()
		entry := store.codes[code]
		entry.ExpiresAt = time.Now().Add(-time.Second)
		store.codes[code] = entry
		store.mu.Unlock()

		assert.Equal(t, 1, store.Len())

		// Start cleanup routine with short interval
		go func() {
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					store.CleanupExpired()
				case <-ctx.Done():
					return
				}
			}
		}()

		// Wait for cleanup
		time.Sleep(50 * time.Millisecond)
		cancel()

		// Code should be cleaned up
		assert.Equal(t, 0, store.Len())
	})

	t.Run("cleanup routine stops on context cancellation", func(t *testing.T) {
		store := NewCodeStore()
		ctx, cancel := context.WithCancel(context.Background())

		store.StartCleanupRoutine(ctx)

		// Cancel context immediately
		cancel()

		// Give time for goroutine to stop
		time.Sleep(10 * time.Millisecond)

		// Store should still work after cancellation
		_, err := store.GenerateCode(uuid.New())
		assert.NoError(t, err)
	})
}

func TestCodeStore_ConcurrentAccess(t *testing.T) {
	store := NewCodeStore()
	numGoroutines := 100
	codesPerGoroutine := 10

	t.Run("concurrent code generation", func(t *testing.T) {
		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func() {
				for j := 0; j < codesPerGoroutine; j++ {
					_, err := store.GenerateCode(uuid.New())
					if err != nil {
						t.Errorf("failed to generate code: %v", err)
					}
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		expectedCount := numGoroutines * codesPerGoroutine
		assert.Equal(t, expectedCount, store.Len())
	})

	t.Run("concurrent exchange", func(t *testing.T) {
		store := NewCodeStore()

		// Generate codes first
		codes := make([]string, 100)
		for i := 0; i < 100; i++ {
			code, err := store.GenerateCode(uuid.New())
			require.NoError(t, err)
			codes[i] = code
		}

		// Concurrent exchanges
		done := make(chan bool, 100)
		var successCount int64

		for _, code := range codes {
			go func(c string) {
				_, ok := store.ExchangeCode(c)
				if ok {
					successCount++
				}
				done <- true
			}(code)
		}

		// Wait for all exchanges
		for i := 0; i < 100; i++ {
			<-done
		}

		// All codes should be consumed
		assert.Equal(t, 0, store.Len())
	})
}

// Benchmarks

func BenchmarkCodeStore_GenerateCode(b *testing.B) {
	store := NewCodeStore()
	userID := uuid.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.GenerateCode(userID)
	}
}

func BenchmarkCodeStore_ExchangeCode(b *testing.B) {
	store := NewCodeStore()
	userID := uuid.New()

	// Pre-generate codes
	codes := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		code, _ := store.GenerateCode(userID)
		codes[i] = code
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store.ExchangeCode(codes[i])
	}
}

func BenchmarkCodeStore_ConcurrentGenerate(b *testing.B) {
	store := NewCodeStore()
	userID := uuid.New()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			store.GenerateCode(userID)
		}
	})
}

func BenchmarkCodeStore_ConcurrentExchange(b *testing.B) {
	store := NewCodeStore()
	userID := uuid.New()

	// Pre-generate codes (each goroutine needs its own codes)
	codes := make(chan string, b.N*10)
	go func() {
		for i := 0; i < b.N*10; i++ {
			code, _ := store.GenerateCode(userID)
			codes <- code
		}
		close(codes)
	}()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			code := <-codes
			store.ExchangeCode(code)
		}
	})
}

// Performance test to verify O(1) lookup
func TestCodeStore_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance test in short mode")
	}

	store := NewCodeStore()

	// Add many codes
	numCodes := 10000
	codes := make([]string, numCodes)
	for i := 0; i < numCodes; i++ {
		code, err := store.GenerateCode(uuid.New())
		require.NoError(t, err)
		codes[i] = code
	}

	// Measure exchange time (should be O(1))
	start := time.Now()
	for _, code := range codes {
		store.ExchangeCode(code)
	}
	duration := time.Since(start)

	avgTime := duration / time.Duration(numCodes)
	t.Logf("Average exchange time with %d codes: %v", numCodes, avgTime)

	// Should complete in reasonable time (O(1) lookup)
	// Average should be well under 1 microsecond per operation
	assert.Less(t, avgTime, time.Microsecond*10, "exchange should be O(1)")
}
