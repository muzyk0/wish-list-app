-- Add automatic UUID generation for reservation tokens
ALTER TABLE reservations
  ALTER COLUMN reservation_token SET DEFAULT gen_random_uuid();

-- Backfill any existing NULL tokens (e.g., created before this fix)
UPDATE reservations
  SET reservation_token = gen_random_uuid()
  WHERE reservation_token IS NULL;
