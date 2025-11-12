---
name: "moai-domain-database"
version: "4.0.0"
created: 2025-11-12
updated: 2025-11-12
status: stable
tier: domain
description: "Enterprise-grade database architecture with PostgreSQL 17, MySQL 8.4 LTS, MongoDB 8.0, Redis 7.4, and modern ORM patterns (SQLAlchemy 2.0, Prisma 5); covers query optimization, connection pooling, indexing strategies, caching patterns, and distributed data management. Enhanced with Context7 MCP for up-to-date stable version documentation."
allowed-tools: "Read, Bash, WebSearch, WebFetch, mcp__context7__resolve-library-id, mcp__context7__get-library-docs"
primary-agent: "database-expert"
secondary-agents: [alfred, qa-validator, doc-syncer]
keywords: [database, postgresql, mysql, mongodb, redis, sqlalchemy, prisma, query-optimization, caching, indexing]
tags: [domain-expert, database]
orchestration: 
can_resume: true
typical_chain_position: "middle"
depends_on: []
---

# moai-domain-database

**Enterprise Database Architecture (v4.0)**

> **Primary Agent**: database-expert  
> **Secondary Agents**: alfred, qa-validator, doc-syncer  
> **Version**: 4.0.0 (Stable 2025-11)  
> **Stable Stack**: PostgreSQL 17, MySQL 8.4 LTS, MongoDB 8.0, Redis 7.4, SQLAlchemy 2.0, Prisma 5

---

## 📖 Progressive Disclosure

### Level 1: Quick Reference (Core Concepts)

**Purpose**: Production-ready database architecture with 2025-11 stable versions. Covers RDBMS (PostgreSQL 17, MySQL 8.4), NoSQL (MongoDB 8.0), caching (Redis 7.4), and modern ORMs (SQLAlchemy 2.0, Prisma 5).

**When to Use:**
- ✅ Designing database schemas with optimal indexing
- ✅ Implementing connection pooling and query optimization
- ✅ Building caching strategies (cache-aside, write-through)
- ✅ Handling N+1 queries and database performance issues
- ✅ Migrating databases with zero downtime

**Quick Start Pattern:**

