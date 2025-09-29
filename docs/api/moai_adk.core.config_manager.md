# moai_adk.core.config_manager

@FEATURE:CONFIG-001 Configuration management utilities for MoAI-ADK

Handles creation and management of various configuration files including
Claude Code settings, MoAI config, package.json, and other project configs.

## Functions

### __init__

Initialize configuration manager.

Args:
    project_dir: 프로젝트 디렉토리 (새로운 기능을 위해 추가)
    security_manager: Security manager instance for validation

```python
__init__(self, project_dir, security_manager)
```

### create_claude_settings

@TASK:CONFIG-CLAUDE-001 Create Claude Code settings.json file

Args:
    settings_path: Full path to settings.json file
    config: Project configuration

Returns:
    bool: True if settings file was created successfully

```python
create_claude_settings(self, settings_path, config)
```

### create_claude_settings_file

Backward-compatible helper to create settings file under project root.

```python
create_claude_settings_file(self, project_path, config)
```

### create_moai_config

@TASK:CONFIG-MOAI-001 Create .moai/config.json configuration file

Args:
    config_path: Full path to config.json file
    config: Project configuration

Returns:
    bool: True if config file was created successfully

```python
create_moai_config(self, config_path, config)
```

### create_moai_config_file

Backward-compatible helper to create config file under project root.

```python
create_moai_config_file(self, project_path, config)
```

### create_package_json

Create package.json for Node.js projects.

Args:
    project_path: Project root path
    config: Project configuration

Returns:
    Path: Path to created package.json

```python
create_package_json(self, project_path, config)
```

### create_initial_indexes

Create initial  TAG index files.

Args:
    project_path: Project root path
    config: Project configuration

Returns:
    List[Path]: List of created index files

```python
create_initial_indexes(self, project_path, config)
```

### setup_steering_config

Setup MoAI steering configuration with 개발 가이드 5 principles.

Args:
    project_path: Project root path

Returns:
    Path: Path to created steering config

```python
setup_steering_config(self, project_path)
```

### _write_json_file

Write data to JSON file with security validation.

Args:
    file_path: Path to write JSON file
    data: Data to write

Returns:
    Path: Path to written file

Raises:
    SecurityError: If security validation fails

```python
_write_json_file(self, file_path, data)
```

### validate_config_file

Validate that a configuration file is valid JSON.

Args:
    file_path: Path to config file to validate

Returns:
    bool: True if valid

```python
validate_config_file(self, file_path)
```

### backup_config_file

Create a backup of a configuration file.

Args:
    file_path: Path to config file to backup

Returns:
    Path: Path to backup file

```python
backup_config_file(self, file_path)
```

### set_mode

개인/팀 모드 설정

Args:
    mode: 설정할 모드 ("personal" 또는 "team")

```python
set_mode(self, mode)
```

### get_mode

현재 모드 반환

Returns:
    현재 설정된 모드 ("personal" 또는 "team")

```python
get_mode(self)
```

### set_option

옵션 설정

Args:
    key: 옵션 키
    value: 옵션 값

```python
set_option(self, key, value)
```

### get_option

옵션 값 반환

Args:
    key: 옵션 키
    default: 기본값

Returns:
    옵션 값 또는 기본값

```python
get_option(self, key, default)
```

## Classes

### ConfigManager

@TASK:CONFIG-MANAGER-001 Manages configuration files for MoAI-ADK installation.
