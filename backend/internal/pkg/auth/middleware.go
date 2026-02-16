package auth

import (
	"errors"
	"strings"

	"wish-list/internal/pkg/apperrors"

	"github.com/labstack/echo/v4"
)

// JWTMiddleware creates a middleware for JWT authentication
func JWTMiddleware(tm *TokenManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return apperrors.Unauthorized("Missing authorization header")
			}

			// Expect format: "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return apperrors.Unauthorized("Invalid authorization header format")
			}

			tokenString := parts[1]

			claims, err := tm.ValidateToken(tokenString)
			if err != nil {
				c.Logger().Warn("Token validation failed: ", err)
				return apperrors.Unauthorized("Invalid or expired token")
			}

			// Add claims to context
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("user_type", claims.UserType)

			return next(c)
		}
	}
}

// OptionalJWTMiddleware creates a middleware for optional JWT authentication
// If no token is provided or the token is invalid, the request continues
// but without user context
func OptionalJWTMiddleware(tm *TokenManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				// No token provided, continue without user context
				return next(c)
			}

			// Expect format: "Bearer <token>"
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				return apperrors.Unauthorized("Invalid authorization header format")
			}

			tokenString := parts[1]

			claims, err := tm.ValidateToken(tokenString)
			if err != nil {
				// Invalid token, continue without user context
				return next(c)
			}

			// Add claims to context
			c.Set("user_id", claims.UserID)
			c.Set("email", claims.Email)
			c.Set("user_type", claims.UserType)

			return next(c)
		}
	}
}

// RequireAuth middleware checks if the user is authenticated
func RequireAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userID := c.Get("user_id")
			if userID == nil {
				return apperrors.Unauthorized("Authentication required")
			}
			return next(c)
		}
	}
}

// RequireUserType middleware checks if the user is of a specific type
func RequireUserType(requiredType string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userType := c.Get("user_type")
			if userType == nil || userType != requiredType {
				return apperrors.Forbidden("Insufficient permissions")
			}
			return next(c)
		}
	}
}

// GetUserFromContext extracts user information from the context
func GetUserFromContext(c echo.Context) (userID, email, userType string, err error) {
	userIDVal := c.Get("user_id")
	emailVal := c.Get("email")
	userTypeVal := c.Get("user_type")

	if userIDVal == nil {
		return "", "", "", errors.New("user not found in context")
	}

	userID, ok := userIDVal.(string)
	if !ok {
		return "", "", "", errors.New("invalid user ID in context")
	}

	if emailVal != nil {
		email, _ = emailVal.(string)
	}

	if userTypeVal != nil {
		userType, _ = userTypeVal.(string)
	} else {
		userType = "user" // Default to regular user
	}

	return userID, email, userType, nil
}
