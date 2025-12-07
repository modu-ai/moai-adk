# Deployment Platform Implementation Examples

## Example 1: Next.js 15 Full-Stack Application on Vercel

### Project Structure

```
my-nextjs-app/
├── app/
│   ├── api/
│   │   ├── auth/
│   │   │   └── route.ts
│   │   ├── products/
│   │   │   └── route.ts
│   │   └── cron/
│   │       └── sync/route.ts
│   ├── (marketing)/
│   │   ├── page.tsx
│   │   └── about/page.tsx
│   ├── dashboard/
│   │   ├── layout.tsx
│   │   └── page.tsx
│   └── layout.tsx
├── middleware.ts
├── vercel.json
├── next.config.js
└── package.json
```

### Configuration Files

next.config.js:
```javascript
/** @type {import('next').NextConfig} */
const nextConfig = {
  experimental: {
    ppr: true, // Partial Prerendering
  },
  images: {
    remotePatterns: [
      {
        protocol: 'https',
        hostname: 'images.example.com',
      },
    ],
    formats: ['image/avif', 'image/webp'],
  },
  async headers() {
    return [
      {
        source: '/api/:path*',
        headers: [
          { key: 'Access-Control-Allow-Origin', value: '*' },
          { key: 'Access-Control-Allow-Methods', value: 'GET, POST, PUT, DELETE, OPTIONS' },
        ],
      },
    ]
  },
}

module.exports = nextConfig
```

vercel.json:
```json
{
  "$schema": "https://openapi.vercel.sh/vercel.json",
  "framework": "nextjs",
  "buildCommand": "pnpm build",
  "installCommand": "pnpm install --frozen-lockfile",
  "regions": ["iad1", "sfo1", "fra1", "sin1"],
  "functions": {
    "app/api/**/*.ts": {
      "memory": 1024,
      "maxDuration": 30
    },
    "app/api/cron/**/*.ts": {
      "memory": 512,
      "maxDuration": 60
    }
  },
  "crons": [
    {
      "path": "/api/cron/sync",
      "schedule": "*/15 * * * *"
    }
  ],
  "headers": [
    {
      "source": "/(.*)",
      "headers": [
        {
          "key": "X-Content-Type-Options",
          "value": "nosniff"
        },
        {
          "key": "X-Frame-Options",
          "value": "DENY"
        },
        {
          "key": "Strict-Transport-Security",
          "value": "max-age=31536000; includeSubDomains"
        }
      ]
    }
  ]
}
```

### Edge Middleware

middleware.ts:
```typescript
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'
import { SignJWT, jwtVerify } from 'jose'

const JWT_SECRET = new TextEncoder().encode(process.env.JWT_SECRET!)

interface UserPayload {
  userId: string
  email: string
  role: string
}

async function verifyToken(token: string): Promise<UserPayload | null> {
  try {
    const { payload } = await jwtVerify(token, JWT_SECRET)
    return payload as unknown as UserPayload
  } catch {
    return null
  }
}

export async function middleware(request: NextRequest) {
  const { pathname } = request.nextUrl

  // Public routes
  if (pathname.startsWith('/api/auth') || pathname === '/login') {
    return NextResponse.next()
  }

  // Protected routes
  if (pathname.startsWith('/dashboard') || pathname.startsWith('/api/')) {
    const token = request.cookies.get('auth-token')?.value

    if (!token) {
      if (pathname.startsWith('/api/')) {
        return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
      }
      return NextResponse.redirect(new URL('/login', request.url))
    }

    const user = await verifyToken(token)
    if (!user) {
      const response = NextResponse.redirect(new URL('/login', request.url))
      response.cookies.delete('auth-token')
      return response
    }

    // Add user info to headers for downstream use
    const requestHeaders = new Headers(request.headers)
    requestHeaders.set('x-user-id', user.userId)
    requestHeaders.set('x-user-role', user.role)

    return NextResponse.next({
      request: { headers: requestHeaders }
    })
  }

  // Geo-based routing
  const country = request.geo?.country || 'US'
  if (country === 'DE' && !pathname.startsWith('/de')) {
    return NextResponse.redirect(new URL(`/de${pathname}`, request.url))
  }

  return NextResponse.next()
}

export const config = {
  matcher: [
    '/((?!_next/static|_next/image|favicon.ico|public).*)',
  ],
}
```

