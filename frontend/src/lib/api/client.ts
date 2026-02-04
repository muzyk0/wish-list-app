// frontend/src/lib/api/client.ts
import createClient from "openapi-fetch";
import type { paths } from "./schema";
import type {
  CreateReservationRequest,
  GetGiftItemsResponse,
  MobileHandoffResponse,
  Reservation,
  WishList,
} from "./types";

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api";

class AuthManager {
  private accessToken: string | null = null;
  private refreshPromise: Promise<string | null> | null = null;

  setAccessToken(token: string): void {
    this.accessToken = token;
  }

  getAccessToken(): string | null {
    return this.accessToken;
  }

  clearAccessToken(): void {
    this.accessToken = null;
  }

  isAuthenticated(): boolean {
    return this.accessToken !== null;
  }

  /**
   * Refresh access token using httpOnly cookie
   * Uses singleton pattern to prevent concurrent refresh requests
   */
  async refreshAccessToken(): Promise<string | null> {
    if (this.refreshPromise) {
      return this.refreshPromise;
    }

    this.refreshPromise = this.doRefresh();

    try {
      return await this.refreshPromise;
    } finally {
      this.refreshPromise = null;
    }
  }

  private async doRefresh(): Promise<string | null> {
    try {
      const client = createClient<paths>({ baseUrl: API_BASE_URL });
      const { data, error } = await client.POST("/auth/refresh", {
        credentials: "include",
      });

      if (error || !data) {
        this.clearAccessToken();
        return null;
      }

      const response = data as { accessToken?: string };
      if (response.accessToken) {
        this.setAccessToken(response.accessToken);
        return response.accessToken;
      }

      return null;
    } catch (error) {
      console.error("Token refresh failed:", error);
      this.clearAccessToken();
      return null;
    }
  }
}

export const authManager = new AuthManager();

/**
 * API Client for Frontend
 * Only includes methods needed for public website (guest access)
 * All authenticated operations (CRUD) are done in Mobile app
 */
class ApiClient {
  private client: ReturnType<typeof createClient<paths>>;

  constructor() {
    this.client = createClient<paths>({ baseUrl: API_BASE_URL });
  }

  private getHeaders(): Record<string, string> {
    const token = authManager.getAccessToken();
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
    };
    if (token) {
      headers.Authorization = `Bearer ${token}`;
    }
    return headers;
  }

  /**
   * Get public wishlist by slug
   * Used for viewing shared wishlists without authentication
   */
  async getPublicWishList(slug: string): Promise<WishList> {
    const { data, error } = await this.client.GET("/public/wishlists/{slug}", {
      params: { path: { slug } },
    });

    if (error || !data) {
      throw new Error(
        (error as { error?: string })?.error ||
          "Failed to fetch public wish list",
      );
    }

    return data;
  }

  /**
   * Get gift items for a public wishlist by slug
   * Used for viewing gift items in shared wishlists with pagination
   */
  async getPublicGiftItems(
    slug: string,
    page = 1,
    limit = 10,
  ): Promise<GetGiftItemsResponse> {
    const { data, error } = await this.client.GET(
      "/public/wishlists/{slug}/gift-items",
      {
        params: {
          path: { slug },
          query: { page, limit },
        },
      },
    );

    if (error || !data) {
      throw new Error(
        (error as { error?: string })?.error || "Failed to fetch gift items",
      );
    }

    return data;
  }

  /**
   * Create guest reservation for a gift item
   * Allows unauthenticated users to reserve items by providing name/email
   */
  async createReservation(
    wishlistId: string,
    itemId: string,
    reservationData?: CreateReservationRequest,
  ): Promise<Reservation> {
    const { data, error } = await this.client.POST(
      "/wishlists/{wishlistId}/gift-items/{itemId}/reservation",
      {
        params: { path: { wishlistId, itemId } },
        body: reservationData,
        headers: this.getHeaders(),
        credentials: "include",
      },
    );

    if (error || !data) {
      throw new Error(
        (error as { error?: string })?.error || "Failed to create reservation",
      );
    }

    return data;
  }

  /**
   * Generate mobile handoff code for Frontend → Mobile auth transfer
   * Used to redirect authenticated users to Mobile app with session
   */
  async mobileHandoff(): Promise<MobileHandoffResponse> {
    const { data, error } = await this.client.POST("/auth/mobile-handoff", {
      headers: this.getHeaders(),
      credentials: "include", // Supported as valid fetch option
    });

    if (error || !data) {
      throw new Error(
        (error as { error?: string })?.error ||
          "Failed to generate handoff code",
      );
    }

    return data;
  }
}

export const apiClient = new ApiClient();
