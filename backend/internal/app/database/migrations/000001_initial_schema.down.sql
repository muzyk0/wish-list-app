-- +migrate Down
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS gift_items;
DROP TABLE IF EXISTS wishlists;
DROP TABLE IF EXISTS templates;
DROP TABLE IF EXISTS users;

-- Disable UUID extension
DROP EXTENSION IF EXISTS "uuid-ossp";