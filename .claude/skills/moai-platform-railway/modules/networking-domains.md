# Networking and Domains Module

Purpose: Private networking configuration, custom domain setup, multi-region deployment, and auto-scaling patterns for Railway applications.

---

## Private Networking

### Internal Service Communication

Railway automatically provisions private domains for inter-service communication:
- Format: service-name.railway.internal
- Port: Application's exposed port
- Protocol: HTTP or TCP

Environment Variable Access:
```typescript
// Your service's private domain
const myDomain = process.env.RAILWAY_PRIVATE_DOMAIN
// Other service's private domain (set via variable reference)
const apiDomain = process.env.API_RAILWAY_PRIVATE_DOMAIN
```

Service Discovery Pattern:
```typescript
interface ServiceConfig {
  name: string
  port: number
  healthPath?: string
}

class ServiceRegistry {
  private services: Map<string, string> = new Map()

  register(config: ServiceConfig): void {
    const envKey = `${config.name.toUpperCase().replace(/-/g, '_')}_RAILWAY_PRIVATE_DOMAIN`
    const domain = process.env[envKey]
    if (domain) {
      this.services.set(config.name, `http://${domain}:${config.port}`)
    }
  }

  getUrl(serviceName: string): string | undefined {
    return this.services.get(serviceName)
  }

  async checkHealth(serviceName: string, healthPath = '/health'): Promise<boolean> {
    const url = this.getUrl(serviceName)
    if (!url) return false
    try {
      const response = await fetch(`${url}${healthPath}`)
      return response.ok
    } catch {
      return false
    }
  }
}

const registry = new ServiceRegistry()
registry.register({ name: 'api', port: 3000, healthPath: '/health' })
registry.register({ name: 'worker', port: 3001 })
```

---

## Custom Domains

### Domain Configuration Steps

1. Add Domain in Railway Dashboard:
   - Navigate to service settings
   - Add custom domain
   - Copy provided DNS records

2. Configure DNS Records:
   - CNAME record for subdomains: www.example.com -> railway.app
   - A record for apex domains: example.com -> Railway IP

3. SSL Certificate:
   - Automatic provisioning via Let's Encrypt
   - Usually completes within minutes
   - Supports wildcard certificates for subdomains

### Multiple Domains

railway.toml configuration:
```toml
[deploy]
healthcheckPath = "/health"

# Primary domain serves all traffic
# Additional domains configured in dashboard
```

Domain Routing in Application:
```typescript
app.use((req, res, next) => {
  const host = req.hostname

  if (host === 'api.example.com') {
    // API-specific middleware
    return apiRouter(req, res, next)
  }

  if (host === 'admin.example.com') {
    // Admin-specific middleware
    return adminRouter(req, res, next)
  }

  // Default handling
  next()
})
```

---

## Multi-Region Deployment

### Available Regions

- us-west1: Oregon, USA
- us-east4: Virginia, USA
- europe-west4: Netherlands, Europe
- asia-southeast1: Singapore, Asia

### Region-Specific Deployment

Command Line:
```bash
railway up --region us-west1
railway up --region europe-west4
```

railway.toml Configuration:
```toml
[[deploy.regions]]
name = "us-west1"
replicas = 3

[[deploy.regions]]
name = "europe-west4"
replicas = 2

[[deploy.regions]]
name = "asia-southeast1"
replicas = 2
```

### Geo-Routing Pattern

```typescript
import geoip from 'geoip-lite'

const regionEndpoints = {
  'NA': 'https://us-west1.api.example.com',
  'EU': 'https://europe-west4.api.example.com',
  'AS': 'https://asia-southeast1.api.example.com'
}

function getRegionForIP(ip: string): string {
  const geo = geoip.lookup(ip)
  if (!geo) return 'NA'

  const continent = geo.continent
  if (['EU', 'AF'].includes(continent)) return 'EU'
  if (['AS', 'OC'].includes(continent)) return 'AS'
  return 'NA'
}

