# Cloudflare Workers 완전 가이드

## 개요

Cloudflare Workers는 전 세계 300개 이상의 Edge 로케이션에서 실행되는 서버리스 실행 환경입니다. V8 JavaScript 엔진을 기반으로 하며, 초고속 Cold Start와 무제한 확장성을 제공합니다.

**핵심 장점**:
- **0ms Cold Start**: 격리된 V8 컨텍스트로 즉시 시작
- **글로벌 Edge Network**: 전 세계 300+ 도시에서 실행
- **무제한 확장**: 동시 요청 수 제한 없음
- **통합 플랫폼**: D1, KV, R2, Durable Objects 완벽 통합
- **개발자 친화적**: TypeScript 완벽 지원

## 왜 Cloudflare Workers인가?

### 1. Edge Computing의 진정한 힘 (Pattern G)

```typescript
// worker.ts - Edge에서 실행되는 API
export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    // 사용자와 가장 가까운 Edge에서 실행
    const url = new URL(request.url)

    // 지리적 위치 정보
    const country = request.cf?.country
    const city = request.cf?.city
    const latitude = request.cf?.latitude
    const longitude = request.cf?.longitude

    return Response.json({
      message: `Hello from ${city}, ${country}!`,
      location: { latitude, longitude },
      edge: request.cf?.colo, // Edge 데이터센터 코드
    })
  },
}
```

### 2. D1 Database와의 완벽한 통합

```typescript
// schema.sql
CREATE TABLE users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  email TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);

// worker.ts
export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const { pathname } = new URL(request.url)

    if (pathname === '/api/users' && request.method === 'GET') {
      // D1 쿼리 (Edge에서 직접 실행)
      const { results } = await env.DB.prepare(
        'SELECT * FROM users ORDER BY created_at DESC LIMIT 10'
      ).all()

      return Response.json(results)
    }

    if (pathname === '/api/users' && request.method === 'POST') {
      const { email, name } = await request.json()

      // 트랜잭션
      await env.DB.batch([
        env.DB.prepare('INSERT INTO users (email, name) VALUES (?, ?)')
          .bind(email, name),
        env.DB.prepare('INSERT INTO activity_log (action) VALUES (?)')
          .bind(`User ${email} created`),
      ])

      return Response.json({ success: true }, { status: 201 })
    }

    return Response.json({ error: 'Not found' }, { status: 404 })
  },
}
```

## 주요 기능

### 1. KV Storage (Key-Value)

**초고속 글로벌 캐시**:

```typescript
// worker.ts
interface Env {
  CACHE: KVNamespace
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url)
    const cacheKey = url.pathname

    // 캐시 확인
    const cached = await env.CACHE.get(cacheKey, { type: 'json' })
    if (cached) {
      return Response.json(cached, {
        headers: { 'X-Cache': 'HIT' },
      })
    }

    // 데이터 가져오기
    const data = await fetchData(url.pathname)

    // 캐시 저장 (TTL: 1시간)
    await env.CACHE.put(
      cacheKey,
      JSON.stringify(data),
      { expirationTtl: 3600 }
    )

    return Response.json(data, {
      headers: { 'X-Cache': 'MISS' },
    })
  },
}
```

### 2. R2 Storage (Object Storage)

**S3 호환 객체 스토리지 (무료 Egress)**:

```typescript
// worker.ts
interface Env {
  BUCKET: R2Bucket
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url)

    // 파일 업로드
    if (request.method === 'PUT') {
      const key = url.pathname.slice(1)
      await env.BUCKET.put(key, request.body, {
        httpMetadata: {
          contentType: request.headers.get('content-type') || 'application/octet-stream',
        },
        customMetadata: {
          uploadedBy: request.headers.get('x-user-id') || 'anonymous',
          uploadedAt: new Date().toISOString(),
        },
      })

      return Response.json({ success: true })
    }

    // 파일 다운로드
    if (request.method === 'GET') {
      const key = url.pathname.slice(1)
      const object = await env.BUCKET.get(key)

      if (!object) {
        return Response.json({ error: 'Not found' }, { status: 404 })
      }

      return new Response(object.body, {
        headers: {
          'Content-Type': object.httpMetadata?.contentType || 'application/octet-stream',
          'ETag': object.httpEtag,
          'Cache-Control': 'public, max-age=31536000',
        },
      })
    }

    return Response.json({ error: 'Method not allowed' }, { status: 405 })
  },
}
```

