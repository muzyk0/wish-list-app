# GitHub PR Issues - Todo List

> **Sources**:
> - Pull Request #8 "Add complete mobile app implementation with Expo Router"
> - Pull Request #7 "Add complete frontend implementation with Next.js"
> **Date Created**: 2026-01-31
> **Last Updated**: 2026-01-31
> **Total Issues**: 62 (Mobile: 40, Frontend: 22)

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

- [x] #### #16: Update redis module version in go.mod
**File**: `backend/go.mod`
**Location**: Line 6
**Priority**: Low
**Issue**: Outdated redis module version.
**Fix**:
- Update `github.com/redis/go-redis/v9` from v9.17.2 to v9.17.3
- Run `go get` or `go mod tidy` to refresh the lockfile
- Ensure go.sum is updated
- Run tests to verify no regressions
**Status**: ✅ **COMPLETED** - Updated redis and other dependencies

- [ ] #### #17: Fix paths in Makefile swagger targets
**File**: `Makefile`
**Location**: Around line 248-266
**Priority**: Medium
**Issue**: The swagger-split and swagger-preview targets reference wrong paths.
**Fix**:
- **swagger-split**: Update existence check from `backend/docs/openapi3.yaml` to `api/openapi3.yaml` and correct echo message to state files were split into `api/split/`
- **swagger-preview**: Update to use `api/openapi3.yaml` in both the if-check and the pnpm preview command

---

## Frontend Issues (22 tasks)

### Security & Configuration (2 issues)

- [ ] #### #41: Fix JWT_SECRET exposed as NEXT_PUBLIC_ environment variable
**File**: `frontend/.env.example`
**Location**: Around line 5-6
**Priority**: Critical
**Issue**: The JWT_SECRET is exposed as NEXT_PUBLIC_JWT_SECRET which bundles it in client-side code, making it accessible to XSS attacks.
**Fix**:
- Rename `NEXT_PUBLIC_JWT_SECRET` to `JWT_SECRET` in .env.example
- Update server-side code to read `process.env.JWT_SECRET` instead
- Ensure the secret is only accessed in API routes, Server Components, or middleware
- Never reference in client-side code or bundled imports
- Update README/docs referencing NEXT_PUBLIC_JWT_SECRET

- [ ] #### #42: Replace localStorage JWT storage with httpOnly cookies
**File**: `frontend/src/lib/api.ts`
**Location**: Around line 23-92
**Priority**: Critical
**Issue**: JWTs are persisted in localStorage (constructor, setToken, logout), which is vulnerable to XSS attacks.
**Fix**:
- Remove localStorage read/write calls from constructor, setToken and logout
- Change login/register to expect backend to set httpOnly cookie
- OR store access token in module-scoped in-memory variable with secure refresh-token flow
- Adjust request() to omit Authorization from localStorage
- Include `credentials: 'include'` when using cookies
- Coordinate backend to set Set-Cookie and implement refresh endpoints

---

### SSR & React Patterns (7 issues)

- [ ] #### #43: Fix window.location.origin SSR crash in WishListDisplay
**File**: `frontend/src/components/wish-list/WishListDisplay.tsx`
**Location**: Around line 109-124
**Priority**: High
**Issue**: Direct access to window.location.origin in render crashes during SSR.
**Fix**:
- Compute safe shareUrl before render using `typeof window !== 'undefined'` guard
- OR use useMemo/useState with useEffect
- Guard navigator.clipboard with `typeof navigator !== 'undefined' && navigator.clipboard`
- Update all window.location.origin references to use new shareUrl variable

- [ ] #### #44: Add 'use client' directive to GiftItemDisplay component
**File**: `frontend/src/components/wish-list/GiftItemDisplay.tsx`
**Location**: Line 1
**Priority**: High
**Issue**: Component with onClick handlers missing 'use client' directive won't work in App Router.
**Fix**: Add `'use client'` as the very first line of the file (above all imports)

- [ ] #### #45: Fix missing cleanup in redirect timeout (page.tsx)
**File**: `frontend/src/app/page.tsx`
**Location**: Around line 11-22
**Priority**: High
**Issue**: handleMobileRedirect sets timeout but never clears it, risking memory leak or unexpected navigation on unmount.
**Fix**:
- Store timeout ID in ref or local variable
- Wrap function in useCallback if appropriate
- Add useEffect cleanup that calls clearTimeout(timeoutId) on unmount
- Simplify visibility check to `document.visibilityState !== 'hidden'`

