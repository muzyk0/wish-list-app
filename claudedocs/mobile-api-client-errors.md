# Mobile API Client Type Errors

**Date**: 2026-01-31
**Status**: Blocking - 42 TypeScript errors preventing mobile app compilation

## Root Cause Analysis

After regenerating API types from backend OpenAPI spec, the mobile API client has fundamental mismatches between expected and actual API structure.

### Issue 1: API Base URL Missing `/api` Prefix

**File**: `mobile/lib/api/api.ts:20`
**Current**: `const API_BASE_URL = process.env.EXPO_PUBLIC_API_URL || 'http://10.0.2.2:8080';`
**Expected**: `const API_BASE_URL = process.env.EXPO_PUBLIC_API_URL || 'http://10.0.2.2:8080/api';`

**Impact**: All API requests are going to wrong URLs.

### Issue 2: Route Path Mismatches

The mobile client uses paths that don't exist in the backend:

| Mobile Client Path | Actual Backend Path | Fix Required |
|-------------------|---------------------|--------------|
| `/v1/users/register` | `/auth/register` | Update all user auth paths |
| `/v1/users/login` | `/auth/login` | Update all user auth paths |
| `/v1/users/me` | `/protected/profile` | Update profile paths |
| `/v1/wishlists` | `/wishlists` | Remove `/v1` prefix |
| `/v1/wishlists/{wishlistId}/items` | `/wishlists/{wishlistId}/gift-items` | Use `/gift-items` |
| `/v1/wishlists/{wishlistId}/items/{itemId}` | `/wishlists/{wishlistId}/gift-items/{itemId}` | Use `/gift-items` |
| `/v1/wishlists/{wishlistId}/items/{itemId}/reserve` | `/wishlists/{wishlistId}/gift-items/{itemId}/reservation` | Use `/reservation` |
| `/v1/wishlists/{wishlistId}/items/{itemId}/cancel-reservation` | `/wishlists/{wishlistId}/gift-items/{itemId}/reservation` | DELETE method |
| `/v1/wishlists/{wishlistId}/items/{itemId}/mark-purchased` | `/gift-items/{id}/purchase` | Different structure |
| `/v1/users/me/reservations` | `/reservations` | Remove `/users/me` |

### Issue 3: Schema Name Mismatches in types.ts

**File**: `mobile/lib/api/types.ts`

The manual type aliases reference schemas that don't exist:

| Expected Schema Name | Actual Schema Name |
|---------------------|-------------------|
| `user_response` | `internal_handlers.AuthResponse` |
| `user_registration` | `internal_handlers.RegisterRequest` |
| `user_login` | `internal_handlers.LoginRequest` |
| `user_update` | Not in schema |
| `wish_list_response` | `wish-list_internal_services.WishListOutput` |
| `public_wish_list_response` | Not in schema |
| `wish_list_create` | `internal_handlers.CreateWishListRequest` |
| `wish_list_update` | `internal_handlers.UpdateWishListRequest` |
| `gift_item_response` | `wish-list_internal_services.GiftItemOutput` |
| `public_gift_item_response` | Not in schema |
| `gift_item_create` | `internal_handlers.CreateGiftItemRequest` |
| `gift_item_update` | `internal_handlers.UpdateGiftItemRequest` |
| `reservation_response` | Not in schema |
| `guest_reservation` | Not in schema |
| `authenticated_reservation` | Not in schema |
| `error` | Not in schema (uses inline types) |
| `pagination` | Not in schema |

### Issue 4: Missing Template Functionality

**File**: `mobile/components/wish-list/TemplateSelector.tsx`

**Errors**:
- Line 33: `Cannot find name 'Template'`
- Line 36: `Property 'getTemplates' does not exist on type 'ApiClient'`
- Line 41: `Property 'updateWishListTemplate' does not exist on type 'ApiClient'`
- Line 61: `Cannot find name 'Template'`

**Cause**: Template endpoints not documented in backend OpenAPI spec or not implemented.

## TypeScript Errors Summary

**Total**: 42 errors
- **API path mismatches**: 24 errors (api.ts)
- **Schema name mismatches**: 14 errors (types.ts)
- **Missing Template implementation**: 4 errors (TemplateSelector.tsx)

## Recommended Fix Strategy

### Option A: Update Mobile Client (Recommended)

**Pros**:
- Matches actual backend implementation
- Uses generated types correctly
- No backend changes needed

**Cons**:
- Requires updating all API method calls in mobile app
- Breaking change for mobile codebase

**Tasks**:
1. Fix API_BASE_URL to include `/api` suffix
2. Update all API method paths to match backend routes
3. Replace manual type aliases in types.ts with direct schema imports
4. Investigate Template endpoints in backend (check if implemented)
5. Update all components using the old API paths

### Option B: Update Backend Routes

**Pros**:
- Mobile client code doesn't need changes
- Maintains mobile API contract

**Cons**:
- Requires backend route restructuring
- Breaking change for any existing clients
- More complex refactoring

**Not Recommended**: Backend routes are established and documented in Swagger.

### Option C: Dual API Support

**Pros**:
- Gradual migration possible
- No breaking changes

**Cons**:
- Maintenance overhead
- Code duplication
- Not sustainable long-term

**Not Recommended**: Adds unnecessary complexity.

## Implementation Priority

### Critical (Blocking)
1. Fix API_BASE_URL to include `/api` prefix
2. Update authentication paths (`/auth/register`, `/auth/login`)
3. Update profile paths (`/protected/profile`)
4. Fix types.ts schema imports

### High Priority
1. Update wishlist CRUD paths
2. Update gift item paths
3. Update reservation paths
4. Fix Template functionality (or remove if not implemented)

### Medium Priority
1. Update test files to use new paths
2. Update documentation
3. Verify all API integrations work correctly

## Next Steps

1. **Create tasks** for each path update in task management system
2. **Implement fixes** starting with critical issues
3. **Run type-check** after each major change
4. **Test API calls** in development environment
5. **Update documentation** with correct API structure
