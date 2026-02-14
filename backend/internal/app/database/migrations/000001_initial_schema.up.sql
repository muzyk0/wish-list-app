-- +migrate Up
-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    avatar_url TEXT,
    is_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_login_at TIMESTAMPTZ,
    deactivated_at TIMESTAMPTZ
);

-- Indexes for Users
CREATE INDEX idx_users_email ON users(email);

-- Templates table
CREATE TABLE templates (
    id VARCHAR(50) PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    preview_image_url TEXT,
    config JSONB NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default template
INSERT INTO templates (id, name, description, config, is_default, created_at, updated_at)
VALUES ('default', 'Default Template', 'Default wish list template', '{}', TRUE, NOW(), NOW());

-- WishLists table
CREATE TABLE wishlists (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    occasion VARCHAR(100),
    occasion_date DATE,
    template_id VARCHAR(50) NOT NULL DEFAULT 'default' REFERENCES templates(id),
    is_public BOOLEAN DEFAULT FALSE,
    public_slug VARCHAR(255) UNIQUE,
    view_count INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for WishLists
CREATE INDEX idx_wishlists_owner_id ON wishlists(owner_id);
CREATE INDEX idx_wishlists_public_slug ON wishlists(public_slug) WHERE is_public = TRUE;

-- GiftItems table
CREATE TABLE gift_items (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    wishlist_id UUID NOT NULL REFERENCES wishlists(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    link TEXT,
    image_url TEXT,
    price DECIMAL(10,2),
    priority INTEGER DEFAULT 0,
    reserved_by_user_id UUID REFERENCES users(id),
    reserved_at TIMESTAMPTZ,
    purchased_by_user_id UUID REFERENCES users(id),
    purchased_at TIMESTAMPTZ,
    purchased_price DECIMAL(10,2),
    notes TEXT,
    position INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes for GiftItems
CREATE INDEX idx_gift_items_wishlist_id ON gift_items(wishlist_id);
CREATE INDEX idx_gift_items_reserved_by_user_id ON gift_items(reserved_by_user_id);
CREATE INDEX idx_gift_items_position ON gift_items(position);

-- Reservations table
CREATE TABLE reservations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    gift_item_id UUID NOT NULL REFERENCES gift_items(id) ON DELETE CASCADE,
    reserved_by_user_id UUID REFERENCES users(id),
    guest_name VARCHAR(200),
    guest_email VARCHAR(255),
    reservation_token UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4(),
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled', 'fulfilled', 'expired')),
    reserved_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMPTZ,
    cancelled_at TIMESTAMPTZ,
    cancelled_reason TEXT,
    notification_sent BOOLEAN DEFAULT FALSE
);

-- Indexes for Reservations
CREATE INDEX idx_reservations_gift_item_id ON reservations(gift_item_id);
CREATE INDEX idx_reservations_reservation_token ON reservations(reservation_token);

-- Add constraint to ensure gift items cannot be both reserved and purchased simultaneously
ALTER TABLE gift_items ADD CONSTRAINT chk_not_reserved_and_purchased 
CHECK (
    NOT (reserved_by_user_id IS NOT NULL AND purchased_by_user_id IS NOT NULL)
);

-- Add constraint to ensure either reserved_by_user_id or (guest_name and guest_email) for reservations
ALTER TABLE reservations ADD CONSTRAINT chk_reservation_identity 
CHECK (
    (reserved_by_user_id IS NOT NULL) OR 
    (guest_name IS NOT NULL AND guest_email IS NOT NULL)
);