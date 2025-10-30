// @CODE:NEXTRA-CONFIG-001 - Nextra 4.0 configuration for Next.js 14
/**
 * Next.js Configuration for MoAI-ADK Documentation Site
 *
 * This configuration integrates Nextra 4.0 documentation theme with Next.js 14.
 *
 * Key Features:
 * - Static Site Generation (SSG) with output: 'export'
 * - Nextra theme with built-in search and code highlighting
 * - LaTeX support for mathematical expressions
 * - Optimized for Vercel deployment
 */

const withNextra = require('nextra')({
  // Nextra theme configuration
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.tsx',

  // Enable LaTeX support for mathematical expressions
  latex: true,

  // Enable search in code blocks
  search: {
    codeblocks: true
  },

  // Show copy button on code blocks by default
  defaultShowCopyCode: true
});

module.exports = withNextra({
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
