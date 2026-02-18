package dto

// UserResponse is a lightweight user DTO for auth responses.
type UserResponse struct {
	ID        string `json:"id" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarUrl string `json:"avatar_url"`
}

// AuthResponse contains user info with access and refresh tokens (used by OAuth flows).
type AuthResponse struct {
	// User information
	User *UserResponse `json:"user" validate:"required"`
	// Access token (short-lived, 15 minutes)
	AccessToken string `json:"accessToken" validate:"required"` //nolint:gosec // API field name for auth response
	// Refresh token (long-lived, 7 days)
	RefreshToken string `json:"refreshToken" validate:"required"` //nolint:gosec // API field name for auth response
}

// RefreshResponse represents the response for token refresh
type RefreshResponse struct {
	AccessToken  string `json:"accessToken" validate:"required"`  //nolint:gosec // API field name for token response
	RefreshToken string `json:"refreshToken" validate:"required"` //nolint:gosec // API field name for token response
}

// HandoffResponse represents the response for mobile handoff code generation
type HandoffResponse struct {
	Code      string `json:"code" validate:"required" example:"a1b2c3d4e5f6..."`
	ExpiresIn int    `json:"expiresIn" validate:"required" example:"60"`
}

// ExchangeResponse represents the response for code exchange
type ExchangeResponse struct {
	AccessToken  string        `json:"accessToken" validate:"required"`  //nolint:gosec // API field name for token response
	RefreshToken string        `json:"refreshToken" validate:"required"` //nolint:gosec // API field name for token response
	User         *UserResponse `json:"user" validate:"required"`
}

// MessageResponse represents a simple message response
type MessageResponse struct {
	Message string `json:"message" validate:"required"`
}
