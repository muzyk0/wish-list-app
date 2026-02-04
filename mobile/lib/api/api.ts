// mobile/lib/api.ts
import createClient, { type Middleware } from "openapi-fetch";
import {
  clearTokens,
  getAccessToken,
  refreshAccessToken,
  setTokens,
} from "./auth";
import { API_BASE_URL } from "./client";
import type { paths } from "./schema";
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
} from "./types";

// Routes that don't require authentication
const UNPROTECTED_ROUTES = ["/auth/login", "/auth/register"];

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
        request.headers.set("Authorization", `Bearer ${token}`);
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
            retryRequest.headers.set("Authorization", `Bearer ${newToken}`);

            // Retry the request with the new token
            return await fetch(retryRequest);
          }

          // If refresh failed, clear tokens and return original 401 response
          await clearTokens();
          return response;
        } catch (error) {
          // If refresh throws an error, clear tokens and return original response
          console.error("Token refresh failed:", error);
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
    const { data, error } = await this.client.POST("/auth/login", {
      body: credentials,
    });

    if (error || !data) {
      throw new Error(
        // biome-ignore lint/suspicious/noExplicitAny: OpenAPI error type
        (error as any)?.error || "Login failed",
      );
    }

    // Store both access and refresh tokens from response
    await setTokens(data.accessToken, data.refreshToken);
    return data;
  }

  async register(userData: UserRegistration): Promise<LoginResponse> {
    const { data, error } = await this.client.POST("/auth/register", {
      body: userData,
    });

    if (error || !data) {
      throw new Error(
        // biome-ignore lint/suspicious/noExplicitAny: OpenAPI error type
        (error as any)?.error || "Registration failed",
      );
    }

    // Store both access and refresh tokens from response
    await setTokens(data.accessToken, data.refreshToken);
    return data;
  }

  async logout(): Promise<void> {
    try {
      await this.client.POST("/auth/logout", {});
    } catch (error) {
      console.error("Logout request failed:", error);
    } finally {
      // Always clear tokens locally
      await clearTokens();
    }
  }

  // User methods
  async getProfile(): Promise<User> {
    const { data, error } = await this.client.GET("/protected/profile", {});

    if (error || !data) {
      throw new Error((error as any)?.error || "Failed to fetch profile");
    }

    return data as User;
  }

  async updateProfile(userData: {
    first_name?: string;
    last_name?: string;
    avatar_url?: string;
  }): Promise<User> {
    const { data, error } = await this.client.PUT("/protected/profile", {
      body: userData,
    });

    if (error || !data) {
      throw new Error((error as any)?.error || "Failed to update profile");
    }

    return data as User;
  }

  async deleteAccount(): Promise<void> {
    const { error } = await this.client.DELETE("/protected/account", {});

    if (error) {
      throw new Error((error as any)?.error || "Failed to delete account");
    }

    // Clear tokens after successful account deletion
    await clearTokens();
  }

  async changeEmail(
    currentPassword: string,
    newEmail: string,
  ): Promise<{ message: string }> {
    const { data, error } = await this.client.POST("/auth/change-email", {
      body: {
        current_password: currentPassword,
        new_email: newEmail,
      },
    });

    if (error || !data) {
      throw new Error((error as any)?.error || "Failed to change email");
    }

    return data as { message: string };
  }

  async changePassword(
    currentPassword: string,
    newPassword: string,
  ): Promise<{ message: string }> {
    const { data, error } = await this.client.POST("/auth/change-password", {
      body: {
        current_password: currentPassword,
        new_password: newPassword,
      },
    });

    if (error || !data) {
      throw new Error((error as any)?.error || "Failed to change password");
    }

    return data as { message: string };
  }

  // Wishlist methods
  async getWishLists(): Promise<WishList[]> {
    const { data, error } = await this.client.GET("/wishlists", {});

    if (error) {
      throw new Error((error as any)?.error || "Failed to fetch wish lists");
    }

    return data ?? [];
  }

  async getWishListById(id: string): Promise<WishList> {
    const { data, error } = await this.client.GET("/wishlists/{id}", {
      params: { path: { id } },
    });

    if (error || !data) {
      throw new Error((error as any)?.error || "Failed to fetch wish list");
    }

    return data as WishList;
  }

  async getPublicWishList(slug: string): Promise<WishList> {
    const { data, error } = await this.client.GET("/public/wishlists/{slug}", {
      params: { path: { slug } },
    });

    if (error || !data) {
      throw new Error(
        (error as any)?.error || "Failed to fetch public wish list",
      );
    }

    return data as WishList;
  }

  async createWishList(data: CreateWishListRequest): Promise<WishList> {
    const { data: responseData, error } = await this.client.POST("/wishlists", {
      body: data,
    });

    if (error || !responseData) {
      throw new Error((error as any)?.error || "Failed to create wish list");
    }

    return responseData as WishList;
  }

  async updateWishList(
    id: string,
    data: UpdateWishListRequest,
  ): Promise<WishList> {
    const { data: responseData, error } = await this.client.PUT(
      "/wishlists/{id}",
      {
        params: { path: { id } },
        body: data,
      },
    );

    if (error || !responseData) {
      throw new Error((error as any)?.error || "Failed to update wish list");
    }

    return responseData as WishList;
  }

  async deleteWishList(id: string): Promise<void> {
    const { error } = await this.client.DELETE("/wishlists/{id}", {
      params: { path: { id } },
    });

    if (error) {
      throw new Error((error as any)?.error || "Failed to delete wish list");
    }
  }

  // Gift item methods
  async getGiftItems(wishlistId: string): Promise<GiftItem[]> {
    const { data, error } = await this.client.GET(
      "/wishlists/{wishlistId}/gift-items",
      {
        params: { path: { wishlistId } },
      },
    );

    if (error || !data) {
      throw new Error((error as any)?.error || "Failed to fetch gift items");
    }

    return ((data as any).data || data) as GiftItem[];
  }

  async getGiftItemById(wishlistId: string, itemId: string): Promise<GiftItem> {
    const { data, error } = await this.client.GET("/gift-items/{id}", {
      params: { path: { id: itemId } },
    });

    if (error || !data) {
      throw new Error((error as any)?.error || "Failed to fetch gift item");
    }

    return data as GiftItem;
  }

  async createGiftItem(
    wishlistId: string,
    data: CreateGiftItemRequest,
  ): Promise<GiftItem> {
    const { data: responseData, error } = await this.client.POST(
      "/wishlists/{wishlistId}/gift-items",
      {
        params: { path: { wishlistId } },
        body: data,
      },
    );

    if (error || !responseData) {
      throw new Error((error as any)?.error || "Failed to create gift item");
    }

    return responseData as GiftItem;
  }

  async updateGiftItem(
    wishlistId: string,
    itemId: string,
    data: UpdateGiftItemRequest,
  ): Promise<GiftItem> {
    const { data: responseData, error } = await this.client.PUT(
      "/gift-items/{id}",
      {
        params: { path: { id: itemId } },
        body: data,
      },
    );

    if (error || !responseData) {
      throw new Error((error as any)?.error || "Failed to update gift item");
    }

    return responseData as GiftItem;
  }

  async deleteGiftItem(wishlistId: string, itemId: string): Promise<void> {
    const { error } = await this.client.DELETE("/gift-items/{id}", {
      params: { path: { id: itemId } },
    });

    if (error) {
      throw new Error((error as any)?.error || "Failed to delete gift item");
    }
  }

  async markGiftItemAsPurchased(
    wishlistId: string,
    itemId: string,
    purchasedPrice: number,
  ): Promise<GiftItem> {
    const { data, error } = await this.client.POST(
      "/gift-items/{id}/purchase",
      {
        params: { path: { id: itemId } },
        body: { purchased_price: purchasedPrice },
      },
    );

    if (error || !data) {
      throw new Error(
        (error as any)?.error || "Failed to mark gift item as purchased",
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
    const { data: responseData, error } = await this.client.POST(
      "/wishlists/{wishlistId}/gift-items/{itemId}/reservation",
      {
        params: { path: { wishlistId, itemId } },
        body: data,
      },
    );

    if (error || !responseData) {
      throw new Error((error as any)?.error || "Failed to create reservation");
    }

    return responseData as Reservation;
  }

  async getReservationsByUser(): Promise<Reservation[]> {
    const { data, error } = await this.client.GET("/reservations", {});

    if (error || !data) {
      throw new Error((error as any)?.error || "Failed to fetch reservations");
    }

    return ((data as any).data || data) as Reservation[];
  }

  async cancelReservation(wishlistId: string, itemId: string): Promise<void> {
    const { error } = await this.client.DELETE(
      "/wishlists/{wishlistId}/gift-items/{itemId}/reservation",
      {
        params: { path: { wishlistId, itemId } },
      },
    );

    if (error) {
      throw new Error((error as any)?.error || "Failed to cancel reservation");
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
