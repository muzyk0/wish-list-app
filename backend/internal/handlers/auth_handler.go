package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"
	"wish-list/internal/auth"
	"wish-list/internal/services"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AuthHandler handles cross-domain authentication endpoints
type AuthHandler struct {
	userService  services.UserServiceInterface
	tokenManager *auth.TokenManager
	codeStore    *auth.CodeStore
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(
	userService services.UserServiceInterface,
	tokenManager *auth.TokenManager,
	codeStore *auth.CodeStore,
) *AuthHandler {
	return &AuthHandler{
		userService:  userService,
		tokenManager: tokenManager,
		codeStore:    codeStore,
	}
}

// RefreshRequest represents the request body for token refresh (mobile clients)
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshResponse represents the response for token refresh
type RefreshResponse struct {
	AccessToken  string `json:"accessToken" validate:"required"`
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// HandoffResponse represents the response for mobile handoff code generation
type HandoffResponse struct {
	Code      string `json:"code" validate:"required" example:"a1b2c3d4e5f6..."`
	ExpiresIn int    `json:"expiresIn" validate:"required" example:"60"`
}

// ExchangeRequest represents the request body for code exchange
type ExchangeRequest struct {
	Code string `json:"code" validate:"required"`
}

// ExchangeResponse represents the response for code exchange
type ExchangeResponse struct {
	AccessToken  string        `json:"accessToken" validate:"required"`
	RefreshToken string        `json:"refreshToken" validate:"required"`
	User         *UserResponse `json:"user" validate:"required"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message" validate:"required"`
}

// ChangeEmailRequest represents the request body for changing email
type ChangeEmailRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=6"`
	NewEmail        string `json:"new_email" validate:"required,email"`
}

// ChangePasswordRequest represents the request body for changing password
type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" validate:"required,min=6"`
	NewPassword     string `json:"new_password" validate:"required,min=6"`
}

// Refresh godoc
//
//	@Summary		Refresh access token
//	@Description	Exchange refresh token for a new access token. Accepts refresh token via httpOnly cookie (web clients) or Authorization Bearer header (mobile clients). Implements token rotation - returns new refresh token on success.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	RefreshResponse		"Token refreshed successfully"
//	@Failure		401	{object}	map[string]string	"Invalid or expired refresh token"
//	@Router			/auth/refresh [post]
func (h *AuthHandler) Refresh(c echo.Context) error {
	var refreshToken string

	// Try to get refresh token from cookie first (web clients)
	cookie, err := c.Cookie("refreshToken")
	if err == nil && cookie.Value != "" {
		refreshToken = cookie.Value
	} else {
		// Try Authorization header (mobile clients)
		authHeader := c.Request().Header.Get("Authorization")
		if token, found := strings.CutPrefix(authHeader, "Bearer "); found {
			refreshToken = token
		} else {
			// Try request body (alternative for mobile)
			var req RefreshRequest
			if err := c.Bind(&req); err == nil && req.RefreshToken != "" {
				refreshToken = req.RefreshToken
			}
		}
	}

	if refreshToken == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "No refresh token provided",
		})
	}

	// Validate refresh token
	claims, err := h.tokenManager.ValidateToken(refreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid or expired refresh token",
		})
	}

	// Generate new access token
	newAccessToken, err := h.tokenManager.GenerateAccessToken(claims.UserID, claims.Email, claims.UserType)
	if err != nil {
		c.Logger().Errorf("Failed to generate access token: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate access token",
		})
	}

	// Generate new refresh token (rotation)
	newTokenID := uuid.New().String()
	newRefreshToken, err := h.tokenManager.GenerateRefreshToken(claims.UserID, claims.Email, claims.UserType, newTokenID)
	if err != nil {
		c.Logger().Errorf("Failed to generate refresh token: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate refresh token",
		})
	}

	// Set new refresh token cookie for web clients
	c.SetCookie(&http.Cookie{
		Name:     "refreshToken",
		Value:    newRefreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})

	// Return both tokens in response for mobile clients
	return c.JSON(http.StatusOK, RefreshResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	})
}

// MobileHandoff godoc
//
//	@Summary		Generate mobile handoff code
//	@Description	Generate a short-lived (60 second) one-time code for transferring authentication from Frontend to Mobile app.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	HandoffResponse		"Handoff code generated"
//	@Failure		401	{object}	map[string]string	"Not authenticated"
//	@Failure		429	{object}	map[string]string	"Rate limit exceeded (10 requests/minute per user)"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Router			/auth/mobile-handoff [post]
func (h *AuthHandler) MobileHandoff(c echo.Context) error {
	// Get authenticated user from context
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Not authenticated",
		})
	}

	// Parse user ID as UUID
	userUUID, err := uuid.Parse(userID)
	if err != nil {
		c.Logger().Errorf("Invalid user ID format: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}

	// Generate handoff code
	code, err := h.codeStore.GenerateCode(userUUID)
	if err != nil {
		c.Logger().Errorf("Failed to generate handoff code: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate handoff code",
		})
	}

	return c.JSON(http.StatusOK, HandoffResponse{
		Code:      code,
		ExpiresIn: 60, // 60 seconds
	})
}

