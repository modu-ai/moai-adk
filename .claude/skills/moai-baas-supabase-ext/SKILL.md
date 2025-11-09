# Skill: moai-baas-supabase-ext

## Metadata

```yaml
skill_id: moai-baas-supabase-ext
skill_name: Supabase Advanced Guide (RLS, Migrations, Realtime, Production Best Practices)
version: 2.0.0
created_date: 2025-11-09
updated_date: 2025-11-09
language: english
triggers:
  - keywords: ["Supabase", "RLS", "Row Level Security", "PostgreSQL", "Migration", "Realtime", "Production", "Deployment"]
  - contexts: ["supabase-detected", "pattern-a", "pattern-d"]
agents:
  - backend-expert
  - database-expert
  - security-expert
freedom_level: high
word_count: 1300
context7_references:
  - url: "https://supabase.com/docs/guides/database/postgres/row-level-security"
    topic: "RLS Policy Writing"
  - url: "https://supabase.com/docs/guides/database/migrations"
    topic: "Migration Safety"
  - url: "https://supabase.com/docs/guides/realtime"
    topic: "Realtime Subscriptions"
  - url: "https://supabase.com/docs/guides/database/connections"
    topic: "Connection Pooling & Supavisor"
  - url: "https://supabase.com/docs/guides/database/postgres/indexes"
    topic: "Database Indexing Strategy"
spec_reference: "@SPEC:BAAS-ECOSYSTEM-001"
```

---

## 📚 Content

### 1. Supabase Architecture (150 words)

**Supabase** is an open-source Firebase alternative built on PostgreSQL with enterprise features.

**Core Components**:
```
┌─────────────────────────────────┐
│ Supabase (PostgreSQL Platform)  │
├─────────────────────────────────┤
│ 1. PostgreSQL Database          │
│    └─ Tables, Functions, Triggers
│                                  │
│ 2. Authentication               │
│    └─ Email, Magic Link, OAuth  │
│                                  │
│ 3. Row Level Security (RLS)     │
│    └─ Policy-based access control
│                                  │
│ 4. Real-time Subscriptions      │
│    └─ Broadcast, Postgres Changes
│                                  │
│ 5. Storage                       │
│    └─ File buckets with CDN     │
│                                  │
│ 6. Edge Functions               │
│    └─ Serverless TypeScript      │
└─────────────────────────────────┘
```

**Edge Functions vs Database Functions**:

| Feature | Edge Functions | Database Functions |
|---------|---|---|
| Language | TypeScript/JavaScript | PL/pgSQL, Python |
| Location | Edge (global) | Database |
| Use Case | HTTP requests | DB triggers |
| Performance | <50ms | Variable |

---

### 2. RLS (Row Level Security) Advanced (300 words)

**RLS Definition**: PostgreSQL feature that controls row-level access based on user roles and policies.

**Core Concept**:
```sql
-- Example: users table
-- Rule: Users can only see their own data

ALTER TABLE users ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Users can view their own data"
ON users FOR SELECT
USING (auth.uid() = id);

CREATE POLICY "Users can update their own data"
ON users FOR UPDATE
USING (auth.uid() = id)
WITH CHECK (auth.uid() = id);
```

**Policy Writing Patterns**:

**Pattern 1: Self-only access (Most Common)**
```sql
CREATE POLICY "Self access"
ON profiles FOR ALL
USING (auth.uid() = user_id);
```

**Pattern 2: Role-based access**
```sql
CREATE POLICY "Admin or owner can delete"
ON posts FOR DELETE
USING (
  auth.uid() = user_id
  OR auth.jwt()->>'role' = 'admin'
);
```

**Pattern 3: Shared data with others**
```sql
CREATE POLICY "Shared with me"
ON documents FOR SELECT
USING (
  user_id = auth.uid()
  OR shared_with @> jsonb_build_array(auth.uid()::text)
);
```

**Debugging RLS 500 Errors**:

```
Issue: "new row violates row-level security policy"
Cause: Missing SELECT policy after INSERT

Solution:
1. Supabase Dashboard → SQL Editor
2. Check logs: SELECT * FROM auth.logs
3. Validate policies:
   SELECT * FROM pg_policies WHERE schemaname='public';
```

