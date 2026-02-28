# Feature Specification: Public Wishlist with Guest Reservation

**Feature Branch**: `005-public-wishlist-reservation`
**Created**: 2026-02-20
**Status**: Draft
**Input**: User description: "We need to create a wishlist interface with items where a gift can be reserved in the public frontend. Currently, we are implementing reservation as a guest; later we will implement it with authorization within the application or a mobile web version via OAuth. Lists need to be displayed via a specific link that the user specifies themselves in the mobile application. Gifts must definitely be linked to their respective wishlist."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - View a Public Wishlist (Priority: P1)

A visitor receives a link to someone's wishlist (e.g., `example.com/public/birthday-2026`) and opens it in their browser. They see the wishlist title, occasion, description, and a list of gift items with their names, images, prices, and availability status. The page loads without requiring any login or account.

**Why this priority**: This is the core read-only experience. Without being able to view a wishlist, no other functionality (reservation, sharing) has value.

**Independent Test**: Can be fully tested by navigating to a valid public slug URL and verifying wishlist metadata and gift items render correctly. Delivers the ability for anyone to browse a shared wishlist.

**Acceptance Scenarios**:

1. **Given** a public wishlist exists with slug "birthday-2026" and 5 gift items, **When** a visitor navigates to `/public/birthday-2026`, **Then** they see the wishlist title, occasion, description, and all 5 gift items with their details (name, image, price, link, priority).
2. **Given** a gift item has a product link, **When** the visitor views the item, **Then** they see a clickable link that opens the product page in a new tab.
3. **Given** a wishlist slug does not exist or the wishlist is not marked as public, **When** a visitor navigates to that URL, **Then** they see a user-friendly "not found" message.
4. **Given** a wishlist has no gift items, **When** a visitor opens the page, **Then** they see an empty state message indicating no gifts have been added yet.
5. **Given** some items in the wishlist are already reserved or purchased, **When** the visitor views the list, **Then** reserved items show a "Reserved" badge and purchased items show a "Purchased" badge, clearly distinguishing them from available items.

---

### User Story 2 - Reserve a Gift as a Guest (Priority: P1)

A visitor sees an available (unreserved, unpurchased) gift on a public wishlist and decides to reserve it. They click a "Reserve" button, fill in their name and email in a dialog, and submit. The system confirms the reservation and provides a reservation token for future reference.

**Why this priority**: Reservation is the primary interactive action on the public wishlist page and the core value proposition for both the gift-giver and the wishlist owner.

**Independent Test**: Can be fully tested by opening a public wishlist, clicking "Reserve" on an available item, filling in guest details, and verifying the reservation is confirmed. Delivers the ability for guests to claim gifts.

**Acceptance Scenarios**:

1. **Given** a visitor is viewing an available gift item, **When** they click "Reserve Gift", **Then** a dialog opens asking for their name and email.
2. **Given** a visitor has filled in a valid name and email, **When** they submit the reservation form, **Then** the system creates the reservation, shows a success message with the reservation token, and the item's status changes to "Reserved" on the page.
3. **Given** a visitor submits the form with an empty name or invalid email, **When** they try to submit, **Then** they see inline validation errors and the form does not submit.
4. **Given** a gift item is already reserved, **When** a visitor views it, **Then** the "Reserve" button is disabled and shows "Reserved".
5. **Given** a gift item is already purchased, **When** a visitor views it, **Then** the reserve button is disabled and shows "Already Purchased".
6. **Given** a reservation succeeds, **When** the visitor checks their browser storage, **Then** the reservation token and details are persisted locally for future reference.

---

### User Story 3 - View My Reservations as a Guest (Priority: P2)

A guest who has previously reserved gifts can view their reservations on a dedicated "My Reservations" page (`/my/reservations`), accessible via a persistent link in the site header or footer. The system retrieves reservations using the tokens stored locally in the browser, showing the gift name, wishlist name, reservation date, and status.

