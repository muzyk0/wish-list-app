# Deep Linking Implementation Guide

**Application**: Wish List Application
**Version**: 1.1.0
**Tasks**: T088 - Deep Linking Support

## Overview

Deep linking allows the web frontend to open specific screens in the mobile app using custom URL schemes. This implementation enables seamless navigation from web to app, providing users with direct access to account management features.

## URL Scheme

**Primary Scheme**: `wishlistapp://`
**Fallback**: Mobile web at `https://lk.domain.com`

## Supported Deep Links

| Deep Link | Mobile Route | Description |
|-----------|--------------|-------------|
| `wishlistapp://home` | `/(tabs)/index` | Home screen |
| `wishlistapp://auth/login` | `/auth/login` | Login screen |
| `wishlistapp://auth/register` | `/auth/register` | Registration screen |
| `wishlistapp://my/reservations` | `/(tabs)/reservations` | User's reservations |
| `wishlistapp://lists` | `/(tabs)/lists` | Wishlists tab |
| `wishlistapp://lists/123` | `/lists/[id]` | View specific wishlist |
| `wishlistapp://lists/123/edit` | `/lists/[id]/edit` | Edit wishlist |
| `wishlistapp://gift-items/456/edit` | `/gift-items/[id]/edit` | Edit gift item |
| `wishlistapp://public/birthday-2026` | `/public/[slug]` | Public wishlist view |
| `wishlistapp://explore` | `/(tabs)/explore` | Explore tab |
| `wishlistapp://profile` | `/(tabs)/profile` | User profile |

## Implementation Details

### 1. Mobile App Configuration (app.json)

```json
{
  "expo": {
    "name": "WishList",
    "slug": "wishlist",
    "scheme": "wishlistapp",
    ...
  }
}
```

**Changes**:
- Updated `name` from "mobile" to "WishList"
- Updated `slug` from "mobile" to "wishlist"
- Updated `scheme` from "mobile" to "wishlistapp"

### 2. iOS Configuration (Info.plist)

**File**: `mobile/ios/mobile/Info.plist`

```xml
<key>CFBundleURLTypes</key>
<array>
  <dict>
    <key>CFBundleURLSchemes</key>
    <array>
      <string>wishlistapp</string>
      <string>mobile</string>
      <string>com.anonymous.mobile</string>
    </array>
  </dict>
</array>
```

**Notes**:
- Primary scheme: `wishlistapp`
- Kept legacy schemes for backward compatibility
- No code changes needed after `expo prebuild`

### 3. Android Configuration (AndroidManifest.xml)

**File**: `mobile/android/app/src/main/AndroidManifest.xml`

```xml
<intent-filter>
  <action android:name="android.intent.action.VIEW"/>
  <category android:name="android.intent.category.DEFAULT"/>
  <category android:name="android.intent.category.BROWSABLE"/>
  <data android:scheme="wishlistapp"/>
</intent-filter>
```

**Notes**:
- Added separate intent-filter for `wishlistapp` scheme
- Kept existing `mobile` scheme for backward compatibility
- Allows app to respond to both URL schemes

### 4. Deep Link Handling (_layout.tsx)

**File**: `mobile/app/_layout.tsx`

**Functionality**:
- Listens for deep link events on app start (cold start)
- Listens for deep link events while app is running (warm start)
- Parses incoming URLs and extracts paths
- Maps web paths to mobile routes
- Navigates to the appropriate screen

**Key Features**:

1. **Route Mapping**: Simple routes are mapped via dictionary
   ```typescript
   const routeMap: { [key: string]: string } = {
     'home': '/(tabs)',
     'auth/login': '/auth/login',
     'my/reservations': '/(tabs)/reservations',
     ...
   };
   ```

2. **Parameterized Routes**: Dynamic routes using Expo Router patterns
   ```typescript
   // Example: wishlistapp://lists/123
   if (path.startsWith('lists/')) {
     const match = path.match(/^lists\/([^\/]+)/);
     if (match && match[1]) {
       const id = match[1];
       // Method 1: String path (inline ID)
       router.push(`/lists/${id}`);

       // Method 2: Object with params (type-safe)
       router.navigate({
         pathname: '/lists/[id]',
         params: { id }
       });
     }
   }
   ```

   **Best Practice**: Use regex matching instead of split() for robust parameter extraction.

   **TypeScript Type Safety**:
   ```typescript
   // ✅ Correct - typed params object
   router.navigate({
     pathname: '/lists/[id]',
     params: { id: '123' }
   });

   // ❌ Wrong - unmatched route pattern
   router.navigate('/lists/[id]'); // Missing params
   ```

3. **Cold Start**: Handles deep links when app is closed
   ```typescript
   Linking.getInitialURL().then((url) => {
     if (url) {
       handleDeepLink({ url });
     }
   });
   ```

