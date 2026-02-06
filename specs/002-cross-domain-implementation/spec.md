# Feature Specification: Cross-Domain Architecture Implementation

**Feature Branch**: `002-cross-domain-implementation`
**Created**: 2026-02-02
**Status**: Draft
**Input**: User description: "Implement cross-domain architecture with secure authentication flow between Frontend (Vercel), Mobile (Expo), and Backend (Render) as defined in docs/plans/"

## Overview

This specification covers the implementation of a cross-domain architecture for the Wish List application where Frontend (Next.js on Vercel), Mobile (Expo/React Native), and Backend (Go/Echo on Render) operate on different domains. The core challenge is enabling secure authentication across these separate domains where httpOnly cookies cannot be shared.

**Architecture Summary**:
- **Frontend (Web)**: Public wishlist viewing, guest reservations, authentication with redirect to Mobile
- **Mobile (App)**: Personal cabinet for list management, authenticated features
- **Backend (API)**: REST API with JWT authentication, CORS configuration, cross-domain token handling

## Clarifications

### Session 2026-02-02

- Q: Should Mobile app redirect users back to Frontend after completing tasks? → A: No return flow - users navigate manually (open browser, use bookmarks)

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Web User Login and Mobile Redirect (Priority: P1)

A registered user visits the Frontend website, logs in, and wants to access their personal cabinet (manage wishlists, items). Since personal cabinet features are in the Mobile app, the system securely transfers their authenticated session to the Mobile app.

**Why this priority**: Core functionality that enables the cross-domain architecture. Without this, users cannot access authenticated features across platforms.

**Independent Test**: Can be fully tested by logging in on web, clicking "Personal Cabinet", and verifying the mobile app opens with the user already authenticated.

**Acceptance Scenarios**:

1. **Given** a registered user on the Frontend login page, **When** they enter valid credentials and submit, **Then** they receive an access token and are marked as authenticated
2. **Given** an authenticated user on Frontend, **When** they click "Personal Cabinet", **Then** the system generates a short-lived handoff code (60 seconds)
3. **Given** a handoff code is generated, **When** the Frontend redirects to Mobile via Universal Link, **Then** the Mobile app receives the code
4. **Given** the Mobile app receives a valid handoff code, **When** it exchanges the code with the Backend, **Then** it receives access and refresh tokens and the user is authenticated
5. **Given** a handoff code was already used, **When** attempting to use it again, **Then** the exchange fails with "Invalid or expired code"

---

### User Story 2 - Token Refresh Flow (Priority: P1)

An authenticated user continues using the application after their short-lived access token expires. The system automatically refreshes their authentication without requiring re-login.

**Why this priority**: Essential for user experience. Without automatic refresh, users would be logged out every 15 minutes.

**Independent Test**: Can be tested by logging in, waiting for access token expiration (15 minutes), then making an API request and verifying it succeeds after automatic token refresh.

**Acceptance Scenarios**:

1. **Given** an authenticated user with an expired access token but valid refresh token, **When** they make an API request, **Then** the system automatically obtains a new access token
2. **Given** a valid refresh token on Frontend (in httpOnly cookie), **When** calling the refresh endpoint, **Then** a new access token is returned and the refresh token is rotated
3. **Given** a valid refresh token on Mobile (in SecureStore), **When** calling the refresh endpoint with Authorization header, **Then** new tokens are returned
4. **Given** an invalid or expired refresh token, **When** attempting to refresh, **Then** the refresh fails and user is redirected to login

---

### User Story 3 - Guest Reservation on Public Wishlist (Priority: P1)

A guest (unauthenticated visitor) views a public wishlist shared via link and reserves a gift item without needing to create an account.

**Why this priority**: Core value proposition - allowing gift givers to reserve items without friction.

**Independent Test**: Can be tested by opening a public wishlist link, selecting an item, entering guest details, and verifying reservation confirmation.

**Acceptance Scenarios**:

1. **Given** a public wishlist URL, **When** an unauthenticated user visits it, **Then** they can view all gift items without logging in
2. **Given** a guest viewing a public wishlist, **When** they click "Reserve" on an available item, **Then** they see a form for name and email
3. **Given** a guest fills reservation form, **When** they submit with valid data, **Then** the item is reserved and they receive a confirmation with a management link
4. **Given** a guest has reserved an item, **When** they click the management link from their email, **Then** they can view and cancel their reservation

---

### User Story 4 - Frontend Secure Token Storage (Priority: P2)

The Frontend application stores authentication tokens securely to prevent XSS attacks from stealing user credentials.

**Why this priority**: Security requirement that protects all authenticated users from credential theft.

