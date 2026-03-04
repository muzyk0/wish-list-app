# Image Domains Configuration

## Location
`frontend/config/image-domains.ts`

## How It Works
- Centralized whitelist of allowed image source domains
- Imported in `next.config.ts` for `images.remotePatterns`
- Easy to add new domains without modifying next.config

## Currently Supported
- AWS S3 (*.amazonaws.com)
- Apple (store.storeimages.cdn-apple.com, *.apple.com)
- Amazon (*.media-amazon.com, *.ssl-images-amazon.com)
- Shopify (cdn.shopify.com)
- CDNs (Cloudinary, Unsplash, Pexels)
- Local development (localhost)

## Adding New Domains
```typescript
// In config/image-domains.ts, add to IMAGE_DOMAINS array:
{ protocol: 'https', hostname: 'new-domain.com' },
{ protocol: 'https', hostname: '*.new-domain.com' }, // wildcards OK
```

## Interface
```typescript
interface RemotePattern {
  protocol: 'http' | 'https';
  hostname: string; // wildcards like *.domain.com supported
}
```
