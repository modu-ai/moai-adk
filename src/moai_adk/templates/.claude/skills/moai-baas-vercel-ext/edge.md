---
name: moai-baas-vercel-ext
description: Edge Functions & Security - Vercel edge computing patterns with security implementation
---

## Edge Functions & Security

### Advanced Edge Request Handling

```typescript
export class EnterpriseVercelManager {
  // Edge function with advanced features
  async handleEdgeRequest(request: VercelRequest): Promise<VercelResponse> {
    try {
      const url = new URL(request.url);
      
      // Security headers
      const securityHeaders = {
        'X-Content-Type-Options': 'nosniff',
        'X-Frame-Options': 'DENY',
        'X-XSS-Protection': '1; mode=block',
        'Referrer-Policy': 'strict-origin-when-cross-origin',
        'Permissions-Policy': 'camera=(), microphone=(), geolocation=()',
      };

      // CORS configuration
      const corsHeaders = this.configureCORS(request);

      // Rate limiting
      const rateLimitResult = await this.checkRateLimit(request);
      if (!rateLimitResult.allowed) {
        return new Response('Rate limit exceeded', {
          status: 429,
          headers: {
            ...securityHeaders,
            'Retry-After': rateLimitResult.retryAfter.toString(),
          },
        });
      }

      // Geographic routing
      const region = this.getOptimalRegion(request);
      
      // Route to appropriate handler
      if (url.pathname.startsWith('/api/')) {
        return await this.handleAPIRequest(request, region);
      }

      // Static file serving with optimization
      if (this.isStaticFile(url.pathname)) {
        return await this.serveStaticFile(url.pathname);
      }

      // SPA fallback
      return await this.serveSPA(request);

    } catch (error) {
      console.error('Edge request error:', error);
      return new Response('Internal Server Error', { status: 500 });
    }
  }
}
```

### CORS Configuration

```typescript
private configureCORS(request: VercelRequest): Record<string, string> {
  const origin = request.headers.get('origin');
  const allowedOrigins = [
    'https://yourdomain.com',
    'https://www.yourdomain.com',
    'https://app.yourdomain.com',
  ];

  if (allowedOrigins.includes(origin || '')) {
    return {
      'Access-Control-Allow-Origin': origin!,
      'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type, Authorization',
      'Access-Control-Allow-Credentials': 'true',
    };
  }

  return {};
}
```

### Rate Limiting

```typescript
private async checkRateLimit(request: VercelRequest): Promise<RateLimitResult> {
  const clientIP = request.headers.get('x-forwarded-for') || 
                   request.headers.get('x-real-ip') || 
                   'unknown';
  
  // Implement sliding window rate limiting
  const key = `rate_limit:${clientIP}`;
  const window = 60000; // 1 minute
  const limit = 100; // requests per minute

  // In production, use Redis or similar distributed cache
  const current = await this.getRateLimitCount(key, window);
  
  if (current >= limit) {
    return {
      allowed: false,
      retryAfter: Math.ceil(window / 1000),
    };
  }

  await this.incrementRateLimitCount(key);
  return { allowed: true };
}
```

### Geographic Routing

```typescript
private getOptimalRegion(request: VercelRequest): string {
  // Geographic routing based on client location
  const country = request.headers.get('x-vercel-ip-country');
  const regionMap: Record<string, string> = {
    'US': 'iad1', // East Coast US
    'CA': 'hnd1', // West Coast US
    'GB': 'lhr1', // United Kingdom
    'DE': 'fra1', // Germany
    'FR': 'cdg1', // France
    'NL': 'arn1', // Netherlands
  };

  return regionMap[country || 'US'] || 'iad1';
}
```

### API Request Handling

```typescript
private async handleAPIRequest(
  request: VercelRequest, 
  region: string
): Promise<VercelResponse> {
  const url = new URL(request.url);
  const pathParts = url.pathname.split('/').filter(Boolean);
  
  // Route to appropriate API handler
  if (pathParts[0] === 'api' && pathParts[1] === 'users') {
    return await this.handleUsersAPI(request, pathParts.slice(2), region);
  }

  if (pathParts[0] === 'api' && pathParts[1] === 'analytics') {
    return await this.handleAnalyticsAPI(request, pathParts.slice(2), region);
  }

  return new Response('API endpoint not found', { status: 404 });
}

private async handleUsersAPI(
  request: VercelRequest,
  pathParts: string[],
  region: string
): Promise<VercelResponse> {
  const userId = pathParts[0];
  
  if (!userId) {
    return new Response('User ID required', { status: 400 });
  }

  try {
    // Fetch user data from database
    const userData = await this.fetchUserData(userId);
    
    if (!userData) {
      return new Response('User not found', { status: 404 });
    }

    // Return user data with proper headers
    return new Response(JSON.stringify(userData), {
      status: 200,
      headers: {
        'Content-Type': 'application/json',
        'Cache-Control': 's-maxage=60, stale-while-revalidate=300',
        'X-Region': region,
      },
    });
  } catch (error) {
    console.error('Users API error:', error);
    return new Response('Internal Server Error', { status: 500 });
  }
}
```

