# Frontend API Client Contract

**Feature**: 005-public-wishlist-reservation
**Date**: 2026-02-20

> The backend API is already defined in `/api/openapi3.yaml`. This document describes the frontend API client methods needed for this feature.

## Existing Methods (no changes needed)

### `apiClient.getPublicWishList(slug: string): Promise<WishList>`

- **Endpoint**: `GET /public/wishlists/{slug}`
- **Auth**: None
- **Returns**: WishList object
- **Errors**: Throws on 404 (not found) or network error

### `apiClient.getPublicGiftItems(slug: string, page?: number, limit?: number): Promise<GetGiftItemsResponse>`

- **Endpoint**: `GET /public/wishlists/{slug}/gift-items`
- **Auth**: None
- **Returns**: Paginated gift items with `{ items, page, limit, total, pages }`
- **Errors**: Throws on 404 (wishlist not found) or network error

### `apiClient.createReservation(wishlistId: string, itemId: string, data?: CreateReservationRequest): Promise<Reservation>`

- **Endpoint**: `POST /reservations/wishlist/{wishlistId}/item/{itemId}`
- **Auth**: Optional (guest provides `guest_name` + `guest_email` in body)
- **Request Body**: `{ guest_name: string, guest_email: string }`
- **Returns**: Reservation object with `reservation_token`
- **Errors**: Throws on 400 (validation), 401 (guest data missing), 500

## New Methods (to be added)

### `apiClient.getGuestReservations(token: string): Promise<ReservationDetailsResponse[]>`

- **Endpoint**: `GET /guest/reservations?token={token}`
- **Auth**: None (token in query param)
- **Returns**: Array of `ReservationDetailsResponse` objects:
  ```typescript
  {
    id: string;
    status: string;
    reserved_at: string;
    expires_at?: string;
    gift_item: {
      id: string;
      name: string;
      image_url?: string;
      price?: string;
    };
    wishlist: {
      id: string;
      title: string;
      owner_first_name?: string;
      owner_last_name?: string;
    };
  }
  ```
- **Errors**: Throws on 400 (invalid token), 500

### `apiClient.cancelReservation(wishlistId: string, itemId: string, data?: CancelReservationRequest): Promise<Reservation>`

- **Endpoint**: `DELETE /reservations/wishlist/{wishlistId}/item/{itemId}`
- **Auth**: Optional (guest provides `reservation_token` in body)
- **Request Body**: `{ reservation_token?: string }` (required for guests)
- **Returns**: Updated reservation object
- **Errors**: Throws on 400 (validation), 401 (unauthorized), 500

## New Type Exports (to be added to `types.ts`)

```typescript
// From generated schema
export type ReservationDetailsResponse = components['schemas'][
  'wish-list_internal_domain_reservation_delivery_http_dto.ReservationDetailsResponse'
];

export type GiftItemSummary = components['schemas'][
  'wish-list_internal_domain_reservation_delivery_http_dto.GiftItemSummary'
];

export type WishListSummary = components['schemas'][
  'wish-list_internal_domain_reservation_delivery_http_dto.WishListSummary'
];

export type CancelReservationRequest = components['schemas'][
  'wish-list_internal_domain_reservation_delivery_http_dto.CancelReservationRequest'
];

export type ReservationStatusResponse = components['schemas'][
  'wish-list_internal_domain_reservation_delivery_http_dto.ReservationStatusResponse'
];
```

## Query Keys Convention

| Query Key | Method | Invalidation Trigger |
|-----------|--------|---------------------|
| `['public-wishlist', slug]` | `getPublicWishList` | - |
| `['public-gift-items', slug]` | `getPublicGiftItems` | After reservation create/cancel |
| `['guest-reservations', token]` | `getGuestReservations` | After reservation cancel |

## Error Handling Strategy

All API client methods follow the same pattern:
1. Call the typed `openapi-fetch` client
2. Check for `error` in response
3. Throw descriptive `Error` with message from server or fallback text
4. UI components catch errors via TanStack Query's `onError` callback or `isError` state
5. Display error via Sonner toast or inline error message
