# Caching and Performance Optimization

> Module: Advanced caching strategies and performance optimization
> Complexity: Advanced
> Time: 15+ minutes
> Dependencies: functools, hashlib, pickle, time, typing

## Intelligent Data Caching

```python
from functools import wraps
import hashlib
import pickle
from typing import Any, Dict, Optional
import time

class DataCache:
 """Intelligent caching system for data operations."""

 def __init__(self, max_size: int = 1000, ttl: int = 3600):
 self.max_size = max_size
 self.ttl = ttl
 self.cache: Dict[str, Dict] = {}

 def _generate_key(self, func_name: str, args: tuple, kwargs: dict) -> str:
 """Generate cache key from function arguments."""
 key_data = {
 'func': func_name,
 'args': args,
 'kwargs': kwargs
 }
 key_str = pickle.dumps(key_data)
 return hashlib.md5(key_str).hexdigest()

 def cache_result(self, ttl: Optional[int] = None):
 """Decorator for caching function results."""
 def decorator(func):
 @wraps(func)
 def wrapper(*args, kwargs):
 cache_key = self._generate_key(func.__name__, args, kwargs)

 # Check cache
 if cache_key in self.cache:
 cache_entry = self.cache[cache_key]
 if time.time() - cache_entry['timestamp'] < (ttl or self.ttl):
 return cache_entry['result']

 # Execute function and cache result
 result = func(*args, kwargs)

 # Manage cache size
 if len(self.cache) >= self.max_size:
 # Remove oldest entry
 oldest_key = min(self.cache.keys(),
 key=lambda k: self.cache[k]['timestamp'])
 del self.cache[oldest_key]

 # Store new entry
 self.cache[cache_key] = {
 'result': result,
 'timestamp': time.time()
 }

 return result

 return wrapper
 return decorator

 def invalidate_pattern(self, pattern: str):
 """Invalidate cache entries matching pattern."""
 keys_to_remove = [
 key for key in self.cache.keys()
 if pattern in key
 ]
 for key in keys_to_remove:
 del self.cache[key]

 def clear_expired(self):
 """Clear expired cache entries."""
 current_time = time.time()
 expired_keys = [
 key for key, entry in self.cache.items()
 if current_time - entry['timestamp'] >= self.ttl
 ]
 for key in expired_keys:
 del self.cache[key]
```

## Advanced Caching Strategies

### Multi-Level Caching

