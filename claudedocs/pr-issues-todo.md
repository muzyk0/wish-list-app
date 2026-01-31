# GitHub PR #8 Issues - Todo List

> **Source**: Pull Request #8 "Add complete mobile app implementation with Expo Router"
> **Date Created**: 2026-01-31
> **Total Issues**: 40

## Overview

This document tracks all issues identified in the GitHub PR review comments from CodeRabbit AI. Issues are organized by component and priority.

---

## Mobile App Issues (23 tasks)

### Profile Screen (3 issues)

- [ ] #### #1: Fix email field persistence in profile update
**File**: `mobile/app/(tabs)/profile.tsx`
**Location**: Around line 50-88
**Priority**: High
**Issue**: The email field is validated in handleUpdateProfile but omitted from the mutation payload, so edits never persist.
**Fix**: Update the mutation function used by updateMutation to include `email: userData.email` in the object sent to `apiClient.updateProfile`. Alternatively, remove the email input/validation if email should not be editable.

- [ ] #### #24: Clear session after account deletion
**File**: `mobile/app/(tabs)/profile.tsx`
**Location**: Around line 88-99
**Priority**: Critical
**Issue**: After successful account deletion, the app doesn't clear the authenticated session and cached profile data.
**Fix**: In the deleteMutation onSuccess handler:
- Call `setAuthUser(null)` or AuthContext signOut method
- Invalidate/clear react-query caches: `queryClient.invalidateQueries(['profile'])` or `queryClient.clear()`
- Navigate to login screen via `navigation.reset` or `navigate('Auth')`

- [ ] #### #25: Fix Avatar rendering with conditional logic
**File**: `mobile/app/(tabs)/profile.tsx`
**Location**: Around line 158-170
**Priority**: Medium
**Issue**: Avatar.Image component does not render children, so fallback initials are not shown.
**Fix**: Replace with conditional rendering:
- If `user?.avatar_url` is truthy: render `Avatar.Image` with `source={{ uri: user.avatar_url }}`
- Otherwise: render `Avatar.Text` with computed initials `((user.first_name.charAt(0) + (user.last_name?.charAt(0) || '')).toUpperCase())` as the label prop

---

### Auth Flow (2 issues)

- [ ] #### #26: Store auth token after login
**File**: `mobile/app/auth/login.tsx`
**Location**: Around line 32-43
**Priority**: Critical
**Issue**: The onSuccess handler for loginUser mutation only navigates and skips storing the auth token.
**Fix**: Update onSuccess to:
1. Extract the token (`data.token` or `data.accessToken`)
2. Persist it using `expo-secure-store` (or AsyncStorage as fallback) under a clear key
3. Handle storage errors
4. Only after successful storage perform `router.push('/(tabs)')`

- [ ] #### #27: Fix linking config
**File**: `mobile/app/linking.ts`
**Location**: Around line 18-58
**Priority**: Medium
**Issue**: The linking config uses a non-standard alias on the reservations screen and repeats full child paths.
**Fix**:
- Remove the alias field from the reservations entry
- Refactor the group to declare a parent path
- Change nested child screen paths to relative paths
- Use only standard React Navigation linking keys (path on groups and relative child paths)

---

### Public Wishlist Screen (5 issues)

- [ ] #### #2: Fix zero price display logic
**File**: `mobile/app/public/[slug].tsx`
**Location**: Around line 70-72
**Priority**: Medium
**Issue**: The JSX currently hides zero prices by checking `item.price !== 0`.
**Fix**: Update the conditional to use a nullish check instead: `item.price != null` so 0 is rendered but null/undefined are still suppressed.

- [ ] #### #3: Fix FlatList data binding
**File**: `mobile/app/public/[slug].tsx`
**Location**: Around line 108-111
**Priority**: High
**Issue**: The FlatList is hardcoded to `data={[]}` which discards the normalized items.
**Fix**: Change the data prop to use the normalized list: `data={giftItems ?? wishList?.giftItems ?? []}` so FlatList receives the fetched items and falls back to an empty array if missing.