**Independent Test**: Can be tested by verifying localStorage contains no tokens and inspecting that access token is only in memory.

**Acceptance Scenarios**:

1. **Given** a user logs in on Frontend, **When** login succeeds, **Then** access token is stored in memory only (not localStorage/sessionStorage)
2. **Given** a user logs in on Frontend, **When** login succeeds, **Then** refresh token is stored in httpOnly cookie (set by Backend)
3. **Given** malicious JavaScript attempts to read tokens, **When** it accesses localStorage and sessionStorage, **Then** no authentication tokens are found
4. **Given** a user closes and reopens the browser tab, **When** the page loads, **Then** the system attempts to refresh the session using the httpOnly cookie

---

### User Story 5 - Mobile Secure Token Storage (Priority: P2)

The Mobile application stores authentication tokens securely using platform-native secure storage.

**Why this priority**: Security requirement for mobile platform that protects credentials from other apps and attackers.

**Independent Test**: Can be tested by verifying tokens are stored in expo-secure-store and not in AsyncStorage or other insecure storage.

**Acceptance Scenarios**:

1. **Given** a user logs in on Mobile, **When** login succeeds, **Then** access and refresh tokens are stored in expo-secure-store
2. **Given** a user logs out on Mobile, **When** logout completes, **Then** all tokens are removed from SecureStore
3. **Given** a user's account is deleted, **When** deletion completes, **Then** all tokens and cached data are cleared

---

### User Story 6 - CORS Protection (Priority: P2)

The Backend API only accepts requests from authorized Frontend and Mobile origins, protecting against cross-site request attacks.

**Why this priority**: Security foundation that prevents malicious websites from making authenticated requests on behalf of users.

**Independent Test**: Can be tested by making API requests from allowed origins (success) and disallowed origins (blocked).

**Acceptance Scenarios**:

1. **Given** a request from https://wishlist.com, **When** sent to Backend API, **Then** it includes correct CORS headers and succeeds
2. **Given** a request from https://malicious.com, **When** sent to Backend API, **Then** it is blocked by CORS policy
3. **Given** a preflight OPTIONS request from allowed origin, **When** sent to Backend, **Then** it returns appropriate Access-Control headers

---

### User Story 7 - User Logout Across Platforms (Priority: P3)

An authenticated user can log out, which clears their session on the current platform.

**Why this priority**: Standard security feature allowing users to end their session.

**Independent Test**: Can be tested by logging out and verifying all tokens are cleared and protected routes redirect to login.

**Acceptance Scenarios**:

1. **Given** an authenticated user on Frontend, **When** they click logout, **Then** their access token is cleared from memory and refresh token cookie is invalidated
2. **Given** an authenticated user on Mobile, **When** they tap logout, **Then** all tokens are removed from SecureStore and they are redirected to login screen

---

### Edge Cases

- What happens when the Mobile app is not installed and user clicks "Personal Cabinet"?
  - System should redirect to App Store/Play Store after a timeout (2.5 seconds)
- How does the system handle expired handoff codes (after 60 seconds)?
  - Returns "Invalid or expired code" error, user must restart the handoff process
- What happens if network connectivity is lost during token refresh?
  - System queues the refresh request and retries when connectivity is restored
- How does the system handle concurrent refresh token requests?
  - Only one refresh request is processed at a time; subsequent requests wait for the first to complete
- What happens if a user's account is deleted while they have active sessions?
  - All tokens become invalid, API requests fail with 401, user is redirected to login

## Requirements *(mandatory)*

### Functional Requirements

**Authentication & Token Management**:
- **FR-001**: System MUST generate short-lived access tokens (15 minutes) for API authentication
- **FR-002**: System MUST generate long-lived refresh tokens (7 days) for session persistence
- **FR-003**: System MUST support token refresh without requiring user re-authentication
- **FR-004**: System MUST implement OAuth-style handoff flow for Frontend-to-Mobile authentication transfer
- **FR-005**: System MUST generate one-time-use handoff codes that expire after 60 seconds
- **FR-006**: Frontend MUST store access tokens in memory only (not localStorage/sessionStorage)
- **FR-007**: Frontend MUST receive refresh tokens via httpOnly cookies from Backend
- **FR-008**: Mobile MUST store both access and refresh tokens in expo-secure-store
- **FR-009**: System MUST invalidate refresh tokens upon logout
- **FR-010**: System MUST clear all tokens and cached data upon account deletion

