# Performance Optimization Examples

Practical implementations for profiling, benchmarking, and optimizing application performance.

---

## Python Profiling with Scalene

### Complete Profiling Workflow

```python
#!/usr/bin/env python3
"""
Production-grade profiling example with Scalene.
Profiles CPU, memory, and GPU usage with detailed reports.
"""

import time
import numpy as np
from typing import List


def cpu_intensive_task(n: int) -> List[int]:
    """Simulate CPU-bound computation."""
    result = []
    for i in range(n):
        # Inefficient fibonacci calculation (intentionally slow)
        result.append(fib_recursive(i % 20))
    return result


def fib_recursive(n: int) -> int:
    """Recursive fibonacci (inefficient)."""
    if n <= 1:
        return n
    return fib_recursive(n - 1) + fib_recursive(n - 2)


def memory_intensive_task(size: int) -> np.ndarray:
    """Simulate memory-bound computation."""
    # Allocate large array
    large_array = np.random.rand(size, size)

    # Matrix multiplication (memory + CPU)
    result = np.dot(large_array, large_array.T)

    return result


def io_intensive_task(filename: str, iterations: int) -> None:
    """Simulate I/O-bound operations."""
    with open(filename, 'w') as f:
        for i in range(iterations):
            f.write(f"Line {i}: " + "x" * 100 + "\n")


def main():
    """Main function to profile."""
    print("Starting profiling demo...")

    # CPU-bound task
    print("1. CPU-intensive task...")
    cpu_result = cpu_intensive_task(1000)

    # Memory-bound task
    print("2. Memory-intensive task...")
    mem_result = memory_intensive_task(1000)

    # I/O-bound task
    print("3. I/O-intensive task...")
    io_intensive_task('/tmp/test_perf.txt', 10000)

    print(f"Completed: CPU result size={len(cpu_result)}, Memory result shape={mem_result.shape}")


if __name__ == "__main__":
    main()
```

### Running Scalene Profiler

```bash
# Install Scalene
pip install scalene

# Basic profiling (terminal output)
scalene profile_demo.py

# Generate HTML report
scalene --html --outfile profile_report.html profile_demo.py

# Profile with reduced sampling (faster, less accurate)
scalene --reduced-profile profile_demo.py

# Profile only CPU (skip memory)
scalene --cpu-only profile_demo.py

# Profile with specific Python version
scalene --python python3.11 profile_demo.py
```

### Interpreting Scalene Output

```python
"""
Scalene reports show:

1. CPU Time (%)
   - Native (C/C++ libraries like NumPy)
   - Python (your code)
   - System (I/O, OS calls)

2. Memory Usage (MB)
   - Peak memory per line
   - Memory growth rate
   - Copy volume (unnecessary copies)

3. GPU Usage (if applicable)
   - GPU kernel time
   - Memory transfers

Example output interpretation:
┌─────────────────────────────────────────────┐
│ Line | CPU % | Memory | Code               │
├─────────────────────────────────────────────┤
│  15  │  45%  │  120MB │ result = np.dot()  │  ← HOTSPOT
│  18  │   5%  │   2MB  │ for i in range()   │
└─────────────────────────────────────────────┘

Action: Optimize line 15 (consider using sparse matrices or reducing array size)
"""
```

---

## Memory Optimization Patterns

### Generator-Based Processing (Python)

```python
import csv
from typing import Iterator, Dict


# ❌ BAD: Load entire file into memory
def process_csv_bad(filename: str) -> list:
    """Loads entire CSV file into memory (memory inefficient)."""
    with open(filename, 'r') as f:
        reader = csv.DictReader(f)
        # This loads ALL rows into memory at once
        return [row for row in reader if float(row['amount']) > 100]


# ✅ GOOD: Use generator for streaming processing
def process_csv_good(filename: str) -> Iterator[Dict[str, str]]:
    """Streams CSV file row-by-row (memory efficient)."""
    with open(filename, 'r') as f:
        reader = csv.DictReader(f)
        for row in reader:
            if float(row['amount']) > 100:
                yield row


# Usage comparison
if __name__ == "__main__":
    # Bad approach: 1GB file → 1GB+ memory usage
    # results = process_csv_bad('large_transactions.csv')

    # Good approach: 1GB file → ~100KB memory usage
    for transaction in process_csv_good('large_transactions.csv'):
        print(f"Processing: {transaction['id']}")
        # Process one row at a time (memory stays constant)
```

