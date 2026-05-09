---
name: configuration
parent: moai-library-nextra
description: Complete Nextra theme.config.tsx reference and configuration patterns
---

# Nextra Theme Configuration Guide

## Overview

Complete reference for configuring Nextra documentation sites using theme.config.tsx (Nextra 3.x) or Layout props (Nextra 4.x with App Router).

## Nextra 3.x Theme Configuration (Pages Router)

### Complete theme.config.tsx Reference

```typescript
// theme.config.tsx
import { DocsThemeConfig } from 'nextra-theme-docs'

const config: DocsThemeConfig = {
  // Branding
  logo: <span className="font-bold">My Documentation</span>,
  logoLink: '/',

  // Project Links
  project: {
    link: 'https://github.com/username/project',
    icon: <GitHubIcon />,  // Optional custom icon
  },

  // Chat/Discord Link
  chat: {
    link: 'https://discord.gg/your-server',
    icon: <DiscordIcon />,
  },

  // Repository Base for Edit Links
  docsRepositoryBase: 'https://github.com/username/project/tree/main/docs',

  // SEO Configuration
  useNextSeoProps() {
    return {
      titleTemplate: '%s - My Documentation',
    }
  },

  // Head Elements
  head: (
    <>
      <meta name="viewport" content="width=device-width, initial-scale=1.0" />
      <meta property="og:title" content="My Documentation" />
      <meta property="og:description" content="Comprehensive project documentation" />
      <link rel="icon" href="/favicon.ico" />
    </>
  ),

  // Primary Color Hue
  primaryHue: {
    dark: 210,
    light: 220,
  },

  // Sidebar Configuration
  sidebar: {
    defaultMenuCollapseLevel: 1,
    toggleButton: true,
    autoCollapse: false,
  },

  // Table of Contents
  toc: {
    title: 'On This Page',
    float: true,
    backToTop: true,
    extraContent: null,
  },

  // Navigation
  navigation: {
    prev: true,
    next: true,
  },

  // Edit Link
  editLink: {
    text: 'Edit this page on GitHub',
  },

  // Feedback
  feedback: {
    content: 'Question? Give us feedback',
    labels: 'feedback',
    useLink: () => 'https://github.com/username/project/issues/new',
  },

  // Footer
  footer: {
    text: (
      <span>
        MIT {new Date().getFullYear()} - My Project
      </span>
    ),
  },

  // Dark Mode
  darkMode: true,

  // Theme Switch Labels
  themeSwitch: {
    useOptions() {
      return {
        light: 'Light',
        dark: 'Dark',
        system: 'System',
      }
    },
  },

  // Git Timestamp
  gitTimestamp: function GitTimestamp({ timestamp }) {
    return (
      <span>
        Last updated: {timestamp.toLocaleDateString()}
      </span>
    )
  },

  // Search Configuration
  search: {
    placeholder: 'Search documentation...',
    loading: 'Loading...',
    emptyResult: 'No results found',
    error: 'Search failed',
  },

  // Navbar Extra Content
  navbar: {
    extraContent: null,
  },

  // Banner
  banner: {
    key: 'release-banner',
    text: 'Version 2.0 is now available!',
    dismissible: true,
  },

  // Internationalization
  i18n: [
    { locale: 'en', text: 'English' },
    { locale: 'ko', text: 'Korean' },
    { locale: 'ja', text: 'Japanese' },
  ],
}

export default config
```

## Nextra 4.x Configuration (App Router)

### Layout Component Props

```typescript
// app/layout.tsx
import { Layout } from 'nextra-theme-docs'
import { getPageMap } from 'nextra/page-map'
import { Banner, Navbar, Footer, Search } from '@/components'

export default async function RootLayout({ children }) {
  const pageMap = await getPageMap()

  return (
    <html lang="en">
      <body>
        <Layout
          pageMap={pageMap}
          banner={<Banner />}
          navbar={<Navbar />}
          footer={<Footer />}
          search={<Search />}
          docsRepositoryBase="https://github.com/username/project/tree/main"
          darkMode={true}
          editLink="Edit this page on GitHub"
          feedback={{
            content: 'Question? Give us feedback',
            labels: 'feedback',
          }}
          i18n={[
            { locale: 'en', name: 'English' },
            { locale: 'ko', name: 'Korean' },
          ]}
          lastUpdated={<LastUpdated />}
          navigation={{ prev: true, next: true }}
          sidebar={{
            autoCollapse: false,
            defaultMenuCollapseLevel: 2,
            defaultOpen: true,
            toggleButton: true,
          }}
          toc={{
            title: 'On This Page',
            float: true,
            backToTop: 'Back to top',
            extraContent: null,
          }}
          themeSwitch={{
            dark: 'Dark',
            light: 'Light',
            system: 'System',
          }}
          nextThemes={{
            attribute: 'class',
            defaultTheme: 'system',
            disableTransitionOnChange: false,
          }}
        >
          {children}
        </Layout>
      </body>
    </html>
  )
}
```

## Configuration Patterns by Use Case

### Minimal Documentation Site

```typescript
const config: DocsThemeConfig = {
  logo: <span>Docs</span>,
  project: { link: 'https://github.com/user/repo' },
  docsRepositoryBase: 'https://github.com/user/repo/tree/main',
}
```

### Enterprise Documentation

```typescript
const config: DocsThemeConfig = {
  logo: <CompanyLogo />,
  logoLink: '/',
  project: { link: 'https://github.com/company/product' },
  docsRepositoryBase: 'https://github.com/company/product/tree/main/docs',

  sidebar: {
    defaultMenuCollapseLevel: 2,
    toggleButton: true,
    autoCollapse: true,
  },

  toc: {
    float: true,
    backToTop: true,
  },

  search: {
    placeholder: 'Search...',
  },

  footer: {
    text: <CompanyFooter />,
  },

  i18n: [
    { locale: 'en', text: 'English' },
    { locale: 'de', text: 'Deutsch' },
    { locale: 'fr', text: 'Francais' },
    { locale: 'ja', text: 'Japanese' },
  ],

  primaryHue: { dark: 200, light: 210 },
  darkMode: true,
}
```

### API Documentation

```typescript
const config: DocsThemeConfig = {
  logo: <span>API Reference</span>,

  // Wider content area for code examples
  toc: {
    float: false,
  },

  // Custom code block styling
  components: {
    pre: CustomCodeBlock,
  },

  // Version selector in navbar
  navbar: {
    extraContent: <VersionSelector />,
  },
}
```

## next.config.js Setup

### Nextra 3.x Configuration

```javascript
// next.config.js
const withNextra = require('nextra')({
  theme: 'nextra-theme-docs',
  themeConfig: './theme.config.tsx',
  staticImage: true,
  flexSearch: true,
  defaultShowCopyCode: true,
})

module.exports = withNextra({
  reactStrictMode: true,
})
```

### Nextra 4.x with Next.js 15

```javascript
// next.config.mjs
import nextra from 'nextra'

const withNextra = nextra({
  // Nextra options
})

export default withNextra({
  // Next.js 15 options
  experimental: {
    turbopack: true,
  },
})
```

## Version Compatibility

Current stable versions as of 2025:
- Nextra 3.x: Compatible with Next.js 13.x and 14.x (Pages Router)
- Nextra 4.x: Compatible with Next.js 14.x and 15.x (App Router, Turbopack support)

Migration from Nextra 3 to 4 requires converting from theme.config.tsx to Layout component props.

---

Last Updated: 2025-12-30
Status: Production Ready
Reference: https://nextra.site/docs/docs-theme/theme-configuration
