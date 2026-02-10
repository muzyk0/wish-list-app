# Data Model: Backend Architecture Migration

**Feature**: 003-backend-arch-migration
**Date**: 2026-02-10

## Overview

This migration does not change the database schema. The "data model" for this feature is the **package/module structure** — how Go packages, interfaces, and types are organized.

## Package Zones

### Zone 1: `internal/app/` — Application Infrastructure

Packages that support the application lifecycle. No business logic.

| Package | Responsibility | Exports |
|---------|---------------|---------|
| `app/config` | Environment variable loading | `Config`, `Load()` |
| `app/database` | DB connection pool, executor interface | `DB`, `Executor`, `New()` |
| `app/database/migrations` | SQL migration files | (raw SQL files) |
| `app/server` | Echo server lifecycle, router | `Server`, `SetupRoutes()` |
| `app/middleware` | HTTP middleware pipeline | `SecurityHeaders()`, `CORS()`, `RateLimit()`, `Timeout()` |
| `app/swagger` | API documentation setup | `InitSwagger()` |
| `app/jobs` | Background services | `AccountCleanupService`, `EmailService` |
| `app` (root) | Application factory | `App`, `New()`, `Run()`, `Shutdown()` |

**Dependency rule**: `app/` may import `pkg/` and `domain/`. Never imported by `pkg/` or `domain/`.

### Zone 2: `internal/pkg/` — Shared Libraries

Reusable packages with zero domain knowledge.

| Package | Responsibility | Exports |
|---------|---------------|---------|
| `pkg/auth` | JWT token management, middleware | `TokenManager`, `CodeStore`, `JWTMiddleware()` |
| `pkg/encryption` | PII encryption with KMS | `Service`, `GetOrCreateDataKey()` |
| `pkg/aws` | AWS S3 client | `S3Client`, `NewS3Client()` |
| `pkg/cache` | Redis caching interface | `CacheInterface`, `RedisCache` |
| `pkg/validation` | Input field validation | `Validator`, `NewValidator()` |
| `pkg/analytics` | Event tracking | `AnalyticsService` |
| `pkg/response` | Standardized HTTP responses | `Success()`, `Error()`, `ValidationError()` |
| `pkg/helpers` | Utility functions | Various test helpers |

**Dependency rule**: `pkg/` MUST NOT import `app/` or `domain/`. Only standard library and external deps.

### Zone 3: `internal/domain/` — Business Domains

Each domain is a self-contained module with the following internal structure:

```
domain/{name}/
├── delivery/
│   └── http/
│       ├── handler.go       # HTTP request handling
│       ├── dto/
│       │   ├── requests.go  # Request DTOs with validation tags
│       │   └── responses.go # Response DTOs with JSON tags
│       └── routes.go        # Route registration function
├── service/
│   ├── {name}_service.go    # Business logic
│   └── {name}_service_test.go
├── repository/
│   ├── {name}_repository.go # Data access with Executor pattern
│   └── {name}_repository_test.go
└── models/
    └── {name}.go            # Domain entity structs
```

**Dependency rule**: A domain may import `pkg/` and `app/database` (for Executor). MUST NOT import another `domain/` package.

## Domain Catalog

### auth

**Purpose**: Authentication and authorization (JWT, OAuth, mobile handoff)
**Contains**: Auth handler, OAuth handler, token-related models
**No repository**: Auth operations use the user domain's service via interface injection
**Cross-domain deps**: Needs `UserServiceInterface` (injected at startup)

### user

**Purpose**: User registration, login, profile management
**Entities**: `User` struct (id, email, first_name, last_name, password_hash, created_at, updated_at, deleted_at)
**Repository**: Full CRUD + email lookup + soft delete
**Encryption**: Optional PII encryption via `pkg/encryption`

### wishlist

**Purpose**: Wishlist CRUD, template-based creation
**Entities**: `WishList` struct (id, owner_id, title, description, is_public, created_at, updated_at)
**Repository**: Full CRUD + owner listing + public access
**Cross-domain deps**: Needs `GiftItemRepositoryInterface`, `ReservationRepositoryInterface`, `TemplateRepositoryInterface`, `CacheInterface` (all injected)

### item

**Purpose**: Gift item management
**Entities**: `GiftItem` struct (id, owner_id, name, description, price, url, image_url, created_at, updated_at)
**Repository**: Full CRUD + owner listing
**Note**: No direct `wishlist_id` — linked via wishlist_items junction table

### wishlist_item

**Purpose**: Many-to-many relationship between wishlists and items
**Entities**: `WishlistItem` struct (id, wishlist_id, gift_item_id, sort_order, created_at)
**Repository**: Link/unlink items to wishlists, ordering
**Cross-domain deps**: Needs `WishListRepositoryInterface`, `GiftItemRepositoryInterface` (injected)

