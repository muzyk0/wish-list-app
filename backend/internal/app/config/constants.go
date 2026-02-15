package config

import "time"

// Constants for the wish-list application
// These values can be overridden via environment variables where indicated

const (
	// Token and Session Expiration

	// HandoffCodeExpiry is the duration for which a mobile handoff code remains valid (60 seconds)
	// This is intentionally short-lived for security
	HandoffCodeExpiry = 60 * time.Second

	// GuestTokenExpiry is the duration for guest/anonymous tokens (24 hours)
	GuestTokenExpiry = 24 * time.Hour

	// RefreshTokenExpiry is the duration for refresh tokens (7 days)
	RefreshTokenExpiry = 7 * 24 * time.Hour

	// AccessTokenExpiry is the duration for access tokens (15 minutes)
	// Short-lived access tokens limit the window of compromise
	AccessTokenExpiry = 15 * time.Minute

	// Database Connection Pool Settings

	// DefaultMaxOpenConns is the default maximum number of open database connections
	DefaultMaxOpenConns = 25

	// DefaultMaxIdleConns is the default maximum number of idle database connections
	DefaultMaxIdleConns = 5

	// DefaultConnMaxLifetime is the default maximum lifetime of a database connection
	DefaultConnMaxLifetime = 5 * time.Minute

	// Query Limits

	// DefaultQueryLimit is the default limit for paginated queries
	DefaultQueryLimit = 100

	// MaxQueryLimit is the maximum allowed limit for paginated queries
	MaxQueryLimit = 1000

	// DefaultPaginationPage is the default page number for pagination
	DefaultPaginationPage = 1

	// DefaultPaginationLimit is the default number of items per page
	DefaultPaginationLimit = 10

	// MaxPaginationLimit is the maximum items per page allowed
	MaxPaginationLimit = 100

	// Rate Limiting

	// RateLimitWindow is the default time window for rate limiting
	RateLimitWindow = time.Minute

	// DefaultRateLimitRequests is the default number of requests allowed per window
	DefaultRateLimitRequests = 100

	// CodeStore Cleanup

	// CodeStoreCleanupInterval is the interval between cleanup runs for expired codes
	CodeStoreCleanupInterval = 30 * time.Second

	// HTTP Client Timeouts

	// DefaultOAuthHTTPTimeout is the default timeout for OAuth HTTP requests
	DefaultOAuthHTTPTimeout = 10 * time.Second

	// DefaultHTTPClientTimeout is the default timeout for general HTTP requests
	DefaultHTTPClientTimeout = 30 * time.Second

	// Security

	// MinPasswordLength is the minimum required password length
	MinPasswordLength = 8

	// MaxNameLength is the maximum allowed length for user names
	MaxNameLength = 100

	// MaxEmailLength is the maximum allowed length for email addresses
	MaxEmailLength = 255

	// Encryption

	// EncryptionKeySize is the required size for AES-256 encryption keys (32 bytes)
	EncryptionKeySize = 32

	// SecureCodeLength is the byte length for generating secure random codes
	SecureCodeLength = 32
)
