# Implementation Plan: Backend Architecture Migration

**Branch**: `003-backend-arch-migration` | **Date**: 2026-02-10 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/003-backend-arch-migration/spec.md`

## Summary

Migrate the Go backend from a flat 3-layer structure (`handlers/`, `services/`, `repositories/`) to a domain-driven architecture with three top-level zones: `internal/app/` (infrastructure), `internal/pkg/` (shared libraries), and `internal/domain/` (business domains). The migration is structural only — no behavior changes, no new features, no schema changes. All 72 Go source files (~26K LOC) are reorganized while preserving identical API behavior and 100% test pass rate.

## Technical Context

**Language/Version**: Go 1.25.5
**Primary Dependencies**: Echo v4.15.0, sqlx v1.4.0, pgx/v5 v5.8.0, golang-jwt/v5 v5.3.1, AWS SDK v2
**Storage**: PostgreSQL (via pgx/sqlx), Redis (caching), AWS S3 (file uploads), AWS KMS (encryption)
**Testing**: testify (assertions), sqlmock (DB mocks), manual mock structs
**Target Platform**: Linux server (Docker/Alpine), deployed to Render
**Project Type**: Web application (backend API only — frontend and mobile are separate)
**Module Path**: `wish-list`
**Performance Goals**: No regression — startup time and request latency must remain identical
**Constraints**: Zero downtime during migration (incremental domain-by-domain approach)
**Scale/Scope**: 72 Go files, 8 domain modules, ~26K LOC to reorganize

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| Code Quality | PASS | Migration improves maintainability via domain isolation |
| Test-First Approach | PASS | All existing tests preserved; import paths updated per domain |
| API Contract Integrity | PASS | Zero API changes — FR-005 guarantees identical endpoints |
| Data Privacy Protection | PASS | Encryption service moves to `internal/pkg/encryption/` — no behavior change |
| Semantic Versioning | PASS | This is a refactor (no API changes) — patch version increment only |
| Specification Checkpoints | PASS | Spec completed and clarified before this plan |

**Gate Result**: PASS — no violations to justify.

## Project Structure

### Documentation (this feature)

```text
specs/003-backend-arch-migration/
├── plan.md              # This file
├── research.md          # Phase 0: migration decisions and rationale
├── data-model.md        # Phase 1: package/module structure mapping
├── quickstart.md        # Phase 1: developer onboarding guide
├── contracts/           # Phase 1: domain module interface contracts
│   └── domain-interfaces.md
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code — Target Structure

