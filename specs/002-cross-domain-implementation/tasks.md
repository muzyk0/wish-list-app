# Tasks: Cross-Domain Architecture Implementation

**Input**: Design documents from `/specs/002-cross-domain-implementation/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/auth-api.yaml

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US7)
- Paths use existing project structure: `backend/`, `frontend/`, `mobile/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Backend infrastructure required for all auth flows

- [X] T001 Add new environment variables to backend/.env.example for JWT_REFRESH_TOKEN_EXPIRY and CORS_ALLOWED_ORIGINS
- [X] T002 [P] Create backend/internal/auth/code_store.go with CodeStore struct for handoff codes
- [X] T003 [P] Create backend/internal/middleware/cors.go with CORS configuration middleware supporting credentials
- [X] T004 [P] Create backend/internal/middleware/rate_limit.go with RateLimiter implementation

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core backend auth endpoints that MUST be complete before Frontend/Mobile work

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T005 Extend backend/internal/auth/token_manager.go to support separate access (15m) and refresh (7d) token generation
- [X] T006 Create backend/internal/handlers/auth_handler.go with AuthHandler struct
- [X] T007 Implement POST /auth/refresh endpoint in backend/internal/handlers/auth_handler.go accepting cookie or Bearer token
- [X] T008 Implement POST /auth/mobile-handoff endpoint in backend/internal/handlers/auth_handler.go
- [X] T009 Implement POST /auth/exchange endpoint in backend/internal/handlers/auth_handler.go
- [X] T010 Implement POST /auth/logout endpoint in backend/internal/handlers/auth_handler.go
- [X] T011 Update backend/internal/handlers/user_handler.go Login method to set httpOnly refresh token cookie
- [X] T012 Register new auth routes and middleware in backend/cmd/server/main.go
- [X] T013 Add Swagger annotations to all new auth endpoints in backend/internal/handlers/auth_handler.go
- [X] T014 Run swag init to regenerate backend/docs/

**Checkpoint**: Backend auth infrastructure ready - Frontend/Mobile can now begin

---

## Phase 3: User Story 1 - Web User Login and Mobile Redirect (Priority: P1) 🎯 MVP

**Goal**: Enable authenticated users on Frontend to securely transfer their session to Mobile app via handoff code

**Independent Test**: Log in on web, click "Personal Cabinet", verify mobile app opens with user authenticated

### Implementation for User Story 1