- [ ] #### #46: Fix missing cleanup in useAuthRedirect hook
**File**: `frontend/src/hooks/useAuthRedirect.ts`
**Location**: Around line 17-36
**Priority**: High
**Issue**: useEffect missing cleanup to avoid state updates after unmount.
**Fix**:
- Create AbortController
- Pass signal to fetch('/api/auth/me')
- In catch handler distinguish AbortError (ignore) from other errors
- Only call setIsAuthenticated when not aborted
- Return cleanup function that calls controller.abort()

- [ ] #### #47: Fix useFormField hook context null handling
**File**: `frontend/src/components/ui/form.tsx`
**Location**: Around line 44-65
**Priority**: High
**Issue**: Hook assumes contexts are truthy and reads fields before null check.
**Fix**:
- Change FormFieldContext and FormItemContext default values to null
- Assert contexts are not null in useFormField before accessing properties
- Throw clear errors when used outside providers
- Read fieldContext.name and itemContext.id only after null checks

- [ ] #### #48: Add forwardRef to Input component
**File**: `frontend/src/components/ui/input.tsx`
**Location**: Around line 5-18
**Priority**: Medium
**Issue**: Without forwarded ref, consumers can't focus the input for form validation.
**Fix**: Use `React.forwardRef<HTMLInputElement, React.ComponentPropsWithoutRef<'input'>>` pattern with `displayName = 'Input'`

- [ ] #### #49: Add forwardRef to Textarea component
**File**: `frontend/src/components/ui/textarea.tsx`
**Location**: Around line 1-16
**Priority**: Medium
**Issue**: Missing ref forwarding for form integration.
**Fix**: Use `React.forwardRef<HTMLTextAreaElement, React.ComponentPropsWithoutRef<'textarea'>>` pattern with `displayName = 'Textarea'`

---

### UI Components (5 issues)

- [ ] #### #50: Fix Button component default type
**File**: `frontend/src/components/ui/button.tsx`
**Location**: Around line 39-58
**Priority**: High
**Issue**: No safe default type so native buttons act as submit buttons in forms.
**Fix**:
- When asChild is false and Comp === 'button', merge `type="button"` as default
- Preserve existing explicit type props
- Keep Slot children untouched

- [ ] #### #51: Fix GiftItemImage className injection and positioning
**File**: `frontend/src/components/wish-list/GiftItemImage.tsx`
**Location**: Around line 9-31
**Priority**: Medium
**Issue**: Injects "undefined" when className missing and uses next/image fill without positioned parent.
**Fix**:
- Use cn utility to safely merge classes (avoid template literal)
- Add "relative" to wrapper div that contains Image
- Update both empty-src branch and main wrapper

- [ ] #### #52: Fix Image component size mismatch in public wishlist
**File**: `frontend/src/app/public/[slug]/page.tsx`
**Location**: Around line 173-179
**Priority**: Medium
**Issue**: Image requests 16×16px while CSS displays 64×64, causing blur.
**Fix**: Update Image props to `width={64} height={64}` to match .w-16 .h-16 CSS

- [ ] #### #53: Remove unused _redirectAttempted state
**File**: `frontend/src/app/public/[slug]/page.tsx`
**Location**: Line 49
**Priority**: Low
**Issue**: Unused state declaration clutters code.
**Fix**: Delete the entire line `const [_redirectAttempted, _setRedirectAttempted] = useState(false)`

- [ ] #### #54: Fix invalid 'path fill' property in SVG style
**File**: `frontend/src/stories/Configure.mdx`
**Location**: Around line 19-33
**Priority**: Medium
**Issue**: Inline style object in RightArrow contains invalid property 'path fill'.
**Fix**: Remove that entry and set `fill="currentColor"` on path element instead

---

### Package Dependencies (3 issues)

- [ ] #### #55: Move class-variance-authority and clsx to dependencies
**File**: `frontend/package.json`
**Location**: Around line 53-54
**Priority**: High
**Issue**: Runtime packages in devDependencies cause production failures.
**Fix**:
- Remove "class-variance-authority" and "clsx" from devDependencies
- Add them under dependencies with same versions (^0.7.1 and ^2.1.1)
- Run pnpm install after update

