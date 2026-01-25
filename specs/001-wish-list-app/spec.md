# Feature Specification: Wish List Application

**Feature Branch**: `001-wish-list-app`
**Created**: 2026-01-12
**Status**: Ready for Implementation
**Input**: User description: "Build a wish-list application for personal use, friends and family. Public site: holiday page, list of gifts, gift reservation (book a gift to avoid duplicates), links, photos, descriptions. Personal account: create and edit wish-lists, choose templates, add purchases, mark items reserved/purchased. Public pages are viewable without auth. Success: minimal friction to create a gift list and avoid duplicate gifts."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create and Share a Gift List (Priority: P1)

A user wants to create an account via mobile app and create a wish list for an occasion (like a birthday or holiday) and share it with friends and family so they can see what gifts are desired and avoid buying duplicates.

**Why this priority**: This is the core functionality of the application - without the ability to create and share wish lists, the app has no value.

**Independent Test**: Can be fully tested by creating a wish list with multiple gift items, sharing the public link, and verifying that others can view the list and reserve gifts without registering.

**Acceptance Scenarios**:

1. **Given** a user has created an account and logged in, **When** they create a new wish list with several gift items (including links, photos, descriptions), **Then** the wish list is created and a public sharing link is generated
2. **Given** a public wish list exists, **When** a visitor (unauthenticated user) visits the public link, **Then** they can view all gift items and see which ones are reserved by others
3. **Given** a gift item is available on a public wish list, **When** a visitor clicks "Reserve this gift", **Then** the gift is marked as reserved and no longer available for others to reserve

---

### User Story 2 - Manage Personal Wish Lists (Priority: P2)

A user wants to create, edit, and manage multiple wish lists from their personal account, including adding/removing items, marking items as purchased, and choosing different templates for presentation.

**Why this priority**: This provides the personal management aspect that allows users to maintain their wish lists over time.

**Independent Test**: Can be fully tested by logging into a personal account, creating multiple wish lists, adding and editing gift items, and verifying all management functions work properly.

**Acceptance Scenarios**:

1. **Given** a user is logged into their account, **When** they create a new wish list, **Then** a new empty wish list is created with default template
2. **Given** a user has a wish list, **When** they add a gift item with link, photo, and description, **Then** the item is added to the list and visible on the public page
3. **Given** a user has reserved a gift on their own list, **When** they mark it as purchased, **Then** the status updates and is reflected on the public view

---

### User Story 3 - Gift Reservation and Tracking (Priority: P3)

A friend or family member wants to browse public wish lists, reserve gifts to avoid duplicates, and track the status of gifts they've reserved.

**Why this priority**: This provides the social aspect that prevents duplicate gifts, which is a key value proposition of the app.

**Independent Test**: Can be fully tested by browsing public wish lists without an account, reserving gifts, and verifying the reservation status is properly tracked.

**Acceptance Scenarios**:

1. **Given** a public wish list exists with available gifts, **When** a visitor reserves a gift, **Then** the gift is marked as reserved with their name and cannot be reserved by others
2. **Given** a gift is reserved by someone, **When** another visitor tries to reserve the same gift, **Then** they receive a clear message that the gift is already reserved
3. **Given** a user has reserved gifts, **When** they visit their reserved gifts section, **Then** they can see all gifts they've reserved across different lists

---

### Edge Cases

