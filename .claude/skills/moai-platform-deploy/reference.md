# Deployment Platform Reference Documentation

## Vercel Platform Reference

### Context7 Integration
Context7 ID: `/vercel/vercel`
Focus: Edge deployment, Next.js optimization, serverless functions

### Edge Network Architecture

Global Edge Locations:
- North America: Washington DC (iad1), San Francisco (sfo1), Toronto (yyz1)
- Europe: Frankfurt (fra1), London (lhr1), Amsterdam (ams1), Stockholm (arn1)
- Asia Pacific: Singapore (sin1), Tokyo (hnd1), Sydney (syd1), Mumbai (bom1)
- South America: Sao Paulo (gru1)

Edge Function Capabilities:
- Runtime: V8 Isolates (lightweight, fast cold starts)
- Memory: Up to 1024 MB (configurable)
- Duration: Up to 30 seconds (Pro/Enterprise)
- Streaming: Response streaming supported
- WebSocket: Limited (use Serverless Functions)

### Next.js Rendering Strategies

Static Site Generation (SSG):
```typescript
// Build-time static generation
export async function generateStaticParams() {
  const posts = await getAllPosts()
  return posts.map((post) => ({
    slug: post.slug
  }))
}

// Static page with build-time data
export default async function Page({ params }: { params: { slug: string } }) {
  const post = await getPost(params.slug)
  return <Post post={post} />
}
```

Incremental Static Regeneration (ISR):
```typescript
// Time-based revalidation
export const revalidate = 3600 // Revalidate every hour

// On-demand revalidation
// pages/api/revalidate.ts
export default async function handler(req, res) {
  const { path, secret } = req.query

  if (secret !== process.env.REVALIDATION_SECRET) {
    return res.status(401).json({ message: 'Invalid token' })
  }

  try {
    await res.revalidate(path)
    return res.json({ revalidated: true })
  } catch (err) {
    return res.status(500).json({ message: 'Error revalidating' })
  }
}
```

Server-Side Rendering (SSR):
```typescript
// Dynamic rendering per request
export const dynamic = 'force-dynamic'

export default async function Page() {
  const data = await fetchDynamicData()
  return <DynamicComponent data={data} />
}
```

### Edge Functions Deep Dive

Edge Middleware Patterns:
```typescript
// middleware.ts
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export function middleware(request: NextRequest) {
  // 1. Authentication check
  const token = request.cookies.get('auth-token')
  if (!token && request.nextUrl.pathname.startsWith('/dashboard')) {
    return NextResponse.redirect(new URL('/login', request.url))
  }

  // 2. Rate limiting at edge
  const ip = request.ip ?? 'unknown'
  const rateLimit = checkRateLimit(ip)
  if (!rateLimit.allowed) {
    return new NextResponse('Too Many Requests', { status: 429 })
  }

  // 3. Feature flags
  const flags = getFeatureFlags(request)
  const response = NextResponse.next()
  response.headers.set('x-feature-flags', JSON.stringify(flags))

  // 4. A/B testing
  const variant = request.cookies.get('ab-variant')?.value
  if (!variant) {
    const newVariant = Math.random() < 0.5 ? 'A' : 'B'
    response.cookies.set('ab-variant', newVariant)
  }

  return response
}
```

Edge API Routes:
```typescript
// app/api/edge-example/route.ts
import { geolocation } from '@vercel/functions'

export const runtime = 'edge'

export async function GET(request: Request) {
  const geo = geolocation(request)

  // Edge-computed response
  const response = {
    location: {
      country: geo.country,
      city: geo.city,
      region: geo.region,
      latitude: geo.latitude,
      longitude: geo.longitude
    },
    timestamp: new Date().toISOString(),
    edge: true
  }

  return Response.json(response, {
    headers: {
      'Cache-Control': 'public, s-maxage=60',
      'CDN-Cache-Control': 'public, max-age=300'
    }
  })
}
```

### Vercel Analytics and Monitoring

Web Vitals Integration:
```typescript
// app/layout.tsx
import { SpeedInsights } from '@vercel/speed-insights/next'
import { Analytics } from '@vercel/analytics/react'

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        {children}
        <SpeedInsights />
        <Analytics />
      </body>
    </html>
  )
}
```

Custom Metrics:
```typescript
import { track } from '@vercel/analytics'

// Track custom events
track('purchase', {
  product_id: 'prod_123',
  amount: 99.99,
  currency: 'USD'
})

// Track page views with metadata
track('page_view', {
  page: '/products',
  category: 'electronics'
})
```

