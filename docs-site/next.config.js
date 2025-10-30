// @CODE:NEXTRA-SITE-001:CONFIG - Nextra 4.0 configuration for Next.js 14
// @CODE:NEXTRA-I18N-012 - next-intl plugin integration
/**
 * Next.js Configuration for MoAI-ADK Documentation Site
 *
 * This configuration integrates:
 * - Nextra 4.0 documentation theme with Next.js 14 (Pages Router)
 * - next-intl for multilingual support (Korean/English)
 *
 * Key Features:
 * - Static Site Generation (SSG) with output: 'export'
 * - Nextra theme configuration in theme.config.tsx
 * - Internationalization (i18n) with next-intl
 * - Optimized for Vercel deployment
 */

const withNextIntl = require('next-intl/plugin')(
  // Path to request configuration (message loader)
  './i18n/request.ts'
);

module.exports = withNextIntl({
  // Enable React strict mode for better development experience
  reactStrictMode: true,

  // Static export for optimal Vercel deployment
  // This generates a fully static site in the 'out' directory
  output: 'export',

  // Image optimization must be disabled for static export
  images: {
    unoptimized: true
  }
});
