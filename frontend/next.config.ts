import type { NextConfig } from 'next';

const path = require('node:path');

const nextConfig: NextConfig = {
  /* config options here */
  reactCompiler: true,
  outputFileTracingRoot: path.join(__dirname),
  // turbopack: {
  //   root: path.join(__dirname),
  // },
  images: {
    remotePatterns: [
      { protocol: 'https', hostname: '**' },
      { protocol: 'http', hostname: '**' },
    ],
  },
};

export default nextConfig;
