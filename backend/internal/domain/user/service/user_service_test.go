package service

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"wish-list/internal/domain/user/models"
	"wish-list/internal/domain/user/repository"
	"wish-list/internal/pkg/logger"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func init() {
	logger.Initialize("test")
}

// --- helpers ---

// testHashPassword hashes a plaintext password with a low cost for fast tests.
func testHashPassword(t *testing.T, password string) string {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)
	return string(hash)
}

// testUUID returns a deterministic UUID string suitable for test fixtures.
func testUUID() string {
	return uuid.New().String()
}

// pgUUID converts a UUID string into a pgtype.UUID.
func pgUUID(t *testing.T, id string) pgtype.UUID {
	t.Helper()
	var pgID pgtype.UUID
	err := pgID.Scan(id)
	require.NoError(t, err)
	return pgID
}

// pgText builds a valid pgtype.Text from a string.
func pgText(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

// makeDBUser constructs a models.User with the given fields for test fixtures.
func makeDBUser(id pgtype.UUID, email, passwordHash, firstName, lastName, avatarURL string) models.User {
	return models.User{
		ID:           id,
		Email:        email,
		PasswordHash: pgText(passwordHash),
		FirstName:    pgText(firstName),
		LastName:     pgText(lastName),
		AvatarUrl:    pgText(avatarURL),
	}
}

type guestReservationLinkerMock struct {
	linkFunc func(ctx context.Context, guestEmail string, userID pgtype.UUID) (int, error)
	calls    []struct {
		guestEmail string
		userID     pgtype.UUID
	}
}

func (m *guestReservationLinkerMock) LinkGuestReservationsToUserByEmail(ctx context.Context, guestEmail string, userID pgtype.UUID) (int, error) {
	m.calls = append(m.calls, struct {
		guestEmail string
		userID     pgtype.UUID
	}{
		guestEmail: guestEmail,
		userID:     userID,
	})

	if m.linkFunc == nil {
		return 0, nil
	}

	return m.linkFunc(ctx, guestEmail, userID)
}

// --- Register tests ---

func TestUserService_Register(t *testing.T) {
	t.Run("returns ErrCredentialsRequired when email is empty", func(t *testing.T) {
		svc := NewUserService(&UserRepositoryInterfaceMock{})

		_, err := svc.Register(context.Background(), RegisterUserInput{
			Email:    "",
			Password: "secret123",
		})

		assert.ErrorIs(t, err, ErrCredentialsRequired)
	})

	t.Run("returns ErrCredentialsRequired when password is empty", func(t *testing.T) {
		svc := NewUserService(&UserRepositoryInterfaceMock{})

		_, err := svc.Register(context.Background(), RegisterUserInput{
			Email:    "user@example.com",
			Password: "",
		})

		assert.ErrorIs(t, err, ErrCredentialsRequired)
	})

	t.Run("returns ErrUserAlreadyExists when email is taken", func(t *testing.T) {
		existingID := pgUUID(t, testUUID())
		existingUser := makeDBUser(existingID, "user@example.com", "hash", "John", "Doe", "")

		mockRepo := &UserRepositoryInterfaceMock{
			GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return &existingUser, nil
			},
		}
		svc := NewUserService(mockRepo)

		_, err := svc.Register(context.Background(), RegisterUserInput{
			Email:    "user@example.com",
			Password: "secret123",
		})

		require.ErrorIs(t, err, ErrUserAlreadyExists)
		assert.Len(t, mockRepo.GetByEmailCalls(), 1)
		assert.Equal(t, "user@example.com", mockRepo.GetByEmailCalls()[0].Email)
	})

	t.Run("propagates unexpected repository error from GetByEmail", func(t *testing.T) {
		dbErr := errors.New("connection refused")
		mockRepo := &UserRepositoryInterfaceMock{
			GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, dbErr
			},
		}
		svc := NewUserService(mockRepo)

		_, err := svc.Register(context.Background(), RegisterUserInput{
			Email:    "user@example.com",
			Password: "secret123",
		})

		require.Error(t, err)
		require.NotErrorIs(t, err, ErrUserAlreadyExists)
		require.NotErrorIs(t, err, ErrCredentialsRequired)
	})

	t.Run("successful registration hashes password and creates user", func(t *testing.T) {
		createdID := pgUUID(t, testUUID())

		mockRepo := &UserRepositoryInterfaceMock{
			GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, repository.ErrUserNotFound
			},
			CreateFunc: func(ctx context.Context, user models.User) (*models.User, error) {
				// Verify password was hashed (not stored in plain text)
				assert.NotEqual(t, "secret123", user.PasswordHash.String)
				assert.True(t, user.PasswordHash.Valid)
				err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash.String), []byte("secret123"))
				require.NoError(t, err, "password hash should validate against original password")

				// Verify email and profile fields
				assert.Equal(t, "user@example.com", user.Email)
				assert.Equal(t, "John", user.FirstName.String)
				assert.True(t, user.FirstName.Valid)
				assert.Equal(t, "Doe", user.LastName.String)
				assert.True(t, user.LastName.Valid)

				created := makeDBUser(createdID, user.Email, user.PasswordHash.String, "John", "Doe", "")
				return &created, nil
			},
		}
		svc := NewUserService(mockRepo)

		output, err := svc.Register(context.Background(), RegisterUserInput{
			Email:     "user@example.com",
			Password:  "secret123",
			FirstName: "John",
			LastName:  "Doe",
		})

		require.NoError(t, err)
		require.NotNil(t, output)
		assert.Equal(t, createdID.String(), output.ID)
		assert.Equal(t, "user@example.com", output.Email)
		assert.Equal(t, "John", output.FirstName)
		assert.Equal(t, "Doe", output.LastName)

		assert.Len(t, mockRepo.GetByEmailCalls(), 1)
		assert.Len(t, mockRepo.CreateCalls(), 1)
	})

	t.Run("propagates Create error", func(t *testing.T) {
		mockRepo := &UserRepositoryInterfaceMock{
			GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, repository.ErrUserNotFound
			},
			CreateFunc: func(ctx context.Context, user models.User) (*models.User, error) {
				return nil, errors.New("database write failure")
			},
		}
		svc := NewUserService(mockRepo)

		_, err := svc.Register(context.Background(), RegisterUserInput{
			Email:    "user@example.com",
			Password: "secret123",
		})

		require.Error(t, err)
		assert.Len(t, mockRepo.CreateCalls(), 1)
	})

	t.Run("links guest reservations by email only for verified user", func(t *testing.T) {
		createdID := pgUUID(t, testUUID())
		linker := &guestReservationLinkerMock{
			linkFunc: func(ctx context.Context, guestEmail string, userID pgtype.UUID) (int, error) {
				return 2, nil
			},
		}

		mockRepo := &UserRepositoryInterfaceMock{
			GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, repository.ErrUserNotFound
			},
			CreateFunc: func(ctx context.Context, user models.User) (*models.User, error) {
				created := makeDBUser(createdID, user.Email, user.PasswordHash.String, "John", "Doe", "")
				created.IsVerified = pgtype.Bool{Bool: true, Valid: true}
				return &created, nil
			},
		}
		svc := NewUserService(mockRepo, linker)

		output, err := svc.Register(context.Background(), RegisterUserInput{
			Email:     "user@example.com",
			Password:  "secret123",
			FirstName: "John",
			LastName:  "Doe",
		})

		require.NoError(t, err)
		require.NotNil(t, output)
		require.Len(t, linker.calls, 1)
		assert.Equal(t, "user@example.com", linker.calls[0].guestEmail)
		assert.Equal(t, createdID, linker.calls[0].userID)
	})

	t.Run("does not link guest reservations for unverified user", func(t *testing.T) {
		createdID := pgUUID(t, testUUID())
		linker := &guestReservationLinkerMock{}

		mockRepo := &UserRepositoryInterfaceMock{
			GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, repository.ErrUserNotFound
			},
			CreateFunc: func(ctx context.Context, user models.User) (*models.User, error) {
				created := makeDBUser(createdID, user.Email, user.PasswordHash.String, "", "", "")
				created.IsVerified = pgtype.Bool{Bool: false, Valid: true}
				return &created, nil
			},
		}
		svc := NewUserService(mockRepo, linker)

		output, err := svc.Register(context.Background(), RegisterUserInput{
			Email:    "user@example.com",
			Password: "secret123",
		})

		require.NoError(t, err)
		require.NotNil(t, output)
		require.Len(t, linker.calls, 0)
	})

	t.Run("does not fail registration when guest reservation linking fails", func(t *testing.T) {
		createdID := pgUUID(t, testUUID())
		linker := &guestReservationLinkerMock{
			linkFunc: func(ctx context.Context, guestEmail string, userID pgtype.UUID) (int, error) {
				return 0, errors.New("linking failed")
			},
		}

		mockRepo := &UserRepositoryInterfaceMock{
			GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, repository.ErrUserNotFound
			},
			CreateFunc: func(ctx context.Context, user models.User) (*models.User, error) {
				created := makeDBUser(createdID, user.Email, user.PasswordHash.String, "", "", "")
				created.IsVerified = pgtype.Bool{Bool: true, Valid: true}
				return &created, nil
			},
		}
		svc := NewUserService(mockRepo, linker)

		output, err := svc.Register(context.Background(), RegisterUserInput{
			Email:    "user@example.com",
			Password: "secret123",
		})

		require.NoError(t, err)
		require.NotNil(t, output)
		require.Len(t, linker.calls, 1)
	})

	t.Run("empty optional fields produce invalid pgtype.Text", func(t *testing.T) {
		createdID := pgUUID(t, testUUID())

		mockRepo := &UserRepositoryInterfaceMock{
			GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, repository.ErrUserNotFound
			},
			CreateFunc: func(ctx context.Context, user models.User) (*models.User, error) {
				// Empty optional fields should NOT be Valid
				assert.False(t, user.FirstName.Valid)
				assert.False(t, user.LastName.Valid)
				assert.False(t, user.AvatarUrl.Valid)

				created := models.User{ID: createdID, Email: user.Email, PasswordHash: user.PasswordHash}
				return &created, nil
			},
		}
		svc := NewUserService(mockRepo)

		output, err := svc.Register(context.Background(), RegisterUserInput{
			Email:    "user@example.com",
			Password: "secret123",
		})

		require.NoError(t, err)
		assert.Empty(t, output.FirstName)
		assert.Empty(t, output.LastName)
	})
}

