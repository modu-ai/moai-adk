---
name: moai-quality-performance
description: "Performance consolidated: Scalene profiling, benchmarking, bundle analysis, Core Web Vitals"
version: 1.0.0
modularized: true
last_updated: 2025-11-24
allowed-tools:
  - Task
  - AskUserQuestion
  - Skill
  - Read
  - Write
  - Edit
compliance_score: 85
modules:
  - profiling-analysis
  - optimization-patterns
  - benchmarking
dependencies:
  - moai-foundation-trust
  - moai-lang-python
  - moai-lang-typescript
deprecated: false
successor: null
category_tier: 3
auto_trigger_keywords:
  - performance
  - optimization
  - profiling
  - scalene
  - memory
  - cpu
  - latency
  - throughput
  - benchmark
  - vitals
agent_coverage:
  - performance-engineer
  - quality-gate
context7_references:
  - scalene
  - lighthouse
invocation_api_version: "1.0"
---

## Quick Reference (30 seconds)

**Enterprise Performance Consolidated**

Unified performance framework consolidating essentials-perf and component-designer with Scalene profiling, benchmarking, Core Web Vitals optimization, and automated performance detection.

**Core Capabilities**:
- ✅ Scalene profiling (CPU, memory, GPU time)
- ✅ Python performance optimization
- ✅ Frontend performance (Lighthouse, CWV)
- ✅ Bundle size analysis
- ✅ Component performance optimization
- ✅ Benchmarking and stress testing
- ✅ Memory leak detection

**When to Use**:
- Profiling slow code sections
- Optimizing database queries
- Reducing bundle size
- Improving Core Web Vitals
- Memory leak detection
- Performance regression testing

**Core Framework**: PROFILE → ANALYZE → OPTIMIZE
```
1. Profile code with Scalene
   ↓
2. Analyze bottlenecks
   ↓
3. Implement optimizations
   ↓
4. Benchmark improvements
   ↓
5. Validate with metrics
```

---

## Core Patterns (5-10 minutes each)

### Pattern 1: Scalene Profiling for Python

**Concept**: Use Scalene to identify CPU-bound, GPU-bound, and memory-intensive code.

```python
# Install and run
# pip install scalene
# scalene script.py

# Example: Slow function
import time

def slow_function():
    """CPU-bound operation - will show high CPU time."""
    total = 0
    for i in range(1000000):
        total += i * i
    return total

def memory_intensive():
    """Memory-intensive operation."""
    large_list = [i for i in range(10000000)]
    return sum(large_list)

if __name__ == "__main__":
    result1 = slow_function()
    result2 = memory_intensive()
```

**Use Case**: Identify performance bottlenecks in Python code.

---

### Pattern 2: Core Web Vitals Optimization

**Concept**: Optimize Largest Contentful Paint (LCP), First Input Delay (FID), and Cumulative Layout Shift (CLS).

```typescript
// ❌ POOR: Large image causes slow LCP
export default function Hero() {
  return <img src="/hero.jpg" alt="Hero" width="1920" height="1080" />;
}

// ✅ GOOD: Optimized image with next/image
import Image from 'next/image';

export default function Hero() {
  return (
    <Image
      src="/hero.jpg"
      alt="Hero"
      width={1920}
      height={1080}
      priority
      placeholder="blur"
    />
  );
}

// Lighthouse score: 95+
```

**Use Case**: Improve Core Web Vitals (PageSpeed score).

---

### Pattern 3: Bundle Size Analysis

**Concept**: Analyze and reduce JavaScript bundle size.

```bash
# Analyze bundle
npm install --save-dev webpack-bundle-analyzer
npx webpack-bundle-analyzer dist/stats.json

# Output shows:
# - react: 42KB (can't optimize)
# - lodash: 71KB (replace with lodash-es or alternatives)
# - moment: 67KB (replace with dayjs or date-fns)
# - Total: 342KB → reduce to 200KB with optimizations
```

```typescript
// ❌ POOR: Imports entire lodash
import { map, filter } from 'lodash';

// ✅ GOOD: Imports individual functions
import map from 'lodash/map';
import filter from 'lodash/filter';

// ✅ BETTER: Use alternatives with smaller size
import { map, filter } from 'lodash-es';
```

**Use Case**: Reduce bundle size from 342KB to <200KB.

---

### Pattern 4: Database Query Optimization

**Concept**: Identify and optimize slow database queries.

```python
# ❌ POOR: N+1 query problem
users = User.query.all()
for user in users:
    # This runs a query for each user (1 + N queries)
    posts = user.posts
    print(f"{user.name}: {len(posts)} posts")

# ✅ GOOD: Eager loading with join
from sqlalchemy.orm import joinedload

users = User.query.options(joinedload(User.posts)).all()
for user in users:
    # Posts already loaded (1 query total)
    posts = user.posts
    print(f"{user.name}: {len(posts)} posts")
```

**Use Case**: Reduce query count from 1000+ to <10.

---

### Pattern 5: Benchmarking and Regression Testing

**Concept**: Measure performance and detect regressions.

```python
import pytest
from benchmark import Timer

class TestPerformance:
    """Performance benchmarks with regression detection."""

    def test_auth_latency(self, benchmark):
        """Authenticate must complete < 50ms."""
        result = benchmark(authenticate, username="alice", password="secret")
        assert benchmark.stats.median < 0.050  # 50ms threshold

    def test_query_performance(self, benchmark):
        """Query 1000 users must complete < 100ms."""
        result = benchmark(
            lambda: User.query.limit(1000).all(),
            min_rounds=5
        )
        assert benchmark.stats.median < 0.100  # 100ms threshold

# Run: pytest --benchmark-only tests/test_performance.py
```

**Use Case**: Detect performance regressions before deployment.

---

## Advanced Documentation

For detailed performance patterns:

- **[modules/profiling-analysis.md](modules/profiling-analysis.md)** - Scalene, cProfile, Memory Profiler
- **[modules/optimization-patterns.md](modules/optimization-patterns.md)** - Query, bundle, and code optimization
- **[modules/benchmarking.md](modules/benchmarking.md)** - Benchmarking and regression testing

---

## Best Practices

### ✅ DO
- Profile before optimizing (measure, don't guess)
- Optimize hot paths (80/20 rule)
- Cache frequently accessed data
- Use indexes for database queries
- Implement lazy loading for images
- Monitor Core Web Vitals
- Benchmark after changes
- Test on production-like hardware

### ❌ DON'T
- Premature optimization (optimize later)
- Ignore profiling results
- Cache without TTL
- Create full-text indexes for small tables
- Load unused libraries
- Ignore bundle size warnings
- Assume local performance = production

---

**Status**: Production Ready
**Generated with**: MoAI-ADK Skill Factory
