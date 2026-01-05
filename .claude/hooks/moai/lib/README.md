# MoAI-ADK Hooks Library

Shared utilities for Claude Code hooks in MoAI-ADK.

## Module Overview

```
lib/
├── __init__.py           # Package exports
├── common.py             # Shared utilities (format_duration, merge_configs)
├── config_manager.py     # Configuration loading and management
├── config_validator.py   # Configuration schema validation
├── exceptions.py         # Exception hierarchy
├── models.py             # Data structures (HookPayload, HookResult)
├── path_utils.py         # Project root detection, safe paths
├── project.py            # Project metadata (language, Git, SPEC)
├── timeout.py            # Cross-platform timeout (Windows/Unix)
├── unified_timeout_manager.py  # Advanced timeout with retry
├── git_operations_manager.py   # Optimized Git operations
├── checkpoint.py         # Risky operation detection
├── language_validator.py # Language config validation
└── tool_registry.py      # Formatter/linter tool registry
```

## Exit Codes

| Code | Meaning | Usage |
|------|---------|-------|
| 0 | Success | Normal completion |
| 1 | Warning/Error | Non-critical error, logged |
| 2 | Critical Error | Blocks operation |
| 3 | Configuration Error | Invalid config |

## Exception Hierarchy

```
HooksBaseError (base)
├── TimeoutError
│   └── HookTimeoutError (with context)
├── GitOperationError
├── ConfigurationError
├── ValidationError
└── SecurityError
```

## Key Functions

### Path Management (`path_utils.py`)

```python
from lib.path_utils import find_project_root, get_safe_moai_path

# Find project root (with caching)
root = find_project_root()

# Get safe path within .moai/
cache_path = get_safe_moai_path("cache/version.json")
```

### Configuration (`config_manager.py`)

```python
from lib.config_manager import ConfigManager, get_config

# Load configuration
config = ConfigManager().load_config()

# Get specific value
timeout = get_config("hooks.timeout_ms", default=5000)
```

### Timeout Handling (`timeout.py`, `unified_timeout_manager.py`)

```python
# Basic timeout
from lib.timeout import CrossPlatformTimeout

with CrossPlatformTimeout(5):
    long_running_operation()

# Advanced timeout with retry
from lib.unified_timeout_manager import get_timeout_manager

manager = get_timeout_manager()
result = manager.execute_with_timeout("hook_name", func, config=config)
```

### Common Utilities (`common.py`)

```python
from lib.common import merge_configs, format_duration

# Merge configs recursively
merged = merge_configs(base_config, override_config)

# Format duration
formatted = format_duration(125.5)  # "2.1m"
```

## Configuration Files

### Main Config
Location: `.moai/config/config.yaml`

### Section Files
Location: `.moai/config/sections/`
- `user.yaml` - User name
- `language.yaml` - Language preferences
- `project.yaml` - Project metadata
- `git-strategy.yaml` - Git workflow
- `quality.yaml` - TDD settings

## Hook Data Flow

```
Claude Code Event
    ↓
stdin (JSON)
    ↓
Hook Process
    ↓
stdout (JSON response)
    ↓
Claude Code
```

### Response Format

```json
{
  "continue": true,
  "systemMessage": "...",
  "hookSpecificOutput": {
    "hookEventName": "PreToolUse",
    "permissionDecision": "allow|deny|ask"
  }
}
```

## Version

- Library Version: 1.0.0
- Last Updated: 2025-01