// --- Login tests ---

func TestUserService_Login(t *testing.T) {
	t.Run("returns ErrCredentialsRequired when email is empty", func(t *testing.T) {
		svc := NewUserService(&UserRepositoryInterfaceMock{})

		_, err := svc.Login(context.Background(), LoginUserInput{
			Email:    "",
			Password: "secret",
		})

		assert.ErrorIs(t, err, ErrCredentialsRequired)
	})

	t.Run("returns ErrCredentialsRequired when password is empty", func(t *testing.T) {
		svc := NewUserService(&UserRepositoryInterfaceMock{})

		_, err := svc.Login(context.Background(), LoginUserInput{
			Email:    "user@example.com",
			Password: "",
		})

		assert.ErrorIs(t, err, ErrCredentialsRequired)
	})

	t.Run("returns ErrInvalidCredentials when user not found", func(t *testing.T) {
		mockRepo := &UserRepositoryInterfaceMock{
			GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, repository.ErrUserNotFound
			},
		}
		svc := NewUserService(mockRepo)

		_, err := svc.Login(context.Background(), LoginUserInput{
			Email:    "unknown@example.com",
			Password: "secret",
		})

		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("returns ErrInvalidCredentials when password hash is invalid (not Valid)", func(t *testing.T) {
		userID := pgUUID(t, testUUID())
		user := models.User{
			ID:    userID,
			Email: "user@example.com",
			PasswordHash: pgtype.Text{
				String: "",
				Valid:  false,
			},
		}

		mockRepo := &UserRepositoryInterfaceMock{
			GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return &user, nil
			},
		}
		svc := NewUserService(mockRepo)

		_, err := svc.Login(context.Background(), LoginUserInput{
			Email:    "user@example.com",
			Password: "secret",
		})

		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("returns ErrInvalidCredentials when password does not match", func(t *testing.T) {
		userID := pgUUID(t, testUUID())
		hash := testHashPassword(t, "correct-password")
		user := makeDBUser(userID, "user@example.com", hash, "John", "Doe", "")

		mockRepo := &UserRepositoryInterfaceMock{
			GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return &user, nil
			},
		}
		svc := NewUserService(mockRepo)

		_, err := svc.Login(context.Background(), LoginUserInput{
			Email:    "user@example.com",
			Password: "wrong-password",
		})

		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})

	t.Run("successful login returns user output", func(t *testing.T) {
		userID := pgUUID(t, testUUID())
		hash := testHashPassword(t, "correct-password")
		user := makeDBUser(userID, "user@example.com", hash, "John", "Doe", "https://avatar.url/img.png")

		mockRepo := &UserRepositoryInterfaceMock{
			GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return &user, nil
			},
		}
		svc := NewUserService(mockRepo)

		output, err := svc.Login(context.Background(), LoginUserInput{
			Email:    "user@example.com",
			Password: "correct-password",
		})

		require.NoError(t, err)
		require.NotNil(t, output)
		assert.Equal(t, userID.String(), output.ID)
		assert.Equal(t, "user@example.com", output.Email)
		assert.Equal(t, "John", output.FirstName)
		assert.Equal(t, "Doe", output.LastName)
		assert.Equal(t, "https://avatar.url/img.png", output.AvatarUrl)
	})

	t.Run("returns ErrInvalidCredentials on any GetByEmail error", func(t *testing.T) {
		mockRepo := &UserRepositoryInterfaceMock{
			GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, errors.New("database timeout")
			},
		}
		svc := NewUserService(mockRepo)

		_, err := svc.Login(context.Background(), LoginUserInput{
			Email:    "user@example.com",
			Password: "secret",
		})

		// Login converts ALL repo errors to ErrInvalidCredentials to avoid user enumeration
		assert.ErrorIs(t, err, ErrInvalidCredentials)
	})
}

