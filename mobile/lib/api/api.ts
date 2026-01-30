// mobile/lib/api.ts
import * as SecureStore from 'expo-secure-store';
import type {
  CreateGiftItemRequest,
  CreateReservationRequest,
  CreateWishListRequest,
  GiftItem,
  LoginRequest,
  LoginResponse,
  PaginatedResponse,
  RegisterRequest,
  Reservation,
  Template,
  UpdateGiftItemRequest,
  UpdateWishListRequest,
  User,
  WishList,
} from '../types';

const API_BASE_URL =
  process.env.EXPO_PUBLIC_API_URL || 'http://10.0.2.2:8080/api';

class ApiClient {
  private token: string | null = null;
  private tokenReady: Promise<void>;
  private resolveTokenReady!: () => void;

  constructor() {
    // Initialize token ready promise
    this.tokenReady = new Promise((resolve) => {
      this.resolveTokenReady = resolve;
    });
    // Load token from secure storage
    this.loadToken();
  }

  private async loadToken() {
    try {
      this.token = await SecureStore.getItemAsync('auth_token');
    } catch (error) {
      console.error('Error loading token:', error);
    } finally {
      this.resolveTokenReady();
    }
  }

  private async request<T>(
    endpoint: string,
    options: RequestInit = {},
  ): Promise<T> {
    // Wait for token to be loaded before making requests
    await this.tokenReady;

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
    try {
      await SecureStore.deleteItemAsync('auth_token');
    } catch (error) {
      console.error('Error removing token:', error);
    }
  }

  private async setToken(token: string): Promise<void> {
    this.token = token;
    try {
      await SecureStore.setItemAsync('auth_token', token);
    } catch (error) {
      console.error('Error saving token:', error);
    }
  }

  // User methods
  async getProfile(): Promise<User> {
    return this.request<User>('/protected/profile');
  }

  async updateProfile(userData: Partial<User>): Promise<User> {
    return this.request<User>('/protected/profile', {
      method: 'PUT',
      body: JSON.stringify(userData),
    });
  }

  async deleteAccount(): Promise<void> {
    await this.request('/protected/account', {
      method: 'DELETE',
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
    return this.request<WishList>(`/public/lists/${slug}`);
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
  async getGiftItems(wishlistId: string): Promise<PaginatedResponse<GiftItem>> {
    return this.request<PaginatedResponse<GiftItem>>(
      `/gift-items/wishlist/${wishlistId}`,
    );
  }

  async getGiftItemById(id: string): Promise<GiftItem> {
    return this.request<GiftItem>(`/gift-items/${id}`);
  }

  async createGiftItem(
    wishlistId: string,
    data: CreateGiftItemRequest,
  ): Promise<GiftItem> {
    return this.request<GiftItem>(`/gift-items/wishlist/${wishlistId}`, {
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

  async markGiftItemAsPurchased(
    giftItemId: string,
    purchasedPrice: number,
  ): Promise<GiftItem> {
    return this.request<GiftItem>(`/gift-items/${giftItemId}/mark-purchased`, {
      method: 'POST',
      body: JSON.stringify({ purchased_price: purchasedPrice }),
    });
  }

  // Template methods
  async getTemplates(): Promise<Template[]> {
    return this.request<Template[]>('/templates');
  }

  async getDefaultTemplate(): Promise<Template> {
    return this.request<Template>('/templates/default');
  }

  async updateWishListTemplate(
    wishListId: string,
    templateId: string,
  ): Promise<WishList> {
    return this.request<WishList>(`/wishlists/${wishListId}/template`, {
      method: 'PUT',
      body: JSON.stringify({ template_id: templateId }),
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

// Helper functions for API calls
export const registerUser = async (userData: {
  email: string;
  password: string;
  firstName?: string;
  lastName?: string;
}) => {
  // The API client will handle the transformation to snake_case
  const registerRequest: RegisterRequest = {
    email: userData.email,
    password: userData.password,
    first_name: userData.firstName,
    last_name: userData.lastName,
  };
  return await apiClient.register(registerRequest);
};

export const loginUser = async (credentials: {
  email: string;
  password: string;
}) => {
  const loginRequest: LoginRequest = {
    email: credentials.email,
    password: credentials.password,
  };
  return await apiClient.login(loginRequest);
};
