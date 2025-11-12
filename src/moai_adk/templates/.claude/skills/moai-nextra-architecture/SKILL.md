---
name: "moai-nextra-architecture"
version: "4.0.0"
status: stable
description: "Enterprise Skill for advanced development"
allowed-tools: "Read, Bash, WebSearch, WebFetch"
---

# Skill: Nextra Architecture Expert

## Metadata

```yaml
skill_id: moai-nextra-architecture
skill_name: Nextra Architecture Expert
version: 1.0.0
created_date: 2025-11-11
updated_date: 2025-11-11
language: english
word_count: 1800
triggers:
  - keywords: [nextra, nextjs documentation, theme config, mdx, static site generation]
  - contexts: [nextra-architecture, docs-optimization, site-generation, documentation-framework]
agents:
  - docs-manager
  - frontend-expert
  - docs-manager
freedom_level: high
context7_references:
  - url: "https://github.com/shuding/nextra"
    topic: "Nextra framework best practices and configuration"
  - url: "https://vercel.com/templates/next.js"
    topic: "Next.js deployment and optimization strategies"
```

## 📚 Content

### Section 1: Nextra Framework Mastery

#### Core Architecture Principles

**Nextra** is a React-based static site generator built on Next.js that specializes in documentation websites. Key architectural advantages:

- **Zero Config MDX**: Markdown + JSX without build configuration
- **File-system Routing**: Automatic route generation from content structure
- **Performance Optimization**: Automatic code splitting and prefetching
- **Theme System**: Pluggable architecture with customizable themes
- **i18n Support**: Built-in internationalization with automatic locale detection

#### Essential Configuration Patterns

**theme.config.tsx** (Core Configuration):
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

**next.config.js** (Build Optimization):
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
      // Custom remark plugins
      require('remark-gfm'),
      require('remark-footnotes'),
    ],
    rehypePlugins: [
      // Custom rehype plugins
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

### Section 2: Advanced Content Architecture

#### Directory Structure Design

**Optimal Nextra Project Structure**:
```
docs/
├── pages/                          # Route definitions (optional)
│   ├── _meta.json                  # Site metadata
│   └── api/                        # Custom API routes
├── public/                         # Static assets
│   ├── images/                     # Images and icons
│   ├── files/                      # Downloadable files
│   └── favicon.ico
├── components/                     # React components
│   ├── ui/                         # Reusable UI components
│   ├── diagrams/                   # Custom diagram components
│   └── interactive/                # Interactive elements
├── lib/                           # Utility functions
│   ├── nextra-config.ts           # Nextra configuration helpers
│   ├── content-utils.ts           # Content processing utilities
│   └── search-client.ts           # Custom search implementation
├── styles/                        # Custom styles
│   ├── globals.css               # Global CSS
│   └── components.css            # Component-specific styles
├── content/                       # Documentation content
│   ├── index.mdx                 # Homepage
│   ├── getting-started/          # Getting started guides
│   ├── guides/                   # How-to guides
│   ├── reference/                # API reference
│   └── tutorials/                # Step-by-step tutorials
├── scripts/                       # Build and utility scripts
│   ├── build-search-index.js     # Search index generation
│   ├── validate-links.js         # Link validation
│   └── generate-sitemap.js       # Sitemap generation
├── types/                        # TypeScript definitions
│   ├── content.d.ts              # Content type definitions
│   └── config.d.ts               # Configuration types
├── .nextra/                      # Nextra cache (auto-generated)
├── package.json
├── tsconfig.json
└── README.md
```

#### Content Type Patterns

**Homepage (content/index.mdx)**:
```mdx
---
title: Welcome to My Documentation
description: Comprehensive guide for getting started with our platform
---

import { Callout } from 'nextra-theme-docs'
import { Tabs, TabItem } from 'nextra-theme-docs'

# Welcome to My Platform

<Callout type="info" emoji="🎯">
  This documentation will help you get up and running in minutes.
</Callout>

## Quick Start

<Tabs items={['npm', 'yarn', 'pnpm']}>
  <TabItem>
    ```bash
    npm install my-platform
    ```
  </TabItem>
  <TabItem>
    ```bash
    yarn add my-platform
    ```
  </TabItem>
  <TabItem>
    ```bash
    pnpm install my-platform
    ```
  </TabItem>
</Tabs>

## Why Choose Our Platform?

- 🚀 **Performance**: Built with modern web technologies
- 🎨 **Customizable**: Extensive theming and configuration options
- 📱 **Mobile-First**: Responsive design for all devices
- 🔍 **Searchable**: Built-in search functionality
- ♿ **Accessible**: WCAG 2.1 compliant

<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 mt-8">
  <div className="p-6 border rounded-lg">
    <h3 className="text-lg font-semibold mb-2">📚 Documentation</h3>
    <p>Comprehensive guides and API references.</p>
  </div>
  <div className="p-6 border rounded-lg">
    <h3 className="text-lg font-semibold mb-2">🛠️ Tools</h3>
    <p>Developer tools and utilities.</p>
  </div>
  <div className="p-6 border rounded-lg">
    <h3 className="text-lg font-semibold mb-2">🤝 Community</h3>
    <p>Join our community for support.</p>
  </div>
</div>
```

**Guide Page Pattern (content/guides/example.mdx)**:
```mdx
---
title: Complete Guide to Feature X
description: Learn how to implement and customize Feature X
---

import { Steps, Step } from 'nextra-theme-docs'
import { FileTree } from 'nextra/components'

# Feature X Guide

<Callout type="warning" emoji="⚠️">
  This feature requires version 2.0.0 or higher.
</Callout>

## Prerequisites

Before you begin, ensure you have:

- Node.js 18+ installed
- Basic understanding of React
- Project configured for TypeScript

## Implementation Steps

<Steps>
  <Step>
    ### Step 1: Installation

    Install the required dependencies:

    ```bash
    npm install feature-x-package
    ```

    Update your configuration:

    <FileTree>
      <FileTree.File name="config.json" active>
        {`{
          "featureX": {
            "enabled": true,
            "options": {
              "theme": "default"
            }
          }
        }}`}
      </FileTree.File>
    </FileTree>
  </Step>

  <Step>
    ### Step 2: Configuration

    Configure the feature according to your needs:

    ```typescript showLineNumbers
    // src/config/featureX.ts
    export const featureXConfig = {
      mode: 'production',
      plugins: ['analytics', 'seo'],
      theme: {
        primary: '#0070f3',
        secondary: '#6f42c1'
      }
    }
    ```

    <Callout type="tip" emoji="💡">
      You can customize the theme using CSS variables.
    </Callout>
  </Step>

  <Step>
    ### Step 3: Usage

    Now you can use the feature in your application:

    ```tsx
    import { FeatureX } from 'feature-x-package'

    export default function MyComponent() {
      return (
        <FeatureX
          config={featureXConfig}
          onReady={() => console.log('Ready!')}
        />
      )
    }
    ```
  </Step>
</Steps>

## Advanced Configuration

For more advanced use cases, refer to the [API Reference](/reference/feature-x).
```

