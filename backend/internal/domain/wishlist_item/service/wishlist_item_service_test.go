package service

import (
	"context"
	"errors"
	"testing"
	"time"

	itemmodels "wish-list/internal/domain/item/models"
	wishlistmodels "wish-list/internal/domain/wishlist/models"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- helpers ---

func strPtr(s string) *string   { return &s }
func f64Ptr(f float64) *float64 { return &f }
func i32Ptr(i int32) *int32     { return &i }

// uuidToPg converts a google/uuid to pgtype.UUID.
func uuidToPg(t *testing.T, u uuid.UUID) pgtype.UUID {
	t.Helper()
	pg := pgtype.UUID{}
	err := pg.Scan(u.String())
	require.NoError(t, err)
	return pg
}

// makeWishlistWI builds a wishlistmodels.WishList owned by ownerID.
func makeWishlistWI(t *testing.T, id, ownerID uuid.UUID, isPublic bool) *wishlistmodels.WishList {
	t.Helper()
	return &wishlistmodels.WishList{
		ID:        uuidToPg(t, id),
		OwnerID:   uuidToPg(t, ownerID),
		Title:     "Test Wishlist",
		IsPublic:  pgtype.Bool{Bool: isPublic, Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
}

// makeGiftItemWI builds a itemmodels.GiftItem with a specific item ID and owner.
func makeGiftItemWI(t *testing.T, itemID, ownerID uuid.UUID) *itemmodels.GiftItem {
	t.Helper()
	return &itemmodels.GiftItem{
		ID:        uuidToPg(t, itemID),
		OwnerID:   uuidToPg(t, ownerID),
		Name:      "Test Item",
		Priority:  pgtype.Int4{Int32: 1, Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}
}

// newTestService creates a WishlistItemService wired to the provided moq mocks.
func newTestService(
	wlRepo *WishListRepositoryInterfaceMock,
	itemRepo *GiftItemRepositoryInterfaceMock,
	wiRepo *WishlistItemRepositoryInterfaceMock,
) *WishlistItemService {
	return NewWishlistItemService(wlRepo, itemRepo, wiRepo)
}

// ============================================================
// GetWishlistItems
// ============================================================

func TestGetWishlistItems_Success_Owner(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	itemID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)
	items := []*itemmodels.GiftItem{makeGiftItemWI(t, itemID, ownerID)}

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, id pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		GetByWishlistFunc: func(_ context.Context, _ pgtype.UUID, page, limit int) ([]*itemmodels.GiftItem, error) {
			return items, nil
		},
		GetByWishlistCountFunc: func(_ context.Context, _ pgtype.UUID) (int64, error) {
			return 1, nil
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, wiRepo)

	result, err := svc.GetWishlistItems(context.Background(), wlID.String(), ownerID.String(), 1, 10)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, int64(1), result.TotalCount)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.Limit)
	assert.Equal(t, 1, result.TotalPages)
	assert.Equal(t, "Test Item", result.Items[0].Name)
}

func TestGetWishlistItems_Success_PublicWishlist_NonOwner(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	wlID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, true) // public

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		GetByWishlistFunc: func(_ context.Context, _ pgtype.UUID, _, _ int) ([]*itemmodels.GiftItem, error) {
			return []*itemmodels.GiftItem{}, nil
		},
		GetByWishlistCountFunc: func(_ context.Context, _ pgtype.UUID) (int64, error) {
			return 0, nil
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, wiRepo)

	result, err := svc.GetWishlistItems(context.Background(), wlID.String(), otherUserID.String(), 0, 0)

	require.NoError(t, err)
	require.NotNil(t, result)
	// Defaults applied: page=1, limit=10
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 10, result.Limit)
}

func TestGetWishlistItems_DefaultPagination(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	wishlist := makeWishlistWI(t, wlID, ownerID, false)

	var capturedPage, capturedLimit int
	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		GetByWishlistFunc: func(_ context.Context, _ pgtype.UUID, page, limit int) ([]*itemmodels.GiftItem, error) {
			capturedPage = page
			capturedLimit = limit
			return []*itemmodels.GiftItem{}, nil
		},
		GetByWishlistCountFunc: func(_ context.Context, _ pgtype.UUID) (int64, error) {
			return 0, nil
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, wiRepo)

	// page=0 and limit=0 should default to page=1, limit=10
	_, err := svc.GetWishlistItems(context.Background(), wlID.String(), ownerID.String(), 0, 0)

	require.NoError(t, err)
	assert.Equal(t, 1, capturedPage)
	assert.Equal(t, 10, capturedLimit)
}

