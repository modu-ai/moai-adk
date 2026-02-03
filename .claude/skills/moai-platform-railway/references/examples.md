# Railway Deployment Examples

Working code examples for common Railway deployment scenarios.

---

## Example 1: Full-Stack Node.js Application

### Project Structure

```
project/
├── apps/
│   ├── api/
│   │   ├── Dockerfile
│   │   ├── railway.toml
│   │   ├── package.json
│   │   └── src/
│   │       └── index.ts
│   └── web/
│       ├── Dockerfile
│       ├── railway.toml
│       └── src/
├── package.json
└── turbo.json
```

### API Service

apps/api/Dockerfile:
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
CMD ["node", "dist/index.js"]
```

apps/api/railway.toml:
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "./Dockerfile"

[deploy]
healthcheckPath = "/health"
healthcheckTimeout = 60
numReplicas = 2

[deploy.resources]
memory = "512Mi"
cpu = "0.5"
```

apps/api/src/index.ts:
```typescript
import express from 'express'
import { Pool } from 'pg'

const app = express()
const pool = new Pool({ connectionString: process.env.DATABASE_URL })

app.use(express.json())

app.get('/health', async (req, res) => {
  try {
    await pool.query('SELECT 1')
    res.json({ status: 'healthy', database: 'connected' })
  } catch (error) {
    res.status(503).json({ status: 'unhealthy', database: 'disconnected' })
  }
})

app.get('/api/users', async (req, res) => {
  const result = await pool.query('SELECT id, name, email FROM users')
  res.json(result.rows)
})

const port = process.env.PORT || 3000
app.listen(port, '0.0.0.0', () => {
  console.log(`API server running on port ${port}`)
})
```

---

## Example 2: Python FastAPI with Worker

### API Service

apps/api/Dockerfile:
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

apps/api/main.py:
```python
from fastapi import FastAPI, BackgroundTasks
from redis import Redis
import json
import os

app = FastAPI()
redis = Redis.from_url(os.environ.get("REDIS_URL", "redis://localhost:6379"))

@app.get("/health")
async def health():
    try:
        redis.ping()
        return {"status": "healthy", "redis": "connected"}
    except Exception:
        return {"status": "unhealthy", "redis": "disconnected"}

@app.post("/api/jobs")
async def create_job(data: dict):
    job_id = str(uuid.uuid4())
    redis.lpush("job_queue", json.dumps({"id": job_id, "data": data}))
    return {"job_id": job_id, "status": "queued"}
```

### Worker Service

apps/worker/main.py:
```python
import json
import time
import os
from redis import Redis

redis = Redis.from_url(os.environ.get("REDIS_URL", "redis://localhost:6379"))

def process_job(job_data: dict):
    print(f"Processing job: {job_data['id']}")
    # Simulate work
    time.sleep(2)
    print(f"Completed job: {job_data['id']}")

def main():
    print("Worker started, waiting for jobs...")
    while True:
        _, message = redis.brpop("job_queue")
        job = json.loads(message)
        process_job(job)

if __name__ == "__main__":
    main()
```

apps/worker/railway.toml:
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "./Dockerfile"

[deploy]
startCommand = "python main.py"
numReplicas = 2

[deploy.resources]
memory = "256Mi"
cpu = "0.25"
```

---

## Example 3: Go Microservice with Volumes

### Main Application

main.go:
```go
package main

import (
    "encoding/json"
    "log"
    "net/http"
    "os"
    "path/filepath"
)

type FileStorage struct {
    basePath string
}

func NewFileStorage() *FileStorage {
    path := os.Getenv("RAILWAY_VOLUME_MOUNT_PATH")
    if path == "" {
        path = "/app/data"
    }
    os.MkdirAll(path, 0755)
    return &FileStorage{basePath: path}
}

func (fs *FileStorage) Save(name string, data []byte) error {
    return os.WriteFile(filepath.Join(fs.basePath, name), data, 0644)
}

func (fs *FileStorage) Load(name string) ([]byte, error) {
    return os.ReadFile(filepath.Join(fs.basePath, name))
}

func main() {
    storage := NewFileStorage()
    
    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
    })
    
    http.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
        if r.Method == "POST" {
            var data struct {
                Name    string `json:"name"`
                Content string `json:"content"`
            }
            json.NewDecoder(r.Body).Decode(&data)
            storage.Save(data.Name, []byte(data.Content))
            json.NewEncoder(w).Encode(map[string]string{"status": "saved"})
        }
    })
    
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Server starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
```

railway.toml:
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "./Dockerfile"

[deploy]
healthcheckPath = "/health"
healthcheckTimeout = 30
numReplicas = 1  # Required for volumes

[[volumes]]
mountPath = "/app/data"
name = "file-storage"
size = "10Gi"

[deploy.resources]
memory = "256Mi"
cpu = "0.25"
```

---

## Example 4: Scheduled Backup Service

### Backup Script

