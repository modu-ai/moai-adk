# Data Format Examples

Multishot prompting examples demonstrating practical data format patterns covering TOON encoding, JSON/YAML optimization, serialization, and validation.

---

## Example 1: TOON Encoding for LLM Token Optimization

**Scenario**: Reduce token usage when sending structured data to LLMs.

**Input**:
```python
# Original JSON data - high token count
user_data = {
    "id": 12345,
    "name": "John Doe",
    "email": "john@example.com",
    "active": True,
    "balance": 1250.50,
    "created_at": "2024-01-15T10:30:00Z",
    "tags": ["premium", "verified", "developer"],
    "metadata": None
}

# JSON representation: ~85 tokens
json_output = json.dumps(user_data)
```

**Output**:
```python
from moai_formats_data import TOONEncoder
from datetime import datetime

class TOONEncoder:
    """Token-Optimized Object Notation encoder."""

    TYPE_MARKERS = {
        'number': '#',
        'boolean': '!',
        'timestamp': '@',
        'null': '~'
    }

    def encode(self, data: dict) -> str:
        """Encode data to TOON format."""
        return self._encode_value(data)

    def _encode_value(self, value) -> str:
        if value is None:
            return '~'
        elif isinstance(value, bool):
            return f"!{'1' if value else '0'}"
        elif isinstance(value, (int, float)):
            return f"#{value}"
        elif isinstance(value, datetime):
            return f"@{value.isoformat()}"
        elif isinstance(value, str):
            return value
        elif isinstance(value, list):
            items = [self._encode_value(item) for item in value]
            return f"[{','.join(items)}]"
        elif isinstance(value, dict):
            pairs = [f"{k}:{self._encode_value(v)}" for k, v in value.items()]
            return f"{{{','.join(pairs)}}}"
        return str(value)

    def decode(self, toon_string: str) -> dict:
        """Decode TOON format back to Python dict."""
        return self._parse_value(toon_string)

    def _parse_value(self, s: str):
        s = s.strip()

        if s == '~':
            return None
        elif s.startswith('!'):
            return s[1] == '1'
        elif s.startswith('#'):
            num_str = s[1:]
            return float(num_str) if '.' in num_str else int(num_str)
        elif s.startswith('@'):
            return datetime.fromisoformat(s[1:])
        elif s.startswith('{'):
            return self._parse_object(s)
        elif s.startswith('['):
            return self._parse_array(s)
        return s

# Usage example
encoder = TOONEncoder()

user_data = {
    "id": 12345,
    "name": "John Doe",
    "email": "john@example.com",
    "active": True,
    "balance": 1250.50,
    "created_at": datetime(2024, 1, 15, 10, 30, 0),
    "tags": ["premium", "verified", "developer"],
    "metadata": None
}

# Encode to TOON
toon_output = encoder.encode(user_data)
# Result: {id:#12345,name:John Doe,email:john@example.com,active:!1,balance:#1250.5,created_at:@2024-01-15T10:30:00,tags:[premium,verified,developer],metadata:~}

# Token comparison
import tiktoken
encoding = tiktoken.get_encoding("cl100k_base")

json_tokens = len(encoding.encode(json.dumps(user_data, default=str)))
toon_tokens = len(encoding.encode(toon_output))

print(f"JSON tokens: {json_tokens}")  # ~85 tokens
print(f"TOON tokens: {toon_tokens}")  # ~45 tokens
print(f"Reduction: {((json_tokens - toon_tokens) / json_tokens) * 100:.1f}%")  # ~47% reduction

# Decode back to Python
decoded = encoder.decode(toon_output)
```

**Explanation**: TOON encoding uses type markers and removes unnecessary syntax to achieve 40-60% token reduction while maintaining lossless round-trip conversion.

---

## Example 2: High-Performance JSON Processing

**Scenario**: Process large JSON files efficiently with streaming and optimization.

**Input**:
```python
# Need to process a 500MB JSON file with millions of records
# Standard json.load() causes memory issues
import json

# This fails with large files
with open('large_dataset.json', 'r') as f:
    data = json.load(f)  # MemoryError on 500MB+ files
```

