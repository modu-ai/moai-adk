# Railway 올인원 플랫폼 완전 가이드

## 개요

Railway는 개발자가 애플리케이션을 빌드, 배포, 스케일링하는 데 필요한 모든 것을 제공하는 통합 플랫폼입니다. Infrastructure as Code, 자동 배포, 통합 모니터링, 그리고 사용한 만큼만 지불하는 가격 정책으로 현대적인 개발 경험을 제공합니다.

**핵심 장점**:
- **Zero Config**: Nixpacks로 자동 빌드 및 배포
- **All-in-One**: 데이터베이스, Redis, Cron, Workers 통합
- **Instant Rollback**: 1초 안에 이전 버전으로 복구
- **Built-in Monitoring**: 실시간 로그, 메트릭, 알림
- **팀 협업**: 환경별 접근 제어, Shared Variables

## 왜 Railway인가?

### 1. 모든 Pattern 지원 (A-H)

Railway는 MoAI-ADK의 모든 아키텍처 패턴을 단일 플랫폼에서 구현할 수 있습니다:

| Pattern | Railway 구현 | 핵심 기능 |
|---------|-------------|----------|
| A: Multi-tenant SaaS | ✅ | 환경 변수, Custom Domains |
| B: Serverless API | ✅ | Auto-scaling, Pay-per-use |
| C: Full-stack | ✅ | Frontend + Backend + DB 통합 |
| D: Microservices | ✅ | Service Mesh, Private Networking |
| E: Event-Driven | ✅ | Cron Jobs, Webhooks |
| F: Real-time Backend | ✅ | WebSocket 지원, Redis Pub/Sub |
| G: Edge Computing | ✅ | Global CDN, Edge Functions (개발 중) |
| H: Enterprise Security | ✅ | VPC, SOC2 준수, SSO |

### 2. 개발자 경험 (DX) 최우선

```bash
# 1. Railway CLI 설치
npm i -g @railway/cli

# 2. 로그인
railway login

# 3. 프로젝트 초기화
railway init

# 4. 배포 (단 하나의 명령)
railway up

# 완료! 자동으로 빌드, 배포, 도메인 할당
```

## 주요 기능

### 1. Nixpacks (자동 빌드)

**설정 없이 모든 언어 지원**:

```json
// package.json만 있으면 자동 인식
{
  "name": "my-app",
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start"
  },
  "dependencies": {
    "next": "14.0.0",
    "react": "18.2.0"
  }
}
```

Railway가 자동으로:
1. Node.js 18+ 감지
2. npm install 실행
3. npm run build 실행
4. npm run start로 서버 시작
5. 포트 자동 감지 ($PORT 환경 변수)

### 2. PostgreSQL, MySQL, MongoDB 통합

```typescript
// lib/db.ts
import { Pool } from 'pg'

// Railway가 자동으로 DATABASE_URL 주입
const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
  ssl: process.env.NODE_ENV === 'production' ? { rejectUnauthorized: false } : false,
})

export async function query<T>(text: string, params?: any[]): Promise<T[]> {
  const result = await pool.query(text, params)
  return result.rows
}

// 사용 예제
const users = await query<User>('SELECT * FROM users WHERE email = $1', [email])
```

**데이터베이스 추가**:
```bash
# PostgreSQL 추가
railway add --plugin postgresql

# MySQL 추가
railway add --plugin mysql

# MongoDB 추가
railway add --plugin mongodb

# Redis 추가
railway add --plugin redis
```

### 3. 환경 변수 관리

```bash
# 환경 변수 설정
railway variables set API_KEY=your-api-key
railway variables set DATABASE_URL=postgresql://...

# 로컬 개발 (.env 자동 생성)
railway run npm run dev

# 환경별 변수 (production, staging, development)
railway variables set --environment production STRIPE_SECRET_KEY=sk_live_...
railway variables set --environment staging STRIPE_SECRET_KEY=sk_test_...
```

**Shared Variables**:
```bash
# 모든 환경에서 공유되는 변수
railway variables set --shared API_ENDPOINT=https://api.example.com
```

### 4. Custom Domains & SSL

