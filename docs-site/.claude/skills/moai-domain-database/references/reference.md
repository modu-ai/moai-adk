# Database Domain Reference

## API Reference

### PostgreSQL Operations

Connection and Pool Management:
```python
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.pool import QueuePool

# Optimized engine configuration
engine = create_async_engine(
    "postgresql+asyncpg://user:pass@localhost/db",
    poolclass=QueuePool,
    pool_size=20,          # Base pool size
    max_overflow=30,       # Additional connections allowed
    pool_pre_ping=True,    # Health check before use
    pool_recycle=3600,     # Recycle connections after 1 hour
    echo=False,            # SQL logging
    connect_args={
        "server_settings": {
            "application_name": "myapp",
            "jit": "off"    # Disable JIT for OLTP
        }
    }
)
```

Query Optimization:
```python
from sqlalchemy import text, select, func
from sqlalchemy.dialects.postgresql import insert

# Bulk upsert with conflict handling
async def upsert_users(session: AsyncSession, users: list):
    stmt = insert(User).values(users)
    stmt = stmt.on_conflict_do_update(
        index_elements=['email'],
        set_={
            'name': stmt.excluded.name,
            'updated_at': func.now()
        }
    )
    await session.execute(stmt)

# Efficient pagination with keyset
async def paginate_users(session: AsyncSession, last_id: int = 0, limit: int = 100):
    result = await session.execute(
        select(User)
        .where(User.id > last_id)
        .order_by(User.id)
        .limit(limit)
    )
    return result.scalars().all()

# Window functions
async def get_ranked_users(session: AsyncSession):
    stmt = text("""
        SELECT id, name, score,
               RANK() OVER (ORDER BY score DESC) as rank,
               PERCENT_RANK() OVER (ORDER BY score DESC) as percentile
        FROM users
        WHERE active = true
    """)
    result = await session.execute(stmt)
    return result.fetchall()
```

### MongoDB Operations

Aggregation Pipelines:
```python
from motor.motor_asyncio import AsyncIOMotorClient

client = AsyncIOMotorClient("mongodb://localhost:27017")
db = client.myapp

# Complex aggregation pipeline
async def get_user_analytics():
    pipeline = [
        {"$match": {"status": "active"}},
        {"$lookup": {
            "from": "orders",
            "localField": "_id",
            "foreignField": "user_id",
            "as": "orders"
        }},
        {"$addFields": {
            "order_count": {"$size": "$orders"},
            "total_spent": {"$sum": "$orders.amount"}
        }},
        {"$group": {
            "_id": "$region",
            "user_count": {"$sum": 1},
            "avg_orders": {"$avg": "$order_count"},
            "total_revenue": {"$sum": "$total_spent"}
        }},
        {"$sort": {"total_revenue": -1}}
    ]
    return await db.users.aggregate(pipeline).to_list(None)

# Efficient bulk operations
async def bulk_update_users(updates: list):
    operations = []
    for update in updates:
        operations.append(
            UpdateOne(
                {"_id": update["id"]},
                {"$set": update["data"]},
                upsert=True
            )
        )
    result = await db.users.bulk_write(operations, ordered=False)
    return result.modified_count
```

Index Management:
```python
# Create compound indexes
async def setup_indexes():
    await db.users.create_indexes([
        IndexModel([("email", 1)], unique=True),
        IndexModel([("region", 1), ("created_at", -1)]),
        IndexModel([("status", 1), ("last_login", -1)]),
        IndexModel(
            [("location", "2dsphere")],
            sparse=True
        )
    ])

# Analyze index usage
async def get_index_stats():
    return await db.command({
        "aggregate": "users",
        "pipeline": [{"$indexStats": {}}],
        "cursor": {}
    })
```

### Redis Operations

Caching Patterns:
```python
import redis.asyncio as redis
import json
from functools import wraps

redis_client = redis.from_url("redis://localhost:6379")

# Cache-aside pattern
async def get_user_cached(user_id: int):
    cache_key = f"user:{user_id}"

    # Try cache first
    cached = await redis_client.get(cache_key)
    if cached:
        return json.loads(cached)

    # Fetch from database
    user = await db.get_user(user_id)
    if user:
        await redis_client.setex(
            cache_key,
            3600,  # 1 hour TTL
            json.dumps(user.dict())
        )
    return user

# Write-through pattern
async def update_user(user_id: int, data: dict):
    # Update database
    user = await db.update_user(user_id, data)

    # Update cache
    cache_key = f"user:{user_id}"
    await redis_client.setex(
        cache_key,
        3600,
        json.dumps(user.dict())
    )
    return user

# Cache invalidation
async def invalidate_user_cache(user_id: int):
    cache_key = f"user:{user_id}"
    await redis_client.delete(cache_key)

    # Invalidate related caches
    pattern = f"user:{user_id}:*"
    keys = await redis_client.keys(pattern)
    if keys:
        await redis_client.delete(*keys)
```

