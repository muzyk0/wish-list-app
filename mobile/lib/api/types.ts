/**
 * Type exports from generated OpenAPI schema
 * This file provides convenient type aliases for the generated schema
 */

import type { components, paths } from './schema';

// User types
export type User =
  components['schemas']['wish-list_internal_domain_user_delivery_http_dto.UserResponse'];
export type UserRegistration =
  components['schemas']['wish-list_internal_domain_user_delivery_http_dto.RegisterRequest'];
export type UserLogin =
  components['schemas']['wish-list_internal_domain_user_delivery_http_dto.LoginRequest'];
export type UserUpdate =
  components['schemas']['wish-list_internal_domain_user_delivery_http_dto.UpdateProfileRequest'];
export type LoginResponse =
  components['schemas']['wish-list_internal_domain_user_delivery_http_dto.AuthResponse'];

// Wish list types
export type WishList =
  components['schemas']['wish-list_internal_domain_wishlist_delivery_http_dto.WishListResponse'];
export type CreateWishListRequest =
  components['schemas']['wish-list_internal_domain_wishlist_delivery_http_dto.CreateWishListRequest'];
export type UpdateWishListRequest =
  components['schemas']['wish-list_internal_domain_wishlist_delivery_http_dto.UpdateWishListRequest'];

// Gift item types
export type GiftItem =
  components['schemas']['wish-list_internal_domain_wishlist_delivery_http_dto.GiftItemResponse'];
export type CreateGiftItemRequest =
  components['schemas']['wish-list_internal_domain_wishlist_item_delivery_http_dto.CreateItemRequest'];
export type UpdateGiftItemRequest =
  components['schemas']['wish-list_internal_domain_item_delivery_http_dto.UpdateItemRequest'];

// Wishlist item types (from /wishlists/{id}/items endpoint - camelCase)
export type WishlistItem =
  components['schemas']['wish-list_internal_domain_wishlist_item_delivery_http_dto.ItemResponse'];
export type PaginatedWishlistItems =
  components['schemas']['wish-list_internal_domain_wishlist_item_delivery_http_dto.PaginatedItemsResponse'];

// Reservation types
export type Reservation =
  components['schemas']['wish-list_internal_domain_reservation_delivery_http_dto.CreateReservationResponse'];
export type CreateReservationRequest =
  components['schemas']['wish-list_internal_domain_reservation_delivery_http_dto.CreateReservationRequest'];
export type CancelReservationRequest =
  components['schemas']['wish-list_internal_domain_reservation_delivery_http_dto.CancelReservationRequest'];
export type ReservationDetails =
  components['schemas']['wish-list_internal_domain_reservation_delivery_http_dto.ReservationDetailsResponse'];
export type ReservationStatus =
  components['schemas']['wish-list_internal_domain_reservation_delivery_http_dto.ReservationStatusResponse'];

// Response types
export type GetGiftItemsResponse =
  components['schemas']['wish-list_internal_domain_wishlist_delivery_http_dto.GetGiftItemsResponse'];
export type UserReservationsResponse =
  components['schemas']['wish-list_internal_domain_reservation_delivery_http_dto.UserReservationsResponse'];

// Path operation types for type-safe API calls
export type RegisterOperation = paths['/auth/register']['post'];
export type LoginOperation = paths['/auth/login']['post'];
export type GetProfileOperation = paths['/protected/profile']['get'];
export type UpdateProfileOperation = paths['/protected/profile']['put'];

// Authentication context type (UI-specific, not from API)
export interface AuthContextType {
  user: User | null;
  token: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  register: (userData: UserRegistration) => Promise<void>;
  isAuthenticated: boolean;
  isLoading: boolean;
}
