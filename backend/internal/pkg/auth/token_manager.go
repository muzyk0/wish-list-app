package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims
type Claims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	UserType string `json:"user_type"`          // "user" or "guest"
	TokenID  string `json:"token_id,omitempty"` // For refresh tokens only (enables rotation/blacklisting)
	jwt.RegisteredClaims
}

// TokenManager handles JWT token operations
type TokenManager struct {
	secret []byte
}

// NewTokenManager creates a new TokenManager
func NewTokenManager(secret string) *TokenManager {
	return &TokenManager{
		secret: []byte(secret),
	}
}

// GenerateGuestToken generates a JWT token for guest users
func (tm *TokenManager) GenerateGuestToken(guestID, guestName, guestEmail string) (string, error) {
	claims := Claims{
		UserID:   guestID,
		Email:    guestEmail,
		UserType: "guest",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), // Guest tokens expire in 24 hours
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "wish-list-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(tm.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signedToken, nil
}

// ValidateToken validates a JWT token and returns the claims
func (tm *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return tm.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GenerateAccessToken generates a short-lived access token (15 minutes)
// for API authentication.
func (tm *TokenManager) GenerateAccessToken(userID, email, userType string) (string, error) {
	expiry := time.Now().Add(15 * time.Minute)

	claims := Claims{
		UserID:   userID,
		Email:    email,
		UserType: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "wish-list-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(tm.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign access token: %w", err)
	}
	return signedToken, nil
}

// GenerateRefreshToken generates a long-lived refresh token (7 days)
// with a unique token ID for rotation support.
func (tm *TokenManager) GenerateRefreshToken(userID, email, userType, tokenID string) (string, error) {
	expiry := time.Now().Add(7 * 24 * time.Hour)

	claims := Claims{
		UserID:   userID,
		Email:    email,
		UserType: userType,
		TokenID:  tokenID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "wish-list-app",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(tm.secret)
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}
	return signedToken, nil
}