### Advanced Caching with Edge Functions

```typescript
// app/api/cached-request/route.ts
import { NextRequest, NextResponse } from 'next/server';
import crypto from 'crypto';

// Simple in-memory cache (use Redis for production)
const edgeCache = new Map<string, { data: any; expiry: number }>();

export async function GET(request: NextRequest) {
  try {
    // Generate cache key
    const cacheKey = generateCacheKey(
      request.nextUrl.pathname,
      'GET',
      Object.fromEntries(request.nextUrl.searchParams)
    );

    // Check cache
    const cached = getCachedResponse(cacheKey);
    if (cached) {
      return NextResponse.json(cached, {
        headers: {
          'X-Cache': 'HIT',
          'Cache-Control': 'public, max-age=3600',
        },
      });
    }

    // Route to appropriate handler
    let response;
    const pathname = request.nextUrl.pathname;

    if (pathname.startsWith('/api/users/')) {
      response = await processUsersRequest(request);
    } else if (pathname.startsWith('/api/analytics/')) {
      response = await processAnalyticsRequest(request);
    } else {
      return NextResponse.json(
        { error: 'Endpoint not found' },
        { status: 404 }
      );
    }

    // Cache successful response
    if (response.ok) {
      const data = await response.json();
      setCachedResponse(cacheKey, data, 3600000); // 1 hour TTL
      return NextResponse.json(data, {
        headers: {
          'X-Cache': 'MISS',
          'Cache-Control': 'public, max-age=3600',
        },
      });
    }

    return response;
  } catch (error) {
    console.error('Cached request error:', error);
    return NextResponse.json(
      { error: 'Internal Server Error' },
      { status: 500 }
    );
  }
}

function generateCacheKey(
  path: string,
  method: string,
  params: Record<string, any>
): string {
  const keyData = `${method}:${path}:${JSON.stringify(
    Object.entries(params).sort()
  )}`;
  return crypto.createHash('md5').update(keyData).digest('hex');
}

function getCachedResponse(key: string): any | null {
  const cached = edgeCache.get(key);
  if (!cached) return null;
  if (Date.now() > cached.expiry) {
    edgeCache.delete(key);
    return null;
  }
  return cached.data;
}

function setCachedResponse(key: string, data: any, ttl: number): void {
  edgeCache.set(key, {
    data,
    expiry: Date.now() + ttl,
  });
}

async function processUsersRequest(request: NextRequest): Promise<Response> {
  // User API processing logic
  return NextResponse.json({ users: [] });
}

async function processAnalyticsRequest(request: NextRequest): Promise<Response> {
  // Analytics API processing logic
  return NextResponse.json({ analytics: {} });
}
```

### Edge Caching Strategy

```typescript
// middleware.ts - Global edge caching
import { NextRequest, NextResponse } from 'next/server';

export function middleware(request: NextRequest) {
  const response = NextResponse.next();

  // Set cache headers based on content type
  const pathname = request.nextUrl.pathname;

  if (pathname.startsWith('/api/')) {
    // API endpoints: shorter cache
    response.headers.set(
      'Cache-Control',
      'public, max-age=60, stale-while-revalidate=300'
    );
  } else if (pathname.match(/\.(js|css|webp|woff2)$/)) {
    // Static assets: long cache
    response.headers.set(
      'Cache-Control',
      'public, max-age=31536000, immutable'
    );
  } else {
    // HTML pages: moderate cache with revalidation
    response.headers.set(
      'Cache-Control',
      'public, max-age=3600, stale-while-revalidate=86400'
    );
  }

  // Add cache validation headers
  response.headers.set('ETag', generateETag(request));
  response.headers.set('X-Edge-Location', 'vercel-cache');

  return response;
}

function generateETag(request: NextRequest): string {
  return `"${Date.now().toString(36)}"`;
}

export const config = {
  matcher: ['/((?!_next/static|favicon.ico).*)'],
};
```

---

**Module**: Edge Functions & Security
**Parent Skill**: moai-baas-vercel-ext
**Last Updated**: 2025-11-21
