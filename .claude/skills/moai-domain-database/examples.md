# Database Domain Examples

Multishot prompting examples demonstrating practical database patterns covering PostgreSQL, MongoDB, Redis, and multi-database architectures.

---

## Example 1: PostgreSQL Advanced Query Optimization

**Scenario**: Optimize a slow query for user analytics with proper indexing.

**Input**:
```sql
-- Slow query: Full table scan
SELECT u.id, u.name, u.email, COUNT(o.id) as order_count, SUM(o.total) as total_spent
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.created_at > '2024-01-01'
  AND u.status = 'active'
GROUP BY u.id
ORDER BY total_spent DESC
LIMIT 100;

-- Query takes 15+ seconds on 1M users, 5M orders
```

**Output**:
```sql
-- Step 1: Create optimized indexes
CREATE INDEX CONCURRENTLY idx_users_status_created
ON users (status, created_at DESC)
WHERE status = 'active';

CREATE INDEX CONCURRENTLY idx_orders_user_total
ON orders (user_id, total)
INCLUDE (id);

-- Step 2: Use materialized view for frequent analytics
CREATE MATERIALIZED VIEW user_order_stats AS
SELECT
    u.id,
    u.name,
    u.email,
    COUNT(o.id) as order_count,
    COALESCE(SUM(o.total), 0) as total_spent,
    MAX(o.created_at) as last_order_date
FROM users u
LEFT JOIN orders o ON u.id = o.user_id
WHERE u.status = 'active'
GROUP BY u.id, u.name, u.email;

CREATE UNIQUE INDEX idx_user_order_stats_id ON user_order_stats(id);
CREATE INDEX idx_user_order_stats_total ON user_order_stats(total_spent DESC);

-- Step 3: Refresh strategy
REFRESH MATERIALIZED VIEW CONCURRENTLY user_order_stats;

-- Step 4: Optimized query (now < 100ms)
SELECT id, name, email, order_count, total_spent
FROM user_order_stats
WHERE total_spent > 0
ORDER BY total_spent DESC
LIMIT 100;
```

**Explanation**: The optimization combines partial indexes, materialized views, and covering indexes to reduce query time from 15 seconds to under 100ms.

---

## Example 2: MongoDB Aggregation Pipeline for Analytics

**Scenario**: Build an analytics pipeline for e-commerce order analysis.

**Input**:
```python
from motor.motor_asyncio import AsyncIOMotorClient
from datetime import datetime, timedelta

# Sample order document structure
order_doc = {
    "_id": "order_123",
    "user_id": "user_456",
    "status": "completed",
    "items": [
        {"product_id": "prod_1", "quantity": 2, "price": 29.99},
        {"product_id": "prod_2", "quantity": 1, "price": 49.99}
    ],
    "total": 109.97,
    "created_at": datetime.utcnow(),
    "shipping_adddess": {"city": "Seoul", "country": "KR"}
}
```

