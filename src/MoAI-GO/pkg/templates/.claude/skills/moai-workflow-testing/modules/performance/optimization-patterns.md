# Optimization Patterns

> Sub-module: Specific optimization strategies and implementation patterns
> Parent: [Performance Optimization](../performance-optimization.md)
> Complexity: Advanced
> Time: 25+ minutes

## Overview

Comprehensive guide to performance optimization patterns, strategies, and best practices for Python applications.

## Optimization Types

### Algorithm Improvement

Pattern: Optimize algorithmic complexity by reducing time complexity from O(n^2) to O(n log n) or better.

```python
# Before: O(n^2) nested loops
def find_duplicates_slow(items):
    duplicates = []
    for i, item1 in enumerate(items):
        for j, item2 in enumerate(items[i+1:], start=i+1):
            if item1 == item2 and item1 not in duplicates:
                duplicates.append(item1)
    return duplicates

# After: O(n) using set
def find_duplicates_fast(items):
    seen = set()
    duplicates = set()
    for item in items:
        if item in seen:
            duplicates.add(item)
        else:
            seen.add(item)
    return list(duplicates)
```

Impact: 10-1000x speedup depending on dataset size

### Caching Strategies

#### Memoization

```python
from functools import lru_cache

@lru_cache(maxsize=128)
def fibonacci_memoized(n):
    if n <= 1:
        return n
    return fibonacci_memoized(n-1) + fibonacci_memoized(n-2)
```

Impact: 50-90% speedup for repeated calls

#### Custom Caching

```python
class CacheOptimizer:
    def __init__(self, max_size=1000):
        self.cache = {}
        self.max_size = max_size
        self.access_count = {}
    
    def get(self, key):
        if key in self.cache:
            self.access_count[key] = self.access_count.get(key, 0) + 1
            return self.cache[key]
        return None
    
    def set(self, key, value):
        if len(self.cache) >= self.max_size:
            lru_key = min(self.access_count, key=self.access_count.get)
            del self.cache[lru_key]
            del self.access_count[lru_key]
        
        self.cache[key] = value
        self.access_count[key] = 0
```

### Concurrency Patterns

#### Multiprocessing for CPU-Bound Tasks

```python
from multiprocessing import Pool

def process_parallel(items):
    with Pool(processes=multiprocessing.cpu_count()) as pool:
        return pool.map(process_item, items)
```

Impact: 2-8x speedup on multi-core systems

#### Asyncio for I/O-Bound Tasks

```python
import asyncio

async def fetch_all_urls(urls):
    async with aiohttp.ClientSession() as session:
        tasks = [fetch_url(session, url) for url in urls]
        return await asyncio.gather(*tasks)
```

Impact: 10-100x speedup for I/O-bound operations

### Memory Optimization

#### Generator Patterns

```python
# Stream data instead of loading into memory
def process_large_file_generator(filename):
    with open(filename) as f:
        for line in f:
            yield process_line(line)
```

Impact: 60-90% memory reduction

#### Memory Pooling

```python
class ObjectPool:
    def __init__(self, factory, max_size=100):
        self.factory = factory
        self.pool = []
        self.max_size = max_size
```

### I/O Optimization

#### Buffered I/O

```python
# Batch writes for efficiency
def write_lines_buffered(lines, filename):
    with open(filename, 'w', buffering=8192) as f:
        f.writelines(lines)
```

Impact: 5-20x speedup for I/O operations

### Data Structure Optimization

#### Appropriate Data Structure Selection

```python
from collections import deque

queue_deque = deque()  # O(1) for popleft
search_set = set()     # O(1) for membership test
```

#### NumPy for Numerical Data

```python
import numpy as np

def sum_arrays(arrays):
    return np.sum(arrays, axis=0)
```

Impact: 10-100x speedup for numerical operations

## Optimization Planning

### Optimization Plan Structure

```python
@dataclass
class OptimizationPlan:
    bottlenecks: List[PerformanceBottleneck]
    execution_order: List[int]
    estimated_total_improvement: str
    implementation_complexity: str
    risk_level: str
    prerequisites: List[str]
    validation_strategy: str
```

### Prioritization Strategy

```python
def _prioritize_bottlenecks(
    self, bottlenecks: List[PerformanceBottleneck]
) -> List[PerformanceBottleneck]:
    severity_order = {'critical': 4, 'high': 3, 'medium': 2, 'low': 1}
    
    return sorted(
        bottlenecks,
        key=lambda x: (
            severity_order.get(x.severity, 0),
            x.impact_score,
            self._get_optimization_priority(x.optimization_type)
        ),
        reverse=True
    )
```

### Execution Order

