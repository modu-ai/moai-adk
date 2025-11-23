# Optimization Techniques

Algorithmic and memory optimization patterns with Context7 best practices.

## Algorithm Optimization

### Big O Complexity Improvements

```python
class AlgorithmOptimizer:
    """Optimize algorithm complexity."""

    def optimize_search(self, data: list, target: any) -> OptimizedSearch:
        """Convert O(n) linear search to O(log n) binary search."""

        # Before: Linear search O(n)
        def linear_search(arr, x):
            for i, item in enumerate(arr):
                if item == x:
                    return i
            return -1

        # After: Binary search O(log n)
        def binary_search(arr, x):
            left, right = 0, len(arr) - 1

            while left <= right:
                mid = (left + right) // 2

                if arr[mid] == x:
                    return mid
                elif arr[mid] < x:
                    left = mid + 1
                else:
                    right = mid - 1

            return -1

        # Ensure data is sorted for binary search
        sorted_data = sorted(data)

        return OptimizedSearch(
            algorithm="binary_search",
            complexity="O(log n)",
            speedup=f"{len(data) / math.log2(len(data)):.2f}x",
            implementation=binary_search
        )

    def optimize_duplicate_detection(self, arr: list) -> OptimizedDetection:
        """Convert O(n²) to O(n) using set."""

        # Before: Nested loop O(n²)
        def find_duplicates_slow(arr):
            duplicates = []
            for i in range(len(arr)):
                for j in range(i + 1, len(arr)):
                    if arr[i] == arr[j] and arr[i] not in duplicates:
                        duplicates.append(arr[i])
            return duplicates

        # After: Set-based O(n)
        def find_duplicates_fast(arr):
            seen = set()
            duplicates = set()

            for item in arr:
                if item in seen:
                    duplicates.add(item)
                seen.add(item)

            return list(duplicates)

        return OptimizedDetection(
            algorithm="set_based_duplicate_detection",
            complexity="O(n)",
            speedup=f"{len(arr)}x for large datasets",
            implementation=find_duplicates_fast
        )
```

## Caching Strategies

### LRU Cache Implementation

```python
from collections import OrderedDict
import time

class LRUCache:
    """Least Recently Used cache implementation."""

    def __init__(self, capacity: int = 1000, ttl_seconds: int = 3600):
        self.capacity = capacity
        self.ttl_seconds = ttl_seconds
        self.cache = OrderedDict()
        self.timestamps = {}
        self.stats = {"hits": 0, "misses": 0, "evictions": 0}

    def get(self, key: str) -> Optional[any]:
        """Get value from cache."""

        if key not in self.cache:
            self.stats["misses"] += 1
            return None

        # Check TTL
        if time.time() - self.timestamps[key] > self.ttl_seconds:
            del self.cache[key]
            del self.timestamps[key]
            self.stats["misses"] += 1
            return None

        # Move to end (most recently used)
        self.cache.move_to_end(key)
        self.stats["hits"] += 1
        return self.cache[key]

    def put(self, key: str, value: any) -> None:
        """Put value in cache."""

        if key in self.cache:
            # Update existing
            self.cache.move_to_end(key)
            self.cache[key] = value
            self.timestamps[key] = time.time()
        else:
            # Add new
            if len(self.cache) >= self.capacity:
                # Evict oldest
                oldest_key = next(iter(self.cache))
                del self.cache[oldest_key]
                del self.timestamps[oldest_key]
                self.stats["evictions"] += 1

            self.cache[key] = value
            self.timestamps[key] = time.time()

    def get_hit_rate(self) -> float:
        """Calculate cache hit rate."""
        total = self.stats["hits"] + self.stats["misses"]
        return self.stats["hits"] / total if total > 0 else 0
```

### Redis Caching Pattern