### Section 3: Performance Optimization

#### Build Performance

**Optimizing Build Speed**:
```typescript
// scripts/optimize-build.js
const { execSync } = require('child_process')
const fs = require('fs')

// Pre-build optimizations
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

// Post-build optimizations
function optimizeOutput() {
  // Generate sitemap
  execSync('node scripts/generate-sitemap.js', { stdio: 'inherit' })

  // Compress static assets
  execSync('gzip -k -r out/', { stdio: 'inherit' })

  // Generate Lighthouse report
  execSync('lighthouse http://localhost:3000 --output=json --output-path=./lighthouse-report.json', {
    stdio: 'inherit'
  })
}
```

**Search Optimization**:
```typescript
// lib/search-client.ts
import type { PagefindResult } from './types'

export class SearchClient {
  private searchIndex: Map<string, PagefindResult> = new Map()

  async initialize() {
    // Load pre-built search index
    const response = await fetch('/search-index.json')
    const index = await response.json()

    this.searchIndex = new Map(
      index.map((item: PagefindResult) => [item.url, item])
    )
  }

  async search(query: string, limit = 10) {
    const results = Array.from(this.searchIndex.values())
      .filter(item =>
        item.title.toLowerCase().includes(query.toLowerCase()) ||
        item.content.toLowerCase().includes(query.toLowerCase())
      )
      .slice(0, limit)

    return {
      results,
      total: results.length
    }
  }
}
```

#### Runtime Performance

