# Performance Optimization Reference

_Last updated: 2025-10-22_

## Profiling Tools by Language (2025)

### Python
| Tool | Purpose | Command |
|------|---------|---------|
| cProfile | CPU profiling | `python -m cProfile -s cumtime script.py` |
| line_profiler | Line-by-line CPU | `kernprof -l -v script.py` |
| memory_profiler | Memory usage | `python -m memory_profiler script.py` |
| py-spy | Sampling profiler | `py-spy top --pid <PID>` |
| Scalene | CPU + memory + GPU | `scalene script.py` |

### JavaScript/TypeScript
| Tool | Purpose | Command |
|------|---------|---------|
| Chrome DevTools | Browser profiling | F12 → Performance tab |
| Node --prof | CPU profiling | `node --prof script.js` |
| clinic.js | Comprehensive | `clinic doctor -- node script.js` |
| 0x | Flame graphs | `0x script.js` |

### Go
| Tool | Purpose | Command |
|------|---------|---------|
| pprof | CPU/memory profiling | `go test -cpuprofile=cpu.prof` |
| trace | Execution tracing | `go test -trace=trace.out` |
| benchstat | Benchmark comparison | `benchstat old.txt new.txt` |

### Rust
| Tool | Purpose | Command |
|------|---------|---------|
| cargo-flamegraph | CPU flame graphs | `cargo flamegraph` |
| valgrind | Memory profiling | `valgrind --tool=massif ./target/release/app` |
| perf | System-level profiling | `perf record ./target/release/app` |

---

## Common Performance Bottlenecks

### 1. N+1 Query Problem
**Symptom**: Multiple database queries in loop

**Solution**: Use eager loading / JOIN
```python
# Bad
for user in users:
    posts = user.posts.all()  # N queries

# Good
users = User.objects.prefetch_related('posts').all()  # 2 queries
```

### 2. Unnecessary Computations
**Symptom**: Repeated expensive calculations

**Solution**: Memoization / caching
```python
from functools import lru_cache

@lru_cache(maxsize=128)
def expensive_function(n):
    # Cached after first call
    return heavy_computation(n)
```

### 3. Memory Leaks
**Symptom**: Increasing memory usage over time

**Solution**: Release references, use weak references
```python
import weakref

# Bad: Circular reference
class Node:
    def __init__(self):
        self.children = []  # Strong references

# Good: Weak references
class Node:
    def __init__(self):
        self._children = weakref.WeakSet()
```

### 4. Blocking I/O
**Symptom**: Slow API calls, file operations

**Solution**: Use async/await, concurrency
```python
import asyncio

async def fetch_all(urls):
    tasks = [fetch_url(url) for url in urls]
    return await asyncio.gather(*tasks)  # Parallel execution
```

---

## Optimization Strategies

### CPU Optimization
1. **Profile first**: Identify actual bottlenecks
2. **Algorithmic improvements**: Better time complexity (O(n²) → O(n log n))
3. **Reduce allocations**: Reuse objects, use object pools
4. **Parallelization**: Multi-threading, multi-processing
5. **JIT compilation**: Use PyPy, Numba for Python hot loops

### Memory Optimization
1. **Use generators**: Lazy evaluation for large datasets
2. **Batch processing**: Process data in chunks
3. **Efficient data structures**: Arrays instead of lists, sets for lookups
4. **Memory pools**: Reuse allocated memory
5. **Compression**: Store data compressed

### Database Optimization
1. **Indexing**: Add indexes on frequently queried columns
2. **Query optimization**: Use EXPLAIN to analyze queries
3. **Connection pooling**: Reuse database connections
4. **Caching**: Redis, Memcached for hot data
5. **Denormalization**: Strategic duplication for read-heavy workloads

---

## Performance Metrics

### Target Metrics

| Metric | Target | Measurement |
|--------|--------|-------------|
| API Response Time | <200ms | p95 latency |
| Database Query | <50ms | Average query time |
| Memory Usage | <1GB | RSS (Resident Set Size) |
| CPU Usage | <70% | Average CPU utilization |
| Throughput | >1000 req/s | Requests per second |

### Monitoring Tools
- **Prometheus**: Metrics collection
- **Grafana**: Visualization
- **New Relic**: APM (Application Performance Monitoring)
- **DataDog**: Infrastructure monitoring
- **Sentry**: Error tracking + performance

---

## Load Testing

### Tools
| Tool | Purpose | Command |
|------|---------|---------|
| k6 | Modern load testing | `k6 run script.js` |
| locust | Python-based | `locust -f locustfile.py` |
| Apache Bench | Simple HTTP testing | `ab -n 1000 -c 10 URL` |
| wrk | HTTP benchmarking | `wrk -t4 -c100 -d30s URL` |

### Example (k6)
```javascript
import http from 'k6/http';
import { check } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 100 },  // Ramp up
    { duration: '3m', target: 100 },  // Stay at 100
    { duration: '1m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<200'],  // 95% under 200ms
  },
};

export default function () {
  const res = http.get('https://api.example.com/users');
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
  });
}
```

---

## References

- [Python Performance Tips](https://wiki.python.org/moin/PythonSpeed/PerformanceTips)
- [Node.js Performance Best Practices](https://nodejs.org/en/docs/guides/simple-profiling/)
- [Chrome DevTools Performance](https://developer.chrome.com/docs/devtools/performance/)
- [Database Indexing Guide](https://use-the-index-luke.com/)

---

_For optimization examples, see examples.md_
