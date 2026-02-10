package handlers

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"time"
	"wish-list/internal/pkg/analytics"
	"wish-list/internal/pkg/auth"
	"wish-list/internal/services"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	service               services.UserServiceInterface
	tokenManager          *auth.TokenManager
	accountCleanupService *services.AccountCleanupService
	analyticsService      *analytics.AnalyticsService
}

func NewUserHandler(service services.UserServiceInterface, tokenManager *auth.TokenManager, accountCleanupService *services.AccountCleanupService, analyticsService *analytics.AnalyticsService) *UserHandler {
	return &UserHandler{
		service:               service,
		tokenManager:          tokenManager,
		accountCleanupService: accountCleanupService,
		analyticsService:      analyticsService,
	}
}

type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarUrl string `json:"avatar_url"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type UpdateProfileRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	AvatarUrl *string `json:"avatar_url"`
}

// UserResponse is the handler-level DTO for user data
type UserResponse struct {
	ID        string `json:"id" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarUrl string `json:"avatar_url"`
}

type AuthResponse struct {
	// User information
	User *UserResponse `json:"user" validate:"required"`
	// Access token (short-lived, 15 minutes)
	AccessToken string `json:"accessToken" validate:"required"`
	// Refresh token (long-lived, 7 days) - also set as httpOnly cookie
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type ProfileResponse struct {
	// User profile information
	User *UserResponse `json:"user" validate:"required"`
}

// toUserResponse maps service layer UserOutput to handler layer UserResponse
func (h *UserHandler) toUserResponse(user *services.UserOutput) *UserResponse {
	if user == nil {
		return nil
	}
	return &UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		AvatarUrl: user.AvatarUrl,
	}
}

// Register godoc
//
//	@Summary		Register a new user
//	@Description	Create a new user account with email and password
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			user	body		RegisterRequest		true	"User registration information"
//	@Success		201		{object}	AuthResponse		"User created successfully"
//	@Failure		400		{object}	map[string]string	"Invalid request body or validation error"
//	@Failure		409		{object}	map[string]string	"User with this email already exists"
//	@Failure		500		{object}	map[string]string	"Internal server error"
//	@Router			/auth/register [post]
func (h *UserHandler) Register(c echo.Context) error {
	var req RegisterRequest
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
	user, err := h.service.Register(ctx, services.RegisterUserInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		AvatarUrl: req.AvatarUrl,
	})

	if err != nil {
		// Detect duplicate user error specifically
		if errors.Is(err, services.ErrUserAlreadyExists) {
			return c.JSON(http.StatusConflict, map[string]string{
				"error": "User with this email already exists",
			})
		}
		// Log detailed error server-side, return generic message to client
		c.Logger().Errorf("Registration failed: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to create user account",
		})
	}

	// Generate access token (15 minutes)
	accessToken, err := h.tokenManager.GenerateAccessToken(user.ID, user.Email, "user")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Could not generate access token",
		})
	}

	// Generate refresh token (7 days)
	tokenID := fmt.Sprintf("%s-%d", user.ID, time.Now().Unix())
	refreshToken, err := h.tokenManager.GenerateRefreshToken(user.ID, user.Email, "user", tokenID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Could not generate refresh token",
		})
	}

	// Set refresh token as httpOnly cookie
	c.SetCookie(&http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})

	// Track user registration analytics
	_ = h.analyticsService.TrackUserRegistration(ctx, user.ID, user.Email)

	response := AuthResponse{
		User:         h.toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return c.JSON(http.StatusCreated, response)
}