// --- GetUser tests ---

func TestUserService_GetUser(t *testing.T) {
	t.Run("returns ErrInvalidUserID for invalid UUID", func(t *testing.T) {
		svc := NewUserService(&UserRepositoryInterfaceMock{})

		_, err := svc.GetUser(context.Background(), "not-a-uuid")

		assert.ErrorIs(t, err, ErrInvalidUserID)
	})

	t.Run("returns ErrInvalidUserID for empty string", func(t *testing.T) {
		svc := NewUserService(&UserRepositoryInterfaceMock{})

		_, err := svc.GetUser(context.Background(), "")

		assert.ErrorIs(t, err, ErrInvalidUserID)
	})

	t.Run("returns ErrUserNotFound when repo returns ErrUserNotFound", func(t *testing.T) {
		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return nil, repository.ErrUserNotFound
			},
		}
		svc := NewUserService(mockRepo)
		validID := testUUID()

		_, err := svc.GetUser(context.Background(), validID)

		assert.ErrorIs(t, err, ErrUserNotFound)
	})

	t.Run("propagates unexpected repo error", func(t *testing.T) {
		dbErr := errors.New("connection lost")
		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return nil, dbErr
			},
		}
		svc := NewUserService(mockRepo)

		_, err := svc.GetUser(context.Background(), testUUID())

		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("successful GetUser returns user output", func(t *testing.T) {
		userIDStr := testUUID()
		userID := pgUUID(t, userIDStr)
		user := makeDBUser(userID, "user@example.com", "hash", "Jane", "Smith", "https://img.url/a.png")

		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				assert.Equal(t, userID, id)
				return &user, nil
			},
		}
		svc := NewUserService(mockRepo)

		output, err := svc.GetUser(context.Background(), userIDStr)

		require.NoError(t, err)
		require.NotNil(t, output)
		assert.Equal(t, userIDStr, output.ID)
		assert.Equal(t, "user@example.com", output.Email)
		assert.Equal(t, "Jane", output.FirstName)
		assert.Equal(t, "Smith", output.LastName)
		assert.Equal(t, "https://img.url/a.png", output.AvatarUrl)
		assert.Len(t, mockRepo.GetByIDCalls(), 1)
	})
}

