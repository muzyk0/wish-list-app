package services

import (
	"context"
	"math/big"
	"testing"

	db "wish-list/internal/db/models"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWishListService_CreateGiftItem(t *testing.T) {
	tests := []struct {
		name          string
		wishlistID    string
		input         CreateGiftItemInput
		mockReturn    *db.GiftItem
		mockError     error
		expectedError bool
	}{
		{
			name:       "successful creation",
			wishlistID: "12345678-1234-5678-9abc-def012345678",
			input: CreateGiftItemInput{
				Name:        "Test Gift",
				Description: "Test Description",
				Link:        "https://example.com/test-gift",
				ImageURL:    "https://example.com/test-gift.jpg",
				Price:       29.99,
				Priority:    8,
				Notes:       "Test notes",
				Position:    1,
			},
			mockReturn: &db.GiftItem{
				ID:          pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true},
				OwnerID:     pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 17}, Valid: true},
				Name:        "Test Gift",
				Description: pgtype.Text{String: "Test Description", Valid: true},
				Link:        pgtype.Text{String: "https://example.com/test-gift", Valid: true},
				ImageUrl:    pgtype.Text{String: "https://example.com/test-gift.jpg", Valid: true},
				Price:       pgtype.Numeric{Int: big.NewInt(2999), Exp: -2, Valid: true},
				Priority:    pgtype.Int4{Int32: 8, Valid: true},
				Notes:       pgtype.Text{String: "Test notes", Valid: true},
				Position:    pgtype.Int4{Int32: 1, Valid: true},
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:       "empty name error",
			wishlistID: "12345678-1234-5678-9abc-def012345678",
			input: CreateGiftItemInput{
				Name:        "",
				Description: "Test Description",
				Link:        "https://example.com/test-gift",
				ImageURL:    "https://example.com/test-gift.jpg",
				Price:       29.99,
				Priority:    8,
				Notes:       "Test notes",
				Position:    1,
			},
			mockReturn:    nil,
			mockError:     nil,
			expectedError: true,
		},
		{
			name:          "invalid wishlist id error",
			wishlistID:    "invalid-uuid",
			input:         CreateGiftItemInput{Name: "Test Gift"},
			mockReturn:    nil,
			mockError:     nil,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWishListRepo := &WishListRepositoryInterfaceMock{}
			mockGiftItemRepo := &GiftItemRepositoryInterfaceMock{}

			if tt.mockReturn != nil || tt.mockError != nil {
				mockGiftItemRepo.CreateFunc = func(ctx context.Context, gi db.GiftItem) (*db.GiftItem, error) {
					return tt.mockReturn, tt.mockError
				}
			}

			service := NewWishListService(mockWishListRepo, mockGiftItemRepo, nil, nil, nil, nil)

			result, err := service.CreateGiftItem(context.Background(), tt.wishlistID, tt.input)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.mockReturn.Name, result.Name)
				assert.Equal(t, tt.mockReturn.Description.String, result.Description)
				assert.Equal(t, tt.mockReturn.Link.String, result.Link)
				assert.Equal(t, tt.mockReturn.ImageUrl.String, result.ImageURL)

				expectedPrice, err := tt.mockReturn.Price.Float64Value()
				require.NoError(t, err)
				assert.True(t, expectedPrice.Valid)
				assert.InDelta(t, expectedPrice.Float64, result.Price, 0.001)

				assert.Equal(t, int(tt.mockReturn.Priority.Int32), result.Priority)
			}
		})
	}
}

func TestWishListService_GetGiftItem(t *testing.T) {
	testUUID := pgtype.UUID{Bytes: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16}, Valid: true}

	tests := []struct {
		name          string
		giftItemID    string
		mockReturn    *db.GiftItem
		mockError     error
		expectedError bool
	}{
		{
			name:       "successful retrieval",
			giftItemID: "12345678-1234-5678-9abc-def012345678",
			mockReturn: &db.GiftItem{
				ID:          testUUID,
				OwnerID:     testUUID,
				Name:        "Test Gift",
				Description: pgtype.Text{String: "Test Description", Valid: true},
				Link:        pgtype.Text{String: "https://example.com/test-gift", Valid: true},
				ImageUrl:    pgtype.Text{String: "https://example.com/test-gift.jpg", Valid: true},
				Price:       pgtype.Numeric{Int: big.NewInt(2999), Exp: -2, Valid: true},
				Priority:    pgtype.Int4{Int32: 8, Valid: true},
				Notes:       pgtype.Text{String: "Test notes", Valid: true},
				Position:    pgtype.Int4{Int32: 1, Valid: true},
			},
			mockError:     nil,
			expectedError: false,
		},
		{
			name:          "invalid gift item id",
			giftItemID:    "invalid-uuid",
			mockReturn:    nil,
			mockError:     nil,
			expectedError: true,
		},
		{
			name:          "gift item not found",
			giftItemID:    "12345678-1234-5678-9abc-def012345678",
			mockReturn:    nil,
			mockError:     assert.AnError,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockWishListRepo := &WishListRepositoryInterfaceMock{}
			mockGiftItemRepo := &GiftItemRepositoryInterfaceMock{}

			if tt.mockReturn != nil || tt.mockError != nil {
				mockGiftItemRepo.GetByIDFunc = func(ctx context.Context, id pgtype.UUID) (*db.GiftItem, error) {
					return tt.mockReturn, tt.mockError
				}
			}

			service := NewWishListService(mockWishListRepo, mockGiftItemRepo, nil, nil, nil, nil)

			result, err := service.GetGiftItem(context.Background(), tt.giftItemID)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.mockReturn.Name, result.Name)
				assert.Equal(t, tt.mockReturn.Description.String, result.Description)
				assert.Equal(t, tt.mockReturn.Link.String, result.Link)
				assert.Equal(t, tt.mockReturn.ImageUrl.String, result.ImageURL)

				expectedPrice, err := tt.mockReturn.Price.Float64Value()
				require.NoError(t, err)
				assert.True(t, expectedPrice.Valid)
				assert.InDelta(t, expectedPrice.Float64, result.Price, 0.001)
			}
		})
	}
}
