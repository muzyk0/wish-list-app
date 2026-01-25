'use client';

import { useEffect } from 'react';

interface MobileRedirectProps {
  redirectPath?: string; // Specific path to redirect to in the mobile app
  fallbackUrl?: string; // Fallback web URL if mobile app is not installed
  children?: React.ReactNode; // Optional content to show while attempting redirect
}

export default function MobileRedirect({
  redirectPath = '',
  fallbackUrl = 'https://lk.domain.com',
  children,
}: MobileRedirectProps) {
  useEffect(() => {
    const redirectToMobile = () => {
      // Construct the app-specific URL scheme
      const appScheme = `wishlistapp://${redirectPath || 'home'}`;

      // Fallback URL for web version
      const webFallback = fallbackUrl;

      // Try to open the mobile app
      window.location.href = appScheme;

      // If the app isn't installed, redirect to the web version after a delay
      setTimeout(() => {
        // If the page is still visible (not hidden), the app wasn't opened
        if (!document.hidden && document.visibilityState !== 'hidden') {
          window.location.href = webFallback;
        }
      }, 1500);
    };

    redirectToMobile();
  }, [redirectPath, fallbackUrl]);

  return children || null;
}
