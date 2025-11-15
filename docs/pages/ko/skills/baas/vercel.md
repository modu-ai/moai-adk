# Vercel Edge Platform 완전 가이드

## 개요

Vercel은 Next.js를 개발한 회사에서 제공하는 엔터프라이즈급 Edge 플랫폼입니다. 전 세계에 분산된 Edge Network를 통해 초고속 성능, 무한 확장성, 그리고 개발자 친화적인 배포 경험을 제공합니다.

**핵심 장점**:
- **Edge-First Architecture**: 사용자와 가장 가까운 Edge에서 실행
- **Zero Configuration**: 설정 없이 즉시 배포 가능
- **Instant Rollback**: 문제 발생 시 1초 안에 롤백
- **Preview Deployments**: 모든 PR에 자동 미리보기 URL 생성
- **Built-in Analytics**: 웹 바이탈 자동 수집

## 왜 Vercel인가?

### 1. Multi-tenant SaaS에 최적화 (Pattern A)

```typescript
// middleware.ts - 테넌트 라우팅 (Edge에서 실행)
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export function middleware(request: NextRequest) {
  const hostname = request.headers.get('host') || ''

  // 서브도메인 추출
  const subdomain = hostname.split('.')[0]

  // 테넌트 정보를 헤더에 추가
  const requestHeaders = new Headers(request.headers)
  requestHeaders.set('x-tenant-id', subdomain)

  return NextResponse.next({
    request: {
      headers: requestHeaders,
    },
  })
}

export const config = {
  matcher: [
    '/((?!api|_next/static|_next/image|favicon.ico).*)',
  ],
}
```

### 2. Serverless API의 강력함 (Pattern B)

```typescript
// app/api/orders/route.ts
import { NextRequest, NextResponse } from 'next/server'
import { auth } from '@/lib/auth'
import { prisma } from '@/lib/prisma'

export const runtime = 'nodejs' // 또는 'edge'
export const maxDuration = 30 // 최대 실행 시간 (초)

export async function POST(request: NextRequest) {
  try {
    // 인증 확인
    const session = await auth(request)
    if (!session) {
      return NextResponse.json(
        { error: 'Unauthorized' },
        { status: 401 }
      )
    }

    // 요청 데이터 파싱
    const body = await request.json()

    // 주문 생성
    const order = await prisma.order.create({
      data: {
        userId: session.userId,
        items: body.items,
        total: body.total,
        status: 'pending',
      },
    })

    return NextResponse.json(order, { status: 201 })
  } catch (error) {
    console.error('Order creation failed:', error)
    return NextResponse.json(
      { error: 'Internal Server Error' },
      { status: 500 }
    )
  }
}
```

## 주요 기능

### 1. Edge Functions

**초고속 글로벌 실행 환경**:

```typescript
// app/api/geo/route.ts
export const runtime = 'edge'

export async function GET(request: Request) {
  // Vercel Edge Network에서 실행
  const geo = request.headers.get('x-vercel-ip-country')
  const city = request.headers.get('x-vercel-ip-city')

  return Response.json({
    country: geo,
    city: city,
    message: `Hello from ${city}, ${geo}!`,
  })
}
```

### 2. Edge Middleware

**요청 인터셉션 및 수정**:

```typescript
// middleware.ts - A/B 테스팅
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export function middleware(request: NextRequest) {
  // 쿠키에서 실험 그룹 확인
  let bucket = request.cookies.get('bucket')?.value

  if (!bucket) {
    // 무작위 할당
    bucket = Math.random() < 0.5 ? 'a' : 'b'
    const response = NextResponse.next()
    response.cookies.set('bucket', bucket, {
      maxAge: 60 * 60 * 24 * 30, // 30일
    })
    return response
  }

  // 그룹에 따라 다른 페이지로 리라이트
  if (bucket === 'b') {
    return NextResponse.rewrite(new URL('/variant-b', request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: '/experiment',
}
```

### 3. Incremental Static Regeneration (ISR)

**정적 + 동적의 장점 결합**:

```typescript
// app/products/[id]/page.tsx
import { prisma } from '@/lib/prisma'

interface PageProps {
  params: { id: string }
}

export const revalidate = 60 // 60초마다 재생성

export async function generateStaticParams() {
  const products = await prisma.product.findMany({
    take: 100, // 상위 100개만 빌드 시 생성
  })

  return products.map((product) => ({
    id: product.id,
  }))
}

export default async function ProductPage({ params }: PageProps) {
  const product = await prisma.product.findUnique({
    where: { id: params.id },
  })

  if (!product) {
    return <div>Product not found</div>
  }

  return (
    <div>
      <h1>{product.name}</h1>
      <p>{product.description}</p>
      <p>${product.price}</p>
    </div>
  )
}
```

### 4. Preview Deployments

**모든 PR에 자동 미리보기**:

```yaml
# .github/workflows/preview.yml
name: Deploy Preview

on:
  pull_request:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Deploy to Vercel
        uses: amondnet/vercel-action@v25
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
          scope: ${{ secrets.VERCEL_SCOPE }}
```

### 5. Environment Variables

**안전한 환경 변수 관리**:

```bash
# Vercel CLI로 환경 변수 설정
vercel env add DATABASE_URL production
vercel env add API_KEY preview development

# .env.local (로컬 개발)
DATABASE_URL="postgresql://..."
NEXT_PUBLIC_API_URL="http://localhost:3000"
```

```typescript
// lib/env.ts - 타입 안전한 환경 변수
import { z } from 'zod'

const envSchema = z.object({
  DATABASE_URL: z.string().url(),
  API_KEY: z.string().min(20),
  NEXT_PUBLIC_API_URL: z.string().url(),
})

export const env = envSchema.parse(process.env)
```

## 시작하기

### 1. Vercel CLI 설치

```bash
npm install -g vercel

# 로그인
vercel login

# 프로젝트 연결
vercel link
```

### 2. Next.js 프로젝트 생성

```bash
npx create-next-app@latest my-vercel-app --typescript --tailwind --app
cd my-vercel-app
```

### 3. 첫 배포

```bash
# 프로덕션 배포
vercel --prod

# 미리보기 배포
vercel
```

### 4. 프로젝트 설정 (`vercel.json`)

```json
{
  "buildCommand": "npm run build",
  "devCommand": "npm run dev",
  "installCommand": "npm install",
  "framework": "nextjs",
  "regions": ["icn1", "sfo1"],
  "env": {
    "DATABASE_URL": "@database-url"
  },
  "build": {
    "env": {
      "NEXT_PUBLIC_API_URL": "https://api.example.com"
    }
  },
  "headers": [
    {
      "source": "/api/(.*)",
      "headers": [
        {
          "key": "Cache-Control",
          "value": "s-maxage=60, stale-while-revalidate"
        }
      ]
    }
  ],
  "redirects": [
    {
      "source": "/old-path",
      "destination": "/new-path",
      "permanent": true
    }
  ],
  "rewrites": [
    {
      "source": "/api/proxy/:path*",
      "destination": "https://external-api.com/:path*"
    }
  ]
}
```

## 사용 가이드

### Pattern A: Multi-tenant SaaS

```typescript
// middleware.ts - 테넌트별 라우팅 및 인증
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { verifyToken } from '@/lib/auth'

export async function middleware(request: NextRequest) {
  const hostname = request.headers.get('host') || ''
  const subdomain = hostname.split('.')[0]

  // 테넌트 검증
  if (subdomain === 'www' || subdomain === 'app') {
    return NextResponse.next()
  }

  // 인증 토큰 확인
  const token = request.cookies.get('auth-token')?.value

  if (!token) {
    return NextResponse.redirect(new URL('/login', request.url))
  }

  try {
    const payload = await verifyToken(token)

    // 테넌트 접근 권한 확인
    if (payload.tenantId !== subdomain) {
      return NextResponse.json(
        { error: 'Forbidden' },
        { status: 403 }
      )
    }

    // 헤더에 테넌트 정보 추가
    const requestHeaders = new Headers(request.headers)
    requestHeaders.set('x-tenant-id', subdomain)
    requestHeaders.set('x-user-id', payload.userId)

    return NextResponse.next({
      request: {
        headers: requestHeaders,
      },
    })
  } catch (error) {
    return NextResponse.redirect(new URL('/login', request.url))
  }
}

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|login|api/auth).*)',
  ],
}
```