- [ ] #### #56: Move @radix-ui/react-slot and lucide-react to dependencies
**File**: `frontend/package.json`
**Location**: Around line 40 and 55
**Priority**: Medium
**Issue**: Runtime UI dependencies incorrectly placed in devDependencies.
**Fix**:
- Move "@radix-ui/react-slot": "^1.2.4" to dependencies
- Move "lucide-react": "^0.562.0" to dependencies
- Run pnpm install after update

- [ ] #### #57: Move postcss to devDependencies
**File**: `frontend/package.json`
**Location**: Line 29
**Priority**: Low
**Issue**: Build-time tool should be in devDependencies.
**Fix**: Move "postcss": "^8.5.6" from dependencies to devDependencies

---

### Configuration & Validation (4 issues)

- [ ] #### #58: Fix /frontend entry in .gitignore
**File**: `frontend/.gitignore`
**Location**: Around line 20-22
**Priority**: Medium
**Issue**: /frontend entry targets non-existent frontend/frontend/ path.
**Fix**: Remove the "/frontend" line from .gitignore (keep /build)

- [ ] #### #59: Add .env*.local pattern to .gitignore
**File**: `frontend/.gitignore`
**Location**: After line 42
**Priority**: Low
**Issue**: Missing pattern to prevent commit of local env files with secrets.
**Fix**: Add `.env*.local` pattern to gitignore

- [ ] #### #60: Use dynamic locale instead of hardcoded lang="en"
**File**: `frontend/src/app/layout.tsx`
**Location**: Around line 29-31
**Priority**: Medium
**Issue**: HTML tag hard-codes lang="en" but app supports i18n.
**Fix**:
- Retrieve current locale from i18n getter, Next.js params, or DEFAULT_LOCALE constant
- Pass to `<html>` element as `lang={locale}`
- Fallback to default locale (e.g., "ru") when none available

- [ ] #### #61: Add URL encoding for reservation token
**File**: `frontend/src/components/wish-list/MyReservations.tsx`
**Location**: Around line 47-51
**Priority**: High
**Issue**: Raw token from localStorage inserted into URL breaks for special characters.
**Fix**: Use `encodeURIComponent(token)` before interpolating into fetch URL

---

### Miscellaneous (1 issue)

- [ ] #### #62: Add missing i18n support to auth pages
**File**: `frontend/src/app/auth/login/page.tsx` and `frontend/src/app/auth/register/page.tsx`
**Location**: Around line 23-31
**Priority**: Medium
**Issue**: Hardcoded English strings don't match i18n support mentioned in PR.
**Fix**:
- Add 'use client' directive
- Import useTranslation from react-i18next
- Replace hardcoded strings with t('auth.login.title'), t('auth.login.description'), etc.
- Add corresponding translation keys to i18n files

---

## Priority Summary

### Critical Priority (7 issues)
- [x] #15: Fix docs package import in backend/cmd/server/main.go ✅
- [x] #63: Fix API_BASE_URL missing /api prefix ✅
- [x] #64: Update authentication API paths ✅
- [x] #65: Update profile API paths ✅
- [x] #66: Fix types.ts schema imports ✅
- [x] #67: Update wishlist and gift item API paths ✅
- [ ] #24: Clear session after account deletion (Mobile)
- [ ] #26: Store auth token after login (Mobile)
- [ ] #36: Replace localStorage with expo-secure-store in ApiClient (Mobile)
- [ ] #41: Fix JWT_SECRET exposed as NEXT_PUBLIC_ environment variable (Frontend)
- [ ] #42: Replace localStorage JWT storage with httpOnly cookies (Frontend)

### High Priority (32 issues)

**Mobile - New Completion Tasks (5 issues)**:
- [ ] #68: Fix type narrowing for optional fields (18 TypeScript errors)
- [ ] #69: Add missing PublicGiftItem and PublicWishList types
- [ ] #71: Add null checks for auth token
- [ ] #72: Add missing API client methods (uploadImage, guest reservations, etc.)
- [ ] #73: Create gift item create screen

**Mobile - Original PR Issues (19 issues)**:
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

