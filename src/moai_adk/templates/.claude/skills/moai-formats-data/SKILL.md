---
name: moai-formats-data
description: Data format specialist covering TOON encoding, JSON/YAML optimization, serialization patterns, and data validation for modern applications
version: 1.0.0
category: library
tags:
  - formats
  - data
  - toon
  - serialization
  - validation
  - optimization
updated: 2025-11-30
status: active
author: MoAI-ADK Team
---

# Data Format Specialist

## Quick Reference (30 seconds)

**Advanced Data Format Management** - Comprehensive data handling covering TOON encoding, JSON/YAML optimization, serialization patterns, and data validation for performance-critical applications.

**Core Capabilities**:
- 🗜️ **TOON Encoding**: 40-60% token reduction vs JSON for LLM communication
- ⚡ **JSON/YAML Optimization**: Efficient serialization and parsing patterns
- ✅ **Data Validation**: Schema validation, type checking, error handling
- 🔄 **Format Conversion**: Seamless transformation between data formats
- 📊 **Performance**: Optimized data structures and caching strategies
- 🎯 **Schema Management**: Dynamic schema generation and evolution

**When to Use**:
- Optimizing data transmission to LLMs within token budgets
- High-performance serialization/deserialization
- Schema validation and data integrity
- Format conversion and data transformation
- Large dataset processing and optimization

---

## Implementation Guide

### TOON (Token-Optimized Object Notation) Encoding

