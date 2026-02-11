package service

import (
	"context"
	"sync"

	"github.com/jackc/pgx/v5/pgtype"

	db "wish-list/internal/shared/db/models"
)

// Ensure, that WishListRepositoryInterfaceMock does implement WishListRepositoryInterface.
var _ WishListRepositoryInterface = &WishListRepositoryInterfaceMock{}

// WishListRepositoryInterfaceMock is a mock implementation of WishListRepositoryInterface.
type WishListRepositoryInterfaceMock struct {
	// GetByIDFunc mocks the GetByID method.
	GetByIDFunc func(ctx context.Context, id pgtype.UUID) (*db.WishList, error)

	// calls tracks calls to the methods.
	calls struct {
		GetByID []struct {
			Ctx context.Context
			ID  pgtype.UUID
		}
	}
	lockGetByID sync.RWMutex
}

// GetByID calls GetByIDFunc.
func (mock *WishListRepositoryInterfaceMock) GetByID(ctx context.Context, id pgtype.UUID) (*db.WishList, error) {
	if mock.GetByIDFunc == nil {
		panic("WishListRepositoryInterfaceMock.GetByIDFunc: method is nil but WishListRepositoryInterface.GetByID was just called")
	}
	callInfo := struct {
		Ctx context.Context
		ID  pgtype.UUID
	}{
		Ctx: ctx,
		ID:  id,
	}
	mock.lockGetByID.Lock()
	mock.calls.GetByID = append(mock.calls.GetByID, callInfo)
	mock.lockGetByID.Unlock()
	return mock.GetByIDFunc(ctx, id)
}

// GetByIDCalls gets all the calls that were made to GetByID.
func (mock *WishListRepositoryInterfaceMock) GetByIDCalls() []struct {
	Ctx context.Context
	ID  pgtype.UUID
} {
	var calls []struct {
		Ctx context.Context
		ID  pgtype.UUID
	}
	mock.lockGetByID.RLock()
	calls = mock.calls.GetByID
	mock.lockGetByID.RUnlock()
	return calls
}

// Ensure, that GiftItemRepositoryInterfaceMock does implement GiftItemRepositoryInterface.
var _ GiftItemRepositoryInterface = &GiftItemRepositoryInterfaceMock{}

// GiftItemRepositoryInterfaceMock is a mock implementation of GiftItemRepositoryInterface.
type GiftItemRepositoryInterfaceMock struct {
	// GetByIDFunc mocks the GetByID method.
	GetByIDFunc func(ctx context.Context, id pgtype.UUID) (*db.GiftItem, error)

	// CreateWithOwnerFunc mocks the CreateWithOwner method.
	CreateWithOwnerFunc func(ctx context.Context, giftItem db.GiftItem) (*db.GiftItem, error)

	// calls tracks calls to the methods.
	calls struct {
		GetByID []struct {
			Ctx context.Context
			ID  pgtype.UUID
		}
		CreateWithOwner []struct {
			Ctx      context.Context
			GiftItem db.GiftItem
		}
	}
	lockGetByID         sync.RWMutex
	lockCreateWithOwner sync.RWMutex
}

// GetByID calls GetByIDFunc.
func (mock *GiftItemRepositoryInterfaceMock) GetByID(ctx context.Context, id pgtype.UUID) (*db.GiftItem, error) {
	if mock.GetByIDFunc == nil {
		panic("GiftItemRepositoryInterfaceMock.GetByIDFunc: method is nil but GiftItemRepositoryInterface.GetByID was just called")
	}
	callInfo := struct {
		Ctx context.Context
		ID  pgtype.UUID
	}{
		Ctx: ctx,
		ID:  id,
	}
	mock.lockGetByID.Lock()
	mock.calls.GetByID = append(mock.calls.GetByID, callInfo)
	mock.lockGetByID.Unlock()
	return mock.GetByIDFunc(ctx, id)
}

// GetByIDCalls gets all the calls that were made to GetByID.
func (mock *GiftItemRepositoryInterfaceMock) GetByIDCalls() []struct {
	Ctx context.Context
	ID  pgtype.UUID
} {
	var calls []struct {
		Ctx context.Context
		ID  pgtype.UUID
	}
	mock.lockGetByID.RLock()
	calls = mock.calls.GetByID
	mock.lockGetByID.RUnlock()
	return calls
}

// CreateWithOwner calls CreateWithOwnerFunc.
func (mock *GiftItemRepositoryInterfaceMock) CreateWithOwner(ctx context.Context, giftItem db.GiftItem) (*db.GiftItem, error) {
	if mock.CreateWithOwnerFunc == nil {
		panic("GiftItemRepositoryInterfaceMock.CreateWithOwnerFunc: method is nil but GiftItemRepositoryInterface.CreateWithOwner was just called")
	}
	callInfo := struct {
		Ctx      context.Context
		GiftItem db.GiftItem
	}{
		Ctx:      ctx,
		GiftItem: giftItem,
	}
	mock.lockCreateWithOwner.Lock()
	mock.calls.CreateWithOwner = append(mock.calls.CreateWithOwner, callInfo)
	mock.lockCreateWithOwner.Unlock()
	return mock.CreateWithOwnerFunc(ctx, giftItem)
}

// CreateWithOwnerCalls gets all the calls that were made to CreateWithOwner.
func (mock *GiftItemRepositoryInterfaceMock) CreateWithOwnerCalls() []struct {
	Ctx      context.Context
	GiftItem db.GiftItem
} {
	var calls []struct {
		Ctx      context.Context
		GiftItem db.GiftItem
	}
	mock.lockCreateWithOwner.RLock()
	calls = mock.calls.CreateWithOwner
	mock.lockCreateWithOwner.RUnlock()
	return calls
}
