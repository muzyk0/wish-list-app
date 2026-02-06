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

// TestGenerateToken removed - replaced by TestGenerateAccessToken and TestGenerateRefreshToken

func TestValidateToken(t *testing.T) {
	secret := "test-secret"
	tm := NewTokenManager(secret)

	userID := "user-123"
	email := "test@example.com"
	userType := "user"

	tokenString, err := tm.GenerateAccessToken(userID, email, userType)
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
	tokenString, err := tm.GenerateAccessToken("user-123", "test@example.com", "user")
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

// Note: Old RefreshToken method removed in favor of cross-domain auth flow.
// Token refresh now handled by AuthHandler with rotation support.

func TestGenerateAccessToken(t *testing.T) {
	secret := "test-secret"
	tm := NewTokenManager(secret)

	userID := "user-123"
	email := "test@example.com"
	userType := "user"

	tokenString, err := tm.GenerateAccessToken(userID, email, userType)
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
	assert.Empty(t, claims.TokenID, "Access tokens should not have TokenID")
}

func TestGenerateRefreshToken(t *testing.T) {
	secret := "test-secret"
	tm := NewTokenManager(secret)

	userID := "user-123"
	email := "test@example.com"
	userType := "user"
	tokenID := "token-id-123"

	tokenString, err := tm.GenerateRefreshToken(userID, email, userType, tokenID)
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
	assert.Equal(t, tokenID, claims.TokenID)
	assert.Equal(t, "wish-list-app", claims.Issuer)
}
