-- name: GetWishList :one
SELECT id, owner_id, title, description, occasion, occasion_date, template_id, is_public, public_slug, view_count, created_at, updated_at
FROM wishlists
WHERE id = $1 LIMIT 1;

-- name: GetWishListByPublicSlug :one
SELECT id, owner_id, title, description, occasion, occasion_date, template_id, is_public, public_slug, view_count, created_at, updated_at
FROM wishlists
WHERE public_slug = $1 AND is_public = true LIMIT 1;

-- name: ListWishListsByOwner :many
SELECT id, owner_id, title, description, occasion, occasion_date, template_id, is_public, public_slug, view_count, created_at, updated_at
FROM wishlists
WHERE owner_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListPublicWishLists :many
SELECT id, owner_id, title, description, occasion, occasion_date, template_id, is_public, public_slug, view_count, created_at, updated_at
FROM wishlists
WHERE is_public = true
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CreateWishList :one
INSERT INTO wishlists (
    owner_id, title, description, occasion, occasion_date, template_id, is_public, public_slug
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING id, owner_id, title, description, occasion, occasion_date, template_id, is_public, public_slug, view_count, created_at, updated_at;

-- name: UpdateWishList :one
UPDATE wishlists SET
    title = $2,
    description = $3,
    occasion = $4,
    occasion_date = $5,
    template_id = $6,
    is_public = $7,
    public_slug = $8,
    updated_at = NOW()
WHERE id = $1
RETURNING id, owner_id, title, description, occasion, occasion_date, template_id, is_public, public_slug, view_count, created_at, updated_at;

-- name: DeleteWishList :exec
DELETE FROM wishlists WHERE id = $1;

-- name: IncrementWishListViewCount :exec
UPDATE wishlists SET view_count = view_count + 1 WHERE id = $1;