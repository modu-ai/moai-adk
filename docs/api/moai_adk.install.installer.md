# moai_adk.install.installer

@FEATURE:INSTALLER-001 Simplified MoAI-ADK Project Installer

@TASK:INSTALL-001 Simplified installation system using embedded package resources instead of symbolic links.
@TASK:INSTALL-002 Ensures perfect compatibility with Claude Code by directly copying all resources through ResourceManager.

## Functions

### __init__

Initialize installation manager

Args:
    config: Project configuration

```python
__init__(self, config)
```

### install

Execute MoAI-ADK project installation

Args:
    progress_callback: Progress callback function

Returns:
    InstallationResult: Installation result

```python
install(self, progress_callback)
```

### _create_basic_structure

Create only auxiliary directories that won't be populated by ResourceManager

```python
_create_basic_structure(self)
```

### _install_claude_resources

Claude Code 리소스 설치

```python
_install_claude_resources(self)
```

### _install_moai_resources

MoAI 리소스 설치

```python
_install_moai_resources(self)
```

### _install_github_workflows

GitHub 워크플로우 설치

```python
_install_github_workflows(self)
```

### _install_project_memory

프로젝트 메모리 파일 생성

```python
_install_project_memory(self)
```

### _write_resource_version_info

Record the installed template/resource version metadata.

```python
_write_resource_version_info(self)
```

### _create_configuration_files

설정 파일 생성

```python
_create_configuration_files(self)
```

### _initialize_git_repository

Git 저장소 초기화

```python
_initialize_git_repository(self)
```

### _verify_installation

설치 검증

```python
_verify_installation(self)
```

### _generate_next_steps

Generate next steps guidance

```python
_generate_next_steps(self)
```

## Classes

### SimplifiedInstaller

@TASK:INSTALLER-MAIN-001 Simplified MoAI-ADK project installation manager

Installation system that directly copies embedded package resources
instead of symbolic links for stable operation across all platforms.
