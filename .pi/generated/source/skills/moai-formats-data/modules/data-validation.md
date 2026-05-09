# Data Validation and Schema Management

> Module: Comprehensive data validation system
> Complexity: Advanced
> Time: 25+ minutes
> Dependencies: typing, dataclasses, enum, re, datetime, jsonschema (optional)

## Advanced Validation System

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

## Validation Patterns and Examples

### Common Validation Schemas

```python
class CommonSchemas:
 """Pre-defined validation schemas for common use cases."""

 @staticmethod
 def user_schema() -> Dict[str, ValidationRule]:
 """User data validation schema."""
 validator = DataValidator()
 return validator.create_schema({
 "id": {"type": "integer", "required": True, "min_value": 1},
 "username": {"type": "string", "required": True, "min_length": 3, "max_length": 50},
 "email": {"type": "email", "required": True},
 "age": {"type": "integer", "required": False, "min_value": 13, "max_value": 120},
 "active": {"type": "boolean", "required": False},
 "preferences": {"type": "object", "required": False},
 "tags": {"type": "array", "required": False}
 })

 @staticmethod
 def api_response_schema() -> Dict[str, ValidationRule]:
 """API response validation schema."""
 validator = DataValidator()
 return validator.create_schema({
 "status": {"type": "string", "required": True, "allowed_values": ["success", "error"]},
 "data": {"type": "object", "required": False},
 "error": {"type": "string", "required": False},
 "timestamp": {"type": "datetime", "required": True},
 "request_id": {"type": "uuid", "required": True}
 })

 @staticmethod
 def config_schema() -> Dict[str, ValidationRule]:
 """Configuration validation schema."""
 validator = DataValidator()
 return validator.create_schema({
 "database": {
 "type": "object",
 "required": True,
 "custom_validator": lambda v: isinstance(v, dict) and "url" in v
 },
 "api_keys": {"type": "object", "required": False},
 "features": {
 "type": "object",
 "required": False,
 "custom_validator": lambda v: all(isinstance(v[k], bool) for k in v)
 },
 "timeouts": {
 "type": "object",
 "required": False,
 "custom_validator": lambda v: all(isinstance(v[k], (int, float)) for k in v)
 }
 })

# Usage examples
def example_validations():
 validator = DataValidator()

 # User data validation
 user_schema = CommonSchemas.user_schema()
 user_data = {
 "id": 123,
 "username": "john_doe",
 "email": "john@example.com",
 "age": 30,
 "active": True
 }

 result = validator.validate(user_data, user_schema)
 print("User validation:", result)

 # Custom validation with regex
 phone_schema = validator.create_schema({
 "phone": {
 "type": "string",
 "required": True,
 "pattern": r'^\+?1?-?\.?\s?\(?(\d{3})\)?[-.\s]?(\d{3})[-.\s]?(\d{4})$'
 }
 })

 # Complex nested validation
 nested_schema = validator.create_schema({
 "user": {"type": "object", "required": True},
 "metadata": {
 "type": "object",
 "required": False,
 "custom_validator": lambda v: isinstance(v, dict) and len(v) <= 10
 },
 "items": {
 "type": "array",
 "required": True,
 "custom_validator": lambda v: len(v) >= 1
 }
 })
```

### Advanced Validation Techniques

