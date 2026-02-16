# Mobile Redirection Architecture

**Application**: Wish List Application
**Version**: 1.1.0
**Feature**: Account Access Redirection Mechanism (T087)

## Overview

The Wish List Application follows a separation of concerns between public functionality (frontend) and private account management (mobile app). This document describes the redirection mechanism that ensures users access account-related features through the mobile app or mobile web version at `lk.domain.com`.

## Architecture Principles

### Public Frontend (Next.js at domain.com)
- **Purpose**: Public wishlist viewing and guest reservations
- **Target Audience**: Anyone with a wishlist link (no account required)
- **Key Features**:
  - View public wishlists via `/public/[slug]`
  - Reserve gifts as a guest
  - View guest reservations (tracked by localStorage token)
  - Browse and discover wishlists

### Private Mobile App (React Native + Mobile Web at lk.domain.com)
- **Purpose**: Account management and wishlist creation
- **Target Audience**: Registered users managing their wishlists
- **Key Features**:
  - User registration and login
  - Create and edit wishlists
  - Add and manage gift items
  - View user reservations
  - Account settings and profile management

## Redirection Mechanism

### 1. MobileRedirect Component

**Location**: `frontend/src/components/common/MobileRedirect.tsx`

**Purpose**: Universal component for redirecting users from web to mobile app

**Usage**:
```tsx
import MobileRedirect from '@/components/common/MobileRedirect';

<MobileRedirect
  redirectPath="auth/login"
  fallbackUrl="https://lk.domain.com/auth/login"
>
  <div>Redirecting to mobile app...</div>
</MobileRedirect>
```

**How it Works**:
1. Attempts to open native mobile app using custom URL scheme: `wishlistapp://[path]`
2. Waits 1.5 seconds to detect if the app opened (checks `document.visibilityState`)
3. If app didn't open (page still visible), redirects to mobile web fallback at `lk.domain.com`

**Parameters**:
- `redirectPath` (optional): Specific path in mobile app (e.g., "auth/login", "my/reservations")
- `fallbackUrl` (optional): Mobile web URL if app is not installed (defaults to "https://lk.domain.com")
- `children` (optional): Content to display during redirection

### 2. useAuthRedirect Hook

**Location**: `frontend/src/hooks/useAuthRedirect.ts`

**Purpose**: Utility hook to check authentication status and optionally trigger redirection

**Usage**:
```tsx
import { useAuthRedirect } from '@/hooks/useAuthRedirect';

const { isAuthenticated, isLoading } = useAuthRedirect(true);
```

**Parameters**:
- `shouldRedirect` (boolean): If true, prepares component for redirection of authenticated users

**Returns**:
- `isAuthenticated` (boolean | null): User authentication status
- `isLoading` (boolean): Whether authentication check is in progress

## Protected Routes

### Account-Only Routes (Redirect to Mobile App)

#### 1. `/auth/login` - Login Page
**File**: `frontend/src/app/auth/login/page.tsx`

**Behavior**:
- Automatically redirects all visitors to mobile app
- Shows "Redirecting to mobile app..." message
- Provides manual link to mobile web version
- Deep link: `wishlistapp://auth/login`
- Fallback: `https://lk.domain.com/auth/login`

#### 2. `/auth/register` - Registration Page
**File**: `frontend/src/app/auth/register/page.tsx`

**Behavior**:
- Automatically redirects all visitors to mobile app
- Shows "Redirecting to mobile app..." message
- Provides manual link to mobile web version
- Deep link: `wishlistapp://auth/register`
- Fallback: `https://lk.domain.com/auth/register`

#### 3. `/my/reservations` - User Reservations Page
**File**: `frontend/src/app/my/reservations/page.tsx`

**Behavior**:
- **Authenticated users**: Redirected to mobile app for account management
  - Deep link: `wishlistapp://my/reservations`
  - Fallback: `https://lk.domain.com/my/reservations`
- **Guest users**: Can view their reservations in the frontend
  - Uses localStorage token to fetch guest reservations
  - No account required

### Public Routes (Stay in Frontend)

#### 1. `/` - Home Page
**File**: `frontend/src/app/page.tsx`

**Behavior**:
- Public landing page explaining the app
- Provides links to mobile app for account management
- Lists quick links to public features
- No automatic redirection

#### 2. `/public/[slug]` - Public Wishlist View
**File**: `frontend/src/app/public/[slug]/page.tsx`

**Behavior**:
- Fully public, no authentication required
- Displays wishlist items and reservation status
- Allows guest users to reserve gifts
- Provides optional link to mobile app for creating own wishlists
- No automatic redirection

