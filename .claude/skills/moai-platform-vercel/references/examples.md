# Vercel Platform Examples

## Edge Functions

### Geo-Based Pricing

```typescript
// app/api/pricing/route.ts
export const runtime = 'edge'

const PRICING_BY_REGION: Record<string, { currency: string; multiplier: number }> = {
  US: { currency: 'USD', multiplier: 1.0 },
  EU: { currency: 'EUR', multiplier: 0.92 },
  GB: { currency: 'GBP', multiplier: 0.79 },
  JP: { currency: 'JPY', multiplier: 149.5 },
  KR: { currency: 'KRW', multiplier: 1320 }
}

export async function GET(request: Request) {
  const country = request.geo?.country ?? 'US'

  // Map country to region
  const region = country === 'GB' ? 'GB' :
                 ['DE', 'FR', 'IT', 'ES', 'NL'].includes(country) ? 'EU' :
                 country

  const pricing = PRICING_BY_REGION[region] ?? PRICING_BY_REGION.US

  return Response.json({
    country,
    currency: pricing.currency,
    basePrice: 99,
    localPrice: Math.round(99 * pricing.multiplier)
  }, {
    headers: {
      'Cache-Control': 'public, s-maxage=3600'
    }
  })
}
```

### Feature Flags at Edge

```typescript
// middleware.ts
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

const FEATURE_FLAGS = {
  newCheckout: { enabled: true, percentage: 50 },
  darkMode: { enabled: true, percentage: 100 },
  betaFeature: { enabled: false, percentage: 0 }
}

export function middleware(request: NextRequest) {
  const response = NextResponse.next()

  // Get or create user ID for consistent flag assignment
  let userId = request.cookies.get('user-id')?.value
  if (!userId) {
    userId = crypto.randomUUID()
    response.cookies.set('user-id', userId, { maxAge: 365 * 24 * 60 * 60 })
  }

  // Compute feature flags
  const flags: Record<string, boolean> = {}
  for (const [name, config] of Object.entries(FEATURE_FLAGS)) {
    if (!config.enabled) {
      flags[name] = false
    } else {
      // Deterministic assignment based on user ID
      const hash = hashString(userId + name)
      flags[name] = (hash % 100) < config.percentage
    }
  }

  // Set flags as header for server components
  response.headers.set('x-feature-flags', JSON.stringify(flags))

  return response
}

function hashString(str: string): number {
  let hash = 0
  for (let i = 0; i < str.length; i++) {
    hash = ((hash << 5) - hash) + str.charCodeAt(i)
    hash |= 0
  }
  return Math.abs(hash)
}
```

## ISR Patterns

### E-Commerce Product Page

```typescript
// app/products/[slug]/page.tsx
import { notFound } from 'next/navigation'
import { revalidateTag } from 'next/cache'

export const revalidate = 300 // 5 minutes base revalidation

interface Product {
  id: string
  slug: string
  name: string
  price: number
  inventory: number
  images: string[]
}

export async function generateStaticParams() {
  const products = await fetch('https://api.store.com/products/popular', {
    next: { revalidate: 3600 }
  }).then(r => r.json())

  return products.map((p: Product) => ({ slug: p.slug }))
}

export async function generateMetadata({ params }: { params: { slug: string } }) {
  const product = await getProduct(params.slug)
  if (!product) return {}

  return {
    title: `${product.name} | Store`,
    description: `Buy ${product.name} for $${product.price}`,
    openGraph: {
      images: product.images[0] ? [product.images[0]] : []
    }
  }
}

async function getProduct(slug: string): Promise<Product | null> {
  const res = await fetch(`https://api.store.com/products/${slug}`, {
    next: { tags: [`product-${slug}`, 'products'] }
  })
  if (!res.ok) return null
  return res.json()
}

export default async function ProductPage({ params }: { params: { slug: string } }) {
  const product = await getProduct(params.slug)
  if (!product) notFound()

  return (
    <div>
      <h1>{product.name}</h1>
      <p>${product.price}</p>
      <p>{product.inventory > 0 ? 'In Stock' : 'Out of Stock'}</p>
    </div>
  )
}
```

### Webhook Handler for On-Demand Revalidation

```typescript
// app/api/webhooks/product-update/route.ts
import { revalidateTag } from 'next/cache'
import { NextRequest } from 'next/server'

