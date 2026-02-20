package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthRateLimiter_Allow(t *testing.T) {
	config := RateLimitConfig{
		Requests:  2,
		Window:    time.Second,
		BurstSize: 3,
	}

	limiter := NewAuthRateLimiter(config)

	t.Run("allows requests within burst limit", func(t *testing.T) {
		identifier := "test-client-1"

		// First 3 requests should be allowed (burst size)
		assert.True(t, limiter.Allow(identifier), "first request should be allowed")
		assert.True(t, limiter.Allow(identifier), "second request should be allowed")
		assert.True(t, limiter.Allow(identifier), "third request should be allowed")
	})

	t.Run("blocks requests exceeding burst limit", func(t *testing.T) {
		identifier := "test-client-2"

		// Use up burst limit
		for range config.BurstSize {
			limiter.Allow(identifier)
		}

		// Next request should be blocked
		assert.False(t, limiter.Allow(identifier), "request exceeding burst should be blocked")
	})

	t.Run("allows requests from different identifiers", func(t *testing.T) {
		limiter := NewAuthRateLimiter(config)

		// Use up burst for first client
		for range config.BurstSize {
			limiter.Allow("client-a")
		}

		// Different client should still be allowed
		assert.True(t, limiter.Allow("client-b"), "different client should be allowed")
	})

	t.Run("resets after window expires", func(t *testing.T) {
		shortConfig := RateLimitConfig{
			Requests:  1,
			Window:    50 * time.Millisecond,
			BurstSize: 1,
		}
		limiter := NewAuthRateLimiter(shortConfig)
		identifier := "test-client-reset"

		// Use up the limit
		limiter.Allow(identifier)
		assert.False(t, limiter.Allow(identifier), "should be blocked immediately after")

		// Wait for window to expire
		time.Sleep(60 * time.Millisecond)

		// Should be allowed again
		assert.True(t, limiter.Allow(identifier), "should be allowed after window reset")
	})
}

func TestAuthRateLimiter_Remaining(t *testing.T) {
	config := RateLimitConfig{
		Requests:  5,
		Window:    time.Minute,
		BurstSize: 10,
	}
	limiter := NewAuthRateLimiter(config)
	identifier := "test-remaining"

	t.Run("returns full burst size for new identifier", func(t *testing.T) {
		assert.Equal(t, config.BurstSize, limiter.Remaining("new-client"))
	})

	t.Run("decreases with each request", func(t *testing.T) {
		limiter.Allow(identifier)
		assert.Equal(t, 9, limiter.Remaining(identifier))

		limiter.Allow(identifier)
		assert.Equal(t, 8, limiter.Remaining(identifier))
	})

	t.Run("returns zero when limit exceeded", func(t *testing.T) {
		// Use up all burst
		for range config.BurstSize {
			limiter.Allow(identifier)
		}

		assert.Equal(t, 0, limiter.Remaining(identifier))
	})
}

func TestAuthRateLimiter_Reset(t *testing.T) {
	config := RateLimitConfig{
		Requests:  2,
		Window:    time.Minute,
		BurstSize: 3,
	}
	limiter := NewAuthRateLimiter(config)
	identifier := "test-reset"

	// Use up burst
	for range config.BurstSize {
		limiter.Allow(identifier)
	}

	// Should be blocked
	assert.False(t, limiter.Allow(identifier))

	// Reset
	limiter.Reset(identifier)

	// Should be allowed again
	assert.True(t, limiter.Allow(identifier))
}

func TestAuthRateLimitMiddleware(t *testing.T) {
	config := RateLimitConfig{
		Requests:  1,
		Window:    time.Minute,
		BurstSize: 2,
	}
	limiter := NewAuthRateLimiter(config)

	e := echo.New()
	middleware := AuthRateLimitMiddleware(limiter, func(c echo.Context) string {
		return c.RealIP()
	})

	// Create test handler
	handler := middleware(func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	t.Run("allows requests within limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "2", rec.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "1", rec.Header().Get("X-RateLimit-Remaining"))
	})

	t.Run("returns 429 when rate limited", func(t *testing.T) {
		// Use up burst
		for range config.BurstSize {
			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			_ = handler(c)
		}

		// Next request should be rate limited
		req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		require.NoError(t, err)
		assert.Equal(t, http.StatusTooManyRequests, rec.Code)
		assert.Contains(t, rec.Body.String(), "Too many requests")
	})
}

func TestNewAuthRateLimiterWithContext(t *testing.T) {
	config := RateLimitConfig{
		Requests:  10,
		Window:    time.Minute,
		BurstSize: 10,
	}

	t.Run("cleanup goroutine exits on context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		limiter := NewAuthRateLimiterWithContext(ctx, config)

		// Add some entries
		for i := range 5 {
			limiter.Allow(string(rune('a' + i)))
		}

		// Cancel context
		cancel()

		// Give goroutine time to exit
		time.Sleep(10 * time.Millisecond)

		// Limiter should still work after cancellation
		assert.True(t, limiter.Allow("new-client"))
	})
}

func TestRateLimitConfigs(t *testing.T) {
	t.Run("login rate limiter has correct config", func(t *testing.T) {
		limiter := NewLoginRateLimiter()
		assert.Equal(t, 5, limiter.config.Requests)
		assert.Equal(t, time.Minute, limiter.config.Window)
		assert.Equal(t, 10, limiter.config.BurstSize)
	})

	t.Run("exchange rate limiter has correct config", func(t *testing.T) {
		limiter := NewExchangeRateLimiter()
		assert.Equal(t, 10, limiter.config.Requests)
		assert.Equal(t, time.Minute, limiter.config.Window)
		assert.Equal(t, 15, limiter.config.BurstSize)
	})

	t.Run("OAuth rate limiter has correct config", func(t *testing.T) {
		limiter := NewOAuthRateLimiter()
		assert.Equal(t, 5, limiter.config.Requests)
		assert.Equal(t, time.Minute, limiter.config.Window)
		assert.Equal(t, 5, limiter.config.BurstSize)
	})
}

func TestIdentifierFunctions(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")

	t.Run("IPIdentifier extracts real IP", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		ip := IPIdentifier(c)
		assert.Equal(t, "192.168.1.1", ip)
	})

	t.Run("UserIdentifier extracts user ID when authenticated", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("userID", "user-123")

		id := UserIdentifier(c)
		assert.Equal(t, "user:user-123", id)
	})

	t.Run("UserIdentifier falls back to IP when not authenticated", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		id := UserIdentifier(c)
		assert.Equal(t, "ip:192.168.1.1", id)
	})
}

func TestAuthRateLimiter_ConcurrentAccess(t *testing.T) {
	config := RateLimitConfig{
		Requests:  100,
		Window:    time.Minute,
		BurstSize: 100,
	}
	limiter := NewAuthRateLimiter(config)
	identifier := "concurrent-test"

	// Run concurrent allows
	done := make(chan bool, 10)
	for range 10 {
		go func() {
			for range 10 {
				limiter.Allow(identifier)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for range 10 {
		<-done
	}

	// Check final count
	// Each goroutine did 10 requests, total 100
	// Burst is 100, so remaining should be 0
	assert.Equal(t, 0, limiter.Remaining(identifier))
}

func BenchmarkAuthRateLimiter_Allow(b *testing.B) {
	config := RateLimitConfig{
		Requests:  1000,
		Window:    time.Minute,
		BurstSize: 1000,
	}
	limiter := NewAuthRateLimiter(config)
	identifier := "bench-client"

	b.ResetTimer()
	for range b.N {
		limiter.Allow(identifier)
	}
}
