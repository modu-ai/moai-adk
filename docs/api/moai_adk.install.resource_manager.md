# moai_adk.install.resource_manager

@FEATURE:RESOURCE-001 MoAI-ADK Resource Manager

@TASK:RESOURCE-001 패키지 내장 리소스를 관리하는 모듈입니다.
@TASK:RESOURCE-002 심볼릭 링크 대신 패키지에서 직접 리소스를 복사하여 관리합니다.

## Functions

### __init__

ResourceManager 초기화

```python
__init__(self)
```

### get_version

패키지 버전 반환

```python
get_version(self)
```

### get_template_path

템플릿 경로 반환 (읽기 전용)

```python
get_template_path(self, template_name)
```

### get_template_content

템플릿 내용 반환

Args:
    template_name: 템플릿 파일명

Returns:
    str: 템플릿 내용 (없으면 None)

```python
get_template_content(self, template_name)
```

### copy_template

템플릿을 대상 경로로 복사

Args:
    template_name: 복사할 템플릿 이름 (.claude, .moai 등)
    target_path: 복사 대상 경로
    overwrite: 기존 파일 덮어쓰기 여부

Returns:
    bool: 복사 성공 여부

```python
copy_template(self, template_name, target_path, overwrite, exclude_subdirs)
```

### _validate_safe_path

경로 안전성 검증

Args:
    target_path: 검증할 경로

Returns:
    bool: 안전한 경로 여부

```python
_validate_safe_path(self, target_path)
```

### copy_claude_resources

Claude Code 관련 리소스를 프로젝트에 복사

Args:
    project_path: 프로젝트 경로
    overwrite: 기존 파일 덮어쓰기 여부

Returns:
    List[Path]: 복사된 파일 경로들

```python
copy_claude_resources(self, project_path, overwrite)
```

### _ensure_hook_permissions

Ensure executable permissions for hook python files.

대상: {claude_root}/hooks/moai/*.py
Windows에서는 무시(권한 비트 미사용)되지만 호출 자체는 안전합니다.

```python
_ensure_hook_permissions(self, claude_root)
```

### copy_moai_resources

MoAI 관련 리소스를 프로젝트에 복사

Args:
    project_path: 프로젝트 경로
    overwrite: 기존 파일 덮어쓰기 여부
    exclude_templates: 템플릿 디렉토리 제외 여부

Returns:
    List[Path]: 복사된 파일 경로들

```python
copy_moai_resources(self, project_path, overwrite, exclude_templates)
```

### copy_github_resources

GitHub 워크플로우 리소스를 프로젝트에 복사

Args:
    project_path: 프로젝트 경로
    overwrite: 기존 파일 덮어쓰기 여부

Returns:
    List[Path]: 복사된 파일 경로들

```python
copy_github_resources(self, project_path, overwrite)
```

### copy_project_memory

프로젝트 메모리 파일(CLAUDE.md) 생성

Args:
    project_path: 프로젝트 경로
    overwrite: 기존 파일 덮어쓰기 여부

Returns:
    bool: 생성 성공 여부

```python
copy_project_memory(self, project_path, overwrite)
```

### copy_memory_templates

Copy stack-specific memory templates into the project.

```python
copy_memory_templates(self, project_path, tech_stack, context, overwrite)
```

### _render_template_with_context

Render a text template with context variables to a target file.

```python
_render_template_with_context(self, template_name, target_path, context, overwrite)
```

### validate_project_resources

프로젝트 리소스 검증 (validate_required_resources와 동일)

Args:
    project_path: 프로젝트 경로

Returns:
    bool: 검증 성공 여부

```python
validate_project_resources(self, project_path)
```

### list_templates

사용 가능한 템플릿 목록 반환

```python
list_templates(self)
```

### validate_required_resources

필수 리소스가 모두 있는지 확인

```python
validate_required_resources(self, project_path)
```

### _validate_clean_installation

@TASK:TEMPLATE-VERIFY-001 설치된 리소스가 깨끗한 초기 상태인지 검증

Args:
    target_path: 검증할 대상 경로

Returns:
    bool: 깨끗한 설치 여부

```python
_validate_clean_installation(self, target_path)
```

### copy_function

```python
copy_function(src, dst)
```

## Classes

### ResourceManager

패키지 내장 리소스 관리 클래스

pip으로 설치된 패키지에서 템플릿과 설정 파일을 관리합니다.
심볼릭 링크를 사용하지 않고 직접 복사 방식을 사용합니다.