- [ ] #### #4: Fix reservation status display
**File**: `mobile/app/public/[slug].tsx`
**Location**: Around line 50-88
**Priority**: High
**Issue**: The component incorrectly reads flat fields on PublicGiftItem.
**Fix**: Update renderGiftItem to use the nested reservation_status object:
- Set `isReserved = !!item.reservation_status?.is_reserved`
- Extract `reservedByName = item.reservation_status?.reserved_by_name`
- Remove any isPurchased logic and UI since PublicGiftItem has no purchase tracking
- Pass `reservedByName` to ReservationButton

- [ ] #### #28: Make Reserve button interactive
**File**: `mobile/app/public/[slug].tsx`
**Location**: Around line 86-92
**Priority**: High
**Issue**: The "Reserve Item" UI is rendered as non-interactive Text so users cannot tap to reserve.
**Fix**: Replace that Text with an interactive component—either wrap in Pressable or use the existing ReservationButton component. Ensure accessibility props and visual styles are preserved.

- [ ] #### #29: Normalize API response in queryFn
**File**: `mobile/app/public/[slug].tsx`
**Location**: Around line 18-35
**Priority**: High
**Issue**: The component expects WishList.giftItems but the API returns items.
**Fix**: Update the queryFn to normalize the response:
```typescript
const resp = await apiClient.getPublicWishList(slug);
return { ...resp, giftItems: resp.items ?? [] } as WishList
```

---

### Components - ImageUpload (5 issues)

- [ ] #### #5: Add typed uploadImage method to ApiClient
**File**: `mobile/components/wish-list/ImageUpload.tsx`
**Location**: Around line 124-129
**Priority**: High
**Issue**: The component uses a fragile cast `((apiClient as any).getAuthToken?.())` to fetch an auth token which bypasses typing and can let uploads proceed unauthenticated.
**Fix**: Add a properly typed `uploadImage(formData: FormData)` method to the ApiClient class that:
- Internally calls the typed `getAuthToken()`
- Attaches Authorization when present
- Posts the form data
- Throws on non-OK responses
- Update ImageUpload.tsx to call `apiClient.uploadImage(formData)`

- [ ] #### #6: Remove debug console statements
**File**: `mobile/components/wish-list/ImageUpload.tsx`
**Location**: Line 51 and other instances
**Priority**: Medium
**Issue**: Stray console.error statements in ImageUpload.tsx.
**Fix**: Remove all raw `console.debug`/`console.error` calls and replace with the app's standard error handling/logging approach: either call the project's logger (e.g., `processLogger.error`) or surface the error to the UI via error state or onError callback.

- [ ] #### #7: Fix getFileSize error handling
**File**: `mobile/components/wish-list/ImageUpload.tsx`
**Location**: Around line 161-169
**Priority**: High
**Issue**: The getFileSize helper currently swallows errors and returns `{size: 0}`, which lets unreadable/oversized files bypass the 10MB check.
**Fix**: Update getFileSize to await `File(uri).info()` correctly and on any failure throw the caught error (or return a sentinel size >10MB) instead of returning `{size: 0}` so the caller's try/catch can handle/display the error.

- [ ] #### #8: Add useEffect to sync imageUri with currentImageUrl prop
**File**: `mobile/components/wish-list/ImageUpload.tsx`
**Location**: Around line 20-22
**Priority**: Medium
**Issue**: The local state imageUri initialized via useState with currentImageUrl won't update if currentImageUrl changes after mount.
**Fix**: Add a useEffect that watches currentImageUrl and calls `setImageUri(currentImageUrl ?? null)` so the component stays in sync when editing or prop changes occur.

- [ ] #### #33: Implement actual file size check in getFileSize helper
**File**: `mobile/components/wish-list/ImageUpload.tsx`
**Location**: Around line 88-99
**Priority**: High
**Issue**: getFileSize currently returns a constant so the 10MB check never runs.
**Fix**: Replace implementation to use expo-file-system's `FileSystem.getInfoAsync(uri, { size: true })` to retrieve and return the actual file size (in bytes). Ensure expo-file-system is installed via `npx expo install expo-file-system` before building.

---

### Components - Other (6 issues)