- [X] T015 [US1] Create frontend/src/lib/auth.ts with AuthManager class storing access token in memory
- [X] T016 [US1] Create frontend/src/lib/mobile-handoff.ts with redirectToPersonalCabinet function
- [X] T017 [US1] Modify frontend/src/lib/api.ts to remove localStorage usage and use AuthManager
- [X] T018 [US1] Modify frontend/src/lib/api.ts to add credentials: 'include' for all fetch calls
- [X] T019 [US1] Update frontend/src/lib/api.ts login method to use AuthManager.setAccessToken()
- [X] T020 [US1] Create mobile/lib/api/auth.ts with SecureStore token management functions
- [X] T021 [US1] Modify mobile/app/_layout.tsx to handle auth deep links (wishlistapp://auth?code=xxx)
- [X] T022 [US1] Implement exchangeCodeForTokens function in mobile/lib/api/auth.ts
- [X] T023 [US1] Update mobile app.json with proper scheme and associatedDomains configuration

**Checkpoint**: User Story 1 complete - Frontend→Mobile handoff works end-to-end

---

## Phase 4: User Story 2 - Token Refresh Flow (Priority: P1)

**Goal**: Enable automatic token refresh on both Frontend and Mobile when access token expires

**Independent Test**: Log in, wait 15+ minutes (or manually expire token), make API request, verify automatic refresh

### Implementation for User Story 2

- [X] T024 [US2] Add refreshAccessToken method to frontend/src/lib/auth.ts with singleton pattern
- [X] T025 [US2] Modify frontend/src/lib/api.ts request method to retry on 401 after refresh attempt
- [X] T026 [US2] Add refresh flow to frontend/src/lib/auth.ts that calls POST /auth/refresh with credentials
- [X] T027 [US2] Implement refreshAccessToken in mobile/lib/api/auth.ts using SecureStore refresh token
- [X] T028 [US2] Modify mobile/lib/api/api.ts to add automatic token refresh on 401 response
- [X] T029 [US2] Create frontend/src/hooks/useAuth.ts hook with authentication state and refresh on mount

**Checkpoint**: User Story 2 complete - Token refresh works automatically on both platforms

---

## Phase 5: User Story 3 - Guest Reservation on Public Wishlist (Priority: P1)

**Goal**: Allow guests to view public wishlists and reserve items without authentication

**Independent Test**: Open public wishlist URL, reserve an item with name/email, verify confirmation received

### Implementation for User Story 3

- [X] T030 [US3] Verify backend/internal/handlers/reservation_handler.go supports guest reservations (existing)
- [X] T031 [US3] Ensure public wishlist endpoints in backend do not require authentication (existing)
- [X] T032 [US3] Verify guest token generation in backend/internal/auth/token_manager.go GenerateGuestToken (existing)
- [X] T033 [US3] Update frontend public wishlist page to allow unauthenticated access
- [X] T034 [US3] Add guest reservation form component to frontend for name/email input

**Checkpoint**: User Story 3 complete - Guest reservation flow works without login

---

## Phase 6: User Story 4 - Frontend Secure Token Storage (Priority: P2)

**Goal**: Ensure Frontend stores tokens securely to prevent XSS attacks

**Independent Test**: Inspect browser storage after login - verify no tokens in localStorage/sessionStorage

### Implementation for User Story 4

- [X] T035 [US4] Audit frontend/src/lib/api.ts for any remaining localStorage references
- [X] T036 [US4] Remove all localStorage.setItem('token', ...) calls from frontend/src/lib/api.ts
- [X] T037 [US4] Verify frontend/src/lib/auth.ts only stores access token in class property
- [X] T038 [US4] Add session restoration on page load in frontend/src/hooks/useAuth.ts via refresh endpoint
- [X] T039 [US4] Update frontend/.env.example with required environment variables

**Checkpoint**: User Story 4 complete - Zero tokens accessible via JavaScript

---

## Phase 7: User Story 5 - Mobile Secure Token Storage (Priority: P2)

**Goal**: Store Mobile tokens in platform-native secure storage (expo-secure-store)

**Independent Test**: Verify tokens stored in SecureStore, not AsyncStorage or other insecure storage

### Implementation for User Story 5

- [ ] T040 [US5] Install expo-secure-store in mobile if not present: npx expo install expo-secure-store
- [ ] T041 [US5] Audit mobile/lib/api/ for any AsyncStorage or insecure storage usage
- [ ] T042 [US5] Ensure mobile/lib/api/auth.ts uses only SecureStore for token storage
- [ ] T043 [US5] Implement clearTokens in mobile/lib/api/auth.ts for logout and account deletion
- [ ] T044 [US5] Update mobile logout flow to call clearTokens

**Checkpoint**: User Story 5 complete - Mobile tokens secured via platform encryption

---

## Phase 8: User Story 6 - CORS Protection (Priority: P2)

**Goal**: Backend only accepts requests from authorized origins

**Independent Test**: Make API request from allowed origin (success) and disallowed origin (blocked)

### Implementation for User Story 6

- [ ] T045 [US6] Implement CORS middleware in backend/internal/middleware/cors.go with allowlist from env
- [ ] T046 [US6] Add Access-Control-Allow-Credentials: true to CORS config
- [ ] T047 [US6] Add development origins (localhost:3000, localhost:8081, localhost:19006) to cors.go
- [ ] T048 [US6] Register CORS middleware before routes in backend/cmd/server/main.go
- [ ] T049 [US6] Test CORS preflight OPTIONS requests return correct headers

**Checkpoint**: User Story 6 complete - CORS blocks unauthorized origins

---

## Phase 9: User Story 7 - User Logout Across Platforms (Priority: P3)

**Goal**: Users can log out, clearing their session on the current platform

**Independent Test**: Log out, verify tokens cleared and protected routes redirect to login

### Implementation for User Story 7

- [ ] T050 [US7] Implement frontend logout that clears AuthManager token and calls POST /auth/logout
- [ ] T051 [US7] Update frontend/src/lib/api.ts logout method to include credentials for cookie clearing
- [ ] T052 [US7] Implement mobile logout in mobile/lib/api/auth.ts calling clearTokens and POST /auth/logout
- [ ] T053 [US7] Add logout redirect to login screen in mobile after token clear

**Checkpoint**: User Story 7 complete - Logout works on both platforms

---

## Phase 10: Polish & Cross-Cutting Concerns

**Purpose**: Quality improvements and verification

- [ ] T054 [P] Add rate limiting middleware to auth routes in backend/cmd/server/main.go
- [ ] T055 [P] Implement background cleanup for expired handoff codes in backend/internal/auth/code_store.go
- [ ] T056 [P] Add health check endpoint GET /health in backend/internal/handlers/health_handler.go (verify exists)
- [ ] T057 [P] Update backend/api/openapi3.yaml with new auth endpoints from contracts/auth-api.yaml
- [ ] T058 Run full test suite: make test
- [ ] T059 Verify no tokens in Frontend localStorage: grep -r "localStorage" frontend/src/
- [ ] T060 Test cross-domain auth flow end-to-end with actual domains
- [ ] T061 Security review: verify XSS protection, token storage, CORS configuration
- [ ] T062 Update CLAUDE.md with new auth flow documentation if needed

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies - can start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 - BLOCKS all user stories
- **Phases 3-9 (User Stories)**: All depend on Phase 2 completion
  - User Stories 1-3 (P1): Can start in parallel after Phase 2
  - User Stories 4-6 (P2): Can start after Phase 2, overlap with P1 stories
  - User Story 7 (P3): Can start after Phase 2
- **Phase 10 (Polish)**: Depends on all user stories being complete

### User Story Dependencies

- **US1 (Handoff)**: Backend endpoints (Phase 2) + Frontend auth + Mobile deep links
- **US2 (Refresh)**: Backend refresh endpoint (Phase 2) + Frontend/Mobile refresh logic
- **US3 (Guest)**: Mostly existing functionality, verify and minor updates
- **US4 (Frontend Security)**: Depends on US1/US2 auth manager being in place
- **US5 (Mobile Security)**: Depends on US1/US2 SecureStore implementation
- **US6 (CORS)**: Backend only, can be done early in Phase 2
- **US7 (Logout)**: Depends on auth manager (US4) and SecureStore (US5) being complete

### Parallel Opportunities

**Phase 1 (all [P] tasks):**
- T002, T003, T004 can run in parallel (different files)

**Phase 2:**
- T005-T011 are sequential (handler depends on token manager)
- T012-T014 sequential (registration, then swagger)

**User Stories:**
- US1-US3 (P1 priority) can run in parallel across Backend/Frontend/Mobile
- US4-US6 (P2 priority) can overlap with P1 stories
- Frontend tasks (T015-T019, T024-T026) can run in parallel with Mobile tasks (T020-T023, T027-T028)

---

## Parallel Example: Phase 1

```bash
# Launch all setup tasks together:
Task: "Create backend/internal/auth/code_store.go with CodeStore struct"
Task: "Create backend/internal/middleware/cors.go with CORS configuration"
Task: "Create backend/internal/middleware/rate_limit.go with RateLimiter"
```

---

## Implementation Strategy

### MVP First (User Stories 1-2)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - Backend auth endpoints)
3. Complete Phase 3: User Story 1 (Frontend→Mobile handoff)
4. Complete Phase 4: User Story 2 (Token refresh)
5. **STOP and VALIDATE**: Test cross-domain auth flow
6. Deploy/demo if ready

### Incremental Delivery

1. Phases 1-2 → Backend auth infrastructure ready
2. Add US1 → Handoff flow works → Demo
3. Add US2 → Refresh works → Demo
4. Add US3 → Guest reservation → Demo
5. Add US4-US6 → Security hardening → Demo
6. Add US7 → Logout → Full feature complete

### Parallel Team Strategy

With multiple developers:
- Developer A (Backend): Phases 1-2, US6 (CORS)
- Developer B (Frontend): US1 Frontend parts, US2 Frontend, US4
- Developer C (Mobile): US1 Mobile parts, US2 Mobile, US5, US7

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Backend Phase 2 MUST complete before Frontend/Mobile work begins
- All code must meet quality standards (Constitution Requirement: Code Quality)
- API contracts defined in contracts/auth-api.yaml (Constitution Requirement: API Contract Integrity)
- No tokens in localStorage - memory only for Frontend (Constitution Requirement: Data Privacy)
- Test with actual cross-domain setup before production deployment