func TestGetWishlistItems_LimitCappedAt100(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	wishlist := makeWishlistWI(t, wlID, ownerID, false)

	var capturedLimit int
	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		GetByWishlistFunc: func(_ context.Context, _ pgtype.UUID, _, limit int) ([]*itemmodels.GiftItem, error) {
			capturedLimit = limit
			return []*itemmodels.GiftItem{}, nil
		},
		GetByWishlistCountFunc: func(_ context.Context, _ pgtype.UUID) (int64, error) {
			return 0, nil
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, wiRepo)

	_, err := svc.GetWishlistItems(context.Background(), wlID.String(), ownerID.String(), 1, 500)

	require.NoError(t, err)
	assert.Equal(t, 100, capturedLimit)
}

func TestGetWishlistItems_InvalidWishlistID(t *testing.T) {
	svc := newTestService(
		&WishListRepositoryInterfaceMock{},
		&GiftItemRepositoryInterfaceMock{},
		&WishlistItemRepositoryInterfaceMock{},
	)

	result, err := svc.GetWishlistItems(context.Background(), "not-a-uuid", uuid.New().String(), 1, 10)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidWishlistItemWLID)
}

func TestGetWishlistItems_InvalidUserID(t *testing.T) {
	svc := newTestService(
		&WishListRepositoryInterfaceMock{},
		&GiftItemRepositoryInterfaceMock{},
		&WishlistItemRepositoryInterfaceMock{},
	)

	result, err := svc.GetWishlistItems(context.Background(), uuid.New().String(), "bad-user-id", 1, 10)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidWishlistItemUser)
}

func TestGetWishlistItems_WishlistNotFound(t *testing.T) {
	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return nil, errors.New("not found")
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, &WishlistItemRepositoryInterfaceMock{})

	result, err := svc.GetWishlistItems(context.Background(), uuid.New().String(), uuid.New().String(), 1, 10)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrWishListNotFound)
}

func TestGetWishlistItems_Forbidden_PrivateWishlist_NonOwner(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	wlID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false) // private

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, &WishlistItemRepositoryInterfaceMock{})

	result, err := svc.GetWishlistItems(context.Background(), wlID.String(), otherUserID.String(), 1, 10)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrWishListForbidden)
}

func TestGetWishlistItems_RepoGetByWishlistError(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	wishlist := makeWishlistWI(t, wlID, ownerID, false)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		GetByWishlistFunc: func(_ context.Context, _ pgtype.UUID, _, _ int) ([]*itemmodels.GiftItem, error) {
			return nil, errors.New("db error")
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, wiRepo)

	result, err := svc.GetWishlistItems(context.Background(), wlID.String(), ownerID.String(), 1, 10)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to get wishlist items")
}

func TestGetWishlistItems_RepoGetByWishlistCountError(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	wishlist := makeWishlistWI(t, wlID, ownerID, false)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		GetByWishlistFunc: func(_ context.Context, _ pgtype.UUID, _, _ int) ([]*itemmodels.GiftItem, error) {
			return []*itemmodels.GiftItem{}, nil
		},
		GetByWishlistCountFunc: func(_ context.Context, _ pgtype.UUID) (int64, error) {
			return 0, errors.New("count error")
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, wiRepo)

	result, err := svc.GetWishlistItems(context.Background(), wlID.String(), ownerID.String(), 1, 10)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to count wishlist items")
}

func TestGetWishlistItems_TotalPagesCalculation(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	wishlist := makeWishlistWI(t, wlID, ownerID, false)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		GetByWishlistFunc: func(_ context.Context, _ pgtype.UUID, _, _ int) ([]*itemmodels.GiftItem, error) {
			return []*itemmodels.GiftItem{}, nil
		},
		GetByWishlistCountFunc: func(_ context.Context, _ pgtype.UUID) (int64, error) {
			return 25, nil
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, wiRepo)

	result, err := svc.GetWishlistItems(context.Background(), wlID.String(), ownerID.String(), 1, 10)

	require.NoError(t, err)
	// 25 items / 10 per page = 3 pages (ceil)
	assert.Equal(t, 3, result.TotalPages)
	assert.Equal(t, int64(25), result.TotalCount)
}