// --- UpdateProfile tests ---

func TestUserService_UpdateProfile(t *testing.T) {
	t.Run("returns ErrInvalidUserID for invalid UUID", func(t *testing.T) {
		svc := NewUserService(&UserRepositoryInterfaceMock{})

		_, err := svc.UpdateProfile(context.Background(), "bad-id", UpdateProfileInput{})

		assert.ErrorIs(t, err, ErrInvalidUserID)
	})

	t.Run("propagates GetByID error", func(t *testing.T) {
		repoErr := errors.New("db error")
		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return nil, repoErr
			},
		}
		svc := NewUserService(mockRepo)

		_, err := svc.UpdateProfile(context.Background(), testUUID(), UpdateProfileInput{})

		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("updates only provided fields", func(t *testing.T) {
		userIDStr := testUUID()
		userID := pgUUID(t, userIDStr)
		originalUser := makeDBUser(userID, "user@example.com", "hash", "OldFirst", "OldLast", "old-avatar.png")

		newFirst := "NewFirst"
		newAvatar := "new-avatar.png"

		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return &originalUser, nil
			},
			UpdateFunc: func(ctx context.Context, user models.User) (*models.User, error) {
				// FirstName should be updated
				assert.Equal(t, "NewFirst", user.FirstName.String)
				assert.True(t, user.FirstName.Valid)

				// LastName should remain unchanged (not provided in input)
				assert.Equal(t, "OldLast", user.LastName.String)

				// AvatarUrl should be updated
				assert.Equal(t, "new-avatar.png", user.AvatarUrl.String)
				assert.True(t, user.AvatarUrl.Valid)

				// Email should remain unchanged
				assert.Equal(t, "user@example.com", user.Email)

				return &user, nil
			},
		}
		svc := NewUserService(mockRepo)

		output, err := svc.UpdateProfile(context.Background(), userIDStr, UpdateProfileInput{
			FirstName: &newFirst,
			AvatarUrl: &newAvatar,
		})

		require.NoError(t, err)
		require.NotNil(t, output)
		assert.Equal(t, "NewFirst", output.FirstName)
		assert.Equal(t, "OldLast", output.LastName)
		assert.Equal(t, "new-avatar.png", output.AvatarUrl)
		assert.Len(t, mockRepo.GetByIDCalls(), 1)
		assert.Len(t, mockRepo.UpdateCalls(), 1)
	})

	t.Run("updates all profile fields", func(t *testing.T) {
		userIDStr := testUUID()
		userID := pgUUID(t, userIDStr)
		originalUser := makeDBUser(userID, "user@example.com", "hash", "Old", "Name", "old.png")

		newFirst := "New"
		newLast := "Person"
		newAvatar := "new.png"

		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return &originalUser, nil
			},
			UpdateFunc: func(ctx context.Context, user models.User) (*models.User, error) {
				return &user, nil
			},
		}
		svc := NewUserService(mockRepo)

		output, err := svc.UpdateProfile(context.Background(), userIDStr, UpdateProfileInput{
			FirstName: &newFirst,
			LastName:  &newLast,
			AvatarUrl: &newAvatar,
		})

		require.NoError(t, err)
		assert.Equal(t, "New", output.FirstName)
		assert.Equal(t, "Person", output.LastName)
		assert.Equal(t, "new.png", output.AvatarUrl)
	})

	t.Run("no fields provided keeps original values", func(t *testing.T) {
		userIDStr := testUUID()
		userID := pgUUID(t, userIDStr)
		originalUser := makeDBUser(userID, "user@example.com", "hash", "Keep", "These", "keep.png")

		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return &originalUser, nil
			},
			UpdateFunc: func(ctx context.Context, user models.User) (*models.User, error) {
				assert.Equal(t, "Keep", user.FirstName.String)
				assert.Equal(t, "These", user.LastName.String)
				assert.Equal(t, "keep.png", user.AvatarUrl.String)
				return &user, nil
			},
		}
		svc := NewUserService(mockRepo)

		output, err := svc.UpdateProfile(context.Background(), userIDStr, UpdateProfileInput{})

		require.NoError(t, err)
		assert.Equal(t, "Keep", output.FirstName)
		assert.Equal(t, "These", output.LastName)
	})

	t.Run("propagates Update error", func(t *testing.T) {
		userIDStr := testUUID()
		userID := pgUUID(t, userIDStr)
		originalUser := makeDBUser(userID, "user@example.com", "hash", "F", "L", "")

		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return &originalUser, nil
			},
			UpdateFunc: func(ctx context.Context, user models.User) (*models.User, error) {
				return nil, errors.New("write failure")
			},
		}
		svc := NewUserService(mockRepo)

		_, err := svc.UpdateProfile(context.Background(), userIDStr, UpdateProfileInput{})

		require.Error(t, err)
		assert.Contains(t, err.Error(), "write failure")
	})
}

