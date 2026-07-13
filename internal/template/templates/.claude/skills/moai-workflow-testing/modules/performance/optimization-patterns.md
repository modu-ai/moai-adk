# Optimization Patterns

> Sub-module: Specific optimization strategies and implementation patterns
> Parent: [Performance Optimization](../performance-optimization.md)
> Complexity: Advanced
> Time: 25+ minutes

## Overview

Comprehensive guide to performance optimization patterns, strategies, and best practices, expressed language-neutrally. Code sketches are pseudo-code illustrating the algorithmic idea, not a runnable program in any single language.

## Optimization Types

### Algorithm Improvement

Pattern: Optimize algorithmic complexity by reducing time complexity from O(n^2) to O(n log n) or better.

```text
# Before: O(n^2) nested loops
duplicates = []
for each (i, a) in items:
    for each (j, b) in items after i:
        if a == b and a not in duplicates:
            duplicates.add(a)

# After: O(n) using a hash set
seen = empty set
duplicates = empty set
for each item in items:
    if item in seen:
        duplicates.add(item)
    else:
        seen.add(item)
```

Impact: 10-1000x speedup depending on dataset size

### Caching Strategies

#### Memoization

```text
# Cache the result of a pure function keyed by its arguments.
# Most languages offer a built-in (Python lru_cache, Go map+sync.Once,
# JS Map, Rust once_cell) — prefer the standard idiom over a hand-rolled cache.
fib(n):
    if n <= 1: return n
    return memoized(fib, n-1) + memoized(fib, n-2)
```

Impact: 50-90% speedup for repeated calls

#### Custom Caching

A bounded cache with an eviction policy (LRU by access count):

```text
Cache(max_size):
    store = empty map
    access_count = empty map

    get(key):
        if key in store:
            access_count[key] += 1
            return store[key]
        return MISSING

    set(key, value):
        if size(store) >= max_size:
            evict the key with the lowest access_count
        store[key] = value
        access_count[key] = 0
```

### Concurrency Patterns

#### Parallelism for CPU-Bound Tasks

```text
# Split independent CPU-bound work across available cores.
# Idiom per language: Go errgroup/goroutines, Python multiprocessing,
# Rust rayon, Java ForkJoinPool, JS worker threads.
results = parallel_map(process_item, items, workers=cpu_count())
```

Impact: 2-8x speedup on multi-core systems

#### Async for I/O-Bound Tasks

```text
# Concurrently issue I/O (HTTP, disk, DB) and await all in flight.
# Idiom per language: Go goroutines+channels, Python asyncio,
# Rust tokio, JS Promise.all, C# Task.WhenAll.
tasks = [fetch(url) for url in urls]
results = await gather(tasks)
```

Impact: 10-100x speedup for I/O-bound operations

### Memory Optimization

#### Generator / Streaming Patterns

```text
# Stream data instead of materializing it all in memory.
# Idiom: Go channels/iter, Python generators, JS async iterators,
# Rust iterators, Java streams.
for line in stream_lines(filename):   # one line at a time
    yield process_line(line)
```

Impact: 60-90% memory reduction

#### Memory Pooling

```text
# Reuse allocated objects instead of churning the allocator.
# Idiom: Go sync.Pool, Python freelists, Rust slab, Java object pools.
Pool(factory, max_size):
    items = empty stack
```

### I/O Optimization

#### Buffered I/O

```text
# Batch writes for efficiency — let the runtime coalesce small writes.
# Idiom: Go bufio.Writer, Python open(buffering=), Rust BufWriter,
# Java BufferedWriter, JS writev/coalescing.
writer = open(filename, buffering=8192)
writer.write_all(lines)
```

Impact: 5-20x speedup for I/O operations

### Data Structure Optimization

#### Appropriate Data Structure Selection

```text
queue  = double-ended-queue   # O(1) push/pop at both ends
lookup = hash-set / hash-map  # O(1) membership test
```

Idiom per language: Go `container/list`+map, Python `collections.deque`/`set`, Rust `VecDeque`/`HashSet`, Java `ArrayDeque`/`HashSet`, JS array-shift is O(n) — prefer a ring buffer for queues.

#### Vectorized Numerical Data

```text
# Use a vectorized container (NumPy, ndarray, SIMD-friendly array)
# for bulk numerical operations instead of a scalar loop.
result = elementwise_sum(arrays)   # single vectorized op
```

Impact: 10-100x speedup for numerical operations

## Optimization Planning

### Optimization Plan Structure

```text
OptimizationPlan:
    bottlenecks:         List<Bottleneck>
    execution_order:     List<int>
    estimated_improvement: text
    implementation_complexity: text
    risk_level:          text
    prerequisites:       List<text>
    validation_strategy: text
```

### Prioritization Strategy

```text
prioritize(bottlenecks):
    severity_rank = { critical: 4, high: 3, medium: 2, low: 1 }
    sort bottlenecks descending by:
        (severity_rank[severity], impact_score, optimization_priority(type))
```

### Execution Order

