# Docker Deployment Module

Purpose: Container build patterns with multi-stage Dockerfiles for Node.js, Python, Go, and Rust applications on Railway.

---

## Build Strategy Selection

Docker Build Use Cases:
- Custom system dependencies required
- Multi-stage builds for optimization
- Specific base images needed
- Binary compilation (Go, Rust)
- Complex build processes

Railpack Build Use Cases:
- Standard runtimes without customization
- Zero-config deployments
- Simple web applications
- Quick iteration during development

---

## Multi-Stage Build Patterns

### Node.js Production Dockerfile

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

Key Optimization Points:
- Separate dependency and build stages for caching
- Production-only dependencies in final image
- Non-root user for security
- Alpine base for minimal image size

### Python Production Dockerfile

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

Poetry Alternatives:
- pip with requirements.txt: pip install -r requirements.txt
- uv package manager: pip install uv && uv pip install -r requirements.txt
- pdm: pip install pdm && pdm install --prod

### Go Production Dockerfile

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

Build Flag Explanations:
- CGO_ENABLED=0: Static binary without C dependencies
- GOOS=linux: Target Linux regardless of build host
- ldflags="-w -s": Strip debug info for smaller binary

### Rust Production Dockerfile

```dockerfile
FROM rust:1.77-slim AS builder
WORKDIR /app
RUN apt-get update && apt-get install -y pkg-config libssl-dev
COPY Cargo.toml Cargo.lock ./
RUN mkdir src && echo "fn main() {}" > src/main.rs
RUN cargo build --release
RUN rm -rf src
COPY . .
RUN touch src/main.rs && cargo build --release

FROM debian:bookworm-slim AS runner
RUN apt-get update && apt-get install -y ca-certificates && rm -rf /var/lib/apt/lists/*
RUN useradd --create-home appuser
WORKDIR /app
COPY --from=builder /app/target/release/app .
USER appuser
EXPOSE 8080
CMD ["./app"]
```

Cargo Build Optimization:
- Dummy main.rs trick caches dependencies
- Release build for production performance
- Minimal runtime image with only binary

---

## Railway Configuration for Docker

railway.toml with Docker Builder:
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "./Dockerfile"
watchPatterns = ["src/**", "Cargo.toml", "Cargo.lock"]

[deploy]
healthcheckPath = "/health"
healthcheckTimeout = 60
restartPolicyType = "ON_FAILURE"
restartPolicyMaxRetries = 5
numReplicas = 2

[deploy.resources]
memory = "512Mi"
cpu = "0.5"
```

Configuration Options:
- dockerfilePath: Custom Dockerfile location (default: ./Dockerfile)
- watchPatterns: Files triggering rebuild on change
- healthcheckPath: Endpoint for container health verification
- restartPolicyType: ON_FAILURE, ALWAYS, or NEVER

---

## Railpack Migration

Note: Nixpacks is deprecated and in maintenance mode. New projects should use Railpack.

Migrating from Nixpacks to Railpack:
```toml
[build]
builder = "RAILPACK"
buildCommand = "npm run build"

[deploy]
startCommand = "npm start"
```

Railpack Benefits:
- Faster build times
- Better caching
- Automatic detection of frameworks
- Continued maintenance and updates

---

## Image Size Optimization

Size Reduction Techniques:
- Use Alpine or slim base images
- Multi-stage builds to exclude build tools
- Clean package manager caches
- Remove unnecessary files in .dockerignore

.dockerignore Example:
```
node_modules
.git
.gitignore
README.md
Dockerfile
docker-compose*.yml
.env*
*.log
.coverage
__pycache__
*.pyc
target
```

Image Size Comparison (typical):
- Node.js (node:20): ~1GB
- Node.js (node:20-alpine): ~150MB
- Python (python:3.12): ~1GB
- Python (python:3.12-slim): ~150MB
- Go (scratch + binary): ~10-20MB
- Rust (debian-slim + binary): ~80-100MB

---

## Health Check Implementation

Application Health Endpoint:
```typescript
app.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    timestamp: new Date().toISOString(),
    service: process.env.RAILWAY_SERVICE_NAME,
    version: process.env.RAILWAY_GIT_COMMIT_SHA?.slice(0, 7)
  })
})
```

Database Connection Check:
```typescript
app.get('/health', async (req, res) => {
  try {
    await db.query('SELECT 1')
    res.json({ status: 'healthy', database: 'connected' })
  } catch (error) {
    res.status(503).json({ status: 'unhealthy', database: 'disconnected' })
  }
})
```

---

## Works Well With

- modules/multi-service.md: Multi-service architecture patterns
- modules/volumes-storage.md: Persistent volume configuration
- moai-lang-python: Python deployment patterns
- moai-lang-typescript: TypeScript deployment patterns
- moai-lang-go: Go deployment patterns
- moai-lang-rust: Rust deployment patterns

---

Version: 1.0.0 | Updated: 2025-12-30
