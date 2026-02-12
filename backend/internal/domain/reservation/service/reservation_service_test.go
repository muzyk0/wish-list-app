package service

import (
	"context"
	"errors"
	"testing"
	"time"

	itemmodels "wish-list/internal/domain/item/models"
	"wish-list/internal/domain/reservation/models"
	"wish-list/internal/domain/reservation/repository"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReservationService_GetReservationStatus(t *testing.T) {
	t.Run("available gift item", func(t *testing.T) {
		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}

		mockGiftItemRepo := &GiftItemRepositoryInterfaceMock{
			GetPublicWishListGiftItemsFunc: func(ctx context.Context, publicSlug string) ([]*itemmodels.GiftItem, error) {
				return []*itemmodels.GiftItem{{ID: giftItemID}}, nil
			},
		}
		mockRepo := &ReservationRepositoryInterfaceMock{
			GetActiveReservationForGiftItemFunc: func(ctx context.Context, id pgtype.UUID) (*models.Reservation, error) {
				return nil, nil
			},
		}

		service := NewReservationService(mockRepo, mockGiftItemRepo)
		status, err := service.GetReservationStatus(context.Background(), "public-slug", giftItemID.String())

		require.NoError(t, err)
		assert.False(t, status.IsReserved)
		assert.Equal(t, "available", status.Status)
		assert.Len(t, mockGiftItemRepo.GetPublicWishListGiftItemsCalls(), 1)
		assert.Len(t, mockRepo.GetActiveReservationForGiftItemCalls(), 1)
	})

	t.Run("reserved gift item", func(t *testing.T) {
		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		token := pgtype.UUID{Bytes: [16]byte{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}, Valid: true}

		activeReservation := &models.Reservation{
			ID:               pgtype.UUID{Bytes: [16]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}, Valid: true},
			GiftItemID:       giftItemID,
			ReservationToken: token,
			Status:           "active",
		}

		mockGiftItemRepo := &GiftItemRepositoryInterfaceMock{
			GetPublicWishListGiftItemsFunc: func(ctx context.Context, publicSlug string) ([]*itemmodels.GiftItem, error) {
				return []*itemmodels.GiftItem{{ID: giftItemID}}, nil
			},
		}
		mockRepo := &ReservationRepositoryInterfaceMock{
			GetActiveReservationForGiftItemFunc: func(ctx context.Context, id pgtype.UUID) (*models.Reservation, error) {
				return activeReservation, nil
			},
			ListGuestReservationsWithDetailsFunc: func(ctx context.Context, t pgtype.UUID) ([]repository.ReservationDetail, error) {
				return []repository.ReservationDetail{}, nil
			},
		}

		service := NewReservationService(mockRepo, mockGiftItemRepo)
		status, err := service.GetReservationStatus(context.Background(), "public-slug", giftItemID.String())

		require.NoError(t, err)
		assert.True(t, status.IsReserved)
		assert.Equal(t, "active", status.Status)
	})

	t.Run("invalid gift item id", func(t *testing.T) {
		mockRepo := &ReservationRepositoryInterfaceMock{}
		mockGiftItemRepo := &GiftItemRepositoryInterfaceMock{}

		service := NewReservationService(mockRepo, mockGiftItemRepo)
		status, err := service.GetReservationStatus(context.Background(), "public-slug", "invalid-uuid")

		require.Error(t, err)
		assert.Nil(t, status)
		assert.Empty(t, mockRepo.GetActiveReservationForGiftItemCalls())
	})
}

