# OAuth 2.0 Implementation Guide

**Date**: 2026-02-03
**Status**: Completed

## Overview

Implemented proper OAuth 2.0 Authorization Code flow for Google and Facebook authentication with backend token exchange and JWT generation.

## Architecture

```
┌─────────┐                  ┌──────────┐                  ┌─────────┐
│ Mobile  │                  │  Google  │                  │ Backend │
│  App    │                  │  OAuth   │                  │   API   │
└────┬────┘                  └────┬─────┘                  └────┬────┘
     │                            │                             │
     │ 1. Auth Request            │                             │
     ├───────────────────────────>│                             │
     │                            │                             │
     │ 2. User Login              │                             │
     │<───────────────────────────┤                             │
     │                            │                             │
     │ 3. Authorization Code      │                             │
     │<───────────────────────────┤                             │
     │ code=abc123                │                             │
     │                            │                             │
     │ 4. Send Code               │                             │
     ├─────────────────────────────────────────────────────────>│
     │ POST /auth/oauth/google    │                             │
     │ { "code": "abc123" }       │                             │
     │                            │                             │
     │                            │ 5. Exchange Code            │
     │                            │<────────────────────────────┤
     │                            │ + client_secret             │
     │                            │                             │
     │                            │ 6. Access Token             │
     │                            │─────────────────────────────>│
     │                            │                             │
     │                            │ 7. Get User Info            │
     │                            │<────────────────────────────┤
     │                            │                             │
     │                            │ 8. User Data                │
     │                            │─────────────────────────────>│
     │                            │                             │
     │                            │ 9. Create/Find User in DB   │
     │                            │                             │
     │                            │ 10. Generate OUR JWTs       │
     │                            │                             │
     │ 11. Return { accessToken, refreshToken }                 │
     │<─────────────────────────────────────────────────────────┤
     │                            │                             │
     │ 12. Store in SecureStore   │                             │
     │                            │                             │
```

## Implementation Details

### Backend

#### 1. OAuth Handler (`backend/internal/handlers/oauth_handler.go`)

**Features**:
- ✅ Google OAuth code exchange
- ✅ Facebook OAuth code exchange
- ✅ User info retrieval from providers
- ✅ User creation/lookup in database
- ✅ JWT token generation (access + refresh)
- ✅ Swagger documentation

**Endpoints**:
- `POST /api/auth/oauth/google` - Exchange Google authorization code
- `POST /api/auth/oauth/facebook` - Exchange Facebook authorization code

**Request**:
```json
{
  "code": "authorization_code_from_provider"
}
```

**Response**:
```json
{
  "accessToken": "jwt_access_token",
  "refreshToken": "jwt_refresh_token",
  "user": {
    "id": "uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "avatar_url": "https://..."
  }
}
```

#### 2. Configuration (`backend/internal/config/config.go`)

Added OAuth configuration fields:
```go
type Config struct {
    // ... existing fields ...
    GoogleClientID       string
    GoogleClientSecret   string
    FacebookClientID     string
    FacebookClientSecret string
    OAuthRedirectURL     string
}
```

Environment variables:
```bash
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
FACEBOOK_CLIENT_ID=your-facebook-app-id
FACEBOOK_CLIENT_SECRET=your-facebook-app-secret
OAUTH_REDIRECT_URL=wishlistapp://oauth
```

#### 3. Dependencies

Added OAuth 2.0 library:
```bash
go get golang.org/x/oauth2
go get golang.org/x/oauth2/google
go get golang.org/x/oauth2/facebook
```

### Mobile

#### 1. OAuth Service (`mobile/lib/oauth-service.ts`)

**Updated flow**:
```typescript
// Before (WRONG):
return { success: true, token: code };  // ❌ Returned code directly

// After (CORRECT):
const response = await fetch(`${backendUrl}/api/auth/oauth/google`, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ code }),
});
const data = await response.json();
return {
  success: true,
  accessToken: data.accessToken,    // ✅ Real tokens
  refreshToken: data.refreshToken,  // ✅ from backend
};
```

**Return Type**:
```typescript
interface OAuthResult {
  success: boolean;
  accessToken?: string;   // ✅ 15 minutes
  refreshToken?: string;  // ✅ 7 days
  error?: string;
}
```

#### 2. Login Screen (`mobile/app/auth/login.tsx`)

**Updated handler**:
```typescript
if (result.success && result.accessToken && result.refreshToken) {
  // Store both tokens properly
  await setTokens(result.accessToken, result.refreshToken);
  router.push('/(tabs)');
}
```

## Security Benefits