### 3. Durable Objects

**상태를 가진 서버리스 객체**:

```typescript
// durable-object.ts
export class Counter {
  state: DurableObjectState
  count: number = 0

  constructor(state: DurableObjectState) {
    this.state = state
    // 상태 복원
    this.state.blockConcurrencyWhile(async () => {
      this.count = (await this.state.storage.get('count')) || 0
    })
  }

  async fetch(request: Request): Promise<Response> {
    const url = new URL(request.url)

    if (url.pathname === '/increment') {
      this.count++
      await this.state.storage.put('count', this.count)
      return Response.json({ count: this.count })
    }

    if (url.pathname === '/decrement') {
      this.count--
      await this.state.storage.put('count', this.count)
      return Response.json({ count: this.count })
    }

    return Response.json({ count: this.count })
  }
}

// worker.ts
export { Counter } from './durable-object'

interface Env {
  COUNTER: DurableObjectNamespace
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    // Durable Object 인스턴스 가져오기
    const id = env.COUNTER.idFromName('global-counter')
    const stub = env.COUNTER.get(id)

    // 요청 전달
    return stub.fetch(request)
  },
}
```

### 4. Queues (메시지 큐)

**비동기 작업 처리**:

```typescript
// producer.ts - 메시지 생성자
interface Env {
  EMAIL_QUEUE: Queue
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const { email, subject, body } = await request.json()

    // 큐에 메시지 추가
    await env.EMAIL_QUEUE.send({
      email,
      subject,
      body,
      timestamp: Date.now(),
    })

    return Response.json({ queued: true })
  },
}

// consumer.ts - 메시지 소비자
export default {
  async queue(batch: MessageBatch<EmailMessage>, env: Env): Promise<void> {
    // 배치로 메시지 처리
    for (const message of batch.messages) {
      try {
        await sendEmail(message.body)
        message.ack() // 처리 완료
      } catch (error) {
        message.retry() // 재시도
      }
    }
  },
}
```

### 5. Analytics Engine

**실시간 분석 데이터 수집**:

```typescript
// worker.ts
interface Env {
  ANALYTICS: AnalyticsEngineDataset
}

export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const start = Date.now()

    // 요청 처리
    const response = await handleRequest(request)

    const duration = Date.now() - start

    // 분석 데이터 기록
    env.ANALYTICS.writeDataPoint({
      indexes: [
        request.cf?.country || 'unknown',
      ],
      blobs: [
        request.url,
        request.method,
      ],
      doubles: [
        duration,
        response.status,
      ],
    })

    return response
  },
}
```

## 시작하기

### 1. Wrangler CLI 설치

```bash
npm install -g wrangler

# 로그인
wrangler login

# 새 프로젝트 생성
npm create cloudflare@latest my-worker
cd my-worker
```

### 2. 프로젝트 구조

```
my-worker/
├── src/
│   └── index.ts          # Worker 코드
├── wrangler.toml         # 설정 파일
├── package.json
└── tsconfig.json
```

### 3. 설정 파일 (`wrangler.toml`)