// Login godoc
//
//	@Summary		User login
//	@Description	Authenticate user with email and password
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			credentials	body		LoginRequest		true	"User login credentials"
//	@Success		200			{object}	AuthResponse		"Login successful"
//	@Failure		400			{object}	map[string]string	"Invalid request body or validation error"
//	@Failure		401			{object}	map[string]string	"Invalid credentials"
//	@Failure		500			{object}	map[string]string	"Internal server error"
//	@Router			/auth/login [post]
func (h *UserHandler) Login(c echo.Context) error {
	var req LoginRequest
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
	user, err := h.service.Login(ctx, services.LoginUserInput{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		// Log the error server-side for debugging with redacted email (avoid PII in logs)
		emailHash := fmt.Sprintf("%x", sha256.Sum256([]byte(req.Email)))[:16]
		c.Logger().Errorf("Login failed for email_hash %s: %v", emailHash, err)
		// Return generic message to avoid leaking information about user existence
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Invalid credentials",
		})
	}

	// Generate access token (15 minutes)
	accessToken, err := h.tokenManager.GenerateAccessToken(user.ID, user.Email, "user")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Could not generate access token",
		})
	}

	// Generate refresh token (7 days)
	tokenID := fmt.Sprintf("%s-%d", user.ID, time.Now().Unix())
	refreshToken, err := h.tokenManager.GenerateRefreshToken(user.ID, user.Email, "user", tokenID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Could not generate refresh token",
		})
	}

	// Set refresh token as httpOnly cookie
	c.SetCookie(&http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
		MaxAge:   7 * 24 * 60 * 60, // 7 days
	})

	// Track user login analytics
	_ = h.analyticsService.TrackUserLogin(ctx, user.ID)

	response := AuthResponse{
		User:         h.toUserResponse(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return c.JSON(http.StatusOK, response)
}

// GetProfile godoc
//
//	@Summary		Get user profile
//	@Description	Get the authenticated user's profile information
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	UserResponse	"User profile"
//	@Failure		401	{object}	map[string]string	"Unauthorized"
//	@Failure		404	{object}	map[string]string	"User not found"
//	@Failure		500	{object}	map[string]string	"Internal server error"
//	@Router			/protected/profile [get]
func (h *UserHandler) GetProfile(c echo.Context) error {
	// Get user from context (after JWT middleware)
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	ctx := c.Request().Context()
	user, err := h.service.GetUser(ctx, userID)
	if err != nil {
		// Check for user not found error specifically
		if errors.Is(err, services.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "User not found",
			})
		}
		// Other errors are internal server errors
		c.Logger().Errorf("Failed to get user profile: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}

	return c.JSON(http.StatusOK, h.toUserResponse(user))
}

// UpdateProfile godoc
//
//	@Summary		Update user profile
//	@Description	Update the authenticated user's profile information
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			profile	body		UpdateProfileRequest	true	"Updated profile information"
//	@Success		200		{object}	UserResponse		"Updated user profile"
//	@Failure		400		{object}	map[string]string		"Invalid request body or validation error"
//	@Failure		401		{object}	map[string]string		"Unauthorized"
//	@Failure		404		{object}	map[string]string		"User not found"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Router			/protected/profile [put]
func (h *UserHandler) UpdateProfile(c echo.Context) error {
	// Get user from context (after JWT middleware)
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	var req UpdateProfileRequest
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
	user, err := h.service.UpdateProfile(ctx, userID, services.UpdateProfileInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		AvatarUrl: req.AvatarUrl,
	})

	if err != nil {
		// Log detailed error server-side, return generic message to client
		c.Logger().Errorf("Failed to update user profile: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}

	return c.JSON(http.StatusOK, h.toUserResponse(user))
}

// DeleteAccount godoc
//
// @Summary      Delete user account
// @Description  Delete the authenticated user's account and all associated data. This action is irreversible.
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      204  {object}  nil  "Account deleted successfully"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /protected/account [delete]
func (h *UserHandler) DeleteAccount(c echo.Context) error {
	// Get user from context (after JWT middleware)
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	ctx := c.Request().Context()
	err = h.accountCleanupService.DeleteUserAccount(ctx, userID, "user_requested_deletion")
	if err != nil {
		// Log detailed error server-side, return generic message to client
		c.Logger().Errorf("Failed to delete user account: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to delete account",
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// ExportUserData godoc
//
// @Summary      Export user data
// @Description  Export the authenticated user's data in JSON format for compliance and personal records
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  interface{}  "User data exported successfully"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Router       /protected/export-data [get]
func (h *UserHandler) ExportUserData(c echo.Context) error {
	// Get user from context (after JWT middleware)
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	ctx := c.Request().Context()
	data, err := h.accountCleanupService.ExportUserData(ctx, userID)
	if err != nil {
		// Log detailed error server-side, return generic message to client
		c.Logger().Errorf("Failed to export user data: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Unable to export user data",
		})
	}

	return c.JSON(http.StatusOK, data)
}
