---
name: content-architecture-optimization
parent: moai-library-nextra
description: Content structure and performance optimization
---

# Module 2: Content Architecture & Optimization

## Directory Structure Design

### Optimal Nextra Project Structure

```
docs/
├── pages/                          # Route definitions
│   ├── _meta.json                  # Site metadata
│   └── api/                        # Custom API routes
├── public/                         # Static assets
│   ├── images/
│   ├── files/
│   └── favicon.ico
├── components/                     # React components
│   ├── ui/
│   ├── diagrams/
│   └── interactive/
├── lib/                           # Utility functions
│   ├── nextra-config.ts
│   ├── content-utils.ts
│   └── search-client.ts
├── styles/
│   ├── globals.css
│   └── components.css
├── content/                       # Documentation content
│   ├── index.mdx
│   ├── getting-started/
│   ├── guides/
│   ├── reference/
│   └── tutorials/
├── scripts/
│   ├── build-search-index.js
│   ├── validate-links.js
│   └── generate-sitemap.js
└── types/
    ├── content.d.ts
    └── config.d.ts
```

## Content Type Patterns

### Homepage Pattern

```mdx
---
title: Welcome to My Documentation
description: Comprehensive guide for getting started
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
```

## Performance Optimization

### Search Optimization

```typescript
// lib/search-client.ts
export class SearchClient {
  private searchIndex: Map<string, PagefindResult> = new Map()

  async initialize() {
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

### Client-Side Optimization

```typescript
// lib/performance.ts
export class PerformanceOptimizer {
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

  static prefetchCriticalResources() {
    const criticalPaths = ['/api/content', '/search-index.json']

    criticalPaths.forEach(path => {
      const link = document.createElement('link')
      link.rel = 'prefetch'
      link.href = path
      document.head.appendChild(link)
    })
  }
}
```

---

**Reference**: [Next.js Performance](https://nextjs.org/docs/advanced-features/measuring-performance)
