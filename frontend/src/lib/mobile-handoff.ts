// frontend/src/lib/mobile-handoff.ts
// Mobile handoff: Transfer authenticated session from Frontend to Mobile app

const API_BASE_URL =
  process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api';

const MOBILE_URL_SCHEME = 'wishlistapp://';

interface HandoffResponse {
  code: string;
  expiresAt: string;
}

/**
 * Generate handoff code and redirect to Mobile app
 * Uses OAuth-style flow:
 * 1. Call POST /auth/mobile-handoff to generate short-lived code
 * 2. Redirect to mobile app via Universal Link: wishlistapp://auth?code=xxx
 * 3. Mobile app exchanges code for tokens via POST /auth/exchange
 */
export async function redirectToPersonalCabinet(): Promise<void> {
  try {
    // Call backend to generate handoff code
    const response = await fetch(`${API_BASE_URL}/auth/mobile-handoff`, {
      method: 'POST',
      credentials: 'include', // Include httpOnly cookie for authentication
      headers: {
        'Content-Type': 'application/json',
      },
    });

    if (!response.ok) {
      throw new Error('Failed to generate handoff code');
    }

    const data: HandoffResponse = await response.json();

    // Redirect to mobile app with code
    const mobileUrl = `${MOBILE_URL_SCHEME}auth?code=${data.code}`;
    window.location.href = mobileUrl;

    // Fallback: If mobile app not installed, show error after delay
    setTimeout(() => {
      alert(
        'Mobile app not found. Please install the Wish List app from the App Store or Google Play.',
      );
    }, 3000);
  } catch (error) {
    console.error('Mobile handoff failed:', error);
    alert('Failed to redirect to mobile app. Please try again.');
  }
}
