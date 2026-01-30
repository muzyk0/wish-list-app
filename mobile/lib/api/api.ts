// mobile/lib/api.ts
import * as SecureStore from 'expo-secure-store';
import createClient from 'openapi-fetch';
import type { paths } from './schema';
import type {
  User,
  UserRegistration,
  UserLogin,
  LoginResponse,
  WishList,
  CreateWishListRequest,
  UpdateWishListRequest,
  GiftItem,
  CreateGiftItemRequest,
  UpdateGiftItemRequest,
  Reservation,
  CreateReservationRequest,
} from './types';

const API_BASE_URL = process.env.EXPO_PUBLIC_API_URL || 'http://10.0.2.2:8080';

class ApiClient {
  private token: string | null = null;
  private tokenReady: Promise<void>;
  private resolveTokenReady!: () => void;
  private client: ReturnType<typeof createClient<paths>>;

  constructor() {
    // Initialize token ready promise
    this.tokenReady = new Promise((resolve) => {
      this.resolveTokenReady = resolve;
    });

    // Create openapi-fetch client
    this.client = createClient<paths>({ baseUrl: API_BASE_URL });

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

  private getHeaders(): Record<string, string> {
    const headers: Record<string, string> = {};
    if (this.token) {
      headers.Authorization = `Bearer ${this.token}`;
    }
    return headers;
  }

  // Authentication methods
  async login(credentials: UserLogin): Promise<LoginResponse> {
    await this.tokenReady;

    const { data, error } = await this.client.POST('/v1/users/login', {
      body: credentials,
      headers: this.getHeaders(),
    });

    if (error || !data) {
      throw new Error(
        // biome-ignore lint/suspicious/noExplicitAny: OpenAPI error type
        (error as any)?.error || 'Login failed',
      );
    }

    const response = data as LoginResponse;
    await this.setToken(response.token);
    return response;
  }

  async register(userData: UserRegistration): Promise<LoginResponse> {
    await this.tokenReady;

    const { data, error } = await this.client.POST('/v1/users/register', {
      body: userData,
      headers: this.getHeaders(),
    });

    if (error || !data) {
      throw new Error(
        // biome-ignore lint/suspicious/noExplicitAny: OpenAPI error type
        (error as any)?.error || 'Registration failed',
      );
    }

    const response = data as LoginResponse;
    await this.setToken(response.token);
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
    await this.tokenReady;

    const { data, error } = await this.client.GET('/v1/users/me', {
      headers: this.getHeaders(),
    });

    if (error || !data) {
      throw new Error((error as any)?.error || 'Failed to fetch profile');
    }

    return data as User;
  }

  async updateProfile(userData: {
    first_name?: string;
    last_name?: string;
    avatar_url?: string;
  }): Promise<User> {
    await this.tokenReady;

    const { data, error } = await this.client.PUT('/v1/users/me', {
      body: userData,
      headers: this.getHeaders(),
    });

    if (error || !data) {
      throw new Error((error as any)?.error || 'Failed to update profile');
    }

    return data as User;
  }

  async deleteAccount(): Promise<void> {
    await this.tokenReady;

    // Note: DELETE /v1/users/me endpoint may not be implemented yet
    // Commenting out for now
    throw new Error('Delete account not implemented');

    /* const { error } = await this.client.DELETE('/v1/users/me', {
      headers: this.getHeaders(),
    });

    if (error) {
      throw new Error(
        (error as any)?.error || 'Failed to delete account',
      );
    } */
  }

  // Wishlist methods
  async getWishLists(): Promise<WishList[]> {
    await this.tokenReady;

    const { data, error } = await this.client.GET('/v1/wishlists', {
      headers: this.getHeaders(),
    });

    if (error || !data) {
      throw new Error((error as any)?.error || 'Failed to fetch wish lists');
    }

    return (data as any).data || [];
  }

  async getWishListById(id: string): Promise<WishList> {
    await this.tokenReady;

    const { data, error } = await this.client.GET('/v1/wishlists/{id}', {
      params: { path: { id } },
      headers: this.getHeaders(),
    });

    if (error || !data) {
      throw new Error((error as any)?.error || 'Failed to fetch wish list');
    }

    return data as WishList;
  }

  async getPublicWishList(slug: string): Promise<WishList> {
    await this.tokenReady;

    const { data, error } = await this.client.GET(
      '/v1/wishlists/public/{slug}',
      {
        params: { path: { slug } },
        headers: this.getHeaders(),
      },
    );

    if (error || !data) {
      throw new Error(
        (error as any)?.error || 'Failed to fetch public wish list',
      );
    }

    return data as WishList;
  }

  async createWishList(data: CreateWishListRequest): Promise<WishList> {
    await this.tokenReady;

    const { data: responseData, error } = await this.client.POST(
      '/v1/wishlists',
      {
        body: data,
        headers: this.getHeaders(),
      },
    );

    if (error || !responseData) {
      throw new Error((error as any)?.error || 'Failed to create wish list');
    }

    return responseData as WishList;
  }

  async updateWishList(
    id: string,
    data: UpdateWishListRequest,
  ): Promise<WishList> {
    await this.tokenReady;

    const { data: responseData, error } = await this.client.PUT(
      '/v1/wishlists/{id}',
      {
        params: { path: { id } },
        body: data,
        headers: this.getHeaders(),
      },
    );

    if (error || !responseData) {
      throw new Error((error as any)?.error || 'Failed to update wish list');
    }

    return responseData as WishList;
  }

  async deleteWishList(id: string): Promise<void> {
    await this.tokenReady;

    const { error } = await this.client.DELETE('/v1/wishlists/{id}', {
      params: { path: { id } },
      headers: this.getHeaders(),
    });

    if (error) {
      throw new Error((error as any)?.error || 'Failed to delete wish list');
    }
  }

  // Gift item methods
  async getGiftItems(wishlistId: string): Promise<GiftItem[]> {
    await this.tokenReady;

    const { data, error } = await this.client.GET(
      '/v1/wishlists/{wishlistId}/items',
      {
        params: { path: { wishlistId } },
        headers: this.getHeaders(),
      },
    );

    if (error || !data) {
      throw new Error((error as any)?.error || 'Failed to fetch gift items');
    }

    return ((data as any).data || data) as GiftItem[];
  }

  async getGiftItemById(wishlistId: string, itemId: string): Promise<GiftItem> {
    await this.tokenReady;

    const { data, error } = await this.client.GET(
      '/v1/wishlists/{wishlistId}/items/{itemId}',
      {
        params: { path: { wishlistId, itemId } },
        headers: this.getHeaders(),
      },
    );

    if (error || !data) {
      throw new Error((error as any)?.error || 'Failed to fetch gift item');
    }

    return data as GiftItem;
  }

  async createGiftItem(
    wishlistId: string,
    data: CreateGiftItemRequest,
  ): Promise<GiftItem> {
    await this.tokenReady;

    const { data: responseData, error } = await this.client.POST(
      '/v1/wishlists/{wishlistId}/items',
      {
        params: { path: { wishlistId } },
        body: data,
        headers: this.getHeaders(),
      },
    );

    if (error || !responseData) {
      throw new Error((error as any)?.error || 'Failed to create gift item');
    }

    return responseData as GiftItem;
  }

  async updateGiftItem(
    wishlistId: string,
    itemId: string,
    data: UpdateGiftItemRequest,
  ): Promise<GiftItem> {
    await this.tokenReady;

    const { data: responseData, error } = await this.client.PUT(
      '/v1/wishlists/{wishlistId}/items/{itemId}',
      {
        params: { path: { wishlistId, itemId } },
        body: data,
        headers: this.getHeaders(),
      },
    );

    if (error || !responseData) {
      throw new Error((error as any)?.error || 'Failed to update gift item');
    }

    return responseData as GiftItem;
  }

  async deleteGiftItem(wishlistId: string, itemId: string): Promise<void> {
    await this.tokenReady;

    const { error } = await this.client.DELETE(
      '/v1/wishlists/{wishlistId}/items/{itemId}',
      {
        params: { path: { wishlistId, itemId } },
        headers: this.getHeaders(),
      },
    );

    if (error) {
      throw new Error((error as any)?.error || 'Failed to delete gift item');
    }
  }

  async markGiftItemAsPurchased(
    wishlistId: string,
    itemId: string,
    purchasedPrice: number,
  ): Promise<GiftItem> {
    await this.tokenReady;

    const { data, error } = await this.client.POST(
      '/v1/wishlists/{wishlistId}/items/{itemId}/mark-purchased',
      {
        params: { path: { wishlistId, itemId } },
        body: { purchased_price: purchasedPrice },
        headers: this.getHeaders(),
      },
    );

    if (error || !data) {
      throw new Error(
        (error as any)?.error || 'Failed to mark gift item as purchased',
      );
    }

    return data as GiftItem;
  }

  // Reservation methods
  async createReservation(
    wishlistId: string,
    itemId: string,
    data: CreateReservationRequest,
  ): Promise<Reservation> {
    await this.tokenReady;

    const { data: responseData, error } = await this.client.POST(
      '/v1/wishlists/{wishlistId}/items/{itemId}/reserve',
      {
        params: { path: { wishlistId, itemId } },
        body: data,
        headers: this.getHeaders(),
      },
    );

    if (error || !responseData) {
      throw new Error((error as any)?.error || 'Failed to create reservation');
    }

    return responseData as Reservation;
  }

  async getReservationsByUser(): Promise<Reservation[]> {
    await this.tokenReady;

    const { data, error } = await this.client.GET('/v1/users/me/reservations', {
      headers: this.getHeaders(),
    });

    if (error || !data) {
      throw new Error((error as any)?.error || 'Failed to fetch reservations');
    }

    return ((data as any).data || data) as Reservation[];
  }

  async cancelReservation(wishlistId: string, itemId: string): Promise<void> {
    await this.tokenReady;

    const { error } = await this.client.POST(
      '/v1/wishlists/{wishlistId}/items/{itemId}/cancel-reservation',
      {
        params: { path: { wishlistId, itemId } },
        headers: this.getHeaders(),
      },
    );

    if (error) {
      throw new Error((error as any)?.error || 'Failed to cancel reservation');
    }
  }
}

export const apiClient = new ApiClient();

// Helper functions for API calls
export const registerUser = async (userData: {
  email: string;
  password: string;
  firstName: string;
  lastName: string;
}) => {
  const registerRequest: UserRegistration = {
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
  const loginRequest: UserLogin = {
    email: credentials.email,
    password: credentials.password,
  };
  return await apiClient.login(loginRequest);
};
