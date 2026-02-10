import type { NextConfig } from 'next';
const path = require('path');

const nextConfig: NextConfig = {
  /* config options here */
  reactCompiler: true,
  outputFileTracingRoot: path.join(__dirname),
  // turbopack: {
  //   root: path.join(__dirname),
  // },
};

export default nextConfig;
