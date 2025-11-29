---
name: moai-domain-database
description: Database specialist covering PostgreSQL, MongoDB, Redis, and advanced data patterns for modern applications
version: 1.0.0
category: domain
tags:
  - database
  - postgresql
  - mongodb
  - redis
  - optimization
  - architecture
updated: 2025-11-30
status: active
author: MoAI-ADK Team
---

# Database Development Specialist

## Quick Reference (30 seconds)

**Modern Database Architecture** - Comprehensive database patterns covering PostgreSQL, MongoDB, Redis, and advanced data modeling for scalable applications.

**Core Capabilities**:
- 🐘 **PostgreSQL**: Advanced SQL, optimization, extensions, indexing strategies
- 🍃 **MongoDB**: Document modeling, aggregation pipelines, sharding
- ⚡ **Redis**: Caching, pub/sub, data structures, performance optimization
- 📊 **Data Modeling**: Schema design, migrations, data integrity
- 🚀 **Performance**: Query optimization, indexing, connection pooling

**When to Use**:
- Database schema design and optimization
- Complex query implementation and tuning
- Data migration and transformation
- Performance optimization for databases
- Multi-database architecture planning

---

## Implementation Guide

### PostgreSQL Advanced Patterns

**Advanced Schema Design**:
```sql
-- User table with proper constraints and indexes
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_login_at TIMESTAMP WITH TIME ZONE
);

-- Optimized indexes
CREATE INDEX idx_users_email ON users (email) WHERE email_verified = TRUE;
CREATE INDEX idx_users_created_at ON users (created_at DESC);
CREATE INDEX idx_users_username_trgm ON users USING gin (username gin_trgm_ops);

-- User profiles with JSONB
CREATE TABLE user_profiles (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    display_name VARCHAR(100),
    bio TEXT,
    avatar_url VARCHAR(500),
    preferences JSONB DEFAULT '{}',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- GIN index for JSONB queries
CREATE INDEX idx_user_profiles_preferences ON user_profiles USING gin (preferences);
```

**Advanced Query Patterns**:
```sql
-- Efficient pagination with cursor-based approach
WITH ranked_users AS (
    SELECT
        u.id,
        u.username,
        u.email,
        up.display_name,
        ROW_NUMBER() OVER (ORDER BY u.created_at DESC) as rn
    FROM users u
    LEFT JOIN user_profiles up ON u.id = up.user_id
    WHERE u.created_at < :cursor OR :cursor IS NULL
)
SELECT * FROM ranked_users
WHERE rn > :offset
ORDER BY created_at DESC
LIMIT :limit;

-- Full-text search with trigrams
SELECT
    u.id,
    u.username,
    up.display_name,
    similarity(u.username, :query) as username_sim,
    ts_rank_cd(search_vector, plainto_tsquery('english', :query)) as search_rank
FROM users u
LEFT JOIN user_profiles up ON u.id = up.user_id,
     to_tsvector('english', u.username || ' ' || COALESCE(up.display_name, '')) search_vector
WHERE u.username % :query
   OR plainto_tsquery('english', :query) @@ search_vector
ORDER BY username_sim DESC, search_rank DESC
LIMIT 20;

-- Window functions for analytics
SELECT
    DATE_TRUNC('month', created_at) as month,
    COUNT(*) as new_users,
    SUM(COUNT(*)) OVER (ORDER BY DATE_TRUNC('month', created_at)) as cumulative_users,
    COUNT(*) - LAG(COUNT(*)) OVER (ORDER BY DATE_TRUNC('month', created_at)) as month_over_month_change
FROM users
WHERE created_at >= NOW() - INTERVAL '1 year'
GROUP BY DATE_TRUNC('month', created_at)
ORDER BY month;
```

**Performance Optimization**:
```sql
-- Partitioning for time-series data
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    event_type VARCHAR(50) NOT NULL,
    data JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
) PARTITION BY RANGE (created_at);

-- Monthly partitions
CREATE TABLE events_2024_01 PARTITION OF events
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

CREATE TABLE events_2024_02 PARTITION OF events
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');

-- Materialized views for analytics
CREATE MATERIALIZED VIEW user_stats AS
SELECT
    u.id,
    u.email,
    u.created_at,
    COUNT(DISTINCT e.id) as event_count,
    MAX(e.created_at) as last_event_at,
    COUNT(DISTINCT CASE WHEN e.event_type = 'login' THEN e.id END) as login_count
FROM users u
LEFT JOIN events e ON u.id = e.user_id
GROUP BY u.id, u.email, u.created_at;

-- Refresh materialized view
CREATE OR REPLACE FUNCTION refresh_user_stats()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY user_stats;
END;
$$ LANGUAGE plpgsql;
```