### Lazy Evaluation with Properties

```python
class DataProcessor:
    """Example of lazy-loading heavy computations."""

    def __init__(self, data_path: str):
        self.data_path = data_path
        self._processed_data = None  # Cache
        self._statistics = None      # Cache

    @property
    def processed_data(self) -> np.ndarray:
        """Lazy-load and cache processed data."""
        if self._processed_data is None:
            print("Loading data (expensive)...")
            raw_data = np.load(self.data_path)
            self._processed_data = self._expensive_transformation(raw_data)
        return self._processed_data

    @property
    def statistics(self) -> dict:
        """Lazy-compute statistics only when needed."""
        if self._statistics is None:
            print("Computing statistics (expensive)...")
            data = self.processed_data  # Uses cached data if available
            self._statistics = {
                'mean': np.mean(data),
                'std': np.std(data),
                'median': np.median(data),
            }
        return self._statistics

    def _expensive_transformation(self, data: np.ndarray) -> np.ndarray:
        """Simulate expensive processing."""
        import time
        time.sleep(2)  # Simulate computation
        return data * 2 + 1


# Usage
processor = DataProcessor('data.npy')

# No data loaded yet (instant)
print("Processor created")

# First access: loads and processes data (slow)
mean_value = processor.statistics['mean']

# Second access: uses cached data (instant)
median_value = processor.statistics['median']
```

---

## Database Query Optimization

### Fixing N+1 Query Problem (SQLAlchemy)

```python
from sqlalchemy import Column, Integer, String, ForeignKey, create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import relationship, sessionmaker, joinedload

Base = declarative_base()


class User(Base):
    __tablename__ = 'users'
    id = Column(Integer, primary_key=True)
    name = Column(String(100))
    posts = relationship('Post', back_populates='user')


class Post(Base):
    __tablename__ = 'posts'
    id = Column(Integer, primary_key=True)
    title = Column(String(200))
    user_id = Column(Integer, ForeignKey('users.id'))
    user = relationship('User', back_populates='posts')


# ❌ BAD: N+1 Query Problem
def get_users_with_posts_bad(session):
    """
    Problem: Executes 1 query for users + N queries for posts (one per user).
    If 100 users → 101 queries!
    """
    users = session.query(User).all()  # 1 query

    for user in users:
        # Each access triggers a separate query (N queries)
        print(f"{user.name}: {len(user.posts)} posts")

    # Total: 1 + N queries


# ✅ GOOD: Eager Loading with joinedload
def get_users_with_posts_good(session):
    """
    Solution: Single JOIN query fetches users + posts together.
    100 users → 1 query!
    """
    users = session.query(User).options(
        joinedload(User.posts)  # Eager load posts in same query
    ).all()

    for user in users:
        # No additional queries (data already loaded)
        print(f"{user.name}: {len(user.posts)} posts")

    # Total: 1 query


# Usage with timing
if __name__ == "__main__":
    import time

    engine = create_engine('postgresql://user:pass@localhost/db')
    Session = sessionmaker(bind=engine)
    session = Session()

    # Bad approach
    start = time.time()
    get_users_with_posts_bad(session)
    print(f"N+1 approach: {time.time() - start:.2f}s")  # ~5.2s for 100 users

    # Good approach
    start = time.time()
    get_users_with_posts_good(session)
    print(f"Eager loading: {time.time() - start:.2f}s")  # ~0.1s for 100 users
```

### Database Indexing Example

