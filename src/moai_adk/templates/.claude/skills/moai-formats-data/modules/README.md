# MoAI Data Format Skill Modules

This directory contains the detailed implementation modules for the moai-formats-data skill, following Claude Code's progressive disclosure pattern.

## Module Structure

Each module focuses on a specific aspect of data format handling:

### Core Modules

- [`toon-encoding.md`](./toon-encoding.md)
 - Complete TOON (Token-Optimized Object Notation) implementation
 - 40-60% token reduction vs JSON for LLM communication
 - Type markers, encoding/decoding, and performance characteristics

- [`json-optimization.md`](./json-optimization.md)
 - High-performance JSON and YAML processing
 - Ultra-fast serialization with orjson
 - Streaming processing for large datasets
 - Memory management and batch processing

- [`data-validation.md`](./data-validation.md)
 - Comprehensive data validation system
 - Schema management and evolution
 - Advanced validation patterns and cross-field validation
 - Performance-optimized batch validation

- [`caching-performance.md`](./caching-performance.md)
 - Intelligent caching strategies
 - Multi-level caching with memory management
 - Cache invalidation and warming strategies
 - Performance monitoring and optimization

## Usage

These modules are referenced from the main skill file for detailed implementation examples and advanced use cases. Each module follows a consistent structure:

1. Module header with complexity, time, and dependencies
2. Core implementation with complete code examples
3. Advanced features and extensions
4. Performance characteristics and best practices
5. Cross-references to related modules

## Integration

All modules work together seamlessly:

```python
# Example of integrating multiple modules
from moai_formats_data import (
 TOONEncoder, # from toon-encoding module
 JSONOptimizer, # from json-optimization module
 DataValidator, # from data-validation module
 SmartCache # from caching-performance module
)

# Combine TOON encoding with validation and caching
encoder = TOONEncoder()
validator = DataValidator()
cache = SmartCache()

@cache.cache.cache_result(ttl=1800)
def process_and_validate_data(data: Dict) -> str:
 # Validate first
 schema = validator.create_schema({
 "user": {"type": "string", "required": True},
 "value": {"type": "number", "required": True}
 })

 result = validator.validate(data, schema)
 if not result['valid']:
 raise ValueError(f"Invalid data: {result['errors']}")

 # Encode optimized format
 return encoder.encode(result['sanitized_data'])
```

## Performance Characteristics

- TOON Encoding: 40-60% token reduction vs JSON
- JSON Processing: 2-5x faster with orjson
- Validation: Optimized batch processing with pattern compilation
- Caching: Intelligent LRU eviction with memory pressure awareness

## Dependencies

Core requirements:
- Python 3.8+
- `typing`, `dataclasses`, `enum`, `re`, `datetime`

Performance libraries (optional):
- `orjson` - Ultra-fast JSON processing
- `PyYAML` - High-performance YAML handling
- `ijson` - Streaming JSON parser

Validation libraries (optional):
- `jsonschema` - JSON Schema validation
- `pydantic` - Type hint based validation