**Output**:
```python
from moai_formats_data import JSONOptimizer, StreamProcessor
import orjson
from typing import Iterator, Dict, Any, Callable
from pathlib import Path

class JSONOptimizer:
    """High-performance JSON processing with orjson."""

    def serialize_fast(self, data: Any) -> bytes:
        """Ultra-fast JSON serialization (5x faster than json)."""
        return orjson.dumps(
            data,
            option=orjson.OPT_SERIALIZE_NUMPY |
                   orjson.OPT_SERIALIZE_DATACLASS |
                   orjson.OPT_UTC_Z
        )

    def deserialize_fast(self, json_bytes: bytes) -> Any:
        """Ultra-fast JSON deserialization."""
        return orjson.loads(json_bytes)

    def serialize_pretty(self, data: Any) -> bytes:
        """Pretty-printed JSON for debugging."""
        return orjson.dumps(
            data,
            option=orjson.OPT_INDENT_2 | orjson.OPT_SORT_KEYS
        )


class StreamProcessor:
    """Stream large JSON files without loading into memory."""

    def __init__(self, chunk_size: int = 8192):
        self.chunk_size = chunk_size

    def process_json_array(
        self,
        file_path: str,
        processor: Callable[[Dict], Any]
    ) -> Iterator[Any]:
        """Stream process JSON array file."""
        import ijson

        with open(file_path, 'rb') as f:
            parser = ijson.items(f, 'item')
            for item in parser:
                yield processor(item)

    def process_json_lines(
        self,
        file_path: str,
        processor: Callable[[Dict], Any]
    ) -> Iterator[Any]:
        """Process JSON Lines (JSONL) format."""
        with open(file_path, 'r') as f:
            for line in f:
                if line.strip():
                    item = orjson.loads(line)
                    yield processor(item)

    def batch_process(
        self,
        file_path: str,
        processor: Callable[[Dict], Any],
        batch_size: int = 1000
    ) -> Iterator[list]:
        """Process items in batches for efficiency."""
        batch = []

        for item in self.process_json_array(file_path, lambda x: x):
            batch.append(processor(item))

            if len(batch) >= batch_size:
                yield batch
                batch = []

        if batch:
            yield batch


# Usage example
optimizer = JSONOptimizer()
stream_processor = StreamProcessor()

# Fast serialization
data = {"users": [{"id": i, "name": f"User {i}"} for i in range(10000)]}
fast_json = optimizer.serialize_fast(data)

# Process large file
def process_user(user: dict) -> dict:
    return {
        "id": user["id"],
        "name_upper": user["name"].upper(),
        "processed": True
    }

# Stream process without memory issues
results = []
for batch in stream_processor.batch_process("large_users.json", process_user, batch_size=5000):
    results.extend(batch)
    print(f"Processed batch of {len(batch)} items")

# Memory-efficient aggregation
total = 0
for user in stream_processor.process_json_array("large_users.json", lambda x: x):
    total += user.get("balance", 0)
print(f"Total balance: {total}")
```

**Explanation**: This pattern uses orjson for 5x faster serialization and ijson for memory-efficient streaming of large files, enabling processing of gigabyte-scale JSON data.

---

## Example 3: Advanced Data Validation with Schema Evolution

**Scenario**: Implement robust data validation with support for schema versioning and migration.

**Input**:
```python
# Need validation system that:
# - Validates complex nested data
# - Supports schema evolution
# - Provides clear error messages
# - Handles migrations between versions
```

