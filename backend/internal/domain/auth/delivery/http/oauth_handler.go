package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"wish-list/internal/app/config"
	"wish-list/internal/domain/auth/delivery/http/dto"
	usermodels "wish-list/internal/domain/user/models"
	"wish-list/internal/domain/user/repository"
	"wish-list/internal/pkg/apperrors"
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

// GuestReservationLinker links guest reservations to an authenticated user by email.
type GuestReservationLinker interface {
	LinkGuestReservationsToUserByEmail(ctx context.Context, guestEmail string, userID pgtype.UUID) (int, error)
}

// OAuthHandler handles OAuth authentication flows
type OAuthHandler struct {
	userRepo          UserRepositoryInterface
	reservationLinker GuestReservationLinker
	tokenManager      *auth.TokenManager
	googleConfig      *oauth2.Config
	fbConfig          *oauth2.Config
	httpTimeout       time.Duration
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
	httpTimeout int,
	reservationLinker ...GuestReservationLinker,
) *OAuthHandler {
	timeout := time.Duration(httpTimeout) * time.Second
	if httpTimeout <= 0 {
		timeout = config.DefaultOAuthHTTPTimeout
	}

	var linker GuestReservationLinker
	if len(reservationLinker) > 0 {
		linker = reservationLinker[0]
	}

	return &OAuthHandler{
		userRepo:          userRepo,
		reservationLinker: linker,
		tokenManager:      tokenManager,
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
		httpTimeout: timeout,
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
	ID       string `json:"id"`
	Email    string `json:"email"`
	Verified bool   `json:"verified"`
	Name     string `json:"name"`
	Picture  struct {
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
// @Failure      400 {object} map[string]string "Invalid or expired authorization code"
// @Failure      502 {object} map[string]string "Failed to communicate with provider"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /auth/oauth/google [post]
func (h *OAuthHandler) GoogleOAuth(c echo.Context) error {
	var req dto.OAuthCodeRequest
	if err := helpers.BindAndValidate(c, &req); err != nil {
		return err
	}

	// Exchange authorization code for token
	ctx := c.Request().Context()
	token, err := h.googleConfig.Exchange(ctx, req.Code)
	if err != nil {
		return h.handleOAuthExchangeError(c, "Google", err)
	}

	// Get user info from Google
	userInfo, err := h.getGoogleUserInfo(ctx, token.AccessToken)
	if err != nil {
		return apperrors.Internal("Failed to get user information").Wrap(err)
	}

	// Verify email
	if !userInfo.VerifiedEmail {
		return apperrors.BadRequest("Email not verified with Google")
	}

	// Create or find user in database
	user, err := h.findOrCreateUser(
		ctx,
		userInfo.Email,
		userInfo.GivenName,
		userInfo.FamilyName,
		userInfo.Picture,
		userInfo.VerifiedEmail,
	)
	if err != nil {
		return apperrors.Internal("Failed to process user").Wrap(err)
	}

	// Generate our own tokens
	userIDStr := uuid.UUID(user.ID.Bytes).String()
	userEmail := user.Email
	if user.Email == "" && user.EncryptedEmail.Valid {
		userEmail = user.EncryptedEmail.String
	}
	accessToken, err := h.tokenManager.GenerateAccessToken(userIDStr, userEmail, "user")
	if err != nil {
		return apperrors.Internal("Failed to generate access token").Wrap(err)
	}

	tokenID := uuid.New().String()
	refreshToken, err := h.tokenManager.GenerateRefreshToken(userIDStr, userEmail, "user", tokenID)
	if err != nil {
		return apperrors.Internal("Failed to generate refresh token").Wrap(err)
	}

	userResp := &dto.UserResponse{
		ID:    userIDStr,
		Email: userEmail,
	}
	if user.FirstName.Valid {
		userResp.FirstName = user.FirstName.String
	}
	if user.LastName.Valid {
		userResp.LastName = user.LastName.String
	}
	if user.AvatarUrl.Valid {
		userResp.AvatarUrl = user.AvatarUrl.String
	}

	return c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userResp,
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
// @Failure      400 {object} map[string]string "Invalid or expired authorization code"
// @Failure      502 {object} map[string]string "Failed to communicate with provider"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /auth/oauth/facebook [post]
func (h *OAuthHandler) FacebookOAuth(c echo.Context) error {
	var req dto.OAuthCodeRequest
	if err := helpers.BindAndValidate(c, &req); err != nil {
		return err
	}

	// Exchange authorization code for token
	ctx := c.Request().Context()
	token, err := h.fbConfig.Exchange(ctx, req.Code)
	if err != nil {
		return h.handleOAuthExchangeError(c, "Facebook", err)
	}

	// Get user info from Facebook
	userInfo, err := h.getFacebookUserInfo(ctx, token.AccessToken)
	if err != nil {
		return apperrors.Internal("Failed to get user information").Wrap(err)
	}

	// Parse name (Facebook returns full name)
	firstName, lastName := parseName(userInfo.Name)

	// Create or find user in database
	user, err := h.findOrCreateUser(
		ctx,
		userInfo.Email,
		firstName,
		lastName,
		userInfo.Picture.Data.URL,
		userInfo.Verified,
	)
	if err != nil {
		return apperrors.Internal("Failed to process user").Wrap(err)
	}

	// Generate our own tokens
	userIDStr := uuid.UUID(user.ID.Bytes).String()
	userEmail := user.Email
	if user.Email == "" && user.EncryptedEmail.Valid {
		userEmail = user.EncryptedEmail.String
	}
	accessToken, err := h.tokenManager.GenerateAccessToken(userIDStr, userEmail, "user")
	if err != nil {
		return apperrors.Internal("Failed to generate access token").Wrap(err)
	}

	tokenID := uuid.New().String()
	refreshToken, err := h.tokenManager.GenerateRefreshToken(userIDStr, userEmail, "user", tokenID)
	if err != nil {
		return apperrors.Internal("Failed to generate refresh token").Wrap(err)
	}

	userResp := &dto.UserResponse{
		ID:    userIDStr,
		Email: userEmail,
	}
	if user.FirstName.Valid {
		userResp.FirstName = user.FirstName.String
	}
	if user.LastName.Valid {
		userResp.LastName = user.LastName.String
	}
	if user.AvatarUrl.Valid {
		userResp.AvatarUrl = user.AvatarUrl.String
	}

	return c.JSON(http.StatusOK, dto.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         userResp,
	})
}

// getGoogleUserInfo fetches user information from Google
func (h *OAuthHandler) getGoogleUserInfo(ctx context.Context, accessToken string) (*GoogleUserInfo, error) {
	client := &http.Client{Timeout: h.httpTimeout}
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

	//nolint:gosec // Intentional external API call to Google OAuth
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
	client := &http.Client{Timeout: h.httpTimeout}
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://graph.facebook.com/me?fields=id,name,email,picture,verified",
		http.NoBody,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)

	//nolint:gosec // Intentional external API call to Facebook OAuth
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
func (h *OAuthHandler) findOrCreateUser(ctx context.Context, email, firstName, lastName, avatarURL string, emailVerified bool) (*usermodels.User, error) {
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

	// Attempt to find existing user by email
	user, err := h.userRepo.GetByEmail(ctx, email)

	if err == nil {
		// User exists - update avatar and/or verification state if needed.
		needsUpdate := false

		// Check both Valid (not NULL) and String (not empty)
		if avatarURL != "" && (!user.AvatarUrl.Valid || user.AvatarUrl.String == "") {
			user.AvatarUrl = pgtype.Text{String: avatarURL, Valid: true}
			needsUpdate = true
		}

		if emailVerified && (!user.IsVerified.Valid || !user.IsVerified.Bool) {
			user.IsVerified = pgtype.Bool{Bool: true, Valid: true}
			needsUpdate = true
		}

		if needsUpdate {
			user, err = h.userRepo.Update(ctx, *user)
			if err != nil {
				return nil, fmt.Errorf("failed to update oauth user profile: %w", err)
			}
		}

		if user.IsVerified.Valid && user.IsVerified.Bool {
			h.linkGuestReservationsByEmail(ctx, email, user.ID)
		}
		return user, nil
	}

	// Distinguish between "user not found" (expected) and database errors (unexpected)
	if !errors.Is(err, repository.ErrUserNotFound) {
		// Other database errors (connection failure, timeout, etc.) should be returned
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	// User doesn't exist - this is the expected path for first-time OAuth users
	// Fall through to user creation below

	// Create new user from OAuth profile data
	// Note: OAuth users don't have passwords
	userID := uuid.New()
	newUser := usermodels.User{
		ID:           pgtype.UUID{Bytes: userID, Valid: true},
		Email:        email,
		PasswordHash: pgtype.Text{String: "", Valid: false}, // No password for OAuth users
		FirstName:    pgtype.Text{String: firstName, Valid: firstName != ""},
		LastName:     pgtype.Text{String: lastName, Valid: lastName != ""},
		AvatarUrl:    pgtype.Text{String: avatarURL, Valid: avatarURL != ""},
		IsVerified:   pgtype.Bool{Bool: emailVerified, Valid: true},
	}

	createdUser, err := h.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	if createdUser.IsVerified.Valid && createdUser.IsVerified.Bool {
		h.linkGuestReservationsByEmail(ctx, email, createdUser.ID)
	}

	return createdUser, nil
}

func (h *OAuthHandler) linkGuestReservationsByEmail(ctx context.Context, email string, userID pgtype.UUID) {
	if h.reservationLinker == nil || !userID.Valid {
		return
	}

	if _, err := h.reservationLinker.LinkGuestReservationsToUserByEmail(ctx, email, userID); err != nil {
		// Best-effort linking: OAuth login should still succeed.
		log.Printf("Warning: failed to link guest reservations for OAuth user %s: %v", email, err)
	}
}

// handleOAuthExchangeError returns appropriate HTTP status code based on error type.
// Client errors (invalid/expired code) return 400 Bad Request.
// Provider/network errors return 502 Bad Gateway for retry indication.
func (h *OAuthHandler) handleOAuthExchangeError(c echo.Context, provider string, err error) error {
	// Log full error for debugging (server-side only)
	c.Logger().Errorf("%s OAuth code exchange failed: %v", provider, err)

	// Check if error indicates client mistake (invalid/expired/revoked code)
	errMsg := strings.ToLower(err.Error())
	clientErrorKeywords := []string{"invalid", "expired", "revoked", "unauthorized", "denied", "malformed"}

	for _, keyword := range clientErrorKeywords {
		if strings.Contains(errMsg, keyword) {
			// Client error - bad request (user needs to re-authenticate)
			return apperrors.BadRequest("Invalid or expired authorization code. Please try logging in again.")
		}
	}

	// Provider/network error - bad gateway (retryable)
	// Examples: timeout, connection refused, DNS failure, provider downtime
	return apperrors.BadGateway("Failed to communicate with authentication provider. Please try again in a moment.")
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
