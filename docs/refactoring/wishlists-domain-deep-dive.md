# Wishlists Domain - Deep Dive & Migration Plan

**Domain**: Wishlists
**Complexity**: Medium-High (10 files, complex relationships)
**Priority**: 5th in migration order (after health, storage, reservations, items)
**Estimated Time**: 1.5-2 hours

---

## 📊 Current State Analysis

### Files Inventory (10 files total)

#### Handlers (3 files)
```
handlers/
├── wishlist_handler.go              (534 lines)
│   ├── Endpoints: POST /wishlists, GET /wishlists, GET /wishlists/{id}
│   │              PUT /wishlists/{id}, DELETE /wishlists/{id}
│   │              GET /public/wishlists/{slug}
│   │              GET /public/wishlists/{slug}/gift-items
│   ├── DTOs: CreateWishListRequest, UpdateWishListRequest
│   │        CreateGiftItemRequest (DUPLICATE ❌)
│   │        UpdateGiftItemRequest (DUPLICATE ❌)
│   │        WishListResponse, GiftItemResponse, GetGiftItemsResponse
│   └── Dependencies: services.WishListServiceInterface, auth
│
├── wishlist_item_handler.go         (~304 lines)
│   ├── Endpoints: GET /wishlists/{id}/items
│   │              POST /wishlists/{id}/items (attach)
│   │              POST /wishlists/{id}/items/new (create+attach)
│   │              DELETE /wishlists/{id}/items/{itemId}
│   ├── DTOs: AttachItemRequest, CreateItemRequest (imported from items)
│   └── Dependencies: services.WishlistItemServiceInterface, auth
│
└── test files: wishlist_handler_test.go
```

#### Services (4 files)
```
services/
├── wishlist_service.go                    (~900 lines)
│   ├── Interface: WishListServiceInterface
│   ├── Methods: CreateWishList, GetWishList, UpdateWishList, DeleteWishList
│   │            GetWishListsByOwner, GetWishListByPublicSlug
│   │            GetGiftItemsByWishList
│   └── Dependencies: repositories (wishlist, wishlistitem, giftitem, template)
│
├── wishlist_service_template_methods.go   (~90 lines)
│   ├── Template processing methods
│   └── Helper methods for wishlist service
│
├── wishlist_item_service.go               (~280 lines)
│   ├── Interface: WishlistItemServiceInterface
│   ├── Methods: GetWishlistItems, AttachItem, DetachItem, CreateItemInWishlist
│   └── Dependencies: repositories (wishlist, wishlistitem, giftitem)
│
└── test files: wishlist_service_test.go, wishlist_item_service_test.go
    mocks: mock_wishlist_repository_test.go, mock_wishlistitem_repository_test.go
```

#### Repositories (3 files)
```
repositories/
├── wishlist_repository.go           (~250 lines)
│   ├── Interface: WishlistRepositoryInterface
│   ├── Methods: Create, GetByID, GetByOwner, Update, Delete, GetByPublicSlug
│   └── Database: wishlists table
│
├── wishlistitem_repository.go       (~180 lines)
│   ├── Interface: WishlistItemRepositoryInterface
│   ├── Methods: AttachItem, DetachItem, GetItemsByWishlist, IsItemInWishlist
│   └── Database: wishlist_items junction table
│
└── template_repository.go           (minimal)
    ├── Template management
    └── Database: templates table
```

---

## 🧩 Domain Relationships

### Intra-Domain Dependencies
```
┌─────────────────────────────────────────────────────────────┐
│                    WISHLISTS DOMAIN                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌──────────────────┐         ┌──────────────────┐         │
│  │  WishlistHandler │         │ WishlistItemHandler│        │
│  └────────┬─────────┘         └────────┬──────────┘        │
│           │                            │                     │
│           ↓                            ↓                     │
│  ┌──────────────────┐         ┌──────────────────┐         │
│  │ WishlistService  │←────────│WishlistItemService│         │
│  └────────┬─────────┘         └────────┬──────────┘        │
│           │                            │                     │
│           ↓                            ↓                     │
│  ┌──────────────────┐         ┌──────────────────┐         │
│  │WishlistRepository│         │WishlistItemRepo  │         │
│  └──────────────────┘         └──────────────────┘         │
│           │                            │                     │
│           └──────────┬─────────────────┘                    │
│                      ↓                                       │
│              ┌─────────────────┐                            │
│              │TemplateRepository│                           │
│              └─────────────────┘                            │
└─────────────────────────────────────────────────────────────┘
```

