# Backend Architecture Violations Analysis

**Analysis Date:** February 1, 2026
**Analyzed Codebase:** `/backend` Go application
**Architecture Standard:** 3-Layer Architecture (Handler-Service-Repository)
**Reference Document:** `/docs/Go-Architecture-Guide.md`

---

## Executive Summary

The codebase violates the core architectural principle: **JSON serialization concerns must stay ONLY in the handler layer**. We found:

- ✅ **Good News**: No HTTP knowledge in services (no `http.` imports)
- ✅ **Good News**: Database models don't have JSON tags
- ❌ **Critical**: Service layer has structs with `json:` tags (4 files)
- ❌ **Critical**: Handlers expose service layer types directly (3+ handlers)
- ⚠️ **Minor**: Some business logic leaked into handlers

---

## Violation Categories

### 🔴 CRITICAL: JSON Tags in Service Layer

**The ONE Non-Negotiable Rule Violated:** JSON serialization concerns exist outside handler layer.

#### Files with Violations:

1. **`backend/internal/services/user_service.go`** (Lines 38-65)
   ```go
   // ❌ WRONG - Service layer has JSON tags
   type RegisterUserInput struct {
       Email     string `json:"email"`
       Password  string `json:"password"`
       FirstName string `json:"first_name"`
       LastName  string `json:"last_name"`
       AvatarUrl string `json:"avatar_url"`
   }

   type LoginUserInput struct {
       Email    string `json:"email"`
       Password string `json:"password"`
   }

   type UpdateUserInput struct {
       Email     *string `json:"email,omitempty"`
       Password  *string `json:"password,omitempty"`
       FirstName *string `json:"first_name,omitempty"`
       LastName  *string `json:"last_name,omitempty"`
       AvatarUrl *string `json:"avatar_url,omitempty"`
   }

   type UserOutput struct {
       ID        string `json:"id"`
       Email     string `json:"email"`
       FirstName string `json:"first_name"`
       LastName  string `json:"last_name"`
       AvatarUrl string `json:"avatar_url"`
   }
   ```

2. **`backend/internal/services/wishlist_service.go`** (Lines 72-156)
   ```go
   // ❌ WRONG - Service layer has JSON tags
   type CreateWishListInput struct {
       Title        string `json:"title"`
       Description  string `json:"description"`
       Occasion     string `json:"occasion"`
       OccasionDate string `json:"occasion_date"`
       TemplateID   string `json:"template_id"`
       IsPublic     bool   `json:"is_public"`
   }

   type WishListOutput struct {
       ID           string `json:"id"`
       OwnerID      string `json:"owner_id"`
       Title        string `json:"title"`
       Description  string `json:"description"`
       Occasion     string `json:"occasion"`
       OccasionDate string `json:"occasion_date"`
       TemplateID   string `json:"template_id"`
       IsPublic     bool   `json:"is_public"`
       PublicSlug   string `json:"public_slug"`
       ViewCount    int64  `json:"view_count"`  // ⚠️ Internal metric exposed!
       CreatedAt    string `json:"created_at"`
       UpdatedAt    string `json:"updated_at"`
   }

   type GiftItemOutput struct {
       ID                string  `json:"id"`
       WishlistID        string  `json:"wishlist_id"`
       Name              string  `json:"name"`
       Description       string  `json:"description"`
       Link              string  `json:"link"`
       ImageURL          string  `json:"image_url"`
       Price             float64 `json:"price"`
       Priority          int     `json:"priority"`
       ReservedByUserID  string  `json:"reserved_by_user_id"`
       ReservedAt        string  `json:"reserved_at"`
       PurchasedByUserID string  `json:"purchased_by_user_id"`
       PurchasedAt       string  `json:"purchased_at"`
       PurchasedPrice    float64 `json:"purchased_price"`
       Notes             string  `json:"notes"`
       Position          int     `json:"position"`
       CreatedAt         string  `json:"created_at"`
       UpdatedAt         string  `json:"updated_at"`
   }

   // Also: CreateGiftItemInput, UpdateGiftItemInput, TemplateOutput
   ```

