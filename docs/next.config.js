const withNextra = require('nextra')({
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.tsx',
  staticImage: true,
  latex: true,
  flexsearch: {
    codeblocks: true,
  },
  codeHighlight: true,
})

module.exports = withNextra({
  reactStrictMode: true,
  swcMinify: true,

  // Image optimization
  images: {
    unoptimized: false,
    formats: ['image/webp', 'image/avif'],
    domains: ['cdn.example.com'],
  },

  // Redirects for better navigation
  redirects: async () => [
    {
      source: '/docs',
      destination: '/getting-started',
      permanent: true,
    },
  ],

  // Performance optimizations
  compress: true,
  poweredByHeader: false,

  // TypeScript
  typescript: {
    tsconfigPath: './tsconfig.json',
  },

  // Experimental features
  experimental: {
    turbotrace: {
      logLevel: 'error',
      logDetail: false,
    },
  },
})
