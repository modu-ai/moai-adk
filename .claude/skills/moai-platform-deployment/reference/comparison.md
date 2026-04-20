# Platform Comparison Guide

Comprehensive comparison of Vercel, Railway, and Convex deployment platforms.

---

## Feature Matrix

| Feature | Vercel | Railway | Convex |
|---------|--------|---------|--------|
| **Primary Use Case** | Edge-First Web Apps | Containerized Apps | Real-Time Backends |
| **Best Framework** | Next.js | Any containerized app | TypeScript apps |
| **Compute Model** | Edge Functions (serverless) | Containers (Docker) | Serverless (managed) |
| **Global Distribution** | 30+ edge locations | Multi-region (5 regions) | Single region (US) |
| **Cold Starts** | Sub-50ms | No cold starts | Fast warm starts |
| **Storage Options** | KV, Blob, Postgres | Volumes, managed DBs | Built-in document DB |
| **Real-Time Sync** | No | No (WebSockets only) | Yes (built-in) |
| **Private Networking** | No | Yes (service mesh) | No |
| **Auto-Scaling** | Yes (serverless) | Yes (configurable) | Yes (automatic) |
| **Docker Support** | Limited | Full (Dockerfile) | No |
| **Custom Runtimes** | Limited | Full | No (TypeScript only) |
| **Preview Deployments** | Yes (built-in) | No | N/A (backend only) |
| **Edge Middleware** | Yes | No | No |
| **Managed Databases** | Postgres, KV, Blob | Postgres, MySQL, Redis | Built-in document DB |
| **File Storage** | Blob storage | Volume mounts | Built-in storage |
| **WebSocket Support** | No | Yes | Yes (via sync) |
| **Cron Jobs** | Yes | Yes | Yes |
| **HTTP Endpoints** | Yes | Yes | Yes (webhooks) |

---

## Pricing Comparison

### Vercel Pricing

**Hobby (Free)**:
- 100GB bandwidth/month
- Serverless Function execution limits
- 1000 edge function invocations/day
- 60-second execution limit

**Pro ($20/month)**:
- 1TB bandwidth/month
- Unlimited serverless functions
- 100-second execution limit
- Team collaboration
- Analytics

**Enterprise**:
- Custom pricing
- SSO/SAML
- SLA
- Dedicated support

**Storage Costs**:
- KV: $0.50/GB read + $0.15/GB stored
- Blob: $0.15/GB stored + $0.18/GB egress
- Postgres: $0.125/hour for smallest instance

### Railway Pricing

**Free Trial**:
- $5 free credit/month
- Community support

**Pay-As-You-Go**:
- $0.000361/second per service (~$31.50/month per service)
- $0.15/GB volume storage
- $0.25/GB network egress

**Databases**:
- Postgres: Included in service pricing
- MySQL: Included in service pricing
- Redis: Included in service pricing

**Example Costs**:
- 2 services (web + api): ~$63/month
- 3 services + 2 databases: ~$157/month

### Convex Pricing

**Free Tier**:
- 500K queries/month
- 500K mutations/month
- 5GB storage
- 3 HTTP actions/day

**Pro ($25/month)**:
- 10M queries/month
- 10M mutations/month
- 100GB storage
- Unlimited HTTP actions

**Enterprise**:
- Custom pricing
- Dedicated support
- SLA

---

## Use Case Scenarios

### Scenario 1: Next.js E-Commerce Site

**Recommended**: Vercel

**Why**:
- Native Next.js optimization
- ISR for product pages
- Global CDN for fast load times
- Preview deployments for A/B testing
- Vercel Postgres for database

**Architecture**:
```
Frontend (Next.js on Vercel)
  → Edge Functions for geo-routing
  → ISR for product pages
  → Vercel Postgres for products/orders
  → Vercel KV for session/cache
```

**Estimated Cost**: $20/month (Pro) + storage costs

### Scenario 2: Microservices API Platform

**Recommended**: Railway

**Why**:
- Full Docker support for custom runtimes
- Private networking for service-to-service communication
- Persistent volumes for stateful services
- Multi-region deployment
- Auto-scaling

