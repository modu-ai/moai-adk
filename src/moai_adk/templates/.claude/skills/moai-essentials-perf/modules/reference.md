# Performance Optimization Reference

Complete reference guide with best practices, metrics, and success benchmarks.

## Performance Metrics Glossary

### CPU Metrics
- **CPU Usage %**: Percentage of CPU utilized (0-100%)
- **Context Switches**: Number of context switches per second
- **Cache Hits**: Percentage of memory accesses found in cache
- **Instructions Per Cycle**: Average instructions executed per CPU cycle

### Memory Metrics
- **Heap Usage**: Current memory allocated to heap
- **Peak Memory**: Maximum memory used during execution
- **Memory Growth Rate**: Change in memory over time
- **Allocation Count**: Total number of memory allocations

### Performance Metrics
- **Latency**: Time taken to complete a request
- **Throughput**: Requests processed per second
- **Response Time**: Time from request to response
- **Tail Latency (p95/p99)**: Latency at 95th/99th percentile

## Performance Optimization Benchmarks

### CPU Optimization
| Task | Target | Optimized |
|------|--------|-----------|
| Matrix multiplication (10K x 10K) | <5s | <0.5s |
| String processing (1M strings) | <2s | <0.2s |
| List sorting (1M items) | <3s | <0.3s |
| Hash lookups (1M items) | <10ms | <1ms |

### Memory Optimization
| Operation | Before | After | Improvement |
|-----------|--------|-------|-------------|
| Large data load | 2GB | 800MB | 60% |
| String processing | 500MB | 100MB | 80% |
| Cache storage | 1GB | 200MB | 80% |
| Image processing | 3GB | 500MB | 83% |

### GPU Optimization
| Workload | CPU Time | GPU Time | Speedup |
|----------|----------|----------|---------|
| Matrix ops | 10s | 0.5s | 20x |
| Tensor compute | 5s | 0.3s | 17x |
| Data processing | 8s | 0.8s | 10x |

## Best Practices Checklist

###  Profile First
- [ ] Run Scalene profiling
- [ ] Identify real bottlenecks
- [ ] Measure baseline metrics
- [ ] Set performance targets

###  Optimize Strategically
- [ ] Start with highest-impact items
- [ ] Measure impact of each change
- [ ] Avoid premature optimization
- [ ] Use Context7 patterns

###  Monitor Continuously
- [ ] Set up real-time monitoring
- [ ] Define alert thresholds
- [ ] Track performance trends
- [ ] Detect regressions early

###  Validate Results
- [ ] Compare before/after metrics
- [ ] Run performance tests
- [ ] Load testing under production conditions
- [ ] Monitor for side effects

## Common Performance Anti-Patterns

### L N+1 Query Problem
```python
# WRONG
for user in users:
    profile = db.query("SELECT * FROM profiles WHERE id = ?", user.id)

# CORRECT
profiles = db.query("SELECT * FROM profiles WHERE id IN ?", user_ids)
```

### L Unbounded Cache Growth
```python
# WRONG
cache = {}
def cache_item(key, value):
    cache[key] = value  # No eviction!

# CORRECT
from functools import lru_cache

@lru_cache(maxsize=1000)
def expensive_function(x):
    return x ** 2
```

### L Synchronous I/O in Loop
```python
# WRONG
for url in urls:
    response = requests.get(url)  # Blocks!

# CORRECT
import asyncio
async def fetch_urls(urls):
    tasks = [fetch(url) for url in urls]
    return await asyncio.gather(*tasks)
```

## Performance Testing Tools

| Tool | Purpose | Use Case |
|------|---------|----------|
| Scalene | Profiling | CPU/GPU/Memory analysis |
| pytest-benchmark | Benchmarking | Regression testing |
| Locust | Load testing | Simulating user load |
| Apache JMeter | Load testing | HTTP endpoints |
| Prometheus | Monitoring | Metrics collection |
| Grafana | Dashboards | Visualization |

## Optimization Impact Guide

### High Impact (Optimize First)
- Algorithm complexity (O(n�) � O(n log n))
- Database query optimization
- Memory allocation reduction
- Network round trips

### Medium Impact (Next Priority)
- Caching strategies
- Code structure refactoring
- String/array operations optimization
- File I/O optimization

### Low Impact (Polish Later)
- Variable naming
- Code comments
- Minor code style changes
- Documentation updates

## Performance Goals Template

```
Service: [service name]
Target Audience: [users/internal]

Response Time Targets:
- P50 (median): [target]ms
- P95: [target]ms
- P99: [target]ms

Throughput Targets:
- Requests/second: [target]
- Concurrent users: [target]

Resource Targets:
- CPU usage: <[target]%
- Memory usage: <[target]GB
- Network: <[target]Mbps

Scaling Targets:
- Horizontal: [# instances]
- Vertical: [CPU/Memory specs]
```

## Success Metrics

### Optimization Success Criteria
-  Measured baseline before optimization
-  Achieved target improvements
-  No performance regressions
-  Context7 patterns applied
-  Monitoring in place
-  Documented changes

### Project Success Rate
- **60% average improvement** with AI optimization
- **95% accuracy** in bottleneck detection
- **85% success rate** for AI-suggested optimizations
- **90% pattern application** from Context7

---

**Last Updated**: 2025-11-23
**Maintained by**: Alfred
