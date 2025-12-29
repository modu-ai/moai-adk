---
name: moai-platform-railway
description: Railway container deployment specialist covering Docker, multi-service architectures, persistent volumes, and auto-scaling. Use when deploying containerized full-stack applications.
version: 1.1.0
category: platform
tags: [railway, docker, containers, multi-service, auto-scaling]
context7-libraries: [railway]
related-skills: [moai-platform-vercel, moai-domain-backend]
updated: 2025-12-30
status: active
allowed-tools: Read, Write, Bash, Grep, Glob
---

# moai-platform-railway: Container Deployment Specialist

## Quick Reference (30 seconds)

Railway Platform Core: Container-first deployment platform with Docker and Nixpacks builds, multi-service architectures, persistent volumes, private networking, and auto-scaling capabilities.

### Railway Optimal Use Cases

Container Workloads:
- Full-stack containerized applications with custom runtimes
- Multi-service architectures with inter-service communication
- Backend services requiring persistent connections (WebSocket, gRPC)
- Custom runtime requirements (Python, Go, Rust, Elixir)
- Database-backed applications with managed PostgreSQL, MySQL, Redis

Infrastructure Requirements:
- Persistent volume storage for stateful workloads
- Private networking for secure service mesh
- Multi-region deployment for global availability
- Auto-scaling based on CPU, memory, or request metrics

### Build Strategy Selection

Docker Build: Custom system dependencies, multi-stage builds, specific base images
Nixpacks Build: Standard runtimes (Node.js, Python, Go), zero-config, faster builds

### Key CLI Commands

```bash
railway login && railway init && railway link
railway up                    # Deploy current directory
railway up --detach          # Deploy without logs
railway up --service api     # Deploy specific service
railway variables --set KEY=value
railway logs --service api
```

---

## Implementation Guide

### Phase 1: Docker Deployment Patterns

Multi-Stage Node.js Dockerfile:
```dockerfile
FROM node:20-alpine AS deps
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production

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
COPY --from=deps /app/node_modules ./node_modules
COPY --from=builder /app/dist ./dist
USER appuser
EXPOSE 3000
CMD ["node", "dist/main.js"]
```

Python Production Dockerfile:
```dockerfile
FROM python:3.12-slim AS builder
WORKDIR /app
RUN pip install --no-cache-dir poetry
COPY pyproject.toml poetry.lock ./
RUN poetry config virtualenvs.create false && poetry install --no-dev

FROM python:3.12-slim AS runner
WORKDIR /app
ENV PYTHONUNBUFFERED=1
COPY --from=builder /usr/local/lib/python3.12/site-packages /usr/local/lib/python3.12/site-packages
COPY . .
RUN useradd --create-home appuser
USER appuser
EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

Go Production Dockerfile:
```dockerfile
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o main .

FROM alpine:latest AS runner
RUN apk --no-cache add ca-certificates && adduser -D appuser
WORKDIR /app
COPY --from=builder /app/main .
USER appuser
EXPOSE 8080
CMD ["./main"]
```

### Phase 2: Nixpacks Configuration

Note: Nixpacks is deprecated and in maintenance mode. New services default to Railpack. Existing services continue to work with Nixpacks. To migrate, set builder = "RAILPACK" in railway.toml.

Node.js nixpacks.toml:
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

Python nixpacks.toml:
```toml
[phases.setup]
nixPkgs = ["python312", "poetry"]
aptPkgs = ["libpq-dev"]

[phases.install]
cmds = ["poetry install --no-dev"]

[start]
cmd = "poetry run gunicorn app:application --bind 0.0.0.0:$PORT"
```

### Phase 3: Railway Configuration

railway.toml with Nixpacks:
```toml
[build]
builder = "NIXPACKS"
buildCommand = "npm run build"
watchPatterns = ["src/**", "package.json"]

[deploy]
startCommand = "npm start"
healthcheckPath = "/health"
healthcheckTimeout = 100
restartPolicyType = "ON_FAILURE"
restartPolicyMaxRetries = 5
numReplicas = 2

[deploy.resources]
memory = "512Mi"
cpu = "0.5"
```

railway.toml with Docker:
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "./Dockerfile"

[deploy]
healthcheckPath = "/api/health"
healthcheckTimeout = 60
restartPolicyType = "ALWAYS"
numReplicas = 3

[deploy.resources]
memory = "1Gi"
cpu = "1"
```

