# Domain-Driven Structure Quick Reference

Quick commands and patterns for working with the new domain-driven structure.

---

## 📁 Directory Structure

```
backend/internal/
├── domains/              # Business domains (bounded contexts)
│   ├── auth/            # Authentication & User Identity
│   ├── wishlists/       # Wishlist Management
│   ├── items/           # Gift Items (Independent Resources)
│   ├── reservations/    # Reservations & Purchases
│   ├── storage/         # File Storage (S3)
│   └── health/          # Health Checks
│
└── shared/              # Cross-cutting infrastructure
    ├── middleware/      # CORS, rate limiting, generic middleware
    ├── config/          # App configuration
    ├── db/              # Database connection, models
    ├── cache/           # Redis client
    ├── encryption/      # Encryption service
    ├── validation/      # Generic validators
    ├── analytics/       # Analytics tracking
    └── aws/             # AWS S3 client
```

---

## 🎯 Domain Structure Template

Each domain follows this pattern:

```
domains/[domain-name]/
├── handlers/            # HTTP handlers (presentation layer)
│   ├── [domain]_handler.go
│   └── [domain]_handler_test.go
├── services/            # Business logic (service layer)
│   ├── [domain]_service.go
│   └── [domain]_service_test.go
├── repositories/        # Data access (repository layer)
│   ├── [domain]_repository.go
│   └── [domain]_repository_test.go
├── dtos/                # Data Transfer Objects (API contracts)
│   ├── requests.go      # Request DTOs
│   └── responses.go     # Response DTOs
└── [domain].go          # Domain exports (public API)
```

---

## 🚀 Quick Commands

### Verify Structure
```bash
# List all domains
ls backend/internal/domains/

# Show domain structure
tree backend/internal/domains -L 2

# Show shared infrastructure
tree backend/internal/shared -L 1
```

### Testing
```bash
# Test specific domain
go test ./internal/domains/auth/...
go test ./internal/domains/wishlists/...
go test ./internal/domains/items/...

# Test all domains
go test ./internal/domains/...

# Test with coverage
go test -cover ./internal/domains/...

# Test shared infrastructure
go test ./internal/shared/...

# Test everything
go test ./...
```

### Building
```bash
# Build application
go build ./cmd/server

# Run application
./server

# Build with verbose output
go build -v ./cmd/server
```

### Code Quality
```bash
# Run linter
golangci-lint run

# Format code
go fmt ./...

# Check for unused imports
go mod tidy

# Vet code
go vet ./...
```

### Import Analysis
```bash
# Find all internal imports
grep -r "wish-list/internal" backend/cmd backend/internal

# Check for old import paths (should return nothing)
grep -r "wish-list/internal/handlers\"" backend/
grep -r "wish-list/internal/services\"" backend/
grep -r "wish-list/internal/repositories\"" backend/

# Detect import cycles
go list -f '{{.ImportPath}}: {{.Imports}}' ./internal/domains/... | grep -i cycle
```

### Swagger/OpenAPI
```bash
# Regenerate Swagger docs
swag init -g cmd/server/main.go -d internal/domains

# Verify Swagger output
ls -la docs/swagger/
cat docs/swagger/swagger.json | jq '.paths | keys'
```

---

## 📝 Import Path Patterns

### Old vs New

| Component | Old Path | New Path |
|-----------|----------|----------|
| Auth Handler | `wish-list/internal/handlers` | `wish-list/internal/domains/auth/handlers` |
| Wishlist Service | `wish-list/internal/services` | `wish-list/internal/domains/wishlists/services` |
| Item Repository | `wish-list/internal/repositories` | `wish-list/internal/domains/items/repositories` |
| Middleware | `wish-list/internal/middleware` | `wish-list/internal/shared/middleware` |
| Database | `wish-list/internal/db` | `wish-list/internal/shared/db` |
| Config | `wish-list/internal/config` | `wish-list/internal/shared/config` |

### Domain Import Examples

```go
// Import domain (use alias to avoid conflicts)
import (
    authDomain "wish-list/internal/domains/auth"
    wishlistsDomain "wish-list/internal/domains/wishlists"
    itemsDomain "wish-list/internal/domains/items"
)

// Use domain exports
authHandler := authDomain.NewAuthHandler(...)
wishlistHandler := wishlistsDomain.NewWishlistHandler(...)

// Import specific layer (within domain code)
import (
    "wish-list/internal/domains/items/dtos"
    "wish-list/internal/domains/items/services"
)

// Import shared infrastructure
import (
    "wish-list/internal/shared/middleware"
    "wish-list/internal/shared/db"
    "wish-list/internal/shared/config"
)
```

