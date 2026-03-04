package config

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"os"
	"strconv"
	"strings"
)

// Config holds the application configuration
type Config struct {
	ServerHost           string
	ServerPort           int
	ServerEnv            string
	DatabaseURL          string
	DatabaseMaxConns     int
	JWTSecret            string //nolint:gosec // Field name matches config key, value loaded from env
	JWTExpiryHours       int
	AWSRegion            string
	AWSAccessKeyID       string
	AWSSecretAccessKey   string
	AWSS3BucketName      string
	CorsAllowedOrigins   []string
	RedisAddr            string
	RedisPassword        string
	RedisDB              int
	CacheTTLMinutes      int
	AnalyticsEnabled     bool
	EncryptionDataKey    string
	KMSKeyID             string
	GoogleClientID       string
	GoogleClientSecret   string
	FacebookClientID     string
	FacebookClientSecret string
	OAuthRedirectURL     string
	OAuthHTTPTimeout     int // Timeout in seconds for OAuth HTTP requests
}

// Load loads the configuration from environment variables
func Load() *Config {
	serverEnv := getEnvOrDefault("SERVER_ENV", "development")
	jwtSecret := getEnvOrDefault("JWT_SECRET", "")

	// Require JWT_SECRET in non-development environments
	if serverEnv != "development" && jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set in production environments")
	}

	// Generate a random JWT secret for development if not provided
	if jwtSecret == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			log.Fatal("Failed to generate random JWT secret:", err)
		}
		jwtSecret = base64.StdEncoding.EncodeToString(b)
		log.Println("WARNING: Using generated temporary JWT secret for development. Set JWT_SECRET for persistence.")
	}

	return &Config{
		ServerHost:         getEnvOrDefault("SERVER_HOST", "localhost"),
		ServerPort:         getIntEnvOrDefault("SERVER_PORT", 8080),
		ServerEnv:          serverEnv,
		DatabaseURL:        getEnvOrDefault("DATABASE_URL", "postgres://user:password@localhost:5432/wishlist_db?sslmode=disable"),
		DatabaseMaxConns:   getIntEnvOrDefault("DATABASE_MAX_CONNECTIONS", 20),
		JWTSecret:          jwtSecret,
		JWTExpiryHours:     getIntEnvOrDefault("JWT_EXPIRY_HOURS", 24),
		AWSRegion:          getEnvOrDefault("AWS_REGION", "us-east-1"),
		AWSAccessKeyID:     getEnvOrDefault("AWS_ACCESS_KEY_ID", ""),
		AWSSecretAccessKey: getEnvOrDefault("AWS_SECRET_ACCESS_KEY", ""),
		AWSS3BucketName:    getEnvOrDefault("AWS_S3_BUCKET_NAME", ""),
		CorsAllowedOrigins: getSliceEnvOrDefault("CORS_ALLOWED_ORIGINS", []string{
			"https://9art.ru",        // Production frontend
			"https://www.9art.ru",    // Production frontend (www)
			"https://lk.9art.ru",     // Personal office (mobile)
			"http://localhost:3000",  // Frontend (dev)
			"http://localhost:19006", // Expo (dev)
			"http://localhost:19000", // Expo (dev)
			"http://localhost:8081",  // React Native Metro (dev)
			"http://127.0.0.1:19006", // Expo alternative (dev)
			"http://127.0.0.1:8081",  // React Native Metro alternative (dev)
		}),
		RedisAddr:            getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		RedisPassword:        getEnvOrDefault("REDIS_PASSWORD", ""),
		RedisDB:              getIntEnvOrDefault("REDIS_DB", 0),
		CacheTTLMinutes:      getIntEnvOrDefault("CACHE_TTL_MINUTES", 15),
		AnalyticsEnabled:     getBoolEnvOrDefault("ANALYTICS_ENABLED", true),
		EncryptionDataKey:    getEnvOrDefault("ENCRYPTION_DATA_KEY", ""),
		KMSKeyID:             getEnvOrDefault("KMS_KEY_ID", ""),
		GoogleClientID:       getEnvOrDefault("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:   getEnvOrDefault("GOOGLE_CLIENT_SECRET", ""),
		FacebookClientID:     getEnvOrDefault("FACEBOOK_CLIENT_ID", ""),
		FacebookClientSecret: getEnvOrDefault("FACEBOOK_CLIENT_SECRET", ""),
		OAuthRedirectURL:     getEnvOrDefault("OAUTH_REDIRECT_URL", "wishlistapp://oauth"),
		OAuthHTTPTimeout:     getIntEnvOrDefault("OAUTH_HTTP_TIMEOUT", 10),
	}
}

// getEnvOrDefault retrieves an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getIntEnvOrDefault retrieves an integer environment variable or returns a default value
func getIntEnvOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// getSliceEnvOrDefault retrieves a slice environment variable (comma-separated) or returns a default value
func getSliceEnvOrDefault(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		slice := strings.Split(value, ",")
		for i, v := range slice {
			slice[i] = strings.TrimSpace(v)
		}
		return slice
	}
	return defaultValue
}

// getBoolEnvOrDefault retrieves a boolean environment variable or returns a default value
func getBoolEnvOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}
