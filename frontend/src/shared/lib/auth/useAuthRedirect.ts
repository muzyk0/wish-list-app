'use client';

import { useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';

/**
 * Hook to check authentication status and optionally redirect authenticated users
 * to the mobile app for account management.
 *
 * @param shouldRedirect - If true, authenticated users will be redirected to mobile app
 * @returns Object with authentication status and loading state
 */
export function useAuthRedirect(shouldRedirect = false) {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null);
  const _router = useRouter();

  useEffect(() => {
    const checkAuth = async () => {
      try {
        const response = await fetch('/api/auth/me');
        const authenticated = response.ok;
        setIsAuthenticated(authenticated);

        // If authenticated and should redirect, navigate to mobile redirect page
        if (authenticated && shouldRedirect) {
          // Don't use router.push as it won't trigger the MobileRedirect component
          // The component itself will handle the redirection
          return;
        }
      } catch {
        setIsAuthenticated(false);
      }
    };

    checkAuth();
  }, [shouldRedirect]);

  return { isAuthenticated, isLoading: isAuthenticated === null };
}
