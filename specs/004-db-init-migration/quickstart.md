# Quickstart: Database Initialization Migration

**Feature**: 004-db-init-migration

## Prerequisites

- PostgreSQL 14+ running locally (or via Docker: `make db-up`)
- Go 1.25.5+ installed
- `DATABASE_URL` environment variable set (or `.env` file in backend/)

## Apply Migration

```bash
# Start database if not running
make db-up

# Run the initialization migration
make migrate-up
```

## Verify Migration

```bash
# Connect to database and check tables
psql $DATABASE_URL -c "\dt"

# Expected output: 5 tables
#  users
#  wishlists
#  gift_items
#  wishlist_items
#  reservations
```

## Rollback Migration

```bash
# Remove all tables created by the migration
make migrate-down
```

## Check Migration Version

```bash
cd backend && go run cmd/migrate/main.go -action version
```

## Files Created

```
backend/internal/app/database/migrations/
├── 000001_init_schema.up.sql    # Creates all tables, indexes, constraints
└── 000001_init_schema.down.sql  # Drops all tables in reverse dependency order
```