// Exchange godoc
//
//	@Summary		Exchange handoff code for tokens
//	@Description	Exchange a handoff code received from Frontend redirect for access and refresh tokens. Code can only be used once.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			request	body		ExchangeRequest		true	"Exchange request"
//	@Success		200		{object}	ExchangeResponse	"Code exchanged successfully"
//	@Failure		400		{object}	map[string]string	"Invalid request body"
//	@Failure		401		{object}	map[string]string	"Invalid or expired code"
//	@Failure		429		{object}	map[string]string	"Rate limit exceeded (10 requests/minute)"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Router			/auth/exchange [post]
func (h *AuthHandler) Exchange(c echo.Context) error {
	var req ExchangeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if req.Code == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Code is required",
		})
	}

	// Exchange code for user ID
	userID, valid := h.codeStore.ExchangeCode(req.Code)
	if !valid {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid or expired code",
		})
	}

	// Get user information
	ctx := c.Request().Context()
	user, err := h.userService.GetUser(ctx, userID.String())
	if err != nil {
		if errors.Is(err, services.ErrUserNotFound) {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "User not found",
			})
		}
		c.Logger().Errorf("Failed to get user: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}

	// Generate access token
	accessToken, err := h.tokenManager.GenerateAccessToken(user.ID, user.Email, "user")
	if err != nil {
		c.Logger().Errorf("Failed to generate access token: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate access token",
		})
	}

	// Generate refresh token
	tokenID := uuid.New().String()
	refreshToken, err := h.tokenManager.GenerateRefreshToken(user.ID, user.Email, "user", tokenID)
	if err != nil {
		c.Logger().Errorf("Failed to generate refresh token: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate refresh token",
		})
	}

	// Create user response
	userResponse := &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarUrl: user.AvatarUrl,
	}

	return c.JSON(http.StatusOK, ExchangeResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userResponse,
	})
}

// Logout godoc
//
//	@Summary		Logout user
//	@Description	Clear refresh token cookie and invalidate session
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	MessageResponse		"Logout successful"
//	@Router			/auth/logout [post]
func (h *AuthHandler) Logout(c echo.Context) error {
	// Clear refresh token cookie
	c.SetCookie(&http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   -1, // Delete cookie
		Expires:  time.Unix(0, 0),
	})

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Logged out successfully",
	})
}

// ChangeEmail godoc
//
//	@Summary		Change user email
//	@Description	Change the authenticated user's email address with password verification. Requires current password to prevent unauthorized changes.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		ChangeEmailRequest	true	"Email change request"
//	@Success		200		{object}	MessageResponse		"Email changed successfully"
//	@Failure		400		{object}	map[string]string	"Invalid request body or validation error"
//	@Failure		401		{object}	map[string]string	"Unauthorized or incorrect password"
//	@Failure		409		{object}	map[string]string	"Email already in use"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Router			/auth/change-email [post]
func (h *AuthHandler) ChangeEmail(c echo.Context) error {
	// Get authenticated user from context
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Not authenticated",
		})
	}

	var req ChangeEmailRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	ctx := c.Request().Context()
	err = h.userService.ChangeEmail(ctx, userID, req.CurrentPassword, req.NewEmail)
	if err != nil {
		// Check for specific errors
		if errors.Is(err, services.ErrInvalidPassword) {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Current password is incorrect",
			})
		}
		if errors.Is(err, services.ErrUserAlreadyExists) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "Email already in use",
			})
		}
		// Log error and return generic message
		c.Logger().Errorf("Failed to change email for user %s: %v", userID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to change email",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Email changed successfully",
	})
}

// ChangePassword godoc
//
//	@Summary		Change user password
//	@Description	Change the authenticated user's password with current password verification. This will invalidate all existing sessions except the current one.
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		ChangePasswordRequest	true	"Password change request"
//	@Success		200		{object}	MessageResponse			"Password changed successfully"
//	@Failure		400		{object}	map[string]string		"Invalid request body or validation error"
//	@Failure		401		{object}	map[string]string		"Unauthorized or incorrect password"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Router			/auth/change-password [post]
func (h *AuthHandler) ChangePassword(c echo.Context) error {
	// Get authenticated user from context
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Not authenticated",
		})
	}

	var req ChangePasswordRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	ctx := c.Request().Context()
	err = h.userService.ChangePassword(ctx, userID, req.CurrentPassword, req.NewPassword)
	if err != nil {
		// Check for specific errors
		if errors.Is(err, services.ErrInvalidPassword) {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "Current password is incorrect",
			})
		}
		// Log error and return generic message
		c.Logger().Errorf("Failed to change password for user %s: %v", userID, err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to change password",
		})
	}

	return c.JSON(http.StatusOK, MessageResponse{
		Message: "Password changed successfully",
	})
}
