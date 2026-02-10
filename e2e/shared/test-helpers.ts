/**
 * Shared test helpers and utilities
 * Used across backend and mobile E2E tests
 */

import { APIRequestContext } from '@playwright/test';

export const API_BASE = 'http://localhost:8080/api';

/**
 * Test user factory
 */
export function createTestUser(prefix: string = 'test') {
  return {
    email: `${prefix}_${Date.now()}@example.com`,
    password: 'Test123456!',
    first_name: 'Test',
    last_name: 'User',
  };
}

/**
 * Register a user and return auth token
 */
export async function registerAndLogin(
  request: APIRequestContext,
  user?: ReturnType<typeof createTestUser>
): Promise<{ token: string; userId: string; user: any }> {
  const testUser = user || createTestUser();

  const response = await request.post(`${API_BASE}/auth/register`, {
    data: testUser,
  });

  const data = await response.json();

  return {
    token: data.accessToken,
    userId: data.user.id,
    user: data.user,
  };
}

/**
 * Create a test wishlist
 */
export async function createTestWishlist(
  request: APIRequestContext,
  token: string,
  data?: Partial<{
    name: string;
    description: string;
    is_public: boolean;
  }>
) {
  const response = await request.post(`${API_BASE}/wishlists`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    data: {
      name: data?.name || 'Test Wishlist',
      description: data?.description || 'For testing',
      is_public: data?.is_public ?? false,
    },
  });

  return await response.json();
}

/**
 * Create a test item
 */
export async function createTestItem(
  request: APIRequestContext,
  token: string,
  data?: Partial<{
    name: string;
    description: string;
    link: string;
    price: number;
    priority: number;
  }>
) {
  const response = await request.post(`${API_BASE}/items`, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
    data: {
      name: data?.name || 'Test Item',
      description: data?.description || 'Test description',
      link: data?.link,
      price: data?.price ?? 99.99,
      priority: data?.priority ?? 0,
    },
  });

  return await response.json();
}

/**
 * Wait for condition with timeout
 */
export async function waitForCondition(
  condition: () => Promise<boolean>,
  options: { timeout?: number; interval?: number } = {}
): Promise<void> {
  const timeout = options.timeout || 5000;
  const interval = options.interval || 100;
  const startTime = Date.now();

  while (Date.now() - startTime < timeout) {
    if (await condition()) {
      return;
    }
    await new Promise((resolve) => setTimeout(resolve, interval));
  }

  throw new Error(`Condition not met within ${timeout}ms`);
}

/**
 * Generate random string
 */
export function randomString(length: number = 10): string {
  return Math.random().toString(36).substring(2, length + 2);
}

/**
 * Assert response status with custom message
 */
export function assertStatus(
  actual: number,
  expected: number,
  context?: string
) {
  if (actual !== expected) {
    throw new Error(
      `Expected status ${expected} but got ${actual}${context ? ` (${context})` : ''}`
    );
  }
}