**Core TOON Implementation**:
```python
from typing import Dict, List, Any, Union
import json
from datetime import datetime

class TOONEncoder:
    """Token-Optimized Object Notation encoder for efficient LLM communication."""

    def __init__(self):
        self.type_markers = {
            'string': '',
            'number': '#',
            'boolean': '!',
            'null': '~',
            'timestamp': '@'
        }

    def encode(self, data: Any) -> str:
        """Encode Python data structure to TOON format."""
        return self._encode_value(data)

    def _encode_value(self, value: Any) -> str:
        """Encode individual values with type optimization."""

        if value is None:
            return '~'

        elif isinstance(value, bool):
            return f'!{str(value)[0]}'  # !t or !f

        elif isinstance(value, (int, float)):
            return f'#{value}'

        elif isinstance(value, str):
            return self._escape_string(value)

        elif isinstance(value, datetime):
            return f'@{value.isoformat()}'

        elif isinstance(value, dict):
            if not value:
                return '{}'
            items = []
            for k, v in value.items():
                encoded_key = self._escape_string(str(k))
                encoded_value = self._encode_value(v)
                items.append(f'{encoded_key}:{encoded_value}')
            return '{' + ','.join(items) + '}'

        elif isinstance(value, list):
            if not value:
                return '[]'
            encoded_items = [self._encode_value(item) for item in value]
            return '[' + '|'.join(encoded_items) + ']'

        else:
            # Fallback to JSON for complex objects
            json_str = json.dumps(value, default=str)
            return f'${json_str}'

    def _escape_string(self, s: str) -> str:
        """Escape special characters in strings."""
        # Replace problematic characters
        s = s.replace('\\', '\\\\')
        s = s.replace(':', '\\:')
        s = s.replace('|', '\\|')
        s = s.replace(',', '\\,')
        s = s.replace('{', '\\{')
        s = s.replace('}', '\\}')
        s = s.replace('[', '\\[')
        s = s.replace(']', '\\]')
        s = s.replace('~', '\\~')
        s = s.replace('#', '\\#')
        s = s.replace('!', '\\!')
        s = s.replace('@', '\\@')
        s = s.replace('$', '\\$')

        return s

    def decode(self, toon_str: str) -> Any:
        """Decode TOON format back to Python data structure."""
        return self._parse_value(toon_str.strip())

    def _parse_value(self, s: str) -> Any:
        """Parse TOON value back to Python type."""
        s = s.strip()

        if not s:
            return None

        # Null value
        if s == '~':
            return None

        # Boolean values
        if s.startswith('!'):
            return s[1:] == 't'

        # Numbers
        if s.startswith('#'):
            num_str = s[1:]
            if '.' in num_str:
                return float(num_str)
            return int(num_str)

        # Timestamps
        if s.startswith('@'):
            try:
                return datetime.fromisoformat(s[1:])
            except ValueError:
                return s[1:]  # Return as string if parsing fails

        # JSON fallback
        if s.startswith('$'):
            return json.loads(s[1:])

        # Arrays
        if s.startswith('[') and s.endswith(']'):
            content = s[1:-1]
            if not content:
                return []
            items = self._split_array_items(content)
            return [self._parse_value(item) for item in items]

        # Objects
        if s.startswith('{') and s.endswith('}'):
            content = s[1:-1]
            if not content:
                return {}
            pairs = self._split_object_pairs(content)
            result = {}
            for pair in pairs:
                if ':' in pair:
                    key, value = pair.split(':', 1)
                    result[self._unescape_string(key)] = self._parse_value(value)
            return result

        # String (default)
        return self._unescape_string(s)

    def _split_array_items(self, content: str) -> List[str]:
        """Split array items handling escaped separators."""
        items = []
        current = []
        escape = False

        for char in content:
            if escape:
                current.append(char)
                escape = False
            elif char == '\\':
                escape = True
            elif char == '|':
                items.append(''.join(current))
                current = []
            else:
                current.append(char)

        if current:
            items.append(''.join(current))

        return items

    def _split_object_pairs(self, content: str) -> List[str]:
        """Split object pairs handling escaped separators."""
        pairs = []
        current = []
        escape = False
        depth = 0

        for char in content:
            if escape:
                current.append(char)
                escape = False
            elif char == '\\':
                escape = True
            elif char in '{[':
                depth += 1
                current.append(char)
            elif char in '}]':
                depth -= 1
                current.append(char)
            elif char == ',' and depth == 0:
                pairs.append(''.join(current))
                current = []
            else:
                current.append(char)

        if current:
            pairs.append(''.join(current))

        return pairs

    def _unescape_string(self, s: str) -> str:
        """Unescape escaped characters in strings."""
        return s.replace('\\:', ':').replace('\\|', '|').replace('\\,', ',') \
                 .replace('\\{', '{').replace('\\}', '}').replace('\\[', '[') \
                 .replace('\\]', ']').replace('\\~', '~').replace('\\#', '#') \
                 .replace('\\!', '!').replace('\\@', '@').replace('\\$', '$') \
                 .replace('\\\\', '\\')

# Usage example and performance comparison
def demonstrate_toon_optimization():
    data = {
        "user": {
            "id": 12345,
            "name": "John Doe",
            "email": "john.doe@example.com",
            "active": True,
            "created_at": datetime.now()
        },
        "permissions": ["read", "write", "admin"],
        "settings": {
            "theme": "dark",
            "notifications": True
        }
    }

    encoder = TOONEncoder()

    # JSON encoding
    json_str = json.dumps(data, default=str)
    json_tokens = len(json_str.split())

    # TOON encoding
    toon_str = encoder.encode(data)
    toon_tokens = len(toon_str.split())

    # Performance comparison
    reduction = (1 - toon_tokens / json_tokens) * 100

    return {
        "json_size": len(json_str),
        "toon_size": len(toon_str),
        "json_tokens": json_tokens,
        "toon_tokens": toon_tokens,
        "token_reduction": reduction,
        "json_str": json_str,
        "toon_str": toon_str
    }
```

### Advanced JSON/YAML Optimization

**High-Performance JSON Handling**:
```python
import orjson  # Ultra-fast JSON library
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
            if key == '$schema':  # Remove schema URL
                continue
            elif key == 'description' and len(value) > 100:  # Truncate long descriptions
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
        self.yaml_loader = yaml.CSafeLoader  # Use C loader for performance
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

### Data Validation and Schema Management

**Advanced Validation System**:
```python
from typing import Type, Dict, Any, List, Optional, Union
from dataclasses import dataclass
from enum import Enum
import re
from datetime import datetime

class ValidationType(Enum):
    STRING = "string"
    INTEGER = "integer"
    FLOAT = "float"
    BOOLEAN = "boolean"
    ARRAY = "array"
    OBJECT = "object"
    DATETIME = "datetime"
    EMAIL = "email"
    URL = "url"
    UUID = "uuid"

@dataclass
class ValidationRule:
    """Individual validation rule configuration."""
    type: ValidationType
    required: bool = True
    min_length: Optional[int] = None
    max_length: Optional[int] = None
    min_value: Optional[Union[int, float]] = None
    max_value: Optional[Union[int, float]] = None
    pattern: Optional[str] = None
    allowed_values: Optional[List[Any]] = None
    custom_validator: Optional[callable] = None

