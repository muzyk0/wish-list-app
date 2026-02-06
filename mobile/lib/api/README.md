# API Client Documentation

## Overview

The API layer provides type-safe HTTP client for the mobile application using **openapi-fetch** with automatic authentication and token refresh.

## Architecture

![Architecture Diagram](diagrams/architecture.puml)

<details>
<summary>View ASCII diagram</summary>

```
┌─────────────────────────────────────────────────────────────┐
│                        API LAYER                             │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌────────────────┐                                          │
│  │  client.ts     │  ← Base client (no middleware)          │
│  │  - baseClient  │                                          │
│  │  - API_BASE_URL│                                          │
│  └────────────────┘                                          │
│         │                                                     │
│         ├──────────────────┬──────────────────────────────── │
│         │                  │                                 │
│         ▼                  ▼                                 │
│  ┌─────────────┐    ┌──────────────┐                       │
│  │  auth.ts    │    │   api.ts     │                       │
│  │             │    │              │                       │
│  │ Uses:       │    │ Creates:     │                       │
│  │ baseClient  │    │ NEW client   │                       │
│  │             │    │ WITH         │                       │
│  │ Functions:  │    │ middleware   │                       │
│  │ • login     │    │              │                       │
│  │ • refresh   │    │ Middleware:  │                       │
│  │ • exchange  │    │ • auth       │                       │
│  │ • logout    │    │ • refresh    │                       │
│  │ • tokens    │    │              │                       │
│  └─────────────┘    │ ApiClient    │                       │
│         │            └──────────────┘                       │
│         │                  │                                 │
│         │                  ▼                                 │
│         └──────────→ Components                             │
│                                                               │
└─────────────────────────────────────────────────────────────┘
```

</details>

## File Structure

```
lib/api/
├── client.ts          # Base openapi-fetch client (single source of truth)
├── auth.ts            # Authentication operations (uses baseClient)
├── api.ts             # ApiClient class (creates client + middleware)
├── schema.ts          # OpenAPI types (auto-generated)
├── types.ts           # Additional TypeScript types
└── index.ts           # Public exports
```

## Two Clients Pattern

### Why Two Clients?

The API layer uses **two separate openapi-fetch clients** to prevent infinite recursion:

1. **baseClient** (in `client.ts`) - Clean HTTP client without middleware
2. **ApiClient.client** (in `api.ts`) - Client with auth & refresh middleware

### The Problem with One Client

![Infinite Recursion Problem](diagrams/infinite-recursion-problem.puml)

If we used a single client with middleware for both auth operations and protected endpoints:

```typescript
// ❌ WRONG: Infinite recursion!

// api.ts
this.client = baseClient;  // Reuse baseClient
this.client.use(refreshMiddleware);  // Add middleware

// What happens:
// 1. GET /protected/profile → 401
// 2. refreshMiddleware → calls refreshAccessToken()
// 3. refreshAccessToken() → baseClient.POST('/auth/refresh')
// 4. baseClient NOW HAS middleware!
// 5. POST /auth/refresh → 401 (refresh token expired)
// 6. refreshMiddleware → calls refreshAccessToken()
// 7. → INFINITE RECURSION! 💥
```

### The Solution

![Two Clients Solution](diagrams/two-clients-solution.puml)

```typescript
// ✅ CORRECT: Two separate clients

// client.ts - Clean client without middleware
export const baseClient = createClient<paths>({ baseUrl: API_BASE_URL });

// auth.ts - Uses clean baseClient
import { baseClient } from './client';
await baseClient.POST('/auth/refresh', ...);  // No middleware

// api.ts - Creates NEW client with middleware
this.client = createClient<paths>({ baseUrl: API_BASE_URL });
this.client.use(authMiddleware);     // Only for this client
this.client.use(refreshMiddleware);  // Only for this client
```

### Client Responsibilities

| Client | Location | Middleware | Purpose |
|--------|----------|------------|---------|
| `baseClient` | auth.ts | ❌ No | Auth operations (prevents recursion) |
| `ApiClient.client` | api.ts | ✅ Yes | Protected endpoints (automatic auth & refresh) |

## Middleware

![Middleware Execution Flow](diagrams/middleware-execution.puml)

### 1. Authentication Middleware

Automatically adds `Authorization: Bearer <token>` header to protected endpoints.

```typescript
private authMiddleware: Middleware = {
  async onRequest({ request }) {
    // Skip auth for unprotected routes
    const url = new URL(request.url);
    const isUnprotected = UNPROTECTED_ROUTES.some((route) =>
      url.pathname.includes(route),
    );

    if (isUnprotected) {
      return request;
    }

    // Add Authorization header for protected routes
    const token = await getAccessToken();
    if (token) {
      request.headers.set('Authorization', `Bearer ${token}`);
    }

    return request;
  },
};
```

**Unprotected Routes:**
```typescript
const UNPROTECTED_ROUTES = ['/auth/login', '/auth/register'];
```

**Why Skip Unprotected Routes?**