```bash
# 커스텀 도메인 추가
railway domain add api.example.com

# Railway가 자동으로:
# 1. SSL 인증서 발급 (Let's Encrypt)
# 2. HTTPS 리다이렉트 설정
# 3. CDN 활성화
```

**설정 예제**:
```
# DNS 레코드 추가 (Cloudflare, Namecheap 등)
CNAME api.example.com -> your-app.up.railway.app
```

### 5. Private Networking

```typescript
// service-a/index.ts - API Gateway
const serviceB = process.env.SERVICE_B_PRIVATE_URL // http://service-b.railway.internal
const serviceC = process.env.SERVICE_C_PRIVATE_URL // http://service-c.railway.internal

app.get('/api/data', async (req, res) => {
  // Private Network를 통한 내부 통신 (무료)
  const [dataB, dataC] = await Promise.all([
    fetch(`${serviceB}/data`).then(r => r.json()),
    fetch(`${serviceC}/data`).then(r => r.json()),
  ])

  res.json({ dataB, dataC })
})

// service-b/index.ts - Internal Service
app.get('/data', async (req, res) => {
  // 외부에서 접근 불가 (Private Network만)
  const data = await fetchDataFromDB()
  res.json(data)
})
```

### 6. Cron Jobs

```typescript
// cron/send-daily-emails.ts
import { sendEmail } from '../lib/email'
import { prisma } from '../lib/prisma'

async function main() {
  const users = await prisma.user.findMany({
    where: { subscribed: true },
  })

  for (const user of users) {
    await sendEmail({
      to: user.email,
      subject: 'Daily Digest',
      body: 'Your daily summary...',
    })
  }

  console.log(`Sent ${users.length} emails`)
}

main()
```

**Railway 설정**:
```bash
# Cron 서비스 추가
railway add

# Cron 표현식 설정 (매일 오전 9시)
railway variables set RAILWAY_CRON_SCHEDULE="0 9 * * *"
```

### 7. Blue-Green Deployments

```yaml
# railway.yml
services:
  web-blue:
    build:
      context: .
    healthcheck:
      path: /health
      interval: 10s
    deploy:
      replicas: 2

  web-green:
    build:
      context: .
    healthcheck:
      path: /health
      interval: 10s
    deploy:
      replicas: 0  # 대기 상태
```

**배포 프로세스**:
1. Green 환경에 새 버전 배포
2. Health check 통과 확인
3. 트래픽을 Blue → Green으로 전환
4. Blue 환경 종료

### 8. Canary Releases

```typescript
// middleware.ts - Canary 트래픽 제어
export function canaryMiddleware(req: Request, res: Response, next: NextFunction) {
  const canaryPercent = parseInt(process.env.CANARY_PERCENT || '0')
  const random = Math.random() * 100

  if (random < canaryPercent) {
    // Canary 버전으로 프록시
    proxy.web(req, res, {
      target: process.env.CANARY_URL,
    })
  } else {
    // 기존 버전
    next()
  }
}

app.use(canaryMiddleware)
```

**Railway 설정**:
```bash
# Canary 서비스 배포
railway up --service canary

# 트래픽 10%로 시작
railway variables set --service canary CANARY_PERCENT=10

# 점진적으로 증가
railway variables set --service canary CANARY_PERCENT=50
railway variables set --service canary CANARY_PERCENT=100
```

## 시작하기

### 1. 프로젝트 생성 및 배포

```bash
# 1. Railway CLI 설치
npm i -g @railway/cli

# 2. 로그인
railway login

# 3. 새 프로젝트 생성
mkdir my-railway-app
cd my-railway-app
railway init

# 4. Next.js 앱 생성
npx create-next-app@latest . --typescript --tailwind --app

# 5. 데이터베이스 추가
railway add --plugin postgresql

# 6. 환경 변수 연결
railway run npm run dev  # 로컬 개발 (DATABASE_URL 자동 주입)

# 7. 배포
railway up
```

### 2. 프로젝트 구조

```
my-railway-app/
├── app/                  # Next.js App Router
│   ├── api/             # API Routes
│   └── page.tsx         # Frontend
├── lib/
│   ├── db.ts            # Database connection
│   └── utils.ts
├── prisma/
│   └── schema.prisma    # Database schema
├── railway.json         # Railway 설정 (선택)
└── package.json
```