**Architecture**:
```
API Gateway (Railway)
  → Service A (Node.js)
  → Service B (Python)
  → Service C (Go)
  → Postgres (Railway)
  → Redis (Railway)
```

**Estimated Cost**: ~$150-200/month (5 services + 2 databases)

### Scenario 3: Real-Time Collaborative Editor

**Recommended**: Convex + Vercel/Railway (frontend)

**Why**:
- Built-in real-time sync
- Optimistic updates
- TypeScript type safety
- Automatic caching
- Reactive queries

**Architecture**:
```
Frontend (Next.js on Vercel)
  → Convex backend (real-time sync)
  → Auth via Clerk/Auth0
```

**Estimated Cost**: $25/month (Convex Pro) + $20/month (Vercel Pro)

### Scenario 4: High-Performance Global API

**Recommended**: Vercel Edge Functions

**Why**:
- Sub-50ms edge latency
- Global distribution (30+ locations)
- Auto-scaling
- Simple pay-as-you-go

**Architecture**:
```
Edge Functions (Vercel)
  → Upstream API or database
  → Edge caching
```

**Estimated Cost**: $20/month (Pro) + compute

### Scenario 5: Legacy App Migration

**Recommended**: Railway

**Why**:
- Full Docker support
- Custom runtime environments
- Direct container migration
- No code refactoring needed

**Architecture**:
```
Existing app in Docker container
  → Deploy to Railway
  → Configure environment variables
  → Set up private networking
```

**Estimated Cost**: ~$63/month per service

---

## Migration Guides

### Migrate from Vercel to Railway

**Step 1: Create Dockerfile**
```dockerfile
# If your app is Node.js/Next.js
FROM node:20-alpine
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build
CMD ["npm", "start"]
```

**Step 2: Configure railway.toml**
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "Dockerfile"
```

**Step 3: Move Environment Variables**
```bash
# Export from Vercel
vercel env pull .env

# Import to Railway
railway variables --import .env
```

**Step 4: Update DNS**
- Remove domain from Vercel
- Add domain to Railway
- Update DNS records

**Step 5: Deploy**
```bash
railway up
```

### Migrate from Railway to Vercel

**Step 1: Vercelize Your App**
- If using Next.js: native support
- If custom Docker: Use Vercel's Docker buildpack
- Otherwise: Rewrite as serverless functions

**Step 2: Handle State**
- Move Railway volumes to Vercel Blob
- Migrate databases to Vercel Postgres
- Use Vercel KV for caching

**Step 3: Configure vercel.json**
```json
{
  "buildCommand": "npm run build",
  "outputDirectory": "dist",
  "functions": {
    "api/**/*.ts": {
      "memory": 1024,
      "maxDuration": 10
    }
  }
}
```

**Step 4: Deploy**
```bash
vercel
```

### Migrate to Convex Backend

**Step 1: Define Schema**
```typescript
// convex/schema.ts
export default defineSchema({
  users: defineTable({
    name: v.string(),
    email: v.string(),
  }),
  posts: defineTable({
    title: v.string(),
    content: v.string(),
    author: v.id("users"),
  }),
})
```

**Step 2: Create Server Functions**
```typescript
// convex/posts.ts
export const list = query({
  handler: async (ctx) => {
    return await ctx.db.query("posts").collect()
  }
})

export const create = mutation({
  args: { title: v.string(), content: v.string() },
  handler: async (ctx, args) => {
    await ctx.db.insert("posts", args)
  }
})
```

**Step 3: Update Frontend**
```typescript
// Replace API calls with Convex hooks
import { useQuery, useMutation } from "convex/react"
import { api } from "../convex/_generated/api"