export async function POST(request: NextRequest) {
  const signature = request.headers.get('x-webhook-signature')
  const body = await request.text()

  // Verify webhook signature
  if (!verifySignature(body, signature)) {
    return Response.json({ error: 'Invalid signature' }, { status: 401 })
  }

  const payload = JSON.parse(body)

  switch (payload.event) {
    case 'product.updated':
    case 'product.deleted':
      revalidateTag(`product-${payload.data.slug}`)
      revalidateTag('products')
      break

    case 'inventory.changed':
      revalidateTag(`product-${payload.data.productSlug}`)
      break

    case 'category.updated':
      revalidateTag(`category-${payload.data.slug}`)
      revalidateTag('categories')
      break
  }

  return Response.json({ revalidated: true })
}

function verifySignature(body: string, signature: string | null): boolean {
  if (!signature) return false
  const crypto = require('crypto')
  const expected = crypto
    .createHmac('sha256', process.env.WEBHOOK_SECRET!)
    .update(body)
    .digest('hex')
  return signature === expected
}
```

## Vercel KV Patterns

### Session Store

```typescript
// lib/session.ts
import { kv } from '@vercel/kv'
import { cookies } from 'next/headers'

interface Session {
  userId: string
  email: string
  role: string
  createdAt: number
}

export async function getSession(): Promise<Session | null> {
  const sessionId = cookies().get('session-id')?.value
  if (!sessionId) return null

  return kv.get<Session>(`session:${sessionId}`)
}

export async function createSession(user: { id: string; email: string; role: string }): Promise<string> {
  const sessionId = crypto.randomUUID()
  const session: Session = {
    userId: user.id,
    email: user.email,
    role: user.role,
    createdAt: Date.now()
  }

  await kv.set(`session:${sessionId}`, session, { ex: 7 * 24 * 60 * 60 }) // 7 days

  cookies().set('session-id', sessionId, {
    httpOnly: true,
    secure: process.env.NODE_ENV === 'production',
    sameSite: 'lax',
    maxAge: 7 * 24 * 60 * 60
  })

  return sessionId
}

export async function destroySession(): Promise<void> {
  const sessionId = cookies().get('session-id')?.value
  if (sessionId) {
    await kv.del(`session:${sessionId}`)
    cookies().delete('session-id')
  }
}
```

### Rate Limiter

```typescript
// lib/rate-limit.ts
import { kv } from '@vercel/kv'

interface RateLimitResult {
  success: boolean
  remaining: number
  reset: number
}

export async function rateLimit(
  identifier: string,
  limit: number = 100,
  window: number = 60
): Promise<RateLimitResult> {
  const key = `ratelimit:${identifier}`
  const now = Math.floor(Date.now() / 1000)

  // Get current count
  const data = await kv.get<{ count: number; reset: number }>(key)

  if (!data || data.reset < now) {
    // New window
    await kv.set(key, { count: 1, reset: now + window }, { ex: window })
    return { success: true, remaining: limit - 1, reset: now + window }
  }

  if (data.count >= limit) {
    return { success: false, remaining: 0, reset: data.reset }
  }

  // Increment count
  await kv.set(key, { count: data.count + 1, reset: data.reset }, { ex: data.reset - now })
  return { success: true, remaining: limit - data.count - 1, reset: data.reset }
}

// Usage in API route
export async function GET(request: Request) {
  const ip = request.headers.get('x-forwarded-for') ?? 'anonymous'
  const result = await rateLimit(ip, 100, 60) // 100 requests per minute

  if (!result.success) {
    return Response.json(
      { error: 'Rate limit exceeded' },
      {
        status: 429,
        headers: {
          'X-RateLimit-Remaining': '0',
          'X-RateLimit-Reset': result.reset.toString()
        }
      }
    )
  }

  return Response.json({ data: 'success' })
}
```

## Vercel Blob Patterns

### Image Upload with Optimization

```typescript
// app/api/upload-image/route.ts
import { put } from '@vercel/blob'
import { NextRequest } from 'next/server'
import sharp from 'sharp'