// T070a: Unit tests for reservation expiration logic
func TestReservationService_ExpirationLogic(t *testing.T) {
	t.Run("expired reservation returns available status", func(t *testing.T) {
		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		token := pgtype.UUID{Bytes: [16]byte{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}, Valid: true}
		reservationID := pgtype.UUID{Bytes: [16]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}, Valid: true}

		expiredReservation := &models.Reservation{
			ID:               reservationID,
			GiftItemID:       giftItemID,
			ReservationToken: token,
			Status:           "active",
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(-1 * time.Hour),
				Valid: true,
			},
		}

		mockGiftItemRepo := &GiftItemRepositoryInterfaceMock{
			GetPublicWishListGiftItemsFunc: func(ctx context.Context, publicSlug string) ([]*itemmodels.GiftItem, error) {
				return []*itemmodels.GiftItem{{ID: giftItemID}}, nil
			},
		}
		mockRepo := &ReservationRepositoryInterfaceMock{
			GetActiveReservationForGiftItemFunc: func(ctx context.Context, id pgtype.UUID) (*models.Reservation, error) {
				return expiredReservation, nil
			},
			UpdateStatusFunc: func(ctx context.Context, resID pgtype.UUID, status string, canceledAt pgtype.Timestamptz, cancelReason pgtype.Text) (*models.Reservation, error) {
				return &models.Reservation{Status: "expired"}, nil
			},
		}

		service := NewReservationService(mockRepo, mockGiftItemRepo)
		status, err := service.GetReservationStatus(context.Background(), "public-slug", giftItemID.String())

		require.NoError(t, err)
		assert.False(t, status.IsReserved)
		assert.Equal(t, "available", status.Status)
		assert.Len(t, mockRepo.UpdateStatusCalls(), 1)
	})

	t.Run("non-expired reservation remains active", func(t *testing.T) {
		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		token := pgtype.UUID{Bytes: [16]byte{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}, Valid: true}

		activeReservation := &models.Reservation{
			ID:               pgtype.UUID{Bytes: [16]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}, Valid: true},
			GiftItemID:       giftItemID,
			ReservationToken: token,
			Status:           "active",
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(24 * time.Hour),
				Valid: true,
			},
		}

		mockGiftItemRepo := &GiftItemRepositoryInterfaceMock{
			GetPublicWishListGiftItemsFunc: func(ctx context.Context, publicSlug string) ([]*itemmodels.GiftItem, error) {
				return []*itemmodels.GiftItem{{ID: giftItemID}}, nil
			},
		}
		mockRepo := &ReservationRepositoryInterfaceMock{
			GetActiveReservationForGiftItemFunc: func(ctx context.Context, id pgtype.UUID) (*models.Reservation, error) {
				return activeReservation, nil
			},
			ListGuestReservationsWithDetailsFunc: func(ctx context.Context, t pgtype.UUID) ([]repository.ReservationDetail, error) {
				return []repository.ReservationDetail{}, nil
			},
		}

		service := NewReservationService(mockRepo, mockGiftItemRepo)
		status, err := service.GetReservationStatus(context.Background(), "public-slug", giftItemID.String())

		require.NoError(t, err)
		assert.True(t, status.IsReserved)
		assert.Equal(t, "active", status.Status)
	})

	t.Run("reservation without expiry date stays active", func(t *testing.T) {
		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		token := pgtype.UUID{Bytes: [16]byte{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}, Valid: true}

		activeReservation := &models.Reservation{
			ID:               pgtype.UUID{Bytes: [16]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}, Valid: true},
			GiftItemID:       giftItemID,
			ReservationToken: token,
			Status:           "active",
			ExpiresAt:        pgtype.Timestamptz{Valid: false},
		}

		mockGiftItemRepo := &GiftItemRepositoryInterfaceMock{
			GetPublicWishListGiftItemsFunc: func(ctx context.Context, publicSlug string) ([]*itemmodels.GiftItem, error) {
				return []*itemmodels.GiftItem{{ID: giftItemID}}, nil
			},
		}
		mockRepo := &ReservationRepositoryInterfaceMock{
			GetActiveReservationForGiftItemFunc: func(ctx context.Context, id pgtype.UUID) (*models.Reservation, error) {
				return activeReservation, nil
			},
			ListGuestReservationsWithDetailsFunc: func(ctx context.Context, t pgtype.UUID) ([]repository.ReservationDetail, error) {
				return []repository.ReservationDetail{}, nil
			},
		}

		service := NewReservationService(mockRepo, mockGiftItemRepo)
		status, err := service.GetReservationStatus(context.Background(), "public-slug", giftItemID.String())

		require.NoError(t, err)
		assert.True(t, status.IsReserved)
		assert.Equal(t, "active", status.Status)
	})
}

