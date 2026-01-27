package services

import (
	"context"
	"testing"
	"time"

	db "wish-list/internal/db/models"
	"wish-list/internal/repositories"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestReservationService_GetReservationStatus(t *testing.T) {
	t.Run("available gift item", func(t *testing.T) {
		mockRepo := new(MockReservationRepository)
		mockGiftItemRepo := new(MockGiftItemRepository)

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}

		// Mock ownership validation
		mockGiftItemRepo.On("GetPublicWishListGiftItems", mock.Anything, "public-slug").Return([]*db.GiftItem{
			{ID: giftItemID},
		}, nil)

		mockRepo.On("GetActiveReservationForGiftItem", mock.Anything, giftItemID).Return(nil, nil)

		status, err := service.GetReservationStatus(context.Background(), "public-slug", giftItemID.String())

		assert.NoError(t, err)
		assert.False(t, status.IsReserved)
		assert.Equal(t, "available", status.Status)

		mockRepo.AssertExpectations(t)
		mockGiftItemRepo.AssertExpectations(t)
	})

	t.Run("reserved gift item", func(t *testing.T) {
		mockRepo := new(MockReservationRepository)
		mockGiftItemRepo := new(MockGiftItemRepository)

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		token := pgtype.UUID{Bytes: [16]byte{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}, Valid: true}

		activeReservation := &db.Reservation{
			ID:               pgtype.UUID{Bytes: [16]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}, Valid: true},
			GiftItemID:       giftItemID,
			ReservationToken: token,
			Status:           "active",
		}

		// Mock ownership validation
		mockGiftItemRepo.On("GetPublicWishListGiftItems", mock.Anything, "public-slug").Return([]*db.GiftItem{
			{ID: giftItemID},
		}, nil)

		mockRepo.On("GetActiveReservationForGiftItem", mock.Anything, giftItemID).Return(activeReservation, nil)
		mockRepo.On("ListGuestReservationsWithDetails", mock.Anything, token).Return([]repositories.ReservationDetail{}, nil)

		status, err := service.GetReservationStatus(context.Background(), "public-slug", giftItemID.String())

		assert.NoError(t, err)
		assert.True(t, status.IsReserved)
		assert.Equal(t, "active", status.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("invalid gift item id", func(t *testing.T) {
		mockRepo := new(MockReservationRepository)
		mockGiftItemRepo := new(MockGiftItemRepository)

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		status, err := service.GetReservationStatus(context.Background(), "public-slug", "invalid-uuid")

		assert.Error(t, err)
		assert.Nil(t, status)

		mockRepo.AssertNotCalled(t, "GetActiveReservationForGiftItem", mock.Anything, mock.Anything)
	})
}

// T070a: Unit tests for reservation expiration logic
func TestReservationService_ExpirationLogic(t *testing.T) {
	t.Run("expired reservation returns available status", func(t *testing.T) {
		mockRepo := new(MockReservationRepository)
		mockGiftItemRepo := new(MockGiftItemRepository)

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		token := pgtype.UUID{Bytes: [16]byte{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}, Valid: true}
		reservationID := pgtype.UUID{Bytes: [16]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}, Valid: true}

		// Create expired reservation (expires in the past)
		expiredReservation := &db.Reservation{
			ID:               reservationID,
			GiftItemID:       giftItemID,
			ReservationToken: token,
			Status:           "active",
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
				Valid: true,
			},
		}

		// Mock ownership validation
		mockGiftItemRepo.On("GetPublicWishListGiftItems", mock.Anything, "public-slug").Return([]*db.GiftItem{
			{ID: giftItemID},
		}, nil)

		mockRepo.On("GetActiveReservationForGiftItem", mock.Anything, giftItemID).Return(expiredReservation, nil)
		mockRepo.On("UpdateStatus", mock.Anything, reservationID, "expired", mock.Anything, mock.Anything).
			Return(&db.Reservation{Status: "expired"}, nil)

		status, err := service.GetReservationStatus(context.Background(), "public-slug", giftItemID.String())

		assert.NoError(t, err)
		assert.False(t, status.IsReserved)
		assert.Equal(t, "available", status.Status)

		mockRepo.AssertExpectations(t)
		mockGiftItemRepo.AssertExpectations(t)
	})

	t.Run("non-expired reservation remains active", func(t *testing.T) {
		mockRepo := new(MockReservationRepository)
		mockGiftItemRepo := new(MockGiftItemRepository)

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		token := pgtype.UUID{Bytes: [16]byte{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}, Valid: true}

		// Create non-expired reservation (expires in the future)
		activeReservation := &db.Reservation{
			ID:               pgtype.UUID{Bytes: [16]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}, Valid: true},
			GiftItemID:       giftItemID,
			ReservationToken: token,
			Status:           "active",
			ExpiresAt: pgtype.Timestamptz{
				Time:  time.Now().Add(24 * time.Hour), // Expires in 24 hours
				Valid: true,
			},
		}

		// Mock ownership validation
		mockGiftItemRepo.On("GetPublicWishListGiftItems", mock.Anything, "public-slug").Return([]*db.GiftItem{
			{ID: giftItemID},
		}, nil)

		mockRepo.On("GetActiveReservationForGiftItem", mock.Anything, giftItemID).Return(activeReservation, nil)
		mockRepo.On("ListGuestReservationsWithDetails", mock.Anything, token).Return([]repositories.ReservationDetail{}, nil)

		status, err := service.GetReservationStatus(context.Background(), "public-slug", giftItemID.String())

		assert.NoError(t, err)
		assert.True(t, status.IsReserved)
		assert.Equal(t, "active", status.Status)

		mockRepo.AssertExpectations(t)
		mockGiftItemRepo.AssertExpectations(t)
	})

	t.Run("reservation without expiry date stays active", func(t *testing.T) {
		mockRepo := new(MockReservationRepository)
		mockGiftItemRepo := new(MockGiftItemRepository)

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		token := pgtype.UUID{Bytes: [16]byte{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}, Valid: true}

		// Mock ownership validation
		mockGiftItemRepo.On("GetPublicWishListGiftItems", mock.Anything, "public-slug").Return([]*db.GiftItem{
			{ID: giftItemID},
		}, nil)

		// Create reservation without expiry date
		activeReservation := &db.Reservation{
			ID:               pgtype.UUID{Bytes: [16]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}, Valid: true},
			GiftItemID:       giftItemID,
			ReservationToken: token,
			Status:           "active",
			ExpiresAt:        pgtype.Timestamptz{Valid: false}, // No expiry
		}

		mockRepo.On("GetActiveReservationForGiftItem", mock.Anything, giftItemID).Return(activeReservation, nil)
		mockRepo.On("ListGuestReservationsWithDetails", mock.Anything, token).Return([]repositories.ReservationDetail{}, nil)

		status, err := service.GetReservationStatus(context.Background(), "public-slug", giftItemID.String())

		assert.NoError(t, err)
		assert.True(t, status.IsReserved)
		assert.Equal(t, "active", status.Status)

		mockRepo.AssertExpectations(t)
		mockGiftItemRepo.AssertExpectations(t)
	})
}

