-- Add manual reservation fields to gift_items
-- Used when the wishlist owner marks an item as reserved by someone offline
-- (e.g., "Grandma said she'll buy the bicycle")
ALTER TABLE gift_items
    ADD COLUMN manual_reserved_by_name VARCHAR(255) NULL,
    ADD COLUMN manual_reservation_note TEXT NULL,
    ADD COLUMN manual_reserved_at TIMESTAMPTZ NULL;
