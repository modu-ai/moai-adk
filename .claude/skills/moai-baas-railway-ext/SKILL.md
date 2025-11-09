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
word_count: 800
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

### 4. Monitoring & Logs (150 words)

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

### 5. Cost Optimization & Scaling (150 words)

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

### 6. Common Issues & Solutions (50 words)

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
- [x] Monitoring & logs streaming
- [x] Cost optimization & scaling strategies
- [x] Common issues & troubleshooting
- [x] 800-word target
- [x] English language (policy compliant)