### Cross-Domain Dependencies

```
┌───────────────────┐
│  WISHLISTS Domain │
└─────────┬─────────┘
          │
          ├──→ AUTH Domain (user authentication)
          │    • auth.GetUserFromContext() - Get authenticated user
          │    • Used in all protected endpoints
          │
          ├──→ ITEMS Domain (gift items)
          │    • GiftItemRepository - Read item data
          │    • CreateItemRequest DTO - Import for CreateItemInWishlist
          │    • ItemResponse DTO - Return item data
          │    • NOTE: Many-to-many via wishlist_items junction table
          │
          └──→ SHARED Infrastructure
               • db.Executor - Transaction support
               • validation - Request validation
```

**Key Insight**: Wishlists and Items have a **many-to-many relationship** via the `wishlist_items` junction table. Items are **independent entities** (owned by users), NOT owned by wishlists.

---

## 🔍 DTO Duplication Analysis

### Critical Issue: Duplicate Item DTOs

#### Problem
```go
// ❌ wishlist_handler.go (lines 42-62)
type CreateGiftItemRequest struct {
    Name        string  `json:"name" validate:"required,max=255"`
    Description string  `json:"description"`
    Link        string  `json:"link" validate:"omitempty,url"`
    ImageURL    string  `json:"image_url" validate:"omitempty,url"`
    Price       float64 `json:"price" validate:"omitempty,min=0"`
    Priority    int     `json:"priority" validate:"omitempty,min=0,max=10"`
    Notes       string  `json:"notes"`
    Position    int     `json:"position" validate:"omitempty,min=0"`
}

// ❌ item_handler.go (lines 30-38)
type CreateItemRequest struct {
    Title       string  `json:"title" validate:"required,min=1,max=255"`  // Different field name!
    Description string  `json:"description" validate:"max=2000"`
    Link        string  `json:"link" validate:"omitempty,url"`
    ImageURL    string  `json:"imageUrl" validate:"omitempty,url"`       // Different JSON tag!
    Price       float64 `json:"price" validate:"omitempty,gte=0"`
    Priority    int     `json:"priority" validate:"omitempty,gte=0,lte=5"` // Different max!
    Notes       string  `json:"notes" validate:"max=1000"`
}
```

**Differences**:
1. Field name: `Name` vs `Title`
2. JSON tag: `image_url` vs `imageUrl`
3. Priority validation: `max=10` vs `lte=5`
4. Missing `Position` field in `CreateItemRequest`

**Impact**: API inconsistency, maintenance overhead, confusion

#### Solution
```go
// ✅ domains/items/dtos/requests.go (Canonical version)
type CreateItemRequest struct {
    Title       string  `json:"title" validate:"required,min=1,max=255"`
    Description string  `json:"description" validate:"max=2000"`
    Link        string  `json:"link" validate:"omitempty,url"`
    ImageURL    string  `json:"imageUrl" validate:"omitempty,url"`
    Price       float64 `json:"price" validate:"omitempty,gte=0"`
    Priority    int     `json:"priority" validate:"omitempty,gte=0,lte=5"`
    Notes       string  `json:"notes" validate:"max=1000"`
}

// ✅ domains/wishlists/handlers/wishlist_handler.go
import itemDtos "wish-list/internal/domains/items/dtos"

// Remove CreateGiftItemRequest definition
// Use itemDtos.CreateItemRequest instead
```

**Decision**: Use Items domain DTO as canonical since:
- Items are independent entities
- More recent API design
- Consistent camelCase JSON tags
- Better validation rules

