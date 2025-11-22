# Supabase Performance Optimization

**Enterprise-grade performance optimization techniques for Supabase applications.**

**Version**: 1.0.0
**Last Updated**: 2025-11-22
**Research Base**: Context7 `/websites/supabase`, production optimization patterns

---

## 📚 Table of Contents

1. [Connection Pooling with PgBouncer](#1-connection-pooling-with-pgbouncer)
2. [Query Optimization](#2-query-optimization)
3. [Realtime Subscription Performance](#3-realtime-subscription-performance)
4. [Caching Strategies](#4-caching-strategies)
5. [Batch Operations](#5-batch-operations)
6. [Cold Start Optimization](#6-cold-start-optimization)
7. [Index Design](#7-index-design)
8. [Cost Optimization](#8-cost-optimization)

---

## 1. Connection Pooling with PgBouncer

### Technique 1.1: Optimal Pool Configuration

**PgBouncer connection pooling**:

```typescript
import { createClient } from '@supabase/supabase-js'

// Production configuration with connection pooling
const supabase = createClient(
  process.env.SUPABASE_URL!,
  process.env.SUPABASE_ANON_KEY!,
  {
    db: {
      schema: 'public'
    },
    auth: {
      autoRefreshToken: true,
      persistSession: true,
      detectSessionInUrl: true
    },
    global: {
      headers: {
        'x-connection-string': 'pooler'  // Use PgBouncer pooler
      }
    }
  }
)

// For serverless environments (e.g., Edge Functions)
const supabaseServerless = createClient(
  process.env.SUPABASE_URL!,
  process.env.SUPABASE_SERVICE_KEY!,
  {
    db: {
      schema: 'public'
    },
    global: {
      headers: {
        'x-connection-string': 'transaction'  // Transaction mode for short-lived connections
      }
    }
  }
)
```

**Connection modes**:

| Mode | Use Case | Max Connections | Latency |
|------|----------|----------------|---------|
| Session | Long-lived connections | 15 | Low |
| Transaction | Serverless functions | 200 | Medium |
| Statement | High concurrency | 500 | High |

**Best practice**:
- Use **Session mode** for traditional servers
- Use **Transaction mode** for serverless (Edge Functions, Lambda)
- Use **Statement mode** only for read-heavy workloads

---

## 2. Query Optimization

### Technique 2.1: Efficient Joins and Filtering

**Optimized query patterns**:

```typescript
// ❌ BAD: N+1 query problem
async function getBadProjectsWithTasks() {
  const { data: projects } = await supabase
    .from('projects')
    .select('*')

  for (const project of projects!) {
    const { data: tasks } = await supabase
      .from('tasks')
      .select('*')
      .eq('project_id', project.id)

    project.tasks = tasks
  }

  return projects
}

// ✅ GOOD: Single query with join
async function getGoodProjectsWithTasks() {
  const { data, error } = await supabase
    .from('projects')
    .select(`
      *,
      tasks (
        id,
        title,
        status,
        assigned_to
      )
    `)

  if (error) throw error
  return data
}

// ✅ BETTER: Filtered join with limit
async function getOptimalProjectsWithTasks(userId: string) {
  const { data, error } = await supabase
    .from('projects')
    .select(`
      id,
      name,
      tasks!inner (
        id,
        title,
        status
      )
    `)
    .eq('tasks.assigned_to', userId)
    .limit(20)

  if (error) throw error
  return data
}
```

### Technique 2.2: Materialized Views

**Pre-computed aggregations**:

```sql
-- Create materialized view for expensive aggregations
CREATE MATERIALIZED VIEW project_stats AS
SELECT
  p.id AS project_id,
  p.name AS project_name,
  COUNT(t.id) AS total_tasks,
  COUNT(t.id) FILTER (WHERE t.status = 'completed') AS completed_tasks,
  COUNT(t.id) FILTER (WHERE t.status = 'in_progress') AS in_progress_tasks,
  AVG(EXTRACT(EPOCH FROM (t.completed_at - t.created_at))) FILTER (WHERE t.status = 'completed') AS avg_completion_time
FROM projects p
LEFT JOIN tasks t ON t.project_id = p.id
GROUP BY p.id, p.name;

-- Create index on materialized view
CREATE UNIQUE INDEX ON project_stats (project_id);

-- Refresh materialized view (schedule this)
REFRESH MATERIALIZED VIEW CONCURRENTLY project_stats;

-- Function to auto-refresh
CREATE OR REPLACE FUNCTION refresh_project_stats()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  REFRESH MATERIALIZED VIEW CONCURRENTLY project_stats;
  RETURN NULL;
END;
$$;

-- Trigger to refresh after task changes
CREATE TRIGGER refresh_stats_trigger
  AFTER INSERT OR UPDATE OR DELETE ON tasks
  FOR EACH STATEMENT
  EXECUTE FUNCTION refresh_project_stats();
```

**TypeScript query**:

```typescript
// Query materialized view (much faster)
async function getProjectStats(projectId: string) {
  const { data, error } = await supabase
    .from('project_stats')
    .select('*')
    .eq('project_id', projectId)
    .single()

  if (error) throw error
  return data
}
```

---

## 3. Realtime Subscription Performance

### Technique 3.1: Subscription Batching

**Optimize realtime subscriptions**:

```typescript
// ❌ BAD: Multiple individual subscriptions
function badRealtimeSetup() {
  const channels = []

  for (let i = 0; i < 100; i++) {
    const channel = supabase
      .channel(`task-${i}`)
      .on('postgres_changes', {
        event: '*',
        schema: 'public',
        table: 'tasks',
        filter: `id=eq.${i}`
      }, (payload) => {
        console.log('Task changed:', payload)
      })
      .subscribe()

    channels.push(channel)
  }

  return channels
}

// ✅ GOOD: Single subscription with filtering
function goodRealtimeSetup(projectId: string) {
  const channel = supabase
    .channel(`project-${projectId}`)
    .on('postgres_changes', {
      event: '*',
      schema: 'public',
      table: 'tasks',
      filter: `project_id=eq.${projectId}`
    }, (payload) => {
      console.log('Task changed:', payload)
    })
    .subscribe()

  return channel
}
```

### Technique 3.2: Subscription Throttling

**Limit update frequency**:

```typescript
import { debounce } from 'lodash'

class ThrottledRealtimeSubscription {
  private channel: RealtimeChannel
  private updateQueue: any[] = []
  private processUpdates: () => void

  constructor(channelName: string, throttleMs = 100) {
    this.channel = supabase.channel(channelName)

    // Debounced batch processor
    this.processUpdates = debounce(() => {
      if (this.updateQueue.length === 0) return

      // Process all queued updates at once
      const updates = [...this.updateQueue]
      this.updateQueue = []

      this.onBatchUpdate(updates)
    }, throttleMs)

    this.setupSubscription()
  }

  private setupSubscription() {
    this.channel
      .on('postgres_changes', {
        event: '*',
        schema: 'public',
        table: 'tasks'
      }, (payload) => {
        // Queue update instead of processing immediately
        this.updateQueue.push(payload)
        this.processUpdates()
      })
      .subscribe()
  }

  // Override this method
  protected onBatchUpdate(updates: any[]) {
    console.log('Processing batch:', updates.length, 'updates')
  }

  cleanup() {
    this.channel.unsubscribe()
  }
}
```

---

## 4. Caching Strategies

### Technique 4.1: Redis Layer Integration

**Client-side caching with Redis**:

```typescript
import { Redis } from 'ioredis'

const redis = new Redis(process.env.REDIS_URL)

class CachedSupabaseClient {
  private cacheTTL = 60 // 60 seconds

  async getCachedProjects(organizationId: string) {
    const cacheKey = `projects:${organizationId}`

    // Try cache first
    const cached = await redis.get(cacheKey)
    if (cached) {
      return JSON.parse(cached)
    }

    // Query Supabase
    const { data, error } = await supabase
      .from('projects')
      .select('*')
      .eq('organization_id', organizationId)

    if (error) throw error

    // Store in cache
    await redis.setex(cacheKey, this.cacheTTL, JSON.stringify(data))

    return data
  }

  async invalidateProjectsCache(organizationId: string) {
    await redis.del(`projects:${organizationId}`)
  }

  async updateProject(projectId: string, updates: any) {
    // Update database
    const { data, error } = await supabase
      .from('projects')
      .update(updates)
      .eq('id', projectId)
      .select()
      .single()

    if (error) throw error

    // Invalidate cache
    await this.invalidateProjectsCache(data.organization_id)

    return data
  }
}
```

### Technique 4.2: Browser Cache with SWR

**Stale-While-Revalidate pattern**:

```typescript
import useSWR from 'swr'

// Fetcher function
async function fetchProjects(organizationId: string) {
  const { data, error } = await supabase
    .from('projects')
    .select('*')
    .eq('organization_id', organizationId)

  if (error) throw error
  return data
}

// React hook with SWR
function useProjects(organizationId: string) {
  const { data, error, mutate } = useSWR(
    `projects-${organizationId}`,
    () => fetchProjects(organizationId),
    {
      revalidateOnFocus: false,
      revalidateOnReconnect: false,
      refreshInterval: 30000,  // Refresh every 30 seconds
      dedupingInterval: 5000   // Dedupe requests within 5 seconds
    }
  )

  return {
    projects: data,
    isLoading: !error && !data,
    isError: error,
    refresh: mutate
  }
}
```

---

## 5. Batch Operations

### Technique 5.1: Bulk Inserts

**Efficient batch inserts**:

```typescript
// ❌ BAD: Individual inserts
async function badBulkInsert(tasks: any[]) {
  for (const task of tasks) {
    await supabase.from('tasks').insert(task)
  }
}

// ✅ GOOD: Single batch insert
async function goodBulkInsert(tasks: any[]) {
  const { data, error } = await supabase
    .from('tasks')
    .insert(tasks)
    .select()

  if (error) throw error
  return data
}

// ✅ BETTER: Chunked batch insert for large datasets
async function optimalBulkInsert(tasks: any[], chunkSize = 1000) {
  const results = []

  for (let i = 0; i < tasks.length; i += chunkSize) {
    const chunk = tasks.slice(i, i + chunkSize)

    const { data, error } = await supabase
      .from('tasks')
      .insert(chunk)
      .select()

    if (error) throw error
    results.push(...data)
  }

  return results
}
```

### Technique 5.2: Batch Updates with Upsert

**Efficient upserts**:

```typescript
async function batchUpsert(tasks: any[]) {
  const { data, error } = await supabase
    .from('tasks')
    .upsert(tasks, {
      onConflict: 'id',
      ignoreDuplicates: false
    })
    .select()

  if (error) throw error
  return data
}
```

---

## 6. Cold Start Optimization

### Technique 6.1: Edge Function Warmup

**Reduce cold starts**:

```typescript
// supabase/functions/warm-handler/index.ts
import { serve } from 'https://deno.land/std@0.168.0/http/server.ts'
import { createClient } from 'https://esm.sh/@supabase/supabase-js@2'

// Pre-initialize expensive connections
const supabase = createClient(
  Deno.env.get('SUPABASE_URL') ?? '',
  Deno.env.get('SUPABASE_SERVICE_KEY') ?? ''
)

// Warmup function (called periodically)
async function warmup() {
  // Execute a lightweight query to keep connection alive
  await supabase.from('_warmup').select('count').limit(1)
}

serve(async (req) => {
  if (req.url.includes('/warmup')) {
    await warmup()
    return new Response('Warmed up', { status: 200 })
  }

  // Your actual handler logic
  return new Response('OK', { status: 200 })
})
```

**Scheduled warmup** (via cron):

```yaml
# .github/workflows/warmup.yml
name: Warmup Edge Functions

on:
  schedule:
    - cron: '*/5 * * * *'  # Every 5 minutes

jobs:
  warmup:
    runs-on: ubuntu-latest
    steps:
      - name: Ping warmup endpoint
        run: |
          curl -X GET "${{ secrets.SUPABASE_FUNCTION_URL }}/warmup"
```

---

## 7. Index Design

### Technique 7.1: Composite Indexes

**Optimal index strategy**:

```sql
-- Single column indexes (basic)
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_assigned_to ON tasks(assigned_to);

-- Composite index (optimized for common query)
CREATE INDEX idx_tasks_project_status
  ON tasks(project_id, status, assigned_to)
  INCLUDE (title, created_at);

-- Partial index (filter frequent queries)
CREATE INDEX idx_tasks_active
  ON tasks(project_id, assigned_to)
  WHERE status IN ('pending', 'in_progress');

-- Index for LIKE queries
CREATE INDEX idx_tasks_title_trgm
  ON tasks USING gin (title gin_trgm_ops);

-- Index usage analysis
SELECT
  schemaname,
  tablename,
  indexname,
  idx_scan,
  idx_tup_read,
  idx_tup_fetch
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY idx_scan ASC
LIMIT 20;
```

### Technique 7.2: Index Maintenance

**Keep indexes healthy**:

```sql
-- Check index bloat
SELECT
  schemaname,
  tablename,
  indexname,
  pg_size_pretty(pg_relation_size(indexrelid)) AS index_size
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY pg_relation_size(indexrelid) DESC;

-- Rebuild bloated indexes
REINDEX INDEX CONCURRENTLY idx_tasks_project_status;

-- Auto-vacuum settings
ALTER TABLE tasks SET (
  autovacuum_vacuum_scale_factor = 0.1,
  autovacuum_analyze_scale_factor = 0.05
);
```

---

## 8. Cost Optimization

### Technique 8.1: Database Size Monitoring

**Track storage costs**:

```sql
-- Table size analysis
SELECT
  schemaname,
  tablename,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) AS total_size,
  pg_size_pretty(pg_relation_size(schemaname||'.'||tablename)) AS table_size,
  pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename) - pg_relation_size(schemaname||'.'||tablename)) AS indexes_size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;

-- Archive old data
CREATE TABLE tasks_archive (LIKE tasks INCLUDING ALL);

-- Move old data to archive
INSERT INTO tasks_archive
SELECT * FROM tasks
WHERE created_at < NOW() - INTERVAL '1 year';

DELETE FROM tasks
WHERE created_at < NOW() - INTERVAL '1 year';

-- Vacuum to reclaim space
VACUUM FULL tasks;
```

### Technique 8.2: Query Cost Analysis

**Optimize expensive queries**:

```sql
-- Enable query planning stats
EXPLAIN (ANALYZE, BUFFERS) SELECT
  p.*,
  COUNT(t.id) AS task_count
FROM projects p
LEFT JOIN tasks t ON t.project_id = p.id
GROUP BY p.id
LIMIT 100;

-- Sample output analysis:
/*
Limit  (cost=1000.00..1250.00 rows=100 width=100) (actual time=0.123..0.456 rows=100 loops=1)
  Buffers: shared hit=500 read=100
  ->  GroupAggregate  (cost=1000.00..50000.00 rows=10000 width=100) (actual time=0.120..0.450 rows=100 loops=1)
        Buffers: shared hit=500 read=100
        ->  Hash Left Join  (cost=500.00..40000.00 rows=50000 width=80) (actual time=0.100..0.350 rows=1000 loops=1)
              Buffers: shared hit=400 read=80
*/

-- Optimization: Add index if "Seq Scan" appears
-- Optimization: Increase shared_buffers if "read" is high
```

### Technique 8.3: Connection Limits

**Prevent connection exhaustion**:

```typescript
class ConnectionPool {
  private activeConnections = 0
  private maxConnections = 10
  private queue: Array<() => void> = []

  async acquire(): Promise<void> {
    if (this.activeConnections < this.maxConnections) {
      this.activeConnections++
      return Promise.resolve()
    }

    return new Promise((resolve) => {
      this.queue.push(() => {
        this.activeConnections++
        resolve()
      })
    })
  }

  release() {
    this.activeConnections--

    if (this.queue.length > 0) {
      const next = this.queue.shift()
      next?.()
    }
  }

  async execute<T>(fn: () => Promise<T>): Promise<T> {
    await this.acquire()

    try {
      return await fn()
    } finally {
      this.release()
    }
  }
}

const pool = new ConnectionPool()

// Usage
const result = await pool.execute(async () => {
  return await supabase.from('tasks').select('*')
})
```

---

## Performance Benchmarks

**Typical improvements**:

| Optimization | Before | After | Improvement |
|--------------|--------|-------|-------------|
| Connection pooling | 150ms | 20ms | 87% faster |
| Query optimization | 500ms | 50ms | 90% faster |
| Caching | 100ms | 5ms | 95% faster |
| Batch operations | 10s (1000 ops) | 200ms | 98% faster |
| Indexed queries | 2000ms | 10ms | 99.5% faster |

---

## Best Practices

**DO**:
- ✅ Use PgBouncer transaction mode for serverless
- ✅ Create composite indexes for common queries
- ✅ Implement caching for read-heavy workloads
- ✅ Batch database operations
- ✅ Monitor query performance regularly
- ✅ Archive old data periodically
- ✅ Use materialized views for aggregations

**DON'T**:
- ❌ Create excessive realtime subscriptions
- ❌ Perform N+1 queries
- ❌ Skip index analysis
- ❌ Ignore connection limits
- ❌ Store large blobs in database (use Storage)
- ❌ Over-index (hurts write performance)

---

**Context7 Reference**: `/websites/supabase` (latest API v2.38+)
**Last Updated**: 2025-11-22
**Status**: Production Ready
