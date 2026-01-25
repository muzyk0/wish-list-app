// User-related types
export interface User {
  id: string;
  email: string;
  first_name?: string;
  last_name?: string;
  avatar_url?: string;
  is_verified: boolean;
  created_at: string;
  updated_at: string;
  last_login_at?: string;
  deactivated_at?: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  user: User;
  token: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  first_name?: string;
  last_name?: string;
}

// Wishlist-related types
export interface WishList {
  id: string;
  owner_id: string;
  title: string;
  description?: string;
  occasion?: string;
  occasion_date?: string; // ISO date string
  template_id: string;
  is_public: boolean;
  public_slug?: string;
  view_count: number;
  created_at: string;
  updated_at: string;
}

export interface CreateWishListRequest {
  title: string;
  description?: string;
  occasion?: string;
  occasion_date?: string;
  template_id?: string;
  is_public?: boolean;
}

export interface UpdateWishListRequest {
  title?: string;
  description?: string;
  occasion?: string;
  occasion_date?: string;
  template_id?: string;
  is_public?: boolean;
}

// Gift item-related types
export interface GiftItem {
  id: string;
  wishlist_id: string;
  name: string;
  description?: string;
  link?: string;
  image_url?: string;
  price?: number;
  priority: number; // 0-10 scale
  reserved_by_user_id?: string;
  reserved_at?: string;
  purchased_by_user_id?: string;
  purchased_at?: string;
  purchased_price?: number;
  notes?: string;
  position: number;
  created_at: string;
  updated_at: string;
}

export interface CreateGiftItemRequest {
  name: string;
  description?: string;
  link?: string;
  image_url?: string;
  price?: number;
  priority?: number;
  notes?: string;
  position?: number;
}

export interface UpdateGiftItemRequest {
  name?: string;
  description?: string;
  link?: string;
  image_url?: string;
  price?: number;
  priority?: number;
  notes?: string;
  position?: number;
}

// Reservation-related types
export interface Reservation {
  id: string;
  giftItemId: string;
  reserved_by_user_id?: string;
  guest_name?: string;
  guestEmail?: string;
  reservationToken: string;
  status: 'active' | 'cancelled' | 'fulfilled' | 'expired';
  reservedAt: string;
  expiresAt?: string;
  cancelledAt?: string;
  cancelledReason?: string;
  notificationSent: boolean;
}

export interface CreateReservationRequest {
  giftItemId: string;
  guestName?: string;
  guestEmail?: string;
}

export interface CancelReservationRequest {
  reservationId: string;
  reason?: string;
}

// Template-related types
export interface Template {
  id: string;
  name: string;
  description?: string;
  preview_image_url?: string;
  config: Record<string, unknown>; // JSON configuration
  is_default: boolean;
  created_at: string;
  updated_at: string;
}

// API response types
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginatedResponse<T> {
  items: T[];
  totalCount: number;
  page: number;
  pageSize: number;
  totalPages: number;
}

// Authentication context type
export interface AuthContextType {
  user: User | null;
  token: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  register: (userData: RegisterRequest) => Promise<void>;
  isAuthenticated: boolean;
  isLoading: boolean;
}
