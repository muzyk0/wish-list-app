# Frontend Security and Quality Improvement Plan

**Generated**: 2026-02-02 (Updated)
**Source**: PR #7 Issues, Cross-Domain Architecture
**Status**: Ready for Implementation
**Depends On**: [00-cross-domain-architecture-plan.md](./00-cross-domain-architecture-plan.md)

## Executive Summary

This plan addresses 17 open frontend issues from PR #7 code review, updated for cross-domain architecture with Backend on Render.

## Architecture Context

```
Frontend (Next.js/Vercel) ──── HTTPS + JWT ────▶ Backend (Go/Render)
       │
       └── Deep Link/Universal Link ────▶ Mobile (Expo/Vercel)
```

**Key Constraints**:
- Different domains = no shared httpOnly cookies for auth
- Must use Bearer token authentication
- Refresh tokens via API, not cookies

---

## Phase 1: Token Management (CRITICAL)

**Estimated Effort**: 2 hours
**Risk Level**: Critical - Security + Cross-domain auth

### Task 1.1: Remove JWT_SECRET from Client Bundle (#41)
**File**: `frontend/.env.example`
**Priority**: Critical
**Effort**: 10 minutes

**Problem**: JWT_SECRET exposed as `NEXT_PUBLIC_JWT_SECRET` bundles it in client-side code.

**Implementation**:
```bash
# 1. Update .env.example
# Remove: NEXT_PUBLIC_JWT_SECRET=your-super-secret-jwt-key-here

# 2. Add API URL instead
NEXT_PUBLIC_API_URL=https://api.wishlist.com

# 3. Search and remove any NEXT_PUBLIC_JWT_SECRET references
grep -r "NEXT_PUBLIC_JWT_SECRET" frontend/src/
```

**Note**: Frontend should NEVER handle JWT signing/verification. Backend handles all auth.

---

### Task 1.2: Implement Secure Token Storage (#42)
**Files**: `frontend/src/lib/api.ts`, `frontend/src/lib/auth.ts`
**Priority**: Critical
**Effort**: 1.5 hours

**Problem**: JWTs in localStorage are vulnerable to XSS attacks.

**Solution**: In-memory access token + refresh token flow via API.

**Implementation**:

```typescript
// frontend/src/lib/auth.ts
class AuthManager {
  private accessToken: string | null = null;
  private refreshPromise: Promise<string | null> | null = null;

  setAccessToken(token: string | null) {
    this.accessToken = token;
  }

  getAccessToken(): string | null {
    return this.accessToken;
  }

  async refreshAccessToken(): Promise<string | null> {
    // Prevent concurrent refresh requests
    if (this.refreshPromise) {
      return this.refreshPromise;
    }

    this.refreshPromise = this.doRefresh();
    const result = await this.refreshPromise;
    this.refreshPromise = null;
    return result;
  }

  private async doRefresh(): Promise<string | null> {
    try {
      const response = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/auth/refresh`, {
        method: 'POST',
        credentials: 'include', // Send refresh token cookie
      });

      if (!response.ok) {
        this.accessToken = null;
        return null;
      }

      const { accessToken } = await response.json();
      this.accessToken = accessToken;
      return accessToken;
    } catch {
      this.accessToken = null;
      return null;
    }
  }

  async logout() {
    try {
      await fetch(`${process.env.NEXT_PUBLIC_API_URL}/auth/logout`, {
        method: 'POST',
        credentials: 'include',
      });
    } finally {
      this.accessToken = null;
    }
  }
}

export const authManager = new AuthManager();
```

```typescript
// frontend/src/lib/api.ts
import { authManager } from './auth';

class ApiClient {
  private baseUrl = process.env.NEXT_PUBLIC_API_URL || '';

  async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const token = authManager.getAccessToken();

    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      credentials: 'include', // Always include cookies for refresh token
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...options.headers,
      },
    });

    // Handle 401 - try refresh
    if (response.status === 401) {
      const newToken = await authManager.refreshAccessToken();
      if (newToken) {
        // Retry with new token
        return this.request(endpoint, options);
      }
      // Redirect to login
      window.location.href = '/auth/login';
      throw new Error('Authentication required');
    }

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Request failed' }));
      throw new Error(error.error || 'Request failed');
    }

    return response.json();
  }

  async login(email: string, password: string) {
    const response = await fetch(`${this.baseUrl}/auth/login`, {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Login failed');
    }

    const { accessToken, user } = await response.json();
    authManager.setAccessToken(accessToken);
    return { user };
  }

  async logout() {
    await authManager.logout();
  }
}

