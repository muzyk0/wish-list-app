// frontend/src/lib/api.ts
import type {
  CreateGiftItemRequest,
  CreateReservationRequest,
  CreateWishListRequest,
  GiftItem,
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  Reservation,
  UpdateGiftItemRequest,
  UpdateWishListRequest,
  User,
  WishList,
} from './types';

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

class ApiClient {
  private token: string | null = null;

  constructor() {
    // Check for token in localStorage on initialization
    if (typeof window !== 'undefined') {
      this.token = localStorage.getItem('token');
    }
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
  ): Promise<T> {
    const url = `${API_BASE_URL}${endpoint}`;

    const headers = {
      'Content-Type': 'application/json',
      ...(this.token && { Authorization: `Bearer ${this.token}` }),
      ...options.headers,
    };

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const errorData = await response.text();
      throw new Error(errorData || `HTTP error! status: ${response.status}`);
    }

    // For successful responses that don't return JSON (like DELETE requests)
    if (response.status === 204) {
      return undefined as T;
    }

    return response.json();
  }

  // Authentication methods
  async login(credentials: LoginRequest): Promise<LoginResponse> {
    const response = await this.request<LoginResponse>('/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    });

    this.setToken(response.token);
    return response;
  }

  async register(userData: RegisterRequest): Promise<LoginResponse> {
    const response = await this.request<LoginResponse>('/auth/register', {
      method: 'POST',
      body: JSON.stringify(userData),
    });

    this.setToken(response.token);
    return response;
  }

  async logout(): Promise<void> {
    this.token = null;
    if (typeof window !== 'undefined') {
      localStorage.removeItem('token');
    }
  }

  private setToken(token: string): void {
    this.token = token;
    if (typeof window !== 'undefined') {
      localStorage.setItem('token', token);
    }
  }

  // User methods
  async getProfile(): Promise<User> {
    return this.request<User>('/users/profile');
  }

  async updateProfile(userData: Partial<User>): Promise<User> {
    return this.request<User>('/users/profile', {
      method: 'PUT',
      body: JSON.stringify(userData),
    });
  }

  // Wishlist methods
  async getWishLists(): Promise<WishList[]> {
    return this.request<WishList[]>('/wishlists');
  }

  async getWishListById(id: string): Promise<WishList> {
    return this.request<WishList>(`/wishlists/${id}`);
  }

  async getPublicWishList(slug: string): Promise<WishList> {
    return this.request<WishList>(`/wishlists/public/${slug}`);
  }

  async createWishList(data: CreateWishListRequest): Promise<WishList> {
    return this.request<WishList>('/wishlists', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateWishList(
    id: string,
    data: UpdateWishListRequest,
  ): Promise<WishList> {
    return this.request<WishList>(`/wishlists/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteWishList(id: string): Promise<void> {
    await this.request(`/wishlists/${id}`, {
      method: 'DELETE',
    });
  }

  // Gift item methods
  async getGiftItems(wishlistId: string): Promise<GiftItem[]> {
    return this.request<GiftItem[]>(`/wishlists/${wishlistId}/gift-items`);
  }

  async getGiftItemById(id: string): Promise<GiftItem> {
    return this.request<GiftItem>(`/gift-items/${id}`);
  }

  async createGiftItem(
    wishlistId: string,
    data: CreateGiftItemRequest,
  ): Promise<GiftItem> {
    return this.request<GiftItem>(`/wishlists/${wishlistId}/gift-items`, {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async updateGiftItem(
    id: string,
    data: UpdateGiftItemRequest,
  ): Promise<GiftItem> {
    return this.request<GiftItem>(`/gift-items/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
    });
  }

  async deleteGiftItem(id: string): Promise<void> {
    await this.request(`/gift-items/${id}`, {
      method: 'DELETE',
    });
  }

  // Reservation methods
  async createReservation(
    data: CreateReservationRequest,
  ): Promise<Reservation> {
    return this.request<Reservation>('/reservations', {
      method: 'POST',
      body: JSON.stringify(data),
    });
  }

  async getReservationsByUser(): Promise<Reservation[]> {
    return this.request<Reservation[]>('/reservations/my');
  }

  async cancelReservation(id: string): Promise<void> {
    await this.request(`/reservations/${id}/cancel`, {
      method: 'POST',
    });
  }
}

export const apiClient = new ApiClient();
