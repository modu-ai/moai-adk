---
name: moai-essentials-perf
description: Performance optimization with profiling, bottleneck detection, and tuning strategies
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
tier: 1
auto-load: "true"
---

# Alfred Performance Optimizer

## What it does

Performance analysis and optimization with profiling tools, bottleneck detection, and language-specific optimization techniques.

## When to use

- "성능 개선해줘", "느린 부분 찾아줘", "최적화 방법은?", "속도 향상", "응답 시간 단축"
- "프로파일링", "병목 지점", "메모리 누수", "CPU 사용률", "쿼리 최적화"
- "Performance tuning", "Profiling", "Bottleneck analysis", "Optimization"
- When application is slow or resource-intensive
- Before production release or scaling

## How it works

**Profiling Tools**:
- **Python**: cProfile, memory_profiler
- **TypeScript**: Chrome DevTools, clinic.js
- **Java**: JProfiler, VisualVM
- **Go**: pprof
- **Rust**: flamegraph, criterion

**Common Performance Issues**:
- **N+1 Query Problem**: Use eager loading/joins
- **Inefficient Loop**: O(n²) → O(n) with Set/Map
- **Memory Leak**: Remove event listeners, close connections

**Optimization Checklist**:
- [ ] Current performance benchmark
- [ ] Bottleneck identification
- [ ] Profiling data collected
- [ ] Algorithm complexity improved (O(n²) → O(n))
- [ ] Unnecessary operations removed
- [ ] Caching applied
- [ ] Async processing introduced
- [ ] Post-optimization benchmark
- [ ] Side effects checked

**Language-specific Optimizations**:
- **Python**: List comprehension, generators, @lru_cache
- **TypeScript**: Memoization, lazy loading, code splitting
- **Java**: Stream API, parallel processing
- **Go**: Goroutines, buffered channels
- **Rust**: Zero-cost abstractions, borrowing

**Performance Targets**:
- API response time: <200ms (P95)
- Page load time: <2s
- Memory usage: <512MB
- CPU usage: <70%

## Profiling Commands

### Python
```bash
# CPU profiling
python -m cProfile -o profile.stats script.py
python -m pstats profile.stats

# Memory profiling
python -m memory_profiler script.py

# Line-by-line profiling
kernprof -l -v script.py

# Flamegraph
py-spy record -o profile.svg -- python script.py
```

### TypeScript/JavaScript
```bash
# Node.js profiling
node --prof script.js
node --prof-process isolate-*.log > processed.txt

# Clinic.js (comprehensive)
clinic doctor -- node script.js
clinic flame -- node script.js

# Chrome DevTools
node --inspect script.js
# Open chrome://inspect
```

### Go
```bash
# CPU profiling
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# Memory profiling
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof

# Trace execution
go run -trace trace.out main.go
go tool trace trace.out
```

### Rust
```bash
# Flamegraph
cargo flamegraph

# Criterion benchmarking
cargo bench

# Memory profiling with valgrind
cargo build --release
valgrind --tool=massif target/release/app
```

### Java
```bash
# JProfiler
java -agentpath:/path/to/jprofiler/libjprofilerti.so MyApp

# VisualVM
jvisualvm

# JMH benchmarking
java -jar target/benchmarks.jar
```

## Optimization Patterns

### O(n²) → O(n) with Set/Map
```python
# ❌ Bad (O(n²))
def find_duplicates(list1, list2):
    duplicates = []
    for item in list1:
        if item in list2:  # O(n) lookup
            duplicates.append(item)
    return duplicates

# ✅ Good (O(n))
def find_duplicates(list1, list2):
    set2 = set(list2)  # O(n) creation
    return [item for item in list1 if item in set2]  # O(1) lookup
```

### N+1 Query → Eager Loading
```python
# ❌ Bad (N+1 queries)
users = User.query.all()  # 1 query
for user in users:
    posts = user.posts  # N queries

# ✅ Good (2 queries)
users = User.query.options(joinedload(User.posts)).all()
for user in users:
    posts = user.posts  # No additional query
```

### Memoization / Caching
```python
# ❌ Bad (recalculates every time)
def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)

# ✅ Good (caches results)
from functools import lru_cache

@lru_cache(maxsize=None)
def fibonacci(n):
    if n <= 1:
        return n
    return fibonacci(n-1) + fibonacci(n-2)
```

