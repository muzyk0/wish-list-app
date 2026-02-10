-- name: GetGiftItem :one
SELECT id, wishlist_id, name, description, link, image_url, price, priority, reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at, purchased_price, notes, position, created_at, updated_at
FROM gift_items
WHERE id = $1 LIMIT 1;

-- name: ListGiftItemsByWishList :many
SELECT id, wishlist_id, name, description, link, image_url, price, priority, reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at, purchased_price, notes, position, created_at, updated_at
FROM gift_items
WHERE wishlist_id = $1
ORDER BY position ASC
LIMIT $2 OFFSET $3;

-- name: CreateGiftItem :one
INSERT INTO gift_items (
    wishlist_id, name, description, link, image_url, price, priority, notes, position
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING id, wishlist_id, name, description, link, image_url, price, priority, reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at, purchased_price, notes, position, created_at, updated_at;

-- name: UpdateGiftItem :one
UPDATE gift_items SET
    name = $2,
    description = $3,
    link = $4,
    image_url = $5,
    price = $6,
    priority = $7,
    notes = $8,
    position = $9,
    updated_at = NOW()
WHERE id = $1
RETURNING id, wishlist_id, name, description, link, image_url, price, priority, reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at, purchased_price, notes, position, created_at, updated_at;

-- name: DeleteGiftItem :exec
DELETE FROM gift_items WHERE id = $1;

-- name: ReserveGiftItem :one
UPDATE gift_items SET
    reserved_by_user_id = $2,
    reserved_at = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING id, wishlist_id, name, description, link, image_url, price, priority, reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at, purchased_price, notes, position, created_at, updated_at;

-- name: UnreserveGiftItem :one
UPDATE gift_items SET
    reserved_by_user_id = NULL,
    reserved_at = NULL,
    updated_at = NOW()
WHERE id = $1
RETURNING id, wishlist_id, name, description, link, image_url, price, priority, reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at, purchased_price, notes, position, created_at, updated_at;

-- name: MarkGiftItemAsPurchased :one
UPDATE gift_items SET
    purchased_by_user_id = $2,
    purchased_at = $3,
    purchased_price = $4,
    reserved_by_user_id = NULL,
    reserved_at = NULL,
    updated_at = NOW()
WHERE id = $1
RETURNING id, wishlist_id, name, description, link, image_url, price, priority, reserved_by_user_id, reserved_at, purchased_by_user_id, purchased_at, purchased_price, notes, position, created_at, updated_at;

-- name: ListPublicWishListGiftItems :many
SELECT gi.id, gi.wishlist_id, gi.name, gi.description, gi.link, gi.image_url, gi.price, gi.priority, gi.reserved_by_user_id, gi.reserved_at, gi.purchased_by_user_id, gi.purchased_at, gi.purchased_price, gi.notes, gi.position, gi.created_at, gi.updated_at
FROM gift_items gi
JOIN wishlists w ON gi.wishlist_id = w.id
WHERE w.public_slug = $1 AND w.is_public = true
ORDER BY gi.position ASC
LIMIT $2 OFFSET $3;