- [ ] #### #9: Implement Template type and API methods for TemplateSelector
**File**: `mobile/components/wish-list/TemplateSelector.tsx`
**Location**: Around line 1-38
**Priority**: High
**Issue**: The `@ts-expect-error` suppressions hide missing API surface and types.
**Fix**:
- Implement and export a Template type in `mobile/lib/api/types.ts` (matching the OpenAPI schema)
- Add `getTemplates()` and `updateWishListTemplate(wishlistId: string, templateId: string)` methods to the ApiClient
- Remove the `@ts-expect-error` comments
- Update TemplateSelector to use the real methods

- [ ] #### #30: Add missing icon mapping and tighten types
**File**: `mobile/components/ui/icon-symbol.tsx`
**Location**: Around line 8-26
**Priority**: Medium
**Issue**: The mapping misses 'bookmark.fill' and the current cast lets unmapped SymbolViewProps['name'] slip through.
**Fix**:
- Add an entry for 'bookmark.fill' to MAPPING
- Tighten typings by deriving IconMapping/IconSymbolName directly from the literal MAPPING: `infer IconSymbolName = keyof typeof MAPPING`
- Ensure compile-time errors for unmapped symbols

- [ ] #### #31: Fix type mismatch in GiftItemDisplay component
**File**: `mobile/components/wish-list/GiftItemDisplay.tsx`
**Location**: Around line 6-45
**Priority**: High
**Issue**: The GiftItem interface and isReserved/isPurchased checks mismatch the API (which uses reserved_by_user_id / purchased_by_user_id).
**Fix**:
- Import and use the shared type from `mobile/lib/types.ts` OR
- Map the incoming API shape to camelCase before rendering
- Remove or actually use the unused wishlistId prop

- [ ] #### #32: Make existingItem optional in GiftItemForm
**File**: `mobile/components/wish-list/GiftItemForm.tsx`
**Location**: Around line 13-17
**Priority**: Medium
**Issue**: existingItem is required but should be optional for create mode.
**Fix**:
- Update GiftItemFormProps to make existingItem optional: `existingItem?: GiftItem`
- Guard the deleteMutation and any access to existingItem.id by checking for existence
- Ensure delete-related UI is conditional on existingItem being present

- [ ] #### #34: Replace inline fetch with TanStack Query mutation in ReservationButton
**File**: `mobile/components/wish-list/ReservationButton.tsx`
**Location**: Around line 34-71
**Priority**: High
**Issue**: Using inline fetch instead of the app's standard API client pattern.
**Fix**:
- Add a dedicated apiClient method: `reserveGiftItem(wishlistId, giftItemId, { guestName, guestEmail })`
- Create a useMutation hook with proper onSuccess/onError handlers
- Remove manual fetch/try/catch
- Use mutation.isLoading instead of local setLoading

- [ ] #### #35: Fix onSuccess handler to use new template ID in TemplateSelector
**File**: `mobile/components/wish-list/TemplateSelector.tsx`
**Location**: Around line 39-47
**Priority**: Medium
**Issue**: The onSuccess handler is calling onTemplateChange with currentTemplateId (the old value).
**Fix**: Change the onSuccess signature to accept the mutation variables: `onSuccess: (_data, templateId) => { ... }` and call `onTemplateChange(templateId)`.

---

### API Client (4 issues)

- [ ] #### #10: Implement or remove deleteAccount method
**File**: `mobile/lib/api/api.ts`
**Location**: Around line 152-167
**Priority**: High
**Issue**: The deleteAccount method currently always throws and never performs the API call.
**Fix**: Either:
- Implement the DELETE request using `this.client.DELETE('/v1/users/me', { headers: this.getHeaders() })`
- Handle the response error appropriately
- Return void on success
- OR remove/feature-flag the method if the endpoint is not supported

- [ ] #### #11: Fix register method response type
**File**: `mobile/lib/api/api.ts`
**Location**: Around line 80-97
**Priority**: High
**Issue**: The register method wrongly treats the POST /v1/users/register response as LoginResponse and persists response.token.
**Fix**:
- Change the method signature/return type from `Promise<LoginResponse>` to `Promise<UserResponse>`
- Cast the POST result to UserResponse
- Remove the call to `this.setToken(response.token)`
- OR coordinate a server/schema change if auto-login is intended

