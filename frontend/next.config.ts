import type { NextConfig } from 'next';
import IMAGE_DOMAINS from './config/image-domains';

const path = require('node:path');

const nextConfig: NextConfig = {
  /* config options here */
  reactCompiler: true,
  outputFileTracingRoot: path.join(__dirname),
  // turbopack: {
  //   root: path.join(__dirname),
  // },
  images: {
    remotePatterns: IMAGE_DOMAINS,
  },
};

export default nextConfig;
