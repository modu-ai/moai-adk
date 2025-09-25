# moai_adk.core.tag_system.migration

@FEATURE:SPEC-009-TAG-MIGRATION-001 - TAG 마이그레이션 도구

GREEN 단계: JSON ↔ SQLite 양방향 변환 및 데이터 무결성 보장

## Functions

### percentage

진행률 퍼센트

```python
percentage(self)
```

### __post_init__

```python
__post_init__(self)
```

### __init__

마이그레이션 도구 초기화

```python
__init__(self, database_path, json_path, backup_directory)
```

### create_backup

백업 생성

```python
create_backup(self, json_path, description)
```

### migrate_json_to_sqlite

JSON에서 SQLite로 마이그레이션

```python
migrate_json_to_sqlite(self, validate_data, strict_mode, progress_callback, batch_size, mode, preserve_existing, conflict_resolution, create_backup, auto_rollback, detailed_reporting, generate_report, plugins, strict_validation)
```

### migrate_sqlite_to_json

SQLite에서 JSON으로 마이그레이션

```python
migrate_sqlite_to_json(self)
```

### _convert_sqlite_to_original_json_format

SQLite 데이터를 원본 JSON 형식으로 변환

```python
_convert_sqlite_to_original_json_format(self, all_tags, db_manager)
```

### _get_category_group_for_stats

통계용 카테고리 그룹 결정

```python
_get_category_group_for_stats(self, category)
```

### get_database_tag_count

데이터베이스 TAG 개수 반환

```python
get_database_tag_count(self)
```

### _validate_json_data

JSON 데이터 검증

```python
_validate_json_data(self, json_data, strict_mode)
```

### _full_migration

전체 마이그레이션

```python
_full_migration(self, db_manager, json_data, result, progress_callback, batch_size)
```

### _incremental_migration

증분 마이그레이션

```python
_incremental_migration(self, db_manager, json_data, result, progress_callback)
```

### _apply_plugins

플러그인 적용

```python
_apply_plugins(self, tag_data)
```

### _validate_with_plugins

플러그인 검증

```python
_validate_with_plugins(self, tag_data)
```

### _generate_detailed_statistics

상세 통계 생성

```python
_generate_detailed_statistics(self, result, db_manager)
```

### _generate_html_report

HTML 리포트 생성

```python
_generate_html_report(self, result)
```

### _perform_rollback

롤백 수행

```python
_perform_rollback(self, backup_info)
```

### _validate_migration_result

마이그레이션 결과 검증 (테스트용)

```python
_validate_migration_result(self)
```

## Classes

### MigrationError

마이그레이션 관련 오류

### DataValidationError

데이터 검증 오류

### ValidationError

검증 오류 정보

### MigrationProgress

마이그레이션 진행률 정보

### ConflictResolution

충돌 해결 정보

### MigrationResult

마이그레이션 결과

### BackupInfo

백업 정보

### BackupManager

백업 관리자

### TagMigrationTool

TAG 마이그레이션 도구

TRUST 원칙 적용:
- Test First: 테스트 요구사항에 맞춘 최소 구현
- Readable: 명확한 마이그레이션 로직
- Unified: 마이그레이션 책임만 담당