### Lazy Loading / Generator
```python
# ❌ Bad (loads all data into memory)
def process_large_file(filename):
    with open(filename) as f:
        lines = f.readlines()  # Loads entire file
    return [process_line(line) for line in lines]

# ✅ Good (processes line by line)
def process_large_file(filename):
    with open(filename) as f:
        for line in f:  # Generator
            yield process_line(line)
```

### Async/Concurrent Processing
```python
# ❌ Bad (sequential, 5 seconds total)
import requests
urls = [url1, url2, url3, url4, url5]
results = [requests.get(url) for url in urls]  # 1s each

# ✅ Good (parallel, 1 second total)
import asyncio
import aiohttp

async def fetch_all(urls):
    async with aiohttp.ClientSession() as session:
        tasks = [session.get(url) for url in urls]
        return await asyncio.gather(*tasks)
```

## Examples

### Example 1: N+1 Query Optimization
User: "N+1 query 문제 해결해줘"

Alfred identifies:
```sql
-- N+1 Problem Detected
SELECT * FROM users LIMIT 100;  -- 1 query
SELECT * FROM posts WHERE user_id = 1;  -- 100 queries!

-- Optimized with JOIN
SELECT users.*, posts.*
FROM users
LEFT JOIN posts ON users.id = posts.user_id
LIMIT 100;  -- 1 query total
```

Result: **100x fewer queries, 50ms → 5ms response time**

### Example 2: Algorithm Complexity Improvement
User: "느린 부분 찾아줘"

Alfred profiles and finds:
```python
# Bottleneck: O(n²) loop (Line 45, takes 5s for 1000 items)
for i in range(len(items)):
    for j in range(len(items)):
        if items[i] == items[j]:
            # ...

# Optimized: O(n) with Set (takes 50ms)
seen = set()
duplicates = set()
for item in items:
    if item in seen:
        duplicates.add(item)
    seen.add(item)
```

Result: **100x faster, 5s → 50ms**

### Example 3: Memory Leak Detection
User: "메모리 계속 증가해"

Alfred profiles:
```bash
python -m memory_profiler app.py

Result:
Line 120: +500MB (global cache never cleared)
Line 250: +300MB (event listeners not removed)

Fix:
1. Implement LRU cache with max_size=1000
2. Add cleanup in __del__ method
```

Result: **Memory usage stable at 100MB**

### Example 4: Database Query Optimization
User: "쿼리 최적화해줘"

Alfred analyzes EXPLAIN:
```sql
-- Before (Full table scan, 2.5s)
SELECT * FROM orders
WHERE customer_id = 123
ORDER BY created_at DESC;

EXPLAIN: Seq Scan on orders (cost=0..10000)

-- After (Index scan, 25ms)
CREATE INDEX idx_orders_customer_created
ON orders (customer_id, created_at DESC);

EXPLAIN: Index Scan using idx_orders_customer_created
```

Result: **100x faster, 2.5s → 25ms**

### Example 5: Frontend Bundle Size Reduction
User: "페이지 로딩 느려"

Alfred analyzes:
```bash
# Before: 5MB bundle, 8s load time
webpack-bundle-analyzer

Result:
- moment.js: 500KB (unused locales)
- lodash: 300KB (importing entire library)

# After: Code splitting + tree shaking
import { debounce } from 'lodash-es'  # Only import what's needed
moment.locale('en')  # Single locale

Result: 2MB bundle, 2s load time
```

Result: **60% smaller bundle, 4x faster load**

## Performance Checklist

**Before Optimization**:
- [ ] Benchmark current performance (baseline)
- [ ] Profile to identify bottlenecks (don't guess!)
- [ ] Prioritize by impact (80/20 rule)

**During Optimization**:
- [ ] Change one thing at a time
- [ ] Measure after each change
- [ ] Document what you tried

**After Optimization**:
- [ ] Run tests (ensure correctness)
- [ ] Benchmark improvement
- [ ] Check for side effects (memory, CPU, I/O)
- [ ] Production testing (gradual rollout)

## Works well with

- moai-essentials-refactor (Optimize + Refactor)
- moai-foundation-trust (Performance tests in TRUST)
