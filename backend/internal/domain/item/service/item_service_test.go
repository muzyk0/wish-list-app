package service

import (
	"context"
	"errors"
	"math/big"
	"testing"
	"time"

	"wish-list/internal/domain/item/models"
	"wish-list/internal/domain/item/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func newValidPgtypeUUID(t *testing.T) (pgtype.UUID, string) {
	t.Helper()
	raw := uuid.New()
	pg := pgtype.UUID{}
	err := pg.Scan(raw.String())
	require.NoError(t, err)
	return pg, raw.String()
}

func pgtypeUUIDFromString(t *testing.T, s string) pgtype.UUID {
	t.Helper()
	pg := pgtype.UUID{}
	err := pg.Scan(s)
	require.NoError(t, err)
	return pg
}

func makeGiftItem(ownerID pgtype.UUID) *models.GiftItem {
	itemID := pgtype.UUID{}
	_ = itemID.Scan(uuid.New().String())
	now := time.Now().UTC()
	return &models.GiftItem{
		ID:          itemID,
		OwnerID:     ownerID,
		Name:        "Test Item",
		Description: pgtype.Text{String: "A description", Valid: true},
		Link:        pgtype.Text{String: "https://example.com", Valid: true},
		ImageUrl:    pgtype.Text{String: "https://example.com/img.jpg", Valid: true},
		Price:       pgtype.Numeric{Int: big.NewInt(1999), Exp: -2, Valid: true},
		Priority:    pgtype.Int4{Int32: 5, Valid: true},
		Notes:       pgtype.Text{String: "Some notes", Valid: true},
		Position:    pgtype.Int4{Int32: 1, Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
	}
}

func newItemService(
	itemRepo *GiftItemRepositoryInterfaceMock,
	wishlistItemRepo *WishlistItemRepositoryInterfaceMock,
) *ItemService {
	return NewItemService(itemRepo, wishlistItemRepo)
}

func stringPtr(s string) *string    { return &s }
func float64Ptr(f float64) *float64 { return &f }
func intPtr(i int) *int             { return &i }

// ---------------------------------------------------------------------------
// GetMyItems
// ---------------------------------------------------------------------------

func TestItemService_GetMyItems_Success(t *testing.T) {
	ownerID, ownerStr := newValidPgtypeUUID(t)
	item := makeGiftItem(ownerID)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByOwnerPaginatedFunc: func(ctx context.Context, oid pgtype.UUID, filters repository.ItemFilters) (*repository.PaginatedResult, error) {
			assert.Equal(t, ownerID, oid)
			assert.Equal(t, 10, filters.Limit)
			assert.Equal(t, 1, filters.Page)
			assert.Equal(t, "created_at", filters.Sort)
			assert.Equal(t, "desc", filters.Order)
			return &repository.PaginatedResult{
				Items:      []*models.GiftItem{item},
				TotalCount: 1,
			}, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.GetMyItems(context.Background(), ownerStr, repository.ItemFilters{})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, int64(1), result.TotalCount)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, item.Name, result.Items[0].Name)
	assert.Equal(t, 10, result.Limit)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 1, result.TotalPages)
	assert.Len(t, itemRepo.GetByOwnerPaginatedCalls(), 1)
}