const posts = useQuery(api.posts.list)
const createPost = useMutation(api.posts.create)
```

**Step 4: Deploy**
```bash
npx convex dev
npx convex deploy
```

---

## Decision Tree

```
Start
 │
 ├─ Need real-time sync?
 │   ├─ Yes → Convex
 │   └─ No → Continue
 │
 ├─ Using Next.js?
 │   ├─ Yes → Vercel (optimal)
 │   └─ No → Continue
 │
 ├─ Need custom runtime (Python/Go/Rust)?
 │   ├─ Yes → Railway
 │   └─ No → Continue
 │
 ├─ Need private networking?
 │   ├─ Yes → Railway
 │   └─ No → Continue
 │
 ├─ Global latency critical (<50ms)?
 │   ├─ Yes → Vercel
 │   └─ No → Continue
 │
 ├─ Need persistent volumes?
 │   ├─ Yes → Railway
 │   └─ No → Continue
 │
 └─ Default → Vercel (easiest)
```

---

## Hybrid Architectures

### Pattern 1: Frontend on Vercel + Backend on Railway

**Use Case**: Next.js frontend with complex backend services

**Architecture**:
```
Next.js (Vercel)
  → API routes → Railway services (private network)
  → Static assets → Vercel CDN
  → Edge middleware for routing
```

**Benefits**:
- Frontend gets Vercel's CDN and edge capabilities
- Backend gets Railway's Docker and private networking
- Optimal tool for each layer

### Pattern 2: Frontend on Vercel/Railway + Convex Backend

**Use Case**: Real-time applications

**Architecture**:
```
React/Next.js (Vercel/Railway)
  → Convex client
  → Real-time sync
  → Auth (Clerk/Auth0)
```

**Benefits**:
- Frontend deployed to preferred platform
- Convex handles real-time backend complexity
- Type safety from database to UI

### Pattern 3: Multi-Cloud (Vercel Edge + Railway + Convex)

**Use Case**: Complex global applications

**Architecture**:
```
Vercel Edge Functions (global routing)
  → Railway (heavy compute)
  → Convex (real-time features)
```

**Benefits**:
- Best-in-class platforms for each need
- Optimized for performance and cost
- Redundancy and disaster recovery

---

## Performance Benchmarks

### Cold Start Times

| Platform | Cold Start | Warm Request |
|----------|-----------|--------------|
| Vercel Edge | <50ms | ~10ms |
| Vercel Serverless | ~200ms | ~50ms |
| Railway (container) | No cold start | ~20ms |
| Convex | ~100ms | ~30ms |

### Global Latency

| Platform | US-East | US-West | Europe | Asia |
|----------|---------|---------|--------|------|
| Vercel Edge | 10ms | 30ms | 40ms | 80ms |
| Railway | 20ms | 40ms | 60ms | 150ms |
| Convex | 30ms | 60ms | 90ms | 200ms |

---

## Choosing the Right Platform

### Choose Vercel If:
- Using Next.js framework
- Need global CDN distribution
- Preview deployments important
- Team collaboration features needed
- Edge compute requirements
- Simple deployment workflow
- Web Vitals monitoring important

### Choose Railway If:
- Need Docker support
- Multi-service architecture
- Custom runtimes required
- Private networking needed
- Persistent volume storage
- WebSocket connections
- Full control over environment
- Migrating existing containerized apps

### Choose Convex If:
- Real-time collaboration features
- TypeScript-first development
- Reactive UI updates
- Optimistic UI patterns
- Want to avoid API boilerplate
- Type safety critical
- Quick prototyping important
- Built-in database sufficient

---

## Final Recommendations

**For New Projects**:
- Web apps: Vercel
- Real-time: Convex
- Microservices: Railway
- Next.js: Vercel

**For Existing Projects**:
- Next.js on custom hosting → Migrate to Vercel
- Dockerized apps → Migrate to Railway
- Need real-time → Add Convex backend

**For Teams**:
- Small team (<5 developers): Vercel (easiest)
- Full-stack team (5-20): Railway (flexible)
- Frontend-focused: Vercel
- Backend-focused: Railway

**For Budget**:
- Free tier available: Vercel or Convex
- Predictable costs: Railway
- Usage-based: Vercel or Convex

---

**Status**: Production Ready
**Version**: 2.0.0
**Updated**: 2026-02-09
