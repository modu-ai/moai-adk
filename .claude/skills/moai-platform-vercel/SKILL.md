---
name: "moai-platform-vercel"
description: "Vercel edge deployment specialist covering Edge Functions, Next.js optimization, preview deployments, ISR, and storage solutions. Use when deploying Next.js applications, implementing edge computing, or configuring Vercel platform features."
version: 2.0.0
category: "platform"
modularized: true
user-invocable: false
tags: ['vercel', 'edge', 'nextjs', 'isr', 'preview', 'cdn', 'kv', 'blob', 'postgres']
context7-libraries: "/vercel/next.js, /vercel/vercel"
related-skills: "moai-platform-railway, moai-lang-typescript, moai-domain-frontend"
updated: 2026-01-08
status: "active"
allowed-tools:
  - Read
  - Write
  - Bash
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# moai-platform-vercel: Vercel Edge Deployment Specialist

## Quick Reference

Vercel Optimization Focus: Edge-first deployment platform with global CDN, Next.js optimized runtime, managed storage solutions, and developer-centric preview workflows.

### Core Capabilities

Edge Functions:
- Global low-latency compute at 30+ edge locations
- Sub-50ms cold start for optimal user experience
- Geo-based routing and personalization
- Edge middleware for request/response transformation

Next.js Optimized Runtime:
- First-class Next.js support with automatic optimizations
- Server Components and App Router integration
- Streaming SSR for improved TTFB
- Built-in image optimization with next/image

Preview Deployments:
- Automatic PR-based preview URLs
- Branch-specific environment variables
- Comment integration for PR reviews
- Instant rollback capabilities

ISR (Incremental Static Regeneration):
- On-demand revalidation for dynamic content
- Stale-while-revalidate caching strategy
- Tag-based cache invalidation
- Background regeneration without user impact

Managed Storage:
- Vercel KV: Redis-compatible key-value store
- Vercel Blob: Object storage for files
- Vercel Postgres: Managed PostgreSQL database

### Quick Decision Guide

Choose Vercel When:
- Next.js is primary framework
- Edge performance is critical requirement
- Preview deployments needed for team collaboration
- Web Vitals monitoring is priority
- Serverless storage solutions needed

---

## Module Index

This skill uses modular documentation for progressive disclosure. Load modules as needed:

### [Edge Functions and Middleware](modules/edge-functions.md)

Covers Edge Runtime configuration, middleware patterns, geo-based content delivery, A/B testing at edge, and authentication at edge. Includes region selection, supported APIs, and cold start optimization.

### [ISR and Caching Strategies](modules/isr-caching.md)

Covers Incremental Static Regeneration patterns, time-based and on-demand revalidation, tag-based cache invalidation, CDN cache headers, streaming with Suspense, and cache debugging techniques.

### [Deployment Configuration](modules/deployment-config.md)

Covers vercel.json configuration, environment variables management, preview and production deployments, function memory and duration settings, cron jobs, monorepo setup, domain configuration, and security headers.

### [Analytics and Speed Insights](modules/analytics-speed.md)

Covers Vercel Analytics integration, Speed Insights for Web Vitals, custom performance monitoring, Core Web Vitals optimization, and third-party analytics integration.

### [KV, Blob, and Postgres Storage](modules/kv-storage.md)

Covers Vercel KV (Redis-compatible), Vercel Blob (object storage), Vercel Postgres (managed database), edge-compatible usage patterns, and common storage patterns.

---

## Implementation Guide

### Essential Configuration

Basic vercel.json:

```json
{
  "$schema": "https://openapi.vercel.sh/vercel.json",
  "framework": "nextjs",
  "regions": ["iad1", "sfo1", "fra1"],
  "functions": {
    "app/api/**/*.ts": {
      "memory": 1024,
      "maxDuration": 30
    }
  }
}
```

### Edge Function Quick Start

```typescript
// app/api/edge/route.ts
export const runtime = 'edge'
export const preferredRegion = ['iad1', 'sfo1']

export async function GET(request: Request) {
  const country = request.geo?.country ?? 'Unknown'
  return Response.json({ country })
}
```

### ISR Quick Start

```typescript
// app/products/[id]/page.tsx
export const revalidate = 60 // Revalidate every 60 seconds

export default async function ProductPage({ params }: { params: { id: string } }) {
  const product = await fetch(`https://api.example.com/products/${params.id}`, {
    next: { tags: [`product-${params.id}`] }
  }).then(r => r.json())

  return <ProductDetail product={product} />
}
```

### Analytics Integration

```typescript
// app/layout.tsx
import { Analytics } from '@vercel/analytics/react'
import { SpeedInsights } from '@vercel/speed-insights/next'

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="en">
      <body>
        {children}
        <Analytics />
        <SpeedInsights />
      </body>
    </html>
  )
}
```

---

## Advanced Patterns

### Blue-Green Deployment

Deploy new version, run smoke tests on preview URL, then switch production alias using Vercel SDK for zero-downtime releases.

### Monorepo with Turborepo

```json
{
  "buildCommand": "cd ../.. && pnpm turbo build --filter=web",
  "installCommand": "cd ../.. && pnpm install",
  "framework": "nextjs"
}
```

### Context7 Integration

Use Context7 MCP tools for latest documentation:

Step 1: Resolve library ID using mcp__context7__resolve-library-id with "vercel" or "next.js"

Step 2: Fetch documentation using mcp__context7__get-library-docs with resolved ID, specific topic, and appropriate token allocation (5000-10000 tokens)

---

## Works Well With

- `moai-platform-railway` - Container-based deployment alternative
- `moai-lang-typescript` - TypeScript patterns for Next.js
- `moai-domain-frontend` - React and Next.js component patterns
- `moai-foundation-quality` - Deployment validation and testing
- `moai-domain-database` - Database patterns for Vercel Postgres

---

## Additional Resources

- [Reference Guide](reference.md) - Complete CLI commands and configuration options
- [Code Examples](examples.md) - Production-ready code snippets

---

Status: Production Ready | Version: 2.0.0 | Updated: 2026-01-06