**Output**:
```python
from motor.motor_asyncio import AsyncIOMotorClient
from datetime import datetime, timedelta
from typing import List, Dict

class OrderAnalytics:
    def __init__(self, client: AsyncIOMotorClient):
        self.db = client.ecommerce
        self.orders = self.db.orders

    async def setup_indexes(self):
        """Create optimized indexes for analytics queries."""
        await self.orders.create_indexes([
            {"keys": [("created_at", -1)]},
            {"keys": [("user_id", 1), ("created_at", -1)]},
            {"keys": [("status", 1), ("created_at", -1)]},
            {"keys": [("shipping_adddess.country", 1), ("created_at", -1)]}
        ])

    async def get_daily_revenue(
        self,
        start_date: datetime,
        end_date: datetime
    ) -> List[Dict]:
        """Get daily revenue breakdown."""
        pipeline = [
            {
                "$match": {
                    "status": "completed",
                    "created_at": {
                        "$gte": start_date,
                        "$lte": end_date
                    }
                }
            },
            {
                "$group": {
                    "_id": {
                        "year": {"$year": "$created_at"},
                        "month": {"$month": "$created_at"},
                        "day": {"$dayOfMonth": "$created_at"}
                    },
                    "total_revenue": {"$sum": "$total"},
                    "order_count": {"$sum": 1},
                    "avg_order_value": {"$avg": "$total"}
                }
            },
            {
                "$sort": {"_id.year": 1, "_id.month": 1, "_id.day": 1}
            },
            {
                "$project": {
                    "_id": 0,
                    "date": {
                        "$dateFromParts": {
                            "year": "$_id.year",
                            "month": "$_id.month",
                            "day": "$_id.day"
                        }
                    },
                    "total_revenue": {"$round": ["$total_revenue", 2]},
                    "order_count": 1,
                    "avg_order_value": {"$round": ["$avg_order_value", 2]}
                }
            }
        ]
        return await self.orders.aggregate(pipeline).to_list(None)

    async def get_top_products(
        self,
        limit: int = 10,
        days: int = 30
    ) -> List[Dict]:
        """Get top selling products."""
        start_date = datetime.utcnow() - timedelta(days=days)

        pipeline = [
            {
                "$match": {
                    "status": "completed",
                    "created_at": {"$gte": start_date}
                }
            },
            {"$unwind": "$items"},
            {
                "$group": {
                    "_id": "$items.product_id",
                    "total_quantity": {"$sum": "$items.quantity"},
                    "total_revenue": {
                        "$sum": {
                            "$multiply": ["$items.quantity", "$items.price"]
                        }
                    },
                    "order_count": {"$sum": 1}
                }
            },
            {"$sort": {"total_revenue": -1}},
            {"$limit": limit},
            {
                "$lookup": {
                    "from": "products",
                    "localField": "_id",
                    "foreignField": "_id",
                    "as": "product_info"
                }
            },
            {
                "$project": {
                    "product_id": "$_id",
                    "product_name": {"$arrayElemAt": ["$product_info.name", 0]},
                    "total_quantity": 1,
                    "total_revenue": {"$round": ["$total_revenue", 2]},
                    "order_count": 1
                }
            }
        ]
        return await self.orders.aggregate(pipeline).to_list(None)

    async def get_customer_segments(self) -> List[Dict]:
        """Segment customers by order frequency and value."""
        pipeline = [
            {"$match": {"status": "completed"}},
            {
                "$group": {
                    "_id": "$user_id",
                    "total_orders": {"$sum": 1},
                    "total_spent": {"$sum": "$total"},
                    "avg_order_value": {"$avg": "$total"},
                    "first_order": {"$min": "$created_at"},
                    "last_order": {"$max": "$created_at"}
                }
            },
            {
                "$addFields": {
                    "segment": {
                        "$switch": {
                            "branches": [
                                {
                                    "case": {
                                        "$and": [
                                            {"$gte": ["$total_orders", 10]},
                                            {"$gte": ["$total_spent", 1000]}
                                        ]
                                    },
                                    "then": "VIP"
                                },
                                {
                                    "case": {
                                        "$and": [
                                            {"$gte": ["$total_orders", 5]},
                                            {"$gte": ["$total_spent", 500]}
                                        ]
                                    },
                                    "then": "Loyal"
                                },
                                {
                                    "case": {"$gte": ["$total_orders", 2]},
                                    "then": "Repeat"
                                }
                            ],
                            "default": "New"
                        }
                    }
                }
            },
            {
                "$group": {
                    "_id": "$segment",
                    "customer_count": {"$sum": 1},
                    "total_revenue": {"$sum": "$total_spent"},
                    "avg_orders": {"$avg": "$total_orders"},
                    "avg_spent": {"$avg": "$total_spent"}
                }
            },
            {"$sort": {"total_revenue": -1}}
        ]
        return await self.orders.aggregate(pipeline).to_list(None)
```

**Explanation**: This pattern demonstrates MongoDB aggregation pipelines for complex analytics with proper indexing, customer segmentation, and revenue analysis.

---

## Example 3: Redis Multi-Layer Caching Strategy

**Scenario**: Implement a sophisticated caching strategy with different TTLs and cache invalidation.

**Input**:
```python
import redis.asyncio as redis
from typing import Optional, Any
from dataclasses import dataclass

@dataclass
class CacheConfig:
    host: str = "localhost"
    port: int = 6379
    db: int = 0
    default_ttl: int = 3600
```