class DataValidator:
    """Comprehensive data validation system."""

    def __init__(self):
        self.compiled_patterns = {}
        self.global_validators = {}

    def create_schema(self, field_definitions: Dict[str, Dict]) -> Dict[str, ValidationRule]:
        """Create validation schema from field definitions."""
        schema = {}

        for field_name, field_config in field_definitions.items():
            validation_type = ValidationType(field_config.get('type', 'string'))
            rule = ValidationRule(
                type=validation_type,
                required=field_config.get('required', True),
                min_length=field_config.get('min_length'),
                max_length=field_config.get('max_length'),
                min_value=field_config.get('min_value'),
                max_value=field_config.get('max_value'),
                pattern=field_config.get('pattern'),
                allowed_values=field_config.get('allowed_values'),
                custom_validator=field_config.get('custom_validator')
            )
            schema[field_name] = rule

        return schema

    def validate(self, data: Any, schema: Dict[str, ValidationRule]) -> Dict[str, Any]:
        """Validate data against schema and return results."""
        errors = {}
        warnings = {}
        sanitized_data = {}

        for field_name, rule in schema.items():
            value = data.get(field_name)

            # Check required fields
            if rule.required and value is None:
                errors[field_name] = f"Field '{field_name}' is required"
                continue

            if value is None:
                continue

            # Type validation
            if not self._validate_type(value, rule.type):
                errors[field_name] = f"Field '{field_name}' must be of type {rule.type.value}"
                continue

            # Length validation for strings
            if rule.type == ValidationType.STRING:
                if rule.min_length and len(value) < rule.min_length:
                    errors[field_name] = f"Field '{field_name}' must be at least {rule.min_length} characters"
                elif rule.max_length and len(value) > rule.max_length:
                    errors[field_name] = f"Field '{field_name}' must be at most {rule.max_length} characters"

            # Value range validation
            if rule.type in [ValidationType.INTEGER, ValidationType.FLOAT]:
                if rule.min_value is not None and value < rule.min_value:
                    errors[field_name] = f"Field '{field_name}' must be at least {rule.min_value}"
                elif rule.max_value is not None and value > rule.max_value:
                    errors[field_name] = f"Field '{field_name}' must be at most {rule.max_value}"

            # Pattern validation
            if rule.pattern:
                if not self._validate_pattern(value, rule.pattern):
                    errors[field_name] = f"Field '{field_name}' does not match required pattern"

            # Allowed values validation
            if rule.allowed_values and value not in rule.allowed_values:
                errors[field_name] = f"Field '{field_name}' must be one of {rule.allowed_values}"

            # Custom validation
            if rule.custom_validator:
                try:
                    custom_result = rule.custom_validator(value)
                    if custom_result is not True:
                        errors[field_name] = custom_result
                except Exception as e:
                    errors[field_name] = f"Custom validation failed: {str(e)}"

            # Sanitize and store valid data
            sanitized_data[field_name] = self._sanitize_value(value, rule.type)

        return {
            'valid': len(errors) == 0,
            'errors': errors,
            'warnings': warnings,
            'sanitized_data': sanitized_data
        }

    def _validate_type(self, value: Any, validation_type: ValidationType) -> bool:
        """Validate value type."""
        type_validators = {
            ValidationType.STRING: lambda v: isinstance(v, str),
            ValidationType.INTEGER: lambda v: isinstance(v, int) and not isinstance(v, bool),
            ValidationType.FLOAT: lambda v: isinstance(v, (int, float)) and not isinstance(v, bool),
            ValidationType.BOOLEAN: lambda v: isinstance(v, bool),
            ValidationType.ARRAY: lambda v: isinstance(v, (list, tuple)),
            ValidationType.OBJECT: lambda v: isinstance(v, dict),
            ValidationType.DATETIME: lambda v: isinstance(v, datetime) or self._is_iso_datetime(v),
            ValidationType.EMAIL: lambda v: isinstance(v, str) and self._is_email(v),
            ValidationType.URL: lambda v: isinstance(v, str) and self._is_url(v),
            ValidationType.UUID: lambda v: isinstance(v, str) and self._is_uuid(v)
        }

        return type_validators.get(validation_type, lambda v: False)(value)

    def _validate_pattern(self, value: str, pattern: str) -> bool:
        """Validate string against regex pattern."""
        if pattern not in self.compiled_patterns:
            self.compiled_patterns[pattern] = re.compile(pattern)

        return bool(self.compiled_patterns[pattern].match(value))

    def _is_iso_datetime(self, value: str) -> bool:
        """Check if string is valid ISO datetime."""
        try:
            datetime.fromisoformat(value.replace('Z', '+00:00'))
            return True
        except ValueError:
            return False

    def _is_email(self, value: str) -> bool:
        """Simple email validation."""
        pattern = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
        return bool(re.match(pattern, value))

    def _is_url(self, value: str) -> bool:
        """Simple URL validation."""
        pattern = r'^https?://[^\s/$.?#].[^\s]*$'
        return bool(re.match(pattern, value))

    def _is_uuid(self, value: str) -> bool:
        """UUID validation."""
        pattern = r'^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
        return bool(re.match(pattern, value.lower()))

    def _sanitize_value(self, value: Any, validation_type: ValidationType) -> Any:
        """Sanitize value based on type."""
        sanitizers = {
            ValidationType.STRING: lambda v: v.strip(),
            ValidationType.INTEGER: lambda v: int(v),
            ValidationType.FLOAT: lambda v: float(v),
            ValidationType.BOOLEAN: lambda v: bool(v),
            ValidationType.ARRAY: lambda v: list(v),
            ValidationType.DATETIME: lambda v: datetime.fromisoformat(v) if isinstance(v, str) else v,
        }

        return sanitizers.get(validation_type, lambda v: v)(value)

