package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
)

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	// Requests is the maximum number of requests allowed in the window
	Requests int
	// Window is the time window for rate limiting
	Window time.Duration
	// BurstSize allows temporary burst above the rate (must be >= Requests)
	BurstSize int
}

// AuthRateLimits defines rate limits for different auth endpoints
var AuthRateLimits = struct {
	Login         RateLimitConfig
	Exchange      RateLimitConfig
	MobileHandoff RateLimitConfig
	Refresh       RateLimitConfig
}{
	Login:         RateLimitConfig{Requests: 5, Window: time.Minute, BurstSize: 10},
	Exchange:      RateLimitConfig{Requests: 10, Window: time.Minute, BurstSize: 15},
	MobileHandoff: RateLimitConfig{Requests: 10, Window: time.Minute, BurstSize: 15},
	Refresh:       RateLimitConfig{Requests: 20, Window: time.Minute, BurstSize: 30},
}

// rateLimitEntry tracks request count for a single identifier
type rateLimitEntry struct {
	count     int
	windowEnd time.Time
}

// AuthRateLimiter implements rate limiting for authentication endpoints
// using a sliding window algorithm with burst support.
type AuthRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateLimitEntry
	config  RateLimitConfig
}

// NewAuthRateLimiter creates a new rate limiter with the given configuration
func NewAuthRateLimiter(config RateLimitConfig) *AuthRateLimiter {
	limiter := &AuthRateLimiter{
		entries: make(map[string]*rateLimitEntry),
		config:  config,
	}

	// Start cleanup goroutine
	go limiter.cleanupLoop()

	return limiter
}

// Allow checks if the request from the given identifier should be allowed.
// Returns true if allowed, false if rate limited.
func (rl *AuthRateLimiter) Allow(identifier string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[identifier]

	if !exists || now.After(entry.windowEnd) {
		// New window
		rl.entries[identifier] = &rateLimitEntry{
			count:     1,
			windowEnd: now.Add(rl.config.Window),
		}
		return true
	}

	// Within existing window
	if entry.count < rl.config.BurstSize {
		entry.count++
		return true
	}

	return false
}

// Remaining returns the number of remaining requests for the identifier
func (rl *AuthRateLimiter) Remaining(identifier string) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	entry, exists := rl.entries[identifier]

	if !exists || now.After(entry.windowEnd) {
		return rl.config.BurstSize
	}

	remaining := rl.config.BurstSize - entry.count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Reset clears the rate limit for the given identifier
func (rl *AuthRateLimiter) Reset(identifier string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	delete(rl.entries, identifier)
}

// cleanupLoop periodically removes expired entries
func (rl *AuthRateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.cleanup()
	}
}

// cleanup removes expired entries
func (rl *AuthRateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	for key, entry := range rl.entries {
		if now.After(entry.windowEnd) {
			delete(rl.entries, key)
		}
	}
}

// AuthRateLimitMiddleware creates Echo middleware for rate limiting auth endpoints.
// The identifier function extracts the rate limit key (e.g., IP address or user ID).
func AuthRateLimitMiddleware(limiter *AuthRateLimiter, identifierFunc func(echo.Context) string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			identifier := identifierFunc(c)

			if !limiter.Allow(identifier) {
				return c.JSON(http.StatusTooManyRequests, map[string]any{
					"error":   "rate limit exceeded",
					"message": "Too many requests. Please try again later.",
				})
			}

			// Add rate limit headers
			remaining := limiter.Remaining(identifier)
			c.Response().Header().Set("X-RateLimit-Limit", strconv.Itoa(limiter.config.BurstSize))
			c.Response().Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))

			return next(c)
		}
	}
}

// IPIdentifier extracts the client IP address for rate limiting
func IPIdentifier(c echo.Context) string {
	return c.RealIP()
}

// UserIdentifier extracts the user ID from context for rate limiting.
// Falls back to IP address if user is not authenticated.
func UserIdentifier(c echo.Context) string {
	if userID := c.Get("userID"); userID != nil {
		if id, ok := userID.(string); ok && id != "" {
			return "user:" + id
		}
	}
	return "ip:" + c.RealIP()
}

// NewLoginRateLimiter creates a rate limiter configured for login endpoint
func NewLoginRateLimiter() *AuthRateLimiter {
	return NewAuthRateLimiter(AuthRateLimits.Login)
}

// NewExchangeRateLimiter creates a rate limiter configured for code exchange endpoint
func NewExchangeRateLimiter() *AuthRateLimiter {
	return NewAuthRateLimiter(AuthRateLimits.Exchange)
}

// NewHandoffRateLimiter creates a rate limiter configured for mobile handoff endpoint
func NewHandoffRateLimiter() *AuthRateLimiter {
	return NewAuthRateLimiter(AuthRateLimits.MobileHandoff)
}

// NewRefreshRateLimiter creates a rate limiter configured for refresh endpoint
func NewRefreshRateLimiter() *AuthRateLimiter {
	return NewAuthRateLimiter(AuthRateLimits.Refresh)
}