### API Routes with Edge Runtime

app/api/products/route.ts:
```typescript
import { NextRequest, NextResponse } from 'next/server'

export const runtime = 'edge'

interface Product {
  id: string
  name: string
  price: number
  category: string
}

// Simulated database (replace with actual database call)
async function getProducts(category?: string): Promise<Product[]> {
  const products: Product[] = [
    { id: '1', name: 'Product A', price: 99.99, category: 'electronics' },
    { id: '2', name: 'Product B', price: 149.99, category: 'electronics' },
    { id: '3', name: 'Product C', price: 29.99, category: 'accessories' },
  ]

  if (category) {
    return products.filter(p => p.category === category)
  }
  return products
}

export async function GET(request: NextRequest) {
  const searchParams = request.nextUrl.searchParams
  const category = searchParams.get('category') || undefined

  try {
    const products = await getProducts(category)

    return NextResponse.json(products, {
      headers: {
        'Cache-Control': 'public, s-maxage=60, stale-while-revalidate=300',
      },
    })
  } catch (error) {
    return NextResponse.json(
      { error: 'Failed to fetch products' },
      { status: 500 }
    )
  }
}

export async function POST(request: NextRequest) {
  const userId = request.headers.get('x-user-id')
  const userRole = request.headers.get('x-user-role')

  if (userRole !== 'admin') {
    return NextResponse.json(
      { error: 'Insufficient permissions' },
      { status: 403 }
    )
  }

  try {
    const body = await request.json()
    const product: Product = {
      id: crypto.randomUUID(),
      name: body.name,
      price: body.price,
      category: body.category,
    }

    // Save to database
    // await db.products.create(product)

    return NextResponse.json(product, { status: 201 })
  } catch (error) {
    return NextResponse.json(
      { error: 'Failed to create product' },
      { status: 500 }
    )
  }
}
```

### Cron Job Handler

app/api/cron/sync/route.ts:
```typescript
import { NextRequest, NextResponse } from 'next/server'

export const runtime = 'nodejs' // Cron jobs need Node.js runtime for longer duration

export async function GET(request: NextRequest) {
  // Verify cron authorization
  const authHeader = request.headers.get('authorization')
  if (authHeader !== `Bearer ${process.env.CRON_SECRET}`) {
    return NextResponse.json({ error: 'Unauthorized' }, { status: 401 })
  }

  console.log('Starting sync job...')

  try {
    // Sync logic
    const startTime = Date.now()

    // 1. Fetch external data
    const externalData = await fetchExternalData()

    // 2. Process and transform
    const processedData = await processData(externalData)

    // 3. Update database
    const updateResult = await updateDatabase(processedData)

    // 4. Cleanup old records
    const cleanupResult = await cleanupOldRecords()

    const duration = Date.now() - startTime

    return NextResponse.json({
      success: true,
      timestamp: new Date().toISOString(),
      duration: `${duration}ms`,
      stats: {
        fetched: externalData.length,
        processed: processedData.length,
        updated: updateResult.count,
        cleaned: cleanupResult.count,
      },
    })
  } catch (error) {
    console.error('Sync job failed:', error)
    return NextResponse.json(
      {
        success: false,
        error: error instanceof Error ? error.message : 'Unknown error',
        timestamp: new Date().toISOString(),
      },
      { status: 500 }
    )
  }
}

async function fetchExternalData() {
  // Implementation
  return []
}

async function processData(data: any[]) {
  // Implementation
  return data
}

async function updateDatabase(data: any[]) {
  // Implementation
  return { count: data.length }
}

async function cleanupOldRecords() {
  // Implementation
  return { count: 0 }
}
```

---

## Example 2: Multi-Service Backend on Railway

