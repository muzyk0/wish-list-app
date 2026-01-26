package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds the application configuration
type Config struct {
	ServerHost         string
	ServerPort         int
	ServerEnv          string
	DatabaseURL        string
	DatabaseMaxConns   int
	JWTSecret          string
	JWTExpiryHours     int
	AWSRegion          string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSS3BucketName    string
	CorsAllowedOrigins []string
	RedisAddr          string
	RedisPassword      string
	RedisDB            int
	CacheTTLMinutes    int
	AnalyticsEnabled   bool
	EncryptionDataKey  string
	KMSKeyID           string
}

// Load loads the configuration from environment variables
func Load() *Config {
	serverEnv := getEnvOrDefault("SERVER_ENV", "development")
	jwtSecret := getEnvOrDefault("JWT_SECRET", "")

	// Require JWT_SECRET in non-development environments
	if serverEnv != "development" && jwtSecret == "" {
		panic("JWT_SECRET must be set in production environments")
	}

	// Use a dev-only default for JWT_SECRET if not set in development
	if jwtSecret == "" {
		jwtSecret = "dev-only-secret-change-in-production"
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
		CorsAllowedOrigins: getSliceEnvOrDefault("CORS_ALLOWED_ORIGINS", []string{"http://localhost:3000", "http://localhost:19006"}),
		RedisAddr:          getEnvOrDefault("REDIS_ADDR", "localhost:6379"),
		RedisPassword:      getEnvOrDefault("REDIS_PASSWORD", ""),
		RedisDB:            getIntEnvOrDefault("REDIS_DB", 0),
		CacheTTLMinutes:    getIntEnvOrDefault("CACHE_TTL_MINUTES", 15),
		AnalyticsEnabled:   getBoolEnvOrDefault("ANALYTICS_ENABLED", true),
		EncryptionDataKey:  getEnvOrDefault("ENCRYPTION_DATA_KEY", ""),
		KMSKeyID:           getEnvOrDefault("KMS_KEY_ID", ""),
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