**Output**:
```python
import redis.asyncio as redis
from typing import Optional, Any, Callable
from dataclasses import dataclass
from functools import wraps
import json
import hashlib
import logging

logger = logging.getLogger(__name__)

@dataclass
class CacheConfig:
    host: str = "localhost"
    port: int = 6379
    db: int = 0
    default_ttl: int = 3600

class CacheLayer:
    """Multi-layer caching with different TTL strategies."""

    # Cache TTL by layer
    LAYER_TTL = {
        "hot": 60,          # 1 minute - frequently accessed
        "warm": 300,        # 5 minutes - moderately accessed
        "cold": 3600,       # 1 hour - rarely accessed
        "persistent": 86400 # 24 hours - static data
    }

    def __init__(self, config: CacheConfig):
        self.redis = redis.Redis(
            host=config.host,
            port=config.port,
            db=config.db,
            decode_responses=True
        )
        self.default_ttl = config.default_ttl

    def _generate_key(
        self,
        prefix: str,
        *args,
        **kwargs
    ) -> str:
        """Generate consistent cache key."""
        key_data = f"{args}:{sorted(kwargs.items())}"
        key_hash = hashlib.md5(key_data.encode()).hexdigest()[:8]
        return f"{prefix}:{key_hash}"

    async def get(self, key: str) -> Optional[Any]:
        """Get value from cache."""
        value = await self.redis.get(key)
        if value:
            return json.loads(value)
        return None

    async def set(
        self,
        key: str,
        value: Any,
        layer: str = "warm"
    ):
        """Set value in cache with layer-specific TTL."""
        ttl = self.LAYER_TTL.get(layer, self.default_ttl)
        await self.redis.setex(
            key,
            ttl,
            json.dumps(value, default=str)
        )

    async def delete(self, key: str):
        """Delete specific key."""
        await self.redis.delete(key)

    async def delete_pattern(self, pattern: str):
        """Delete all keys matching pattern."""
        cursor = 0
        while True:
            cursor, keys = await self.redis.scan(
                cursor=cursor,
                match=pattern,
                count=100
            )
            if keys:
                await self.redis.delete(*keys)
            if cursor == 0:
                break

    def cached(
        self,
        prefix: str,
        layer: str = "warm"
    ):
        """Decorator for caching function results."""
        def decorator(func: Callable):
            @wraps(func)
            async def wrapper(*args, **kwargs):
                cache_key = self._generate_key(prefix, *args, **kwargs)

                # Try cache first
                cached_value = await self.get(cache_key)
                if cached_value is not None:
                    logger.debug(f"Cache hit: {cache_key}")
                    return cached_value

                # Execute function
                result = await func(*args, **kwargs)

                # Cache result
                if result is not None:
                    await self.set(cache_key, result, layer)
                    logger.debug(f"Cache set: {cache_key}")

                return result
            return wrapper
        return decorator


class UserCacheService:
    """User-specific caching with invalidation patterns."""

    def __init__(self, cache: CacheLayer, db_session):
        self.cache = cache
        self.db = db_session

    @property
    def _prefix(self) -> str:
        return "user"

    async def get_user(self, user_id: int) -> Optional[dict]:
        """Get user with caching."""
        cache_key = f"{self._prefix}:{user_id}"

        # Check cache
        cached = await self.cache.get(cache_key)
        if cached:
            return cached

        # Query database
        user = await self.db.get_user(user_id)
        if user:
            await self.cache.set(cache_key, user.to_dict(), "warm")

        return user.to_dict() if user else None

    async def get_user_profile(self, user_id: int) -> Optional[dict]:
        """Get full user profile (cold cache for expensive queries)."""
        cache_key = f"{self._prefix}:profile:{user_id}"

        cached = await self.cache.get(cache_key)
        if cached:
            return cached

        # Expensive query with joins
        profile = await self.db.get_full_profile(user_id)
        if profile:
            await self.cache.set(cache_key, profile, "cold")

        return profile

    async def update_user(self, user_id: int, data: dict) -> dict:
        """Update user and invalidate related caches."""
        # Update database
        user = await self.db.update_user(user_id, data)

        # Invalidate caches
        await self.cache.delete(f"{self._prefix}:{user_id}")
        await self.cache.delete(f"{self._prefix}:profile:{user_id}")
        await self.cache.delete_pattern(f"users:list:*")

        return user.to_dict()

    async def get_active_users_count(self) -> int:
        """Get count with hot cache (frequently accessed)."""
        cache_key = "users:count:active"

        cached = await self.cache.get(cache_key)
        if cached is not None:
            return cached

        count = await self.db.count_active_users()
        await self.cache.set(cache_key, count, "hot")

        return count


class DistributedLock:
    """Redis-based distributed locking."""

    def __init__(self, redis_client: redis.Redis):
        self.redis = redis_client

    async def acquire(
        self,
        lock_name: str,
        timeout: int = 10,
        blocking: bool = True,
        block_timeout: int = 5
    ) -> Optional[str]:
        """Acquire a distributed lock."""
        import uuid
        lock_id = str(uuid.uuid4())
        lock_key = f"lock:{lock_name}"

        if blocking:
            end_time = asyncio.get_event_loop().time() + block_timeout
            while asyncio.get_event_loop().time() < end_time:
                acquired = await self.redis.set(
                    lock_key,
                    lock_id,
                    nx=True,
                    ex=timeout
                )
                if acquired:
                    return lock_id
                await asyncio.sleep(0.1)
            return None
        else:
            acquired = await self.redis.set(
                lock_key,
                lock_id,
                nx=True,
                ex=timeout
            )
            return lock_id if acquired else None

    async def release(self, lock_name: str, lock_id: str) -> bool:
        """Release a distributed lock."""
        lock_key = f"lock:{lock_name}"

        # Lua script for atomic check-and-delete
        script = """
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
        """
        result = await self.redis.eval(script, 1, lock_key, lock_id)
        return result == 1
```