// ============================================================
// AttachItem
// ============================================================

func TestAttachItem_Success(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	itemID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)
	item := makeGiftItemWI(t, itemID, ownerID)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*itemmodels.GiftItem, error) {
			return item, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		IsAttachedFunc: func(_ context.Context, _, _ pgtype.UUID) (bool, error) {
			return false, nil
		},
		AttachFunc: func(_ context.Context, _, _ pgtype.UUID) error {
			return nil
		},
	}

	svc := newTestService(wlRepo, itemRepo, wiRepo)

	err := svc.AttachItem(context.Background(), wlID.String(), itemID.String(), ownerID.String())

	require.NoError(t, err)
	assert.Len(t, wiRepo.AttachCalls(), 1)
}

func TestAttachItem_InvalidWishlistID(t *testing.T) {
	svc := newTestService(
		&WishListRepositoryInterfaceMock{},
		&GiftItemRepositoryInterfaceMock{},
		&WishlistItemRepositoryInterfaceMock{},
	)

	err := svc.AttachItem(context.Background(), "bad-wl-id", uuid.New().String(), uuid.New().String())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidWishlistItemWLID)
}

func TestAttachItem_InvalidItemID(t *testing.T) {
	svc := newTestService(
		&WishListRepositoryInterfaceMock{},
		&GiftItemRepositoryInterfaceMock{},
		&WishlistItemRepositoryInterfaceMock{},
	)

	err := svc.AttachItem(context.Background(), uuid.New().String(), "bad-item-id", uuid.New().String())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidWishlistItemID)
}

func TestAttachItem_InvalidUserID(t *testing.T) {
	svc := newTestService(
		&WishListRepositoryInterfaceMock{},
		&GiftItemRepositoryInterfaceMock{},
		&WishlistItemRepositoryInterfaceMock{},
	)

	err := svc.AttachItem(context.Background(), uuid.New().String(), uuid.New().String(), "bad-user-id")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidWishlistItemUser)
}

func TestAttachItem_WishlistNotFound(t *testing.T) {
	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return nil, errors.New("not found")
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, &WishlistItemRepositoryInterfaceMock{})

	err := svc.AttachItem(context.Background(), uuid.New().String(), uuid.New().String(), uuid.New().String())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrWishListNotFound)
}

func TestAttachItem_WishlistForbidden_NotOwner(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	wlID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, &WishlistItemRepositoryInterfaceMock{})

	err := svc.AttachItem(context.Background(), wlID.String(), uuid.New().String(), otherUserID.String())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrWishListForbidden)
}

func TestAttachItem_ItemNotFound(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*itemmodels.GiftItem, error) {
			return nil, errors.New("not found")
		},
	}

	svc := newTestService(wlRepo, itemRepo, &WishlistItemRepositoryInterfaceMock{})

	err := svc.AttachItem(context.Background(), wlID.String(), uuid.New().String(), ownerID.String())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrItemNotFound)
}

func TestAttachItem_ItemForbidden_NotOwner(t *testing.T) {
	ownerID := uuid.New()
	itemOwnerID := uuid.New()
	wlID := uuid.New()
	itemID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)
	item := makeGiftItemWI(t, itemID, itemOwnerID) // different owner

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*itemmodels.GiftItem, error) {
			return item, nil
		},
	}

	svc := newTestService(wlRepo, itemRepo, &WishlistItemRepositoryInterfaceMock{})

	err := svc.AttachItem(context.Background(), wlID.String(), itemID.String(), ownerID.String())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrItemForbidden)
}

func TestAttachItem_AlreadyAttached(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	itemID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)
	item := makeGiftItemWI(t, itemID, ownerID)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*itemmodels.GiftItem, error) {
			return item, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		IsAttachedFunc: func(_ context.Context, _, _ pgtype.UUID) (bool, error) {
			return true, nil // already attached
		},
	}

	svc := newTestService(wlRepo, itemRepo, wiRepo)

	err := svc.AttachItem(context.Background(), wlID.String(), itemID.String(), ownerID.String())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrItemAlreadyAttached)
}

