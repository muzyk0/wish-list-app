package service

import (
	"context"
	"errors"
	"fmt"

	"wish-list/internal/domain/user/models"
	"wish-list/internal/domain/user/repository"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

// Sentinel errors
var (
	ErrUserAlreadyExists   = errors.New("user with this email already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrInvalidPassword     = errors.New("invalid password")
	ErrCredentialsRequired = errors.New("email and password are required")
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidUserID       = errors.New("invalid user id")
)

// UserServiceInterface defines the interface for user-related operations
type UserServiceInterface interface {
	Register(ctx context.Context, input RegisterUserInput) (*UserOutput, error)
	Login(ctx context.Context, input LoginUserInput) (*UserOutput, error)
	GetUser(ctx context.Context, userID string) (*UserOutput, error)
	UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (*UserOutput, error)
	ChangeEmail(ctx context.Context, userID, currentPassword, newEmail string) error
	ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error
	DeleteUser(ctx context.Context, userID string) error
}

type UserService struct {
	repo repository.UserRepositoryInterface
}

func NewUserService(repo repository.UserRepositoryInterface) *UserService {
	return &UserService{
		repo: repo,
	}
}

type RegisterUserInput struct {
	Email     string
	Password  string
	FirstName string
	LastName  string
	AvatarUrl string
}

type LoginUserInput struct {
	Email    string
	Password string
}

type UpdateUserInput struct {
	Email     *string
	Password  *string
	FirstName *string
	LastName  *string
	AvatarUrl *string
}

type UpdateProfileInput struct {
	FirstName *string
	LastName  *string
	AvatarUrl *string
}

type UserOutput struct {
	ID        string
	Email     string
	FirstName string
	LastName  string
	AvatarUrl string
}

func (s *UserService) Register(ctx context.Context, input RegisterUserInput) (*UserOutput, error) {
	// Validate input
	if input.Email == "" || input.Password == "" {
		return nil, ErrCredentialsRequired
	}

	// Check if user already exists
	existingUser, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		// If error is "user not found", continue with registration
		if !errors.Is(err, repository.ErrUserNotFound) {
			// Surface other database errors
			return nil, fmt.Errorf("failed to check existing user: %w", err)
		}
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := models.User{
		Email: input.Email,
		PasswordHash: pgtype.Text{
			String: string(hashedPassword),
			Valid:  true,
		},
		FirstName: pgtype.Text{
			String: input.FirstName,
			Valid:  input.FirstName != "",
		},
		LastName: pgtype.Text{
			String: input.LastName,
			Valid:  input.LastName != "",
		},
		AvatarUrl: pgtype.Text{
			String: input.AvatarUrl,
			Valid:  input.AvatarUrl != "",
		},
	}

	createdUser, err := s.repo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	output := &UserOutput{
		ID:        createdUser.ID.String(),
		Email:     createdUser.Email,
		FirstName: createdUser.FirstName.String,
		LastName:  createdUser.LastName.String,
		AvatarUrl: createdUser.AvatarUrl.String,
	}

	return output, nil
}

func (s *UserService) Login(ctx context.Context, input LoginUserInput) (*UserOutput, error) {
	// Validate input
	if input.Email == "" || input.Password == "" {
		return nil, ErrCredentialsRequired
	}

	// Get user by email
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check if password hash is valid
	if !user.PasswordHash.Valid {
		return nil, ErrInvalidCredentials
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(input.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	output := &UserOutput{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName.String,
		LastName:  user.LastName.String,
		AvatarUrl: user.AvatarUrl.String,
	}

	return output, nil
}

func (s *UserService) GetUser(ctx context.Context, userID string) (*UserOutput, error) {
	id := pgtype.UUID{}
	if err := id.Scan(userID); err != nil {
		return nil, ErrInvalidUserID
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	output := &UserOutput{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName.String,
		LastName:  user.LastName.String,
		AvatarUrl: user.AvatarUrl.String,
	}

	return output, nil
}

func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	id := pgtype.UUID{}
	if err := id.Scan(userID); err != nil {
		return ErrInvalidUserID
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

// UpdateProfile updates only non-sensitive profile information (firstName, lastName, avatarUrl)
func (s *UserService) UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (*UserOutput, error) {
	id := pgtype.UUID{}
	if err := id.Scan(userID); err != nil {
		return nil, ErrInvalidUserID
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update only profile fields (no email or password)
	if input.FirstName != nil {
		user.FirstName = pgtype.Text{
			String: *input.FirstName,
			Valid:  true,
		}
	}
	if input.LastName != nil {
		user.LastName = pgtype.Text{
			String: *input.LastName,
			Valid:  true,
		}
	}
	if input.AvatarUrl != nil {
		user.AvatarUrl = pgtype.Text{
			String: *input.AvatarUrl,
			Valid:  true,
		}
	}

	updatedUser, err := s.repo.Update(ctx, *user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	output := &UserOutput{
		ID:        updatedUser.ID.String(),
		Email:     updatedUser.Email,
		FirstName: updatedUser.FirstName.String,
		LastName:  updatedUser.LastName.String,
		AvatarUrl: updatedUser.AvatarUrl.String,
	}

	return output, nil
}

// ChangeEmail changes the user's email address with password verification
func (s *UserService) ChangeEmail(ctx context.Context, userID, currentPassword, newEmail string) error {
	id := pgtype.UUID{}
	if err := id.Scan(userID); err != nil {
		return ErrInvalidUserID
	}

	// Get current user
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify current password
	if !user.PasswordHash.Valid {
		return ErrInvalidPassword
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(currentPassword)); err != nil {
		return ErrInvalidPassword
	}

	// Check if new email is already in use by another account
	existingUser, err := s.repo.GetByEmail(ctx, newEmail)
	if err == nil && existingUser.ID != user.ID {
		return ErrUserAlreadyExists
	}

	// Update email
	user.Email = newEmail

	_, err = s.repo.Update(ctx, *user)
	if err != nil {
		return fmt.Errorf("failed to update user email: %w", err)
	}

	return nil
}

// ChangePassword changes the user's password with current password verification
func (s *UserService) ChangePassword(ctx context.Context, userID, currentPassword, newPassword string) error {
	id := pgtype.UUID{}
	if err := id.Scan(userID); err != nil {
		return ErrInvalidUserID
	}

	// Get current user
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Verify current password
	if !user.PasswordHash.Valid {
		return ErrInvalidPassword
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(currentPassword)); err != nil {
		return ErrInvalidPassword
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	user.PasswordHash = pgtype.Text{
		String: string(hashedPassword),
		Valid:  true,
	}

	_, err = s.repo.Update(ctx, *user)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	return nil
}
