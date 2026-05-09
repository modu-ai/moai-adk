# JSON/YAML Optimization Implementation

> Module: High-performance JSON and YAML processing
> Complexity: Advanced
> Time: 20+ minutes
> Dependencies: orjson, PyYAML, ijson, typing, dataclasses, functools

## High-Performance JSON Handling

```python
import orjson # Ultra-fast JSON library
import yaml
from typing import Any, Dict, List
from dataclasses import dataclass
from functools import lru_cache

class JSONOptimizer:
 """Optimized JSON processing for high-performance applications."""

 def __init__(self):
 self._compression_cache = {}

 def serialize_fast(self, obj: Any) -> bytes:
 """Ultra-fast JSON serialization using orjson."""
 return orjson.dumps(
 obj,
 option=orjson.OPT_SERIALIZE_NUMPY |
 orjson.OPT_SERIALIZE_DATACLASS |
 orjson.OPT_SERIALIZE_UUID |
 orjson.OPT_NON_STR_KEYS
 )

 def deserialize_fast(self, data: bytes) -> Any:
 """Ultra-fast JSON deserialization."""
 return orjson.loads(data)

 @lru_cache(maxsize=1024)
 def compress_schema(self, schema: Dict) -> Dict:
 """Cache and compress JSON schemas for repeated use."""
 return self._optimize_schema(schema)

 def _optimize_schema(self, schema: Dict) -> Dict:
 """Remove redundant schema properties and optimize structure."""
 optimized = {}

 for key, value in schema.items():
 if key == '$schema': # Remove schema URL
 continue
 elif key == 'description' and len(value) > 100: # Truncate long descriptions
 optimized[key] = value[:100] + '...'
 elif isinstance(value, dict):
 optimized[key] = self._optimize_schema(value)
 elif isinstance(value, list):
 optimized[key] = [self._optimize_schema(item) if isinstance(item, dict) else item for item in value]
 else:
 optimized[key] = value

 return optimized

class YAMLOptimizer:
 """Optimized YAML processing for configuration management."""

 def __init__(self):
 self.yaml_loader = yaml.CSafeLoader # Use C loader for performance
 self.yaml_dumper = yaml.CSafeDumper

 def load_fast(self, stream) -> Any:
 """Fast YAML loading with optimized loader."""
 return yaml.load(stream, Loader=self.yaml_loader)

 def dump_fast(self, data: Any, stream=None) -> str:
 """Fast YAML dumping with optimized dumper."""
 return yaml.dump(
 data,
 stream=stream,
 Dumper=self.yaml_dumper,
 default_flow_style=False,
 sort_keys=False,
 allow_unicode=True
 )

 def merge_configs(self, *configs: Dict) -> Dict:
 """Intelligently merge multiple YAML configurations."""
 result = {}

 for config in configs:
 result = self._deep_merge(result, config)

 return result

 def _deep_merge(self, base: Dict, overlay: Dict) -> Dict:
 """Deep merge dictionaries with conflict resolution."""
 result = base.copy()

 for key, value in overlay.items():
 if key in result:
 if isinstance(result[key], dict) and isinstance(value, dict):
 result[key] = self._deep_merge(result[key], value)
 elif isinstance(result[key], list) and isinstance(value, list):
 result[key] = result[key] + value
 else:
 result[key] = value
 else:
 result[key] = value

 return result

# Streaming data processor for large datasets
class StreamProcessor:
 """Memory-efficient streaming processor for large data files."""

 def __init__(self, chunk_size: int = 8192):
 self.chunk_size = chunk_size

 def process_json_stream(self, file_path: str, processor_func):
 """Process large JSON files in streaming mode."""
 import ijson

 with open(file_path, 'rb') as file:
 parser = ijson.parse(file)

 current_object = {}
 for prefix, event, value in parser:
 if prefix.endswith('.item'):
 if event == 'start_map':
 current_object = {}
 elif event == 'map_key':
 self._current_key = value
 elif event in ['string', 'number', 'boolean', 'null']:
 current_object[self._current_key] = value
 elif event == 'end_map':
 processor_func(current_object)

 def process_csv_stream(self, file_path: str, processor_func):
 """Process large CSV files in streaming mode."""
 import csv

 with open(file_path, 'r', encoding='utf-8') as file:
 reader = csv.DictReader(file)
 for row in reader:
 processor_func(row)

 def aggregate_json_stream(self, file_path: str, aggregation_key: str) -> Dict:
 """Aggregate streaming JSON data by key."""
 aggregates = {}

 def aggregate_processor(item):
 key = item.get(aggregation_key, 'unknown')
 if key not in aggregates:
 aggregates[key] = {
 'count': 0,
 'items': []
 }
 aggregates[key]['count'] += 1
 aggregates[key]['items'].append(item)

 self.process_json_stream(file_path, aggregate_processor)
 return aggregates
```

## Performance Optimization Techniques

### Memory Management

