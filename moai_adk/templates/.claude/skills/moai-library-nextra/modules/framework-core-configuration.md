---
name: framework-core-configuration
parent: moai-library-nextra
description: Nextra framework setup and theme configuration
---

# Module 1: Framework Core & Configuration

## Core Architecture Principles

Nextra is a React-based static site generator built on Next.js that specializes in documentation websites. Key architectural advantages:

- Zero Config MDX: Markdown + JSX without build configuration
- File-system Routing: Automatic route generation from content structure
- Performance Optimization: Automatic code splitting and prefetching
- Theme System: Pluggable architecture with customizable themes
- i18n Support: Built-in internationalization with automatic locale detection

## Essential Configuration Patterns

### theme.config.tsx (Core Configuration)

```typescript
import { DocsThemeConfig } from 'nextra-theme-docs'

const config: DocsThemeConfig = {
 logo: <span>My Documentation</span>,
 project: {
 link: 'https://github.com/username/project',
 },
 docsRepositoryBase: 'https://github.com/username/project/tree/main',
 useNextSeoProps: true,
 head: (
 <>
 <meta name="viewport" content="width=device-width, initial-scale=1.0" />
 <meta property="og:title" content="My Documentation" />
 <meta property="og:description" content="Comprehensive documentation" />
 </>
 ),
 footer: {
 text: 'Built with Nextra and Next.js',
 },
 sidebar: {
 defaultMenuCollapseLevel: 1,
 toggleButton: true,
 },
 toc: {
 backToTop: true,
 extraContent: (
 <div style={{ fontSize: '0.8em', marginTop: '1em' }}>
 Last updated: {new Date().toLocaleDateString()}
 </div>
 ),
 },
 editLink: {
 text: 'Edit this page on GitHub →',
 },
 feedback: {
 content: 'Question? Give us feedback →',
 labels: 'feedback',
 },
 navigation: {
 prev: true,
 next: true,
 },
 darkMode: true,
 themeSwitch: {
 component: <ThemeSwitch />,
 useOptions() {
 return {
 light: 'Light',
 dark: 'Dark',
 system: 'System',
 }
 },
 },
}

export default config
```

### next.config.js (Build Optimization)

```javascript
const withNextra = require('nextra')({
 theme: 'nextra-theme-docs',
 themeConfig: './theme.config.tsx',
 // Performance optimizations
 staticImage: true,
 flexibleSearch: true,
 defaultShowCopyCode: true,
 readingTime: true,
 // MDX configuration
 mdxOptions: {
 remarkPlugins: [
 require('remark-gfm'),
 require('remark-footnotes'),
 ],
 rehypePlugins: [
 require('rehype-slug'),
 require('rehype-autolink-headings'),
 ],
 },
})

module.exports = withNextra({
 // Next.js optimizations
 experimental: {
 appDir: true,
 serverComponentsExternalPackages: ['nextra'],
 },
 images: {
 domains: ['example.com'],
 formats: ['image/webp', 'image/avif'],
 },
 // Build performance
 swcMinify: true,
 compiler: {
 removeConsole: process.env.NODE_ENV === 'production',
 },
 // Static export configuration
 output: 'export',
 trailingSlash: true,
 images: {
 unoptimized: true,
 },
})
```

## Custom Theme Development

```typescript
// components/custom-theme.tsx
import type { DocsThemeConfig } from 'nextra-theme-docs'

export const customTheme: DocsThemeConfig = {
 // Custom components
 components: {
 h1: (props) => <h1 className="custom-heading" {...props} />,
 pre: (props) => <CodeBlock {...props} />,
 },
 
 // Custom CSS
 primaryHue: {
 dark: 210,
 light: 210
 },
 
 // Custom navigation
 navbar: {
 extraContent: () => <CustomNavButtons />
 }
}
```

## Build Optimization

```typescript
// scripts/optimize-build.js
const { execSync } = require('child_process')

function optimizeBuild() {
 // Clear Next.js cache
 execSync('rm -rf .next', { stdio: 'inherit' })

 // Generate search index incrementally
 execSync('node scripts/build-search-index.js --incremental', {
 stdio: 'inherit'
 })

 // Optimize images
 execSync('node scripts/optimize-images.js', { stdio: 'inherit' })
}

function optimizeOutput() {
 // Generate sitemap
 execSync('node scripts/generate-sitemap.js', { stdio: 'inherit' })

 // Compress static assets
 execSync('gzip -k -r out/', { stdio: 'inherit' })
}
```

---

Reference: [Nextra Official Documentation](https://nextra.site/)
