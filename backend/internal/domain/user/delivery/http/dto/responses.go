package dto

import (
	userservice "wish-list/internal/domain/user/service"
)

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error string `json:"error" example:"Error message" validate:"required"`
}

// UserResponse is the handler-level DTO for user data
type UserResponse struct {
	ID        string `json:"id" validate:"required"`
	Email     string `json:"email" validate:"required,email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarUrl string `json:"avatar_url"`
}

// UserResponseFromDomain maps service layer UserOutput to handler layer UserResponse
func UserResponseFromDomain(user *userservice.UserOutput) *UserResponse {
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

// AuthResponse contains user info with access and refresh tokens
type AuthResponse struct {
	// User information
	User *UserResponse `json:"user" validate:"required"`
	// Access token (short-lived, 15 minutes)
	AccessToken string `json:"accessToken" validate:"required"`
	// Refresh token (long-lived, 7 days) - also set as httpOnly cookie
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// ProfileResponse wraps user profile information
type ProfileResponse struct {
	// User profile information
	User *UserResponse `json:"user" validate:"required"`
}

// ExportedGiftItemResponse represents a gift item in the data export
type ExportedGiftItemResponse struct {
	ID          string  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string  `json:"name" example:"iPhone 15 Pro"`
	Description string  `json:"description" example:"256GB, Blue Titanium"`
	Link        string  `json:"link" example:"https://apple.com/iphone-15-pro"`
	ImageURL    string  `json:"image_url" example:"https://example.com/image.jpg"`
	Price       float64 `json:"price" example:"999.99"`
	Priority    int32   `json:"priority" example:"3"`
	CreatedAt   string  `json:"created_at" example:"2024-01-01T12:00:00Z"`
}

// ExportedWishlistResponse represents a wishlist in the data export
type ExportedWishlistResponse struct {
	ID          string                      `json:"id" example:"550e8400-e29b-41d4-a716-446655440001"`
	Title       string                      `json:"title" example:"Birthday 2024"`
	Description string                      `json:"description" example:"My birthday wishlist"`
	Occasion    string                      `json:"occasion" example:"Birthday"`
	IsPublic    bool                        `json:"is_public" example:"true"`
	PublicSlug  string                      `json:"public_slug" example:"birthday-2024-abc123"`
	CreatedAt   string                      `json:"created_at" example:"2024-01-01T12:00:00Z"`
	GiftItems   []ExportedGiftItemResponse  `json:"gift_items"`
}

// ExportedUserResponse represents user info in the data export
type ExportedUserResponse struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440002"`
	Email     string `json:"email" example:"user@example.com"`
	Name      string `json:"name" example:"John Doe"`
	CreatedAt string `json:"created_at" example:"2024-01-01T12:00:00Z"`
	UpdatedAt string `json:"updated_at" example:"2024-01-01T12:00:00Z"`
}

// ExportUserDataResponse represents the complete user data export
type ExportUserDataResponse struct {
	User         ExportedUserResponse        `json:"user" validate:"required"`
	Wishlists    []ExportedWishlistResponse  `json:"wishlists" validate:"required"`
	ExportedAt   string                      `json:"exported_at" example:"2024-01-01T12:00:00Z" validate:"required"`
	ExportFormat string                      `json:"export_format" example:"json" validate:"required"`
}

// ExportUserDataResponseFromMap converts map[string]any from service to typed response
func ExportUserDataResponseFromMap(data map[string]any) ExportUserDataResponse {
	response := ExportUserDataResponse{
		ExportedAt:   data["exported_at"].(string),
		ExportFormat: data["export_format"].(string),
		Wishlists:    []ExportedWishlistResponse{},
	}

	// Convert user data
	if userData, ok := data["user"].(map[string]any); ok {
		response.User = ExportedUserResponse{
			ID:        userData["id"].(string),
			Email:     userData["email"].(string),
			Name:      userData["name"].(string),
			CreatedAt: userData["created_at"].(string),
			UpdatedAt: userData["updated_at"].(string),
		}
	}

	// Convert wishlists data
	if wishlistsData, ok := data["wishlists"].([]map[string]any); ok {
		for _, wl := range wishlistsData {
			wishlist := ExportedWishlistResponse{
				ID:          wl["id"].(string),
				Title:       wl["title"].(string),
				Description: wl["description"].(string),
				Occasion:    wl["occasion"].(string),
				IsPublic:    wl["is_public"].(bool),
				PublicSlug:  wl["public_slug"].(string),
				CreatedAt:   wl["created_at"].(string),
				GiftItems:   []ExportedGiftItemResponse{},
			}

			// Convert gift items
			if itemsData, ok := wl["gift_items"].([]map[string]any); ok {
				for _, item := range itemsData {
					giftItem := ExportedGiftItemResponse{
						ID:          item["id"].(string),
						Name:        item["name"].(string),
						Description: item["description"].(string),
						Link:        item["link"].(string),
						ImageURL:    item["image_url"].(string),
						Price:       item["price"].(float64),
						Priority:    item["priority"].(int32),
						CreatedAt:   item["created_at"].(string),
					}
					wishlist.GiftItems = append(wishlist.GiftItems, giftItem)
				}
			}

			response.Wishlists = append(response.Wishlists, wishlist)
		}
	}

	return response
}
