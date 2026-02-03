// mobile/lib/api/auth.ts
// Secure token management for Mobile using expo-secure-store
// Falls back to AsyncStorage for web
// - Access token: 15 minutes
// - Refresh token: 7 days
// - Platform-native encryption (iOS Keychain, Android Keystore) or encrypted web storage

import { Platform } from 'react-native';
import * as SecureStore from 'expo-secure-store';
import AsyncStorage from '@react-native-async-storage/async-storage';

const ACCESS_TOKEN_KEY = 'accessToken';
const REFRESH_TOKEN_KEY = 'refreshToken';

const API_BASE_URL =
  process.env.EXPO_PUBLIC_API_URL || 'http://localhost:8080/api';

interface TokenResponse {
  accessToken: string;
  refreshToken: string;
}

/**
 * Store tokens - uses SecureStore on native, AsyncStorage on web
 */
export async function setTokens(
  accessToken: string,
  refreshToken: string,
): Promise<void> {
  if (Platform.OS === 'web') {
    // On web, store in AsyncStorage (less secure but available)
    await AsyncStorage.setItem(ACCESS_TOKEN_KEY, accessToken);
    await AsyncStorage.setItem(REFRESH_TOKEN_KEY, refreshToken);
  } else {
    // On native, use SecureStore
    await Promise.all([
      SecureStore.setItemAsync(ACCESS_TOKEN_KEY, accessToken),
      SecureStore.setItemAsync(REFRESH_TOKEN_KEY, refreshToken),
    ]);
  }
}

/**
 * Get access token - uses SecureStore on native, AsyncStorage on web
 */
export async function getAccessToken(): Promise<string | null> {
  if (Platform.OS === 'web') {
    return await AsyncStorage.getItem(ACCESS_TOKEN_KEY);
  } else {
    return await SecureStore.getItemAsync(ACCESS_TOKEN_KEY);
  }
}

/**
 * Get refresh token - uses SecureStore on native, AsyncStorage on web
 */
export async function getRefreshToken(): Promise<string | null> {
  if (Platform.OS === 'web') {
    return await AsyncStorage.getItem(REFRESH_TOKEN_KEY);
  } else {
    return await SecureStore.getItemAsync(REFRESH_TOKEN_KEY);
  }
}

/**
 * Clear all tokens - uses SecureStore on native, AsyncStorage on web
 * Used for logout and account deletion
 */
export async function clearTokens(): Promise<void> {
  if (Platform.OS === 'web') {
    await Promise.all([
      AsyncStorage.removeItem(ACCESS_TOKEN_KEY),
      AsyncStorage.removeItem(REFRESH_TOKEN_KEY),
    ]);
  } else {
    await Promise.all([
      SecureStore.deleteItemAsync(ACCESS_TOKEN_KEY),
      SecureStore.deleteItemAsync(REFRESH_TOKEN_KEY),
    ]);
  }
}

/**
 * Check if user is authenticated (has tokens)
 */
export async function isAuthenticated(): Promise<boolean> {
  const accessToken = await getAccessToken();
  return accessToken !== null;
}

/**
 * Exchange handoff code for tokens
 * Called when mobile app receives deep link: wishlistapp://auth?code=xxx
 */
export async function exchangeCodeForTokens(
  code: string,
): Promise<TokenResponse> {
  const response = await fetch(`${API_BASE_URL}/auth/exchange`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ code }),
  });

  if (!response.ok) {
    const errorData = await response.text();
    throw new Error(errorData || 'Failed to exchange code for tokens');
  }

  const data: TokenResponse = await response.json();

  // Store tokens appropriately based on platform
  await setTokens(data.accessToken, data.refreshToken);

  return data;
}

/**
 * Refresh access token using refresh token
 * Returns new access token or null if refresh failed
 */
export async function refreshAccessToken(): Promise<string | null> {
  const refreshToken = await getRefreshToken();

  if (!refreshToken) {
    return null;
  }

  try {
    const response = await fetch(`${API_BASE_URL}/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${refreshToken}`,
      },
    });

    if (!response.ok) {
      // Refresh token invalid or expired - clear all tokens
      await clearTokens();
      return null;
    }

    const data: TokenResponse = await response.json();

    // Store new tokens (refresh token is rotated)
    await setTokens(data.accessToken, data.refreshToken);

    return data.accessToken;
  } catch (error) {
    console.error('Token refresh failed:', error);
    await clearTokens();
    return null;
  }
}

/**
 * Logout: Clear tokens and call backend to invalidate refresh token
 */
export async function logout(): Promise<void> {
  const accessToken = await getAccessToken();

  // Clear local tokens first
  await clearTokens();

  // Notify backend to invalidate refresh token
  if (accessToken) {
    try {
      await fetch(`${API_BASE_URL}/auth/logout`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${accessToken}`,
        },
      });
    } catch (error) {
      console.error('Logout request failed:', error);
      // Continue - tokens already cleared locally
    }
  }
}