## Deep Linking Strategy

### URL Scheme: `wishlistapp://[path]`

**Supported Deep Links**:

| Deep Link | Purpose | Fallback URL |
|-----------|---------|--------------|
| `wishlistapp://home` | App home screen | `https://lk.domain.com` |
| `wishlistapp://auth/login` | Login screen | `https://lk.domain.com/auth/login` |
| `wishlistapp://auth/register` | Registration screen | `https://lk.domain.com/auth/register` |
| `wishlistapp://my/reservations` | User's reservations | `https://lk.domain.com/my/reservations` |
| `wishlistapp://reserve?itemId={id}` | Reserve gift item | `https://lk.domain.com/lists/{listId}/reserve/{itemId}` |

### Mobile App Configuration

**iOS (Info.plist)**:
```xml
<key>CFBundleURLTypes</key>
<array>
  <dict>
    <key>CFBundleURLSchemes</key>
    <array>
      <string>wishlistapp</string>
    </array>
  </dict>
</array>
```

**Android (AndroidManifest.xml)**:
```xml
<intent-filter>
  <action android:name="android.intent.action.VIEW" />
  <category android:name="android.intent.category.DEFAULT" />
  <category android:name="android.intent.category.BROWSABLE" />
  <data android:scheme="wishlistapp" />
</intent-filter>
```

## Mobile Web Version (lk.domain.com)

### Purpose
- Fallback for users without the native app installed
- Provides same account management functionality as mobile app
- Responsive mobile-first design
- Progressive Web App (PWA) capabilities

### Technology Stack
- React Native Web (same codebase as mobile app)
- Hosted at `lk.domain.com` subdomain
- Full account management features
- Optimized for mobile browsers

## User Flows

### Flow 1: Guest Viewing and Reserving Gifts
```
1. User receives wishlist link: domain.com/public/birthday-2026
2. User clicks link → Opens in browser
3. Frontend displays wishlist (NO REDIRECTION)
4. User reserves gift → Stored as guest reservation (localStorage token)
5. User can view reservations at /my/reservations (NO ACCOUNT NEEDED)
```

### Flow 2: User Wanting to Create Wishlist
```
1. User visits domain.com
2. User clicks "Manage Your Wishlists" or "Open Mobile App"
3. Browser attempts deep link: wishlistapp://home
4. If app installed: Opens native app
5. If app not installed: Redirects to lk.domain.com (mobile web)
6. User registers/logs in → Creates wishlists
```

### Flow 3: User Trying to Login from Frontend
```
1. User visits domain.com/auth/login
2. MobileRedirect component activates automatically
3. Shows "Redirecting to mobile app..." message
4. Attempts deep link: wishlistapp://auth/login
5. Waits 1.5 seconds for app to open
6. If app doesn't open → Redirects to lk.domain.com/auth/login
```

### Flow 4: Authenticated User Accessing Reservations
```
1. Authenticated user visits domain.com/my/reservations
2. useAuthRedirect hook detects authentication
3. MobileRedirect component activates
4. Deep link: wishlistapp://my/reservations
5. Fallback: lk.domain.com/my/reservations
```

## Implementation Details

### Authentication Check

The redirection system uses a simple authentication check:

```typescript
const response = await fetch('/api/auth/me');
const isAuthenticated = response.ok;
```

If the user has a valid JWT token in cookies/localStorage, the backend returns 200 OK. Otherwise, it returns 401 Unauthorized.

### Visibility Detection

To determine if the mobile app opened, the system checks if the browser tab lost focus:

```typescript
setTimeout(() => {
  if (!document.hidden && document.visibilityState !== 'hidden') {
    // App didn't open, redirect to web version
    window.location.href = fallbackUrl;
  }
}, 1500);
```

**Why this works**:
- When a native app opens, the browser tab goes to background
- `document.hidden` becomes `true` and `visibilityState` becomes `'hidden'`
- If these values remain unchanged after 1.5 seconds, the app didn't open

### Guest Reservation Tracking

Guest users can reserve gifts without an account. The system tracks these using:

1. **Reservation Token**: Unique token stored in localStorage
2. **Backend API**: `/api/guest/reservations?token={token}`
3. **Persistence**: Token survives page reloads and browser restarts

**Flow**:
```typescript
// When guest reserves a gift
const token = crypto.randomUUID();
localStorage.setItem('reservationToken', token);

// When viewing reservations
const token = localStorage.getItem('reservationToken');
const response = await fetch(`/api/guest/reservations?token=${token}`);
```

## Security Considerations

