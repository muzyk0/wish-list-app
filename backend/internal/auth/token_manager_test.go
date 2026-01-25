package auth

import (
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenManager(t *testing.T) {
	secret := "test-secret"
	tm := NewTokenManager(secret)

	assert.Equal(t, []byte(secret), tm.secret)
}

func TestGenerateToken(t *testing.T) {
	secret := "test-secret"
	tm := NewTokenManager(secret)

	userID := "user-123"
	email := "test@example.com"
	userType := "user"
	expiryHours := 1

	tokenString, err := tm.GenerateToken(userID, email, userType, expiryHours)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	require.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(*Claims)
	require.True(t, ok)

	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, userType, claims.UserType)
	assert.Equal(t, "wish-list-app", claims.Issuer)
}

func TestValidateToken(t *testing.T) {
	secret := "test-secret"
	tm := NewTokenManager(secret)

	userID := "user-123"
	email := "test@example.com"
	userType := "user"

	tokenString, err := tm.GenerateToken(userID, email, userType, 1)
	require.NoError(t, err)

	claims, err := tm.ValidateToken(tokenString)
	require.NoError(t, err)

	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, userType, claims.UserType)
	assert.Equal(t, "wish-list-app", claims.Issuer)
}

func TestValidateTokenInvalid(t *testing.T) {
	secret := "test-secret"
	tm := NewTokenManager(secret)

	// Test with invalid token
	_, err := tm.ValidateToken("invalid-token")
	require.Error(t, err)

	// Test with wrong secret
	wrongTm := NewTokenManager("wrong-secret")
	tokenString, err := tm.GenerateToken("user-123", "test@example.com", "user", 1)
	require.NoError(t, err)

	_, err = wrongTm.ValidateToken(tokenString)
	assert.Error(t, err)
}

func TestGenerateGuestToken(t *testing.T) {
	secret := "test-secret"
	tm := NewTokenManager(secret)

	guestID := "guest-123"
	guestName := "Guest User"
	guestEmail := "guest@example.com"

	tokenString, err := tm.GenerateGuestToken(guestID, guestName, guestEmail)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Parse and validate the token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	require.NoError(t, err)
	assert.True(t, token.Valid)

	claims, ok := token.Claims.(*Claims)
	require.True(t, ok)

	assert.Equal(t, guestID, claims.UserID)
	assert.Equal(t, guestEmail, claims.Email)
	assert.Equal(t, "guest", claims.UserType)
	assert.Equal(t, "wish-list-app", claims.Issuer)
}

func TestRefreshToken(t *testing.T) {
	secret := "test-secret"
	tm := NewTokenManager(secret)

	userID := "user-123"
	email := "test@example.com"
	userType := "user"

	// Generate initial token with 1 hour expiry
	initialToken, err := tm.GenerateToken(userID, email, userType, 1)
	require.NoError(t, err)

	// Refresh token with 2 hours expiry
	refreshedToken, err := tm.RefreshToken(initialToken, 2)
	require.NoError(t, err)
	assert.NotEqual(t, initialToken, refreshedToken)

	// Validate refreshed token
	claims, err := tm.ValidateToken(refreshedToken)
	require.NoError(t, err)

	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.Equal(t, userType, claims.UserType)
}
