package dto

import (
	userservice "wish-list/internal/domain/user/service"
)

// RegisterRequest represents the user registration request
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarUrl string `json:"avatar_url"`
}

// ToDomain converts the request DTO to a service input
func (r *RegisterRequest) ToDomain() userservice.RegisterUserInput {
	return userservice.RegisterUserInput{
		Email:     r.Email,
		Password:  r.Password,
		FirstName: r.FirstName,
		LastName:  r.LastName,
		AvatarUrl: r.AvatarUrl,
	}
}

// LoginRequest represents the user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// ToDomain converts the request DTO to a service input
func (r *LoginRequest) ToDomain() userservice.LoginUserInput {
	return userservice.LoginUserInput{
		Email:    r.Email,
		Password: r.Password,
	}
}

// UpdateProfileRequest represents the profile update request
type UpdateProfileRequest struct {
	FirstName *string `json:"first_name"`
	LastName  *string `json:"last_name"`
	AvatarUrl *string `json:"avatar_url"`
}

// ToDomain converts the request DTO to a service input
func (r *UpdateProfileRequest) ToDomain() userservice.UpdateProfileInput {
	return userservice.UpdateProfileInput{
		FirstName: r.FirstName,
		LastName:  r.LastName,
		AvatarUrl: r.AvatarUrl,
	}
}