3. **`backend/internal/services/reservation_service.go`** (Lines 55-68)
   ```go
   // ❌ WRONG - Service layer has JSON tags
   type ReservationOutput struct {
       ID               pgtype.UUID        `json:"id"`
       GiftItemID       pgtype.UUID        `json:"giftItemId"`
       ReservedByUserID pgtype.UUID        `json:"reservedByUserId"`
       GuestName        *string            `json:"guestName"`
       GuestEmail       *string            `json:"guestEmail"`
       ReservationToken pgtype.UUID        `json:"reservationToken"`
       Status           string             `json:"status"`
       ReservedAt       pgtype.Timestamptz `json:"reservedAt"`
       ExpiresAt        pgtype.Timestamptz `json:"expiresAt"`
       CanceledAt       pgtype.Timestamptz `json:"canceledAt"`
       CancelReason     pgtype.Text        `json:"cancelReason"`
       NotificationSent pgtype.Bool        `json:"notificationSent"`
   }
   ```

4. **`backend/internal/repositories/reservation_repository.go`**
   - Repository layer also has JSON tags (double violation!)

**Impact:**
- Services are coupled to API representation
- Can't reuse services for other transports (gRPC, CLI, etc.)
- API changes require service layer modifications
- Violates separation of concerns

---

### 🔴 CRITICAL: Handlers Exposing Service Layer Types Directly

Handlers are returning service layer structs with JSON tags directly to clients, bypassing the DTO mapping layer.

#### 1. **`backend/internal/handlers/user_handler.go`**

**Lines 54, 61:** Response structs reference service types:
```go
// ❌ WRONG - Handler embedding service type
type AuthResponse struct {
    User *services.UserOutput `json:"user" validate:"required"`
    Token string `json:"token" validate:"required"`
}

type ProfileResponse struct {
    User *services.UserOutput `json:"user" validate:"required"`
}
```

**Line 235:** Direct service output return:
```go
// ❌ WRONG - Returning service output directly
func (h *UserHandler) GetProfile(c echo.Context) error {
    // ...
    user, err := h.service.GetUser(ctx, userID)
    // ...
    return c.JSON(http.StatusOK, user)  // Service type exposed!
}
```

**Lines 93-99, 163-166, 277-283:** Mapping request DTOs to service Input types:
```go
// ✅ This part is CORRECT - mapping from handler DTO to service input
user, err := h.service.Register(ctx, services.RegisterUserInput{
    Email:     req.Email,
    Password:  req.Password,
    FirstName: req.FirstName,
    LastName:  req.LastName,
    AvatarUrl: req.AvatarUrl,
})
```

#### 2. **`backend/internal/handlers/wishlist_handler.go`**

**Line 65:** Response struct references service type:
```go
// ❌ WRONG - Handler embedding service type
type GetGiftItemsResponse struct {
    Items []*services.GiftItemOutput `json:"items" validate:"required"`
    Total int                        `json:"total" validate:"required"`
    Page  int                        `json:"page" validate:"required"`
    Limit int                        `json:"limit" validate:"required"`
    Pages int                        `json:"pages" validate:"required"`
}
```

**Lines 84, 139:** Swagger docs reference service types:
```go
// ❌ WRONG - API documentation coupled to service layer
// @Success 201 {object} services.WishListOutput
// @Success 200 {object} services.WishListOutput
```

**Line 129:** Direct service output return:
```go
// ❌ WRONG - Returning service output directly
return c.JSON(http.StatusCreated, wishList)
```

#### 3. **`backend/internal/handlers/reservation_handler.go`**

**Line 121:** Using service type directly:
```go
// ❌ WRONG - Handler using service type
var reservation *services.ReservationOutput
```