**Frontend (8 issues)**:
- [ ] #43: Fix window.location.origin SSR crash in WishListDisplay
- [ ] #44: Add 'use client' directive to GiftItemDisplay component
- [ ] #45: Fix missing cleanup in redirect timeout (page.tsx)
- [ ] #46: Fix missing cleanup in useAuthRedirect hook
- [ ] #47: Fix useFormField hook context null handling
- [ ] #50: Fix Button component default type
- [ ] #55: Move class-variance-authority and clsx to dependencies
- [ ] #61: Add URL encoding for reservation token

### Medium Priority (26 issues)

**Mobile - New Completion Tasks (5 issues)**:
- [ ] #70: Fix field naming mismatch (guest_name → guestName)
- [ ] #74: Implement reservation details screen
- [ ] #75: Add image upload functionality
- [ ] #76: Fix or remove Template functionality
- [ ] #81: Implement search/discover functionality
- [ ] #84: Add error handling screens

**Mobile - Original PR Issues (14 issues)**:
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

**Frontend (7 issues)**:
- [ ] #48: Add forwardRef to Input component
- [ ] #49: Add forwardRef to Textarea component
- [ ] #51: Fix GiftItemImage className injection and positioning
- [ ] #52: Fix Image component size mismatch in public wishlist
- [ ] #54: Fix invalid 'path fill' property in SVG style
- [ ] #56: Move @radix-ui/react-slot and lucide-react to dependencies
- [ ] #58: Fix /frontend entry in .gitignore
- [ ] #60: Use dynamic locale instead of hardcoded lang="en"
- [ ] #62: Add missing i18n support to auth pages

### Low Priority (13 issues)

**Mobile - New Completion Tasks (7 issues)**:
- [ ] #77: Add pagination support to list endpoints
- [ ] #78: Implement filtering and sorting
- [ ] #79: Standardize API response formats
- [ ] #80: Add batch operations
- [ ] #82: Add settings screen
- [ ] #83: Create onboarding flow
- [ ] #85: Improve loading and empty states

**Original PR Issues (6 issues)**:
- [ ] #13: Add language identifiers to code blocks in api/README.md (Mobile)
- [x] #16: Update redis module version in go.mod ✅ (Backend)
- [ ] #40: Move nodeLinker config from pnpm-workspace.yaml to .npmrc (Mobile)
- [ ] #53: Remove unused _redirectAttempted state (Frontend)
- [ ] #57: Move postcss to devDependencies (Frontend)
- [ ] #59: Add .env*.local pattern to .gitignore (Frontend)

---

## Mobile API Client Errors (RESOLVED → NEW TASKS)

