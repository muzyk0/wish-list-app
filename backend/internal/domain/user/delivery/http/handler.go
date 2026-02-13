package http

import (
	"crypto/sha256"
	"errors"
	"fmt"
	nethttp "net/http"
	"time"

	"wish-list/internal/app/jobs"
	"wish-list/internal/domain/user/delivery/http/dto"
	userservice "wish-list/internal/domain/user/service"
	"wish-list/internal/pkg/analytics"
	"wish-list/internal/pkg/auth"
	"wish-list/internal/pkg/helpers"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service               userservice.UserServiceInterface
	tokenManager          *auth.TokenManager
	accountCleanupService *jobs.AccountCleanupService
	analyticsService      *analytics.AnalyticsService
}

func NewHandler(service userservice.UserServiceInterface, tokenManager *auth.TokenManager, accountCleanupService *jobs.AccountCleanupService, analyticsService *analytics.AnalyticsService) *Handler {
	return &Handler{
		service:               service,
		tokenManager:          tokenManager,
		accountCleanupService: accountCleanupService,
		analyticsService:      analyticsService,
	}
}

// Register godoc
//
//	@Summary		Register a new user
//	@Description	Create a new user account with email and password
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			user	body		dto.RegisterRequest		true	"User registration information"
//	@Success		201		{object}	dto.AuthResponse		"User created successfully"
//	@Failure		400		{object}	map[string]string		"Invalid request body or validation error"
//	@Failure		409		{object}	map[string]string		"User with this email already exists"
//	@Failure		500		{object}	map[string]string		"Internal server error"
//	@Router			/auth/register [post]
func (h *Handler) Register(c echo.Context) error {
	var req dto.RegisterRequest
	if err := helpers.BindAndValidate(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	user, err := h.service.Register(ctx, req.ToDomain())

	if err != nil {
		// Detect duplicate user error specifically
		if errors.Is(err, userservice.ErrUserAlreadyExists) {
			return c.JSON(nethttp.StatusConflict, map[string]string{
				"error": "User with this email already exists",
			})
		}
		// Log detailed error server-side, return generic message to client
		c.Logger().Errorf("Registration failed: %v", err)
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": "Failed to create user account",
		})
	}

	// Generate access token (15 minutes)
	accessToken, err := h.tokenManager.GenerateAccessToken(user.ID, user.Email, "user")
	if err != nil {
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": "Could not generate access token",
		})
	}

	// Generate refresh token (7 days)
	tokenID := fmt.Sprintf("%s-%d", user.ID, time.Now().Unix())
	refreshToken, err := h.tokenManager.GenerateRefreshToken(user.ID, user.Email, "user", tokenID)
	if err != nil {
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": "Could not generate refresh token",
		})
	}

	// Set refresh token as httpOnly cookie
	c.SetCookie(auth.NewRefreshTokenCookie(refreshToken))

	// Track user registration analytics
	_ = h.analyticsService.TrackUserRegistration(ctx, user.ID, user.Email)

	response := dto.AuthResponse{
		User:         dto.UserResponseFromDomain(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return c.JSON(nethttp.StatusCreated, response)
}

// Login godoc
//
//	@Summary		User login
//	@Description	Authenticate user with email and password
//	@Tags			Authentication
//	@Accept			json
//	@Produce		json
//	@Param			credentials	body		dto.LoginRequest		true	"User login credentials"
//	@Success		200			{object}	dto.AuthResponse		"Login successful"
//	@Failure		400			{object}	map[string]string		"Invalid request body or validation error"
//	@Failure		401			{object}	map[string]string		"Invalid credentials"
//	@Failure		500			{object}	map[string]string		"Internal server error"
//	@Router			/auth/login [post]
func (h *Handler) Login(c echo.Context) error {
	var req dto.LoginRequest
	if err := helpers.BindAndValidate(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	user, err := h.service.Login(ctx, req.ToDomain())

	if err != nil {
		// Log the error server-side for debugging with redacted email (avoid PII in logs)
		emailHash := fmt.Sprintf("%x", sha256.Sum256([]byte(req.Email)))[:16]
		c.Logger().Errorf("Login failed for email_hash %s: %v", emailHash, err)
		// Return generic message to avoid leaking information about user existence
		return c.JSON(nethttp.StatusUnauthorized, map[string]string{
			"error": "Invalid credentials",
		})
	}

	// Generate access token (15 minutes)
	accessToken, err := h.tokenManager.GenerateAccessToken(user.ID, user.Email, "user")
	if err != nil {
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": "Could not generate access token",
		})
	}

	// Generate refresh token (7 days)
	tokenID := fmt.Sprintf("%s-%d", user.ID, time.Now().Unix())
	refreshToken, err := h.tokenManager.GenerateRefreshToken(user.ID, user.Email, "user", tokenID)
	if err != nil {
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": "Could not generate refresh token",
		})
	}

	// Set refresh token as httpOnly cookie
	c.SetCookie(auth.NewRefreshTokenCookie(refreshToken))

	// Track user login analytics
	_ = h.analyticsService.TrackUserLogin(ctx, user.ID)

	response := dto.AuthResponse{
		User:         dto.UserResponseFromDomain(user),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return c.JSON(nethttp.StatusOK, response)
}

