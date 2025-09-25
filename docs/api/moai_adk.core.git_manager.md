# moai_adk.core.git_manager

@FEATURE:GIT-001 Git repository management utilities for MoAI-ADK

Handles Git repository initialization, validation, and operations
with security validation and error handling.

## Functions

### __init__

Initialize Git manager.

Args:
    project_dir: 프로젝트 디렉토리 (새로운 기능을 위해 추가)
    config: 설정 관리자 인스턴스 (새로운 기능을 위해 추가)
    security_manager: Security manager instance for validation
    file_manager: File manager for .gitignore creation

```python
__init__(self, project_dir, config, security_manager, file_manager)
```

### initialize_repository

Initialize git repository if not already initialized.

Args:
    project_path: Project root path

Returns:
    bool: True if git repo exists or was successfully initialized

```python
initialize_repository(self, project_path)
```

### _check_git_available

Check if git is available in the system.

```python
_check_git_available(self)
```

### _initialize_repository

Initialize git repository with security validation.

```python
_initialize_repository(self, project_path)
```

### _offer_git_installation

Offer to install git and attempt installation if user agrees.

```python
_offer_git_installation(self)
```

### _get_git_install_command

Get Git installation command based on OS.

```python
_get_git_install_command(self, os_name)
```

### _check_command_exists

Check if a command exists in the system.

```python
_check_command_exists(self, command)
```

### _install_git_with_command

Install git using the provided command.

```python
_install_git_with_command(self, install_cmd, os_name)
```

### check_git_status

Check the status of Git repository.

Args:
    project_path: Project root path

Returns:
    dict: Git status information

```python
check_git_status(self, project_path)
```

### get_git_info

Get comprehensive Git repository information.

Args:
    project_path: Project root path

Returns:
    dict: Git repository information

```python
get_git_info(self, project_path)
```

### _get_remote_info

Get Git remote information.

```python
_get_remote_info(self, project_path)
```

### create_gitignore

Create a comprehensive .gitignore file.

Args:
    gitignore_path: Path where .gitignore should be created

Returns:
    bool: True if successful

```python
create_gitignore(self, gitignore_path)
```

### commit_with_lock

잠금 시스템을 사용한 안전한 커밋

Args:
    message: 커밋 메시지
    files: 커밋할 파일 목록 (None이면 모든 변경사항)
    mode: Git 모드 ("personal" 또는 "team")

```python
commit_with_lock(self, message, files, mode)
```

### set_strategy

모드에 따른 Git 전략 설정

Args:
    mode: Git 모드 ("personal" 또는 "team")

```python
set_strategy(self, mode)
```

### get_mode

현재 Git 모드 반환

Returns:
    현재 Git 모드 ("personal", "team", 또는 "unknown")

```python
get_mode(self)
```

## Classes

### GitManager

@TASK:GIT-MANAGER-001 Manages Git operations for MoAI-ADK installation.