**Why this priority**: Guests need a way to recall what they reserved, especially if they reserved items across multiple wishlists. This reduces duplicate gifting and supports guest accountability.

**Independent Test**: Can be fully tested by making a reservation, then navigating to a "My Reservations" section and verifying the reserved item appears with correct details.

**Acceptance Scenarios**:

1. **Given** a guest has reserved 2 gifts across different wishlists, **When** they navigate to the reservations view, **Then** they see both reservations with gift name, wishlist title, reservation date, and status.
2. **Given** a guest has no reservations stored locally, **When** they navigate to the reservations view, **Then** they see an empty state with a message like "You haven't reserved any gifts yet".
3. **Given** a guest's reservation token has expired or been canceled, **When** the system fetches reservations, **Then** the expired/canceled reservation shows its updated status accordingly.

---

### User Story 4 - Cancel a Reservation as a Guest (Priority: P3)

A guest who previously reserved a gift decides they can no longer fulfill it. They can cancel the reservation using their reservation token, freeing the item for others.

**Why this priority**: Cancellation is important for flexibility but is a secondary action. Most users will reserve and follow through. This prevents items from being permanently "locked" by guests who change their mind.

**Independent Test**: Can be fully tested by reserving an item, then canceling the reservation and verifying the item returns to "available" status.

**Acceptance Scenarios**:

1. **Given** a guest has a valid reservation, **When** they initiate cancellation and confirm, **Then** the reservation is canceled and the item becomes available again on the public wishlist.
2. **Given** a guest attempts to cancel with an invalid or expired token, **When** the cancellation request is sent, **Then** the system shows an appropriate error message.

---

### Edge Cases

- What happens when two guests try to reserve the same item simultaneously? The system should allow only one reservation to succeed; the second receives a "this item is already reserved" error.
- What happens when a guest clears their browser data? Their local reservation tokens are lost. The reservations still exist on the server but the guest loses the ability to view or cancel them from this browser. A message should guide them to contact the wishlist owner if needed.
- What happens when the wishlist owner removes an item that has been reserved? The reservation should be invalidated and the guest's reservation view should reflect this.
- What happens when the wishlist owner makes a public wishlist private? The public URL should no longer display the wishlist; visitors see a "not found" page.
- What happens when a gift item has no image? A placeholder icon or graphic should be shown instead.
- What happens on a slow or failed network connection during reservation? The user should see loading states and meaningful error messages with an option to retry.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST display a public wishlist with all its gift items when accessed via the wishlist's public slug URL, without requiring authentication.
- **FR-002**: System MUST show each gift item's name, description, price, image (or placeholder), product link, and priority level.
- **FR-003**: System MUST visually indicate the reservation status of each item (available, reserved, or purchased) using distinct badges or labels.
- **FR-004**: System MUST allow unauthenticated visitors to reserve an available gift by providing their name; email is optional.
- **FR-005**: System MUST validate guest name (required, max 200 characters) and email (optional, valid format when provided) before submitting a reservation.
- **FR-006**: System MUST return a reservation token upon successful reservation and persist it in the guest's browser for future reference.
- **FR-007**: System MUST prevent reserving an item that is already reserved or purchased.
- **FR-008**: System MUST provide a dedicated "My Reservations" page at `/my/reservations` where guests can view their previously made reservations using stored reservation tokens. This page MUST be accessible via a persistent link in the site header or footer.
- **FR-009**: System MUST allow guests to cancel their own reservations using the reservation token.
- **FR-010**: System MUST display a user-friendly error page when a wishlist slug is not found or the wishlist is not public.
- **FR-011**: System MUST display loading states while wishlist and item data is being fetched.
- **FR-012**: System MUST sort gift items by their position/order as set by the wishlist owner.
- **FR-013**: System MUST be fully usable on mobile browsers (responsive design).
- **FR-014**: System MUST update the item's visual status immediately after a successful reservation without requiring a full page reload.
- **FR-015**: System MUST support both Russian and English languages with a language switcher. Russian is the default language.
- **FR-016**: The public wishlist page MUST have its own independent visual styling, not bound to the main site's design. The styling architecture MUST support future theming (e.g., holiday or occasion-based designs).

