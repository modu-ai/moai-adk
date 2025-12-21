---
name: moai-formats-data
description: Data format specialist covering TOON encoding, JSON/YAML optimization, serialization patterns, and data validation for modern applications
version: 1.0.0
category: library
allowed-tools: Read, Write, Edit, Grep, Glob
tags:
 - formats
 - data
 - toon
 - serialization
 - validation
 - optimization
updated: 2025-12-06
status: active
author: MoAI-ADK Team
---

# Data Format Specialist

## Quick Reference (30 seconds)

Advanced Data Format Management - Comprehensive data handling covering TOON encoding, JSON/YAML optimization, serialization patterns, and data validation for performance-critical applications.

Core Capabilities:
- TOON Encoding: 40-60% token reduction vs JSON for LLM communication
- JSON/YAML Optimization: Efficient serialization and parsing patterns
- Data Validation: Schema validation, type checking, error handling
- Format Conversion: Seamless transformation between data formats
- Performance: Optimized data structures and caching strategies
- Schema Management: Dynamic schema generation and evolution

When to Use:
- Optimizing data transmission to LLMs within token budgets
- High-performance serialization/deserialization
- Schema validation and data integrity
- Format conversion and data transformation
- Large dataset processing and optimization

Quick Start:
```python
# TOON encoding (40-60% token reduction)
from moai_formats_data import TOONEncoder
encoder = TOONEncoder()
compressed = encoder.encode({"user": "John", "age": 30})
original = encoder.decode(compressed)

# Fast JSON processing
from moai_formats_data import JSONOptimizer
optimizer = JSONOptimizer()
fast_json = optimizer.serialize_fast(large_dataset)

# Data validation
from moai_formats_data import DataValidator
validator = DataValidator()
schema = validator.create_schema({"name": {"type": "string", "required": True}})
result = validator.validate({"name": "John"}, schema)
```

---

## Implementation Guide (5 minutes)

### Core Concepts

TOON (Token-Optimized Object Notation):
- Custom binary-compatible format optimized for LLM token usage
- Type markers: # (numbers), ! (booleans), @ (timestamps), ~ (null)
- 40-60% size reduction vs JSON for typical data structures
- Lossless round-trip encoding/decoding

Performance Optimization:
- Ultra-fast JSON processing with orjson (2-5x faster than standard json)
- Streaming processing for large datasets using ijson
- Intelligent caching with LRU eviction and memory management
- Schema compression and validation optimization

Data Validation:
- Type-safe validation with custom rules and patterns
- Schema evolution and migration support
- Cross-field validation and dependency checking
- Performance-optimized batch validation

### Basic Implementation

```python
from moai_formats_data import TOONEncoder, JSONOptimizer, DataValidator
from datetime import datetime

# 1. TOON Encoding for LLM optimization
encoder = TOONEncoder()
data = {
 "user": {"id": 123, "name": "John", "active": True, "created": datetime.now()},
 "permissions": ["read", "write", "admin"]
}

# Encode and compare sizes
toon_data = encoder.encode(data)
original_data = encoder.decode(toon_data)

# 2. Fast JSON Processing
optimizer = JSONOptimizer()

# Ultra-fast serialization
json_bytes = optimizer.serialize_fast(data)
parsed_data = optimizer.deserialize_fast(json_bytes)

# Schema compression for repeated validation
schema = {"type": "object", "properties": {"name": {"type": "string"}}}
compressed_schema = optimizer.compress_schema(schema)

# 3. Data Validation
validator = DataValidator()

# Create validation schema
user_schema = validator.create_schema({
 "username": {"type": "string", "required": True, "min_length": 3},
 "email": {"type": "email", "required": True},
 "age": {"type": "integer", "required": False, "min_value": 13}
})

# Validate data
user_data = {"username": "john_doe", "email": "john@example.com", "age": 30}
result = validator.validate(user_data, user_schema)

if result['valid']:
 print("Data is valid!")
 sanitized = result['sanitized_data']
else:
 print("Validation errors:", result['errors'])
```

### Common Use Cases

API Response Optimization:
```python
# Optimize API responses for LLM consumption
def optimize_api_response(data: Dict) -> str:
 encoder = TOONEncoder()
 return encoder.encode(data)

# Parse optimized responses
def parse_optimized_response(toon_data: str) -> Dict:
 encoder = TOONEncoder()
 return encoder.decode(toon_data)
```

