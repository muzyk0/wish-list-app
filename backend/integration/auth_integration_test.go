//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"wish-list/internal/app/config"
	"wish-list/internal/app/database"
	authhttp "wish-list/internal/domain/auth/delivery/http"
	authdto "wish-list/internal/domain/auth/delivery/http/dto"
	userrepo "wish-list/internal/domain/user/repository"
	userservice "wish-list/internal/domain/user/service"
	"wish-list/internal/pkg/auth"
)

// setupTestServer creates a test server with required dependencies
func setupTestServer(t *testing.T) (*echo.Echo, func()) {
	e := echo.New()

	// Use test database
	db, err := database.NewDB("postgresql://test:test@localhost:5433/wishlist_test?sslmode=disable")
	if err != nil {
		t.Skip("Integration test requires database: ", err)
	}

	// Setup services
	userRepo := userrepo.NewUserRepository(db)
	userSvc := userservice.NewUserService(userRepo)
	tokenManager := auth.NewTokenManager("test-secret-key-for-testing-only")
	codeStore := auth.NewCodeStore()

	// Setup handlers
	authHandler := authhttp.NewHandler(userSvc, tokenManager, codeStore)

	// Register routes
	authGroup := e.Group("/api/auth")
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/refresh", authHandler.RefreshToken)

	cleanup := func() {
		db.Close()
	}

	return e, cleanup
}

func TestAuthFlow_EndToEnd(t *testing.T) {
	e, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("complete registration and login flow", func(t *testing.T) {
		// Step 1: Register a new user
		registerReq := authdto.RegisterRequest{
			Email:     "test@example.com",
			Password:  "SecurePass123!",
			FirstName: "Test",
			LastName:  "User",
		}

		body, _ := json.Marshal(registerReq)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusCreated, rec.Code)

		var registerResp authdto.AuthResponse
		err := json.Unmarshal(rec.Body.Bytes(), &registerResp)
		require.NoError(t, err)
		assert.NotEmpty(t, registerResp.AccessToken)
		assert.NotEmpty(t, registerResp.RefreshToken)

		// Step 2: Login with the same credentials
		loginReq := authdto.LoginRequest{
			Email:    "test@example.com",
			Password: "SecurePass123!",
		}

		body, _ = json.Marshal(loginReq)
		req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var loginResp authdto.AuthResponse
		err = json.Unmarshal(rec.Body.Bytes(), &loginResp)
		require.NoError(t, err)
		assert.NotEmpty(t, loginResp.AccessToken)
		assert.NotEmpty(t, loginResp.RefreshToken)

		// Step 3: Refresh the token
		refreshReq := authdto.RefreshRequest{
			RefreshToken: loginResp.RefreshToken,
		}

		body, _ = json.Marshal(refreshReq)
		req = httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var refreshResp authdto.AuthResponse
		err = json.Unmarshal(rec.Body.Bytes(), &refreshResp)
		require.NoError(t, err)
		assert.NotEmpty(t, refreshResp.AccessToken)
		assert.NotEmpty(t, refreshResp.RefreshToken)
	})
}

func TestAuthFlow_InvalidCredentials(t *testing.T) {
	e, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("login with wrong password", func(t *testing.T) {
		// First register a user
		registerReq := authdto.RegisterRequest{
			Email:    "test2@example.com",
			Password: "CorrectPass123!",
		}

		body, _ := json.Marshal(registerReq)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)

		// Try to login with wrong password
		loginReq := authdto.LoginRequest{
			Email:    "test2@example.com",
			Password: "WrongPass123!",
		}

		body, _ = json.Marshal(loginReq)
		req = httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusUnauthorized, rec.Code)
	})
}

func TestAuthFlow_DuplicateRegistration(t *testing.T) {
	e, cleanup := setupTestServer(t)
	defer cleanup()

	t.Run("prevent duplicate email registration", func(t *testing.T) {
		// Register first user
		registerReq := authdto.RegisterRequest{
			Email:    "duplicate@example.com",
			Password: "SecurePass123!",
		}

		body, _ := json.Marshal(registerReq)
		req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusCreated, rec.Code)

		// Try to register with same email
		req = httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec = httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusConflict, rec.Code)
	})
}

// TestTokenExpiry tests that expired tokens are rejected
func TestAuthFlow_TokenExpiry(t *testing.T) {
	// This test would require a shorter token expiry for testing
	// Skipping for now as it requires configuration changes
	t.Skip("Token expiry test requires configurable expiry times")
}

// TestRateLimiting tests that rate limiting is applied
func TestAuthFlow_RateLimiting(t *testing.T) {
	t.Skip("Rate limiting integration test requires running server")
}