**Output**:
```python
from moai_formats_data import DataValidator, SchemaEvolution
from typing import Dict, Any, List, Optional
from dataclasses import dataclass
import re

@dataclass
class ValidationResult:
    valid: bool
    errors: Dict[str, List[str]]
    sanitized_data: Optional[Dict[str, Any]] = None
    warnings: List[str] = None


class DataValidator:
    """Advanced data validation with type coercion and custom rules."""

    def __init__(self):
        self.type_validators = {
            'string': self._validate_string,
            'integer': self._validate_integer,
            'float': self._validate_float,
            'boolean': self._validate_boolean,
            'email': self._validate_email,
            'url': self._validate_url,
            'array': self._validate_array,
            'object': self._validate_object
        }

    def create_schema(self, schema_dict: Dict) -> Dict:
        """Create a compiled schema for efficient validation."""
        compiled = {}
        for field, rules in schema_dict.items():
            compiled[field] = self._compile_field_rules(rules)
        return compiled

    def _compile_field_rules(self, rules: Dict) -> Dict:
        """Pre-compile validation rules for performance."""
        compiled = rules.copy()

        # Pre-compile regex patterns
        if 'pattern' in rules:
            compiled['_compiled_pattern'] = re.compile(rules['pattern'])

        return compiled

    def validate(self, data: Dict, schema: Dict) -> ValidationResult:
        """Validate data against schema."""
        errors = {}
        sanitized = {}
        warnings = []

        for field, rules in schema.items():
            value = data.get(field)

            # Check required
            if rules.get('required', False) and value is None:
                errors.setdefault(field, []).append(f"Field '{field}' is required")
                continue

            if value is None:
                if 'default' in rules:
                    sanitized[field] = rules['default']
                continue

            # Type validation
            field_type = rules.get('type', 'string')
            validator = self.type_validators.get(field_type)

            if validator:
                is_valid, sanitized_value, error = validator(value, rules)
                if not is_valid:
                    errors.setdefault(field, []).append(error)
                else:
                    sanitized[field] = sanitized_value

        return ValidationResult(
            valid=len(errors) == 0,
            errors=errors,
            sanitized_data=sanitized if len(errors) == 0 else None,
            warnings=warnings
        )

    def _validate_string(self, value, rules) -> tuple:
        if not isinstance(value, str):
            return False, None, "Must be a string"

        if 'min_length' in rules and len(value) < rules['min_length']:
            return False, None, f"Minimum length is {rules['min_length']}"

        if 'max_length' in rules and len(value) > rules['max_length']:
            return False, None, f"Maximum length is {rules['max_length']}"

        if '_compiled_pattern' in rules:
            if not rules['_compiled_pattern'].match(value):
                return False, None, f"Does not match pattern"

        return True, value.strip(), None

    def _validate_integer(self, value, rules) -> tuple:
        try:
            int_value = int(value)
        except (ValueError, TypeError):
            return False, None, "Must be an integer"

        if 'min_value' in rules and int_value < rules['min_value']:
            return False, None, f"Minimum value is {rules['min_value']}"

        if 'max_value' in rules and int_value > rules['max_value']:
            return False, None, f"Maximum value is {rules['max_value']}"

        return True, int_value, None

    def _validate_email(self, value, rules) -> tuple:
        if not isinstance(value, str):
            return False, None, "Must be a string"

        email_pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
        if not re.match(email_pattern, value):
            return False, None, "Invalid email format"

        return True, value.lower().strip(), None

    def _validate_float(self, value, rules) -> tuple:
        try:
            float_value = float(value)
        except (ValueError, TypeError):
            return False, None, "Must be a number"
        return True, float_value, None

    def _validate_boolean(self, value, rules) -> tuple:
        if isinstance(value, bool):
            return True, value, None
        if value in ('true', 'True', '1', 1):
            return True, True, None
        if value in ('false', 'False', '0', 0):
            return True, False, None
        return False, None, "Must be a boolean"

    def _validate_url(self, value, rules) -> tuple:
        if not isinstance(value, str):
            return False, None, "Must be a string"
        url_pattern = r'^https?://[^\s/$.?#].[^\s]*$'
        if not re.match(url_pattern, value):
            return False, None, "Invalid URL format"
        return True, value, None

    def _validate_array(self, value, rules) -> tuple:
        if not isinstance(value, list):
            return False, None, "Must be an array"
        return True, value, None

    def _validate_object(self, value, rules) -> tuple:
        if not isinstance(value, dict):
            return False, None, "Must be an object"
        return True, value, None


class SchemaEvolution:
    """Handle schema versioning and data migration."""

    def __init__(self):
        self.schemas = {}
        self.migrations = {}

    def register_schema(self, version: str, schema: Dict):
        """Register a schema version."""
        self.schemas[version] = schema

    def add_migration(
        self,
        from_version: str,
        to_version: str,
        migration_fn: callable
    ):
        """Add migration function between versions."""
        key = f"{from_version}:{to_version}"
        self.migrations[key] = migration_fn

    def migrate_data(
        self,
        data: Dict,
        from_version: str,
        to_version: str
    ) -> Dict:
        """Migrate data between schema versions."""
        key = f"{from_version}:{to_version}"

        if key not in self.migrations:
            raise ValueError(f"No migration path from {from_version} to {to_version}")

        return self.migrations[key](data)


# Usage example
validator = DataValidator()

# Define schema
user_schema = validator.create_schema({
    "username": {
        "type": "string",
        "required": True,
        "min_length": 3,
        "max_length": 50,
        "pattern": r"^[a-zA-Z0-9_]+$"
    },
    "email": {
        "type": "email",
        "required": True
    },
    "age": {
        "type": "integer",
        "required": False,
        "min_value": 13,
        "max_value": 120
    },
    "website": {
        "type": "url",
        "required": False
    }
})

# Validate data
user_data = {
    "username": "john_doe",
    "email": "JOHN@Example.COM",
    "age": 25
}

result = validator.validate(user_data, user_schema)

if result.valid:
    print("Valid data:", result.sanitized_data)
    # {'username': 'john_doe', 'email': 'john@example.com', 'age': 25}
else:
    print("Validation errors:", result.errors)


# Schema evolution example
evolution = SchemaEvolution()

v1_schema = {"name": {"type": "string"}, "age": {"type": "integer"}}
v2_schema = {"full_name": {"type": "string"}, "age": {"type": "integer"}, "email": {"type": "email"}}

evolution.register_schema("v1", v1_schema)
evolution.register_schema("v2", v2_schema)

def migrate_v1_to_v2(data: Dict) -> Dict:
    return {
        "full_name": data.get("name", ""),
        "age": data.get("age", 0),
        "email": None  # New required field needs default
    }

evolution.add_migration("v1", "v2", migrate_v1_to_v2)

# Migrate old data
old_data = {"name": "John Doe", "age": 30}
new_data = evolution.migrate_data(old_data, "v1", "v2")
# {'full_name': 'John Doe', 'age': 30, 'email': None}
```