### Constitution Requirements

- **CR-001**: Code Quality - All code MUST meet high standards of quality, maintainability, and readability
- **CR-002**: Test-First - Unit tests MUST be written for all business logic before implementation
- **CR-003**: API Contracts - All API contracts MUST be explicitly defined using OpenAPI/Swagger specifications
- **CR-004**: Data Privacy - No personally identifiable information (PII) MAY be stored without encryption
- **CR-005**: Semantic Versioning - All releases MUST follow semantic versioning (MAJOR.MINOR.PATCH) standards
- **CR-006**: Specification Checkpoints - Features MUST be fully specified before implementation begins

### Key Entities

- **Wishlist**: A collection of desired gifts created by a user. Has a title, description, occasion, occasion date, public/private flag, and a unique public slug for sharing. Gifts are linked to wishlists through a many-to-many relationship.
- **Gift Item**: An individual gift within a wishlist. Has a name, description, price, image, product link, priority, and position. Each item tracks its reservation and purchase status. A gift item belongs to a wishlist owner and can appear in multiple wishlists.
- **Reservation**: A claim on a gift item by a guest or authenticated user. Contains guest name, guest email, reservation token, status (active/canceled), and timestamp. Reservations do not auto-expire; they remain active until explicitly canceled by the guest or the wishlist owner. Each reservation is tied to a specific gift item within a specific wishlist.
- **Guest**: An unauthenticated visitor who can view public wishlists and reserve gifts. Identified by name and email during reservation. Their reservation tokens are stored in the browser's local storage.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Visitors can view a complete public wishlist with all items in under 3 seconds on standard connections.
- **SC-002**: Guests can complete a gift reservation (from clicking "Reserve" to seeing confirmation) in under 30 seconds.
- **SC-003**: 100% of gift items in a wishlist display correct reservation status (available, reserved, or purchased) in real-time after any status change.
- **SC-004**: The public wishlist page is fully functional and readable on screens from 320px width (mobile) to desktop without horizontal scrolling.
- **SC-005**: Guests can view their past reservations and see correct status for each one.
- **SC-006**: Guests can successfully cancel their own reservations, and the item returns to "available" status immediately.
- **SC-007**: Invalid or non-existent wishlist URLs display a clear error page instead of a broken/blank page.

## Clarifications

### Session 2026-02-20

- Q: What language should the public wishlist page use? → A: Both Russian and English with i18n and a language switcher; Russian is the default language.
- Q: Do reservations expire automatically after a certain time? → A: No auto-expiration; reservations stay active until explicitly canceled by the guest or the wishlist owner.
- Q: Where does the "My Reservations" UI live? → A: Separate `/my/reservations` page, accessible via a persistent link in the site header or footer.

## Assumptions

- The backend API already supports all required endpoints for public wishlists, gift items, and reservations (confirmed by OpenAPI spec analysis).
- Wishlist owners create and manage wishlists and gift items exclusively through the mobile application; the frontend (web) is read-only for wishlist data.
- The `public_slug` is set by the wishlist owner in the mobile app and uniquely identifies a public wishlist.
- Guest reservation tokens are stored in browser `localStorage`; if the user clears browser data, they lose access to their reservation history from that browser.
- The currency symbol displayed alongside prices is "$" (USD) as the default. Multi-currency support is out of scope for this feature, but the i18n infrastructure supports future localization.
- The public wishlist page has its own independent styling. Future iterations may allow wishlist owners to select occasion-based themes (e.g., birthday, New Year); for now, a clean default design is sufficient, but the styling architecture should not prevent theming later.
- Email confirmation/notification to guests upon reservation is handled by the backend and is out of scope for the frontend implementation.
- Authenticated reservation (via OAuth) is explicitly out of scope for this feature and will be implemented in a future iteration.