- [ ] #### #12: Fix getReservationsByUser return type
**File**: `mobile/lib/api/api.ts`
**Location**: Around line 420-432
**Priority**: High
**Issue**: getReservationsByUser is typed to return Reservation[] but the API returns reservation_details_response with nested gift_item and wishlist.
**Fix**:
- Add a new type alias ReservationDetails in `mobile/lib/api/types.ts` that matches reservation_details_response
- Change the signature to `Promise<ReservationDetails[]>`
- Cast the parsed payload to `ReservationDetails[]`
- Ensure ReservationDetails includes the nested gift_item and wishlist shapes

- [ ] #### #36: Replace localStorage with expo-secure-store in ApiClient
**File**: `mobile/lib/api/api.ts`
**Location**: Around line 25-112
**Priority**: Critical
**Issue**: The constructor calls async loadToken() without awaiting and uses localStorage (not available in RN).
**Fix**:
- Import SecureStore
- Replace localStorage.getItem/setItem/removeItem with SecureStore.getItemAsync, setItemAsync and deleteItemAsync
- Initialize a Promise/resolve pair: `this.tokenReady` and `this.resolveTokenReady`
- Call `resolveTokenReady()` when loadToken completes
- Await `this.tokenReady` in `request<T>(...)` so requests wait for token initialization

---

### OAuth Service (2 issues)

- [ ] #### #37: Use environment variable for redirectUri
**File**: `mobile/lib/oauth-service.ts`
**Location**: Around line 49-53
**Priority**: High
**Issue**: The redirectUri currently hardcodes the native scheme.
**Fix**:
- Read the app scheme from an environment variable (e.g., `process.env.APP_SCHEME` or `EXPO_APP_SCHEME`)
- Use that value for the native field when constructing redirectUri
- If the env var is missing or empty, throw or log a fatal error and exit early

- [ ] #### #38: Replace manual OAuth flows with AuthSession.AuthRequest
**File**: `mobile/lib/oauth-service.ts`
**Location**: Around line 55-80
**Priority**: High
**Issue**: The Google and Facebook flows currently build authUrl strings and call WebBrowser.openAuthSessionAsync, then manually parse the returned URL.
**Fix**:
- Create an AuthRequest: `new AuthSession.AuthRequest({ clientId, redirectUrl, scopes, extraParams, usePKCE: true })`
- Use `request.promptAsync(discovery)` which handles PKCE, automatic state validation and code extraction
- Remove manual authUrl construction and URL parsing

---

### Types & Configuration (2 issues)

- [ ] #### #39: Fix CreateReservationRequest to use snake_case
**File**: `mobile/lib/types.ts`
**Location**: Around line 110-134
**Priority**: High
**Issue**: The CreateReservationRequest type uses camelCase but the API expects snake_case.
**Fix**:
- Update the interface to use `gift_item_id`, `guest_name`, and `guest_email` (instead of camelCase)
- Verify other request types follow the same snake_case convention
- Adjust any usages of CreateReservationRequest to set the new snake_case properties

- [ ] #### #40: Move nodeLinker config from pnpm-workspace.yaml to .npmrc
**File**: `mobile/pnpm-workspace.yaml`
**Location**: Line 1
**Priority**: Low
**Issue**: The pnpm workspace file incorrectly contains the nodeLinker: hoisted setting.
**Fix**:
- Move this configuration into `.npmrc` as `node-linker=hoisted` (or create .npmrc if missing)
- Remove the nodeLinker entry from pnpm-workspace.yaml
- If this repo is a monorepo, ensure pnpm-workspace.yaml instead defines workspace packages array

---

### Deep Linking (1 issue)

- [ ] #### #23: Fix deep-link ID extraction in _layout.tsx
**File**: `mobile/app/_layout.tsx`
**Location**: Around line 52-55
**Priority**: High
**Issue**: The deep-link logic incorrectly extracts the id via `path.split('/')[2]`, which returns "edit" for paths like "gift-items/{id}/edit".
**Fix**:
- Change the parsing to use a regex match: `/^gift-items\/([^\/]+)\/edit/`
- Assign `id = match[1]` if present
- Call `router.push(\`/gift-items/${id}/edit\`)` with proper guards and null-check

---