### Before (Insecure)
- ❌ Authorization code returned to app as "token"
- ❌ No backend validation
- ❌ client_secret would be in mobile app (if implemented)
- ❌ No control over token generation
- ❌ Single token (no refresh)

### After (Secure)
- ✅ Authorization code exchanged by backend
- ✅ client_secret stays on backend (never exposed)
- ✅ Backend validates with OAuth provider
- ✅ Backend controls user creation and authentication
- ✅ Standardized JWT tokens (access + refresh)
- ✅ Tokens follow cross-domain auth strategy

## Setup Instructions

### 1. Backend Configuration

Create/update `backend/.env`:
```bash
# Google OAuth (https://console.cloud.google.com/apis/credentials)
GOOGLE_CLIENT_ID=your-google-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-google-client-secret

# Facebook OAuth (https://developers.facebook.com/apps/)
FACEBOOK_CLIENT_ID=your-facebook-app-id
FACEBOOK_CLIENT_SECRET=your-facebook-app-secret

# OAuth redirect URL (must match mobile app scheme)
OAUTH_REDIRECT_URL=wishlistapp://oauth
```

### 2. Google OAuth Setup

1. Go to https://console.cloud.google.com/apis/credentials
2. Create OAuth 2.0 Client ID
3. Configure for Android/iOS:
   - **Android**: Add package name and SHA-1 certificate fingerprint
   - **iOS**: Add bundle ID and App Store ID
4. Add redirect URI: `wishlistapp://oauth`
5. Copy Client ID and Client Secret to `.env`

### 3. Facebook OAuth Setup

1. Go to https://developers.facebook.com/apps/
2. Create new app or use existing
3. Add "Facebook Login" product
4. Configure OAuth redirect URIs:
   - Add `wishlistapp://oauth`
5. Copy App ID and App Secret to `.env`

### 4. Mobile Configuration

Update `mobile/.env`:
```bash
EXPO_PUBLIC_API_URL=http://your-backend-url/api
EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID=your-google-client-id
EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID=your-android-client-id
EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID=your-ios-client-id
EXPO_PUBLIC_FACEBOOK_CLIENT_ID=your-facebook-app-id
EXPO_PUBLIC_APP_SCHEME=wishlistapp
```

## Testing

### Test Google OAuth Flow

1. Start backend: `cd backend && go run ./cmd/server`
2. Start mobile: `cd mobile && npx expo start`
3. Click "Sign in with Google"
4. Complete OAuth in browser
5. Verify redirect to app with tokens stored

### Test with curl

```bash
# Simulate mobile app sending code to backend
curl -X POST http://localhost:8080/api/auth/oauth/google \
  -H "Content-Type: application/json" \
  -d '{"code":"test_authorization_code"}'
```

## Swagger Documentation

After starting backend, view OAuth endpoints:
```
http://localhost:8080/swagger/index.html
```

Search for "OAuth" to see:
- `POST /auth/oauth/google`
- `POST /auth/oauth/facebook`

## Next Steps

### Future Enhancements

1. **Apple Sign In** - Implement `POST /auth/oauth/apple` endpoint
2. **Token Blacklisting** - Add refresh token revocation on logout
3. **OAuth Scopes** - Request additional permissions as needed
4. **Provider Linking** - Allow linking multiple OAuth providers to one account
5. **OAuth State** - Add CSRF protection with state parameter

### Apple Sign In Implementation

For reference, the flow would be:
```typescript
import * as AppleAuthentication from 'expo-apple-authentication';

const credential = await AppleAuthentication.signInAsync({
  requestedScopes: [
    AppleAuthentication.AppleAuthenticationScope.FULL_NAME,
    AppleAuthentication.AppleAuthenticationScope.EMAIL,
  ],
});

// Send identityToken to backend
const response = await fetch(`${backendUrl}/api/auth/oauth/apple`, {
  method: 'POST',
  body: JSON.stringify({ identityToken: credential.identityToken }),
});
```

## Troubleshooting

### "Failed to exchange authorization code"
- Check that client_secret is correct in `.env`
- Verify redirect URI matches exactly
- Ensure OAuth app is not in test mode

### "Invalid redirect URI"
- Mobile app scheme must match `OAUTH_REDIRECT_URL`
- Check `app.json` scheme configuration
- Verify OAuth provider settings

### "User email not verified"
- Google requires verified email
- User must verify email with provider first

## References

- [OAuth 2.0 RFC](https://datatracker.ietf.org/doc/html/rfc6749)
- [Google OAuth 2.0](https://developers.google.com/identity/protocols/oauth2)
- [Facebook Login](https://developers.facebook.com/docs/facebook-login/)
- [Expo AuthSession](https://docs.expo.dev/versions/latest/sdk/auth-session/)
