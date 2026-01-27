package services

import (
	"context"
	"errors"
	db "wish-list/internal/db/models"
	"wish-list/internal/repositories"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

// Sentinel errors
var (
	ErrUserAlreadyExists = errors.New("user with this email already exists")
)

// UserServiceInterface defines the interface for user-related operations
type UserServiceInterface interface {
	Register(ctx context.Context, input RegisterUserInput) (*UserOutput, error)
	Login(ctx context.Context, input LoginUserInput) (*UserOutput, error)
	GetUser(ctx context.Context, userID string) (*UserOutput, error)
	UpdateUser(ctx context.Context, userID string, input UpdateUserInput) (*UserOutput, error)
	DeleteUser(ctx context.Context, userID string) error
}

type UserService struct {
	repo repositories.UserRepositoryInterface
}

func NewUserService(repo repositories.UserRepositoryInterface) *UserService {
	return &UserService{
		repo: repo,
	}
}

type RegisterUserInput struct {
	Email     string `json:"email"`
	Password  string `json:"password"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarUrl string `json:"avatar_url"`
}

type LoginUserInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserInput struct {
	Email     *string `json:"email,omitempty"`
	Password  *string `json:"password,omitempty"`
	FirstName *string `json:"first_name,omitempty"`
	LastName  *string `json:"last_name,omitempty"`
	AvatarUrl *string `json:"avatar_url,omitempty"`
}

type UserOutput struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarUrl string `json:"avatar_url"`
}

func (s *UserService) Register(ctx context.Context, input RegisterUserInput) (*UserOutput, error) {
	// Validate input
	if input.Email == "" || input.Password == "" {
		return nil, errors.New("email and password are required")
	}

	// Check if user already exists
	existingUser, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		// If error is "user not found", continue with registration
		if err.Error() != "user not found" {
			// Surface other database errors
			return nil, err
		}
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user
	user := db.User{
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
		return nil, err
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
		return nil, errors.New("email and password are required")
	}

	// Get user by email
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Compare password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte(input.Password)); err != nil {
		return nil, errors.New("invalid email or password")
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
		return nil, errors.New("invalid user id")
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
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

func (s *UserService) UpdateUser(ctx context.Context, userID string, input UpdateUserInput) (*UserOutput, error) {
	id := pgtype.UUID{}
	if err := id.Scan(userID); err != nil {
		return nil, errors.New("invalid user id")
	}

	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if input.Email != nil && *input.Email != "" {
		user.Email = *input.Email
	}
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

	// If password is provided, hash and update it
	if input.Password != nil && *input.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(*input.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, errors.New("failed to hash password")
		}
		user.PasswordHash = pgtype.Text{
			String: string(hashedPassword),
			Valid:  true,
		}
	}

	updatedUser, err := s.repo.Update(ctx, *user)
	if err != nil {
		return nil, err
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

func (s *UserService) DeleteUser(ctx context.Context, userID string) error {
	id := pgtype.UUID{}
	if err := id.Scan(userID); err != nil {
		return errors.New("invalid user id")
	}

	return s.repo.Delete(ctx, id)
}