**BUT - Lines 164-199: This handler DOES IT RIGHT! ✅**
```go
// ✅ CORRECT - Handler maps service output to handler DTO
response := CreateReservationResponse{
    ID:               reservation.ID.String(),
    GiftItemID:       reservation.GiftItemID.String(),
    ReservedByUserID: nil,
    GuestName:        reservation.GuestName,
    GuestEmail:       reservation.GuestEmail,
    ReservationToken: reservation.ReservationToken.String(),
    Status:           reservation.Status,
    ReservedAt:       reservation.ReservedAt.Time.Format("2006-01-02T15:04:05Z07:00"),
    // ... proper type conversions
}

return c.JSON(http.StatusOK, response)
```

**Impact:**
- Service layer changes break API contracts
- Can't have different API representations for same data
- Can't control which fields are exposed
- Security risk: internal fields like `ViewCount` exposed

---

### ⚠️ MINOR: Business Logic in Handlers

Some business rules exist in handler layer instead of service layer.

#### **`backend/internal/handlers/reservation_handler.go`** (Lines 143-147)

```go
// ⚠️ Business validation in handler
if req.GuestName == nil || req.GuestEmail == nil {
    return c.JSON(http.StatusBadRequest, map[string]string{
        "error": "Guest name and email are required for unauthenticated reservations",
    })
}
```

**Should be:** This business rule belongs in the service layer. The handler should only validate input format.

**Impact:**
- Business logic can't be reused across different transports
- Service layer is incomplete (missing validation)
- Harder to test business rules (requires HTTP mocking)

---

## What's Going RIGHT ✅

### 1. **No HTTP Knowledge in Services**
- ✅ Services don't import `net/http`
- ✅ Services use sentinel errors, not HTTP status codes
- ✅ Services return business errors, handlers map to HTTP codes

### 2. **Database Models Are Clean**
- ✅ Database models only have `db:` tags
- ✅ No JSON tags in `backend/internal/db/models/`

### 3. **Reservation Handler Mapping (Partial)**
- ✅ `reservation_handler.go` lines 164-199 show CORRECT DTO mapping
- ✅ Proper type conversions (pgtype.UUID → string)
- ✅ Date formatting in handler layer

### 4. **Handler Request DTOs**
- ✅ Handlers define their own request DTOs with validation tags
- ✅ Handlers map request DTOs to service input types

---

## Refactoring Plan

### Phase 1: Remove JSON Tags from Service Layer (CRITICAL)

**Priority:** 🔴 **HIGH** - Violates core architecture principle

#### Step 1.1: Create Business Structs (No Tags)

Create new files in services without JSON tags:

**`backend/internal/services/types.go`** (new file)
```go
package services

import "github.com/jackc/pgx/v5/pgtype"

// User domain types - NO JSON tags
type RegisterUserInput struct {
    Email     string
    Password  string
    FirstName string
    LastName  string
    AvatarUrl string
}

type LoginUserInput struct {
    Email    string
    Password string
}

type UpdateUserInput struct {
    Email     *string
    Password  *string
    FirstName *string
    LastName  *string
    AvatarUrl *string
}

type UserOutput struct {
    ID        string
    Email     string
    FirstName string
    LastName  string
    AvatarUrl string
}

// WishList domain types - NO JSON tags
type CreateWishListInput struct {
    Title        string
    Description  string
    Occasion     string
    OccasionDate string
    TemplateID   string
    IsPublic     bool
}

type WishListOutput struct {
    ID           string
    OwnerID      string
    Title        string
    Description  string
    Occasion     string
    OccasionDate string
    TemplateID   string
    IsPublic     bool
    PublicSlug   string
    ViewCount    int64  // Internal metric - handlers should exclude
    CreatedAt    string
    UpdatedAt    string
}

// Reservation domain types - NO JSON tags
type ReservationOutput struct {
    ID               pgtype.UUID
    GiftItemID       pgtype.UUID
    ReservedByUserID pgtype.UUID
    GuestName        *string
    GuestEmail       *string
    ReservationToken pgtype.UUID
    Status           string
    ReservedAt       pgtype.Timestamptz
    ExpiresAt        pgtype.Timestamptz
    CanceledAt       pgtype.Timestamptz
    CancelReason     pgtype.Text
    NotificationSent pgtype.Bool
}
```