### 3. Railway 설정 (`railway.json`)

```json
{
  "$schema": "https://railway.app/railway.schema.json",
  "build": {
    "builder": "NIXPACKS",
    "buildCommand": "npm run build"
  },
  "deploy": {
    "startCommand": "npm run start",
    "healthcheckPath": "/api/health",
    "healthcheckTimeout": 30,
    "restartPolicyType": "ON_FAILURE",
    "restartPolicyMaxRetries": 3
  }
}
```

### 4. 로컬 개발 환경

```bash
# .env 자동 생성 (Railway에서 변수 가져옴)
railway run npm run dev

# 또는 직접 링크
railway link

# 변수 확인
railway variables
```

## 사용 가이드

### Pattern A: Multi-tenant SaaS

```typescript
// middleware.ts - 테넌트 격리
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export async function middleware(request: NextRequest) {
  const hostname = request.headers.get('host') || ''
  const subdomain = hostname.split('.')[0]

  // 테넌트별 데이터베이스 URL
  const tenantDBUrl = process.env[`DB_${subdomain.toUpperCase()}`]

  if (!tenantDBUrl) {
    return NextResponse.json(
      { error: 'Invalid tenant' },
      { status: 404 }
    )
  }

  // 요청 헤더에 테넌트 정보 추가
  const requestHeaders = new Headers(request.headers)
  requestHeaders.set('x-tenant-id', subdomain)
  requestHeaders.set('x-tenant-db', tenantDBUrl)

  return NextResponse.next({
    request: {
      headers: requestHeaders,
    },
  })
}
```

**Railway 설정**:
```bash
# 테넌트별 데이터베이스
railway variables set DB_ACME=postgresql://...
railway variables set DB_GLOBEX=postgresql://...
railway variables set DB_INITECH=postgresql://...

# Custom domains
railway domain add acme.myapp.com
railway domain add globex.myapp.com
railway domain add initech.myapp.com
```

### Pattern D: Microservices

```yaml
# railway.yml - 마이크로서비스 아키텍처
services:
  api-gateway:
    build:
      context: ./services/api-gateway
    domains:
      - api.example.com
    env:
      - AUTH_SERVICE_URL=${{services.auth.url}}
      - USER_SERVICE_URL=${{services.users.url}}
      - ORDER_SERVICE_URL=${{services.orders.url}}

  auth:
    build:
      context: ./services/auth
    env:
      - DATABASE_URL=${{plugins.auth-db.DATABASE_URL}}

  users:
    build:
      context: ./services/users
    env:
      - DATABASE_URL=${{plugins.users-db.DATABASE_URL}}
      - REDIS_URL=${{plugins.redis.REDIS_URL}}

  orders:
    build:
      context: ./services/orders
    env:
      - DATABASE_URL=${{plugins.orders-db.DATABASE_URL}}

plugins:
  auth-db:
    type: postgresql
  users-db:
    type: postgresql
  orders-db:
    type: postgresql
  redis:
    type: redis
```

### Pattern E: Event-Driven Architecture

```typescript
// worker/process-events.ts
import { Redis } from 'ioredis'
import { prisma } from '../lib/prisma'

const redis = new Redis(process.env.REDIS_URL!)

async function processEvents() {
  console.log('Event worker started')

  while (true) {
    // Redis List에서 이벤트 가져오기 (블로킹)
    const result = await redis.brpop('events', 0)

    if (result) {
      const [, eventData] = result
      const event = JSON.parse(eventData)

      try {
        await handleEvent(event)
      } catch (error) {
        console.error('Event processing failed:', error)
        // Dead letter queue로 이동
        await redis.lpush('events:failed', eventData)
      }
    }
  }
}

async function handleEvent(event: any) {
  switch (event.type) {
    case 'user.created':
      await sendWelcomeEmail(event.data)
      break
    case 'order.placed':
      await processOrder(event.data)
      break
    default:
      console.log('Unknown event type:', event.type)
  }
}

processEvents()
```

**Railway 설정**:
```bash
# API 서비스 (이벤트 발행자)
railway add --service api

# Worker 서비스 (이벤트 소비자)
railway add --service worker

# Redis 추가
railway add --plugin redis

# 환경 변수 공유
railway variables set --shared REDIS_URL=${{plugins.redis.REDIS_URL}}
```