```python
import redis
import json
import hashlib

class RedisCache:
    """Redis-based distributed cache."""

    def __init__(self, host: str = "localhost", port: int = 6379):
        self.redis_client = redis.Redis(host=host, port=port, decode_responses=True)
        self.default_ttl = 3600  # 1 hour

    def cache_function_result(self, ttl: int = None):
        """Decorator for caching function results."""

        def decorator(func):
            @functools.wraps(func)
            async def wrapper(*args, **kwargs):
                # Generate cache key
                cache_key = self.generate_cache_key(func.__name__, args, kwargs)

                # Try to get from cache
                cached_value = self.redis_client.get(cache_key)
                if cached_value:
                    return json.loads(cached_value)

                # Execute function
                result = await func(*args, **kwargs)

                # Store in cache
                self.redis_client.setex(
                    cache_key,
                    ttl or self.default_ttl,
                    json.dumps(result)
                )

                return result

            return wrapper
        return decorator

    def generate_cache_key(self, func_name: str, args: tuple, kwargs: dict) -> str:
        """Generate consistent cache key."""

        key_data = {
            "function": func_name,
            "args": str(args),
            "kwargs": str(sorted(kwargs.items()))
        }

        key_string = json.dumps(key_data, sort_keys=True)
        key_hash = hashlib.md5(key_string.encode()).hexdigest()

        return f"cache:{func_name}:{key_hash}"
```

## Memory Optimization

### Memory-Efficient Data Structures

```python
class MemoryOptimizer:
    """Optimize memory usage with efficient data structures."""

    def use_generators(self, large_dataset: list) -> Generator:
        """Use generators instead of lists for large datasets."""

        # Before: Loads all data in memory O(n) space
        def process_all_data(data):
            results = []
            for item in data:
                processed = expensive_operation(item)
                results.append(processed)
            return results

        # After: Generator O(1) space
        def process_data_generator(data):
            for item in data:
                processed = expensive_operation(item)
                yield processed

        return process_data_generator(large_dataset)

    def use_slots(self, class_definition: type) -> type:
        """Use __slots__ to reduce memory overhead."""

        # Before: Regular class (uses __dict__)
        class RegularPoint:
            def __init__(self, x, y):
                self.x = x
                self.y = y

        # After: __slots__ class (56 bytes vs 136 bytes per instance)
        class OptimizedPoint:
            __slots__ = ('x', 'y')

            def __init__(self, x, y):
                self.x = x
                self.y = y

        return OptimizedPoint

    def use_array_instead_of_list(self, numeric_data: list) -> array.array:
        """Use array.array for homogeneous numeric data."""

        import array

        # Before: List (each int is Python object)
        numbers_list = [1, 2, 3, 4, 5] * 1000000  # ~40MB

        # After: Array (compact C array)
        numbers_array = array.array('i', numbers_list)  # ~4MB

        return numbers_array
```

### Memory Pool Pattern

```python
class MemoryPool:
    """Object pooling for memory efficiency."""

    def __init__(self, object_class: type, pool_size: int = 100):
        self.object_class = object_class
        self.available_objects = [object_class() for _ in range(pool_size)]
        self.in_use = set()
        self.stats = {
            "created": pool_size,
            "reused": 0,
            "peak_usage": 0
        }

    def acquire(self) -> any:
        """Get object from pool."""

        if self.available_objects:
            obj = self.available_objects.pop()
            self.stats["reused"] += 1
        else:
            obj = self.object_class()
            self.stats["created"] += 1

        self.in_use.add(id(obj))
        self.stats["peak_usage"] = max(self.stats["peak_usage"], len(self.in_use))

        return obj

    def release(self, obj: any) -> None:
        """Return object to pool."""

        obj.reset()  # Reset object state
        self.in_use.discard(id(obj))
        self.available_objects.append(obj)

    def get_efficiency(self) -> float:
        """Calculate pool efficiency."""
        return self.stats["reused"] / (self.stats["created"] + self.stats["reused"])
```

## Network Optimization

### Connection Pooling
