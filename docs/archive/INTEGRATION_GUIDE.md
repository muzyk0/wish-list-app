# API Routes Refactoring - Integration Guide

## Overview

This guide shows how to integrate all the refactored components into your application.

## ✅ Completed Components

### Phase 1: Database
- ✅ Migration: `000005_refactor_gift_items_many_to_many.up.sql`
- ✅ Models: `GiftItem`, `WishlistItem`, `Reservation` updated

### Phase 2-4: Backend Code
- ✅ Handlers: `item_handler.go`, `wishlist_item_handler.go`
- ✅ Services: `item_service.go`, `wishlist_item_service.go`
- ✅ Repositories: `giftitem_repository_extended.go`, `wishlistitem_repository.go`

## 🔧 Integration Steps

### Step 1: Update Repository Interfaces

Add new methods to existing repository interfaces:

**File**: `/backend/internal/repositories/giftitem_repository.go`

Add to `GiftItemRepositoryInterface`:
```go
GetByOwnerPaginated(ctx context.Context, ownerID pgtype.UUID, filters services.ItemFilters) (*PaginatedResult, error)
SoftDelete(ctx context.Context, id pgtype.UUID) error
GetUnattached(ctx context.Context, ownerID pgtype.UUID) ([]*db.GiftItem, error)
CreateWithOwner(ctx context.Context, giftItem db.GiftItem) (*db.GiftItem, error)
UpdateWithNewSchema(ctx context.Context, giftItem *db.GiftItem) (*db.GiftItem, error)
```

### Step 2: Update main.go - Initialize Services

**File**: `/backend/cmd/server/main.go`

Find the section where services are initialized (around line 170-190) and add:

```go
// Initialize repositories
wishlistRepo := repositories.NewWishListRepository(database)
giftItemRepo := repositories.NewGiftItemRepository(database)
wishlistItemRepo := repositories.NewWishlistItemRepository(database) // NEW
templateRepo := repositories.NewTemplateRepository(database)
reservationRepo := repositories.NewReservationRepository(database)
userRepo := repositories.NewUserRepository(database)

// Initialize services
wishListService := services.NewWishListService(
	wishlistRepo,
	giftItemRepo,
	templateRepo,
	emailService,
	reservationRepo,
	cache,
)

// NEW: Initialize item service
itemService := services.NewItemService(
	giftItemRepo,
	wishlistItemRepo,
)

// NEW: Initialize wishlist-item service
wishlistItemService := services.NewWishlistItemService(
	wishlistRepo,
	giftItemRepo,
	wishlistItemRepo,
)

reservationService := services.NewReservationService(
	giftItemRepo,
	wishlistRepo,
	reservationRepo,
	emailService,
)
```

### Step 3: Update main.go - Initialize Handlers

After service initialization, add handler initialization:

```go
// Initialize handlers
wishListHandler := handlers.NewWishListHandler(wishListService)
itemHandler := handlers.NewItemHandler(itemService) // NEW
wishlistItemHandler := handlers.NewWishlistItemHandler(wishlistItemService) // NEW
reservationHandler := handlers.NewReservationHandler(reservationService)
userHandler := handlers.NewUserHandler(userService, codeStore, *authCfg)
```

### Step 4: Update main.go - Register Routes

Find the route registration section (around line 310-344) and replace with:

```go
// ==================== ITEMS (Independent Resource) ====================
itemsGroup := e.Group("/api/items")
itemsGroup.Use(auth.JWTMiddleware(tokenManager))
itemsGroup.GET("", itemHandler.GetMyItems)                      // List my items with filters
itemsGroup.POST("", itemHandler.CreateItem)                     // Create item without wishlist
itemsGroup.GET("/:id", itemHandler.GetItem)                     // Get specific item
itemsGroup.PUT("/:id", itemHandler.UpdateItem)                  // Update item
itemsGroup.DELETE("/:id", itemHandler.DeleteItem)               // Soft delete (archive)
itemsGroup.POST("/:id/mark-purchased", itemHandler.MarkItemAsPurchased) // Mark as purchased

// ==================== WISHLISTS ====================
wishlistsGroup := e.Group("/api/wishlists")
wishlistsGroup.Use(auth.JWTMiddleware(tokenManager))
wishlistsGroup.GET("", wishListHandler.GetWishListsByOwner)     // List my wishlists
wishlistsGroup.POST("", wishListHandler.CreateWishList)         // Create wishlist
wishlistsGroup.GET("/:id", wishListHandler.GetWishList)         // Get specific wishlist
wishlistsGroup.PUT("/:id", wishListHandler.UpdateWishList)      // Update wishlist
wishlistsGroup.DELETE("/:id", wishListHandler.DeleteWishList)   // Delete wishlist

// Wishlist-Items Relationships (Many-to-Many)
wishlistsGroup.GET("/:id/items", wishlistItemHandler.GetWishlistItems)              // Get items in wishlist
wishlistsGroup.POST("/:id/items", wishlistItemHandler.AttachItemToWishlist)         // Attach existing item
wishlistsGroup.POST("/:id/items/new", wishlistItemHandler.CreateItemInWishlist)     // Create + attach
wishlistsGroup.DELETE("/:id/items/:itemId", wishlistItemHandler.DetachItemFromWishlist) // Detach

// ==================== PUBLIC ====================
publicWishlistGroup := e.Group("/api/public/wishlists")
publicWishlistGroup.GET("/:slug", wishListHandler.GetWishListByPublicSlug)
publicWishlistGroup.GET("/:slug/items", wishlistItemHandler.GetWishlistItems) // Reuse same handler

// ==================== RESERVATIONS ====================
reservationsGroup := e.Group("/api/reservations")
reservationsGroup.Use(auth.JWTMiddleware(tokenManager))
reservationsGroup.POST("/wishlists/:wishlistId/items/:itemId", reservationHandler.CreateReservation)
reservationsGroup.DELETE("/wishlists/:wishlistId/items/:itemId", reservationHandler.CancelReservation)
reservationsGroup.GET("", reservationHandler.GetMyReservations)

// Public reservation status
publicReservationsGroup := e.Group("/api/public/reservations")
publicReservationsGroup.GET("/wishlists/:slug/items/:itemId", reservationHandler.GetReservationStatus)

// ==================== DEPRECATED (Remove after migration) ====================
// OLD: /api/gift-items/wishlist/:wishlistId
// These routes will be removed in future version
```

### Step 5: Update Repository Method Calls

**IMPORTANT**: After running the migration, the following repository methods will need updates:

#### GiftItemRepository.Create

**Old (with wishlist_id)**:
```go
func (r *GiftItemRepository) Create(ctx context.Context, giftItem db.GiftItem) (*db.GiftItem, error) {
	query := `INSERT INTO gift_items (wishlist_id, name, ...) VALUES ($1, $2, ...)`
	// ...
}
```

**New (with owner_id)** - Use `CreateWithOwner` instead:
```go
// Use the new method from giftitem_repository_extended.go
createdItem, err := itemRepo.CreateWithOwner(ctx, item)
```

#### GiftItemRepository.Update

**Replace calls to `Update`** with `UpdateWithNewSchema`:
```go
// Old
updatedItem, err := s.itemRepo.Update(ctx, item)

// New
updatedItem, err := s.itemRepo.UpdateWithNewSchema(ctx, &item)
```

### Step 6: Run Database Migration

```bash
cd backend
make migrate-up
```

Verify migration succeeded:
```sql
-- Check new schema
\d gift_items        -- Should have owner_id, archived_at
\d wishlist_items    -- Should exist
\d reservations      -- Should have wishlist_id
```

### Step 7: Update Item Service to Use New Methods

**File**: `/backend/internal/services/item_service.go`

Update `CreateItem` method:
```go
// OLD
createdItem, err := s.itemRepo.Create(ctx, item)

// NEW
createdItem, err := s.itemRepo.CreateWithOwner(ctx, item)
```

Update `UpdateItem` method:
```go
// OLD
updatedItem, err := s.itemRepo.Update(ctx, item)

// NEW
updatedItem, err := s.itemRepo.UpdateWithNewSchema(ctx, item)
```

### Step 8: Regenerate Swagger Documentation

```bash
cd backend
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

Verify:
```bash
# Check that new routes are documented
grep -A 5 "/items" docs/swagger.yaml
grep -A 5 "/wishlists/{id}/items" docs/swagger.yaml
```

### Step 9: Copy OpenAPI Specs to API Directory

```bash
cp backend/docs/swagger.yaml api/openapi3.yaml
cp backend/docs/swagger.json api/openapi3.json
```

### Step 10: Regenerate Frontend/Mobile API Clients

**Frontend**:
```bash
cd frontend
pnpm generate:api
```

**Mobile**:
```bash
cd mobile
pnpm generate:api
```

## 🧪 Testing Checklist

### Backend Tests

```bash
cd backend

