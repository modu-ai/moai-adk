---
title: Next.js 16 Advanced Patterns
description: Master Next.js 16 App Router, Server Components, streaming, and middleware best practices
freedom_level: high
tier: framework
updated: 2025-10-31
---

# Next.js 16 Advanced Patterns

## Overview

Next.js 16 introduces significant improvements to the App Router architecture, Server Components, Turbopack performance, and the network boundary model. This skill covers advanced patterns for building production-grade Next.js applications with optimal performance and developer experience.

## Key Patterns

### 1. Server Components & Client Components Strategy

**Pattern**: Default to Server Components, selectively opt into Client Components only when needed.

```typescript
// app/dashboard/page.tsx (Server Component by default)
import { ClientChart } from './ClientChart'

async function getData() {
  const res = await fetch('https://api.example.com/data', {
    next: { revalidate: 3600 } // ISR with 1-hour cache
  })
  return res.json()
}

export default async function DashboardPage() {
  const data = await getData()
  
  return (
    <div>
      <h1>Dashboard</h1>
      {/* Server-rendered content */}
      <ServerSummary data={data} />
      {/* Client-side interactive chart */}
      <ClientChart data={data} />
    </div>
  )
}

// app/dashboard/ClientChart.tsx
'use client'

export function ClientChart({ data }: { data: any }) {
  const [filter, setFilter] = useState('all')
  return <Chart data={data} filter={filter} />
}
```

**Benefits**: Smaller JavaScript bundle, better SEO, faster Time to First Byte (TTFB).

### 2. Streaming with Suspense

**Pattern**: Stream content progressively to improve perceived performance.

```typescript
// app/products/page.tsx
import { Suspense } from 'react'
import { ProductList } from './ProductList'
import { RecommendedProducts } from './RecommendedProducts'

export default function ProductsPage() {
  return (
    <div>
      <h1>Products</h1>
      
      {/* Critical content loads first */}
      <Suspense fallback={<ProductsSkeleton />}>
        <ProductList />
      </Suspense>
      
      {/* Non-critical content streams later */}
      <Suspense fallback={<RecommendedSkeleton />}>
        <RecommendedProducts />
      </Suspense>
    </div>
  )
}

// app/products/ProductList.tsx
async function ProductList() {
  const products = await fetch('https://api.example.com/products')
  return <div>{/* Render products */}</div>
}
```

**Benefits**: Users see content faster, reduces waterfall loading, improves LCP (Largest Contentful Paint).

### 3. Proxy.ts Network Boundary (Next.js 16)

**Pattern**: Use `proxy.ts` to define explicit network boundaries (replaces `middleware.ts`).

```typescript
// proxy.ts (Next.js 16+)
import { NextRequest } from 'next/server'

export function proxy(request: NextRequest) {
  const url = request.nextUrl
  
  // Geolocation-based routing
  const country = request.geo?.country || 'US'
  if (country === 'CN' && url.pathname.startsWith('/api')) {
    return Response.redirect(new URL('/api-cn' + url.pathname, request.url))
  }
  
  // A/B testing
  if (url.pathname === '/pricing') {
    const variant = Math.random() > 0.5 ? 'A' : 'B'
    url.searchParams.set('variant', variant)
    return NextResponse.rewrite(url)
  }
  
  return NextResponse.next()
}

export const config = {
  matcher: ['/api/:path*', '/pricing']
}
```

**Migration**: Rename `middleware.ts` → `proxy.ts` and rename exported function to `proxy`. The `middleware.ts` pattern is deprecated and will be removed in future versions.

### 4. Parallel Routes & Intercepting Routes

**Pattern**: Load multiple page segments in parallel and intercept routes for modals.

```typescript
// app/dashboard/layout.tsx
export default function DashboardLayout({
  children,
  analytics,
  team
}: {
  children: React.ReactNode
  analytics: React.ReactNode
  team: React.ReactNode
}) {
  return (
    <div>
      <div>{children}</div>
      <div className="grid grid-cols-2">
        <div>{analytics}</div>
        <div>{team}</div>
      </div>
    </div>
  )
}

// app/dashboard/@analytics/page.tsx (Parallel route)
export default async function AnalyticsSlot() {
  const data = await fetchAnalytics()
  return <AnalyticsChart data={data} />
}

// app/@modal/(.)photos/[id]/page.tsx (Intercepting route)
export default function PhotoModal({ params }: { params: { id: string } }) {
  return (
    <dialog open>
      <Photo id={params.id} />
    </dialog>
  )
}
```

