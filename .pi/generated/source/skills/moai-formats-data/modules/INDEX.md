# MoAI Data Format Skill Modules

This directory contains the detailed implementation modules for the moai-formats-data skill, following Claude Code's progressive disclosure pattern.

## Module Architecture

Each module focuses on a specific aspect of data format handling, providing comprehensive implementation guidance, advanced patterns, and production-ready examples.

### Progressive Disclosure Structure

**SKILL.md** (247 lines):
- Quick Reference: 30-second overview of core capabilities
- Implementation Guide: 5-minute basic usage patterns
- Advanced Features Overview: Links to detailed modules

**Modules** (detailed deep-dives):
- Complete implementation examples
- Advanced patterns and extensions
- Performance characteristics and optimization
- Integration patterns and best practices

**Supporting Files**:
- `reference.md`: Extended reference documentation (585 lines)
- `examples.md`: Complete working examples (804 lines)

---

## Core Modules

### 1. TOON Encoding Module

**File**: [`toon-encoding.md`](./toon-encoding.md) (308 lines)

**Purpose**: Token-Optimized Object Notation for LLM communication

**Key Features**:
- 40-60% token reduction vs JSON for typical data structures
- Custom type markers for optimized representation
- Lossless round-trip encoding/decoding
- Streaming and batch processing support

**Topics Covered**:
- Core TOON encoding algorithm
- Type markers and value representation
- Custom type handlers (UUID, Decimal, etc.)
- Streaming TOON processing
- Performance benchmarks and optimization
- Integration with LLM workflows

**When to Use**:
- Transmitting data to LLMs within token budgets
- Optimizing API responses for AI consumption
- Reducing context window usage
- Large dataset processing for LLM pipelines

**Complexity**: Intermediate | **Time**: 15 minutes | **Dependencies**: None

---

### 2. JSON/YAML Optimization Module

**File**: [`json-optimization.md`](./json-optimization.md) (374 lines)

**Purpose**: High-performance JSON and YAML processing

**Key Features**:
- Ultra-fast serialization with orjson (2-5x faster)
- Streaming processing for large datasets
- Schema compression and caching
- Memory-efficient parsing

**Topics Covered**:
- Fast JSON serialization/deserialization
- Streaming JSON processing with ijson
- YAML configuration management
- Schema compression techniques
- Format conversion utilities
- Memory management strategies

**When to Use**:
- Processing large JSON files (>100MB)
- High-performance API responses
- Configuration file management
- Data transformation pipelines

**Complexity**: Intermediate | **Time**: 20 minutes | **Dependencies**: orjson, ijson, PyYAML

---

### 3. Data Validation Module

**File**: [`data-validation.md`](./data-validation.md) (485 lines)

**Purpose**: Comprehensive data validation and schema management

**Key Features**:
- Type-safe validation with custom rules
- Schema evolution and migration
- Cross-field validation
- Batch validation optimization

**Topics Covered**:
- Schema creation and management
- Type checking and validation rules
- Cross-field validation patterns
- Schema evolution strategies
- Custom validation rules
- Batch processing optimization
- Error handling and reporting

**When to Use**:
- Validating user input and API requests
- Schema-driven data processing
- Data integrity verification
- Configuration validation

**Complexity**: Advanced | **Time**: 30 minutes | **Dependencies**: jsonschema, pydantic (optional)

---

### 4. Caching and Performance Module

**File**: [`caching-performance.md`](./caching-performance.md) (459 lines)

**Purpose**: Intelligent caching strategies and performance optimization

**Key Features**:
- Multi-level caching with LRU eviction
- Memory-aware cache management
- Cache warming and invalidation
- Performance monitoring

**Topics Covered**:
- Intelligent caching strategies
- LRU cache implementation
- Memory pressure management
- Cache warming patterns
- Invalidation strategies
- Performance monitoring and metrics
- Benchmarking techniques

**When to Use**:
- Expensive data processing operations
- Repeated validation or serialization
- High-performance requirements
- Memory-constrained environments

**Complexity**: Advanced | **Time**: 25 minutes | **Dependencies**: functools, hashlib

---

## Integration Patterns

### Combining Multiple Modules

All modules work together seamlessly for comprehensive data format management:

```python
from moai_formats_data import (
    TOONEncoder,        # from toon-encoding module
    JSONOptimizer,      # from json-optimization module
    DataValidator,      # from data-validation module
    SmartCache          # from caching-performance module
)

# Complete data processing pipeline
class DataProcessor:
    def __init__(self):
        self.encoder = TOONEncoder()
        self.optimizer = JSONOptimizer()
        self.validator = DataValidator()
        self.cache = SmartCache(max_memory_mb=50)

    @SmartCache.cache_result(ttl=1800)  # 30 minutes
    def process_and_validate(self, data: Dict) -> str:
        # Step 1: Validate input
        schema = self.validator.create_schema({
            "user": {"type": "string", "required": True},
            "value": {"type": "number", "required": True}
        })

        result = self.validator.validate(data, schema)
        if not result['valid']:
            raise ValueError(f"Invalid data: {result['errors']}")

        # Step 2: Optimize format
        optimized = self.optimizer.serialize_fast(result['sanitized_data'])

        # Step 3: Encode for LLM
        return self.encoder.encode(result['sanitized_data'])

# Usage
processor = DataProcessor()
result = processor.process_and_validate({"user": "john", "value": 42})
```

### Module Dependencies

**TOON Encoding**:
- Standalone module
- No external dependencies required
- Optional: datetime for timestamp handling

**JSON/YAML Optimization**:
- Optional: orjson for ultra-fast JSON
- Optional: ijson for streaming processing
- Optional: PyYAML for YAML support

**Data Validation**:
- Optional: jsonschema for schema validation
- Optional: pydantic for type hint validation
- Optional: cerberus for lightweight validation

**Caching and Performance**:
- Built-in: functools, hashlib
- No external dependencies required

---

## Performance Characteristics

### Benchmarks

**TOON Encoding**:
- Token reduction: 40-60% vs JSON
- Encoding speed: ~1MB/s
- Decoding speed: ~2MB/s
- Memory overhead: ~10% vs JSON

**JSON Processing** (with orjson):
- Serialization: 2-5x faster than standard json
- Deserialization: 2-3x faster than standard json
- Memory usage: ~30% lower overhead
- Streaming: Constant memory usage

**Data Validation**:
- Single validation: ~0.1ms per record
- Batch validation: ~0.05ms per record (compiled)
- Schema compilation: ~10ms one-time cost
- Memory usage: ~1KB per schema

**Caching**:
- Cache hit: ~0.001ms (in-memory)
- Cache miss: Variable (depends on operation)
- Memory overhead: ~20% vs uncached
- LRU eviction: O(1) complexity

---

## Best Practices

### When to Use Each Module

**TOON Encoding**:
- Optimal for: LLM communication, token budget optimization
- Less optimal for: Human-readable data, long-term storage
- Best practice: Use for LLM contexts, JSON for everything else

**JSON/YAML Optimization**:
- Optimal for: High-performance APIs, large dataset processing
- Less optimal for: Simple one-off operations
- Best practice: Use orjson for hot paths, standard json for cold paths

**Data Validation**:
- Optimal for: User input, API requests, configuration validation
- Less optimal for: Trusted internal data
- Best practice: Validate at boundaries, validate once

**Caching and Performance**:
- Optimal for: Expensive operations, repeated queries
- Less optimal for: Unique operations, real-time data
- Best practice: Cache judiciously, monitor hit rates

### Common Pitfalls

**TOON Encoding**:
- Don't use for human-readable configuration
- Don't use for long-term data storage
- Remember to validate before encoding

**JSON/YAML Optimization**:
- Don't optimize prematurely
- Don't use orjson if compatibility is critical
- Don't forget error handling for large files

**Data Validation**:
- Don't validate internal data repeatedly
- Don't create overly complex schemas
- Don't ignore validation errors

**Caching and Performance**:
- Don't cache volatile data
- Don't ignore cache invalidation
- Don't cache without monitoring

---

## Version History

**v2.0.0 (2026-01-06)**:
- Expanded README from 98 to 250+ lines
- Added comprehensive integration patterns
- Added performance characteristics section
- Added best practices and common pitfalls
- Enhanced module cross-references

**v1.0.0 (2025-12-06)**:
- Initial modular structure
- Four core modules established
- Basic integration examples

---

## Contributing

When adding new modules or updating existing ones:

1. Follow the progressive disclosure pattern
2. Include complete working examples
3. Add performance characteristics
4. Document dependencies clearly
5. Cross-reference related modules
6. Keep modules under 500 lines

---

**Status**: Active | **Last Updated**: 2026-01-06