func TestAttachItem_IsAttachedRepoError(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	itemID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)
	item := makeGiftItemWI(t, itemID, ownerID)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*itemmodels.GiftItem, error) {
			return item, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		IsAttachedFunc: func(_ context.Context, _, _ pgtype.UUID) (bool, error) {
			return false, errors.New("db error")
		},
	}

	svc := newTestService(wlRepo, itemRepo, wiRepo)

	err := svc.AttachItem(context.Background(), wlID.String(), itemID.String(), ownerID.String())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check attachment")
}

func TestAttachItem_AttachRepoError(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	itemID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)
	item := makeGiftItemWI(t, itemID, ownerID)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*itemmodels.GiftItem, error) {
			return item, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		IsAttachedFunc: func(_ context.Context, _, _ pgtype.UUID) (bool, error) {
			return false, nil
		},
		AttachFunc: func(_ context.Context, _, _ pgtype.UUID) error {
			return errors.New("db error")
		},
	}

	svc := newTestService(wlRepo, itemRepo, wiRepo)

	err := svc.AttachItem(context.Background(), wlID.String(), itemID.String(), ownerID.String())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to attach item")
}

// ============================================================
// CreateItemInWishlist
// ============================================================

func TestCreateItemInWishlist_Success(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	createdItemID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)
	now := time.Now()
	createdItem := &itemmodels.GiftItem{
		ID:        uuidToPg(t, createdItemID),
		OwnerID:   uuidToPg(t, ownerID),
		Name:      "New Item",
		Priority:  pgtype.Int4{Int32: 3, Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	itemRepo := &GiftItemRepositoryInterfaceMock{
		CreateWithOwnerFunc: func(_ context.Context, item itemmodels.GiftItem) (*itemmodels.GiftItem, error) {
			return createdItem, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		AttachFunc: func(_ context.Context, _, _ pgtype.UUID) error {
			return nil
		},
	}

	svc := newTestService(wlRepo, itemRepo, wiRepo)

	input := CreateItemInput{
		Title:    "New Item",
		Priority: i32Ptr(3),
	}
	result, err := svc.CreateItemInWishlist(context.Background(), wlID.String(), ownerID.String(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "New Item", result.Name)
	assert.Equal(t, 3, result.Priority)
	assert.Len(t, itemRepo.CreateWithOwnerCalls(), 1)
	assert.Len(t, wiRepo.AttachCalls(), 1)
}

func TestCreateItemInWishlist_Success_WithAllFields(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	createdItemID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)
	now := time.Now()
	createdItem := &itemmodels.GiftItem{
		ID:          uuidToPg(t, createdItemID),
		OwnerID:     uuidToPg(t, ownerID),
		Name:        "Full Item",
		Description: pgtype.Text{String: "A description", Valid: true},
		Link:        pgtype.Text{String: "https://example.com", Valid: true},
		ImageUrl:    pgtype.Text{String: "https://example.com/img.jpg", Valid: true},
		Priority:    pgtype.Int4{Int32: 5, Valid: true},
		Notes:       pgtype.Text{String: "Some notes", Valid: true},
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
	}

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}

	var capturedItem itemmodels.GiftItem
	itemRepo := &GiftItemRepositoryInterfaceMock{
		CreateWithOwnerFunc: func(_ context.Context, item itemmodels.GiftItem) (*itemmodels.GiftItem, error) {
			capturedItem = item
			return createdItem, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		AttachFunc: func(_ context.Context, _, _ pgtype.UUID) error {
			return nil
		},
	}

	svc := newTestService(wlRepo, itemRepo, wiRepo)

	input := CreateItemInput{
		Title:       "Full Item",
		Description: strPtr("A description"),
		Link:        strPtr("https://example.com"),
		ImageURL:    strPtr("https://example.com/img.jpg"),
		Price:       f64Ptr(19.99),
		Priority:    i32Ptr(5),
		Notes:       strPtr("Some notes"),
	}
	result, err := svc.CreateItemInWishlist(context.Background(), wlID.String(), ownerID.String(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "Full Item", result.Name)
	// Verify captured item has correct owner and fields
	assert.Equal(t, uuidToPg(t, ownerID), capturedItem.OwnerID)
	assert.Equal(t, "Full Item", capturedItem.Name)
	assert.True(t, capturedItem.Description.Valid)
	assert.Equal(t, "A description", capturedItem.Description.String)
	assert.True(t, capturedItem.Link.Valid)
	assert.True(t, capturedItem.ImageUrl.Valid)
	assert.True(t, capturedItem.Notes.Valid)
	assert.True(t, capturedItem.Price.Valid)
}

func TestCreateItemInWishlist_EmptyTitle(t *testing.T) {
	svc := newTestService(
		&WishListRepositoryInterfaceMock{},
		&GiftItemRepositoryInterfaceMock{},
		&WishlistItemRepositoryInterfaceMock{},
	)

	input := CreateItemInput{Title: ""}
	result, err := svc.CreateItemInWishlist(context.Background(), uuid.New().String(), uuid.New().String(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrWishlistItemTitleRequired)
}

func TestCreateItemInWishlist_InvalidWishlistID(t *testing.T) {
	svc := newTestService(
		&WishListRepositoryInterfaceMock{},
		&GiftItemRepositoryInterfaceMock{},
		&WishlistItemRepositoryInterfaceMock{},
	)

	input := CreateItemInput{Title: "Some Item"}
	result, err := svc.CreateItemInWishlist(context.Background(), "bad-wl-id", uuid.New().String(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidWishlistItemWLID)
}

func TestCreateItemInWishlist_InvalidUserID(t *testing.T) {
	svc := newTestService(
		&WishListRepositoryInterfaceMock{},
		&GiftItemRepositoryInterfaceMock{},
		&WishlistItemRepositoryInterfaceMock{},
	)

	input := CreateItemInput{Title: "Some Item"}
	result, err := svc.CreateItemInWishlist(context.Background(), uuid.New().String(), "bad-user-id", input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrInvalidWishlistItemUser)
}

func TestCreateItemInWishlist_WishlistNotFound(t *testing.T) {
	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return nil, errors.New("not found")
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, &WishlistItemRepositoryInterfaceMock{})

	input := CreateItemInput{Title: "Some Item"}
	result, err := svc.CreateItemInWishlist(context.Background(), uuid.New().String(), uuid.New().String(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrWishListNotFound)
}

func TestCreateItemInWishlist_WishlistForbidden(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	wlID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, &WishlistItemRepositoryInterfaceMock{})

	input := CreateItemInput{Title: "Some Item"}
	result, err := svc.CreateItemInWishlist(context.Background(), wlID.String(), otherUserID.String(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrWishListForbidden)
}

func TestCreateItemInWishlist_CreateItemRepoError(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	itemRepo := &GiftItemRepositoryInterfaceMock{
		CreateWithOwnerFunc: func(_ context.Context, _ itemmodels.GiftItem) (*itemmodels.GiftItem, error) {
			return nil, errors.New("db error")
		},
	}

	svc := newTestService(wlRepo, itemRepo, &WishlistItemRepositoryInterfaceMock{})

	input := CreateItemInput{Title: "Some Item"}
	result, err := svc.CreateItemInWishlist(context.Background(), wlID.String(), ownerID.String(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to create item")
}

func TestCreateItemInWishlist_AttachRepoError(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	createdItemID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)
	now := time.Now()
	createdItem := &itemmodels.GiftItem{
		ID:        uuidToPg(t, createdItemID),
		OwnerID:   uuidToPg(t, ownerID),
		Name:      "Item",
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	itemRepo := &GiftItemRepositoryInterfaceMock{
		CreateWithOwnerFunc: func(_ context.Context, _ itemmodels.GiftItem) (*itemmodels.GiftItem, error) {
			return createdItem, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		AttachFunc: func(_ context.Context, _, _ pgtype.UUID) error {
			return errors.New("attach failed")
		},
	}

	svc := newTestService(wlRepo, itemRepo, wiRepo)

	input := CreateItemInput{Title: "Item"}
	result, err := svc.CreateItemInWishlist(context.Background(), wlID.String(), ownerID.String(), input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to attach item to wishlist")
}

// ============================================================
// DetachItem
// ============================================================

func TestDetachItem_Success(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	itemID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		IsAttachedFunc: func(_ context.Context, _, _ pgtype.UUID) (bool, error) {
			return true, nil
		},
		DetachFunc: func(_ context.Context, _, _ pgtype.UUID) error {
			return nil
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, wiRepo)

	err := svc.DetachItem(context.Background(), wlID.String(), itemID.String(), ownerID.String())

	require.NoError(t, err)
	assert.Len(t, wiRepo.DetachCalls(), 1)
}

func TestDetachItem_InvalidWishlistID(t *testing.T) {
	svc := newTestService(
		&WishListRepositoryInterfaceMock{},
		&GiftItemRepositoryInterfaceMock{},
		&WishlistItemRepositoryInterfaceMock{},
	)

	err := svc.DetachItem(context.Background(), "bad-wl-id", uuid.New().String(), uuid.New().String())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidWishlistItemWLID)
}

func TestDetachItem_InvalidItemID(t *testing.T) {
	svc := newTestService(
		&WishListRepositoryInterfaceMock{},
		&GiftItemRepositoryInterfaceMock{},
		&WishlistItemRepositoryInterfaceMock{},
	)

	err := svc.DetachItem(context.Background(), uuid.New().String(), "bad-item-id", uuid.New().String())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidWishlistItemID)
}

func TestDetachItem_InvalidUserID(t *testing.T) {
	svc := newTestService(
		&WishListRepositoryInterfaceMock{},
		&GiftItemRepositoryInterfaceMock{},
		&WishlistItemRepositoryInterfaceMock{},
	)

	err := svc.DetachItem(context.Background(), uuid.New().String(), uuid.New().String(), "bad-user-id")

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidWishlistItemUser)
}

func TestDetachItem_WishlistNotFound(t *testing.T) {
	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return nil, errors.New("not found")
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, &WishlistItemRepositoryInterfaceMock{})

	err := svc.DetachItem(context.Background(), uuid.New().String(), uuid.New().String(), uuid.New().String())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrWishListNotFound)
}

func TestDetachItem_WishlistForbidden_NotOwner(t *testing.T) {
	ownerID := uuid.New()
	otherUserID := uuid.New()
	wlID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, &WishlistItemRepositoryInterfaceMock{})

	err := svc.DetachItem(context.Background(), wlID.String(), uuid.New().String(), otherUserID.String())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrWishListForbidden)
}

func TestDetachItem_ItemNotInWishlist(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	itemID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		IsAttachedFunc: func(_ context.Context, _, _ pgtype.UUID) (bool, error) {
			return false, nil // not attached
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, wiRepo)

	err := svc.DetachItem(context.Background(), wlID.String(), itemID.String(), ownerID.String())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrItemNotInWishlist)
}

func TestDetachItem_IsAttachedRepoError(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	itemID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		IsAttachedFunc: func(_ context.Context, _, _ pgtype.UUID) (bool, error) {
			return false, errors.New("db error")
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, wiRepo)

	err := svc.DetachItem(context.Background(), wlID.String(), itemID.String(), ownerID.String())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check attachment")
}

func TestDetachItem_DetachRepoError(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	itemID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		IsAttachedFunc: func(_ context.Context, _, _ pgtype.UUID) (bool, error) {
			return true, nil
		},
		DetachFunc: func(_ context.Context, _, _ pgtype.UUID) error {
			return errors.New("detach failed")
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, wiRepo)

	err := svc.DetachItem(context.Background(), wlID.String(), itemID.String(), ownerID.String())

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to detach item")
}

// ============================================================
// Edge cases and additional coverage
// ============================================================

func TestGetWishlistItems_Forbidden_IsPublicInvalid(t *testing.T) {
	// When IsPublic.Valid is false, the wishlist is treated as private.
	ownerID := uuid.New()
	otherUserID := uuid.New()
	wlID := uuid.New()

	wishlist := &wishlistmodels.WishList{
		ID:        uuidToPg(t, wlID),
		OwnerID:   uuidToPg(t, ownerID),
		Title:     "Wishlist",
		IsPublic:  pgtype.Bool{Bool: false, Valid: false}, // Valid=false
		CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
	}

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, &WishlistItemRepositoryInterfaceMock{})

	result, err := svc.GetWishlistItems(context.Background(), wlID.String(), otherUserID.String(), 1, 10)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.ErrorIs(t, err, ErrWishListForbidden)
}