```python
class MultiLevelCache:
 """Multi-level caching with memory and persistent storage."""

 def __init__(self, memory_size: int = 1000, persistent_path: str = None):
 self.memory_cache = DataCache(max_size=memory_size)
 self.persistent_path = persistent_path
 self.persistent_cache = {}

 if persistent_path:
 self._load_persistent_cache()

 def _load_persistent_cache(self):
 """Load persistent cache from disk."""
 try:
 with open(self.persistent_path, 'rb') as f:
 self.persistent_cache = pickle.load(f)
 except (FileNotFoundError, pickle.PickleError):
 self.persistent_cache = {}

 def _save_persistent_cache(self):
 """Save persistent cache to disk."""
 try:
 with open(self.persistent_path, 'wb') as f:
 pickle.dump(self.persistent_cache, f)
 except (IOError, pickle.PickleError):
 pass # Silently fail for cache operations

 def get(self, key: str, use_memory: bool = True, use_persistent: bool = True) -> Any:
 """Get value from cache levels."""
 # Try memory cache first
 if use_memory and key in self.memory_cache.cache:
 entry = self.memory_cache.cache[key]
 if time.time() - entry['timestamp'] < self.memory_cache.ttl:
 return entry['result']

 # Try persistent cache
 if use_persistent and key in self.persistent_cache:
 entry = self.persistent_cache[key]
 if time.time() - entry['timestamp'] < entry.get('ttl', self.memory_cache.ttl):
 # Promote to memory cache
 self.memory_cache.cache[key] = entry
 return entry['result']

 return None

 def set(self, key: str, value: Any, persist: bool = False, ttl: int = None):
 """Set value in cache levels."""
 # Always set in memory cache
 self.memory_cache.cache[key] = {
 'result': value,
 'timestamp': time.time()
 }

 # Optionally persist
 if persist and self.persistent_path:
 self.persistent_cache[key] = {
 'result': value,
 'timestamp': time.time(),
 'ttl': ttl or self.memory_cache.ttl
 }
 self._save_persistent_cache()

class SmartCache:
 """Smart caching with memory pressure and access pattern analysis."""

 def __init__(self, max_memory_mb: int = 100, max_items: int = 10000):
 self.max_memory_mb = max_memory_mb
 self.max_items = max_items
 self.cache = {}
 self.access_count = {}
 self.last_access = {}
 self.estimated_sizes = {}

 def _estimate_size(self, obj: Any) -> int:
 """Estimate memory size of an object."""
 try:
 return len(pickle.dumps(obj))
 except pickle.PickleError:
 return 1024 # Default estimate

 def _get_memory_usage(self) -> int:
 """Get current memory usage in bytes."""
 return sum(self.estimated_sizes.values())

 def _evict_lru(self, count: int = 1):
 """Evict least recently used items."""
 if not self.cache:
 return

 # Sort by last access time
 sorted_items = sorted(
 self.cache.keys(),
 key=lambda k: self.last_access.get(k, 0)
 )

 for key in sorted_items[:count]:
 self.remove(key)

 def get(self, key: str) -> Any:
 """Get item and update access statistics."""
 if key in self.cache:
 entry = self.cache[key]
 current_time = time.time()

 # Check expiration
 if current_time - entry['timestamp'] > entry.get('ttl', float('inf')):
 self.remove(key)
 return None

 # Update access statistics
 self.access_count[key] = self.access_count.get(key, 0) + 1
 self.last_access[key] = current_time

 return entry['value']

 return None

 def set(self, key: str, value: Any, ttl: int = None):
 """Set item with memory management."""
 current_time = time.time()
 item_size = self._estimate_size(value)

 # Check memory constraints
 if (self._get_memory_usage() + item_size > self.max_memory_mb * 1024 * 1024 or
 len(self.cache) >= self.max_items):

 # Calculate eviction score (access count * recency)
 def eviction_score(k):
 count = self.access_count.get(k, 0)
 last_access = self.last_access.get(k, current_time)
 recency = current_time - last_access
 return count / (recency + 1) # Avoid division by zero

 # Evict items with lowest scores
 if self.cache:
 sorted_by_score = sorted(self.cache.keys(), key=eviction_score)
 evict_count = max(1, len(self.cache) // 10) # Evict 10% or at least 1
 for key in sorted_by_score[:evict_count]:
 self.remove(key)

 # Store item
 self.cache[key] = {
 'value': value,
 'timestamp': current_time,
 'ttl': ttl or float('inf')
 }
 self.access_count[key] = 1
 self.last_access[key] = current_time
 self.estimated_sizes[key] = item_size

 def remove(self, key: str):
 """Remove item from cache."""
 self.cache.pop(key, None)
 self.access_count.pop(key, None)
 self.last_access.pop(key, None)
 self.estimated_sizes.pop(key, None)

 def get_stats(self) -> Dict[str, Any]:
 """Get cache statistics."""
 return {
 'items': len(self.cache),
 'memory_usage_mb': self._get_memory_usage() / (1024 * 1024),
 'memory_limit_mb': self.max_memory_mb,
 'hit_rate': self._calculate_hit_rate(),
 'top_accessed': sorted(
 self.access_count.items(),
 key=lambda x: x[1],
 reverse=True
 )[:10]
 }

 def _calculate_hit_rate(self) -> float:
 """Calculate cache hit rate (simplified)."""
 total_accesses = sum(self.access_count.values())
 if total_accesses == 0:
 return 0.0
 # This is a simplified calculation - real implementation would track hits/misses
 return min(0.8, total_accesses / (total_accesses + 10)) # Dummy calculation
```

### Cache Invalidation Strategies

