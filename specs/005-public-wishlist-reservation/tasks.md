# Tasks: Public Wishlist with Guest Reservation

**Input**: Design documents from `/specs/005-public-wishlist-reservation/`
**Prerequisites**: spec.md (user stories), data-model.md (frontend entities), contracts/frontend-api.md (API client methods), research.md (decisions), quickstart.md (test scenarios)

**Tests**: Not explicitly requested in the feature specification. Test tasks are omitted. Constitution CR-002 (Test-First) applies to backend; this feature is frontend-only with existing backend endpoints.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3, US4)
- Include exact file paths in descriptions

## Path Conventions

- **Web app (frontend-only feature)**: `frontend/src/`
- All paths relative to repository root

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: API client extensions, type exports, i18n keys, and shared utilities needed by all user stories

- [ ] T001 [P] Add new type exports (`ReservationDetailsResponse`, `GiftItemSummary`, `WishListSummary`, `CancelReservationRequest`, `ReservationStatusResponse`) in `frontend/src/lib/api/types.ts`
- [ ] T002 [P] Add `getGuestReservations(token)` method to API client in `frontend/src/lib/api/client.ts`
- [ ] T003 [P] Add `cancelReservation(wishlistId, itemId, data)` method to API client in `frontend/src/lib/api/client.ts`
- [ ] T004 [P] Add public wishlist translation keys (titles, labels, badges, empty states, errors) to `frontend/src/i18n/locales/en.json`
- [ ] T005 [P] Add public wishlist translation keys (titles, labels, badges, empty states, errors) to `frontend/src/i18n/locales/ru.json`
- [ ] T006 [P] Create `guest-reservations` localStorage utility module with `getStoredReservations()`, `addReservation()`, `removeReservation()`, `getAllTokens()` helpers in `frontend/src/lib/guest-reservations.ts`

**Checkpoint**: Foundation ready — API client methods, types, i18n keys, and localStorage utilities available for all stories

---

## Phase 2: User Story 1 — View a Public Wishlist (Priority: P1) 🎯 MVP

**Goal**: A visitor opens a public wishlist link and sees the wishlist title, occasion, description, and all gift items with their names, images, prices, and availability status. No login required.

**Independent Test**: Navigate to `/public/{slug}` → verify wishlist metadata renders, gift items display with correct details and status badges, 404 page shown for invalid/private slugs, empty state shown for wishlists with no items.

### Implementation for User Story 1

- [ ] T007 [P] [US1] Create `WishlistHeader` component (title, occasion, occasion date, description, item count) in `frontend/src/components/public-wishlist/WishlistHeader.tsx`
- [ ] T008 [P] [US1] Create `GiftItemCard` component (name, image/placeholder, price, product link, priority badge, reservation status badge) in `frontend/src/components/public-wishlist/GiftItemCard.tsx`
- [ ] T009 [P] [US1] Create `WishlistNotFound` component (user-friendly 404 message) in `frontend/src/components/public-wishlist/WishlistNotFound.tsx`
- [ ] T010 [P] [US1] Create `WishlistEmptyState` component (no items message) in `frontend/src/components/public-wishlist/WishlistEmptyState.tsx`
- [ ] T011 [P] [US1] Create `GiftItemSkeleton` loading component in `frontend/src/components/public-wishlist/GiftItemSkeleton.tsx`
- [ ] T012 [US1] Refactor `frontend/src/app/public/[slug]/page.tsx` to use new components (`WishlistHeader`, `GiftItemCard`, `WishlistNotFound`, `WishlistEmptyState`, `GiftItemSkeleton`), add i18n with `useTranslation`, wire up TanStack Query for `getPublicWishList` and `getPublicGiftItems`
- [ ] T013 [US1] Add CSS custom properties for theming support (`--wishlist-primary`, `--wishlist-bg`, `--wishlist-accent`) in `frontend/src/app/public/[slug]/layout.tsx` or a dedicated CSS file per R-003 decision
- [ ] T014 [US1] Ensure responsive design (320px mobile to desktop) for all public wishlist components; verify no horizontal scrolling

**Checkpoint**: User Story 1 complete — visitors can view any public wishlist with items, statuses, and a polished responsive UI

---

## Phase 3: User Story 2 — Reserve a Gift as a Guest (Priority: P1)

**Goal**: A visitor clicks "Reserve Gift" on an available item, fills in name and email in a dialog, submits, and sees confirmation. The item status updates immediately without page reload. Reservation token is stored in localStorage.

