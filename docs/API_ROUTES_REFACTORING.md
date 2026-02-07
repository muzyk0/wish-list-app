# API Routes Refactoring Plan

## Overview

Refactor API routes to follow REST best practices with items as independent resources that can be attached to multiple wishlists (many-to-many relationship).

## Key Changes

1. **Items as Primary Resource**: Items can exist independently without wishlists
2. **Many-to-Many**: Items can be attached to multiple wishlists
3. **Soft Delete**: Items are archived instead of physically deleted
4. **Query Parameters**: Pagination, filtering, sorting support
5. **Public/Private**: Clear separation with `/api/public/*` prefix

## New Route Structure

### Items (Primary Resource)

```
GET    /api/items                         List my items with filters
POST   /api/items                         Create item without wishlist
GET    /api/items/:id                     Get specific item
PUT    /api/items/:id                     Update item
DELETE /api/items/:id                     Soft delete (archive) item
POST   /api/items/:id/mark-purchased      Mark item as purchased globally
```

**Query Parameters for GET /api/items:**
- `page` (int): Page number, default 1
- `limit` (int): Items per page, default 10, max 100
- `sort` (string): Sort field (created_at, updated_at, title, price)
- `order` (string): Sort order (asc, desc), default desc
- `unattached` (bool): Filter items not attached to any wishlist
- `search` (string): Search in title/description (future)
- `include_archived` (bool): Include archived items, default false

**Example Requests:**
```bash
# All my items, paginated
GET /api/items?page=1&limit=20

# Latest created items
GET /api/items?sort=created_at&order=desc&limit=10

# Items not attached to any wishlist
GET /api/items?unattached=true

# Search for iPhone
GET /api/items?search=iPhone
```

### Wishlists (Primary Resource)

```
GET    /api/wishlists                     List my wishlists
POST   /api/wishlists                     Create wishlist
GET    /api/wishlists/:id                 Get specific wishlist
PUT    /api/wishlists/:id                 Update wishlist
DELETE /api/wishlists/:id                 Delete wishlist
```

### Wishlist-Items Relationships (Many-to-Many)

```
GET    /api/wishlists/:id/items                    List items in wishlist
POST   /api/wishlists/:id/items                    Attach existing item
POST   /api/wishlists/:id/items/new                Create item and attach
DELETE /api/wishlists/:id/items/:itemId            Detach item from wishlist
```

**POST /api/wishlists/:id/items** (Attach existing):
```json
{
  "itemId": "existing-item-uuid"
}
```

**POST /api/wishlists/:id/items/new** (Create and attach):
```json
{
  "title": "iPhone 15 Pro",
  "description": "256GB, Blue Titanium",
  "price": 999.99,
  "url": "https://apple.com/...",
  "imageUrl": "https://...",
  "priority": "high"
}
```

**GET /api/wishlists/:id/items** Query Parameters:
- `page` (int): Page number
- `limit` (int): Items per page

### Public Routes

```
GET /api/public/wishlists/:slug              Get public wishlist by slug
GET /api/public/wishlists/:slug/items        List items in public wishlist
```

### Reservations (Unchanged)

```
POST   /api/reservations/wishlists/:wishlistId/items/:itemId    Create reservation
DELETE /api/reservations/wishlists/:wishlistId/items/:itemId    Cancel reservation
GET    /api/reservations                                         Get my reservations

# Public
GET /api/public/reservations/wishlists/:slug/items/:itemId      Get reservation status
```

## Database Schema Changes

### Add archived_at to gift_items

```sql
ALTER TABLE gift_items
ADD COLUMN archived_at TIMESTAMP NULL DEFAULT NULL;

CREATE INDEX idx_gift_items_archived_at ON gift_items(archived_at);
```

### Verify wishlist_items join table

```sql
-- Should already exist, verify structure:
CREATE TABLE IF NOT EXISTS wishlist_items (
    wishlist_id UUID NOT NULL REFERENCES wishlists(id) ON DELETE CASCADE,
    gift_item_id UUID NOT NULL REFERENCES gift_items(id) ON DELETE CASCADE,
    added_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (wishlist_id, gift_item_id)
);

CREATE INDEX idx_wishlist_items_wishlist_id ON wishlist_items(wishlist_id);
CREATE INDEX idx_wishlist_items_gift_item_id ON wishlist_items(gift_item_id);
```