---

## 🔍 Finding Code

### By Feature (Domain)

| Feature | Location |
|---------|----------|
| User authentication | `domains/auth/` |
| OAuth login | `domains/auth/handlers/oauth_handler.go` |
| User management | `domains/auth/` |
| Wishlist CRUD | `domains/wishlists/` |
| Gift items | `domains/items/` |
| Reservations | `domains/reservations/` |
| File uploads | `domains/storage/` |
| Health checks | `domains/health/` |

### By Layer

| Layer | Pattern | Example |
|-------|---------|---------|
| Handlers | `domains/*/handlers/` | `domains/auth/handlers/auth_handler.go` |
| Services | `domains/*/services/` | `domains/wishlists/services/wishlist_service.go` |
| Repositories | `domains/*/repositories/` | `domains/items/repositories/giftitem_repository.go` |
| DTOs | `domains/*/dtos/` | `domains/items/dtos/requests.go` |
| Tests | `domains/*/*_test.go` | `domains/auth/handlers/auth_handler_test.go` |

### Search Commands

```bash
# Find all handlers
find backend/internal/domains -name "*_handler.go" -type f

# Find all services
find backend/internal/domains -name "*_service.go" -type f

# Find all repositories
find backend/internal/domains -name "*_repository.go" -type f

# Find all DTOs
find backend/internal/domains -path "*/dtos/*.go" -type f

# Find all tests
find backend/internal/domains -name "*_test.go" -type f

# Search for specific function across all domains
grep -r "func CreateWishlist" backend/internal/domains/

# Search for specific type
grep -r "type CreateItemRequest" backend/internal/domains/
```

---

## 🛠️ Common Tasks

### Adding a New Endpoint to Existing Domain

**Example**: Add `DELETE /items/:id` endpoint

```bash
# 1. Add DTO (if needed)
# Edit: domains/items/dtos/requests.go or responses.go

# 2. Add service method
# Edit: domains/items/services/item_service.go
func (s *ItemService) DeleteItem(ctx context.Context, itemID, userID string) error {
    // Business logic
}

# 3. Add handler method
# Edit: domains/items/handlers/item_handler.go
func (h *ItemHandler) DeleteItem(c echo.Context) error {
    // HTTP handling
}

# 4. Register route
# Edit: cmd/server/main.go
items.DELETE("/:id", itemHandler.DeleteItem)

# 5. Add Swagger annotations
# In domains/items/handlers/item_handler.go

# 6. Add tests
# Edit: domains/items/handlers/item_handler_test.go
# Edit: domains/items/services/item_service_test.go

# 7. Run tests
go test ./internal/domains/items/...

# 8. Regenerate Swagger
swag init -g cmd/server/main.go -d internal/domains
```

---

### Adding a New Domain

**Example**: Add `notifications` domain

```bash
# 1. Create domain structure
mkdir -p backend/internal/domains/notifications/{handlers,services,repositories,dtos}

# 2. Create domain files
touch backend/internal/domains/notifications/handlers/notification_handler.go
touch backend/internal/domains/notifications/services/notification_service.go
touch backend/internal/domains/notifications/repositories/notification_repository.go
touch backend/internal/domains/notifications/dtos/requests.go
touch backend/internal/domains/notifications/dtos/responses.go
touch backend/internal/domains/notifications/notifications.go

# 3. Implement domain export
# Edit: domains/notifications/notifications.go
package notifications

import (
    "wish-list/internal/domains/notifications/handlers"
    "wish-list/internal/domains/notifications/services"
    "wish-list/internal/domains/notifications/repositories"
    "wish-list/internal/shared/db"
)

func NewNotificationHandler(db *db.DB) *handlers.NotificationHandler {
    repo := repositories.NewNotificationRepository(db)
    service := services.NewNotificationService(repo)
    return handlers.NewNotificationHandler(service)
}

# 4. Implement layers (handler, service, repository)

# 5. Register in main.go
# Edit: cmd/server/main.go
import notificationsDomain "wish-list/internal/domains/notifications"

notificationHandler := notificationsDomain.NewNotificationHandler(database)
api.GET("/notifications", notificationHandler.List)

# 6. Add tests
touch backend/internal/domains/notifications/handlers/notification_handler_test.go
touch backend/internal/domains/notifications/services/notification_service_test.go

# 7. Run tests
go test ./internal/domains/notifications/...
```

