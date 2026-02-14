-- Fix typos in reservation column names
-- Change cancelled_at to canceled_at (American spelling)
-- Change cancelled_reason to cancel_reason (consistency with codebase)

ALTER TABLE reservations RENAME COLUMN cancelled_at TO canceled_at;
ALTER TABLE reservations RENAME COLUMN cancelled_reason TO cancel_reason;
