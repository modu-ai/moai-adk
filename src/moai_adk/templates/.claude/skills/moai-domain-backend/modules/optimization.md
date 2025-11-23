# Backend Performance Optimization

## Application-Level Optimization

### Async/Await Optimization

**Problem**: Synchronous operations block event loop, reducing throughput.

**Solution**: Use async/await for I/O-bound operations.

```python
import asyncio
from fastapi import FastAPI
import aiohttp

app = FastAPI()

# Good: Parallel async requests
async def fetch_multiple_services(user_id: int):
    """Fetch data from multiple services in parallel"""
    async with aiohttp.ClientSession() as session:
        tasks = [
            fetch_user_service(session, user_id),
            fetch_order_service(session, user_id),
            fetch_payment_service(session, user_id)
        ]
        # Run all requests in parallel
        results = await asyncio.gather(*tasks)
        return results

@app.get("/user-profile/{user_id}")
async def get_user_profile(user_id: int):
    user_data, orders, payments = await fetch_multiple_services(user_id)
    return {
        "user": user_data,
        "orders": orders,
        "payments": payments
    }

# Bad: Sequential requests (3x slower)
@app.get("/user-profile-slow/{user_id}")
async def get_user_profile_slow(user_id: int):
    user_data = await fetch_user_service(user_id)
    orders = await fetch_order_service(user_id)
    payments = await fetch_payment_service(user_id)
    return {"user": user_data, "orders": orders, "payments": payments}
```

**Performance Impact**: 3x faster response time for concurrent operations.

---

### Connection Pooling

**Database Connection Pool Configuration**:
```python
from sqlalchemy import create_engine
from sqlalchemy.pool import QueuePool

# Optimized connection pool
engine = create_engine(
    'postgresql://user:password@localhost:5432/mydb',
    poolclass=QueuePool,
    pool_size=20,           # Core connections to keep open
    max_overflow=10,        # Additional connections when needed
    pool_timeout=30,        # Wait max 30s for connection
    pool_recycle=3600,      # Recycle connections after 1 hour
    pool_pre_ping=True,     # Check connection health before use
    echo_pool=False         # Disable logging in production
)

# Connection pool sizing formula
# pool_size = (number_of_workers * 2) + spare_connections
# For 10 workers: pool_size = 20 + 2 = 22

# Redis connection pool
import redis

redis_pool = redis.ConnectionPool(
    host='localhost',
    port=6379,
    max_connections=50,  # Max concurrent Redis connections
    socket_connect_timeout=5,
    socket_keepalive=True
)

redis_client = redis.Redis(connection_pool=redis_pool)
```

**Performance Metrics**:
- Connection establishment: ~10-50ms reduced to ~1ms via pooling
- Memory: Fewer open connections = lower memory footprint
- Latency: Reused connections have near-zero overhead

---

### Query Optimization

**N+1 Query Problem Solution**:
```python
from sqlalchemy.orm import selectinload, joinedload

# Bad: N+1 queries
def get_orders_slow(user_id: int):
    """This executes N+1 queries!"""
    user = db.session.query(User).filter_by(id=user_id).first()
    orders = db.session.query(Order).filter_by(user_id=user_id).all()

    # This loops fires database query for EACH order
    for order in orders:
        items = order.items  # N additional queries!

# Good: Eager loading
def get_orders_optimized(user_id: int):
    """Single query with JOINs"""
    user = db.session.query(User).options(
        selectinload(User.orders).selectinload(Order.items)
    ).filter_by(id=user_id).first()

    # Now accessing order.items doesn't trigger queries
    return user

# Alternative: Explicit JOIN
from sqlalchemy import select

def get_orders_with_join():
    """Explicit join for total control"""
    stmt = select(Order).options(
        joinedload(Order.items),
        joinedload(Order.user)
    )
    return db.session.execute(stmt).unique().scalars().all()
```

**Performance Improvement**: 100+ queries reduced to 1-3 queries.

---

## Database Optimization

### Indexing Strategy

```python
# Analyze slow queries
async def find_slow_queries(db_connection):
    """Identify queries needing optimization"""
    slow_queries = await db_connection.fetch("""
        SELECT query, calls, total_time, mean_time
        FROM pg_stat_statements
        WHERE mean_time > 100  -- queries > 100ms
        ORDER BY total_time DESC
        LIMIT 20
    """)

    for query in slow_queries:
        print(f"Slow query: {query['query'][:100]}")
        print(f"  Calls: {query['calls']}, Mean time: {query['mean_time']}ms")

# Index recommendations based on query patterns
sql_recommendations = """
-- Query: SELECT * FROM orders WHERE user_id = ? AND status = ?
-- Solution: Composite index on (user_id, status)
CREATE INDEX idx_orders_user_status ON orders(user_id, status);

-- Query: SELECT * FROM users ORDER BY created_at DESC LIMIT 10
-- Solution: BRIN index for sequential data
CREATE INDEX idx_users_created USING BRIN ON users(created_at);

-- Query: Full-text search on product description
-- Solution: GiST index for text search
CREATE INDEX idx_products_search ON products USING GiST(to_tsvector('english', description));
"""
```

### Partitioning Strategy

```sql
-- Partition by date (for logs, events)
CREATE TABLE events (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(50),
    created_at TIMESTAMP NOT NULL,
    data JSONB
) PARTITION BY RANGE (created_at);

-- Create monthly partitions automatically
CREATE TABLE events_2025_01 PARTITION OF events
    FOR VALUES FROM ('2025-01-01') TO ('2025-02-01');

-- Query automatically uses correct partition
SELECT * FROM events WHERE created_at >= '2025-01-15';

-- Partition by hash (for scalability)
CREATE TABLE orders (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    created_at TIMESTAMP
) PARTITION BY HASH (user_id);

CREATE TABLE orders_0 PARTITION OF orders
    FOR VALUES WITH (MODULUS 4, REMAINDER 0);
CREATE TABLE orders_1 PARTITION OF orders
    FOR VALUES WITH (MODULUS 4, REMAINDER 1);
-- ... etc for REMAINDER 2, 3
```