---

### Cross-Domain Communication

**Pattern**: Domain A needs data from Domain B

**❌ Bad**: Direct import
```go
// domains/reservations/services/reservation_service.go
import "wish-list/internal/domains/items/repositories"  // ❌ Direct coupling

func (s *ReservationService) Reserve(itemID string) {
    item := itemRepo.GetByID(itemID)  // ❌ Tight coupling
}
```

**✅ Good**: Interface-based dependency injection
```go
// domains/reservations/services/reservation_service.go
type ItemFetcher interface {  // ✅ Define interface in consumer
    GetByID(ctx context.Context, id string) (*Item, error)
}

type ReservationService struct {
    itemFetcher ItemFetcher  // ✅ Depend on interface
}

// cmd/server/main.go
itemRepo := itemsDomain.NewItemRepository(db)
reservationService := reservationsDomain.NewReservationService(
    itemRepo,  // ✅ Inject at composition root
)
```

---

### DTO Sharing Between Domains

**Pattern**: Multiple domains use the same entity (e.g., Item)

**Option 1**: Cross-domain import (when DTOs are identical)
```go
// domains/wishlists/handlers/wishlist_handler.go
import "wish-list/internal/domains/items/dtos"

func (h *WishlistHandler) AddItem(c echo.Context) error {
    var req dtos.CreateItemRequest  // Use items domain DTO
    // ...
}
```

**Option 2**: Domain-specific DTOs (when semantics differ)
```go
// domains/wishlists/dtos/requests.go
type AddItemToWishlistRequest struct {
    ItemID     string `json:"item_id"`      // Reference existing item
    WishlistID string `json:"wishlist_id"`
}

// vs

// domains/items/dtos/requests.go
type CreateItemRequest struct {
    Title string `json:"title"`  // Create new item
    Price float64 `json:"price"`
}
```

**Rule of Thumb**: Use cross-domain import if DTOs are identical, create domain-specific DTOs if they differ semantically.

---

## 🚨 Common Pitfalls

### Import Cycles
```bash
# ❌ domains/auth imports domains/wishlists
# ❌ domains/wishlists imports domains/auth
# → Import cycle!

# ✅ Fix: Extract shared interface to shared/interfaces/
mkdir -p backend/internal/shared/interfaces
# Move interface to shared, both domains import shared
```

### Shared Business Logic
```bash
# ❌ Don't put domain logic in shared/
# shared/business/item_validator.go  # ❌ Wrong!

# ✅ Domain logic stays in domain
# domains/items/services/item_validator.go  # ✅ Correct!

# ✅ Only infrastructure in shared/
# shared/validation/validator.go  # ✅ Generic validation utility
```

### DTO Duplication
```bash
# ❌ Same DTO in multiple domains
# domains/items/dtos/requests.go: CreateItemRequest
# domains/wishlists/dtos/requests.go: CreateItemRequest

# ✅ Single source of truth
# domains/items/dtos/requests.go: CreateItemRequest (canonical)
# domains/wishlists/ imports items/dtos
```

---

## 📚 Documentation

### Where to Find Things

| Document | Location | Purpose |
|----------|----------|---------|
| **Comprehensive Plan** | `/docs/refactoring/domain-driven-structure-plan.md` | Full rationale, benefits, migration strategy |
| **Migration Checklist** | `/docs/refactoring/migration-checklist.md` | Step-by-step migration tracking |
| **Quick Reference** | `/docs/refactoring/quick-reference.md` | This document |
| **Architecture Guide** | `/docs/Go-Architecture-Guide.md` | 3-layer architecture principles |
| **Project Instructions** | `CLAUDE.md` | Overall project conventions |

### Related Concepts

- **3-Layer Architecture**: Handler → Service → Repository
- **Domain-Driven Design**: Bounded contexts, ubiquitous language
- **Dependency Injection**: Compose dependencies at application root
- **Interface Segregation**: Define interfaces in consumer, not provider

---

## 🔗 Useful Links

### Internal
- Go Architecture Guide: `/docs/Go-Architecture-Guide.md`
- Database Schema: `/docs/plans/00-cross-domain-architecture-plan.md`
- Project Structure: `CLAUDE.md`

### External
- [Domain-Driven Design by Martin Fowler](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Package Oriented Design by Arden Labs](https://www.ardanlabs.com/blog/2017/02/package-oriented-design.html)

---

**Last Updated**: 2026-02-09
**Version**: 1.0