## Soft Delete Behavior

### Archive Item
```
DELETE /api/items/:id
→ Sets archived_at = NOW()
→ Item remains in database
→ Removed from all queries by default
→ Still visible in wishlists until detached
```

### Query Archived Items
```
GET /api/items?include_archived=true
→ Returns both active and archived items
```

### Restore Item (Future)
```
POST /api/items/:id/restore
→ Sets archived_at = NULL
→ Item becomes active again
```

## Handler Structure

### ItemHandler (NEW)

```go
type ItemHandler struct {
    service services.ItemServiceInterface
}

// Endpoints
func (h *ItemHandler) GetMyItems(c echo.Context) error
func (h *ItemHandler) CreateItem(c echo.Context) error
func (h *ItemHandler) GetItem(c echo.Context) error
func (h *ItemHandler) UpdateItem(c echo.Context) error
func (h *ItemHandler) DeleteItem(c echo.Context) error // Soft delete
func (h *ItemHandler) MarkPurchased(c echo.Context) error
```

### WishlistHandler Updates

```go
// New methods
func (h *WishlistHandler) GetWishlistItems(c echo.Context) error
func (h *WishlistHandler) AttachItemToWishlist(c echo.Context) error
func (h *WishlistHandler) CreateItemInWishlist(c echo.Context) error
func (h *WishlistHandler) DetachItem(c echo.Context) error

// Existing methods remain unchanged
func (h *WishlistHandler) GetMyWishlists(c echo.Context) error
func (h *WishlistHandler) CreateWishlist(c echo.Context) error
// ... etc
```

## Service Layer Changes

### ItemService (NEW)

```go
type ItemService struct {
    itemRepo repositories.GiftItemRepositoryInterface
    wishlistItemRepo repositories.WishlistItemRepositoryInterface
}

func (s *ItemService) GetMyItems(ctx context.Context, userID string, filters ItemFilters) (*PaginatedItems, error)
func (s *ItemService) CreateItem(ctx context.Context, userID string, input CreateItemInput) (*ItemOutput, error)
func (s *ItemService) GetItem(ctx context.Context, itemID string) (*ItemOutput, error)
func (s *ItemService) UpdateItem(ctx context.Context, itemID string, input UpdateItemInput) (*ItemOutput, error)
func (s *ItemService) SoftDeleteItem(ctx context.Context, itemID string, userID string) error
func (s *ItemService) MarkPurchased(ctx context.Context, itemID string, userID string, purchasedPrice float64) (*ItemOutput, error)
```

### WishlistService Updates

```go
// New methods
func (s *WishlistService) GetWishlistItems(ctx context.Context, wishlistID string, page, limit int) (*PaginatedItems, error)
func (s *WishlistService) AttachItem(ctx context.Context, wishlistID, itemID, userID string) error
func (s *WishlistService) CreateItemInWishlist(ctx context.Context, wishlistID, userID string, input CreateItemInput) (*ItemOutput, error)
func (s *WishlistService) DetachItem(ctx context.Context, wishlistID, itemID, userID string) error
```

## Repository Layer Changes

### GiftItemRepository Updates

```go
// New methods
func (r *GiftItemRepository) GetByOwner(ctx context.Context, ownerID pgtype.UUID, filters ItemFilters) ([]*db.GiftItem, error)
func (r *GiftItemRepository) GetByOwnerPaginated(ctx context.Context, ownerID pgtype.UUID, filters ItemFilters) (*PaginatedResult, error)
func (r *GiftItemRepository) SoftDelete(ctx context.Context, id pgtype.UUID) error
func (r *GiftItemRepository) GetUnattached(ctx context.Context, ownerID pgtype.UUID) ([]*db.GiftItem, error)

// Update existing queries to exclude archived by default
// WHERE archived_at IS NULL
```

### WishlistItemRepository (NEW)

