-- Add updated_at column to reservations table for tracking modification timestamps
ALTER TABLE reservations
ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

-- Create index on updated_at for efficient timestamp-based queries
CREATE INDEX idx_reservations_updated_at ON reservations(updated_at);
