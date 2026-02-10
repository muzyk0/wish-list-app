/**
 * E2E Tests for CORS Protection (Phase 8 - User Story 6)
 *
 * Test Strategy:
 * 1. Verify allowed origins receive correct CORS headers
 * 2. Verify disallowed origins are blocked
 * 3. Verify credentials are enabled for cross-domain cookies
 * 4. Test preflight OPTIONS requests
 */

import { test, expect } from '@playwright/test';

const API_BASE_URL = 'http://localhost:8080';
const ALLOWED_ORIGINS = [
  'http://localhost:3000',    // Frontend
  'http://localhost:19006',   // Mobile Expo
  'http://localhost:8081',    // Mobile development
];
const DISALLOWED_ORIGIN = 'http://malicious-site.com';

test.describe('CORS Protection - Phase 8', () => {

  test('T045: CORS middleware allows requests from configured origins', async ({ request }) => {
    for (const origin of ALLOWED_ORIGINS) {
      const response = await request.fetch(`${API_BASE_URL}/api/auth/login`, {
        method: 'OPTIONS',
        headers: {
          'Origin': origin,
          'Access-Control-Request-Method': 'POST',
          'Access-Control-Request-Headers': 'Content-Type',
        },
      });

      // Should allow the request
      expect([200, 204]).toContain(response.status());

      // Verify CORS headers are present
      const headers = response.headers();
      expect(headers['access-control-allow-origin']).toBe(origin);

      console.log(`✓ Allowed origin ${origin} received CORS headers`);
    }
  });

  test('T046: CORS middleware sets Access-Control-Allow-Credentials: true', async ({ request }) => {
    const response = await request.fetch(`${API_BASE_URL}/api/auth/refresh`, {
      method: 'OPTIONS',
      headers: {
        'Origin': 'http://localhost:3000',
        'Access-Control-Request-Method': 'POST',
      },
    });

    const headers = response.headers();

    // Credentials MUST be true for httpOnly cookies to work cross-domain
    expect(headers['access-control-allow-credentials']).toBe('true');

    console.log('✓ Credentials enabled for cross-domain cookies');
  });

  test('T047: All development origins are whitelisted', async ({ request }) => {
    const expectedOrigins = [
      'http://localhost:3000',    // Frontend Next.js
      'http://localhost:19006',   // Mobile Expo default port
      'http://localhost:8081',    // Mobile alternative port
    ];

    for (const origin of expectedOrigins) {
      const response = await request.fetch(`${API_BASE_URL}/api/auth/login`, {
        method: 'OPTIONS',
        headers: {
          'Origin': origin,
          'Access-Control-Request-Method': 'POST',
        },
      });

      const headers = response.headers();
      expect(headers['access-control-allow-origin']).toBe(origin);

      console.log(`✓ Development origin ${origin} is whitelisted`);
    }
  });

  test('T049: Preflight OPTIONS requests return correct CORS headers', async ({ request }) => {
    const response = await request.fetch(`${API_BASE_URL}/api/auth/login`, {
      method: 'OPTIONS',
      headers: {
        'Origin': 'http://localhost:3000',
        'Access-Control-Request-Method': 'POST',
        'Access-Control-Request-Headers': 'Content-Type, Authorization',
      },
    });

    const headers = response.headers();

    // Verify all required CORS headers
    expect(headers['access-control-allow-origin']).toBe('http://localhost:3000');
    expect(headers['access-control-allow-credentials']).toBe('true');
    expect(headers['access-control-allow-methods']).toContain('POST');
    expect(headers['access-control-max-age']).toBe('86400'); // 24 hours

    console.log('✓ Preflight OPTIONS returns correct CORS headers');
  });

  test('Disallowed origin does NOT receive CORS headers', async ({ request }) => {
    const response = await request.fetch(`${API_BASE_URL}/api/auth/login`, {
      method: 'OPTIONS',
      headers: {
        'Origin': DISALLOWED_ORIGIN,
        'Access-Control-Request-Method': 'POST',
      },
    });

    const headers = response.headers();

    // Disallowed origin should NOT receive Access-Control-Allow-Origin header
    // or it should not match the requesting origin
    const allowOriginHeader = headers['access-control-allow-origin'];
    expect(allowOriginHeader).not.toBe(DISALLOWED_ORIGIN);

    console.log('✓ Disallowed origin blocked (no matching CORS headers)');
  });

  test('CORS headers support all required HTTP methods', async ({ request }) => {
    const methods = ['GET', 'POST', 'PUT', 'DELETE', 'OPTIONS'];

    const response = await request.fetch(`${API_BASE_URL}/api/wishlists`, {
      method: 'OPTIONS',
      headers: {
        'Origin': 'http://localhost:3000',
        'Access-Control-Request-Method': 'POST',
      },
    });

    const headers = response.headers();
    const allowedMethods = headers['access-control-allow-methods'] || '';

    for (const method of methods) {
      expect(allowedMethods).toContain(method);
    }

    console.log('✓ All required HTTP methods are allowed');
  });

  test('CORS headers expose Authorization header', async ({ request }) => {
    // Register a user first to get auth headers
    const registerResponse = await request.post(`${API_BASE_URL}/api/auth/register`, {
      data: {
        email: `cors-auth-test-${Date.now()}@example.com`,
        password: 'TestPassword123!',
        first_name: 'CORS',
        last_name: 'Auth',
      },
      headers: {
        'Origin': 'http://localhost:3000',
      },
    });

    expect(registerResponse.ok()).toBeTruthy();
    const registerData = await registerResponse.json();
    expect(registerData).toHaveProperty('accessToken');

    // Now make a request that includes Authorization header
    const authResponse = await request.get(`${API_BASE_URL}/api/auth/profile`, {
      headers: {
        'Origin': 'http://localhost:3000',
        'Authorization': `Bearer ${registerData.accessToken}`,
      },
    });

    const headers = authResponse.headers();
    const exposedHeaders = headers['access-control-expose-headers'] || '';

    // Authorization header should be exposed for JWT tokens
    expect(exposedHeaders.toLowerCase()).toContain('authorization');

    console.log('✓ Authorization header is exposed via CORS');
  });

  test('Real cross-origin request works with credentials', async ({ request }) => {
    // This simulates a real cross-domain request from Frontend to Backend
    // with credentials (cookies) enabled

    // First, register a test user
    const registerResponse = await request.post(`${API_BASE_URL}/api/auth/register`, {
      data: {
        email: `cors-test-${Date.now()}@example.com`,
        password: 'TestPassword123!',
        first_name: 'CORS',
        last_name: 'Test',
      },
      headers: {
        'Origin': 'http://localhost:3000',
      },
    });

    expect(registerResponse.ok()).toBeTruthy();

    const registerData = await registerResponse.json();
    expect(registerData).toHaveProperty('accessToken');
    expect(registerData).toHaveProperty('refreshToken');

    // Verify CORS headers were set on the actual request
    const corsHeaders = registerResponse.headers();
    expect(corsHeaders['access-control-allow-origin']).toBe('http://localhost:3000');
    expect(corsHeaders['access-control-allow-credentials']).toBe('true');

    console.log('✓ Real cross-origin request works with credentials');
  });

  test('CORS protection maintains security across all auth endpoints', async ({ request }) => {
    const authEndpoints = [
      '/api/auth/login',
      '/api/auth/register',
      '/api/auth/refresh',
      '/api/auth/logout',
      '/api/auth/mobile-handoff',
      '/api/auth/exchange',
    ];

    for (const endpoint of authEndpoints) {
      const response = await request.fetch(`${API_BASE_URL}${endpoint}`, {
        method: 'OPTIONS',
        headers: {
          'Origin': 'http://localhost:3000',
          'Access-Control-Request-Method': 'POST',
        },
      });

      const headers = response.headers();

      // Each auth endpoint should have CORS protection
      expect(headers['access-control-allow-origin']).toBe('http://localhost:3000');
      expect(headers['access-control-allow-credentials']).toBe('true');
    }

    console.log('✓ CORS protection applied to all auth endpoints');
  });
});