**Independent Test**: Open public wishlist → click "Reserve" on available item → fill name + email → submit → verify success toast, "Reserved" badge appears, `guest_reservations` array in localStorage contains new entry.

**Depends on**: Phase 2 (US1 — public wishlist page must render items with reserve buttons)

### Implementation for User Story 2

- [ ] T015 [US2] Refactor `GuestReservationDialog` in `frontend/src/components/guest/GuestReservationDialog.tsx` — add i18n translations, use `apiClient.createReservation()`, integrate with `guest-reservations` localStorage utility (R-005), add Zod validation for name (required, max 255) and email (required, valid format)
- [ ] T016 [US2] Add "Reserve Gift" button to `GiftItemCard` component in `frontend/src/components/public-wishlist/GiftItemCard.tsx` — disabled with "Reserved"/"Already Purchased" label when item is not available, opens `GuestReservationDialog` when available
- [ ] T017 [US2] Wire reservation mutation in `frontend/src/app/public/[slug]/page.tsx` — use TanStack Query `useMutation` for `createReservation`, invalidate `public-gift-items` query on success, show Sonner success toast with reservation token
- [ ] T018 [US2] Handle race condition: if `createReservation` returns error (item already reserved), show "already reserved" toast and refresh item list

**Checkpoint**: User Story 2 complete — guests can reserve available items, see immediate status updates, tokens stored locally

---

## Phase 4: User Story 3 — View My Reservations as a Guest (Priority: P2)

**Goal**: A guest navigates to `/my/reservations` and sees all their past reservations (from localStorage tokens), with gift name, wishlist title, reservation date, and status. Empty state shown if no reservations.

**Independent Test**: Make a reservation (US2) → navigate to `/my/reservations` → verify reserved item appears with correct details. Clear localStorage → verify empty state renders.

### Implementation for User Story 3

- [ ] T019 [US3] Rewrite `MyReservations` component in `frontend/src/components/wish-list/MyReservations.tsx` — read all tokens from `guest-reservations` localStorage utility, call `apiClient.getGuestReservations(token)` for each token, aggregate and display results, add i18n, show loading/error/empty states per R-005 decision
- [ ] T020 [US3] Update `frontend/src/app/my/reservations/page.tsx` to use rewritten `MyReservations` component with proper data fetching and i18n
- [ ] T021 [US3] Add "My Reservations" navigation link in `frontend/src/widgets/Footer.tsx` (persistent link per FR-008)
- [ ] T022 [P] [US3] Add "My Reservations" navigation link in `frontend/src/widgets/Header.tsx` (persistent link per FR-008)

**Checkpoint**: User Story 3 complete — guests can view all their reservations across wishlists from any page via header/footer link

---

## Phase 5: User Story 4 — Cancel a Reservation as a Guest (Priority: P3)

**Goal**: A guest cancels a reservation from the My Reservations page. The item returns to "available" on the public wishlist.

**Independent Test**: Reserve an item (US2) → go to My Reservations (US3) → click Cancel → confirm → verify reservation shows "canceled" status, item returns to "available" on public wishlist page.

**Depends on**: Phase 4 (US3 — My Reservations page must display reservations with cancel action)

### Implementation for User Story 4

- [ ] T023 [US4] Add cancel button and confirmation dialog to `MyReservations` component in `frontend/src/components/wish-list/MyReservations.tsx` — use `apiClient.cancelReservation(wishlistId, itemId, { reservation_token })`, invalidate `guest-reservations` query on success, show Sonner toast, handle invalid/expired token errors
- [ ] T024 [US4] Update localStorage on cancellation — remove canceled reservation from `guest_reservations` array via utility in `frontend/src/lib/guest-reservations.ts`, or keep with "canceled" status for display
- [ ] T025 [US4] Handle edge case: cancellation with invalid/expired token shows appropriate i18n error message and guides user to contact wishlist owner

**Checkpoint**: User Story 4 complete — guests can cancel reservations, items return to available status

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Quality, accessibility, and validation improvements across all stories

