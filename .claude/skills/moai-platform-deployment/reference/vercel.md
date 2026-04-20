# Vercel Platform Reference

Comprehensive guide to Vercel edge deployment platform.

---

## Edge Functions and Middleware

### Edge Runtime Configuration

Edge Functions execute globally at 30+ edge locations with sub-50ms cold starts.

**Runtime Declaration**:
```typescript
export const runtime = "edge"
export const preferredRegion = ["iad1", "sfo1", "fra1"]
```

**Supported Edge APIs**:
- Request/Response manipulation
- Geo-location data (request.geo)
- Headers and cookies
- URL rewriting and redirects
- A/B testing at edge

### Middleware Patterns

**Authentication Middleware**:
```typescript
import { NextResponse } from "next/server"
import type { NextRequest } from "next/server"

export function middleware(request: NextRequest) {
  const token = request.cookies.get("auth-token")

  if (!token && request.nextUrl.pathname.startsWith("/dashboard")) {
    return NextResponse.redirect(new URL("/login", request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: ["/dashboard/:path*", "/api/:path*"]
}
```

**Geo-Based Content Delivery**:
```typescript
export function middleware(request: NextRequest) {
  const country = request.geo?.country || "US"
  const response = NextResponse.next()

  response.headers.set("x-user-country", country)

  if (country === "GB") {
    return NextResponse.rewrite(new URL("/gb", request.url))
  }

  return response
}
```

**A/B Testing at Edge**:
```typescript
export function middleware(request: NextRequest) {
  const bucket = Math.random() < 0.5 ? "a" : "b"
  const response = NextResponse.next()

  response.cookies.set("ab-test-bucket", bucket)

  if (bucket === "b") {
    return NextResponse.rewrite(new URL("/variant-b", request.url))
  }

  return response
}
```

---

## ISR and Caching Strategies

### Incremental Static Regeneration

**Time-Based Revalidation**:
```typescript
export const revalidate = 3600 // Revalidate every hour

export default async function Page() {
  const data = await fetch("https://api.example.com/data")
  return <div>{/* Render data */}</div>
}
```

**On-Demand Revalidation**:
```typescript
// app/api/revalidate/route.ts
import { revalidatePath, revalidateTag } from "next/cache"
import { NextRequest } from "next/server"

export async function POST(request: NextRequest) {
  const tag = request.nextUrl.searchParams.get("tag")

  if (tag) {
    revalidateTag(tag)
    return Response.json({ revalidated: true, tag })
  }

  const path = request.nextUrl.searchParams.get("path")
  if (path) {
    revalidatePath(path)
    return Response.json({ revalidated: true, path })
  }

  return Response.json({ revalidated: false })
}
```

**Tag-Based Cache Invalidation**:
```typescript
export default async function Page() {
  const posts = await fetch("https://api.example.com/posts", {
    next: { tags: ["posts"] }
  })

  return <div>{/* Render posts */}</div>
}

// Invalidate all posts caches
await revalidateTag("posts")
```

### CDN Cache Headers

**Custom Cache Control**:
```typescript
export async function GET() {
  const data = await fetchData()

  return Response.json(data, {
    headers: {
      "Cache-Control": "public, s-maxage=3600, stale-while-revalidate=86400"
    }
  })
}
```

**Streaming with Suspense**:
```typescript
import { Suspense } from "react"

export default function Page() {
  return (
    <div>
      <Suspense fallback={<Loading />}>
        <SlowComponent />
      </Suspense>
    </div>
  )
}
```

---

## Deployment Configuration

### vercel.json Complete Reference

```json
{
  "$schema": "https://openapi.vercel.sh/vercel.json",
  "framework": "nextjs",
  "buildCommand": "npm run build",
  "installCommand": "npm install",
  "outputDirectory": ".next",

  "regions": ["iad1", "sfo1", "fra1"],

  "functions": {
    "app/api/**/*.ts": {
      "memory": 1024,
      "maxDuration": 10
    }
  },

  "crons": [
    {
      "path": "/api/cron/hourly",
      "schedule": "0 * * * *"
    }
  ],

  "redirects": [
    {
      "source": "/old-path",
      "destination": "/new-path",
      "permanent": true
    }
  ],

  "headers": [
    {
      "source": "/(.*)",
      "headers": [
        {
          "key": "X-Frame-Options",
          "value": "DENY"
        },
        {
          "key": "X-Content-Type-Options",
          "value": "nosniff"
        }
      ]
    }
  ]
}
```

### Environment Variables

**Variable Types**:
- Production: Used in production deployments
- Preview: Used in preview deployments
- Development: Used in local development

**Setting Variables**:
```bash
# Via CLI
vercel env add SECRET_KEY production

# Via Dashboard
# Project Settings → Environment Variables
```