func TestItemService_GetMyItems_DefaultFilters(t *testing.T) {
	_, ownerStr := newValidPgtypeUUID(t)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByOwnerPaginatedFunc: func(ctx context.Context, oid pgtype.UUID, filters repository.ItemFilters) (*repository.PaginatedResult, error) {
			// Verify defaults were applied
			assert.Equal(t, 10, filters.Limit)
			assert.Equal(t, 1, filters.Page)
			assert.Equal(t, "created_at", filters.Sort)
			assert.Equal(t, "desc", filters.Order)
			return &repository.PaginatedResult{Items: []*models.GiftItem{}, TotalCount: 0}, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.GetMyItems(context.Background(), ownerStr, repository.ItemFilters{})

	require.NoError(t, err)
	assert.Equal(t, int64(0), result.TotalCount)
	assert.Empty(t, result.Items)
}

func TestItemService_GetMyItems_LimitCap(t *testing.T) {
	_, ownerStr := newValidPgtypeUUID(t)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByOwnerPaginatedFunc: func(ctx context.Context, oid pgtype.UUID, filters repository.ItemFilters) (*repository.PaginatedResult, error) {
			assert.Equal(t, 100, filters.Limit, "limit exceeding 100 should be capped to 100")
			return &repository.PaginatedResult{Items: []*models.GiftItem{}, TotalCount: 0}, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	_, err := svc.GetMyItems(context.Background(), ownerStr, repository.ItemFilters{Limit: 200})

	require.NoError(t, err)
}

func TestItemService_GetMyItems_CustomFilters(t *testing.T) {
	_, ownerStr := newValidPgtypeUUID(t)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByOwnerPaginatedFunc: func(ctx context.Context, oid pgtype.UUID, filters repository.ItemFilters) (*repository.PaginatedResult, error) {
			assert.Equal(t, 25, filters.Limit)
			assert.Equal(t, 3, filters.Page)
			assert.Equal(t, "price", filters.Sort)
			assert.Equal(t, "asc", filters.Order)
			return &repository.PaginatedResult{Items: []*models.GiftItem{}, TotalCount: 0}, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	_, err := svc.GetMyItems(context.Background(), ownerStr, repository.ItemFilters{
		Limit: 25,
		Page:  3,
		Sort:  "price",
		Order: "asc",
	})

	require.NoError(t, err)
}

func TestItemService_GetMyItems_InvalidUserID(t *testing.T) {
	itemRepo := &GiftItemRepositoryInterfaceMock{}
	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})

	result, err := svc.GetMyItems(context.Background(), "not-a-uuid", repository.ItemFilters{})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrInvalidItemUser))
	assert.Empty(t, itemRepo.GetByOwnerPaginatedCalls(), "repo should not be called for invalid UUID")
}

func TestItemService_GetMyItems_RepoError(t *testing.T) {
	_, ownerStr := newValidPgtypeUUID(t)
	repoErr := errors.New("database connection lost")

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByOwnerPaginatedFunc: func(ctx context.Context, oid pgtype.UUID, filters repository.ItemFilters) (*repository.PaginatedResult, error) {
			return nil, repoErr
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.GetMyItems(context.Background(), ownerStr, repository.ItemFilters{})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get items")
}

func TestItemService_GetMyItems_TotalPages(t *testing.T) {
	ownerID, ownerStr := newValidPgtypeUUID(t)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByOwnerPaginatedFunc: func(ctx context.Context, oid pgtype.UUID, filters repository.ItemFilters) (*repository.PaginatedResult, error) {
			return &repository.PaginatedResult{
				Items:      []*models.GiftItem{makeGiftItem(ownerID)},
				TotalCount: 25,
			}, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.GetMyItems(context.Background(), ownerStr, repository.ItemFilters{})

	require.NoError(t, err)
	// 25 items / 10 per page = 3 pages (ceiling)
	assert.Equal(t, 3, result.TotalPages)
}

// ---------------------------------------------------------------------------
// CreateItem
// ---------------------------------------------------------------------------

func TestItemService_CreateItem_Success(t *testing.T) {
	ownerID, ownerStr := newValidPgtypeUUID(t)
	returnedItem := makeGiftItem(ownerID)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		CreateFunc: func(ctx context.Context, gi models.GiftItem) (*models.GiftItem, error) {
			assert.Equal(t, ownerID, gi.OwnerID)
			assert.Equal(t, "Birthday Gift", gi.Name)
			assert.Equal(t, "A nice present", gi.Description.String)
			assert.True(t, gi.Description.Valid)
			assert.Equal(t, "https://shop.com", gi.Link.String)
			assert.True(t, gi.Link.Valid)
			assert.True(t, gi.Price.Valid)
			assert.Equal(t, int32(3), gi.Priority.Int32)
			return returnedItem, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.CreateItem(context.Background(), ownerStr, CreateItemInput{
		Title:       "Birthday Gift",
		Description: "A nice present",
		Link:        "https://shop.com",
		ImageURL:    "https://shop.com/img.jpg",
		Price:       49.99,
		Priority:    3,
		Notes:       "Wrap it nicely",
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, returnedItem.Name, result.Name)
	assert.Len(t, itemRepo.CreateCalls(), 1)
}

func TestItemService_CreateItem_EmptyTitle(t *testing.T) {
	itemRepo := &GiftItemRepositoryInterfaceMock{}
	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})

	result, err := svc.CreateItem(context.Background(), uuid.New().String(), CreateItemInput{
		Title: "",
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrItemTitleRequired))
	assert.Empty(t, itemRepo.CreateCalls(), "repo should not be called when title is empty")
}

func TestItemService_CreateItem_InvalidUserID(t *testing.T) {
	itemRepo := &GiftItemRepositoryInterfaceMock{}
	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})

	result, err := svc.CreateItem(context.Background(), "bad-uuid", CreateItemInput{
		Title: "Valid Title",
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrInvalidItemUser))
	assert.Empty(t, itemRepo.CreateCalls())
}

