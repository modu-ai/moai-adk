const withNextra = require('nextra')({
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.tsx',
  staticImage: true,
  latex: true,
  codeHighlight: true,
})

module.exports = withNextra({
  // Internationalization (i18n) configuration
  i18n: {
    locales: ['ko', 'en', 'ja', 'zh'],
    defaultLocale: 'ko',
  },

  // Image optimization
  images: {
    unoptimized: false,
  },

  // Redirects for root path
  async redirects() {
    return [
      {
        source: '/',
        destination: '/ko',
        permanent: false,
      },
    ]
  },

  // Headers for cache and security
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
            value: 'SAMEORIGIN',
          },
          {
            key: 'X-XSS-Protection',
            value: '1; mode=block',
          },
        ],
      },
    ]
  },

  reactStrictMode: true,
})
