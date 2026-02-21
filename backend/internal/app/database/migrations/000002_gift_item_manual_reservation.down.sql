-- Revert manual reservation fields from gift_items
ALTER TABLE gift_items
    DROP COLUMN IF EXISTS manual_reserved_by_name,
    DROP COLUMN IF EXISTS manual_reservation_note,
    DROP COLUMN IF EXISTS manual_reserved_at;