### Vercel Cron Jobs

Scheduled Functions:
```json
// vercel.json
{
  "crons": [
    {
      "path": "/api/cron/daily-cleanup",
      "schedule": "0 0 * * *"
    },
    {
      "path": "/api/cron/hourly-sync",
      "schedule": "0 * * * *"
    },
    {
      "path": "/api/cron/weekly-report",
      "schedule": "0 9 * * 1"
    }
  ]
}
```

Cron Handler:
```typescript
// app/api/cron/daily-cleanup/route.ts
import { NextRequest } from 'next/server'

export async function GET(request: NextRequest) {
  // Verify cron secret
  const authHeader = request.headers.get('authorization')
  if (authHeader !== `Bearer ${process.env.CRON_SECRET}`) {
    return new Response('Unauthorized', { status: 401 })
  }

  try {
    await performCleanup()
    return Response.json({ success: true, timestamp: new Date().toISOString() })
  } catch (error) {
    return Response.json({ error: error.message }, { status: 500 })
  }
}
```

### Domain and SSL Configuration

Custom Domain Setup:
```bash
# Add custom domain
vercel domains add example.com

# Configure DNS (A record)
# @ -> 76.76.21.21

# Configure www redirect
vercel domains add www.example.com --redirect example.com

# Verify SSL certificate
vercel certs ls
```

Advanced Domain Configuration:
```json
// vercel.json
{
  "redirects": [
    {
      "source": "/old-blog/:path*",
      "destination": "/blog/:path*",
      "permanent": true
    }
  ],
  "headers": [
    {
      "source": "/(.*)",
      "headers": [
        {
          "key": "Strict-Transport-Security",
          "value": "max-age=63072000; includeSubDomains; preload"
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

---

## Railway Platform Reference

### Context7 Integration
Context7 ID: `/railwayapp/docs`
Focus: Container deployment, multi-service architecture, managed databases

### Container Build Options

Dockerfile Build:
```dockerfile
# Production-optimized multi-stage build
FROM node:20-alpine AS base
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

FROM base AS deps
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile

FROM base AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN pnpm build

FROM base AS runner
WORKDIR /app
ENV NODE_ENV=production
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 appuser
COPY --from=builder --chown=appuser:nodejs /app/dist ./dist
COPY --from=builder --chown=appuser:nodejs /app/node_modules ./node_modules
USER appuser
EXPOSE 3000
CMD ["node", "dist/main.js"]
```

Nixpacks Build Configuration:
```toml
# nixpacks.toml
[phases.setup]
nixPkgs = ["nodejs-20_x", "pnpm", "python310"]
aptPkgs = ["libpq-dev"]

[phases.install]
cmds = [
  "pnpm install --frozen-lockfile",
  "pnpm prisma generate"
]

[phases.build]
cmds = [
  "pnpm build"
]

[start]
cmd = "pnpm start:prod"

[variables]
NODE_ENV = "production"
```

### Railway Service Configuration

Complete railway.toml:
```toml
[build]
builder = "NIXPACKS"
buildCommand = "pnpm build"
watchPatterns = ["src/**", "prisma/**"]

[deploy]
startCommand = "pnpm start:prod"
healthcheckPath = "/api/health"
healthcheckTimeout = 300
restartPolicyType = "ON_FAILURE"
restartPolicyMaxRetries = 10
numReplicas = 2

[deploy.resources]
memory = "1Gi"
cpu = "1"

[[deploy.scaling]]
minReplicas = 1
maxReplicas = 5
metric = "cpu"
target = 80
```

### Multi-Region Deployment

Region Configuration:
```python
# Railway supports these regions
RAILWAY_REGIONS = {
    "us-west1": "Oregon, USA",
    "us-east4": "Virginia, USA",
    "europe-west4": "Netherlands, EU",
    "asia-southeast1": "Singapore, APAC"
}

# Multi-region service configuration
class MultiRegionDeployer:
    def __init__(self, railway_client):
        self.client = railway_client

    async def deploy_multi_region(self, service_config: dict):
        """Deploy service across multiple regions."""
        deployments = []

        for region in service_config['regions']:
            deployment = await self.client.deploy(
                project_id=service_config['project_id'],
                service_id=service_config['service_id'],
                region=region,
                environment=service_config['environment']
            )
            deployments.append(deployment)

        return deployments
```

### Railway Database Services

PostgreSQL Configuration:
```python
# Database connection with connection pooling
import asyncpg
from contextlib import asynccontextmanager