```python
class AdvancedValidator:
 """Advanced validation with complex rules and cross-field validation."""

 def __init__(self):
 self.base_validator = DataValidator()

 def validate_with_context(self, data: Dict, schema: Dict, context: Dict = None) -> Dict:
 """Validate data with additional context information."""
 result = self.base_validator.validate(data, schema)

 # Add context-aware validations
 if context:
 result.update(self._context_validation(data, context))

 return result

 def _context_validation(self, data: Dict, context: Dict) -> Dict:
 """Perform context-aware validations."""
 context_errors = {}

 # Example: Validate that user has permission for requested action
 if 'user_role' in context and 'requested_action' in data:
 user_role = context['user_role']
 action = data['requested_action']

 if not self._has_permission(user_role, action):
 context_errors['permission'] = f"User role '{user_role}' cannot perform action '{action}'"

 # Example: Validate business rules
 if 'business_hours' in context and 'timestamp' in data:
 timestamp = data['timestamp']
 business_hours = context['business_hours']

 if not self._is_within_business_hours(timestamp, business_hours):
 context_errors['business_hours'] = "Action must be performed during business hours"

 return {'context_errors': context_errors}

 def validate_dependencies(self, data: Dict, dependencies: Dict) -> List[str]:
 """Validate field dependencies (e.g., if field A exists, field B must exist)."""
 errors = []

 for field, required_fields in dependencies.items():
 if field in data:
 for required_field in required_fields:
 if required_field not in data:
 errors.append(f"Field '{required_field}' is required when '{field}' is present")

 return errors

 def validate_conditional_requirements(self, data: Dict, conditions: Dict) -> List[str]:
 """Validate conditional requirements based on field values."""
 errors = []

 for field, condition_rules in conditions.items():
 if field in data:
 field_value = data[field]
 for condition, required_fields in condition_rules.items():
 if self._evaluate_condition(field_value, condition):
 for required_field in required_fields:
 if required_field not in data:
 errors.append(f"Field '{required_field}' is required when '{field}' {condition}")

 return errors

 def _has_permission(self, role: str, action: str) -> bool:
 """Check if user role has permission for action."""
 permissions = {
 'admin': ['read', 'write', 'delete', 'admin'],
 'editor': ['read', 'write'],
 'viewer': ['read']
 }
 return action in permissions.get(role, [])

 def _is_within_business_hours(self, timestamp: datetime, business_hours: Dict) -> bool:
 """Check if timestamp is within business hours."""
 if not isinstance(timestamp, datetime):
 return True # Skip validation if not a datetime

 weekday = timestamp.weekday() # 0 = Monday, 6 = Sunday
 hour = timestamp.hour

 if weekday >= 5: # Weekend
 return business_hours.get('weekend_enabled', False)

 start_hour = business_hours.get('start_hour', 9)
 end_hour = business_hours.get('end_hour', 17)

 return start_hour <= hour < end_hour

 def _evaluate_condition(self, value: Any, condition: str) -> bool:
 """Evaluate a condition string against a value."""
 # Simple implementation - could be enhanced with more complex condition parsing
 if condition.startswith('>'):
 try:
 return value > int(condition[1:])
 except ValueError:
 return False
 elif condition.startswith('<'):
 try:
 return value < int(condition[1:])
 except ValueError:
 return False
 elif condition.startswith('=='):
 return value == condition[2:]
 else:
 # Default: check if value equals condition string
 return str(value) == condition
```

### Performance Optimization for Validation

```python
class OptimizedValidator:
 """Performance-optimized validation with caching and batching."""

 def __init__(self):
 self.base_validator = DataValidator()
 self.schema_cache = {}
 self.pattern_cache = {}

 def validate_batch(self, data_list: List[Dict], schema: Dict) -> List[Dict]:
 """Validate multiple data items efficiently."""
 results = []

 # Pre-compile patterns for better performance
 self._compile_schema_patterns(schema)

 for data in data_list:
 result = self.base_validator.validate(data, schema)
 results.append(result)

 return results

 def _compile_schema_patterns(self, schema: Dict):
 """Pre-compile regex patterns in schema."""
 for field_name, rule in schema.items():
 if rule.pattern and rule.pattern not in self.pattern_cache:
 self.pattern_cache[rule.pattern] = re.compile(rule.pattern)

 def get_cached_schema(self, schema_key: str, schema_factory: callable) -> Dict:
 """Get schema from cache or create and cache it."""
 if schema_key not in self.schema_cache:
 self.schema_cache[schema_key] = schema_factory()
 return self.schema_cache[schema_key]
```

---

Module: `modules/data-validation.md`
Related: [TOON Encoding](./toon-encoding.md) | [JSON Optimization](./json-optimization.md)