// GetProfile godoc
//
//	@Summary		Get user profile
//	@Description	Get the authenticated user's profile information
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.UserResponse		"User profile"
//	@Failure		401	{object}	map[string]string		"Unauthorized"
//	@Failure		404	{object}	map[string]string		"User not found"
//	@Failure		500	{object}	map[string]string		"Internal server error"
//	@Router			/protected/profile [get]
func (h *Handler) GetProfile(c echo.Context) error {
	userID := auth.MustGetUserID(c)

	ctx := c.Request().Context()
	user, err := h.service.GetUser(ctx, userID)
	if err != nil {
		// Check for user not found error specifically
		if errors.Is(err, userservice.ErrUserNotFound) {
			return c.JSON(nethttp.StatusNotFound, map[string]string{
				"error": "User not found",
			})
		}
		// Other errors are internal server errors
		c.Logger().Errorf("Failed to get user profile: %v", err)
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}

	return c.JSON(nethttp.StatusOK, dto.UserResponseFromDomain(user))
}

// UpdateProfile godoc
//
//	@Summary		Update user profile
//	@Description	Update the authenticated user's profile information
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			profile	body		dto.UpdateProfileRequest	true	"Updated profile information"
//	@Success		200		{object}	dto.UserResponse			"Updated user profile"
//	@Failure		400		{object}	map[string]string			"Invalid request body or validation error"
//	@Failure		401		{object}	map[string]string			"Unauthorized"
//	@Failure		404		{object}	map[string]string			"User not found"
//	@Failure		500		{object}	map[string]string			"Internal server error"
//	@Router			/protected/profile [put]
func (h *Handler) UpdateProfile(c echo.Context) error {
	userID := auth.MustGetUserID(c)

	var req dto.UpdateProfileRequest
	if err := helpers.BindAndValidate(c, &req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	user, err := h.service.UpdateProfile(ctx, userID, req.ToDomain())

	if err != nil {
		// Log detailed error server-side, return generic message to client
		c.Logger().Errorf("Failed to update user profile: %v", err)
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": "Internal server error",
		})
	}

	return c.JSON(nethttp.StatusOK, dto.UserResponseFromDomain(user))
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
func (h *Handler) DeleteAccount(c echo.Context) error {
	userID := auth.MustGetUserID(c)

	ctx := c.Request().Context()
	err := h.accountCleanupService.DeleteUserAccount(ctx, userID, "user_requested_deletion")
	if err != nil {
		// Log detailed error server-side, return generic message to client
		c.Logger().Errorf("Failed to delete user account: %v", err)
		return c.JSON(nethttp.StatusInternalServerError, map[string]string{
			"error": "Failed to delete account",
		})
	}

	return c.NoContent(nethttp.StatusNoContent)
}

// ExportUserData godoc
//
// @Summary      Export user data
// @Description  Export the authenticated user's data in JSON format for compliance and personal records
// @Tags         User
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  dto.ExportUserDataResponse  "User data exported successfully"
// @Failure      401  {object}  dto.ErrorResponse  "Unauthorized"
// @Failure      500  {object}  dto.ErrorResponse  "Internal server error"
// @Router       /protected/export-data [get]
func (h *Handler) ExportUserData(c echo.Context) error {
	userID := auth.MustGetUserID(c)

	ctx := c.Request().Context()
	data, err := h.accountCleanupService.ExportUserData(ctx, userID)
	if err != nil {
		// Log detailed error server-side, return generic message to client
		c.Logger().Errorf("Failed to export user data: %v", err)
		return c.JSON(nethttp.StatusInternalServerError, dto.ErrorResponse{
			Error: "Unable to export user data",
		})
	}

	// Convert map to typed response
	response := dto.ExportUserDataResponseFromMap(data)
	return c.JSON(nethttp.StatusOK, response)
}