// T070b: Unit tests for concurrency controls for simultaneous reservations
func TestReservationService_ConcurrencyControls(t *testing.T) {
	t.Run("create reservation on already reserved item fails", func(t *testing.T) {
		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		wishlistID := pgtype.UUID{Bytes: [16]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}, Valid: true}

		giftItem := &itemmodels.GiftItem{ID: giftItemID}
		existingReservation := &models.Reservation{
			ID:         pgtype.UUID{Valid: true},
			GiftItemID: giftItemID,
			Status:     "active",
		}

		mockGiftItemRepo := &GiftItemRepositoryInterfaceMock{
			GetByWishListFunc: func(ctx context.Context, wlID pgtype.UUID) ([]*itemmodels.GiftItem, error) {
				return []*itemmodels.GiftItem{giftItem}, nil
			},
		}
		mockRepo := &ReservationRepositoryInterfaceMock{
			GetActiveReservationForGiftItemFunc: func(ctx context.Context, id pgtype.UUID) (*models.Reservation, error) {
				return existingReservation, nil
			},
		}

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		guestName := "Test User"
		guestEmail := "test@example.com"
		input := CreateReservationInput{
			WishListID: wishlistID.String(),
			GiftItemID: giftItemID.String(),
			UserID:     pgtype.UUID{Valid: false},
			GuestName:  &guestName,
			GuestEmail: &guestEmail,
		}

		_, err := service.CreateReservation(context.Background(), input)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrGiftItemAlreadyReserved))
	})

	t.Run("authenticated user reservation uses atomic operation", func(t *testing.T) {
		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		userID := pgtype.UUID{Bytes: [16]byte{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}, Valid: true}
		wishlistID := pgtype.UUID{Bytes: [16]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}, Valid: true}

		giftItem := &itemmodels.GiftItem{ID: giftItemID}
		reservedItem := &itemmodels.GiftItem{
			ID:               giftItemID,
			ReservedByUserID: userID,
		}
		createdReservation := &models.Reservation{
			ID:               pgtype.UUID{Bytes: [16]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}, Valid: true},
			GiftItemID:       giftItemID,
			ReservedByUserID: userID,
			Status:           "active",
		}

		mockGiftItemRepo := &GiftItemRepositoryInterfaceMock{
			GetByWishListFunc: func(ctx context.Context, wlID pgtype.UUID) ([]*itemmodels.GiftItem, error) {
				return []*itemmodels.GiftItem{giftItem}, nil
			},
			ReserveIfNotReservedFunc: func(ctx context.Context, giID, uID pgtype.UUID) (*itemmodels.GiftItem, error) {
				return reservedItem, nil
			},
		}
		mockRepo := &ReservationRepositoryInterfaceMock{
			CreateFunc: func(ctx context.Context, reservation models.Reservation) (*models.Reservation, error) {
				return createdReservation, nil
			},
		}

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		input := CreateReservationInput{
			WishListID: wishlistID.String(),
			GiftItemID: giftItemID.String(),
			UserID:     userID,
		}

		reservation, err := service.CreateReservation(context.Background(), input)

		require.NoError(t, err)
		assert.NotNil(t, reservation)
		assert.Equal(t, "active", reservation.Status)
		assert.Len(t, mockGiftItemRepo.ReserveIfNotReservedCalls(), 1)
		assert.Len(t, mockRepo.CreateCalls(), 1)
	})

	t.Run("concurrent reservation attempts are handled atomically", func(t *testing.T) {
		t.Skip("Full concurrency testing requires integration tests with real database")
	})
}

