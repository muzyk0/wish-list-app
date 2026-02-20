# Quick Start: Cross-Domain Authentication

**Feature**: 002-cross-domain-implementation
**Date**: 2026-02-02

## Prerequisites

- Go 1.21+ installed
- Node.js 18+ and pnpm installed
- Expo CLI installed (`npm install -g expo-cli`)
- Docker for local PostgreSQL
- Environment variables configured

## Environment Setup

### Backend (.env)

```bash
# Database
DATABASE_URL=postgresql://postgres:password@localhost:5432/wishlist?sslmode=disable

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-at-least-32-chars
JWT_ACCESS_TOKEN_EXPIRY=15m
JWT_REFRESH_TOKEN_EXPIRY=168h  # 7 days

# CORS (comma-separated origins)
CORS_ALLOWED_ORIGINS=http://localhost:3000,http://localhost:8081,http://localhost:19006

# Environment
ENV=development
```

### Frontend (.env.local)

```bash
# API URL
NEXT_PUBLIC_API_URL=http://localhost:8080/api

# Mobile handoff
NEXT_PUBLIC_MOBILE_SCHEME=wishlistapp
NEXT_PUBLIC_MOBILE_UNIVERSAL_LINK=https://wishlist.com/app
```

### Mobile (app.config.js or .env)

```bash
# API URL
EXPO_PUBLIC_API_URL=http://localhost:8080/api
```

---

## Running Locally

### 1. Start Database

```bash
make db-up
# or
docker-compose -f database/docker-compose.yml up -d
```

### 2. Start Backend

```bash
cd backend
go run ./cmd/server
# Server starts at http://localhost:8080
```

### 3. Start Frontend

```bash
cd frontend
pnpm install
pnpm dev
# Frontend starts at http://localhost:3000
```

### 4. Start Mobile

```bash
cd mobile
pnpm install
npx expo start
# Scan QR code or press 'i' for iOS simulator
```

---

## Testing Authentication Flows

### 1. Login Flow (Frontend)

```bash
# Register a test user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","first_name":"Test"}'

# Login - note the Set-Cookie header
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}' \
  -c cookies.txt -v
```

### 2. Token Refresh Flow

```bash
# Using cookie from login
curl -X POST http://localhost:8080/api/auth/refresh \
  -b cookies.txt -c cookies.txt

# Using Bearer token (mobile-style)
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Authorization: Bearer <refresh_token>"
```

### 3. Mobile Handoff Flow

```bash
# Step 1: Generate handoff code (requires access token)
curl -X POST http://localhost:8080/api/auth/mobile-handoff \
  -H "Authorization: Bearer <access_token>"
# Returns: {"code":"abc123...","expiresIn":60}

# Step 2: Exchange code for tokens (within 60 seconds)
curl -X POST http://localhost:8080/api/auth/exchange \
  -H "Content-Type: application/json" \
  -d '{"code":"abc123..."}'
# Returns: {"accessToken":"...","refreshToken":"...","user":{...}}
```

### 4. CORS Validation

```bash
# Test preflight request
curl -X OPTIONS http://localhost:8080/api/auth/login \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: POST" \
  -v

# Should return:
# Access-Control-Allow-Origin: http://localhost:3000
# Access-Control-Allow-Credentials: true
```

### 5. Health Check

```bash
curl http://localhost:8080/healthz
# Returns: {"status":"healthy"}
```

---

## Deep Link Testing

### iOS Simulator

```bash
xcrun simctl openurl booted "wishlistapp://auth?code=test123"
```

### Android Emulator

```bash
adb shell am start -W -a android.intent.action.VIEW \
  -d "wishlistapp://auth?code=test123"
```

---

## Code Examples

### Frontend: Auth Manager

```typescript
// lib/auth.ts
class AuthManager {
  private accessToken: string | null = null;

  getAccessToken(): string | null {
    return this.accessToken;
  }

  setAccessToken(token: string | null): void {
    this.accessToken = token;
  }

  async refreshAccessToken(): Promise<string | null> {
    const response = await fetch('/api/auth/refresh', {
      method: 'POST',
      credentials: 'include', // Send cookies
    });

    if (!response.ok) {
      this.accessToken = null;
      return null;
    }

    const { accessToken } = await response.json();
    this.accessToken = accessToken;
    return accessToken;
  }
}

export const authManager = new AuthManager();
```

### Mobile: SecureStore Usage

```typescript
// lib/api/auth.ts
import * as SecureStore from 'expo-secure-store';

const ACCESS_TOKEN_KEY = 'accessToken';
const REFRESH_TOKEN_KEY = 'refreshToken';

export async function getAccessToken(): Promise<string | null> {
  return SecureStore.getItemAsync(ACCESS_TOKEN_KEY);
}

export async function setTokens(access: string, refresh: string): Promise<void> {
  await SecureStore.setItemAsync(ACCESS_TOKEN_KEY, access);
  await SecureStore.setItemAsync(REFRESH_TOKEN_KEY, refresh);
}

export async function clearTokens(): Promise<void> {
  await SecureStore.deleteItemAsync(ACCESS_TOKEN_KEY);
  await SecureStore.deleteItemAsync(REFRESH_TOKEN_KEY);
}
```

### Backend: Code Store

```go
// internal/auth/code_store.go
type CodeStore struct {
    mu    sync.RWMutex
    codes map[string]codeEntry
}

func (cs *CodeStore) Set(code string, userID uuid.UUID, ttl time.Duration) {
    cs.mu.Lock()
    defer cs.mu.Unlock()
    cs.codes[code] = codeEntry{
        UserID:    userID,
        ExpiresAt: time.Now().Add(ttl),
    }
}

func (cs *CodeStore) GetAndDelete(code string) (uuid.UUID, bool) {
    cs.mu.Lock()
    defer cs.mu.Unlock()

    entry, ok := cs.codes[code]
    if !ok || time.Now().After(entry.ExpiresAt) {
        delete(cs.codes, code)
        return uuid.Nil, false
    }

    delete(cs.codes, code)
    return entry.UserID, true
}
```

---

## Troubleshooting

### CORS Errors

**Symptom**: Browser console shows "CORS policy" errors

**Solution**:
1. Check `CORS_ALLOWED_ORIGINS` includes your frontend URL
2. Ensure `Access-Control-Allow-Credentials: true` is set
3. For development, include `http://localhost:3000`

### Cookie Not Sent

**Symptom**: Refresh token cookie not included in requests

**Solution**:
1. Ensure `credentials: 'include'` in fetch options
2. Check cookie `SameSite=None; Secure` attributes
3. For local development, may need HTTPS

### Deep Link Not Working

**Symptom**: App doesn't open when clicking Universal Link

**Solution**:
1. Check `associatedDomains` in app.json
2. Verify AASA file at `/.well-known/apple-app-site-association`
3. For testing, use custom scheme: `wishlistapp://`

### Token Refresh Loop

**Symptom**: Continuous 401 errors and refresh attempts

**Solution**:
1. Check `refreshPromise` singleton pattern implemented
2. Verify refresh token hasn't expired (7 day limit)
3. Ensure proper error handling clears tokens on failure

---

## Next Steps

1. Run `make test` to verify all tests pass
2. Review [contracts/auth-api.yaml](./contracts/auth-api.yaml) for API documentation
3. See [research.md](./research.md) for technical decisions
4. Run `/speckit.tasks` to generate implementation tasks
