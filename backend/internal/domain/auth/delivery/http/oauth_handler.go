package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"wish-list/internal/domain/auth/delivery/http/dto"
	usermodels "wish-list/internal/domain/user/models"
	"wish-list/internal/pkg/auth"
	"wish-list/internal/pkg/helpers"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/facebook"
	"golang.org/x/oauth2/google"
)

// UserRepositoryInterface defines what the OAuth handler needs from the user repository.
type UserRepositoryInterface interface {
	GetByEmail(ctx context.Context, email string) (*usermodels.User, error)
	Create(ctx context.Context, user usermodels.User) (*usermodels.User, error)
	Update(ctx context.Context, user usermodels.User) (*usermodels.User, error)
}

// OAuthHandler handles OAuth authentication flows
type OAuthHandler struct {
	userRepo     UserRepositoryInterface
	tokenManager *auth.TokenManager
	googleConfig *oauth2.Config
	fbConfig     *oauth2.Config
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(
	userRepo UserRepositoryInterface,
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
// @Param			request	body	dto.OAuthCodeRequest	true	"Authorization code from Google"
// @Success      200 {object} dto.AuthResponse "Authentication successful"
// @Failure      400 {object} map[string]string "Invalid request"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /auth/oauth/google [post]
func (h *OAuthHandler) GoogleOAuth(c echo.Context) error {
	var req dto.OAuthCodeRequest
	if err := helpers.BindAndValidate(c, &req); err != nil {
		return err
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

	return c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &dto.UserResponse{
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
// @Param			request	body	dto.OAuthCodeRequest	true	"Authorization code from Facebook"
// @Success      200 {object} dto.AuthResponse "Authentication successful"
// @Failure      400 {object} map[string]string "Invalid request"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /auth/oauth/facebook [post]
func (h *OAuthHandler) FacebookOAuth(c echo.Context) error {
	var req dto.OAuthCodeRequest
	if err := helpers.BindAndValidate(c, &req); err != nil {
		return err
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

	return c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: &dto.UserResponse{
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
func (h *OAuthHandler) findOrCreateUser(email, firstName, lastName, avatarURL string) (*usermodels.User, error) {
	ctx := context.Background()

	// Validate email format
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, fmt.Errorf("invalid email format: %w", err)
	}

	// Sanitize and validate names
	firstName = strings.TrimSpace(firstName)
	lastName = strings.TrimSpace(lastName)

	const maxNameLength = 100
	if len(firstName) > maxNameLength {
		firstName = firstName[:maxNameLength]
	}
	if len(lastName) > maxNameLength {
		lastName = lastName[:maxNameLength]
	}

	// Validate avatar URL if present
	if avatarURL != "" {
		if _, err := url.ParseRequestURI(avatarURL); err != nil {
			avatarURL = ""
		}
	}

	// Try to find existing user by email
	user, err := h.userRepo.GetByEmail(ctx, email)
	if err == nil {
		// User exists, update avatar if provided and not set
		if avatarURL != "" && user.AvatarUrl.String == "" {
			user.AvatarUrl = pgtype.Text{String: avatarURL, Valid: true}
			user, err = h.userRepo.Update(ctx, *user)
			if err != nil {
				return nil, fmt.Errorf("failed to update user avatar: %w", err)
			}
		}
		return user, nil
	}

	// Handle repository errors other than "not found"
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// User doesn't exist, create new one
	// Note: OAuth users don't have passwords
	userID := uuid.New()
	newUser := usermodels.User{
		ID:           pgtype.UUID{Bytes: userID, Valid: true},
		Email:        email,
		PasswordHash: pgtype.Text{String: "", Valid: false}, // No password for OAuth users
		FirstName:    pgtype.Text{String: firstName, Valid: firstName != ""},
		LastName:     pgtype.Text{String: lastName, Valid: lastName != ""},
		AvatarUrl:    pgtype.Text{String: avatarURL, Valid: avatarURL != ""},
	}

	createdUser, err := h.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

// parseName splits full name into first and last name
func parseName(fullName string) (firstName, lastName string) {
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