// --- ChangeEmail tests ---

func TestUserService_ChangeEmail(t *testing.T) {
	const currentPassword = "current-password"

	t.Run("returns ErrInvalidUserID for invalid UUID", func(t *testing.T) {
		svc := NewUserService(&UserRepositoryInterfaceMock{})

		err := svc.ChangeEmail(context.Background(), "bad-id", currentPassword, "new@example.com")

		assert.ErrorIs(t, err, ErrInvalidUserID)
	})

	t.Run("propagates GetByID error", func(t *testing.T) {
		repoErr := errors.New("db unavailable")
		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return nil, repoErr
			},
		}
		svc := NewUserService(mockRepo)

		err := svc.ChangeEmail(context.Background(), testUUID(), currentPassword, "new@example.com")

		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("returns ErrInvalidPassword when password hash is not valid", func(t *testing.T) {
		userIDStr := testUUID()
		userID := pgUUID(t, userIDStr)
		user := models.User{
			ID:           userID,
			Email:        "old@example.com",
			PasswordHash: pgtype.Text{Valid: false},
		}

		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return &user, nil
			},
		}
		svc := NewUserService(mockRepo)

		err := svc.ChangeEmail(context.Background(), userIDStr, currentPassword, "new@example.com")

		assert.ErrorIs(t, err, ErrInvalidPassword)
	})

	t.Run("returns ErrInvalidPassword when current password is wrong", func(t *testing.T) {
		userIDStr := testUUID()
		userID := pgUUID(t, userIDStr)
		hash := testHashPassword(t, "actual-password")
		user := makeDBUser(userID, "old@example.com", hash, "F", "L", "")

		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return &user, nil
			},
		}
		svc := NewUserService(mockRepo)

		err := svc.ChangeEmail(context.Background(), userIDStr, "wrong-password", "new@example.com")

		assert.ErrorIs(t, err, ErrInvalidPassword)
	})

	t.Run("returns ErrUserAlreadyExists when new email is taken by another user", func(t *testing.T) {
		userIDStr := testUUID()
		userID := pgUUID(t, userIDStr)
		hash := testHashPassword(t, currentPassword)
		user := makeDBUser(userID, "old@example.com", hash, "F", "L", "")

		otherUserID := pgUUID(t, testUUID())
		otherUser := makeDBUser(otherUserID, "new@example.com", "other-hash", "O", "U", "")

		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return &user, nil
			},
			GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return &otherUser, nil
			},
		}
		svc := NewUserService(mockRepo)

		err := svc.ChangeEmail(context.Background(), userIDStr, currentPassword, "new@example.com")

		assert.ErrorIs(t, err, ErrUserAlreadyExists)
	})

	t.Run("allows changing to same email (own account)", func(t *testing.T) {
		userIDStr := testUUID()
		userID := pgUUID(t, userIDStr)
		hash := testHashPassword(t, currentPassword)
		user := makeDBUser(userID, "same@example.com", hash, "F", "L", "")

		// GetByEmail returns the same user (same ID), so it should be allowed
		sameUser := makeDBUser(userID, "same@example.com", hash, "F", "L", "")

		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return &user, nil
			},
			GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return &sameUser, nil
			},
			UpdateFunc: func(ctx context.Context, u models.User) (*models.User, error) {
				return &u, nil
			},
		}
		svc := NewUserService(mockRepo)

		err := svc.ChangeEmail(context.Background(), userIDStr, currentPassword, "same@example.com")

		assert.NoError(t, err)
	})

	t.Run("successful email change", func(t *testing.T) {
		userIDStr := testUUID()
		userID := pgUUID(t, userIDStr)
		hash := testHashPassword(t, currentPassword)
		user := makeDBUser(userID, "old@example.com", hash, "F", "L", "")

		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return &user, nil
			},
			GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, repository.ErrUserNotFound
			},
			UpdateFunc: func(ctx context.Context, u models.User) (*models.User, error) {
				assert.Equal(t, "new@example.com", u.Email)
				return &u, nil
			},
		}
		svc := NewUserService(mockRepo)

		err := svc.ChangeEmail(context.Background(), userIDStr, currentPassword, "new@example.com")

		require.NoError(t, err)
		assert.Len(t, mockRepo.UpdateCalls(), 1)
		assert.Equal(t, "new@example.com", mockRepo.UpdateCalls()[0].User.Email)
	})

	t.Run("propagates Update error", func(t *testing.T) {
		userIDStr := testUUID()
		userID := pgUUID(t, userIDStr)
		hash := testHashPassword(t, currentPassword)
		user := makeDBUser(userID, "old@example.com", hash, "F", "L", "")

		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return &user, nil
			},
			GetByEmailFunc: func(ctx context.Context, email string) (*models.User, error) {
				return nil, repository.ErrUserNotFound
			},
			UpdateFunc: func(ctx context.Context, u models.User) (*models.User, error) {
				return nil, errors.New("update failed")
			},
		}
		svc := NewUserService(mockRepo)

		err := svc.ChangeEmail(context.Background(), userIDStr, currentPassword, "new@example.com")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "update failed")
	})
}

