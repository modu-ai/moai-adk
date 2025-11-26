# API Reference

## Overview

Complete API reference for skill integration and usage.

## Core APIs

### Initialization

```python
from moai_adk import SkillLoader

loader = SkillLoader()
skill = loader.load("skill-name")
```

### Execution

```python
result = skill.execute(
    input_data=data,
    options={"mode": "strict"}
)
```

## Configuration

### Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| mode | str | "normal" | Execution mode |
| timeout | int | 30 | Timeout in seconds |
| retry | bool | true | Enable retries |

## Error Handling

```python
try:
    result = skill.execute(data)
except SkillExecutionError as e:
    logger.error(f"Execution failed: {e}")
    handle_error(e)
```

---
**Last Updated**: 2025-11-23
**Status**: Production Ready
