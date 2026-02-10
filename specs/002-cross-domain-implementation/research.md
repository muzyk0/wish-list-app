# Technical Research: Cross-Domain Architecture

**Feature**: 002-cross-domain-implementation
**Date**: 2026-02-02

## Research Summary

This document captures technical decisions for implementing cross-domain authentication between Frontend (Vercel), Mobile (Expo), and Backend (Render).

---

## 1. Cross-Domain Cookie Strategy

### Decision
Use httpOnly cookies for refresh tokens on Frontend (same SameSite policy), Bearer tokens in Authorization header for Mobile.

### Rationale
- Frontend and Backend are on different domains (wishlist.com vs api.wishlist.com)
- Subdomains allow `SameSite=None; Secure` cookies with `Access-Control-Allow-Credentials: true`
- Mobile cannot use cookies reliably; Authorization header is standard practice

### Alternatives Considered
| Alternative | Rejected Because |
|-------------|------------------|
| localStorage for all tokens | XSS vulnerability, tokens accessible to JavaScript |
| Cookies only (no Bearer) | Mobile apps don't support cookies well |
| Same domain (proxy all through Frontend) | Added latency, complexity, Vercel function limits |

### Implementation Notes
```go
// Backend cookie settings for cross-subdomain
c.SetCookie(&http.Cookie{
    Name:     "refreshToken",
    Value:    token,
    Path:     "/",
    Domain:   ".wishlist.com", // Allows api.wishlist.com to set for wishlist.com
    HttpOnly: true,
    Secure:   true,
    SameSite: http.SameSiteNoneMode,
    MaxAge:   7 * 24 * 60 * 60, // 7 days
})
```

---

## 2. Token Refresh Strategy

### Decision
- Access token: 15 minutes, stored in memory (Frontend) / SecureStore (Mobile)
- Refresh token: 7 days, httpOnly cookie (Frontend) / SecureStore (Mobile)
- Rotation: Issue new refresh token on each refresh call

### Rationale
- Short access tokens minimize impact of token theft
- Refresh token rotation prevents token replay attacks
- 7-day refresh aligns with typical "remember me" UX expectations

### Alternatives Considered
| Alternative | Rejected Because |
|-------------|------------------|
| Longer access tokens (1h+) | Higher security risk if compromised |
| No refresh rotation | Enables token replay attacks |
| Sliding expiration only | Complex state management, sync issues |

### Implementation Notes
```typescript
// Frontend: Automatic refresh on 401
if (response.status === 401) {
    const newToken = await authManager.refreshAccessToken();
    if (newToken) {
        return this.request(endpoint, options); // Retry
    }
    throw new Error('Authentication required');
}
```

---

## 3. Handoff Code Implementation

### Decision
- In-memory store with automatic cleanup (not Redis)
- Cryptographically secure random codes (32 bytes, base64url encoded)
- One-time use with 60-second expiry
- Constant-time comparison to prevent timing attacks

### Rationale
- In-memory sufficient for MVP scale (codes live <60s, low volume)
- Redis adds infrastructure complexity without clear benefit at current scale
- Crypto-random prevents prediction attacks

### Alternatives Considered
| Alternative | Rejected Because |
|-------------|------------------|
| Redis-backed store | Overkill for short-lived codes, adds complexity |
| Database-backed | Unnecessary persistence, cleanup overhead |
| JWT-based codes | Larger size, no revocation benefit |

### Implementation Notes
```go
type CodeStore struct {
    mu    sync.RWMutex
    codes map[string]codeEntry
}

type codeEntry struct {
    UserID    uuid.UUID
    ExpiresAt time.Time
}

// Generate using crypto/rand
func generateSecureCode(length int) string {
    bytes := make([]byte, length)
    if _, err := rand.Read(bytes); err != nil {
        panic(err) // Should never fail
    }
    return base64.RawURLEncoding.EncodeToString(bytes)
}
```

---

## 4. CORS Configuration

### Decision
- Explicit allow list from environment variable
- Include `localhost:3000`, `localhost:8081`, `localhost:19006` for development
- Credentials mode enabled for cookie handling
- Preflight caching (24 hours)

### Rationale
- Environment-based config allows different settings per deployment
- Credentials required for cross-domain cookies
- Long preflight cache reduces OPTIONS request overhead

### Alternatives Considered
| Alternative | Rejected Because |
|-------------|------------------|
| Wildcard (*) origins | Not compatible with credentials mode, security risk |
| Per-request origin validation | Complex, error-prone |
| No CORS (proxy through Frontend) | Added latency, function complexity |

### Implementation Notes
```go
middleware.CORSConfig{
    AllowOrigins:     strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ","),
    AllowMethods:     []string{GET, POST, PUT, PATCH, DELETE, OPTIONS},
    AllowHeaders:     []string{Origin, Content-Type, Accept, Authorization},
    AllowCredentials: true,
    MaxAge:           86400, // 24 hours
}
```