## API Documentation Issues (8 tasks)

### README Documentation (1 issue)

- [ ] #### #13: Add language identifiers to code blocks in api/README.md
**File**: `api/README.md`
**Location**: Around line 31-43 and 58-60
**Priority**: Low
**Issue**: Fenced code blocks lack language identifiers.
**Fix**:
- Update the directory tree block to use ` ```text ` or ` ```plaintext `
- Change the URL block to use ` ```text `
- Ensure snippets render correctly in Markdown viewers

---

### OpenAPI Schema Issues (7 issues)

- [ ] #### #14: Fix scheme-relative URLs in OpenAPI files
**Files**:
- `api/split/openapi.yaml` (line 23-24)
- `api/openapi3.json` (line 1789-1793)
- `api/openapi3.yaml` (line 1113-1114)
**Priority**: Medium
**Issue**: The servers.url currently uses scheme-relative URL ("//localhost:8080/api").
**Fix**: Update to include explicit scheme (e.g., "http://localhost:8080/api" or "https://localhost:8080/api") in all three files.

- [ ] #### #18: Fix empty pagination schema in UserReservationsResponse
**File**: `api/split/components/schemas/internal_handlers.UserReservationsResponse.yaml`
**Location**: Line 6
**Priority**: Medium
**Issue**: The pagination schema is empty.
**Fix**: Update pagination to be a typed object matching what the backend returns:
```yaml
pagination:
  type: object
  properties:
    page:
      type: integer
    limit:
      type: integer
    total:
      type: integer
    totalPages:
      type: integer
  required:
    - page
    - limit
    - total
    - totalPages