---

## 🗂️ Target Structure

### After Migration

```
domains/wishlists/
├── handlers/
│   ├── wishlist_handler.go          # Wishlist CRUD endpoints
│   ├── wishlist_handler_test.go
│   └── wishlist_item_handler.go     # Wishlist-Item relationships
│
├── services/
│   ├── wishlist_service.go                    # Wishlist business logic
│   ├── wishlist_service_test.go
│   ├── wishlist_service_template_methods.go   # Template helpers
│   ├── wishlist_item_service.go               # Wishlist-Item logic
│   └── wishlist_item_service_test.go
│
├── repositories/
│   ├── wishlist_repository.go                 # Wishlist data access
│   ├── wishlist_repository_test.go
│   ├── wishlistitem_repository.go             # Junction table data access
│   └── template_repository.go                 # Template data access
│
├── dtos/
│   ├── requests.go                # Wishlist-specific requests
│   │   ├── CreateWishListRequest
│   │   ├── UpdateWishListRequest
│   │   └── AttachItemRequest
│   │
│   └── responses.go               # Wishlist-specific responses
│       ├── WishListResponse
│       ├── GetGiftItemsResponse   # Paginated items in wishlist
│       └── (Remove GiftItemResponse - import from items/dtos)
│
├── mocks/                         # Test mocks
│   ├── mock_wishlist_repository_test.go
│   └── mock_wishlistitem_repository_test.go
│
└── wishlists.go                   # Domain export (public API)
```

---

## 📋 Detailed Migration Steps

### Step 1: Create Domain Structure (5 mins)

```bash
# Create wishlists domain folders
mkdir -p backend/internal/domains/wishlists/{handlers,services,repositories,dtos,mocks}

# Verify
tree backend/internal/domains/wishlists -L 1
```

---

### Step 2: Move Handler Files (15 mins)

```bash
# Move handlers
mv backend/internal/handlers/wishlist_handler.go \
   backend/internal/domains/wishlists/handlers/

mv backend/internal/handlers/wishlist_handler_test.go \
   backend/internal/domains/wishlists/handlers/

mv backend/internal/handlers/wishlist_item_handler.go \
   backend/internal/domains/wishlists/handlers/
```

**Update imports in moved handler files:**

```go
// ❌ OLD imports in wishlist_handler.go
import (
    "wish-list/internal/auth"
    "wish-list/internal/services"
)

// ✅ NEW imports
import (
    "wish-list/internal/domains/auth"               // Auth domain
    "wish-list/internal/domains/wishlists/services" // Same domain
    authShared "wish-list/internal/shared/auth"     // Auth middleware (if moved to shared)
)
```

**Package stays the same**: `package handlers`

---

### Step 3: Move Service Files (15 mins)

```bash
# Move services
mv backend/internal/services/wishlist_service.go \
   backend/internal/domains/wishlists/services/

mv backend/internal/services/wishlist_service_test.go \
   backend/internal/domains/wishlists/services/

mv backend/internal/services/wishlist_service_template_methods.go \
   backend/internal/domains/wishlists/services/

mv backend/internal/services/wishlist_item_service.go \
   backend/internal/domains/wishlists/services/

mv backend/internal/services/wishlist_item_service_test.go \
   backend/internal/domains/wishlists/services/
```

**Update imports in service files:**

```go
// ❌ OLD imports in wishlist_service.go
import (
    "wish-list/internal/repositories"
    "wish-list/internal/db"
)

// ✅ NEW imports
import (
    "wish-list/internal/domains/wishlists/repositories"
    "wish-list/internal/domains/items/repositories"  // For GiftItemRepository
    "wish-list/internal/shared/db"
)
```

---

### Step 4: Move Repository Files (15 mins)

