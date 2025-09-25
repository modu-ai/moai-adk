# moai_adk.core.version_sync

@FEATURE:VERSION-SYNC-001 MoAI-ADK Automated Version Synchronization System
Automated version synchronization system for MoAI-ADK project files

## Functions

### main

CLI entry point

```python
main()
```

### __init__

Initialize version sync manager

Args:
    project_root: Project root directory. Auto-detected if None

```python
__init__(self, project_root)
```

### _find_project_root

Find project root directory containing pyproject.toml

```python
_find_project_root(self)
```

### _load_version_patterns

Define version patterns - file-specific replacement rules

```python
_load_version_patterns(self)
```

### sync_all_versions

전체 프로젝트의 버전 정보 동기화

Args:
    dry_run: True면 실제 변경하지 않고 시뮬레이션만
    
Returns:
    Dict[파일패턴, 변경된파일리스트]

```python
sync_all_versions(self, dry_run)
```

### _sync_pattern

특정 파일 패턴에 대해 버전 동기화 수행

```python
_sync_pattern(self, file_pattern, replacements, dry_run)
```

### _should_skip_file

파일 스킵 조건 확인

```python
_should_skip_file(self, file_path)
```

### _sync_file

단일 파일의 버전 동기화

```python
_sync_file(self, file_path, replacements, dry_run)
```

### verify_sync

버전 동기화 검증 - 남은 불일치 확인

```python
verify_sync(self)
```

### _find_version_mismatches

버전 불일치 파일 찾기

```python
_find_version_mismatches(self, pattern, expected)
```

### create_version_update_script

버전 업데이트용 스크립트 생성

```python
create_version_update_script(self)
```

## Classes

### VersionSyncManager

@TASK:VERSION-SYNC-MANAGER-001 Version synchronization manager class