**Explanation**: This pattern demonstrates multi-layer caching with different TTLs, cache invalidation strategies, and distributed locking for concurrent access control.

---

## Common Patterns

### Pattern 1: Polyglot Persistence

Use the right database for each use case:

```python
class DataRouter:
    """Route data to appropriate database based on access pattern."""

    def __init__(
        self,
        postgres: AsyncSession,
        mongodb: AsyncIOMotorClient,
        redis: redis.Redis
    ):
        self.postgres = postgres  # Relational data
        self.mongodb = mongodb    # Document data
        self.redis = redis        # Real-time data

    async def get_user_complete(self, user_id: int) -> dict:
        """Aggregate data from multiple databases."""

        # Structured user data from PostgreSQL
        user = await self.postgres.execute(
            select(User).where(User.id == user_id)
        )
        user_data = user.scalar_one_or_none()

        if not user_data:
            return None

        # Flexible profile from MongoDB
        profile = await self.mongodb.profiles.find_one(
            {"user_id": str(user_id)}
        )

        # Real-time status from Redis
        status = await self.redis.hgetall(f"user:status:{user_id}")

        return {
            "user": user_data.to_dict(),
            "profile": profile,
            "status": status
        }

    async def save_activity(self, user_id: int, activity: dict):
        """Save activity to appropriate store."""

        # High-frequency real-time counter in Redis
        await self.redis.hincrby(
            f"user:activity:{user_id}",
            "page_views",
            1
        )

        # Activity log in MongoDB (flexible schema)
        await self.mongodb.activities.insert_one({
            "user_id": str(user_id),
            **activity,
            "timestamp": datetime.utcnow()
        })
```

### Pattern 2: Connection Pool Management