```bash
# Move repositories
mv backend/internal/repositories/wishlist_repository.go \
   backend/internal/domains/wishlists/repositories/

mv backend/internal/repositories/wishlist_repository_test.go \
   backend/internal/domains/wishlists/repositories/

mv backend/internal/repositories/wishlistitem_repository.go \
   backend/internal/domains/wishlists/repositories/

mv backend/internal/repositories/template_repository.go \
   backend/internal/domains/wishlists/repositories/
```

**Update imports in repository files:**

```go
// ❌ OLD imports in wishlist_repository.go
import (
    "wish-list/internal/db"
)

// ✅ NEW imports
import (
    "wish-list/internal/shared/db"
)
```

---

### Step 5: Move Test Mocks (5 mins)

```bash
# Move mocks
mv backend/internal/services/mock_wishlist_repository_test.go \
   backend/internal/domains/wishlists/mocks/

mv backend/internal/services/mock_wishlistitem_repository_test.go \
   backend/internal/domains/wishlists/mocks/
```

**Update package in mock files:**

```go
// ❌ OLD package
package services

// ✅ NEW package
package mocks
```

**Update test imports to use mocks:**

```go
// In wishlist_service_test.go
import (
    "wish-list/internal/domains/wishlists/mocks"
    "wish-list/internal/domains/wishlists/services"
)

func TestWishlistService(t *testing.T) {
    mockRepo := &mocks.MockWishlistRepository{}
    // ...
}
```

---

### Step 6: Extract & Consolidate DTOs (30 mins)

#### 6.1 Create DTO Files

```bash
touch backend/internal/domains/wishlists/dtos/requests.go
touch backend/internal/domains/wishlists/dtos/responses.go
```

#### 6.2 Extract Wishlist-Specific DTOs

**`domains/wishlists/dtos/requests.go`**:

```go
package dtos

// CreateWishListRequest represents wishlist creation input
type CreateWishListRequest struct {
    Title        string `json:"title" validate:"required,max=200"`
    Description  string `json:"description"`
    Occasion     string `json:"occasion"`
    OccasionDate string `json:"occasion_date"`
    TemplateID   string `json:"template_id" default:"default"`
    IsPublic     bool   `json:"is_public"`
}

// UpdateWishListRequest represents wishlist update input (partial)
type UpdateWishListRequest struct {
    Title        *string `json:"title" validate:"omitempty,max=200"`
    Description  *string `json:"description"`
    Occasion     *string `json:"occasion"`
    OccasionDate *string `json:"occasion_date"`
    TemplateID   *string `json:"template_id"`
    IsPublic     *bool   `json:"is_public"`
}

// AttachItemRequest represents attaching existing item to wishlist
type AttachItemRequest struct {
    ItemID string `json:"itemId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}
```

**`domains/wishlists/dtos/responses.go`**:

```go
package dtos

// WishListResponse represents wishlist in API responses
type WishListResponse struct {
    ID           string `json:"id" validate:"required"`
    OwnerID      string `json:"owner_id" validate:"required"`
    Title        string `json:"title" validate:"required"`
    Description  string `json:"description"`
    Occasion     string `json:"occasion"`
    OccasionDate string `json:"occasion_date"`
    TemplateID   string `json:"template_id"`
    IsPublic     bool   `json:"is_public"`
    PublicSlug   string `json:"public_slug"`
    ViewCount    string `json:"view_count" validate:"required"`
    ItemCount    int    `json:"item_count" example:"5"`
    CreatedAt    string `json:"created_at" validate:"required"`
    UpdatedAt    string `json:"updated_at" validate:"required"`
}

// GetGiftItemsResponse represents paginated items in wishlist
type GetGiftItemsResponse struct {
    Items []*GiftItemResponse `json:"items" validate:"required"`
    Total int                 `json:"total" validate:"required"`
    Page  int                 `json:"page" validate:"required"`
    Limit int                 `json:"limit" validate:"required"`
    Pages int                 `json:"pages" validate:"required"`
}

// NOTE: GiftItemResponse removed - import from domains/items/dtos instead
```

#### 6.3 Update Handler Imports

**In `wishlist_handler.go`**:

```go
package handlers

