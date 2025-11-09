# Skill: moai-baas-railway-ext

## Metadata

```yaml
skill_id: moai-baas-railway-ext
skill_name: Railway Full-Stack Platform & Deployment
version: 1.0.0
created_date: 2025-11-09
updated_date: 2025-11-09
language: english
triggers:
  - keywords: ["Railway", "Full-stack", "Deployment", "PostgreSQL", "Docker"]
  - contexts: ["railway-detected", "pattern-c", "full-stack-platform"]
agents:
  - devops-expert
  - backend-expert
  - database-expert
freedom_level: high
word_count: 1200
context7_references:
  - url: "https://docs.railway.app"
    topic: "Railway Documentation"
  - url: "https://docs.railway.app/databases/postgresql"
    topic: "PostgreSQL on Railway"
  - url: "https://docs.railway.app/deploy/deployments"
    topic: "Deployment Process"
  - url: "https://docs.railway.app/develop/variables"
    topic: "Environment Variables"
  - url: "https://docs.railway.app/reference/cli"
    topic: "Railway CLI Reference"
spec_reference: "@SPEC:BAAS-ECOSYSTEM-001"
```

---

## 📚 Content

### 1. Railway Platform Overview (100 words)

**Railway** is a simple all-in-one platform for deploying full-stack applications with integrated databases and monitoring.

**Core Philosophy**:
```
Railway Approach:
  Git Push → Auto-Deploy → Production Live
  (No infrastructure management needed)
```

**All-in-One Stack**:
- PostgreSQL database (managed)
- Backend hosting (Node.js, Python, Go, etc.)
- Environment variables & secrets
- Monitoring & logs
- Custom domains
- Auto-scaling

**Why Railway**:
- ✅ Simplest deployment experience
- ✅ Built-in PostgreSQL
- ✅ Super cheap ($5-50/month)
- ✅ Git integration (auto-deploy)
- ✅ Perfect for monoliths

---

### 2. Project Setup & Database (200 words)

**Initial Setup**:

```bash
# 1. Install Railway CLI
npm install -g @railway/cli

# 2. Login
railway login

# 3. Create new project
railway init

# 4. Connect Git repository
railway link

# 5. Add PostgreSQL database
railway add postgresql
```

**Environment Configuration**:

```yaml
# railway.json
{
  "build": {
    "builder": "dockerfile"
  },
  "deploy": {
    "startCommand": "npm start",
    "restartPolicyType": "on_failure"
  }
}
```

**PostgreSQL Setup**:

```typescript
// src/db.ts
import { Pool } from "pg";

const pool = new Pool({
  connectionString: process.env.DATABASE_URL,
  ssl: true, // Railway uses SSL
});

// Run migrations
async function runMigrations() {
  const client = await pool.connect();

  try {
    await client.query(`
      CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        email TEXT UNIQUE NOT NULL,
        name TEXT,
        created_at TIMESTAMP DEFAULT NOW()
      );
    `);
  } finally {
    client.release();
  }
}

export { pool, runMigrations };
```

**Docker Setup** (Railway Requirement):

```dockerfile
# Dockerfile
FROM node:18-alpine

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production

COPY . .

EXPOSE 3000

CMD ["npm", "start"]
```

---

### 3. Deployment & Preview Environments (200 words)

**Git-based Deployment**:

```bash
# Push to GitHub (Railway watches)
git add .
git commit -m "Update feature"
git push origin main

# Railway auto-deploys in <5 minutes
# View deployment logs
railway logs --service=api
```

**Multiple Environments**:

```bash
# Create staging environment
railway env create staging

# Switch to staging
railway env staging

# Add postgres to staging
railway add postgresql

# Deploy to staging
railway deploy --service=api

# Production stays on main branch
# Preview on feature branches
```

**Environment Variables**:

```bash
# Set variables
railway variables set DATABASE_URL="..."
railway variables set API_KEY="secret123"
railway variables set NODE_ENV="production"

# Reference in code
const apiKey = process.env.API_KEY;
const dbUrl = process.env.DATABASE_URL;
```

**Health Checks**:

```typescript
// server.ts
import express from "express";

const app = express();

// Health check endpoint
app.get("/health", (req, res) => {
  res.json({ status: "ok", timestamp: new Date() });
});

app.listen(3000, () => {
  console.log("Server running on port 3000");
});
```

**Preview Deployments**:

```bash
# Create PR with feature branch
git push origin feature/new-api

# Railway automatically deploys preview at:
# https://[project]-[branch].railway.app

# Once merged to main, production deploys
# https://[project].railway.app
```

---

### 4. Advanced Deployment Strategies (250 words)