```typescript
// app/api/data/route.ts - 테넌트별 데이터 분리
import { NextRequest, NextResponse } from 'next/server'
import { prisma } from '@/lib/prisma'

export async function GET(request: NextRequest) {
  const tenantId = request.headers.get('x-tenant-id')
  const userId = request.headers.get('x-user-id')

  if (!tenantId || !userId) {
    return NextResponse.json(
      { error: 'Unauthorized' },
      { status: 401 }
    )
  }

  const data = await prisma.data.findMany({
    where: {
      tenantId,
      OR: [
        { visibility: 'public' },
        { ownerId: userId },
      ],
    },
  })

  return NextResponse.json(data)
}
```

### Pattern B: Serverless API

```typescript
// app/api/webhooks/stripe/route.ts
import { NextRequest, NextResponse } from 'next/server'
import Stripe from 'stripe'
import { buffer } from 'node:stream/consumers'

const stripe = new Stripe(process.env.STRIPE_SECRET_KEY!, {
  apiVersion: '2023-10-16',
})

export const runtime = 'nodejs'
export const maxDuration = 30

export async function POST(request: NextRequest) {
  const sig = request.headers.get('stripe-signature')!
  const webhookSecret = process.env.STRIPE_WEBHOOK_SECRET!

  let event: Stripe.Event

  try {
    // 원시 body 읽기
    const body = await buffer(request.body!)

    // 서명 검증
    event = stripe.webhooks.constructEvent(
      body,
      sig,
      webhookSecret
    )
  } catch (err) {
    console.error('Webhook signature verification failed:', err)
    return NextResponse.json(
      { error: 'Webhook error' },
      { status: 400 }
    )
  }

  // 이벤트 처리
  switch (event.type) {
    case 'payment_intent.succeeded':
      const paymentIntent = event.data.object
      await handlePaymentSuccess(paymentIntent)
      break

    case 'payment_intent.payment_failed':
      const failedPayment = event.data.object
      await handlePaymentFailure(failedPayment)
      break

    default:
      console.log(`Unhandled event type: ${event.type}`)
  }

  return NextResponse.json({ received: true })
}

async function handlePaymentSuccess(paymentIntent: Stripe.PaymentIntent) {
  // 결제 성공 로직
  console.log('Payment succeeded:', paymentIntent.id)
}

async function handlePaymentFailure(paymentIntent: Stripe.PaymentIntent) {
  // 결제 실패 로직
  console.log('Payment failed:', paymentIntent.id)
}
```

## 코드 예제

### 1. Edge Caching 전략

```typescript
// app/api/products/route.ts
export const runtime = 'edge'
export const dynamic = 'force-static'
export const revalidate = 3600 // 1시간

export async function GET() {
  const products = await fetch('https://api.example.com/products', {
    next: { revalidate: 3600 },
  }).then(res => res.json())

  return Response.json(products, {
    headers: {
      'Cache-Control': 's-maxage=3600, stale-while-revalidate=86400',
    },
  })
}
```

### 2. 이미지 최적화

```typescript
// app/gallery/page.tsx
import Image from 'next/image'

export default function Gallery() {
  return (
    <div className="grid grid-cols-3 gap-4">
      {images.map((img) => (
        <Image
          key={img.id}
          src={img.url}
          alt={img.alt}
          width={400}
          height={300}
          sizes="(max-width: 768px) 100vw, 33vw"
          quality={85}
          placeholder="blur"
          blurDataURL={img.blurUrl}
        />
      ))}
    </div>
  )
}
```

### 3. 스트리밍 SSR

```typescript
// app/dashboard/page.tsx
import { Suspense } from 'react'

async function SlowComponent() {
  const data = await fetch('https://slow-api.com/data')
  return <div>{/* 렌더링 */}</div>
}

export default function Dashboard() {
  return (
    <div>
      <h1>Dashboard</h1>

      {/* 빠른 컴포넌트는 즉시 표시 */}
      <FastComponent />

      {/* 느린 컴포넌트는 준비되는 대로 스트리밍 */}
      <Suspense fallback={<Loading />}>
        <SlowComponent />
      </Suspense>
    </div>
  )
}
```