### Project Structure (Monorepo)

```
my-platform/
├── apps/
│   ├── api/
│   │   ├── src/
│   │   │   ├── main.ts
│   │   │   ├── routes/
│   │   │   └── middleware/
│   │   ├── Dockerfile
│   │   ├── railway.toml
│   │   └── package.json
│   ├── worker/
│   │   ├── src/
│   │   │   ├── main.ts
│   │   │   └── jobs/
│   │   ├── Dockerfile
│   │   ├── railway.toml
│   │   └── package.json
│   └── web/
│       ├── src/
│       ├── Dockerfile
│       └── package.json
├── packages/
│   ├── shared/
│   │   └── src/
│   └── database/
│       └── prisma/
├── docker-compose.yml
├── pnpm-workspace.yaml
└── turbo.json
```

### API Service

apps/api/Dockerfile:
```dockerfile
FROM node:20-alpine AS base
RUN corepack enable && corepack prepare pnpm@latest --activate

FROM base AS deps
WORKDIR /app
COPY pnpm-lock.yaml ./
COPY package.json ./
COPY apps/api/package.json ./apps/api/
COPY packages/shared/package.json ./packages/shared/
COPY packages/database/package.json ./packages/database/
RUN pnpm install --frozen-lockfile --filter api...

FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY --from=deps /app/apps/api/node_modules ./apps/api/node_modules
COPY . .
RUN pnpm turbo build --filter=api

FROM base AS runner
WORKDIR /app
ENV NODE_ENV=production

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 api

COPY --from=builder --chown=api:nodejs /app/apps/api/dist ./dist
COPY --from=builder --chown=api:nodejs /app/apps/api/node_modules ./node_modules
COPY --from=builder --chown=api:nodejs /app/packages/database/prisma ./prisma

USER api

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:3000/health || exit 1

EXPOSE 3000
CMD ["node", "dist/main.js"]
```

apps/api/railway.toml:
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "./apps/api/Dockerfile"
watchPatterns = ["apps/api/**", "packages/**"]

[deploy]
startCommand = "node dist/main.js"
healthcheckPath = "/health"
healthcheckTimeout = 100
restartPolicyType = "ON_FAILURE"
restartPolicyMaxRetries = 5
numReplicas = 2

[deploy.resources]
memory = "1Gi"
cpu = "1"

[[deploy.scaling]]
minReplicas = 2
maxReplicas = 10
metric = "cpu"
target = 75

[[deploy.scaling]]
minReplicas = 2
maxReplicas = 8
metric = "requests_per_second"
target = 500
```

apps/api/src/main.ts:
```typescript
import express from 'express'
import cors from 'cors'
import helmet from 'helmet'
import { PrismaClient } from '@prisma/client'
import Redis from 'ioredis'

const app = express()
const prisma = new PrismaClient()
const redis = new Redis(process.env.REDIS_URL!)

// Middleware
app.use(helmet())
app.use(cors({
  origin: process.env.ALLOWED_ORIGINS?.split(',') || ['http://localhost:3000'],
  credentials: true,
}))
app.use(express.json())

// Health check
app.get('/health', async (req, res) => {
  try {
    await prisma.$queryRaw`SELECT 1`
    await redis.ping()

    res.json({
      status: 'healthy',
      service: 'api',
      version: process.env.RAILWAY_GIT_COMMIT_SHA || 'local',
      timestamp: new Date().toISOString(),
      checks: {
        database: 'ok',
        redis: 'ok',
      },
    })
  } catch (error) {
    res.status(503).json({
      status: 'unhealthy',
      error: error instanceof Error ? error.message : 'Unknown error',
    })
  }
})

// API Routes
app.get('/api/users', async (req, res) => {
  try {
    // Check cache first
    const cacheKey = 'users:all'
    const cached = await redis.get(cacheKey)

    if (cached) {
      return res.json(JSON.parse(cached))
    }

    const users = await prisma.user.findMany({
      select: {
        id: true,
        email: true,
        name: true,
        createdAt: true,
      },
      take: 100,
    })

    // Cache for 5 minutes
    await redis.setex(cacheKey, 300, JSON.stringify(users))

    res.json(users)
  } catch (error) {
    console.error('Failed to fetch users:', error)
    res.status(500).json({ error: 'Internal server error' })
  }
})