# Schema evolution manager
class SchemaEvolution:
    """Manage schema evolution and migration."""

    def __init__(self):
        self.version_history = {}
        self.migrations = {}

    def register_schema(self, version: str, schema: Dict):
        """Register schema version."""
        self.version_history[version] = {
            'schema': schema,
            'timestamp': datetime.now(),
            'version': version
        }

    def add_migration(self, from_version: str, to_version: str, migration_func: callable):
        """Add migration function between schema versions."""
        migration_key = f"{from_version}->{to_version}"
        self.migrations[migration_key] = migration_func

    def migrate_data(self, data: Dict, from_version: str, to_version: str) -> Dict:
        """Migrate data between schema versions."""
        current_data = data.copy()
        current_version = from_version

        while current_version != to_version:
            # Find next migration path
            migration_key = f"{current_version}->{to_version}"
            if migration_key not in self.migrations:
                raise ValueError(f"No migration path from {current_version} to {to_version}")

            migration_func = self.migrations[migration_key]
            current_data = migration_func(current_data)
            current_version = to_version

        return current_data
```

---

## Performance Optimization

### Caching and Memoization

**Intelligent Data Caching**:
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
            def wrapper(*args, **kwargs):
                cache_key = self._generate_key(func.__name__, args, kwargs)

                # Check cache
                if cache_key in self.cache:
                    cache_entry = self.cache[cache_key]
                    if time.time() - cache_entry['timestamp'] < (ttl or self.ttl):
                        return cache_entry['result']

                # Execute function and cache result
                result = func(*args, **kwargs)

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

---

## Works Well With

- **moai-domain-backend** - Backend data serialization and API responses
- **moai-domain-database** - Database data format optimization
- **moai-integration-mcp** - MCP data serialization and transmission
- **moai-docs-generation** - Documentation data formatting
- **moai-foundation-core** - Core data architecture principles

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

**Core Libraries**:
- **orjson**: Ultra-fast JSON parsing and serialization
- **PyYAML**: YAML processing with C-based loaders
- **ijson**: Streaming JSON parser for large files
- **python-dateutil**: Advanced datetime parsing
- **regex**: Advanced regular expression support

**Performance Tools**:
- **lru_cache**: Built-in memoization
- **pickle**: Object serialization
- **hashlib**: Hash generation for caching
- **functools**: Function decorators and utilities

**Validation Libraries**:
- **jsonschema**: JSON Schema validation
- **cerberus**: Lightweight data validation
- **marshmallow**: Object serialization/deserialization
- **pydantic**: Data validation using Python type hints

---

**Status**: Production Ready
**Last Updated**: 2025-11-30
**Maintained by**: MoAI-ADK Data Team