import (
    // ... other imports
    "wish-list/internal/domains/wishlists/dtos"
    itemDtos "wish-list/internal/domains/items/dtos"  // For GiftItemResponse
    "wish-list/internal/domains/wishlists/services"
)

// Remove DTO definitions (CreateWishListRequest, UpdateWishListRequest, etc.)
// Use dtos.CreateWishListRequest instead

func (h *WishListHandler) CreateWishList(c echo.Context) error {
    var req dtos.CreateWishListRequest  // ✅ Use DTO package
    // ...
}
```

#### 6.4 Remove Duplicate Item DTOs

**In `wishlist_handler.go`**:

```go
// ❌ DELETE these duplicate definitions (lines 42-62)
// type CreateGiftItemRequest struct { ... }
// type UpdateGiftItemRequest struct { ... }
// type GiftItemResponse struct { ... }

// ✅ Import from items domain instead
import itemDtos "wish-list/internal/domains/items/dtos"

func (h *WishListHandler) toGiftItemResponse(item *services.GiftItemOutput) *itemDtos.ItemResponse {
    // Map to items domain DTO
}
```

---

### Step 7: Create Domain Export File (10 mins)

**`domains/wishlists/wishlists.go`**:

```go
package wishlists

import (
    "wish-list/internal/domains/wishlists/handlers"
    "wish-list/internal/domains/wishlists/repositories"
    "wish-list/internal/domains/wishlists/services"
    "wish-list/internal/shared/db"
)

// NewWishlistHandler creates a fully initialized wishlist handler with all dependencies
func NewWishlistHandler(database *db.DB) *handlers.WishListHandler {
    // Initialize repositories
    wishlistRepo := repositories.NewWishlistRepository(database)
    wishlistItemRepo := repositories.NewWishlistItemRepository(database)
    templateRepo := repositories.NewTemplateRepository(database)

    // Initialize service
    wishlistService := services.NewWishlistService(
        wishlistRepo,
        wishlistItemRepo,
        templateRepo,
    )

    // Return handler
    return handlers.NewWishListHandler(wishlistService)
}

// NewWishlistItemHandler creates wishlist-item relationship handler
func NewWishlistItemHandler(database *db.DB) *handlers.WishlistItemHandler {
    // Initialize repositories
    wishlistRepo := repositories.NewWishlistRepository(database)
    wishlistItemRepo := repositories.NewWishlistItemRepository(database)

    // May need GiftItemRepository from items domain
    // itemRepo := itemsDomain.NewGiftItemRepository(database)

    // Initialize service
    wishlistItemService := services.NewWishlistItemService(
        wishlistRepo,
        wishlistItemRepo,
        // itemRepo,  // If needed
    )

    // Return handler
    return handlers.NewWishlistItemHandler(wishlistItemService)
}

// Export interfaces for external use
type (
    WishlistServiceInterface     = services.WishListServiceInterface
    WishlistItemServiceInterface = services.WishlistItemServiceInterface
    WishlistRepositoryInterface  = repositories.WishlistRepositoryInterface
)
```

---

### Step 8: Update Main Application (15 mins)

**In `cmd/server/main.go`**:

```go
// ❌ OLD imports
import (
    "wish-list/internal/handlers"
    "wish-list/internal/services"
    "wish-list/internal/repositories"
)

// Initialize handlers
wishlistRepo := repositories.NewWishlistRepository(db)
wishlistService := services.NewWishlistService(wishlistRepo)
wishlistHandler := handlers.NewWishlistHandler(wishlistService)

// ✅ NEW imports
import (
    wishlistsDomain "wish-list/internal/domains/wishlists"
    itemsDomain "wish-list/internal/domains/items"
)

// Initialize handlers (much simpler!)
wishlistHandler := wishlistsDomain.NewWishlistHandler(db)
wishlistItemHandler := wishlistsDomain.NewWishlistItemHandler(db)

