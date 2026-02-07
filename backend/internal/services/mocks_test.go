package services

import (
	"context"

	db "wish-list/internal/db/models"
	"wish-list/internal/repositories"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/mock"
)

// MockGiftItemRepository is a mock implementation of GiftItemRepositoryInterface
type MockGiftItemRepository struct {
	mock.Mock
}

func (m *MockGiftItemRepository) Create(ctx context.Context, giftItem db.GiftItem) (*db.GiftItem, error) {
	args := m.Called(ctx, giftItem)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.GiftItem), args.Error(1)
}

func (m *MockGiftItemRepository) GetByID(ctx context.Context, id pgtype.UUID) (*db.GiftItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.GiftItem), args.Error(1)
}

func (m *MockGiftItemRepository) GetByWishList(ctx context.Context, wishlistID pgtype.UUID) ([]*db.GiftItem, error) {
	args := m.Called(ctx, wishlistID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.GiftItem), args.Error(1)
}

func (m *MockGiftItemRepository) Update(ctx context.Context, giftItem db.GiftItem) (*db.GiftItem, error) {
	args := m.Called(ctx, giftItem)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.GiftItem), args.Error(1)
}

func (m *MockGiftItemRepository) Delete(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGiftItemRepository) DeleteWithExecutor(ctx context.Context, executor db.Executor, id pgtype.UUID) error {
	args := m.Called(ctx, executor, id)
	return args.Error(0)
}

func (m *MockGiftItemRepository) Reserve(ctx context.Context, giftItemID, userID pgtype.UUID) (*db.GiftItem, error) {
	args := m.Called(ctx, giftItemID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.GiftItem), args.Error(1)
}

func (m *MockGiftItemRepository) Unreserve(ctx context.Context, giftItemID pgtype.UUID) (*db.GiftItem, error) {
	args := m.Called(ctx, giftItemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.GiftItem), args.Error(1)
}

func (m *MockGiftItemRepository) MarkAsPurchased(ctx context.Context, giftItemID, userID pgtype.UUID, purchasedPrice pgtype.Numeric) (*db.GiftItem, error) {
	args := m.Called(ctx, giftItemID, userID, purchasedPrice)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.GiftItem), args.Error(1)
}

func (m *MockGiftItemRepository) GetPublicWishListGiftItems(ctx context.Context, publicSlug string) ([]*db.GiftItem, error) {
	args := m.Called(ctx, publicSlug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.GiftItem), args.Error(1)
}

func (m *MockGiftItemRepository) ReserveIfNotReserved(ctx context.Context, giftItemID, userID pgtype.UUID) (*db.GiftItem, error) {
	args := m.Called(ctx, giftItemID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.GiftItem), args.Error(1)
}

func (m *MockGiftItemRepository) DeleteWithReservationNotification(ctx context.Context, giftItemID pgtype.UUID) ([]*db.Reservation, error) {
	args := m.Called(ctx, giftItemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.Reservation), args.Error(1)
}

func (m *MockGiftItemRepository) GetByOwnerPaginated(ctx context.Context, ownerID pgtype.UUID, filters repositories.ItemFilters) (*repositories.PaginatedResult, error) {
	args := m.Called(ctx, ownerID, filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repositories.PaginatedResult), args.Error(1)
}

func (m *MockGiftItemRepository) SoftDelete(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGiftItemRepository) GetUnattached(ctx context.Context, ownerID pgtype.UUID) ([]*db.GiftItem, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.GiftItem), args.Error(1)
}

func (m *MockGiftItemRepository) CreateWithOwner(ctx context.Context, giftItem db.GiftItem) (*db.GiftItem, error) {
	args := m.Called(ctx, giftItem)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.GiftItem), args.Error(1)
}

func (m *MockGiftItemRepository) UpdateWithNewSchema(ctx context.Context, giftItem *db.GiftItem) (*db.GiftItem, error) {
	args := m.Called(ctx, giftItem)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.GiftItem), args.Error(1)
}

// MockReservationRepository is a mock implementation of ReservationRepositoryInterface
type MockReservationRepository struct {
	mock.Mock
}

func (m *MockReservationRepository) Create(ctx context.Context, reservation db.Reservation) (*db.Reservation, error) {
	args := m.Called(ctx, reservation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Reservation), args.Error(1)
}

func (m *MockReservationRepository) GetByID(ctx context.Context, id pgtype.UUID) (*db.Reservation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Reservation), args.Error(1)
}

func (m *MockReservationRepository) GetByToken(ctx context.Context, token pgtype.UUID) (*db.Reservation, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Reservation), args.Error(1)
}

func (m *MockReservationRepository) GetByGiftItem(ctx context.Context, giftItemID pgtype.UUID) ([]*db.Reservation, error) {
	args := m.Called(ctx, giftItemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.Reservation), args.Error(1)
}

func (m *MockReservationRepository) GetActiveReservationForGiftItem(ctx context.Context, giftItemID pgtype.UUID) (*db.Reservation, error) {
	args := m.Called(ctx, giftItemID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Reservation), args.Error(1)
}

func (m *MockReservationRepository) GetReservationsByUser(ctx context.Context, userID pgtype.UUID, limit, offset int) ([]*db.Reservation, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.Reservation), args.Error(1)
}

func (m *MockReservationRepository) UpdateStatus(ctx context.Context, reservationID pgtype.UUID, status string, canceledAt pgtype.Timestamptz, cancelReason pgtype.Text) (*db.Reservation, error) {
	args := m.Called(ctx, reservationID, status, canceledAt, cancelReason)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Reservation), args.Error(1)
}

func (m *MockReservationRepository) UpdateStatusByToken(ctx context.Context, token pgtype.UUID, status string, canceledAt pgtype.Timestamptz, cancelReason pgtype.Text) (*db.Reservation, error) {
	args := m.Called(ctx, token, status, canceledAt, cancelReason)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.Reservation), args.Error(1)
}

func (m *MockReservationRepository) ListUserReservationsWithDetails(ctx context.Context, userID pgtype.UUID, limit, offset int) ([]repositories.ReservationDetail, error) {
	args := m.Called(ctx, userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.ReservationDetail), args.Error(1)
}

func (m *MockReservationRepository) ListGuestReservationsWithDetails(ctx context.Context, token pgtype.UUID) ([]repositories.ReservationDetail, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]repositories.ReservationDetail), args.Error(1)
}

func (m *MockReservationRepository) CountUserReservations(ctx context.Context, userID pgtype.UUID) (int, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return 0, args.Error(1)
	}
	return args.Int(0), args.Error(1)
}
