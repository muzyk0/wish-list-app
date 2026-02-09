-- Remove index on updated_at
DROP INDEX IF EXISTS idx_reservations_updated_at;

-- Remove updated_at column from reservations table
ALTER TABLE reservations
DROP COLUMN IF EXISTS updated_at;