---

## 5. Rate Limiting Strategy

### Decision
- IP-based rate limiting for auth endpoints
- User-based rate limiting for authenticated endpoints (mobile-handoff)
- In-memory token bucket algorithm (golang.org/x/time/rate)

### Rationale
- Prevents brute-force attacks on login
- Limits handoff code generation abuse
- Token bucket allows bursts while maintaining average rate

### Rate Limits
| Endpoint | Limit | Burst | Scope |
|----------|-------|-------|-------|
| POST /auth/login | 5/min | 10 | Per IP |
| POST /auth/exchange | 10/min | 15 | Per IP |
| POST /auth/mobile-handoff | 10/min | 15 | Per User |
| POST /auth/refresh | 20/min | 30 | Per IP |

### Alternatives Considered
| Alternative | Rejected Because |
|-------------|------------------|
| Redis-based rate limiting | Infrastructure overhead, in-memory sufficient |
| Fixed window | Less smooth, allows bursts at window boundaries |
| No rate limiting | Security risk |

---

## 6. Mobile Deep Link Handling

### Decision
- Custom URL scheme: `wishlistapp://`
- Universal Links (iOS) and App Links (Android) for HTTPS-based links
- Handle both cold start and warm start scenarios

### Rationale
- Custom scheme provides reliable fallback
- Universal/App Links provide seamless UX (no app chooser)
- Both scenarios must work for complete user experience

### Implementation Notes
```typescript
// expo-router _layout.tsx
useEffect(() => {
    // Cold start
    Linking.getInitialURL().then(url => {
        if (url) handleDeepLink(url);
    });

    // Warm start
    const subscription = Linking.addEventListener('url', ({ url }) => {
        handleDeepLink(url);
    });

    return () => subscription.remove();
}, []);
```

### Required Configuration
```json
// app.json
{
    "expo": {
        "scheme": "wishlistapp",
        "ios": {
            "associatedDomains": ["applinks:wishlist.com"]
        },
        "android": {
            "intentFilters": [{
                "action": "VIEW",
                "autoVerify": true,
                "data": [{ "scheme": "https", "host": "wishlist.com" }],
                "category": ["BROWSABLE", "DEFAULT"]
            }]
        }
    }
}
```

---

## 7. Frontend Token Storage

### Decision
- Access token: JavaScript variable (class property)
- Refresh token: httpOnly cookie (set by Backend)
- No localStorage/sessionStorage for any tokens

### Rationale
- Memory storage protects against XSS (tokens not accessible via document.*)
- httpOnly cookies protect refresh tokens from JavaScript access
- Tab/window refresh handled by automatic refresh call on page load

### Implementation Notes
```typescript
class AuthManager {
    private accessToken: string | null = null;

    // Singleton prevents duplicate refresh requests
    private refreshPromise: Promise<string | null> | null = null;

    async refreshAccessToken(): Promise<string | null> {
        if (this.refreshPromise) {
            return this.refreshPromise;
        }
        this.refreshPromise = this.doRefresh();
        const result = await this.refreshPromise;
        this.refreshPromise = null;
        return result;
    }
}
```

---

## 8. Mobile Secure Storage

### Decision
- Use `expo-secure-store` for both access and refresh tokens
- Clear on logout and account deletion
- No fallback to AsyncStorage

### Rationale
- expo-secure-store uses iOS Keychain and Android Keystore
- Platform-native encryption without additional setup
- AsyncStorage is unencrypted, security risk

### Limitations
- Key/value size limit (2KB on Android)
- No web support (but not needed for mobile-only features)

### Implementation Notes
```typescript
import * as SecureStore from 'expo-secure-store';

const ACCESS_TOKEN_KEY = 'accessToken';
const REFRESH_TOKEN_KEY = 'refreshToken';

export async function setTokens(access: string, refresh: string) {
    await SecureStore.setItemAsync(ACCESS_TOKEN_KEY, access);
    await SecureStore.setItemAsync(REFRESH_TOKEN_KEY, refresh);
}

export async function clearTokens() {
    await SecureStore.deleteItemAsync(ACCESS_TOKEN_KEY);
    await SecureStore.deleteItemAsync(REFRESH_TOKEN_KEY);
}
```

---

## Summary of Decisions

| Topic | Decision |
|-------|----------|
| Cookie strategy | httpOnly + SameSite=None for Frontend, Bearer for Mobile |
| Token lifetimes | Access: 15min, Refresh: 7 days with rotation |
| Handoff codes | In-memory, crypto-random, 60s expiry, one-time use |
| CORS | Explicit allow list, credentials enabled |
| Rate limiting | Token bucket, IP/User scoped |
| Deep links | Custom scheme + Universal/App Links |
| Frontend storage | Memory for access, httpOnly cookie for refresh |
| Mobile storage | expo-secure-store for all tokens |