#### Step 1.2: Files to Modify

1. **`backend/internal/services/user_service.go`**
   - Remove JSON tags from all structs (lines 38-65)
   - Service logic remains unchanged
   - Tests should still pass (no behavior change)

2. **`backend/internal/services/wishlist_service.go`**
   - Remove JSON tags from all structs (lines 72-156)
   - Service logic remains unchanged

3. **`backend/internal/services/reservation_service.go`**
   - Remove JSON tags from ReservationOutput (lines 55-68)

4. **`backend/internal/repositories/reservation_repository.go`**
   - Remove JSON tags from repository types

**Testing:** Run service tests - they should still pass (no behavior change)
```bash
make test-backend
```

---

### Phase 2: Create Handler DTOs and Mapping Functions (CRITICAL)

**Priority:** 🔴 **HIGH** - Handlers must not expose service types

#### Step 2.1: User Handler Refactoring

**`backend/internal/handlers/user_handler.go`**

**Create Handler-Specific DTOs:**
```go
// Handler-level response DTOs with JSON tags
type UserResponse struct {
    ID        string `json:"id"`
    Email     string `json:"email"`
    FirstName string `json:"first_name"`
    LastName  string `json:"last_name"`
    AvatarUrl string `json:"avatar_url"`
}

type AuthResponse struct {
    User  *UserResponse `json:"user"`
    Token string        `json:"token"`
}
```

**Create Mapping Functions:**
```go
// Private helper: maps service output to handler DTO
func (h *UserHandler) toUserResponse(user *services.UserOutput) *UserResponse {
    if user == nil {
        return nil
    }
    return &UserResponse{
        ID:        user.ID,
        Email:     user.Email,
        FirstName: user.FirstName,
        LastName:  user.LastName,
        AvatarUrl: user.AvatarUrl,
    }
}
```

**Update Handler Methods:**
```go
// Before (WRONG):
return c.JSON(http.StatusOK, user)

// After (CORRECT):
return c.JSON(http.StatusOK, h.toUserResponse(user))
```

**Files to modify:**
- Lines 54, 61: Replace `services.UserOutput` with `UserResponse`
- Line 235: Add `h.toUserResponse(user)` mapping
- Lines 126-131, 189-194: Map service output to handler DTO

#### Step 2.2: WishList Handler Refactoring

**`backend/internal/handlers/wishlist_handler.go`**

**Create Handler-Specific DTOs:**
```go
// Handler-level response DTOs with JSON tags
type WishListResponse struct {
    ID           string `json:"id"`
    OwnerID      string `json:"owner_id"`
    Title        string `json:"title"`
    Description  string `json:"description"`
    Occasion     string `json:"occasion"`
    OccasionDate string `json:"occasion_date"`
    TemplateID   string `json:"template_id"`
    IsPublic     bool   `json:"is_public"`
    PublicSlug   string `json:"public_slug"`
    // ViewCount intentionally EXCLUDED (internal metric)
    CreatedAt    string `json:"created_at"`
    UpdatedAt    string `json:"updated_at"`
}

type GiftItemResponse struct {
    ID          string  `json:"id"`
    WishlistID  string  `json:"wishlist_id"`
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Link        string  `json:"link"`
    ImageURL    string  `json:"image_url"`
    Price       float64 `json:"price"`
    Priority    int     `json:"priority"`
    // Conditionally include reservation fields based on context
    Notes     string `json:"notes"`
    Position  int    `json:"position"`
    CreatedAt string `json:"created_at"`
    UpdatedAt string `json:"updated_at"`
}

type GetGiftItemsResponse struct {
    Items []*GiftItemResponse `json:"items"`
    Total int                 `json:"total"`
    Page  int                 `json:"page"`
    Limit int                 `json:"limit"`
    Pages int                 `json:"pages"`
}
```

