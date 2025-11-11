---
name: moai-essentials-perf
version: 4.0.0
created: 2025-10-22
updated: 2025-11-11
status: active
description: "Performance optimization and profiling with Context7. Analyze bottlenecks, optimize algorithms, improve memory usage, profile code execution. Use when optimizing Python code, analyzing performance issues, or improving system efficiency."
keywords: ['performance', 'optimization', 'profiling', 'benchmarking', 'memory-management', 'cpu-optimization', 'context7', 'mcp-integration']
allowed-tools: "Read, Write, Edit, Glob, Bash, AskUserQuestion, mcp__context7__resolve-library-id, mcp__context7__get-library-docs, WebFetch"
---

# Performance Optimization and Profiling

## Quick Start

1. **Install profiling tools**: `pip install py-spy memory-profiler line-profiler`
2. **Run performance analysis**: Use built-in profiling patterns
3. **Optimize systematically**: Apply Context7-backed optimization strategies

## Core Performance Patterns

### Pattern 1: CPU Profiling with Py-Spy

```python
# Profile running Python process
py-spy top --pid 12345  # Real-time CPU usage
py-spy record -o profile.svg --pid 12345  # Flame graph
py-spy dump --pid 12345  # Stack traces
```

### Pattern 2: Memory Profiling

```python
# Memory line profiler
@profile
def memory_intensive_function():
    data = []
    for i in range(1000000):
        data.append([i] * 100)
    return data

# Run with: python -m memory_profiler script.py
```

### Pattern 3: Algorithm Optimization

```python
# Before: O(n²) nested loops
def find_duplicates_slow(arr):
    duplicates = []
    for i in range(len(arr)):
        for j in range(i+1, len(arr)):
            if arr[i] == arr[j] and arr[i] not in duplicates:
                duplicates.append(arr[i])
    return duplicates

# After: O(n) with set lookup
def find_duplicates_fast(arr):
    seen = set()
    duplicates = set()
    for item in arr:
        if item in seen:
            duplicates.add(item)
        seen.add(item)
    return list(duplicates)
```

## Context7 Integration Examples

### Performance Libraries with Context7

```python
# Get latest optimization techniques from Context7
def optimize_with_context7():
    """
    Context7-backed performance optimization patterns
    - Latest Python optimization techniques
    - Algorithm complexity analysis
    - Memory usage best practices
    """
    pass
```

## Performance Optimization Strategies

### 1. Algorithm Complexity

| Complexity | Use Case | Example |
|------------|----------|---------|
| O(1) | Direct access | Dictionary lookups |
| O(log n) | Search operations | Binary search |
| O(n) | Single pass | List comprehensions |
| O(n log n) | Sorting | Efficient sorting algorithms |
| O(n²) | Nested operations | Matrix operations (use libraries) |

### 2. Memory Optimization

```python
# Memory-efficient data processing
import numpy as np
from itertools import islice

# Process large files in chunks
def process_large_file(filename, chunk_size=1000):
    with open(filename) as f:
        while True:
            chunk = list(islice(f, chunk_size))
            if not chunk:
                break
            yield process_chunk(chunk)

# Use generators instead of lists
def generate_numbers(n):
    for i in range(n):
        yield i  # Memory: O(1) vs O(n) for list
```

### 3. I/O Optimization

```python
# Async I/O for concurrent operations
import asyncio
import aiohttp

async def fetch_multiple_urls(urls):
    async with aiohttp.ClientSession() as session:
        tasks = [fetch_url(session, url) for url in urls]
        return await asyncio.gather(*tasks)

# Buffering for file operations
def write_efficiently(data, filename):
    with open(filename, 'w', buffering=8192) as f:
        for item in data:
            f.write(f"{item}\n")
```

## Profiling Tools Integration

### Built-in Profilers

```python
import cProfile
import pstats
from functools import wraps

def profile_function(func):
    """Decorator to profile function execution"""
    @wraps(func)
    def wrapper(*args, **kwargs):
        pr = cProfile.Profile()
        pr.enable()
        result = func(*args, **kwargs)
        pr.disable()

        stats = pstats.Stats(pr)
        stats.sort_stats('cumulative')
        stats.print_stats(10)  # Top 10 functions
        return result
    return wrapper

# Usage
@profile_function
def slow_function():
    import time
    time.sleep(1)
    return sum(range(1000000))
```

### External Profilers

```bash
# Line-by-line profiling
kernprof -l -v script.py

# Memory profiling
mprof run script.py
mprof plot

# CPU profiling with flame graphs
py-spy record -o profile.svg -- python script.py
```

## Performance Anti-patterns

### ❌ Common Mistakes

```python
# 1. Inefficient string concatenation
result = ""
for item in large_list:
    result += str(item)  # O(n²) due to reallocation

# ✅ Better: Use join
result = "".join(str(item) for item in large_list)

# 2. Repeated calculations in loops
for i in range(len(data)):
    processed = expensive_calculation(data[i])
    # ... other operations

# ✅ Better: Pre-compute or cache
@lru_cache(maxsize=None)
def expensive_calculation_cached(item):
    return expensive_calculation(item)
```

## Context7 Research Integration

### Performance Research Patterns

```python
# Research-backed optimization techniques
def apply_research_optimizations():
    """
    Latest findings from performance research:
    - JIT compilation with Numba for numerical code
    - Vectorization with NumPy/Pandas
    - Parallel processing with multiprocessing
    - Async patterns for I/O-bound tasks
    """
    pass
```

## Real-World Examples

### Example 1: Web API Performance

```python
import asyncio
import aiohttp
from aiohttp import web

async def handle_request(request):
    # Async processing for better concurrency
    data = await fetch_data_async()
    processed = await process_data_async(data)
    return web.json_response(processed)

async def fetch_data_async():
    async with aiohttp.ClientSession() as session:
        async with session.get('https://api.example.com/data') as resp:
            return await resp.json()
```

### Example 2: Data Processing Pipeline

```python
import pandas as pd
import numpy as np
from multiprocessing import Pool

def optimize_data_pipeline(data_path):
    # Vectorized operations
    df = pd.read_csv(data_path)

    # Use NumPy for numerical operations
    df['processed'] = np.vectorize(complex_calculation)(df['column'])

    # Parallel processing for heavy computations
    with Pool() as pool:
        results = pool.map(process_chunk, np.array_split(df, 4))

    return pd.concat(results)
```

## Performance Checklist

- [ ] Profile before optimizing
- [ ] Identify bottlenecks with data
- [ ] Consider algorithm complexity first
- [ ] Use appropriate data structures
- [ ] Apply caching where beneficial
- [ ] Optimize I/O operations
- [ ] Consider parallel processing
- [ ] Validate improvements with benchmarks
- [ ] Monitor memory usage
- [ ] Document trade-offs

## Works Well With

- `Skill("moai-essentials-debug")` - Debugging performance issues
- `Skill("moai-essentials-refactor")` - Refactoring for performance
- `Skill("moai-lang-python")` - Python-specific optimizations
- `Skill("moai-domain-backend")` - Backend performance patterns

---

**Reference**: Performance optimization best practices with Context7 integration
**Version**: 4.0.0 Enterprise
