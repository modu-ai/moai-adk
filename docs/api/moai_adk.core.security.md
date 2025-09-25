# moai_adk.core.security

@FEATURE:SECURITY-001 Security utilities for MoAI-ADK

Provides basic path validation and safe operations for a local development tool.

## Functions

### __init__

```python
__init__(self)
```

### _get_critical_paths

Get system critical paths that should never be deleted.

```python
_get_critical_paths(self)
```

### validate_path_safety

Basic path validation to prevent path traversal.

Args:
    path: Path to validate
    base_path: Base path that should contain the target

Returns:
    bool: True if path is safe to use

```python
validate_path_safety(self, path, base_path)
```

### safe_rmtree

Safely remove a directory tree after validation.

Args:
    path: Directory to remove

Returns:
    bool: True if removal was successful

```python
safe_rmtree(self, path)
```

### validate_file_creation

Validate that a file can be safely created.

Args:
    file_path: Path where file will be created
    base_path: Base project directory

Returns:
    bool: True if file creation is safe

```python
validate_file_creation(self, file_path, base_path)
```

### validate_subprocess_path

Validate that a subprocess can be safely executed within the given paths.

Args:
    path: Path where subprocess will operate
    base_path: Base project directory

Returns:
    bool: True if subprocess execution is safe

```python
validate_subprocess_path(self, path, base_path)
```

### sanitize_filename

Sanitize filename to prevent filesystem issues.

Args:
    filename: Original filename

Returns:
    str: Sanitized filename

```python
sanitize_filename(self, filename)
```

### safe_subprocess_run

Safe subprocess execution with basic validation.

Args:
    command: Command to execute
    *args: Additional arguments
    **kwargs: Additional keyword arguments

Returns:
    subprocess result

```python
safe_subprocess_run(self, command)
```

### sanitize_command_args

Sanitize command arguments for safe execution.

Args:
    args: Command arguments to sanitize

Returns:
    Sanitized arguments

```python
sanitize_command_args(self, args)
```

### validate_path_safety_enhanced

Enhanced path validation with additional checks.

Args:
    path: Path to validate
    base_path: Base path that should contain the target
    allow_creation: Whether to allow path creation

Returns:
    bool: True if path is safe to use

```python
validate_path_safety_enhanced(self, path, base_path, allow_creation)
```

### validate_file_size

Validate file size to prevent resource exhaustion.

Args:
    file_path: Path to file
    max_size_mb: Maximum allowed size in megabytes

Returns:
    bool: True if file size is acceptable

```python
validate_file_size(self, file_path, max_size_mb)
```

## Classes

### SecurityError

Security-related exception.

### SecurityManager

@TASK:SECURITY-MANAGER-001 Manages basic security operations for local development environment.
