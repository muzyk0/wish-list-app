// mobile/lib/api.ts
import createClient, { type Middleware } from 'openapi-fetch';
import {
  clearTokens,
  getAccessToken,
  refreshAccessToken,
  setTokens,
} from './auth';
import { API_BASE_URL } from './client';
import type { paths } from './generated-schema';
import type {
  CreateGiftItemRequest,
  CreateReservationRequest,
  CreateWishListRequest,
  GiftItem,
  LoginResponse,
  MarkManualReservationRequest,
  PaginatedGiftItems,
  Reservation,
  UpdateGiftItemRequest,
  UpdateWishListRequest,
  User,
  UserLogin,
  UserRegistration,
  WishList,
  WishlistItem,
} from './types';

// Routes that don't require authentication
const UNPROTECTED_ROUTES = ['/auth/login', '/auth/register'];

class ApiClient {
  private client: ReturnType<typeof createClient<paths>>;
  private refreshPromise: Promise<string | null> | null = null;
  private isRefreshing = false;

  constructor() {
    // Create NEW openapi-fetch client WITH middleware
    // Note: We don't use baseClient from client.ts because:
    // 1. baseClient is used by auth.ts WITHOUT middleware (prevents infinite recursion)
    // 2. This client needs middleware for automatic auth & token refresh
    // 3. Middleware should only apply to protected endpoints, not auth operations
    this.client = createClient<paths>({ baseUrl: API_BASE_URL });

    // Register authentication middleware (adds Authorization header)
    this.client.use(this.authMiddleware);

    // Register token refresh middleware (handles 401 errors)
    this.client.use(this.refreshMiddleware);
  }

  /**
   * Authentication middleware - adds Authorization header to protected routes
   */
  private authMiddleware: Middleware = {
    async onRequest({ request }) {
      // Skip auth for unprotected routes
      const url = new URL(request.url);
      const isUnprotected = UNPROTECTED_ROUTES.some((route) =>
        url.pathname.includes(route),
      );

      if (isUnprotected) {
        return request;
      }

      // Add Authorization header for protected routes
      const token = await getAccessToken();
      if (token) {
        request.headers.set('Authorization', `Bearer ${token}`);
      }

      return request;
    },
  };

  /**
   * Token refresh middleware - handles 401 errors and automatic token refresh
   */
  private refreshMiddleware: Middleware = {
    onResponse: async ({ request, response }) => {
      // Check for 401 Unauthorized
      if (response.status === 401 && !this.isRefreshing) {
        this.isRefreshing = true;

        try {
          // Attempt to refresh the token (singleton pattern prevents multiple concurrent refreshes)
          if (!this.refreshPromise) {
            this.refreshPromise = refreshAccessToken();
          }

          const newToken = await this.refreshPromise;

          if (newToken) {
            // Clone the original request with the new token
            const retryRequest = request.clone();
            retryRequest.headers.set('Authorization', `Bearer ${newToken}`);

            // Retry the request with the new token
            return await fetch(retryRequest);
          }

          // If refresh failed, clear tokens and return original 401 response
          await clearTokens();
          return response;
        } catch (error) {
          // If refresh throws an error, clear tokens and return original response
          console.error('Token refresh failed:', error);
          await clearTokens();
          return response;
        } finally {
          this.isRefreshing = false;
          this.refreshPromise = null;
        }
      }

      return response;
    },
  };

  // Authentication methods
  async login(credentials: UserLogin): Promise<LoginResponse> {
    const { data, error } = await this.client.POST('/auth/login', {
      body: credentials,
    });

    if (error) {
      throw error;
    }

    if (!data) {
      throw new Error('No data received from login');
    }

    // Store both access and refresh tokens from response
    await setTokens(data.accessToken, data.refreshToken);
    return data;
  }

  async register(userData: UserRegistration): Promise<LoginResponse> {
    const { data, error } = await this.client.POST('/auth/register', {
      body: userData,
    });

    if (error) {
      throw error;
    }

    if (!data) {
      throw new Error('No data received from registration');
    }

    // Store both access and refresh tokens from response
    await setTokens(data.accessToken, data.refreshToken);
    return data;
  }

  async logout(): Promise<void> {
    try {
      await this.client.POST('/auth/logout', {});
    } catch (error) {
      console.error('Logout request failed:', error);
    } finally {
      // Always clear tokens locally
      await clearTokens();
    }
  }

