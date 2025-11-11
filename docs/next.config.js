const nextra = require("nextra");
const withNextra = nextra.default || nextra;

/** @type {import('next').NextConfig} */
const nextConfig = {
  // Static export configuration
  output: 'export',
  trailingSlash: true,

  // Image optimization (compatible with static export)
  images: {
    // For static export, we need unoptimized: true
    // but we'll handle optimization through our custom component
    unoptimized: true,
    // Configure remote patterns for external images
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'moai-adk.gooslab.ai',
      },
    ],
  },

  // Bundle optimization
  compiler: {
    // Remove console logs in production
    removeConsole: process.env.NODE_ENV === 'production',
    // Optimize React components
    reactRemoveProperties: process.env.NODE_ENV === 'production',
  },

  // Webpack optimizations
  webpack: (config, { isServer, dev }) => {
    // Enable webpack bundle analyzer in development
    if (dev) {
      config.optimization = {
        ...config.optimization,
        usedExports: true,
        sideEffects: false,
      };
    }

    // Optimize bundle splitting
    config.optimization.splitChunks = {
      chunks: 'all',
      cacheGroups: {
        vendor: {
          test: /[\\/]node_modules[\\/]/,
          name: 'vendors',
          chunks: 'all',
          priority: 10,
        },
        common: {
          name: 'common',
          minChunks: 2,
          chunks: 'all',
          priority: 5,
        },
      },
    };

    // Exclude Pagefind from server bundle
    if (isServer) {
      config.externals.push({
        'pagefind': 'commonjs pagefind',
      });
    }

    return config;
  },

  // Compression
  compress: true,

  // Performance monitoring
  poweredByHeader: false,

  // Security headers
  headers: async () => [
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
      ],
    },
    {
      source: '/_next/static/(.*)',
      headers: [
        {
          key: 'Cache-Control',
          value: 'public, max-age=31536000, immutable',
        },
      ],
    },
    {
      source: '/pagefind/(.*)',
      headers: [
        {
          key: 'Cache-Control',
          value: 'public, max-age=86400',
        },
      ],
    },
  ],

  // Rewrites for Pagefind search
  async rewrites() {
    return [
      {
        source: '/pagefind/:path*',
        destination: '/pagefind/:path*',
      },
    ];
  },
};

module.exports = withNextra({
  staticImage: true,
  latex: true,
  codeHighlight: true,
  // Nextra 4.x compatible options
  defaultShowCopyCode: true,
  search: {
    // Use custom Pagefind search
    codeblocks: false,
  },
})(nextConfig);