### MongoDB Advanced Patterns

**Document Modeling**:
```javascript
// User schema with embedded and referenced patterns
const userSchema = new Schema({
  _id: ObjectId,
  username: { type: String, required: true, unique: true },
  email: { type: String, required: true, unique: true },
  profile: {
    displayName: String,
    bio: String,
    avatar: String,
    preferences: {
      theme: { type: String, enum: ['light', 'dark'], default: 'light' },
      language: { type: String, default: 'en' },
      notifications: {
        email: { type: Boolean, default: true },
        push: { type: Boolean, default: false }
      }
    }
  },
  security: {
    passwordHash: String,
    lastLoginAt: Date,
    failedLoginAttempts: { type: Number, default: 0 },
    lockedUntil: Date
  },
  activity: {
    lastSeenAt: Date,
    loginCount: { type: Number, default: 0 }
  }
}, {
  timestamps: true,
  // Optimized indexes
  index: [
    { username: 1 },
    { email: 1 },
    { 'security.lastLoginAt': -1 },
    { 'activity.lastSeenAt': -1 }
  ]
});

// Post schema with comments embedded for performance
const postSchema = new Schema({
  _id: ObjectId,
  authorId: { type: ObjectId, ref: 'User', required: true },
  title: { type: String, required: true },
  content: { type: String, required: true },
  tags: [String],
  metadata: {
    viewCount: { type: Number, default: 0 },
    likeCount: { type: Number, default: 0 },
    commentCount: { type: Number, default: 0 }
  },
  // Embed recent comments for performance
  recentComments: [{
    _id: ObjectId,
    authorId: ObjectId,
    authorName: String,
    content: String,
    createdAt: { type: Date, default: Date.now }
  }],
  status: { type: String, enum: ['draft', 'published', 'archived'], default: 'draft' }
}, {
  timestamps: true,
  index: [
    { authorId: 1, createdAt: -1 },
    { status: 1, createdAt: -1 },
    { tags: 1 },
    { 'metadata.viewCount': -1 }
  ]
});
```

**Advanced Aggregation Pipelines**:
```javascript
// User analytics with complex aggregation
const getUserAnalytics = async (userId, timeRange) => {
  return await User.aggregate([
    // Match specific user
    { $match: { _id: userId } },

    // Lookup posts and engagement
    {
      $lookup: {
        from: 'posts',
        localField: '_id',
        foreignField: 'authorId',
        as: 'posts'
      }
    },

    // Unwind posts for processing
    { $unwind: '$posts' },

    // Filter by time range
    {
      $match: {
        'posts.createdAt': {
          $gte: timeRange.start,
          $lte: timeRange.end
        }
      }
    },

    // Group by user and calculate metrics
    {
      $group: {
        _id: '$_id',
        username: { $first: '$username' },
        totalPosts: { $sum: 1 },
        totalViews: { $sum: '$posts.metadata.viewCount' },
        totalLikes: { $sum: '$posts.metadata.likeCount' },
        avgViewsPerPost: { $avg: '$posts.metadata.viewCount' },
        tags: { $push: '$posts.tags' }
      }
    },

    // Flatten and count tags
    {
      $addFields: {
        allTags: { $reduce: {
          input: '$tags',
          initialValue: [],
          in: { $concatArrays: ['$$value', '$$this'] }
        }}
      }
    },

    {
      $addFields: {
        uniqueTags: { $setUnion: ['$allTags', []] },
        topTags: { $slice: [
          { $sortArray: { input: { $objectToArray: { $size: '$allTags' } }, sortBy: { count: -1 } } },
          5
        ]}
      }
    }
  ]);
};

// Time series aggregation for activity tracking
const getActivityTrends = async (userId, granularity = 'daily') => {
  const groupFormat = granularity === 'daily' ?
    { $dateToString: { format: '%Y-%m-%d', date: '$createdAt' } } :
    { $dateToString: { format: '%Y-%U', date: '$createdAt' } };

  return await Activity.aggregate([
    { $match: { userId } },
    {
      $group: {
        _id: groupFormat,
        totalEvents: { $sum: 1 },
        eventTypes: { $addToSet: '$eventType' },
        uniqueSessions: { $addToSet: '$sessionId' }
      }
    },
    {
      $addFields: {
        eventCount: { $size: '$eventTypes' },
        sessionCount: { $size: '$uniqueSessions' }
      }
    },
    { $sort: { _id: 1 } }
  ]);
};
```