Configuration Management:
```python
# Fast YAML configuration loading
from moai_formats_data import YAMLOptimizer

yaml_optimizer = YAMLOptimizer()
config = yaml_optimizer.load_fast("config.yaml")

# Merge multiple configurations
merged = yaml_optimizer.merge_configs(base_config, env_config, user_config)
```

Large Dataset Processing:
```python
# Stream processing for large JSON files
from moai_formats_data import StreamProcessor

processor = StreamProcessor(chunk_size=8192)

# Process file line by line without loading into memory
def process_item(item):
 print(f"Processing: {item['id']}")

processor.process_json_stream("large_dataset.json", process_item)
```

---

## Advanced Features (10+ minutes)

### Advanced TOON Features

Custom Type Handlers:
```python
# Extend TOON encoder with custom types
class CustomTOONEncoder(TOONEncoder):
 def _encode_value(self, value):
 # Handle UUID objects
 if hasattr(value, 'hex') and len(value.hex) == 32: # UUID
 return f'${value.hex}'

 # Handle Decimal objects
 elif hasattr(value, 'as_tuple'): # Decimal
 return f'&{str(value)}'

 return super()._encode_value(value)

 def _parse_value(self, s):
 # Parse custom UUIDs
 if s.startswith('$') and len(s) == 33:
 import uuid
 return uuid.UUID(s[1:])

 # Parse custom Decimals
 elif s.startswith('&'):
 from decimal import Decimal
 return Decimal(s[1:])

 return super()._parse_value(s)
```

Streaming TOON Processing:
```python
# Process TOON data in streaming mode
def stream_toon_data(data_generator):
 encoder = TOONEncoder()
 for data in data_generator:
 yield encoder.encode(data)

# Batch TOON processing
def batch_encode_toon(data_list: List[Dict], batch_size: int = 1000):
 encoder = TOONEncoder()
 results = []

 for i in range(0, len(data_list), batch_size):
 batch = data_list[i:i + batch_size]
 encoded_batch = [encoder.encode(item) for item in batch]
 results.extend(encoded_batch)

 return results
```

### Advanced Validation Patterns

Cross-Field Validation:
```python
# Validate relationships between fields
class CrossFieldValidator:
 def __init__(self):
 self.base_validator = DataValidator()

 def validate_user_data(self, data: Dict) -> Dict:
 # Base validation
 schema = self.base_validator.create_schema({
 "password": {"type": "string", "required": True, "min_length": 8},
 "confirm_password": {"type": "string", "required": True},
 "email": {"type": "email", "required": True}
 })

 result = self.base_validator.validate(data, schema)

 # Cross-field validation
 if data.get("password") != data.get("confirm_password"):
 result['errors']['password_mismatch'] = "Passwords do not match"
 result['valid'] = False

 return result
```

Schema Evolution:
```python
# Handle schema changes over time
from moai_formats_data import SchemaEvolution

evolution = SchemaEvolution()

# Define schema versions
v1_schema = {"name": {"type": "string"}, "age": {"type": "integer"}}
v2_schema = {"full_name": {"type": "string"}, "age": {"type": "integer"}, "email": {"type": "email"}}

# Register schemas
evolution.register_schema("v1", v1_schema)
evolution.register_schema("v2", v2_schema)

# Add migration function
def migrate_v1_to_v2(data: Dict) -> Dict:
 return {
 "full_name": data["name"],
 "age": data["age"],
 "email": None # New required field
 }

evolution.add_migration("v1", "v2", migrate_v1_to_v2)

# Migrate data
old_data = {"name": "John Doe", "age": 30}
new_data = evolution.migrate_data(old_data, "v1", "v2")
```

### Performance Optimization

Intelligent Caching:
```python
from moai_formats_data import SmartCache

# Create cache with memory constraints
cache = SmartCache(max_memory_mb=50, max_items=10000)

@cache.cache.cache_result(ttl=1800) # 30 minutes
def expensive_data_processing(data: Dict) -> Dict:
 # Simulate expensive computation
 time.sleep(0.1)
 return {"processed": True, "data": data}

# Cache statistics
print(cache.get_stats())

# Cache warming
def warm_common_data():
 common_queries = [
 {"type": "user", "id": 1},
 {"type": "user", "id": 2},
 {"type": "config", "key": "app"}
 ]

 for query in common_queries:
 expensive_data_processing(query)

warm_common_data()
```

Batch Processing Optimization:
```python
# Optimized batch validation
def validate_batch_optimized(data_list: List[Dict], schema: Dict) -> List[Dict]:
 validator = DataValidator()

 # Pre-compile patterns for performance
 validator._compile_schema_patterns(schema)

 # Process in batches for memory efficiency
 batch_size = 1000
 results = []

 for i in range(0, len(data_list), batch_size):
 batch = data_list[i:i + batch_size]
 batch_results = [validator.validate(data, schema) for data in batch]
 results.extend(batch_results)

 return results
```