app.post('/api/users', async (req, res) => {
  try {
    const { email, name } = req.body

    const user = await prisma.user.create({
      data: { email, name },
    })

    // Invalidate cache
    await redis.del('users:all')

    // Publish event for worker
    await redis.publish('user:created', JSON.stringify(user))

    res.status(201).json(user)
  } catch (error) {
    console.error('Failed to create user:', error)
    res.status(500).json({ error: 'Internal server error' })
  }
})

// Graceful shutdown
process.on('SIGTERM', async () => {
  console.log('SIGTERM received, shutting down gracefully...')
  await prisma.$disconnect()
  await redis.quit()
  process.exit(0)
})

const PORT = process.env.PORT || 3000
app.listen(PORT, () => {
  console.log(`API server running on port ${PORT}`)
  console.log(`Environment: ${process.env.NODE_ENV}`)
  console.log(`Replica: ${process.env.RAILWAY_REPLICA_ID || 'local'}`)
})
```

### Worker Service

apps/worker/railway.toml:
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "./apps/worker/Dockerfile"

[deploy]
startCommand = "node dist/main.js"
restartPolicyType = "ON_FAILURE"
restartPolicyMaxRetries = 10
numReplicas = 1

[deploy.resources]
memory = "512Mi"
cpu = "0.5"
```

apps/worker/src/main.ts:
```typescript
import Redis from 'ioredis'
import { PrismaClient } from '@prisma/client'

const prisma = new PrismaClient()
const redis = new Redis(process.env.REDIS_URL!)
const subscriber = new Redis(process.env.REDIS_URL!)

interface Job {
  type: string
  payload: any
  timestamp: number
}

class JobProcessor {
  private handlers: Map<string, (payload: any) => Promise<void>>

  constructor() {
    this.handlers = new Map()
    this.registerHandlers()
  }

  private registerHandlers() {
    this.handlers.set('user:created', this.handleUserCreated.bind(this))
    this.handlers.set('email:send', this.handleEmailSend.bind(this))
    this.handlers.set('report:generate', this.handleReportGenerate.bind(this))
  }

  async handleUserCreated(payload: any) {
    console.log('Processing user:created event:', payload)

    // Send welcome email
    await this.queueJob('email:send', {
      to: payload.email,
      template: 'welcome',
      data: { name: payload.name },
    })

    // Update analytics
    await prisma.analytics.create({
      data: {
        event: 'user_signup',
        userId: payload.id,
        metadata: payload,
      },
    })
  }

  async handleEmailSend(payload: any) {
    console.log('Sending email:', payload)
    // Email sending logic
  }

  async handleReportGenerate(payload: any) {
    console.log('Generating report:', payload)
    // Report generation logic
  }

  async queueJob(type: string, payload: any) {
    const job: Job = {
      type,
      payload,
      timestamp: Date.now(),
    }
    await redis.lpush('job:queue', JSON.stringify(job))
  }

  async processQueue() {
    while (true) {
      try {
        // Block and wait for job
        const result = await redis.brpop('job:queue', 30)

        if (result) {
          const [, jobData] = result
          const job: Job = JSON.parse(jobData)

          const handler = this.handlers.get(job.type)
          if (handler) {
            await handler(job.payload)
            console.log(`Processed job: ${job.type}`)
          } else {
            console.warn(`No handler for job type: ${job.type}`)
          }
        }
      } catch (error) {
        console.error('Error processing job:', error)
        await new Promise(resolve => setTimeout(resolve, 1000))
      }
    }
  }

  async subscribeToPubSub() {
    subscriber.subscribe('user:created', 'order:placed')

    subscriber.on('message', async (channel, message) => {
      try {
        const payload = JSON.parse(message)
        const handler = this.handlers.get(channel)

        if (handler) {
          await handler(payload)
        }
      } catch (error) {
        console.error(`Error handling ${channel}:`, error)
      }
    })
  }
}

async function main() {
  console.log('Worker starting...')
  console.log(`Replica: ${process.env.RAILWAY_REPLICA_ID || 'local'}`)

  const processor = new JobProcessor()

  // Start both queue processing and pub/sub
  await Promise.all([
    processor.processQueue(),
    processor.subscribeToPubSub(),
  ])
}

// Graceful shutdown
process.on('SIGTERM', async () => {
  console.log('SIGTERM received, shutting down worker...')
  await prisma.$disconnect()
  await redis.quit()
  await subscriber.quit()
  process.exit(0)
})

main().catch(console.error)
```