**Testing Policies with pgTAP**:

```sql
-- Policy validation with pgTAP
CREATE OR REPLACE FUNCTION test_rls()
RETURNS void AS $$
DECLARE
  user_id uuid := 'xxx';
BEGIN
  -- Verify user only sees own data
  ASSERT (
    SELECT COUNT(*) FROM profiles
    WHERE user_id = auth.uid()
  ) = 1;
END;
$$ LANGUAGE plpgsql;
```

**RLS Security Best Practices**:
- ✅ Enable RLS on all tables
- ✅ Define SELECT, INSERT, UPDATE, DELETE policies per table
- ✅ Always include auth.uid() checks
- ✅ Validate JWT claims (`auth.jwt()->>'role'`)
- ❌ Never expose Service Role tokens

---

### 3. Database Functions (200 words)

**Database Functions**: Expose PostgreSQL functions as RPC (Remote Procedure Call) endpoints.

**Use Cases**:
- Complex business logic
- Atomic operations
- Multi-table updates

**Example: Create tweet with counter increment**

```sql
CREATE OR REPLACE FUNCTION create_tweet(
  p_content TEXT,
  p_user_id UUID
)
RETURNS tweets AS $$
DECLARE
  v_tweet tweets;
BEGIN
  -- Insert tweet
  INSERT INTO tweets (content, user_id, created_at)
  VALUES (p_content, p_user_id, NOW())
  RETURNING * INTO v_tweet;

  -- Increment user tweet count (single transaction)
  UPDATE users
  SET tweet_count = tweet_count + 1
  WHERE id = p_user_id;

  RETURN v_tweet;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

**Client invocation**:
```typescript
const { data, error } = await supabase.rpc('create_tweet', {
  p_content: 'Hello World',
  p_user_id: userId
});
```

**Triggers**: Automated function execution

```sql
CREATE OR REPLACE FUNCTION update_user_stats()
RETURNS TRIGGER AS $$
BEGIN
  -- Increment count on every new tweet
  UPDATE users
  SET tweet_count = tweet_count + 1
  WHERE id = NEW.user_id;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_tweet_created
AFTER INSERT ON tweets
FOR EACH ROW
EXECUTE FUNCTION update_user_stats();
```

---

### 4. Migrations (200 words)

**Migrations**: Database schema versioning and tracking.

**Strategy 1: Migration-first (Recommended)**

```bash
# 1. Create migration
supabase migration new add_user_table

# 2. Write SQL
cat supabase/migrations/20250101120000_add_user_table.sql
CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
  email TEXT UNIQUE NOT NULL,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

# 3. Test locally
supabase db reset

# 4. Deploy to production
supabase db push
```

**Strategy 2: Dashboard-first (Avoid)**

```
Creating tables directly in Supabase Dashboard
→ No migration files
→ Can't sync with team
→ Can't deploy to production
```

**Safe migration patterns**:

```sql
-- ❌ Risky: Data loss possible
ALTER TABLE users DROP COLUMN email;

-- ✅ Safe: Step-by-step changes
-- Step 1: Add new column
ALTER TABLE users ADD COLUMN email_new TEXT;

-- Step 2: Migrate data
UPDATE users SET email_new = email;

-- Step 3: Remove old column (next deploy)
ALTER TABLE users DROP COLUMN email;
```

**Rollback strategy**:
```bash
# Rollback to previous migration
supabase db push --version 20250101110000
```

---

### 5. Realtime (100 words)

**Realtime**: WebSocket-based real-time data synchronization.

**Two modes**:

**Mode 1: Broadcast** (Message passing)
```typescript
// User 1: Broadcast message
supabase.realtime.channel('game').send({
  type: 'broadcast',
  event: 'player_moved',
  payload: { x: 100, y: 200 }
});

// User 2: Receive message
channel.on('broadcast', { event: 'player_moved' }, (payload) => {
  console.log('Player moved:', payload);
});
```

**Mode 2: Postgres Changes** (DB change detection)
```typescript
supabase
  .channel('public:messages')
  .on(
    'postgres_changes',
    { event: 'INSERT', schema: 'public', table: 'messages' },
    (payload) => {
      console.log('New message:', payload.new);
    }
  )
  .subscribe();