```text
backend/
├── cmd/
│   ├── server/main.go                    # Minimal entry point (delegates to app)
│   └── migrate/main.go                   # Migration CLI (unchanged)
│
├── internal/
│   ├── app/                              # APPLICATION INFRASTRUCTURE
│   │   ├── app.go                        # Application factory & lifecycle
│   │   ├── config/
│   │   │   └── config.go                 # Configuration loading (from shared/config/)
│   │   ├── database/
│   │   │   ├── postgres.go               # DB connection & pooling (from shared/db/models/db.go)
│   │   │   ├── executor.go               # Executor interface (from shared/db/models/)
│   │   │   └── migrations/               # SQL migration files (from shared/db/migrations/)
│   │   ├── server/
│   │   │   ├── server.go                 # Echo server setup
│   │   │   └── router.go                 # Central router (calls domain route registrars)
│   │   ├── middleware/
│   │   │   ├── middleware.go             # Middleware pipeline (from shared/middleware/)
│   │   │   ├── cors.go
│   │   │   └── rate_limit.go
│   │   ├── swagger/
│   │   │   └── swagger.go               # Swagger initialization
│   │   └── jobs/
│   │       ├── account_cleanup.go        # Background cleanup (from services/account_cleanup_service.go)
│   │       └── email_service.go          # Email sending (from services/email_service.go)
│   │
│   ├── pkg/                              # SHARED LIBRARIES (domain-agnostic)
│   │   ├── auth/
│   │   │   ├── token_manager.go          # JWT operations (from auth/)
│   │   │   ├── code_store.go             # Mobile handoff codes (from auth/)
│   │   │   └── middleware.go             # JWT middleware (from auth/)
│   │   ├── encryption/
│   │   │   ├── service.go               # PII encryption (from shared/encryption/)
│   │   │   └── kms.go                   # AWS KMS integration
│   │   ├── aws/
│   │   │   └── s3.go                    # S3 client (from shared/aws/)
│   │   ├── cache/
│   │   │   └── redis.go                 # Redis cache (from shared/cache/)
│   │   ├── validation/
│   │   │   └── validator.go             # Input validation (from shared/validation/)
│   │   ├── analytics/
│   │   │   └── analytics.go            # Event tracking (from shared/analytics/)
│   │   ├── response/
│   │   │   └── response.go             # Standardized HTTP responses (NEW)
│   │   └── helpers/
│   │       └── helpers.go              # Test/utility helpers (from shared/helpers/)
│   │
│   └── domain/                           # BUSINESS DOMAINS
│       ├── auth/
│       │   ├── delivery/
│       │   │   └── http/
│       │   │       ├── handler.go        # Auth endpoints (from handlers/auth_handler.go)
│       │   │       ├── oauth_handler.go  # OAuth endpoints (from handlers/oauth_handler.go)
│       │   │       ├── dto/
│       │   │       │   ├── requests.go
│       │   │       │   └── responses.go
│       │   │       └── routes.go         # Route registration (NEW)
│       │   └── models/
│       │       └── auth.go               # Auth-specific models (tokens, sessions)
│       │
│       ├── user/
│       │   ├── delivery/
│       │   │   └── http/
│       │   │       ├── handler.go        # User endpoints (from handlers/user_handler.go)
│       │   │       ├── dto/
│       │   │       │   ├── requests.go
│       │   │       │   └── responses.go
│       │   │       └── routes.go
│       │   ├── service/
│       │   │   ├── user_service.go       # (from services/user_service.go)
│       │   │   └── user_service_test.go
│       │   ├── repository/
│       │   │   ├── user_repository.go    # (from repositories/user_repository.go)
│       │   │   └── user_repository_test.go
│       │   └── models/
│       │       └── user.go               # User entity (from shared/db/models/models.go)
│       │
│       ├── wishlist/
│       │   ├── delivery/
│       │   │   └── http/
│       │   │       ├── handler.go        # (from handlers/wishlist_handler.go)
│       │   │       ├── dto/
│       │   │       │   ├── requests.go
│       │   │       │   └── responses.go
│       │   │       └── routes.go
│       │   ├── service/
│       │   │   ├── wishlist_service.go   # (from services/wishlist_service.go)
│       │   │   ├── wishlist_service_test.go
│       │   │   └── mock_wishlist_repository_test.go
│       │   ├── repository/
│       │   │   ├── wishlist_repository.go # (from repositories/wishlist_repository.go)
│       │   │   └── wishlist_repository_test.go
│       │   └── models/
│       │       └── wishlist.go           # (from shared/db/models/models.go)
│       │
│       ├── item/
│       │   ├── delivery/
│       │   │   └── http/
│       │   │       ├── handler.go        # (from handlers/item_handler.go)
│       │   │       ├── dto/
│       │   │       │   ├── requests.go
│       │   │       │   └── responses.go
│       │   │       └── routes.go
│       │   ├── service/
│       │   │   ├── item_service.go       # (from services/item_service.go)
│       │   │   └── item_service_test.go
│       │   ├── repository/
│       │   │   ├── giftitem_repository.go # (from repositories/giftitem_repository.go)
│       │   │   └── giftitem_repository_test.go
│       │   └── models/
│       │       └── item.go               # (from shared/db/models/models.go)
│       │
│       ├── wishlist_item/
│       │   ├── delivery/
│       │   │   └── http/
│       │   │       ├── handler.go        # (from handlers/wishlist_item_handler.go)
│       │   │       ├── dto/
│       │   │       │   ├── requests.go
│       │   │       │   └── responses.go
│       │   │       └── routes.go
│       │   ├── service/
│       │   │   ├── wishlist_item_service.go # (from services/wishlist_item_service.go)
│       │   │   └── wishlist_item_service_test.go
│       │   ├── repository/
│       │   │   ├── wishlistitem_repository.go # (from repositories/wishlistitem_repository.go)
│       │   │   └── wishlistitem_repository_test.go
│       │   └── models/
│       │       └── wishlist_item.go      # Junction table model
│       │
│       ├── reservation/
│       │   ├── delivery/
│       │   │   └── http/
│       │   │       ├── handler.go        # (from handlers/reservation_handler.go)
│       │   │       ├── dto/
│       │   │       │   ├── requests.go
│       │   │       │   └── responses.go
│       │   │       └── routes.go
│       │   ├── service/
│       │   │   ├── reservation_service.go # (from services/reservation_service.go)
│       │   │   └── reservation_service_test.go
│       │   ├── repository/
│       │   │   ├── reservation_repository.go # (from repositories/reservation_repository.go)
│       │   │   └── reservation_repository_test.go
│       │   └── models/
│       │       └── reservation.go        # (from shared/db/models/models.go)
│       │
│       ├── health/
│       │   ├── delivery/
│       │   │   └── http/
│       │   │       ├── handler.go        # (from domains/health/handlers/)
│       │   │       └── routes.go
│       │   └── models/
│       │       └── health.go
│       │
│       └── storage/
│           ├── delivery/
│           │   └── http/
│           │       ├── handler.go        # (from domains/storage/handlers/)
│           │       ├── dto/
│           │       │   ├── requests.go
│           │       │   └── responses.go
│           │       └── routes.go
│           └── models/
│               └── storage.go
│
├── docs/                                  # Generated Swagger documentation (unchanged)
├── library-docs/                          # Reference documentation (unchanged)
├── go.mod                                 # Module: wish-list (unchanged)
├── Dockerfile                             # Updated build paths
└── Makefile                               # Updated test/build commands
```

