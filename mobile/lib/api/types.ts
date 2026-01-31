/**
 * Type exports from generated OpenAPI schema
 * This file provides convenient type aliases for the generated schema
 */

import type { components, paths } from './schema';

// User types
export type User = Required<components['schemas']['user_response']>;
export type UserRegistration = components['schemas']['user_registration'];
export type UserLogin = components['schemas']['user_login'];
export type UserUpdate = components['schemas']['user_update'];
export type LoginResponse = Required<components['schemas']['login_response']>;

// Wish list types
export type WishList = Required<components['schemas']['wish_list_response']>;
export type WishListWithItems =
  components['schemas']['wish_list_with_items_response'];
export type PublicWishList = components['schemas']['public_wish_list_response'];
export type CreateWishListRequest = components['schemas']['wish_list_create'];
export type UpdateWishListRequest = components['schemas']['wish_list_update'];

// Gift item types
export type GiftItem = Required<components['schemas']['gift_item_response']>;
export type PublicGiftItem = components['schemas']['public_gift_item_response'];
export type CreateGiftItemRequest = components['schemas']['gift_item_create'];
export type UpdateGiftItemRequest = components['schemas']['gift_item_update'];

// Reservation types
export type Reservation = Required<
  components['schemas']['reservation_response']
>;
export type GuestReservation = components['schemas']['guest_reservation'];
export type AuthenticatedReservation =
  components['schemas']['authenticated_reservation'];
export type CreateReservationRequest =
  | GuestReservation
  | AuthenticatedReservation;

// Utility types
export type ApiError = components['schemas']['error'];
export type Pagination = components['schemas']['pagination'];

// Path operation types for type-safe API calls
export type RegisterOperation = paths['/v1/users/register']['post'];
export type LoginOperation = paths['/v1/users/login']['post'];
export type GetProfileOperation = paths['/v1/users/me']['get'];
export type UpdateProfileOperation = paths['/v1/users/me']['put'];

// Legacy compatibility types (for gradual migration)
export interface PaginatedResponse<T> {
  items: T[];
  pagination: Pagination;
}

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