# Test items endpoints
curl -X GET http://localhost:8080/api/items \
  -H "Authorization: Bearer $TOKEN"

curl -X POST http://localhost:8080/api/items \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test Item","price":99.99}'

# Test wishlist-items relationships
curl -X POST http://localhost:8080/api/wishlists/$WISHLIST_ID/items \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"itemId":"'$ITEM_ID'"}'

curl -X GET http://localhost:8080/api/wishlists/$WISHLIST_ID/items \
  -H "Authorization: Bearer $TOKEN"

curl -X DELETE http://localhost:8080/api/wishlists/$WISHLIST_ID/items/$ITEM_ID \
  -H "Authorization: Bearer $TOKEN"

# Test soft delete
curl -X DELETE http://localhost:8080/api/items/$ITEM_ID \
  -H "Authorization: Bearer $TOKEN"

# Verify item is archived (not in default queries)
curl -X GET http://localhost:8080/api/items \
  -H "Authorization: Bearer $TOKEN"

# Include archived items
curl -X GET "http://localhost:8080/api/items?include_archived=true" \
  -H "Authorization: Bearer $TOKEN"
```

### Database Verification

```sql
-- Verify owner_id exists
SELECT id, owner_id, name, archived_at FROM gift_items LIMIT 5;

-- Verify wishlist_items associations
SELECT * FROM wishlist_items LIMIT 5;

-- Verify reservations have wishlist_id
SELECT id, wishlist_id, gift_item_id FROM reservations LIMIT 5;

-- Count unattached items
SELECT COUNT(*) FROM gift_items gi
WHERE NOT EXISTS (
	SELECT 1 FROM wishlist_items wi WHERE wi.gift_item_id = gi.id
);
```

## 🔄 Migration Path for Existing Clients

### For Mobile App

**Old code**:
```typescript
// OLD: Get items in wishlist
const { data } = await client.GET('/gift-items/wishlist/{wishlistId}', {
  params: { path: { wishlistId: '123' } }
});
```

**New code**:
```typescript
// NEW: Get items in wishlist
const { data } = await client.GET('/wishlists/{id}/items', {
  params: { path: { id: '123' } }
});

// NEW: Create item without wishlist
const { data } = await client.POST('/items', {
  body: { title: 'iPhone 15', price: 999 }
});

// NEW: Attach existing item to wishlist
await client.POST('/wishlists/{id}/items', {
  params: { path: { id: wishlistId } },
  body: { itemId: existingItemId }
});
```

### For Frontend

Same changes as mobile - regenerated API client will have new methods.

## ⚠️ Breaking Changes

### Deprecated Endpoints

These endpoints will be **removed** in next version:
- ❌ `GET /api/gift-items/wishlist/:wishlistId` → Use `GET /api/wishlists/:id/items`
- ❌ `POST /api/gift-items/wishlist/:wishlistId` → Use `POST /api/wishlists/:id/items/new`

### Database Schema Changes

- ❌ `gift_items.wishlist_id` **removed**
- ✅ `gift_items.owner_id` **added**
- ✅ `gift_items.archived_at` **added**
- ✅ `wishlist_items` table **created**
- ✅ `reservations.wishlist_id` **added**

## 📝 Next Steps

1. **Run migration**: `make migrate-up`
2. **Update main.go**: Follow Steps 2-4 above
3. **Test locally**: Use curl commands from Testing Checklist
4. **Regenerate clients**: Frontend + Mobile API clients
5. **Update client code**: Replace old endpoint calls
6. **Deploy**: Backend → Mobile → Frontend (in this order)

## 🆘 Troubleshooting

### Migration Failed

```bash
# Rollback
make migrate-down

# Check logs
cat backend/logs/migration.log

# Fix issues and retry
make migrate-up
```

### "column wishlist_id does not exist"

This means old code is still using old schema. Check:
1. Did you update repository method calls? (CreateWithOwner, UpdateWithNewSchema)
2. Did you restart the backend after code changes?

### "item already attached to this wishlist"

This is expected behavior - you can't attach the same item twice. Use different item or detach first.

### API clients not updated

```bash
# Regenerate from scratch
rm -rf frontend/lib/api/schema.ts
rm -rf mobile/lib/api/schema.ts
cd frontend && pnpm generate:api
cd mobile && pnpm generate:api
```

## 📚 Additional Resources

- [API Routes Refactoring Plan](./API_ROUTES_REFACTORING.md)
- [Backend Architecture Guide](./Go-Architecture-Guide.md)
- [Migration Scripts](../backend/internal/db/migrations/)