// --- ChangePassword tests ---

func TestUserService_ChangePassword(t *testing.T) {
	const currentPassword = "old-password"
	const newPassword = "new-password"

	t.Run("returns ErrInvalidUserID for invalid UUID", func(t *testing.T) {
		svc := NewUserService(&UserRepositoryInterfaceMock{})

		err := svc.ChangePassword(context.Background(), "not-uuid", currentPassword, newPassword)

		assert.ErrorIs(t, err, ErrInvalidUserID)
	})

	t.Run("propagates GetByID error", func(t *testing.T) {
		repoErr := errors.New("db down")
		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return nil, repoErr
			},
		}
		svc := NewUserService(mockRepo)

		err := svc.ChangePassword(context.Background(), testUUID(), currentPassword, newPassword)

		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("returns ErrInvalidPassword when password hash is not valid", func(t *testing.T) {
		userIDStr := testUUID()
		userID := pgUUID(t, userIDStr)
		user := models.User{
			ID:           userID,
			Email:        "user@example.com",
			PasswordHash: pgtype.Text{Valid: false},
		}

		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return &user, nil
			},
		}
		svc := NewUserService(mockRepo)

		err := svc.ChangePassword(context.Background(), userIDStr, currentPassword, newPassword)

		assert.ErrorIs(t, err, ErrInvalidPassword)
	})

	t.Run("returns ErrInvalidPassword when current password is wrong", func(t *testing.T) {
		userIDStr := testUUID()
		userID := pgUUID(t, userIDStr)
		hash := testHashPassword(t, "actual-password")
		user := makeDBUser(userID, "user@example.com", hash, "F", "L", "")

		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return &user, nil
			},
		}
		svc := NewUserService(mockRepo)

		err := svc.ChangePassword(context.Background(), userIDStr, "wrong-password", newPassword)

		assert.ErrorIs(t, err, ErrInvalidPassword)
	})

	t.Run("successful password change hashes and stores new password", func(t *testing.T) {
		userIDStr := testUUID()
		userID := pgUUID(t, userIDStr)
		hash := testHashPassword(t, currentPassword)
		user := makeDBUser(userID, "user@example.com", hash, "F", "L", "")

		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return &user, nil
			},
			UpdateFunc: func(ctx context.Context, u models.User) (*models.User, error) {
				// The new password should be hashed, not stored in plain text
				assert.True(t, u.PasswordHash.Valid)
				assert.NotEqual(t, newPassword, u.PasswordHash.String)
				assert.NotEqual(t, hash, u.PasswordHash.String) // should differ from old hash

				// Verify the new hash validates against the new password
				err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash.String), []byte(newPassword))
				assert.NoError(t, err, "new hash should validate against new password")

				return &u, nil
			},
		}
		svc := NewUserService(mockRepo)

		err := svc.ChangePassword(context.Background(), userIDStr, currentPassword, newPassword)

		require.NoError(t, err)
		assert.Len(t, mockRepo.UpdateCalls(), 1)
	})

	t.Run("propagates Update error", func(t *testing.T) {
		userIDStr := testUUID()
		userID := pgUUID(t, userIDStr)
		hash := testHashPassword(t, currentPassword)
		user := makeDBUser(userID, "user@example.com", hash, "F", "L", "")

		mockRepo := &UserRepositoryInterfaceMock{
			GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.User, error) {
				return &user, nil
			},
			UpdateFunc: func(ctx context.Context, u models.User) (*models.User, error) {
				return nil, errors.New("write error")
			},
		}
		svc := NewUserService(mockRepo)

		err := svc.ChangePassword(context.Background(), userIDStr, currentPassword, newPassword)

		assert.Error(t, err)
	})
}