### Redis Advanced Patterns

**Caching Strategies**:
```javascript
class CacheManager {
  constructor(redisClient) {
    this.redis = redisClient;
  }

  // Multi-layer caching with fallback
  async getWithFallback(key, fetchFunction, ttl = 3600) {
    // Try memory cache first
    const memoryCache = this.getMemoryCache(key);
    if (memoryCache) return memoryCache;

    // Try Redis cache
    const redisCache = await this.redis.get(key);
    if (redisCache) {
      const data = JSON.parse(redisCache);
      this.setMemoryCache(key, data, ttl / 10); // Shorter memory TTL
      return data;
    }

    // Fetch from source
    const data = await fetchFunction();

    // Set both caches
    await this.redis.setex(key, ttl, JSON.stringify(data));
    this.setMemoryCache(key, data, ttl / 10);

    return data;
  }

  // Write-through caching
  async setWithWriteThrough(key, data, fetchFunction, ttl = 3600) {
    // Update source first
    await fetchFunction(data);

    // Update caches
    const pipeline = this.redis.pipeline();
    pipeline.setex(key, ttl, JSON.stringify(data));

    // Invalidate related cache keys
    const relatedKeys = await this.getRelatedKeys(key);
    relatedKeys.forEach(relatedKey => {
      pipeline.del(relatedKey);
    });

    await pipeline.exec();
    this.setMemoryCache(key, data, ttl / 10);
  }

  // Rate limiting with sliding window
  async checkRateLimit(key, limit, windowMs) {
    const now = Date.now();
    const pipeline = this.redis.pipeline();

    // Remove old entries
    pipeline.zremrangebyscore(key, 0, now - windowMs);

    // Add current request
    pipeline.zadd(key, now, now);

    // Count current window requests
    pipeline.zcard(key);

    // Set expiration
    pipeline.expire(key, Math.ceil(windowMs / 1000));

    const results = await pipeline.exec();
    const currentCount = results[2][1];

    return {
      allowed: currentCount <= limit,
      count: currentCount,
      remaining: Math.max(0, limit - currentCount),
      resetTime: now + windowMs
    };
  }
}
```

**Advanced Data Structures**:
```javascript
class RedisDataManager {
  constructor(redisClient) {
    this.redis = redisClient;
  }

  // Leaderboard with time decay
  async addToTimeDecayLeaderboard(key, memberId, score, decayRate = 0.95) {
    const timestamp = Date.now();
    const decayedScore = score * Math.pow(decayRate, (Date.now() - timestamp) / (1000 * 60 * 60));

    await this.redis.zadd(key, decayedScore, memberId);

    // Remove old entries
    const weekAgo = Date.now() - (7 * 24 * 60 * 60 * 1000);
    await this.redis.zremrangebyscore(key, 0, weekAgo);
  }

  // Real-time analytics with HyperLogLog
  async trackUniqueVisitors(pageKey, userId) {
    // Track unique users with HyperLogLog
    await this.redis.pfadd(`${pageKey}:unique`, userId);

    // Track total visits with regular counter
    await this.redis.incr(`${pageKey}:total`);

    // Track user activity set for recent activity
    const activityKey = `${pageKey}:activity:${Math.floor(Date.now() / (60 * 1000))}`;
    await this.redis.sadd(activityKey, userId);
    await this.redis.expire(activityKey, 300); // 5 minutes
  }

  async getAnalytics(pageKey) {
    const pipeline = this.redis.pipeline();

    // Unique visitors estimate
    pipeline.pfcount(`${pageKey}:unique`);

    // Total page views
    pipeline.get(`${pageKey}:total`);

    // Recent active users (last 5 minutes)
    const now = Math.floor(Date.now() / (60 * 1000));
    for (let i = 0; i < 5; i++) {
      pipeline.scard(`${pageKey}:activity:${now - i}`);
    }

    const results = await pipeline.exec();

    const uniqueVisitors = results[0][1];
    const totalViews = parseInt(results[1][1]) || 0;
    const recentActivity = results.slice(2).map(r => r[1]).reduce((a, b) => a + b, 0);

    return {
      uniqueVisitors,
      totalViews,
      recentActiveUsers: recentActivity
    };
  }

  // Distributed locking with timeout and retry
  async acquireLock(key, ttl = 30000, retryCount = 3, retryDelay = 100) {
    const lockKey = `lock:${key}`;
    const lockValue = `${Date.now()}-${Math.random()}`;

    for (let attempt = 0; attempt < retryCount; attempt++) {
      const result = await this.redis.set(
        lockKey,
        lockValue,
        'PX', ttl,
        'NX'
      );

      if (result === 'OK') {
        return {
          acquired: true,
          lockValue,
          release: async () => await this.releaseLock(lockKey, lockValue)
        };
      }

      // Wait before retry
      if (attempt < retryCount - 1) {
        await new Promise(resolve => setTimeout(resolve, retryDelay));
      }
    }

    return { acquired: false };
  }

  async releaseLock(lockKey, lockValue) {
    const script = `
      if redis.call("get", KEYS[1]) == ARGV[1] then
        return redis.call("del", KEYS[1])
      else
        return 0
      end
    `;

    return await this.redis.eval(script, 1, lockKey, lockValue);
  }
}
```

