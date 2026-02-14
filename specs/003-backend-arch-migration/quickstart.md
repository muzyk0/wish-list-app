# Quickstart: Backend Architecture Migration

**Feature**: 003-backend-arch-migration
**Date**: 2026-02-10

## What Changed

The backend was reorganized from a flat 3-layer structure to a domain-driven architecture:

```
BEFORE                              AFTER
internal/                           internal/
├── handlers/     (all handlers)    ├── app/        (infrastructure)
├── services/     (all services)    ├── pkg/        (shared libraries)
├── repositories/ (all repos)       └── domain/     (business modules)
├── shared/       (everything else)     ├── auth/
├── auth/         (JWT)                 ├── user/
└── domains/      (partial)             ├── wishlist/
                                        ├── item/
                                        ├── wishlist_item/
                                        ├── reservation/
                                        ├── health/
                                        └── storage/
```

## Finding Code

### "Where is the wishlist handler?"

```
internal/domain/wishlist/delivery/http/handler.go
```

### "Where are the user service tests?"

```
internal/domain/user/service/user_service_test.go
```

### "Where is the database connection setup?"

```
internal/app/database/postgres.go
```

### "Where is the JWT middleware?"

```
internal/pkg/auth/middleware.go
```

### "Where is the account cleanup background job?"

```
internal/app/jobs/account_cleanup.go
```

## Domain Module Structure

Every business domain follows this pattern:

```
domain/{name}/
├── delivery/http/
│   ├── handler.go       → HTTP request handling
│   ├── dto/
│   │   ├── requests.go  → Input validation structs
│   │   └── responses.go → API response structs
│   └── routes.go        → Route registration
├── service/
│   ├── {name}_service.go      → Business logic
│   └── {name}_service_test.go → Unit tests
├── repository/
│   ├── {name}_repository.go      → Database access
│   └── {name}_repository_test.go → DB mock tests
└── models/
    └── {name}.go        → Domain entity structs
```

## Three Rules

1. **Domains don't import other domains.** Cross-domain communication is wired at the app layer via interface injection.
2. **`pkg/` doesn't import `domain/` or `app/`.** Shared libraries are dependency-free.
3. **`app/` wires everything together.** It imports all domains and shared packages to initialize the application.

## Import Direction

```
app/ ──imports──→ domain/
app/ ──imports──→ pkg/
domain/ ──imports──→ pkg/
domain/ ──imports──→ app/database (Executor interface only)
pkg/ ──imports──→ (external deps only)
```

## Running

Nothing changed for running, testing, or deploying:

```bash
# Start server
make backend

# Run all tests
make test-backend

# Run single domain tests
go test ./backend/internal/domain/wishlist/...

# Build
make build-backend
```

## Adding a New Domain

1. Create directory: `internal/domain/{name}/delivery/http/dto/`, `service/`, `repository/`, `models/`
2. Write models in `models/{name}.go`
3. Write repository in `repository/{name}_repository.go`
4. Write service in `service/{name}_service.go`
5. Write DTOs in `delivery/http/dto/`
6. Write handler in `delivery/http/handler.go`
7. Write route registration in `delivery/http/routes.go`
8. Register domain routes in `internal/app/server/router.go`
9. Wire dependencies in `internal/app/app.go`
