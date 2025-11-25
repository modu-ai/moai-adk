# Advanced Patterns - Configuration Schema Management

**Enterprise configuration validation, schema design, and best practices**

---

## JSON Schema Validation Patterns

### Basic Schema Definition

```python
CONFIG_SCHEMA = {
    "type": "object",
    "properties": {
        "api": {
            "type": "object",
            "properties": {
                "host": {"type": "string"},
                "port": {"type": "integer", "minimum": 1, "maximum": 65535},
                "timeout": {"type": "integer", "default": 30}
            },
            "required": ["host", "port"]
        },
        "database": {
            "type": "object",
            "properties": {
                "connection_string": {"type": "string"},
                "pool_size": {"type": "integer", "minimum": 1}
            }
        }
    },
    "required": ["api"]
}
```

### Validation with jsonschema

```python
import jsonschema

def validate_config(config: dict) -> bool:
    try:
        jsonschema.validate(config, CONFIG_SCHEMA)
        return True
    except jsonschema.ValidationError as e:
        logger.error(f"Config validation failed: {e.message}")
        return False
```

---

## Schema Design Patterns

### Hierarchical Configuration

```yaml
app:
  name: MyApp
  version: 1.0.0
  
  server:
    host: localhost
    port: 8000
  
  database:
    provider: postgresql
    connection:
      host: localhost
      port: 5432
  
  logging:
    level: INFO
    handlers:
      - console
      - file
```

### Environment-Specific Schemas

```python
SCHEMAS = {
    'development': {
        'debug': True,
        'database_url': 'sqlite:///dev.db'
    },
    'production': {
        'debug': False,
        'database_url': 'postgresql://prod_server/db'
    }
}
```

### Configuration Inheritance

```python
BASE_CONFIG = {
    'timeout': 30,
    'retries': 3,
    'log_level': 'INFO'
}

PRODUCTION_CONFIG = {
    **BASE_CONFIG,
    'timeout': 60,
    'log_level': 'ERROR'
}
```

---

## Type-Safe Configuration

### Pydantic Models

```python
from pydantic import BaseModel, Field

class APIConfig(BaseModel):
    host: str = "localhost"
    port: int = Field(8000, ge=1, le=65535)
    timeout: int = 30

class DatabaseConfig(BaseModel):
    connection_string: str
    pool_size: int = 10

class AppConfig(BaseModel):
    api: APIConfig
    database: DatabaseConfig
    debug: bool = False

# Usage
config = AppConfig(
    api=APIConfig(host="0.0.0.0", port=8080),
    database=DatabaseConfig(connection_string="postgres://...")
)
```

---

## Configuration Validation

### Multi-Stage Validation

```python
class ConfigValidator:
    def validate(self, config: dict) -> ValidationResult:
        # Stage 1: Schema validation
        schema_valid = self._validate_schema(config)
        if not schema_valid:
            return ValidationResult(passed=False, stage="schema")
        
        # Stage 2: Business logic validation
        logic_valid = self._validate_logic(config)
        if not logic_valid:
            return ValidationResult(passed=False, stage="logic")
        
        # Stage 3: Dependencies validation
        deps_valid = self._validate_dependencies(config)
        if not deps_valid:
            return ValidationResult(passed=False, stage="dependencies")
        
        return ValidationResult(passed=True)
```

### Custom Validators

```python
def validate_port_not_in_use(port: int) -> bool:
    """Check if port is available."""
    import socket
    sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    result = sock.connect_ex(('localhost', port))
    sock.close()
    return result != 0  # Port is available if connection fails

def validate_database_connection(connection_string: str) -> bool:
    """Test database connection."""
    try:
        conn = create_connection(connection_string)
        conn.close()
        return True
    except Exception:
        return False
```

---

## Configuration Loading Strategies

### Priority-Based Loading

```python
def load_config() -> dict:
    """Load config with priority: env vars > config file > defaults."""
    
    config = DEFAULT_CONFIG.copy()
    
    # Override with file config
    if os.path.exists('config.json'):
        config.update(load_json_file('config.json'))
    
    # Override with environment variables
    for key in config:
        env_value = os.getenv(key.upper())
        if env_value:
            config[key] = env_value
    
    return config
```

### Configuration Merging

```python
def merge_configs(*configs):
    """Deep merge multiple configuration dicts."""
    result = {}
    for config in configs:
        for key, value in config.items():
            if isinstance(value, dict) and key in result:
                result[key] = merge_configs(result[key], value)
            else:
                result[key] = value
    return result
```

---

## Error Handling & Defaults

### Safe Configuration Access

```python
class SafeConfig:
    def __init__(self, config: dict):
        self.config = config
    
    def get(self, key: str, default=None, type_cast=None):
        """Safely get config value with optional type casting."""
        value = self.config.get(key, default)
        if type_cast and value is not None:
            try:
                return type_cast(value)
            except (ValueError, TypeError):
                logger.warning(f"Failed to cast {key} to {type_cast}")
                return default
        return value
```

---

## Context7 Integration

### Schema Best Practices

```python
async def research_config_patterns():
    """Get latest configuration patterns from Context7."""
    
    docs = await context7.get_library_docs(
        context7_library_id="/pydantic/pydantic",
        topic="configuration schema validation best practices 2025",
        tokens=3000
    )
    
    return extract_patterns(docs)
```

---

**Last Updated**: 2025-11-22
**Status**: Production Ready