backup.ts:
```typescript
import { exec } from 'child_process'
import { promisify } from 'util'
import { createReadStream } from 'fs'
import { S3Client, PutObjectCommand } from '@aws-sdk/client-s3'

const execAsync = promisify(exec)

const s3 = new S3Client({
  region: process.env.AWS_REGION!,
  credentials: {
    accessKeyId: process.env.AWS_ACCESS_KEY_ID!,
    secretAccessKey: process.env.AWS_SECRET_ACCESS_KEY!
  }
})

async function backupDatabase(): Promise<void> {
  const timestamp = new Date().toISOString().replace(/[:.]/g, '-')
  const backupFile = `/tmp/backup-${timestamp}.sql.gz`
  
  // Create PostgreSQL dump
  await execAsync(
    `pg_dump "${process.env.DATABASE_URL}" | gzip > ${backupFile}`
  )
  
  // Upload to S3
  await s3.send(new PutObjectCommand({
    Bucket: process.env.BACKUP_BUCKET!,
    Key: `database-backups/backup-${timestamp}.sql.gz`,
    Body: createReadStream(backupFile)
  }))
  
  // Cleanup
  await execAsync(`rm ${backupFile}`)
  
  console.log(`Backup completed: backup-${timestamp}.sql.gz`)
}

backupDatabase().catch(console.error)
```

railway.toml:
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "./Dockerfile"

[deploy]
startCommand = "npx ts-node backup.ts"
cronSchedule = "0 0 * * *"  # Daily at midnight
numReplicas = 1

[deploy.resources]
memory = "256Mi"
cpu = "0.25"
```

---

## Example 5: Multi-Region API with Read Replicas

### Connection Manager

db.ts:
```typescript
import { Pool } from 'pg'

const primaryPool = new Pool({
  connectionString: process.env.DATABASE_URL,
  max: 10,
  idleTimeoutMillis: 30000
})

const replicaPool = new Pool({
  connectionString: process.env.DATABASE_REPLICA_URL || process.env.DATABASE_URL,
  max: 20,
  idleTimeoutMillis: 30000
})

export async function query(
  sql: string,
  params?: any[],
  options: { forceWrite?: boolean } = {}
): Promise<any> {
  const isWrite = !sql.trim().toLowerCase().startsWith('select')
  const pool = (isWrite || options.forceWrite) ? primaryPool : replicaPool
  
  const result = await pool.query(sql, params)
  return result.rows
}

export async function transaction<T>(
  callback: (query: (sql: string, params?: any[]) => Promise<any>) => Promise<T>
): Promise<T> {
  const client = await primaryPool.connect()
  try {
    await client.query('BEGIN')
    const result = await callback((sql, params) => client.query(sql, params).then(r => r.rows))
    await client.query('COMMIT')
    return result
  } catch (error) {
    await client.query('ROLLBACK')
    throw error
  } finally {
    client.release()
  }
}
```

railway.toml:
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "./Dockerfile"

[deploy]
healthcheckPath = "/health"
healthcheckTimeout = 60

[[deploy.regions]]
name = "us-west1"
replicas = 3

[[deploy.regions]]
name = "europe-west4"
replicas = 2

[deploy.scaling]
minReplicas = 2
maxReplicas = 10
targetRequestsPerSecond = 100

[deploy.resources]
memory = "1Gi"
cpu = "1"
```

---

## Example 6: WebSocket Service

### WebSocket Server

websocket.ts:
```typescript
import { WebSocketServer } from 'ws'
import { createServer } from 'http'
import express from 'express'

const app = express()
const server = createServer(app)
const wss = new WebSocketServer({ server })

const clients = new Set<WebSocket>()

app.get('/health', (req, res) => {
  res.json({
    status: 'healthy',
    clients: clients.size,
    replica: process.env.RAILWAY_REPLICA_ID
  })
})

wss.on('connection', (ws, req) => {
  clients.add(ws)
  console.log(`Client connected. Total: ${clients.size}`)
  
  ws.on('message', (message) => {
    // Broadcast to all clients
    const data = message.toString()
    clients.forEach(client => {
      if (client !== ws && client.readyState === 1) {
        client.send(data)
      }
    })
  })
  
  ws.on('close', () => {
    clients.delete(ws)
    console.log(`Client disconnected. Total: ${clients.size}`)
  })
  
  ws.on('error', console.error)
})

const port = process.env.PORT || 3000
server.listen(port, '0.0.0.0', () => {
  console.log(`WebSocket server running on port ${port}`)
})
```

railway.toml:
```toml
[build]
builder = "DOCKERFILE"
dockerfilePath = "./Dockerfile"

[deploy]
healthcheckPath = "/health"
healthcheckTimeout = 30
numReplicas = 3

[deploy.resources]
memory = "512Mi"
cpu = "0.5"
```

---

## GitHub Actions Workflow

Complete CI/CD pipeline:

```yaml
name: Railway Deploy

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

env:
  RAILWAY_TOKEN: ${{ secrets.RAILWAY_TOKEN }}

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
      - run: npm ci
      - run: npm test
      - run: npm run lint

  deploy:
    needs: test
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install Railway CLI
        run: npm i -g @railway/cli
      - name: Deploy to Railway
        run: railway up --detach
      - name: Verify Deployment
        run: |
          sleep 30
          railway status
```

---

Status: Production Ready | Updated: 2025-12-30
