# Feature Specification: Database Initialization Migration

**Feature Branch**: `004-db-init-migration`
**Created**: 2026-02-17
**Status**: Draft
**Input**: User description: "It is necessary to check all database models @backend/internal/domain/< ... >/models/ (as example), add initialization migration for all current entities in the database."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Developer Deploys Fresh Database (Priority: P1)

A developer or DevOps engineer sets up the application on a new environment (local, staging, production). They run the database migration tool, and all required tables, indexes, and constraints are created automatically from a single initialization migration file. The database is ready for the application to use without any manual SQL execution.

**Why this priority**: Without the initialization migration, there is no reproducible way to create the database schema. This is the foundational requirement that enables all other database operations.

**Independent Test**: Can be fully tested by running `make migrate-up` against an empty PostgreSQL database and verifying all 5 tables exist with correct columns, constraints, and indexes.

**Acceptance Scenarios**:

1. **Given** an empty PostgreSQL database, **When** the initialization migration is run, **Then** all 5 tables (users, wishlists, gift_items, wishlist_items, reservations) are created with correct columns and types.
2. **Given** an empty PostgreSQL database, **When** the initialization migration is run, **Then** all primary keys, foreign keys, unique constraints, and indexes are created.
3. **Given** a database with the initialization migration already applied, **When** the migration is run again, **Then** it is skipped (idempotent via golang-migrate versioning).

---

### User Story 2 - Developer Rolls Back Migration (Priority: P2)

A developer needs to roll back the initialization migration (e.g., during development or testing). They run the down migration, and all tables and related objects are cleanly removed.

**Why this priority**: Rollback capability is essential for safe development workflow and CI/CD pipelines.

**Independent Test**: Can be tested by running `make migrate-up` followed by `make migrate-down` and verifying the database is empty.

**Acceptance Scenarios**:

1. **Given** a database with the initialization migration applied, **When** the down migration is run, **Then** all tables, indexes, and constraints are removed.
2. **Given** a database with data in all tables, **When** the down migration is run, **Then** all tables are dropped (CASCADE) without errors.

---

### User Story 3 - Model-Migration Consistency Verification (Priority: P2)

All Go model structs in `backend/internal/domain/*/models/` accurately correspond to the columns defined in the initialization migration. There are no missing columns, extra columns, or type mismatches between the code and the migration SQL.

**Why this priority**: Schema-code mismatches cause runtime errors that are difficult to diagnose.

**Independent Test**: Can be verified by comparing each Go struct's `db:` tags against the migration SQL column definitions.

**Acceptance Scenarios**:

1. **Given** the Go model structs and the migration SQL, **When** compared side by side, **Then** every `db:"column_name"` tag in every model struct has a corresponding column in the migration with a compatible PostgreSQL type.
2. **Given** the migration SQL, **When** compared against model structs, **Then** no column exists in the migration that lacks a corresponding struct field (and vice versa).

---

### Edge Cases