**Benefits**:
- Faster queries on subsets of data
- Easier maintenance and archiving
- Better parallelization of queries

---

## Caching Optimization

### Cache Invalidation Strategies

```python
from datetime import datetime, timedelta

class CacheInvalidationStrategy:
    """Different cache invalidation patterns"""

    # 1. Time-Based (TTL)
    async def cache_with_ttl(self, key: str, data, ttl_seconds: int = 3600):
        """Simple time-based expiration"""
        self.cache.setex(key, ttl_seconds, json.dumps(data))

    # 2. Event-Based Invalidation
    async def invalidate_user_cache(self, user_id: int):
        """Invalidate all caches related to user"""
        pattern = f"user:{user_id}:*"
        keys = self.cache.keys(pattern)
        if keys:
            self.cache.delete(*keys)

    # 3. LRU Cache with size limits
    from functools import lru_cache

    @lru_cache(maxsize=128)  # Keep only 128 most-used items
    def expensive_computation(self, param: str):
        """Automatic eviction of least-recently-used items"""
        return complex_calculation(param)

    # 4. Probabilistic expiration (cache stampede prevention)
    async def cached_with_early_expiry(self, key: str, data, ttl: int = 3600):
        """Refresh cache before it expires (probabilistic)"""
        import random
        refresh_probability = 0.1  # 10% chance to refresh early
        actual_ttl = ttl if random.random() > refresh_probability else ttl // 2
        self.cache.setex(key, actual_ttl, json.dumps(data))
```

### Multi-Tier Caching

```python
import time

class MultiTierCache:
    """L1: Local (memory) → L2: Redis → L3: Database"""

    def __init__(self):
        self.l1_cache = {}  # In-process cache
        self.l2_client = redis.Redis()

    async def get(self, key: str):
        """Attempt retrieval from each tier"""
        # L1: Check local cache
        if key in self.l1_cache:
            print(f"L1 HIT: {key}")
            return self.l1_cache[key]

        # L2: Check Redis
        l2_value = self.l2_client.get(key)
        if l2_value:
            print(f"L2 HIT: {key}")
            value = json.loads(l2_value)
            self.l1_cache[key] = value  # Populate L1
            return value

        # L3: Database miss - fetch from source
        print(f"L3 MISS: {key} - Fetching from database")
        value = await self.fetch_from_db(key)

        # Populate both caches
        self.l1_cache[key] = value
        self.l2_client.setex(key, 3600, json.dumps(value))

        return value

    async def invalidate(self, key: str):
        """Invalidate across all tiers"""
        if key in self.l1_cache:
            del self.l1_cache[key]
        self.l2_client.delete(key)
```

---

## Monitoring & Profiling

### Performance Profiling

```python
import time
from functools import wraps
import statistics

def profile_performance(func):
    """Decorator to profile function performance"""
    func.call_times = []

    @wraps(func)
    def wrapper(*args, **kwargs):
        start_time = time.perf_counter()
        result = func(*args, **kwargs)
        elapsed = (time.perf_counter() - start_time) * 1000  # Convert to ms

        func.call_times.append(elapsed)

        # Print stats every 100 calls
        if len(func.call_times) % 100 == 0:
            avg = statistics.mean(func.call_times)
            p95 = statistics.quantiles(func.call_times, n=20)[18]  # 95th percentile
            p99 = statistics.quantiles(func.call_times, n=100)[98]  # 99th percentile

            print(f"{func.__name__} - Avg: {avg:.2f}ms, P95: {p95:.2f}ms, P99: {p99:.2f}ms")

        return result

    return wrapper

# Usage
@profile_performance
async def process_order(order_id: int):
    # Processing logic
    pass
```

### Memory Optimization

```python
import sys

class MemoryOptimizedDataStructure:
    """Use __slots__ to reduce memory overhead"""

    # Without slots: ~296 bytes per instance
    class User:
        def __init__(self, name, email):
            self.name = name
            self.email = email

    # With slots: ~80 bytes per instance (73% reduction!)
    class OptimizedUser:
        __slots__ = ['name', 'email']

        def __init__(self, name, email):
            self.name = name
            self.email = email

    def demonstrate(self):
        regular = self.User("John", "john@example.com")
        optimized = self.OptimizedUser("John", "john@example.com")

        print(f"Regular User: {sys.getsizeof(regular)} bytes")
        print(f"Optimized User: {sys.getsizeof(optimized)} bytes")
        # Output: Regular User: 296 bytes, Optimized User: 80 bytes
```

---

## Scaling Patterns

### Read Replicas Strategy

```python
# Primary (write) connection
primary_engine = create_engine("postgresql://user:pass@primary:5432/db")

# Read replica connections (load balanced)
read_replicas = [
    create_engine("postgresql://user:pass@replica-1:5432/db"),
    create_engine("postgresql://user:pass@replica-2:5432/db"),
    create_engine("postgresql://user:pass@replica-3:5432/db")
]

async def execute_write_query(query):
    """Always write to primary"""
    return await primary_engine.execute(query)

async def execute_read_query(query):
    """Distribute reads across replicas"""
    import random
    replica = random.choice(read_replicas)
    return await replica.execute(query)
```

---

**Last Updated**: 2025-11-22
**Version**: 4.0.0
