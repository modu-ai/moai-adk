---
name: moai-platform-deploy
description: Deployment platform specialist covering Vercel and Railway for edge deployment, containerized applications, CI/CD automation, and cross-platform deployment strategies. Use when deploying Next.js apps, configuring edge functions, or setting up container-based infrastructure.
version: 1.0.0
category: platform
updated: 2025-12-07
status: active
allowed-tools: Read, Write, Bash, Grep, Glob
---

# moai-platform-deploy: Deployment Platform Specialist

## Quick Reference (30 seconds)

Deployment Platform Unification: Comprehensive deployment patterns for Vercel (edge-first) and Railway (container-first) platforms with CI/CD automation, multi-environment management, and zero-downtime deployment strategies.

### Platform Selection Matrix

Vercel Optimal Use Cases:
- Next.js applications with SSR/SSG/ISR patterns
- Edge Functions for global low-latency APIs
- JAMstack static sites with CDN distribution
- Preview deployments for PR-based workflows
- Web Vitals monitoring and analytics

Railway Optimal Use Cases:
- Full-stack containerized applications
- Multi-service architectures with databases
- Backend services with persistent connections
- Custom runtime requirements (Python, Go, Rust)
- Multi-region container deployment

### Quick Platform Decision

Edge Performance Critical:
- YES: Vercel (Global Edge Network)
- Next.js Framework: Vercel (Optimized Runtime)

Container Flexibility Required:
- YES: Railway (Docker/Nixpacks)
- Multi-Service Architecture: Railway (Service Mesh)

### Key Capabilities
- Edge Deployment: Vercel Edge Functions with global CDN
- Container Deployment: Railway Docker and Nixpacks builds
- CI/CD Automation: Git-based deployment pipelines
- Environment Management: Staging, preview, production environments
- Scaling Strategies: Vertical, horizontal, and auto-scaling patterns
- Zero-Downtime: Blue-green and canary release patterns

---

## Implementation Guide

### Phase 1: Vercel Edge Deployment

Edge Network Architecture:
- Global CDN with 30+ edge locations
- Edge Functions for compute at the edge
- Automatic SSL and custom domain management
- Image optimization and static asset caching

Next.js Optimization Patterns:

ISR (Incremental Static Regeneration):
```typescript
// app/products/[id]/page.tsx
export const revalidate = 60 // Revalidate every 60 seconds

export async function generateStaticParams() {
  const products = await fetchTopProducts()
  return products.map(p => ({ id: p.id }))
}

export default async function ProductPage({ params }) {
  const product = await fetchProduct(params.id)
  return <ProductDetail product={product} />
}
```

Edge Functions for API Routes:
```typescript
// app/api/geo/route.ts
export const runtime = 'edge'

export async function GET(request: Request) {
  const geo = request.geo
  return Response.json({
    country: geo?.country,
    city: geo?.city,
    region: geo?.region
  })
}
```

Middleware for Edge Logic:
```typescript
// middleware.ts
import { NextResponse } from 'next/server'
import type { NextRequest } from 'next/server'

export function middleware(request: NextRequest) {
  const country = request.geo?.country || 'US'

  // Geo-based routing
  if (country === 'DE') {
    return NextResponse.rewrite(new URL('/de', request.url))
  }

  // A/B testing at the edge
  const bucket = Math.random() < 0.5 ? 'control' : 'variant'
  const response = NextResponse.next()
  response.cookies.set('ab-test', bucket)

  return response
}

export const config = {
  matcher: ['/((?!api|_next/static|favicon.ico).*)']
}
```

### Phase 2: Vercel Configuration

vercel.json Configuration:
```json
{
  "$schema": "https://openapi.vercel.sh/vercel.json",
  "framework": "nextjs",
  "buildCommand": "pnpm build",
  "installCommand": "pnpm install",
  "outputDirectory": ".next",
  "regions": ["iad1", "sfo1", "fra1"],
  "functions": {
    "api/**/*.ts": {
      "memory": 1024,
      "maxDuration": 30
    }
  },
  "crons": [
    {
      "path": "/api/cron/cleanup",
      "schedule": "0 0 * * *"
    }
  ],
  "headers": [
    {
      "source": "/api/(.*)",
      "headers": [
        { "key": "Cache-Control", "value": "s-maxage=60" }
      ]
    }
  ],
  "rewrites": [
    { "source": "/blog/:slug", "destination": "/posts/:slug" }
  ]
}
```

