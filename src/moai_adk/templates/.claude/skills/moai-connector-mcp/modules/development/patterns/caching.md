# Caching Patterns

Caching strategies for improving MCP server performance.

---

## In-Memory Caching

### Simple Cache

```python
from functools import wraps
from datetime import datetime, timedelta
from typing import Optional
import hashlib

class SimpleCache:
    def __init__(self, ttl_seconds: int = 300):
        self.cache = {}
        self.ttl = ttl_seconds

    def get(self, key: str) -> Optional[any]:
        """Get value from cache."""
        if key not in self.cache:
            return None

        value, timestamp = self.cache[key]
        age = (datetime.now() - timestamp).total_seconds()

        if age > self.ttl:
            del self.cache[key]
            return None

        return value

    def set(self, key: str, value: any):
        """Set value in cache."""
        self.cache[key] = (value, datetime.now())

    def clear(self):
        """Clear all cache."""
        self.cache.clear()

    def decorator(self, ttl_seconds: Optional[int] = None):
        """Decorator for caching function results."""
        def decorating_function(func):
            @wraps(func)
            def wrapper(*args, **kwargs):
                # Generate cache key
                key = f"{func.__name__}:{str(args)}:{str(kwargs)}"
                key = hashlib.md5(key.encode()).hexdigest()

                # Check cache
                cached = self.get(key)
                if cached is not None:
                    return cached

                # Call function
                result = func(*args, **kwargs)

                # Store in cache
                self.set(key, result)
                return result

            return wrapper
        return decorating_function

cache = SimpleCache(ttl_seconds=600)

@server.tool()
@cache.decorator()
def expensive_query(table: str, query: str) -> list[dict]:
    """Query database (cached for 10 minutes)."""
    return execute_query(table, query)
```

---

## Distributed Caching with Redis

### Redis Cache

```python
import redis
import json
from typing import Optional

class RedisCache:
    def __init__(self, host: str = "localhost", port: int = 6379, db: int = 0):
        self.client = redis.Redis(host=host, port=port, db=db)

    def get(self, key: str) -> Optional[dict]:
        """Get from cache."""
        value = self.client.get(key)
        if value:
            return json.loads(value)
        return None

    def set(self, key: str, value: dict, ttl_seconds: int = 300):
        """Set cache with TTL."""
        self.client.setex(
            key,
            ttl_seconds,
            json.dumps(value)
        )

    def delete(self, key: str):
        """Delete from cache."""
        self.client.delete(key)

    def invalidate_pattern(self, pattern: str):
        """Invalidate cache by pattern."""
        for key in self.client.scan_iter(match=pattern):
            self.client.delete(key)

redis_cache = RedisCache()

@server.tool()
def cached_user_query(user_id: str) -> dict:
    """Get user (cached in Redis)."""
    cache_key = f"user:{user_id}"

    # Try cache
    cached = redis_cache.get(cache_key)
    if cached:
        return cached

    # Fetch from database
    user = fetch_user(user_id)

    # Cache for 1 hour
    redis_cache.set(cache_key, user, ttl_seconds=3600)

    return user

@server.tool()
def update_user(user_id: str, data: dict) -> dict:
    """Update user and invalidate cache."""
    updated = update_user_data(user_id, data)

    # Invalidate related caches
    redis_cache.delete(f"user:{user_id}")
    redis_cache.invalidate_pattern(f"user:{user_id}:*")

    return updated
```

---

## Cache Invalidation Strategies

### Time-Based Invalidation

```python
@server.tool()
def get_report(report_type: str) -> dict:
    """Get report (cached for 1 hour)."""
    cache_key = f"report:{report_type}"

    cached = redis_cache.get(cache_key)
    if cached:
        return cached

    report = generate_report(report_type)
    redis_cache.set(cache_key, report, ttl_seconds=3600)  # 1 hour

    return report
```

### Event-Based Invalidation

```python
@server.tool()
def create_user(email: str, name: str) -> dict:
    """Create user and invalidate user list cache."""
    user = save_user(email, name)

    # Invalidate list caches
    redis_cache.delete("users:all")
    redis_cache.invalidate_pattern("users:page:*")

    return user
```

