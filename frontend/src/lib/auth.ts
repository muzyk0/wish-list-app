// frontend/src/lib/auth.ts
// AuthManager: Secure token management for Frontend
// - Access token stored in memory (prevents XSS)
// - Refresh token in httpOnly cookie (set by Backend)
// - Singleton pattern prevents duplicate refresh requests

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

class AuthManager {
  private accessToken: string | null = null;
  private refreshPromise: Promise<string | null> | null = null;

  /**
   * Set access token in memory
   */
  setAccessToken(token: string): void {
    this.accessToken = token;
  }

  /**
   * Get current access token
   */
  getAccessToken(): string | null {
    return this.accessToken;
  }

  /**
   * Clear access token from memory
   */
  clearAccessToken(): void {
    this.accessToken = null;
  }

  /**
   * Check if user is authenticated
   */
  isAuthenticated(): boolean {
    return this.accessToken !== null;
  }

  /**
   * Refresh access token using httpOnly cookie
   * Uses singleton pattern to prevent concurrent refresh requests
   */
  async refreshAccessToken(): Promise<string | null> {
    // Return existing promise if refresh is already in progress
    if (this.refreshPromise) {
      return this.refreshPromise;
    }

    // Create new refresh promise
    this.refreshPromise = this.doRefresh();

    // Wait for result and clear promise
    try {
      const result = await this.refreshPromise;
      return result;
    } finally {
      this.refreshPromise = null;
    }
  }

  /**
   * Internal refresh implementation
   */
  private async doRefresh(): Promise<string | null> {
    try {
      const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
        method: 'POST',
        credentials: 'include', // Send httpOnly cookie
        headers: {
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        this.clearAccessToken();
        return null;
      }

      const data = await response.json();
      if (data.accessToken) {
        this.setAccessToken(data.accessToken);
        return data.accessToken;
      }

      return null;
    } catch (error) {
      console.error('Token refresh failed:', error);
      this.clearAccessToken();
      return null;
    }
  }
}

// Export singleton instance
export const authManager = new AuthManager();