## 코드 예제

### 1. Health Check 엔드포인트

```typescript
// app/api/health/route.ts
import { prisma } from '@/lib/prisma'
import { redis } from '@/lib/redis'

export async function GET() {
  try {
    // Database health
    await prisma.$queryRaw`SELECT 1`

    // Redis health
    await redis.ping()

    return Response.json({
      status: 'healthy',
      timestamp: new Date().toISOString(),
      services: {
        database: 'up',
        redis: 'up',
      },
    })
  } catch (error) {
    return Response.json(
      {
        status: 'unhealthy',
        error: error instanceof Error ? error.message : 'Unknown error',
      },
      { status: 503 }
    )
  }
}
```

### 2. Observability (로깅 및 메트릭)

```typescript
// lib/logger.ts
import pino from 'pino'

export const logger = pino({
  level: process.env.LOG_LEVEL || 'info',
  formatters: {
    level: (label) => {
      return { level: label }
    },
  },
  // Railway가 자동으로 JSON 로그 수집
  serializers: {
    req: (req) => ({
      method: req.method,
      url: req.url,
      headers: req.headers,
    }),
    res: (res) => ({
      statusCode: res.statusCode,
    }),
    err: pino.stdSerializers.err,
  },
})

// middleware.ts
import { logger } from './lib/logger'

app.use((req, res, next) => {
  const start = Date.now()

  res.on('finish', () => {
    const duration = Date.now() - start

    logger.info({
      req,
      res,
      duration,
    })
  })

  next()
})
```

### 3. 데이터베이스 마이그레이션

```typescript
// scripts/migrate.ts
import { exec } from 'child_process'
import { promisify } from 'util'

const execAsync = promisify(exec)

async function migrate() {
  try {
    console.log('Running database migrations...')

    // Prisma migrate
    await execAsync('npx prisma migrate deploy')

    console.log('Migrations completed successfully')
    process.exit(0)
  } catch (error) {
    console.error('Migration failed:', error)
    process.exit(1)
  }
}

migrate()
```

**Railway 설정**:
```json
// package.json
{
  "scripts": {
    "build": "npm run migrate && next build",
    "migrate": "prisma migrate deploy"
  }
}
```

### 4. Background Jobs

```typescript
// worker/jobs/cleanup.ts
import { prisma } from '../../lib/prisma'

export async function cleanupExpiredSessions() {
  const result = await prisma.session.deleteMany({
    where: {
      expiresAt: {
        lt: new Date(),
      },
    },
  })

  console.log(`Deleted ${result.count} expired sessions`)
}

// worker/index.ts
import { CronJob } from 'cron'
import { cleanupExpiredSessions } from './jobs/cleanup'

// 매시간 실행
new CronJob('0 * * * *', async () => {
  await cleanupExpiredSessions()
}, null, true)

console.log('Worker started')
```

### 5. Rate Limiting (Redis)

```typescript
// lib/rate-limit.ts
import { Redis } from 'ioredis'

const redis = new Redis(process.env.REDIS_URL!)

export async function rateLimit(
  key: string,
  limit: number,
  window: number
): Promise<boolean> {
  const current = await redis.incr(key)

  if (current === 1) {
    await redis.expire(key, window)
  }

  return current <= limit
}

// middleware.ts
import { rateLimit } from './lib/rate-limit'

app.use(async (req, res, next) => {
  const ip = req.ip
  const key = `rate-limit:${ip}`

  const allowed = await rateLimit(key, 100, 60) // 100 requests per minute

  if (!allowed) {
    return res.status(429).json({ error: 'Too many requests' })
  }

  next()
})
```

## Monitoring & Observability

### 1. 실시간 로그

```bash
# 전체 로그 스트리밍
railway logs

# 특정 서비스
railway logs --service api

# 마지막 100줄
railway logs --tail 100

# 에러만 필터링
railway logs | grep ERROR
```

### 2. 메트릭 대시보드