---

## Example 3: Blue-Green Deployment Automation

### Vercel Blue-Green Script

scripts/vercel-blue-green.ts:
```typescript
import { Vercel } from '@vercel/sdk'

interface DeploymentConfig {
  projectId: string
  teamId: string
  productionAlias: string
}

class BlueGreenDeployer {
  private vercel: Vercel
  private config: DeploymentConfig

  constructor(config: DeploymentConfig) {
    this.vercel = new Vercel({ token: process.env.VERCEL_TOKEN! })
    this.config = config
  }

  async deploy() {
    console.log('Starting blue-green deployment...')

    // 1. Get current production deployment
    const currentProduction = await this.getCurrentProduction()
    console.log(`Current production: ${currentProduction?.id}`)

    // 2. Create new deployment (green)
    const newDeployment = await this.createDeployment()
    console.log(`New deployment created: ${newDeployment.id}`)
    console.log(`Preview URL: ${newDeployment.url}`)

    // 3. Wait for deployment to be ready
    await this.waitForReady(newDeployment.id)
    console.log('Deployment is ready')

    // 4. Run smoke tests
    const testsPassed = await this.runSmokeTests(newDeployment.url)
    if (!testsPassed) {
      console.error('Smoke tests failed, aborting deployment')
      await this.deleteDeployment(newDeployment.id)
      throw new Error('Smoke tests failed')
    }
    console.log('Smoke tests passed')

    // 5. Switch production alias
    await this.switchAlias(newDeployment.id)
    console.log(`Switched ${this.config.productionAlias} to new deployment`)

    // 6. Keep old deployment for quick rollback (delete after 24h)
    if (currentProduction) {
      console.log(`Previous deployment ${currentProduction.id} kept for rollback`)
    }

    return newDeployment
  }

  private async getCurrentProduction() {
    const aliases = await this.vercel.aliases.list({
      projectId: this.config.projectId,
      teamId: this.config.teamId,
    })

    const productionAlias = aliases.aliases.find(
      a => a.alias === this.config.productionAlias
    )

    return productionAlias?.deployment
  }

  private async createDeployment() {
    return this.vercel.deployments.create({
      name: this.config.projectId,
      teamId: this.config.teamId,
      gitSource: {
        type: 'github',
        ref: process.env.GITHUB_REF || 'main',
        repo: process.env.GITHUB_REPOSITORY!,
      },
    })
  }

  private async waitForReady(deploymentId: string, timeout = 300000) {
    const startTime = Date.now()

    while (Date.now() - startTime < timeout) {
      const deployment = await this.vercel.deployments.get(deploymentId)

      if (deployment.readyState === 'READY') {
        return
      }

      if (deployment.readyState === 'ERROR') {
        throw new Error(`Deployment failed: ${deployment.errorMessage}`)
      }

      await new Promise(resolve => setTimeout(resolve, 5000))
    }

    throw new Error('Deployment timeout')
  }

  private async runSmokeTests(url: string): Promise<boolean> {
    const endpoints = [
      '/',
      '/api/health',
      '/api/products',
    ]

    for (const endpoint of endpoints) {
      try {
        const response = await fetch(`https://${url}${endpoint}`)
        if (!response.ok) {
          console.error(`Smoke test failed for ${endpoint}: ${response.status}`)
          return false
        }
      } catch (error) {
        console.error(`Smoke test error for ${endpoint}:`, error)
        return false
      }
    }

    return true
  }

  private async switchAlias(deploymentId: string) {
    await this.vercel.aliases.create({
      alias: this.config.productionAlias,
      deploymentId,
      teamId: this.config.teamId,
    })
  }

  private async deleteDeployment(deploymentId: string) {
    await this.vercel.deployments.delete(deploymentId, {
      teamId: this.config.teamId,
    })
  }

  async rollback(deploymentId: string) {
    console.log(`Rolling back to deployment ${deploymentId}`)
    await this.switchAlias(deploymentId)
    console.log('Rollback complete')
  }
}

