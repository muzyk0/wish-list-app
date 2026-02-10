# Domain Interface Contracts

**Feature**: 003-backend-arch-migration
**Date**: 2026-02-10

## Overview

This document defines the interface contracts between domains. Per the clarification decision, cross-domain communication uses **interface injection at startup**. Each consuming domain defines the interface it needs, and the app layer wires the concrete implementations.

## Convention

Following Go's idiomatic pattern: **"Accept interfaces, return structs."**

- Interfaces are defined by the **consumer**, not the provider
- Interface names use the `Interface` suffix (matching existing project convention)
- Constructors return interfaces (existing pattern: `NewXxxRepository() XxxRepositoryInterface`)

## Domain Route Registration Contract

Every domain that serves HTTP endpoints MUST export a route registration function:

```go
// Pattern for every domain's routes.go
package http

import "github.com/labstack/echo/v4"

func RegisterRoutes(g *echo.Group, h *Handler, authMiddleware echo.MiddlewareFunc) {
    // Domain-specific route registration
}
```

The central router calls each domain's `RegisterRoutes()` with the appropriate Echo group and middleware.

## Cross-Domain Interface Contracts

### User Domain → Auth Domain

Auth domain needs user lookup for login/registration:

```go
// Defined in: domain/auth/service/
type UserServiceInterface interface {
    GetByEmail(ctx context.Context, email string) (*usermodels.User, error)
    GetByID(ctx context.Context, id pgtype.UUID) (*usermodels.User, error)
    Create(ctx context.Context, user *usermodels.User) error
}
```

### Item Domain → Wishlist, WishlistItem, Reservation Domains

Multiple domains need to look up gift items:

```go
// Defined in consuming domain's service package
type GiftItemRepositoryInterface interface {
    GetByID(ctx context.Context, id pgtype.UUID) (*itemmodels.GiftItem, error)
    GetByOwnerID(ctx context.Context, ownerID pgtype.UUID) ([]itemmodels.GiftItem, error)
    // Additional methods as needed by the specific consumer
}
```

### Wishlist Domain → WishlistItem Domain

WishlistItem domain needs wishlist existence verification:

```go
// Defined in: domain/wishlist_item/service/
type WishListRepositoryInterface interface {
    GetByID(ctx context.Context, id pgtype.UUID) (*wishlistmodels.WishList, error)
}
```

### Reservation Domain → Wishlist Domain

Wishlist domain needs reservation data for display:

```go
// Defined in: domain/wishlist/service/
type ReservationRepositoryInterface interface {
    GetByWishlistID(ctx context.Context, wishlistID pgtype.UUID) ([]reservationmodels.Reservation, error)
}
```

### Cache Interface → Wishlist Domain

Wishlist domain uses caching:

```go
// Defined in: pkg/cache/
type CacheInterface interface {
    Get(ctx context.Context, key string) (string, error)
    Set(ctx context.Context, key string, value string, ttl time.Duration) error
    Delete(ctx context.Context, key string) error
}
```

### Encryption Interface → User, Reservation Domains

Domains with PII use optional encryption:

```go
// Defined in: pkg/encryption/
type EncryptionServiceInterface interface {
    Encrypt(plaintext string) (string, error)
    Decrypt(ciphertext string) (string, error)
}
```

## App Layer Wiring Contract

The `app.go` or `main.go` wires all cross-domain dependencies:

```go
// Pseudocode for app layer wiring
func (a *App) initializeDomains() {
    // Repositories (depend on database only)
    userRepo := userrepository.New(a.db)
    wishlistRepo := wishlistrepository.New(a.db)
    itemRepo := itemrepository.New(a.db)
    wishlistItemRepo := wishlistitemrepository.New(a.db)
    reservationRepo := reservationrepository.New(a.db)
    templateRepo := wishlistrepository.NewTemplateRepository(a.db)

    // Services (depend on repositories, cross-domain via interfaces)
    userService := userservice.New(userRepo)
    wishlistService := wishlistservice.New(wishlistRepo, itemRepo, templateRepo, a.emailService, reservationRepo, a.cache)
    itemService := itemservice.New(itemRepo, wishlistItemRepo)
    wishlistItemService := wishlistitemservice.New(wishlistRepo, itemRepo, wishlistItemRepo)
    reservationService := reservationservice.New(reservationRepo, itemRepo)

    // Handlers (depend on services)
    a.userHandler = userhandler.New(userService, a.tokenManager, a.cleanupService, a.analytics)
    a.authHandler = authhandler.New(userService, a.tokenManager, a.codeStore)
    a.wishlistHandler = wishlisthandler.New(wishlistService)
    a.itemHandler = itemhandler.New(itemService)
    a.wishlistItemHandler = wishlistitemhandler.New(wishlistItemService)
    a.reservationHandler = reservationhandler.New(reservationService)
    a.healthHandler = healthhandler.New(a.db)
    a.storageHandler = storagehandler.New(a.s3Client)
}
```

## Existing Interface Compliance

All existing repository interfaces in the current codebase already follow the interface-based constructor pattern. The migration moves these interfaces into domain packages without changing their method signatures. This ensures SC-001 (100% test pass rate) is maintained.
