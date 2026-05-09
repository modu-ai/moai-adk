# Data Formats Reference

## API Reference

### TOONEncoder Class

Token-Optimized Object Notation encoder for LLM communication.

Initialization:
```python
from moai_formats_data import TOONEncoder

encoder = TOONEncoder(
    use_type_markers=True,      # Enable type prefix markers
    compress_keys=False,        # Enable key abbreviation
    datetime_format='iso',      # Datetime serialization format
    decimal_places=2            # Float precision
)
```

Methods:

encode(data):
- Parameters: data (dict, list, or primitive)
- Returns: str - TOON encoded string
- Achieves 40-60% token reduction vs JSON

decode(toon_string):
- Parameters: toon_string (str)
- Returns: dict/list/primitive - Original data
- Lossless round-trip decoding

encode_batch(data_list):
- Parameters: data_list (list of dicts)
- Returns: list of TOON encoded strings
- Optimized for batch processing

Type Markers:
```
# - Number (integer or float)
! - Boolean
@ - Timestamp/datetime
~ - Null value
$ - UUID (custom extension)
& - Decimal (custom extension)
```

Example:
```python
encoder = TOONEncoder()

# Original data
data = {
    "user_id": 12345,
    "active": True,
    "balance": 99.50,
    "created_at": datetime(2025, 1, 1),
    "metadata": None
}

# Encoded TOON
encoded = encoder.encode(data)
# Result: "user_id:#12345|active:!1|balance:#99.5|created_at:@2025-01-01T00:00:00|metadata:~"

# Decode back
decoded = encoder.decode(encoded)
# Returns original data structure
```

### JSONOptimizer Class

High-performance JSON processing with orjson.

Initialization:
```python
from moai_formats_data import JSONOptimizer

optimizer = JSONOptimizer(
    use_orjson=True,            # Use orjson for performance
    sort_keys=False,            # Sort dictionary keys
    indent=None,                # Pretty print indentation
    default_handler=None        # Custom type handler
)
```

Methods:

serialize_fast(data):
- Parameters: data (any serializable object)
- Returns: bytes - JSON encoded bytes
- 2-5x faster than standard json module

deserialize_fast(json_bytes):
- Parameters: json_bytes (bytes or str)
- Returns: dict/list - Parsed data

compress_schema(schema):
- Parameters: schema (dict) - JSON Schema
- Returns: bytes - Compressed schema for reuse

stream_parse(file_path, item_callback):
- Parameters: file_path (str), item_callback (callable)
- Returns: int - Number of items processed
- Memory-efficient streaming for large files

Example:
```python
optimizer = JSONOptimizer()

# Fast serialization
data = {"users": [{"id": i, "name": f"user_{i}"} for i in range(10000)]}
json_bytes = optimizer.serialize_fast(data)

# Fast deserialization
parsed = optimizer.deserialize_fast(json_bytes)

# Streaming large files
def process_user(user):
    print(f"Processing: {user['id']}")

count = optimizer.stream_parse("large_users.json", process_user)
print(f"Processed {count} users")
```

### DataValidator Class

Schema validation with custom rules.

Initialization:
```python
from moai_formats_data import DataValidator

validator = DataValidator(
    strict_mode=False,          # Fail on unknown fields
    coerce_types=True,          # Auto-convert compatible types
    error_limit=10              # Max errors to collect
)
```

Methods:

create_schema(rules):
- Parameters: rules (dict) - Field validation rules
- Returns: Schema object for validation

validate(data, schema):
- Parameters: data (dict), schema (Schema object)
- Returns: dict with keys: valid, errors, sanitized_data

add_custom_validator(name, validator_func):
- Parameters: name (str), validator_func (callable)
- Registers custom validation function

Rule Types:
```python
schema = validator.create_schema({
    # String validation
    "username": {
        "type": "string",
        "required": True,
        "min_length": 3,
        "max_length": 50,
        "pattern": r"^[a-zA-Z0-9_]+$"
    },

    # Email validation
    "email": {
        "type": "email",
        "required": True
    },

    # Number validation
    "age": {
        "type": "integer",
        "required": False,
        "min_value": 13,
        "max_value": 120
    },

    # Enum validation
    "role": {
        "type": "enum",
        "values": ["admin", "user", "guest"],
        "default": "user"
    },

    # Nested object
    "profile": {
        "type": "object",
        "schema": {
            "bio": {"type": "string", "max_length": 500},
            "avatar_url": {"type": "url"}
        }
    },

    # Array validation
    "tags": {
        "type": "array",
        "items": {"type": "string"},
        "min_items": 1,
        "max_items": 10,
        "unique": True
    }
})
```

### YAMLOptimizer Class

Optimized YAML processing.