func TestCreateItemInWishlist_NoPriceSet(t *testing.T) {
	// Verify that when Price is 0, the price field is not set on the item model.
	ownerID := uuid.New()
	wlID := uuid.New()
	createdItemID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)
	now := time.Now()
	createdItem := &itemmodels.GiftItem{
		ID:        uuidToPg(t, createdItemID),
		OwnerID:   uuidToPg(t, ownerID),
		Name:      "No Price Item",
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}

	var capturedItem itemmodels.GiftItem
	itemRepo := &GiftItemRepositoryInterfaceMock{
		CreateWithOwnerFunc: func(_ context.Context, item itemmodels.GiftItem) (*itemmodels.GiftItem, error) {
			capturedItem = item
			return createdItem, nil
		},
	}
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		AttachFunc: func(_ context.Context, _, _ pgtype.UUID) error {
			return nil
		},
	}

	svc := newTestService(wlRepo, itemRepo, wiRepo)

	input := CreateItemInput{
		Title: "No Price Item",
		// Price nil means no price
	}
	result, err := svc.CreateItemInWishlist(context.Background(), wlID.String(), ownerID.String(), input)

	require.NoError(t, err)
	require.NotNil(t, result)
	// Price.Valid should remain false when Price <= 0
	assert.False(t, capturedItem.Price.Valid)
}

