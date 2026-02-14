-- Revert reservation column name fixes

ALTER TABLE reservations RENAME COLUMN cancel_reason TO cancelled_reason;
ALTER TABLE reservations RENAME COLUMN canceled_at TO cancelled_at;
