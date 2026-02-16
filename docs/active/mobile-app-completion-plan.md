# Mobile App Completion Plan

**Date**: 2026-01-31
**Status**: In Progress - Critical fixes completed, remaining work identified

## Current State Analysis

### ✅ Completed (Tasks #63-#67)
- API_BASE_URL fixed with `/api` prefix
- Authentication paths updated (`/auth/*`)
- Profile paths updated (`/protected/*`)
- Type imports fixed (14 errors resolved)
- Wishlist/gift item paths updated
- **Result**: Reduced from 42 to 27 TypeScript errors (-36%)

### ⚠️ Remaining TypeScript Errors (27)

**Type Narrowing Issues (18 errors)**:
- Optional fields need null checks: `view_count`, `email`, `priority`, `avatar_url`, etc.
- Quick fix: Add optional chaining `?.` or default values

**Missing Types (2 errors)**:
- `PublicGiftItem` - needed for public wishlist viewing
- `PublicWishList` - needed for public wishlist metadata

**Template Functionality (4 errors)**:
- Template type not defined
- `getTemplates()` method missing
- `updateWishListTemplate()` method missing

**Field Naming (1 error)**:
- ReservationButton using `guest_name` instead of `guestName`

**Auth Token (2 errors)**:
- Token field is optional, needs null checks in setToken calls

## Missing API Client Methods

Based on backend endpoints, these methods are missing:

1. **Image Upload**
   - `uploadImage(file: File): Promise<{ url: string }>`
   - Endpoint: `POST /s3/upload`

2. **Guest Reservations**
   - `getGuestReservations(token: string): Promise<Reservation[]>`
   - Endpoint: `GET /reservations/guest?token={token}`

3. **Public Reservation Status**
   - `getReservationStatus(slug: string, itemId: string): Promise<ReservationStatus>`
   - Endpoint: `GET /public/wishlists/{slug}/gift-items/{itemId}/reservation-status`

4. **Account Deletion** (needs path fix)
   - Current: `DELETE /protected/account`
   - Check if backend has this endpoint or use different path

5. **Template Management** (if implemented in backend)
   - `getTemplates(): Promise<Template[]>`
   - `getTemplateById(id: string): Promise<Template>`
   - `updateWishListTemplate(wishlistId: string, templateId: string): Promise<WishList>`

## Missing Screens/Pages

### Critical (Core Functionality)
1. **Gift Item Create Screen**
   - Path: `/gift-items/create`
   - Currently missing - only edit screen exists
   - Needed for adding items to wishlists

2. **Reservation Details Screen**
   - Path: `/reservations/[id]`
   - View reservation details
   - Cancel reservation option

3. **Search/Discover Screen**
   - Exists as `/explore.tsx` but may need implementation
   - Search public wishlists
   - Browse popular lists

### Nice-to-Have
4. **Settings Screen**
   - App preferences
   - Notification settings
   - Privacy settings

5. **Onboarding Flow**
   - Welcome screens for new users
   - Tutorial/guide

6. **Error/Not Found Pages**
   - 404 handler
   - Error boundary screens

## REST API Best Practices Review

### ✅ Good Practices Already Followed
- RESTful resource naming (`/wishlists`, `/gift-items`)
- Proper HTTP methods (GET, POST, PUT, DELETE)
- Nested resources for relationships (`/wishlists/{id}/gift-items`)
- Public vs protected routes separation

### ⚠️ Areas for Improvement

**1. Inconsistent Resource Naming**
- **Issue**: Mix of nested and flat resources
  - Gift items: Both `/gift-items/{id}` AND `/wishlists/{id}/gift-items`
- **Recommendation**: Choose one pattern consistently
  - Option A: Always use nested routes when there's a parent relationship
  - Option B: Use flat routes with query parameters for filtering

**2. Missing Pagination**
- **Issue**: No pagination for lists that could grow large
  - GET `/wishlists` - needs pagination
  - GET `/wishlists/{id}/gift-items` - needs pagination
  - GET `/reservations` - needs pagination
- **Recommendation**: Add `?page=1&limit=20` query parameters

**3. Missing Filtering/Sorting**
- **Issue**: No way to filter or sort results
- **Recommendation**: Add query parameters
  - `GET /wishlists?sort=created_at&order=desc`
  - `GET /wishlists?is_public=true`