### reservation

**Purpose**: Item reservation management
**Entities**: `Reservation` struct (id, gift_item_id, wishlist_id, reserver_name, reserver_email, status, created_at, updated_at)
**Repository**: Create/update/delete reservations, prevent doubles
**Encryption**: Optional PII encryption (reserver_name, reserver_email)
**Cross-domain deps**: Needs `GiftItemRepositoryInterface` (injected)

### health

**Purpose**: Application health check endpoint
**Lightweight domain**: No service or repository layer — handler queries DB directly
**Entities**: None (returns health status)

### storage

**Purpose**: File upload via S3
**Lightweight domain**: Handler uses `pkg/aws` directly
**Entities**: Upload metadata (presigned URLs)

## Interface Contracts (Cross-Domain)

Each domain that is consumed by another defines a service or repository interface that the consuming domain depends on:

```
user domain exports:
  → UserServiceInterface (consumed by auth domain)

item domain exports:
  → GiftItemRepositoryInterface (consumed by wishlist, wishlist_item, reservation domains)

wishlist domain exports:
  → WishListRepositoryInterface (consumed by wishlist_item domain)

reservation domain exports:
  → ReservationRepositoryInterface (consumed by wishlist domain)

template (part of wishlist domain):
  → TemplateRepositoryInterface (consumed within wishlist domain)
```

All interfaces are defined in the consuming domain's service package (not the providing domain), following the Go idiom of "accept interfaces, return structs."

## Migration File Mapping

| Current Path | Target Path |
|-------------|-------------|
| `shared/db/models/models.go` (User) | `domain/user/models/user.go` |
| `shared/db/models/models.go` (WishList) | `domain/wishlist/models/wishlist.go` |
| `shared/db/models/models.go` (GiftItem) | `domain/item/models/item.go` |
| `shared/db/models/models.go` (Reservation) | `domain/reservation/models/reservation.go` |
| `shared/db/models/models.go` (WishlistItem) | `domain/wishlist_item/models/wishlist_item.go` |
| `shared/db/models/db.go` (DB, Executor) | `app/database/postgres.go`, `app/database/executor.go` |
| `shared/config/config.go` | `app/config/config.go` |
| `shared/middleware/*.go` | `app/middleware/*.go` |
| `shared/encryption/*.go` | `pkg/encryption/*.go` |
| `shared/aws/*.go` | `pkg/aws/*.go` |
| `shared/cache/*.go` | `pkg/cache/*.go` |
| `shared/validation/*.go` | `pkg/validation/*.go` |
| `shared/analytics/*.go` | `pkg/analytics/*.go` |
| `shared/helpers/*.go` | `pkg/helpers/*.go` |
| `auth/*.go` | `pkg/auth/*.go` |
| `handlers/user_handler.go` | `domain/user/delivery/http/handler.go` |
| `handlers/auth_handler.go` | `domain/auth/delivery/http/handler.go` |
| `handlers/oauth_handler.go` | `domain/auth/delivery/http/oauth_handler.go` |
| `handlers/wishlist_handler.go` | `domain/wishlist/delivery/http/handler.go` |
| `handlers/item_handler.go` | `domain/item/delivery/http/handler.go` |
| `handlers/wishlist_item_handler.go` | `domain/wishlist_item/delivery/http/handler.go` |
| `handlers/reservation_handler.go` | `domain/reservation/delivery/http/handler.go` |
| `services/user_service.go` | `domain/user/service/user_service.go` |
| `services/wishlist_service.go` | `domain/wishlist/service/wishlist_service.go` |
| `services/item_service.go` | `domain/item/service/item_service.go` |
| `services/wishlist_item_service.go` | `domain/wishlist_item/service/wishlist_item_service.go` |
| `services/reservation_service.go` | `domain/reservation/service/reservation_service.go` |
| `services/account_cleanup_service.go` | `app/jobs/account_cleanup.go` |
| `services/email_service.go` | `app/jobs/email_service.go` |
| `repositories/user_repository.go` | `domain/user/repository/user_repository.go` |
| `repositories/wishlist_repository.go` | `domain/wishlist/repository/wishlist_repository.go` |
| `repositories/giftitem_repository.go` | `domain/item/repository/giftitem_repository.go` |
| `repositories/wishlistitem_repository.go` | `domain/wishlist_item/repository/wishlistitem_repository.go` |
| `repositories/reservation_repository.go` | `domain/reservation/repository/reservation_repository.go` |
| `repositories/template_repository.go` | `domain/wishlist/repository/template_repository.go` |
| `domains/health/*` | `domain/health/delivery/http/*` |
| `domains/storage/*` | `domain/storage/delivery/http/*` |