// Usage
const deployer = new BlueGreenDeployer({
  projectId: process.env.VERCEL_PROJECT_ID!,
  teamId: process.env.VERCEL_TEAM_ID!,
  productionAlias: 'example.com',
})

deployer.deploy()
  .then(deployment => {
    console.log('Deployment successful:', deployment.id)
  })
  .catch(error => {
    console.error('Deployment failed:', error)
    process.exit(1)
  })
```

---

## Example 4: GitHub Actions CI/CD Pipeline

### Complete Deployment Workflow

.github/workflows/deploy.yml:
```yaml
name: Production Deployment

on:
  push:
    branches: [main]
  workflow_dispatch:
    inputs:
      environment:
        description: 'Deployment environment'
        required: true
        default: 'production'
        type: choice
        options:
          - staging
          - production

env:
  TURBO_TOKEN: ${{ secrets.TURBO_TOKEN }}
  TURBO_TEAM: ${{ secrets.TURBO_TEAM }}

jobs:
  test:
    name: Test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup pnpm
        uses: pnpm/action-setup@v2
        with:
          version: 8

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'pnpm'

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Run tests
        run: pnpm turbo test

      - name: Run linting
        run: pnpm turbo lint

      - name: Type check
        run: pnpm turbo type-check

  build:
    name: Build
    needs: test
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup pnpm
        uses: pnpm/action-setup@v2
        with:
          version: 8

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: 20
          cache: 'pnpm'

      - name: Install dependencies
        run: pnpm install --frozen-lockfile

      - name: Build all packages
        run: pnpm turbo build

      - name: Upload build artifacts
        uses: actions/upload-artifact@v4
        with:
          name: build-artifacts
          path: |
            apps/web/.next
            apps/api/dist
          retention-days: 1

  deploy-web:
    name: Deploy Web (Vercel)
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: ${{ github.event.inputs.environment || 'production' }}
      url: ${{ steps.deploy.outputs.url }}
    steps:
      - uses: actions/checkout@v4

      - name: Download build artifacts
        uses: actions/download-artifact@v4
        with:
          name: build-artifacts

      - name: Install Vercel CLI
        run: npm i -g vercel@latest

      - name: Pull Vercel Environment
        run: |
          vercel pull --yes \
            --environment=${{ github.event.inputs.environment || 'production' }} \
            --token=${{ secrets.VERCEL_TOKEN }}
        working-directory: apps/web

      - name: Deploy to Vercel
        id: deploy
        run: |
          if [ "${{ github.event.inputs.environment }}" = "production" ]; then
            URL=$(vercel deploy --prebuilt --prod --token=${{ secrets.VERCEL_TOKEN }})
          else
            URL=$(vercel deploy --prebuilt --token=${{ secrets.VERCEL_TOKEN }})
          fi
          echo "url=$URL" >> $GITHUB_OUTPUT
        working-directory: apps/web

  deploy-api:
    name: Deploy API (Railway)
    needs: build
    runs-on: ubuntu-latest
    environment:
      name: ${{ github.event.inputs.environment || 'production' }}
    steps:
      - uses: actions/checkout@v4

      - name: Download build artifacts
        uses: actions/download-artifact@v4
        with:
          name: build-artifacts

      - name: Install Railway CLI
        run: npm i -g @railway/cli

      - name: Deploy to Railway
        run: |
          railway up \
            --service api \
            --environment ${{ github.event.inputs.environment || 'production' }} \
            --detach
        env:
          RAILWAY_TOKEN: ${{ secrets.RAILWAY_TOKEN }}
        working-directory: apps/api

  deploy-worker:
    name: Deploy Worker (Railway)
    needs: deploy-api
    runs-on: ubuntu-latest
    environment:
      name: ${{ github.event.inputs.environment || 'production' }}
    steps:
      - uses: actions/checkout@v4

      - name: Install Railway CLI
        run: npm i -g @railway/cli

      - name: Deploy Worker to Railway
        run: |
          railway up \
            --service worker \
            --environment ${{ github.event.inputs.environment || 'production' }} \
            --detach
        env:
          RAILWAY_TOKEN: ${{ secrets.RAILWAY_TOKEN }}
        working-directory: apps/worker

  smoke-tests:
    name: Smoke Tests
    needs: [deploy-web, deploy-api]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Run smoke tests
        run: |
          # Wait for deployments to be ready
          sleep 30

          # Test web
          WEB_URL="${{ needs.deploy-web.outputs.url }}"
          curl --fail --silent --show-error "$WEB_URL" > /dev/null

          # Test API health
          API_URL="${{ secrets.API_URL }}"
          curl --fail --silent --show-error "$API_URL/health" > /dev/null

          echo "All smoke tests passed!"

  notify:
    name: Notify
    needs: [smoke-tests]
    runs-on: ubuntu-latest
    if: always()
    steps:
      - name: Send Slack notification
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          fields: repo,message,commit,author,action,eventName,ref,workflow
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

