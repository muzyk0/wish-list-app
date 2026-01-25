# Authentication Guide

This document provides information about the authentication system in the mobile application, including OAuth integration and environment configuration.

## Table of Contents
- [Overview](#overview)
- [OAuth Providers](#oauth-providers)
- [Environment Variables](#environment-variables)
- [Implementation Details](#implementation-details)
- [Security Considerations](#security-considerations)

## Overview

The mobile application supports multiple authentication methods:
- Traditional email/password authentication
- OAuth with Google, Facebook, and Apple
- Session management and token handling

## OAuth Providers

### Google OAuth
- Uses Google's OAuth 2.0 protocol
- Supports web, Android, and iOS platforms
- Requires Google Cloud Console project setup

### Facebook OAuth
- Uses Facebook's OAuth 2.0 protocol
- Supports web, Android, and iOS platforms
- Requires Facebook Developer account setup

### Apple OAuth
- Uses Apple's Sign In with Apple
- iOS only (currently not implemented in demo)
- Requires Apple Developer account setup

## Environment Variables

The following environment variables need to be configured in your `.env` file:

```env
# OAuth - Google
EXPO_PUBLIC_GOOGLE_WEB_CLIENT_ID=your_google_web_client_id
EXPO_PUBLIC_GOOGLE_ANDROID_CLIENT_ID=your_google_android_client_id
EXPO_PUBLIC_GOOGLE_IOS_CLIENT_ID=your_google_ios_client_id

# OAuth - Facebook
EXPO_PUBLIC_FACEBOOK_CLIENT_ID=your_facebook_client_id
```

### Setting Up OAuth Providers

#### Google OAuth
1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the Google+ API
4. Create OAuth 2.0 credentials for:
   - Web application (for web builds)
   - Android application (for Android builds)
   - iOS application (for iOS builds)
5. Add your app's redirect URIs to the OAuth configuration

#### Facebook OAuth
1. Go to [Facebook Developers](https://developers.facebook.com/)
2. Create a new app or select an existing one
3. Add the "Facebook Login" product
4. Configure the OAuth redirect URIs
5. Get your App ID and configure it as `EXPO_PUBLIC_FACEBOOK_CLIENT_ID`

#### Apple OAuth (iOS Only)
1. Go to [Apple Developer Portal](https://developer.apple.com/account/)
2. Create an App ID with "Sign In with Apple" enabled
3. Configure your app for Sign In with Apple
4. Note: This requires additional setup in your app configuration

## Implementation Details

### OAuth Service (`lib/oauth-service.ts`)
- Handles the OAuth flows for different providers
- Manages platform-specific configurations
- Provides error handling and user feedback
- Returns authentication tokens for backend verification

### OAuth Buttons (`components/OAuthButton.tsx`)
- Reusable component for OAuth provider buttons
- Proper theming support for light/dark modes
- Loading states and error handling
- Responsive design for different screen sizes

### Integration Points
- Login screen (`app/auth/login.tsx`)
- Register screen (`app/auth/register.tsx`)
- Both screens include OAuth buttons with proper error handling

## Security Considerations

### Token Handling
- OAuth tokens should be exchanged for app-specific tokens on your backend
- Never store sensitive tokens in plain text
- Use secure storage for persistent authentication

### Redirect URIs
- Configure proper redirect URIs for each platform
- Validate redirect URIs on your backend
- Use platform-specific app schemes for native apps

### Client Secrets
- Never expose client secrets in client-side code
- Use backend services for token verification
- Implement proper server-side validation

## Troubleshooting

### Common Issues
- Invalid redirect URI: Ensure your OAuth configuration matches your app's scheme
- Missing client ID: Add required environment variables to your `.env` file
- Platform-specific errors: Check platform-specific OAuth configurations

### Debugging OAuth Flows
- Enable logging in the OAuth service for debugging
- Verify your OAuth provider configurations
- Check network connectivity and CORS settings

## Testing OAuth Locally

1. Add OAuth client IDs to your `.env` file
2. Run the app with `npx expo start`
3. Test OAuth flows on your preferred platform
4. Monitor the logs for any OAuth-related errors

For local development, you may need to configure additional redirect URIs for localhost testing.