  // User methods
  async getProfile(): Promise<User> {
    const { data, error } = await this.client.GET('/protected/profile', {});

    if (error) {
      throw error;
    }

    if (!data) {
      throw new Error('No data received from profile');
    }

    return data;
  }

  async updateProfile(userData: {
    first_name?: string;
    last_name?: string;
    avatar_url?: string;
  }): Promise<User> {
    const { data, error } = await this.client.PUT('/protected/profile', {
      body: userData,
    });

    if (error) {
      throw error;
    }

    if (!data) {
      throw new Error('No data received from profile update');
    }

    return data;
  }

  async deleteAccount(): Promise<void> {
    const { error } = await this.client.DELETE('/protected/account', {});

    if (error) {
      throw error;
    }

    // Clear tokens after successful account deletion
    await clearTokens();
  }

  async changeEmail({
    currentPassword,
    newEmail,
  }: {
    currentPassword: string;
    newEmail: string;
  }): Promise<{ message: string }> {
    const { data, error } = await this.client.POST('/auth/change-email', {
      body: {
        current_password: currentPassword,
        new_email: newEmail,
      },
    });

    if (error) {
      throw error;
    }

    if (!data) {
      throw new Error('No data received from email change');
    }

    return data;
  }

  async changePassword({
    currentPassword,
    newPassword,
  }: {
    currentPassword: string;
    newPassword: string;
  }): Promise<{ message: string }> {
    const { data, error } = await this.client.POST('/auth/change-password', {
      body: {
        current_password: currentPassword,
        new_password: newPassword,
      },
    });

    if (error) {
      throw error;
    }

    if (!data) {
      throw new Error('No data received from password change');
    }

    return data;
  }

  // Wishlist methods
  async getWishLists(): Promise<WishList[]> {
    const { data, error } = await this.client.GET('/wishlists', {});

    if (error) {
      throw error;
    }

    return data ?? [];
  }

  async getWishListById(id: string): Promise<WishList> {
    const { data, error } = await this.client.GET('/wishlists/{id}', {
      params: { path: { id } },
    });

    if (error) {
      throw error;
    }

    if (!data) {
      throw new Error('No data received from wishlist');
    }

    return data;
  }

  async getPublicWishList(slug: string): Promise<WishList> {
    const { data, error } = await this.client.GET('/public/wishlists/{slug}', {
      params: { path: { slug } },
    });

    if (error) {
      throw error;
    }

    if (!data) {
      throw new Error('No data received from public wishlist');
    }

    return data;
  }

  async createWishList(data: CreateWishListRequest): Promise<WishList> {
    const { data: responseData, error } = await this.client.POST('/wishlists', {
      body: data,
    });

    if (error) {
      throw error;
    }

    if (!responseData) {
      throw new Error('No data received from wishlist creation');
    }

    return responseData;
  }

  async updateWishList(
    id: string,
    data: UpdateWishListRequest,
  ): Promise<WishList> {
    const { data: responseData, error } = await this.client.PUT(
      '/wishlists/{id}',
      {
        params: { path: { id } },
        body: data,
      },
    );

    if (error) {
      throw error;
    }

    if (!responseData) {
      throw new Error('No data received from wishlist update');
    }

    return responseData;
  }

  async deleteWishList(id: string): Promise<void> {
    const { error } = await this.client.DELETE('/wishlists/{id}', {
      params: { path: { id } },
    });

    if (error) {
      throw error;
    }
  }

  // Gift item methods
  async getGiftItems(wishlistId: string) {
    const { data, error } = await this.client.GET('/wishlists/{id}/items', {
      params: { path: { id: wishlistId } },
    });

    if (error) {
      throw error;
    }

    return data;
  }

  async getGiftItemById(
    _wishlistId: string,
    itemId: string,
  ): Promise<GiftItem> {
    const { data, error } = await this.client.GET('/items/{id}', {
      params: { path: { id: itemId } },
    });

    if (error) {
      throw error;
    }

    if (!data) {
      throw new Error('No data received from gift item');
    }

    return data;
  }

  async getUserGiftItems(options?: {
    page?: number;
    limit?: number;
    sort?: string;
    order?: string;
    unattached?: boolean;
    attached?: boolean;
    include_archived?: boolean;
    search?: string;
  }): Promise<PaginatedGiftItems> {
    const { data, error } = await this.client.GET('/items', {
      params: {
        query: options,
      },
    });

    if (error) {
      throw error;
    }

    if (!data) {
      throw new Error('No data received from user gift items');
    }

    return {
      ...data,
      items: (data.items as GiftItem[]) || [],
    };
  }

