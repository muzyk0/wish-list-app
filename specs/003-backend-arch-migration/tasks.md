# Tasks: Backend Architecture Migration

**Input**: Design documents from `/specs/003-backend-arch-migration/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/domain-interfaces.md, quickstart.md

**Tests**: This migration restructures existing code. No new tests are required. Existing tests are verified via US6 (100% pass rate with updated imports only).

**Organization**: Tasks are grouped by user story. US2 (app infrastructure) and US3 (shared libraries) are foundational and MUST complete before domain migration. US1/US4/US5 are delivered together per domain. US6 is final verification.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Create target directory structure and verify clean baseline

- [X] T001 Verify clean build baseline: run `go build ./...` and `go test ./...` from `backend/` to confirm green state before any changes
- [X] T002 Create `internal/app/` directory tree: `config/`, `database/migrations/`, `server/`, `middleware/`, `swagger/`, `jobs/`
- [X] T003 [P] Create `internal/pkg/` directory tree: `auth/`, `encryption/`, `aws/`, `cache/`, `validation/`, `analytics/`, `response/`, `helpers/`
- [X] T004 [P] Create `internal/domain/` directory tree for all 8 domains, each with `delivery/http/dto/`, `service/`, `repository/`, `models/` subdirectories

**Checkpoint**: Empty directory structure ready. Build still green.

---

## Phase 2: US2 - Application Infrastructure is Centralized (Priority: P1) 🎯

**Goal**: Extract all application-level infrastructure into `internal/app/` so domains have a stable foundation to depend on

**Independent Test**: `go build ./...` passes after each move; app boots with config, database, and middleware from new locations

### Implementation for US2

- [X] T005 [US2] Move `backend/internal/shared/config/config.go` → `backend/internal/app/config/config.go`, update package declaration and all imports referencing `wish-list/internal/shared/config`
- [X] T006 [US2] Extract DB connection and pool from `backend/internal/shared/db/models/db.go` → `backend/internal/app/database/postgres.go`, update package to `database`
- [X] T007 [US2] Extract `Executor` interface from `backend/internal/shared/db/models/db.go` → `backend/internal/app/database/executor.go`
- [X] T008 [US2] Move SQL migration files from `backend/internal/shared/db/migrations/` → `backend/internal/app/database/migrations/` (file copy, no code changes)
- [X] T009 [US2] Move middleware files from `backend/internal/shared/middleware/` → `backend/internal/app/middleware/`, update package declaration and imports
- [X] T010 [US2] Create `backend/internal/app/server/server.go` — extract Echo server setup from `backend/cmd/server/main.go` (server creation, middleware pipeline, graceful shutdown)
- [X] T011 [US2] Create `backend/internal/app/server/router.go` — extract route registration skeleton from `main.go` (will be filled in during domain migration)
- [X] T012 [US2] Move Swagger initialization to `backend/internal/app/swagger/swagger.go`, update package and imports
- [X] T013 [US2] Move `backend/internal/services/account_cleanup_service.go` → `backend/internal/app/jobs/account_cleanup.go`, update package to `jobs`, update imports
- [X] T014 [US2] Move `backend/internal/services/email_service.go` → `backend/internal/app/jobs/email_service.go`, update package to `jobs`, update imports
- [X] T015 [US2] Create `backend/internal/app/app.go` — application factory struct with `New()`, `Run()`, `Shutdown()` methods, wiring repositories/services/handlers (initial skeleton, refined in Phase 5)
- [X] T016 [US2] Update `backend/cmd/server/main.go` to import from `internal/app/` paths, verify `go build ./...` passes

**Checkpoint**: All app infrastructure in `internal/app/`. Build passes. Old `shared/config`, `shared/db`, `shared/middleware` imports updated.

---

## Phase 3: US3 - Shared Libraries are Reusable Across Domains (Priority: P2)

**Goal**: Extract all reusable utility packages into `internal/pkg/` with zero domain dependencies

**Independent Test**: `go build ./internal/pkg/...` compiles with no imports from `internal/domain/` or `internal/app/`

### Implementation for US3

- [X] T017 [P] [US3] Move `backend/internal/auth/token_manager.go` → `backend/internal/pkg/auth/token_manager.go`, update package to `auth` under pkg path
- [X] T018 [P] [US3] Move `backend/internal/auth/code_store.go` → `backend/internal/pkg/auth/code_store.go`, update package declaration
- [X] T019 [P] [US3] Move `backend/internal/auth/middleware.go` → `backend/internal/pkg/auth/middleware.go`, update package declaration
- [X] T020 [P] [US3] Move remaining `backend/internal/auth/*.go` files (if any) → `backend/internal/pkg/auth/`, update all package declarations
- [X] T021 [P] [US3] Move `backend/internal/shared/encryption/` → `backend/internal/pkg/encryption/`, update package declarations and imports
- [X] T022 [P] [US3] Move `backend/internal/shared/aws/` → `backend/internal/pkg/aws/`, update package declarations and imports
- [X] T023 [P] [US3] Move `backend/internal/shared/cache/` → `backend/internal/pkg/cache/`, update package declarations and imports
- [X] T024 [P] [US3] Move `backend/internal/shared/validation/` → `backend/internal/pkg/validation/`, update package declarations and imports
- [X] T025 [P] [US3] Move `backend/internal/shared/analytics/` → `backend/internal/pkg/analytics/`, update package declarations and imports
- [X] T026 [P] [US3] Move `backend/internal/shared/helpers/` → `backend/internal/pkg/helpers/`, update package declarations and imports
- [X] T027 [US3] Create `backend/internal/pkg/response/response.go` — extract standardized HTTP response helpers (Success, Error, ValidationError) from handler common patterns
- [X] T028 [US3] Update all remaining imports across codebase referencing old `internal/auth/` and `internal/shared/` paths to new `internal/pkg/` paths
- [X] T029 [US3] Verify `go build ./...` passes and `go vet ./...` clean after all pkg/ moves

**Checkpoint**: All shared libraries in `internal/pkg/`. No pkg/ file imports domain/ or app/. Build passes.

---

## Phase 4: US1/US4/US5 - Domain Migration (Priority: P1/P2/P3) 🎯 MVP

**Goal**: Migrate all 8 domains to self-contained modules with models, DTOs, and self-registered routes

**Independent Test**: Each domain compiles independently; `go test ./internal/domain/{name}/...` passes after each domain migration

**Migration order**: Leaf domains first (no cross-domain deps), then core entities, then dependent domains.

### 4A: Health Domain (lightweight, no dependencies)

- [x] T030 [US1] Move `backend/internal/domains/health/handlers/health_handler.go` → `backend/internal/domain/health/delivery/http/handler.go`, update package to `http` (under health delivery)
- [x] T031 [P] [US1] Move `backend/internal/domains/health/handlers/health_handler_test.go` → `backend/internal/domain/health/delivery/http/handler_test.go`, update package and imports
- [x] T032 [US5] Create `backend/internal/domain/health/delivery/http/routes.go` — `RegisterRoutes(g *echo.Group, h *Handler)` function extracting health route definitions
- [x] T032a [US1] Inline `backend/internal/domains/health/health.go` factory function into handler constructor or routes.go, then delete the factory file
- [x] T033 [US1] Create `backend/internal/domain/health/models/health.go` — health check response model if handler uses custom types, otherwise skip
- [x] T034 [US1] Update `backend/internal/app/server/router.go` to call `health.RegisterRoutes()`, verify `go build ./...` passes

### 4B: Storage Domain (lightweight, depends on pkg/aws)

- [x] T035 [US1] Move `backend/internal/domains/storage/handlers/s3_handler.go` → `backend/internal/domain/storage/delivery/http/handler.go`, update package and imports to use `internal/pkg/aws`
- [x] T036 [P] [US1] Move `backend/internal/domains/storage/handlers/s3_handler_test.go` → `backend/internal/domain/storage/delivery/http/handler_test.go`, update package and imports
- [x] T037 [US4] Create `backend/internal/domain/storage/delivery/http/dto/requests.go` and `responses.go` — extract upload request/response types from handler
- [x] T038 [US5] Create `backend/internal/domain/storage/delivery/http/routes.go` — `RegisterRoutes()` function for storage endpoints
- [x] T039 [US1] Create `backend/internal/domain/storage/models/storage.go` — upload metadata model
- [x] T040 [US1] Update router to call `storage.RegisterRoutes()`, verify `go build ./...` passes

### 4C: User Domain (core entity, other domains reference user)

- [x] T041 [US1] Extract `User` struct from `backend/internal/shared/db/models/models.go` → `backend/internal/domain/user/models/user.go`, update package to `models` (under user)
- [x] T042 [US1] Move `backend/internal/repositories/user_repository.go` → `backend/internal/domain/user/repository/user_repository.go`, update package, update model imports to `domain/user/models`, update DB/Executor imports to `app/database`
- [x] T043 [P] [US1] Move `backend/internal/repositories/user_repository_test.go` (if exists) → `backend/internal/domain/user/repository/user_repository_test.go`, update imports
- [x] T044 [US1] Move `backend/internal/services/user_service.go` → `backend/internal/domain/user/service/user_service.go`, update package, update repository/model imports
- [x] T045 [P] [US1] Move `backend/internal/services/user_service_test.go` → `backend/internal/domain/user/service/user_service_test.go`, update imports
- [x] T045a [P] [US1] Move `backend/internal/services/mock_user_repository_test.go` → `backend/internal/domain/user/service/mock_user_repository_test.go`, update package declaration
- [x] T046 [US1] Move `backend/internal/handlers/user_handler.go` → `backend/internal/domain/user/delivery/http/handler.go`, update package, update service/model imports
- [x] T047 [P] [US1] Move `backend/internal/handlers/user_handler_test.go` → `backend/internal/domain/user/delivery/http/handler_test.go`, update imports
- [x] T047a [US1] Convert `backend/internal/handlers/test_helpers_test.go` → `backend/internal/pkg/helpers/testutil.go` (rename from `_test.go` to regular file so it can be imported by domain handler tests), update package to `helpers`
- [x] T048 [US4] Create `backend/internal/domain/user/delivery/http/dto/requests.go` — extract request binding structs from user handler into DTO types with `ToDomain()` methods
- [x] T049 [US4] Create `backend/internal/domain/user/delivery/http/dto/responses.go` — extract response structs from user handler into DTO types with `FromDomain()` methods
- [x] T050 [US5] Create `backend/internal/domain/user/delivery/http/routes.go` — `RegisterRoutes(g *echo.Group, h *Handler, authMiddleware echo.MiddlewareFunc)` with user endpoints
- [x] T051 [US1] Update router to call `user.RegisterRoutes()`, verify `go build ./...` and `go test ./internal/domain/user/...` pass

### 4D: Item Domain (referenced by wishlist, wishlist_item, reservation)

- [x] T052 [US1] Extract `GiftItem` struct from `backend/internal/shared/db/models/models.go` → `backend/internal/domain/item/models/item.go`
- [x] T053 [US1] Move `backend/internal/repositories/giftitem_repository.go` → `backend/internal/domain/item/repository/giftitem_repository.go`, update package, model imports, DB imports
- [x] T054 [P] [US1] Move giftitem repository test (if exists) → `backend/internal/domain/item/repository/giftitem_repository_test.go`, update imports
- [x] T055 [US1] Move `backend/internal/services/item_service.go` → `backend/internal/domain/item/service/item_service.go`, update package and imports
- [x] T056 [P] [US1] Move `backend/internal/services/item_service_test.go` → `backend/internal/domain/item/service/item_service_test.go`, update imports
- [x] T056a [P] [US1] Skip — `giftitem_service_test.go` tests WishListService methods, belongs to Phase 4F (wishlist domain)
- [x] T056b [P] [US1] Move `backend/internal/services/mock_giftitem_repository_test.go` → `backend/internal/domain/item/service/mock_giftitem_repository_test.go`, update package declaration
- [x] T057 [US1] Move `backend/internal/handlers/item_handler.go` → `backend/internal/domain/item/delivery/http/handler.go`, update package and imports
- [x] T058 [P] [US1] Skip — no item handler test file exists (no `item_handler_test.go` in handlers/)
- [x] T059 [US4] Create `backend/internal/domain/item/delivery/http/dto/requests.go` and `responses.go` — extract item DTOs with conversion methods
- [x] T060 [US5] Create `backend/internal/domain/item/delivery/http/routes.go` — `RegisterRoutes()` for item endpoints
- [x] T061 [US1] Update router, verify build and `go test ./internal/domain/item/...` pass

### 4E: WishlistItem Domain (junction table, depends on item + wishlist interfaces)

- [X] T062 [US1] Extract `WishlistItem` struct from `backend/internal/shared/db/models/models.go` → `backend/internal/domain/wishlist_item/models/wishlist_item.go`
- [X] T063 [US1] Move `backend/internal/repositories/wishlistitem_repository.go` → `backend/internal/domain/wishlist_item/repository/wishlistitem_repository.go`, update package, model/DB imports
- [X] T064 [P] [US1] Move wishlistitem repository test (if exists) → `backend/internal/domain/wishlist_item/repository/wishlistitem_repository_test.go`, update imports
- [X] T065 [US1] Move `backend/internal/services/wishlist_item_service.go` → `backend/internal/domain/wishlist_item/service/wishlist_item_service.go`, update package and imports; define `WishListRepositoryInterface` and `GiftItemRepositoryInterface` in service package for cross-domain deps
- [X] T066 [P] [US1] Move `backend/internal/services/wishlist_item_service_test.go` → `backend/internal/domain/wishlist_item/service/wishlist_item_service_test.go`, update imports
- [X] T066a [P] [US1] Regenerate `backend/internal/services/mock_wishlistitem_repository_test.go` → `backend/internal/domain/wishlist_item/service/mock_wishlistitem_repository_test.go` with go:generate annotation in backend/internal/repositories/wishlistitem_repository.go
- [X] T067 [US1] Move `backend/internal/handlers/wishlist_item_handler.go` → `backend/internal/domain/wishlist_item/delivery/http/handler.go`, update package and imports
- [X] T068 [P] [US1] Skip — no wishlist_item handler test file exists (no `wishlist_item_handler_test.go` in handlers/)
- [X] T069 [US4] Create `backend/internal/domain/wishlist_item/delivery/http/dto/requests.go` and `responses.go` — extract DTOs with conversion methods
- [X] T070 [US5] Create `backend/internal/domain/wishlist_item/delivery/http/routes.go` — `RegisterRoutes()` for wishlist_item endpoints
- [X] T071 [US1] Update router, verify build and `go test ./internal/domain/wishlist_item/...` pass

### 4F: Wishlist Domain (depends on item, reservation, cache interfaces)

- [X] T072 [US1] Extract `WishList` struct from `backend/internal/shared/db/models/models.go` → `backend/internal/domain/wishlist/models/wishlist.go`
- [X] T073 [US1] Move `backend/internal/repositories/wishlist_repository.go` → `backend/internal/domain/wishlist/repository/wishlist_repository.go`, update package, model/DB imports
- [X] T074 [P] [US1] Move `backend/internal/repositories/wishlist_repository_test.go` → `backend/internal/domain/wishlist/repository/wishlist_repository_test.go`, update imports
- [X] T075 [US1] Delete `backend/internal/repositories/template_repository.go` — template feature removed per business decision. Remove `TemplateRepositoryInterface` references and `templateRepo` field from `wishlist_service.go`, remove `templateRepo` parameter from `NewWishListService()` constructor, update all callers
- [X] T075a [US1] Delete `backend/internal/services/wishlist_service_template_methods.go` — all template methods (`GetTemplates`, `GetDefaultTemplate`, `UpdateWishListTemplate`) removed along with template repository
- [X] T076 [US1] Move `backend/internal/services/wishlist_service.go` → `backend/internal/domain/wishlist/service/wishlist_service.go`, update package and imports; define `GiftItemRepositoryInterface`, `ReservationRepositoryInterface`, `CacheInterface` in service package for cross-domain deps
- [X] T077 [P] [US1] Move `backend/internal/services/wishlist_service_test.go` → `backend/internal/domain/wishlist/service/wishlist_service_test.go`, update imports (remove template-related test cases if any)
- [X] T077a [P] [US1] Regenerate `backend/internal/services/mock_wishlist_repository_test.go` → `backend/internal/domain/wishlist/service/mock_wishlist_repository_test.go` with go:generate annotation in backend/internal/repositories/wishlist_repository.go
- [X] T078 [US1] Move `backend/internal/handlers/wishlist_handler.go` → `backend/internal/domain/wishlist/delivery/http/handler.go`, update package and imports
- [X] T079 [P] [US1] Move wishlist handler test (if exists) → `backend/internal/domain/wishlist/delivery/http/handler_test.go`, update imports
- [X] T080 [US4] Create `backend/internal/domain/wishlist/delivery/http/dto/requests.go` and `responses.go` — extract wishlist DTOs with conversion methods
- [X] T081 [US5] Create `backend/internal/domain/wishlist/delivery/http/routes.go` — `RegisterRoutes()` for wishlist endpoints
- [X] T082 [US1] Update router, verify build and `go test ./internal/domain/wishlist/...` pass

### 4G: Reservation Domain (depends on item interface)

- [ ] T083 [US1] Extract `Reservation` struct from `backend/internal/shared/db/models/models.go` → `backend/internal/domain/reservation/models/reservation.go`
- [ ] T084 [US1] Move `backend/internal/repositories/reservation_repository.go` → `backend/internal/domain/reservation/repository/reservation_repository.go`, update package, model/DB imports
- [ ] T085 [P] [US1] Move reservation repository test (if exists) → `backend/internal/domain/reservation/repository/reservation_repository_test.go`, update imports
- [ ] T086 [US1] Move `backend/internal/services/reservation_service.go` → `backend/internal/domain/reservation/service/reservation_service.go`, update package and imports; define `GiftItemRepositoryInterface` in service package
- [ ] T087 [P] [US1] Move `backend/internal/services/reservation_service_test.go` → `backend/internal/domain/reservation/service/reservation_service_test.go`, update imports
- [ ] T087a [P] [US1] Regenerate `backend/internal/services/mock_reservation_repository_test.go` → `backend/internal/domain/reservation/service/mock_reservation_repository_test.go` with go:generate annotation in with go:generate annotation in backend/internal/repositories/reservation_repository.go
- [ ] T088 [US1] Move `backend/internal/handlers/reservation_handler.go` → `backend/internal/domain/reservation/delivery/http/handler.go`, update package and imports
- [ ] T089 [P] [US1] Move reservation handler test (if exists) → `backend/internal/domain/reservation/delivery/http/handler_test.go`, update imports
- [ ] T090 [US4] Create `backend/internal/domain/reservation/delivery/http/dto/requests.go` and `responses.go` — extract reservation DTOs with conversion methods
- [ ] T091 [US5] Create `backend/internal/domain/reservation/delivery/http/routes.go` — `RegisterRoutes()` for reservation endpoints
- [ ] T092 [US1] Update router, verify build and `go test ./internal/domain/reservation/...` pass

### 4H: Auth Domain (depends on user service interface, token manager)

- [ ] T093 [US1] Create `backend/internal/domain/auth/models/auth.go` — auth-specific models (token responses, session types) extracted from handler structs
- [ ] T094 [US1] Move `backend/internal/handlers/auth_handler.go` → `backend/internal/domain/auth/delivery/http/handler.go`, update package, import `internal/pkg/auth` for token manager, define `UserServiceInterface` in handler or service package
- [ ] T095 [P] [US1] Skip — no auth handler test file exists (no `auth_handler_test.go` in handlers/)
- [ ] T096 [US1] Move `backend/internal/handlers/oauth_handler.go` → `backend/internal/domain/auth/delivery/http/oauth_handler.go`, update package and imports
- [ ] T097 [P] [US1] Skip — no oauth handler test file exists (no `oauth_handler_test.go` in handlers/)
- [ ] T098 [US4] Create `backend/internal/domain/auth/delivery/http/dto/requests.go` and `responses.go` — extract auth/oauth request and response DTOs with conversion methods
- [ ] T099 [US5] Create `backend/internal/domain/auth/delivery/http/routes.go` — `RegisterRoutes()` for auth and oauth endpoints
- [ ] T100 [US1] Update router, verify build and `go test ./internal/domain/auth/...` pass

**Checkpoint**: All 8 domains migrated. Each domain is self-contained with models, repository, service, handler, DTOs, and routes. Build passes.

---

## Phase 5: US6 - Application Wiring & Test Verification (Priority: P1) 🎯

**Goal**: Wire all domains together in app layer, finalize main.go, verify 100% test pass rate

**Independent Test**: `go test ./...` passes with 100% of pre-existing tests green

### Implementation for US6

- [ ] T101 [US6] Finalize `backend/internal/app/app.go` — complete dependency wiring: instantiate all repositories, services, and handlers per contracts/domain-interfaces.md pseudocode; inject cross-domain interfaces
- [ ] T102 [US6] Finalize `backend/internal/app/server/router.go` — call all 8 domain `RegisterRoutes()` functions with correct Echo groups and auth middleware
- [ ] T103 [US6] Update `backend/cmd/server/main.go` — minimize to: load config → `app.New()` → `app.Run()` with graceful shutdown
- [ ] T104 [US6] Update `backend/cmd/migrate/main.go` (if exists) — update migration path to `internal/app/database/migrations/`
- [ ] T105 [US6] Run `go build ./...` — verify full compilation with zero errors
- [ ] T106 [US6] Run `go test ./...` — verify 100% test pass rate (SC-001)
- [ ] T107 [US6] Run `go vet ./...` — verify no vet warnings

**Checkpoint**: Application builds, all tests pass, no vet warnings. Full functionality preserved.

---

## Phase 6: Cleanup & Deletion

**Purpose**: Remove old directories and files that have been fully migrated

- [ ] T108 Delete `backend/internal/handlers/` directory (all handlers moved to domain/)
- [ ] T109 [P] Delete `backend/internal/services/` directory (all services moved to domain/ or app/jobs/)
- [ ] T110 [P] Delete `backend/internal/repositories/` directory (all repositories moved to domain/)
- [ ] T111 [P] Delete `backend/internal/shared/` directory (split into app/ and pkg/)
- [ ] T112 [P] Delete `backend/internal/auth/` directory (moved to pkg/auth/)
- [ ] T113 [P] Delete `backend/internal/domains/` directory (moved to domain/)
- [ ] T114 Remove entity structs from `backend/internal/shared/db/models/models.go` that were extracted to domain models (User, WishList, GiftItem, Reservation, WishlistItem) and Template — if file is now empty, delete it
- [ ] T115 Run `go build ./...` and `go test ./...` after all deletions to confirm nothing was missed

**Checkpoint**: Old directory structure fully removed. Only new structure remains.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Final quality checks, build infrastructure updates, and documentation

- [ ] T116 [P] Update `backend/Dockerfile` — update migration file COPY path to `internal/app/database/migrations/`, verify Docker build passes
- [ ] T117 [P] Update `backend/Makefile` — update test/build commands if any paths changed
- [ ] T118 Update Swagger: run `swag init` with updated `--dir` and `--parseDependency` flags pointing to new handler locations, verify generated docs are identical
- [ ] T119 [P] Verify import direction rules (SC-006): run static analysis to confirm `pkg/` has zero imports from `domain/` or `app/`
- [ ] T120 [P] Verify cross-domain isolation (SC-002): run static analysis to confirm no `domain/X` package imports `domain/Y` package
- [ ] T121 Verify central router (SC-007): confirm `router.go` contains only `RegisterRoutes()` calls, no domain-specific route definitions
- [ ] T122 Start server and manually verify key endpoints return identical responses (SC-004): `GET /api/v1/health`, auth flow, wishlist CRUD, item CRUD, reservation flow
- [ ] T123 Verify startup time (SC-005): compare server startup time before and after migration, confirm no measurable regression
- [ ] T124 Update `CLAUDE.md` backend structure section to reflect new `internal/app/`, `internal/pkg/`, `internal/domain/` layout
- [ ] T125 Run quickstart.md validation — confirm all example paths in quickstart.md are accurate
- [ ] T126 [P] Update version identifiers per CR-005 — increment patch version in API docs (`@version` annotation) and any version constants to reflect structural refactor

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **US2 - App Infrastructure (Phase 2)**: Depends on Setup — BLOCKS all other phases
- **US3 - Shared Libraries (Phase 3)**: Depends on Phase 2 (pkg/ packages reference app/database for Executor)
- **Domain Migration (Phase 4)**: Depends on Phases 2 AND 3 (domains import from app/ and pkg/)
  - 4A (health) and 4B (storage): Can start as soon as Phase 3 completes
  - 4C (user): Can start after Phase 3
  - 4D (item): Can start after Phase 3
  - 4E (wishlist_item): Can start after 4D (needs item models for interface definition)
  - 4F (wishlist): Can start after 4D and 4G (needs item and reservation interfaces); template repo deleted in T075/T075a before domain move
  - 4G (reservation): Can start after 4D (needs item interface)
  - 4H (auth): Can start after 4C (needs user service interface)
- **US6 - Wiring & Verification (Phase 5)**: Depends on ALL Phase 4 sub-phases
- **Cleanup (Phase 6)**: Depends on Phase 5 (only delete after verification)
- **Polish (Phase 7)**: Depends on Phase 6

### Domain Migration Dependency Graph

```
Phase 3 (pkg/) complete
        │
        ├──→ 4A health ──────────────────────────┐
        ├──→ 4B storage ─────────────────────────┤
        ├──→ 4C user ───────────→ 4H auth ──────┤
        ├──→ 4D item ───┬──→ 4E wishlist_item ──┤
        │               ├──→ 4G reservation ────┤──→ Phase 5
        │               └──→ 4F wishlist ────────┤
        │                    (also needs 4G) ─────┘
```

### Parallel Opportunities

**Within Phase 1**: T002, T003, T004 are parallel (different directories)
**Within Phase 2**: T005-T014 are mostly sequential (import chain dependencies), but T013+T014 (jobs) can parallel with T009 (middleware)
**Within Phase 3**: T017-T026 are ALL parallel (independent pkg/ packages, no cross-deps)
**Within Phase 4**:
  - 4A + 4B can run in parallel
  - 4C + 4D can run in parallel
  - Within each domain: model + repo + test moves are parallel; service depends on model; handler depends on service; DTOs + routes depend on handler
**Within Phase 6**: T108-T113 are ALL parallel (independent directory deletions)
**Within Phase 7**: T116, T117, T119, T120 are parallel

---

## Implementation Strategy

### MVP First (Phases 1-5)

1. Complete Phase 1: Setup directory tree
2. Complete Phase 2: App infrastructure (US2)
3. Complete Phase 3: Shared libraries (US3)
4. Complete Phase 4: All 8 domains (US1 + US4 + US5)
5. Complete Phase 5: Wire together and verify (US6)
6. **STOP and VALIDATE**: `go build ./...` + `go test ./...` + manual endpoint verification
7. Continue to Phase 6-7 only after validation passes

### Incremental Domain Migration

Within Phase 4, migrate one domain at a time:
1. Move files → update imports → verify build → verify domain tests
2. Commit after each domain is fully migrated
3. This ensures the application compiles at every step (research decision R9)

### Rollback Strategy

- Git commit after each completed domain migration
- If a domain migration breaks the build, `git stash` and investigate
- Old and new paths can coexist temporarily during incremental migration

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- US1 (domain structure), US4 (DTOs), US5 (routes) are co-delivered per domain in Phase 4
- This is a structural migration — no new test logic needed, only import path updates
- Commit after each domain or logical group (research decision R9)
- All code must build after each task: `go build ./...` is the primary validation gate
- SC-001: 100% test pass rate is verified in Phase 5 (T106)
- SC-002: Cross-domain isolation verified in Phase 7 (T120)
- SC-006: Import direction verified in Phase 7 (T119)
- SC-007: Router cleanliness verified in Phase 7 (T121)
- Template deletion: T075/T075a remove template_repository.go and wishlist_service_template_methods.go per business decision (not a structural move)
- Shared test helpers: T047a converts handlers/test_helpers_test.go to importable pkg/helpers/testutil.go
- Mock files: Each mock_*_repository_test.go moves with its consuming service's test package
- Total: 135 tasks across 7 phases covering all 6 user stories and 7 success criteria
