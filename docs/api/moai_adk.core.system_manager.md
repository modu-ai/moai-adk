# moai_adk.core.system_manager

@FEATURE:SYSTEM-001 System utilities and environment checks for MoAI-ADK

Handles Node.js/npm detection, ccusage availability checks,
and other system-level validations.

## Functions

### __init__

Initialize system manager.

```python
__init__(self)
```

### check_nodejs_and_npm

Check if Node.js and npm are installed, and verify ccusage can be used.

Returns:
    bool: True if Node.js environment is properly set up

```python
check_nodejs_and_npm(self)
```

### _check_command_exists

Check if a command exists in the system.

```python
_check_command_exists(self, command)
```

### _validate_nodejs_environment

Validate Node.js environment and ccusage availability.

```python
_validate_nodejs_environment(self)
```

### _test_ccusage_availability

Test ccusage package accessibility.

```python
_test_ccusage_availability(self)
```

### get_system_info

Get comprehensive system information.

Returns:
    dict: System information including OS, Python, Node.js, etc.

```python
get_system_info(self)
```

### _get_nodejs_info

Get Node.js environment information.

```python
_get_nodejs_info(self)
```

### _quick_ccusage_test

Quick test for ccusage availability without output.

```python
_quick_ccusage_test(self)
```

### _get_package_managers_info

Get information about available package managers.

```python
_get_package_managers_info(self)
```

### check_python_version

Check if Python version meets minimum requirements.

Args:
    min_version: Minimum required version as tuple (major, minor)

Returns:
    bool: True if version is sufficient

```python
check_python_version(self, min_version)
```

### detect_project_type

Detect project type based on existing files.

Args:
    project_path: Path to project directory

Returns:
    dict: Detected project information

```python
detect_project_type(self, project_path)
```

### _analyze_package_json

Analyze package.json for frameworks and dependencies.

```python
_analyze_package_json(self, package_json_path)
```

### should_create_package_json

Check if package.json should be created based on project configuration.

Args:
    config: Project configuration

Returns:
    bool: True if package.json should be created

```python
should_create_package_json(self, config)
```

## Classes

### SystemManager

@TASK:SYSTEM-MANAGER-001 Manages system-level checks and validations.