class DatabasePool:
    def __init__(self):
        self.pool = None

    async def initialize(self):
        self.pool = await asyncpg.create_pool(
            dsn=os.environ['DATABASE_URL'],
            min_size=5,
            max_size=20,
            command_timeout=60,
            server_settings={
                'application_name': 'railway-app',
                'jit': 'off'
            }
        )

    @asynccontextmanager
    async def connection(self):
        async with self.pool.acquire() as conn:
            yield conn

    async def close(self):
        await self.pool.close()
```

Redis Configuration:
```typescript
// Redis connection with retry logic
import Redis from 'ioredis'

const redis = new Redis(process.env.REDIS_URL, {
  maxRetriesPerRequest: 3,
  retryDelayOnFailover: 100,
  enableReadyCheck: true,
  connectTimeout: 10000,
  lazyConnect: true,
  reconnectOnError: (err) => {
    const targetErrors = ['READONLY', 'ECONNRESET', 'ETIMEDOUT']
    return targetErrors.some(e => err.message.includes(e))
  }
})

redis.on('error', (err) => console.error('Redis Error:', err))
redis.on('connect', () => console.log('Redis connected'))
```

### Private Networking

Internal Service Communication:
```typescript
// Railway provides private DNS for internal communication
const INTERNAL_SERVICES = {
  api: process.env.API_PRIVATE_URL || 'http://api.railway.internal:3000',
  worker: process.env.WORKER_PRIVATE_URL || 'http://worker.railway.internal:3001',
  cache: process.env.REDIS_PRIVATE_URL || 'redis://redis.railway.internal:6379'
}

// gRPC internal service
const grpcClient = new ServiceClient(
  INTERNAL_SERVICES.api,
  grpc.credentials.createInsecure()
)
```

TCP Proxy Configuration:
```toml
# For non-HTTP services (databases, custom protocols)
[deploy]
tcpProxyApplicationPort = 5432

# For WebSocket support
[deploy]
wsProxyPath = "/ws"
```

### Environment Management

Environment Configuration:
```bash
# Create environments
railway environment create staging
railway environment create production

# Set environment-specific variables
railway variables set DATABASE_URL="postgresql://..." --environment production
railway variables set API_KEY="prod_key" --environment production

railway variables set DATABASE_URL="postgresql://..." --environment staging
railway variables set API_KEY="staging_key" --environment staging

# Link local development to environment
railway link --environment staging
```

Secrets Management:
```python
# Accessing Railway secrets in application
import os
from cryptography.fernet import Fernet

class SecretsManager:
    def __init__(self):
        self.encryption_key = os.environ.get('ENCRYPTION_KEY')
        self.fernet = Fernet(self.encryption_key) if self.encryption_key else None

    def get_secret(self, key: str) -> str:
        """Get decrypted secret from environment."""
        encrypted_value = os.environ.get(f'{key}_ENCRYPTED')
        if encrypted_value and self.fernet:
            return self.fernet.decrypt(encrypted_value.encode()).decode()
        return os.environ.get(key, '')

    def rotate_secrets(self, secrets: dict):
        """Rotate secrets with zero-downtime."""
        # Railway supports instant environment variable updates
        for key, value in secrets.items():
            encrypted = self.fernet.encrypt(value.encode()).decode()
            # Update via Railway API
            self.update_railway_variable(f'{key}_ENCRYPTED', encrypted)
```

---

## Cross-Platform Deployment Strategies

### Hybrid Deployment Pattern

Vercel for Frontend + Railway for Backend:
```yaml
# GitHub Actions workflow for hybrid deployment
name: Hybrid Deployment

on:
  push:
    branches: [main]

jobs:
  deploy-frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Deploy to Vercel
        uses: amondnet/vercel-action@v25
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.VERCEL_ORG_ID }}
          vercel-project-id: ${{ secrets.VERCEL_PROJECT_ID }}
          working-directory: ./apps/web

  deploy-backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install Railway CLI
        run: npm i -g @railway/cli

      - name: Deploy API to Railway
        run: railway up --service api
        env:
          RAILWAY_TOKEN: ${{ secrets.RAILWAY_TOKEN }}
        working-directory: ./apps/api
```

### Rollback Strategies

Vercel Instant Rollback:
```typescript
import { Vercel } from '@vercel/sdk'