**Blue-Green Deployment** (Zero-downtime updates):

```bash
# Current setup: Green (live) → Blue (staging)
# Step 1: Deploy new version to Blue environment
railway env create blue
railway env blue
git checkout production  # or deploy specific version
railway deploy --service=api

# Step 2: Run smoke tests on Blue
curl https://blue-[project].railway.app/health
npm run integration-tests

# Step 3: Switch traffic from Green to Blue
# Railway → Project Settings → Domains → Switch target
# OR use traffic splitting

# Step 4: Keep Green as rollback target for 24 hours
railway env green
# Keep this running for quick rollback if needed

# Step 5: Delete Green after confidence period
railway env delete green
```

**Canary Deployment** (Gradual rollout):

```bash
# Step 1: Deploy new version as canary
railway env create canary
railway env canary
git push canary-deployment

# Step 2: Route 10% traffic to canary
# Railway → Project Settings → Domains
# Advanced → Traffic splitting: 90% Green, 10% Blue

# Step 3: Monitor metrics for canary
railway logs --service=api --follow
# Watch for errors, latency, memory usage

# Step 4: Gradually increase canary traffic
# 10% → 25% → 50% → 100%
# Each step: monitor for 30 minutes

# Step 5: Complete rollout
# Full traffic to canary, delete old version
railway env delete green
```

**Database Migrations During Deploy**:

```typescript
// Safe migration pattern
async function migrate() {
  const client = await pool.connect();

  try {
    // Step 1: Create new column (non-blocking)
    await client.query(`
      ALTER TABLE users ADD COLUMN status VARCHAR(50) DEFAULT 'active';
    `);

    // Step 2: Backfill gradually (during traffic)
    await client.query(`
      UPDATE users SET status = 'active'
      WHERE status IS NULL
      LIMIT 10000;
    `);

    // Step 3: Add constraint after backfill complete
    await client.query(`
      ALTER TABLE users ALTER COLUMN status SET NOT NULL;
    `);

  } finally {
    client.release();
  }
}
```

**Deployment Checklist**:
- ✅ Test build locally with Docker
- ✅ Run full test suite in CI/CD
- ✅ Database migration tested in staging
- ✅ Environment variables verified
- ✅ Dependencies up-to-date
- ✅ Health check endpoint responding
- ✅ Monitoring alerts configured
- ✅ Rollback procedure documented

---

### 5. Custom Domains & SSL Configuration (150 words)

**Setup Custom Domain**:

```bash
# Step 1: Add custom domain to Railway
railway domain add example.com

# Step 2: Get DNS records from Railway
# Railway Dashboard → Project → Domains → example.com
# Shows: CNAME target

# Step 3: Update DNS provider (Namecheap, Route53, etc.)
# CNAME: example.com → [railway-provided-cname]
# Wait 5-30 minutes for DNS propagation

# Verify with:
nslookup example.com
dig example.com
```

**SSL Certificate (Automatic)**:

Railway automatically provisions Let's Encrypt certificates:
- ✅ Automatic renewal (before expiry)
- ✅ HTTPS enabled by default
- ✅ HTTP → HTTPS redirect automatic
- ✅ No manual certificate management needed

**Subdomain Routing**:

```typescript
// Route based on subdomain
app.get("*", (req, res, next) => {
  const subdomain = req.subdomains[0]; // e.g., "api" from api.example.com

  if (subdomain === "api") {
    // Route to API handlers
    return apiRouter.handle(req, res);
  }

  if (subdomain === "www" || !subdomain) {
    // Route to web app
    return webRouter.handle(req, res);
  }

  // Unknown subdomain
  res.status(404).send("Not found");
});
```

**Domain Troubleshooting**:
- ❌ SSL not showing: Wait 10 minutes after DNS propagation
- ❌ DNS not resolving: Check CNAME matches exactly
- ❌ Redirect loop: Verify protocol (http vs https)

---

### 6. Advanced Monitoring & Alerting (150 words)

**Configure Uptime Monitoring**:

```bash
# Railway Dashboard → Alerts
# Create uptime alert for /health endpoint

# Alert conditions:
# - Response time > 2s
# - Status code != 200
# - Service down > 5 minutes
```

**Log Aggregation Setup**:

```bash
# Export logs to external service
# Step 1: Configure log drain in Railway
railway variable set LOG_DRAIN_URL="https://your-drain.com"

# Step 2: Send structured logs
export LOG_LEVEL=info
export LOG_FORMAT=json  # For log parsing
```

**Performance Monitoring Metrics**:

```sql
-- Database performance queries
-- Monitor connection pool usage
SELECT count(*) FROM pg_stat_activity WHERE state = 'active';

-- Query performance analysis
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'user@example.com';

-- Cache hit ratio
SELECT
  sum(heap_blks_read) as heap_read,
  sum(heap_blks_hit)  as heap_hit,
  sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)) as ratio
FROM pg_statio_user_tables;
```

**Error Rate Tracking**:

```typescript
// Log errors with context
import logger from "pino";

const log = logger({
  level: process.env.LOG_LEVEL || "info",
  transport: {
    target: "pino-pretty",
  },
});

app.use((err, req, res, next) => {
  log.error({
    error: err.message,
    status: err.status || 500,
    path: req.path,
    method: req.method,
    timestamp: new Date(),
  });

  res.status(err.status || 500).json({
    error: err.message,
  });
});
```

**Alert Thresholds**:
- CPU > 80%: Investigate
- Memory > 90%: Scale up
- Error rate > 1%: Page on-call
- Response time p99 > 5s: Investigate queries

---

### 7. Monitoring & Logs (150 words)

**View Real-time Logs**:

```bash
# Stream logs
railway logs --follow

# Filter by service
railway logs --service=api
railway logs --service=postgres

# View deployment history
railway logs --tail 50
```

**Database Monitoring**:

```sql
-- Check connection count
SELECT COUNT(*) FROM pg_stat_activity;

-- Find slow queries
SELECT query, mean_exec_time FROM pg_stat_statements
ORDER BY mean_exec_time DESC LIMIT 10;

-- Database size
SELECT pg_size_pretty(pg_database_size('postgres'));
```

**Railway Dashboard**:

- Dashboard → Logs: Real-time application logs
- Dashboard → Metrics: CPU, memory, requests/sec
- Dashboard → Deployments: Deployment history
- Dashboard → Alerts: Configure uptime monitoring

---

### 8. Cost Optimization & Scaling (150 words)

**Railway Pricing**:

```
Free tier:       $5/month usage
Pro:             $0.04/GB RAM/hour
Database:        Included (5GB included)

Example costs:
├─ API server (0.5GB) = $35/month
├─ PostgreSQL (10GB) = $5/month
└─ Total = ~$40/month
```

**Cost Reduction Strategies**:

```bash
# 1. Use smaller compute (256MB RAM = ~$18/month)
railway scale memory=256

# 2. Auto-scale based on traffic
railway scale --cpu-limit=1 --memory-limit=512

# 3. Cleanup old deployments (logs take space)
railway logs --cleanup

# 4. Use Railway's free PostgreSQL (5GB included)

# 5. Set database backup retention
# Dashboard → PostgreSQL → Backup Retention → 7 days
```

**Scaling Considerations**:

- ✅ Railway auto-scales (adds more instances if needed)
- ✅ Database auto-backup daily
- ✅ Suitable for: MVP, small startups, monolithic apps
- ⚠️ Limited for: Microservices (use Docker Compose instead)
- ⚠️ Not for: Very high traffic (consider load balancing)

---

### 9. Common Issues & Solutions (50 words)

| Issue | Solution |
|-------|----------|
| **Connection timeout** | Check DATABASE_URL is set; verify PostgreSQL service running |
| **Port already in use** | Railway uses random ports; check `echo $PORT` |
| **Out of memory** | Scale up with `railway scale memory=512` |
| **Slow deployment** | Check Docker build logs with `railway logs --follow` |

---

## 🎯 Usage

### Invocation from Agents
```python
Skill("moai-baas-railway-ext")
# Load when Pattern C (Railway all-in-one) detected
```

### Context7 Integration
When Railway detected:
- Project setup & PostgreSQL integration
- Git-based deployment workflow
- Environment configuration & secrets
- Monitoring & scaling strategies

---

## 📚 Reference Materials

- [Railway Documentation](https://docs.railway.app)
- [PostgreSQL on Railway](https://docs.railway.app/databases/postgresql)
- [Deployment Process](https://docs.railway.app/deploy/deployments)
- [Environment Variables](https://docs.railway.app/develop/variables)
- [Railway CLI Reference](https://docs.railway.app/reference/cli)

---

## ✅ Validation Checklist

- [x] Railway platform overview & philosophy
- [x] Project setup & database provisioning
- [x] Git-based deployment & preview environments
- [x] Advanced deployment strategies (blue-green, canary)
- [x] Custom domains & SSL configuration
- [x] Advanced monitoring & alerting
- [x] Monitoring & logs streaming
- [x] Cost optimization & scaling strategies
- [x] Common issues & troubleshooting
- [x] 1200-word target (from 800)
- [x] English language (policy compliant)
