-- +migrate Down
-- Remove encrypted PII fields from reservations table
ALTER TABLE reservations DROP COLUMN IF EXISTS encrypted_guest_email;
ALTER TABLE reservations DROP COLUMN IF EXISTS encrypted_guest_name;

-- Remove encrypted PII fields from users table
ALTER TABLE users DROP COLUMN IF EXISTS encrypted_last_name;
ALTER TABLE users DROP COLUMN IF EXISTS encrypted_first_name;
ALTER TABLE users DROP COLUMN IF EXISTS encrypted_email;