4. **Warm Start**: Handles deep links when app is already running
   ```typescript
   const subscription = Linking.addEventListener('url', handleDeepLink);
   ```

### 5. Linking Configuration (linking.ts)

**File**: `mobile/app/linking.ts`

**Purpose**: Documents the linking configuration and URL structure

**Contents**:
- Prefixes: `['wishlistapp://', 'https://lk.domain.com']`
- Screen mapping configuration
- URL examples with descriptions

**Note**: This file serves as documentation. The actual routing logic is in `_layout.tsx` for better control.

### 6. Declarative Navigation with Link Component

**Using Link Component for Internal Navigation**:

```typescript
import { Link } from 'expo-router';
import { View, Text, Pressable } from 'react-native';

export default function ListsScreen() {
  return (
    <View>
      {/* Method 1: String href (inline ID) */}
      <Link href="/lists/123">
        View List 123
      </Link>

      {/* Method 2: Object href with params (type-safe) */}
      <Link
        href={{
          pathname: '/lists/[id]',
          params: { id: '123' }
        }}
      >
        View List (typed)
      </Link>

      {/* Method 3: With query parameters */}
      <Link
        href={{
          pathname: '/lists/[id]',
          params: { id: '123', edit: 'true' }
        }}
      >
        Edit List
      </Link>

      {/* Method 4: Using asChild with custom component */}
      <Link href="/lists/123" asChild>
        <Pressable>
          <Text>Custom Button</Text>
        </Pressable>
      </Link>

      {/* Method 5: Relative navigation */}
      <Link href="./settings">
        Settings (relative)
      </Link>
    </View>
  );
}
```

**Imperative Navigation with router**:

```typescript
import { router } from 'expo-router';
import { Pressable, Text } from 'react-native';

export default function Screen() {
  return (
    <Pressable
      onPress={() =>
        router.navigate({
          pathname: '/lists/[id]',
          params: { id: '123' }
        })
      }
    >
      <Text>Navigate to List</Text>
    </Pressable>
  );
}
```

### 7. Accessing Route Parameters

**File**: Any screen component with dynamic routes (e.g., `app/lists/[id]/index.tsx`)

**Using useLocalSearchParams Hook** (Recommended):
```typescript
import { useLocalSearchParams } from 'expo-router';
import { View, Text } from 'react-native';

export default function ListDetails() {
  // Access dynamic route parameters and query params
  const { id, edit } = useLocalSearchParams();

  return (
    <View>
      <Text>List ID: {id}</Text>
      <Text>Edit mode: {edit}</Text>
    </View>
  );
}
```

**Key Features**:
- Returns URL parameters for the currently focused route
- Includes both route parameters (`[id]`) and query parameters (`?edit=true`)
- Type-safe with TypeScript
- Re-renders when parameters change

**Getting Current Pathname**:
```typescript
import { usePathname } from 'expo-router';

export default function Route() {
  // Returns normalized pathname without query params
  const pathname = usePathname(); // e.g., "/lists/123"

  return <Text>Current path: {pathname}</Text>;
}
```

**Example Routes and Parameters**:
```typescript
// Route: /lists/[id]?edit=true
const params = useLocalSearchParams();
// params = { id: "123", edit: "true" }

// Route: /user/[id]/posts/[postId]
const params = useLocalSearchParams();
// params = { id: "user123", postId: "post456" }
```

## Testing Deep Links

### iOS Simulator

```bash
# Open a deep link in iOS Simulator
xcrun simctl openurl booted wishlistapp://auth/login

# Test with parameters
xcrun simctl openurl booted wishlistapp://lists/123
```

### Android Emulator

```bash
# Open a deep link in Android Emulator
adb shell am start -W -a android.intent.action.VIEW -d "wishlistapp://auth/login"

# Test with parameters
adb shell am start -W -a android.intent.action.VIEW -d "wishlistapp://lists/123"
```

### Physical Devices

**Method 1: Test Links via Safari/Chrome**
1. Open Safari (iOS) or Chrome (Android)
2. Type deep link in address bar: `wishlistapp://auth/login`
3. App should open to the specified screen

**Method 2: Test from Frontend**
1. Deploy frontend to device-accessible URL
2. Visit auth/login page
3. Allow automatic redirect or click manual link
4. App should open to login screen

### Test Script

Create a test file for comprehensive testing:

```bash
#!/bin/bash
# test-deep-links.sh

echo "Testing Deep Links..."

# iOS
xcrun simctl openurl booted wishlistapp://home
sleep 2
xcrun simctl openurl booted wishlistapp://auth/login
sleep 2
xcrun simctl openurl booted wishlistapp://lists
sleep 2
xcrun simctl openurl booted wishlistapp://lists/123

# Android
# adb shell am start -W -a android.intent.action.VIEW -d "wishlistapp://home"
# ...
```