  async createGiftItem(
    wishlistId: string,
    data: CreateGiftItemRequest,
  ): Promise<WishlistItem> {
    const { data: responseData, error } = await this.client.POST(
      '/wishlists/{id}/items/new',
      {
        params: { path: { id: wishlistId } },
        body: data,
      },
    );

    if (error) {
      throw error;
    }

    if (!responseData) {
      throw new Error('No data received from gift item creation');
    }

    return responseData;
  }

  async updateGiftItem(
    _wishlistId: string,
    itemId: string,
    data: UpdateGiftItemRequest,
  ): Promise<GiftItem> {
    const { data: responseData, error } = await this.client.PUT('/items/{id}', {
      params: { path: { id: itemId } },
      body: data,
    });

    if (error) {
      throw error;
    }

    if (!responseData) {
      throw new Error('No data received from gift item update');
    }

    return responseData;
  }

  async deleteGiftItem(_wishlistId: string, itemId: string): Promise<void> {
    const { error } = await this.client.DELETE('/items/{id}', {
      params: { path: { id: itemId } },
    });

    if (error) {
      throw error;
    }
  }

  async createStandaloneGiftItem(
    data: CreateGiftItemRequest,
  ): Promise<GiftItem> {
    const { data: responseData, error } = await this.client.POST('/items', {
      body: data,
    });

    if (error) {
      throw error;
    }

    if (!responseData) {
      throw new Error('No data received from standalone gift item creation');
    }

    return responseData;
  }

  async attachGiftItemToWishlist(
    wishlistId: string,
    itemId: string,
  ): Promise<void> {
    const { error } = await this.client.POST('/wishlists/{id}/items', {
      params: { path: { id: wishlistId } },
      body: { item_id: itemId },
    });

    if (error) {
      throw error;
    }
  }

  async markGiftItemAsPurchased(
    _wishlistId: string,
    itemId: string,
    purchasedPrice: number,
  ): Promise<GiftItem> {
    const { data, error } = await this.client.POST(
      '/items/{id}/mark-purchased',
      {
        params: { path: { id: itemId } },
        body: { purchased_price: purchasedPrice },
      },
    );

    if (error) {
      throw error;
    }

    if (!data) {
      throw new Error('No data received from mark purchased');
    }

    return data;
  }

  async markItemAsManuallyReserved(
    wishlistId: string,
    itemId: string,
    data: MarkManualReservationRequest,
  ): Promise<WishlistItem> {
    const { data: responseData, error } = await this.client.PATCH(
      '/wishlists/{id}/items/{itemId}/mark-reserved',
      {
        params: { path: { id: wishlistId, itemId } },
        body: data,
      },
    );

    if (error) {
      throw error;
    }

    if (!responseData) {
      throw new Error('No data received from mark manual reservation');
    }

    return responseData;
  }

  // Reservation methods
  async createReservation(
    wishlistId: string,
    itemId: string,
    data: CreateReservationRequest,
  ): Promise<Reservation> {
    const { data: responseData, error } = await this.client.POST(
      '/public/reservations/wishlist/{wishlistId}/item/{itemId}',
      {
        params: { path: { wishlistId, itemId } },
        body: data,
      },
    );

    if (error) {
      throw error;
    }

    if (!responseData) {
      throw new Error('No data received from reservation creation');
    }

    return responseData;
  }

  async getReservationsByUser(): Promise<Reservation[]> {
    const { data, error } = await this.client.GET('/reservations/user', {});

    if (error) {
      throw error;
    }

    if (!data) {
      return [];
    }

    // Response is wrapped in { data: [...], pagination: {...} }
    if ('data' in data && Array.isArray(data.data)) {
      return data.data as unknown as Reservation[];
    }

    // Fallback if response is direct array
    return data as unknown as Reservation[];
  }

  async cancelReservation(wishlistId: string, itemId: string): Promise<void> {
    const { error } = await this.client.DELETE(
      '/public/reservations/wishlist/{wishlistId}/item/{itemId}',
      {
        params: { path: { wishlistId, itemId } },
      },
    );

    if (error) {
      throw error;
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