// Test CreateReservation function
func TestReservationService_CreateReservation(t *testing.T) {
	t.Run("successful guest reservation", func(t *testing.T) {
		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		wishlistID := pgtype.UUID{Bytes: [16]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}, Valid: true}

		giftItem := &itemmodels.GiftItem{ID: giftItemID}
		createdReservation := &models.Reservation{
			ID:         pgtype.UUID{Bytes: [16]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}, Valid: true},
			GiftItemID: giftItemID,
			Status:     "active",
		}

		mockGiftItemRepo := &GiftItemRepositoryInterfaceMock{
			GetByWishListFunc: func(ctx context.Context, wlID pgtype.UUID) ([]*itemmodels.GiftItem, error) {
				return []*itemmodels.GiftItem{giftItem}, nil
			},
		}
		mockRepo := &ReservationRepositoryInterfaceMock{
			GetActiveReservationForGiftItemFunc: func(ctx context.Context, id pgtype.UUID) (*models.Reservation, error) {
				return nil, nil
			},
			CreateFunc: func(ctx context.Context, reservation models.Reservation) (*models.Reservation, error) {
				return createdReservation, nil
			},
		}

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		guestName := "Test Guest"
		guestEmail := "guest@example.com"
		input := CreateReservationInput{
			WishListID: wishlistID.String(),
			GiftItemID: giftItemID.String(),
			UserID:     pgtype.UUID{Valid: false},
			GuestName:  &guestName,
			GuestEmail: &guestEmail,
		}

		reservation, err := service.CreateReservation(context.Background(), input)

		require.NoError(t, err)
		assert.NotNil(t, reservation)
		assert.Equal(t, "active", reservation.Status)
	})

	t.Run("guest reservation requires name and email", func(t *testing.T) {
		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		wishlistID := pgtype.UUID{Bytes: [16]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}, Valid: true}

		giftItem := &itemmodels.GiftItem{ID: giftItemID}

		mockGiftItemRepo := &GiftItemRepositoryInterfaceMock{
			GetByWishListFunc: func(ctx context.Context, wlID pgtype.UUID) ([]*itemmodels.GiftItem, error) {
				return []*itemmodels.GiftItem{giftItem}, nil
			},
		}
		mockRepo := &ReservationRepositoryInterfaceMock{
			GetActiveReservationForGiftItemFunc: func(ctx context.Context, id pgtype.UUID) (*models.Reservation, error) {
				return nil, nil
			},
		}

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		input := CreateReservationInput{
			WishListID: wishlistID.String(),
			GiftItemID: giftItemID.String(),
			UserID:     pgtype.UUID{Valid: false},
			GuestName:  nil,
			GuestEmail: nil,
		}

		_, err := service.CreateReservation(context.Background(), input)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrGuestInfoRequired))
	})

	t.Run("invalid gift item id", func(t *testing.T) {
		mockRepo := &ReservationRepositoryInterfaceMock{}
		mockGiftItemRepo := &GiftItemRepositoryInterfaceMock{}

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		input := CreateReservationInput{
			WishListID: "list-123",
			GiftItemID: "invalid-uuid",
			UserID:     pgtype.UUID{Valid: false},
		}

		_, err := service.CreateReservation(context.Background(), input)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrInvalidGiftItemID))
	})
}

// Test CancelReservation function
func TestReservationService_CancelReservation(t *testing.T) {
	t.Run("successful cancellation by guest with token", func(t *testing.T) {
		token := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		giftItemID := pgtype.UUID{Bytes: [16]byte{20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35}, Valid: true}
		wishlistID := pgtype.UUID{Bytes: [16]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}, Valid: true}

		giftItem := &itemmodels.GiftItem{ID: giftItemID}
		canceledReservation := &models.Reservation{
			ID:               pgtype.UUID{Bytes: [16]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}, Valid: true},
			ReservationToken: token,
			Status:           "canceled",
		}

		mockGiftItemRepo := &GiftItemRepositoryInterfaceMock{
			GetByWishListFunc: func(ctx context.Context, wlID pgtype.UUID) ([]*itemmodels.GiftItem, error) {
				return []*itemmodels.GiftItem{giftItem}, nil
			},
		}
		mockRepo := &ReservationRepositoryInterfaceMock{
			UpdateStatusByTokenFunc: func(ctx context.Context, t pgtype.UUID, status string, canceledAt pgtype.Timestamptz, cancelReason pgtype.Text) (*models.Reservation, error) {
				return canceledReservation, nil
			},
		}

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		input := CancelReservationInput{
			WishListID:       wishlistID.String(),
			GiftItemID:       giftItemID.String(),
			UserID:           pgtype.UUID{Valid: false},
			ReservationToken: &token,
		}

		reservation, err := service.CancelReservation(context.Background(), input)

		require.NoError(t, err)
		assert.NotNil(t, reservation)
		assert.Equal(t, "canceled", reservation.Status)
		assert.Len(t, mockRepo.UpdateStatusByTokenCalls(), 1)
	})

	t.Run("cancellation requires user ID or token", func(t *testing.T) {
		giftItemID := pgtype.UUID{Bytes: [16]byte{20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35}, Valid: true}
		wishlistID := pgtype.UUID{Bytes: [16]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}, Valid: true}

		giftItem := &itemmodels.GiftItem{ID: giftItemID}

		mockGiftItemRepo := &GiftItemRepositoryInterfaceMock{
			GetByWishListFunc: func(ctx context.Context, wlID pgtype.UUID) ([]*itemmodels.GiftItem, error) {
				return []*itemmodels.GiftItem{giftItem}, nil
			},
		}
		mockRepo := &ReservationRepositoryInterfaceMock{}

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		input := CancelReservationInput{
			WishListID:       wishlistID.String(),
			GiftItemID:       giftItemID.String(),
			UserID:           pgtype.UUID{Valid: false},
			ReservationToken: nil,
		}

		_, err := service.CancelReservation(context.Background(), input)

		require.Error(t, err)
		assert.True(t, errors.Is(err, ErrMissingUserOrToken))
	})
}