// Register routes (same as before)
api := e.Group("/api/v1")
wishlists := api.Group("/wishlists", authMiddleware)
wishlists.POST("", wishlistHandler.CreateWishList)
wishlists.GET("", wishlistHandler.GetWishListsByOwner)
wishlists.GET("/:id", wishlistHandler.GetWishList)
wishlists.PUT("/:id", wishlistHandler.UpdateWishList)
wishlists.DELETE("/:id", wishlistHandler.DeleteWishList)

// Wishlist-item relationships
wishlists.GET("/:id/items", wishlistItemHandler.GetWishlistItems)
wishlists.POST("/:id/items", wishlistItemHandler.AttachItemToWishlist)
wishlists.POST("/:id/items/new", wishlistItemHandler.CreateItemInWishlist)
wishlists.DELETE("/:id/items/:itemId", wishlistItemHandler.DetachItemFromWishlist)

// Public routes
public := e.Group("/public")
public.GET("/wishlists/:slug", wishlistHandler.GetWishListByPublicSlug)
public.GET("/wishlists/:slug/gift-items", wishlistHandler.GetGiftItemsByPublicSlug)
```

---

### Step 9: Validation (10 mins)

```bash
# Run domain tests
go test ./internal/domains/wishlists/...

# Run all tests
go test ./...

# Check for import cycles
go list -f '{{.ImportPath}}: {{.Imports}}' ./internal/domains/wishlists/... | grep -i cycle

# Build application
go build ./cmd/server

# Lint
golangci-lint run ./internal/domains/wishlists/...
```

**Validation Checklist**:
- [ ] All tests pass (wishlist_service_test.go, wishlist_handler_test.go)
- [ ] No import cycles
- [ ] Build succeeds
- [ ] Handlers return correct HTTP status codes
- [ ] DTO validation still works
- [ ] Cross-domain imports work (items DTOs)

---

### Step 10: Commit (2 mins)

```bash
git add backend/internal/domains/wishlists/
git add backend/cmd/server/main.go
git commit -m "refactor(wishlists): migrate to domain structure

- Moved 10 files to domains/wishlists/
- Consolidated DTOs to wishlists/dtos/
- Removed duplicate item DTOs (now import from items domain)
- Updated imports and domain exports
- All tests passing"
```

---

## ⚠️ Migration Challenges & Solutions

### Challenge 1: Cross-Domain Item DTOs

**Problem**: Wishlists domain uses Item DTOs (CreateItemRequest, ItemResponse)

**Solution**:
```go
// In wishlist_item_handler.go
import itemDtos "wish-list/internal/domains/items/dtos"

func (h *WishlistItemHandler) CreateItemInWishlist(c echo.Context) error {
    var req itemDtos.CreateItemRequest  // Import from items domain
    // ...
}
```

**Risk**: Import cycle if items domain imports wishlists domain
**Mitigation**: Items domain should NOT import wishlists (one-way dependency)

---

### Challenge 2: GiftItemRepository Dependency

**Problem**: WishlistService and WishlistItemService both use GiftItemRepository

**Current Code**:
```go
// wishlist_service.go
import "wish-list/internal/repositories"

type WishListService struct {
    giftItemRepo repositories.GiftItemRepositoryInterface
}
```

**Solution**: Import from items domain repository

```go
// wishlist_service.go
import (
    "wish-list/internal/domains/wishlists/repositories"
    itemRepos "wish-list/internal/domains/items/repositories"
)

