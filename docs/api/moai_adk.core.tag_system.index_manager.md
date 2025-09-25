# moai_adk.core.tag_system.index_manager

@FEATURE:INDEX-MANAGER-001 - 실시간 TAG 인덱스 관리 최소 구현

watchdog 기반 파일 감지 및 인덱스 실시간 갱신

## Functions

### __init__

```python
__init__(self, manager)
```

### is_watching

파일 감시 상태 반환

```python
is_watching(self)
```

### initialize_index

빈 인덱스 구조 생성

```python
initialize_index(self)
```

### start_watching

파일 감시 시작

```python
start_watching(self)
```

### stop_watching

파일 감시 중지

```python
stop_watching(self)
```

### process_file_change

파일 변경 처리

Args:
    file_path: 변경된 파일 경로
    event_type: 이벤트 타입 (created, modified, deleted)

```python
process_file_change(self, file_path, event_type)
```

### load_index

인덱스 로드

```python
load_index(self)
```

### save_index

인덱스 저장

```python
save_index(self, index_data)
```

### validate_index_schema

인덱스 스키마 검증

```python
validate_index_schema(self, index_data)
```

### _update_file_in_index

파일을 인덱스에 업데이트

```python
_update_file_in_index(self, file_path)
```

### _remove_file_from_index

인덱스에서 파일 제거

```python
_remove_file_from_index(self, file_path)
```

### _remove_file_tags_from_categories

카테고리에서 특정 파일의 TAG 제거

```python
_remove_file_tags_from_categories(self, file_path, index_data)
```

### _get_category_group

카테고리 그룹 결정

```python
_get_category_group(self, category)
```

### on_created

```python
on_created(self, event)
```

### on_modified

```python
on_modified(self, event)
```

### on_deleted

```python
on_deleted(self, event)
```

## Classes

### WatcherStatus

파일 감시 상태

### IndexUpdateEvent

인덱스 업데이트 이벤트

### TagIndexManager

실시간 TAG 인덱스 관리 최소 구현

TRUST 원칙 적용:
- Test First: 테스트 요구사항에 맞춘 최소 구현
- Readable: 명확한 인덱스 관리 로직
- Unified: 인덱스 관리 책임만 담당

### _TagFileEventHandler

내부 파일 이벤트 핸들러
