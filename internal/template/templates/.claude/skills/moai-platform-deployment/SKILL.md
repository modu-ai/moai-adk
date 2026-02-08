---
name: moai-platform-deployment
description: >
  Deployment and hosting platform specialist covering Vercel, Railway, and Convex.
  Use when deploying applications, configuring edge functions, setting up continuous deployment,
  managing serverless infrastructure, containerized deployments, real-time backends, or choosing
  deployment platforms. Covers edge computing (Vercel), container orchestration (Railway), and
  reactive backends (Convex).
license: MIT
metadata:
  version: "2.0.0"
  category: "platform"
  status: "active"
  updated: "2026-02-09"
  platforms: "Vercel, Railway, Convex"
  tags: "deployment, hosting, vercel, railway, convex, edge, containers, serverless, real-time"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 4500

# MoAI Extension: Triggers
triggers:
  keywords: ["deploy", "deployment", "hosting", "vercel", "railway", "convex", "edge functions", "containers", "docker", "serverless", "real-time", "preview deployment", "continuous deployment"]
  agents: ["expert-devops", "expert-backend", "expert-frontend"]
  phases: ["run", "sync"]

user-invocable: false
---

# Deployment Platform Specialist

Comprehensive deployment platform guide covering Vercel (edge-first), Railway (container-first), and Convex (real-time backend).

---

## Quick Platform Selection

### When to Use Each Platform

**Vercel** - Edge-First Deployment:
- Next.js applications with SSR/SSG
- Global CDN distribution required
- Sub-50ms edge latency critical
- Preview deployments for team collaboration
- Managed storage needs (KV, Blob, Postgres)

**Railway** - Container-First Deployment:
- Full-stack containerized applications
- Custom runtime environments
- Multi-service architectures
- Persistent volume storage
- WebSocket/gRPC long-lived connections

**Convex** - Real-Time Backend:
- Collaborative real-time applications
- Reactive data synchronization
- TypeScript-first backend needs
- Optimistic UI updates
- Document-oriented data models

---

## Decision Guide

### By Application Type

**Web Applications (Frontend + API)**:
- Next.js → Vercel (optimal integration)
- React/Vue with custom API → Railway (flexible)
- Real-time collaborative → Convex + Vercel

**Mobile Backends**:
- REST/GraphQL → Railway (stable connections)
- Real-time sync → Convex (reactive queries)
- Edge API → Vercel (global latency)

**Full-Stack Monoliths**:
- Containerized → Railway (Docker support)
- Serverless → Vercel (Next.js API routes)
- Real-time → Convex (built-in reactivity)

### By Infrastructure Needs

**Compute Requirements**:
- Edge compute → Vercel (30+ edge locations)
- Custom runtimes → Railway (Docker flexibility)
- Serverless TypeScript → Convex (managed runtime)

**Storage Requirements**:
- Redis/KV → Vercel KV or Railway
- PostgreSQL → Vercel Postgres or Railway
- File storage → Vercel Blob or Railway volumes
- Document DB → Convex (built-in)

**Networking Requirements**:
- CDN distribution → Vercel (built-in)
- Private networking → Railway (service mesh)
- Real-time WebSocket → Convex (built-in) or Railway

---

## Common Deployment Patterns

### Pattern 1: Next.js with Database

**Stack**: Vercel + Vercel Postgres/KV

**Setup**:
1. Deploy Next.js app to Vercel
2. Provision Vercel Postgres for database
3. Use Vercel KV for session/cache
4. Configure environment variables
5. Enable ISR for dynamic content

**Best For**: Web apps with standard database needs, e-commerce, content sites

### Pattern 2: Containerized Multi-Service

**Stack**: Railway + Docker

**Setup**:
1. Create multi-stage Dockerfile
2. Configure railway.toml for services
3. Set up private networking
4. Configure persistent volumes
5. Enable auto-scaling

**Best For**: Microservices, complex backends, custom tech stacks

### Pattern 3: Real-Time Collaborative App

**Stack**: Convex + Vercel/Railway (frontend)

**Setup**:
1. Initialize Convex backend
2. Define schema and server functions
3. Deploy frontend to Vercel/Railway
4. Configure Convex provider
5. Implement optimistic updates

**Best For**: Collaborative tools, live dashboards, chat applications

### Pattern 4: Hybrid Edge + Container

**Stack**: Vercel (frontend/edge) + Railway (backend services)

**Setup**:
1. Deploy Next.js frontend to Vercel
2. Deploy backend services to Railway
3. Configure CORS and API endpoints
4. Set up edge middleware for routing
5. Use private networking for Railway

**Best For**: High-performance apps, global distribution with complex backends

### Pattern 5: Serverless Full-Stack

**Stack**: Vercel (frontend + API routes) + Convex (backend)