```toml
name = "my-worker"
main = "src/index.ts"
compatibility_date = "2024-01-01"

# KV 네임스페이스
[[kv_namespaces]]
binding = "CACHE"
id = "your-kv-namespace-id"
preview_id = "your-preview-kv-namespace-id"

# D1 데이터베이스
[[d1_databases]]
binding = "DB"
database_name = "my-database"
database_id = "your-database-id"

# R2 버킷
[[r2_buckets]]
binding = "BUCKET"
bucket_name = "my-bucket"

# Durable Objects
[[durable_objects.bindings]]
name = "COUNTER"
class_name = "Counter"
script_name = "my-worker"

# 환경 변수
[vars]
ENVIRONMENT = "production"

# 비밀 환경 변수 (wrangler secret put으로 설정)
# API_KEY = "..."

# 라우트 설정
routes = [
  { pattern = "example.com/api/*", zone_name = "example.com" }
]
```

### 4. 로컬 개발

```bash
# 개발 서버 시작
wrangler dev

# 특정 포트로 시작
wrangler dev --port 8787

# 리모트 모드 (실제 KV, D1 사용)
wrangler dev --remote
```

### 5. 배포

```bash
# 프로덕션 배포
wrangler deploy

# 특정 환경으로 배포
wrangler deploy --env staging
```

## 사용 가이드

### Pattern G: Edge Computing

```typescript
// worker.ts - 글로벌 API Gateway
interface Env {
  CACHE: KVNamespace
  DB: D1Database
  RATE_LIMIT: DurableObjectNamespace
}

export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    const url = new URL(request.url)

    // 1. CORS 처리
    if (request.method === 'OPTIONS') {
      return handleCORS(request)
    }

    // 2. Rate Limiting (Durable Objects)
    const rateLimitOk = await checkRateLimit(request, env)
    if (!rateLimitOk) {
      return Response.json(
        { error: 'Too many requests' },
        { status: 429 }
      )
    }

    // 3. 캐시 확인 (KV)
    const cacheKey = `${request.method}:${url.pathname}`
    const cached = await env.CACHE.get(cacheKey)
    if (cached) {
      return new Response(cached, {
        headers: {
          'Content-Type': 'application/json',
          'X-Cache': 'HIT',
          'Access-Control-Allow-Origin': '*',
        },
      })
    }

    // 4. 데이터베이스 쿼리 (D1)
    const data = await fetchFromDatabase(url.pathname, env.DB)

    // 5. 캐시 저장 (백그라운드)
    ctx.waitUntil(
      env.CACHE.put(cacheKey, JSON.stringify(data), {
        expirationTtl: 300, // 5분
      })
    )

    return Response.json(data, {
      headers: {
        'X-Cache': 'MISS',
        'Access-Control-Allow-Origin': '*',
      },
    })
  },
}

async function checkRateLimit(request: Request, env: Env): Promise<boolean> {
  const ip = request.headers.get('cf-connecting-ip') || 'unknown'
  const id = env.RATE_LIMIT.idFromName(ip)
  const stub = env.RATE_LIMIT.get(id)

  const response = await stub.fetch('https://dummy/check')
  return response.ok
}

function handleCORS(request: Request): Response {
  return new Response(null, {
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type, Authorization',
      'Access-Control-Max-Age': '86400',
    },
  })
}
```

## 코드 예제

### 1. WebSocket 프록시

```typescript
export default {
  async fetch(request: Request): Promise<Response> {
    const upgradeHeader = request.headers.get('Upgrade')
    if (upgradeHeader !== 'websocket') {
      return Response.json({ error: 'Expected websocket' }, { status: 400 })
    }

    const webSocketPair = new WebSocketPair()
    const [client, server] = Object.values(webSocketPair)

    server.accept()

    // 메시지 처리
    server.addEventListener('message', (event) => {
      console.log('Received:', event.data)
      server.send(`Echo: ${event.data}`)
    })

    return new Response(null, {
      status: 101,
      webSocket: client,
    })
  },
}
```

### 2. Image Resizing

```typescript
export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    const url = new URL(request.url)
    const width = parseInt(url.searchParams.get('w') || '800')
    const quality = parseInt(url.searchParams.get('q') || '85')

    // R2에서 원본 이미지 가져오기
    const imageKey = url.pathname.slice(1)
    const object = await env.BUCKET.get(imageKey)

    if (!object) {
      return Response.json({ error: 'Not found' }, { status: 404 })
    }

    // Cloudflare Image Resizing 사용
    const resized = await fetch(request.url, {
      cf: {
        image: {
          width,
          quality,
          format: 'auto',
        },
      },
    })

    return resized
  },
}
```

