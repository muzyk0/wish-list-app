package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"

	"wish-list/internal/auth"
	db "wish-list/internal/db/models"
	"wish-list/internal/repositories"
)

// OAuthHandler handles OAuth authentication flows
type OAuthHandler struct {
	userRepo     repositories.UserRepositoryInterface
	tokenManager *auth.TokenManager
	googleConfig *oauth2.Config
	fbConfig     *oauth2.Config
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(
	userRepo repositories.UserRepositoryInterface,
	tokenManager *auth.TokenManager,
	googleClientID string,
	googleClientSecret string,
	fbClientID string,
	fbClientSecret string,
	redirectURL string,
) *OAuthHandler {
	return &OAuthHandler{
		userRepo:     userRepo,
		tokenManager: tokenManager,
		googleConfig: &oauth2.Config{
			ClientID:     googleClientID,
			ClientSecret: googleClientSecret,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
			Endpoint: google.Endpoint,
		},
		fbConfig: &oauth2.Config{
			ClientID:     fbClientID,
			ClientSecret: fbClientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{"email", "public_profile"},
			Endpoint:     facebook.Endpoint,
		},
	}
}

// OAuthCodeRequest represents the request body for OAuth code exchange
type OAuthCodeRequest struct {
	Code string `json:"code" validate:"required"`
}

// GoogleUserInfo represents user info from Google OAuth
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// FacebookUserInfo represents user info from Facebook OAuth
type FacebookUserInfo struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture struct {
		Data struct {
			URL string `json:"url"`
		} `json:"data"`
	} `json:"picture"`
}

// GoogleOAuth handles Google OAuth code exchange
// @Summary      Google OAuth authentication
// @Description  Exchange Google authorization code for access and refresh tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param			request	body	OAuthCodeRequest	true	"Authorization code from Google"
// @Success      200 {object} AuthResponse "Authentication successful"
// @Failure      400 {object} map[string]string "Invalid request"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /auth/oauth/google [post]
func (h *OAuthHandler) GoogleOAuth(c echo.Context) error {
	var req OAuthCodeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Code is required",
		})
	}

	// Exchange authorization code for token
	ctx := context.Background()
	token, err := h.googleConfig.Exchange(ctx, req.Code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to exchange authorization code",
		})
	}

	// Get user info from Google
	userInfo, err := h.getGoogleUserInfo(ctx, token.AccessToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get user information",
		})
	}

	// Verify email
	if !userInfo.VerifiedEmail {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Email not verified with Google",
		})
	}

	// Create or find user in database
	user, err := h.findOrCreateUser(
		userInfo.Email,
		userInfo.GivenName,
		userInfo.FamilyName,
		userInfo.Picture,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process user",
		})
	}

	// Generate our own tokens
	userIDStr := uuid.UUID(user.ID.Bytes).String()
	userEmail := user.Email
	if user.Email == "" && user.EncryptedEmail.Valid {
		// Use encrypted email if available (for encrypted scenarios)
		userEmail = user.EncryptedEmail.String
	}
	accessToken, err := h.tokenManager.GenerateAccessToken(userIDStr, userEmail, "user")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate access token",
		})
	}

	tokenID := uuid.New().String()
	refreshToken, err := h.tokenManager.GenerateRefreshToken(userIDStr, userEmail, "user", tokenID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate refresh token",
		})
	}

	// Return response
	return c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &UserResponse{
			ID:        userIDStr,
			Email:     userEmail,
			FirstName: user.FirstName.String,
			LastName:  user.LastName.String,
			AvatarUrl: user.AvatarUrl.String,
		},
	})
}