async function rollbackToDeployment(deploymentId: string) {
  const vercel = new Vercel({ token: process.env.VERCEL_TOKEN })

  // Get deployment details
  const deployment = await vercel.deployments.get(deploymentId)

  // Create new deployment from same git ref
  const rollback = await vercel.deployments.create({
    name: deployment.name,
    gitSource: {
      type: 'github',
      ref: deployment.meta.githubCommitRef,
      repo: deployment.meta.githubRepo
    }
  })

  // Promote to production
  await vercel.aliases.assign({
    alias: 'production.example.com',
    deployment: rollback.id
  })

  return rollback
}
```

Railway Rollback:
```bash
# List recent deployments
railway deployments list

# Rollback to specific deployment
railway rollback <deployment-id>

# Rollback to previous deployment
railway rollback --previous
```

### Disaster Recovery

Multi-Platform Failover:
```python
class DisasterRecovery:
    def __init__(self):
        self.primary = 'vercel'
        self.secondary = 'railway'
        self.health_check_interval = 30

    async def monitor_and_failover(self):
        """Monitor primary and failover to secondary if needed."""
        while True:
            primary_healthy = await self.check_health(self.primary)

            if not primary_healthy:
                await self.trigger_failover()
                await self.notify_team('Primary platform unhealthy, failover initiated')

            await asyncio.sleep(self.health_check_interval)

    async def check_health(self, platform: str) -> bool:
        """Check platform health status."""
        endpoints = {
            'vercel': 'https://api.vercel.com/v1/status',
            'railway': 'https://railway.app/api/health'
        }

        try:
            async with aiohttp.ClientSession() as session:
                async with session.get(endpoints[platform]) as response:
                    return response.status == 200
        except Exception:
            return False

    async def trigger_failover(self):
        """Execute failover to secondary platform."""
        # Update DNS to point to secondary
        await self.update_dns_records(self.secondary)

        # Scale up secondary
        await self.scale_secondary_resources()

        # Notify monitoring systems
        await self.update_monitoring_config()
```

---

## Performance Optimization

### Vercel Edge Caching

Cache Headers Configuration:
```typescript
// Optimal caching strategy
export async function GET(request: Request) {
  const data = await fetchData()

  return Response.json(data, {
    headers: {
      // CDN cache for 1 hour
      'Cache-Control': 'public, s-maxage=3600, stale-while-revalidate=86400',
      // Browser cache for 5 minutes
      'CDN-Cache-Control': 'public, max-age=300',
      // Vary by accept-encoding for compression
      'Vary': 'Accept-Encoding'
    }
  })
}
```

### Railway Resource Optimization

Autoscaling Configuration:
```toml
# railway.toml
[[deploy.scaling]]
minReplicas = 1
maxReplicas = 10
metric = "cpu"
target = 70

[[deploy.scaling]]
minReplicas = 2
maxReplicas = 8
metric = "memory"
target = 80

[[deploy.scaling]]
minReplicas = 1
maxReplicas = 5
metric = "requests_per_second"
target = 1000
```

---

## Security Best Practices

### Vercel Security Headers

Security Configuration:
```json
// vercel.json
{
  "headers": [
    {
      "source": "/(.*)",
      "headers": [
        {
          "key": "Content-Security-Policy",
          "value": "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'"
        },
        {
          "key": "X-Frame-Options",
          "value": "DENY"
        },
        {
          "key": "X-Content-Type-Options",
          "value": "nosniff"
        },
        {
          "key": "Referrer-Policy",
          "value": "strict-origin-when-cross-origin"
        },
        {
          "key": "Permissions-Policy",
          "value": "camera=(), microphone=(), geolocation=()"
        }
      ]
    }
  ]
}
```

### Railway Security

Container Security:
```dockerfile
# Security-hardened Dockerfile
FROM node:20-alpine AS runner

# Run as non-root user
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 appuser

# Set secure permissions
WORKDIR /app
RUN chown -R appuser:nodejs /app

# No root access
USER appuser

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:3000/health || exit 1

EXPOSE 3000
CMD ["node", "dist/main.js"]
```

---

## Cost Optimization

### Vercel Pricing Optimization

Usage-Based Pricing Strategy:
- Edge Function Invocations: Optimize with caching
- Bandwidth: Enable compression, optimize images
- Build Minutes: Use incremental builds
- Serverless Function Duration: Optimize cold starts

### Railway Pricing Optimization

Resource Optimization:
- Right-size memory and CPU allocations
- Use sleep/wake for development environments
- Implement autoscaling for variable workloads
- Share databases between staging environments

---

Last Updated: 2025-12-07
Platforms Documented: Vercel, Railway
Context7 IDs: /vercel/vercel, /railwayapp/docs
