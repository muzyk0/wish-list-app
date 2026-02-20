-- Migration: 000001_init_schema (DOWN)
-- Purpose: Rollback database schema initialization
-- Date: 2026-02-17
-- Note: Tables dropped in reverse dependency order to respect foreign key constraints

-- Drop tables in reverse dependency order
-- (deepest dependencies first, working up to users)

-- ============================================================================
-- Drop reservations (depends on wishlists, gift_items, users)
-- ============================================================================
DROP TABLE IF EXISTS reservations CASCADE;

-- ============================================================================
-- Drop wishlist_items junction table (depends on wishlists, gift_items)
-- ============================================================================
DROP TABLE IF EXISTS wishlist_items CASCADE;

-- ============================================================================
-- Drop gift_items (depends on users)
-- ============================================================================
DROP TABLE IF EXISTS gift_items CASCADE;

-- ============================================================================
-- Drop wishlists (depends on users)
-- ============================================================================
DROP TABLE IF EXISTS wishlists CASCADE;

-- ============================================================================
-- Drop users (no dependencies)
-- ============================================================================
DROP TABLE IF EXISTS users CASCADE;

-- ============================================================================
-- Extension Management
-- ============================================================================
-- Note: We do NOT drop pgcrypto extension here
-- Rationale: Other migrations or system features may depend on it
-- Dropping shared extensions can break unrelated functionality
