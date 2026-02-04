/**
 * Type exports from generated OpenAPI schema
 * Frontend uses only public/guest operations - all authenticated CRUD is in Mobile
 */

import type { components } from "./schema";

// Wish list type from API (without gift_items - fetched separately)
export type WishList =
  components["schemas"]["internal_handlers.WishListResponse"];

// Gift item type
export type GiftItem =
  components["schemas"]["internal_handlers.GiftItemResponse"];

// Gift items response with pagination
export type GetGiftItemsResponse =
  components["schemas"]["internal_handlers.GetGiftItemsResponse"];

// Reservation types
export type Reservation =
  components["schemas"]["internal_handlers.CreateReservationResponse"];
export type CreateReservationRequest =
  components["schemas"]["internal_handlers.CreateReservationRequest"];

// Auth types (for mobile handoff)
export type MobileHandoffResponse =
  components["schemas"]["internal_handlers.HandoffResponse"];
