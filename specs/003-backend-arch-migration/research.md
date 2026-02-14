# Research: Backend Architecture Migration

**Feature**: 003-backend-arch-migration
**Date**: 2026-02-10

## Research Questions & Decisions

### R1: Domain Directory Naming Convention

**Question**: Should the domain directory be `internal/domain/` (singular, as in the example doc) or `internal/domains/` (plural, as in the current codebase)?

**Decision**: Use `internal/domain/` (singular).

**Rationale**: The example architecture document uses singular `domain/`. Go convention favors singular package names (e.g., `net/http`, not `net/https`). The current `internal/domains/` was created during the partial migration attempt and only contains 2 modules.

**Alternatives considered**:
- `internal/domains/` (plural) — matches current code but deviates from Go convention and the target architecture document.

---

### R2: Models Location — Shared vs Per-Domain

**Question**: The current `internal/shared/db/models/models.go` contains all entity structs (User, WishList, GiftItem, Reservation, WishlistItem). After migration, should each domain own its models or should there be a shared models package?

**Decision**: Each domain owns its own model structs. Shared database infrastructure (`DB`, `Executor` interface) stays in `internal/app/database/`.

**Rationale**: Domain isolation requires that each domain is self-contained. If models are shared, changing one domain's data structure risks breaking another. The executor interface and DB connection are genuinely cross-cutting infrastructure, not domain models.

**Alternatives considered**:
- Shared models package in `internal/pkg/models/` — rejected because it creates coupling between domains and violates FR-007 (no cross-domain imports).
- Keep `shared/db/models/` as-is — rejected because it doesn't align with domain-driven design and prevents independent domain evolution.

**Migration approach**: Split `models.go` entity by entity. Types referenced across domains (e.g., `pgtype.UUID`) are standard library types, not custom models, so no duplication issue.

---

### R3: DTO Extraction Strategy

**Question**: Currently, handlers directly serialize service/repository models to JSON. The target architecture requires separate DTOs. How should DTOs be created during migration?

**Decision**: Extract existing JSON struct tags and response-building logic from handlers into DTO structs. Each domain gets `dto/requests.go` and `dto/responses.go` with explicit `ToDomain()` and `FromDomain()` conversion methods.

**Rationale**: This is the approach used in the example architecture document. It creates a clean boundary between the API contract and internal models, satisfying FR-002 and User Story 4.

**Alternatives considered**:
- Keep current approach (no DTOs, handlers serialize models directly) — rejected because it doesn't match the target architecture and makes it impossible to evolve internal models independently of the API.
- Auto-generate DTOs from models — rejected because it doesn't add a meaningful boundary; changes to models would auto-propagate to the API contract.

**Implementation note**: For this migration, DTOs will initially mirror the model fields exactly (to maintain FR-005 — identical API responses). The DTO layer provides a future extension point without requiring it to diverge immediately.

---

### R4: Route Registration Pattern

**Question**: How should each domain register its routes? The current codebase has a `setupRoutes()` function in `main.go` that registers all routes centrally.

**Decision**: Each domain has a `routes.go` file with a `RegisterRoutes(g *echo.Group)` function. The central router calls each domain's registrar.

**Rationale**: This is the standard pattern in Go Echo applications and matches the example architecture. It reduces main.go complexity and allows domains to own their route definitions.

**Pattern**:
```go
// internal/domain/wishlist/delivery/http/routes.go
func RegisterRoutes(g *echo.Group, h *Handler, authMiddleware echo.MiddlewareFunc) {
    wishlists := g.Group("/wishlists")
    wishlists.POST("", h.Create, authMiddleware)
    wishlists.GET("/:id", h.GetByID)
    // ...
}
```

```go
// internal/app/server/router.go
func SetupRoutes(e *echo.Echo, deps Dependencies) {
    api := e.Group("/api/v1")
    wishlistHandler := deps.WishlistHandler
    wishlist.RegisterRoutes(api, wishlistHandler, deps.AuthMiddleware)
    // ... other domains
}
```

**Alternatives considered**:
- Interface-based route registration (each domain implements `Router` interface) — rejected as over-engineering for 8 domains.
- Automatic route discovery via reflection — rejected as too magical and hard to debug.

---

### R5: Background Services Placement

**Question**: Where should `account_cleanup_service.go` and `email_service.go` live? They aren't HTTP-serving domains.

