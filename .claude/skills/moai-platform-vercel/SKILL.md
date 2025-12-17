---
name: moai-platform-vercel
description: Vercel edge deployment specialist covering Edge Functions, Next.js optimization, preview deployments, and ISR. Use when deploying Next.js or edge-first applications.
version: 1.0.0
category: platform
tags: vercel, edge, nextjs, isr, preview, cdn
context7-libraries: /vercel/next.js, /vercel/vercel
related-skills: moai-platform-railway, moai-lang-typescript
updated: 2025-12-07
status: active
allowed-tools: Read, Write, Bash, Grep, Glob
---

# moai-platform-vercel: Vercel Edge Deployment Specialist

## Quick Reference (30 seconds)

Vercel Optimization Focus: Edge-first deployment platform with global CDN, Next.js optimized runtime, and developer-centric preview workflows.

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

### Quick Decision Guide

Choose Vercel When:
- Next.js is primary framework
- Edge performance is critical requirement
- Preview deployments needed for team collaboration
- Web Vitals monitoring is priority

---

## Implementation Guide

### Phase 1: Edge Functions Architecture

Edge Runtime Configuration:
```typescript
// app/api/edge-handler/route.ts
export const runtime = 'edge'
export const preferredRegion = ['iad1', 'sfo1', 'fra1']

export async function GET(request: Request) {
  const { geo, ip } = request

  return Response.json({
    country: geo?.country ?? 'Unknown',
    city: geo?.city ?? 'Unknown',
    region: geo?.region ?? 'Unknown',
    ip: ip ?? 'Unknown',
    timestamp: new Date().toISOString()
  })
}
```

Edge Function Patterns:

Geo-Based Content Delivery:
```typescript
// app/api/localized/route.ts
export const runtime = 'edge'

const CONTENT_BY_REGION: Record<string, { currency: string; locale: string }> = {
  US: { currency: 'USD', locale: 'en-US' },
  DE: { currency: 'EUR', locale: 'de-DE' },
  JP: { currency: 'JPY', locale: 'ja-JP' },
  KR: { currency: 'KRW', locale: 'ko-KR' }
}

export async function GET(request: Request) {
  const country = request.geo?.country ?? 'US'
  const config = CONTENT_BY_REGION[country] ?? CONTENT_BY_REGION.US

  return Response.json(config, {
    headers: {
      'Cache-Control': 'public, s-maxage=3600',
      'CDN-Cache-Control': 'public, max-age=86400'
    }
  })
}
```

A/B Testing at Edge:
```typescript
// middleware.ts
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export function middleware(request: NextRequest) {
  const existingBucket = request.cookies.get('ab-bucket')?.value

  if (!existingBucket) {
    const bucket = Math.random() < 0.5 ? 'control' : 'variant'
    const response = NextResponse.next()
    response.cookies.set('ab-bucket', bucket, {
      maxAge: 60 * 60 * 24 * 30, // 30 days
      httpOnly: true,
      sameSite: 'lax'
    })
    return response
  }

  return NextResponse.next()
}

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)']
}
```

### Phase 2: Next.js Optimization Patterns

ISR Implementation:
```typescript
// app/products/[id]/page.tsx
import { notFound } from 'next/navigation'

// Revalidate every 60 seconds
export const revalidate = 60

// Generate static params for top products
export async function generateStaticParams() {
  const products = await fetch('https://api.example.com/products/top').then(r => r.json())
  return products.map((p: { id: string }) => ({ id: p.id }))
}

// Dynamic metadata for SEO
export async function generateMetadata({ params }: { params: { id: string } }) {
  const product = await fetchProduct(params.id)
  if (!product) return {}

  return {
    title: product.name,
    description: product.description,
    openGraph: {
      images: [product.imageUrl]
    }
  }
}

export default async function ProductPage({ params }: { params: { id: string } }) {
  const product = await fetchProduct(params.id)
  if (!product) notFound()

  return <ProductDetail product={product} />
}

async function fetchProduct(id: string) {
  const res = await fetch(`https://api.example.com/products/${id}`, {
    next: { tags: [`product-${id}`] }
  })
  if (!res.ok) return null
  return res.json()
}
```

On-Demand Revalidation:
```typescript
// app/api/revalidate/route.ts
import { revalidateTag, revalidatePath } from 'next/cache'
import { NextRequest } from 'next/server'

export async function POST(request: NextRequest) {
  const { tag, path, secret } = await request.json()

  // Validate webhook secret
  if (secret !== process.env.REVALIDATION_SECRET) {
    return Response.json({ error: 'Invalid secret' }, { status: 401 })
  }

  try {
    if (tag) {
      revalidateTag(tag)
      return Response.json({ revalidated: true, tag })
    }

    if (path) {
      revalidatePath(path)
      return Response.json({ revalidated: true, path })
    }

    return Response.json({ error: 'Missing tag or path' }, { status: 400 })
  } catch (error) {
    return Response.json({ error: 'Revalidation failed' }, { status: 500 })
  }
}
```

Streaming with Suspense:
```typescript
// app/dashboard/page.tsx
import { Suspense } from 'react'

export default function DashboardPage() {
  return (
    <div className="dashboard">
      <h1>Dashboard</h1>

      <Suspense fallback={<MetricsSkeleton />}>
        <Metrics />
      </Suspense>

      <Suspense fallback={<ChartSkeleton />}>
        <AnalyticsChart />
      </Suspense>

      <Suspense fallback={<TableSkeleton />}>
        <RecentOrders />
      </Suspense>
    </div>
  )
}