export async function POST(request: NextRequest) {
  const formData = await request.formData()
  const file = formData.get('file') as File

  if (!file) {
    return Response.json({ error: 'No file provided' }, { status: 400 })
  }

  // Validate file type
  if (!file.type.startsWith('image/')) {
    return Response.json({ error: 'Invalid file type' }, { status: 400 })
  }

  // Convert to buffer and optimize
  const buffer = Buffer.from(await file.arrayBuffer())
  const optimized = await sharp(buffer)
    .resize(1200, 1200, { fit: 'inside', withoutEnlargement: true })
    .webp({ quality: 80 })
    .toBuffer()

  // Upload optimized image
  const blob = await put(`images/${Date.now()}.webp`, optimized, {
    access: 'public',
    contentType: 'image/webp'
  })

  // Generate thumbnail
  const thumbnail = await sharp(buffer)
    .resize(200, 200, { fit: 'cover' })
    .webp({ quality: 70 })
    .toBuffer()

  const thumbBlob = await put(`thumbnails/${Date.now()}.webp`, thumbnail, {
    access: 'public',
    contentType: 'image/webp'
  })

  return Response.json({
    url: blob.url,
    thumbnail: thumbBlob.url
  })
}
```

## Vercel Postgres Patterns

### Full-Text Search

```typescript
// app/api/search/route.ts
import { sql } from '@vercel/postgres'

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url)
  const query = searchParams.get('q')

  if (!query) {
    return Response.json({ error: 'Query required' }, { status: 400 })
  }

  const { rows } = await sql`
    SELECT id, title, description,
           ts_rank(search_vector, websearch_to_tsquery('english', ${query})) as rank
    FROM products
    WHERE search_vector @@ websearch_to_tsquery('english', ${query})
    ORDER BY rank DESC
    LIMIT 20
  `

  return Response.json(rows)
}
```

### Pagination with Cursors

```typescript
// app/api/posts/route.ts
import { sql } from '@vercel/postgres'

export async function GET(request: Request) {
  const { searchParams } = new URL(request.url)
  const cursor = searchParams.get('cursor')
  const limit = Math.min(parseInt(searchParams.get('limit') ?? '20'), 100)

  let query
  if (cursor) {
    query = sql`
      SELECT id, title, created_at
      FROM posts
      WHERE created_at < ${cursor}
      ORDER BY created_at DESC
      LIMIT ${limit + 1}
    `
  } else {
    query = sql`
      SELECT id, title, created_at
      FROM posts
      ORDER BY created_at DESC
      LIMIT ${limit + 1}
    `
  }

  const { rows } = await query
  const hasMore = rows.length > limit
  const posts = hasMore ? rows.slice(0, -1) : rows
  const nextCursor = hasMore ? posts[posts.length - 1].created_at : null

  return Response.json({
    posts,
    nextCursor,
    hasMore
  })
}
```

## GitHub Actions Complete Workflow

```yaml
# .github/workflows/vercel.yml
name: Vercel Deployment

on:
  push:
    branches: [main, develop]
  pull_request:
    types: [opened, synchronize, reopened]

env:
  VERCEL_ORG_ID: ${{ secrets.VERCEL_ORG_ID }}
  VERCEL_PROJECT_ID: ${{ secrets.VERCEL_PROJECT_ID }}

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v2
      - uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: pnpm

      - run: pnpm install
      - run: pnpm test
      - run: pnpm lint

  deploy-preview:
    needs: test
    if: github.event_name == 'pull_request'
    runs-on: ubuntu-latest
    environment:
      name: Preview
      url: ${{ steps.deploy.outputs.url }}
    steps:
      - uses: actions/checkout@v4
      - run: npm i -g vercel@latest
      - run: vercel pull --yes --environment=preview --token=${{ secrets.VERCEL_TOKEN }}
      - run: vercel build --token=${{ secrets.VERCEL_TOKEN }}
      - id: deploy
        run: echo "url=$(vercel deploy --prebuilt --token=${{ secrets.VERCEL_TOKEN }})" >> $GITHUB_OUTPUT

  deploy-production:
    needs: test
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    environment:
      name: Production
      url: https://your-domain.com
    steps:
      - uses: actions/checkout@v4
      - run: npm i -g vercel@latest
      - run: vercel pull --yes --environment=production --token=${{ secrets.VERCEL_TOKEN }}
      - run: vercel build --prod --token=${{ secrets.VERCEL_TOKEN }}
      - run: vercel deploy --prebuilt --prod --token=${{ secrets.VERCEL_TOKEN }}
```
