# Multi-Service Architecture Module

Purpose: Patterns for deploying multi-service applications with monorepo support, service communication, and orchestration on Railway.

---

## Monorepo Structure

Recommended Directory Layout:
```
project/
├── apps/
│   ├── api/
│   │   ├── Dockerfile
│   │   ├── railway.toml
│   │   └── src/
│   ├── worker/
│   │   ├── Dockerfile
│   │   ├── railway.toml
│   │   └── src/
│   ├── scheduler/
│   │   ├── Dockerfile
│   │   ├── railway.toml
│   │   └── src/
│   └── web/
│       ├── Dockerfile
│       ├── railway.toml
│       └── src/
├── packages/
│   ├── shared/
│   └── database/
├── package.json
└── turbo.json
```

Each service requires its own railway.toml in its directory for independent deployment configuration.

---

## Service Configuration Patterns

### API Service

apps/api/railway.toml:
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "./Dockerfile"

[deploy]
startCommand = "node dist/main.js"
healthcheckPath = "/health"
healthcheckTimeout = 60
numReplicas = 3

[deploy.resources]
memory = "1Gi"
cpu = "1"
```

### Worker Service

apps/worker/railway.toml:
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "./Dockerfile"

[deploy]
startCommand = "node dist/worker.js"
numReplicas = 2

[deploy.resources]
memory = "512Mi"
cpu = "0.5"
```

Worker services typically do not need health checks or public endpoints.

### Scheduler Service

apps/scheduler/railway.toml:
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "./Dockerfile"

[deploy]
startCommand = "node dist/scheduler.js"
cronSchedule = "*/5 * * * *"
numReplicas = 1
```

Cron Schedule Format: minute hour day month weekday
- */5 * * * *: Every 5 minutes
- 0 * * * *: Every hour
- 0 0 * * *: Daily at midnight
- 0 0 * * 0: Weekly on Sunday

---

## Service Communication

### Variable References

Railway provides variable reference syntax for cross-service communication:
- Format: ${{ServiceName.VARIABLE_NAME}}
- Example: ${{api.DATABASE_URL}}

Setting Variable References:
1. Go to service settings in Railway dashboard
2. Add environment variable
3. Use reference syntax for dynamic values

Common Reference Patterns:
- Database URL: ${{Postgres.DATABASE_URL}}
- Redis URL: ${{Redis.REDIS_URL}}
- API Internal URL: ${{api.RAILWAY_PRIVATE_DOMAIN}}

### Private Networking

Internal Communication Pattern:
```typescript
const getInternalUrl = (service: string, port = 3000): string => {
  const domain = process.env[`${service.toUpperCase()}_RAILWAY_PRIVATE_DOMAIN`]
  return domain ? `http://${domain}:${port}` : `http://localhost:${port}`
}

// Usage
const apiUrl = getInternalUrl('api', 3000)
const workerUrl = getInternalUrl('worker', 3001)
```

Service Discovery Environment Variables:
- RAILWAY_PRIVATE_DOMAIN: Internal hostname for this service
- {SERVICE}_RAILWAY_PRIVATE_DOMAIN: Internal hostname for other services

### Internal API Calls

```typescript
import axios from 'axios'

class ServiceClient {
  private baseUrl: string

  constructor(serviceName: string, port: number = 3000) {
    const domain = process.env[`${serviceName.toUpperCase()}_RAILWAY_PRIVATE_DOMAIN`]
    this.baseUrl = domain ? `http://${domain}:${port}` : `http://localhost:${port}`
  }

  async get<T>(path: string): Promise<T> {
    const response = await axios.get(`${this.baseUrl}${path}`)
    return response.data
  }

  async post<T>(path: string, data: any): Promise<T> {
    const response = await axios.post(`${this.baseUrl}${path}`, data)
    return response.data
  }
}

// Usage
const apiClient = new ServiceClient('api', 3000)
const users = await apiClient.get('/users')
```

---

## Message Queue Patterns

### Redis-Based Queue

Worker Implementation:
```typescript
import { Queue, Worker } from 'bullmq'
import Redis from 'ioredis'

const connection = new Redis(process.env.REDIS_URL)

const emailQueue = new Queue('emails', { connection })

// Producer (API service)
await emailQueue.add('welcome', { userId: 123, email: 'user@example.com' })

// Consumer (Worker service)
const worker = new Worker('emails', async (job) => {
  if (job.name === 'welcome') {
    await sendWelcomeEmail(job.data.email)
  }
}, { connection })
```

### Event-Based Communication

Pub/Sub Pattern:
```typescript
import Redis from 'ioredis'

const publisher = new Redis(process.env.REDIS_URL)
const subscriber = new Redis(process.env.REDIS_URL)

// Publisher (any service)
await publisher.publish('user:created', JSON.stringify({ userId: 123 }))

// Subscriber (worker service)
subscriber.subscribe('user:created')
subscriber.on('message', (channel, message) => {
  const data = JSON.parse(message)
  console.log(`User created: ${data.userId}`)
})
```

---

## Database Sharing

Shared Database Access:
```typescript
// packages/database/index.ts
import { Pool } from 'pg'

export const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
  max: 20,
  idleTimeoutMillis: 30000,
  connectionTimeoutMillis: 2000
})

export const query = (text: string, params?: any[]) => pool.query(text, params)
```

Service-Specific Migrations:
- API service owns schema migrations
- Worker and scheduler services use read-only access where possible
- Use database roles for access control

---

## Deployment Ordering

Service Dependency Order:
1. Database and cache services (PostgreSQL, Redis)
2. Shared infrastructure services
3. Core API services
4. Worker and background services
5. Frontend and edge services

Railway automatically handles dependency ordering when using variable references.

---

## Health Monitoring

Service Health Endpoint:
```typescript
app.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    service: process.env.RAILWAY_SERVICE_NAME,
    replica: process.env.RAILWAY_REPLICA_ID,
    timestamp: new Date().toISOString()
  })
})
```

Dependency Health Check:
```typescript
app.get('/health/ready', async (req, res) => {
  const checks = {
    database: await checkDatabase(),
    redis: await checkRedis(),
    api: await checkApiService()
  }

  const allHealthy = Object.values(checks).every(v => v)
  res.status(allHealthy ? 200 : 503).json({ status: allHealthy ? 'ready' : 'not_ready', checks })
})
```

---

## Works Well With

- modules/docker-deployment.md: Container build patterns
- modules/networking-domains.md: Private networking configuration
- modules/volumes-storage.md: Shared storage patterns
- moai-domain-backend: Backend architecture patterns

---

Version: 1.0.0 | Updated: 2025-12-30
