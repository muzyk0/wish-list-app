package http

import (
	"context"
	"errors"
	"testing"

	usermodels "wish-list/internal/domain/user/models"
	userrepo "wish-list/internal/domain/user/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type userRepoMock struct {
	getByEmailFunc func(ctx context.Context, email string) (*usermodels.User, error)
	createFunc     func(ctx context.Context, user usermodels.User) (*usermodels.User, error)
	updateFunc     func(ctx context.Context, user usermodels.User) (*usermodels.User, error)
}

func (m *userRepoMock) GetByEmail(ctx context.Context, email string) (*usermodels.User, error) {
	return m.getByEmailFunc(ctx, email)
}

func (m *userRepoMock) Create(ctx context.Context, user usermodels.User) (*usermodels.User, error) {
	return m.createFunc(ctx, user)
}

func (m *userRepoMock) Update(ctx context.Context, user usermodels.User) (*usermodels.User, error) {
	return m.updateFunc(ctx, user)
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

func TestOAuthHandler_findOrCreateUser_LinksGuestReservations_ExistingUser(t *testing.T) {
	existingUserID := uuid.New()
	existingUser := &usermodels.User{
		ID:    pgtype.UUID{Bytes: existingUserID, Valid: true},
		Email: "oauth@example.com",
		IsVerified: pgtype.Bool{
			Bool:  true,
			Valid: true,
		},
		AvatarUrl: pgtype.Text{
			String: "",
			Valid:  false,
		},
	}

	repo := &userRepoMock{
		getByEmailFunc: func(ctx context.Context, email string) (*usermodels.User, error) {
			return existingUser, nil
		},
		createFunc: func(ctx context.Context, user usermodels.User) (*usermodels.User, error) {
			t.Fatal("create should not be called for existing user")
			return nil, nil
		},
		updateFunc: func(ctx context.Context, user usermodels.User) (*usermodels.User, error) {
			return &user, nil
		},
	}
	linker := &guestReservationLinkerMock{}
	handler := &OAuthHandler{
		userRepo:          repo,
		reservationLinker: linker,
	}

	user, err := handler.findOrCreateUser(context.Background(), "oauth@example.com", "John", "Doe", "https://example.com/avatar.png", true)

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Len(t, linker.calls, 1)
	assert.Equal(t, "oauth@example.com", linker.calls[0].guestEmail)
	assert.Equal(t, existingUser.ID, linker.calls[0].userID)
}

func TestOAuthHandler_findOrCreateUser_LinksGuestReservations_NewUser(t *testing.T) {
	createdUserID := uuid.New()
	repo := &userRepoMock{
		getByEmailFunc: func(ctx context.Context, email string) (*usermodels.User, error) {
			return nil, userrepo.ErrUserNotFound
		},
		createFunc: func(ctx context.Context, user usermodels.User) (*usermodels.User, error) {
			return &usermodels.User{
				ID:        pgtype.UUID{Bytes: createdUserID, Valid: true},
				Email:     user.Email,
				FirstName: user.FirstName,
				LastName:  user.LastName,
				AvatarUrl: user.AvatarUrl,
				IsVerified: pgtype.Bool{
					Bool:  true,
					Valid: true,
				},
			}, nil
		},
		updateFunc: func(ctx context.Context, user usermodels.User) (*usermodels.User, error) {
			return &user, nil
		},
	}
	linker := &guestReservationLinkerMock{}
	handler := &OAuthHandler{
		userRepo:          repo,
		reservationLinker: linker,
	}

	user, err := handler.findOrCreateUser(context.Background(), "new@example.com", "Jane", "Smith", "", true)

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Len(t, linker.calls, 1)
	assert.Equal(t, "new@example.com", linker.calls[0].guestEmail)
	assert.Equal(t, createdUserID, uuid.UUID(linker.calls[0].userID.Bytes))
}

func TestOAuthHandler_findOrCreateUser_LinkingErrorIsNonFatal(t *testing.T) {
	existingUser := &usermodels.User{
		ID:    pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Email: "oauth@example.com",
		IsVerified: pgtype.Bool{
			Bool:  true,
			Valid: true,
		},
	}

	repo := &userRepoMock{
		getByEmailFunc: func(ctx context.Context, email string) (*usermodels.User, error) {
			return existingUser, nil
		},
		createFunc: func(ctx context.Context, user usermodels.User) (*usermodels.User, error) {
			return nil, nil
		},
		updateFunc: func(ctx context.Context, user usermodels.User) (*usermodels.User, error) {
			return &user, nil
		},
	}
	linker := &guestReservationLinkerMock{
		linkFunc: func(ctx context.Context, guestEmail string, userID pgtype.UUID) (int, error) {
			return 0, errors.New("link failure")
		},
	}
	handler := &OAuthHandler{
		userRepo:          repo,
		reservationLinker: linker,
	}

	user, err := handler.findOrCreateUser(context.Background(), "oauth@example.com", "John", "Doe", "", true)

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Len(t, linker.calls, 1)
}

func TestOAuthHandler_findOrCreateUser_InvalidEmail(t *testing.T) {
	repo := &userRepoMock{
		getByEmailFunc: func(ctx context.Context, email string) (*usermodels.User, error) {
			t.Fatal("GetByEmail should not be called for invalid email")
			return nil, nil
		},
		createFunc: func(ctx context.Context, user usermodels.User) (*usermodels.User, error) {
			t.Fatal("Create should not be called for invalid email")
			return nil, nil
		},
		updateFunc: func(ctx context.Context, user usermodels.User) (*usermodels.User, error) {
			return &user, nil
		},
	}
	handler := &OAuthHandler{
		userRepo: repo,
	}

	user, err := handler.findOrCreateUser(context.Background(), "not-an-email", "John", "Doe", "", false)

	require.Error(t, err)
	assert.Nil(t, user)
}

func TestOAuthHandler_findOrCreateUser_DoesNotLinkWhenEmailNotVerified(t *testing.T) {
	existingUser := &usermodels.User{
		ID:    pgtype.UUID{Bytes: uuid.New(), Valid: true},
		Email: "oauth@example.com",
		IsVerified: pgtype.Bool{
			Bool:  false,
			Valid: true,
		},
	}

	repo := &userRepoMock{
		getByEmailFunc: func(ctx context.Context, email string) (*usermodels.User, error) {
			return existingUser, nil
		},
		createFunc: func(ctx context.Context, user usermodels.User) (*usermodels.User, error) {
			return nil, nil
		},
		updateFunc: func(ctx context.Context, user usermodels.User) (*usermodels.User, error) {
			return &user, nil
		},
	}
	linker := &guestReservationLinkerMock{}
	handler := &OAuthHandler{
		userRepo:          repo,
		reservationLinker: linker,
	}

	user, err := handler.findOrCreateUser(context.Background(), "oauth@example.com", "John", "Doe", "", false)

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Len(t, linker.calls, 0)
}
