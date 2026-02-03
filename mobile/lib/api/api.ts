// mobile/lib/api.ts
import createClient from 'openapi-fetch';
import {
  clearTokens,
  getAccessToken,
  refreshAccessToken,
  setTokens,
} from './auth';
import type { paths } from './schema';
import type {
  CreateGiftItemRequest,
  CreateReservationRequest,
  CreateWishListRequest,
  GiftItem,
  LoginResponse,
  Reservation,
  UpdateGiftItemRequest,
  UpdateWishListRequest,
  User,
  UserLogin,
  UserRegistration,
  WishList,
} from './types';

const API_BASE_URL =
  process.env.EXPO_PUBLIC_API_URL || 'http://10.0.2.2:8080/api';

class ApiClient {
  private client: ReturnType<typeof createClient<paths>>;
  private refreshPromise: Promise<string | null> | null = null;

  constructor() {
    // Create openapi-fetch client
    this.client = createClient<paths>({ baseUrl: API_BASE_URL });
  }

  private async getHeaders(): Promise<Record<string, string>> {
    const token = await getAccessToken();
    const headers: Record<string, string> = {};
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }
    return headers;
  }

  /**
   * Make request with automatic token refresh on 401
   */
  private async requestWithRetry<T>(
    makeRequest: () => Promise<{ data?: T; error?: unknown }>,
  ): Promise<{ data?: T; error?: unknown }> {
    // First attempt
    let result = await makeRequest();

    // If 401 and not already refreshing, try to refresh token
    if (
      result.error &&
      (result.error as any)?.status === 401 &&
      !this.refreshPromise
    ) {
      // Refresh token using singleton pattern
      if (!this.refreshPromise) {
        this.refreshPromise = refreshAccessToken();
      }

      try {
        const newToken = await this.refreshPromise;
        if (newToken) {
          // Retry request with new token
          result = await makeRequest();
        }
      } finally {
        this.refreshPromise = null;
      }
    }

    return result;
  }

  // Authentication methods
  async login(credentials: UserLogin): Promise<LoginResponse> {
    const headers = await this.getHeaders();

    const { data, error } = await this.client.POST('/auth/login', {
      body: credentials,
      headers,
    });

    if (error || !data) {
      throw new Error(
        // biome-ignore lint/suspicious/noExplicitAny: OpenAPI error type
        (error as any)?.error || 'Login failed',
      );
    }

    // Store both access and refresh tokens from response
    // Assuming backend returns both in response
    await setTokens(data.accessToken, data.refreshToken); // TODO: Backend should return separate refresh token
    return data;
  }

  async register(userData: UserRegistration): Promise<LoginResponse> {
    const headers = await this.getHeaders();

    const { data, error } = await this.client.POST('/auth/register', {
      body: userData,
      headers,
    });

    if (error || !data) {
      throw new Error(
        // biome-ignore lint/suspicious/noExplicitAny: OpenAPI error type
        (error as any)?.error || 'Registration failed',
      );
    }

    // Store both access and refresh tokens from response
    await setTokens(data.accessToken, data.refreshToken); // TODO: Backend should return separate refresh token
    return data;
  }

  async logout(): Promise<void> {
    try {
      const headers = await this.getHeaders();
      await this.client.POST('/auth/logout', { headers });
    } catch (error) {
      console.error('Logout request failed:', error);
    } finally {
      // Always clear tokens locally
      await clearTokens();
    }
  }

  // User methods
  async getProfile(): Promise<User> {
    const { data, error } = await this.requestWithRetry(async () => {
      const headers = await this.getHeaders();
      return this.client.GET('/protected/profile', { headers });
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
    const { data, error } = await this.requestWithRetry(async () => {
      const headers = await this.getHeaders();
      return this.client.PUT('/protected/profile', {
        body: userData,
        headers,
      });
    });

    if (error || !data) {
      throw new Error((error as any)?.error || 'Failed to update profile');
    }

    return data as User;
  }

  async deleteAccount(): Promise<void> {
    const { error } = await this.requestWithRetry(async () => {
      const headers = await this.getHeaders();
      return this.client.DELETE('/protected/account', { headers });
    });

    if (error) {
      throw new Error((error as any)?.error || 'Failed to delete account');
    }

    // Clear tokens after successful account deletion
    await clearTokens();
  }

  async changeEmail(
    currentPassword: string,
    newEmail: string,
  ): Promise<{ message: string }> {
    const { data, error } = await this.requestWithRetry(async () => {
      const headers = await this.getHeaders();
      return this.client.POST('/auth/change-email', {
        body: {
          current_password: currentPassword,
          new_email: newEmail,
        },
        headers,
      });
    });

    if (error || !data) {
      throw new Error((error as any)?.error || 'Failed to change email');
    }

    return data as { message: string };
  }

  async changePassword(
    currentPassword: string,
    newPassword: string,
  ): Promise<{ message: string }> {
    const { data, error } = await this.requestWithRetry(async () => {
      const headers = await this.getHeaders();
      return this.client.POST('/auth/change-password', {
        body: {
          current_password: currentPassword,
          new_password: newPassword,
        },
        headers,
      });
    });

    if (error || !data) {
      throw new Error((error as any)?.error || 'Failed to change password');
    }

    return data as { message: string };
  }

  // Wishlist methods
  async getWishLists(): Promise<WishList[]> {
    const { data, error } = await this.requestWithRetry(async () => {
      const headers = await this.getHeaders();
      return this.client.GET('/wishlists', { headers });
    });

    if (error || !data) {
      throw new Error((error as any)?.error || 'Failed to fetch wish lists');
    }

    return data;
  }

  async getWishListById(id: string): Promise<WishList> {
    const { data, error } = await this.requestWithRetry(async () => {
      const headers = await this.getHeaders();
      return this.client.GET('/wishlists/{id}', {
        params: { path: { id } },
        headers,
      });
    });

    if (error || !data) {
      throw new Error((error as any)?.error || 'Failed to fetch wish list');
    }

    return data as WishList;
  }

  async getPublicWishList(slug: string): Promise<WishList> {
    const { data, error } = await this.requestWithRetry(async () => {
      const headers = await this.getHeaders();
      return this.client.GET('/public/wishlists/{slug}', {
        params: { path: { slug } },
        headers,
      });
    });

    if (error || !data) {
      throw new Error(
        (error as any)?.error || 'Failed to fetch public wish list',
      );
    }

    return data as WishList;
  }

  async createWishList(data: CreateWishListRequest): Promise<WishList> {
    const { data: responseData, error } = await this.requestWithRetry(
      async () => {
        const headers = await this.getHeaders();
        return this.client.POST('/wishlists', {
          body: data,
          headers,
        });
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
    const { data: responseData, error } = await this.requestWithRetry(
      async () => {
        const headers = await this.getHeaders();
        return this.client.PUT('/wishlists/{id}', {
          params: { path: { id } },
          body: data,
          headers,
        });
      },
    );

    if (error || !responseData) {
      throw new Error((error as any)?.error || 'Failed to update wish list');
    }

    return responseData as WishList;
  }

  async deleteWishList(id: string): Promise<void> {
    const { error } = await this.requestWithRetry(async () => {
      const headers = await this.getHeaders();
      return this.client.DELETE('/wishlists/{id}', {
        params: { path: { id } },
        headers,
      });
    });

    if (error) {
      throw new Error((error as any)?.error || 'Failed to delete wish list');
    }
  }

  // Gift item methods
  async getGiftItems(wishlistId: string): Promise<GiftItem[]> {
    const { data, error } = await this.requestWithRetry(async () => {
      const headers = await this.getHeaders();
      return this.client.GET('/wishlists/{wishlistId}/gift-items', {
        params: { path: { wishlistId } },
        headers,
      });
    });

    if (error || !data) {
      throw new Error((error as any)?.error || 'Failed to fetch gift items');
    }

    return ((data as any).data || data) as GiftItem[];
  }

  async getGiftItemById(wishlistId: string, itemId: string): Promise<GiftItem> {
    const { data, error } = await this.requestWithRetry(async () => {
      const headers = await this.getHeaders();
      return this.client.GET('/gift-items/{id}', {
        params: { path: { id: itemId } },
        headers,
      });
    });

    if (error || !data) {
      throw new Error((error as any)?.error || 'Failed to fetch gift item');
    }

    return data as GiftItem;
  }

  async createGiftItem(
    wishlistId: string,
    data: CreateGiftItemRequest,
  ): Promise<GiftItem> {
    const { data: responseData, error } = await this.requestWithRetry(
      async () => {
        const headers = await this.getHeaders();
        return this.client.POST('/wishlists/{wishlistId}/gift-items', {
          params: { path: { wishlistId } },
          body: data,
          headers,
        });
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
    const { data: responseData, error } = await this.requestWithRetry(
      async () => {
        const headers = await this.getHeaders();
        return this.client.PUT('/gift-items/{id}', {
          params: { path: { id: itemId } },
          body: data,
          headers,
        });
      },
    );

    if (error || !responseData) {
      throw new Error((error as any)?.error || 'Failed to update gift item');
    }

    return responseData as GiftItem;
  }

  async deleteGiftItem(wishlistId: string, itemId: string): Promise<void> {
    const { error } = await this.requestWithRetry(async () => {
      const headers = await this.getHeaders();
      return this.client.DELETE('/gift-items/{id}', {
        params: { path: { id: itemId } },
        headers,
      });
    });

    if (error) {
      throw new Error((error as any)?.error || 'Failed to delete gift item');
    }
  }

  async markGiftItemAsPurchased(
    wishlistId: string,
    itemId: string,
    purchasedPrice: number,
  ): Promise<GiftItem> {
    const { data, error } = await this.requestWithRetry(async () => {
      const headers = await this.getHeaders();
      return this.client.POST('/gift-items/{id}/purchase', {
        params: { path: { id: itemId } },
        body: { purchased_price: purchasedPrice },
        headers,
      });
    });

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
    const { data: responseData, error } = await this.requestWithRetry(
      async () => {
        const headers = await this.getHeaders();
        return this.client.POST(
          '/wishlists/{wishlistId}/gift-items/{itemId}/reservation',
          {
            params: { path: { wishlistId, itemId } },
            body: data,
            headers,
          },
        );
      },
    );

    if (error || !responseData) {
      throw new Error((error as any)?.error || 'Failed to create reservation');
    }

    return responseData as Reservation;
  }

  async getReservationsByUser(): Promise<Reservation[]> {
    const { data, error } = await this.requestWithRetry(async () => {
      const headers = await this.getHeaders();
      return this.client.GET('/reservations', { headers });
    });

    if (error || !data) {
      throw new Error((error as any)?.error || 'Failed to fetch reservations');
    }

    return ((data as any).data || data) as Reservation[];
  }

  async cancelReservation(wishlistId: string, itemId: string): Promise<void> {
    const { error } = await this.requestWithRetry(async () => {
      const headers = await this.getHeaders();
      return this.client.DELETE(
        '/wishlists/{wishlistId}/gift-items/{itemId}/reservation',
        {
          params: { path: { wishlistId, itemId } },
          headers,
        },
      );
    });

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
