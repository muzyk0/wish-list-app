-- Rollback migration: Revert gift_items to one-to-many with wishlists
-- WARNING: This will lose many-to-many associations!
-- If an item is in multiple wishlists, only the first association will be preserved.

-- Step 1: Add wishlist_id back to gift_items
ALTER TABLE gift_items
ADD COLUMN wishlist_id UUID REFERENCES wishlists(id) ON DELETE CASCADE;

-- Step 2: Populate wishlist_id from wishlist_items (take first association)
UPDATE gift_items gi
SET wishlist_id = (
    SELECT wishlist_id
    FROM wishlist_items wi
    WHERE wi.gift_item_id = gi.id
    ORDER BY wi.added_at ASC
    LIMIT 1
);

-- Step 3: Make wishlist_id NOT NULL (items must belong to a wishlist)
-- Note: Items without wishlist association will be deleted!
DELETE FROM gift_items WHERE wishlist_id IS NULL;

ALTER TABLE gift_items
ALTER COLUMN wishlist_id SET NOT NULL;

-- Step 4: Drop owner_id and archived_at from gift_items
ALTER TABLE gift_items
DROP COLUMN owner_id;

ALTER TABLE gift_items
DROP COLUMN archived_at;

-- Step 5: Remove wishlist_id from reservations
ALTER TABLE reservations
DROP COLUMN wishlist_id;

-- Step 6: Drop wishlist_items join table
DROP TABLE wishlist_items;

-- Step 7: Recreate the original index
CREATE INDEX idx_gift_items_wishlist_id ON gift_items(wishlist_id);