func TestItemService_CreateItem_ZeroPrice(t *testing.T) {
	ownerID, ownerStr := newValidPgtypeUUID(t)
	returnedItem := makeGiftItem(ownerID)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		CreateFunc: func(ctx context.Context, gi models.GiftItem) (*models.GiftItem, error) {
			// Price should not be set when zero
			assert.False(t, gi.Price.Valid, "price should not be set when input is 0")
			return returnedItem, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.CreateItem(context.Background(), ownerStr, CreateItemInput{
		Title: "Free Gift",
		Price: 0,
	})

	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestItemService_CreateItem_OptionalFieldsEmpty(t *testing.T) {
	ownerID, ownerStr := newValidPgtypeUUID(t)
	returnedItem := makeGiftItem(ownerID)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		CreateFunc: func(ctx context.Context, gi models.GiftItem) (*models.GiftItem, error) {
			assert.False(t, gi.Description.Valid)
			assert.False(t, gi.Link.Valid)
			assert.False(t, gi.ImageUrl.Valid)
			assert.False(t, gi.Notes.Valid)
			return returnedItem, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	_, err := svc.CreateItem(context.Background(), ownerStr, CreateItemInput{
		Title: "Minimal Item",
	})

	require.NoError(t, err)
}

func TestItemService_CreateItem_RepoError(t *testing.T) {
	_, ownerStr := newValidPgtypeUUID(t)
	repoErr := errors.New("insert failed")

	itemRepo := &GiftItemRepositoryInterfaceMock{
		CreateFunc: func(ctx context.Context, gi models.GiftItem) (*models.GiftItem, error) {
			return nil, repoErr
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.CreateItem(context.Background(), ownerStr, CreateItemInput{
		Title: "Will Fail",
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create item")
}

// ---------------------------------------------------------------------------
// GetItem
// ---------------------------------------------------------------------------

func TestItemService_GetItem_Success(t *testing.T) {
	ownerID, ownerStr := newValidPgtypeUUID(t)
	item := makeGiftItem(ownerID)
	itemIDStr := item.ID.String()

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			assert.Equal(t, item.ID, id)
			return item, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.GetItem(context.Background(), itemIDStr, ownerStr)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, item.Name, result.Name)
	assert.Equal(t, ownerID.String(), result.OwnerID)
	assert.Len(t, itemRepo.GetByIDCalls(), 1)
}

func TestItemService_GetItem_InvalidItemID(t *testing.T) {
	itemRepo := &GiftItemRepositoryInterfaceMock{}
	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})

	result, err := svc.GetItem(context.Background(), "not-valid", uuid.New().String())

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrItemNotFound))
	assert.Empty(t, itemRepo.GetByIDCalls())
}

func TestItemService_GetItem_InvalidUserID(t *testing.T) {
	itemID := pgtype.UUID{}
	_ = itemID.Scan(uuid.New().String())

	itemRepo := &GiftItemRepositoryInterfaceMock{}
	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})

	result, err := svc.GetItem(context.Background(), itemID.String(), "bad-user-id")

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrInvalidItemUser))
	assert.Empty(t, itemRepo.GetByIDCalls())
}

func TestItemService_GetItem_NotFound(t *testing.T) {
	_, ownerStr := newValidPgtypeUUID(t)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return nil, errors.New("not found")
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.GetItem(context.Background(), uuid.New().String(), ownerStr)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrItemNotFound))
}

func TestItemService_GetItem_Forbidden(t *testing.T) {
	ownerID, _ := newValidPgtypeUUID(t)
	_, differentUserStr := newValidPgtypeUUID(t)
	item := makeGiftItem(ownerID)
	itemIDStr := item.ID.String()

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return item, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.GetItem(context.Background(), itemIDStr, differentUserStr)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrItemForbidden))
}

// ---------------------------------------------------------------------------
// UpdateItem
// ---------------------------------------------------------------------------