Environment Variables Management:
```bash
# Production environment
vercel env add DATABASE_URL production
vercel env add API_SECRET production

# Preview environments (PR deployments)
vercel env add DATABASE_URL preview
vercel env add API_SECRET preview

# Development environment
vercel env add DATABASE_URL development
```

### Phase 3: Railway Container Deployment

Container Deployment Patterns:

Dockerfile Configuration:
```dockerfile
# Multi-stage build for production
FROM node:20-alpine AS base
WORKDIR /app
COPY package*.json ./

FROM base AS deps
RUN npm ci --only=production

FROM base AS builder
RUN npm ci
COPY . .
RUN npm run build

FROM base AS runner
ENV NODE_ENV=production
COPY --from=deps /app/node_modules ./node_modules
COPY --from=builder /app/dist ./dist
USER node
EXPOSE 3000
CMD ["node", "dist/main.js"]
```

Nixpacks Configuration (nixpacks.toml):
```toml
[phases.setup]
nixPkgs = ["nodejs-20_x", "pnpm"]

[phases.install]
cmds = ["pnpm install --frozen-lockfile"]

[phases.build]
cmds = ["pnpm build"]

[start]
cmd = "pnpm start"
```

Railway Configuration (railway.toml):
```toml
[build]
builder = "NIXPACKS"
watchPatterns = ["src/**", "package.json"]

[deploy]
startCommand = "npm start"
healthcheckPath = "/health"
healthcheckTimeout = 100
restartPolicyType = "ON_FAILURE"
restartPolicyMaxRetries = 5

[deploy.resources]
memory = "512Mi"
cpu = "0.5"
```

### Phase 4: Multi-Service Architecture (Railway)

Service Mesh Configuration:
```yaml
# railway.yaml
services:
  web:
    build:
      dockerfile: ./apps/web/Dockerfile
    deploy:
      replicas: 2
      resources:
        memory: 512Mi
    depends_on:
      - api
      - redis

  api:
    build:
      dockerfile: ./apps/api/Dockerfile
    deploy:
      replicas: 3
      resources:
        memory: 1Gi
    environment:
      DATABASE_URL: ${{Postgres.DATABASE_URL}}
      REDIS_URL: ${{Redis.REDIS_URL}}

  worker:
    build:
      dockerfile: ./apps/worker/Dockerfile
    deploy:
      replicas: 2
    environment:
      REDIS_URL: ${{Redis.REDIS_URL}}
```

Private Networking:
```typescript
// Internal service communication
const API_URL = process.env.RAILWAY_PRIVATE_DOMAIN
  ? `http://${process.env.RAILWAY_PRIVATE_DOMAIN}:3000`
  : 'http://localhost:3000'

// Health check endpoint
app.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    service: process.env.RAILWAY_SERVICE_NAME,
    replica: process.env.RAILWAY_REPLICA_ID
  })
})
```

### Phase 5: CI/CD Automation

GitHub Actions for Vercel:
```yaml
name: Vercel Production Deployment
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install Vercel CLI
        run: npm i -g vercel@latest

      - name: Pull Vercel Environment
        run: vercel pull --yes --environment=production --token=${{ secrets.VERCEL_TOKEN }}

      - name: Build Project
        run: vercel build --prod --token=${{ secrets.VERCEL_TOKEN }}

      - name: Deploy to Vercel
        run: vercel deploy --prebuilt --prod --token=${{ secrets.VERCEL_TOKEN }}
```

Railway Auto-Deploy Configuration:
```yaml
name: Railway Deployment
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install Railway CLI
        run: npm i -g @railway/cli

      - name: Deploy to Railway
        run: railway up --detach
        env:
          RAILWAY_TOKEN: ${{ secrets.RAILWAY_TOKEN }}