**CORS & Cross-Domain**:
- **FR-011**: Backend MUST configure CORS to allow requests from Frontend domain(s)
- **FR-012**: Backend MUST set `Access-Control-Allow-Credentials: true` for cookie handling
- **FR-013**: Backend MUST support multiple allowed origins via environment configuration

**Guest Functionality**:
- **FR-014**: System MUST allow unauthenticated users to view public wishlists
- **FR-015**: System MUST allow guests to reserve items with name and email only
- **FR-016**: System MUST generate guest tokens for reservation management
- **FR-017**: System MUST allow guests to cancel their reservations using guest token

**API Endpoints**:
- **FR-018**: Backend MUST provide `POST /auth/login` returning accessToken, refreshToken, and user data
- **FR-019**: Backend MUST provide `POST /auth/refresh` accepting refresh token via cookie or header
- **FR-020**: Backend MUST provide `POST /auth/mobile-handoff` generating handoff codes for authenticated users
- **FR-021**: Backend MUST provide `POST /auth/exchange` accepting handoff code and returning tokens
- **FR-022**: Backend MUST provide `POST /auth/logout` clearing refresh token cookie
- **FR-023**: Backend MUST provide `GET /health` for deployment health checks

**Rate Limiting**:
- **FR-024**: System MUST rate limit `/auth/login` to 5 requests per minute per IP
- **FR-025**: System MUST rate limit `/auth/mobile-handoff` to 10 requests per minute per user
- **FR-026**: System MUST rate limit `/auth/exchange` to 10 requests per minute per IP

### Constitution Requirements

- **CR-001**: Code Quality - All code MUST meet high standards of quality, maintainability, and readability
- **CR-002**: Test-First - Unit tests MUST be written for all business logic before implementation
- **CR-003**: API Contracts - All API contracts MUST be explicitly defined using OpenAPI/Swagger specifications
- **CR-004**: Data Privacy - No personally identifiable information (PII) MAY be stored without encryption
- **CR-005**: Semantic Versioning - All releases MUST follow semantic versioning (MAJOR.MINOR.PATCH) standards
- **CR-006**: Specification Checkpoints - Features MUST be fully specified before implementation begins

### Key Entities

- **User**: Registered application user with email, name, and authentication credentials
- **AccessToken**: Short-lived JWT (15 min) containing user ID, used for API authorization
- **RefreshToken**: Long-lived JWT (7 days) used to obtain new access tokens without re-authentication
- **HandoffCode**: One-time-use cryptographically secure code (60s expiry) for cross-app authentication transfer
- **GuestToken**: Token issued to unauthenticated users for managing their reservations
- **Reservation**: Gift item reservation made by either authenticated user or guest

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can complete the Frontend-to-Mobile authentication handoff in under 5 seconds (excluding app install time)
- **SC-002**: Access token refresh happens automatically without user intervention in 100% of valid cases
- **SC-003**: Zero authentication tokens are accessible via JavaScript in the browser (XSS protection)
- **SC-004**: System supports 10,000 concurrent authenticated users without authentication failures
- **SC-005**: Guest users can complete a reservation in under 60 seconds
- **SC-006**: All public wishlist pages load within 2 seconds for unauthenticated users
- **SC-007**: CORS policy correctly blocks 100% of requests from unauthorized origins
- **SC-008**: Rate limiting prevents more than specified request limits with 100% accuracy
- **SC-009**: Token refresh failure rate is below 0.1% for valid refresh tokens
- **SC-010**: All authentication flows work consistently across Chrome, Safari, Firefox, and Edge browsers

## Assumptions

1. **Domain Configuration**: Production domains (wishlist.com, api.wishlist.com) are already configured and have valid SSL certificates
2. **Universal Links/App Links**: Apple App Site Association and Android Asset Links files will be configured as part of deployment
3. **Mobile App Distribution**: Mobile app will be available in App Store and Play Store for the handoff redirect
4. **Email Service**: Email service for guest reservation confirmations is already configured
5. **AWS S3**: S3 bucket for image uploads is already configured
6. **PostgreSQL**: Database is already provisioned on Render with required schema
7. **Environment Variables**: All required environment variables can be configured on Vercel and Render
8. **HTTPS Only**: All production traffic uses HTTPS; HTTP is not supported

## Out of Scope

1. Magic link authentication (passwordless login)
2. OAuth2 integration with third-party providers (Google, GitHub, etc.)
3. Multi-factor authentication (MFA)
4. Session management across multiple devices (single device per platform is supported)
5. Token blacklisting with Redis (in-memory store used for handoff codes)
6. Real-time notifications (WebSocket/SSE)
7. Offline support for Mobile app
8. Mobile → Frontend return flow (users navigate back to Frontend manually via browser)