```python
def _create_optimization_execution_order(
    self, bottlenecks: List[PerformanceBottleneck]
) -> List[int]:
    type_groups = defaultdict(list)
    for i, bottleneck in enumerate(bottlenecks):
        type_groups[bottleneck.optimization_type].append(i)
    
    execution_order = []
    type_order = [
        OptimizationType.ALGORITHM_IMPROVEMENT,
        OptimizationType.DATA_STRUCTURE_CHANGE,
        OptimizationType.CACHING,
        OptimizationType.MEMORY_OPTIMIZATION,
        OptimizationType.CONCURRENCY,
        OptimizationType.I_O_OPTIMIZATION,
        OptimizationType.DATABASE_OPTIMIZATION
    ]
    
    for opt_type in type_order:
        if opt_type in type_groups:
            execution_order.extend(type_groups[opt_type])
    
    return execution_order
```

## Implementation Strategies

### Risk Assessment

```python
def _assess_optimization_risk(
    self, bottlenecks: List[PerformanceBottleneck]
) -> str:
    high_risk_types = {
        OptimizationType.ALGORITHM_IMPROVEMENT,
        OptimizationType.DATA_STRUCTURE_CHANGE,
        OptimizationType.CONCURRENCY
    }
    
    high_risk_count = sum(
        1 for b in bottlenecks
        if b.optimization_type in high_risk_types and b.impact_score > 0.3
    )
    
    if high_risk_count > 3:
        return "high"
    elif high_risk_count > 1:
        return "medium"
    else:
        return "low"
```

### Prerequisites Identification

```python
def _identify_optimization_prerequisites(
    self, bottlenecks: List[PerformanceBottleneck]
) -> List[str]:
    prerequisites = [
        "Create comprehensive performance benchmarks",
        "Ensure version control with current implementation",
        "Set up performance testing environment"
    ]
    
    optimization_types = set(b.optimization_type for b in bottlenecks)
    
    if OptimizationType.CONCURRENCY in optimization_types:
        prerequisites.extend([
            "Review thread safety and shared resource access",
            "Implement proper synchronization mechanisms"
        ])
    
    if OptimizationType.DATABASE_OPTIMIZATION in optimization_types:
        prerequisites.extend([
            "Create database backup before optimization",
            "Set up database performance monitoring"
        ])
    
    return prerequisites
```

### Validation Strategy

```python
def _create_validation_strategy(
    self, bottlenecks: List[PerformanceBottleneck]
) -> str:
    return """
    Validation Strategy:
    1. Baseline Performance Measurement
    2. Incremental Testing
    3. Automated Performance Testing
    4. Functional Validation
    5. Production Monitoring
    """
```

## Intelligent Optimization

### AI-Powered Suggestions

```python
class IntelligentOptimizer(PerformanceProfiler):
    async def get_ai_optimization_suggestions(
        self, bottlenecks: List[PerformanceBottleneck],
        codebase_context: Dict[str, Any]
    ) -> Dict[str, Any]:
        if not self.context7:
            return self._get_rule_based_suggestions(bottlenecks)
        
        optimization_patterns = await self.context7.get_library_docs(
            context7_library_id="/performance/python-profiling",
            topic="advanced performance optimization patterns 2025",
            tokens=5000
        )
        
        algorithm_patterns = await self.context7.get_library_docs(
            context7_library_id="/algorithms/python",
            topic="algorithm optimization big-O complexity reduction",
            tokens=3000
        )
        
        return await self._generate_ai_suggestions(
            bottlenecks, optimization_patterns, algorithm_patterns, codebase_context
        )
```

### Algorithm Improvement Suggestions

```python
def _suggest_algorithm_improvement(
    self, bottleneck: PerformanceBottleneck, algo_patterns: Dict
) -> Dict[str, Any]:
    function_name = bottleneck.function_name.lower()
    suggestions = []
    
    if any(keyword in function_name for keyword in ["search", "find"]):
        suggestions.extend([
            "Consider using binary search for sorted data",
            "Implement hash-based lookup for O(1) average case",
            "Use trie structures for prefix searches"
        ])
    
    elif any(keyword in function_name for keyword in ["sort", "order"]):
        suggestions.extend([
            "Consider using Timsort (Python's built-in sort)",
            "Use radix sort for uniform integer data",
            "Implement bucket sort for uniformly distributed data"
        ])
    
    return {
        'bottleneck': bottleneck.function_name,
        'suggestions': suggestions,
        'estimated_improvement': "30-90% depending on algorithm",
        'implementation_complexity': "medium to high"
    }
```

### Data Structure Optimization

```python
def _suggest_data_structure_improvement(
    self, bottleneck: PerformanceBottleneck, opt_patterns: Dict
) -> Dict[str, Any]:
    return {
        'bottleneck': bottleneck.function_name,
        'suggestions': [
            "Use generators instead of lists for large datasets",
            "Implement lazy loading for expensive data structures",
            "Use memoryviews or numpy arrays for numerical data",
            "Consider using collections.deque for queue operations",
            "Use set/dict for O(1) lookups instead of list searches"
        ],
        'estimated_improvement': "30-80% memory reduction",
        'implementation_complexity': "low to medium"
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