```sql
-- Before optimization: Slow query
SELECT * FROM users WHERE email = 'user@example.com';
-- Query time: 850ms (table scan on 1M rows)

-- Create index on frequently queried column
CREATE INDEX idx_users_email ON users(email);

-- After optimization: Fast query
SELECT * FROM users WHERE email = 'user@example.com';
-- Query time: 2ms (index lookup)

-- Composite index for multi-column queries
CREATE INDEX idx_users_status_created ON users(status, created_at);

-- Now this query is fast
SELECT * FROM users WHERE status = 'active' AND created_at > '2024-01-01';

-- Check index usage (PostgreSQL)
EXPLAIN ANALYZE SELECT * FROM users WHERE email = 'user@example.com';

-- Output shows if index is used:
-- Index Scan using idx_users_email on users  (cost=0.42..8.44 rows=1 width=100)
```

---

## Caching Strategies

### In-Memory Cache with LRU (Python)

```python
from functools import lru_cache
import time
from typing import Dict


# ✅ Simple LRU Cache (Standard Library)
@lru_cache(maxsize=128)
def expensive_computation(x: int) -> int:
    """Cache results of expensive function."""
    print(f"Computing for {x}...")
    time.sleep(1)  # Simulate expensive operation
    return x ** 2 + x + 1


# Usage
result1 = expensive_computation(5)  # Takes 1s (cache miss)
result2 = expensive_computation(5)  # Instant (cache hit)
result3 = expensive_computation(10) # Takes 1s (cache miss)

# Check cache stats
print(expensive_computation.cache_info())
# CacheInfo(hits=1, misses=2, maxsize=128, currsize=2)


# ✅ Custom Cache with TTL (Time-To-Live)
class TTLCache:
    """Cache with expiration time."""

    def __init__(self, ttl_seconds: int = 300):
        self.ttl_seconds = ttl_seconds
        self.cache: Dict[str, tuple] = {}

    def get(self, key: str):
        """Get cached value if not expired."""
        if key in self.cache:
            value, timestamp = self.cache[key]
            if time.time() - timestamp < self.ttl_seconds:
                return value
            else:
                # Expired, remove from cache
                del self.cache[key]
        return None

    def set(self, key: str, value):
        """Set cached value with current timestamp."""
        self.cache[key] = (value, time.time())

    def clear(self):
        """Clear all cached values."""
        self.cache.clear()


# Usage
cache = TTLCache(ttl_seconds=60)  # 60-second expiration

def get_user_data(user_id: str) -> dict:
    """Fetch user data with caching."""
    cached = cache.get(user_id)
    if cached:
        print(f"Cache hit for user {user_id}")
        return cached

    print(f"Cache miss, fetching user {user_id}")
    # Simulate expensive database query
    user_data = {'id': user_id, 'name': f'User {user_id}'}

    cache.set(user_id, user_data)
    return user_data


# Test caching
user = get_user_data('123')  # Cache miss
user = get_user_data('123')  # Cache hit (instant)
time.sleep(61)
user = get_user_data('123')  # Cache miss (expired after 60s)
```

### Redis Cache Integration

```python
import redis
import json
from typing import Optional


class RedisCache:
    """Redis-backed cache for distributed systems."""

    def __init__(self, host: str = 'localhost', port: int = 6379, db: int = 0):
        self.client = redis.Redis(
            host=host,
            port=port,
            db=db,
            decode_responses=True
        )

    def get(self, key: str) -> Optional[dict]:
        """Get cached value from Redis."""
        value = self.client.get(key)
        if value:
            return json.loads(value)
        return None

    def set(self, key: str, value: dict, ttl_seconds: int = 3600):
        """Set cached value in Redis with TTL."""
        serialized = json.dumps(value)
        self.client.setex(key, ttl_seconds, serialized)

    def delete(self, key: str):
        """Delete cached value."""
        self.client.delete(key)

    def exists(self, key: str) -> bool:
        """Check if key exists in cache."""
        return bool(self.client.exists(key))


# Usage Example
cache = RedisCache()


def get_product_details(product_id: str) -> dict:
    """Fetch product details with Redis caching."""
    cache_key = f"product:{product_id}"

    # Try cache first
    cached = cache.get(cache_key)
    if cached:
        print(f"✓ Cache hit: {product_id}")
        return cached

    print(f"✗ Cache miss: {product_id}")

    # Simulate expensive database query
    product = {
        'id': product_id,
        'name': f'Product {product_id}',
        'price': 99.99,
    }

    # Store in cache (1 hour TTL)
    cache.set(cache_key, product, ttl_seconds=3600)

    return product


# Test
product1 = get_product_details('ABC123')  # Cache miss → DB query
product2 = get_product_details('ABC123')  # Cache hit → Instant
```