```python
from sqlalchemy.ext.asyncio import create_async_engine
from motor.motor_asyncio import AsyncIOMotorClient
import redis.asyncio as redis

class DatabasePool:
    """Manage database connection pools."""

    _instance = None

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance

    def __init__(self):
        if hasattr(self, '_initialized'):
            return
        self._initialized = True

        # PostgreSQL pool
        self.postgres = create_async_engine(
            DATABASE_URL,
            pool_size=20,
            max_overflow=30,
            pool_pre_ping=True,
            pool_recycle=3600
        )

        # MongoDB connection
        self.mongodb = AsyncIOMotorClient(
            MONGODB_URL,
            maxPoolSize=50,
            minPoolSize=10
        )

        # Redis pool
        self.redis = redis.ConnectionPool.from_url(
            REDIS_URL,
            max_connections=100
        )

    async def health_check(self) -> dict:
        """Check all database connections."""
        results = {}

        # PostgreSQL
        try:
            async with self.postgres.connect() as conn:
                await conn.execute("SELECT 1")
            results["postgres"] = "healthy"
        except Exception as e:
            results["postgres"] = f"unhealthy: {e}"

        # MongoDB
        try:
            await self.mongodb.admin.command("ping")
            results["mongodb"] = "healthy"
        except Exception as e:
            results["mongodb"] = f"unhealthy: {e}"

        # Redis
        try:
            r = redis.Redis(connection_pool=self.redis)
            await r.ping()
            results["redis"] = "healthy"
        except Exception as e:
            results["redis"] = f"unhealthy: {e}"

        return results
```

### Pattern 3: Database Migration Strategy

```python
from alembic import command
from alembic.config import Config

class MigrationManager:
    """Manage database migrations safely."""

    def __init__(self, alembic_cfg_path: str):
        self.alembic_cfg = Config(alembic_cfg_path)

    def upgrade(self, revision: str = "head"):
        """Apply pending migrations."""
        command.upgrade(self.alembic_cfg, revision)

    def downgrade(self, revision: str):
        """Rollback to specific revision."""
        command.downgrade(self.alembic_cfg, revision)

    def create_migration(self, message: str):
        """Auto-generate migration from model changes."""
        command.revision(
            self.alembic_cfg,
            message=message,
            autogenerate=True
        )

    def current(self) -> str:
        """Get current migration revision."""
        return command.current(self.alembic_cfg)
```

---

## Anti-Patterns (Patterns to Avoid)

### Anti-Pattern 1: Missing Database Indexes

**Problem**: Queries without proper indexes cause full table scans.

```python
# Incorrect approach - no index consideration
async def find_users_by_email(email: str):
    return await session.execute(
        select(User).where(User.email == email)
    )
```

**Solution**: Create appropriate indexes.

```sql
-- Correct approach - add index
CREATE UNIQUE INDEX idx_users_email ON users (email);
```

```python
# And verify query uses index
async def find_users_by_email(email: str):
    # This will now use the index
    return await session.execute(
        select(User).where(User.email == email)
    )
```

### Anti-Pattern 2: Ignoring Connection Limits

**Problem**: Opening unlimited database connections.

```python
# Incorrect approach - new connection per request
async def get_data():
    engine = create_async_engine(DATABASE_URL)
    async with engine.connect() as conn:
        return await conn.execute(query)
```

**Solution**: Use connection pooling.

```python
# Correct approach - shared connection pool
engine = create_async_engine(
    DATABASE_URL,
    pool_size=20,
    max_overflow=10
)

async def get_data():
    async with engine.connect() as conn:
        return await conn.execute(query)
```

### Anti-Pattern 3: Caching Without Invalidation

**Problem**: Stale data in cache after updates.

```python
# Incorrect approach - cache without invalidation
async def update_user(user_id: int, data: dict):
    await db.update_user(user_id, data)
    # Cache still has old data!
```

**Solution**: Always invalidate cache on updates.

```python
# Correct approach - invalidate on update
async def update_user(user_id: int, data: dict):
    await db.update_user(user_id, data)

    # Invalidate related caches
    await cache.delete(f"user:{user_id}")
    await cache.delete(f"user:profile:{user_id}")
    await cache.delete_pattern("users:list:*")
```

---

## Performance Benchmarks

### Query Performance Comparison

```python
# Before optimization: 15,000ms
# After indexing: 150ms
# With materialized view: 15ms
# With Redis cache: 1ms

async def benchmark_user_analytics():
    import time

    # Cold query (no cache)
    start = time.time()
    result = await get_user_analytics_no_cache(user_id=1)
    cold_time = time.time() - start

    # Warm query (cached)
    start = time.time()
    result = await get_user_analytics_cached(user_id=1)
    warm_time = time.time() - start

    return {
        "cold_query_ms": cold_time * 1000,
        "warm_query_ms": warm_time * 1000,
        "speedup_factor": cold_time / warm_time
    }
```

---

*For additional patterns and database-specific optimizations, see the `modules/` directory.*