func TestAttachItem_VerifiesCorrectIDsPassedToRepo(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	itemID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)
	item := makeGiftItemWI(t, itemID, ownerID)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, id pgtype.UUID) (*wishlistmodels.WishList, error) {
			assert.Equal(t, uuidToPg(t, wlID), id)
			return wishlist, nil
		},
	}
	itemRepo := &GiftItemRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, id pgtype.UUID) (*itemmodels.GiftItem, error) {
			assert.Equal(t, uuidToPg(t, itemID), id)
			return item, nil
		},
	}

	var capturedWLID, capturedItemID pgtype.UUID
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		IsAttachedFunc: func(_ context.Context, wl, it pgtype.UUID) (bool, error) {
			return false, nil
		},
		AttachFunc: func(_ context.Context, wl, it pgtype.UUID) error {
			capturedWLID = wl
			capturedItemID = it
			return nil
		},
	}

	svc := newTestService(wlRepo, itemRepo, wiRepo)

	err := svc.AttachItem(context.Background(), wlID.String(), itemID.String(), ownerID.String())

	require.NoError(t, err)
	assert.Equal(t, uuidToPg(t, wlID), capturedWLID)
	assert.Equal(t, uuidToPg(t, itemID), capturedItemID)
}

func TestDetachItem_VerifiesCorrectIDsPassedToRepo(t *testing.T) {
	ownerID := uuid.New()
	wlID := uuid.New()
	itemID := uuid.New()

	wishlist := makeWishlistWI(t, wlID, ownerID, false)

	wlRepo := &WishListRepositoryInterfaceMock{
		GetByIDFunc: func(_ context.Context, _ pgtype.UUID) (*wishlistmodels.WishList, error) {
			return wishlist, nil
		},
	}

	var capturedWLID, capturedItemID pgtype.UUID
	wiRepo := &WishlistItemRepositoryInterfaceMock{
		IsAttachedFunc: func(_ context.Context, _, _ pgtype.UUID) (bool, error) {
			return true, nil
		},
		DetachFunc: func(_ context.Context, wl, it pgtype.UUID) error {
			capturedWLID = wl
			capturedItemID = it
			return nil
		},
	}

	svc := newTestService(wlRepo, &GiftItemRepositoryInterfaceMock{}, wiRepo)

	err := svc.DetachItem(context.Background(), wlID.String(), itemID.String(), ownerID.String())

	require.NoError(t, err)
	assert.Equal(t, uuidToPg(t, wlID), capturedWLID)
	assert.Equal(t, uuidToPg(t, itemID), capturedItemID)
}
