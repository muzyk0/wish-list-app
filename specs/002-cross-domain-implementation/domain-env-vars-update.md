# Domain Configuration via Environment Variables

**Date**: 2026-02-03
**Status**: ✅ COMPLETE

## Summary

Replaced hardcoded domain values in mobile app configuration with environment variables, enabling flexible deployment across different environments (development, staging, production).

## Changes Made

### 1. Mobile App Configuration

**Before** (`mobile/app.json`):
```json
{
  "expo": {
    "ios": {
      "associatedDomains": [
        "applinks:wishlist.com",
        "applinks:www.wishlist.com"
      ]
    },
    "android": {
      "intentFilters": [{
        "data": [
          { "scheme": "https", "host": "wishlist.com" },
          { "scheme": "https", "host": "www.wishlist.com" }
        ]
      }]
    }
  }
}
```

**After** (`mobile/app.config.js`):
```javascript
const WEB_DOMAIN = process.env.EXPO_PUBLIC_WEB_DOMAIN || 'wishlist.com';
const WWW_DOMAIN = process.env.EXPO_PUBLIC_WWW_DOMAIN || `www.${WEB_DOMAIN}`;

module.exports = {
  expo: {
    ios: {
      associatedDomains: [
        `applinks:${WEB_DOMAIN}`,
        `applinks:${WWW_DOMAIN}`,
      ]
    },
    android: {
      intentFilters: [{
        data: [
          { scheme: 'https', host: WEB_DOMAIN },
          { scheme: 'https', host: WWW_DOMAIN }
        ]
      }]
    }
  }
};
```

### 2. Environment Variables

**Added to `mobile/.env.example`**:
```bash
# Web Domain (for Universal Links and App Links)
EXPO_PUBLIC_WEB_DOMAIN=wishlist.com
EXPO_PUBLIC_WWW_DOMAIN=www.wishlist.com
```

**Added to `mobile/.env`**:
```bash
# Web Domain (for Universal Links and App Links)
EXPO_PUBLIC_WEB_DOMAIN=wishlist.com
EXPO_PUBLIC_WWW_DOMAIN=www.wishlist.com
```

## Benefits

1. **Environment Flexibility**: Different domains for dev/staging/prod
2. **No Hardcoding**: Easy to change without modifying source code
3. **Team Collaboration**: Each developer can use their own domain
4. **CI/CD Friendly**: Inject variables at build time

## Usage Examples

### Development (localhost)
```bash
EXPO_PUBLIC_WEB_DOMAIN=localhost:3000
EXPO_PUBLIC_WWW_DOMAIN=localhost:3000
```

### Staging
```bash
EXPO_PUBLIC_WEB_DOMAIN=staging.wishlist.com
EXPO_PUBLIC_WWW_DOMAIN=www.staging.wishlist.com
```

### Production
```bash
EXPO_PUBLIC_WEB_DOMAIN=wishlist.com
EXPO_PUBLIC_WWW_DOMAIN=www.wishlist.com
```

## Testing

**Verify environment variable support**:
```bash
# Test with custom domain
EXPO_PUBLIC_WEB_DOMAIN=example.com node -e "
  const config = require('./app.config.js');
  console.log('iOS domains:', config.expo.ios.associatedDomains);
  console.log('Android hosts:', config.expo.android.intentFilters[0].data.slice(0,2).map(d => d.host));
"
```

**Expected output**:
```
iOS domains: [ 'applinks:example.com', 'applinks:www.example.com' ]
Android hosts: [ 'example.com', 'www.example.com' ]
```

## Migration Notes

### For Existing Deployments

1. **Add environment variables** to deployment configuration:
   - Vercel: Add `EXPO_PUBLIC_WEB_DOMAIN` and `EXPO_PUBLIC_WWW_DOMAIN` in project settings
   - EAS Build: Add variables to `eas.json` or environment
   - Local: Update `.env` file

2. **Rebuild mobile app** to pick up new configuration:
   ```bash
   cd mobile
   npx expo prebuild --clean
   ```

3. **Update Universal/App Links** files on web server:
   - iOS: `.well-known/apple-app-site-association`
   - Android: `.well-known/assetlinks.json`

### Breaking Changes

- `app.json` is now `app.config.js` - Git ignored files may need updating
- Expo CLI will automatically use `app.config.js` over `app.json`
- No breaking changes for existing functionality

## Files Changed

**Created**:
- `mobile/app.config.js` - Dynamic configuration with env vars

**Modified**:
- `mobile/.env.example` - Added domain variables
- `mobile/.env` - Added domain variables
- `specs/002-cross-domain-implementation/phase3-completion-report.md` - Updated docs

**Deleted**:
- `mobile/app.json` - Replaced by app.config.js

## Related Configuration

### Backend CORS

The backend already supports environment-based CORS configuration:

```bash
# backend/.env
CORS_ALLOWED_ORIGINS=https://wishlist.com,https://www.wishlist.com
```

### Frontend Environment

Frontend API URL is already configurable:

```bash
# frontend/.env.local
NEXT_PUBLIC_API_URL=https://api.wishlist.com/api
```

## Validation

✅ Syntax validation: `node -e "require('./app.config.js')"`
✅ Environment override: Works correctly
✅ Formatting: Applied via Biome
✅ Default values: Fall back to `wishlist.com` if not set

## Next Steps

When deploying to production:
1. Set environment variables in deployment platform
2. Verify Universal/App Links configuration
3. Test deep linking with production domains
4. Update DNS and SSL certificates if needed

## Documentation Updated

- [x] Phase 3 completion report
- [x] Environment variables section
- [x] Deployment notes
- [x] Files changed list
- [x] This document
