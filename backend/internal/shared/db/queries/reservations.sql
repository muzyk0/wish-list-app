-- name: CreateReservation :one
INSERT INTO reservations (
    gift_item_id, reserved_by_user_id, guest_name, guest_email, status, reserved_at, expires_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING id, gift_item_id, reserved_by_user_id, guest_name, guest_email, reservation_token, status, reserved_at, expires_at, canceled_at, cancel_reason, notification_sent;

-- name: GetReservation :one
SELECT id, gift_item_id, reserved_by_user_id, guest_name, guest_email, reservation_token, status, reserved_at, expires_at, canceled_at, cancel_reason, notification_sent
FROM reservations
WHERE id = $1 LIMIT 1;

-- name: GetReservationByToken :one
SELECT id, gift_item_id, reserved_by_user_id, guest_name, guest_email, reservation_token, status, reserved_at, expires_at, canceled_at, cancel_reason, notification_sent
FROM reservations
WHERE reservation_token = $1 LIMIT 1;

-- name: GetReservationsByGiftItem :many
SELECT id, gift_item_id, reserved_by_user_id, guest_name, guest_email, reservation_token, status, reserved_at, expires_at, canceled_at, cancel_reason, notification_sent
FROM reservations
WHERE gift_item_id = $1
ORDER BY reserved_at DESC;

-- name: GetReservationsByUser :many
SELECT r.id, r.gift_item_id, r.reserved_by_user_id, r.guest_name, r.guest_email, r.reservation_token, r.status, r.reserved_at, r.expires_at, r.canceled_at, r.cancel_reason, r.notification_sent
FROM reservations r
JOIN gift_items gi ON r.gift_item_id = gi.id
JOIN wishlists w ON gi.wishlist_id = w.id
WHERE r.reserved_by_user_id = $1 AND r.status = 'active'
ORDER BY r.reserved_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateReservationStatus :one
UPDATE reservations SET
    status = $2,
    canceled_at = CASE WHEN $2 = 'cancelled' THEN $3 ELSE canceled_at END,
    cancel_reason = CASE WHEN $2 = 'cancelled' THEN $4 ELSE cancel_reason END,
    updated_at = NOW()
WHERE id = $1
RETURNING id, gift_item_id, reserved_by_user_id, guest_name, guest_email, reservation_token, status, reserved_at, expires_at, canceled_at, cancel_reason, notification_sent;

-- name: UpdateReservationStatusByToken :one
UPDATE reservations SET
    status = $2,
    canceled_at = CASE WHEN $2 = 'cancelled' THEN $3 ELSE canceled_at END,
    cancel_reason = CASE WHEN $2 = 'cancelled' THEN $4 ELSE cancel_reason END,
    updated_at = NOW()
WHERE reservation_token = $1
RETURNING id, gift_item_id, reserved_by_user_id, guest_name, guest_email, reservation_token, status, reserved_at, expires_at, canceled_at, cancel_reason, notification_sent;

-- name: GetActiveReservationForGiftItem :one
SELECT id, gift_item_id, reserved_by_user_id, guest_name, guest_email, reservation_token, status, reserved_at, expires_at, canceled_at, cancel_reason, notification_sent
FROM reservations
WHERE gift_item_id = $1 AND status = 'active'
LIMIT 1;

-- name: ListUserReservationsWithDetails :many
SELECT
    r.id,
    r.gift_item_id,
    r.reserved_by_user_id,
    r.guest_name,
    r.guest_email,
    r.reservation_token,
    r.status,
    r.reserved_at,
    r.expires_at,
    r.canceled_at,
    r.cancel_reason,
    r.notification_sent,
    gi.name as gift_item_name,
    gi.image_url as gift_item_image_url,
    gi.price as gift_item_price,
    w.title as wishlist_title,
    u.first_name as owner_first_name,
    u.last_name as owner_last_name
FROM reservations r
JOIN gift_items gi ON r.gift_item_id = gi.id
JOIN wishlists w ON gi.wishlist_id = w.id
LEFT JOIN users u ON w.owner_id = u.id
WHERE r.reserved_by_user_id = $1 AND r.status IN ('active', 'cancelled')
ORDER BY r.reserved_at DESC
LIMIT $2 OFFSET $3;

-- name: ListGuestReservationsWithDetails :many
SELECT
    r.id,
    r.gift_item_id,
    r.reserved_by_user_id,
    r.guest_name,
    r.guest_email,
    r.reservation_token,
    r.status,
    r.reserved_at,
    r.expires_at,
    r.canceled_at,
    r.cancel_reason,
    r.notification_sent,
    gi.name as gift_item_name,
    gi.image_url as gift_item_image_url,
    gi.price as gift_item_price,
    w.title as wishlist_title,
    u.first_name as owner_first_name,
    u.last_name as owner_last_name
FROM reservations r
JOIN gift_items gi ON r.gift_item_id = gi.id
JOIN wishlists w ON gi.wishlist_id = w.id
LEFT JOIN users u ON w.owner_id = u.id
WHERE r.reservation_token = $1 AND r.status IN ('active', 'cancelled')
ORDER BY r.reserved_at DESC;