# Skill: moai-baas-neon-ext

## Metadata

```yaml
skill_id: moai-baas-neon-ext
skill_name: Neon Serverless Postgres & Development Branching
version: 1.0.0
created_date: 2025-11-09
updated_date: 2025-11-09
language: english
triggers:
  - keywords: ["Neon", "PostgreSQL", "Database branching", "Serverless", "Postgres"]
  - contexts: ["neon-detected", "pattern-b", "postgres-branching"]
agents:
  - database-expert
  - backend-expert
  - devops-expert
freedom_level: high
word_count: 1000
context7_references:
  - url: "https://neon.tech/docs/get-started-with-neon"
    topic: "Getting Started with Neon"
  - url: "https://neon.tech/docs/manage/branches"
    topic: "Database Branching Guide"
  - url: "https://neon.tech/docs/reference/pg-connection"
    topic: "PostgreSQL Connection Pooling"
  - url: "https://neon.tech/docs/serverless/python"
    topic: "Serverless Python Integration"
  - url: "https://neon.tech/docs/manage/projects"
    topic: "Project Management"
spec_reference: "@SPEC:BAAS-ECOSYSTEM-001"
```

---

## 📚 Content

### 1. Neon Architecture & Branching Concepts (150 words)

**Neon** is a serverless PostgreSQL platform with native database branching for development workflows.

**Core Philosophy**:
```
Traditional PostgreSQL:
  Production (main DB) → Manual backup/restore
  (Time-consuming, error-prone)

Neon Branching:
  Main branch → Dev branch (instant copy)
  (Zero-copy, instant restore, isolation)
```

**Key Features**:
- **Serverless Postgres**: Runs anywhere (AWS, Vercel Edge, Cloudflare Workers)
- **Database Branching**: Create isolated dev instances in seconds
- **Connection Pooling**: Built-in PgBouncer (eliminates connection limits)
- **Autoscaling**: Auto-pause unused databases ($0/month)
- **Compute Sharing**: Multiple branches share single compute

**Architecture Stack**:
```
┌──────────────────────────────┐
│ Neon (Serverless Postgres)   │
├──────────────────────────────┤
│ 1. Main Branch (Production)  │
│    └─ 10GB included         │
│                              │
│ 2. Dev Branches (Instant)   │
│    └─ Per-developer copy    │
│                              │
│ 3. Connection Pooling       │
│    └─ PgBouncer (free)      │
│                              │
│ 4. Autoscaling Compute      │
│    └─ Scale to zero         │
└──────────────────────────────┘
```

---

### 2. Database Schema & Branching Workflow (200 words)

**PostgreSQL Schema Setup**:

```sql
-- Production schema (main branch)
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT UNIQUE NOT NULL,
  name TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE posts (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id),
  title TEXT NOT NULL,
  content TEXT,
  published BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for production performance
CREATE INDEX idx_posts_user_id ON posts(user_id);
CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
CREATE INDEX idx_users_email ON users(email);
```

**Development Branching Workflow**:

```bash
# 1. Create dev branch from main
neon branch create --parent-branch-id main --branch-name feature/new-column

# 2. Get connection string for dev branch
NEON_DATABASE_URL=$(neon connection-string --branch-name feature/new-column)

# 3. Run migrations on dev branch
psql $NEON_DATABASE_URL < migrations/add_user_role.sql

# 4. Test locally with dev data
export DATABASE_URL=$NEON_DATABASE_URL
npm run test

# 5. Merge back to main (schema only, data separate)
# Manual schema review → Apply to main branch

# 6. Cleanup dev branch
neon branch delete --branch-name feature/new-column
```

**Per-Developer Workflow** (Solves "database lock" problem):

```bash
# Developer A gets isolated branch
neon branch create --parent-branch-id main --branch-name dev-alice

# Developer B gets separate isolated branch
neon branch create --parent-branch-id main --branch-name dev-bob

# Both can test schema changes independently
# No conflicts, no waiting
```

---

### 3. Connection Pooling & Optimization (150 words)

**PgBouncer Connection Pooling** (Built-in, Free):

```
┌──────────────────────────────────┐
│ Client Connections (unlimited)   │
├──────────────────────────────────┤
│ PgBouncer Connection Pool        │
│ ├─ min_pool_size: 0              │
│ ├─ max_pool_size: 25             │
│ └─ Idle timeout: 5 minutes       │
├──────────────────────────────────┤
│ PostgreSQL Database              │
│ (max 100 connections)            │
└──────────────────────────────────┘
```

**Connection String with Pooling**:

```
# Pooling enabled (use for serverless/edge functions)
postgresql://user:password@ep-xxx.us-east-1.neon.tech/dbname?sslmode=require&sslcert=/path/to/cert

# Transaction mode (safest for serverless)
postgresql://user:password@ep-xxx.us-east-1.neon.tech:6432/dbname?sslmode=require
#                                                        ^^^^
#                                         Port 6432 = pooling mode
```

**Performance Tips**:
- ✅ Use port 6432 for connection pooling (unlimited connections)
- ✅ Set `idle_in_transaction_session_timeout = 60s` to prevent timeouts
- ✅ Use `SET SESSION CHARACTERISTICS` for transaction isolation
- ✅ Connection pooling increases throughput 10-100x for serverless

---

### 4. Production Deployment Strategy (200 words)

**CI/CD Branching Strategy**:

```yaml
# .github/workflows/deploy.yml
name: Deploy to Production

on:
  push:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    env:
      DATABASE_URL: ${{ secrets.NEON_STAGING_URL }}
    steps:
      - uses: actions/checkout@v3

      # 1. Create temporary test branch
      - name: Create test branch
        run: |
          TEST_BRANCH="test-${GITHUB_SHA:0:8}"
          neon branch create --parent-branch-id main --branch-name $TEST_BRANCH
          echo "TEST_BRANCH=$TEST_BRANCH" >> $GITHUB_ENV

      # 2. Run migrations on test branch
      - name: Run migrations
        run: |
          psql $(neon connection-string --branch-name $TEST_BRANCH) < migrations/*

      # 3. Run test suite
      - name: Test
        run: npm run test

      # 4. Cleanup test branch
      - name: Cleanup
        run: neon branch delete --branch-name $TEST_BRANCH

  deploy:
    needs: test
    if: success()
    steps:
      # Apply schema changes to main branch
      - name: Migrate production
        run: psql ${{ secrets.NEON_PRODUCTION_URL }} < migrations/*

      # Deploy application
      - name: Deploy app
        run: npm run deploy
```

**Zero-Downtime Deployment**:

```sql
-- Step 1: Add new column (non-blocking)
ALTER TABLE users ADD COLUMN status VARCHAR(50) DEFAULT 'active';

-- Step 2: Backfill data (done gradually)
UPDATE users SET status = 'active' WHERE status IS NULL LIMIT 1000;

-- Step 3: Add constraint after backfill complete
ALTER TABLE users ALTER COLUMN status SET NOT NULL;

-- Step 4: Create index after data is stable
CREATE INDEX idx_users_status ON users(status);
```

**Monitoring Queries**:

```sql
-- Monitor active connections
SELECT datname, usename, count(*) FROM pg_stat_activity GROUP BY datname, usename;

-- Check slow queries
SELECT query, mean_time, calls FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;

-- Database size
SELECT pg_size_pretty(pg_database_size('dbname'));
```

---

### 5. Cost Optimization & Scaling (150 words)

**Neon Pricing Model**:

| Component | Cost |
|-----------|------|
| **Compute** | Auto-pause to $0/month (perfect for dev) |
| **Storage** | $0.30/GB/month (10GB free) |
| **Egress** | $0.10/GB (first 1GB/month free) |
| **Branches** | Unlimited free (per-developer cost reduction) |

**Cost Optimization Strategies**:

```javascript
// Strategy 1: Auto-pause for dev branches (free)
neon branch create --autoscale-limit 0.25  // Pause after 5 min inactivity

// Strategy 2: Remove unused branches
neon branch delete --branch-name dev-old-feature

// Strategy 3: Monitor storage usage
neon projects list --format json | jq '.projects[].databases[].size_bytes'

// Strategy 4: Archive large tables
-- Move old data to separate archival table or S3
ALTER TABLE logs DETACH PARTITION logs_2024_q1;
```

**Scaling Considerations**:

- ✅ Serverless scaling: Automatic (from $0 to $30/month)
- ✅ Per-developer cost: Nearly free (autoscaling + unlimited free branches)
- ✅ Production database size: $3/GB/month (100GB = $30/month)
- ✅ Egress to Vercel/Cloudflare: Free (same AWS region)

---

### 6. Common Issues & Solutions (150 words)

| Issue | Solution |
|-------|----------|
| **Connection timeout** | Use port 6432 (pooling mode), increase `statement_timeout` |
| **Slow queries on branch** | Indexes not copied; recreate on branch after creation |
| **Transaction conflicts** | Use connection pooling mode (port 6432), avoid long transactions |
| **Branch creation slow** | First branch slow (~30s), subsequent instant |
| **Data size growing** | Monitor with `pg_database_size()`, archive old data |
| **Autoscale limits hit** | Increase `autoscale_limit` in project settings |

**Debugging Connection Issues**:

```bash
# Test connection
psql -h ep-xxx.us-east-1.neon.tech -U postgres -d neondb \
  -c "SELECT version();"

# Check connection limits
SELECT count(*) FROM pg_stat_activity;

# Monitor pooling status
SHOW pool_mode;  -- Returns 'session' or 'transaction'
```

---

## 🎯 Usage

### Invocation from Agents
```python
Skill("moai-baas-neon-ext")
# Load when Pattern B (Neon + Clerk + Vercel) detected
```

### Context7 Integration
When Neon platform detected:
- Database branching for dev/production separation
- Connection pooling for serverless integration
- Schema migration strategies
- Cost optimization via autoscaling

---

## 📚 Reference Materials

- [Neon Getting Started](https://neon.tech/docs/get-started-with-neon)
- [Database Branching Guide](https://neon.tech/docs/manage/branches)
- [PostgreSQL Connection Pooling](https://neon.tech/docs/reference/pg-connection)
- [Serverless Integration](https://neon.tech/docs/serverless/python)
- [Project Management](https://neon.tech/docs/manage/projects)

---

## ✅ Validation Checklist

- [x] Neon architecture & branching concepts
- [x] Database schema & branching workflow
- [x] Connection pooling optimization
- [x] Production deployment strategy
- [x] Cost optimization & scaling
- [x] Common issues & solutions
- [x] 1000-word target
- [x] English language (policy compliant)
