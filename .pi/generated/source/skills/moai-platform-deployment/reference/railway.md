# Railway Platform Reference

Comprehensive guide to Railway container deployment platform.

---

## Docker Deployment

### Multi-Stage Dockerfile Patterns

**Node.js with TypeScript**:
```dockerfile
# Builder stage
FROM node:20-alpine AS builder
WORKDIR /app

# Install dependencies
COPY package*.json ./
RUN npm ci --only=production && \
    npm cache clean --force

# Build application
COPY . .
RUN npm run build

# Runner stage
FROM node:20-alpine
WORKDIR /app

# Security: Create non-root user
RUN addgroup -g 1001 -S nodejs && \
    adduser -S appuser -u 1001

# Environment
ENV NODE_ENV=production

# Copy artifacts
COPY --from=builder --chown=appuser:nodejs /app/node_modules ./node_modules
COPY --from=builder --chown=appuser:nodejs /app/dist ./dist
COPY --chown=appuser:nodejs package*.json ./

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD node -e "require('http').get('http://localhost:3000/health', (r) => process.exit(r.statusCode === 200 ? 0 : 1))"

# Switch to non-root user
USER appuser

EXPOSE 3000
CMD ["node", "dist/main.js"]
```

**Python FastAPI**:
```dockerfile
# Builder stage
FROM python:3.11-slim AS builder
WORKDIR /app

# Install dependencies
RUN pip install --no-cache-dir poetry
COPY pyproject.toml poetry.lock ./
RUN poetry config virtualenvs.in-project true && \
    poetry install --no-dev --no-interaction

# Runner stage
FROM python:3.11-slim
WORKDIR /app

# Security: Create non-root user
RUN useradd -m -u 1001 appuser

# Copy virtual environment
COPY --from=builder --chown=appuser:appuser /app/.venv ./.venv
COPY --chown=appuser:appuser . .

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD python -c "import requests; requests.get('http://localhost:8000/health')"

USER appuser
EXPOSE 8000

CMD ["./.venv/bin/uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

**Go Service**:
```dockerfile
# Builder stage
FROM golang:1.21-alpine AS builder
WORKDIR /app

# Dependencies
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Runner stage
FROM alpine:latest
WORKDIR /app

# Security: Install CA certificates and create non-root user
RUN apk --no-cache add ca-certificates && \
    addgroup -g 1001 appuser && \
    adduser -S -u 1001 -G appuser appuser

# Copy binary
COPY --from=builder --chown=appuser:appuser /app/main .

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

USER appuser
EXPOSE 8080

CMD ["./main"]
```

### Build Optimization

**Layer Caching Strategy**:
1. Copy dependency files first (package.json, go.mod, etc.)
2. Install dependencies (cached if files unchanged)
3. Copy source code
4. Build application

**Image Size Reduction**:
- Use Alpine base images
- Multi-stage builds (separate builder and runner)
- .dockerignore to exclude unnecessary files
- Clean package manager caches
- Use specific dependency versions

---

## Multi-Service Architecture

### railway.toml Configuration

**Complete Configuration**:
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "Dockerfile"

[deploy]
healthcheckPath = "/health"
healthcheckTimeout = 100
restartPolicyType = "ON_FAILURE"
numReplicas = 2
startCommand = "node dist/main.js"

[deploy.resources]
memory = "2GB"
cpu = "2.0"

[[deploy.regions]]
name = "us-west"
replicas = 2

[[deploy.regions]]
name = "eu-central"
replicas = 1

[deploy.scaling]
minReplicas = 1
maxReplicas = 10
targetCPUUtilization = 80

[[volumes]]
mountPath = "/app/data"
name = "app-data"
size = "10GB"
```

### Service Communication

**Private Networking**:
```typescript
// Helper function for internal service URLs
function getInternalUrl(serviceName: string, port = 3000): string {
  const privateHost = process.env[`${serviceName.toUpperCase()}_RAILWAY_PRIVATE_DOMAIN`]
  return privateHost ? `http://${privateHost}:${port}` : `http://localhost:${port}`
}

// Usage
const apiUrl = getInternalUrl("api")
const authUrl = getInternalUrl("auth", 4000)
```

**Variable References**:
```yaml
# Backend service
DATABASE_URL: ${{Postgres.DATABASE_URL}}
REDIS_URL: ${{Redis.REDIS_URL}}

# Frontend service
API_URL: ${{backend.RAILWAY_PRIVATE_DOMAIN}}
```

### Monorepo Deployment

**Turborepo with Railway**:
```toml
[build]
builder = "NIXPACKS"
buildCommand = "turbo run build --filter=api"

[deploy]
startCommand = "turbo run start --filter=api"
```

---

## Volumes and Storage

### Persistent Volumes

**Configuration**:
```toml
[[volumes]]
mountPath = "/app/data"
name = "app-data"
size = "10GB"
```

**Usage in Application**:
```typescript
import fs from "fs/promises"
import path from "path"

const dataDir = process.env.RAILWAY_VOLUME_MOUNT_PATH || "./data"

// Write file
await fs.writeFile(
  path.join(dataDir, "file.txt"),
  "content",
  "utf-8"
)

// Read file
const content = await fs.readFile(
  path.join(dataDir, "file.txt"),
  "utf-8"
)
```

### SQLite on Volumes

**Configuration**:
```typescript
import Database from "better-sqlite3"
import path from "path"

