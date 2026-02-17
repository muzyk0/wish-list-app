# Implementation Plan: Database Initialization Migration

**Branch**: `004-db-init-migration` | **Date**: 2026-02-17 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/004-db-init-migration/spec.md`

## Summary

Create a single initialization migration (up/down) that defines the complete database schema for all 5 existing domain entities. The migration must match 1:1 with the Go model structs in `backend/internal/domain/*/models/`. No API changes, no new code beyond the SQL migration files.

## Technical Context

**Language/Version**: Go 1.25.5 (migration runner), SQL (migration files)
**Primary Dependencies**: golang-migrate/v4, lib/pq (migration driver)
**Storage**: PostgreSQL 14+ with pgcrypto extension
**Testing**: Manual verification via `make migrate-up` / `make migrate-down`, struct-to-SQL comparison
**Target Platform**: Linux server (production), macOS (development)
**Project Type**: Web application (backend component only)
**Performance Goals**: Migration completes in under 5 seconds
**Constraints**: Must match existing Go model structs exactly; no schema drift
**Scale/Scope**: 5 tables, ~70 columns total, 9 foreign keys, 5 explicit indexes

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Requirement | Status | Notes |
|------------|--------|-------|
| Code Quality | PASS | SQL will be well-formatted and commented |
| Test-First Approach | PASS | Migration verified against model structs before merge |
| API Contract Integrity | PASS | No API changes; schema supports existing contracts |
| Data Privacy Protection | PASS | Schema includes encrypted_* columns for all PII fields |
| Semantic Versioning | N/A | Infrastructure change, no release version impact |
| Specification Checkpoints | PASS | Spec completed and clarified before planning |

## Project Structure

### Documentation (this feature)

```text
specs/004-db-init-migration/
├── plan.md              # This file
├── research.md          # Phase 0 output - decisions and rationale
├── data-model.md        # Phase 1 output - complete schema definition
├── quickstart.md        # Phase 1 output - usage instructions
└── checklists/
    └── requirements.md  # Specification quality checklist
```

### Source Code (repository root)

```text
backend/internal/app/database/migrations/
├── 000001_init_schema.up.sql    # CREATE tables, indexes, constraints
└── 000001_init_schema.down.sql  # DROP tables in reverse dependency order
```

**Structure Decision**: Only 2 SQL files are created. No Go code changes needed - the existing migration runner (`backend/cmd/migrate/main.go`) already reads from the `migrations/` directory.

## Implementation Phases

### Phase 1: Write Up Migration

Create `000001_init_schema.up.sql` with:

1. Enable `pgcrypto` extension (`CREATE EXTENSION IF NOT EXISTS pgcrypto`)
2. Create `users` table (14 columns)
3. Create `wishlists` table (11 columns) with FK to users
4. Create `gift_items` table (17 columns) with FKs to users
5. Create `wishlist_items` junction table (3 columns) with composite PK and FKs
6. Create `reservations` table (16 columns) with FKs to users, wishlists, gift_items
7. Create explicit indexes on FK columns and query-targeted columns

**Order matters**: Tables must be created in dependency order (users → wishlists/gift_items → wishlist_items → reservations).

### Phase 2: Write Down Migration

Create `000001_init_schema.down.sql` with:

1. Drop tables in reverse dependency order:
   - `DROP TABLE IF EXISTS reservations CASCADE`
   - `DROP TABLE IF EXISTS wishlist_items CASCADE`
   - `DROP TABLE IF EXISTS gift_items CASCADE`
   - `DROP TABLE IF EXISTS wishlists CASCADE`
   - `DROP TABLE IF EXISTS users CASCADE`
2. Do NOT drop pgcrypto extension (shared resource)

### Phase 3: Verification

1. Compare every `db:"column_name"` tag in Go models against SQL columns
2. Verify type compatibility (pgtype.UUID ↔ UUID, pgtype.Text ↔ TEXT, etc.)
3. Run `make migrate-up` against empty database
4. Run `make migrate-down` to verify clean rollback
5. Run `make migrate-up` again to verify idempotency via golang-migrate

## Type Mapping Reference

| Go (pgtype) | PostgreSQL | Notes |
|-------------|-----------|-------|
| pgtype.UUID | UUID | gen_random_uuid() default |
| string | VARCHAR(255) | For NOT NULL text fields |
| pgtype.Text | TEXT | Nullable text |
| pgtype.Bool | BOOLEAN | |
| pgtype.Int4 | INTEGER | |
| pgtype.Numeric | NUMERIC(12,2) | Price fields |
| pgtype.Date | DATE | |
| pgtype.Timestamptz | TIMESTAMPTZ | |

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|-----------|--------|------------|
| Column mismatch with Go structs | Low | High | Side-by-side comparison in Phase 3 |
| Wrong FK ON DELETE behavior | Low | Medium | Documented in research.md, explicit per-FK |
| Missing index on hot query path | Medium | Low | Can add indexes in future migration |
| pgcrypto not available | Very Low | High | PostgreSQL 13+ includes it by default |
