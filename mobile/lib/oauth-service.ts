import * as AuthSession from 'expo-auth-session';
import { Platform } from 'react-native';

// Define OAuth configuration
const OAUTH_CONFIG = {
  // Google OAuth
  googleWebClientId:
    process.env.EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID || 'GOOGLE_WEB_CLIENT_ID',
  googleAndroidClientId:
    process.env.EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID ||
    'GOOGLE_ANDROID_CLIENT_ID',
  googleIOSClientId:
    process.env.EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID || 'GOOGLE_IOS_CLIENT_ID',

  // Facebook OAuth
  facebookClientId:
    process.env.EXPO_PUBLIC_FACEBOOK_CLIENT_ID || 'FACEBOOK_CLIENT_ID',

  // Backend API URL for OAuth callback
  backendUrl: process.env.EXPO_PUBLIC_API_URL || 'http://localhost:8080',
};

// Google OAuth discovery endpoint
const googleDiscovery = AuthSession.discovery({
  authorizationEndpoint: 'https://accounts.google.com/o/oauth2/v2/auth',
  tokenEndpoint: 'https://oauth2.googleapis.com/token',
  revocationEndpoint: 'https://oauth2.googleapis.com/revoke',
});

// Google OAuth flow
export const startGoogleOAuth = async (): Promise<{
  success: boolean;
  token?: string;
  error?: string;
}> => {
  try {
    // Determine the correct client ID based on platform
    let clientId = '';
    if (Platform.OS === 'web') {
      clientId = OAUTH_CONFIG.googleWebClientId;
    } else if (Platform.OS === 'android') {
      clientId = OAUTH_CONFIG.googleAndroidClientId;
    } else if (Platform.OS === 'ios') {
      clientId = OAUTH_CONFIG.googleIOSClientId;
    }

    if (!clientId || clientId.includes('CLIENT_ID')) {
      return {
        success: false,
        error:
          'OAuth is not configured. Please set up your Google client ID in environment variables.',
      };
    }

    // Get app scheme from environment variable
    const appScheme =
      process.env.EXPO_PUBLIC_APP_SCHEME || process.env.APP_SCHEME;
    if (!appScheme) {
      throw new Error(
        'APP_SCHEME environment variable is required for OAuth. Please set EXPO_PUBLIC_APP_SCHEME in your .env file.',
      );
    }

    // Create redirect URI
    const redirectUri = AuthSession.makeRedirectUri({
      native: `${appScheme}://oauth`,
      preferLocalhost: true,
    });

    // Create AuthRequest with PKCE enabled
    const request = new AuthSession.AuthRequest({
      clientId,
      redirectUri,
      scopes: ['openid', 'profile', 'email'],
      usePKCE: true,
      extraParams: {
        access_type: 'offline', // Request refresh token
      },
    });

    // Prompt for authorization
    const result = await request.promptAsync(googleDiscovery);

    if (result.type === 'success') {
      const { code } = result.params;

      if (code) {
        // In a real implementation, you would exchange the code for tokens
        // The code is now extracted automatically by AuthRequest
        return { success: true, token: code };
      } else {
        return { success: false, error: 'Authorization code not received' };
      }
    } else if (result.type === 'dismiss' || result.type === 'cancel') {
      return { success: false, error: 'OAuth flow was cancelled' };
    } else if (result.type === 'error') {
      return {
        success: false,
        error: result.params?.error_description || 'OAuth flow failed',
      };
    } else {
      return { success: false, error: 'OAuth flow failed' };
    }
    // biome-ignore lint/suspicious/noExplicitAny: Error type
  } catch (error: any) {
    console.error('Google OAuth error:', error);
    return {
      success: false,
      error: error.message || 'An error occurred during OAuth',
    };
  }
};

// Facebook OAuth discovery endpoint
const facebookDiscovery = AuthSession.discovery({
  authorizationEndpoint: 'https://www.facebook.com/v18.0/dialog/oauth',
  tokenEndpoint: 'https://graph.facebook.com/v18.0/oauth/access_token',
});

// Facebook OAuth flow
export const startFacebookOAuth = async (): Promise<{
  success: boolean;
  token?: string;
  error?: string;
}> => {
  try {
    const facebookClientId = OAUTH_CONFIG.facebookClientId;

    if (!facebookClientId || facebookClientId.includes('CLIENT_ID')) {
      return {
        success: false,
        error:
          'OAuth is not configured. Please set up your Facebook client ID in environment variables.',
      };
    }

    // Get app scheme from environment variable
    const appScheme =
      process.env.EXPO_PUBLIC_APP_SCHEME || process.env.APP_SCHEME;
    if (!appScheme) {
      throw new Error(
        'APP_SCHEME environment variable is required for OAuth. Please set EXPO_PUBLIC_APP_SCHEME in your .env file.',
      );
    }

    // Create redirect URI
    const redirectUri = AuthSession.makeRedirectUri({
      native: `${appScheme}://oauth`,
      preferLocalhost: true,
    });

    // Create AuthRequest with PKCE enabled
    const request = new AuthSession.AuthRequest({
      clientId: facebookClientId,
      redirectUri,
      scopes: ['email', 'public_profile'],
      usePKCE: true,
    });

    // Prompt for authorization
    const result = await request.promptAsync(facebookDiscovery);

    if (result.type === 'success') {
      const { code } = result.params;

      if (code) {
        // In a real implementation, you would exchange the code for tokens
        // The code is now extracted automatically by AuthRequest
        return { success: true, token: code };
      } else {
        return { success: false, error: 'Authorization code not received' };
      }
    } else if (result.type === 'dismiss' || result.type === 'cancel') {
      return { success: false, error: 'OAuth flow was cancelled' };
    } else if (result.type === 'error') {
      return {
        success: false,
        error: result.params?.error_description || 'OAuth flow failed',
      };
    } else {
      return { success: false, error: 'OAuth flow failed' };
    }
    // biome-ignore lint/suspicious/noExplicitAny: Error type
  } catch (error: any) {
    console.error('Facebook OAuth error:', error);
    return {
      success: false,
      error: error.message || 'An error occurred during OAuth',
    };
  }
};

// Apple OAuth flow (iOS only)
export const startAppleOAuth = async (): Promise<{
  success: boolean;
  token?: string;
  error?: string;
}> => {
  // Apple Sign In requires additional setup and is iOS-specific
  // This is a placeholder implementation
  if (Platform.OS !== 'ios') {
    return {
      success: false,
      error: 'Apple Sign In is only available on iOS devices',
    };
  }

  // For now, return a mock response
  // In a real implementation, you'd use expo-apple-authentication
  return {
    success: false,
    error:
      'Apple Sign In requires additional setup and is not implemented in this demo',
  };
};
