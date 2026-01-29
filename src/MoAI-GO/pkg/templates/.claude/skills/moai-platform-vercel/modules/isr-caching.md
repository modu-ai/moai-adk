# ISR and Caching Strategies

## Overview

Incremental Static Regeneration (ISR) enables static pages to be updated after build time without rebuilding the entire site. Combined with Vercel's global CDN, ISR provides the best of static and dynamic rendering.

## Basic ISR Configuration

### Time-Based Revalidation

```typescript
// app/products/[id]/page.tsx
import { notFound } from 'next/navigation'

// Revalidate every 60 seconds
export const revalidate = 60

export default async function ProductPage({ params }: { params: { id: string } }) {
  const product = await fetchProduct(params.id)
  if (!product) notFound()

  return <ProductDetail product={product} />
}

async function fetchProduct(id: string) {
  const res = await fetch(`https://api.example.com/products/${id}`, {
    next: { revalidate: 60 }
  })
  if (!res.ok) return null
  return res.json()
}
```

### Static Params Generation

```typescript
// app/products/[id]/page.tsx

// Generate static params for top products at build time
export async function generateStaticParams() {
  const products = await fetch('https://api.example.com/products/top').then(r => r.json())
  return products.map((p: { id: string }) => ({ id: p.id }))
}
```

## On-Demand Revalidation

### Tag-Based Revalidation

```typescript
// app/products/[id]/page.tsx
async function fetchProduct(id: string) {
  const res = await fetch(`https://api.example.com/products/${id}`, {
    next: { tags: [`product-${id}`, 'products'] }
  })
  if (!res.ok) return null
  return res.json()
}
```

### Revalidation API Endpoint

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

### Webhook Integration

```typescript
// app/api/webhooks/cms/route.ts
import { revalidateTag } from 'next/cache'
import { NextRequest } from 'next/server'
import crypto from 'crypto'

export async function POST(request: NextRequest) {
  const body = await request.text()
  const signature = request.headers.get('x-webhook-signature')

  // Verify webhook signature
  const expectedSignature = crypto
    .createHmac('sha256', process.env.WEBHOOK_SECRET!)
    .update(body)
    .digest('hex')

  if (signature !== expectedSignature) {
    return Response.json({ error: 'Invalid signature' }, { status: 401 })
  }

  const payload = JSON.parse(body)

  // Revalidate based on content type
  if (payload.type === 'product') {
    revalidateTag(`product-${payload.id}`)
    revalidateTag('products')
  } else if (payload.type === 'blog') {
    revalidateTag(`post-${payload.slug}`)
    revalidateTag('blog')
  }

  return Response.json({ revalidated: true })
}
```

## Caching Strategies

### Fetch Cache Configuration

```typescript
// Force cache (default for static)
fetch(url, { cache: 'force-cache' })

// No cache (always fresh)
fetch(url, { cache: 'no-store' })

// Time-based revalidation
fetch(url, { next: { revalidate: 3600 } })

// Tag-based revalidation
fetch(url, { next: { tags: ['posts'] } })
```

### Route Segment Config

```typescript
// app/api/data/route.ts

// Force dynamic rendering
export const dynamic = 'force-dynamic'

// Force static rendering
export const dynamic = 'force-static'

// Set revalidation interval
export const revalidate = 60

// Configure runtime
export const runtime = 'edge' // or 'nodejs'
```

## CDN Cache Headers

### Response Headers

```typescript
// app/api/data/route.ts
export async function GET() {
  const data = await fetchData()

  return Response.json(data, {
    headers: {
      // Cache at Vercel's CDN for 1 hour
      'Cache-Control': 'public, s-maxage=3600',
      // Serve stale while revalidating
      'Stale-While-Revalidate': '60',
      // CDN-specific cache control
      'CDN-Cache-Control': 'public, max-age=86400'
    }
  })
}
```

### vercel.json Cache Headers

```json
{
  "headers": [
    {
      "source": "/api/(.*)",
      "headers": [
        {
          "key": "Cache-Control",
          "value": "s-maxage=60, stale-while-revalidate"
        }
      ]
    },
    {
      "source": "/(.*)",
      "headers": [
        {
          "key": "Cache-Control",
          "value": "public, max-age=31536000, immutable"
        }
      ]
    }
  ]
}
```

## Streaming with Suspense

### Progressive Rendering

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

### Loading States

```typescript
// app/dashboard/loading.tsx
export default function Loading() {
  return (
    <div className="animate-pulse">
      <div className="h-8 bg-gray-200 rounded w-1/4 mb-4" />
      <div className="h-64 bg-gray-200 rounded mb-4" />
      <div className="h-32 bg-gray-200 rounded" />
    </div>
  )
}
```

## Cache Debugging

### Check Cache Status

Response headers indicate cache status:
- `x-vercel-cache: HIT` - Served from cache
- `x-vercel-cache: MISS` - Not in cache, fetched from origin
- `x-vercel-cache: STALE` - Served stale while revalidating
- `x-vercel-cache: BYPASS` - Cache bypassed

### Force Cache Bypass

```bash
# Add header to bypass cache
curl -H "x-vercel-skip-cache: 1" https://your-app.vercel.app/api/data
```

## Best Practices

### ISR Strategy Selection

- Use time-based revalidation for content that changes predictably
- Use on-demand revalidation for content updated via CMS webhooks
- Use `no-store` for personalized or real-time data
- Combine ISR with client-side revalidation for hybrid approach

### Performance Optimization

- Set appropriate revalidation intervals based on content freshness requirements
- Use tag-based revalidation for granular cache control
- Implement stale-while-revalidate for better user experience
- Monitor cache hit rates in Vercel Analytics

## Context7 Integration

For latest ISR and caching documentation, use:
- Library: `/vercel/next.js`
- Topics: ISR, caching, revalidation, streaming
- Token allocation: 8000-12000 for comprehensive coverage