// --- DeleteUser tests ---

func TestUserService_DeleteUser(t *testing.T) {
	t.Run("returns ErrInvalidUserID for invalid UUID", func(t *testing.T) {
		svc := NewUserService(&UserRepositoryInterfaceMock{})

		err := svc.DeleteUser(context.Background(), "invalid")

		assert.ErrorIs(t, err, ErrInvalidUserID)
	})

	t.Run("returns ErrInvalidUserID for empty string", func(t *testing.T) {
		svc := NewUserService(&UserRepositoryInterfaceMock{})

		err := svc.DeleteUser(context.Background(), "")

		assert.ErrorIs(t, err, ErrInvalidUserID)
	})

	t.Run("successful delete delegates to repo", func(t *testing.T) {
		userIDStr := testUUID()
		expectedPgID := pgUUID(t, userIDStr)

		mockRepo := &UserRepositoryInterfaceMock{
			DeleteFunc: func(ctx context.Context, id pgtype.UUID) error {
				assert.Equal(t, expectedPgID, id)
				return nil
			},
		}
		svc := NewUserService(mockRepo)

		err := svc.DeleteUser(context.Background(), userIDStr)

		require.NoError(t, err)
		assert.Len(t, mockRepo.DeleteCalls(), 1)
	})

	t.Run("propagates repo error", func(t *testing.T) {
		repoErr := fmt.Errorf("failed to delete user: %w", repository.ErrUserNotFound)
		mockRepo := &UserRepositoryInterfaceMock{
			DeleteFunc: func(ctx context.Context, id pgtype.UUID) error {
				return repoErr
			},
		}
		svc := NewUserService(mockRepo)

		err := svc.DeleteUser(context.Background(), testUUID())

		require.Error(t, err)
		assert.ErrorIs(t, err, repository.ErrUserNotFound)
	})
}