**Setup**:
1. Build Next.js app with API routes
2. Initialize Convex for data layer
3. Configure authentication (Clerk/Auth0)
4. Deploy frontend to Vercel
5. Connect Convex client

**Best For**: Rapid prototyping, startups, real-time web apps

---

## Essential Configuration

### Vercel Quick Start

**vercel.json**:
```json
{
  "$schema": "https://openapi.vercel.sh/vercel.json",
  "framework": "nextjs",
  "regions": ["iad1", "sfo1", "fra1"],
  "functions": {
    "app/api/**/*.ts": {
      "memory": 1024,
      "maxDuration": 10
    }
  }
}
```

**Edge Function**:
```typescript
export const runtime = "edge"
export const preferredRegion = ["iad1", "sfo1"]

export async function GET(request: Request) {
  const country = request.geo?.country || "Unknown"
  return Response.json({ country })
}
```

### Railway Quick Start

**railway.toml**:
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "Dockerfile"

[deploy]
healthcheckPath = "/health"
healthcheckTimeout = 100
restartPolicyType = "ON_FAILURE"
numReplicas = 2

[deploy.resources]
memory = "2GB"
cpu = "2.0"
```

**Multi-Stage Dockerfile**:
```dockerfile
# Builder stage
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Runner stage
FROM node:20-alpine
WORKDIR /app
ENV NODE_ENV=production
RUN addgroup -g 1001 -S nodejs && adduser -S appuser -u 1001
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/dist ./dist
USER appuser
EXPOSE 3000
CMD ["node", "dist/main.js"]
```

### Convex Quick Start

**convex/schema.ts**:
```typescript
import { defineSchema, defineTable } from "convex/server"
import { v } from "convex/values"

export default defineSchema({
  messages: defineTable({
    text: v.string(),
    userId: v.id("users"),
    timestamp: v.number(),
  })
    .index("by_timestamp", ["timestamp"])
    .searchIndex("search_text", {
      searchField: "text",
      filterFields: ["userId"],
    }),
})
```

**React Integration**:
```typescript
import { useQuery, useMutation } from "convex/react"
import { api } from "../convex/_generated/api"

export function Messages() {
  const messages = useQuery(api.messages.list)
  const sendMessage = useMutation(api.messages.send)

  if (!messages) return <div>Loading...</div>

  return (
    <div>
      {messages.map((msg) => (
        <div key={msg._id}>{msg.text}</div>
      ))}
    </div>
  )
}
```

---

## CI/CD Integration

### GitHub Actions - Vercel

```yaml
name: Deploy to Vercel
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: amondnet/vercel-action@v25
        with:
          vercel-token: ${{ secrets.VERCEL_TOKEN }}
          vercel-org-id: ${{ secrets.ORG_ID }}
          vercel-project-id: ${{ secrets.PROJECT_ID }}
```

### GitHub Actions - Railway

```yaml
name: Deploy to Railway
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: npm install -g @railway/cli
      - run: railway up --detach
        env:
          RAILWAY_TOKEN: ${{ secrets.RAILWAY_TOKEN }}
```

### GitHub Actions - Convex

```yaml
name: Deploy to Convex
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - run: npm ci
      - run: npx convex deploy
        env:
          CONVEX_DEPLOY_KEY: ${{ secrets.CONVEX_DEPLOY_KEY }}
```

---

## Advanced Patterns

### Blue-Green Deployment (Vercel)

Deploy new version, test on preview URL, then switch production alias using Vercel SDK for zero-downtime releases.

### Multi-Region (Railway)

Configure deployment regions in railway.toml:
```toml
[deploy.regions]
name = "us-west"
replicas = 2

[[deploy.regions]]
name = "eu-central"
replicas = 1
```

### Optimistic Updates (Convex)

```typescript
const sendMessage = useMutation(api.messages.send)

const handleSend = (text: string) => {
  sendMessage({ text })
    .then(() => console.log("Sent"))
    .catch(() => console.log("Failed, rolled back"))
}
```

---

## Platform-Specific Details

For detailed platform-specific patterns, configuration options, and advanced use cases, see:

- **reference/vercel.md** - Edge Functions, ISR, Analytics, Storage
- **reference/railway.md** - Docker, Multi-Service, Volumes, Scaling
- **reference/convex.md** - Reactive Queries, Server Functions, File Storage
- **reference/comparison.md** - Feature Matrix, Pricing, Migration Guides

---

## Works Well With

- moai-domain-backend for backend architecture patterns
- moai-domain-frontend for frontend integration
- moai-lang-typescript for TypeScript best practices
- moai-lang-python for Python deployment (Railway)
- moai-platform-auth for authentication integration
- moai-platform-database for database patterns

---

Status: Production Ready
Version: 2.0.0
Updated: 2026-02-09
Platforms: Vercel, Railway, Convex