- What happens when a user deletes a wish list that has items reserved by others?
- How does the system handle image uploads that exceed size limits or unsupported formats?
- What happens when a wish list owner marks a reserved gift as purchased?
- How does the system handle concurrent reservations of the same gift?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow users to create accounts with authentication via the mobile application, with account access available through both mobile app and web redirect mechanisms
- **FR-002**: System MUST allow authenticated users to create and manage multiple wish lists
- **FR-003**: System MUST allow users to add gift items to wish lists with links, photos, and descriptions
- **FR-004**: System MUST generate public URLs for wish lists that can be shared
- **FR-005**: System MUST allow visitors (unauthenticated users) to view public wish lists
- **FR-006**: System MUST allow visitors to reserve gift items to prevent duplicates
- **FR-007**: System MUST display reservation status clearly on public wish lists
- **FR-008**: System MUST allow wish list owners to mark items as purchased
- **FR-009**: System MUST provide at least 3 distinct presentation templates for wish lists. Each template MUST offer configurable options for: (1) layout structure (grid, list, or card-based views), (2) color schemes (minimum 3 different palettes per template), and (3) typography choices (minimum 2 font families per template). Templates are considered distinct if they differ in at least one of: layout structure, color palette options, or design elements (spacing, visual components, decorative elements).
- **FR-010**: System MUST store and serve images for gift items
- **FR-011**: System MUST support image uploads up to 10MB in size with supported formats: JPEG, PNG, GIF (static and animated), WEBP. Animated GIFs MUST be preserved as animated when uploaded.
- **FR-012**: System MUST retain user data for 2 years after account inactivity before automatic deletion with prior notification
- **FR-013**: When a wish list owner removes a gift item that has active reservations, the system MUST notify all reservation holders via email that their reserved gift is no longer available
- **FR-015**: System MUST provide a mobile web interface at lk.domain.com that offers the same account management functionality as the mobile app
- **FR-016**: System MUST encrypt all personally identifiable information (PII) at rest, including user email addresses, user names, guest reservation names, and guest reservation emails using industry-standard encryption (AES-256 or equivalent)
- **FR-017**: When a wish list owner deletes a wish list that contains gift items with active reservations, the system MUST: (1) prevent deletion and display a warning showing the count of active reservations, (2) provide an option to "Delete anyway and notify reservation holders", (3) if proceeding, send email notifications to all reservation holders that their reserved gifts are no longer available with the list name and gift item details
- **FR-018**: When an image upload fails due to exceeding the 10MB size limit, the system MUST return HTTP 413 (Payload Too Large) with error message "Image size exceeds 10MB limit. Please upload a smaller image." When an unsupported format is uploaded, the system MUST return HTTP 415 (Unsupported Media Type) with error message "Unsupported image format. Please upload JPEG, PNG, GIF, or WEBP."
- **FR-019**: When a wish list owner marks a reserved gift item as purchased, the system MUST: (1) update the gift item status to "purchased", (2) send email notification to the reservation holder confirming their gift was received, (3) maintain the reservation record for historical tracking, (4) display "Purchased - Thank you [reserver name]!" on the public wish list view

### Constitution Requirements

- **CR-001**: Code Quality - All code MUST meet high standards of quality, maintainability, and readability
- **CR-002**: Test-First - Unit tests MUST be written for all business logic before implementation
- **CR-003**: API Contracts - All API contracts MUST be explicitly defined using OpenAPI/Swagger specifications
- **CR-004**: Data Privacy - No personally identifiable information (PII) MAY be stored without encryption
- **CR-005**: Semantic Versioning - All releases MUST follow semantic versioning (MAJOR.MINOR.PATCH) standards
- **CR-006**: Specification Checkpoints - Features MUST be fully specified before implementation begins

### Key Entities

- **User**: An account holder who can create and manage wish lists; has profile information and authentication details
- **Wish List**: A collection of gift items for a specific occasion; has owner, title, description, template, and sharing settings
- **Gift Item**: An item on a wish list; has name, description, link, photo, price, reservation status, and purchase status
- **Reservation**: A record of a guest reserving a gift item; has guest identifier, reserved item, timestamp, and status
- **Template**: A presentation style for wish lists; has layout, colors, and design elements

### Glossary

- **User**: An authenticated account holder who can create and manage wish lists
- **Visitor**: An unauthenticated person viewing public wish lists and reserving gifts (may provide name/email for reservation tracking without creating an account)
- **Guest Reservation**: A reservation made by a visitor without authentication, tracked by email/name only
- **Authenticated Reservation**: A reservation made by a logged-in user (future functionality)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can create a new wish list with 5 gift items in under 5 minutes
- **SC-002**: 95% of users successfully complete the wish list creation process on their first attempt
- **SC-003**: Users report 80% reduction in duplicate gifts received after using the app for one holiday season
- **SC-004**: At least 70% of gift items on public lists have clear reservation status visible to visitors
- **SC-005**: System supports up to 10,000 concurrent users browsing public wish lists without performance degradation, where "concurrent user" is defined as a unique session making an average of 10 HTTP requests per minute (viewing lists, reserving gifts, loading images). Performance degradation is defined as p95 response time exceeding 200ms or error rate exceeding 0.1%.
- **SC-006**: Users can reserve a gift from a public list in 2 clicks or less