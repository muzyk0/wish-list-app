package services

import (
	"context"
	"testing"

	db "wish-list/internal/db/models"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockWishListRepository is a mock implementation of WishListRepositoryInterface
type MockWishListRepository struct {
	mock.Mock
}

func (m *MockWishListRepository) Create(ctx context.Context, wishList db.WishList) (*db.WishList, error) {
	args := m.Called(ctx, wishList)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.WishList), args.Error(1)
}

func (m *MockWishListRepository) GetByID(ctx context.Context, id pgtype.UUID) (*db.WishList, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.WishList), args.Error(1)
}

func (m *MockWishListRepository) GetByOwner(ctx context.Context, ownerID pgtype.UUID) ([]*db.WishList, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.WishList), args.Error(1)
}

func (m *MockWishListRepository) GetByPublicSlug(ctx context.Context, publicSlug string) (*db.WishList, error) {
	args := m.Called(ctx, publicSlug)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.WishList), args.Error(1)
}

func (m *MockWishListRepository) Update(ctx context.Context, wishList db.WishList) (*db.WishList, error) {
	args := m.Called(ctx, wishList)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*db.WishList), args.Error(1)
}

func (m *MockWishListRepository) Delete(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWishListRepository) DeleteWithExecutor(ctx context.Context, executor db.Executor, id pgtype.UUID) error {
	args := m.Called(ctx, executor, id)
	return args.Error(0)
}

func (m *MockWishListRepository) IncrementViewCount(ctx context.Context, id pgtype.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockWishListRepository) GetByOwnerWithItemCount(ctx context.Context, ownerID pgtype.UUID) ([]*db.WishListWithItemCount, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*db.WishListWithItemCount), args.Error(1)
}

func TestWishListService_CreateWishList(t *testing.T) {
	tests := []struct {
		name          string
		input         CreateWishListInput
		userID        string
		mockReturn    *db.WishList
		mockError     error
		expectedError bool
	}{
		{
			name: "successful creation",
			input: CreateWishListInput{
				Title:        "Test List",
				Description:  "Test Description",
				Occasion:     "Birthday",
				OccasionDate: "2026-12-25",
				TemplateID:   "default",
				IsPublic:     true,
			},
			userID: "01020304-0506-0708-090a-0b0c0d0e0f10",
			mockReturn: &db.WishList{
				ID:          pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
				OwnerID:     pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
				Title:       "Test List",
				Description: pgtype.Text{String: "Test Description", Valid: true},
				Occasion:    pgtype.Text{String: "Birthday", Valid: true},
				TemplateID:  "default",
				IsPublic:    pgtype.Bool{Bool: true, Valid: true},
				PublicSlug:  pgtype.Text{String: "test-list-1234", Valid: true},
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name: "empty title error",
			input: CreateWishListInput{
				Title:        "",
				Description:  "Test Description",
				Occasion:     "Birthday",
				OccasionDate: "2026-12-25",
				TemplateID:   "default",
				IsPublic:     true,
			},
			userID:        "test-user-id",
			mockReturn:    nil,
			mockError:     nil,
			expectedError: true,
		},
		{
			name: "invalid user id error",
			input: CreateWishListInput{
				Title:        "Test List",
				Description:  "Test Description",
				Occasion:     "Birthday",
				OccasionDate: "2026-12-25",
				TemplateID:   "default",
				IsPublic:     true,
			},
			userID:        "invalid-user-id",
			mockReturn:    nil,
			mockError:     nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWishListRepo := new(MockWishListRepository)
			mockGiftItemRepo := new(MockGiftItemRepository)

			if tt.mockReturn != nil || tt.mockError != nil {
				mockWishListRepo.On("Create", mock.Anything, mock.AnythingOfType("db.WishList")).Return(tt.mockReturn, tt.mockError)
			}

			service := NewWishListService(mockWishListRepo, mockGiftItemRepo, nil, nil, nil, nil)

			result, err := service.CreateWishList(context.Background(), tt.userID, tt.input)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.mockReturn.Title, result.Title)
				assert.Equal(t, tt.mockReturn.Description.String, result.Description)
				assert.Equal(t, tt.mockReturn.Occasion.String, result.Occasion)
				assert.Equal(t, tt.mockReturn.TemplateID, result.TemplateID)
				assert.Equal(t, tt.mockReturn.IsPublic.Bool, result.IsPublic)
			}

			mockWishListRepo.AssertExpectations(t)
		})
	}
}

func TestWishListService_GetWishList(t *testing.T) {
	// Create a valid UUID for testing
	testUUID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}

	tests := []struct {
		name          string
		wishListID    string
		mockReturn    *db.WishList
		mockError     error
		expectedError bool
	}{
		{
			name:       "successful retrieval",
			wishListID: "12345678-1234-5678-9abc-def012345678",
			mockReturn: &db.WishList{
				ID:          testUUID,
				OwnerID:     testUUID,
				Title:       "Test List",
				Description: pgtype.Text{String: "Test Description", Valid: true},
				Occasion:    pgtype.Text{String: "Birthday", Valid: true},
				TemplateID:  "default",
				IsPublic:    pgtype.Bool{Bool: true, Valid: true},
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "invalid wishlist id",
			wishListID:    "invalid-uuid",
			mockReturn:    nil,
			mockError:     nil,
			expectedError: true,
		},
		{
			name:          "wishlist not found",
			wishListID:    "12345678-1234-5678-9abc-def012345678",
			mockReturn:    nil,
			mockError:     assert.AnError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWishListRepo := new(MockWishListRepository)
			mockGiftItemRepo := new(MockGiftItemRepository)

			if tt.mockReturn != nil || tt.mockError != nil {
				mockWishListRepo.On("GetByID", mock.Anything, mock.AnythingOfType("pgtype.UUID")).Return(tt.mockReturn, tt.mockError)
			}

			service := NewWishListService(mockWishListRepo, mockGiftItemRepo, nil, nil, nil, nil)

			result, err := service.GetWishList(context.Background(), tt.wishListID)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.mockReturn.Title, result.Title)
				assert.Equal(t, tt.mockReturn.Description.String, result.Description)
				assert.Equal(t, tt.mockReturn.Occasion.String, result.Occasion)
			}

			mockWishListRepo.AssertExpectations(t)
		})
	}
}
