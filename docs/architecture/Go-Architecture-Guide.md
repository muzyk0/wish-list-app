# Go Backend Architecture Guide

A practical guide to building maintainable Go web services using a simplified 3-layer architecture.

---

## Table of Contents

1. [Overview](#overview)
2. [Layer Responsibilities](#layer-responsibilities)
3. [Code Examples](#code-examples)
4. [Project Structure](#project-structure)
5. [Data Flow](#data-flow)
6. [Validation Strategy](#validation-strategy)
7. [Testing Approach](#testing-approach)
8. [Security Considerations](#security-considerations)
9. [When to Evolve](#when-to-evolve)
10. [FAQ](#faq)

---

## Overview

### Architecture Philosophy

This guide describes a **simple, pragmatic 3-layer Go architecture** designed for medium-sized web services. The philosophy is "less abstraction when possible" — we avoid mandatory domain/entity models unless complexity demands them.

### Core Principles

1. **Handler-Service-Repository Pattern**: Three clear layers with distinct responsibilities
2. **Minimal Abstraction**: Database models can flow through layers without intermediate domain objects
3. **Pragmatic Trade-offs**: Simplicity over theoretical purity for CRUD-heavy applications

### The ONE Non-Negotiable Rule

> **JSON serialization concerns stay ONLY in the handler layer.**

This is the single most important architectural constraint. Database models may flow through your application, but handlers MUST map them to API-specific DTOs before returning responses. This rule ensures:

- Services remain reusable across different transports (REST, gRPC, CLI)
- API changes don't ripple through business logic
- Sensitive data is never accidentally exposed

---

## Layer Responsibilities

### Handler Layer (Transport)

The handler layer is the **ONLY** layer that knows about HTTP, JSON, and API contracts.

**Responsibilities:**
- Define Request/Response DTOs with `json:` tags
- Perform input validation using struct tags (`validate:"required"`)
- Map between DTOs and business/data structures
- Return appropriate HTTP status codes
- Handle authentication/authorization context extraction

**What handlers should NEVER do:**
- Contain business logic
- Directly access the database
- Expose database models to clients

```go
package handler

// Handler owns all JSON serialization concerns
type WishListHandler struct {
    service *service.WishListService
}

// Request DTO - only place for json: and validate: tags together
type CreateWishListRequest struct {
    Title       string `json:"title" validate:"required,min=3,max=100"`
    Description string `json:"description,omitempty" validate:"max=500"`
    IsPublic    bool   `json:"is_public"`
}

// Response DTO - explicit control over API contract
type WishListResponse struct {
    ID          string `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description,omitempty"`
    IsPublic    bool   `json:"is_public"`
    CreatedAt   string `json:"created_at"`
}
```

### Service Layer (Business Logic)

The service layer contains **ALL** business rules and workflows. It is completely unaware of HTTP, JSON, or any transport mechanism.

**Responsibilities:**
- Implement business rules and validation
- Orchestrate operations across multiple repositories
- Handle transactions when needed
- Return database models or simple business structs

**What services should NEVER have:**
- Structs with `json:` tags
- HTTP-related imports
- Knowledge of request/response formats

```go
package service

import "project/database"

// Service works with database models directly
type WishListService struct {
    repo   *database.WishListRepository
    limits *LimitsConfig
}

// Returns database model - no JSON concerns here
func (s *WishListService) GetWishList(id, userID string) (*database.WishList, error) {
    wishlist, err := s.repo.FindByID(id)
    if err != nil {
        return nil, err
    }

    // Business rule: access control
    if !wishlist.IsPublic && wishlist.OwnerID != userID {
        return nil, ErrAccessDenied
    }

    return wishlist, nil
}
```

### Repository Layer (Data Access)

The repository layer handles all database operations using models with `db:` tags.

**Responsibilities:**
- Execute database queries
- Map database rows to structs
- Handle database-specific error translation

**Location:** Repositories live in the `database/` package alongside models and migrations.

```go
package database

import "github.com/jmoiron/sqlx"

type WishListRepository struct {
    db *sqlx.DB
}

func (r *WishListRepository) FindByID(id string) (*WishList, error) {
    var wishlist WishList
    err := r.db.Get(&wishlist,
        "SELECT * FROM wishlists WHERE id = $1", id)
    if err != nil {
        return nil, err
    }
    return &wishlist, nil
}
```

### Database Package

The database package is the single source of truth for data structures and persistence.

**Contains:**
- Database models (structs with `db:` tags only)
- Migrations (SQL files)
- Repository implementations
- Connection/transaction helpers

```go
package database

import "time"

// Database model - ONLY db: tags, NO json: tags
type WishList struct {
    ID          string    `db:"id"`
    OwnerID     string    `db:"owner_id"`
    Title       string    `db:"title"`
    Description string    `db:"description"`
    IsPublic    bool      `db:"is_public"`
    CreatedAt   time.Time `db:"created_at"`
    UpdatedAt   time.Time `db:"updated_at"`
    // Internal fields that should never reach clients
    ViewCount   int64     `db:"view_count"`
    DeletedAt   *time.Time `db:"deleted_at"`
}
```

---

## Code Examples

### Good Pattern: Complete Handler Implementation

```go
package handler

import (
    "net/http"
    "time"

    "github.com/labstack/echo/v4"
    "project/internal/service"
)

// Request DTO with validation
type CreateWishListRequest struct {
    Title        string  `json:"title" validate:"required,min=3,max=100"`
    Description  string  `json:"description,omitempty" validate:"max=500"`
    IsPublic     bool    `json:"is_public"`
    OccasionDate *string `json:"occasion_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

// Response DTO - explicit field selection
type WishListResponse struct {
    ID          string  `json:"id"`
    Title       string  `json:"title"`
    Description string  `json:"description,omitempty"`
    IsPublic    bool    `json:"is_public"`
    CreatedAt   string  `json:"created_at"`
    UpdatedAt   string  `json:"updated_at"`
}

// List response with pagination
type WishListListResponse struct {
    Items      []WishListResponse `json:"items"`
    TotalCount int                `json:"total_count"`
    Page       int                `json:"page"`
    PageSize   int                `json:"page_size"`
}

type WishListHandler struct {
    service *service.WishListService
}

func (h *WishListHandler) Create(c echo.Context) error {
    // 1. Bind and validate request
    var req CreateWishListRequest
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "invalid request body",
        })
    }

    if err := c.Validate(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": err.Error(),
        })
    }

    // 2. Extract user context
    userID := c.Get("user_id").(string)

    // 3. Map DTO to database model
    wishlist := &database.WishList{
        OwnerID:     userID,
        Title:       req.Title,
        Description: req.Description,
        IsPublic:    req.IsPublic,
    }

    // 4. Call service
    created, err := h.service.Create(wishlist)
    if err != nil {
        return h.handleServiceError(c, err)
    }

    // 5. Map database model to response DTO
    resp := h.toResponse(created)

    return c.JSON(http.StatusCreated, resp)
}

func (h *WishListHandler) GetByID(c echo.Context) error {
    id := c.Param("id")
    userID := c.Get("user_id").(string)

    // Service returns database model
    wishlist, err := h.service.GetWishList(id, userID)
    if err != nil {
        return h.handleServiceError(c, err)
    }

    // Handler maps to API response
    return c.JSON(http.StatusOK, h.toResponse(wishlist))
}

// Private helper: maps database model to response DTO
func (h *WishListHandler) toResponse(wl *database.WishList) WishListResponse {
    return WishListResponse{
        ID:          wl.ID,
        Title:       wl.Title,
        Description: wl.Description,
        IsPublic:    wl.IsPublic,
        CreatedAt:   wl.CreatedAt.Format(time.RFC3339),
        UpdatedAt:   wl.UpdatedAt.Format(time.RFC3339),
        // Note: ViewCount and DeletedAt are NOT exposed
    }
}

func (h *WishListHandler) handleServiceError(c echo.Context, err error) error {
    switch {
    case errors.Is(err, service.ErrNotFound):
        return c.JSON(http.StatusNotFound, map[string]string{
            "error": "wishlist not found",
        })
    case errors.Is(err, service.ErrAccessDenied):
        return c.JSON(http.StatusForbidden, map[string]string{
            "error": "access denied",
        })
    case errors.Is(err, service.ErrLimitExceeded):
        return c.JSON(http.StatusUnprocessableEntity, map[string]string{
            "error": "wishlist limit exceeded",
        })
    default:
        // Log the actual error internally
        c.Logger().Error(err)
        return c.JSON(http.StatusInternalServerError, map[string]string{
            "error": "internal server error",
        })
    }
}
```

### Good Pattern: Service Layer

```go
package service

import (
    "errors"

    "project/database"
)

// Sentinel errors for business conditions
var (
    ErrNotFound      = errors.New("wishlist not found")
    ErrAccessDenied  = errors.New("access denied")
    ErrLimitExceeded = errors.New("wishlist limit exceeded")
)

type WishListService struct {
    repo     *database.WishListRepository
    maxLists int
}

func NewWishListService(repo *database.WishListRepository, maxLists int) *WishListService {
    return &WishListService{
        repo:     repo,
        maxLists: maxLists,
    }
}

// Create applies business rules and persists
func (s *WishListService) Create(wishlist *database.WishList) (*database.WishList, error) {
    // Business validation (not input format validation)
    if wishlist.OwnerID == "" {
        return nil, errors.New("owner ID is required")
    }

    // Business rule: check user's wishlist limit
    count, err := s.repo.CountByOwner(wishlist.OwnerID)
    if err != nil {
        return nil, err
    }
    if count >= s.maxLists {
        return nil, ErrLimitExceeded
    }

    // Persist and return
    return s.repo.Create(wishlist)
}

// GetWishList retrieves with access control
func (s *WishListService) GetWishList(id, requestingUserID string) (*database.WishList, error) {
    wishlist, err := s.repo.FindByID(id)
    if err != nil {
        if errors.Is(err, database.ErrNoRows) {
            return nil, ErrNotFound
        }
        return nil, err
    }

    // Business rule: access control
    if !wishlist.IsPublic && wishlist.OwnerID != requestingUserID {
        return nil, ErrAccessDenied
    }

    // Business logic: track views (fire and forget)
    go s.repo.IncrementViewCount(id)

    return wishlist, nil
}

// Delete with ownership verification
func (s *WishListService) Delete(id, requestingUserID string) error {
    wishlist, err := s.repo.FindByID(id)
    if err != nil {
        if errors.Is(err, database.ErrNoRows) {
            return ErrNotFound
        }
        return err
    }

    // Business rule: only owner can delete
    if wishlist.OwnerID != requestingUserID {
        return ErrAccessDenied
    }

    return s.repo.SoftDelete(id)
}
```

### Anti-Patterns to Avoid

```go
// ❌ WRONG: JSON tags in service layer
package service

type WishListOutput struct {
    ID   string `json:"id"`   // JSON tags don't belong in services!
    Name string `json:"name"`
}

func (s *WishListService) GetFormatted(id string) (*WishListOutput, error) {
    // Service shouldn't know about JSON formatting
    return &WishListOutput{...}, nil
}
```

```go
// ❌ WRONG: Handler exposing database model directly
func (h *WishListHandler) GetByID(c echo.Context) error {
    wishlist, _ := h.service.GetWishList(id, userID)

    // DANGEROUS: Exposes ALL fields including:
    // - ViewCount (internal metric)
    // - DeletedAt (soft delete flag)
    // - Potentially sensitive data
    return c.JSON(200, wishlist)
}
```

```go
// ❌ WRONG: Business logic in handler
func (h *WishListHandler) Create(c echo.Context) error {
    var req CreateRequest
    c.Bind(&req)

    // Business logic leaked into handler!
    count, _ := h.repo.CountByOwner(userID)
    if count >= 10 {
        return c.JSON(400, "too many wishlists")
    }

    // Handlers shouldn't access repositories directly
    wishlist, _ := h.repo.Create(...)
    return c.JSON(201, wishlist)
}
```

```go
// ❌ WRONG: HTTP knowledge in service
package service

import "net/http"

func (s *WishListService) GetWishList(id string) (int, interface{}) {
    wishlist, err := s.repo.FindByID(id)
    if err != nil {
        return http.StatusNotFound, nil  // Service shouldn't know HTTP!
    }
    return http.StatusOK, wishlist
}
```

---

## Project Structure

```
project/
├── cmd/
│   └── server/
│       └── main.go              # Application bootstrap
│
├── database/                    # ALL database concerns
│   ├── migrations/
│   │   ├── 000001_create_users.up.sql
│   │   ├── 000001_create_users.down.sql
│   │   ├── 000002_create_wishlists.up.sql
│   │   └── 000002_create_wishlists.down.sql
│   ├── models.go                # DB structs (db: tags only)
│   ├── connection.go            # DB connection setup
│   ├── errors.go                # Database-specific errors
│   ├── user_repository.go
│   └── wishlist_repository.go
│
├── internal/
│   ├── handler/
│   │   ├── dto/                 # Request/Response structs
│   │   │   ├── user.go
│   │   │   └── wishlist.go
│   │   ├── middleware/
│   │   │   ├── auth.go
│   │   │   ├── logging.go
│   │   │   └── recovery.go
│   │   ├── user.go              # User handlers
│   │   ├── wishlist.go          # Wishlist handlers
│   │   └── routes.go            # Route registration
│   │
│   ├── service/
│   │   ├── errors.go            # Business error definitions
│   │   ├── user.go
│   │   └── wishlist.go
│   │
│   └── config/
│       └── config.go            # Application configuration
│
├── docs/
│   ├── openapi.yaml             # API specification
│   └── swagger/                 # Generated Swagger UI
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Key Structure Decisions

1. **`database/` at root level**: Contains models, migrations, and repositories together. This co-location is practical and acceptable.

2. **`internal/handler/dto/`**: Separates DTOs from handler logic for better organization in larger projects. Optional for smaller projects.

3. **No `domain/` or `entity/` packages**: We intentionally skip these layers for simplicity. Database models serve as our data structures.

4. **Service errors in `service/errors.go`**: Centralized business error definitions make handler error mapping cleaner.

---

## Data Flow

### Visual Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                        HTTP Request                              │
│                    (JSON body, headers)                          │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      HANDLER LAYER                               │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Request DTO (json: + validate: tags)                   │    │
│  │  CreateWishListRequest{Title, Description, IsPublic}    │    │
│  └─────────────────────────────────────────────────────────┘    │
│                         │                                        │
│                         │ Maps to database model                 │
│                         ▼                                        │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      SERVICE LAYER                               │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  database.WishList (no json: tags)                      │    │
│  │  Business rules, validation, orchestration              │    │
│  └─────────────────────────────────────────────────────────┘    │
│                         │                                        │
│                         ▼                                        │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                    REPOSITORY LAYER                              │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  database.WishList (db: tags)                           │    │
│  │  SQL queries, row scanning                              │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                       DATABASE                                   │
│                    (PostgreSQL, etc.)                            │
└─────────────────────────────────────────────────────────────────┘
                                │
                                │ Returns database model
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      SERVICE LAYER                               │
│              Returns database.WishList to handler                │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                      HANDLER LAYER                               │
│  ┌─────────────────────────────────────────────────────────┐    │
│  │  Maps database model to Response DTO                    │    │
│  │  WishListResponse{ID, Title, CreatedAt} (json: tags)    │    │
│  │  - Formats dates as strings                             │    │
│  │  - Excludes sensitive fields (ViewCount, DeletedAt)     │    │
│  └─────────────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────┐
│                       HTTP Response                              │
│                         (JSON)                                   │
└─────────────────────────────────────────────────────────────────┘
```

### Data Transformation Points

| Location | Input | Output | Transformation |
|----------|-------|--------|----------------|
| Handler (request) | JSON body | Request DTO | Binding + validation |
| Handler → Service | Request DTO | Database model | Field mapping |
| Service → Repository | Database model | Database model | None (pass-through) |
| Repository → DB | Database model | SQL | Query execution |
| DB → Repository | SQL rows | Database model | Row scanning |
| Repository → Service | Database model | Database model | None (pass-through) |
| Service → Handler | Database model | Database model | Business logic applied |
| Handler (response) | Database model | Response DTO | **Critical mapping point** |

---

## Validation Strategy

Validation happens at two distinct layers with different purposes.

### Handler Layer: Input Validation

Validates the **format and structure** of incoming data.

```go
package handler

// Struct tag validation for request format
type CreateWishListRequest struct {
    Title        string  `json:"title" validate:"required,min=3,max=100"`
    Description  string  `json:"description,omitempty" validate:"max=500"`
    IsPublic     bool    `json:"is_public"`
    OccasionDate *string `json:"occasion_date,omitempty" validate:"omitempty,datetime=2006-01-02"`
    Tags         []string `json:"tags,omitempty" validate:"max=10,dive,min=1,max=30"`
}

func (h *WishListHandler) Create(c echo.Context) error {
    var req CreateWishListRequest

    // 1. Binding validation (JSON structure)
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": "invalid JSON format",
        })
    }

    // 2. Struct validation (field constraints)
    if err := c.Validate(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{
            "error": formatValidationError(err),
        })
    }

    // Request is now guaranteed to be well-formed
    // ...
}
```

**What handler validation checks:**
- Required fields are present
- String lengths are within bounds
- Date formats are correct
- Arrays don't exceed size limits
- Nested objects are valid

### Service Layer: Business Validation

Validates **business rules and constraints** that require domain knowledge or database access.

```go
package service

func (s *WishListService) Create(wishlist *database.WishList) (*database.WishList, error) {
    // Business rule: user must exist
    exists, err := s.userRepo.Exists(wishlist.OwnerID)
    if err != nil {
        return nil, err
    }
    if !exists {
        return nil, ErrInvalidOwner
    }

    // Business rule: wishlist limit per user
    count, err := s.repo.CountByOwner(wishlist.OwnerID)
    if err != nil {
        return nil, err
    }
    if count >= s.config.MaxWishlistsPerUser {
        return nil, ErrLimitExceeded
    }

    // Business rule: title uniqueness per user
    existing, err := s.repo.FindByOwnerAndTitle(wishlist.OwnerID, wishlist.Title)
    if err != nil && !errors.Is(err, database.ErrNoRows) {
        return nil, err
    }
    if existing != nil {
        return nil, ErrDuplicateTitle
    }

    return s.repo.Create(wishlist)
}
```

**What service validation checks:**
- User has permission for the operation
- Resource limits are not exceeded
- Business invariants are maintained
- Related entities exist
- State transitions are valid

### Validation Summary

| Aspect | Handler Layer | Service Layer |
|--------|---------------|---------------|
| Purpose | Input format | Business rules |
| Trigger | Every request | After input validation |
| Examples | Required, length, format | Limits, permissions, uniqueness |
| Errors | 400 Bad Request | 403/409/422 depending on rule |
| Database | Never | When needed |

---

## Testing Approach

### Handler Tests

Test HTTP behavior, status codes, and response formats. Mock the service layer.

```go
package handler_test

import (
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "github.com/labstack/echo/v4"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestWishListHandler_Create(t *testing.T) {
    tests := []struct {
        name           string
        requestBody    string
        mockSetup      func(*MockWishListService)
        expectedStatus int
        expectedBody   string
    }{
        {
            name:        "success",
            requestBody: `{"title":"Birthday Wishes","is_public":true}`,
            mockSetup: func(m *MockWishListService) {
                m.On("Create", mock.Anything).Return(&database.WishList{
                    ID:        "123",
                    Title:     "Birthday Wishes",
                    IsPublic:  true,
                    CreatedAt: time.Now(),
                }, nil)
            },
            expectedStatus: http.StatusCreated,
            expectedBody:   `"id":"123"`,
        },
        {
            name:           "invalid json",
            requestBody:    `{invalid}`,
            mockSetup:      func(m *MockWishListService) {},
            expectedStatus: http.StatusBadRequest,
            expectedBody:   `"error"`,
        },
        {
            name:        "validation error - title too short",
            requestBody: `{"title":"Ab"}`,
            mockSetup:   func(m *MockWishListService) {},
            expectedStatus: http.StatusBadRequest,
            expectedBody:   `"error"`,
        },
        {
            name:        "service error - limit exceeded",
            requestBody: `{"title":"Another List"}`,
            mockSetup: func(m *MockWishListService) {
                m.On("Create", mock.Anything).Return(nil, service.ErrLimitExceeded)
            },
            expectedStatus: http.StatusUnprocessableEntity,
            expectedBody:   `"error"`,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup
            mockService := new(MockWishListService)
            tt.mockSetup(mockService)
            handler := NewWishListHandler(mockService)

            e := echo.New()
            req := httptest.NewRequest(http.MethodPost, "/wishlists",
                strings.NewReader(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")
            rec := httptest.NewRecorder()
            c := e.NewContext(req, rec)
            c.Set("user_id", "user-123")

            // Execute
            err := handler.Create(c)

            // Assert
            assert.NoError(t, err)
            assert.Equal(t, tt.expectedStatus, rec.Code)
            assert.Contains(t, rec.Body.String(), tt.expectedBody)
            mockService.AssertExpectations(t)
        })
    }
}
```

### Service Tests

Test business logic with mocked repositories. No HTTP concerns.

```go
package service_test

func TestWishListService_Create(t *testing.T) {
    tests := []struct {
        name        string
        input       *database.WishList
        mockSetup   func(*MockWishListRepo, *MockUserRepo)
        expectError error
    }{
        {
            name: "success",
            input: &database.WishList{
                OwnerID: "user-123",
                Title:   "My List",
            },
            mockSetup: func(wlRepo *MockWishListRepo, userRepo *MockUserRepo) {
                userRepo.On("Exists", "user-123").Return(true, nil)
                wlRepo.On("CountByOwner", "user-123").Return(0, nil)
                wlRepo.On("Create", mock.Anything).Return(&database.WishList{
                    ID:      "wl-123",
                    OwnerID: "user-123",
                    Title:   "My List",
                }, nil)
            },
            expectError: nil,
        },
        {
            name: "limit exceeded",
            input: &database.WishList{
                OwnerID: "user-123",
                Title:   "Another List",
            },
            mockSetup: func(wlRepo *MockWishListRepo, userRepo *MockUserRepo) {
                userRepo.On("Exists", "user-123").Return(true, nil)
                wlRepo.On("CountByOwner", "user-123").Return(10, nil) // At limit
            },
            expectError: ErrLimitExceeded,
        },
        {
            name: "invalid owner",
            input: &database.WishList{
                OwnerID: "nonexistent",
                Title:   "My List",
            },
            mockSetup: func(wlRepo *MockWishListRepo, userRepo *MockUserRepo) {
                userRepo.On("Exists", "nonexistent").Return(false, nil)
            },
            expectError: ErrInvalidOwner,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockWLRepo := new(MockWishListRepo)
            mockUserRepo := new(MockUserRepo)
            tt.mockSetup(mockWLRepo, mockUserRepo)

            svc := NewWishListService(mockWLRepo, mockUserRepo, 10)

            result, err := svc.Create(tt.input)

            if tt.expectError != nil {
                assert.ErrorIs(t, err, tt.expectError)
                assert.Nil(t, result)
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
            }
        })
    }
}
```

### Repository Tests (Integration)

Test actual database interactions. Use test containers or a test database.

```go
package database_test

func TestWishListRepository_Create(t *testing.T) {
    // Setup test database (e.g., using testcontainers)
    db := setupTestDB(t)
    defer db.Close()

    repo := database.NewWishListRepository(db)

    // Test
    wishlist := &database.WishList{
        OwnerID:     "user-123",
        Title:       "Test List",
        Description: "Test Description",
        IsPublic:    true,
    }

    created, err := repo.Create(wishlist)

    // Assert
    assert.NoError(t, err)
    assert.NotEmpty(t, created.ID)
    assert.Equal(t, wishlist.Title, created.Title)
    assert.False(t, created.CreatedAt.IsZero())

    // Verify persistence
    fetched, err := repo.FindByID(created.ID)
    assert.NoError(t, err)
    assert.Equal(t, created.ID, fetched.ID)
}
```

### Testing Summary

| Layer | Test Type | Mocks | Database |
|-------|-----------|-------|----------|
| Handler | Unit | Service | No |
| Service | Unit | Repository | No |
| Repository | Integration | None | Yes (test DB) |

---

## Security Considerations

### 1. Data Exposure Control

Handlers must **explicitly choose** which fields to expose. This is the primary security benefit of mapping to DTOs.

```go
// Database model with sensitive fields
type User struct {
    ID           string    `db:"id"`
    Email        string    `db:"email"`
    PasswordHash string    `db:"password_hash"`  // Never expose!
    CreatedAt    time.Time `db:"created_at"`
    LastLoginIP  string    `db:"last_login_ip"`  // PII - be careful
    FailedLogins int       `db:"failed_logins"`  // Internal metric
}

// Public response - explicit field selection
type UserResponse struct {
    ID        string `json:"id"`
    Email     string `json:"email"`
    CreatedAt string `json:"created_at"`
    // PasswordHash, LastLoginIP, FailedLogins intentionally omitted
}

// Admin response - more fields, still controlled
type AdminUserResponse struct {
    ID           string `json:"id"`
    Email        string `json:"email"`
    CreatedAt    string `json:"created_at"`
    LastLoginIP  string `json:"last_login_ip,omitempty"`
    FailedLogins int    `json:"failed_logins"`
    // PasswordHash still never exposed!
}
```

### 2. Mass Assignment Prevention

Use separate Request DTOs to prevent clients from setting fields they shouldn't.

```go
// ❌ DANGEROUS: Binding directly to database model
func (h *UserHandler) Update(c echo.Context) error {
    var user database.User
    c.Bind(&user)  // Client could set IsAdmin, PasswordHash, etc.!
    h.repo.Update(&user)
}

// ✅ SAFE: Binding to controlled DTO
type UpdateUserRequest struct {
    Name  string `json:"name" validate:"max=100"`
    Email string `json:"email" validate:"email"`
    // No IsAdmin, no PasswordHash
}

func (h *UserHandler) Update(c echo.Context) error {
    var req UpdateUserRequest
    c.Bind(&req)

    // Explicit field mapping
    user.Name = req.Name
    user.Email = req.Email
    // IsAdmin unchanged
}
```

### 3. Validation Layers

```
┌─────────────────────────────────────────────────────────────┐
│                    HANDLER LAYER                             │
│  Input Validation                                            │
│  - Format constraints (length, pattern, type)                │
│  - Required fields                                           │
│  - Prevents malformed requests from reaching business logic  │
│  - Returns 400 Bad Request                                   │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    SERVICE LAYER                             │
│  Business Validation                                         │
│  - Authorization checks                                      │
│  - Resource limits                                           │
│  - Business invariants                                       │
│  - Returns 403/409/422 depending on violation                │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                   REPOSITORY LAYER                           │
│  Database Constraints                                        │
│  - Unique constraints                                        │
│  - Foreign key constraints                                   │
│  - Not null constraints                                      │
│  - Last line of defense                                      │
└─────────────────────────────────────────────────────────────┘
```

### 4. Error Message Security

Never expose internal details in error messages.

```go
// ❌ WRONG: Exposes internal information
return c.JSON(500, map[string]string{
    "error": fmt.Sprintf("database error: %v", err),
    // Could expose: table names, query structure, connection details
})

// ✅ CORRECT: Generic message, log details internally
c.Logger().Error("database error", "error", err, "user_id", userID)
return c.JSON(500, map[string]string{
    "error": "internal server error",
})
```

---

## When to Evolve

This simplified architecture works well for many projects, but there are signs you may need more abstraction.

### Signs You Need Domain Models

1. **Database schema changes break service logic frequently**
   - Your business rules are tightly coupled to table structure
   - Adding a column requires changes in multiple services

2. **Multiple data sources**
   - You're aggregating data from DB + cache + external APIs
   - Different sources return different structures for the same concept

3. **Complex business logic detached from storage**
   - Business calculations don't map to database fields
   - You're computing derived values that have no column

4. **Multiple delivery mechanisms**
   - Same data served via REST API + gRPC + WebSocket
   - Different transports need different service interfaces

5. **Domain-Driven Design makes sense**
   - Complex domain with aggregates and value objects
   - Ubiquitous language is different from database naming

### Evolution Path

```
Current (Simplified):
Handler → Service → Repository
           ↓
    database.WishList (db: tags)

Evolved (With Domain):
Handler → Service → Repository
           ↓            ↓
    domain.WishList   database.WishListRow (db: tags)
    (no tags)              ↓
                    Mapper functions
```

When evolving, add domain models gradually:
1. Start with the most complex entities
2. Keep simple CRUD entities using database models
3. Don't refactor everything at once

---

## FAQ

### Q: Can services return database models?

**A:** Yes, for simplicity. The critical rule is that **handlers must map them to DTOs** before returning to clients. This ensures:
- Sensitive fields are never exposed accidentally
- API format is decoupled from database schema
- Services remain reusable

### Q: Where should migrations live?

**A:** In the `database/` package alongside models. This co-location is practical:
- Models and migrations evolve together
- Easier to verify consistency
- Clear ownership of database schema

### Q: How do I handle optional fields?

**A:** Use pointers and `omitempty`:

```go
// Request DTO
type CreateRequest struct {
    Title       string  `json:"title" validate:"required"`
    Description *string `json:"description,omitempty"`  // Optional
}

// Response DTO
type Response struct {
    ID          string `json:"id"`
    Description string `json:"description,omitempty"`  // Omit if empty
}
```

### Q: How do I handle different API representations?

**A:** Create multiple DTOs for different contexts:

```go
// Full representation for detail endpoints
type WishListDetailResponse struct {
    ID          string              `json:"id"`
    Title       string              `json:"title"`
    Description string              `json:"description,omitempty"`
    Items       []WishListItem      `json:"items"`
    CreatedAt   string              `json:"created_at"`
}

// Compact representation for list endpoints
type WishListSummaryResponse struct {
    ID        string `json:"id"`
    Title     string `json:"title"`
    ItemCount int    `json:"item_count"`
}

// Public representation (no owner details)
type WishListPublicResponse struct {
    ID          string `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description,omitempty"`
}
```

### Q: When should we add domain models?

**A:** When service logic becomes complex and independent of storage:
- Business rules don't map to database columns
- You're aggregating from multiple sources
- You need different representations in services vs repositories
- Domain language diverges from database naming

Start simple, evolve when complexity demands it.

### Q: Should I use an ORM or raw SQL?

**A:** This architecture works with both:
- **sqlx**: Lightweight, gives you control, database models with `db:` tags
- **GORM**: More features, uses `gorm:` tags in models
- **Raw SQL**: Maximum control, manual scanning

The principles remain the same: database concerns stay in the database package, JSON concerns stay in handlers.

### Q: How do I handle transactions?

**A:** Transactions are a service layer concern:

```go
func (s *OrderService) CreateOrder(order *database.Order, items []database.OrderItem) error {
    tx, err := s.db.BeginTx(ctx, nil)
    if err != nil {
        return err
    }
    defer tx.Rollback()

    // Use transaction-aware repository methods
    createdOrder, err := s.orderRepo.CreateTx(tx, order)
    if err != nil {
        return err
    }

    for _, item := range items {
        item.OrderID = createdOrder.ID
        if err := s.itemRepo.CreateTx(tx, &item); err != nil {
            return err
        }
    }

    return tx.Commit()
}
```

### Q: Can handlers call multiple services?

**A:** Yes, handlers can orchestrate multiple service calls:

```go
func (h *OrderHandler) Create(c echo.Context) error {
    // Validate stock availability
    available, err := h.inventoryService.CheckAvailability(items)
    if err != nil || !available {
        return c.JSON(400, "items not available")
    }

    // Create order
    order, err := h.orderService.Create(orderData)
    if err != nil {
        return h.handleError(c, err)
    }

    // Send notifications (async)
    go h.notificationService.SendOrderConfirmation(order.ID)

    return c.JSON(201, h.toResponse(order))
}
```

However, if orchestration becomes complex, consider a dedicated orchestration service.

---

## Summary

This architecture prioritizes **simplicity and pragmatism** over theoretical purity:

1. **Three layers**: Handler (HTTP/JSON) → Service (Business) → Repository (Database)
2. **One rule**: JSON serialization only in handlers
3. **Database models flow**: Through service layer, but handlers map to DTOs
4. **Co-located database concerns**: Models, migrations, and repositories together
5. **Evolve when needed**: Add domain models only when complexity demands

The goal is maintainable code that's easy to understand, test, and modify. Start simple, measure complexity, and evolve the architecture only when you have concrete needs.
