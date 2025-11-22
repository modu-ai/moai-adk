# Internal Communications - Practical Examples

## Example 1: Structured Logging

```python
# Structured logging for debugging
import json
import logging

def log_structured(level: str, message: str, **kwargs):
    log_entry = {
        'timestamp': datetime.now().isoformat(),
        'level': level,
        'message': message,
        **kwargs
    }
    print(json.dumps(log_entry))
```

## Example 2: Error Reporting

```python
# Report errors with context
def report_error(error: Exception, context: dict):
    error_report = {
        'error_type': type(error).__name__,
        'message': str(error),
        'context': context,
        'traceback': traceback.format_exc()
    }
    log_to_service(error_report)
```

**Learn More**: See advanced-patterns.md for detailed communication strategies.