---

## Async I/O Optimization (Node.js)

### Parallel vs Sequential Async Operations

```typescript
import fs from "fs/promises";
import fetch from "node-fetch";

// ❌ BAD: Sequential async operations
async function fetchUserDataSequential(userIds: string[]): Promise<any[]> {
  const results = [];

  for (const userId of userIds) {
    // Each request waits for previous to complete (slow!)
    const response = await fetch(`https://api.example.com/users/${userId}`);
    const data = await response.json();
    results.push(data);
  }

  return results;
  // Time for 10 users with 100ms latency each: 1000ms
}

// ✅ GOOD: Parallel async operations
async function fetchUserDataParallel(userIds: string[]): Promise<any[]> {
  // Launch all requests simultaneously
  const promises = userIds.map(async (userId) => {
    const response = await fetch(`https://api.example.com/users/${userId}`);
    return response.json();
  });

  // Wait for all to complete
  return Promise.all(promises);
  // Time for 10 users with 100ms latency each: ~100ms (10x faster!)
}

// ✅ BETTER: Parallel with concurrency limit (prevent overwhelming server)
async function fetchUserDataBatched(
  userIds: string[],
  batchSize: number = 5
): Promise<any[]> {
  const results: any[] = [];

  for (let i = 0; i < userIds.length; i += batchSize) {
    const batch = userIds.slice(i, i + batchSize);

    // Process batch in parallel
    const batchResults = await Promise.all(
      batch.map(async (userId) => {
        const response = await fetch(`https://api.example.com/users/${userId}`);
        return response.json();
      })
    );

    results.push(...batchResults);
  }

  return results;
  // Controlled parallelism: Max 5 concurrent requests
}

// Usage
const userIds = Array.from({ length: 100 }, (_, i) => `user-${i}`);

console.time("sequential");
await fetchUserDataSequential(userIds.slice(0, 10));
console.timeEnd("sequential"); // ~1000ms

console.time("parallel");
await fetchUserDataParallel(userIds.slice(0, 10));
console.timeEnd("parallel"); // ~100ms

console.time("batched");
await fetchUserDataBatched(userIds, 10);
console.timeEnd("batched"); // ~1000ms but safer
```

---

## Benchmarking

### Python Benchmarking with `timeit`

```python
import timeit
import statistics


def benchmark_function(func, *args, iterations=1000):
    """Benchmark function execution time."""
    # Warm-up run
    func(*args)

    # Run benchmark
    times = timeit.repeat(
        lambda: func(*args),
        repeat=5,
        number=iterations
    )

    # Calculate statistics
    mean_time = statistics.mean(times) / iterations
    median_time = statistics.median(times) / iterations
    stdev_time = statistics.stdev(times) / iterations

    print(f"Function: {func.__name__}")
    print(f"  Mean:   {mean_time * 1000:.3f}ms")
    print(f"  Median: {median_time * 1000:.3f}ms")
    print(f"  StdDev: {stdev_time * 1000:.3f}ms")
    print(f"  Iterations: {iterations}")

    return mean_time


# Example: Compare list comprehension vs map
def using_list_comp(n):
    return [i**2 for i in range(n)]

def using_map(n):
    return list(map(lambda i: i**2, range(n)))


# Benchmark both approaches
print("=== Benchmarking ===")
time1 = benchmark_function(using_list_comp, 10000)
time2 = benchmark_function(using_map, 10000)

speedup = time2 / time1
print(f"\nSpeedup: {speedup:.2f}x")
```

---

**See also**: [reference.md](./reference.md) for APM tools and production monitoring strategies
