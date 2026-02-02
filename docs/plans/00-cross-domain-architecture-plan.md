# Cross-Domain Architecture Plan

**Generated**: 2026-02-02
**Priority**: Critical - Must be implemented before other plans
**Status**: Ready for Implementation

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           WISH LIST APPLICATION                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────┐    ┌─────────────────────┐    ┌─────────────────┐ │
│  │   Frontend (Web)    │    │   Mobile (App)      │    │   Backend       │ │
│  │   Next.js           │    │   React Native/Expo │    │   Go/Echo       │ │
│  │                     │    │                     │    │                 │ │
│  │   Domain: TBD       │    │   Scheme:           │    │   Domain: TBD   │ │
│  │   Host: Vercel      │    │   wishlistapp://    │    │   Host: Render  │ │
│  │                     │    │   Host: Vercel      │    │                 │ │
│  ├─────────────────────┤    ├─────────────────────┤    ├─────────────────┤ │
│  │ FEATURES:           │    │ FEATURES:           │    │ ENDPOINTS:      │ │
│  │ • Public wishlist   │    │ • Create wishlists  │    │ • /auth/*       │ │
│  │   view              │    │ • Manage holidays   │    │ • /wishlists/*  │ │
│  │ • Guest reservation │    │ • Add gift items    │    │ • /gift-items/* │ │
│  │ • Auth reservation  │    │ • View reservations │    │ • /reservations │ │
│  │ • My reservations   │    │ • Profile settings  │    │ • /public/*     │ │
│  │ • Cancel booking    │    │ • Account deletion  │    │ • /s3/*         │ │
│  │ • Redirect to LC    │    │                     │    │                 │ │
│  └─────────────────────┘    └─────────────────────┘    └─────────────────┘ │
│           │                          │                          ▲          │
│           │                          │                          │          │
│           └──────────────────────────┴──────────────────────────┘          │
│                              HTTPS + JWT + CORS                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Domain Configuration

### Production Domains (Example)

| Component | Domain | Provider |
|-----------|--------|----------|
| Frontend | `wishlist.com` | Vercel |
| Mobile (Web) | N/A (native app) | App Stores |
| Backend | `api.wishlist.com` | Render |

### Environment Variables

**Backend (.env)**:
```bash
# CORS Origins - comma separated
CORS_ALLOWED_ORIGINS=https://wishlist.com,https://www.wishlist.com

# JWT Settings
JWT_SECRET=<strong-secret>
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=7d

# Deep Link Scheme for Mobile
MOBILE_DEEP_LINK_SCHEME=wishlistapp
```

**Frontend (.env.local)**:
```bash
# API Base URL
NEXT_PUBLIC_API_URL=https://api.wishlist.com

# Mobile Deep Link (for redirect to personal cabinet)
NEXT_PUBLIC_MOBILE_SCHEME=wishlistapp
NEXT_PUBLIC_MOBILE_UNIVERSAL_LINK=https://wishlist.com/app
```

**Mobile (app.json / .env)**:
```json
{
  "expo": {
    "scheme": "wishlistapp",
    "extra": {
      "apiUrl": "https://api.wishlist.com"
    }
  }
}
```

---

## Authentication Architecture

### Token Strategy

**Access Token**:
- Short-lived (15 minutes)
- Stored in memory (Frontend) / SecureStore (Mobile)
- Sent via `Authorization: Bearer <token>` header

**Refresh Token**:
- Long-lived (7 days)
- Stored in httpOnly cookie (Frontend) / SecureStore (Mobile)
- Used only to get new access tokens

### Token Flow Diagram

```
┌──────────────────────────────────────────────────────────────────────────┐
│                          AUTHENTICATION FLOW                              │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                           │
│  1. LOGIN (Frontend or Mobile)                                           │
│  ┌──────────┐         POST /auth/login          ┌──────────┐             │
│  │ Client   │ ──────────────────────────────────▶│ Backend  │             │
│  │          │ { email, password }                │          │             │
│  │          │ ◀────────────────────────────────── │          │             │
│  └──────────┘ { accessToken, user }              └──────────┘             │
│               + Set-Cookie: refreshToken (httpOnly)                       │
│                                                                           │
│  2. API REQUEST                                                           │
│  ┌──────────┐   Authorization: Bearer <accessToken>  ┌──────────┐        │
│  │ Client   │ ──────────────────────────────────────▶│ Backend  │        │
│  │          │ ◀────────────────────────────────────── │          │        │
│  └──────────┘            { data }                    └──────────┘        │
│                                                                           │
│  3. TOKEN REFRESH (when access token expires)                             │
│  ┌──────────┐       POST /auth/refresh              ┌──────────┐         │
│  │ Client   │ ──────────────────────────────────────▶│ Backend  │         │
│  │          │ Cookie: refreshToken                   │          │         │
│  │          │ ◀────────────────────────────────────── │          │         │
│  └──────────┘ { accessToken }                        └──────────┘         │
│               + Set-Cookie: refreshToken (new, httpOnly)                  │
│                                                                           │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## Frontend ↔ Mobile Redirect (Personal Cabinet)

### OAuth-Style Flow

When user wants to access "Personal Cabinet" from Frontend:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                    FRONTEND → MOBILE REDIRECT FLOW                       │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  1. User clicks "Personal Cabinet" on Frontend                          │
│                                                                          │
│  2. Frontend requests auth code from Backend                             │
│     POST /auth/mobile-handoff                                            │
│     Authorization: Bearer <accessToken>                                  │
│     Response: { code: "abc123", expiresIn: 60 }                         │
│                                                                          │
│  3. Frontend redirects to Mobile via Universal Link                      │
│     https://wishlist.com/app/auth?code=abc123                           │
│     OR Deep Link: wishlistapp://auth?code=abc123                        │
│                                                                          │
│  4. Mobile exchanges code for tokens                                     │
│     POST /auth/exchange                                                  │
│     Body: { code: "abc123" }                                            │
│     Response: { accessToken, refreshToken, user }                       │
│                                                                          │
│  5. Mobile stores tokens in SecureStore                                  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

### Implementation

**Backend - New Endpoints**:

```go
// POST /auth/mobile-handoff
// Generates short-lived code for Frontend → Mobile handoff
func (h *AuthHandler) MobileHandoff(c echo.Context) error {
    userID := getUserIDFromToken(c)

    // Generate random code
    code := generateSecureCode()

    // Store code with expiry (60 seconds)
    h.codeStore.Set(code, userID, 60*time.Second)

    return c.JSON(http.StatusOK, map[string]interface{}{
        "code":      code,
        "expiresIn": 60,
    })
}

// POST /auth/exchange
// Exchanges handoff code for tokens
func (h *AuthHandler) ExchangeCode(c echo.Context) error {
    var req struct {
        Code string `json:"code" validate:"required"`
    }

    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
    }

    // Get user ID from code
    userID, ok := h.codeStore.Get(req.Code)
    if !ok {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "Invalid or expired code"})
    }

    // Delete code (one-time use)
    h.codeStore.Delete(req.Code)

    // Generate tokens
    accessToken, refreshToken, err := h.authService.GenerateTokens(userID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to generate tokens"})
    }

    return c.JSON(http.StatusOK, map[string]interface{}{
        "accessToken":  accessToken,
        "refreshToken": refreshToken,
        "user":         user,
    })
}
```

**Frontend - Redirect to Mobile**:

```typescript
// lib/mobile-handoff.ts
export async function redirectToPersonalCabinet() {
  const apiClient = getApiClient();

  // Get handoff code
  const { code } = await apiClient.post('/auth/mobile-handoff');

  // Try Universal Link first, fall back to App Store
  const universalLink = `${process.env.NEXT_PUBLIC_MOBILE_UNIVERSAL_LINK}/auth?code=${code}`;
  const deepLink = `${process.env.NEXT_PUBLIC_MOBILE_SCHEME}://auth?code=${code}`;
  const appStoreLink = 'https://apps.apple.com/app/wishlist/id123456789';

  // Attempt to open app
  window.location.href = universalLink;

  // If app not installed, redirect to app store after timeout
  setTimeout(() => {
    // Check if page is still visible (app didn't open)
    if (!document.hidden) {
      window.location.href = appStoreLink;
    }
  }, 2500);
}
```

**Mobile - Handle Redirect**:

```typescript
// app/_layout.tsx
import * as Linking from 'expo-linking';
import * as SecureStore from 'expo-secure-store';

function handleDeepLink(url: string) {
  const { path, queryParams } = Linking.parse(url);

  if (path === 'auth' && queryParams?.code) {
    // Exchange code for tokens
    exchangeCodeForTokens(queryParams.code as string);
  }
}

async function exchangeCodeForTokens(code: string) {
  const response = await fetch(`${API_URL}/auth/exchange`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ code }),
  });

  if (response.ok) {
    const { accessToken, refreshToken, user } = await response.json();

    // Store tokens
    await SecureStore.setItemAsync('accessToken', accessToken);
    await SecureStore.setItemAsync('refreshToken', refreshToken);

    // Navigate to home
    router.replace('/(tabs)');
  } else {
    // Show error, redirect to login
    router.replace('/auth/login');
  }
}
```

---

## CORS Configuration

### Backend CORS Middleware

```go
// middleware/cors.go
func CORSMiddleware(allowedOrigins []string) echo.MiddlewareFunc {
    return middleware.CORSWithConfig(middleware.CORSConfig{
        AllowOrigins:     allowedOrigins,
        AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions},
        AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
        AllowCredentials: true, // Required for cookies
        MaxAge:           86400,
    })
}

// main.go
func main() {
    // ...

    origins := strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), ",")
    e.Use(CORSMiddleware(origins))

    // ...
}
```

---

## Guest Reservation Flow

Guests can reserve items without authentication:

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        GUEST RESERVATION FLOW                            │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  1. Guest views public wishlist                                          │
│     GET /public/wishlists/{slug}                                         │
│     (No authentication required)                                         │
│                                                                          │
│  2. Guest reserves item                                                  │
│     POST /public/wishlists/{slug}/gift-items/{id}/reserve                │
│     Body: { guestName: "John", guestEmail: "john@example.com" }         │
│     Response: { reservationId, guestToken }                             │
│                                                                          │
│  3. Guest receives email with management link                            │
│     Link: https://wishlist.com/reservations/manage?token={guestToken}   │
│                                                                          │
│  4. Guest can cancel reservation using token                             │
│     DELETE /public/reservations/{id}?token={guestToken}                  │
│                                                                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

## Security Considerations

### Token Security

| Concern | Mitigation |
|---------|------------|
| XSS stealing tokens | Access token in memory only, refresh in httpOnly cookie |
| CSRF attacks | SameSite=Lax cookie, CORS restrictions |
| Token leakage in URLs | Handoff codes are one-time use, expire in 60s |
| Cross-domain cookie issues | Use token refresh endpoint with credentials |

### CORS Security

| Origin | Allowed |
|--------|---------|
| `https://wishlist.com` | ✅ Yes |
| `https://www.wishlist.com` | ✅ Yes |
| `http://localhost:3000` | ✅ Dev only |
| `https://malicious.com` | ❌ No |

### Rate Limiting

| Endpoint | Limit |
|----------|-------|
| `/auth/login` | 5/minute per IP |
| `/auth/mobile-handoff` | 10/minute per user |
| `/auth/exchange` | 10/minute per IP |
| `/public/*/reserve` | 10/minute per IP |

---

## Implementation Tasks

### Phase 1: Backend Auth Updates (4 hours)

- [ ] **1.1** Add refresh token support to login endpoint
- [ ] **1.2** Create `/auth/refresh` endpoint
- [ ] **1.3** Create `/auth/mobile-handoff` endpoint
- [ ] **1.4** Create `/auth/exchange` endpoint
- [ ] **1.5** Configure CORS middleware for multiple origins
- [ ] **1.6** Add rate limiting middleware

### Phase 2: Frontend Token Management (3 hours)

- [ ] **2.1** Implement in-memory access token storage
- [ ] **2.2** Add automatic token refresh on 401
- [ ] **2.3** Create mobile handoff redirect function
- [ ] **2.4** Update API client to handle token refresh

### Phase 3: Mobile Auth Updates (2 hours)

- [ ] **3.1** Handle auth deep links in `_layout.tsx`
- [ ] **3.2** Implement code exchange flow
- [ ] **3.3** Update SecureStore token management
- [ ] **3.4** Add token refresh logic

### Phase 4: Testing (2 hours)

- [ ] **4.1** Test Frontend → Mobile handoff flow
- [ ] **4.2** Test token refresh across all clients
- [ ] **4.3** Test CORS configuration
- [ ] **4.4** Test guest reservation flow

---

## Verification Commands

```bash
# Backend - Test CORS
curl -I -X OPTIONS https://api.wishlist.com/auth/login \
  -H "Origin: https://wishlist.com" \
  -H "Access-Control-Request-Method: POST"

# Should return:
# Access-Control-Allow-Origin: https://wishlist.com
# Access-Control-Allow-Credentials: true

# Test token refresh
curl -X POST https://api.wishlist.com/auth/refresh \
  -H "Cookie: refreshToken=<token>" \
  --include

# Test handoff code generation
curl -X POST https://api.wishlist.com/auth/mobile-handoff \
  -H "Authorization: Bearer <accessToken>"
```

---

## Dependencies

This plan MUST be completed before:
- Plan 01 (Frontend Security) - Token storage strategy depends on this
- Plan 02 (Mobile) - Deep link auth handling depends on this
- Plan 03 (Backend) - CORS and new endpoints depend on this

---

## Notes

- Universal Links require Apple App Site Association file hosted at `/.well-known/apple-app-site-association`
- Android App Links require Digital Asset Links file at `/.well-known/assetlinks.json`
- Consider adding magic link login as alternative to password for better cross-device UX
