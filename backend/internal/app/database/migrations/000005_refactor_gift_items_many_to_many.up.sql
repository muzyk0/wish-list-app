-- Migration: Refactor gift_items to support many-to-many with wishlists (FIXED VERSION)
-- This migration:
-- 1. Creates wishlist_items join table
-- 2. Adds owner_id to gift_items
-- 3. Adds archived_at for soft delete
-- 4. Migrates existing wishlist_id associations to join table
-- 5. Removes wishlist_id from gift_items
-- 6. Updates reservations with proper conflict handling

-- Step 1: Create wishlist_items join table for many-to-many
CREATE TABLE wishlist_items (
                                wishlist_id UUID NOT NULL REFERENCES wishlists(id) ON DELETE CASCADE,
                                gift_item_id UUID NOT NULL REFERENCES gift_items(id) ON DELETE CASCADE,
                                added_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                PRIMARY KEY (wishlist_id, gift_item_id)
);

CREATE INDEX idx_wishlist_items_wishlist_id ON wishlist_items(wishlist_id);
CREATE INDEX idx_wishlist_items_gift_item_id ON wishlist_items(gift_item_id);

-- Step 2: Add owner_id to gift_items (will be populated from wishlist owner)
ALTER TABLE gift_items
    ADD COLUMN owner_id UUID REFERENCES users(id) ON DELETE CASCADE;

-- Step 3: Add archived_at for soft delete
ALTER TABLE gift_items
    ADD COLUMN archived_at TIMESTAMPTZ NULL DEFAULT NULL;

CREATE INDEX idx_gift_items_archived_at ON gift_items(archived_at);
CREATE INDEX idx_gift_items_owner_id ON gift_items(owner_id);

-- Step 4: Populate owner_id from wishlist owner
UPDATE gift_items gi
SET owner_id = w.owner_id
FROM wishlists w
WHERE gi.wishlist_id = w.id;

-- Step 5: Migrate existing wishlist_id associations to join table
INSERT INTO wishlist_items (wishlist_id, gift_item_id, added_at)
SELECT wishlist_id, id, created_at
FROM gift_items
WHERE wishlist_id IS NOT NULL;

-- Step 6: Make owner_id NOT NULL (all items should have owner now)
-- First check if any items don't have owner_id
DO $$
    DECLARE
        orphan_count INT;
    BEGIN
        SELECT COUNT(*) INTO orphan_count
        FROM gift_items
        WHERE owner_id IS NULL;

        IF orphan_count > 0 THEN
            RAISE EXCEPTION 'Migration failed: % gift items have no owner_id. Manual intervention required.', orphan_count;
        END IF;
    END $$;

ALTER TABLE gift_items
    ALTER COLUMN owner_id SET NOT NULL;

-- Step 7: Remove wishlist_id foreign key and column
ALTER TABLE gift_items
    DROP CONSTRAINT gift_items_wishlist_id_fkey;

ALTER TABLE gift_items
    DROP COLUMN wishlist_id;

-- Step 8: Update reservations to reference wishlist+item pair
-- Add wishlist_id to reservations table
ALTER TABLE reservations
    ADD COLUMN wishlist_id UUID REFERENCES wishlists(id) ON DELETE CASCADE;

-- FIX #1: Check for items with multiple wishlists that have reservations
DO $$
    DECLARE
        conflict_count INT;
        conflict_items TEXT;
    BEGIN
        -- Count items that are in multiple wishlists AND have active reservations
        SELECT COUNT(DISTINCT gi.id), STRING_AGG(DISTINCT gi.name, ', ')
        INTO conflict_count, conflict_items
        FROM gift_items gi
                 INNER JOIN wishlist_items wi ON wi.gift_item_id = gi.id
                 INNER JOIN reservations r ON r.gift_item_id = gi.id
        WHERE r.wishlist_id IS NULL  -- Only check unassigned reservations
        GROUP BY gi.id
        HAVING COUNT(DISTINCT wi.wishlist_id) > 1;

        IF conflict_count > 0 THEN
            RAISE WARNING 'Found % items with ambiguous wishlist assignments: %', conflict_count, conflict_items;
            RAISE WARNING 'These reservations will be assigned to the first wishlist found.';
            RAISE WARNING 'Consider manually verifying reservation assignments after migration.';
        END IF;
    END $$;

-- Populate wishlist_id for reservations
-- For items in multiple wishlists, this picks the first one (by added_at)
UPDATE reservations r
SET wishlist_id = (
    SELECT wi.wishlist_id
    FROM wishlist_items wi
    WHERE wi.gift_item_id = r.gift_item_id
    ORDER BY wi.added_at ASC  -- Use oldest association
    LIMIT 1
)
WHERE r.wishlist_id IS NULL;

-- FIX #2: Check for reservations that couldn't get a wishlist_id
DO $$
    DECLARE
        null_count INT;
        orphan_items TEXT;
    BEGIN
        SELECT COUNT(*), STRING_AGG(DISTINCT gi.name, ', ')
        INTO null_count, orphan_items
        FROM reservations r
                 LEFT JOIN gift_items gi ON gi.id = r.gift_item_id
        WHERE r.wishlist_id IS NULL;

        IF null_count > 0 THEN
            RAISE EXCEPTION 'Migration failed: % reservations could not be assigned to wishlists. Items: %. Manual intervention required.',
                null_count, COALESCE(orphan_items, 'deleted items');
        END IF;
    END $$;

-- Make wishlist_id NOT NULL (safe now after validation)
ALTER TABLE reservations
    ALTER COLUMN wishlist_id SET NOT NULL;

-- Create index for reservations lookups
CREATE INDEX idx_reservations_wishlist_id ON reservations(wishlist_id);

-- Step 9: Add unique constraint to prevent duplicate active reservations per wishlist+item
CREATE UNIQUE INDEX idx_reservations_unique_active
    ON reservations(wishlist_id, gift_item_id, reserved_by_user_id)
    WHERE status = 'active';

-- Step 10: Create a log table for migration audit
CREATE TABLE IF NOT EXISTS migration_audit_log (
                                                   id SERIAL PRIMARY KEY,
                                                   migration_version VARCHAR(50),
                                                   event_type VARCHAR(50),
                                                   message TEXT,
                                                   created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Log successful migration
INSERT INTO migration_audit_log (migration_version, event_type, message)
VALUES ('000005', 'SUCCESS', 'Successfully migrated gift_items to many-to-many structure');

-- Display migration summary
DO $$
    DECLARE
        total_items INT;
        total_wishlists INT;
        total_associations INT;
        total_reservations INT;
    BEGIN
        SELECT COUNT(*) INTO total_items FROM gift_items;
        SELECT COUNT(*) INTO total_wishlists FROM wishlists;
        SELECT COUNT(*) INTO total_associations FROM wishlist_items;
        SELECT COUNT(*) INTO total_reservations FROM reservations;

        RAISE NOTICE '=== Migration Summary ===';
        RAISE NOTICE 'Total gift items: %', total_items;
        RAISE NOTICE 'Total wishlists: %', total_wishlists;
        RAISE NOTICE 'Total wishlist-item associations: %', total_associations;
        RAISE NOTICE 'Total reservations: %', total_reservations;
        RAISE NOTICE '========================';
    END $$;
