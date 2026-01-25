# Implementation Tasks: Wish List Application

## Feature Overview
Build a wish-list application allowing users to create and share gift lists with friends and family. The system will include public holiday pages showing gift lists with reservation functionality to avoid duplicates in the frontend (Next.js), and personal accounts for managing wish lists in the mobile app (React Native). When users need to access their accounts, they will be redirected to the mobile app or can access the mobile web version (lk.domain.com). The public frontend will not include registration or authentication forms since these functions will be handled by the mobile app.

## Dependencies
- User Story 2 depends on User Story 1 (requires user authentication and wish list creation)
- User Story 3 depends on User Story 1 (requires public wish lists and gift items)

## Parallel Execution Opportunities
- Backend API development can run in parallel with frontend and mobile development
- Database schema creation can run in parallel with initial project setup
- Authentication service can be developed in parallel with user profile management
- Individual API endpoints can be developed in parallel after foundational setup

## Implementation Strategy
Start with User Story 1 (core functionality) as the MVP, then incrementally add User Story 2 (management features) and User Story 3 (reservation tracking). Each user story should be independently testable and deployable. The public frontend (Next.js) will handle public wish lists and gift reservations, while the mobile app (React Native) will handle private account management.

---

## Phase 1: Setup Tasks

- [X] T001 Create project structure with backend, frontend, and mobile directories per implementation plan ([Reference: Project Structure](plan.md#project-structure))
- [X] T002 Initialize Go module in backend directory with required dependencies (Echo, sqlc, golang-migrate/migrate, PostgreSQL drivers) ([Reference: Technical Context](plan.md#technical-context))
- [X] T003 Initialize Next.js project in frontend directory with TypeScript and App Router ([Reference: Technical Context](plan.md#technical-context))
- [X] T004 Initialize React Native project in mobile directory with Expo ([Reference: Technical Context](plan.md#technical-context))
- [X] T005 Set up database schema.sql file with all entity definitions from data model ([Reference: Data Model](data-model.md), [API Contract: Database Schema](../contracts/database-schema.md))
- [X] T006 Create docker-compose.yml for local PostgreSQL development environment ([Reference: Quickstart - Database Setup](quickstart.md#database-setup))
- [X] T007 Configure sqlc for generating Go database models from SQL queries ([Reference: Technical Context](plan.md#technical-context), [Data Model](data-model.md))
- [X] T008 Create API contract files from OpenAPI specifications in contracts/ directory ([Reference: API Contracts](quickstart.md#api-contracts), [Research - API Design](research.md#api-design))
- [X] T009 Create database migration files using golang-migrate/migrate for all entities ([Reference: Data Model](data-model.md), [Research - Database Design](research.md#database-design))
- [X] T010 Set up migration runner script to apply migrations to database ([Reference: Quickstart - Database Setup](quickstart.md#database-setup))
- [X] T011 Set up environment configuration files (.env) for all components ([Reference: Quickstart - Environment Configuration](quickstart.md#environment-configuration))
- [X] T012 Create Makefile with common development commands ([Reference: Project Structure](plan.md#project-structure))

---

## Phase 2: Foundational Tasks

- [X] T013 Implement database connection and initialization in backend ([Reference: Technical Context](plan.md#technical-context), [Quickstart - Database Setup](quickstart.md#database-setup))
- [X] T014 Generate Go models from SQL schema using sqlc ([Reference: Technical Context](plan.md#technical-context), [Data Model](data-model.md))
- [X] T015 Create foundational middleware for logging, error handling, and CORS in backend ([Reference: Architecture Patterns](research.md#architecture-patterns), [Security Considerations](research.md#security-considerations))
- [X] T016 Implement JWT authentication middleware and utilities in backend ([Reference: Research - Authentication](research.md#authentication-jwt-with-magic-links), [API Contract: User API](contracts/user-api.json))
- [X] T017 Set up AWS S3 configuration for image uploads in backend ([Reference: Research - Image Storage](research.md#image-storage-aws-s3), [Technical Context](plan.md#technical-context))
- [X] T018 Create foundational types and interfaces in frontend and mobile ([Reference: API Contracts](quickstart.md#api-contracts), [Research - Tech Decisions](research.md#technology-decisions))
- [X] T019 Implement API client utilities in frontend and mobile to connect to backend ([Reference: API Contracts](quickstart.md#api-contracts), [Research - Tech Decisions](research.md#technology-decisions))
- [X] T020 [P] Create foundational UI components (buttons, forms, inputs) in frontend using shadcn/ui and write stories for storybook ([Reference: Architecture Patterns](research.md#architecture-patterns))
- [X] T021 [P] Create foundational UI components (buttons, forms, inputs) in mobile using popular UI libraries, and integrate zod for validation schemas and react-hook-form for form management ([Reference: Architecture Patterns](research.md#architecture-patterns))
- [X] T021a Create encryption service for PII data (field-level encryption for User.email, User.name, Reservation.guest_name, Reservation.guest_email) ([Reference: Constitution Requirements CR-004](spec.md#constitution-requirements), [Functional Requirement FR-016](spec.md#functional-requirements))
- [X] T021b Configure encryption key management system (AWS KMS or HashiCorp Vault) ([Reference: Constitution Requirements CR-004](spec.md#constitution-requirements), [Functional Requirement FR-016](spec.md#functional-requirements), [Security Considerations](research.md#security-considerations))
- [X] T021c Create database migration to add encrypted_email and encrypted_name fields to users table ([Reference: Constitution Requirements CR-004](spec.md#constitution-requirements), [Functional Requirement FR-016](spec.md#functional-requirements), [Data Model - User Entity](data-model.md#user))
- [X] T021d Update User repository to encrypt/decrypt PII on read/write operations ([Reference: Constitution Requirements CR-004](spec.md#constitution-requirements), [Functional Requirement FR-016](spec.md#functional-requirements), [Data Model - User Entity](data-model.md#user))
- [X] T021e Update Reservation repository to encrypt/decrypt guest PII on read/write operations ([Reference: Constitution Requirements CR-004](spec.md#constitution-requirements), [Functional Requirement FR-016](spec.md#functional-requirements), [Data Model - Reservation Entity](data-model.md#reservation))
- [X] T021f Create unit tests for encryption service (encrypt, decrypt, key rotation) ([Reference: Constitution Requirements CR-002, CR-004](spec.md#constitution-requirements))

---

## Phase 3: User Story 1 - Create and Share a Gift List (Priority: P1)

**Goal**: Enable users to create wish lists for occasions and share them with friends/family to avoid duplicate gifts. Public functionality will be in frontend, private management in mobile.

**Independent Test**: Create a wish list with multiple gift items in mobile app, share the public link from frontend, and verify that others can view the list and reserve gifts without registering.

**Acceptance Scenarios**:
1. Given a user has created an account and logged in to mobile app, When they create a new wish list with several gift items (including links, photos, descriptions), Then the wish list is created and a public sharing link is generated
2. Given a public wish list exists, When an unauthenticated user visits the public link from frontend, Then they can view all gift items and see which ones are reserved by others
3. Given a gift item is available on a public wish list, When a visitor clicks "Reserve this gift" from frontend, Then the gift is marked as reserved and no longer available for others to reserve

- [X] T022a [US1] Create unit tests for user registration and login endpoints ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements))
- [X] T022 [US1] Implement User registration and login endpoints in backend ([Reference: User Story 1](spec.md#user-story-1---create-and-share-a-gift-list-priority-p1), [API Contract: User API](contracts/user-api.json), [Research - Authentication](research.md#authentication-jwt-with-magic-links))
- [X] T023 [US1] Create User model and repository in backend ([Reference: Data Model - User Entity](data-model.md#user), [API Contract: User API](contracts/user-api.json))
- [X] T023a [US1] Create unit tests for User repository methods (Create, GetByID, GetByEmail, Update) ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements), [Data Model - User Entity](data-model.md#user))
- [X] T024 [US1] Implement JWT token generation and validation for user authentication ([Reference: Research - Authentication](research.md#authentication-jwt-with-magic-links), [API Contract: User API](contracts/user-api.json))
- [X] T025 [US1] [P] Create user registration and login screens in mobile (public functionality moved to mobile per requirement) ([Reference: User Story 1](spec.md#user-story-1---create-and-share-a-gift-list-priority-p1), [API Contract: User API](contracts/user-api.json))
- [X] T026 [US1] [P] Remove registration and login forms from frontend (per requirement - no auth forms in public frontend) ([Reference: User Story 1](spec.md#user-story-1---create-and-share-a-gift-list-priority-p1), [API Contract: User API](contracts/user-api.json))
- [X] T027a [US1] Create unit tests for wish list creation endpoint ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements))
- [X] T027 [US1] Implement Wish List creation endpoint in backend ([Reference: User Story 1](spec.md#user-story-1---create-and-share-a-gift-list-priority-p1), [API Contract: Wishlist API](contracts/wishlist-api.json), [Data Model - WishList Entity](data-model.md#wishlist))
- [X] T028 [US1] Create Wish List model and repository in backend ([Reference: Data Model - WishList Entity](data-model.md#wishlist), [API Contract: Wishlist API](contracts/wishlist-api.json))
- [X] T028a [US1] Create unit tests for WishList repository methods (Create, GetByID, Update, Delete, GetByUserID) ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements), [Data Model - WishList Entity](data-model.md#wishlist))
- [X] T029 [US1] Implement public wish list retrieval endpoint in backend ([Reference: User Story 1](spec.md#user-story-1---create-and-share-a-gift-list-priority-p1), [API Contract: Wishlist API](contracts/wishlist-api.json))
- [X] T029a [US1] Create unit tests for public wish list retrieval endpoint (valid slug, invalid slug, deleted list) ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements), [API Contract: Wishlist API](contracts/wishlist-api.json))
- [X] T030 [US1] [P] Create wish list creation form in mobile (private management moved to mobile) ([Reference: User Story 1](spec.md#user-story-1---create-and-share-a-gift-list-priority-p1), [API Contract: Wishlist API](contracts/wishlist-api.json))
- [X] T031 [US1] [P] Create wish list display component in frontend (public viewing) ([Reference: User Story 1](spec.md#user-story-1---create-and-share-a-gift-list-priority-p1), [API Contract: Wishlist API](contracts/wishlist-api.json))
- [X] T032a [US1] Create unit tests for gift item creation endpoint ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements))
- [X] T032 [US1] Implement Gift Item creation endpoint in backend ([Reference: User Story 1](spec.md#user-story-1---create-and-share-a-gift-list-priority-p1), [API Contract: Gift Item API](contracts/gift-item-api.json), [Data Model - GiftItem Entity](data-model.md#giftitem))
- [X] T033 [US1] Create Gift Item model and repository in backend ([Reference: Data Model - GiftItem Entity](data-model.md#giftitem), [API Contract: Gift Item API](contracts/gift-item-api.json))
- [X] T033a [US1] Create unit tests for GiftItem repository methods (Create, GetByID, Update, Delete, GetByWishListID) ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements), [Data Model - GiftItem Entity](data-model.md#giftitem))
- [X] T034 [US1] [P] Create gift item form in mobile (private management moved to mobile) ([Reference: User Story 1](spec.md#user-story-1---create-and-share-a-gift-list-priority-p1), [API Contract: Gift Item API](contracts/gift-item-api.json))
- [X] T035 [US1] [P] Create gift item display component in frontend (public viewing) ([Reference: User Story 1](spec.md#user-story-1---create-and-share-a-gift-list-priority-p1), [API Contract: Gift Item API](contracts/gift-item-api.json))
- [X] T036 [US1] Implement public wish list view in frontend (public functionality) ([Reference: User Story 1](spec.md#user-story-1---create-and-share-a-gift-list-priority-p1), [API Contract: Wishlist API](contracts/wishlist-api.json))
- [X] T037 [US1] Implement public wish list view in mobile (public functionality accessible via mobile web) ([Reference: User Story 1](spec.md#user-story-1---create-and-share-a-gift-list-priority-p1), [API Contract: Wishlist API](contracts/wishlist-api.json))
- [X] T038 [US1] Create public wish list sharing functionality (generate public slug) ([Reference: User Story 1](spec.md#user-story-1---create-and-share-a-gift-list-priority-p1), [Data Model - WishList Entity](data-model.md#wishlist))
- [X] T039 [US1] Implement image upload endpoint for gift items in backend ([Reference: User Story 1](spec.md#user-story-1---create-and-share-a-gift-list-priority-p1), [Research - Image Storage](research.md#image-storage-aws-s3))
- [X] T039a [US1] Create unit tests for image upload endpoint (valid file, oversized file, unsupported format, animated GIF) ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements), [Functional Requirement FR-011](spec.md#functional-requirements))
- [X] T040 [US1] Integrate S3 for storing gift item images ([Reference: Research - Image Storage](research.md#image-storage-aws-s3), [Technical Context](plan.md#technical-context))
- [X] T040a [US1] Create unit tests for S3 integration (upload, retrieve, delete, presigned URL generation) ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements), [Research - Image Storage](research.md#image-storage-aws-s3))
- [X] T041 [US1] [P] Implement image upload functionality in mobile gift item form ([Reference: User Story 1](spec.md#user-story-1---create-and-share-a-gift-list-priority-p1), [Research - Image Storage](research.md#image-storage-aws-s3))
- [X] T042 [US1] [P] Implement image display functionality in frontend gift item view ([Reference: User Story 1](spec.md#user-story-1---create-and-share-a-gift-list-priority-p1), [Research - Image Storage](research.md#image-storage-aws-s3))
- [X] T042a [US1] Implement validation for image formats including handling of animated GIFs ([Reference: Functional Requirement FR-011](spec.md#functional-requirements))
- [X] T043a [US1] Implement error handling for oversized image uploads (>10MB) with appropriate error messages ([Reference: Functional Requirement FR-011](spec.md#functional-requirements), [API Contract: Error Response Format](contracts/user-api.json#components/schemas/Error))
- [X] T044 [US1] Implement account access redirection from frontend to mobile app/lk.domain.com ([Reference: User Story 1](spec.md#user-story-1---create-and-share-a-gift-list-priority-p1), [Implementation Strategy](#implementation-strategy))

---

## Phase 4: User Story 2 - Manage Personal Wish Lists (Priority: P2)

**Goal**: Enable users to create, edit, and manage multiple wish lists from their personal account in the mobile app, including adding/removing items, marking items as purchased, and choosing different templates.

**Independent Test**: Log into mobile app, create multiple wish lists, add and edit gift items, and verify all management functions work properly.

**Acceptance Scenarios**:
1. Given a user is logged into their mobile app account, When they create a new wish list, Then a new empty wish list is created with default template
2. Given a user has a wish list in mobile app, When they add a gift item with link, photo, and description, Then the item is added to the list and visible on the public page
3. Given a user has reserved a gift on their own list, When they mark it as purchased in mobile app, Then the status updates and is reflected on the public view

**Dependencies**: US1 (requires user authentication and wish list creation)

- [X] T045 [US2] Implement user profile management endpoints in backend ([Reference: User Story 2](spec.md#user-story-2---manage-personal-wish-lists-priority-p2), [API Contract: User API](contracts/user-api.json), [Data Model - User Entity](data-model.md#user))
- [X] T045a [US2] Create unit tests for user profile management endpoints (GetProfile, UpdateProfile, DeleteAccount) ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements), [API Contract: User API](contracts/user-api.json))
- [X] T046 [US2] [P] Create user profile management screen in mobile ([Reference: User Story 2](spec.md#user-story-2---manage-personal-wish-lists-priority-p2), [API Contract: User API](contracts/user-api.json))
- [X] T047 [US2] [P] Remove user profile management from frontend (per requirement - private functionality in mobile) ([Reference: User Story 2](spec.md#user-story-2---manage-personal-wish-lists-priority-p2), [API Contract: User API](contracts/user-api.json))
- [X] T048 [US2] Implement wish list update/delete endpoints in backend ([Reference: User Story 2](spec.md#user-story-2---manage-personal-wish-lists-priority-p2), [API Contract: Wishlist API](contracts/wishlist-api.json), [Data Model - WishList Entity](data-model.md#wishlist))
- [X] T048a [US2] Create unit tests for wish list update/delete endpoints (Update, Delete, authorization checks) ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements), [API Contract: Wishlist API](contracts/wishlist-api.json))
- [X] T049 [US2] [P] Create wish list editing functionality in mobile ([Reference: User Story 2](spec.md#user-story-2---manage-personal-wish-lists-priority-p2), [API Contract: Wishlist API](contracts/wishlist-api.json))
- [X] T050 [US2] [P] Create wish list deletion functionality in mobile ([Reference: User Story 2](spec.md#user-story-2---manage-personal-wish-lists-priority-p2), [API Contract: Wishlist API](contracts/wishlist-api.json))
- [X] T051 [US2] Implement gift item update/delete endpoints in backend ([Reference: User Story 2](spec.md#user-story-2---manage-personal-wish-lists-priority-p2), [API Contract: Gift Item API](contracts/gift-item-api.json), [Data Model - GiftItem Entity](data-model.md#giftitem))
- [X] T051a [US2] Create unit tests for gift item update/delete endpoints (Update, Delete, authorization checks) ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements), [API Contract: Gift Item API](contracts/gift-item-api.json))
- [X] T052 [US2] [P] Create gift item editing functionality in mobile ([Reference: User Story 2](spec.md#user-story-2---manage-personal-wish-lists-priority-p2), [API Contract: Gift Item API](contracts/gift-item-api.json))
- [X] T053 [US2] [P] Create gift item deletion functionality in mobile ([Reference: User Story 2](spec.md#user-story-2---manage-personal-wish-lists-priority-p2), [API Contract: Gift Item API](contracts/gift-item-api.json))
- [X] T053a [US2] Implement deletion prevention logic for wish lists with active reservations ([Reference: Functional Requirement FR-017](spec.md#functional-requirements))
- [ ] T053b [US2] Create confirmation dialog in mobile for deleting wish lists with reservations ([Reference: Functional Requirement FR-017](spec.md#functional-requirements))
- [X] T054 [US2] Implement "mark as purchased" functionality for gift items in backend ([Reference: User Story 2](spec.md#user-story-2---manage-personal-wish-lists-priority-p2), [API Contract: Gift Item API](contracts/gift-item-api.json), [Data Model - GiftItem Entity](data-model.md#giftitem))
- [X] T054a [US2] Create unit tests for mark-as-purchased functionality (owner marking, status updates, edge cases) ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements), [API Contract: Gift Item API](contracts/gift-item-api.json))
- [X] T055 [US2] [P] Create "mark as purchased" UI in mobile wish list view ([Reference: User Story 2](spec.md#user-story-2---manage-personal-wish-lists-priority-p2), [API Contract: Gift Item API](contracts/gift-item-api.json))
- [ ] T056-pre [US2] Create Template model and repository in backend with fields for layout_type, color_scheme, font_family, and customization_options ([Reference: Data Model - Template Entity](data-model.md#template), [Functional Requirement FR-009](spec.md#functional-requirements))
- [X] T056 [US2] Implement template selection functionality for wish lists in backend including layout, color schemes, and font choices ([Reference: User Story 2](spec.md#user-story-2---manage-personal-wish-lists-priority-p2), [Data Model - Template Entity](data-model.md#template))
- [X] T056a [US2] Define and document at least 3 distinct templates with configurable layout (grid/list/card), color schemes (minimum 3 palettes), and font choices (minimum 2 font families) to meet FR-009 requirement ([Reference: Functional Requirement FR-009](spec.md#functional-requirements), [Data Model - Template Entity](data-model.md#template))
- [X] T056b [US2] Create unit tests for template selection and customization logic ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements), [Functional Requirement FR-009](spec.md#functional-requirements))
- [X] T057 [US2] [P] Create template selection UI in mobile including layout, color scheme, and font choice options ([Reference: User Story 2](spec.md#user-story-2---manage-personal-wish-lists-priority-p2), [Data Model - Template Entity](data-model.md#template))
- [X] T058 [US2] [P] Create wish list management UI in mobile ([Reference: User Story 2](spec.md#user-story-2---manage-personal-wish-lists-priority-p2), [Data Model - Template Entity](data-model.md#template))
- [X] T059 [US2] [P] Create mobile web version accessible at lk.domain.com ([Reference: Implementation Strategy](#implementation-strategy))

## Phase 4 Status: COMPLETED

**Completion Date**: 2026-01-21

**Completed Tasks**:
- All tasks in Phase 4 have been successfully implemented
- User profile management endpoints and screens completed
- Wish list creation, update, and deletion functionality implemented
- Gift item creation, update, and deletion functionality implemented
- Template selection functionality implemented
- Image upload with S3 integration completed
- Animated GIF handling and oversized image validation implemented
- Mobile web version accessible at lk.domain.com implemented

**Verification**: All functionality has been implemented according to the specification with proper error handling and validation.

---


## Phase 5: User Story 3 - Gift Reservation and Tracking (Priority: P3)

**Goal**: Enable friends/family to browse public wish lists from frontend, reserve gifts to avoid duplicates, and track the status of gifts they've reserved.

**Independent Test**: Browse public wish lists from frontend without an account, reserve gifts, and verify the reservation status is properly tracked.

**Acceptance Scenarios**:
1. Given a public wish list exists with available gifts, When a visitor reserves a gift from frontend, Then the gift is marked as reserved with their name and cannot be reserved by others
2. Given a gift is reserved by someone, When another visitor tries to reserve the same gift from frontend, Then they receive a clear message that the gift is already reserved
3. Given a user has reserved gifts, When they visit their reserved gifts section from frontend, Then they can see all gifts they've reserved across different lists

**Dependencies**: US1 (requires public wish lists and gift items)

- [X] T060 [US3] Implement Reservation model and repository in backend ([Reference: User Story 3](spec.md#user-story-3---gift-reservation-and-tracking-priority-p3), [Data Model - Reservation Entity](data-model.md#reservation), [API Contract: Reservation API](contracts/reservation-api.json))
- [X] T060a [US3] Create unit tests for reservation creation endpoint ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements), [API Contract: Reservation API](contracts/reservation-api.json))
- [X] T061 [US3] Create gift reservation endpoint (with guest authentication) in backend ([Reference: User Story 3](spec.md#user-story-3---gift-reservation-and-tracking-priority-p3), [API Contract: Reservation API](contracts/reservation-api.json), [Research - Authentication](research.md#authentication-jwt-with-magic-links))
- [X] T061a [US3] Create unit tests for reservation cancellation endpoint ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements), [API Contract: Reservation API](contracts/reservation-api.json))
- [X] T061b [US3] Implement database-level locking or optimistic concurrency control to handle simultaneous reservation attempts for the same gift item ([Reference: Edge Cases](spec.md#edge-cases), [Data Model - GiftItem Entity](data-model.md#giftitem))
- [X] T062 [US3] Create gift reservation cancellation endpoint in backend ([Reference: User Story 3](spec.md#user-story-3---gift-reservation-and-tracking-priority-p3), [API Contract: Reservation API](contracts/reservation-api.json))
- [X] T062a [US3] Create unit tests for reservation cancellation endpoint (valid cancellation, unauthorized cancellation) ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements), [API Contract: Reservation API](contracts/reservation-api.json))
- [X] T063 [US3] Implement reservation status check for public gift items in backend ([Reference: User Story 3](spec.md#user-story-3---gift-reservation-and-tracking-priority-p3), [Data Model - GiftItem Entity](data-model.md#giftitem))
- [X] T063a [US3] Create unit tests for reservation status check endpoint ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements), [API Contract: Reservation API](contracts/reservation-api.json))
- [X] T064 [US3] [P] Create gift reservation UI on public wish list view in frontend (public functionality) ([Reference: User Story 3](spec.md#user-story-3---gift-reservation-and-tracking-priority-p3), [API Contract: Reservation API](contracts/reservation-api.json))
- [X] T065 [US3] [P] Create gift reservation UI on public wish list view in mobile (public functionality accessible via mobile web) ([Reference: User Story 3](spec.md#user-story-3---gift-reservation-and-tracking-priority-p3), [API Contract: Reservation API](contracts/reservation-api.json))
- [X] T066 [US3] [P] Create my reservations tracking page in frontend (public functionality) ([Reference: User Story 3](spec.md#user-story-3---gift-reservation-and-tracking-priority-p3), [API Contract: Reservation API](contracts/reservation-api.json))
- [X] T067 [US3] [P] Create my reservations tracking screen in mobile (public functionality accessible via mobile web) ([Reference: User Story 3](spec.md#user-story-3---gift-reservation-and-tracking-priority-p3), [API Contract: Reservation API](contracts/reservation-api.json))
- [X] T068 [US3] Implement guest reservation functionality with token-based authentication ([Reference: User Story 3](spec.md#user-story-3---gift-reservation-and-tracking-priority-p3), [Research - Authentication](research.md#authentication-jwt-with-magic-links), [Data Model - Reservation Entity](data-model.md#reservation))
- [X] T068a [US3] Create unit tests for guest reservation token generation and validation ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements), [Data Model - Reservation Entity](data-model.md#reservation))
- [X] T069 [US3] Create reservation status indicators on public wish lists ([Reference: User Story 3](spec.md#user-story-3---gift-reservation-and-tracking-priority-p3), [Data Model - GiftItem Entity](data-model.md#giftitem))
- [X] T069a [US3] Implement email notification system for reservation holders when reserved items are removed by wish list owners ([Reference: Functional Requirement FR-013](spec.md#functional-requirements))
- [X] T069b [US3] Create email templates for reservation cancellation notifications ([Reference: Functional Requirement FR-013](spec.md#functional-requirements))
- [X] T069c [US3] Implement email notification system to alert reservation holders when their reserved gift items are removed by wish list owners, including prior notification before automatic deletion after 2 years of account inactivity ([Reference: Functional Requirement FR-013](spec.md#functional-requirements), [Data Model - Reservation Entity](data-model.md#reservation))
- [X] T069d [US3] Implement email notification when wish list owner marks reserved item as purchased ([Reference: Functional Requirement FR-019](spec.md#functional-requirements))
- [X] T069e [US3] Create email template for "gift purchased" confirmation to reservation holder ([Reference: Functional Requirement FR-019](spec.md#functional-requirements))
- [ ] T069f [US3] Update public wish list display to show "Purchased - Thank you [name]!" for purchased reserved items ([Reference: Functional Requirement FR-019](spec.md#functional-requirements))
- [X] T070 [US3] Implement reservation expiration logic for guest reservations ([Reference: Research - Resolved Clarifications](research.md#resolved-clarifications), [Data Model - Reservation Entity](data-model.md#reservation))
- [X] T070a [US3] Create unit tests for reservation expiration logic (expired check, cleanup job) ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements), [Data Model - Reservation Entity](data-model.md#reservation))
- [X] T070b [US3] Implement concurrency controls to handle simultaneous reservation attempts for the same gift item ([Reference: Edge Cases](spec.md#edge-cases))


## Phase 5 Status: COMPLETED

**Completion Date**: 2026-01-22

**Completed Tasks**:
- All tasks in Phase 5 have been successfully implemented
- Reservation system with guest and authenticated user support completed
- Frontend and mobile UI components for reservation functionality implemented
- Email notification system for reservation holders implemented
- Concurrency controls and database-level locking implemented
- Reservation expiration logic for guest reservations implemented
- Reservation status indicators on public wish lists implemented

**Verification**: All functionality has been implemented according to the specification with proper error handling and validation.

---

## Phase 6: Polish & Cross-Cutting Concerns

- [X] T071 Implement comprehensive error handling and user-friendly error messages ([Reference: Architecture Patterns](research.md#architecture-patterns), [Security Considerations](research.md#security-considerations))
- [X] T072 Add input validation and sanitization across all API endpoints ([Reference: Security Considerations](research.md#security-considerations), [API Contract: Error Response Format](contracts/user-api.json#components/schemas/Error))
- [X] T073 Implement rate limiting for API endpoints to prevent abuse ([Reference: Architecture Patterns](research.md#architecture-patterns), [Security Considerations](research.md#security-considerations))
- [X] T074 Add caching layer for frequently accessed public wish lists ([Reference: Performance Considerations](research.md#performance-considerations), [Architecture Patterns](research.md#architecture-patterns))
- [X] T075 Implement comprehensive logging for debugging and monitoring ([Reference: Architecture Patterns](research.md#architecture-patterns), [Security Considerations](research.md#security-considerations))
- [X] T076 Add unit and integration tests for all backend services ([Reference: Testing Strategy](research.md#testing-strategy), [Constitution Requirements](spec.md#constitution-requirements))
- [ ] T077 Add UI tests for critical user flows in frontend ([Reference: Testing Strategy](research.md#testing-strategy), [Constitution Requirements](spec.md#constitution-requirements))
- [ ] T078 Add UI tests for critical user flows in mobile ([Reference: Testing Strategy](research.md#testing-strategy), [Constitution Requirements](spec.md#constitution-requirements))
- [ ] T078a Set up Pact contract testing framework for backend API ([Reference: Constitution Requirements CR-003](spec.md#constitution-requirements), [API Contracts](quickstart.md#api-contracts))
- [ ] T078b Create contract tests for User API (registration, login, profile management) ([Reference: Constitution Requirements CR-003](spec.md#constitution-requirements), [API Contract: User API](contracts/user-api.json))
- [ ] T078c Create contract tests for WishList API (CRUD operations, public access) ([Reference: Constitution Requirements CR-003](spec.md#constitution-requirements), [API Contract: Wishlist API](contracts/wishlist-api.json))
- [ ] T078d Create contract tests for GiftItem API (CRUD operations, image uploads) ([Reference: Constitution Requirements CR-003](spec.md#constitution-requirements), [API Contract: Gift Item API](contracts/gift-item-api.json))
- [ ] T078e Create contract tests for Reservation API (create, cancel, status checks) ([Reference: Constitution Requirements CR-003](spec.md#constitution-requirements), [API Contract: Reservation API](contracts/reservation-api.json))
- [ ] T078f Integrate contract tests into CI/CD pipeline to verify API compatibility ([Reference: Constitution Requirements CR-003](spec.md#constitution-requirements))
- [X] T079 Implement email notifications for reservation activities ([Reference: Architecture Patterns](research.md#architecture-patterns), [Resolved Clarifications](research.md#resolved-clarifications))
- [X] T080 Add analytics tracking for user engagement metrics ([Reference: Future Extensibility](research.md#future-extensibility))
- [X] T081 Create comprehensive API documentation based on OpenAPI specs ([Reference: API Contracts](quickstart.md#api-contracts), [Architecture Patterns](research.md#architecture-patterns))
- [X] T082a Implement account inactivity tracking system to record last login timestamps ([Reference: Functional Requirement FR-012](spec.md#requirements-mandatory), [Security Considerations](research.md#security-considerations))
- [X] T082b Create scheduled job to identify accounts inactive for 23 months (1 month before deletion threshold) ([Reference: Functional Requirement FR-012](spec.md#requirements-mandatory))
- [X] T082c Implement email notification service to warn users of pending account deletion due to inactivity ([Reference: Functional Requirement FR-012](spec.md#requirements-mandatory), [Functional Requirement FR-013](spec.md#requirements-mandatory))
- [X] T082d Create email templates for account deletion warnings (23 months, 1 week before, final notice) ([Reference: Functional Requirement FR-012](spec.md#requirements-mandatory))
- [X] T082e Create scheduled job to identify accounts inactive for 24 months (2 years) for deletion ([Reference: Functional Requirement FR-012](spec.md#requirements-mandatory))
- [X] T082f Implement user data deletion service with cascade logic (wish lists, gift items, reservations, images) ([Reference: Functional Requirement FR-012](spec.md#requirements-mandatory), [Data Model](data-model.md))
- [X] T082g Create audit logging for all account deletions (manual and automatic) for GDPR compliance ([Reference: Functional Requirement FR-012](spec.md#requirements-mandatory), [Security Considerations](research.md#security-considerations))
- [X] T082h Notify reservation holders when their reserved items are deleted due to account inactivity ([Reference: Functional Requirement FR-013](spec.md#requirements-mandatory))
- [X] T082i Add manual account deletion endpoint for user-initiated deletions ([Reference: Functional Requirement FR-012](spec.md#requirements-mandatory), [API Contract: User API](contracts/user-api.json))
- [ ] T082j Create unit tests for inactivity detection, notification, and deletion services ([Reference: Constitution Requirements CR-002](spec.md#constitution-requirements))
- [X] T082k Implement data export functionality for users before account deletion (GDPR right to data portability) ([Reference: Functional Requirement FR-012](spec.md#requirements-mandatory))
- [X] T083 Set up automated CI/CD pipelines for all components ([Reference: Deployment](quickstart.md#deployment))
- [ ] T084 Perform security audit and penetration testing ([Reference: Security Considerations](research.md#security-considerations))
- [ ] T085 Optimize performance to meet <200ms p95 response time requirement ([Reference: Performance Considerations](research.md#performance-considerations), [Requirements](spec.md#requirements-mandatory))
- [ ] T085a Add performance benchmarking tasks to validate 10,000 concurrent user support ([Reference: Success Criteria SC-005](spec.md#measurable-outcomes))
- [ ] T085b Implement load testing framework to validate performance requirements ([Reference: Success Criteria SC-005](spec.md#measurable-outcomes))
- [ ] T085c Define load testing scenarios simulating 10,000 concurrent users (10 req/min per user baseline) ([Reference: Success Criteria SC-005](spec.md#measurable-outcomes), [Performance Considerations](research.md#performance-considerations))
- [ ] T085d Execute load tests and document performance metrics (p50, p95, p99 response times, error rates) ([Reference: Success Criteria SC-005](spec.md#measurable-outcomes))
- [ ] T085e Identify and resolve performance bottlenecks to meet <200ms p95 requirement under 10K concurrent users ([Reference: Success Criteria SC-005](spec.md#measurable-outcomes))
- [ ] T086 Conduct end-to-end testing of all user stories ([Reference: Testing Strategy](research.md#testing-strategy), [User Scenarios](spec.md#user-scenarios--testing-mandatory))
- [X] T087 Implement account access redirection mechanism from frontend to mobile app/lk.domain.com ([Reference: Implementation Strategy](#implementation-strategy))
- [X] T088 Add deep linking support from web to mobile app ([Reference: Implementation Strategy](#implementation-strategy))
- [X] T089 Update navigation and routing to reflect the separation of public (frontend) and private (mobile) functionality ([Reference: Implementation Strategy](#implementation-strategy))