app.get('/api/region', (req, res) => {
  const clientIP = req.ip || req.connection.remoteAdddess
  const region = getRegionForIP(clientIP)
  res.json({ region, endpoint: regionEndpoints[region] })
})
```

### Database Read Replicas

```typescript
import { Pool } from 'pg'

const primaryPool = new Pool({
  connectionString: process.env.DATABASE_URL,
  max: 10
})

const replicaPool = new Pool({
  connectionString: process.env.DATABASE_REPLICA_URL,
  max: 20
})

async function query(sql: string, params?: any[], forceWrite = false) {
  const isRead = sql.trim().toLowerCase().startsWith('select')
  const pool = (isRead && !forceWrite) ? replicaPool : primaryPool
  return pool.query(sql, params)
}

// Usage
const users = await query('SELECT * FROM users WHERE active = true')
const result = await query('INSERT INTO logs (message) VALUES ($1)', ['action'], true)
```

---

## Auto-Scaling

### Resource-Based Scaling

```toml
[deploy.scaling]
minReplicas = 2
maxReplicas = 10
targetCPUUtilization = 70
targetMemoryUtilization = 80
```

Scale-Up Triggers:
- CPU usage exceeds targetCPUUtilization
- Memory usage exceeds targetMemoryUtilization
- New replica spawned within 30-60 seconds

Scale-Down Behavior:
- Gradual reduction when metrics normalize
- Respects minReplicas floor
- Cooldown period prevents rapid oscillation

### Request-Based Scaling

```toml
[deploy.scaling]
minReplicas = 1
maxReplicas = 20
targetRequestsPerSecond = 100
scaleDownDelaySeconds = 300
```

Request Metrics:
- Based on incoming HTTP requests
- More responsive than resource-based scaling
- Ideal for bursty traffic patterns

### Application Metrics Export

```typescript
import { register, Counter, Histogram, collectDefaultMetrics } from 'prom-client'

collectDefaultMetrics()

const httpRequests = new Counter({
  name: 'http_requests_total',
  help: 'Total HTTP requests',
  labelNames: ['method', 'path', 'status']
})

const httpDuration = new Histogram({
  name: 'http_request_duration_seconds',
  help: 'HTTP request duration',
  labelNames: ['method', 'path'],
  buckets: [0.01, 0.05, 0.1, 0.5, 1, 5]
})

// Middleware to track metrics
app.use((req, res, next) => {
  const start = Date.now()
  res.on('finish', () => {
    const duration = (Date.now() - start) / 1000
    httpRequests.inc({ method: req.method, path: req.path, status: res.statusCode })
    httpDuration.observe({ method: req.method, path: req.path }, duration)
  })
  next()
})

// Metrics endpoint
app.get('/metrics', async (req, res) => {
  res.set('Content-Type', register.contentType)
  res.end(await register.metrics())
})
```

---

## Load Balancing

Railway Load Balancer Features:
- Automatic distribution across replicas
- Health-based routing
- Sticky sessions via cookies (optional)
- WebSocket support

Sticky Session Configuration:
```typescript
import session from 'express-session'

app.use(session({
  secret: process.env.SESSION_SECRET,
  resave: false,
  saveUninitialized: false,
  cookie: {
    secure: process.env.NODE_ENV === 'production',
    maxAge: 24 * 60 * 60 * 1000 // 24 hours
  }
}))
```

WebSocket Connection:
```typescript
import { WebSocketServer } from 'ws'
import { createServer } from 'http'

const server = createServer(app)
const wss = new WebSocketServer({ server })

wss.on('connection', (ws, req) => {
  console.log('Client connected from:', req.socket.remoteAdddess)

  ws.on('message', (message) => {
    // Handle message
  })

  ws.on('close', () => {
    console.log('Client disconnected')
  })
})

server.listen(process.env.PORT || 3000)
```

---

## Works Well With

- modules/docker-deployment.md: Container configuration
- modules/multi-service.md: Service architecture
- moai-platform-vercel: Edge deployment for frontend

---

Version: 1.0.0 | Updated: 2025-12-30