```go
type WishlistItemRepository struct {
    db *sqlx.DB
}

func (r *WishlistItemRepository) Attach(ctx context.Context, wishlistID, itemID pgtype.UUID) error
func (r *WishlistItemRepository) Detach(ctx context.Context, wishlistID, itemID pgtype.UUID) error
func (r *WishlistItemRepository) GetByWishlist(ctx context.Context, wishlistID pgtype.UUID, page, limit int) ([]*db.GiftItem, error)
func (r *WishlistItemRepository) IsAttached(ctx context.Context, wishlistID, itemID pgtype.UUID) (bool, error)
func (r *WishlistItemRepository) GetWishlistsForItem(ctx context.Context, itemID pgtype.UUID) ([]pgtype.UUID, error)
```

## Migration Strategy

### Step 1: Add archived_at column
```sql
-- Migration: 000X_add_archived_at_to_gift_items.up.sql
ALTER TABLE gift_items
ADD COLUMN archived_at TIMESTAMP NULL DEFAULT NULL;

CREATE INDEX idx_gift_items_archived_at ON gift_items(archived_at);
```

### Step 2: Verify wishlist_items table
```bash
# Check if table exists and has correct structure
# If not, create it with the schema above
```

### Step 3: Migrate existing data
```sql
-- Ensure all existing items have proper associations in wishlist_items
-- This should already be done, but verify
INSERT INTO wishlist_items (wishlist_id, gift_item_id)
SELECT wishlist_id, id
FROM gift_items
WHERE wishlist_id IS NOT NULL
ON CONFLICT DO NOTHING;
```

## Testing Checklist

### Items CRUD
- [ ] Create item without wishlist
- [ ] Get my items with pagination
- [ ] Filter unattached items
- [ ] Sort by created_at, price
- [ ] Update item
- [ ] Soft delete item (archived_at set)
- [ ] Archived items not in default queries
- [ ] Mark item as purchased (global status)

### Wishlist-Items Relationships
- [ ] Attach existing item to wishlist
- [ ] Create new item in wishlist
- [ ] Get items in wishlist (paginated)
- [ ] Detach item from wishlist
- [ ] Item visible in multiple wishlists
- [ ] Detach doesn't delete item

### Public Access
- [ ] Get public wishlist by slug
- [ ] Get public wishlist items
- [ ] Private wishlists not accessible

### Reservations
- [ ] Create reservation for wishlist+item
- [ ] Cancel reservation
- [ ] Get my reservations
- [ ] Public reservation status

### Authorization
- [ ] Only owner can CRUD their items
- [ ] Only owner can manage wishlist associations
- [ ] Public access works without auth

## Breaking Changes

### For Mobile/Frontend Clients

**Old endpoints (DEPRECATED):**
```
GET /api/gift-items/wishlist/:wishlistId  → /api/wishlists/:id/items
POST /api/gift-items/wishlist/:wishlistId → /api/wishlists/:id/items/new
```

**Migration Guide:**
1. Update API client by regenerating from new OpenAPI spec
2. Replace old endpoint calls with new structure
3. Test all item-related operations

## Rollout Plan

### Phase 1: Backend (Week 1)
- [ ] Database migrations
- [ ] Repository layer
- [ ] Service layer
- [ ] Handlers
- [ ] Routes registration
- [ ] Unit tests

### Phase 2: Documentation (Week 1)
- [ ] Swagger annotations
- [ ] Regenerate OpenAPI specs
- [ ] Update API documentation

### Phase 3: Clients (Week 2)
- [ ] Regenerate mobile API client
- [ ] Regenerate frontend API client
- [ ] Update mobile app code
- [ ] Update frontend code
- [ ] Integration testing

### Phase 4: Deployment (Week 2)
- [ ] Deploy backend
- [ ] Deploy mobile app
- [ ] Deploy frontend
- [ ] Monitor for issues

## Success Metrics

- [ ] All tests pass
- [ ] API documentation accurate
- [ ] Mobile app works with new routes
- [ ] Frontend works with new routes
- [ ] No breaking changes for public API users
- [ ] Performance maintained or improved
- [ ] Soft delete works correctly
- [ ] Many-to-many associations work

## Future Enhancements

1. **Search**: Full-text search in items
2. **Tags**: Tag system for items
3. **Archived Items Management**: UI for viewing/restoring archived items
4. **Bulk Operations**: Attach multiple items at once
5. **Item History**: Track changes to items
6. **Smart Suggestions**: Recommend items based on wishlists
