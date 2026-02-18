# Data Model: Database Initialization Migration

**Feature**: 004-db-init-migration
**Date**: 2026-02-17

## Entity Relationship Diagram

```
┌──────────────┐       ┌──────────────────┐       ┌──────────────────┐
│    users     │       │    wishlists      │       │   gift_items     │
│──────────────│       │──────────────────│       │──────────────────│
│ id (PK)      │◄──┐   │ id (PK)          │       │ id (PK)          │
│ email        │   ├───│ owner_id (FK)     │   ┌──│ owner_id (FK)    │
│ encrypted_*  │   │   │ title             │   │  │ name             │
│ password_hash│   │   │ description       │   │  │ reserved_by (FK) │──┐
│ first_name   │   │   │ occasion          │   │  │ purchased_by (FK)│──┤
│ ...          │   │   │ occasion_date     │   │  │ ...              │  │
│ created_at   │   │   │ is_public         │   │  │ archived_at      │  │
│ updated_at   │   │   │ public_slug       │   │  │ created_at       │  │
└──────────────┘   │   │ view_count        │   │  │ updated_at       │  │
       ▲           │   │ created_at        │   │  └──────────────────┘  │
       │           │   │ updated_at        │   │          ▲             │
       │           │   └──────────────────┘   │          │             │
       │           │          ▲                │          │             │
       │           │          │                │  ┌──────────────────┐  │
       │           │          │                │  │ wishlist_items   │  │
       │           │          │                │  │──────────────────│  │
       │           │          └────────────────┼──│ wishlist_id (FK) │  │
       │           │                           └──│ gift_item_id(FK) │  │
       │           │                              │ added_at         │  │
       │           │                              └──────────────────┘  │
       │           │                                                    │
       │           │   ┌──────────────────┐                             │
       │           │   │  reservations    │                             │
       │           │   │──────────────────│                             │
       │           ├───│ reserved_by (FK) │◄────────────────────────────┘
       │           │   │ wishlist_id (FK) │
       │           │   │ gift_item_id(FK) │
       │           │   │ guest_name       │
       │           │   │ encrypted_*      │
       │           │   │ reservation_token│
       │           │   │ status           │
       │           │   │ ...              │
       │           │   └──────────────────┘
       │           │
       └───────────┘
```

## Tables

### 1. users

| Column | Type | Constraints | Default | Notes |
|--------|------|-------------|---------|-------|
| id | UUID | PK | gen_random_uuid() | |
| email | VARCHAR(255) | NOT NULL, UNIQUE | | Login lookup field |
| encrypted_email | TEXT | | | PII encrypted copy |
| password_hash | TEXT | | | Nullable for magic-link users |
| first_name | TEXT | | | |
| encrypted_first_name | TEXT | | | PII encrypted copy |
| last_name | TEXT | | | |
| encrypted_last_name | TEXT | | | PII encrypted copy |
| avatar_url | TEXT | | | |
| is_verified | BOOLEAN | NOT NULL | false | |
| created_at | TIMESTAMPTZ | NOT NULL | NOW() | |
| updated_at | TIMESTAMPTZ | NOT NULL | NOW() | |
| last_login_at | TIMESTAMPTZ | | | |
| deactivated_at | TIMESTAMPTZ | | | Soft deactivation |

**Go Model**: `backend/internal/domain/user/models/user.go` → `User` struct

### 2. wishlists

| Column | Type | Constraints | Default | Notes |
|--------|------|-------------|---------|-------|
| id | UUID | PK | gen_random_uuid() | |
| owner_id | UUID | NOT NULL, FK → users(id) ON DELETE CASCADE | | |
| title | VARCHAR(255) | NOT NULL | | |
| description | TEXT | | | |
| occasion | TEXT | | | |
| occasion_date | DATE | | | |
| is_public | BOOLEAN | NOT NULL | false | |
| public_slug | TEXT | UNIQUE | | Nullable, unique when set |
| view_count | INTEGER | NOT NULL | 0 | |
| created_at | TIMESTAMPTZ | NOT NULL | NOW() | |
| updated_at | TIMESTAMPTZ | NOT NULL | NOW() | |

**Go Model**: `backend/internal/domain/wishlist/models/wishlist.go` → `WishList` struct

### 3. gift_items

