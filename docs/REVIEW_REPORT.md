# API Routes Refactoring - Pre-Integration Review Report

## 📋 Executive Summary

**Status**: ✅ Ready for integration with **2 critical fixes** required
**Risk Level**: 🟡 Medium (migration has edge cases)
**Recommendation**: Apply fixes → Test on staging → Deploy to production

---

## ✅ Swagger Annotations Review

### Item Handler (`item_handler.go`)

| Endpoint | Route | Status | Notes |
|----------|-------|--------|-------|
| GetMyItems | `GET /items` | ✅ Fixed | Fixed `include_archived` parameter formatting |
| CreateItem | `POST /items` | ✅ OK | All annotations correct |
| GetItem | `GET /items/{id}` | ✅ OK | Path parameter correct |
| UpdateItem | `PUT /items/{id}` | ✅ OK | Request body validation present |
| DeleteItem | `DELETE /items/{id}` | ✅ OK | Soft delete documented |
| MarkPurchased | `POST /items/{id}/mark-purchased` | ✅ OK | Purchase request model present |

**Issues Found & Fixed**:
- ✅ **FIXED**: Line 91 - `include_archived` parameter had incorrect spacing

### Wishlist Item Handler (`wishlist_item_handler.go`)

| Endpoint | Route | Status | Notes |
|----------|-------|--------|-------|
| GetWishlistItems | `GET /wishlists/{id}/items` | ✅ OK | Pagination params documented |
| AttachItem | `POST /wishlists/{id}/items` | ✅ OK | Request model correct |
| CreateItemInWishlist | `POST /wishlists/{id}/items/new` | ✅ OK | Different route from attach |
| DetachItem | `DELETE /wishlists/{id}/items/{itemId}` | ✅ OK | Two path params handled |

**All annotations correct** ✅

### Response Models Validation

All referenced response models exist:
- ✅ `PaginatedItemsResponse` - defined in `item_handler.go`
- ✅ `ItemResponse` - defined in `item_handler.go`
- ✅ `CreateItemRequest` - defined in `item_handler.go`
- ✅ `UpdateItemRequest` - defined in `item_handler.go`
- ✅ `AttachItemRequest` - defined in `wishlist_item_handler.go`

---

## 🔴 Migration SQL Critical Issues

### Issue #1: Reservation Wishlist Association Logic

**Location**: Lines 61-65 in `000005_refactor_gift_items_many_to_many.up.sql`

```sql
UPDATE reservations r
SET wishlist_id = wi.wishlist_id
FROM wishlist_items wi
WHERE r.gift_item_id = wi.gift_item_id
  AND r.wishlist_id IS NULL;
```

**Problem**: If an item is attached to multiple wishlists, this query will pick an **arbitrary** wishlist (whichever is returned first), which may not be the correct wishlist for the reservation.

**Risk**: 🔴 **HIGH** - Reservations may be linked to wrong wishlists, causing:
- Users seeing wrong reservation status
- Items appearing as reserved in wrong wishlists
- Confusion when managing reservations

**Solution Required**: Need to handle multi-wishlist items properly.

**Recommended Fix**:
```sql
-- Option A: Keep only reservations that can be unambiguously assigned
-- Delete reservations for items that are in multiple wishlists (edge case)
DELETE FROM reservations r
WHERE r.gift_item_id IN (
    SELECT gift_item_id
    FROM wishlist_items
    GROUP BY gift_item_id
    HAVING COUNT(*) > 1
);

-- Then do the update for remaining items
UPDATE reservations r
SET wishlist_id = wi.wishlist_id
FROM wishlist_items wi
WHERE r.gift_item_id = wi.gift_item_id
  AND r.wishlist_id IS NULL;

-- Option B: Prompt user to manually resolve conflicts (safer)
-- Log items with multiple wishlists and reservations
DO $$
DECLARE
    conflict_count INT;
BEGIN
    SELECT COUNT(DISTINCT r.id) INTO conflict_count
    FROM reservations r
    WHERE r.gift_item_id IN (
        SELECT gift_item_id
        FROM wishlist_items
        GROUP BY gift_item_id
        HAVING COUNT(*) > 1
    );

    IF conflict_count > 0 THEN
        RAISE NOTICE 'WARNING: % reservations need manual wishlist assignment', conflict_count;
    END IF;
END $$;
```