```python
from moai_formats_data import YAMLOptimizer

yaml_optimizer = YAMLOptimizer(
    use_c_loader=True,          # Use LibYAML C extension
    preserve_order=True,        # Maintain key order
    default_flow_style=False    # Block style output
)

# Load YAML file
config = yaml_optimizer.load_fast("config.yaml")

# Merge multiple configs
merged = yaml_optimizer.merge_configs(
    base_config,
    env_config,
    override_config
)

# Dump to YAML
yaml_str = yaml_optimizer.dump_fast(data)
```

### StreamProcessor Class

Memory-efficient large file processing.

```python
from moai_formats_data import StreamProcessor

processor = StreamProcessor(
    chunk_size=8192,            # Read buffer size
    max_memory_mb=100           # Memory limit for buffering
)

# Process JSON array stream
def handle_item(item):
    # Process each item
    pass

processor.process_json_stream("large_file.json", handle_item)

# Process NDJSON (newline-delimited JSON)
processor.process_ndjson_stream("events.ndjson", handle_item)

# Process CSV with type inference
processor.process_csv_stream("data.csv", handle_item, infer_types=True)
```

---

## Configuration Options

### TOON Configuration

```yaml
# config/toon.yaml
encoding:
  use_type_markers: true
  compress_keys: false
  key_separator: ":"
  field_separator: "|"
  array_separator: ","

types:
  number_marker: "#"
  boolean_marker: "!"
  timestamp_marker: "@"
  null_marker: "~"
  uuid_marker: "$"
  decimal_marker: "&"

datetime:
  format: "iso"               # iso, unix, custom
  timezone: "UTC"
  custom_format: "%Y-%m-%dT%H:%M:%S%z"

performance:
  batch_size: 1000
  use_string_builder: true
  cache_patterns: true
```

### JSON Configuration

```yaml
# config/json.yaml
serialization:
  library: "orjson"           # orjson, ujson, standard
  sort_keys: false
  ensure_ascii: false
  indent: null

deserialization:
  parse_float: "float"        # float, decimal
  parse_int: "int"            # int, str
  strict: false

streaming:
  chunk_size: 8192
  buffer_size: 65536
  use_mmap: true              # Memory-mapped file reading

caching:
  enabled: true
  max_size_mb: 50
  ttl_seconds: 3600
```

### Validation Configuration

```yaml
# config/validation.yaml
behavior:
  strict_mode: false
  coerce_types: true
  strip_unknown: false
  error_limit: 50

defaults:
  string_max_length: 10000
  array_max_items: 1000
  object_max_depth: 10

custom_types:
  phone:
    pattern: "^\\+?[1-9]\\d{1,14}$"
    message: "Invalid phone number format"
  postal_code:
    pattern: "^[0-9]{5}(-[0-9]{4})?$"
    message: "Invalid postal code"

performance:
  compile_patterns: true
  cache_schemas: true
  parallel_validation: false
```

---

## Integration Patterns

### LLM Data Optimization

```python
from moai_formats_data import TOONEncoder, DataValidator

class LLMDataPreparer:
    def __init__(self, max_tokens: int = 4000):
        self.encoder = TOONEncoder()
        self.validator = DataValidator()
        self.max_tokens = max_tokens

    def prepare_for_llm(self, data: dict) -> str:
        # Validate data structure
        validation_result = self.validator.validate(data, self.schema)
        if not validation_result['valid']:
            raise ValueError(f"Invalid data: {validation_result['errors']}")

        # Encode to TOON format
        encoded = self.encoder.encode(data)

        # Check token budget
        estimated_tokens = len(encoded.split())
        if estimated_tokens > self.max_tokens:
            # Reduce data complexity
            reduced = self._reduce_complexity(data)
            encoded = self.encoder.encode(reduced)

        return encoded

    def _reduce_complexity(self, data: dict) -> dict:
        """Remove low-priority fields to fit token budget."""
        priority_fields = ['id', 'name', 'type', 'status']
        return {k: v for k, v in data.items() if k in priority_fields}

    def parse_llm_response(self, response: str) -> dict:
        """Parse TOON-encoded LLM response."""
        return self.encoder.decode(response)
```

### API Response Optimization

```python
from fastapi import FastAPI
from moai_formats_data import TOONEncoder, JSONOptimizer

app = FastAPI()
encoder = TOONEncoder()
optimizer = JSONOptimizer()

@app.get("/api/users/{user_id}")
async def get_user(user_id: int, format: str = "json"):
    user = await fetch_user(user_id)

    if format == "toon":
        # Return TOON format for LLM clients
        return Response(
            content=encoder.encode(user),
            media_type="application/x-toon"
        )
    else:
        # Return optimized JSON
        return Response(
            content=optimizer.serialize_fast(user),
            media_type="application/json"
        )

@app.post("/api/users/batch")
async def process_users(request: Request):
    # Stream process large request body
    users = []

    async for chunk in request.stream():
        data = optimizer.deserialize_fast(chunk)
        users.extend(data.get('users', []))

    # Process in batches
    results = await process_users_batch(users)
    return {"processed": len(results)}
```