| Column | Type | Constraints | Default | Notes |
|--------|------|-------------|---------|-------|
| id | UUID | PK | gen_random_uuid() | |
| owner_id | UUID | NOT NULL, FK → users(id) ON DELETE CASCADE | | Item owner |
| name | VARCHAR(255) | NOT NULL | | |
| description | TEXT | | | |
| link | TEXT | | | URL to product |
| image_url | TEXT | | | |
| price | NUMERIC(12,2) | | | |
| priority | INTEGER | | | |
| reserved_by_user_id | UUID | FK → users(id) ON DELETE SET NULL | | |
| reserved_at | TIMESTAMPTZ | | | |
| purchased_by_user_id | UUID | FK → users(id) ON DELETE SET NULL | | |
| purchased_at | TIMESTAMPTZ | | | |
| purchased_price | NUMERIC(12,2) | | | |
| notes | TEXT | | | |
| position | INTEGER | | | Ordering within wishlist |
| archived_at | TIMESTAMPTZ | | | Soft delete |
| created_at | TIMESTAMPTZ | NOT NULL | NOW() | |
| updated_at | TIMESTAMPTZ | NOT NULL | NOW() | |

**Go Model**: `backend/internal/domain/item/models/item.go` → `GiftItem` struct

### 4. wishlist_items (Junction Table)

| Column | Type | Constraints | Default | Notes |
|--------|------|-------------|---------|-------|
| wishlist_id | UUID | PK, FK → wishlists(id) ON DELETE CASCADE | | |
| gift_item_id | UUID | PK, FK → gift_items(id) ON DELETE CASCADE | | |
| added_at | TIMESTAMPTZ | NOT NULL | NOW() | |

**Composite PK**: (wishlist_id, gift_item_id)

**Go Model**: `backend/internal/domain/wishlist_item/models/wishlist_item.go` → `WishlistItem` struct

### 5. reservations

| Column | Type | Constraints | Default | Notes |
|--------|------|-------------|---------|-------|
| id | UUID | PK | gen_random_uuid() | |
| wishlist_id | UUID | NOT NULL, FK → wishlists(id) ON DELETE CASCADE | | Reservation context |
| gift_item_id | UUID | NOT NULL, FK → gift_items(id) ON DELETE CASCADE | | |
| reserved_by_user_id | UUID | FK → users(id) ON DELETE SET NULL | | Nullable for guests |
| guest_name | TEXT | | | |
| encrypted_guest_name | TEXT | | | PII encrypted |
| guest_email | TEXT | | | |
| encrypted_guest_email | TEXT | | | PII encrypted |
| reservation_token | UUID | UNIQUE | | Anonymous access token |
| status | VARCHAR(50) | NOT NULL | 'active' | active/canceled/fulfilled/expired |
| reserved_at | TIMESTAMPTZ | NOT NULL | NOW() | |
| expires_at | TIMESTAMPTZ | | | |
| canceled_at | TIMESTAMPTZ | | | |
| cancel_reason | TEXT | | | |
| notification_sent | BOOLEAN | NOT NULL | false | |
| updated_at | TIMESTAMPTZ | NOT NULL | NOW() | |

**Go Model**: `backend/internal/domain/reservation/models/reservation.go` → `Reservation` struct

## Indexes

| Table | Index Name | Columns | Type | Notes |
|-------|-----------|---------|------|-------|
| users | users_pkey | id | PK (implicit) | |
| users | users_email_key | email | UNIQUE (implicit) | |
| wishlists | wishlists_pkey | id | PK (implicit) | |
| wishlists | idx_wishlists_owner_id | owner_id | BTREE | FK lookups |
| wishlists | wishlists_public_slug_key | public_slug | UNIQUE (implicit) | |
| gift_items | gift_items_pkey | id | PK (implicit) | |
| gift_items | idx_gift_items_owner_id | owner_id | BTREE | FK lookups |
| wishlist_items | wishlist_items_pkey | (wishlist_id, gift_item_id) | PK (implicit) | Composite |
| reservations | reservations_pkey | id | PK (implicit) | |
| reservations | idx_reservations_gift_item_id | gift_item_id | BTREE | JOIN target |
| reservations | idx_reservations_wishlist_id | wishlist_id | BTREE | JOIN target |
| reservations | idx_reservations_reserved_by | reserved_by_user_id | BTREE | "My reservations" queries |
| reservations | reservations_token_key | reservation_token | UNIQUE (implicit) | |
| reservations | idx_reservations_status | status | BTREE | Filter queries |

## Foreign Key Summary

| Source Table | Column | References | ON DELETE |
|-------------|--------|-----------|-----------|
| wishlists | owner_id | users(id) | CASCADE |
| gift_items | owner_id | users(id) | CASCADE |
| gift_items | reserved_by_user_id | users(id) | SET NULL |
| gift_items | purchased_by_user_id | users(id) | SET NULL |
| wishlist_items | wishlist_id | wishlists(id) | CASCADE |
| wishlist_items | gift_item_id | gift_items(id) | CASCADE |
| reservations | wishlist_id | wishlists(id) | CASCADE |
| reservations | gift_item_id | gift_items(id) | CASCADE |
| reservations | reserved_by_user_id | users(id) | SET NULL |