```

- [ ] #### #19: Add missing 401 response to gift-items GET endpoint
**File**: `api/split/paths/gift-items_{id}.yaml`
**Location**: Around line 63-86
**Priority**: Medium
**Issue**: The GET operation is missing a 401 Unauthorized response even though it requires BearerAuth.
**Fix**: Add a '401' response entry under responses (alongside '200', '403', '404') with description like "Authentication required or invalid".

- [ ] #### #20: Add missing 404 response to wishlists DELETE endpoint
**File**: `api/split/paths/wishlists_{id}.yaml`
**Location**: Around line 1-41
**Priority**: Medium
**Issue**: Missing 404 response to document "Wish list not found" case.
**Fix**: Add a '404' entry with the same application/json content schema used for other error responses so DELETE is consistent with GET and PUT.

- [ ] #### #21: Fix security requirements for reservation endpoints
**File**: `api/split/paths/wishlists_{wishlistId}_gift-items_{itemId}_reservation.yaml`
**Location**: Around line 1-122
**Priority**: High
**Issue**: The POST (CreateReservation) and DELETE (CancelReservation) operations require BearerAuth which contradicts documented guest flows.
**Fix**: Update both operations to support optional authentication by replacing the single-entry security with an array allowing either BearerAuth or no-security (include an empty security object alongside the BearerAuth entry).

- [ ] #### #22: Fix security requirements for gift items GET endpoint
**File**: `api/split/paths/wishlists_{wishlistId}_gift-items.yaml`
**Location**: Around line 1-57
**Priority**: High
**Issue**: The GET operation declares required auth via "security: - BearerAuth: []" but the handler allows unauthenticated reads for public wishlists.
**Fix**:
- Remove the operation-level "security: - BearerAuth: []" (or make it optional)
- Add a '401' response entry matching the structure of existing error responses

---

## Backend Issues (3 tasks)

- [x] #### #15: Fix docs package import in backend/cmd/server/main.go
**File**: `backend/cmd/server/main.go`
**Location**: Around line 26-32
**Priority**: Critical
**Issue**: The build is failing because `_ "wish-list/docs"` cannot be resolved.
**Fix**: Either:
- Commit the generated docs package into the repo, OR
- Add a pre-build step in CI to generate them (run `swag init`)
- Keep the blank import to register docs with echo-swagger
- OR remove it if embedded docs are not wanted
**Status**: ✅ **COMPLETED** - Added Swagger docs generation to CI workflow

- [ ] #### #16: Update redis module version in go.mod
**File**: `backend/go.mod`
**Location**: Line 6
**Priority**: Low
**Issue**: Outdated redis module version.
**Fix**:
- Update `github.com/redis/go-redis/v9` from v9.17.2 to v9.17.3
- Run `go get` or `go mod tidy` to refresh the lockfile
- Ensure go.sum is updated
- Run tests to verify no regressions

- [ ] #### #17: Fix paths in Makefile swagger targets
**File**: `Makefile`
**Location**: Around line 248-266
**Priority**: Medium
**Issue**: The swagger-split and swagger-preview targets reference wrong paths.
**Fix**:
- **swagger-split**: Update existence check from `backend/docs/openapi3.yaml` to `api/openapi3.yaml` and correct echo message to state files were split into `api/split/`
- **swagger-preview**: Update to use `api/openapi3.yaml` in both the if-check and the pnpm preview command

---

## Priority Summary

### Critical Priority (4 issues)
- [x] #15: Fix docs package import in backend/cmd/server/main.go ✅
- [ ] #24: Clear session after account deletion
- [ ] #26: Store auth token after login
- [ ] #36: Replace localStorage with expo-secure-store in ApiClient

### High Priority (19 issues)
- [ ] #1: Fix email field persistence in profile update
- [ ] #3: Fix FlatList data binding in public wishlist
- [ ] #4: Fix reservation status display in public wishlist
- [ ] #5: Add typed uploadImage method to ApiClient
- [ ] #7: Fix getFileSize error handling in ImageUpload
- [ ] #9: Implement Template type and API methods for TemplateSelector
- [ ] #10: Implement or remove deleteAccount method in ApiClient
- [ ] #11: Fix register method response type in ApiClient
- [ ] #12: Fix getReservationsByUser return type in ApiClient
- [ ] #21: Fix security requirements for reservation endpoints
- [ ] #22: Fix security requirements for gift items GET endpoint
- [ ] #23: Fix deep-link ID extraction in _layout.tsx
- [ ] #28: Make Reserve button interactive in public wishlist
- [ ] #29: Normalize API response in public wishlist queryFn
- [ ] #31: Fix type mismatch in GiftItemDisplay component
- [ ] #33: Implement actual file size check in getFileSize helper
- [ ] #34: Replace inline fetch with TanStack Query mutation in ReservationButton
- [ ] #37: Use environment variable for redirectUri in oauth-service.ts
- [ ] #38: Replace manual OAuth flows with AuthSession.AuthRequest
- [ ] #39: Fix CreateReservationRequest to use snake_case in types.ts

### Medium Priority (14 issues)
- [ ] #2: Fix zero price display logic in public wishlist
- [ ] #6: Remove debug console statements in ImageUpload
- [ ] #8: Add useEffect to sync imageUri with currentImageUrl prop
- [ ] #14: Fix scheme-relative URLs in OpenAPI files
- [ ] #17: Fix paths in Makefile swagger targets
- [ ] #18: Fix empty pagination schema in UserReservationsResponse
- [ ] #19: Add missing 401 response to gift-items GET endpoint
- [ ] #20: Add missing 404 response to wishlists DELETE endpoint
- [ ] #25: Fix Avatar rendering with conditional logic in profile.tsx
- [ ] #27: Fix linking config in linking.ts
- [ ] #30: Add missing icon mapping and tighten types in icon-symbol.tsx
- [ ] #32: Make existingItem optional in GiftItemForm
- [ ] #35: Fix onSuccess handler to use new template ID in TemplateSelector

### Low Priority (3 issues)
- [ ] #13: Add language identifiers to code blocks in api/README.md
- [ ] #16: Update redis module version in go.mod
- [ ] #40: Move nodeLinker config from pnpm-workspace.yaml to .npmrc

---

## Next Steps

1. **Start with Critical Priority issues** - These are blocking or security-critical
2. **Address High Priority issues** - Core functionality and type safety
3. **Work through Medium Priority issues** - Important but not blocking
4. **Complete Low Priority issues** - Polish and documentation

## Notes

- All issues are tracked in the task management system (tasks #1-#40)
- Use `TaskUpdate` to mark tasks as in_progress when starting work
- Use `TaskUpdate` to mark tasks as completed when finished
- This document should be updated as issues are resolved