- [ ] T026 [P] Verify all components meet responsive design requirements (320px–desktop) — test on multiple viewport sizes
- [ ] T027 [P] Verify i18n completeness — every user-visible string uses `t()`, language switcher works on all public pages
- [ ] T028 [P] Verify edge cases: no-image items show placeholder, slow network shows loading states, failed requests show retry-able error messages
- [ ] T029 [P] Verify localStorage loss scenario — "My Reservations" empty state includes guidance message about contacting wishlist owner
- [ ] T030 Run `cd frontend && pnpm type-check` to verify TypeScript correctness across all new/modified files
- [ ] T031 Run `cd frontend && pnpm format` to ensure Biome formatting compliance
- [ ] T032 Validate quickstart.md scenarios end-to-end (view wishlist → reserve → view reservations → cancel)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — can start immediately
- **US1 (Phase 2)**: Depends on Phase 1 (needs types, i18n keys)
- **US2 (Phase 3)**: Depends on Phase 2 (needs GiftItemCard with reserve button slot)
- **US3 (Phase 4)**: Depends on Phase 1 (needs API client methods, localStorage utility) — can start in parallel with US2
- **US4 (Phase 5)**: Depends on Phase 4 (needs My Reservations page to add cancel action)
- **Polish (Phase 6)**: Depends on all user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: After Phase 1 → no dependencies on other stories
- **User Story 2 (P1)**: After US1 (needs the public wishlist page with item cards)
- **User Story 3 (P2)**: After Phase 1 → no dependency on US1/US2 for implementation (but integration test benefits from US2)
- **User Story 4 (P3)**: After US3 (needs My Reservations page as the cancellation UI host)

### Within Each User Story

- Components marked [P] can be created in parallel (different files)
- Page-level wiring depends on components being ready
- i18n keys must exist before components using `t()` (covered by Phase 1)

### Parallel Opportunities

- **Phase 1**: All 6 tasks (T001–T006) can run in parallel — they touch different files
- **Phase 2 (US1)**: T007, T008, T009, T010, T011 can run in parallel (different component files); T012 depends on them
- **Phase 3 (US2)**: T015 and T016 can start in parallel; T017 depends on both
- **Phase 4 (US3)**: T021 and T022 can run in parallel with T019; T020 depends on T019
- **US2 and US3 can be developed in parallel** by different developers (different pages/components)

---

## Parallel Example: Phase 1

```
# All setup tasks in parallel (different files):
T001: Add type exports in types.ts
T002: Add getGuestReservations in client.ts
T003: Add cancelReservation in client.ts
T004: Add EN translation keys in en.json
T005: Add RU translation keys in ru.json
T006: Create localStorage utility in guest-reservations.ts
```

## Parallel Example: User Story 1

```
# All component files in parallel:
T007: WishlistHeader.tsx
T008: GiftItemCard.tsx
T009: WishlistNotFound.tsx
T010: WishlistEmptyState.tsx
T011: GiftItemSkeleton.tsx

# Then sequential:
T012: Wire everything in page.tsx (depends on T007-T011)
T013: Add theming CSS properties
T014: Responsive design verification
```

---

## Implementation Strategy

### MVP First (User Stories 1 + 2)

1. Complete Phase 1: Setup (types, API client, i18n, localStorage utility)
2. Complete Phase 2: User Story 1 — View Public Wishlist
3. Complete Phase 3: User Story 2 — Reserve a Gift
4. **STOP and VALIDATE**: Visitors can view wishlists and reserve items
5. Deploy/demo if ready — this is the core value proposition

### Incremental Delivery

1. Phase 1 → Foundation ready
2. US1 (View Wishlist) → Test independently → First deployable increment
3. US2 (Reserve Gift) → Test independently → Core feature complete (MVP!)
4. US3 (My Reservations) → Test independently → Guest accountability
5. US4 (Cancel Reservation) → Test independently → Full feature complete
6. Polish → Production-quality release

### Suggested MVP Scope

**User Stories 1 + 2** together form the MVP — a visitor can view a public wishlist and reserve a gift. This delivers the core value proposition. US3 and US4 are enhancements.

---

## Notes

- This is a **frontend-only** feature. The backend API is complete and verified (R-001).
- All components use the existing `openapi-fetch` typed client — no raw `fetch()` calls.
- The existing `GuestReservationDialog.tsx` needs refactoring, not replacement (R-005).
- The existing `MyReservations.tsx` needs rewriting due to incorrect API calls (R-005).
- The public wishlist page has independent styling with CSS custom properties for future theming (R-003).
- i18n extends the existing i18next setup — no new libraries needed (R-002).
- Guest reservation tokens use localStorage, not cookies (R-004).
- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
