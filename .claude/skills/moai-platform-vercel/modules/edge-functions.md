# Edge Functions and Middleware

## Overview

Vercel Edge Functions execute at the network edge (30+ global locations) with sub-50ms cold starts. They provide low-latency compute for request transformation, geo-based personalization, and authentication checks.

## Edge Runtime Configuration

### Basic Edge Function

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

### Region Selection

Available regions for preferredRegion configuration:
- `iad1` - Washington, D.C. (US East)
- `sfo1` - San Francisco (US West)
- `fra1` - Frankfurt (Europe)
- `hnd1` - Tokyo (Asia Pacific)
- `syd1` - Sydney (Australia)
- `gru1` - Sao Paulo (South America)

## Edge Middleware

### Request Transformation

```typescript
// middleware.ts
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export function middleware(request: NextRequest) {
  const response = NextResponse.next()

  // Add custom headers
  response.headers.set('x-custom-header', 'my-value')

  // Geo-based locale detection
  const country = request.geo?.country ?? 'US'
  response.cookies.set('user-country', country)

  return response
}

export const config = {
  matcher: ['/((?!api|_next/static|_next/image|favicon.ico).*)']
}
```

### A/B Testing at Edge

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

## Geo-Based Content Delivery

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

## Authentication at Edge

```typescript
// middleware.ts
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export async function middleware(request: NextRequest) {
  const token = request.cookies.get('auth-token')?.value

  if (!token && request.nextUrl.pathname.startsWith('/dashboard')) {
    return NextResponse.redirect(new URL('/login', request.url))
  }

  // Optional: Verify JWT at edge for stateless auth
  if (token) {
    try {
      // Use edge-compatible JWT library
      const payload = await verifyToken(token)
      const response = NextResponse.next()
      response.headers.set('x-user-id', payload.sub)
      return response
    } catch {
      return NextResponse.redirect(new URL('/login', request.url))
    }
  }

  return NextResponse.next()
}
```

## Edge Function Limitations

### Supported APIs

Edge runtime supports Web APIs but not Node.js APIs:
- Fetch API (fetch, Request, Response, Headers)
- URL and URLSearchParams
- TextEncoder and TextDecoder
- crypto (Web Crypto API)
- ReadableStream and WritableStream

### Not Supported

- Node.js built-in modules (fs, path, etc.)
- Native Node.js addons
- Synchronous file operations
- Long-running processes (max 30 seconds)

### Memory and Duration

- Maximum memory: 128MB
- Maximum duration: 30 seconds (Hobby), 60 seconds (Pro/Enterprise)
- Cold start: typically under 50ms

## Best Practices

### Minimize Cold Starts

- Keep dependencies minimal
- Use dynamic imports for rarely-used code
- Prefer edge-compatible libraries

### Caching Strategy

```typescript
// Set appropriate cache headers
return Response.json(data, {
  headers: {
    // Cache at CDN for 1 hour
    'Cache-Control': 'public, s-maxage=3600',
    // Browser cache for 5 minutes
    'CDN-Cache-Control': 'public, max-age=300',
    // Stale-while-revalidate for smooth updates
    'Stale-While-Revalidate': '60'
  }
})
```

### Error Handling

```typescript
export async function GET(request: Request) {
  try {
    const result = await processRequest(request)
    return Response.json(result)
  } catch (error) {
    console.error('Edge function error:', error)
    return Response.json(
      { error: 'Internal Server Error' },
      { status: 500 }
    )
  }
}
```

## Context7 Integration

For latest Edge Functions documentation, use:
- Library: `/vercel/vercel`
- Topics: edge-functions, middleware, geo-detection
- Token allocation: 5000-10000 for comprehensive coverage