**Decision**: Move to `internal/app/jobs/`. They are application-level infrastructure that coordinates across domains.

**Rationale**: Per the clarification session, background jobs and cross-cutting services reside in the application layer. The account cleanup service needs repositories from multiple domains (user, wishlist, item, reservation), making it a cross-domain coordinator — exactly what the app layer is for.

**Alternatives considered**:
- Create a `jobs` domain — rejected because these services don't serve HTTP endpoints and don't fit the domain module pattern (no handler, no routes).
- Keep in `services/` flat directory — rejected because it maintains the old structure.
- Split into individual domains (cleanup in user domain, email standalone) — rejected because account cleanup spans 4 domains and email is used across multiple domains.

---

### R6: SQL Query Files Location

**Question**: `internal/shared/db/queries/` contains SQL reference files (users.sql, wishlists.sql, etc.). Where should they go?

**Decision**: Distribute to each domain's repository directory as reference documentation. These files are not executed programmatically — they serve as documentation for the SQL embedded in repository Go files.

**Rationale**: Keeping query documentation close to the repository that executes those queries improves discoverability and maintains domain isolation.

**Alternatives considered**:
- Keep in `internal/app/database/queries/` — rejected because queries are domain-specific, not infrastructure.
- Delete them (SQL is already in Go files) — acceptable but loses useful documentation.

---

### R7: Existing Partial Migration (health, storage)

**Question**: `internal/domains/health/` and `internal/domains/storage/` are already partially migrated. What needs to change?

**Decision**: Move from `internal/domains/` to `internal/domain/` (singular) and align the internal structure with the standard domain pattern (delivery/http/, models/).

**Current structure** (health):
```
domains/health/
├── health.go              # Factory
└── handlers/
    ├── health_handler.go
    └── health_handler_test.go
```

**Target structure**:
```
domain/health/
├── delivery/
│   └── http/
│       ├── handler.go
│       ├── handler_test.go
│       └── routes.go
└── models/
    └── health.go
```

**Rationale**: Consistent structure across all domains. The factory function pattern moves into the handler constructor or a domain-level initialization file.

---

### R8: Dockerfile Impact

**Question**: Does the migration require Dockerfile changes?

**Decision**: Minimal changes needed — only update the migration file copy path if migrations move.

**Current Dockerfile** copies:
- Binary from `cmd/server/main.go`
- Migrations from `internal/db/migrations` (at runtime)

**After migration**:
- Binary path unchanged: `cmd/server/main.go`
- Migrations path: `internal/app/database/migrations/` (update COPY instruction)

**Alternatives considered**:
- Change entry point to `cmd/api/main.go` — deferred per spec assumptions. Can be done separately.

---

### R9: Import Path Update Strategy

**Question**: With ~72 Go files changing package paths, what's the safest approach to update imports?

**Decision**: Use `goimports` + manual verification per domain. Migrate one domain at a time, run `go build ./...` after each domain, and fix compilation errors before proceeding.

**Rationale**: Go's compiler catches all import errors at build time. The incremental approach (one domain at a time) limits the blast radius of any single migration step.

**Alternatives considered**:
- Big-bang migration (move everything at once) — rejected because a single compilation error could block all progress and make debugging difficult.
- Automated refactoring tool (gorename, gopls) — useful for individual renames but doesn't handle directory restructuring well.

**Order of operations per file move**:
1. Create target directory
2. Copy file to new location
3. Update package declaration
4. Update internal imports within the file
5. Run `go build ./...` to find all broken imports in OTHER files
6. Fix broken imports
7. Run `go test ./...` to verify
8. Delete old file

## Technology Stack Confirmation

No new technologies introduced. All existing dependencies remain:

| Component | Package | Version | Change |
|-----------|---------|---------|--------|
| Web framework | echo/v4 | v4.15.0 | No change |
| Database driver | pgx/v5 | v5.8.0 | No change |
| SQL mapper | sqlx | v1.4.0 | No change |
| JWT | golang-jwt/v5 | v5.3.1 | No change |
| AWS SDK | aws-sdk-go-v2 | v1.41.1 | No change |
| Redis | go-redis/v9 | v9.17.3 | No change |
| Migrations | golang-migrate/v4 | v4.19.1 | No change |
| Validation | validator/v10 | v10.30.1 | No change |
| Swagger | swaggo/swag | v1.16.6 | No change |
| Testing | testify | latest | No change |