export const apiClient = new ApiClient();
```

**Acceptance Criteria**:
- [ ] No localStorage/sessionStorage for tokens
- [ ] Access token stored in memory only
- [ ] Automatic token refresh on 401
- [ ] XSS cannot steal authentication

---

### Task 1.3: Add Mobile Handoff Function
**File**: `frontend/src/lib/mobile-handoff.ts`
**Priority**: High
**Effort**: 30 minutes

**Implementation**:
```typescript
// frontend/src/lib/mobile-handoff.ts
import { apiClient } from './api';

const MOBILE_SCHEME = process.env.NEXT_PUBLIC_MOBILE_SCHEME || 'wishlistapp';
const MOBILE_UNIVERSAL_LINK = process.env.NEXT_PUBLIC_MOBILE_UNIVERSAL_LINK || 'https://wishlist.com/app';
const APP_STORE_URL = 'https://apps.apple.com/app/wishlist/id123456789';
const PLAY_STORE_URL = 'https://play.google.com/store/apps/details?id=com.wishlist.app';

export async function redirectToPersonalCabinet(path: string = '/') {
  try {
    // Get handoff code from backend
    const { code } = await apiClient.post<{ code: string; expiresIn: number }>(
      '/auth/mobile-handoff'
    );

    // Build redirect URL
    const encodedPath = encodeURIComponent(path);
    const universalLink = `${MOBILE_UNIVERSAL_LINK}/auth?code=${code}&redirect=${encodedPath}`;
    const deepLink = `${MOBILE_SCHEME}://auth?code=${code}&redirect=${encodedPath}`;

    // Try Universal Link first
    window.location.href = universalLink;

    // Fallback to app store if app not installed
    setTimeout(() => {
      if (!document.hidden) {
        const isIOS = /iPad|iPhone|iPod/.test(navigator.userAgent);
        window.location.href = isIOS ? APP_STORE_URL : PLAY_STORE_URL;
      }
    }, 2500);
  } catch (error) {
    console.error('Failed to create handoff:', error);
    throw error;
  }
}
```

---

## Phase 2: Dependency Corrections (High Priority)

**Estimated Effort**: 15 minutes
**Risk Level**: High - Production build failures

### Task 2.1: Fix Runtime Dependencies (#55, #56, #57)
**File**: `frontend/package.json`
**Priority**: High
**Effort**: 10 minutes

**Problem**: Runtime packages in devDependencies cause production failures on Vercel.

**Implementation**:
```bash
cd frontend
pnpm remove class-variance-authority clsx @radix-ui/react-slot lucide-react
pnpm add class-variance-authority clsx @radix-ui/react-slot lucide-react
pnpm remove postcss
pnpm add -D postcss
```

**Acceptance Criteria**:
- [ ] class-variance-authority in dependencies
- [ ] clsx in dependencies
- [ ] @radix-ui/react-slot in dependencies
- [ ] lucide-react in dependencies
- [ ] postcss in devDependencies
- [ ] Vercel production build succeeds

---

## Phase 3: Component Fixes (High Priority)

**Estimated Effort**: 30 minutes
**Risk Level**: High - Runtime errors

### Task 3.1: Add 'use client' to GiftItemDisplay (#44)
**File**: `frontend/src/components/wish-list/GiftItemDisplay.tsx`
**Priority**: High
**Effort**: 2 minutes

```typescript
'use client'

// ... rest of file
```

---

### Task 3.2: Fix Button Default Type (#50)
**File**: `frontend/src/components/ui/button.tsx`
**Priority**: High
**Effort**: 10 minutes

```typescript
const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  ({ className, variant, size, asChild = false, type = 'button', ...props }, ref) => {
    const Comp = asChild ? Slot : 'button'
    return (
      <Comp
        className={cn(buttonVariants({ variant, size, className }))}
        ref={ref}
        type={asChild ? undefined : type}
        {...props}
      />
    )
  }
)
```

---

### Task 3.3: Fix useAuthRedirect Cleanup (#46)
**File**: `frontend/src/hooks/useAuthRedirect.ts`
**Priority**: High
**Effort**: 15 minutes

```typescript
import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { apiClient } from '@/lib/api';