func TestItemService_UpdateItem_Success_AllFields(t *testing.T) {
	ownerID, ownerStr := newValidPgtypeUUID(t)
	existingItem := makeGiftItem(ownerID)
	itemIDStr := existingItem.ID.String()

	updatedItem := makeGiftItem(ownerID)
	updatedItem.ID = existingItem.ID
	updatedItem.Name = "Updated Title"
	updatedItem.Description = pgtype.Text{String: "Updated desc", Valid: true}

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return existingItem, nil
		},
		UpdateWithNewSchemaFunc: func(ctx context.Context, gi *models.GiftItem) (*models.GiftItem, error) {
			assert.Equal(t, "Updated Title", gi.Name)
			assert.Equal(t, "Updated desc", gi.Description.String)
			assert.True(t, gi.Description.Valid)
			assert.Equal(t, "https://new.link", gi.Link.String)
			assert.Equal(t, int32(7), gi.Priority.Int32)
			return updatedItem, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.UpdateItem(context.Background(), itemIDStr, ownerStr, UpdateItemInput{
		Title:       stringPtr("Updated Title"),
		Description: stringPtr("Updated desc"),
		Link:        stringPtr("https://new.link"),
		ImageURL:    stringPtr("https://new.img.jpg"),
		Price:       float64Ptr(99.99),
		Priority:    intPtr(7),
		Notes:       stringPtr("New notes"),
	})

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, itemRepo.GetByIDCalls(), 1)
	assert.Len(t, itemRepo.UpdateWithNewSchemaCalls(), 1)
}

func TestItemService_UpdateItem_PartialUpdate(t *testing.T) {
	ownerID, ownerStr := newValidPgtypeUUID(t)
	existingItem := makeGiftItem(ownerID)
	existingItem.Name = "Original Title"
	existingItem.Description = pgtype.Text{String: "Original desc", Valid: true}
	itemIDStr := existingItem.ID.String()

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return existingItem, nil
		},
		UpdateWithNewSchemaFunc: func(ctx context.Context, gi *models.GiftItem) (*models.GiftItem, error) {
			// Only title should change; description should remain the same
			assert.Equal(t, "New Title Only", gi.Name)
			assert.Equal(t, "Original desc", gi.Description.String)
			return gi, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.UpdateItem(context.Background(), itemIDStr, ownerStr, UpdateItemInput{
		Title: stringPtr("New Title Only"),
	})

	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestItemService_UpdateItem_ClearOptionalField(t *testing.T) {
	ownerID, ownerStr := newValidPgtypeUUID(t)
	existingItem := makeGiftItem(ownerID)
	itemIDStr := existingItem.ID.String()

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return existingItem, nil
		},
		UpdateWithNewSchemaFunc: func(ctx context.Context, gi *models.GiftItem) (*models.GiftItem, error) {
			// Setting empty string should result in Valid=false
			assert.False(t, gi.Description.Valid, "empty description should set Valid=false")
			return gi, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	_, err := svc.UpdateItem(context.Background(), itemIDStr, ownerStr, UpdateItemInput{
		Description: stringPtr(""),
	})

	require.NoError(t, err)
}

func TestItemService_UpdateItem_InvalidItemID(t *testing.T) {
	itemRepo := &GiftItemRepositoryInterfaceMock{}
	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})

	result, err := svc.UpdateItem(context.Background(), "bad-id", uuid.New().String(), UpdateItemInput{
		Title: stringPtr("X"),
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrItemNotFound))
}

func TestItemService_UpdateItem_InvalidUserID(t *testing.T) {
	itemRepo := &GiftItemRepositoryInterfaceMock{}
	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})

	result, err := svc.UpdateItem(context.Background(), uuid.New().String(), "bad-user", UpdateItemInput{
		Title: stringPtr("X"),
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrInvalidItemUser))
}

func TestItemService_UpdateItem_NotFound(t *testing.T) {
	_, ownerStr := newValidPgtypeUUID(t)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return nil, errors.New("not found")
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.UpdateItem(context.Background(), uuid.New().String(), ownerStr, UpdateItemInput{
		Title: stringPtr("X"),
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrItemNotFound))
}

func TestItemService_UpdateItem_Forbidden(t *testing.T) {
	ownerID, _ := newValidPgtypeUUID(t)
	_, differentUserStr := newValidPgtypeUUID(t)
	existingItem := makeGiftItem(ownerID)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return existingItem, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.UpdateItem(context.Background(), existingItem.ID.String(), differentUserStr, UpdateItemInput{
		Title: stringPtr("Stolen Title"),
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrItemForbidden))
	assert.Empty(t, itemRepo.UpdateWithNewSchemaCalls(), "update should not be called when forbidden")
}