- What happens when the migration is run against a database that already has some tables manually created? (golang-migrate versioning handles this - migration either runs fully or not at all)
- How does the system handle running the up migration on a database with an incompatible PostgreSQL version? (PostgreSQL 14+ assumed based on feature usage)
- What happens if the down migration is run when tables have cross-table foreign key references? (CASCADE drop order handles this)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST include a numbered migration file pair (up/down) in `backend/internal/app/database/migrations/` following golang-migrate naming convention (`000001_init_schema.up.sql` / `000001_init_schema.down.sql`).
- **FR-002**: The up migration MUST create the `users` table with all columns matching the User model struct: id (UUID, PK), email (VARCHAR, UNIQUE, NOT NULL), encrypted_email (TEXT), password_hash (TEXT), first_name (TEXT), encrypted_first_name (TEXT), last_name (TEXT), encrypted_last_name (TEXT), avatar_url (TEXT), is_verified (BOOLEAN, DEFAULT false), created_at (TIMESTAMPTZ, DEFAULT NOW()), updated_at (TIMESTAMPTZ, DEFAULT NOW()), last_login_at (TIMESTAMPTZ), deactivated_at (TIMESTAMPTZ).
- **FR-003**: The up migration MUST create the `wishlists` table with all columns matching the WishList model struct: id (UUID, PK), owner_id (UUID, FK to users, NOT NULL), title (VARCHAR, NOT NULL), description (TEXT), occasion (TEXT), occasion_date (DATE), is_public (BOOLEAN, DEFAULT false), public_slug (TEXT, UNIQUE), view_count (INTEGER, DEFAULT 0), created_at (TIMESTAMPTZ, DEFAULT NOW()), updated_at (TIMESTAMPTZ, DEFAULT NOW()).
- **FR-004**: The up migration MUST create the `gift_items` table with all columns matching the GiftItem model struct: id (UUID, PK), owner_id (UUID, FK to users, NOT NULL), name (VARCHAR, NOT NULL), description (TEXT), link (TEXT), image_url (TEXT), price (NUMERIC), priority (INTEGER), reserved_by_user_id (UUID, FK to users), reserved_at (TIMESTAMPTZ), purchased_by_user_id (UUID, FK to users), purchased_at (TIMESTAMPTZ), purchased_price (NUMERIC), notes (TEXT), position (INTEGER), archived_at (TIMESTAMPTZ), created_at (TIMESTAMPTZ, DEFAULT NOW()), updated_at (TIMESTAMPTZ, DEFAULT NOW()).
- **FR-005**: The up migration MUST create the `wishlist_items` junction table with columns: wishlist_id (UUID, FK to wishlists, NOT NULL), gift_item_id (UUID, FK to gift_items, NOT NULL), added_at (TIMESTAMPTZ, DEFAULT NOW()), with a composite primary key on (wishlist_id, gift_item_id).
- **FR-006**: The up migration MUST create the `reservations` table with all columns matching the Reservation model struct: id (UUID, PK), wishlist_id (UUID, FK to wishlists, NOT NULL), gift_item_id (UUID, FK to gift_items, NOT NULL), reserved_by_user_id (UUID, FK to users), guest_name (TEXT), encrypted_guest_name (TEXT), guest_email (TEXT), encrypted_guest_email (TEXT), reservation_token (UUID, UNIQUE), status (VARCHAR, NOT NULL, DEFAULT 'active'), reserved_at (TIMESTAMPTZ, DEFAULT NOW()), expires_at (TIMESTAMPTZ), canceled_at (TIMESTAMPTZ), cancel_reason (TEXT), notification_sent (BOOLEAN, DEFAULT false), updated_at (TIMESTAMPTZ, DEFAULT NOW()).
- **FR-007**: The up migration MUST create appropriate indexes for foreign key columns and frequently queried columns (owner_id, public_slug, reservation_token, status, email).
- **FR-008**: The up migration MUST enable the `pgcrypto` extension for UUID generation support via `gen_random_uuid()`.
- **FR-009**: The down migration MUST drop all tables in reverse dependency order (reservations, wishlist_items, gift_items, wishlists, users) using CASCADE.
- **FR-010**: All foreign key constraints MUST use appropriate ON DELETE behavior (CASCADE for ownership relationships, SET NULL for optional references).

### Constitution Requirements

- **CR-001**: Code Quality - Migration SQL MUST be well-formatted, commented, and follow PostgreSQL best practices.
- **CR-002**: Test-First - Migration MUST be verified against model structs before merging.
- **CR-003**: API Contracts - Database schema MUST align with existing API contract expectations.
- **CR-006**: Specification Checkpoints - Feature MUST be fully specified before implementation begins.

### Key Entities

- **Users**: Application users with PII encryption support (encrypted_email, encrypted_first_name, encrypted_last_name). Central entity referenced by wishlists, items, and reservations.
- **Wishlists**: Collections of gift items owned by a user, with optional public access via unique slug. Tracks view counts and supports occasion dates.
- **Gift Items**: Individual gift entries owned by a user, associated with wishlists via junction table. Supports reservation tracking, purchase tracking, ordering (position), and soft deletion (archived_at).
- **Wishlist Items**: Junction table enabling many-to-many relationship between wishlists and gift items. Tracks when items were added.
- **Reservations**: Records of item reservations by authenticated users or anonymous guests, scoped to a specific wishlist context. Supports PII encryption for guest data, expiration, and cancellation tracking.

## Assumptions

- PostgreSQL 14+ is the target database version.
- golang-migrate is the migration tool (already configured in the project via Makefile).
- The migration numbering starts at `000001` since the migrations directory is currently empty.
- UUID primary keys are generated application-side, but `pgcrypto` extension is enabled for flexibility with `gen_random_uuid()`.
- Default values for `created_at`/`updated_at` use `NOW()` for convenience, though the application may override them.
- `email` in the users table is stored as plaintext for login lookup; `encrypted_email` is the encrypted copy for display/export.
- Tables are created in dependency order: users first (no FK dependencies), then wishlists and gift_items (depend on users), then wishlist_items (depends on both), then reservations (depends on all).

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Running `make migrate-up` on an empty database creates all 5 tables with 100% column coverage matching the Go model structs.
- **SC-002**: Running `make migrate-down` after `migrate-up` cleanly removes all database objects without errors.
- **SC-003**: Every `db:"column_name"` tag across all 5 model structs has a corresponding column in the migration SQL with a compatible PostgreSQL type.
- **SC-004**: The migration completes in under 5 seconds on a standard development machine.
- **SC-005**: The application starts successfully and can perform CRUD operations on all entities after the migration is applied.