### Manual Invalidation

```python
@server.tool()
def clear_cache(pattern: str) -> dict:
    """Clear cache matching pattern."""
    redis_cache.invalidate_pattern(pattern)
    return {"status": "success", "pattern": pattern}
```

---

## Cache Warming

### Preload Frequently Used Data

```python
async def warm_cache():
    """Preload frequently accessed data."""
    logger.info("Warming cache...")

    # Preload popular items
    popular = fetch_popular_items(limit=100)
    for item in popular:
        key = f"item:{item['id']}"
        redis_cache.set(key, item, ttl_seconds=3600)

    # Preload user counts
    stats = {
        "total_users": count_users(),
        "active_users": count_active_users(),
        "total_posts": count_posts()
    }
    redis_cache.set("stats:global", stats, ttl_seconds=1800)

    logger.info("Cache warming complete")

# Call on startup
if __name__ == "__main__":
    asyncio.run(warm_cache())
    server.run()
```

---

## Cache-Aside Pattern

```python
@server.tool()
def get_product(product_id: str) -> dict:
    """Get product using cache-aside pattern."""
    cache_key = f"product:{product_id}"

    # 1. Try cache
    cached = redis_cache.get(cache_key)
    if cached:
        logger.info(f"Cache hit for {product_id}")
        return cached

    # 2. Cache miss, fetch from source
    logger.info(f"Cache miss for {product_id}")
    product = fetch_product_from_db(product_id)

    if not product:
        raise ValueError(f"Product not found: {product_id}")

    # 3. Load into cache
    redis_cache.set(cache_key, product, ttl_seconds=3600)

    return product
```

---

## Write-Through Caching

```python
@server.tool()
def update_product(product_id: str, data: dict) -> dict:
    """Update product with write-through cache."""
    # 1. Update in database
    updated = update_product_db(product_id, data)

    # 2. Update in cache
    cache_key = f"product:{product_id}"
    redis_cache.set(cache_key, updated, ttl_seconds=3600)

    return updated
```

---

## Caching Considerations

### Cache Key Design

```python
def build_cache_key(*args, **kwargs) -> str:
    """Build consistent cache key."""
    key_parts = [
        f"{arg}" for arg in args
    ] + [
        f"{k}:{v}" for k, v in sorted(kwargs.items())
    ]

    key_str = ":".join(key_parts)
    return hashlib.md5(key_str.encode()).hexdigest()

# Usage
cache_key = build_cache_key("user", 123, role="admin")
# Produces: "user:123:role:admin" (hashed)
```

### Cache Size Limits

```python
class BoundedCache:
    def __init__(self, max_size: int = 1000):
        self.cache = {}
        self.max_size = max_size
        self.access_times = {}

    def set(self, key: str, value: any):
        """Set with LRU eviction."""
        if len(self.cache) >= self.max_size:
            # Evict least recently used
            lru_key = min(self.access_times, key=self.access_times.get)
            del self.cache[lru_key]
            del self.access_times[lru_key]

        self.cache[key] = value
        self.access_times[key] = datetime.now()

    def get(self, key: str):
        """Get and update access time."""
        if key in self.cache:
            self.access_times[key] = datetime.now()
            return self.cache[key]
        return None
```

---

## Best Practices

✅ **Caching**:
- Cache frequently accessed, expensive data
- Use appropriate TTLs (not too short, not too long)
- Implement cache invalidation strategy
- Monitor cache hit/miss rates
- Use distributed cache for production
- Document cache keys and TTLs

✅ **Performance**:
- Measure cache effectiveness
- Avoid cache stampede (use locks)
- Implement cache warming
- Monitor memory usage
- Set reasonable size limits

❌ **Avoid**:
- Caching sensitive data
- Very long TTLs
- Uncontrolled cache growth
- Ignoring cache invalidation
- Caching everything
- Synchronous cache operations

---

**Last Updated**: 2025-11-27