func TestItemService_UpdateItem_RepoError(t *testing.T) {
	ownerID, ownerStr := newValidPgtypeUUID(t)
	existingItem := makeGiftItem(ownerID)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return existingItem, nil
		},
		UpdateWithNewSchemaFunc: func(ctx context.Context, gi *models.GiftItem) (*models.GiftItem, error) {
			return nil, errors.New("update failed")
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.UpdateItem(context.Background(), existingItem.ID.String(), ownerStr, UpdateItemInput{
		Title: stringPtr("X"),
	})

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to update item")
}

// ---------------------------------------------------------------------------
// SoftDeleteItem
// ---------------------------------------------------------------------------

func TestItemService_SoftDeleteItem_Success(t *testing.T) {
	ownerID, ownerStr := newValidPgtypeUUID(t)
	existingItem := makeGiftItem(ownerID)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return existingItem, nil
		},
		SoftDeleteFunc: func(ctx context.Context, id pgtype.UUID) error {
			assert.Equal(t, existingItem.ID, id)
			return nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	err := svc.SoftDeleteItem(context.Background(), existingItem.ID.String(), ownerStr)

	require.NoError(t, err)
	assert.Len(t, itemRepo.GetByIDCalls(), 1)
	assert.Len(t, itemRepo.SoftDeleteCalls(), 1)
}

func TestItemService_SoftDeleteItem_InvalidItemID(t *testing.T) {
	itemRepo := &GiftItemRepositoryInterfaceMock{}
	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})

	err := svc.SoftDeleteItem(context.Background(), "bad-id", uuid.New().String())

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrItemNotFound))
	assert.Empty(t, itemRepo.GetByIDCalls())
}

func TestItemService_SoftDeleteItem_InvalidUserID(t *testing.T) {
	itemRepo := &GiftItemRepositoryInterfaceMock{}
	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})

	err := svc.SoftDeleteItem(context.Background(), uuid.New().String(), "bad-user")

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidItemUser))
}

func TestItemService_SoftDeleteItem_NotFound(t *testing.T) {
	_, ownerStr := newValidPgtypeUUID(t)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return nil, errors.New("not found")
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	err := svc.SoftDeleteItem(context.Background(), uuid.New().String(), ownerStr)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrItemNotFound))
}

func TestItemService_SoftDeleteItem_Forbidden(t *testing.T) {
	ownerID, _ := newValidPgtypeUUID(t)
	_, differentUserStr := newValidPgtypeUUID(t)
	existingItem := makeGiftItem(ownerID)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return existingItem, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	err := svc.SoftDeleteItem(context.Background(), existingItem.ID.String(), differentUserStr)

	require.Error(t, err)
	assert.True(t, errors.Is(err, ErrItemForbidden))
	assert.Empty(t, itemRepo.SoftDeleteCalls(), "soft delete should not be called when forbidden")
}

func TestItemService_SoftDeleteItem_RepoError(t *testing.T) {
	ownerID, ownerStr := newValidPgtypeUUID(t)
	existingItem := makeGiftItem(ownerID)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return existingItem, nil
		},
		SoftDeleteFunc: func(ctx context.Context, id pgtype.UUID) error {
			return errors.New("archive failed")
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	err := svc.SoftDeleteItem(context.Background(), existingItem.ID.String(), ownerStr)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to archive item")
}

// ---------------------------------------------------------------------------
// MarkPurchased
// ---------------------------------------------------------------------------

func TestItemService_MarkPurchased_Success(t *testing.T) {
	ownerID, _ := newValidPgtypeUUID(t)
	_, buyerStr := newValidPgtypeUUID(t)
	existingItem := makeGiftItem(ownerID)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return existingItem, nil
		},
		UpdateWithNewSchemaFunc: func(ctx context.Context, gi *models.GiftItem) (*models.GiftItem, error) {
			assert.True(t, gi.PurchasedByUserID.Valid, "PurchasedByUserID should be set")
			buyerID := pgtypeUUIDFromString(t, buyerStr)
			assert.Equal(t, buyerID.Bytes, gi.PurchasedByUserID.Bytes)
			assert.True(t, gi.PurchasedAt.Valid, "PurchasedAt should be set")
			assert.WithinDuration(t, time.Now(), gi.PurchasedAt.Time, 5*time.Second)
			assert.True(t, gi.PurchasedPrice.Valid, "PurchasedPrice should be set")
			return gi, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.MarkPurchased(context.Background(), existingItem.ID.String(), buyerStr, 29.99)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.True(t, result.IsPurchased)
	assert.Len(t, itemRepo.GetByIDCalls(), 1)
	assert.Len(t, itemRepo.UpdateWithNewSchemaCalls(), 1)
}

