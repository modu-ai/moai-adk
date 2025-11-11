const nextra = require("nextra");
const withNextra = nextra.default || nextra;

/** @type {import('next').NextConfig} */
const nextConfig = {
  // Development mode - no static export
  // Remove output: 'export' for development

  // Enable experimental features for Nextra 4.x with Next.js 15
  experimental: {
    webpackBuildWorker: true,
    // Next.js 15 specific optimizations
    optimizePackageImports: ['nextra', 'nextra-theme-docs'],
  },

  // Removed invalid turbo configuration
  // Turbopack is automatically configured by Next.js 15

  // Image optimization
  images: {
    unoptimized: false, // Enable optimization in development
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'moai-adk.gooslab.ai',
      },
    ],
  },

  // Bundle optimization - simplified for development
  compiler: {
    removeConsole: false, // Keep console logs in development
    reactRemoveProperties: false,
  },

  webpack: (config, { isServer, dev }) => {
    // Simplified webpack config for development
    // Remove usedExports to avoid cacheUnaffected conflict
    if (dev) {
      config.optimization = {
        ...config.optimization,
        sideEffects: false,
      };
    }

    // Exclude Pagefind from server bundle
    if (isServer) {
      config.externals.push({
        'pagefind': 'commonjs pagefind',
      });
    }

    return config;
  },

  compress: true,
  poweredByHeader: false,

  // Custom headers for better SEO and accessibility
  async headers() {
    return [
      {
        source: '/(.*)',
        headers: [
          {
            key: 'X-Content-Type-Options',
            value: 'nosniff',
          },
          {
            key: 'X-Frame-Options',
            value: 'DENY',
          },
          {
            key: 'X-XSS-Protection',
            value: '1; mode=block',
          },
          {
            key: 'Referrer-Policy',
            value: 'strict-origin-when-cross-origin',
          },
        ],
      },
    ];
  },
};

// Configure Nextra for 4.6.0 - theme is now configured differently
module.exports = withNextra({
  staticImage: true,
  latex: true,
  codeHighlight: true,
  defaultShowCopyCode: true,
  search: {
    codeblocks: false,
  },
})(nextConfig);