```

### Phase 6: Zero-Downtime Deployment

Blue-Green Deployment (Vercel):
```typescript
// Vercel Alias Switching
import { Vercel } from '@vercel/sdk'

const vercel = new Vercel({ token: process.env.VERCEL_TOKEN })

async function blueGreenDeploy(projectId: string, deploymentId: string) {
  // Create new deployment (green)
  const deployment = await vercel.deployments.create({
    name: projectId,
    gitSource: { type: 'github', repo: 'org/repo' }
  })

  // Run smoke tests on preview URL
  const smokeTestPassed = await runSmokeTests(deployment.url)

  if (smokeTestPassed) {
    // Switch production alias to green
    await vercel.aliases.assign({
      alias: 'production.example.com',
      deployment: deploymentId
    })
  }
}
```

Canary Releases (Railway):
```python
# Gradual traffic shifting
class CanaryDeployer:
    def __init__(self, railway_client):
        self.client = railway_client

    async def deploy_canary(self, service_id: str, percentage: int = 10):
        """Deploy canary with gradual traffic shift."""

        # Deploy new version as canary
        canary = await self.client.deploy(
            service_id=service_id,
            canary=True,
            traffic_percentage=percentage
        )

        # Monitor metrics
        metrics = await self.monitor_canary(canary.id, duration=300)

        if metrics.error_rate < 0.01:  # Less than 1% error rate
            # Gradually increase traffic
            for pct in [25, 50, 75, 100]:
                await self.client.update_traffic(canary.id, pct)
                await asyncio.sleep(60)
        else:
            # Rollback
            await self.client.rollback(canary.id)
```

---

## Monorepo Deployment Patterns

### Turborepo with Vercel

Root Configuration (turbo.json):
```json
{
  "$schema": "https://turbo.build/schema.json",
  "pipeline": {
    "build": {
      "dependsOn": ["^build"],
      "outputs": [".next/**", "dist/**"]
    },
    "deploy:web": {
      "dependsOn": ["build"],
      "cache": false
    },
    "deploy:api": {
      "dependsOn": ["build"],
      "cache": false
    }
  }
}
```

Per-App vercel.json:
```json
{
  "buildCommand": "cd ../.. && pnpm turbo build --filter=web",
  "installCommand": "cd ../.. && pnpm install",
  "framework": "nextjs",
  "outputDirectory": ".next"
}
```

### pnpm Workspaces with Railway

Workspace Configuration (pnpm-workspace.yaml):
```yaml
packages:
  - 'apps/*'
  - 'packages/*'
```

Selective Deployment:
```bash
# Deploy specific workspace app
railway up --service web --cwd apps/web
railway up --service api --cwd apps/api
```

---

## Context7 Integration

Vercel Documentation Access:
```python
async def get_vercel_patterns():
    """Get latest Vercel deployment patterns."""
    docs = await mcp__context7__get_library_docs(
        context7CompatibleLibraryID="/vercel/vercel",
        topic="edge functions deployment next.js optimization",
        tokens=5000
    )
    return docs
```

Railway Documentation Access:
```python
async def get_railway_patterns():
    """Get latest Railway container patterns."""
    docs = await mcp__context7__get_library_docs(
        context7CompatibleLibraryID="/railwayapp/docs",
        topic="container deployment nixpacks multi-region",
        tokens=4000
    )
    return docs
```

---

## Works Well With

- `moai-platform-baas` - BaaS provider integration (Auth, Database)
- `moai-domain-frontend` - Frontend framework deployment patterns
- `moai-domain-backend` - Backend service architecture
- `moai-foundation-quality` - Deployment validation and testing

## Advanced Patterns

For multi-region failover, custom domains, database migrations, secrets rotation, and disaster recovery patterns, see [reference.md](reference.md) and [examples.md](examples.md).

---

Status: Production Ready | Version: 1.0.0 | Updated: 2025-12-07