### Issue #2: NOT NULL Constraint Without Validation

**Location**: Line 69 in migration

```sql
ALTER TABLE reservations
ALTER COLUMN wishlist_id SET NOT NULL;
```

**Problem**: This will **fail** if any reservations couldn't get a wishlist_id assigned in the previous UPDATE.

**Risk**: 🟡 **MEDIUM** - Migration will fail if:
- Items were deleted but reservations still exist
- Items are not in any wishlist

**Solution Required**: Add validation before setting NOT NULL.

**Recommended Fix**:
```sql
-- Check for null values before setting NOT NULL
DO $$
DECLARE
    null_count INT;
BEGIN
    SELECT COUNT(*) INTO null_count
    FROM reservations
    WHERE wishlist_id IS NULL;

    IF null_count > 0 THEN
        RAISE EXCEPTION 'Cannot set wishlist_id to NOT NULL: % reservations have NULL wishlist_id', null_count;
    END IF;
END $$;

-- Only set NOT NULL if no NULLs exist
ALTER TABLE reservations
ALTER COLUMN wishlist_id SET NOT NULL;
```

---

## 🟢 Architecture Changes Review

### Schema Changes

| Change | Impact | Risk | Mitigation |
|--------|--------|------|------------|
| Add `owner_id` to `gift_items` | ✅ Items belong to users | Low | Populated from wishlist owner |
| Remove `wishlist_id` from `gift_items` | 🔴 Breaking change | High | Many-to-many via join table |
| Add `archived_at` to `gift_items` | ✅ Soft delete | Low | NULL by default |
| Create `wishlist_items` table | ✅ Many-to-many | Low | Proper indexes added |
| Add `wishlist_id` to `reservations` | 🟡 Changes reservation logic | Medium | **Needs fix** (see Issue #1) |

### Code Quality

✅ **Strengths**:
- Clean separation of concerns (handler → service → repository)
- Proper error handling with sentinel errors
- Pagination implemented correctly
- Soft delete pattern properly implemented
- Many-to-many relationship handled correctly
- All methods have proper authorization checks

⚠️ **Potential Improvements**:
- Consider adding transaction support for CreateItemInWishlist
- Add rate limiting for bulk operations
- Consider adding item count cache in wishlists

### Breaking Changes for Clients

| Old Endpoint | New Endpoint | Breaking? | Migration |
|--------------|--------------|-----------|-----------|
| `GET /api/gift-items/wishlist/:id` | `GET /api/wishlists/:id/items` | ✅ Yes | Update API calls |
| `POST /api/gift-items/wishlist/:id` | `POST /api/wishlists/:id/items/new` | ✅ Yes | Update API calls |
| - | `GET /api/items` | ✅ New | Add new feature |
| - | `POST /api/items` | ✅ New | Add new feature |

---

## 🔧 Required Fixes Before Integration

### Fix #1: Update Migration SQL (CRITICAL)

**File**: `/backend/internal/db/migrations/000005_refactor_gift_items_many_to_many.up.sql`

**Lines to modify**: 54-77

**Action**: Add validation and conflict handling for reservation wishlist assignment.

### Fix #2: Verify No Orphaned Reservations

**Before running migration**, check for orphaned reservations:

```sql
-- Check for items with multiple wishlists that have reservations
SELECT
    gi.id AS item_id,
    gi.name AS item_name,
    COUNT(DISTINCT wi.wishlist_id) AS wishlist_count,
    COUNT(DISTINCT r.id) AS reservation_count
FROM gift_items gi
LEFT JOIN wishlist_items wi ON wi.gift_item_id = gi.id
LEFT JOIN reservations r ON r.gift_item_id = gi.id
GROUP BY gi.id, gi.name
HAVING COUNT(DISTINCT wi.wishlist_id) > 1 AND COUNT(DISTINCT r.id) > 0;
```

If any results, **manually resolve** before migration.

---

## 📊 Test Plan

### Unit Tests Required

| Component | Test Coverage | Priority |
|-----------|---------------|----------|
| ItemService | ✅ All methods | High |
| WishlistItemService | ✅ All methods | High |
| ItemHandler | ⚠️ Missing | High |
| WishlistItemHandler | ⚠️ Missing | High |
| GiftItemRepository (extended) | ⚠️ Missing | Medium |
| WishlistItemRepository | ⚠️ Missing | Medium |

**Recommendation**: Create handler tests before deployment.

### Integration Tests Checklist

- [ ] Create item without wishlist
- [ ] List my items with pagination
- [ ] Filter unattached items
- [ ] Soft delete item
- [ ] Item not in default queries after soft delete
- [ ] Include archived items query
- [ ] Attach existing item to wishlist
- [ ] Create item in wishlist
- [ ] Item appears in wishlist
- [ ] Detach item from wishlist
- [ ] Item still exists after detach
- [ ] Attach same item to multiple wishlists
- [ ] Each wishlist shows the item
- [ ] Mark item as purchased (global)
- [ ] Purchase status reflects in all wishlists

### Edge Cases to Test

- [ ] Attach item already in wishlist (should fail with 409)
- [ ] Detach item not in wishlist (should fail with 404)
- [ ] Delete item while attached to wishlists (should archive)
- [ ] Access denied tests (wrong owner)
- [ ] Public wishlist access (no auth required)
- [ ] Pagination edge cases (empty, single page, multiple pages)

---

## 🚦 Go/No-Go Criteria

### ✅ GO Criteria

- [x] All Swagger annotations correct
- [x] Response models defined
- [x] Code follows architecture patterns
- [x] Proper authorization checks
- [ ] Migration SQL issues fixed (**Required**)
- [ ] No orphaned reservations in DB
- [ ] Handler unit tests added
- [ ] Integration tests passing

### 🔴 NO-GO Criteria

- Migration SQL issues not fixed
- Orphaned reservations exist without resolution plan
- Breaking changes not communicated to frontend/mobile teams

---

## 📝 Recommended Integration Sequence

### 1. Pre-Integration (30 min)

```bash
# Fix migration SQL (apply fixes from this report)
# Add handler unit tests
# Check for orphaned reservations in production DB
```

### 2. Staging Deployment (1 hour)

```bash
# Deploy to staging environment
cd backend
make migrate-up  # Run on staging DB
make backend     # Start server
# Run integration tests
```

### 3. Smoke Tests (15 min)

```bash
# Test critical paths
curl http://staging/api/items -H "Authorization: Bearer $TOKEN"
curl http://staging/api/wishlists/$ID/items -H "Authorization: Bearer $TOKEN"
```

### 4. Client Updates (1 hour)

```bash
# Regenerate API clients
cd frontend && pnpm generate:api
cd mobile && pnpm generate:api
# Update client code to use new endpoints
# Test frontend + mobile against staging
```

### 5. Production Deployment

```bash
# Deploy in order:
# 1. Backend (with migration)
# 2. Frontend (with new API calls)
# 3. Mobile (publish new version)
```

---

## 🎯 Final Recommendation

**Status**: ⚠️ **Ready with fixes**

**Action Items** (Priority Order):

1. **🔴 CRITICAL**: Fix migration SQL (Issue #1 and #2)
2. **🟡 HIGH**: Check production DB for orphaned reservations
3. **🟢 MEDIUM**: Add handler unit tests
4. **🟢 LOW**: Run integration tests on staging

**Timeline Estimate**:
- Fixes: 1-2 hours
- Testing: 2-3 hours
- Total: **3-5 hours** before production deployment

**Risk Assessment**: 🟡 **Medium Risk** (fixable issues identified)

Once fixes applied: **✅ APPROVED for staging deployment**

---

## 📞 Questions to Resolve

1. **Reservation conflicts**: How to handle reservations for items in multiple wishlists?
   - Option A: Delete conflicting reservations (data loss)
   - Option B: Manually assign to correct wishlist (time-consuming)
   - Option C: Keep first wishlist found (current behavior, may be wrong)

2. **Rollback strategy**: If migration fails in production, what's the rollback plan?
   - Migration has `.down.sql` but will lose many-to-many data
   - Consider backup strategy before migration

3. **Client communication**: When will frontend/mobile teams be notified of breaking changes?
   - Need coordinated deployment
   - Old endpoints should return proper deprecation notices

---

**Review Date**: 2026-02-06
**Reviewer**: Claude Code
**Next Review**: After fixes applied