**Client-Side Optimizations**:
```typescript
// lib/performance.ts
export class PerformanceOptimizer {
  // Lazy load images
  static optimizeImages() {
    if ('IntersectionObserver' in window) {
      const imageObserver = new IntersectionObserver((entries) => {
        entries.forEach(entry => {
          if (entry.isIntersecting) {
            const img = entry.target as HTMLImageElement
            if (img.dataset.src) {
              img.src = img.dataset.src
              imageObserver.unobserve(img)
            }
          }
        })
      })

      document.querySelectorAll('img[data-src]').forEach(img => {
        imageObserver.observe(img)
      })
    }
  }

  // Prefetch critical resources
  static prefetchCriticalResources() {
    const criticalPaths = ['/api/content', '/search-index.json']

    criticalPaths.forEach(path => {
      const link = document.createElement('link')
      link.rel = 'prefetch'
      link.href = path
      document.head.appendChild(link)
    })
  }

  // Optimize scroll performance
  static optimizeScrolling() {
    let ticking = false

    window.addEventListener('scroll', () => {
      if (!ticking) {
        requestAnimationFrame(() => {
          // Update scroll-based elements
          this.updateScrollPosition()
          ticking = false
        })
        ticking = true
      }
    })
  }
}
```

### Section 4: Mobile Optimization

#### Responsive Design Patterns

**Mobile-First Layout Components**:
```tsx
// components/ui/MobileLayout.tsx
import { useState } from 'react'
import { useRouter } from 'next/router'

export function MobileLayout({ children }) {
  const [sidebarOpen, setSidebarOpen] = useState(false)
  const router = useRouter()

  return (
    <div className="min-h-screen bg-background">
      {/* Mobile Header */}
      <header className="sticky top-0 z-50 border-b bg-background md:hidden">
        <div className="flex h-14 items-center px-4">
          <button
            onClick={() => setSidebarOpen(true)}
            className="mr-4 p-2 rounded-md hover:bg-accent"
            aria-label="Open sidebar"
          >
            <svg className="h-5 w-5" fill="none" stroke="currentColor">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2}
                d="M4 6h16M4 12h16M4 18h16" />
            </svg>
          </button>
          <h1 className="text-lg font-semibold">Documentation</h1>
        </div>
      </header>

      {/* Mobile Sidebar Overlay */}
      {sidebarOpen && (
        <div className="fixed inset-0 z-50 md:hidden">
          <div
            className="fixed inset-0 bg-black/50"
            onClick={() => setSidebarOpen(false)}
          />
          <div className="fixed left-0 top-0 h-full w-64 bg-background border-r">
            {/* Sidebar content here */}
          </div>
        </div>
      )}

      {/* Main Content */}
      <main className="flex-1 md:flex">
        {/* Desktop Sidebar (always visible) */}
        <aside className="hidden md:block w-64 border-r">
          {/* Desktop sidebar content */}
        </aside>

        {/* Content Area */}
        <div className="flex-1 px-4 py-6 md:px-8 lg:px-12">
          <div className="max-w-4xl mx-auto">
            {children}
          </div>
        </div>
      </main>
    </div>
  )
}
```

**Touch-Friendly Interactions**:
```css
/* styles/mobile.css}
@media (max-width: 768px) {
  /* Larger tap targets */
  button, a, input, select, textarea {
    min-height: 44px;
    min-width: 44px;
    padding: 12px;
  }

  /* Improved spacing */
  .mobile-spacing > * + * {
    margin-top: 1.5rem;
  }

  /* Touch-friendly navigation */
  .mobile-nav {
    touch-action: manipulation;
    -webkit-tap-highlight-color: transparent;
  }

  /* Readable font sizes */
  body {
    font-size: 16px;
    line-height: 1.6;
  }

  /* Improved readability */
  .content {
    max-width: 100%;
    padding: 0 1rem;
  }

  /* Horizontal scroll prevention */
  .no-overflow {
    overflow-x: hidden;
  }
}
```

### Section 5: Accessibility Implementation

#### WCAG 2.1 Compliance

**Semantic HTML Structure**:
```tsx
// components/ui/AccessibleContent.tsx
export function AccessibleContent({ title, children }) {
  return (
    <main id="main-content" role="main" aria-labelledby="page-title">
      <header>
        <h1 id="page-title" className="sr-only">
          {title}
        </h1>
      </header>

      <div className="prose prose-lg max-w-none">
        {children}
      </div>

      {/* Skip to main content link */}
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:absolute focus:top-4 focus:left-4
                   bg-primary text-primary-foreground px-4 py-2 rounded-md"
      >
        Skip to main content
      </a>
    </main>
  )
}
```