// T070b: Unit tests for concurrency controls for simultaneous reservations
func TestReservationService_ConcurrencyControls(t *testing.T) {
	t.Run("create reservation on already reserved item fails", func(t *testing.T) {
		mockRepo := new(MockReservationRepository)
		mockGiftItemRepo := new(MockGiftItemRepository)

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		wishlistID := pgtype.UUID{Bytes: [16]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}, Valid: true}

		// Create gift item
		giftItem := &db.GiftItem{
			ID: giftItemID,
		}

		// Active reservation exists
		existingReservation := &db.Reservation{
			ID:         pgtype.UUID{Valid: true},
			GiftItemID: giftItemID,
			Status:     "active",
		}

		// Mock ownership validation
		mockGiftItemRepo.On("GetByWishList", mock.Anything, wishlistID).Return([]*db.GiftItem{giftItem}, nil)
		mockRepo.On("GetActiveReservationForGiftItem", mock.Anything, giftItemID).Return(existingReservation, nil)

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

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already reserved")

		mockRepo.AssertExpectations(t)
		mockGiftItemRepo.AssertExpectations(t)
	})

	t.Run("authenticated user reservation uses atomic operation", func(t *testing.T) {
		mockRepo := new(MockReservationRepository)
		mockGiftItemRepo := new(MockGiftItemRepository)

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		userID := pgtype.UUID{Bytes: [16]byte{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17}, Valid: true}
		wishlistID := pgtype.UUID{Bytes: [16]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}, Valid: true}

		// Create gift item
		giftItem := &db.GiftItem{
			ID: giftItemID,
		}

		reservedItem := &db.GiftItem{
			ID:               giftItemID,
			ReservedByUserID: userID,
		}

		createdReservation := &db.Reservation{
			ID:               pgtype.UUID{Bytes: [16]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}, Valid: true},
			GiftItemID:       giftItemID,
			ReservedByUserID: userID,
			Status:           "active",
		}

		// Mock ownership validation
		mockGiftItemRepo.On("GetByWishList", mock.Anything, wishlistID).Return([]*db.GiftItem{giftItem}, nil)
		mockGiftItemRepo.On("ReserveIfNotReserved", mock.Anything, giftItemID, userID).Return(reservedItem, nil)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("db.Reservation")).Return(createdReservation, nil)

		input := CreateReservationInput{
			WishListID: wishlistID.String(),
			GiftItemID: giftItemID.String(),
			UserID:     userID,
		}

		reservation, err := service.CreateReservation(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, reservation)
		assert.Equal(t, "active", reservation.Status)

		mockRepo.AssertExpectations(t)
		mockGiftItemRepo.AssertExpectations(t)
	})

	t.Run("concurrent reservation attempts are handled atomically", func(t *testing.T) {
		// This test validates that the repository uses proper locking
		// In a real integration test, we would spawn multiple goroutines
		// For unit tests, we verify the atomic operation is called
		t.Skip("Full concurrency testing requires integration tests with real database")
	})
}

