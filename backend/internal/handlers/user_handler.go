package handlers

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"wish-list/internal/analytics"
	"wish-list/internal/auth"
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
	Email     *string `json:"email" validate:"omitempty,email"`
	Password  *string `json:"password" validate:"omitempty,min=6"`
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	AvatarUrl *string `json:"avatar_url"`
}

type AuthResponse struct {
	User  *services.UserOutput `json:"user"`
	Token string               `json:"token"`
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param user body RegisterRequest true "User registration information"
// @Success 201 {object} AuthResponse "User created successfully"
// @Failure 400 {object} map[string]string "Invalid request body or validation error"
// @Failure 409 {object} map[string]string "User with this email already exists"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/register [post]
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

	// Generate JWT token using token manager
	tokenString, err := h.tokenManager.GenerateToken(user.ID, user.Email, "user", 72) // 72 hours expiry
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Could not generate token",
		})
	}

	// Track user registration analytics
	_ = h.analyticsService.TrackUserRegistration(ctx, user.ID, user.Email)

	response := AuthResponse{
		User:  user,
		Token: tokenString,
	}

	return c.JSON(http.StatusCreated, response)
}

// Login godoc
// @Summary User login
// @Description Authenticate user with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "User login credentials"
// @Success 200 {object} AuthResponse "Login successful"
// @Failure 400 {object} map[string]string "Invalid request body or validation error"
// @Failure 401 {object} map[string]string "Invalid credentials"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /auth/login [post]
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

	// Generate JWT token using token manager
	tokenString, err := h.tokenManager.GenerateToken(user.ID, user.Email, "user", 72) // 72 hours expiry
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Could not generate token",
		})
	}

	// Track user login analytics
	_ = h.analyticsService.TrackUserLogin(ctx, user.ID)

	response := AuthResponse{
		User:  user,
		Token: tokenString,
	}

	return c.JSON(http.StatusOK, response)
}

// GetProfile godoc
// @Summary Get user profile
// @Description Get the authenticated user's profile information
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} services.UserOutput "User profile"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /protected/profile [get]
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

	return c.JSON(http.StatusOK, user)
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update the authenticated user's profile information
// @Tags User
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param profile body UpdateProfileRequest true "Updated profile information"
// @Success 200 {object} services.UserOutput "Updated user profile"
// @Failure 400 {object} map[string]string "Invalid request body or validation error"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /protected/profile [put]
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
	user, err := h.service.UpdateUser(ctx, userID, services.UpdateUserInput{
		Email:     req.Email,
		Password:  req.Password,
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

	return c.JSON(http.StatusOK, user)
}

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
