# Reference Documentation

## Overview

Complete API and configuration reference.

## API Reference

### Core Functions

```python
def initialize_skill(config: Dict) -> Skill:
    """Initialize skill with configuration."""
    pass

def execute_skill(skill: Skill, input_data: Dict) -> Result:
    """Execute skill with input data."""
    pass
```

## Configuration Reference

### Settings

| Setting | Type | Default | Description |
|---------|------|---------|-------------|
| mode | str | "normal" | Execution mode |
| timeout | int | 30 | Timeout seconds |
| retries | int | 3 | Retry attempts |

## Integration Patterns

### Basic Integration

```python
from moai_adk import load_skill

skill = load_skill("skill-name")
result = skill.execute(data)
```

---
**Last Updated**: 2025-11-23
**Status**: Production Ready