**Create Mapping Functions:**
```go
func (h *WishListHandler) toWishListResponse(wl *services.WishListOutput) *WishListResponse {
    return &WishListResponse{
        ID:           wl.ID,
        OwnerID:      wl.OwnerID,
        Title:        wl.Title,
        Description:  wl.Description,
        Occasion:     wl.Occasion,
        OccasionDate: wl.OccasionDate,
        TemplateID:   wl.TemplateID,
        IsPublic:     wl.IsPublic,
        PublicSlug:   wl.PublicSlug,
        // ViewCount intentionally excluded
        CreatedAt:    wl.CreatedAt,
        UpdatedAt:    wl.UpdatedAt,
    }
}

func (h *WishListHandler) toGiftItemResponse(item *services.GiftItemOutput) *GiftItemResponse {
    return &GiftItemResponse{
        ID:          item.ID,
        WishlistID:  item.WishlistID,
        Name:        item.Name,
        Description: item.Description,
        Link:        item.Link,
        ImageURL:    item.ImageURL,
        Price:       item.Price,
        Priority:    item.Priority,
        Notes:       item.Notes,
        Position:    item.Position,
        CreatedAt:   item.CreatedAt,
        UpdatedAt:   item.UpdatedAt,
    }
}
```

**Update Handler Methods:**
```go
// Line 129 - Before (WRONG):
return c.JSON(http.StatusCreated, wishList)

// After (CORRECT):
return c.JSON(http.StatusCreated, h.toWishListResponse(wishList))
```

**Files to modify:**
- Line 65: Replace `services.GiftItemOutput` with `GiftItemResponse`
- Lines 84, 139: Update Swagger annotations to use handler types
- Line 129 and all other return statements: Add mapping

---

### Phase 3: Move Business Logic to Service Layer (MINOR)

**Priority:** 🟡 **MEDIUM** - Improves architecture but not critical

#### Step 3.1: Guest Validation in Service

**`backend/internal/handlers/reservation_handler.go`** (Lines 143-147)

**Before (WRONG):**
```go
// Handler contains business logic
if req.GuestName == nil || req.GuestEmail == nil {
    return c.JSON(http.StatusBadRequest, map[string]string{
        "error": "Guest name and email are required for unauthenticated reservations",
    })
}
```

**After (CORRECT):**

**In Handler:**
```go
// Handler only validates input format
reservation, err := h.service.CreateReservation(ctx, services.CreateReservationInput{
    WishListID: wishListID,
    GiftItemID: giftItemID,
    UserID:     pgtype.UUID{Valid: false},
    GuestName:  req.GuestName,
    GuestEmail: req.GuestEmail,
})

if err != nil {
    if errors.Is(err, services.ErrGuestInfoRequired) {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "Guest name and email are required for unauthenticated reservations",
        })
    }
    // ... other error handling
}
```

**In Service:**
```go
// backend/internal/services/reservation_service.go
var ErrGuestInfoRequired = errors.New("guest name and email are required")

func (s *ReservationService) CreateReservation(ctx context.Context, input CreateReservationInput) (*ReservationOutput, error) {
    // Business validation
    if !input.UserID.Valid {
        if input.GuestName == nil || *input.GuestName == "" || input.GuestEmail == nil || *input.GuestEmail == "" {
            return nil, ErrGuestInfoRequired
        }
    }
    // ... rest of logic
}
```

---

## Implementation Order

### Week 1: Service Layer Cleanup (CRITICAL)
1. ✅ Create `backend/internal/services/types.go` with clean types (no JSON tags)
2. ✅ Remove JSON tags from `user_service.go`
3. ✅ Remove JSON tags from `wishlist_service.go`
4. ✅ Remove JSON tags from `reservation_service.go`
5. ✅ Run tests: `make test-backend` (should pass)

