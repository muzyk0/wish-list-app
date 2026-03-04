// frontend/src/hooks/useAuth.ts
// React hook for authentication state management
// - Provides authentication state across components
// - Automatically attempts token refresh on mount
// - Handles loading and error states

import { useCallback, useEffect, useState } from 'react';
import { authManager } from '@/shared/api';

interface UseAuthReturn {
  isAuthenticated: boolean;
  isLoading: boolean;
  error: string | null;
  refreshAuth: () => Promise<void>;
}

/**
 * Hook for managing authentication state
 *
 * Features:
 * - Automatic token refresh attempt on mount (silent)
 * - Loading state during refresh
 * - Error state if refresh fails
 * - Manual refresh capability
 *
 * @returns Authentication state and control methods
 *
 * @example
 * ```tsx
 * function ProtectedPage() {
 *   const { isAuthenticated, isLoading } = useAuth();
 *
 *   if (isLoading) return <Spinner />;
 *   if (!isAuthenticated) return <LoginPrompt />;
 *   return <ProtectedContent />;
 * }
 * ```
 */
export function useAuth(): UseAuthReturn {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(
    authManager.isAuthenticated(),
  );
  const [isLoading, setIsLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  const refreshAuth = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const newToken = await authManager.refreshAccessToken();

      if (newToken) {
        setIsAuthenticated(true);
      } else {
        setIsAuthenticated(false);
      }
    } catch (err) {
      setError(
        err instanceof Error ? err.message : 'Failed to refresh authentication',
      );
      setIsAuthenticated(false);
    } finally {
      setIsLoading(false);
    }
  }, []);

  // Attempt to refresh token on mount (silent authentication)
  useEffect(() => {
    const initAuth = async () => {
      // If already authenticated (access token in memory), no need to refresh
      if (authManager.isAuthenticated()) {
        setIsLoading(false);
        return;
      }

      // Otherwise, try to refresh using httpOnly cookie
      await refreshAuth();
    };

    initAuth();
  }, [refreshAuth]);

  return {
    isAuthenticated,
    isLoading,
    error,
    refreshAuth,
  };
}