Railway Dashboard에서 자동으로 수집:
- **CPU 사용률**: 실시간 CPU 사용량
- **메모리 사용률**: RAM 사용량
- **네트워크 I/O**: 인바운드/아웃바운드 트래픽
- **디스크 I/O**: 읽기/쓰기 작업
- **응답 시간**: P50, P95, P99 latency

### 3. 알림 설정

```bash
# Webhook 알림 (Slack, Discord 등)
railway webhooks create \
  --url https://hooks.slack.com/services/... \
  --events deployment.success,deployment.failure
```

### 4. Custom Metrics (Prometheus)

```typescript
// lib/metrics.ts
import promClient from 'prom-client'

const register = new promClient.Registry()

export const httpRequestDuration = new promClient.Histogram({
  name: 'http_request_duration_seconds',
  help: 'Duration of HTTP requests in seconds',
  labelNames: ['method', 'route', 'status'],
  registers: [register],
})

export const activeConnections = new promClient.Gauge({
  name: 'active_connections',
  help: 'Number of active connections',
  registers: [register],
})

// app/api/metrics/route.ts
import { register } from '@/lib/metrics'

export async function GET() {
  return new Response(await register.metrics(), {
    headers: {
      'Content-Type': register.contentType,
    },
  })
}
```

## Cost Optimization

### 1. 리소스 최적화

```json
// railway.json - 리소스 제한
{
  "deploy": {
    "numReplicas": 2,
    "resources": {
      "cpu": 1,      // vCPU 개수
      "memory": 512  // MB
    }
  }
}
```

### 2. Autoscaling 설정

```bash
# Horizontal autoscaling
railway variables set RAILWAY_AUTOSCALE_MIN=1
railway variables set RAILWAY_AUTOSCALE_MAX=10
railway variables set RAILWAY_AUTOSCALE_CPU_THRESHOLD=70
```

### 3. 사용량 모니터링

```bash
# 현재 사용량 확인
railway status

# 비용 예측
railway billing
```

## Best Practices

### 1. Secrets 관리

```bash
# ✅ Railway Variables 사용
railway variables set API_KEY=secret-key

# ❌ .env 파일 커밋 금지
echo ".env" >> .gitignore
```

### 2. Health Checks

```typescript
// 항상 health check 엔드포인트 제공
export async function GET() {
  return Response.json({ status: 'ok' })
}
```

### 3. Graceful Shutdown

```typescript
// server.ts
const server = app.listen(port)

process.on('SIGTERM', () => {
  console.log('SIGTERM received, shutting down gracefully')

  server.close(() => {
    console.log('Server closed')
    process.exit(0)
  })
})
```

### 4. 데이터베이스 연결 관리

```typescript
// Connection pooling
const pool = new Pool({
  max: 10,
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 2000,
})
```

## 문제 해결

### 1. 배포 실패

```bash
# 로그 확인
railway logs --deployment last

# 이전 버전으로 롤백
railway rollback
```

### 2. 메모리 부족

```json
// railway.json - 메모리 증가
{
  "deploy": {
    "resources": {
      "memory": 1024  // 512 → 1024 MB
    }
  }
}
```

### 3. 데이터베이스 연결 실패

```typescript
// 연결 재시도 로직
async function connectWithRetry(maxRetries = 5) {
  for (let i = 0; i < maxRetries; i++) {
    try {
      await prisma.$connect()
      console.log('Database connected')
      return
    } catch (error) {
      console.log(`Connection attempt ${i + 1} failed`)
      await new Promise(resolve => setTimeout(resolve, 1000 * (i + 1)))
    }
  }
  throw new Error('Failed to connect to database')
}
```

## 다음 단계

- [Vercel 가이드](/ko/skills/baas/vercel) - Frontend에 특화된 플랫폼
- [Neon 가이드](/ko/skills/baas/neon) - Serverless PostgreSQL 전문
- [BaaS 개요](/ko/skills/baas) - 플랫폼 비교 및 선택 가이드
- [Pattern D: Microservices](/ko/skills/patterns/pattern-d) - 마이크로서비스 아키텍처 상세

## 참고 자료

- [Railway 공식 문서](https://docs.railway.app/)
- [Railway CLI 가이드](https://docs.railway.app/develop/cli)
- [Nixpacks 문서](https://nixpacks.com/)
- [Railway 커뮤니티](https://discord.gg/railway)