### 1. No Sensitive Data in Deep Links
- Deep links contain only navigation paths, not tokens or user data
- Authentication tokens remain in HTTP-only cookies

### 2. HTTPS Enforcement
- All fallback URLs use HTTPS
- Mobile web version (lk.domain.com) enforces TLS

### 3. CORS Configuration
- Backend accepts requests from both domain.com and lk.domain.com
- Credentials are included in cross-origin requests

### 4. Token Security
- JWT tokens are HTTP-only cookies (not accessible to JavaScript)
- Guest reservation tokens are non-sensitive UUIDs
- No account access with guest tokens

## Testing

### Manual Testing Checklist

**Test 1: Login Redirection**
- [ ] Visit `/auth/login` from browser
- [ ] Verify redirection message appears
- [ ] Verify deep link attempt occurs
- [ ] Verify fallback to lk.domain.com after 1.5s (if no app)

**Test 2: Guest Reservations**
- [ ] Visit public wishlist without account
- [ ] Reserve a gift item
- [ ] Visit `/my/reservations` as guest
- [ ] Verify guest reservations display correctly
- [ ] Verify no redirection occurs for guests

**Test 3: Authenticated User Reservations**
- [ ] Login to mobile app
- [ ] Visit `/my/reservations` from browser
- [ ] Verify redirection to mobile app occurs
- [ ] Verify user can access their account reservations in app

**Test 4: Public Wishlist**
- [ ] Visit public wishlist as guest
- [ ] Verify no automatic redirection
- [ ] Verify all gift items display correctly
- [ ] Verify manual "Open Mobile App" button works

### Automated Testing

**Component Tests**:
```typescript
// Test MobileRedirect component
describe('MobileRedirect', () => {
  it('should attempt deep link on mount', () => {
    // Mock window.location.href
    // Verify deep link is constructed correctly
  });

  it('should fallback after timeout', async () => {
    // Mock setTimeout
    // Verify fallback URL is used after 1.5s
  });
});
```

**Integration Tests**:
```typescript
// Test auth redirection flow
describe('Auth Routes', () => {
  it('should redirect login page to mobile app', () => {
    // Visit /auth/login
    // Verify MobileRedirect renders
    // Verify deep link is attempted
  });
});
```

## Troubleshooting

### Issue: Redirection Loop
**Symptoms**: Page keeps redirecting back and forth
**Cause**: Both frontend and mobile web triggering redirects
**Solution**: Ensure mobile web (lk.domain.com) does NOT have redirection logic

### Issue: Deep Link Not Working on iOS
**Symptoms**: Fallback URL always loads, app never opens
**Cause**: URL scheme not configured in Info.plist
**Solution**: Add `wishlistapp` scheme to CFBundleURLTypes

### Issue: Guest Reservations Not Persisting
**Symptoms**: Reservations disappear after page reload
**Cause**: localStorage token lost or backend API issue
**Solution**: Check localStorage is enabled, verify token in API requests

### Issue: Authenticated Users Seeing Guest Flow
**Symptoms**: Authenticated users not redirected to mobile app
**Cause**: `/api/auth/me` endpoint failing or cookies not sent
**Solution**: Verify JWT token in cookies, check CORS configuration

## Future Enhancements

### 1. Universal Links (iOS) and App Links (Android)
Instead of custom URL schemes, use standard HTTPS links that open the app:
- iOS: `https://lk.domain.com/auth/login` opens app if installed
- Android: `https://lk.domain.com/auth/login` opens app if installed
- **Benefit**: More reliable, works in all contexts

### 2. Smart Banner
Add iOS Smart App Banner to prompt app installation:
```html
<meta name="apple-itunes-app" content="app-id=myAppStoreID">
```

### 3. Progressive Web App (PWA)
Enhance mobile web version with:
- Service workers for offline functionality
- Add to Home Screen capability
- Push notifications

### 4. Deferred Deep Linking
For users who don't have the app:
- Redirect to App Store/Play Store
- After installation, open app to originally intended destination
- Track installs from web redirects

## Conclusion

The mobile redirection architecture ensures a seamless experience for users by:
1. **Public Functionality**: Accessible to everyone via frontend (domain.com)
2. **Private Functionality**: Secured in mobile app (native or lk.domain.com)
3. **Smart Redirection**: Automatically guides users to the right interface
4. **Graceful Fallbacks**: Always provides alternative access via mobile web

This separation enhances security, improves user experience, and allows each platform to focus on its core strengths.

---

**Implementation**: T087 - Account Access Redirection Mechanism
**Related Requirements**: FR-001, FR-015
**Documentation Date**: 2026-01-23