async function Metrics() {
  const data = await fetch('https://api.example.com/metrics', {
    next: { revalidate: 30 }
  }).then(r => r.json())

  return <MetricsDisplay data={data} />
}
```

### Phase 3: Vercel Configuration

vercel.json Configuration:
```json
{
  "$schema": "https://openapi.vercel.sh/vercel.json",
  "framework": "nextjs",
  "buildCommand": "pnpm build",
  "installCommand": "pnpm install",
  "outputDirectory": ".next",
  "regions": ["iad1", "sfo1", "fra1", "hnd1"],
  "functions": {
    "app/api/**/*.ts": {
      "memory": 1024,
      "maxDuration": 30
    },
    "app/api/heavy/**/*.ts": {
      "memory": 3008,
      "maxDuration": 60
    }
  },
  "crons": [
    {
      "path": "/api/cron/cleanup",
      "schedule": "0 0 * * *"
    },
    {
      "path": "/api/cron/sync",
      "schedule": "*/15 * * * *"
    }
  ],
  "headers": [
    {
      "source": "/api/(.*)",
      "headers": [
        { "key": "Access-Control-Allow-Origin", "value": "*" },
        { "key": "Cache-Control", "value": "s-maxage=60, stale-while-revalidate" }
      ]
    },
    {
      "source": "/(.*)",
      "headers": [
        { "key": "X-Content-Type-Options", "value": "nosniff" },
        { "key": "X-Frame-Options", "value": "DENY" },
        { "key": "X-XSS-Protection", "value": "1; mode=block" }
      ]
    }
  ],
  "rewrites": [
    { "source": "/blog/:slug", "destination": "/posts/:slug" },
    { "source": "/api/v1/:path*", "destination": "/api/:path*" }
  ],
  "redirects": [
    { "source": "/old-page", "destination": "/new-page", "permanent": true }
  ]
}
```

Environment Variables Management:
```bash
# Production environment
vercel env add DATABASE_URL production
vercel env add API_SECRET production
vercel env add NEXT_PUBLIC_API_URL production

# Preview environments (PR deployments)
vercel env add DATABASE_URL preview
vercel env add API_SECRET preview

# Development environment
vercel env add DATABASE_URL development

# Pull environment for local development
vercel env pull .env.local
```

### Phase 4: Preview Deployments

GitHub Integration Setup:
```yaml
# .github/workflows/vercel-preview.yml
name: Vercel Preview Deployment
on:
  pull_request:
    types: [opened, synchronize, reopened]

jobs:
  deploy-preview:
    runs-on: ubuntu-latest
    environment:
      name: Preview
      url: ${{ steps.deploy.outputs.url }}
    steps:
      - uses: actions/checkout@v4

      - name: Install Vercel CLI
        run: npm i -g vercel@latest

      - name: Pull Vercel Environment
        run: vercel pull --yes --environment=preview --token=${{ secrets.VERCEL_TOKEN }}

      - name: Build Project
        run: vercel build --token=${{ secrets.VERCEL_TOKEN }}

      - name: Deploy to Vercel
        id: deploy
        run: |
          url=$(vercel deploy --prebuilt --token=${{ secrets.VERCEL_TOKEN }})
          echo "url=$url" >> $GITHUB_OUTPUT

      - name: Comment PR with Preview URL
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              owner: context.repo.owner,
              repo: context.repo.repo,
              issue_number: context.issue.number,
              body: `Preview deployed: ${{ steps.deploy.outputs.url }}`
            })
```

Production Deployment:
```yaml
# .github/workflows/vercel-production.yml
name: Vercel Production Deployment
on:
  push:
    branches: [main]

jobs:
  deploy-production:
    runs-on: ubuntu-latest
    environment:
      name: Production
      url: https://your-domain.com
    steps:
      - uses: actions/checkout@v4

      - name: Install Vercel CLI
        run: npm i -g vercel@latest

      - name: Pull Vercel Environment
        run: vercel pull --yes --environment=production --token=${{ secrets.VERCEL_TOKEN }}

      - name: Build Project
        run: vercel build --prod --token=${{ secrets.VERCEL_TOKEN }}

      - name: Deploy to Vercel
        run: vercel deploy --prebuilt --prod --token=${{ secrets.VERCEL_TOKEN }}
```

### Phase 5: Web Vitals Monitoring

Analytics Integration:
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

Custom Web Vitals Reporting:
```typescript
// app/components/WebVitals.tsx
'use client'

import { useReportWebVitals } from 'next/web-vitals'

export function WebVitals() {
  useReportWebVitals((metric) => {
    const body = JSON.stringify({
      name: metric.name,
      value: metric.value,
      rating: metric.rating,
      delta: metric.delta,
      id: metric.id,
      navigationType: metric.navigationType
    })

    // Send to custom analytics endpoint
    if (navigator.sendBeacon) {
      navigator.sendBeacon('/api/vitals', body)
    } else {
      fetch('/api/vitals', { body, method: 'POST', keepalive: true })
    }
  })

  return null
}
```

---

## Advanced Patterns

### Monorepo with Turborepo

```json
{
  "buildCommand": "cd ../.. && pnpm turbo build --filter=web",
  "installCommand": "cd ../.. && pnpm install",
  "framework": "nextjs"
}
```

### Blue-Green Deployment

Deploy new version, run smoke tests on preview URL, then switch production alias using Vercel SDK aliases.assign() method for zero-downtime releases.

### Context7 Integration

Use `/vercel/vercel` for edge function patterns and `/vercel/next.js` for App Router, ISR, and streaming patterns with appropriate token allocation.

---

## Works Well With

- `moai-platform-railway` - Container-based deployment alternative
- `moai-lang-typescript` - TypeScript patterns for Next.js
- `moai-domain-frontend` - React and Next.js component patterns
- `moai-foundation-quality` - Deployment validation and testing

---

Status: Production Ready | Version: 1.0.0 | Updated: 2025-12-07
