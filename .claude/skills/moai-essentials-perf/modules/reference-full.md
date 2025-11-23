# Performance Optimization Reference Guide

## Performance Tuning Checklist (40+ Items)

### CPU Optimization Checklist

- [ ] Profile CPU usage with Scalene
- [ ] Identify functions consuming >5% CPU
- [ ] Optimize algorithmic complexity (O(n²) → O(n log n))
- [ ] Use appropriate data structures (dict vs list for lookups)
- [ ] Implement caching for frequently called functions
- [ ] Avoid unnecessary computations in loops
- [ ] Use list comprehensions instead of loops where appropriate
- [ ] Leverage built-in functions (sum, max, min) instead of manual loops
- [ ] Use generators for large datasets
- [ ] Enable JIT compilation (PyPy for Python)
- [ ] Profile with cProfile for detailed call graphs
- [ ] Optimize hot paths identified by profiler
- [ ] Use multiprocessing for CPU-bound tasks
- [ ] Implement thread pooling for I/O-bound concurrency

### Memory Optimization Checklist

- [ ] Profile memory usage with Scalene or memray
- [ ] Use __slots__ for classes with many instances
- [ ] Replace lists with arrays for homogeneous numeric data
- [ ] Implement object pooling for frequently created objects
- [ ] Use generators instead of lists for large datasets
- [ ] Clear large data structures when no longer needed
- [ ] Avoid circular references that prevent garbage collection
- [ ] Use weak references where appropriate
- [ ] Compress data in memory for large datasets
- [ ] Stream process large files instead of loading entirely
- [ ] Monitor memory leaks with continuous profiling
- [ ] Set appropriate cache sizes with TTL
- [ ] Use memory-mapped files for very large datasets
- [ ] Implement pagination for database queries

### Network Optimization Checklist

- [ ] Implement connection pooling
- [ ] Use HTTP keep-alive connections
- [ ] Enable compression (gzip) for HTTP responses
- [ ] Batch API requests where possible
- [ ] Implement request coalescing
- [ ] Use CDN for static assets
- [ ] Enable HTTP/2 or HTTP/3
- [ ] Implement retry logic with exponential backoff
- [ ] Use async/await for concurrent network operations
- [ ] Set appropriate timeouts
- [ ] Implement circuit breakers for external services
- [ ] Cache DNS lookups
- [ ] Use connection reuse
- [ ] Monitor network latency and throughput

### Database Optimization Checklist

- [ ] Add indexes for frequently queried columns
- [ ] Use prepared statements
- [ ] Implement query result caching
- [ ] Optimize JOIN operations
- [ ] Use connection pooling
- [ ] Batch database operations
- [ ] Avoid N+1 query problems
- [ ] Use appropriate isolation levels
- [ ] Implement read replicas for read-heavy workloads
- [ ] Partition large tables
- [ ] Analyze and optimize slow queries
- [ ] Use database query profiling tools

## Performance Metrics Reference

### CPU Metrics

| Metric | Description | Target | Tool |
|--------|-------------|--------|------|
| CPU Usage % | Percentage of CPU utilized | <70% | psutil, top |
| Context Switches | Number of context switches/sec | <10000 | vmstat |
| CPU Time | Time spent in CPU cycles | Minimize | cProfile |
| IPC | Instructions per cycle | Maximize | perf |
| Cache Misses | L1/L2/L3 cache miss rate | <5% | perf, cachegrind |

### Memory Metrics

| Metric | Description | Target | Tool |
|--------|-------------|--------|------|
| Heap Usage | Memory allocated to heap | Monitor growth | psutil |
| RSS | Resident Set Size | <80% available | psutil |
| Peak Memory | Maximum memory used | Monitor | Scalene, memray |
| Allocation Rate | Memory allocations/sec | Minimize | Scalene |
| Memory Leaks | Unbounded growth | 0 leaks | memray, valgrind |
| GC Pauses | Garbage collection pauses | <100ms | gc module |

### Network Metrics

| Metric | Description | Target | Tool |
|--------|-------------|--------|------|
| Latency | Time to first byte | <50ms | curl, httpstat |
| Throughput | Requests per second | Maximize | ab, wrk |
| Bandwidth | Data transfer rate | <80% capacity | iftop, nload |
| Connection Pool | Active connections | Optimize | monitoring |
| TCP Retransmits | Packet retransmissions | <0.1% | netstat |

### Application Metrics

| Metric | Description | Target | Tool |
|--------|-------------|--------|------|
| Response Time P50 | Median response time | <100ms | Application logs |
| Response Time P95 | 95th percentile | <200ms | Prometheus |
| Response Time P99 | 99th percentile | <500ms | Prometheus |
| Error Rate | Percentage of failures | <0.1% | Monitoring |
| Saturation | Resource utilization | <80% | Multiple tools |

## Performance Optimization Patterns

### Pattern 1: Lazy Loading

```python
class LazyLoader:
    """Load resources only when needed."""

    def __init__(self):
        self._data = None

    @property
    def data(self):
        """Lazy load data on first access."""
        if self._data is None:
            self._data = self.load_expensive_data()
        return self._data

    def load_expensive_data(self):
        # Expensive operation
        return load_data()
```

### Pattern 2: Memoization

