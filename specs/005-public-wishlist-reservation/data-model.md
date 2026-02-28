# Data Model: Public Wishlist with Guest Reservation

**Feature**: 005-public-wishlist-reservation
**Date**: 2026-02-20

> This feature is frontend-only. The backend data model already exists. This document describes the frontend data shapes as consumed from the API.

## Entities

### WishList (read-only from API)

Retrieved via `GET /public/wishlists/{slug}`.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string (UUID) | Yes | Unique identifier |
| title | string | Yes | Wishlist title (max 200 chars) |
| description | string | No | Wishlist description |
| occasion | string | No | Occasion type (birthday, wedding, etc.) |
| occasion_date | string (ISO date) | No | Date of the occasion |
| is_public | boolean | Yes | Must be true for public access |
| public_slug | string | Yes | Unique URL-friendly slug |
| owner_id | string (UUID) | Yes | Wishlist creator's user ID |
| item_count | integer | No | Number of gift items |
| view_count | string | Yes | Number of page views |
| created_at | string (ISO datetime) | Yes | Creation timestamp |
| updated_at | string (ISO datetime) | Yes | Last update timestamp |

### GiftItem (read-only from API)

Retrieved via `GET /public/wishlists/{slug}/gift-items` (paginated).

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string (UUID) | Yes | Unique identifier |
| name | string | Yes | Gift item name |
| description | string | No | Item description |
| price | number | No | Price in USD |
| image_url | string | No | URL to item image |
| link | string | No | Product URL |
| notes | string | No | Additional notes from owner |
| priority | integer | No | Priority level (0-10) |
| position | integer | No | Display order in wishlist |
| wishlist_id | string (UUID) | Yes | Parent wishlist ID |
| reserved_at | string (ISO datetime) | No | When item was reserved (null = available) |
| reserved_by_user_id | string (UUID) | No | Who reserved (null = available) |
| is_reserved | boolean | Yes | Aggregated reservation flag (covers guest/auth/manual reservation) |
| purchased_at | string (ISO datetime) | No | When item was purchased (null = not purchased) |
| purchased_by_user_id | string (UUID) | No | Who purchased |
| purchased_price | number | No | Actual purchase price |
| created_at | string (ISO datetime) | Yes | Creation timestamp |
| updated_at | string (ISO datetime) | Yes | Last update timestamp |

**Derived States**:
- `available`: `is_reserved === false && purchased_by_user_id === null`
- `reserved`: `is_reserved === true && purchased_by_user_id === null`
- `purchased`: `purchased_by_user_id !== null`

### Reservation (created via API, stored partially in localStorage)

Created via `POST /reservations/wishlist/{wishlistId}/item/{itemId}`.

**API Response Shape** (`CreateReservationResponse`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string (UUID) | Yes | Reservation identifier |
| gift_item_id | string (UUID) | Yes | Reserved item |
| status | string | Yes | "active" or "canceled" |
| reservation_token | string | Yes | Token for guest access |
| reserved_at | string (ISO datetime) | Yes | Reservation timestamp |
| guest_name | string | No | Guest's name |
| guest_email | string | No | Guest's email |
| reserved_by_user_id | string (UUID) | No | Auth user ID (null for guests) |
| notification_sent | boolean | Yes | Whether email was sent |
| canceled_at | string (ISO datetime) | No | Cancellation timestamp |
| cancel_reason | string | No | Cancellation reason |
| expires_at | string (ISO datetime) | No | Not used (no auto-expiration) |

**Guest Reservation Details** (from `GET /guest/reservations?token=X`):

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string (UUID) | Yes | Reservation ID |
| status | string | Yes | "active" or "canceled" |
| reserved_at | string (ISO datetime) | Yes | Reservation timestamp |
| expires_at | string (ISO datetime) | No | Not used |
| gift_item | GiftItemSummary | Yes | Summary of reserved item |
| wishlist | WishListSummary | Yes | Summary of parent wishlist |

**GiftItemSummary**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string (UUID) | Yes | Item ID |
| name | string | Yes | Item name |
| image_url | string | No | Item image |
| price | string | No | Item price |

**WishListSummary**:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| id | string (UUID) | Yes | Wishlist ID |
| title | string | Yes | Wishlist title |
| owner_first_name | string | No | Owner's first name |
| owner_last_name | string | No | Owner's last name |

### LocalStorageReservation (client-side only)

Stored in `localStorage` under key `guest_reservations` as a JSON array.

| Field | Type | Description |
|-------|------|-------------|
| itemId | string (UUID) | Reserved item ID |
| itemName | string | Item name (for offline display) |
| reservationToken | string | Server-issued token |
| reservedAt | string (ISO datetime) | When reserved |
| guestName | string | Guest's name |
| wishlistId | string (UUID) | Parent wishlist ID |

## State Transitions

### Gift Item Availability (from guest's perspective)

```
Available ──(guest reserves)──> Reserved
Available ──(owner marks purchased)──> Purchased
Reserved ──(guest cancels)──> Available
Reserved ──(owner marks purchased)──> Purchased
```

### Reservation Lifecycle

```
(none) ──(POST create)──> Active
Active ──(DELETE cancel)──> Canceled
```

No auto-expiration. Reservations remain active until explicitly canceled.

## Relationships

```
WishList (1) ──────── (N) GiftItem       via GET /public/wishlists/{slug}/gift-items
GiftItem (1) ──────── (0..1) Reservation  one active reservation per item per wishlist
Guest    (1) ──────── (N) Reservation     a guest can reserve multiple items
```
