package handlers

import (
	"net/http"
	"wish-list/internal/auth"
	"wish-list/internal/services"

	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	service               services.UserServiceInterface
	tokenManager          *auth.TokenManager
	accountCleanupService *services.AccountCleanupService
}

func NewUserHandler(service services.UserServiceInterface, tokenManager *auth.TokenManager, accountCleanupService *services.AccountCleanupService) *UserHandler {
	return &UserHandler{
		service:               service,
		tokenManager:          tokenManager,
		accountCleanupService: accountCleanupService,
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

type AuthResponse struct {
	User  *services.UserOutput `json:"user"`
	Token string               `json:"token"`
}

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
		return c.JSON(http.StatusConflict, map[string]string{
			"error": err.Error(),
		})
	}

	// Generate JWT token using token manager
	tokenString, err := h.tokenManager.GenerateToken(user.ID, user.Email, "user", 72) // 72 hours expiry
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Could not generate token",
		})
	}

	response := AuthResponse{
		User:  user,
		Token: tokenString,
	}

	return c.JSON(http.StatusCreated, response)
}

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
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	// Generate JWT token using token manager
	tokenString, err := h.tokenManager.GenerateToken(user.ID, user.Email, "user", 72) // 72 hours expiry
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Could not generate token",
		})
	}

	response := AuthResponse{
		User:  user,
		Token: tokenString,
	}

	return c.JSON(http.StatusOK, response)
}

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
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "User not found",
		})
	}

	return c.JSON(http.StatusOK, user)
}

func (h *UserHandler) UpdateProfile(c echo.Context) error {
	// Get user from context (after JWT middleware)
	userID, _, _, err := auth.GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "Unauthorized",
		})
	}

	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	ctx := c.Request().Context()
	user, err := h.service.UpdateUser(ctx, userID, services.RegisterUserInput{
		Email:     req.Email,
		Password:  req.Password,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		AvatarUrl: req.AvatarUrl,
	})

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
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
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
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
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, data)
}
