# Tasks: Database Initialization Migration

**Input**: Design documents from `/specs/004-db-init-migration/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, quickstart.md

**Tests**: No automated test tasks included. Verification is manual (migrate up/down) and structural (model-to-SQL comparison).

**Organization**: Tasks follow the 3 user stories from spec.md. Due to the nature of this feature (2 SQL files), most tasks map to a single deliverable but are broken down for traceability.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Migration files**: `backend/internal/app/database/migrations/`
- **Model files**: `backend/internal/domain/{name}/models/`
- **Migration runner**: `backend/cmd/migrate/main.go` (existing, no changes needed)

---

## Phase 1: Setup

**Purpose**: Ensure migration infrastructure is ready

- [ ] T001 Verify migrations directory exists and is empty at `backend/internal/app/database/migrations/`
- [ ] T002 Verify migration runner works by running `cd backend && go run cmd/migrate/main.go -action version` (should report no version or error gracefully)

**Checkpoint**: Migration infrastructure confirmed working

---

## Phase 2: User Story 1 - Developer Deploys Fresh Database (Priority: P1) MVP

**Goal**: Create the up migration that builds the complete database schema from scratch.

**Independent Test**: Run `make migrate-up` against an empty PostgreSQL database and verify all 5 tables exist with correct columns, constraints, and indexes.

### Implementation for User Story 1

- [ ] T003 [US1] Create `backend/internal/app/database/migrations/000001_init_schema.up.sql` with pgcrypto extension enablement: `CREATE EXTENSION IF NOT EXISTS pgcrypto`
- [ ] T004 [US1] Add `users` table DDL to `backend/internal/app/database/migrations/000001_init_schema.up.sql` with all 14 columns matching `backend/internal/domain/user/models/user.go` User struct (see data-model.md for exact types, constraints, defaults)
- [ ] T005 [US1] Add `wishlists` table DDL to `backend/internal/app/database/migrations/000001_init_schema.up.sql` with all 11 columns matching `backend/internal/domain/wishlist/models/wishlist.go` WishList struct, FK to users(id) ON DELETE CASCADE
- [ ] T006 [US1] Add `gift_items` table DDL to `backend/internal/app/database/migrations/000001_init_schema.up.sql` with all 17 columns matching `backend/internal/domain/item/models/item.go` GiftItem struct, FKs to users(id) with CASCADE/SET NULL per research.md
- [ ] T007 [US1] Add `wishlist_items` junction table DDL to `backend/internal/app/database/migrations/000001_init_schema.up.sql` with composite PK (wishlist_id, gift_item_id), FKs to wishlists(id) and gift_items(id) ON DELETE CASCADE, matching `backend/internal/domain/wishlist_item/models/wishlist_item.go`
- [ ] T008 [US1] Add `reservations` table DDL to `backend/internal/app/database/migrations/000001_init_schema.up.sql` with all 16 columns matching `backend/internal/domain/reservation/models/reservation.go` Reservation struct, FKs per research.md
- [ ] T009 [US1] Add explicit indexes to `backend/internal/app/database/migrations/000001_init_schema.up.sql`: idx_wishlists_owner_id, idx_gift_items_owner_id, idx_reservations_gift_item_id, idx_reservations_wishlist_id, idx_reservations_reserved_by, idx_reservations_status (see data-model.md Indexes table for full list)
- [ ] T010 [US1] Run `make migrate-up` against empty database and verify all 5 tables created with `\dt` in psql

**Checkpoint**: Up migration complete. Fresh database deployment works.

---

## Phase 3: User Story 2 - Developer Rolls Back Migration (Priority: P2)

**Goal**: Create the down migration that cleanly removes all database objects.

**Independent Test**: Run `make migrate-up` then `make migrate-down` and verify database is empty.

### Implementation for User Story 2

- [ ] T011 [US2] Create `backend/internal/app/database/migrations/000001_init_schema.down.sql` with DROP TABLE statements in reverse dependency order: reservations, wishlist_items, gift_items, wishlists, users (all with IF EXISTS and CASCADE)
- [ ] T012 [US2] Run `make migrate-down` after `make migrate-up` and verify all tables are removed
- [ ] T013 [US2] Run `make migrate-up` again after rollback to verify re-apply works (idempotency via golang-migrate versioning)

**Checkpoint**: Full migrate up → down → up cycle works cleanly.

---

## Phase 4: User Story 3 - Model-Migration Consistency Verification (Priority: P2)

**Goal**: Verify 1:1 correspondence between Go model struct `db:` tags and migration SQL columns.

**Independent Test**: Side-by-side comparison of every struct field against migration columns.

### Verification for User Story 3

- [ ] T014 [US3] Compare `backend/internal/domain/user/models/user.go` User struct db tags against `users` table columns in migration — verify all 14 columns match in name and compatible PostgreSQL type
- [ ] T015 [P] [US3] Compare `backend/internal/domain/wishlist/models/wishlist.go` WishList struct db tags against `wishlists` table columns in migration — verify all 11 columns match
- [ ] T016 [P] [US3] Compare `backend/internal/domain/item/models/item.go` GiftItem struct db tags against `gift_items` table columns in migration — verify all 17 columns match
- [ ] T017 [P] [US3] Compare `backend/internal/domain/wishlist_item/models/wishlist_item.go` WishlistItem struct db tags against `wishlist_items` table columns in migration — verify all 3 columns match
- [ ] T018 [P] [US3] Compare `backend/internal/domain/reservation/models/reservation.go` Reservation struct db tags against `reservations` table columns in migration — verify all 16 columns match

**Checkpoint**: All 5 model structs verified against migration SQL. Zero drift.

---

## Phase 5: Polish & Cross-Cutting Concerns

**Purpose**: Final validation and cleanup

- [ ] T019 Add SQL comments to `backend/internal/app/database/migrations/000001_init_schema.up.sql` documenting table purposes and non-obvious design decisions (PII encryption columns, junction table rationale, soft delete pattern)
- [ ] T020 Validate quickstart.md instructions by following them end-to-end on a clean database
- [ ] T021 Run full application (`make backend`) after migration to verify CRUD operations work against the new schema

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - start immediately
- **US1 (Phase 2)**: Depends on Setup — creates the up migration file
- **US2 (Phase 3)**: Depends on US1 — needs the up migration to exist before testing rollback
- **US3 (Phase 4)**: Depends on US1 — needs the up migration SQL to compare against models
- **Polish (Phase 5)**: Depends on US1 and US2 completion

### User Story Dependencies

- **US1 (P1)**: Independent after Setup. Creates the up migration. **BLOCKS US2 and US3.**
- **US2 (P2)**: Depends on US1 (needs up migration to test rollback)
- **US3 (P2)**: Depends on US1 (needs up migration SQL to compare). **Can run in parallel with US2.**

### Within User Story 1

- T003 → T004 → T005 → T006 → T007 → T008 → T009 → T010 (sequential: single file, dependency order)

### Parallel Opportunities

- **US2 and US3 can run in parallel** after US1 completes (different concerns: rollback testing vs. struct comparison)
- **T015, T016, T017, T018 can run in parallel** within US3 (independent model comparisons)

---

## Parallel Example: User Story 3

```text
# After US1 is complete, launch all model comparisons in parallel:
Task: "Compare wishlist model against migration" (T015)
Task: "Compare gift_item model against migration" (T016)
Task: "Compare wishlist_item model against migration" (T017)
Task: "Compare reservation model against migration" (T018)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T002)
2. Complete Phase 2: User Story 1 (T003-T010)
3. **STOP and VALIDATE**: `make migrate-up` creates all tables correctly
4. Database is immediately usable for development

### Incremental Delivery

1. US1 → Database can be deployed to any environment
2. US2 → Rollback capability confirmed for safe CI/CD
3. US3 → Schema-code consistency verified, no drift
4. Polish → Documentation and full app validation

---

## Notes

- This feature creates exactly **2 files**: `000001_init_schema.up.sql` and `000001_init_schema.down.sql`
- No Go code changes needed — existing migration runner handles the new files automatically
- All column definitions must reference `data-model.md` as the authoritative source
- FK ON DELETE behavior must follow `research.md` decisions (CASCADE for ownership, SET NULL for optional references)
- Do NOT drop pgcrypto extension in down migration (shared resource)
- Commit after completing each user story phase
