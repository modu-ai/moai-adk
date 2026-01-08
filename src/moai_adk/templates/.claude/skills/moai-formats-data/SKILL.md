---
name: "moai-formats-data"
description: "Data format specialist covering TOON encoding, JSON/YAML optimization, serialization patterns, and data validation for modern applications. Use when optimizing data for LLM transmission, implementing high-performance serialization, validating data schemas, or converting between data formats."
version: 2.0.0
category: "library"
modularized: true
user-invocable: false
tags: ['formats', 'data', 'toon', 'serialization', 'validation', 'optimization']
updated: 2026-01-08
allowed-tools:
  - Read
  - Write
  - Edit
  - Grep
  - Glob
  - mcp__context7__resolve-library-id
  - mcp__context7__get-library-docs
status: "active"
author: "MoAI-ADK Team"
---

# Data Format Specialist

## Quick Reference (30 seconds)

Advanced Data Format Management - Comprehensive data handling covering TOON encoding, JSON/YAML optimization, serialization patterns, and data validation for performance-critical applications.

**Core Capabilities:**
- **TOON Encoding**: 40-60% token reduction vs JSON for LLM communication
- **JSON/YAML Optimization**: Efficient serialization and parsing patterns
- **Data Validation**: Schema validation, type checking, error handling
- **Format Conversion**: Seamless transformation between data formats
- **Performance**: Optimized data structures and caching strategies
- **Schema Management**: Dynamic schema generation and evolution

**When to Use:**
- Optimizing data transmission to LLMs within token budgets
- High-performance serialization/deserialization
- Schema validation and data integrity
- Format conversion and data transformation
- Large dataset processing and optimization

**Quick Start:**
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

**TOON (Token-Optimized Object Notation):**
- Custom binary-compatible format optimized for LLM token usage
- Type markers: # (numbers), ! (booleans), @ (timestamps), ~ (null)
- 40-60% size reduction vs JSON for typical data structures
- Lossless round-trip encoding/decoding

**Performance Optimization:**
- Ultra-fast JSON processing with orjson (2-5x faster than standard json)
- Streaming processing for large datasets using ijson
- Intelligent caching with LRU eviction and memory management
- Schema compression and validation optimization

**Data Validation:**
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

**API Response Optimization:**
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

**Configuration Management:**
```python
# Fast YAML configuration loading
from moai_formats_data import YAMLOptimizer

yaml_optimizer = YAMLOptimizer()
config = yaml_optimizer.load_fast("config.yaml")

# Merge multiple configurations
merged = yaml_optimizer.merge_configs(base_config, env_config, user_config)
```

**Large Dataset Processing:**
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

## Advanced Features Overview

### Advanced TOON Features

See [`modules/toon-encoding.md`](./modules/toon-encoding.md) for:
- Custom type handlers (UUID, Decimal, etc.)
- Streaming TOON processing
- Batch TOON encoding
- Performance characteristics and benchmarks

### Advanced Validation Patterns

See [`modules/data-validation.md`](./modules/data-validation.md) for:
- Cross-field validation
- Schema evolution and migration
- Custom validation rules
- Batch validation optimization

### Performance Optimization

See [`modules/caching-performance.md`](./modules/caching-performance.md) for:
- Intelligent caching strategies
- Cache warming and invalidation
- Memory management
- Performance monitoring

### JSON/YAML Advanced Features

See [`modules/json-optimization.md`](./modules/json-optimization.md) for:
- Streaming JSON processing
- Memory-efficient parsing
- Schema compression
- Format conversion utilities

---

## Works Well With

- **moai-domain-backend** - Backend data serialization and API responses
- **moai-domain-database** - Database data format optimization
- **moai-foundation-core** - MCP data serialization and transmission patterns
- **moai-workflow-docs** - Documentation data formatting
- **moai-foundation-context** - Context optimization for token budgets

---

## Module References

**Core Implementation Modules:**
- [`modules/toon-encoding.md`](./modules/toon-encoding.md) - TOON encoding implementation (308 lines)
- [`modules/json-optimization.md`](./modules/json-optimization.md) - High-performance JSON/YAML (374 lines)
- [`modules/data-validation.md`](./modules/data-validation.md) - Advanced validation and schemas (485 lines)
- [`modules/caching-performance.md`](./modules/caching-performance.md) - Caching strategies (459 lines)

**Supporting Files:**
- [`modules/README.md`](./modules/README.md) - Module overview and integration patterns (98 lines)
- [`reference.md`](./reference.md) - Extended reference documentation (585 lines)
- [`examples.md`](./examples.md) - Complete working examples (804 lines)

---

## Technology Stack

**Core Libraries:**
- orjson: Ultra-fast JSON parsing and serialization
- PyYAML: YAML processing with C-based loaders
- ijson: Streaming JSON parser for large files
- python-dateutil: Advanced datetime parsing
- regex: Advanced regular expression support

**Performance Tools:**
- lru_cache: Built-in memoization
- pickle: Object serialization
- hashlib: Hash generation for caching
- functools: Function decorators and utilities

**Validation Libraries:**
- jsonschema: JSON Schema validation
- cerberus: Lightweight data validation
- marshmallow: Object serialization/deserialization
- pydantic: Data validation using Python type hints

---

## Version History

**v2.0.0 (2026-01-06):**
- Modularized skill with progressive disclosure pattern
- Reduced SKILL.md from 493 to 247 lines (50% reduction)
- Organized advanced content into focused modules
- Enhanced module cross-references and integration patterns
- Updated Context7 MCP tools in allowed-tools

**v1.0.0 (2025-12-06):**
- Initial release with monolithic structure
- TOON encoding, JSON/YAML optimization, data validation
- Basic caching and performance features

---

**Status:** Production Ready
**Last Updated:** 2026-01-06
**Maintained by:** MoAI-ADK Data Team
