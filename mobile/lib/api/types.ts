/**
 * Type exports from generated OpenAPI schema
 * This file provides convenient type aliases for the generated schema
 */

import type { components, paths } from './schema';

// User types
export type User = components['schemas']['internal_handlers.UserResponse'];
export type UserRegistration =
  components['schemas']['internal_handlers.RegisterRequest'];
export type UserLogin = components['schemas']['internal_handlers.LoginRequest'];
export type UserUpdate =
  components['schemas']['internal_handlers.UpdateProfileRequest'];
export type LoginResponse =
  components['schemas']['internal_handlers.AuthResponse'];

// Wish list types
export type WishList =
  components['schemas']['internal_handlers.WishListResponse'];
export type CreateWishListRequest =
  components['schemas']['internal_handlers.CreateWishListRequest'];
export type UpdateWishListRequest =
  components['schemas']['internal_handlers.UpdateWishListRequest'];

// Gift item types
export type GiftItem =
  components['schemas']['internal_handlers.GiftItemResponse'];
export type CreateGiftItemRequest =
  components['schemas']['internal_handlers.CreateGiftItemRequest'];
export type UpdateGiftItemRequest =
  components['schemas']['internal_handlers.UpdateGiftItemRequest'];

// Reservation types
export type Reservation =
  components['schemas']['internal_handlers.CreateReservationResponse'];
export type CreateReservationRequest =
  components['schemas']['internal_handlers.CreateReservationRequest'];
export type CancelReservationRequest =
  components['schemas']['internal_handlers.CancelReservationRequest'];
export type ReservationDetails =
  components['schemas']['internal_handlers.ReservationDetailsResponse'];
export type ReservationStatus =
  components['schemas']['internal_handlers.ReservationStatusResponse'];

// Response types
export type GetGiftItemsResponse =
  components['schemas']['internal_handlers.GetGiftItemsResponse'];
export type UserReservationsResponse =
  components['schemas']['internal_handlers.UserReservationsResponse'];

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
