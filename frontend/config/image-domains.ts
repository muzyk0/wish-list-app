/**
 * Whitelisted domains for next/image remotePatterns
 * Add new domains here as we support more product sources
 */

interface RemotePattern {
  protocol: 'http' | 'https';
  hostname: string;
}

const IMAGE_DOMAINS: RemotePattern[] = [
  // AWS S3 (our own uploads)
  { protocol: 'https', hostname: '*.amazonaws.com' },

  // Local development
  { protocol: 'http', hostname: 'localhost' },

  // E-commerce & Product Sources
  // Apple
  { protocol: 'https', hostname: 'store.storeimages.cdn-apple.com' },
  { protocol: 'https', hostname: '*.apple.com' },

  // Amazon
  { protocol: 'https', hostname: 'images-na.ssl-images-amazon.com' },
  { protocol: 'https', hostname: '*.media-amazon.com' },

  // Common CDNs & image hosts
  { protocol: 'https', hostname: 'cdn.shopify.com' },
  { protocol: 'https', hostname: '*.cloudinary.com' },
  { protocol: 'https', hostname: 'images.unsplash.com' },
  { protocol: 'https', hostname: 'images.pexels.com' },

  // E-commerce platforms
  { protocol: 'https', hostname: 'images.benvenue.de' }, // MediaMarkt
  { protocol: 'https', hostname: 'images-eu.ssl-images-amazon.com' },
];

export default IMAGE_DOMAINS;