export function useAuthRedirect(redirectTo: string = '/auth/login') {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null);
  const router = useRouter();

  useEffect(() => {
    const controller = new AbortController();

    async function checkAuth() {
      try {
        await apiClient.get('/auth/me', { signal: controller.signal });

        if (!controller.signal.aborted) {
          setIsAuthenticated(true);
        }
      } catch (error) {
        if (error instanceof Error && error.name === 'AbortError') {
          return;
        }
        if (!controller.signal.aborted) {
          setIsAuthenticated(false);
          router.push(redirectTo);
        }
      }
    }

    checkAuth();

    return () => {
      controller.abort();
    };
  }, [redirectTo, router]);

  return isAuthenticated;
}
```

---

## Phase 4: Component Improvements (Medium Priority)

**Estimated Effort**: 20 minutes

### Task 4.1: Add forwardRef to Input (#48)
**File**: `frontend/src/components/ui/input.tsx`

```typescript
const Input = React.forwardRef<HTMLInputElement, InputProps>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        type={type}
        className={cn(/* ... */)}
        ref={ref}
        {...props}
      />
    )
  }
)
Input.displayName = 'Input'
```

---

### Task 4.2: Add forwardRef to Textarea (#49)
**File**: `frontend/src/components/ui/textarea.tsx`

```typescript
const Textarea = React.forwardRef<HTMLTextAreaElement, TextareaProps>(
  ({ className, ...props }, ref) => {
    return (
      <textarea
        className={cn(/* ... */)}
        ref={ref}
        {...props}
      />
    )
  }
)
Textarea.displayName = 'Textarea'
```

---

## Phase 5: Configuration Fixes (Low Priority)

### Task 5.1: Fix .gitignore (#58, #59)
**File**: `frontend/.gitignore`

```gitignore
# Remove incorrect line:
# /frontend

# Add env pattern:
.env*.local
```

---

## Phase 6: Vercel Deployment Configuration

### Task 6.1: Add Vercel Configuration
**File**: `frontend/vercel.json`
**Priority**: Medium
**Effort**: 15 minutes

```json
{
  "framework": "nextjs",
  "regions": ["fra1"],
  "env": {
    "NEXT_PUBLIC_API_URL": "@api-url",
    "NEXT_PUBLIC_MOBILE_SCHEME": "wishlistapp",
    "NEXT_PUBLIC_MOBILE_UNIVERSAL_LINK": "@mobile-universal-link"
  },
  "headers": [
    {
      "source": "/(.*)",
      "headers": [
        {
          "key": "X-Content-Type-Options",
          "value": "nosniff"
        },
        {
          "key": "X-Frame-Options",
          "value": "DENY"
        },
        {
          "key": "X-XSS-Protection",
          "value": "1; mode=block"
        }
      ]
    }
  ]
}
```

---

## Implementation Checklist

### Critical (Complete First)
- [ ] Task 1.1: Remove JWT_SECRET from client
- [ ] Task 1.2: Implement secure token storage
- [ ] Task 1.3: Add mobile handoff function

### High Priority (Complete Second)
- [ ] Task 2.1: Dependency corrections
- [ ] Task 3.1: GiftItemDisplay 'use client'
- [ ] Task 3.2: Button default type
- [ ] Task 3.3: useAuthRedirect cleanup

### Medium Priority (Complete Third)
- [ ] Task 4.1: Input forwardRef
- [ ] Task 4.2: Textarea forwardRef
- [ ] Task 6.1: Vercel configuration

### Low Priority (Complete Last)
- [ ] Task 5.1: .gitignore fixes

---

## Verification Commands

```bash
cd frontend

# 1. Type checking
npm run type-check

# 2. Lint
npm run lint

# 3. Build (simulates Vercel)
npm run build

# 4. Verify no secrets in bundle
grep -r "JWT_SECRET" .next/ | wc -l  # Should be 0
grep -r "NEXT_PUBLIC_JWT" .next/ | wc -l  # Should be 0

# 5. Test with Vercel CLI
npx vercel build
```

---

## Notes

- Backend must implement `/auth/refresh` and `/auth/mobile-handoff` endpoints first
- CORS must be configured on Backend to allow Frontend domain
- Test token refresh flow in development before deploying