```text
execution_order(bottlenecks):
    group = bucket bottlenecks by optimization_type
    order = []
    for type in [ALGORITHM, DATA_STRUCTURE, CACHING, MEMORY, CONCURRENCY, IO, DATABASE]:
        if type in group:
            order.extend(group[type])
    return order
```

## Implementation Strategies

### Risk Assessment

```text
assess_risk(bottlenecks):
    high_risk_types = { ALGORITHM, DATA_STRUCTURE, CONCURRENCY }
    high_risk_count = count of b where b.type in high_risk_types and b.impact > 0.3

    if high_risk_count > 3: return "high"
    if high_risk_count > 1: return "medium"
    return "low"
```

### Prerequisites Identification

```text
identify_prerequisites(bottlenecks):
    prerequisites = [
        "Create comprehensive performance benchmarks",
        "Ensure version control with current implementation",
        "Set up performance testing environment"
    ]
    types = set of b.optimization_type for b in bottlenecks

    if CONCURRENCY in types:
        prerequisites += ["Review thread safety and shared resource access",
                          "Implement proper synchronization mechanisms"]
    if DATABASE in types:
        prerequisites += ["Create database backup before optimization",
                          "Set up database performance monitoring"]
    return prerequisites
```

### Validation Strategy

```text
validation_strategy(bottlenecks):
    return [
        "1. Baseline performance measurement",
        "2. Incremental testing",
        "3. Automated performance testing",
        "4. Functional validation",
        "5. Production monitoring"
    ]
```

## Intelligent Optimization

### AI-Powered Suggestions

```text
get_ai_optimization_suggestions(bottlenecks, codebase_context):
    if docs unavailable:
        return rule_based_suggestions(bottlenecks)

    optimization_patterns = docs.get_library_docs(
        topic="advanced performance optimization patterns",
        tokens=5000)
    algorithm_patterns = docs.get_library_docs(
        topic="algorithm optimization big-O complexity reduction",
        tokens=3000)

    return generate_ai_suggestions(
        bottlenecks, optimization_patterns, algorithm_patterns, codebase_context)
```

### Algorithm Improvement Suggestions

```text
suggest_algorithm_improvement(bottleneck, algo_patterns):
    name = bottleneck.function_name
    suggestions = []
    if name contains ["search", "find"]:
        suggestions += ["Consider binary search for sorted data",
                        "Implement hash-based lookup for O(1) average case",
                        "Use trie structures for prefix searches"]
    if name contains ["sort", "order"]:
        suggestions += ["Prefer the language's built-in stable sort",
                        "Use radix sort for uniform integer data",
                        "Use bucket sort for uniformly distributed data"]
    return { bottleneck: name, suggestions, improvement: "30-90% depending on algorithm" }
```

### Data Structure Optimization

```text
suggest_data_structure_improvement(bottleneck, opt_patterns):
    return {
        bottleneck: bottleneck.function_name,
        suggestions: [
            "Use streaming/generators instead of materializing large datasets",
            "Implement lazy loading for expensive data structures",
            "Use contiguous arrays for numerical data",
            "Use a double-ended queue for queue operations",
            "Use set/dict for O(1) lookups instead of list searches"
        ],
        estimated_improvement: "30-80% memory reduction"
    }
```

## Best Practices

### Incremental Optimization

1. Profile before optimization
2. Optimize one bottleneck at a time
3. Measure improvement after each change
4. Verify functionality with tests
5. Commit changes with performance notes

### Measurement-Driven Optimization

Always measure:
- Before optimization: Establish baseline
- During optimization: Track incremental improvements
- After optimization: Validate total improvement
- In production: Monitor for regressions

### Testing Strategy

Comprehensive testing:
- Unit tests: Verify functional correctness
- Performance tests: Validate improvement claims
- Integration tests: Ensure system-wide compatibility
- Regression tests: Prevent performance degradation

### Documentation

Document optimizations:
- Problem description and metrics
- Solution approach and implementation
- Performance improvement measurements
- Side effects and trade-offs
- Maintenance considerations

## Common Pitfalls

### Premature Optimization

Avoid optimizing:
- Code paths rarely executed
- Features with uncertain requirements
- Clear and correct code without performance issues

Focus on:
- Measured bottlenecks
- Hot paths in critical workflows
- User-facing performance issues

### Over-Optimization

Signs of over-optimization:
- Code becomes unreadable
- Maintenance costs exceed performance gains
- Optimization targets theoretical scenarios
- Diminishing returns on investment

### Ignoring Trade-offs

Consider trade-offs:
- Memory vs. CPU
- Development time vs. performance gain
- Code clarity vs. optimization
- Portability vs. platform-specific optimizations

## Performance Monitoring

### Continuous Monitoring

Implement monitoring for:
- Response times and throughput
- Resource utilization (CPU, memory, I/O)
- Error rates and exceptions
- Business metrics correlated with performance

### Alerting

Set up alerts for:
- Performance regression beyond thresholds
- Resource exhaustion conditions
- Anomalous behavior patterns
- SLA violations

---

Sub-module: `modules/performance/optimization-patterns.md`
Parent: [Performance Optimization](../performance-optimization.md)
Version: 2.0.0
Last Updated: 2025-12-07