**Accessing in Code**:
```typescript
const apiKey = process.env.API_KEY
```

### Monorepo Setup with Turborepo

```json
{
  "buildCommand": "turbo run build --filter=web",
  "installCommand": "pnpm install",
  "framework": "nextjs",
  "outputDirectory": "apps/web/.next"
}
```

---

## Analytics and Speed Insights

### Vercel Analytics Integration

**Installation**:
```bash
npm install @vercel/analytics
```

**Setup**:
```typescript
// app/layout.tsx
import { Analytics } from "@vercel/analytics/react"

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        {children}
        <Analytics />
      </body>
    </html>
  )
}
```

### Speed Insights for Web Vitals

**Installation**:
```bash
npm install @vercel/speed-insights
```

**Setup**:
```typescript
import { SpeedInsights } from "@vercel/speed-insights/next"

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        {children}
        <SpeedInsights />
      </body>
    </html>
  )
}
```

**Custom Performance Monitoring**:
```typescript
import { sendToAnalytics } from "@vercel/analytics"

export function reportWebVitals(metric) {
  sendToAnalytics(metric)
}
```

---

## Vercel Storage Solutions

### Vercel KV (Redis)

**Setup**:
```bash
npm install @vercel/kv
```

**Usage**:
```typescript
import { kv } from "@vercel/kv"

// Set value
await kv.set("user:123", { name: "John", email: "john@example.com" })

// Get value
const user = await kv.get("user:123")

// Increment counter
await kv.incr("page-views")

// Expire key
await kv.expire("session:abc", 3600) // 1 hour
```

### Vercel Blob (Object Storage)

**Setup**:
```bash
npm install @vercel/blob
```

**Upload**:
```typescript
import { put } from "@vercel/blob"

export async function POST(request: Request) {
  const file = await request.blob()
  const blob = await put("avatar.png", file, { access: "public" })

  return Response.json({ url: blob.url })
}
```

**List Files**:
```typescript
import { list } from "@vercel/blob"

const { blobs } = await list({ prefix: "avatars/" })
```

### Vercel Postgres

**Setup**:
```bash
npm install @vercel/postgres
```

**Usage**:
```typescript
import { sql } from "@vercel/postgres"

const users = await sql`SELECT * FROM users WHERE id = ${userId}`

// Transaction
await sql.transaction(async (client) => {
  await client`INSERT INTO users (name) VALUES (${name})`
  await client`INSERT INTO profiles (user_id) VALUES (${userId})`
})
```

---

## CLI Commands

### Deployment

```bash
# Deploy current directory
vercel

# Deploy with production
vercel --prod

# Deploy with build environment variables
vercel --build-env API_URL=https://api.example.com
```

### Environment Variables

```bash
# Add variable
vercel env add SECRET_KEY

# List variables
vercel env ls

# Pull variables to .env.local
vercel env pull
```

### Logs and Domains

```bash
# View logs
vercel logs <deployment-url>

# Add domain
vercel domains add example.com

# List domains
vercel domains ls
```

---

## Best Practices

### Performance Optimization

1. **Use Edge Functions for latency-sensitive operations**
2. **Enable ISR for dynamic content with caching**
3. **Implement streaming SSR with Suspense**
4. **Optimize images with next/image**
5. **Use Vercel Analytics to monitor Core Web Vitals**

### Security Headers

Always configure security headers in vercel.json:
- X-Frame-Options: DENY
- X-Content-Type-Options: nosniff
- Strict-Transport-Security: max-age=31536000
- Content-Security-Policy: Default policy

### Cost Optimization

1. **Use appropriate function memory settings**
2. **Configure maxDuration based on needs**
3. **Implement efficient ISR strategies**
4. **Use edge caching where possible**
5. **Monitor bandwidth usage with Analytics**

---

## Troubleshooting

### Common Issues

**Build Failures**:
- Check build logs in Vercel dashboard
- Verify buildCommand in vercel.json
- Ensure dependencies are in package.json

**Function Timeouts**:
- Increase maxDuration in vercel.json
- Optimize database queries
- Use streaming responses

**Cache Issues**:
- Clear cache with revalidatePath/revalidateTag
- Check Cache-Control headers
- Verify ISR configuration

---

## Context7 Integration

For latest Vercel documentation:

```typescript
// Step 1: Resolve library
const libraries = await mcp__context7__resolve_library_id({
  libraryName: "vercel",
  query: "edge functions middleware"
})

// Step 2: Get documentation
const docs = await mcp__context7__get_library_docs({
  libraryId: libraries[0].id,
  topic: "edge-middleware",
  maxTokens: 8000
})
```

---

**Status**: Production Ready
**Platform**: Vercel Edge Deployment
**Updated**: 2026-02-09
