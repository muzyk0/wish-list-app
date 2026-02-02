# Mobile App Completion Plan

**Generated**: 2026-02-02 (Updated)
**Source**: PR #8 Issues, Cross-Domain Architecture
**Status**: Phase 1 In Progress
**Depends On**: [00-cross-domain-architecture-plan.md](./00-cross-domain-architecture-plan.md)

## Architecture Context

```
Mobile App (Expo/React Native)
├── Personal Cabinet (authenticated)
│   ├── Create wishlists
│   ├── Manage holidays/events
│   ├── Add gift items
│   └── View received reservations
├── Auth Flow
│   ├── Direct login/register
│   └── Handoff from Frontend (code exchange)
└── Deployment: Vercel (Expo Web) + App Stores (Native)
```

---

## Progress Summary

| Phase | Status | Errors Fixed | Issues Resolved |
|-------|--------|--------------|-----------------|
| Critical API Fixes | ✅ Complete | 15 | 5 (#63-#67) |
| Phase 1: Auth & Type Safety | 🔄 In Progress | 0/18 | 0/5 |
| Phase 2: Essential Features | ⏳ Pending | - | 0/4 |
| Phase 3: PR Issues | ⏳ Pending | - | 0/23 |
| Phase 4: Polish | ⏳ Pending | - | 0/8 |

---

## Phase 1: Authentication & Type Safety (CRITICAL)

**Estimated Effort**: 6 hours
**Goal**: Secure auth flow + zero TypeScript errors

### Task 1.1: Implement Auth Code Exchange (NEW)
**Files**: `mobile/app/_layout.tsx`, `mobile/lib/api/auth.ts`
**Priority**: Critical
**Effort**: 1.5 hours

**Purpose**: Handle redirect from Frontend with auth code.

```typescript
// mobile/lib/api/auth.ts
import * as SecureStore from 'expo-secure-store';
import { API_URL } from './config';

const ACCESS_TOKEN_KEY = 'accessToken';
const REFRESH_TOKEN_KEY = 'refreshToken';

export async function exchangeCodeForTokens(code: string): Promise<boolean> {
  try {
    const response = await fetch(`${API_URL}/auth/exchange`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code }),
    });

    if (!response.ok) {
      console.error('Code exchange failed:', response.status);
      return false;
    }

    const { accessToken, refreshToken } = await response.json();

    await SecureStore.setItemAsync(ACCESS_TOKEN_KEY, accessToken);
    await SecureStore.setItemAsync(REFRESH_TOKEN_KEY, refreshToken);

    return true;
  } catch (error) {
    console.error('Code exchange error:', error);
    return false;
  }
}

export async function getAccessToken(): Promise<string | null> {
  return SecureStore.getItemAsync(ACCESS_TOKEN_KEY);
}

export async function getRefreshToken(): Promise<string | null> {
  return SecureStore.getItemAsync(REFRESH_TOKEN_KEY);
}

export async function clearTokens(): Promise<void> {
  await SecureStore.deleteItemAsync(ACCESS_TOKEN_KEY);
  await SecureStore.deleteItemAsync(REFRESH_TOKEN_KEY);
}

export async function refreshAccessToken(): Promise<string | null> {
  try {
    const refreshToken = await getRefreshToken();
    if (!refreshToken) return null;

    const response = await fetch(`${API_URL}/auth/refresh`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${refreshToken}`,
      },
    });

    if (!response.ok) {
      await clearTokens();
      return null;
    }

    const { accessToken, refreshToken: newRefreshToken } = await response.json();

    await SecureStore.setItemAsync(ACCESS_TOKEN_KEY, accessToken);
    if (newRefreshToken) {
      await SecureStore.setItemAsync(REFRESH_TOKEN_KEY, newRefreshToken);
    }

    return accessToken;
  } catch {
    await clearTokens();
    return null;
  }
}
```

```typescript
// mobile/app/_layout.tsx - Deep Link Handling
import * as Linking from 'expo-linking';
import { useEffect } from 'react';
import { router } from 'expo-router';
import { exchangeCodeForTokens } from '@/lib/api/auth';