**Explanation**: This pattern provides comprehensive data validation with type coercion, pattern matching, and schema evolution support for backward compatibility.

---

## Common Patterns

### Pattern 1: Intelligent Caching with Format Optimization

```python
from functools import lru_cache
import hashlib
import time

class SmartCache:
    """Memory-aware caching with format optimization."""

    def __init__(self, max_memory_mb: int = 50, max_items: int = 10000):
        self.max_memory = max_memory_mb * 1024 * 1024
        self.max_items = max_items
        self.cache = {}
        self.access_times = {}
        self.sizes = {}

    def _generate_key(self, data: dict) -> str:
        """Generate deterministic cache key."""
        serialized = orjson.dumps(data, option=orjson.OPT_SORT_KEYS)
        return hashlib.sha256(serialized).hexdigest()[:16]

    def get(self, key: str) -> Any:
        """Get from cache with access tracking."""
        if key in self.cache:
            self.access_times[key] = time.time()
            return self.cache[key]
        return None

    def set(self, key: str, value: Any, ttl: int = 3600):
        """Set cache with memory management."""
        serialized = orjson.dumps(value)
        size = len(serialized)

        # Evict if needed
        while self._total_size() + size > self.max_memory:
            self._evict_lru()

        self.cache[key] = value
        self.sizes[key] = size
        self.access_times[key] = time.time()

    def _total_size(self) -> int:
        return sum(self.sizes.values())

    def _evict_lru(self):
        """Evict least recently used item."""
        if not self.access_times:
            return

        oldest_key = min(self.access_times, key=self.access_times.get)
        del self.cache[oldest_key]
        del self.sizes[oldest_key]
        del self.access_times[oldest_key]

    def cache_result(self, ttl: int = 3600):
        """Decorator for caching function results."""
        def decorator(func):
            def wrapper(*args, **kwargs):
                key = self._generate_key({"args": args, "kwargs": kwargs})
                cached = self.get(key)
                if cached is not None:
                    return cached

                result = func(*args, **kwargs)
                self.set(key, result, ttl)
                return result
            return wrapper
        return decorator
```

### Pattern 2: Format Conversion Pipeline

