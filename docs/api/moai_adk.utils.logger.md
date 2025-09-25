# moai_adk.utils.logger

@FEATURE:LOGGER-UTILITIES-001 🗿 MoAI-ADK Logging Utilities
@TASK:STRUCTURED-LOGGING-001 Provides structured logging functionality for MoAI-ADK operations

This module provides:
- Configured logger instances with color formatting
- Project-specific logging setup
- Silent mode support for automated operations
- Structured logging for debugging and audit trails

## Functions

### get_logger

@TASK:GET-LOGGER-001 Get a configured logger instance

Args:
    name: Logger name
    level: Logging level (DEBUG, INFO, WARNING, ERROR, CRITICAL)

Returns:
    Configured logger instance

```python
get_logger(name, level)
```

### setup_project_logging

@TASK:PROJECT-LOGGING-001 Setup project-specific logging

Args:
    project_path: Path to the project
    silent: Whether to enable silent mode

Returns:
    Project logger instance

```python
setup_project_logging(project_path, silent)
```