function RootLayout() {
  useEffect(() => {
    // Handle cold start deep link
    Linking.getInitialURL().then((url) => {
      if (url) handleDeepLink(url);
    });

    // Handle warm start deep link
    const subscription = Linking.addEventListener('url', ({ url }) => {
      handleDeepLink(url);
    });

    return () => subscription.remove();
  }, []);

  async function handleDeepLink(url: string) {
    const { path, queryParams } = Linking.parse(url);

    // Handle auth redirect from Frontend
    if (path === 'auth' && queryParams?.code) {
      const success = await exchangeCodeForTokens(queryParams.code as string);

      if (success) {
        // Navigate to redirect path or home
        const redirectPath = queryParams.redirect as string || '/(tabs)';
        router.replace(redirectPath);
      } else {
        router.replace('/auth/login');
      }
      return;
    }

    // Handle other deep links (public wishlist, etc.)
    if (path?.startsWith('lists/')) {
      const match = path.match(/^lists\/([^\/]+)/);
      if (match?.[1]) {
        router.navigate({
          pathname: '/lists/[id]',
          params: { id: match[1] },
        });
      }
    }
  }

  return (/* ... existing layout ... */);
}
```

---

### Task 1.2: Replace localStorage with SecureStore (#36)
**File**: `mobile/lib/api/api.ts`
**Priority**: Critical
**Effort**: 45 minutes

**Problem**: React Native doesn't have localStorage.

```typescript
// mobile/lib/api/api.ts
import { getAccessToken, refreshAccessToken, clearTokens } from './auth';

