// User-related types
export interface User {
  id: string;
  email: string;
  firstName?: string;
  lastName?: string;
  avatarUrl?: string;
  isVerified: boolean;
  createdAt: string;
  updatedAt: string;
  lastLoginAt?: string;
  deactivatedAt?: string;
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
  firstName?: string;
  lastName?: string;
}

// Wishlist-related types
export interface WishList {
  id: string;
  ownerId: string;
  title: string;
  description?: string;
  occasion?: string;
  occasionDate?: string; // ISO date string
  templateId: string;
  isPublic: boolean;
  publicSlug?: string;
  viewCount: number;
  createdAt: string;
  updatedAt: string;
}

export interface CreateWishListRequest {
  title: string;
  description?: string;
  occasion?: string;
  occasionDate?: string;
  templateId?: string;
  isPublic?: boolean;
}

export interface UpdateWishListRequest {
  title?: string;
  description?: string;
  occasion?: string;
  occasionDate?: string;
  templateId?: string;
  isPublic?: boolean;
}

// Gift item-related types
export interface GiftItem {
  id: string;
  wishlistId: string;
  name: string;
  description?: string;
  link?: string;
  imageUrl?: string;
  price?: number;
  priority: number; // 0-10 scale
  reservedByUserId?: string;
  reservedAt?: string;
  purchasedByUserId?: string;
  purchasedAt?: string;
  purchasedPrice?: number;
  notes?: string;
  position: number;
  createdAt: string;
  updatedAt: string;
}

export interface CreateGiftItemRequest {
  name: string;
  description?: string;
  link?: string;
  imageUrl?: string;
  price?: number;
  priority?: number;
  notes?: string;
  position?: number;
}

export interface UpdateGiftItemRequest {
  name?: string;
  description?: string;
  link?: string;
  imageUrl?: string;
  price?: number;
  priority?: number;
  notes?: string;
  position?: number;
}

// Reservation-related types
export interface Reservation {
  id: string;
  giftItemId: string;
  reservedByUserId?: string;
  guestName?: string;
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
  previewImageUrl?: string;
  config: Record<string, unknown>; // JSON configuration
  isDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

// API response types
export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
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