### 4. Edge Config (실시간 설정 업데이트)

```typescript
// lib/config.ts
import { get } from '@vercel/edge-config'

export async function getFeatureFlags() {
  const flags = await get('feature-flags')
  return flags || {}
}

// middleware.ts
import { getFeatureFlags } from '@/lib/config'

export async function middleware(request: NextRequest) {
  const flags = await getFeatureFlags()

  if (flags.maintenanceMode) {
    return NextResponse.rewrite(new URL('/maintenance', request.url))
  }

  return NextResponse.next()
}
```

### 5. Analytics 통합

```typescript
// app/layout.tsx
import { Analytics } from '@vercel/analytics/react'
import { SpeedInsights } from '@vercel/speed-insights/next'

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="ko">
      <body>
        {children}
        <Analytics />
        <SpeedInsights />
      </body>
    </html>
  )
}
```

## Best Practices

### 1. Edge vs Node.js Runtime 선택

**Edge Runtime 사용 시기**:
- 짧은 실행 시간 (<50ms)
- 전 세계 사용자 대상
- 간단한 로직 (인증, 리다이렉트)

**Node.js Runtime 사용 시기**:
- 복잡한 로직
- 데이터베이스 연결 필요
- 큰 npm 패키지 사용

### 2. 캐싱 전략

```typescript
// 정적 데이터 - 무한 캐싱
export const dynamic = 'force-static'

// 동적 데이터 - 재검증 주기 설정
export const revalidate = 60

// 실시간 데이터 - 캐싱 안 함
export const dynamic = 'force-dynamic'
```

### 3. 환경 변수 보안

```typescript
// ✅ 올바른 방법
const apiKey = process.env.API_KEY // 서버에서만 접근

// ❌ 잘못된 방법
const apiKey = process.env.NEXT_PUBLIC_API_KEY // 클라이언트에 노출
```

### 4. 배포 전 체크리스트

```bash
# 1. 빌드 테스트
npm run build

# 2. 타입 체크
npm run type-check

# 3. 린트 검사
npm run lint

# 4. 환경 변수 확인
vercel env ls

# 5. 미리보기 배포
vercel

# 6. 프로덕션 배포
vercel --prod
```

## 문제 해결

### 1. 빌드 시간 초과

```json
// vercel.json
{
  "builds": [
    {
      "src": "package.json",
      "use": "@vercel/node",
      "config": {
        "maxLambdaSize": "50mb",
        "includeFiles": ["public/**"]
      }
    }
  ]
}
```

### 2. 메모리 부족

```typescript
// next.config.js
module.exports = {
  experimental: {
    // 메모리 사용량 줄이기
    optimizeCss: true,
    optimizePackageImports: ['@mui/material', 'lodash'],
  },
}
```

### 3. Cold Start 최적화

```typescript
// 연결 풀링
import { Pool } from '@neondatabase/serverless'

const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
  max: 1, // Serverless에서는 1개 연결만 유지
})

export async function query(text: string, params: any[]) {
  const client = await pool.connect()
  try {
    return await client.query(text, params)
  } finally {
    client.release()
  }
}
```

## 다음 단계

- [Railway 가이드](/ko/skills/baas/railway) - 더 많은 제어가 필요하다면
- [Cloudflare Workers 가이드](/ko/skills/baas/cloudflare) - Edge Computing 전문
- [BaaS 개요](/ko/skills/baas) - 다른 플랫폼 비교
- [Pattern A: Multi-tenant SaaS](/ko/skills/patterns/pattern-a) - 아키텍처 상세 가이드
- [Pattern B: Serverless API](/ko/skills/patterns/pattern-b) - API 설계 패턴

## 참고 자료

- [Vercel 공식 문서](https://vercel.com/docs)
- [Next.js 문서](https://nextjs.org/docs)
- [Edge Functions 가이드](https://vercel.com/docs/functions/edge-functions)
- [Vercel CLI 문서](https://vercel.com/docs/cli)