**Keyboard Navigation**:
```tsx
// components/ui/KeyboardNavigation.tsx
import { useEffect } from 'react'

export function KeyboardNavigation() {
  useEffect(() => {
    const handleKeyDown = (event: KeyboardEvent) => {
      // Focus trap for modals
      if (event.key === 'Tab') {
        const focusableElements = document.querySelectorAll(
          'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
        )
        const firstElement = focusableElements[0] as HTMLElement
        const lastElement = focusableElements[focusableElements.length - 1] as HTMLElement

        if (event.shiftKey && document.activeElement === firstElement) {
          event.preventDefault()
          lastElement.focus()
        } else if (!event.shiftKey && document.activeElement === lastElement) {
          event.preventDefault()
          firstElement.focus()
        }
      }

      // Escape key to close modals
      if (event.key === 'Escape') {
        const modal = document.querySelector('[role="dialog"]')
        if (modal) {
          const closeButton = modal.querySelector('button[aria-label*="Close"]')
          closeButton?.dispatchEvent(new MouseEvent('click'))
        }
      }
    }

    document.addEventListener('keydown', handleKeyDown)
    return () => document.removeEventListener('keydown', handleKeyDown)
  }, [])

  return null
}
```

### Section 6: SEO Optimization

#### Meta Tag Configuration

**Dynamic SEO Implementation**:
```typescript
// lib/seo.ts
import type { Metadata } from 'next'

export function generateSEOMetadata({
  title,
  description,
  path = '',
  image,
  keywords = [],
  publishedTime,
  modifiedTime
}: {
  title: string
  description: string
  path?: string
  image?: string
  keywords?: string[]
  publishedTime?: string
  modifiedTime?: string
}): Metadata {
  const baseUrl = 'https://docs.example.com'
  const url = `${baseUrl}${path}`

  return {
    title,
    description,
    keywords: keywords.join(', '),
    authors: [{ name: 'Documentation Team' }],
    creator: 'Documentation Platform',
    publisher: 'Your Company',

    openGraph: {
      type: 'article',
      locale: 'en_US',
      url,
      title,
      description,
      siteName: 'Documentation',
      images: image ? [{
        url: image,
        width: 1200,
        height: 630,
        alt: title,
      }] : [],
    },

    twitter: {
      card: 'summary_large_image',
      title,
      description,
      images: image ? [image] : [],
    },

    alternates: {
      canonical: url,
      languages: {
        'en': `${baseUrl}/en${path}`,
        'ko': `${baseUrl}/ko${path}`,
        'ja': `${baseUrl}/ja${path}`,
      },
    },

    robots: {
      index: true,
      follow: true,
      googleBot: {
        index: true,
        follow: true,
        'max-video-preview': -1,
        'max-image-preview': 'large',
        'max-snippet': -1,
      },
    },
  }
}
```

## 🎯 Usage

### From Agents

```python
# docs-manager agent
Skill("moai-nextra-architecture")

# Analyze project structure
architecture = NextraArchitecture()
structure_analysis = architecture.analyze_source_project("./src")

# Design optimal Nextra configuration
config = architecture.design_theme_config(structure_analysis)
directory_structure = architecture.design_content_structure(structure_analysis)

# Generate configuration files
architecture.generate_theme_config(config)
architecture.create_directory_structure(directory_structure)
```

### Integration Examples

```bash
# Initialize Nextra documentation project
npx nextra-init docs --theme docs --typescript

# Generate optimized configuration
node scripts/generate-nextra-config.js --source ./src --output ./docs

# Build and validate documentation
npm run build:docs
npm run validate:docs
npm run test:docs
```

## 📚 Reference Materials

- [Nextra Official Documentation](https://nextra.site/)
- [Next.js Performance Optimization](https://nextjs.org/docs/advanced-features/measuring-performance)
- [MDN Web Accessibility](https://developer.mozilla.org/en-US/docs/Web/Accessibility)
- [Google SEO Starter Guide](https://developers.google.com/search/docs/fundamentals/seo-starter-guide)

## ✅ Validation Checklist

- [x] Complete Nextra configuration patterns documented
- [x] Performance optimization strategies included
- [x] Mobile-first responsive design examples provided
- [x] WCAG 2.1 accessibility implementation shown
- [x] SEO optimization best practices integrated
- [x] Real-world usage examples included
- [x] TypeScript configuration templates provided
- [x] Build optimization strategies documented