### Integration Patterns

LLM Integration:
```python
# Prepare data for LLM consumption
def prepare_for_llm(data: Dict, max_tokens: int = 2000) -> str:
 encoder = TOONEncoder()
 toon_data = encoder.encode(data)

 # Check token count
 estimated_tokens = len(toon_data.split())

 if estimated_tokens > max_tokens:
 # Implement data reduction strategy
 reduced_data = reduce_data_complexity(data, max_tokens)
 toon_data = encoder.encode(reduced_data)

 return toon_data

def reduce_data_complexity(data: Dict, max_tokens: int) -> Dict:
 """Reduce data complexity to fit token budget."""
 # Implement selective field removal
 priority_fields = ["id", "name", "email", "status"]
 reduced = {k: v for k, v in data.items() if k in priority_fields}

 # Further reduction if needed
 encoder = TOONEncoder()
 while len(encoder.encode(reduced).split()) > max_tokens:
 # Remove least important fields
 if len(reduced) <= 1:
 break
 reduced.popitem()

 return reduced
```

Database Integration:
```python
# Optimize database queries with format conversion
def optimize_db_response(db_data: List[Dict]) -> Dict:
 # Convert database results to optimized format
 optimizer = JSONOptimizer()

 # Compress and cache schema
 common_schema = {"type": "object", "properties": {"id": {"type": "integer"}}}
 compressed_schema = optimizer.compress_schema(common_schema)

 # Process in batches
 processor = StreamProcessor()
 processed_data = []

 for item in db_data:
 # Apply validation and transformation
 processed_item = transform_db_item(item)
 processed_data.append(processed_item)

 return {
 "data": processed_data,
 "count": len(processed_data),
 "schema": compressed_schema
 }
```

---

## Works Well With

- moai-domain-backend - Backend data serialization and API responses
- moai-domain-database - Database data format optimization
- moai-foundation-core - MCP data serialization and transmission patterns
- moai-docs-generation - Documentation data formatting
- moai-foundation-core - Core data architecture principles

---

## Module References

Core Implementation Modules:
- [`modules/toon-encoding.md`](./modules/toon-encoding.md) - TOON encoding implementation and examples
- [`modules/json-optimization.md`](./modules/json-optimization.md) - High-performance JSON/YAML processing
- [`modules/data-validation.md`](./modules/data-validation.md) - Advanced validation and schema management
- [`modules/caching-performance.md`](./modules/caching-performance.md) - Caching strategies and performance optimization

---

## Usage Examples

### CLI Usage
```bash
# Encode data to TOON format
moai-formats encode-toon --input data.json --output data.toon

# Validate data against schema
moai-formats validate --schema schema.json --data data.json

# Convert between formats
moai-formats convert --input data.json --output data.yaml --format yaml

# Optimize JSON structure
moai-formats optimize-json --input large-data.json --output optimized.json
```

### Python API
```python
from moai_formats_data import TOONEncoder, DataValidator, JSONOptimizer

# TOON encoding
encoder = TOONEncoder()
toon_data = encoder.encode({"user": "John", "age": 30})
original_data = encoder.decode(toon_data)

# Data validation
validator = DataValidator()
schema = validator.create_schema({
 "name": {"type": "string", "required": True, "min_length": 2},
 "email": {"type": "email", "required": True}
})
result = validator.validate({"name": "John", "email": "john@example.com"}, schema)

# JSON optimization
optimizer = JSONOptimizer()
fast_json = optimizer.serialize_fast(large_dataset)
parsed_data = optimizer.deserialize_fast(fast_json)
```

---

## Technology Stack

Core Libraries:
- orjson: Ultra-fast JSON parsing and serialization
- PyYAML: YAML processing with C-based loaders
- ijson: Streaming JSON parser for large files
- python-dateutil: Advanced datetime parsing
- regex: Advanced regular expression support

Performance Tools:
- lru_cache: Built-in memoization
- pickle: Object serialization
- hashlib: Hash generation for caching
- functools: Function decorators and utilities

Validation Libraries:
- jsonschema: JSON Schema validation
- cerberus: Lightweight data validation
- marshmallow: Object serialization/deserialization
- pydantic: Data validation using Python type hints

---

Status: Production Ready
Last Updated: 2025-11-30
Maintained by: MoAI-ADK Data Team