type WishListService struct {
    wishlistRepo     repositories.WishlistRepositoryInterface
    wishlistItemRepo repositories.WishlistItemRepositoryInterface
    giftItemRepo     itemRepos.GiftItemRepositoryInterface  // From items domain
    templateRepo     repositories.TemplateRepositoryInterface
}
```

**Risk**: Import cycle if items service imports wishlist service
**Mitigation**: Use dependency injection at composition root (main.go)

---

### Challenge 3: Service Output DTOs

**Problem**: Service layer has its own DTOs (WishListOutput, GiftItemOutput)

**Current Code**:
```go
// services/wishlist_service.go
type WishListOutput struct {
    ID       string
    OwnerID  string
    Title    string
    // ...
}
```

**Decision**: Keep service DTOs internal to domain

**Rationale**:
- Service DTOs != Handler DTOs (different concerns)
- Service layer uses database models (pgtype.UUID, pgtype.Text)
- Handler layer converts to JSON-friendly types
- Clean architecture: service layer independent of HTTP

**No Change Needed**: Service DTOs stay in service files

---

### Challenge 4: Template Repository

**Problem**: Template repository is used only by wishlists

**Options**:
1. Keep in wishlists/repositories/ (current plan)
2. Move to shared/ if other domains need templates

**Decision**: Keep in wishlists domain for now

**Rationale**:
- Only wishlists use templates currently
- Easy to move to shared/ later if needed
- YAGNI (You Aren't Gonna Need It)

---

## 🧪 Testing Strategy

### Unit Tests to Update

1. **`wishlist_service_test.go`**:
   - Update repository mock imports
   - Update service imports
   - Verify service business logic unchanged

2. **`wishlist_item_service_test.go`**:
   - Update repository mock imports
   - Update service imports
   - Verify wishlist-item logic unchanged

3. **`wishlist_handler_test.go`**:
   - Update handler imports
   - Update DTO imports
   - Update service mock imports
   - Verify HTTP responses unchanged

### Integration Test (Optional)

```go
// domains/wishlists/integration_test.go
func TestWishlistsIntegration(t *testing.T) {
    // Test full domain stack: handler → service → repository → DB
    // Verify cross-domain interaction with items domain
}
```

---

## 📈 Success Metrics

### Quantitative
- [ ] All 10 files moved successfully
- [ ] 3 test files passing (wishlist_service_test, wishlist_item_service_test, wishlist_handler_test)
- [ ] 0 import cycles
- [ ] Build succeeds
- [ ] All endpoints return correct status codes

### Qualitative
- [ ] DTOs consolidated (removed duplicates)
- [ ] Clear separation from items domain
- [ ] Handler DTOs in dtos/ folder
- [ ] Service DTOs kept internal
- [ ] Domain export provides clean API
- [ ] Cross-domain imports are one-way (wishlists → items, NOT items → wishlists)

---

## 🔄 Rollback Plan

If issues arise during migration:

```bash
# Revert to last checkpoint
git log --oneline | head -10  # Find commit before wishlists migration
git revert <commit-hash>

# OR reset to checkpoint
git reset --hard <checkpoint-commit>

# Clean up created folders
rm -rf backend/internal/domains/wishlists
```

---

## 📝 Post-Migration Checklist

- [ ] All wishlists tests passing
- [ ] All cross-domain tests passing (items, reservations)
- [ ] Swagger docs updated
- [ ] main.go simplified
- [ ] Import cycles checked (zero)
- [ ] DTO duplication resolved
- [ ] Documentation updated (CLAUDE.md)
- [ ] Team notified about new structure

---

## 🎓 Key Learnings

### Architectural Insights

1. **Many-to-Many Complexity**: Junction table (wishlist_items) requires careful repository coordination

2. **Cross-Domain DTOs**: Importing DTOs from another domain is acceptable when entities have clear ownership (items are owned by users, not wishlists)

3. **Service vs Handler DTOs**: Keep separate! Service DTOs work with database types, handler DTOs with JSON

4. **Domain Exports**: Simplify dependency injection by providing factory functions at domain boundary

### Best Practices Applied

1. **Dependency Direction**: Wishlists depends on Items, NOT the other way around
2. **Interface Segregation**: WishlistServiceInterface and WishlistItemServiceInterface separate concerns
3. **Repository Pattern**: Clean database abstraction with Executor pattern for transactions
4. **DTO Consolidation**: Remove duplicates, import from owning domain

---

**Estimated Migration Time**: 1.5-2 hours
**Actual Time**: _____ hours
**Issues Encountered**: _________________________________
**Completed By**: _____________________ **Date**: _____

---

**Next Domain**: Auth (most complex, 11 files) - See `auth-domain-deep-dive.md`