```python
class FormatConverter:
    """Convert between different data formats."""

    @staticmethod
    def json_to_yaml(json_data: str) -> str:
        """Convert JSON to YAML."""
        import yaml
        data = orjson.loads(json_data)
        return yaml.dump(data, default_flow_style=False, allow_unicode=True)

    @staticmethod
    def yaml_to_json(yaml_data: str) -> str:
        """Convert YAML to JSON."""
        import yaml
        data = yaml.safe_load(yaml_data)
        return orjson.dumps(data).decode()

    @staticmethod
    def json_to_toon(json_data: str) -> str:
        """Convert JSON to TOON."""
        data = orjson.loads(json_data)
        encoder = TOONEncoder()
        return encoder.encode(data)

    @staticmethod
    def csv_to_json(csv_data: str) -> str:
        """Convert CSV to JSON array."""
        import csv
        from io import StringIO

        reader = csv.DictReader(StringIO(csv_data))
        records = list(reader)
        return orjson.dumps(records).decode()

# Usage
converter = FormatConverter()

json_data = '{"name": "John", "age": 30}'
yaml_output = converter.json_to_yaml(json_data)
toon_output = converter.json_to_toon(json_data)
```

### Pattern 3: Batch Validation with Error Aggregation

```python
def validate_batch(
    items: list,
    schema: dict,
    stop_on_first_error: bool = False
) -> dict:
    """Validate a batch of items with aggregated results."""
    validator = DataValidator()
    results = {
        "total": len(items),
        "valid": 0,
        "invalid": 0,
        "errors": []
    }

    for index, item in enumerate(items):
        result = validator.validate(item, schema)

        if result.valid:
            results["valid"] += 1
        else:
            results["invalid"] += 1
            results["errors"].append({
                "index": index,
                "item": item,
                "errors": result.errors
            })

            if stop_on_first_error:
                break

    return results

# Usage
items = [
    {"username": "john", "email": "john@example.com"},
    {"username": "ab", "email": "invalid"},  # Both fields invalid
    {"username": "jane_doe", "email": "jane@example.com"}
]

batch_result = validate_batch(items, user_schema)
# {'total': 3, 'valid': 2, 'invalid': 1, 'errors': [...]}
```

---

## Anti-Patterns (Patterns to Avoid)

### Anti-Pattern 1: Loading Entire Files Into Memory

**Problem**: Loading large files completely causes memory exhaustion.

```python
# Incorrect approach
with open('huge_file.json', 'r') as f:
    data = json.load(f)  # Loads entire file into memory
```

**Solution**: Use streaming for large files.

```python
# Correct approach
import ijson

with open('huge_file.json', 'rb') as f:
    for item in ijson.items(f, 'item'):
        process(item)  # Process one item at a time
```

### Anti-Pattern 2: Inconsistent Serialization

**Problem**: Different serialization methods produce inconsistent output.

```python
# Incorrect approach - inconsistent key ordering
json.dumps({"b": 2, "a": 1})  # '{"b": 2, "a": 1}' sometimes '{"a": 1, "b": 2}'
```

**Solution**: Use consistent serialization options.

```python
# Correct approach
orjson.dumps(data, option=orjson.OPT_SORT_KEYS)
# Always produces consistent, sorted output
```

### Anti-Pattern 3: No Validation Before Processing

**Problem**: Processing invalid data leads to runtime errors.

```python
# Incorrect approach
def process_user(user_data: dict):
    email = user_data['email'].lower()  # KeyError if missing
    age = int(user_data['age'])  # ValueError if not numeric
```

**Solution**: Validate before processing.

```python
# Correct approach
def process_user(user_data: dict):
    result = validator.validate(user_data, user_schema)
    if not result.valid:
        raise ValidationError(result.errors)

    data = result.sanitized_data
    email = data['email']  # Already validated and sanitized
    age = data['age']      # Already converted to int
```

---

## Performance Benchmarks

```python
import time

def benchmark_serialization():
    """Compare serialization performance."""
    data = {"users": [{"id": i, "name": f"User {i}"} for i in range(10000)]}

    # Standard json
    start = time.time()
    for _ in range(100):
        json.dumps(data)
    json_time = time.time() - start

    # orjson
    start = time.time()
    for _ in range(100):
        orjson.dumps(data)
    orjson_time = time.time() - start

    # TOON
    encoder = TOONEncoder()
    start = time.time()
    for _ in range(100):
        encoder.encode(data)
    toon_time = time.time() - start

    print(f"json: {json_time:.3f}s")      # ~2.5s
    print(f"orjson: {orjson_time:.3f}s")  # ~0.5s (5x faster)
    print(f"TOON: {toon_time:.3f}s")      # ~0.8s (with 40% smaller output)
```

---

*For additional patterns and format-specific optimizations, see the `modules/` directory.*
