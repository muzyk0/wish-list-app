-- Migration: 000001_init_schema
-- Purpose: Initialize complete database schema for wish list application
-- Date: 2026-02-17

-- Enable pgcrypto extension for UUID generation and encryption functions
CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ============================================================================
-- Table: users
-- Purpose: Store user account information with PII encryption support
-- ============================================================================
CREATE TABLE users (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email                VARCHAR(255) NOT NULL UNIQUE,
    encrypted_email      TEXT,                      -- PII encrypted copy of email
    password_hash        TEXT,                       -- Nullable for magic-link/OAuth users
    first_name           TEXT,
    encrypted_first_name TEXT,                       -- PII encrypted copy of first name
    last_name            TEXT,
    encrypted_last_name  TEXT,                       -- PII encrypted copy of last name
    avatar_url           TEXT,
    is_verified          BOOLEAN NOT NULL DEFAULT false,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at        TIMESTAMPTZ,
    deactivated_at       TIMESTAMPTZ                 -- Soft deactivation timestamp
);

-- ============================================================================
-- Table: wishlists
-- Purpose: Store wishlist metadata owned by users
-- ============================================================================
CREATE TABLE wishlists (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id       UUID NOT NULL,
    title          VARCHAR(255) NOT NULL,
    description    TEXT,
    occasion       TEXT,
    occasion_date  DATE,
    is_public      BOOLEAN NOT NULL DEFAULT false,
    public_slug    TEXT UNIQUE,                    -- Nullable, unique when set
    view_count     INTEGER NOT NULL DEFAULT 0,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_wishlists_owner
        FOREIGN KEY (owner_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- ============================================================================
-- Table: gift_items
-- Purpose: Store gift items owned by users, can be added to multiple wishlists
-- Note: Items belong to users, not to specific wishlists (many-to-many via junction table)
-- ============================================================================
CREATE TABLE gift_items (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id              UUID NOT NULL,
    name                  VARCHAR(255) NOT NULL,
    description           TEXT,
    link                  TEXT,
    image_url             TEXT,
    price                 NUMERIC(12,2),
    priority              INTEGER,
    reserved_by_user_id   UUID,
    reserved_at           TIMESTAMPTZ,
    purchased_by_user_id  UUID,
    purchased_at          TIMESTAMPTZ,
    purchased_price       NUMERIC(12,2),
    notes                 TEXT,
    position              INTEGER,                  -- Ordering within wishlist
    archived_at           TIMESTAMPTZ,              -- Soft delete timestamp
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_gift_items_owner
        FOREIGN KEY (owner_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_gift_items_reserved_by
        FOREIGN KEY (reserved_by_user_id)
        REFERENCES users(id)
        ON DELETE SET NULL,

    CONSTRAINT fk_gift_items_purchased_by
        FOREIGN KEY (purchased_by_user_id)
        REFERENCES users(id)
        ON DELETE SET NULL
);

-- ============================================================================
-- Table: wishlist_items (Junction Table)
-- Purpose: Many-to-many relationship between wishlists and gift items
-- Note: Composite primary key ensures each item appears only once per wishlist
-- ============================================================================
CREATE TABLE wishlist_items (
    wishlist_id  UUID NOT NULL,
    gift_item_id UUID NOT NULL,
    added_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    PRIMARY KEY (wishlist_id, gift_item_id),

    CONSTRAINT fk_wishlist_items_wishlist
        FOREIGN KEY (wishlist_id)
        REFERENCES wishlists(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_wishlist_items_gift_item
        FOREIGN KEY (gift_item_id)
        REFERENCES gift_items(id)
        ON DELETE CASCADE
);

-- ============================================================================
-- Table: reservations
-- Purpose: Track reservations for gift items in specific wishlists
-- Note: Supports both authenticated users and guests (via guest_name/guest_email)
-- ============================================================================
CREATE TABLE reservations (
    id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wishlist_id           UUID NOT NULL,               -- Context: which wishlist this reservation is for
    gift_item_id          UUID NOT NULL,
    reserved_by_user_id   UUID,                        -- Nullable for guest reservations
    guest_name            TEXT,
    encrypted_guest_name  TEXT,                        -- PII encrypted copy
    guest_email           TEXT,
    encrypted_guest_email TEXT,                        -- PII encrypted copy
    reservation_token     UUID UNIQUE,                 -- Anonymous access token
    status                VARCHAR(50) NOT NULL DEFAULT 'active',
    reserved_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at            TIMESTAMPTZ,
    canceled_at           TIMESTAMPTZ,
    cancel_reason         TEXT,
    notification_sent     BOOLEAN NOT NULL DEFAULT false,
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_reservations_wishlist
        FOREIGN KEY (wishlist_id)
        REFERENCES wishlists(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_reservations_gift_item
        FOREIGN KEY (gift_item_id)
        REFERENCES gift_items(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_reservations_reserved_by
        FOREIGN KEY (reserved_by_user_id)
        REFERENCES users(id)
        ON DELETE SET NULL
);

-- ============================================================================
-- Indexes
-- Purpose: Optimize query performance for common access patterns
-- Note: UNIQUE constraints create implicit indexes, explicit indexes for FK columns
-- ============================================================================

-- Wishlists indexes
CREATE INDEX idx_wishlists_owner_id ON wishlists(owner_id);

-- Gift items indexes
CREATE INDEX idx_gift_items_owner_id ON gift_items(owner_id);

-- Reservations indexes
CREATE INDEX idx_reservations_gift_item_id ON reservations(gift_item_id);
CREATE INDEX idx_reservations_wishlist_id ON reservations(wishlist_id);
CREATE INDEX idx_reservations_reserved_by ON reservations(reserved_by_user_id);
CREATE INDEX idx_reservations_status ON reservations(status);

-- ============================================================================
-- Schema Notes
-- ============================================================================
-- 1. PII Encryption: encrypted_* columns store encrypted copies of personal data
-- 2. Soft Delete: archived_at and deactivated_at allow soft deletion without data loss
-- 3. Junction Table: wishlist_items enables items to belong to multiple wishlists
-- 4. ON DELETE Behavior:
--    - CASCADE: Owner relationships (delete user → delete their data)
--    - SET NULL: Optional references (preserve history when referenced user deleted)
-- 5. Status Values: reservations.status = 'active'|'canceled'|'fulfilled'|'expired'