```python
class MemoryEfficientProcessor:
 """Memory-optimized data processing for large datasets."""

 def __init__(self, max_memory_mb: int = 512):
 self.max_memory_mb = max_memory_mb
 self.current_memory = 0

 def batch_process(self, data_generator, batch_size: int = 1000):
 """Process data in batches to control memory usage."""
 batch = []

 for item in data_generator:
 batch.append(item)

 # Check memory usage and process batch
 if len(batch) >= batch_size:
 yield self._process_batch(batch)
 batch.clear()

 # Process remaining items
 if batch:
 yield self._process_batch(batch)

 def _process_batch(self, batch: List[Any]) -> Any:
 """Process a single batch of data."""
 # Implement batch processing logic
 pass

 def _estimate_memory_usage(self, data: Any) -> int:
 """Estimate memory usage of data structure."""
 import sys
 return sys.getsizeof(data)
```

### Caching Strategies

```python
from functools import wraps
import hashlib
import pickle
import time

class SmartCache:
 """Intelligent caching with memory management and expiration."""

 def __init__(self, max_size: int = 1000, ttl: int = 3600):
 self.cache = {}
 self.access_times = {}
 self.max_size = max_size
 self.ttl = ttl

 def get_cache_key(self, func_name: str, args: tuple, kwargs: dict) -> str:
 """Generate consistent cache key."""
 key_data = f"{func_name}:{pickle.dumps(args)}:{pickle.dumps(kwargs)}"
 return hashlib.md5(key_data.encode()).hexdigest()

 def get(self, key: str) -> Any:
 """Get value from cache with expiration check."""
 if key in self.cache:
 entry = self.cache[key]
 if time.time() - entry['timestamp'] < self.ttl:
 self.access_times[key] = time.time()
 return entry['value']
 else:
 del self.cache[key]
 del self.access_times[key]
 return None

 def set(self, key: str, value: Any):
 """Set value in cache with LRU eviction."""
 # Evict old entries if at capacity
 while len(self.cache) >= self.max_size:
 oldest_key = min(self.access_times.keys(),
 key=lambda k: self.access_times[k])
 del self.cache[oldest_key]
 del self.access_times[oldest_key]

 self.cache[key] = {
 'value': value,
 'timestamp': time.time()
 }
 self.access_times[key] = time.time()

def cached(ttl: int = 3600):
 """Decorator for smart function caching."""
 cache = SmartCache(ttl=ttl)

 def decorator(func):
 @wraps(func)
 def wrapper(*args, kwargs):
 cache_key = cache.get_cache_key(func.__name__, args, kwargs)

 # Try to get from cache
 cached_result = cache.get(cache_key)
 if cached_result is not None:
 return cached_result

 # Execute and cache result
 result = func(*args, kwargs)
 cache.set(cache_key, result)
 return result

 return wrapper
 return decorator
```

### Format-Specific Optimizations

```python
class FormatSpecificOptimizer:
 """Optimizations for specific data formats."""

 def optimize_for_json(self, data: Dict) -> Dict:
 """Optimize data structure for JSON serialization."""
 optimized = {}

 for key, value in data.items():
 # Convert datetime objects to ISO strings
 if hasattr(value, 'isoformat'):
 optimized[key] = value.isoformat()
 # Convert sets to lists for JSON compatibility
 elif isinstance(value, set):
 optimized[key] = list(value)
 # Handle None values
 elif value is None:
 # Skip None values for compactness (optional)
 continue
 else:
 optimized[key] = value

 return optimized

 def optimize_for_yaml(self, data: Dict) -> Dict:
 """Optimize data structure for YAML serialization."""
 optimized = {}

 for key, value in data.items():
 # Preserve complex objects for YAML's better type support
 if hasattr(value, '__dict__'):
 optimized[key] = value
 # Use multi-line strings for long text
 elif isinstance(value, str) and len(value) > 100:
 optimized[key] = self._format_multiline_string(value)
 else:
 optimized[key] = value

 return optimized

 def _format_multiline_string(self, text: str) -> str:
 """Format long strings as YAML multi-line."""
 # Simple implementation - could be enhanced with smarter line breaks
 lines = text.split('\n')
 if len(lines) == 1:
 return text
 return '|\n' + '\n'.join(' ' + line for line in lines)

# Performance monitoring
class PerformanceMonitor:
 """Monitor and optimize data processing performance."""

 def __init__(self):
 self.metrics = {}

 def time_operation(self, operation_name: str):
 """Decorator to time operations."""
 def decorator(func):
 @wraps(func)
 def wrapper(*args, kwargs):
 start_time = time.time()
 result = func(*args, kwargs)
 duration = time.time() - start_time

 # Record metrics
 if operation_name not in self.metrics:
 self.metrics[operation_name] = []
 self.metrics[operation_name].append(duration)

 return result
 return wrapper
 return decorator

 def get_stats(self, operation_name: str) -> Dict:
 """Get performance statistics for an operation."""
 if operation_name not in self.metrics:
 return {}

 durations = self.metrics[operation_name]
 return {
 'count': len(durations),
 'total_time': sum(durations),
 'avg_time': sum(durations) / len(durations),
 'min_time': min(durations),
 'max_time': max(durations)
 }
```

## Best Practices

1. Use orjson for JSON: 2-5x faster than standard json module
2. Stream large files: Use ijson for memory-efficient processing
3. Cache intelligently: Implement smart caching with expiration
4. Batch processing: Process data in manageable chunks
5. Profile regularly: Monitor performance bottlenecks
6. Choose right format: YAML for config, JSON for data exchange, TOON for LLM communication

---

Module: `modules/json-optimization.md`
Related: [TOON Encoding](./toon-encoding.md) | [Data Validation](./data-validation.md)
