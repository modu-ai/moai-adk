---
name: moai-essentials-perf
version: 4.0.0
updated: 2025-11-20
status: stable
description: Performance optimization with profiling, memory analysis, and benchmarking
allowed-tools: [Read, Bash, WebSearch, WebFetch]
---

# Performance Optimization Expert

**Application Profiling & Performance Tuning**

> **Focus**: CPU/Memory Profiling, Benchmarking, Optimization Patterns  
> **Tools**: Python (Scalene, cProfile), Node.js (Clinic.js), Go (pprof)

---

## Overview

Systematic approach to identifying and fixing performance bottlenecks.

### Core Techniques

1.  **Profiling**: Measure CPU, memory, I/O usage.
2.  **Benchmarking**: Compare performance before/after changes.
3.  **Optimization**: Apply targeted improvements (caching, lazy loading, parallelism).

---

## Implementation Patterns

### 1. Python Profiling (Scalene)

**Installation**:

```bash
pip install scalene
```

**Usage**:

```bash
# Profile CPU, memory, and GPU
scalene my_script.py

# HTML report
scalene --html --outfile profile.html my_script.py
```

**Code-level profiling**:

```python
from scalene import scalene_profiler

scalene_profiler.start()
# Code to profile
result = expensive_computation()
scalene_profiler.stop()
```

### 2. Memory Optimization

**Common Issues**:

- Large data structures held in memory
- Circular references preventing GC
- Inefficient data structures (lists vs generators)

**Example (Python)**:

```python
# Bad: Load entire file into memory
data = [line for line in open('large_file.txt')]

# Good: Use generator
def read_lines(filename):
    with open(filename) as f:
        for line in f:
            yield line.strip()
```

### 3. Database Query Optimization

**N+1 Query Problem**:

```python
# Bad: N+1 queries
users = User.query.all()
for user in users:
    posts = Post.query.filter_by(user_id=user.id).all()  # N queries

# Good: Eager loading
users = User.query.options(joinedload(User.posts)).all()
```

**Indexing**:

```sql
-- Create index on frequently queried columns
CREATE INDEX idx_user_email ON users(email);
```

### 4. Caching Strategies

**In-Memory Cache (Python)**:

```python
from functools import lru_cache

@lru_cache(maxsize=128)
def expensive_function(x):
    # Expensive computation
    return result
```

**Redis Cache**:

```python
import redis

cache = redis.Redis(host='localhost', port=6379)

def get_data(key):
    cached = cache.get(key)
    if cached:
        return cached

    data = fetch_from_database(key)
    cache.setex(key, 3600, data)  # Cache for 1 hour
    return data
```

### 5. Async I/O (Node.js)

**Avoid blocking I/O**:

```javascript
// Bad: Synchronous
const data = fs.readFileSync("large_file.txt", "utf8");

// Good: Asynchronous
const data = await fs.promises.readFile("large_file.txt", "utf8");
```

---

## Best Practices

1.  **Measure First**: Don't optimize without profiling.
2.  **Target Hotspots**: Focus on the 20% causing 80% slowdown.
3.  **Benchmark**: Use realistic workloads, not microbenchmarks.
4.  **Monitor Production**: Use APM tools (New Relic, Datadog).

---

## Validation Checklist

- [ ] **Profiled**: Code profiled with appropriate tool?
- [ ] **Baseline**: Performance baseline established?
- [ ] **Optimized**: Hotspots identified and fixed?
- [ ] **Benchmarked**: Improvements measured with benchmarks?
- [ ] **Monitored**: Production metrics tracked?

---

## Related Skills

- `moai-domain-backend`: API optimization
- `moai-domain-monitoring`: APM setup
- `moai-devops-docker`: Container resource limits

---

**Last Updated**: 2025-11-20
