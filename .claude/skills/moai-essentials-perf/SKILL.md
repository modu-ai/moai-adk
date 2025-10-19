---
name: moai-essentials-perf
description: Performance optimization with profiling, bottleneck detection, and tuning strategies
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# Alfred Performance Optimizer

## What it does

Performance analysis and optimization with profiling tools, bottleneck detection, and language-specific optimization techniques.

## When to use

- "성능 개선해줘", "느린 부분 찾아줘", "최적화 방법은?"
- "프로파일링", "병목 지점", "메모리 누수"

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

## Performance Profiling Commands

| Language | Tool | Command | Output |
|----------|------|---------|--------|
| **Python** | cProfile | `python -m cProfile -s cumtime script.py` | Function time ranking |
| **Python** | memory_profiler | `python -m memory_profiler script.py` | Line-by-line memory |
| **TypeScript** | clinic.js | `clinic doctor -- node app.js` | HTML report |
| **TypeScript** | Node profiler | `node --prof app.js && node --prof-process` | Flame graph |
| **Java** | JProfiler | GUI tool | CPU, memory, threads |
| **Java** | async-profiler | `java -jar async-profiler.jar -d 30 ...` | Native code analysis |
| **Go** | pprof | `go tool pprof http://localhost:6060/debug/pprof/profile` | CPU profiling |
| **Go** | pprof memory | `go tool pprof http://localhost:6060/debug/pprof/heap` | Memory leaks |
| **Rust** | flamegraph | `cargo install flamegraph && cargo flamegraph` | Visualization |

## Examples

### Example 1: N+1 Query Problem → Eager Loading with JOIN

**❌ Before (N+1 Problem)**:
```python
# @CODE:QUERY-OPTIMIZATION-001: N+1 쿼리 문제
def get_orders_with_customers(limit: int):
    """고객 주문 목록 조회 (느림)"""
    orders = db.query("SELECT * FROM orders LIMIT %s", limit)

    # 문제: 각 주문마다 고객 정보 조회 → N+1 쿼리
    for order in orders:
        order.customer = db.query(
            "SELECT * FROM customers WHERE id = %s",
            order.customer_id
        )[0]

    return orders

# 성능: 1 + 100 쿼리 = 5.2초 ⏱️
```

**✅ After (JOIN 최적화)**:
```python
# @CODE:QUERY-OPTIMIZATION-001: JOIN으로 개선
def get_orders_with_customers(limit: int):
    query = """
        SELECT o.*, c.* FROM orders o
        INNER JOIN customers c ON o.customer_id = c.id
        LIMIT %s
    """
    return db.query(query, limit)

# 성능: 1 쿼리 = 0.3초 ✅ (94% 개선!)
```

### Example 2: Algorithm O(n²) → O(n)

**❌ Before**:
```typescript
// 중첩 루프 = O(n²)
function findDuplicates(arr: number[]): number[] {
    const dups: number[] = [];
    for (let i = 0; i < arr.length; i++) {
        for (let j = i + 1; j < arr.length; j++) {
            if (arr[i] === arr[j]) dups.push(arr[i]);
        }
    }
    return dups;
}

// 100,000 items: 7,800ms 💥
```

**✅ After**:
```typescript
// Set 사용 = O(n)
function findDuplicates(arr: number[]): number[] {
    const seen = new Set();
    const dups = new Set();
    for (const num of arr) {
        if (seen.has(num)) dups.add(num);
        else seen.add(num);
    }
    return Array.from(dups);
}

// 100,000 items: 10ms ✅ (780배 빠름!)
```

### Example 3: Memory Leak → Proper Cleanup

**❌ Before**:
```go
// 메모리 누수: 리스너 제거 없음
var listeners []func()
func subscribe(fn func()) {
    listeners = append(listeners, fn)  // 무한 누적
}
// 1시간 후: 50MB → 5GB 💥
```

**✅ After**:
```go
// 구독 해제 구현
func unsubscribe(index int) {
    listeners = append(listeners[:index], listeners[index+1:]...)
}
// 메모리 안정적 유지 ✅
```

### Example 4: Database Index Missing

**❌ Before (500ms)**:
```sql
SELECT * FROM orders WHERE customer_id = 123;
-- Table Scan: 전체 1M 행 스캔
```

**✅ After (2ms)**:
```sql
CREATE INDEX idx_customer_id ON orders(customer_id);
-- Index Scan: 250배 빠름 ✅
```

### Example 5: Caching → Response Time

**❌ Before (100ms per request)**:
```python
def get_user(id):
    return db.query("SELECT * FROM users WHERE id = %s", id)

# 100 concurrent: 10,000ms 💥
```

**✅ After (1ms cache hit)**:
```python
import redis
cache = redis.Redis()

def get_user(id):
    if cached := cache.get(f"user:{id}"):
        return json.loads(cached)  # 1ms ✅

    user = db.query("SELECT * FROM users WHERE id = %s", id)
    cache.setex(f"user:{id}", 3600, json.dumps(user))
    return user

# 100 concurrent with 90% cache hit: 110ms ✅
# 응답시간: 8초 → 5ms (1600배!)
```

## Keywords

"성능 최적화", "N+1 쿼리", "병목 지점", "프로파일링", "메모리 누수", "캐싱", "알고리즘", "인덱싱", "performance profiling", "bottleneck detection", "O(n) complexity"

## Reference

- Performance profiling: `.moai/memory/development-guide.md#성능-최적화`
- Caching strategies: CLAUDE.md#캐싱-전략
- Optimization techniques: `.moai/memory/development-guide.md#알고리즘-최적화`

## Works well with

- moai-essentials-review (코드 품질 검증)
- moai-essentials-refactor (리팩토링)
- moai-essentials-debug (성능 오류 디버깅)
