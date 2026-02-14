-- +migrate Up
-- Add encrypted PII fields for users table
-- Store encrypted versions of sensitive data

ALTER TABLE users ADD COLUMN encrypted_email TEXT;
ALTER TABLE users ADD COLUMN encrypted_first_name TEXT;
ALTER TABLE users ADD COLUMN encrypted_last_name TEXT;

-- Add encrypted PII fields for reservations table
-- Store encrypted versions of guest information

ALTER TABLE reservations ADD COLUMN encrypted_guest_name TEXT;
ALTER TABLE reservations ADD COLUMN encrypted_guest_email TEXT;

-- Create indexes for encrypted fields to support lookups
-- Note: These indexes are on the encrypted values, so they won't support
-- equality searches on plaintext. Consider using hashing for searchable encryption
-- if needed in the future.

-- Add comment to document encryption requirement
COMMENT ON COLUMN users.encrypted_email IS 'AES-256 encrypted email address for PII protection (CR-004)';
COMMENT ON COLUMN users.encrypted_first_name IS 'AES-256 encrypted first name for PII protection (CR-004)';
COMMENT ON COLUMN users.encrypted_last_name IS 'AES-256 encrypted last name for PII protection (CR-004)';
COMMENT ON COLUMN reservations.encrypted_guest_name IS 'AES-256 encrypted guest name for PII protection (CR-004)';
COMMENT ON COLUMN reservations.encrypted_guest_email IS 'AES-256 encrypted guest email for PII protection (CR-004)';