1. **Security Best Practice** - Don't send credentials when not needed
2. **Backend Logic** - Auth endpoints may reject requests with existing tokens
3. **Cleaner API** - Endpoints receive only what they need
4. **Easier Debugging** - Clear separation of public/protected endpoints

### 2. Token Refresh Middleware

Automatically refreshes access token on 401 errors and retries the request.

```typescript
private refreshMiddleware: Middleware = {
  onResponse: async ({ request, response }) => {
    // Check for 401 Unauthorized
    if (response.status === 401 && !this.isRefreshing) {
      this.isRefreshing = true;  // Prevent concurrent refreshes

      try {
        // Attempt to refresh the token
        if (!this.refreshPromise) {
          this.refreshPromise = refreshAccessToken();
        }

        const newToken = await this.refreshPromise;

        if (newToken) {
          // Clone the original request with the new token
          const retryRequest = request.clone();
          retryRequest.headers.set('Authorization', `Bearer ${newToken}`);

          // Retry the request with the new token
          return await fetch(retryRequest);
        }

        // If refresh failed, clear tokens
        await clearTokens();
        return response;
      } finally {
        this.isRefreshing = false;
        this.refreshPromise = null;
      }
    }

    return response;
  },
};
```

**Key Features:**

1. **Singleton Pattern** - `refreshPromise` prevents multiple concurrent refresh requests
2. **Automatic Retry** - Original request is retried with new token
3. **Graceful Failure** - Clears tokens if refresh fails

### Why Native `fetch()` in Retry?

The middleware uses native `fetch()` for retrying requests instead of `this.client.GET/POST/etc`:

```typescript
// ✅ CORRECT: Native fetch
return await fetch(retryRequest);

// ❌ WRONG: Would need to determine method and extract params
return await this.client.GET(???, ???);  // How to get path/params?
```

**Reasons:**

1. **Unknown HTTP Method** - We have a `Request` object but don't know which openapi-fetch method to call
2. **Type Safety Loss** - Can't extract typed path from URL string
3. **Complex Parameter Extraction** - Would need to parse URL, body, headers
4. **Infinite Loop Risk** - Would go through middleware again
5. **Simplicity** - Native fetch is straightforward and works

**This is the correct pattern recommended by openapi-fetch documentation.**

## Usage

### Authentication Operations

```typescript
// Login
const response = await apiClient.login({
  email: 'user@example.com',
  password: 'password123',
});

// Register
const response = await apiClient.register({
  email: 'user@example.com',
  password: 'password123',
  first_name: 'John',
  last_name: 'Doe',
});

// Logout
await apiClient.logout();
```

### Protected Endpoints

```typescript
// Get user profile (automatic auth + refresh on 401)
const user = await apiClient.getProfile();

// Get wishlists (automatic auth + refresh on 401)
const wishlists = await apiClient.getWishLists();

// Create wishlist
const wishlist = await apiClient.createWishList({
  name: 'My Wishlist',
  description: 'Birthday wishlist',
});
```

### Token Management

```typescript
// Get access token
const token = await getAccessToken();

// Check if authenticated
const isAuth = await isAuthenticated();

// Clear all tokens (logout)
await clearTokens();
```

## Flow Examples

### Successful Request Flow

![Successful Request Flow](diagrams/successful-request-flow.puml)

<details>
<summary>View text flow</summary>

```
1. User: apiClient.getProfile()
2. authMiddleware: Add Authorization header
3. Backend: 200 OK
4. Return: User profile data
```

</details>

### Token Refresh Flow

![Token Refresh Flow](diagrams/token-refresh-flow.puml)

<details>
<summary>View text flow</summary>

```
1. User: apiClient.getProfile()
2. authMiddleware: Add Authorization: Bearer <expired_token>
3. Backend: 401 Unauthorized
4. refreshMiddleware: Detect 401
5. refreshMiddleware: Call refreshAccessToken()
6. refreshAccessToken(): baseClient.POST('/auth/refresh')
7. Backend: Return new tokens
8. refreshMiddleware: Retry original request with new token
9. Backend: 200 OK
10. Return: User profile data
```

**Note:** The retry uses native `fetch()` with the new token, not `this.client`, to avoid middleware re-execution.

</details>

### Failed Refresh Flow

![Failed Refresh Flow](diagrams/failed-refresh-flow.puml)

<details>
<summary>View text flow</summary>

```
1. User: apiClient.getProfile()
2. authMiddleware: Add Authorization: Bearer <expired_token>
3. Backend: 401 Unauthorized
4. refreshMiddleware: Detect 401
5. refreshMiddleware: Call refreshAccessToken()
6. refreshAccessToken(): baseClient.POST('/auth/refresh')
7. Backend: 401 Unauthorized (refresh token also expired)
8. refreshAccessToken(): Clear tokens, return null
9. refreshMiddleware: Return 401 to caller
10. App: Redirect to login screen
```

</details>

## Security Features

### 1. Token Storage

- **Native**: Expo SecureStore (iOS Keychain, Android Keystore)
- **Web**: AsyncStorage (less secure but available)

### 2. Token Isolation

