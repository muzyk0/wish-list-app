/**
 * Type exports from generated OpenAPI schema
 * Frontend uses only public/guest operations - all authenticated CRUD is in Mobile
 */

import type { components } from './generated-schema';

// Wish list type from API (without gift_items - fetched separately)
export type WishList =
  components['schemas']['wish-list_internal_domain_wishlist_delivery_http_dto.WishListResponse'];

// Gift item type
export type GiftItem =
  components['schemas']['wish-list_internal_domain_wishlist_delivery_http_dto.GiftItemResponse'];

// Gift items response with pagination
export type GetGiftItemsResponse =
  components['schemas']['wish-list_internal_domain_wishlist_delivery_http_dto.GetGiftItemsResponse'];

// Reservation types
export type Reservation =
  components['schemas']['wish-list_internal_domain_reservation_delivery_http_dto.CreateReservationResponse'];
export type CreateReservationRequest =
  components['schemas']['wish-list_internal_domain_reservation_delivery_http_dto.CreateReservationRequest'];

// Auth types (for mobile handoff)
export type MobileHandoffResponse =
  components['schemas']['wish-list_internal_domain_auth_delivery_http_dto.HandoffResponse'];

// Guest reservation detail types
export type ReservationDetailsResponse =
  components['schemas']['wish-list_internal_domain_reservation_delivery_http_dto.ReservationDetailsResponse'];

export type GiftItemSummary =
  components['schemas']['wish-list_internal_domain_reservation_delivery_http_dto.GiftItemSummary'];

export type WishListSummary =
  components['schemas']['wish-list_internal_domain_reservation_delivery_http_dto.WishListSummary'];

export type CancelReservationRequest =
  components['schemas']['wish-list_internal_domain_reservation_delivery_http_dto.CancelReservationRequest'];

export type ReservationStatusResponse =
  components['schemas']['wish-list_internal_domain_reservation_delivery_http_dto.ReservationStatusResponse'];