### Phase 4: Multi-Service Architecture

Multi-Service Monorepo (railway.toml per service):

Each service in a monorepo requires its own railway.toml in its directory.

API Service (apps/api/railway.toml):
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "./Dockerfile"

[deploy]
startCommand = "node dist/main.js"
healthcheckPath = "/health"
numReplicas = 3
```

Worker Service (apps/worker/railway.toml):
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "./Dockerfile"

[deploy]
startCommand = "node dist/worker.js"
numReplicas = 2
```

Scheduler Service (apps/scheduler/railway.toml):
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "./Dockerfile"

[deploy]
startCommand = "node dist/scheduler.js"
cronSchedule = "*/5 * * * *"
```

Service Variable References:
Use Railway's variable reference syntax in service settings for cross-service communication with format ${{ServiceName.VARIABLE_NAME}}.

Private Networking:
```typescript
const getInternalUrl = (service: string, port = 3000): string => {
  const domain = process.env[`${service.toUpperCase()}_RAILWAY_PRIVATE_DOMAIN`]
  return domain ? `http://${domain}:${port}` : `http://localhost:${port}`
}

app.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    service: process.env.RAILWAY_SERVICE_NAME,
    replica: process.env.RAILWAY_REPLICA_ID
  })
})
```

### Phase 5: Persistent Volumes

Volume Configuration:
```toml
[deploy]
startCommand = "npm start"

[[volumes]]
mountPath = "/app/data"
name = "app-data"
size = "10Gi"

[[volumes]]
mountPath = "/app/uploads"
name = "user-uploads"
size = "50Gi"
```

Persistent Storage Pattern:
```typescript
import { join } from 'path'
import { existsSync, mkdirSync, writeFileSync, readFileSync } from 'fs'

const VOLUME_PATH = process.env.RAILWAY_VOLUME_MOUNT_PATH || '/app/data'

class PersistentStorage {
  constructor() {
    if (!existsSync(VOLUME_PATH)) mkdirSync(VOLUME_PATH, { recursive: true })
  }
  write(file: string, data: Buffer | string) { writeFileSync(join(VOLUME_PATH, file), data) }
  read(file: string): Buffer { return readFileSync(join(VOLUME_PATH, file)) }
}
```

### Phase 6: Auto-Scaling

Resource-Based Scaling:
```toml
[deploy.scaling]
minReplicas = 2
maxReplicas = 10
targetCPUUtilization = 70
targetMemoryUtilization = 80
```

Request-Based Scaling:
```toml
[deploy.scaling]
minReplicas = 1
maxReplicas = 20
targetRequestsPerSecond = 100
scaleDownDelaySeconds = 300
```

Application Metrics:
```typescript
import { register, Counter, Histogram } from 'prom-client'

const httpRequests = new Counter({
  name: 'http_requests_total',
  help: 'Total HTTP requests',
  labelNames: ['method', 'status']
})

app.get('/metrics', async (req, res) => {
  res.set('Content-Type', register.contentType)
  res.end(await register.metrics())
})
```

---

## Multi-Region Deployment

Available Regions: us-west1 (Oregon), us-east4 (Virginia), europe-west4 (Netherlands), asia-southeast1 (Singapore)

```bash
railway up --region us-west1
railway up --region europe-west4
```

Region Configuration:
```toml
[[deploy.regions]]
name = "us-west1"
replicas = 3

[[deploy.regions]]
name = "europe-west4"
replicas = 2
```

Database Read Replica:
```typescript
const primaryPool = new Pool({ connectionString: process.env.DATABASE_URL })
const replicaPool = new Pool({ connectionString: process.env.DATABASE_REPLICA_URL })

async function query(sql: string, params?: any[]) {
  const isRead = sql.trim().toLowerCase().startsWith('select')
  return (isRead ? replicaPool : primaryPool).query(sql, params)
}
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

Rollback Commands:
```bash
railway deployments list
railway rollback <deployment-id>
railway rollback --previous
```

---

## Works Well With

- `moai-platform-vercel` - Edge deployment for frontend applications
- `moai-domain-backend` - Backend service architecture patterns
- `moai-lang-python` - Python FastAPI deployment configurations
- `moai-lang-typescript` - TypeScript Node.js deployment patterns

---

Status: Production Ready | Version: 1.1.0 | Updated: 2025-12-30
