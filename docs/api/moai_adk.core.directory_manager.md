# moai_adk.core.directory_manager

@FEATURE:DIRECTORY-001 Directory management utilities for MoAI-ADK

Handles directory creation, structure setup, and safe directory operations
with security validation and proper error handling.

## Functions

### __init__

Initialize directory manager.

Args:
    security_manager: Security manager instance for validation

```python
__init__(self, security_manager)
```

### create_project_directory

Create the main project directory, preserving existing files when safe.

Args:
    config: Project configuration containing path and settings

```python
create_project_directory(self, config)
```

### _handle_force_overwrite

Handle force overwrite operation with Git directory preservation.

Args:
    project_path: Path to project directory

```python
_handle_force_overwrite(self, project_path)
```

### create_directory_structure

Create the complete MoAI-ADK directory structure.

Args:
    base_path: Base project path

Returns:
    List[Path]: List of created directories

```python
create_directory_structure(self, base_path)
```

### ensure_directory_exists

Ensure a directory exists, creating it if necessary with security validation.

Args:
    directory: Directory path to ensure exists
    base_path: Base path for security validation (optional)

Returns:
    bool: True if directory exists or was created successfully

```python
ensure_directory_exists(self, directory, base_path)
```

### get_directory_info

Get information about a directory.

Args:
    directory: Directory to analyze

Returns:
    dict: Directory information including size, file counts, etc.

```python
get_directory_info(self, directory)
```

### clean_directory

Clean directory contents while preserving specified patterns.

Args:
    directory: Directory to clean
    preserve_patterns: List of glob patterns to preserve

Returns:
    bool: True if successful

```python
clean_directory(self, directory, preserve_patterns)
```

### create_backup_directory

Create a backup directory with timestamp.

Args:
    source_path: Source directory to backup
    backup_base: Base directory for backups (default: parent of source)

Returns:
    Path: Path to created backup directory

```python
create_backup_directory(self, source_path, backup_base)
```

## Classes

### DirectoryManager

@TASK:DIRECTORY-MANAGER-001 Manages directory operations for MoAI-ADK installation.