func TestItemService_MarkPurchased_InvalidItemID(t *testing.T) {
	itemRepo := &GiftItemRepositoryInterfaceMock{}
	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})

	result, err := svc.MarkPurchased(context.Background(), "bad-id", uuid.New().String(), 10.0)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrItemNotFound))
}

func TestItemService_MarkPurchased_InvalidUserID(t *testing.T) {
	itemRepo := &GiftItemRepositoryInterfaceMock{}
	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})

	result, err := svc.MarkPurchased(context.Background(), uuid.New().String(), "bad-user", 10.0)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrInvalidItemUser))
}

func TestItemService_MarkPurchased_ItemNotFound(t *testing.T) {
	_, buyerStr := newValidPgtypeUUID(t)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return nil, errors.New("not found")
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.MarkPurchased(context.Background(), uuid.New().String(), buyerStr, 10.0)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, ErrItemNotFound))
}

func TestItemService_MarkPurchased_RepoUpdateError(t *testing.T) {
	ownerID, _ := newValidPgtypeUUID(t)
	_, buyerStr := newValidPgtypeUUID(t)
	existingItem := makeGiftItem(ownerID)

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return existingItem, nil
		},
		UpdateWithNewSchemaFunc: func(ctx context.Context, gi *models.GiftItem) (*models.GiftItem, error) {
			return nil, errors.New("update failed")
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.MarkPurchased(context.Background(), existingItem.ID.String(), buyerStr, 29.99)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to mark item as purchased")
}

// ---------------------------------------------------------------------------
// convertToOutput (tested through public methods)
// ---------------------------------------------------------------------------

func TestItemService_ConvertToOutput_NullableFields(t *testing.T) {
	ownerID, ownerStr := newValidPgtypeUUID(t)
	item := &models.GiftItem{
		ID:        pgtypeUUIDFromString(t, uuid.New().String()),
		OwnerID:   ownerID,
		Name:      "Minimal",
		CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		// All optional fields left as zero values (Valid=false)
	}

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return item, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.GetItem(context.Background(), item.ID.String(), ownerStr)

	require.NoError(t, err)
	assert.Equal(t, "Minimal", result.Name)
	assert.Equal(t, "", result.Description)
	assert.Equal(t, "", result.Link)
	assert.Equal(t, "", result.ImageURL)
	assert.Equal(t, float64(0), result.Price)
	assert.Equal(t, 0, result.Priority)
	assert.Equal(t, "", result.Notes)
	assert.False(t, result.IsPurchased)
	assert.False(t, result.IsArchived)
}

func TestItemService_ConvertToOutput_PurchasedAndArchived(t *testing.T) {
	ownerID, ownerStr := newValidPgtypeUUID(t)
	buyerID := pgtypeUUIDFromString(t, uuid.New().String())
	now := time.Now().UTC()

	item := &models.GiftItem{
		ID:                pgtypeUUIDFromString(t, uuid.New().String()),
		OwnerID:           ownerID,
		Name:              "Purchased Item",
		PurchasedByUserID: buyerID,
		PurchasedAt:       pgtype.Timestamptz{Time: now, Valid: true},
		ArchivedAt:        pgtype.Timestamptz{Time: now, Valid: true},
		CreatedAt:         pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:         pgtype.Timestamptz{Time: now, Valid: true},
	}

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return item, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.GetItem(context.Background(), item.ID.String(), ownerStr)

	require.NoError(t, err)
	assert.True(t, result.IsPurchased, "IsPurchased should be true when PurchasedByUserID is valid")
	assert.True(t, result.IsArchived, "IsArchived should be true when ArchivedAt is valid")
}

func TestItemService_ConvertToOutput_PriceConversion(t *testing.T) {
	ownerID, ownerStr := newValidPgtypeUUID(t)
	now := time.Now().UTC()

	item := &models.GiftItem{
		ID:        pgtypeUUIDFromString(t, uuid.New().String()),
		OwnerID:   ownerID,
		Name:      "Priced Item",
		Price:     pgtype.Numeric{Int: big.NewInt(4999), Exp: -2, Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(ctx context.Context, id pgtype.UUID) (*models.GiftItem, error) {
			return item, nil
		},
	}

	svc := newItemService(itemRepo, &WishlistItemRepositoryInterfaceMock{})
	result, err := svc.GetItem(context.Background(), item.ID.String(), ownerStr)

	require.NoError(t, err)
	assert.InDelta(t, 49.99, result.Price, 0.001)
}