Advanced Redis Data Structures:
```python
# Rate limiting with sliding window
async def check_rate_limit(user_id: str, limit: int = 100, window: int = 60):
    key = f"ratelimit:{user_id}"
    now = time.time()

    pipe = redis_client.pipeline()
    pipe.zremrangebyscore(key, 0, now - window)
    pipe.zadd(key, {str(now): now})
    pipe.zcard(key)
    pipe.expire(key, window)

    results = await pipe.execute()
    count = results[2]

    return count <= limit

# Distributed locking
async def acquire_lock(resource: str, timeout: int = 30):
    lock_key = f"lock:{resource}"
    lock_id = str(uuid.uuid4())

    acquired = await redis_client.set(
        lock_key,
        lock_id,
        nx=True,
        ex=timeout
    )
    return lock_id if acquired else None

async def release_lock(resource: str, lock_id: str):
    lock_key = f"lock:{resource}"
    script = """
    if redis.call("get", KEYS[1]) == ARGV[1] then
        return redis.call("del", KEYS[1])
    else
        return 0
    end
    """
    return await redis_client.eval(script, 1, lock_key, lock_id)
```

---

## Configuration Options

### PostgreSQL Configuration

```yaml
# config/postgresql.yaml
connection:
  host: localhost
  port: 5432
  database: myapp
  user: ${DB_USER}
  password: ${DB_PASSWORD}
  sslmode: require

pool:
  min_size: 10
  max_size: 50
  max_overflow: 20
  pool_recycle: 3600
  pool_pre_ping: true
  pool_timeout: 30

performance:
  statement_cache_size: 1024
  prepared_statement_cache_size: 256

replication:
  read_replicas:
    - host: replica1.db.local
      port: 5432
    - host: replica2.db.local
      port: 5432
  load_balance: round_robin
```

### MongoDB Configuration

```yaml
# config/mongodb.yaml
connection:
  uri: mongodb://localhost:27017
  database: myapp

pool:
  max_pool_size: 100
  min_pool_size: 10
  max_idle_time_ms: 60000
  wait_queue_timeout_ms: 10000

replication:
  replica_set: rs0
  read_preference: secondaryPreferred
  write_concern:
    w: majority
    j: true
    wtimeout: 5000

sharding:
  enabled: true
  shard_key: user_id
  chunk_size_mb: 64
```

### Redis Configuration

```yaml
# config/redis.yaml
connection:
  url: redis://localhost:6379
  password: ${REDIS_PASSWORD}
  db: 0

pool:
  max_connections: 100
  min_idle: 10
  max_idle: 50
  connection_timeout: 5
  socket_timeout: 5

cluster:
  enabled: false
  nodes:
    - redis://node1:6379
    - redis://node2:6379
    - redis://node3:6379

sentinel:
  enabled: false
  master_name: mymaster
  sentinels:
    - host: sentinel1
      port: 26379
```

---

## Integration Patterns

### Multi-Database Transaction

```python
from contextlib import asynccontextmanager

class TransactionManager:
    def __init__(self, pg_session, mongo_client, redis_client):
        self.pg = pg_session
        self.mongo = mongo_client
        self.redis = redis_client

    @asynccontextmanager
    async def transaction(self):
        pg_transaction = None
        mongo_session = None
        try:
            # Start PostgreSQL transaction
            pg_transaction = await self.pg.begin()

            # Start MongoDB session
            mongo_session = await self.mongo.start_session()
            mongo_session.start_transaction()

            yield {
                'pg': self.pg,
                'mongo': mongo_session,
                'redis': self.redis
            }

            # Commit all
            await pg_transaction.commit()
            await mongo_session.commit_transaction()

        except Exception as e:
            # Rollback all
            if pg_transaction:
                await pg_transaction.rollback()
            if mongo_session:
                await mongo_session.abort_transaction()

            # Cleanup Redis operations (compensating transaction)
            raise
        finally:
            if mongo_session:
                await mongo_session.end_session()
```

### CQRS Pattern