**Structure Decision**: Domain-driven layout under `internal/` with three zones: `app/` (infrastructure), `pkg/` (shared libraries), `domain/` (business modules). Each domain is fully self-contained with delivery/service/repository/models layers. This structure matches the target architecture from the example document while preserving all existing patterns (executor pattern, sentinel errors, constructor DI).

## Migration Phases

### Phase A: Foundation (internal/app/ + internal/pkg/)

Create application infrastructure and shared library packages first, since all domains depend on them.

**Step A1**: Create `internal/app/` structure
- Move `shared/config/` → `app/config/`
- Move `shared/db/models/db.go` + executor → `app/database/`
- Move `shared/db/migrations/` → `app/database/migrations/`
- Move `shared/middleware/` → `app/middleware/`
- Create `app/server/` (server.go, router.go) — extract from main.go
- Move `services/account_cleanup_service.go` → `app/jobs/account_cleanup.go`
- Move `services/email_service.go` → `app/jobs/email_service.go`
- Create `app/app.go` — application factory

**Step A2**: Create `internal/pkg/` structure
- Move `auth/` → `pkg/auth/` (token_manager, code_store, middleware)
- Move `shared/encryption/` → `pkg/encryption/`
- Move `shared/aws/` → `pkg/aws/`
- Move `shared/cache/` → `pkg/cache/`
- Move `shared/validation/` → `pkg/validation/`
- Move `shared/analytics/` → `pkg/analytics/`
- Move `shared/helpers/` → `pkg/helpers/`
- Create `pkg/response/` — extract standardized response helpers

**Step A3**: Update all import paths across the codebase
- Update `cmd/server/main.go` to use new `app/` and `pkg/` paths
- Verify compilation: `go build ./...`

### Phase B: Domain Migration (one domain at a time)

Each domain follows the same pattern: create directory → move files → extract DTOs → create routes.go → update imports → verify tests.

**Migration order** (leaf domains first, then domains with cross-domain dependencies):
1. **health** — already partially migrated, no dependencies
2. **storage** — already partially migrated, depends on pkg/aws only
3. **user** — core entity, other domains reference user
4. **item** — referenced by wishlist and reservation
5. **wishlist_item** — junction table, depends on item and wishlist interfaces
6. **wishlist** — depends on item, reservation interfaces + cache
7. **reservation** — depends on item interface
8. **auth** — depends on user service interface, token manager

**Per-domain migration steps**:
1. Create directory tree: `domain/{name}/delivery/http/dto/`, `service/`, `repository/`, `models/`
2. Move handler → `delivery/http/handler.go`
3. Move service → `service/{name}_service.go`
4. Move repository → `repository/{name}_repository.go`
5. Move tests alongside their source files
6. Extract model structs from `shared/db/models/models.go` → `models/{name}.go`
7. Extract DTOs from handler into `delivery/http/dto/requests.go` and `responses.go`
8. Create `delivery/http/routes.go` — register domain routes
9. Update all import paths
10. Run `go test ./internal/domain/{name}/...` — verify tests pass

### Phase C: Application Wiring & Cleanup

**Step C1**: Update `cmd/server/main.go`
- Delegate initialization to `app.New()` and `app.Run()`
- Keep main.go minimal (load config, create app, run)

**Step C2**: Update router
- Create `app/server/router.go` that calls each domain's `RegisterRoutes()`
- Remove all domain-specific route definitions from central location

**Step C3**: Update Swagger
- Update `swag init` source path for new handler locations
- Regenerate Swagger docs, verify identical output

**Step C4**: Update Dockerfile
- Update build paths if any entry point changes
- Update migration file copy path
- Verify Docker build passes

**Step C5**: Delete old directories
- Remove `internal/handlers/` (empty after migration)
- Remove `internal/services/` (empty after migration)
- Remove `internal/repositories/` (empty after migration)
- Remove `internal/shared/` (empty after split to app/ + pkg/)
- Remove `internal/auth/` (moved to pkg/auth/)
- Remove `internal/domains/` (moved to domain/)

**Step C6**: Final verification
- `go build ./...` — compilation passes
- `go test ./...` — all tests pass
- `go vet ./...` — no warnings
- Start server, hit all endpoints — identical behavior

## Cross-Domain Dependency Resolution

Based on the clarification: **interface injection at startup**.

| Domain | Needs From | Resolution |
|--------|-----------|------------|
| wishlist | item repo, reservation repo, cache | Interfaces injected via service constructor |
| wishlist_item | wishlist repo, item repo | Interfaces injected via service constructor |
| reservation | item repo | Interface injected via service constructor |
| auth | user service | Interface injected via handler constructor |
| app/jobs/cleanup | user repo, wishlist repo, item repo, reservation repo | Interfaces injected at app startup |

Each domain defines its own interface for what it needs. The app layer imports all domains and wires them together.

## Complexity Tracking

No constitution violations. No complexity justifications needed.