// Test CreateReservation function
func TestReservationService_CreateReservation(t *testing.T) {
	t.Run("successful guest reservation", func(t *testing.T) {
		mockRepo := new(MockReservationRepository)
		mockGiftItemRepo := new(MockGiftItemRepository)

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		wishlistID := pgtype.UUID{Bytes: [16]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}, Valid: true}

		giftItem := &db.GiftItem{
			ID: giftItemID,
		}

		createdReservation := &db.Reservation{
			ID:         pgtype.UUID{Bytes: [16]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}, Valid: true},
			GiftItemID: giftItemID,
			Status:     "active",
		}

		// Mock ownership validation
		mockGiftItemRepo.On("GetByWishList", mock.Anything, wishlistID).Return([]*db.GiftItem{giftItem}, nil)
		mockRepo.On("GetActiveReservationForGiftItem", mock.Anything, giftItemID).Return(nil, nil)
		mockRepo.On("Create", mock.Anything, mock.AnythingOfType("db.Reservation")).Return(createdReservation, nil)

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

		assert.NoError(t, err)
		assert.NotNil(t, reservation)
		assert.Equal(t, "active", reservation.Status)

		mockRepo.AssertExpectations(t)
		mockGiftItemRepo.AssertExpectations(t)
	})

	t.Run("guest reservation requires name and email", func(t *testing.T) {
		mockRepo := new(MockReservationRepository)
		mockGiftItemRepo := new(MockGiftItemRepository)

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		giftItemID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		wishlistID := pgtype.UUID{Bytes: [16]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}, Valid: true}

		giftItem := &db.GiftItem{
			ID: giftItemID,
		}

		// Mock ownership validation
		mockGiftItemRepo.On("GetByWishList", mock.Anything, wishlistID).Return([]*db.GiftItem{giftItem}, nil)
		mockRepo.On("GetActiveReservationForGiftItem", mock.Anything, giftItemID).Return(nil, nil)

		// Missing guest details
		input := CreateReservationInput{
			WishListID: wishlistID.String(),
			GiftItemID: giftItemID.String(),
			UserID:     pgtype.UUID{Valid: false},
			GuestName:  nil,
			GuestEmail: nil,
		}

		_, err := service.CreateReservation(context.Background(), input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "guest name and email are required")
	})

	t.Run("invalid gift item id", func(t *testing.T) {
		mockRepo := new(MockReservationRepository)
		mockGiftItemRepo := new(MockGiftItemRepository)

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		input := CreateReservationInput{
			WishListID: "list-123",
			GiftItemID: "invalid-uuid",
			UserID:     pgtype.UUID{Valid: false},
		}

		_, err := service.CreateReservation(context.Background(), input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid gift item id")
	})
}

// Test CancelReservation function
func TestReservationService_CancelReservation(t *testing.T) {
	t.Run("successful cancellation by guest with token", func(t *testing.T) {
		mockRepo := new(MockReservationRepository)
		mockGiftItemRepo := new(MockGiftItemRepository)

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		token := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}
		giftItemID := pgtype.UUID{Bytes: [16]byte{20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35}, Valid: true}
		wishlistID := pgtype.UUID{Bytes: [16]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}, Valid: true}

		giftItem := &db.GiftItem{
			ID: giftItemID,
		}

		canceledReservation := &db.Reservation{
			ID:               pgtype.UUID{Bytes: [16]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18}, Valid: true},
			ReservationToken: token,
			Status:           "cancelled",
		}

		// Mock ownership validation
		mockGiftItemRepo.On("GetByWishList", mock.Anything, wishlistID).Return([]*db.GiftItem{giftItem}, nil)
		mockRepo.On("UpdateStatusByToken", mock.Anything, token, "cancelled", mock.Anything, mock.Anything).
			Return(canceledReservation, nil)

		input := CancelReservationInput{
			WishListID:       wishlistID.String(),
			GiftItemID:       giftItemID.String(),
			UserID:           pgtype.UUID{Valid: false},
			ReservationToken: &token,
		}

		reservation, err := service.CancelReservation(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, reservation)
		assert.Equal(t, "cancelled", reservation.Status)

		mockRepo.AssertExpectations(t)
	})

	t.Run("cancellation requires user ID or token", func(t *testing.T) {
		mockRepo := new(MockReservationRepository)
		mockGiftItemRepo := new(MockGiftItemRepository)

		service := NewReservationService(mockRepo, mockGiftItemRepo)

		giftItemID := pgtype.UUID{Bytes: [16]byte{20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35}, Valid: true}
		wishlistID := pgtype.UUID{Bytes: [16]byte{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25}, Valid: true}

		giftItem := &db.GiftItem{
			ID: giftItemID,
		}

		// Mock ownership validation
		mockGiftItemRepo.On("GetByWishList", mock.Anything, wishlistID).Return([]*db.GiftItem{giftItem}, nil)

		input := CancelReservationInput{
			WishListID:       wishlistID.String(),
			GiftItemID:       giftItemID.String(),
			UserID:           pgtype.UUID{Valid: false},
			ReservationToken: nil,
		}

		_, err := service.CancelReservation(context.Background(), input)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "either user ID or reservation token must be provided")

		mockGiftItemRepo.AssertExpectations(t)
	})
}
