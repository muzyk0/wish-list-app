# Research: Database Initialization Migration

**Feature**: 004-db-init-migration
**Date**: 2026-02-17

## Research Topics

### 1. golang-migrate File Naming Convention

**Decision**: Use `000001_init_schema.up.sql` / `000001_init_schema.down.sql` format.

**Rationale**: golang-migrate requires files named `{version}_{title}.{direction}.sql`. The version must be a numeric prefix. Since the migrations directory is empty, we start at `000001`. Six-digit zero-padded format allows up to 999,999 migrations.

**Alternatives considered**:
- Timestamp-based versioning (e.g., `20260217120000_init_schema.up.sql`) - Rejected: sequential numbering is simpler for a fresh start and is the project's established convention (Makefile uses `migrate create` which defaults to sequential).

### 2. PostgreSQL UUID Generation Strategy

**Decision**: Enable `pgcrypto` extension and use `gen_random_uuid()` as default for UUID primary keys.

**Rationale**: `gen_random_uuid()` is built into PostgreSQL 13+ via `pgcrypto`. The application currently generates UUIDs in Go code, but having the database default provides a safety net and enables direct SQL inserts during development/testing.

**Alternatives considered**:
- `uuid-ossp` extension with `uuid_generate_v4()` - Rejected: `pgcrypto` is lighter and `gen_random_uuid()` is functionally equivalent.
- No database-side UUID generation (app-only) - Rejected: defaults are useful for ad-hoc operations and testing.

### 3. Foreign Key ON DELETE Behavior

**Decision**: Use specific ON DELETE behavior per relationship type:

| FK Column | References | ON DELETE |
|-----------|-----------|-----------|
| `wishlists.owner_id` | `users.id` | CASCADE |
| `gift_items.owner_id` | `users.id` | CASCADE |
| `gift_items.reserved_by_user_id` | `users.id` | SET NULL |
| `gift_items.purchased_by_user_id` | `users.id` | SET NULL |
| `wishlist_items.wishlist_id` | `wishlists.id` | CASCADE |
| `wishlist_items.gift_item_id` | `gift_items.id` | CASCADE |
| `reservations.wishlist_id` | `wishlists.id` | CASCADE |
| `reservations.gift_item_id` | `gift_items.id` | CASCADE |
| `reservations.reserved_by_user_id` | `users.id` | SET NULL |

**Rationale**: Ownership relationships (owner_id) use CASCADE so deleting a user removes their data. Optional references (reserved_by, purchased_by) use SET NULL to preserve the item/reservation record even if the referencing user is deleted. Junction table entries CASCADE with both parents.

**Alternatives considered**:
- RESTRICT on all FKs - Rejected: too rigid, forces manual cleanup before user deletion.
- CASCADE on all FKs - Rejected: would delete reservation history when a guest user is removed.

### 4. Index Strategy

**Decision**: Create indexes on:
- `users.email` (UNIQUE constraint provides implicit index)
- `wishlists.owner_id` (frequent lookups by owner)
- `wishlists.public_slug` (UNIQUE constraint provides implicit index, partial WHERE public_slug IS NOT NULL)
- `gift_items.owner_id` (frequent lookups by owner)
- `reservations.gift_item_id` (join target)
- `reservations.wishlist_id` (join target)
- `reservations.reserved_by_user_id` (lookups for "my reservations")
- `reservations.reservation_token` (UNIQUE constraint provides implicit index)
- `reservations.status` (filter queries)

**Rationale**: Indexes cover all foreign key columns (PostgreSQL doesn't auto-index FK columns unlike some databases) and columns used in WHERE/JOIN clauses in existing repository queries.

**Alternatives considered**:
- Composite indexes (e.g., `reservations(gift_item_id, status)`) - Deferred: premature optimization without query profiling data. Can be added in future migrations.

### 5. Down Migration Strategy

**Decision**: Drop tables in reverse dependency order using `DROP TABLE IF EXISTS ... CASCADE`.

**Rationale**: CASCADE handles any remaining FK constraints. `IF EXISTS` ensures idempotency. Reverse order (reservations → wishlist_items → gift_items → wishlists → users) respects dependency chain even without CASCADE.

**Alternatives considered**:
- Single `DROP SCHEMA public CASCADE; CREATE SCHEMA public;` - Rejected: too destructive, removes other objects (extensions, types) and would break golang-migrate's schema_migrations table.

### 6. Extension Management in Down Migration

**Decision**: Do NOT drop `pgcrypto` extension in down migration.

**Rationale**: Other parts of the system or other migrations may depend on `pgcrypto`. Dropping it could break unrelated functionality. Extensions are lightweight and shared resources.

**Alternatives considered**:
- Drop `pgcrypto` in down migration - Rejected: shared extension, removing could break other consumers.