**4. Batch Operations Missing**
- **Issue**: Can only delete one item at a time
- **Recommendation**: Add batch endpoints
  - `DELETE /gift-items?ids=1,2,3`
  - `PUT /gift-items/batch` for bulk updates

**5. PATCH vs PUT**
- **Issue**: Using PUT for partial updates
- **Recommendation**: Use PATCH for partial updates, PUT for full replacement
  - Current: `PUT /gift-items/{id}` with partial body
  - Better: `PATCH /gift-items/{id}` for partial updates

**6. Response Consistency**
- **Issue**: Some endpoints return arrays, others return wrapped objects
  - `GET /wishlists` returns `{data: [...], pagination: {...}}`
  - `GET /wishlists/{id}` returns object directly
- **Recommendation**: Consistent response format
  ```json
  {
    "data": {...},
    "meta": {"timestamp": "..."}
  }
  ```

**7. Error Response Format**
- **Recommendation**: Standardize error responses
  ```json
  {
    "error": {
      "code": "VALIDATION_ERROR",
      "message": "Invalid request",
      "details": [...]
    }
  }
  ```

## Implementation Priority

### Phase 1: Critical Fixes (Blocking) - Tasks #68-#72
1. **#68**: Fix type narrowing for optional fields (18 errors)
2. **#69**: Add missing PublicGiftItem and PublicWishList types
3. **#70**: Fix field naming (guest_name → guestName)
4. **#71**: Add null checks for auth token
5. **#72**: Add missing API client methods (uploadImage, getGuestReservations, etc.)

### Phase 2: Essential Features - Tasks #73-#76
1. **#73**: Create gift item create screen
2. **#74**: Implement reservation details screen
3. **#75**: Add image upload functionality
4. **#76**: Fix Template functionality (if backend supports it)

### Phase 3: API Improvements - Tasks #77-#80
1. **#77**: Add pagination support to list endpoints
2. **#78**: Implement filtering and sorting
3. **#79**: Standardize response formats
4. **#80**: Add batch operations for efficiency

### Phase 4: Polish - Tasks #81-#85
1. **#81**: Implement search/discover functionality
2. **#82**: Add settings screen
3. **#83**: Create onboarding flow
4. **#84**: Add error handling screens
5. **#85**: Improve loading and empty states

## Type Safety Improvements

### Current Issues
```typescript
// ❌ Problem: Optional fields accessed without checks
item.view_count  // Error: possibly undefined
user.email       // Error: possibly undefined

// ✅ Solution: Add null checks
item.view_count ?? 0
user.email ?? ''
user?.email || 'No email'
```

### Recommended Type Guards
```typescript
// Add to lib/api/types.ts
export function isPublicWishList(list: any): list is PublicWishList {
  return list && typeof list.public_slug === 'string';
}

export function hasEmail(user: User): user is User & { email: string } {
  return typeof user.email === 'string';
}
```

## Backend Endpoint Recommendations

### New Endpoints Needed
1. `GET /wishlists/{id}/stats` - wishlist statistics
2. `GET /users/me/stats` - user activity statistics
3. `POST /wishlists/{id}/share` - generate shareable link
4. `GET /templates` - list available templates
5. `GET /templates/{id}` - get template details

### Endpoints to Review
1. `DELETE /protected/account` - verify this exists in backend
2. Template endpoints - check if implemented
3. Batch operations - consider adding

## Testing Requirements

### Unit Tests Needed
- API client method tests
- Type guard tests
- Helper function tests

### Integration Tests Needed
- Auth flow end-to-end
- Wishlist CRUD operations
- Reservation flow
- Public wishlist viewing

### E2E Tests Needed
- User registration → login → create wishlist → add items
- Guest viewing public wishlist → making reservation
- User viewing their reservations → canceling

## Documentation Updates Needed

1. **API Client Documentation**
   - Document all methods
   - Add usage examples
   - Error handling guide

2. **Screen Documentation**
   - Navigation flow diagrams
   - Component usage examples

3. **Type System Documentation**
   - Type definitions reference
   - Type guard usage

## Success Metrics

**Current**: 27 TypeScript errors, core functionality works
**Phase 1 Target**: 0 TypeScript errors, all critical features work
**Phase 2 Target**: All essential screens implemented
**Phase 3 Target**: API best practices implemented
**Phase 4 Target**: Production-ready polish complete