### Week 2: User Handler DTOs (CRITICAL)
1. ✅ Create `UserResponse` DTO in `user_handler.go`
2. ✅ Create `toUserResponse()` mapping function
3. ✅ Update all handler methods to use DTO
4. ✅ Update tests
5. ✅ Run tests: `cd backend && go test ./internal/handlers/...`

### Week 3: WishList Handler DTOs (CRITICAL)
1. ✅ Create `WishListResponse`, `GiftItemResponse` DTOs
2. ✅ Create mapping functions
3. ✅ Update all handler methods
4. ✅ Update Swagger annotations
5. ✅ Run tests

### Week 4: Validation & Documentation (MEDIUM)
1. ✅ Move business validation from handlers to services
2. ✅ Update API documentation (remove service type references)
3. ✅ Add architecture compliance tests
4. ✅ Update `/docs/Go-Architecture-Guide.md` with code examples from project

---

## Validation Checklist

After refactoring, verify compliance:

- [ ] **No JSON tags in services**: `grep -r 'json:' backend/internal/services/` returns nothing
- [ ] **No JSON tags in repositories**: `grep -r 'json:' backend/internal/repositories/` returns nothing
- [ ] **No service types in handlers**: `grep -r 'services\.\w*Output' backend/internal/handlers/*.go` shows only internal usage, not in responses
- [ ] **All handlers have DTOs**: Every handler file has `Response` structs with JSON tags
- [ ] **All handlers map outputs**: Search for direct service returns: `grep -r 'c\.JSON.*service\.' backend/internal/handlers/`
- [ ] **No HTTP in services**: `grep -r 'net/http' backend/internal/services/` returns nothing
- [ ] **Tests pass**: `make test-backend` returns success

---

## Architecture Compliance Test

Create automated test to prevent future violations:

**`backend/internal/handlers/architecture_test.go`** (new file)
```go
package handlers_test

import (
    "os"
    "path/filepath"
    "strings"
    "testing"
)

func TestNoServiceTypesInHandlerResponses(t *testing.T) {
    handlerFiles, _ := filepath.Glob("*.go")

    for _, file := range handlerFiles {
        if strings.HasSuffix(file, "_test.go") {
            continue
        }

        content, _ := os.ReadFile(file)

        // Check for service types in handler responses
        violations := []string{
            "services.UserOutput",
            "services.WishListOutput",
            "services.GiftItemOutput",
            "services.ReservationOutput",
        }

        for _, violation := range violations {
            if strings.Contains(string(content), violation) {
                t.Errorf("Handler %s contains service type %s - handlers must define their own DTOs", file, violation)
            }
        }
    }
}
```

---

## Summary of Changes

| Component | Current State | Target State | Impact |
|-----------|--------------|--------------|--------|
| **Service Types** | Have `json:` tags | No `json:` tags | Services reusable, transport-agnostic |
| **Handler DTOs** | Use service types | Define own DTOs | API decoupled from business logic |
| **Handler Mapping** | Mixed (some map, some don't) | All map service → DTO | Consistent, secure, controllable |
| **Business Logic** | Some in handlers | All in services | Testable, reusable |
| **API Documentation** | References service types | References handler DTOs | Accurate, stable |

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Breaking API changes | High | High | Version API, deprecate old endpoints |
| Test failures | Medium | Medium | Update tests alongside refactoring |
| Missing field mappings | Medium | High | Code review, integration tests |
| Performance regression | Low | Low | Mapping is negligible overhead |

---

## Success Metrics

- ✅ Zero JSON tags in service layer
- ✅ Zero service types in handler response signatures
- ✅ 100% handler methods use DTO mapping
- ✅ All tests passing
- ✅ API documentation references only handler types
- ✅ Architecture compliance test passes

---

## References

- [Go Architecture Guide](/docs/Go-Architecture-Guide.md)
- [Backend Best Practices](/CLAUDE.md#backend-best-practices--patterns)
- [Conventional Commits](/CLAUDE.md#conventional-commits)