---

## Example 5: Canary Release Implementation

### Railway Canary Deployment

scripts/railway-canary.py:
```python
import asyncio
import aiohttp
import os
from dataclasses import dataclass
from typing import Optional

@dataclass
class CanaryMetrics:
    error_rate: float
    latency_p50: float
    latency_p99: float
    request_count: int

class RailwayCanaryDeployer:
    def __init__(self):
        self.api_url = "https://backboard.railway.app/graphql/v2"
        self.token = os.environ["RAILWAY_TOKEN"]
        self.project_id = os.environ["RAILWAY_PROJECT_ID"]
        self.service_id = os.environ["RAILWAY_SERVICE_ID"]

    async def deploy_canary(
        self,
        traffic_percentage: int = 10,
        monitoring_duration: int = 300,
        error_threshold: float = 0.01
    ):
        """Deploy canary with gradual traffic shift."""
        print(f"Starting canary deployment with {traffic_percentage}% traffic")

        # 1. Create canary deployment
        canary_id = await self.create_canary_deployment()
        print(f"Canary deployment created: {canary_id}")

        # 2. Wait for canary to be ready
        await self.wait_for_ready(canary_id)
        print("Canary is ready")

        # 3. Route traffic to canary
        await self.set_traffic_split(canary_id, traffic_percentage)
        print(f"Routing {traffic_percentage}% traffic to canary")

        # 4. Monitor canary
        metrics = await self.monitor_canary(canary_id, monitoring_duration)
        print(f"Canary metrics: {metrics}")

        # 5. Decision: promote or rollback
        if metrics.error_rate <= error_threshold:
            await self.promote_canary(canary_id)
            print("Canary promoted to production")
            return True
        else:
            await self.rollback_canary(canary_id)
            print(f"Canary rolled back due to high error rate: {metrics.error_rate}")
            return False

    async def create_canary_deployment(self) -> str:
        """Create new canary deployment."""
        query = """
        mutation DeployService($serviceId: String!) {
            deployService(serviceId: $serviceId) {
                id
                status
            }
        }
        """
        result = await self.graphql_request(query, {
            "serviceId": self.service_id
        })
        return result["deployService"]["id"]

    async def wait_for_ready(self, deployment_id: str, timeout: int = 600):
        """Wait for deployment to be ready."""
        start_time = asyncio.get_event_loop().time()

        while asyncio.get_event_loop().time() - start_time < timeout:
            status = await self.get_deployment_status(deployment_id)

            if status == "READY":
                return
            elif status == "FAILED":
                raise Exception("Deployment failed")

            await asyncio.sleep(10)

        raise Exception("Deployment timeout")

    async def get_deployment_status(self, deployment_id: str) -> str:
        """Get deployment status."""
        query = """
        query GetDeployment($deploymentId: String!) {
            deployment(id: $deploymentId) {
                status
            }
        }
        """
        result = await self.graphql_request(query, {
            "deploymentId": deployment_id
        })
        return result["deployment"]["status"]

    async def set_traffic_split(self, canary_id: str, percentage: int):
        """Set traffic split between stable and canary."""
        query = """
        mutation SetTrafficSplit($serviceId: String!, $splits: [TrafficSplitInput!]!) {
            setTrafficSplit(serviceId: $serviceId, splits: $splits) {
                success
            }
        }
        """
        await self.graphql_request(query, {
            "serviceId": self.service_id,
            "splits": [
                {"deploymentId": canary_id, "percentage": percentage}
            ]
        })

    async def monitor_canary(
        self,
        canary_id: str,
        duration: int
    ) -> CanaryMetrics:
        """Monitor canary metrics for specified duration."""
        print(f"Monitoring canary for {duration} seconds...")

        error_count = 0
        total_requests = 0
        latencies = []

        start_time = asyncio.get_event_loop().time()

        while asyncio.get_event_loop().time() - start_time < duration:
            metrics = await self.fetch_metrics(canary_id)

            error_count += metrics.get("errors", 0)
            total_requests += metrics.get("requests", 0)
            latencies.extend(metrics.get("latencies", []))

            await asyncio.sleep(10)

        error_rate = error_count / total_requests if total_requests > 0 else 0
        latencies.sort()

        return CanaryMetrics(
            error_rate=error_rate,
            latency_p50=latencies[len(latencies) // 2] if latencies else 0,
            latency_p99=latencies[int(len(latencies) * 0.99)] if latencies else 0,
            request_count=total_requests
        )

    async def fetch_metrics(self, deployment_id: str) -> dict:
        """Fetch metrics for deployment."""
        query = """
        query GetMetrics($deploymentId: String!) {
            deploymentMetrics(deploymentId: $deploymentId) {
                requests
                errors
                latencies
            }
        }
        """
        result = await self.graphql_request(query, {
            "deploymentId": deployment_id
        })
        return result.get("deploymentMetrics", {})

    async def promote_canary(self, canary_id: str):
        """Promote canary to 100% traffic."""
        await self.set_traffic_split(canary_id, 100)

    async def rollback_canary(self, canary_id: str):
        """Rollback canary deployment."""
        await self.set_traffic_split(canary_id, 0)
        # Optionally delete the canary deployment

    async def graphql_request(self, query: str, variables: dict) -> dict:
        """Make GraphQL request to Railway API."""
        async with aiohttp.ClientSession() as session:
            async with session.post(
                self.api_url,
                json={"query": query, "variables": variables},
                headers={
                    "Authorization": f"Bearer {self.token}",
                    "Content-Type": "application/json"
                }
            ) as response:
                result = await response.json()
                if "errors" in result:
                    raise Exception(f"GraphQL error: {result['errors']}")
                return result.get("data", {})


async def main():
    deployer = RailwayCanaryDeployer()

    success = await deployer.deploy_canary(
        traffic_percentage=10,
        monitoring_duration=300,  # 5 minutes
        error_threshold=0.01  # 1% error rate threshold
    )

    if success:
        print("Canary deployment successful!")
    else:
        print("Canary deployment failed, rolled back to stable")
        exit(1)


if __name__ == "__main__":
    asyncio.run(main())
```

---

Last Updated: 2025-12-07
Examples: Production-ready implementations for Vercel and Railway
Patterns: Blue-Green, Canary, Multi-Service, CI/CD