### Database Integration

```python
from moai_formats_data import JSONOptimizer, DataValidator

class DatabaseCache:
    def __init__(self, redis_client):
        self.redis = redis_client
        self.optimizer = JSONOptimizer()
        self.validator = DataValidator()

    async def get_cached(self, key: str, schema=None):
        """Get cached data with optional validation."""
        cached = await self.redis.get(key)
        if not cached:
            return None

        data = self.optimizer.deserialize_fast(cached)

        if schema:
            result = self.validator.validate(data, schema)
            if not result['valid']:
                # Invalid cached data, delete and return None
                await self.redis.delete(key)
                return None
            return result['sanitized_data']

        return data

    async def set_cached(self, key: str, data: dict, ttl: int = 3600):
        """Cache data with fast serialization."""
        serialized = self.optimizer.serialize_fast(data)
        await self.redis.setex(key, ttl, serialized)

    async def cache_query_result(self, query_key: str, query_func, ttl: int = 3600):
        """Cache database query results."""
        cached = await self.get_cached(query_key)
        if cached:
            return cached

        result = await query_func()
        await self.set_cached(query_key, result, ttl)
        return result
```

---

## Troubleshooting

### TOON Encoding Issues

Issue: Decode fails with "Invalid type marker"
Symptoms: ValueError during decode
Solution:
- Check for unsupported data types in input
- Ensure encode/decode use same TOONEncoder configuration
- Verify no data corruption during transmission
- Use try/except with fallback to JSON

Issue: Large token savings not achieved
Symptoms: Only 10-20% reduction instead of 40-60%
Solution:
- Check data structure (nested objects benefit most)
- Enable key compression for repeated keys
- Remove redundant whitespace in string values
- Consider data structure optimization

### JSON Performance Issues

Issue: Serialization slower than expected
Symptoms: High CPU usage, slow response times
Solution:
- Verify orjson is installed: pip install orjson
- Check for non-serializable types requiring fallback
- Use bytes output instead of string when possible
- Batch small objects for single serialization call

Issue: Memory exhaustion with large files
Symptoms: MemoryError, system slowdown
Solution:
- Use streaming parser (ijson) for large files
- Process in chunks with StreamProcessor
- Enable memory-mapped file reading
- Implement pagination for large datasets

### Validation Issues

Issue: Custom validator not triggered
Symptoms: Custom rules ignored during validation
Solution:
- Verify validator registered before schema creation
- Check validator function signature (must accept value, return bool)
- Ensure field type matches custom validator name
- Debug with validator.get_registered_validators()

Issue: Schema evolution breaks existing data
Symptoms: Validation failures after schema update
Solution:
- Use SchemaEvolution for versioned migrations
- Implement data migration functions between versions
- Add backwards-compatible default values
- Test migration paths with production data samples

### Performance Optimization

Large Dataset Processing:
- Use NDJSON format for line-by-line processing
- Enable parallel processing for independent items
- Implement early termination for search operations
- Cache compiled regex patterns and schemas

Memory Management:
- Use generators instead of lists for streaming
- Clear cache periodically with cache.clear()
- Monitor memory with process.memory_info()
- Set explicit memory limits in configuration

---

## External Resources

### Core Libraries
- orjson: https://github.com/ijl/orjson
- ujson: https://github.com/ultrajson/ultrajson
- ijson: https://github.com/ICRAR/ijson
- PyYAML: https://pyyaml.org/wiki/PyYAMLDocumentation

### Validation Libraries
- Pydantic: https://docs.pydantic.dev/
- Cerberus: https://docs.python-cerberus.org/
- Marshmallow: https://marshmallow.readthedocs.io/
- JSON Schema: https://json-schema.org/

### Performance Tools
- memory_profiler: https://pypi.org/project/memory-profiler/
- line_profiler: https://github.com/pyutils/line_profiler
- py-spy: https://github.com/benfred/py-spy

### Serialization Standards
- JSON Specification: https://www.json.org/
- YAML Specification: https://yaml.org/spec/
- MessagePack: https://msgpack.org/
- Protocol Buffers: https://protobuf.dev/

### Best Practices
- Google JSON Style Guide: https://google.github.io/styleguide/jsoncstyleguide.xml
- JSON API Specification: https://jsonapi.org/
- OpenAPI Data Types: https://swagger.io/docs/specification/data-models/

---

Version: 1.0.0
Last Updated: 2025-12-06