## Debugging Deep Links

### Enable Logging

Add debug logs to `_layout.tsx`:

```typescript
const handleDeepLink = (event: { url: string }) => {
  console.log('🔗 Deep link received:', event.url);

  const { path, queryParams } = Linking.parse(event.url);
  console.log('📍 Parsed path:', path);
  console.log('🔍 Query params:', queryParams);

  // ... rest of handling logic
};
```

### Check URL Registration

**iOS**:
```bash
# Check registered URL schemes
plutil -p ios/mobile/Info.plist | grep -A10 CFBundleURLSchemes
```

**Android**:
```bash
# Check intent filters
grep -A5 "android:scheme" android/app/src/main/AndroidManifest.xml
```

### Verify Expo Configuration

```bash
# Check Expo configuration
cat app.json | grep -A3 "scheme"
```

## Common Issues and Solutions

### Issue 1: Deep Link Not Opening App

**Symptoms**: Clicking deep link does nothing or opens browser

**Solutions**:
1. Verify URL scheme is registered in `app.json`
2. Run `expo prebuild` to regenerate native files
3. Rebuild the app: `expo run:ios` or `expo run:android`
4. Check Info.plist (iOS) or AndroidManifest.xml (Android) for correct scheme

### Issue 2: App Opens But Doesn't Navigate

**Symptoms**: App opens to default screen, not target screen

**Solutions**:
1. Check `_layout.tsx` has deep link handling code
2. Verify route mapping includes the target path
3. Add debug logs to `handleDeepLink` function
4. Check for typos in path matching logic

### Issue 3: Parameterized Routes Not Working

**Symptoms**: Routes like `/lists/123` don't work

**Solutions**:

1. **Use regex matching for parameter extraction** (not split):
   ```typescript
   // ✅ Correct - robust parameter extraction
   const match = path.match(/^lists\/([^\/]+)/);
   if (match && match[1]) {
     const id = match[1];
     router.push(`/lists/${id}`);
   }

   // ❌ Wrong - brittle parsing
   const id = path.split('/')[1]; // May get wrong segment
   ```

2. **Use typed navigation with params object**:
   ```typescript
   // ✅ Type-safe navigation
   router.navigate({
     pathname: '/lists/[id]',
     params: { id: '123' }
   });

   // ✅ Also valid - inline ID
   router.push('/lists/123');

   // ❌ Wrong - missing params
   router.navigate('/lists/[id]');
   ```

3. **Verify file-based route structure**:
   - Check file exists: `app/lists/[id]/index.tsx`
   - Folder name must match param: `[id]` not `[listId]`
   - Use `useLocalSearchParams()` in component to access `id`

4. **Test with both navigation methods**:
   ```typescript
   // Test inline
   router.push('/lists/123');

   // Test with params object
   router.navigate({
     pathname: '/lists/[id]',
     params: { id: '123' }
   });
   ```

5. **Debug parameter extraction**:
   ```typescript
   const { id } = useLocalSearchParams();
   console.log('Route parameter:', id);
   // Should log: "123"
   ```

### Issue 4: Cold Start vs Warm Start Inconsistency

**Symptoms**: Works when app is running, not when closed

**Solutions**:
1. Ensure `Linking.getInitialURL()` is called
2. Check that `handleDeepLink` works synchronously
3. Test with `await Linking.getInitialURL()` in useEffect
4. Verify no race conditions with navigation

### Issue 5: Android Deep Links Fail

**Symptoms**: Works on iOS but not Android

**Solutions**:
1. Check AndroidManifest.xml has correct intent-filter
2. Verify `android:scheme` attribute is set
3. Test with `adb shell` command directly
4. Ensure package name matches in AndroidManifest.xml

## Security Considerations

### 1. URL Validation

Always validate deep link parameters:

```typescript
const handleDeepLink = (event: { url: string }) => {
  const { path } = Linking.parse(event.url);

  // Validate path format
  if (!isValidPath(path)) {
    console.warn('Invalid deep link path:', path);
    return;
  }

  // Sanitize parameters
  const sanitizedId = sanitizeId(extractId(path));

  // Navigate safely
  router.push(`/lists/${sanitizedId}`);
};
```

### 2. Authentication Check

Protect authenticated routes:

```typescript
const handleDeepLink = (event: { url: string }) => {
  const { path } = Linking.parse(event.url);

  // Check if route requires auth
  if (requiresAuth(path)) {
    const isAuthenticated = checkAuthStatus();

    if (!isAuthenticated) {
      // Redirect to login with return path
      router.push(`/auth/login?returnTo=${encodeURIComponent(path)}`);
      return;
    }
  }

  // Navigate to target
  navigateToPath(path);
};
```