~~⛔ **CRITICAL**: Mobile app is completely broken~~
✅ **FIXED**: Critical API contract issues resolved (Tasks #63-#67 completed)
⚠️ **REMAINING**: 27 TypeScript errors, missing features, type improvements needed

**Progress**: 42 errors → 27 errors (-36% reduction)
**Status**: App compiles and core functionality works, polish needed

**See detailed completion plan**: `claudedocs/mobile-app-completion-plan.md`

### Critical Fixes Completed (Issues #63-#67)

- [x] #### #63: Fix API_BASE_URL missing /api prefix
**File**: `mobile/lib/api/api.ts:20`
**Priority**: Critical (Blocking)
**Issue**: Base URL `http://10.0.2.2:8080` should be `http://10.0.2.2:8080/api`
**Impact**: All API requests going to wrong URLs, resulting in 404 errors
**Fix**: Update API_BASE_URL constant to include `/api` suffix

- [x] #### #64: Update authentication API paths
**File**: `mobile/lib/api/api.ts`
**Priority**: Critical (Blocking)
**Status**: ✅ **COMPLETED**
**Fix Applied**: Updated paths from `/v1/users/*` to `/auth/*`

- [x] #### #65: Update profile API paths
**File**: `mobile/lib/api/api.ts`
**Priority**: Critical (Blocking)
**Status**: ✅ **COMPLETED**
**Fix Applied**: Updated from `/v1/users/me` to `/protected/profile`

- [x] #### #66: Fix types.ts schema imports
**File**: `mobile/lib/api/types.ts`
**Priority**: Critical (Blocking)
**Status**: ✅ **COMPLETED** - Resolved 14 type errors
**Fix Applied**: Replaced manual aliases with generated schema imports

- [x] #### #67: Update wishlist and gift item API paths
**File**: `mobile/lib/api/api.ts`
**Priority**: High (Blocking)
**Status**: ✅ **COMPLETED**
**Fix Applied**: Removed `/v1` prefix, changed `/items/*` to `/gift-items/*`, updated endpoints

---

### Phase 1: Remaining Critical Fixes (Issues #68-#72)

- [ ] #### #68: Fix type narrowing for optional fields (18 TypeScript errors)
**Files**: Multiple screens and components
**Priority**: High
**Issue**: Optional fields accessed without null checks causing 18 compilation errors
- `item.view_count` is possibly undefined
- `user.email` is possibly undefined
- `item.priority` is possibly undefined
- And 15 more similar errors
**Fix**: Add optional chaining `?.` or default values `?? 0`
**Impact**: Resolves 18 of remaining 27 TypeScript errors

- [ ] #### #69: Add missing PublicGiftItem and PublicWishList types
**File**: `mobile/lib/api/types.ts`
**Priority**: High
**Issue**: Types referenced in public wishlist screen but not defined
**Fix**: Export types from schema:
```typescript
export type PublicWishList = components['schemas']['wish-list_internal_services.WishListOutput'];
export type PublicGiftItem = components['schemas']['wish-list_internal_services.GiftItemOutput'];
```

- [ ] #### #70: Fix field naming mismatch in ReservationButton
**File**: `mobile/components/wish-list/ReservationButton.tsx:38`
**Priority**: Medium
**Issue**: Using `guest_name` instead of `guestName` (snake_case vs camelCase)
**Fix**: Change field to match schema: `guestName`

- [ ] #### #71: Add null checks for auth token
**File**: `mobile/lib/api/api.ts:76, 96`
**Priority**: High
**Issue**: Auth response token is optional but setToken assumes it exists
**Fix**: Add null checks before calling `setToken()`

- [ ] #### #72: Add missing API client methods
**File**: `mobile/lib/api/api.ts`
**Priority**: High
**Issue**: Several backend endpoints not exposed in API client
**Fix**: Add methods for:
1. `uploadImage(file: File)` - POST /s3/upload
2. `getGuestReservations(token: string)` - GET /reservations/guest
3. `getReservationStatus(slug: string, itemId: string)` - GET /public/wishlists/{slug}/gift-items/{itemId}/reservation-status

---

### Phase 2: Essential Features (Issues #73-#76)

- [ ] #### #73: Create gift item create screen
**Path**: `mobile/app/gift-items/create.tsx`
**Priority**: High
**Issue**: Only edit screen exists, no way to create new gift items
**Fix**: Create new screen with form for adding gift items to wishlists
**Components needed**: GiftItemForm (already exists, make reusable)

- [ ] #### #74: Implement reservation details screen
**Path**: `mobile/app/reservations/[id]/index.tsx`
**Priority**: Medium
**Issue**: No way to view individual reservation details
**Fix**: Create detail screen showing:
- Gift item information
- Wishlist owner details
- Reservation date and status
- Cancel reservation button

- [ ] #### #75: Add image upload functionality
**Files**: `mobile/components/wish-list/ImageUpload.tsx`, `mobile/lib/api/api.ts`
**Priority**: High
**Issue**: ImageUpload component doesn't actually upload to S3
**Fix**:
1. Add uploadImage API method
2. Integrate with ImageUpload component
3. Handle upload progress and errors

- [ ] #### #76: Fix or remove Template functionality
**Files**: `mobile/components/wish-list/TemplateSelector.tsx`, `mobile/lib/api/api.ts`
**Priority**: Medium
**Issue**: 4 TypeScript errors - Template type and API methods missing
**Fix**: Either:
- Option A: Implement if backend supports templates
- Option B: Remove TemplateSelector component if not needed

---

### Phase 3: API Improvements (Issues #77-#80)

- [ ] #### #77: Add pagination support to list endpoints
**Files**: Backend handlers, mobile API client
**Priority**: Medium
**Issue**: No pagination for lists that could grow large (wishlists, gift items, reservations)
**Fix**: Add query parameters `?page=1&limit=20` to:
- GET /wishlists
- GET /wishlists/{id}/gift-items
- GET /reservations

- [ ] #### #78: Implement filtering and sorting
**Files**: Backend handlers, mobile API client
**Priority**: Low
**Issue**: No way to filter or sort results
**Fix**: Add query parameters:
- `?sort=created_at&order=desc`
- `?is_public=true`
- `?status=active`

- [ ] #### #79: Standardize API response formats
**Files**: Backend handlers
**Priority**: Low
**Issue**: Inconsistent response wrapping (some wrapped, some direct)
**Fix**: Standardize all responses:
```json
{
  "data": {...},
  "meta": {"timestamp": "..."}
}
```

- [ ] #### #80: Add batch operations
**Files**: Backend handlers, mobile API client
**Priority**: Low
**Issue**: Can only delete/update one item at a time
**Fix**: Add batch endpoints:
- DELETE /gift-items?ids=1,2,3
- PATCH /gift-items/batch

---

### Phase 4: Polish & UX (Issues #81-#85)

- [ ] #### #81: Implement search/discover functionality
**File**: `mobile/app/(tabs)/explore.tsx`
**Priority**: Medium
**Issue**: Explore screen exists but not implemented
**Fix**: Add:
- Search bar for finding public wishlists
- Browse popular/recent wishlists
- Filter by occasion/category

- [ ] #### #82: Add settings screen
**Path**: `mobile/app/settings/index.tsx`
**Priority**: Low
**Issue**: No settings screen for user preferences
**Fix**: Create settings screen with:
- Notification preferences
- Privacy settings
- Theme selection
- Language selection

- [ ] #### #83: Create onboarding flow
**Path**: `mobile/app/onboarding/*`
**Priority**: Low
**Issue**: No onboarding for new users
**Fix**: Create welcome screens:
- App introduction
- Feature highlights
- Tutorial walkthrough

- [ ] #### #84: Add error handling screens
**Files**: Error boundaries, 404 pages
**Priority**: Medium
**Issue**: No proper error handling UI
**Fix**: Add:
- Global error boundary
- 404 not found screen
- Network error screen
- Permission denied screen

- [ ] #### #85: Improve loading and empty states
**Files**: All list screens
**Priority**: Low
**Issue**: Generic loading indicators
**Fix**: Add:
- Skeleton loaders for lists
- Empty state illustrations
- Loading state animations
- Pull-to-refresh indicators

---

## Next Steps

1. ~~**Fix blocking mobile API client issues FIRST** (#63-#67)~~ ✅ **COMPLETED**
2. **Start with Critical Priority issues** (6 issues) - Security vulnerabilities and blocking problems
   - **Frontend Security**: #41 JWT_SECRET exposure and #42 localStorage XSS vulnerability are urgent
   - **Mobile Auth**: #24 session cleanup, #26 token storage, #36 secure storage
2. **Address High Priority issues** (27 issues) - Core functionality and type safety
   - **Frontend**: SSR crashes (#43), React patterns (#44-47), component defaults (#50), dependencies (#55), URL encoding (#61)
   - **Mobile**: Form persistence, API types, deep linking, reservation functionality
3. **Work through Medium Priority issues** (21 issues) - Important but not blocking
   - **Frontend**: ref forwarding, i18n support, image sizing, dependency organization
   - **Mobile**: display logic, error handling, API documentation
4. **Complete Low Priority issues** (5 issues) - Polish and documentation

## Notes

### Task Management
- All issues are tracked in the task management system (tasks #1-#85)
- Mobile PR issues: #1-#40 (from PR #8)
- Frontend issues: #41-#62 (from PR #7)
- Mobile completion: #63-#85 (new implementation tasks)
- Backend issues: #15-#17, #77-#80 (from PR #8 + API improvements)
- Use `TaskUpdate` to mark tasks as in_progress when starting work
- Use `TaskUpdate` to mark tasks as completed when finished
- This document should be updated as issues are resolved

### Sources
- **PR #8** (Mobile): 40 issues from CodeRabbit AI review
- **PR #7** (Frontend): 22 issues from CodeRabbit AI review (filtered from 17 actionable + 26 nitpicks)
- **Mobile Completion** (#63-#85): 23 tasks from comprehensive analysis after API fix
- **Completed**: 7 issues (#15, #16, #63-#67)

### Priority Focus
- **Critical issues are security-related** - JWT exposure, XSS vulnerabilities, authentication flaws
- **High priority issues affect functionality** - SSR crashes, missing client directives, wrong dependencies
- **Medium priority issues improve quality** - ref forwarding, i18n, proper patterns
- **Low priority issues are polish** - documentation, cleanup, style improvements