**Benefits**: Better UX with modal interception, parallel data fetching reduces total load time.

### 5. Server Actions & Progressive Enhancement

**Pattern**: Use Server Actions for form handling with zero JavaScript fallback.

```typescript
// app/actions.ts
'use server'

import { revalidatePath } from 'next/cache'
import { redirect } from 'next/navigation'

export async function createPost(formData: FormData) {
  const title = formData.get('title') as string
  const content = formData.get('content') as string
  
  // Server-side validation
  if (!title || title.length < 3) {
    return { error: 'Title must be at least 3 characters' }
  }
  
  await db.posts.create({ title, content })
  
  revalidatePath('/posts')
  redirect('/posts')
}

// app/posts/new/page.tsx
import { createPost } from '@/app/actions'

export default function NewPostPage() {
  return (
    <form action={createPost}>
      <input name="title" required />
      <textarea name="content" required />
      <button type="submit">Create Post</button>
    </form>
  )
}
```

**Benefits**: Works without JavaScript, better accessibility, automatic loading states.

### 6. Incremental Static Regeneration (ISR) Strategy

**Pattern**: Use ISR with proper revalidation strategies for dynamic content.

```typescript
// app/blog/[slug]/page.tsx
export async function generateStaticParams() {
  const posts = await fetch('https://api.example.com/posts').then(res => res.json())
  return posts.map((post: any) => ({ slug: post.slug }))
}

export default async function BlogPost({ params }: { params: { slug: string } }) {
  const post = await fetch(`https://api.example.com/posts/${params.slug}`, {
    next: { 
      revalidate: 60, // Revalidate every 60 seconds
      tags: ['posts', `post-${params.slug}`]
    }
  }).then(res => res.json())
  
  return <article>{post.content}</article>
}

// Manual revalidation in Server Action
import { revalidateTag } from 'next/cache'

export async function updatePost(slug: string) {
  await db.posts.update(slug, data)
  revalidateTag(`post-${slug}`) // Invalidate specific post cache
}
```

### 7. Edge Runtime Optimization

**Pattern**: Use Edge Runtime for latency-sensitive routes.

```typescript
// app/api/geolocation/route.ts
export const runtime = 'edge'

export async function GET(request: Request) {
  const { geo } = request as any
  
  return Response.json({
    country: geo?.country || 'Unknown',
    city: geo?.city || 'Unknown',
    latency: Date.now()
  })
}

// app/api/personalization/route.ts (Node.js runtime)
export const runtime = 'nodejs'
export const dynamic = 'force-dynamic'

export async function GET() {
  const recommendations = await complexMLModel()
  return Response.json(recommendations)
}
```

**Decision Matrix**:
- **Edge**: Geolocation, A/B testing, auth checks, simple transformations
- **Node.js**: Database queries, file system access, complex computations

## Checklist

- [ ] Audit components: Are 80%+ Server Components? Move interactivity to Client Components only
- [ ] Implement Suspense boundaries for slow data fetching (>500ms)
- [ ] Migrate `middleware.ts` → `proxy.ts` for Next.js 16 compatibility
- [ ] Enable Turbopack in development: `next dev --turbo`
- [ ] Use ISR with cache tags for fine-grained invalidation
- [ ] Profile edge vs Node.js runtime: measure cold start times and latency
- [ ] Test Server Actions work without JavaScript (disable JS in DevTools)

## Resources

- **Official Next.js 16 Release**: https://nextjs.org/blog/next-16
- **App Router Documentation**: https://nextjs.org/docs/app/getting-started/server-and-client-components
- **Server Actions Guide**: https://nextjs.org/docs/app/building-your-application/data-fetching/server-actions-and-mutations
- **Middleware Documentation**: https://nextjs.org/docs/app/building-your-application/routing/middleware
- **Performance Best Practices (2025)**: https://medium.com/@GoutamSingha/next-js-best-practices-in-2025-build-faster-cleaner-scalable-apps-7efbad2c3820

---

**Version**: 1.0.0  
**Last Updated**: 2025-10-31  
**Model Recommendation**: Sonnet (deep reasoning for architecture decisions)