### 3. A/B Testing

```typescript
export default {
  async fetch(request: Request): Promise<Response> {
    // 쿠키 확인
    const cookie = request.headers.get('Cookie')
    let variant = cookie?.match(/variant=([ab])/)?.[1]

    if (!variant) {
      // 무작위 할당
      variant = Math.random() < 0.5 ? 'a' : 'b'
    }

    // Variant에 따라 다른 Origin으로 프록시
    const origin = variant === 'a'
      ? 'https://variant-a.example.com'
      : 'https://variant-b.example.com'

    const response = await fetch(origin + new URL(request.url).pathname, {
      headers: request.headers,
    })

    // Variant 쿠키 설정
    const newResponse = new Response(response.body, response)
    newResponse.headers.set(
      'Set-Cookie',
      `variant=${variant}; Max-Age=2592000; Path=/; SameSite=Lax`
    )

    return newResponse
  },
}
```

## Best Practices

### 1. CPU 시간 제한 준수

Workers는 10ms-50ms CPU 시간 제한이 있습니다:

```typescript
// ✅ 올바른 방법 - I/O는 CPU 시간에 포함되지 않음
const data = await fetch('https://api.example.com/data')

// ❌ 잘못된 방법 - 무거운 계산은 CPU 시간 소모
for (let i = 0; i < 1000000; i++) {
  // 복잡한 계산...
}
```

### 2. 효율적인 KV 사용

```typescript
// ✅ 배치 읽기
const values = await env.CACHE.get(['key1', 'key2', 'key3'], { type: 'json' })

// ❌ 순차 읽기
const value1 = await env.CACHE.get('key1')
const value2 = await env.CACHE.get('key2')
const value3 = await env.CACHE.get('key3')
```

### 3. 백그라운드 작업

```typescript
export default {
  async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
    // 응답 즉시 반환
    const response = Response.json({ success: true })

    // 백그라운드에서 처리
    ctx.waitUntil(
      (async () => {
        await logAnalytics(request, env)
        await updateCache(request, env)
      })()
    )

    return response
  },
}
```

## 문제 해결

### 1. CPU 시간 초과

```typescript
// 긴 작업을 여러 Workers로 분할
export default {
  async fetch(request: Request, env: Env): Promise<Response> {
    // 작업을 큐에 추가
    await env.WORK_QUEUE.send({
      task: 'process-data',
      data: largeDataset,
    })

    return Response.json({ queued: true })
  },
}
```

### 2. KV 쓰기 지연

KV는 eventually consistent이므로 쓰기 후 즉시 읽으면 이전 값이 반환될 수 있습니다:

```typescript
// 중요한 데이터는 Durable Objects 사용
const id = env.IMPORTANT_DATA.idFromName('key')
const stub = env.IMPORTANT_DATA.get(id)
await stub.fetch('https://dummy/write', { method: 'POST', body: data })
```

## 다음 단계

- [Vercel 가이드](/ko/skills/baas/vercel) - Next.js와 함께 사용
- [Railway 가이드](/ko/skills/baas/railway) - 더 복잡한 백엔드가 필요하다면
- [Pattern G: Edge Computing](/ko/skills/patterns/pattern-g) - 아키텍처 상세 가이드
- [BaaS 개요](/ko/skills/baas) - 다른 플랫폼 비교

## 참고 자료

- [Cloudflare Workers 문서](https://developers.cloudflare.com/workers/)
- [Wrangler CLI 가이드](https://developers.cloudflare.com/workers/wrangler/)
- [D1 Database 문서](https://developers.cloudflare.com/d1/)
- [Durable Objects 가이드](https://developers.cloudflare.com/durable-objects/)