```

**Performance**: 1000+ concurrent connections, RLS automatically enforced.

---

### 6. Production Best Practices (300 words)

**Connection Pooling with Supavisor**:

Production deployments must use Supavisor for connection management:

```typescript
// Supabase connection string with Supavisor
const supabase = createClient(
  process.env.SUPABASE_URL,
  process.env.SUPABASE_ANON_KEY,
  {
    db: {
      schema: 'public',
    }
  }
);

// Connection pooling settings (from Dashboard → Database → Connection Pooling)
// Min pool: 5, Max pool: 20, Timeout: 3s
```

**Database Indexing Strategy**:

Smart indexing prevents slow queries and reduces costs:

```sql
-- 1. Identify slow queries using Supabase Logs
-- Dashboard → Logs → Database → Sort by duration

-- 2. Create composite indexes for common filters
CREATE INDEX idx_posts_user_created ON posts(user_id, created_at DESC);

-- 3. Use EXPLAIN QUERY PLAN to verify
EXPLAIN QUERY PLAN
SELECT * FROM posts WHERE user_id = $1 ORDER BY created_at DESC;

-- 4. Monitor index bloat
SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read, idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;
```

**RLS Performance Optimization**:

RLS policies can slow queries if poorly written:

```sql
-- ❌ Slow: Subquery in USING clause
CREATE POLICY "slow_policy" ON posts FOR SELECT
USING (user_id IN (SELECT id FROM users WHERE status = 'active'));

-- ✅ Fast: Direct column comparison
CREATE POLICY "fast_policy" ON posts FOR SELECT
USING (user_id = auth.uid() AND auth.jwt()->>'status' = 'active');
```

**Monitoring with Supabase Logs**:

```typescript
// Check database performance metrics
// Dashboard → Logs → Database
// - Query execution time
// - Connection usage
// - Replication lag
```

**Backup Strategy**:

```bash
# Automatic daily backups (included in paid plans)
# Dashboard → Database → Backups

# Manual backup for critical data
pg_dump --dbname=$DATABASE_URL > backup.sql

# Restore
psql --dbname=$DATABASE_URL < backup.sql
```

---

### 7. Security & Cost Optimization (100 words)

**Cost Reduction**:

```sql
-- Monthly cost depends on:
-- 1. Database size (included 0-500MB free)
-- 2. Egress bandwidth
-- 3. Realtime connections

-- Monitor costs:
-- Dashboard → Database → Usage → Database size
-- Kill idle connections to reduce bandwidth
SELECT pid, now() - query_start AS duration, query
FROM pg_stat_activity
WHERE state = 'idle' AND query_start < now() - INTERVAL '15 minutes';
```

**Service Role Key Security**:

- ✅ Store in `.env` (never in code)
- ✅ Rotate quarterly
- ❌ Never expose in browser code
- ❌ Never share via email or logs

---

## 🎯 Usage

### Agent Invocation

```python
# From backend-expert or database-expert
Skill("moai-baas-supabase-ext")

# Auto-loaded when Supabase patterns detected
```

### Context7 Auto-loading

When Supabase detected, these docs auto-loaded:
- RLS policy writing guide
- Migration best practices
- Realtime subscriptions
- Connection pooling
- Database indexing strategy

---

## 📚 Reference Materials

- [Supabase RLS Documentation](https://supabase.com/docs/guides/database/postgres/row-level-security)
- [Migration Guide](https://supabase.com/docs/guides/database/migrations)
- [Realtime Documentation](https://supabase.com/docs/guides/realtime)
- [Connection Pooling Guide](https://supabase.com/docs/guides/database/connections)
- [Database Indexing](https://supabase.com/docs/guides/database/postgres/indexes)

---

## ✅ Validation Checklist

- [x] Supabase architecture overview
- [x] RLS advanced patterns
- [x] Database functions & triggers
- [x] Safe migrations & rollback
- [x] Realtime subscriptions
- [x] Production best practices (connection pooling, indexing, RLS optimization)
- [x] Backup strategy & monitoring
- [x] Security & cost optimization
- [x] 1300+ word target (from 1000)
- [x] English language (policy compliant)