\`\`\`python
# PostgreSQL 17 with SQLAlchemy 2.0 async connection pooling
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker

# Create async engine with connection pooling
engine = create_async_engine(
    "postgresql+asyncpg://user:pass@localhost/db",
    pool_size=20,
    max_overflow=10,
    pool_pre_ping=True,  # Auto-reconnect on connection loss
    echo=False
)

# Session factory with async context manager
AsyncSessionLocal = sessionmaker(
    engine, class_=AsyncSession, expire_on_commit=False
)

async def get_user(user_id: int):
    async with AsyncSessionLocal() as session:
        result = await session.execute(
            select(User).where(User.id == user_id)
        )
        return result.scalar_one_or_none()
\`\`\`

---

### Level 2: Practical Implementation (Common Patterns)

#### Pattern 1: PostgreSQL 17 Advanced Indexing (GiST, GIN, BRIN)

**Context7 Reference**: [PostgreSQL 17.5 Official Docs](https://www.postgresql.org/docs/current/)

\`\`\`sql
-- PostgreSQL 17: Create optimized indexes for different use cases
-- B-Tree Index (default, best for equality/range queries)
CREATE INDEX CONCURRENTLY idx_users_email ON users (email);

-- GIN Index (Generalized Inverted Index for full-text search, JSONB)
CREATE INDEX CONCURRENTLY idx_products_search 
ON products USING GIN (to_tsvector('english', name || ' ' || description));

-- GiST Index (Geometric Search Tree for spatial/range data)
CREATE INDEX CONCURRENTLY idx_locations_point 
ON locations USING GIST (coordinates);

-- BRIN Index (Block Range Index for time-series/append-only data)
CREATE INDEX CONCURRENTLY idx_logs_timestamp 
ON logs USING BRIN (created_at) WITH (pages_per_range = 128);

-- Partial Index (index subset of rows for specific queries)
CREATE INDEX CONCURRENTLY idx_active_orders 
ON orders (user_id, created_at) 
WHERE status = 'active' AND deleted_at IS NULL;

-- Covering Index (include extra columns for index-only scans)
CREATE INDEX CONCURRENTLY idx_orders_user_covering 
ON orders (user_id) INCLUDE (total_amount, status);
\`\`\`

**Performance Impact**:
- B-Tree: 10-100x faster for equality/range queries
- GIN: 50-500x faster for full-text search
- BRIN: 95% storage reduction for time-series data
- Partial: 70% index size reduction with same performance

---

#### Pattern 2: SQLAlchemy 2.0 Async Connection Pooling

**Context7 Reference**: [SQLAlchemy 2.0 Connection Pooling](https://docs.sqlalchemy.org/en/20/core/pooling.html)

\`\`\`python
# SQLAlchemy 2.0 production-grade async connection pooling
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker, selectinload
from sqlalchemy import select, event
from sqlalchemy.pool import NullPool, QueuePool
import logging

logger = logging.getLogger(__name__)

# Production engine with optimized pooling
engine = create_async_engine(
    "postgresql+asyncpg://user:pass@localhost:5432/db",
    poolclass=QueuePool,
    pool_size=20,              # Core connections
    max_overflow=10,           # Additional connections under load
    pool_timeout=30,           # Wait timeout for connection
    pool_recycle=3600,         # Recycle connections every hour
    pool_pre_ping=True,        # Test connection before use
    echo_pool=False,           # Disable pool logging in production
    connect_args={
        "server_settings": {
            "application_name": "myapp",
            "jit": "off"       # Disable JIT for predictable performance
        },
        "command_timeout": 60,
        "timeout": 10
    }
)

# Session factory with async context manager
AsyncSessionLocal = sessionmaker(
    bind=engine,
    class_=AsyncSession,
    expire_on_commit=False,    # Keep objects usable after commit
    autoflush=False,           # Manual flush control
    autocommit=False
)

# Connection pool monitoring event
@event.listens_for(engine.sync_engine, "connect")
def receive_connect(dbapi_conn, connection_record):
    logger.info(f"New DB connection established: {id(dbapi_conn)}")

@event.listens_for(engine.sync_engine, "checkout")
def receive_checkout(dbapi_conn, connection_record, connection_proxy):
    logger.debug(f"Connection checked out from pool: {id(dbapi_conn)}")

# Async session context manager
async def get_session() -> AsyncSession:
    async with AsyncSessionLocal() as session:
        try:
            yield session
            await session.commit()
        except Exception:
            await session.rollback()
            raise
        finally:
            await session.close()

# Example: Efficient query with eager loading (prevent N+1)
async def get_user_with_posts(user_id: int):
    async with AsyncSessionLocal() as session:
        result = await session.execute(
            select(User)
            .options(selectinload(User.posts))  # Eager load posts
            .where(User.id == user_id)
        )
        return result.scalar_one_or_none()
\`\`\`

**Best Practices**:
- Use \`pool_pre_ping=True\` for automatic reconnection
- Set \`pool_recycle\` to prevent stale connections
- Monitor pool metrics: size, overflow, checkouts, timeouts
- Use \`selectinload()\` or \`joinedload()\` to prevent N+1 queries

---

#### Pattern 3: Prisma 5 Query Optimization & Relations

**Context7 Reference**: [Prisma 5 Docs](https://www.prisma.io/docs)

\`\`\`typescript
// Prisma 5: Optimized query patterns with relations
import { PrismaClient } from '@prisma/client'

const prisma = new PrismaClient({
  log: [
    { level: 'query', emit: 'event' },
    { level: 'error', emit: 'stdout' }
  ],
  datasources: {
    db: {
      url: process.env.DATABASE_URL
    }
  }
})

// Query optimization: Use select to fetch only needed fields
async function getOptimizedUser(userId: number) {
  return await prisma.user.findUnique({
    where: { id: userId },
    select: {
      id: true,
      email: true,
      name: true,
      // Exclude heavy fields like 'bio', 'avatar'
    }
  })
}

// Prevent N+1: Use include with nested relations
async function getUserWithPostsAndComments(userId: number) {
  return await prisma.user.findUnique({
    where: { id: userId },
    include: {
      posts: {
        include: {
          comments: {
            take: 10,  // Limit comments per post
            orderBy: { createdAt: 'desc' }
          }
        },
        where: { published: true },
        orderBy: { createdAt: 'desc' }
      }
    }
  })
}

// Batch queries: Reduce database round-trips
async function batchGetUsers(userIds: number[]) {
  return await prisma.user.findMany({
    where: {
      id: { in: userIds }
    }
  })
}

// Transactions: Ensure atomicity
async function createUserWithProfile(userData: any, profileData: any) {
  return await prisma.\$transaction(async (tx) => {
    const user = await tx.user.create({
      data: userData
    })
    
    const profile = await tx.profile.create({
      data: {
        ...profileData,
        userId: user.id
      }
    })
    
    return { user, profile }
  })
}

// Raw queries with type safety (for complex queries)
async function getTopUsers(limit: number = 10) {
  return await prisma.\$queryRaw<Array<{id: number, name: string, postCount: bigint}>>\`
    SELECT u.id, u.name, COUNT(p.id)::bigint as "postCount"
    FROM "User" u
    LEFT JOIN "Post" p ON u.id = p."authorId"
    GROUP BY u.id
    ORDER BY "postCount" DESC
    LIMIT \${limit}
  \`
}

// Connection pool monitoring
prisma.\$on('query', (e) => {
  console.log(\`Query: \${e.query}\`)
  console.log(\`Duration: \${e.duration}ms\`)
})
\`\`\`

**Performance Tips**:
- Use \`select\` to fetch only required fields (reduces payload size)
- Use \`include\` strategically to prevent N+1 queries
- Leverage \`findMany\` with \`in\` operator for batch queries
- Use transactions for multi-step operations
- Monitor slow queries with \`prisma.\$on('query')\` event

---

#### Pattern 4: Redis 7.4 Caching Strategies

**Context7 Reference**: [Redis 7.4 Official Docs](https://redis.io/docs/latest/)

\`\`\`python
# Redis 7.4: Production caching patterns
import redis.asyncio as redis
from typing import Optional
import json
import hashlib

class RedisCache:
    def __init__(self, redis_url: str):
        self.client = redis.from_url(
            redis_url,
            encoding="utf-8",
            decode_responses=True,
            max_connections=50,
            socket_keepalive=True,
            socket_connect_timeout=5,
            retry_on_timeout=True
        )
    
    # Pattern 1: Cache-Aside (Lazy Loading)
    async def get_user_cache_aside(self, user_id: int):
        cache_key = f"user:{user_id}"
        
        # Try cache first
        cached = await self.client.get(cache_key)
        if cached:
            return json.loads(cached)
        
        # Cache miss: fetch from database
        user = await db.get_user(user_id)  # Database query
        
        # Store in cache with TTL
        await self.client.setex(
            cache_key,
            3600,  # 1 hour TTL
            json.dumps(user)
        )
        return user
    
    # Pattern 2: Write-Through (Update cache on write)
    async def update_user_write_through(self, user_id: int, data: dict):
        # Update database first
        updated_user = await db.update_user(user_id, data)
        
        # Update cache immediately
        cache_key = f"user:{user_id}"
        await self.client.setex(
            cache_key,
            3600,
            json.dumps(updated_user)
        )
        return updated_user
    
    # Pattern 3: Write-Behind (Async cache invalidation)
    async def invalidate_user_cache(self, user_id: int):
        cache_key = f"user:{user_id}"
        await self.client.delete(cache_key)
    
    # Pattern 4: Cache Stampede Prevention (using locking)
    async def get_with_lock(self, key: str, fetch_fn, ttl: int = 3600):
        # Try cache
        cached = await self.client.get(key)
        if cached:
            return json.loads(cached)
        
        # Acquire lock to prevent stampede
        lock_key = f"lock:{key}"
        lock = await self.client.set(lock_key, "1", nx=True, ex=10)
        
        if lock:
            try:
                # Fetch fresh data
                data = await fetch_fn()
                await self.client.setex(key, ttl, json.dumps(data))
                return data
            finally:
                await self.client.delete(lock_key)
        else:
            # Wait for lock holder to populate cache
            await asyncio.sleep(0.1)
            return await self.get_with_lock(key, fetch_fn, ttl)
    
    # Pattern 5: Multi-level caching (L1: local, L2: Redis)
    async def get_multi_level(self, key: str, l1_cache: dict, ttl: int = 300):
        # L1: In-memory cache
        if key in l1_cache:
            return l1_cache[key]
        
        # L2: Redis cache
        cached = await self.client.get(key)
        if cached:
            data = json.loads(cached)
            l1_cache[key] = data  # Populate L1
            return data
        
        # L3: Database
        data = await db.fetch(key)
        await self.client.setex(key, ttl, json.dumps(data))
        l1_cache[key] = data
        return data

# Usage example
cache = RedisCache("redis://localhost:6379/0")

async def get_user(user_id: int):
    return await cache.get_user_cache_aside(user_id)
\`\`\`

**Cache Strategy Selection**:
- **Cache-Aside**: Read-heavy workloads, tolerate stale data
- **Write-Through**: Data consistency critical, moderate write load
- **Write-Behind**: High write throughput, tolerate eventual consistency
- **Cache Stampede Prevention**: High-traffic keys, expensive computations

---

#### Pattern 5: MongoDB 8.0 Aggregation Pipeline Optimization

**Context7 Reference**: [MongoDB 8.0 Docs](https://www.mongodb.com/docs/)

\`\`\`javascript
// MongoDB 8.0: Optimized aggregation pipeline patterns
const { MongoClient } = require('mongodb')

const client = new MongoClient('mongodb://localhost:27017', {
  maxPoolSize: 50,
  minPoolSize: 10,
  serverSelectionTimeoutMS: 5000,
  socketTimeoutMS: 45000
})

// Pattern 1: Efficient \$lookup with pipeline
async function getUsersWithRecentPosts() {
  const db = client.db('myapp')
  
  return await db.collection('users').aggregate([
    // Stage 1: Match active users only
    { \$match: { status: 'active' } },
    
    // Stage 2: Lookup with optimized pipeline
    {
      \$lookup: {
        from: 'posts',
        let: { userId: '\$_id' },
        pipeline: [
          { \$match: {
            \$expr: { \$eq: ['\$authorId', '\$\$userId'] },
            published: true
          }},
          { \$sort: { createdAt: -1 } },
          { \$limit: 5 },
          { \$project: { title: 1, createdAt: 1 } }
        ],
        as: 'recentPosts'
      }
    },
    
    // Stage 3: Project only needed fields
    {
      \$project: {
        _id: 1,
        name: 1,
        email: 1,
        recentPosts: 1
      }
    }
  ]).toArray()
}

// Pattern 2: \$facet for multiple aggregations in single query
async function getDashboardStats() {
  const db = client.db('myapp')
  
  return await db.collection('orders').aggregate([
    {
      \$facet: {
        totalRevenue: [
          { \$group: { _id: null, total: { \$sum: '\$amount' } } }
        ],
        topProducts: [
          { \$unwind: '\$items' },
          { \$group: { 
            _id: '\$items.productId',
            count: { \$sum: '\$items.quantity' }
          }},
          { \$sort: { count: -1 } },
          { \$limit: 10 }
        ],
        revenueByMonth: [
          { \$group: {
            _id: { \$dateToString: { format: '%Y-%m', date: '\$createdAt' } },
            revenue: { \$sum: '\$amount' }
          }},
          { \$sort: { _id: 1 } }
        ]
      }
    }
  ]).toArray()
}

// Pattern 3: Optimize with indexes (compound index)
async function createOptimizedIndexes() {
  const db = client.db('myapp')
  
  // Compound index for common query pattern
  await db.collection('orders').createIndex(
    { userId: 1, status: 1, createdAt: -1 },
    { background: true, name: 'idx_user_status_date' }
  )
  
  // Text index for full-text search
  await db.collection('products').createIndex(
    { name: 'text', description: 'text' },
    { name: 'idx_product_text_search' }
  )
}

// Pattern 4: Use \$merge for incremental aggregations
async function updateDailySalesReport(date) {
  const db = client.db('myapp')
  
  await db.collection('orders').aggregate([
    {
      \$match: {
        createdAt: {
          \$gte: new Date(date),
          \$lt: new Date(date.getTime() + 86400000)
        }
      }
    },
    {
      \$group: {
        _id: {
          date: { \$dateToString: { format: '%Y-%m-%d', date: '\$createdAt' } },
          productId: '\$productId'
        },
        totalSales: { \$sum: '\$amount' },
        orderCount: { \$sum: 1 }
      }
    },
    {
      \$merge: {
        into: 'daily_sales_reports',
        on: '_id',
        whenMatched: 'replace',
        whenNotMatched: 'insert'
      }
    }
  ])
}
\`\`\`

**MongoDB 8.0 Performance Tips**:
- Use \`\$match\` early to reduce pipeline data volume
- Create compound indexes matching query patterns
- Use \`\$lookup\` pipeline syntax (more efficient than basic)
- Leverage \`\$facet\` to avoid multiple queries
- Use \`\$merge\` for incremental materialized views

---

#### Pattern 6: Query Optimization with EXPLAIN ANALYZE

\`\`\`sql
-- PostgreSQL 17: Analyze query performance
EXPLAIN (ANALYZE, BUFFERS, VERBOSE, COSTS) 
SELECT u.id, u.name, COUNT(p.id) as post_count
FROM users u
LEFT JOIN posts p ON u.id = p.author_id
WHERE u.created_at > '2024-01-01'
  AND u.status = 'active'
GROUP BY u.id, u.name
ORDER BY post_count DESC
LIMIT 10;

-- Expected output analysis:
-- Look for:
-- 1. Seq Scan → create index
-- 2. High "cost" values → optimize query
-- 3. "Buffers: shared hit" vs "read" → cache hit ratio
-- 4. "Rows Removed by Filter" → improve WHERE clause
\`\`\`

**Optimization Checklist**:
- ✅ Index on \`users.created_at\` and \`users.status\`
- ✅ Index on \`posts.author_id\` (foreign key)
- ✅ Use covering index to avoid heap lookups
- ✅ Check \`shared_buffers\` and \`effective_cache_size\` settings

---

#### Pattern 7: Database Migration (Zero-Downtime)

\`\`\`python
# Alembic migration for adding column without downtime
from alembic import op
import sqlalchemy as sa

def upgrade():
    # Step 1: Add column as nullable
    op.add_column('users', 
        sa.Column('verified', sa.Boolean(), nullable=True)
    )
    
    # Step 2: Populate existing rows in batches
    connection = op.get_bind()
    connection.execute(
        sa.text("UPDATE users SET verified = false WHERE verified IS NULL")
    )
    
    # Step 3: Make column NOT NULL (after all rows populated)
    op.alter_column('users', 'verified',
        existing_type=sa.Boolean(),
        nullable=False,
        server_default=sa.false()
    )

def downgrade():
    op.drop_column('users', 'verified')
\`\`\`

**Zero-Downtime Migration Steps**:
1. Add column as nullable with default
2. Deploy code that uses new column
3. Backfill existing rows in batches
4. Add NOT NULL constraint
5. Remove default (if needed)

---

#### Pattern 8: N+1 Query Prevention

\`\`\`python
# SQLAlchemy 2.0: Prevent N+1 with selectinload
from sqlalchemy.orm import selectinload

# BAD: N+1 query problem
users = await session.execute(select(User))
for user in users.scalars():
    print(user.posts)  # Each iteration triggers separate query!

# GOOD: Eager loading with selectinload
result = await session.execute(
    select(User).options(selectinload(User.posts))
)
users = result.scalars().all()
for user in users:
    print(user.posts)  # Already loaded, no extra queries

# GOOD: joinedload for one-to-one relations
result = await session.execute(
    select(User).options(joinedload(User.profile))
)
\`\`\`

**Eager Loading Strategies**:
- \`selectinload()\`: Separate query with IN clause (good for one-to-many)
- \`joinedload()\`: Single query with LEFT JOIN (good for one-to-one)
- \`subqueryload()\`: Subquery for complex relations

---

#### Pattern 9: Transaction Management & Isolation Levels

\`\`\`python
# SQLAlchemy 2.0: Transaction isolation levels
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker

engine = create_engine(
    "postgresql://localhost/db",
    isolation_level="REPEATABLE READ"  # or "READ COMMITTED", "SERIALIZABLE"
)

# Explicit transaction control
async with AsyncSessionLocal() as session:
    async with session.begin():
        # All operations in this block are in single transaction
        user = await session.get(User, user_id)
        user.balance -= 100
        
        order = Order(user_id=user.id, amount=100)
        session.add(order)
        
        # Commit happens automatically at end of block
        # Rollback happens automatically on exception

# Manual transaction control
async with AsyncSessionLocal() as session:
    try:
        user = await session.get(User, user_id)
        user.balance -= 100
        await session.commit()
    except Exception:
        await session.rollback()
        raise
\`\`\`

**Isolation Level Selection**:
- **READ COMMITTED**: Default, prevents dirty reads
- **REPEATABLE READ**: Prevents non-repeatable reads
- **SERIALIZABLE**: Full isolation, highest consistency (slowest)

---

#### Pattern 10: Database Sharding & Partitioning

\`\`\`sql
-- PostgreSQL 17: Declarative partitioning (time-based)
CREATE TABLE logs (
    id BIGSERIAL,
    user_id INTEGER,
    action TEXT,
    created_at TIMESTAMP NOT NULL,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Create partitions for each month
CREATE TABLE logs_2024_01 PARTITION OF logs
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE logs_2024_02 PARTITION OF logs
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

CREATE TABLE logs_2024_03 PARTITION OF logs
    FOR VALUES FROM ('2024-03-01') TO ('2024-04-01');

-- Indexes on partitions are created automatically
CREATE INDEX idx_logs_user ON logs (user_id, created_at);

-- Query automatically routes to correct partition
SELECT * FROM logs 
WHERE created_at >= '2024-02-15' 
  AND created_at < '2024-02-20';
\`\`\`

**Partitioning Benefits**:
- 90% faster queries (pruning irrelevant partitions)
- Easier data archival (drop old partitions)
- Parallel query execution across partitions
- Improved maintenance operations (VACUUM, ANALYZE)

---

### Level 3: Advanced Patterns (Expert Reference)

#### Advanced Topic 1: Connection Pool Tuning

**Formula for pool sizing**:
\`\`\`
pool_size = (core_count * 2) + effective_spindle_count
max_overflow = pool_size * 0.5

For cloud databases:
pool_size = min(20, max_connections / num_app_instances)
\`\`\`

**Monitoring metrics**:
- Pool checkouts/sec
- Pool overflow count
- Connection wait time
- Connection lifetime

---

#### Advanced Topic 2: Query Performance Profiling

\`\`\`python
# SQLAlchemy 2.0: Query profiling with events
from sqlalchemy import event
import time

@event.listens_for(engine, "before_cursor_execute")
def receive_before_cursor_execute(conn, cursor, statement, params, context, executemany):
    context._query_start_time = time.time()

@event.listens_for(engine, "after_cursor_execute")
def receive_after_cursor_execute(conn, cursor, statement, params, context, executemany):
    duration = time.time() - context._query_start_time
    if duration > 1.0:  # Log slow queries (> 1 second)
        logger.warning(f"Slow query ({duration:.2f}s): {statement}")
\`\`\`

---

## 🎯 Best Practices Checklist

**Must-Have:**
- ✅ Use connection pooling with \`pool_pre_ping=True\`
- ✅ Create indexes for all foreign keys and WHERE clauses
- ✅ Always use parameterized queries (prevent SQL injection)
- ✅ Monitor slow query logs (>100ms threshold)
- ✅ Use database migrations for schema changes

**Recommended:**
- ✅ Implement caching layer (Redis) for read-heavy workloads
- ✅ Use eager loading (selectinload/joinedload) to prevent N+1
- ✅ Set up read replicas for read-heavy applications
- ✅ Use EXPLAIN ANALYZE for query optimization
- ✅ Implement database health checks in production

**Security:**
- 🔒 Never store plaintext credentials (use environment variables)
- 🔒 Use SSL/TLS for database connections
- 🔒 Implement row-level security for multi-tenant apps
- 🔒 Rotate database credentials regularly
- 🔒 Audit database access logs for suspicious activity

---

## 🔗 Context7 MCP Integration

**When to Use Context7 for This Skill:**

This skill benefits from Context7 when:
- Verifying latest stable versions (PostgreSQL 17, MySQL 8.4 LTS, MongoDB 8.0, Redis 7.4)
- Checking ORM best practices (SQLAlchemy 2.0, Prisma 5)
- Validating query optimization techniques
- Confirming connection pooling configurations

**Example Usage:**

\`\`\`python
# Fetch latest PostgreSQL 17 documentation
from moai_adk.integrations import Context7Helper

helper = Context7Helper()
docs = await helper.get_docs(
    library_id="/websites/postgresql-current",
    topic="indexing performance optimization",
    tokens=3000
)
\`\`\`

**Relevant Libraries:**

| Library | Context7 ID | Use Case | Version |
|---------|-------------|----------|---------|
| PostgreSQL | \`/websites/postgresql-current\` | RDBMS indexing, query optimization | 17.5 (Stable) |
| SQLAlchemy | \`/websites/sqlalchemy_en_20\` | Python ORM, async patterns | 2.0 (Stable) |
| Prisma | \`/prisma/prisma\` | TypeScript ORM, type safety | 5.x (Stable) |
| Redis | \`/websites/redis_io\` | Caching strategies, data structures | 7.4 (Stable) |
| MongoDB | \`/mongodb/mongo\` | NoSQL aggregation, indexing | 8.0 (Stable) |

---

## 📊 Decision Tree

**When to use moai-domain-database:**

\`\`\`
Start
  ├─ Need database design?
  │   ├─ YES → Use Pattern 1 (PostgreSQL Indexing)
  │   └─ NO → Continue
  ├─ Performance issues?
  │   ├─ YES → Use Pattern 6 (EXPLAIN ANALYZE)
  │   └─ NO → Continue
  ├─ N+1 query problem?
  │   ├─ YES → Use Pattern 8 (Eager Loading)
  │   └─ NO → Continue
  ├─ Need caching?
  │   ├─ YES → Use Pattern 4 (Redis Strategies)
  │   └─ NO → Continue
  └─ Complex aggregations?
      ├─ YES → Use Pattern 5 (MongoDB Pipeline)
      └─ NO → Review Level 1 patterns
\`\`\`

---

## 🔄 Integration with Other Skills

**Prerequisite Skills:**
- Skill("moai-foundation-python") – Python async/await fundamentals
- Skill("moai-foundation-typescript") – TypeScript type system for Prisma

**Complementary Skills:**
- Skill("moai-domain-backend") – API design with database integration
- Skill("moai-domain-testing") – Database testing strategies

**Next Steps:**
- Skill("moai-domain-devops") – Database deployment and monitoring
- Skill("moai-domain-security") – Database security hardening

---

## 📚 Official References

### Stable Version Documentation (2025-11)

**PostgreSQL 17**:
- [Official Docs](https://www.postgresql.org/docs/17/)
- [Performance Tips](https://www.postgresql.org/docs/17/performance-tips.html)
- [Indexing](https://www.postgresql.org/docs/17/indexes.html)

**SQLAlchemy 2.0**:
- [Core Documentation](https://docs.sqlalchemy.org/en/20/)
- [Async Engine](https://docs.sqlalchemy.org/en/20/orm/extensions/asyncio.html)
- [Connection Pooling](https://docs.sqlalchemy.org/en/20/core/pooling.html)

**Prisma 5**:
- [Official Docs](https://www.prisma.io/docs)
- [Query Optimization](https://www.prisma.io/docs/guides/performance-and-optimization)
- [Relations](https://www.prisma.io/docs/concepts/components/prisma-schema/relations)

**Redis 7.4**:
- [Official Docs](https://redis.io/docs/latest/)
- [Caching Patterns](https://redis.io/docs/latest/develop/reference/client-side-caching/)
- [Best Practices](https://redis.io/docs/latest/develop/whats-new/)

**MongoDB 8.0**:
- [Official Docs](https://www.mongodb.com/docs/manual/)
- [Aggregation](https://www.mongodb.com/docs/manual/aggregation/)
- [Indexing](https://www.mongodb.com/docs/manual/indexes/)

---

## 📈 Version History

**v4.0.0** (2025-11-12)
- ✨ Updated to 2025-11 stable versions (PostgreSQL 17, MySQL 8.4 LTS, MongoDB 8.0, Redis 7.4)
- ✨ Added SQLAlchemy 2.0 async patterns with connection pooling
- ✨ Added Prisma 5 query optimization patterns
- ✨ Enhanced with Context7 MCP stable version documentation
- ✨ 10+ production-ready database patterns
- ✨ Optimized to 900 lines, 30 KB (within Enterprise metrics)
- ✨ Added zero-downtime migration patterns
- ✨ Added database sharding and partitioning examples

---

**Generated with**: MoAI-ADK Skill Factory v4.0  
**Last Updated**: 2025-11-12  
**Maintained by**: Primary Agent (database-expert)  
**Stable Stack**: PostgreSQL 17, MySQL 8.4 LTS, MongoDB 8.0, Redis 7.4, SQLAlchemy 2.0, Prisma 5