- Access token: 15 minutes
- Refresh token: 7 days
- Automatic refresh on 401 errors
- Clear tokens on failed refresh

### 3. Credential Handling

- No credentials sent to unprotected endpoints
- Authorization header only on protected routes
- Tokens cleared on logout and account deletion

### 4. Race Condition Prevention

- Singleton refresh promise prevents concurrent token refreshes
- `isRefreshing` flag prevents infinite recursion
- All concurrent 401 requests await the same refresh operation

## Type Safety

All API methods are fully type-safe using auto-generated OpenAPI types:

```typescript
// TypeScript knows the exact shape of request/response
const wishlist = await apiClient.createWishList({
  name: 'string',        // ✅ Type-safe
  description: 'string', // ✅ Type-safe
  // invalidField: 123   // ❌ TypeScript error
});

// Return type is inferred
const user: User = await apiClient.getProfile();
```

## Error Handling

### Authentication Errors

```typescript
try {
  await apiClient.login({ email, password });
} catch (error) {
  // Handle login failure
  console.error('Login failed:', error.message);
}
```

### API Errors

```typescript
try {
  const wishlists = await apiClient.getWishLists();
} catch (error) {
  // Handle API error
  console.error('Failed to fetch wishlists:', error.message);
}
```

### Token Refresh Errors

Token refresh errors are handled automatically:
- Failed refresh → tokens cleared
- User redirected to login screen
- No manual error handling needed

## Testing

### Mock baseClient

```typescript
jest.mock('./client', () => ({
  baseClient: {
    POST: jest.fn().mockResolvedValue({
      data: { accessToken: 'token', refreshToken: 'refresh' }
    }),
  },
}));
```

### Mock ApiClient

```typescript
const mockApiClient = {
  login: jest.fn(),
  getProfile: jest.fn(),
  getWishLists: jest.fn(),
};
```

### Test Middleware

```typescript
describe('authMiddleware', () => {
  it('should add Authorization header to protected routes', async () => {
    // Test implementation
  });

  it('should skip Authorization header for unprotected routes', async () => {
    // Test implementation
  });
});

describe('refreshMiddleware', () => {
  it('should refresh token on 401', async () => {
    // Test implementation
  });

  it('should clear tokens on failed refresh', async () => {
    // Test implementation
  });
});
```

## Performance Considerations

### baseClient (Lightweight)

- No middleware overhead
- Direct HTTP calls
- Faster for auth operations

### ApiClient (Feature-Rich)

- Small middleware overhead
- Automatic token management
- Better UX (seamless token refresh)

### Optimization Tips

1. **Singleton Pattern** - Only one refresh request at a time
2. **Request Cloning** - Efficient retry mechanism
3. **Conditional Middleware** - Skip unnecessary auth checks

## Common Pitfalls

### ❌ Don't Mutate baseClient

```typescript
// ❌ WRONG: Mutates shared baseClient
import { baseClient } from './client';
baseClient.use(someMiddleware);  // Affects all consumers!
```

### ❌ Don't Mix Clients

```typescript
// ❌ WRONG: Using wrong client for auth
import { apiClient } from './api';
await apiClient.login();  // Should use baseClient in auth.ts
```

### ✅ Do Use Correct Client

```typescript
// ✅ CORRECT: Auth operations use baseClient
// auth.ts
await baseClient.POST('/auth/login', ...);

// ✅ CORRECT: Protected endpoints use ApiClient
// components
await apiClient.getProfile();
```

## Configuration

### Change API Base URL

```typescript
// client.ts
const API_BASE_URL =
  process.env.EXPO_PUBLIC_API_URL || 'http://10.0.2.2:8080/api';
```

### Add Global Headers

```typescript
// client.ts
export const baseClient = createClient<paths>({
  baseUrl: API_BASE_URL,
  headers: {
    'X-App-Version': '1.0.0',
    'X-Platform': Platform.OS,
  },
});
```

### Add Timeout

```typescript
// client.ts
export const baseClient = createClient<paths>({
  baseUrl: API_BASE_URL,
  signal: AbortSignal.timeout(10000), // 10s timeout
});
```

## Best Practices

1. **Always use apiClient for protected endpoints** - Don't bypass middleware
2. **Use baseClient only in auth.ts** - Keep it clean for auth operations
3. **Don't send tokens to unprotected routes** - Security best practice
4. **Let middleware handle token refresh** - Don't manually refresh
5. **Clear tokens on logout** - Always clean up
6. **Type-safe everywhere** - Use generated OpenAPI types

## Summary

| Component | Purpose | Middleware | Usage |
|-----------|---------|------------|-------|
| `baseClient` | Auth operations | No | auth.ts only |
| `ApiClient` | Protected endpoints | Yes | Components, hooks |
| `authMiddleware` | Add auth headers | - | Protected routes |
| `refreshMiddleware` | Auto token refresh | - | Handle 401 errors |

**Key Principles:**
- ✅ Two clients prevent infinite recursion
- ✅ Middleware provides seamless UX
- ✅ Type-safe API layer
- ✅ Security best practices
- ✅ Automatic token management