```python
# Command side (writes)
class UserCommandHandler:
    def __init__(self, pg_session):
        self.pg = pg_session

    async def create_user(self, command: CreateUserCommand):
        user = User(**command.dict())
        self.pg.add(user)
        await self.pg.flush()

        # Publish event for read model update
        await event_bus.publish("user.created", {
            "id": user.id,
            "email": user.email,
            "name": user.name
        })
        return user.id

# Query side (reads)
class UserQueryHandler:
    def __init__(self, mongo_db, redis_client):
        self.mongo = mongo_db
        self.redis = redis_client

    async def get_user_profile(self, user_id: int):
        # Check cache
        cached = await self.redis.get(f"profile:{user_id}")
        if cached:
            return json.loads(cached)

        # Query read model
        profile = await self.mongo.user_profiles.find_one({"_id": user_id})
        if profile:
            await self.redis.setex(
                f"profile:{user_id}",
                3600,
                json.dumps(profile)
            )
        return profile

# Event handler to sync read model
class UserEventHandler:
    def __init__(self, mongo_db):
        self.mongo = mongo_db

    async def on_user_created(self, event: dict):
        await self.mongo.user_profiles.insert_one({
            "_id": event["id"],
            "email": event["email"],
            "name": event["name"],
            "created_at": datetime.utcnow()
        })
```

---

## Troubleshooting

### PostgreSQL Issues

Issue: Connection pool exhausted
Symptoms: "too many connections" errors, request timeouts
Solution:
- Increase max_pool_size gradually
- Monitor active connections: SELECT count(*) FROM pg_stat_activity
- Check for connection leaks in application code
- Consider using PgBouncer for connection pooling

Issue: Slow queries
Symptoms: High response times, CPU spikes
Solution:
- Run EXPLAIN ANALYZE on slow queries
- Check for missing indexes: SELECT * FROM pg_stat_user_indexes WHERE idx_scan = 0
- Update table statistics: ANALYZE table_name
- Consider query rewrites or denormalization

Issue: Lock contention
Symptoms: Queries waiting, deadlocks
Solution:
- Monitor locks: SELECT * FROM pg_locks WHERE NOT granted
- Use shorter transactions
- Add proper row-level locking hints
- Consider optimistic locking patterns

### MongoDB Issues

Issue: Slow aggregation pipelines
Symptoms: Long query times, high memory usage
Solution:
- Add indexes for $match stages
- Use $project early to reduce document size
- Enable allowDiskUse for large result sets
- Consider $merge for materialized views

Issue: Write performance degradation
Symptoms: Slow inserts, high disk I/O
Solution:
- Check index count and optimize
- Use bulk writes instead of individual operations
- Consider write concern adjustment
- Monitor oplog size for replication

### Redis Issues

Issue: Memory pressure
Symptoms: OOM errors, evictions
Solution:
- Set maxmemory and maxmemory-policy
- Implement TTL for all keys
- Use more efficient data structures
- Consider Redis Cluster for horizontal scaling

Issue: Connection timeouts
Symptoms: Connection refused, timeout errors
Solution:
- Increase tcp-backlog value
- Check network latency
- Use connection pooling
- Monitor slow log: SLOWLOG GET 10

---

## External Resources

### PostgreSQL
- Official Documentation: https://www.postgresql.org/docs/
- Performance Tuning: https://wiki.postgresql.org/wiki/Performance_Optimization
- pg_stat_statements: https://www.postgresql.org/docs/current/pgstatstatements.html
- PgBouncer: https://www.pgbouncer.org/

### MongoDB
- Official Documentation: https://www.mongodb.com/docs/
- Aggregation Reference: https://www.mongodb.com/docs/manual/aggregation/
- Performance Best Practices: https://www.mongodb.com/docs/manual/administration/analyzing-mongodb-performance/
- MongoDB University: https://university.mongodb.com/

### Redis
- Official Documentation: https://redis.io/docs/
- Redis Best Practices: https://redis.io/docs/management/optimization/
- Redis Cluster: https://redis.io/docs/management/scaling/
- Redis Insight: https://redis.com/redis-enterprise/redis-insight/

### Tools
- pgAdmin: https://www.pgadmin.org/
- MongoDB Compass: https://www.mongodb.com/products/compass
- RedisInsight: https://redis.com/redis-enterprise/redis-insight/
- DBeaver: https://dbeaver.io/

### Books and Courses
- High Performance MySQL: https://www.oreilly.com/library/view/high-performance-mysql/
- MongoDB: The Definitive Guide: https://www.oreilly.com/library/view/mongodb-the-definitive/
- Redis in Action: https://www.manning.com/books/redis-in-action

---

Version: 1.0.0
Last Updated: 2025-12-06
