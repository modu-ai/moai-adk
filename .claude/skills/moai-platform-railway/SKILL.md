---
name: "moai-platform-railway"
description: "Railway container deployment specialist covering Docker, multi-service architectures, persistent volumes, and auto-scaling. Use when deploying containerized full-stack applications, configuring multi-region deployments, or setting up persistent storage."
version: 2.0.0
category: "platform"
modularized: true
user-invocable: false
tags: ['railway', 'docker', 'containers', 'multi-service', 'auto-scaling']
context7-libraries: ['/railwayapp/railway']
related-skills: ['moai-platform-vercel', 'moai-domain-backend']
updated: 2026-01-08
status: "active"
allowed-tools:
  - Read
  - Write
  - Bash
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
---

# moai-platform-railway: Container Deployment Specialist

## Quick Reference

Railway Platform Core: Container-first deployment platform with Docker and Railpack builds, multi-service architectures, persistent volumes, private networking, and auto-scaling capabilities.

### Railway Optimal Use Cases

Container Workloads:
- Full-stack containerized applications with custom runtimes
- Multi-service architectures with inter-service communication
- Backend services requiring persistent connections (WebSocket, gRPC)
- Database-backed applications with managed PostgreSQL, MySQL, Redis

Infrastructure Requirements:
- Persistent volume storage for stateful workloads
- Private networking for secure service mesh
- Multi-region deployment for global availability
- Auto-scaling based on CPU, memory, or request metrics

### Build Strategy Selection

Docker Build: Custom system dependencies, multi-stage builds, specific base images
Railpack Build: Standard runtimes (Node.js, Python, Go), zero-config, faster builds

Note: Nixpacks is deprecated. New services default to Railpack.

### Key CLI Commands

```bash
railway login && railway init && railway link
railway up                    # Deploy current directory
railway up --detach          # Deploy without logs
railway variables --set KEY=value
railway logs --service api
railway rollback --previous
```

---

## Module Index

### Docker Deployment (modules/docker-deployment.md)
Multi-stage Dockerfiles for Node.js, Python, Go, and Rust. Build optimization, image size reduction, health checks, and Railpack migration.

### Multi-Service Architecture (modules/multi-service.md)
Monorepo deployments, service communication, variable references, private networking, and message queue patterns.

### Volumes and Storage (modules/volumes-storage.md)
Persistent volume configuration, file storage service, SQLite on volumes, backup/restore patterns.

### Networking and Domains (modules/networking-domains.md)
Private networking, custom domains with SSL, multi-region deployment, auto-scaling, WebSocket support.

---

## Implementation Guide

### Phase 1: Project Setup

```bash
railway login && railway init && railway link
```

railway.toml:
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "./Dockerfile"

[deploy]
healthcheckPath = "/health"
healthcheckTimeout = 60
restartPolicyType = "ON_FAILURE"
numReplicas = 2

[deploy.resources]
memory = "512Mi"
cpu = "0.5"
```

### Phase 2: Container Configuration

Quick Start Dockerfile (Node.js):
```dockerfile
FROM node:20-alpine AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:20-alpine AS runner
WORKDIR /app
ENV NODE_ENV=production
RUN addgroup --system --gid 1001 nodejs && adduser --system --uid 1001 appuser
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/dist ./dist
USER appuser
EXPOSE 3000
CMD ["node", "dist/main.js"]
```

For detailed patterns, see modules/docker-deployment.md.

### Phase 3: Multi-Service Setup

Service Variable References:
```
${{Postgres.DATABASE_URL}}
${{Redis.REDIS_URL}}
${{api.RAILWAY_PRIVATE_DOMAIN}}
```

Private Networking:
```typescript
const getInternalUrl = (service: string, port = 3000): string => {
  const domain = process.env[`${service.toUpperCase()}_RAILWAY_PRIVATE_DOMAIN`]
  return domain ? `http://${domain}:${port}` : `http://localhost:${port}`
}
```

For detailed patterns, see modules/multi-service.md.

### Phase 4: Storage and Scaling

Volume Configuration:
```toml
[[volumes]]
mountPath = "/app/data"
name = "app-data"
size = "10Gi"
```

Auto-Scaling:
```toml
[deploy.scaling]
minReplicas = 2
maxReplicas = 10
targetCPUUtilization = 70
```

Multi-Region:
```toml
[[deploy.regions]]
name = "us-west1"
replicas = 3

[[deploy.regions]]
name = "europe-west4"
replicas = 2
```

---

## CI/CD Integration

GitHub Actions:
```yaml
name: Railway Deploy
on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - run: npm i -g @railway/cli
      - run: railway up --detach
        env:
          RAILWAY_TOKEN: ${{ secrets.RAILWAY_TOKEN }}
```

---

## Context7 Documentation Access

Step 1: Use mcp__context7__resolve-library-id with "railway"
Step 2: Use mcp__context7__get-library-docs with resolved ID and topic

---

## Works Well With

- moai-platform-vercel: Edge deployment for frontend
- moai-domain-backend: Backend architecture patterns
- moai-lang-python: Python FastAPI deployment
- moai-lang-typescript: TypeScript Node.js patterns
- moai-lang-go: Go service deployment

---

## Additional Resources

- reference.md: Extended documentation and configuration
- examples.md: Working code examples for common scenarios

---

Status: Production Ready | Version: 2.0.0 | Updated: 2025-12-30
