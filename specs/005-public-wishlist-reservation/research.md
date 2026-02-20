# Research: Public Wishlist with Guest Reservation

**Feature**: 005-public-wishlist-reservation
**Date**: 2026-02-20

## R-001: Backend API Readiness

**Decision**: All required backend endpoints exist and are production-ready.

**Rationale**: Verified against `/api/openapi3.yaml`. The following endpoints cover all feature requirements:

| Endpoint | Method | Purpose | Auth Required |
|----------|--------|---------|---------------|
| `/public/wishlists/{slug}` | GET | Fetch public wishlist metadata | No |
| `/public/wishlists/{slug}/gift-items` | GET | Fetch paginated gift items | No |
| `/public/reservations/list/{slug}/item/{itemId}` | GET | Check item reservation status | No |
| `/reservations/wishlist/{wishlistId}/item/{itemId}` | POST | Create reservation (guest or auth) | Optional |
| `/reservations/wishlist/{wishlistId}/item/{itemId}` | DELETE | Cancel reservation (guest via token) | Optional |
| `/guest/reservations` | GET | Fetch guest reservations by token | No (token in query) |

**Alternatives Considered**: Building a BFF (Backend for Frontend) layer. Rejected because the existing API already serves the frontend's needs directly.

## R-002: i18n Implementation Approach

**Decision**: Extend existing i18next setup with new translation keys for public wishlist pages.

**Rationale**: The frontend already has a fully configured i18next setup with:
- `i18next` 25.8, `react-i18next` 16.5, `i18next-browser-languagedetector` 8.2
- Provider in `providers.tsx`, language switcher in Header
- Russian and English JSON files at `src/i18n/locales/{en,ru}.json`
- Fallback language: Russian (`ru`)
- Detection order: localStorage → navigator → htmlTag

No new libraries needed. Simply add translation keys under new prefixes.

**Alternatives Considered**:
- `next-intl` (route-based i18n): Rejected because the app already uses i18next and route-based locale prefixes would break existing URLs.
- Separate translation files per feature: Rejected for simplicity; single file per locale is sufficient at current scale (~100 keys).

## R-003: Independent Styling Architecture

**Decision**: Create a separate `public-wishlist/` component directory with self-contained Tailwind classes. Use CSS custom properties for theme-able values.

**Rationale**: The spec requires visual independence from the main site and future theming capability. By:
1. Isolating public wishlist components in their own directory
2. Using Tailwind utility classes scoped to these components
3. Defining a small set of CSS variables (e.g., `--wishlist-primary`, `--wishlist-bg`) that can be overridden per theme
4. Keeping the existing shadcn/ui primitives (Card, Badge, Button, Dialog) but applying different style variants

This enables future occasion-based theming (birthday, wedding, New Year) without affecting the landing page.

**Alternatives Considered**:
- CSS Modules per component: Rejected because the project uses Tailwind exclusively.
- Separate Tailwind config for public pages: Rejected as over-engineering for current scope.
- Theme context with React Context API: Deferred to the future theming feature.

## R-004: Guest Reservation Token Storage Strategy

**Decision**: Store reservation tokens in localStorage as a JSON array, keyed by `guest_reservations`.

**Rationale**: The existing `GuestReservationDialog.tsx` already uses this pattern (line 94-105). Each reservation entry stores:
```json
{
  "itemId": "uuid",
  "itemName": "Gift Name",
  "reservationToken": "token-string",
  "reservedAt": "ISO-date",
  "guestName": "John",
  "guestEmail": "john@example.com",
  "wishlistId": "uuid",
  "wishlistSlug": "birthday-2026"
}
```

The `/my/reservations` page reads all stored tokens and calls `GET /guest/reservations?token=X` to get server-verified status for each.

**Alternatives Considered**:
- Session cookies: Rejected because cross-domain architecture prevents cookie sharing.
- IndexedDB: Over-engineering for simple token storage (<1KB per entry).
- No client-side storage (server-only via email lookup): Rejected because it requires authentication or email verification flow.

## R-005: MyReservations Component Refactoring

**Decision**: Rewrite `MyReservations.tsx` to use the `apiClient` methods instead of raw `fetch()` calls with incorrect URLs.

**Rationale**: Current implementation has bugs:
1. Calls `/api/auth/me` — this endpoint doesn't exist (correct: the auth check happens via `useAuthRedirect` in the parent page)
2. Calls `/api/users/me/reservations` — doesn't exist (correct: `GET /reservations/user` for authenticated, `GET /guest/reservations?token=X` for guests)
3. Calls `/api/reservations/{id}/cancel` — doesn't exist (correct: `DELETE /reservations/wishlist/{wishlistId}/item/{itemId}`)
4. Reads a single `reservationToken` from localStorage — should read the `guest_reservations` array

The component needs to:
1. Read all tokens from the `guest_reservations` localStorage array
2. Call `apiClient.getGuestReservations(token)` for each unique token
3. Aggregate and display results
4. Support cancellation via `apiClient.cancelReservation(wishlistId, itemId, { reservation_token: token })`

**Alternatives Considered**: Creating a new component from scratch. Rejected because the existing component has the correct UI structure; only the data-fetching logic needs fixing.

## R-006: Public Wishlist Page Refactoring Scope

**Decision**: Extract the inline rendering from `page.tsx` into reusable components and add i18n, while keeping the existing TanStack Query data-fetching pattern.

**Rationale**: The current `/public/[slug]/page.tsx` (250 lines) mixes data fetching, loading/error states, and item rendering in a single file. Extracting into:
- `PublicWishlistPage.tsx` — main orchestration component
- `WishlistHeader.tsx` — title, occasion, description, badges
- `GiftItemCard.tsx` — individual item card with reserve button

This improves testability (each component testable independently), i18n integration (each component uses `useTranslation`), and future theming (swap card designs without touching page logic).

The existing `GuestReservationDialog.tsx` is well-structured and only needs i18n translation wrapping.

**Alternatives Considered**: Reusing `WishListDisplay.tsx` and `GiftItemDisplay.tsx`. Rejected because those components have a different interface contract (they expect `gift_items` nested in the wishlist object) and are tightly coupled to an older data shape. The new components align with the actual API response shape.