---

## Advanced Patterns

### Database Connection Management

**PostgreSQL Connection Pooling**:
```python
import asyncpg
from asyncpg.pool import Pool

class DatabaseManager:
    def __init__(self, database_url: str):
        self.database_url = database_url
        self.pool: Pool = None

    async def initialize_pool(self):
        self.pool = await asyncpg.create_pool(
            self.database_url,
            min_size=10,
            max_size=20,
            command_timeout=60,
            server_settings={
                'application_name': 'moai_app',
                'jit': 'off'  # Disable JIT for better performance on simple queries
            }
        )

    async def execute_query(self, query: str, *args):
        async with self.pool.acquire() as connection:
            return await connection.fetch(query, *args)

    async def execute_transaction(self, queries: List[Tuple[str, tuple]]):
        async with self.pool.acquire() as connection:
            async with connection.transaction():
                results = []
                for query, params in queries:
                    result = await connection.fetch(query, *params)
                    results.append(result)
                return results
```

### Database Migration Management

**Version Controlled Migrations**:
```python
from pathlib import Path
import hashlib

class MigrationManager:
    def __init__(self, pool, migrations_dir: str):
        self.pool = pool
        self.migrations_dir = Path(migrations_dir)

    async def initialize_migration_table(self):
        await self.pool.execute("""
            CREATE TABLE IF NOT EXISTS schema_migrations (
                version VARCHAR(255) PRIMARY KEY,
                checksum VARCHAR(64) NOT NULL,
                executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
            )
        """)

    async def get_pending_migrations(self):
        # Get executed migrations
        executed = await self.pool.fetch(
            "SELECT version, checksum FROM schema_migrations ORDER BY version"
        )
        executed_map = {row['version']: row['checksum'] for row in executed}

        # Get available migration files
        migration_files = sorted(self.migrations_dir.glob("*.sql"))
        pending = []

        for file_path in migration_files:
            version = file_path.stem
            content = file_path.read_text()
            checksum = hashlib.sha256(content.encode()).hexdigest()

            if version not in executed_map:
                pending.append((version, content, checksum))
            elif executed_map[version] != checksum:
                raise ValueError(f"Migration {version} has been modified after execution")

        return pending

    async def execute_migration(self, version: str, content: str, checksum: str):
        async with self.pool.acquire() as conn:
            async with conn.transaction():
                await conn.execute(content)
                await conn.execute(
                    "INSERT INTO schema_migrations (version, checksum) VALUES ($1, $2)",
                    version, checksum
                )
```

---

## Works Well With

- **moai-domain-backend** - Backend integration patterns
- **moai-integration-mcp** - MCP server for database operations
- **moai-quality-security** - Database security and compliance
- **moai-system-universal** - Performance optimization strategies
- **moai-foundation-core** - Architectural principles

---

## Technology Stack

**Primary Technologies**:
- **PostgreSQL 16+**: Advanced SQL, JSONB, partitioning, extensions
- **MongoDB 7+**: Document modeling, aggregation, sharding
- **Redis 7+**: Caching, pub/sub, advanced data structures
- **Connection Libraries**: asyncpg, motor, redis-py, SQLAlchemy
- **Migration Tools**: Alembic, custom migration managers
- **Monitoring**: pgAdmin, MongoDB Compass, Redis Insight

**Performance Tools**:
- PostgreSQL EXPLAIN ANALYZE, pg_stat_statements
- MongoDB Compass performance analyzer
- Redis slow log and monitoring
- Connection pooling and query optimization

---

**Status**: Production Ready
**Last Updated**: 2025-11-30
**Maintained by**: MoAI-ADK Database Team