test.describe('CORS Protection - Edge Cases', () => {

  test('Missing Origin header does not break requests', async ({ request }) => {
    // Some requests (like from server-side) may not have Origin header
    const response = await request.get(`${API_BASE_URL}/healthz`, {
      // No Origin header
    });

    // Request should still succeed
    expect(response.ok()).toBeTruthy();

    console.log('✓ Requests without Origin header work correctly');
  });

  test('Case sensitivity in Origin matching', async ({ request }) => {
    // Test that origin matching is case-sensitive (security requirement)
    const response = await request.fetch(`${API_BASE_URL}/api/auth/login`, {
      method: 'OPTIONS',
      headers: {
        'Origin': 'http://LOCALHOST:3000', // Different case
        'Access-Control-Request-Method': 'POST',
      },
    });

    const headers = response.headers();
    const allowOriginHeader = headers['access-control-allow-origin'];

    // Should NOT match due to case difference (security best practice)
    expect(allowOriginHeader).not.toBe('http://LOCALHOST:3000');

    console.log('✓ Origin matching is case-sensitive');
  });

  test('Port number matters in origin matching', async ({ request }) => {
    // Different port should be treated as different origin
    const response = await request.fetch(`${API_BASE_URL}/api/auth/login`, {
      method: 'OPTIONS',
      headers: {
        'Origin': 'http://localhost:9999', // Wrong port
        'Access-Control-Request-Method': 'POST',
      },
    });

    const headers = response.headers();
    const allowOriginHeader = headers['access-control-allow-origin'];

    // Should NOT match due to wrong port
    expect(allowOriginHeader).not.toBe('http://localhost:9999');

    console.log('✓ Port number is enforced in origin matching');
  });
});