```python
class CacheInvalidator:
 """Advanced cache invalidation strategies."""

 def __init__(self, cache: DataCache):
 self.cache = cache
 self.tags = {} # key -> set of tags
 self.tag_to_keys = {} # tag -> set of keys

 def set_with_tags(self, key: str, value: Any, tags: List[str], ttl: int = None):
 """Set cache value with tags for invalidation."""
 # Store in cache
 if hasattr(self.cache, 'set'):
 self.cache.set(key, value, ttl)
 else:
 self.cache.cache[key] = {
 'result': value,
 'timestamp': time.time()
 }

 # Manage tags
 self.tags[key] = set(tags)
 for tag in tags:
 if tag not in self.tag_to_keys:
 self.tag_to_keys[tag] = set()
 self.tag_to_keys[tag].add(key)

 def invalidate_by_tag(self, tag: str):
 """Invalidate all cache entries with specific tag."""
 if tag in self.tag_to_keys:
 keys_to_invalidate = self.tag_to_keys[tag]
 for key in keys_to_invalidate:
 self.cache.cache.pop(key, None)
 self.tags.pop(key, None)

 # Clear tag mapping
 del self.tag_to_keys[tag]

 def invalidate_by_pattern(self, pattern: str, is_regex: bool = False):
 """Invalidate cache entries matching pattern."""
 keys_to_remove = []

 if is_regex:
 import re
 regex = re.compile(pattern)
 keys_to_remove = [key for key in self.cache.cache.keys() if regex.match(key)]
 else:
 keys_to_remove = [key for key in self.cache.cache.keys() if pattern in key]

 for key in keys_to_remove:
 self.cache.cache.pop(key, None)
 # Clean up tags
 if key in self.tags:
 for tag in self.tags[key]:
 self.tag_to_keys[tag].discard(key)
 self.tags.pop(key, None)

 def invalidate_dependencies(self, key: str):
 """Invalidate entries that depend on this key."""
 # Implement dependency tracking logic
 pass

class CacheWarmer:
 """Cache warming for predictable access patterns."""

 def __init__(self, cache: DataCache):
 self.cache = cache
 self.warming_strategies = {}

 def register_warm_strategy(self, name: str, warm_func: callable):
 """Register a cache warming strategy."""
 self.warming_strategies[name] = warm_func

 def warm_cache(self, strategy_name: str, *args, kwargs):
 """Warm cache using specific strategy."""
 if strategy_name in self.warming_strategies:
 warm_func = self.warming_strategies[strategy_name]
 warm_func(self.cache, *args, kwargs)

 def warm_all_strategies(self):
 """Warm cache using all registered strategies."""
 for name, warm_func in self.warming_strategies.items():
 try:
 warm_func(self.cache)
 except Exception as e:
 print(f"Cache warming failed for {name}: {e}")

# Example usage
@SmartCache(max_memory_mb=50).cache.cache_result(ttl=1800)
def expensive_computation(data: Dict) -> Dict:
 """Example function with caching."""
 # Simulate expensive operation
 time.sleep(0.1)
 return {"result": sum(data.values()), "timestamp": time.time()}

def cache_warming_example(cache: DataCache):
 """Example cache warming function."""
 # Pre-load common data
 common_data = [
 {"a": 1, "b": 2},
 {"x": 10, "y": 20},
 {"p": 100, "q": 200}
 ]

 for data in common_data:
 cache.cache.cache[f"computation_{hash(str(data))}"] = {
 'result': {"result": sum(data.values()), "timestamp": time.time()},
 'timestamp': time.time()
 }
```

## Performance Monitoring

```python
class PerformanceMonitor:
 """Monitor cache and data processing performance."""

 def __init__(self):
 self.metrics = {
 'cache_hits': 0,
 'cache_misses': 0,
 'operation_times': {},
 'memory_usage': []
 }

 def record_cache_hit(self):
 """Record a cache hit."""
 self.metrics['cache_hits'] += 1

 def record_cache_miss(self):
 """Record a cache miss."""
 self.metrics['cache_misses'] += 1

 def time_operation(self, operation_name: str):
 """Decorator to time operations."""
 def decorator(func):
 @wraps(func)
 def wrapper(*args, kwargs):
 start_time = time.time()
 result = func(*args, kwargs)
 duration = time.time() - start_time

 if operation_name not in self.metrics['operation_times']:
 self.metrics['operation_times'][operation_name] = []
 self.metrics['operation_times'][operation_name].append(duration)

 return result
 return wrapper
 return decorator

 def get_cache_hit_rate(self) -> float:
 """Calculate cache hit rate."""
 total = self.metrics['cache_hits'] + self.metrics['cache_misses']
 return self.metrics['cache_hits'] / total if total > 0 else 0.0

 def get_performance_report(self) -> Dict[str, Any]:
 """Get comprehensive performance report."""
 report = {
 'cache_hit_rate': self.get_cache_hit_rate(),
 'total_requests': self.metrics['cache_hits'] + self.metrics['cache_misses']
 }

 # Operation statistics
 for op_name, times in self.metrics['operation_times'].items():
 report[f'{op_name}_avg_time'] = sum(times) / len(times)
 report[f'{op_name}_max_time'] = max(times)
 report[f'{op_name}_min_time'] = min(times)

 return report
```

## Best Practices

1. Use multi-level caching: Memory + persistent storage
2. Implement smart eviction: LRU with memory pressure awareness
3. Tag-based invalidation: Group related cache entries
4. Monitor performance: Track hit rates and operation times
5. Warm caches strategically: Pre-load predictable data
6. Set appropriate TTLs: Balance freshness and performance

---

Module: `modules/caching-performance.md`
Related: [JSON Optimization](./json-optimization.md) | [TOON Encoding](./toon-encoding.md)