```python
from functools import lru_cache

class Memoizer:
    """Cache function results."""

    @lru_cache(maxsize=1000)
    def expensive_function(self, x: int) -> int:
        """Cached expensive computation."""
        return x ** 2
```

### Pattern 3: Batch Processing

```python
class BatchProcessor:
    """Process items in batches."""

    async def process_batch(self, items: List[Item]) -> List[Result]:
        """Process items in optimized batches."""

        batch_size = 100
        results = []

        for i in range(0, len(items), batch_size):
            batch = items[i:i + batch_size]
            batch_results = await self.process_items(batch)
            results.extend(batch_results)

        return results
```

### Pattern 4: Connection Pooling

```python
class ConnectionPool:
    """Reuse connections efficiently."""

    def __init__(self, max_connections: int = 20):
        self.pool = Queue(maxsize=max_connections)
        self.initialize_pool()

    def get_connection(self):
        """Get connection from pool."""
        return self.pool.get()

    def release_connection(self, conn):
        """Return connection to pool."""
        self.pool.put(conn)
```

### Pattern 5: Async/Await

```python
class AsyncProcessor:
    """Use async for concurrent operations."""

    async def process_concurrently(self, items: List[Item]) -> List[Result]:
        """Process items concurrently."""

        tasks = [self.process_item(item) for item in items]
        return await asyncio.gather(*tasks)
```

## Performance Testing Tools

### Profiling Tools

| Tool | Purpose | Language | Use Case |
|------|---------|----------|----------|
| Scalene | CPU/GPU/Memory profiling | Python | Comprehensive profiling |
| cProfile | Function-level profiling | Python | Detailed call graphs |
| memray | Memory profiling | Python | Memory leak detection |
| perf | System-level profiling | Linux | Hardware metrics |
| py-spy | Sampling profiler | Python | Production profiling |

### Load Testing Tools

| Tool | Purpose | Protocol | Use Case |
|------|---------|----------|----------|
| Locust | Load testing | HTTP/WebSocket | User simulation |
| Apache JMeter | Load testing | HTTP/JDBC/FTP | Enterprise testing |
| k6 | Load testing | HTTP/gRPC | Modern protocols |
| wrk | Benchmarking | HTTP | High-performance benchmarking |
| ab (Apache Bench) | Benchmarking | HTTP | Quick benchmarks |

### Monitoring Tools

| Tool | Purpose | Type | Use Case |
|------|---------|------|----------|
| Prometheus | Metrics collection | Time-series DB | System monitoring |
| Grafana | Visualization | Dashboard | Metrics visualization |
| Jaeger | Distributed tracing | Tracing | Microservices |
| ELK Stack | Log aggregation | Logging | Centralized logging |
| New Relic | APM | SaaS | Application performance |

## Common Performance Anti-Patterns

### Anti-Pattern 1: N+1 Query Problem

```python
# WRONG: N+1 queries
for user in users:
    profile = db.query("SELECT * FROM profiles WHERE user_id = ?", user.id)

# CORRECT: Single query
user_ids = [user.id for user in users]
profiles = db.query("SELECT * FROM profiles WHERE user_id IN (?)", user_ids)
```

### Anti-Pattern 2: Unbounded Cache

```python
# WRONG: No cache eviction
cache = {}
def cache_item(key, value):
    cache[key] = value  # Grows indefinitely

# CORRECT: LRU cache with size limit
from functools import lru_cache

@lru_cache(maxsize=1000)
def cached_function(x):
    return x ** 2
```

### Anti-Pattern 3: Synchronous I/O in Loop

```python
# WRONG: Blocking I/O in loop
for url in urls:
    response = requests.get(url)  # Blocks execution

# CORRECT: Async I/O
async def fetch_urls(urls):
    tasks = [fetch(url) for url in urls]
    return await asyncio.gather(*tasks)
```

### Anti-Pattern 4: String Concatenation in Loop

```python
# WRONG: String concatenation (creates new string each iteration)
result = ""
for item in items:
    result += str(item)  # O(n²) complexity

# CORRECT: Use join (O(n) complexity)
result = "".join(str(item) for item in items)
```

### Anti-Pattern 5: Premature Optimization

```python
# WRONG: Optimizing before profiling
# Spending time optimizing code that runs 0.001% of total time

# CORRECT: Profile first, then optimize
# 1. Run profiler (Scalene, cProfile)
# 2. Identify bottlenecks (>5% CPU/memory)
# 3. Optimize only high-impact areas
# 4. Measure improvement
```

## Performance Benchmarks

### Typical Performance Targets

| Operation | Target | Tool |
|-----------|--------|------|
| API Response Time (P95) | <200ms | Load testing |
| Database Query | <50ms | Query profiling |
| Memory Usage | <2GB per service | Memory profiling |
| CPU Usage | <70% average | System monitoring |
| Throughput | >1000 req/s | Load testing |
| Error Rate | <0.1% | Monitoring |

### Success Criteria

- ✅ 60% average performance improvement with optimization
- ✅ 95% accuracy in bottleneck detection
- ✅ 85% success rate for applied optimizations
- ✅ 90% reduction in memory usage for optimized code
- ✅ 50% improvement in response time after optimization

---

**Last Updated**: 2025-11-23
**Status**: Production Ready
**Lines**: 310
**Code Examples**: 5+ practical patterns and anti-patterns