const volumePath = process.env.RAILWAY_VOLUME_MOUNT_PATH || "./data"
const dbPath = path.join(volumePath, "app.db")

const db = new Database(dbPath)

// Optimize for SSD volumes
db.pragma("journal_mode = WAL")
db.pragma("synchronous = NORMAL")
```

### Backup Patterns

**Automated Backups**:
```typescript
import { execSync } from "child_process"
import path from "path"

async function backupVolume() {
  const volumePath = process.env.RAILWAY_VOLUME_MOUNT_PATH
  const backupPath = path.join(volumePath, "backups")
  const timestamp = new Date().toISOString()

  // Create backup directory
  await fs.mkdir(backupPath, { recursive: true })

  // Backup database
  execSync(
    `sqlite3 ${volumePath}/app.db ".backup '${backupPath}/backup-${timestamp}.db'"`
  )
}

// Schedule backups
import cron from "node-cron"
cron.schedule("0 2 * * *", backupVolume) // Daily at 2 AM
```

---

## Networking and Domains

### Custom Domains with SSL

**Setup via CLI**:
```bash
# Add custom domain
railway domain add example.com

# List domains
railway domain list

# Remove domain
railway domain remove example.com
```

**DNS Configuration**:
```
# CNAME record
example.com → <your-service>.railway.app

# Or A record (get IP from Railway dashboard)
example.com → <railway-ip-address>
```

### WebSocket Support

**Express WebSocket Server**:
```typescript
import express from "express"
import { WebSocketServer } from "ws"

const app = express()
const server = app.listen(process.env.PORT || 3000)

const wss = new WebSocketServer({ server })

wss.on("connection", (ws) => {
  console.log("Client connected")

  ws.on("message", (data) => {
    console.log("Received:", data)
    ws.send(`Echo: ${data}`)
  })

  ws.on("close", () => {
    console.log("Client disconnected")
  })
})
```

### Auto-Scaling Configuration

**CPU-Based Scaling**:
```toml
[deploy.scaling]
minReplicas = 1
maxReplicas = 10
targetCPUUtilization = 80
```

**Memory-Based Scaling**:
```toml
[deploy.scaling]
minReplicas = 2
maxReplicas = 20
targetMemoryUtilization = 75
```

---

## CLI Commands

### Project Management

```bash
# Login
railway login

# Initialize project
railway init

# Link to existing project
railway link

# List services
railway service list

# Switch service
railway service switch <service-name>
```

### Deployment

```bash
# Deploy current directory
railway up

# Deploy without waiting for logs
railway up --detach

# Deploy specific service
railway up --service api

# Rollback to previous deployment
railway rollback --previous
```

### Environment Variables

```bash
# Set variable
railway variables --set API_KEY=secret

# List variables
railway variables

# Delete variable
railway variables --delete API_KEY
```

### Logs and Status

```bash
# View logs
railway logs

# Follow logs
railway logs --follow

# View logs for specific service
railway logs --service api

# View deployment status
railway status
```

---

## CI/CD Integration

### GitHub Actions

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

      - name: Install Railway CLI
        run: npm install -g @railway/cli

      - name: Deploy to Railway
        run: railway up --detach
        env:
          RAILWAY_TOKEN: ${{ secrets.RAILWAY_TOKEN }}
```

### GitLab CI

```yaml
deploy:
  stage: deploy
  image: node:20
  script:
    - npm install -g @railway/cli
    - railway up --detach
  environment:
    name: production
  variables:
    RAILWAY_TOKEN: $CI_RAILWAY_TOKEN
  only:
    - main
```

---

## Best Practices

### Dockerfile Optimization

1. **Use multi-stage builds** to reduce final image size
2. **Order layers by change frequency** (dependencies before source code)
3. **Use .dockerignore** to exclude unnecessary files
4. **Pin specific versions** for reproducible builds
5. **Run as non-root user** for security

### Resource Allocation

**Memory Guidelines**:
- Lightweight API: 512MB - 1GB
- Standard web app: 1GB - 2GB
- Data-intensive: 2GB - 4GB
- ML inference: 4GB+

**CPU Guidelines**:
- Background jobs: 0.5 - 1.0 CPU
- API servers: 1.0 - 2.0 CPU
- Compute-heavy: 2.0+ CPU

### Health Checks

Always implement health checks:
```typescript
app.get("/health", (req, res) => {
  // Check database connection
  // Check external services
  // Return 200 if healthy, 503 if unhealthy
  res.status(200).json({ status: "healthy" })
})
```

---

## Troubleshooting

### Common Issues

**Build Failures**:
- Check Dockerfile syntax
- Verify base image availability
- Review build logs in Railway dashboard
- Test build locally: `docker build .`

**Deployment Crashes**:
- Check application logs: `railway logs`
- Verify health check endpoint
- Review resource allocation (memory/CPU)
- Check environment variables

**Connection Issues**:
- Verify private networking configuration
- Check service variable references
- Test with Railway's internal DNS

---

## Context7 Integration

For latest Railway documentation:

```typescript
// Step 1: Resolve library
const libraries = await mcp__context7__resolve_library_id({
  libraryName: "railway",
  query: "docker deployment multi-service"
})

// Step 2: Get documentation
const docs = await mcp__context7__get_library_docs({
  libraryId: libraries[0].id,
  topic: "docker-deployment",
  maxTokens: 8000
})
```

---

**Status**: Production Ready
**Platform**: Railway Container Deployment
**Updated**: 2026-02-09