### 3. Rate Limiting

Prevent deep link abuse:

```typescript
let lastDeepLinkTime = 0;
const RATE_LIMIT_MS = 1000; // 1 second

const handleDeepLink = (event: { url: string }) => {
  const now = Date.now();

  if (now - lastDeepLinkTime < RATE_LIMIT_MS) {
    console.warn('Deep link rate limit exceeded');
    return;
  }

  lastDeepLinkTime = now;

  // Process deep link
  processDeepLink(event.url);
};
```

### 4. URL Scheme Hijacking Prevention

**Mitigation**:
- Use unique, branded scheme: `wishlistapp://` not generic like `myapp://`
- Implement Universal Links (iOS) and App Links (Android) for HTTPS fallback
- Validate incoming URLs against whitelist of allowed patterns
- Never execute code directly from deep link parameters

## Universal Links (Future Enhancement)

### Why Universal Links?

**Benefits over Custom URL Schemes**:
- More reliable: Works even when app not installed (fallback to web)
- Secure: Domain ownership verified by OS
- Better UX: Seamless transition between web and app
- SEO friendly: Same URLs work for both web and app

### Implementation Plan

1. **iOS Universal Links Configuration**:

   Add to `app.json`:
   ```json
   {
     "expo": {
       "ios": {
         "associatedDomains": [
           "applinks:lk.domain.com",
           "webcredentials:lk.domain.com"
         ],
         "entitlements": {
           "com.apple.developer.associated-domains": [
             "applinks:lk.domain.com"
           ]
         }
       }
     }
   }
   ```

   **Steps**:
   - Host `apple-app-site-association` file at `https://lk.domain.com/.well-known/apple-app-site-association`
   - File must be served with `Content-Type: application/json`
   - No `.json` extension required
   - Verify file is accessible: `curl https://lk.domain.com/.well-known/apple-app-site-association`

2. **Android App Links Configuration**:

   Add to `app.json`:
   ```json
   {
     "expo": {
       "android": {
         "intentFilters": [
           {
             "action": "VIEW",
             "autoVerify": true,
             "data": [
               {
                 "scheme": "https",
                 "host": "lk.domain.com",
                 "pathPrefix": "/lists"
               }
             ],
             "category": ["BROWSABLE", "DEFAULT"]
           }
         ]
       }
     }
   }
   ```

   **Steps**:
   - Host `assetlinks.json` at `https://lk.domain.com/.well-known/assetlinks.json`
   - Must be served with `Content-Type: application/json`
   - File must include app package name and SHA-256 certificate fingerprints
   - Verify with: `curl https://lk.domain.com/.well-known/assetlinks.json`
   - Test with: `adb shell pm verify-app-links --re-verify <package-name>`

3. **Configuration with Wildcard Subdomains**:
   ```json
   {
     "expo": {
       "android": {
         "intentFilters": [
           {
             "action": "VIEW",
             "autoVerify": true,
             "data": [
               {
                 "scheme": "https",
                 "host": "*.domain.com",
                 "pathPrefix": "/records"
               }
             ],
             "category": ["BROWSABLE", "DEFAULT"]
           }
         ]
       }
     }
   }
   ```

   **Use Cases**:
   - Multi-tenant apps with subdomain per tenant
   - Staging/production environments
   - Regional domains

## Performance Considerations

### 1. Fast Navigation

- Deep link handling should be fast (<100ms)
- Avoid async operations in critical path
- Cache route mappings
- Use optimized path parsing

### 2. Memory Efficiency

- Clean up event listeners on unmount
- Avoid memory leaks with useEffect cleanup
- Don't store large state in deep link handler

### 3. User Experience

- Show loading indicator during navigation
- Handle errors gracefully
- Provide fallback navigation if target screen unavailable
- Preserve user's place in app if deep link fails

## Testing Checklist

- [ ] Test all documented deep links on iOS
- [ ] Test all documented deep links on Android
- [ ] Test cold start (app closed)
- [ ] Test warm start (app running in background)
- [ ] Test with valid parameters
- [ ] Test with invalid parameters
- [ ] Test with missing parameters
- [ ] Test authentication-required routes
- [ ] Test public routes
- [ ] Test from web frontend
- [ ] Test from mobile browser
- [ ] Test from email links
- [ ] Test from SMS links
- [ ] Test concurrent deep links
- [ ] Test during slow navigation
- [ ] Test error scenarios

## Conclusion

Deep linking implementation successfully connects the web frontend to the mobile app, providing seamless navigation between platforms. The implementation supports both custom URL schemes and is ready for future Universal Links/App Links enhancement.

---

**Task**: T088 - Add deep linking support from web to mobile app
**Status**: ✅ Complete
**Implementation Date**: 2026-01-23
**Platform Support**: iOS, Android, Web (Expo Router)