class ApiClient {
  private baseUrl: string;

  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }

  async request<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
    const token = await getAccessToken();

    const response = await fetch(`${this.baseUrl}${endpoint}`, {
      ...options,
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...options.headers,
      },
    });

    // Handle 401 - try refresh
    if (response.status === 401) {
      const newToken = await refreshAccessToken();
      if (newToken) {
        return this.request(endpoint, options);
      }
      throw new Error('Authentication required');
    }

    if (!response.ok) {
      const error = await response.json().catch(() => ({ error: 'Request failed' }));
      throw new Error(error.error || error.message || 'Request failed');
    }

    return response.json();
  }

  async get<T>(endpoint: string, options?: RequestInit): Promise<T> {
    return this.request(endpoint, { ...options, method: 'GET' });
  }

  async post<T>(endpoint: string, body?: unknown, options?: RequestInit): Promise<T> {
    return this.request(endpoint, {
      ...options,
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  async put<T>(endpoint: string, body?: unknown, options?: RequestInit): Promise<T> {
    return this.request(endpoint, {
      ...options,
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  async delete<T>(endpoint: string, options?: RequestInit): Promise<T> {
    return this.request(endpoint, { ...options, method: 'DELETE' });
  }
}

export const apiClient = new ApiClient(process.env.EXPO_PUBLIC_API_URL || '');
```

---

### Task 1.3: Store Auth Token After Login (#26)
**File**: `mobile/app/auth/login.tsx`
**Priority**: Critical
**Effort**: 20 minutes

```typescript
import * as SecureStore from 'expo-secure-store';
import { router } from 'expo-router';
import { useMutation, useQueryClient } from '@tanstack/react-query';

const loginMutation = useMutation({
  mutationFn: async ({ email, password }: { email: string; password: string }) => {
    const response = await fetch(`${API_URL}/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });

    if (!response.ok) {
      const error = await response.json();
      throw new Error(error.error || 'Login failed');
    }

    return response.json();
  },
  onSuccess: async (data) => {
    // Store tokens
    await SecureStore.setItemAsync('accessToken', data.accessToken);
    if (data.refreshToken) {
      await SecureStore.setItemAsync('refreshToken', data.refreshToken);
    }

    // Invalidate queries
    queryClient.invalidateQueries({ queryKey: ['profile'] });

    // Navigate to main app
    router.replace('/(tabs)');
  },
  onError: (error) => {
    Alert.alert('Login Failed', error.message);
  },
});
```

---

### Task 1.4: Clear Session After Account Deletion (#24)
**File**: `mobile/app/(tabs)/profile.tsx`
**Priority**: Critical
**Effort**: 15 minutes

```typescript
import { clearTokens } from '@/lib/api/auth';

const deleteMutation = useMutation({
  mutationFn: () => apiClient.delete('/account'),
  onSuccess: async () => {
    // Clear all tokens
    await clearTokens();

    // Clear all cached data
    queryClient.clear();

    // Navigate to auth
    router.replace('/auth/login');
  },
  onError: (error) => {
    Alert.alert('Error', 'Failed to delete account. Please try again.');
  },
});
```

---

### Task 1.5: Fix Type Narrowing for Optional Fields (#68)
**Files**: Multiple screens and components
**Priority**: High
**Effort**: 1 hour

**Pattern**:
```typescript
// Before (error)
<Text>{item.view_count} views</Text>

// After (fixed)
<Text>{item.view_count ?? 0} views</Text>

// Conditional rendering
{item.description != null && <Text>{item.description}</Text>}
```

**Files to Update**:
```bash
grep -rn "\.view_count\|\.email\|\.priority\|\.avatar_url" mobile/app/ mobile/components/
```

---

### Task 1.6: Add Missing Public Types (#69)
**File**: `mobile/lib/api/types.ts`
**Priority**: High
**Effort**: 15 minutes

```typescript
// mobile/lib/api/types.ts
import { components } from './generated/schema';

// Public types
export type PublicWishList = components['schemas']['wish-list_internal_services.WishListOutput'];
export type PublicGiftItem = components['schemas']['wish-list_internal_services.GiftItemOutput'];

// Type guards
export function isPublicWishList(list: unknown): list is PublicWishList {
  return typeof list === 'object' && list !== null && 'public_slug' in list;
}
```

---

## Phase 2: Essential Features

**Estimated Effort**: 8 hours
**Goal**: Core app functionality complete

### Task 2.1: Create Gift Item Create Screen (#73)
**Path**: `mobile/app/gift-items/create.tsx`
**Priority**: High
**Effort**: 2 hours

```typescript
import { useLocalSearchParams, router } from 'expo-router';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { GiftItemForm } from '@/components/wish-list/GiftItemForm';
import { apiClient } from '@/lib/api/api';

export default function CreateGiftItemScreen() {
  const { wishlistId } = useLocalSearchParams<{ wishlistId: string }>();
  const queryClient = useQueryClient();

  const createMutation = useMutation({
    mutationFn: (data: CreateGiftItemRequest) =>
      apiClient.post(`/wishlists/${wishlistId}/gift-items`, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['wishlist', wishlistId] });
      router.back();
    },
  });

  return (
    <GiftItemForm
      onSubmit={(data) => createMutation.mutate(data)}
      isLoading={createMutation.isPending}
    />
  );
}
```

---

### Task 2.2: Implement Reservation Details Screen (#74)
**Path**: `mobile/app/reservations/[id]/index.tsx`
**Priority**: Medium
**Effort**: 2 hours

**Features**:
- Gift item information
- Wishlist owner details
- Reservation date and status
- Cancel reservation button

---

### Task 2.3: Add Image Upload Functionality (#75)
**Files**: `mobile/components/wish-list/ImageUpload.tsx`, `mobile/lib/api/api.ts`
**Priority**: High
**Effort**: 2 hours

```typescript
// mobile/lib/api/api.ts - Add upload method
async uploadImage(uri: string): Promise<{ url: string }> {
  const token = await getAccessToken();

  const formData = new FormData();
  const filename = uri.split('/').pop() || 'image.jpg';
  const match = /\.(\w+)$/.exec(filename);
  const type = match ? `image/${match[1]}` : 'image/jpeg';

  formData.append('file', {
    uri,
    name: filename,
    type,
  } as any);

  const response = await fetch(`${this.baseUrl}/s3/upload`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${token}`,
    },
    body: formData,
  });

  if (!response.ok) {
    throw new Error('Image upload failed');
  }

  return response.json();
}
```

---

### Task 2.4: Fix or Remove Template Functionality (#76)
**Files**: `mobile/components/wish-list/TemplateSelector.tsx`
**Priority**: Medium
**Effort**: 1 hour

**Decision**: Check if backend supports templates. If not, remove component.

---

## Phase 3: PR Issues Resolution

**Estimated Effort**: 12 hours

### Critical Issues (Covered in Phase 1)
- #36: expo-secure-store ✅
- #26: Auth token storage ✅
- #24: Session cleanup ✅

### High Priority Issues

| Issue | Description | File | Effort |
|-------|-------------|------|--------|
| #10 | deleteAccount implementation | profile.tsx | 30m |
| #11 | Register response type | register.tsx | 15m |
| #1 | Email persistence | login.tsx | 20m |
| #3 | FlatList data binding | lists | 15m |
| #4 | Reservation status display | ReservationCard | 20m |
| #5 | Typed uploadImage method | api.ts | 30m |
| #23 | Deep-link extraction | _layout.tsx | 30m |
| #28 | Reserve button interactive | ReserveButton | 15m |

### Medium Priority Issues

| Issue | Description | File | Effort |
|-------|-------------|------|--------|
| #2 | Zero price display | GiftItemCard | 10m |
| #6 | Console statements | Multiple | 30m |
| #8 | imageUri sync | ImageUpload | 15m |
| #25 | Avatar rendering | Profile | 15m |
| #27 | Linking config | app.json | 15m |

---

## Phase 4: Polish & UX

**Estimated Effort**: 16 hours

| Task | Description | Effort |
|------|-------------|--------|
| #81 | Search/discover functionality | 4h |
| #82 | Settings screen | 2h |
| #83 | Onboarding flow | 4h |
| #84 | Error handling screens | 2h |
| #85 | Loading and empty states | 4h |

---

## App Configuration for Cross-Domain

### Task: Update app.json
**File**: `mobile/app.json`

```json
{
  "expo": {
    "name": "Wish List",
    "slug": "wishlist",
    "scheme": "wishlistapp",
    "extra": {
      "apiUrl": "https://api.wishlist.com"
    },
    "ios": {
      "bundleIdentifier": "com.wishlist.app",
      "associatedDomains": [
        "applinks:wishlist.com",
        "applinks:www.wishlist.com"
      ]
    },
    "android": {
      "package": "com.wishlist.app",
      "intentFilters": [
        {
          "action": "VIEW",
          "autoVerify": true,
          "data": [
            { "scheme": "https", "host": "wishlist.com", "pathPrefix": "/app" },
            { "scheme": "https", "host": "www.wishlist.com", "pathPrefix": "/app" },
            { "scheme": "wishlistapp" }
          ],
          "category": ["BROWSABLE", "DEFAULT"]
        }
      ]
    }
  }
}
```

---

## Verification Commands

```bash
cd mobile

# 1. Type checking
npm run type-check

# 2. Lint
npm run lint

# 3. Build check (EAS)
npx eas build --platform ios --profile preview --local

# 4. Run on simulator
npx expo start --ios

# 5. Test deep link
xcrun simctl openurl booted "wishlistapp://auth?code=test123"
```

---

## Success Criteria

### Phase 1 Complete When:
- [ ] Auth code exchange works from Frontend redirect
- [ ] 0 TypeScript errors
- [ ] SecureStore used for all tokens
- [ ] Auth flow works end-to-end

### Phase 2 Complete When:
- [ ] Gift item CRUD works
- [ ] Reservations viewable/cancellable
- [ ] Image upload functional

### Phase 3 Complete When:
- [ ] All PR #8 critical/high issues resolved

### Phase 4 Complete When:
- [ ] App is production-ready
- [ ] Good UX with loading/error states
- [ ] Onboarding implemented

---

## Dependencies

### External Dependencies
- Backend: `/auth/exchange` endpoint (Phase 1)
- Backend: `/auth/refresh` endpoint (Phase 1)
- Backend: CORS configured for Expo web (if used)
- Apple: App Site Association file
- Google: Asset Links file
