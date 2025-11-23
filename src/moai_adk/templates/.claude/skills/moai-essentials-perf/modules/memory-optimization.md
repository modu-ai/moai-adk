# Memory Optimization Strategies

Complete guide to reducing memory usage and preventing memory leaks.

## Memory Profiling Basics

```python
import tracemalloc

# Start memory tracing
tracemalloc.start()

# Your code here
data = [list(range(1000)) for _ in range(1000)]

# Get memory snapshot
current, peak = tracemalloc.get_traced_memory()
print(f"Current: {current / 1e6:.1f} MB")
print(f"Peak: {peak / 1e6:.1f} MB")

tracemalloc.stop()
```

## Memory Optimization Patterns

### Pattern 1: Data Structure Optimization

```python
# ❌ WRONG - Inefficient storage
def slow_approach():
    # Lists consume more memory than necessary
    large_list = [i for i in range(1000000)]

    # Storing duplicates
    data = {'items': large_list, 'copy': large_list[:]}

    return data

# ✅ CORRECT - Memory-efficient storage
def fast_approach():
    import numpy as np

    # NumPy arrays use less memory
    large_array = np.arange(1000000)

    # Use references instead of copies
    data = {'items': large_array, 'view': large_array[::2]}

    return data
```

### Pattern 2: Generator-Based Processing

```python
# ❌ WRONG - Load everything into memory
def process_all_data():
    with open('large_file.txt') as f:
        data = f.readlines()  # Loads entire file

    for line in data:
        process_line(line)

# ✅ CORRECT - Process line by line
def stream_process_data():
    with open('large_file.txt') as f:
        for line in f:  # Generator - one line at a time
            process_line(line)

# ✅ Generator function
def read_chunks(filepath, chunk_size=1024):
    """Read file in chunks."""
    with open(filepath, 'rb') as f:
        while True:
            chunk = f.read(chunk_size)
            if not chunk:
                break
            yield chunk
```

### Pattern 3: Object Pool Pattern

```python
class ObjectPool:
    """Reuse objects instead of creating new ones."""

    def __init__(self, object_class, size=100):
        self.available = [object_class() for _ in range(size)]
        self.in_use = []

    def acquire(self):
        """Get object from pool."""
        if self.available:
            obj = self.available.pop()
        else:
            obj = object_class()

        self.in_use.append(obj)
        return obj

    def release(self, obj):
        """Return object to pool."""
        obj.reset()
        self.in_use.remove(obj)
        self.available.append(obj)
```

## Memory Leak Detection

### Using Scalene for Memory Leaks

```bash
# Detect memory leaks with Scalene
scalene --memory --trace-allocations script.py
```

### Python Memory Leak Detection

```python
import gc
import objgraph

class MemoryLeakDetector:
    def detect_leaks(self):
        """Detect growing memory leaks."""

        # Force garbage collection
        gc.collect()

        # Get object counts
        initial_counts = objgraph.show_most_common_types(limit=5)

        # Run operation
        self.operation_that_might_leak()

        # Check object growth
        gc.collect()
        final_counts = objgraph.show_most_common_types(limit=5)

        # Compare counts
        growth = self.compare_counts(initial_counts, final_counts)
        return growth
```

## Caching Strategies

### Pattern 1: LRU Cache

```python
from functools import lru_cache

@lru_cache(maxsize=128)
def expensive_operation(x):
    """Cache results to avoid recomputation."""
    # Expensive computation
    return x ** 2 + x + 1

# Usage
result = expensive_operation(42)
result = expensive_operation(42)  # Cache hit!
```

### Pattern 2: Memory-Aware Caching

```python
class MemoryAwareCache:
    """Cache with memory limit."""

    def __init__(self, max_memory_mb=100):
        self.cache = {}
        self.max_memory = max_memory_mb * 1e6
        self.current_memory = 0

    def get(self, key):
        return self.cache.get(key)

    def set(self, key, value):
        """Cache with memory awareness."""
        import sys

        value_size = sys.getsizeof(value)

        # Evict if necessary
        while self.current_memory + value_size > self.max_memory:
            self.evict_oldest()

        self.cache[key] = value
        self.current_memory += value_size

    def evict_oldest(self):
        """Remove oldest entry."""
        if self.cache:
            key = next(iter(self.cache))
            value = self.cache.pop(key)
            import sys
            self.current_memory -= sys.getsizeof(value)
```

## Best Practices

### ✅ DO
- Profile memory usage regularly
- Use generators for large datasets
- Reuse objects when possible
- Implement caching for expensive operations
- Monitor memory growth over time
- Clear caches periodically
- Use weak references for caches

### ❌ DON'T
- Hold references to large objects longer than necessary
- Create copies of large data structures
- Ignore memory warnings
- Cache unlimited amounts of data
- Load entire files into memory
- Create circular references without cleanup

---

**Related Tools**: tracemalloc, memory_profiler, objgraph