// FacebookOAuth handles Facebook OAuth code exchange
// @Summary      Facebook OAuth authentication
// @Description  Exchange Facebook authorization code for access and refresh tokens
// @Tags         Auth
// @Accept       json
// @Produce      json
// @Param			request	body	OAuthCodeRequest	true	"Authorization code from Facebook"
// @Success      200 {object} AuthResponse "Authentication successful"
// @Failure      400 {object} map[string]string "Invalid request"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /auth/oauth/facebook [post]
func (h *OAuthHandler) FacebookOAuth(c echo.Context) error {
	var req OAuthCodeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	if err := c.Validate(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Code is required",
		})
	}

	// Exchange authorization code for token
	ctx := context.Background()
	token, err := h.fbConfig.Exchange(ctx, req.Code)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to exchange authorization code",
		})
	}

	// Get user info from Facebook
	userInfo, err := h.getFacebookUserInfo(ctx, token.AccessToken)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get user information",
		})
	}

	// Parse name (Facebook returns full name)
	firstName, lastName := parseName(userInfo.Name)

	// Create or find user in database
	user, err := h.findOrCreateUser(
		userInfo.Email,
		firstName,
		lastName,
		userInfo.Picture.Data.URL,
	)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to process user",
		})
	}

	// Generate our own tokens
	userIDStr := uuid.UUID(user.ID.Bytes).String()
	userEmail := user.Email
	if user.Email == "" && user.EncryptedEmail.Valid {
		// Use encrypted email if available (for encrypted scenarios)
		userEmail = user.EncryptedEmail.String
	}
	accessToken, err := h.tokenManager.GenerateAccessToken(userIDStr, userEmail, "user")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate access token",
		})
	}

	tokenID := uuid.New().String()
	refreshToken, err := h.tokenManager.GenerateRefreshToken(userIDStr, userEmail, "user", tokenID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate refresh token",
		})
	}

	// Return response
	return c.JSON(http.StatusOK, AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &UserResponse{
			ID:        userIDStr,
			Email:     userEmail,
			FirstName: user.FirstName.String,
			LastName:  user.LastName.String,
			AvatarUrl: user.AvatarUrl.String,
		},
	})
}

// getGoogleUserInfo fetches user information from Google
func (h *OAuthHandler) getGoogleUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://www.googleapis.com/oauth2/v2/userinfo",
		http.NoBody,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("google API error: %s", string(body))
	}

	var userInfo GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// getFacebookUserInfo fetches user information from Facebook
func (h *OAuthHandler) getFacebookUserInfo(ctx context.Context, accessToken string) (*FacebookUserInfo, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://graph.facebook.com/me?fields=id,name,email,picture",
		http.NoBody,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("facebook API error: %s", string(body))
	}

	var userInfo FacebookUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, err
	}

	return &userInfo, nil
}

// findOrCreateUser finds existing user or creates new one from OAuth data
func (h *OAuthHandler) findOrCreateUser(email, firstName, lastName, avatarURL string) (*db.User, error) {
	ctx := context.Background()

	// Try to find existing user by email
	user, err := h.userRepo.GetByEmail(ctx, email)
	if err == nil {
		// User exists, update avatar if provided and not set
		if avatarURL != "" && user.AvatarUrl.String == "" {
			user.AvatarUrl = pgtype.Text{String: avatarURL, Valid: true}
			// Update user in database
			user, err = h.userRepo.Update(ctx, *user)
			if err != nil {
				return nil, err
			}
		}
		return user, nil
	}

	// User doesn't exist, create new one
	// Note: OAuth users don't have passwords
	userID := uuid.New()
	newUser := db.User{
		ID:           pgtype.UUID{Bytes: userID, Valid: true},
		Email:        email,
		PasswordHash: pgtype.Text{String: "", Valid: false}, // No password for OAuth users
		FirstName:    pgtype.Text{String: firstName, Valid: firstName != ""},
		LastName:     pgtype.Text{String: lastName, Valid: lastName != ""},
		AvatarUrl:    pgtype.Text{String: avatarURL, Valid: avatarURL != ""},
	}

	createdUser, err := h.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, err
	}

	return createdUser, nil
}

// parseName splits full name into first and last name
func parseName(fullName string) (firstName, lastName string) {
	// Simple parsing - split by space
	parts := splitName(fullName)
	if len(parts) == 0 {
		return "", ""
	}
	if len(parts) == 1 {
		return parts[0], ""
	}
	return parts[0], parts[len(parts)-1]
}

// splitName helper function
func splitName(name string) []string {
	var parts []string
	current := ""
	for _, char := range name {
		if char == ' ' {
			if current != "" {
				parts = append(parts, current)
				current = ""
			}
		} else {
			current += string(char)
		}
	}
	if current != "" {
		parts = append(parts, current)
	}
	return parts
}
