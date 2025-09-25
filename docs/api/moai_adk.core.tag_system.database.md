# moai_adk.core.tag_system.database

@FEATURE:SPEC-009-TAG-DATABASE-001 - SQLite 기반 TAG 데이터베이스 관리자

GREEN 단계: 테스트를 통과시키는 최소 구현

## Functions

### __init__

```python
__init__(self, database)
```

### get_connection

스레드별 독립적인 연결 반환

```python
get_connection(self)
```

### close

데이터베이스 연결 종료

```python
close(self)
```

### execute

배치용 데이터 추가

```python
execute(self, category, identifier, description, file_path, line_number)
```

### executemany

배치 쿼리 실행

```python
executemany(self, query, params_list)
```

### commit

트랜잭션 커밋

```python
commit(self)
```

### initialize

데이터베이스 초기화

```python
initialize(self, create_indexes)
```

### _create_update_trigger

updated_at 자동 업데이트 트리거 생성

```python
_create_update_trigger(self)
```

### get_schema

스키마 정보 반환

```python
get_schema(self)
```

### get_indexes

인덱스 목록 반환

```python
get_indexes(self)
```

### insert_tag

TAG 삽입 (성능 모니터링 포함)

```python
insert_tag(self, category, identifier, description, file_path, line_number)
```

### get_tag_by_id

ID로 TAG 조회

```python
get_tag_by_id(self, tag_id)
```

### search_tags_by_category

카테고리별 TAG 검색

```python
search_tags_by_category(self, category)
```

### search_tags_by_identifier

식별자로 TAG 검색

```python
search_tags_by_identifier(self, identifier)
```

### search_tags_by_file

파일 경로로 TAG 검색

```python
search_tags_by_file(self, file_path)
```

### search_tags_by_pattern

패턴으로 TAG 검색

```python
search_tags_by_pattern(self, pattern)
```

### search_tags_by_line_range

줄 번호 범위로 TAG 검색

```python
search_tags_by_line_range(self, start_line, end_line)
```

### create_reference

TAG 참조 관계 생성

```python
create_reference(self, source_tag_id, target_tag_id, reference_type)
```

### get_references_by_source

소스 TAG의 참조 관계 조회

```python
get_references_by_source(self, source_tag_id)
```

### update_tag

TAG 업데이트

```python
update_tag(self, tag_id)
```

### delete_tag

TAG 삭제 (CASCADE로 참조도 삭제됨)

```python
delete_tag(self, tag_id)
```

### get_all_tags

모든 TAG 조회

```python
get_all_tags(self)
```

### transaction

트랜잭션 컨텍스트 매니저

```python
transaction(self)
```

### complex_search

복합 검색

```python
complex_search(self, category, file_pattern, line_range)
```

### search_tags_by_file_pattern

파일 패턴으로 검색

```python
search_tags_by_file_pattern(self, pattern)
```

### prepared_insert

Prepared statement 매니저

```python
prepared_insert(self)
```

### bulk_insert_tags

대량 TAG 삽입

```python
bulk_insert_tags(self, tags_data)
```

### get_tag_metadata

TAG 메타데이터 조회 (확장용)

```python
get_tag_metadata(self, tag_id)
```

### __enter__

```python
__enter__(self)
```

### __exit__

```python
__exit__(self, exc_type, exc_val, exc_tb)
```

## Classes

### DatabaseError

데이터베이스 관련 오류

### TransactionError

트랜잭션 관련 오류

### TagSearchResult

TAG 검색 결과

### DatabaseConnection

데이터베이스 연결 관리 (Thread-Safe 개선)

TRUST 원칙 적용:
- Secured: 스레드별 별도 연결로 안전성 확보
- Readable: 명확한 연결 관리 로직

### TagDatabase

SQLite 데이터베이스 추상화

### TagDatabaseManager

SQLite 기반 TAG 데이터베이스 관리자

TRUST 원칙 적용:
- Test First: 테스트가 요구하는 최소 기능만 구현
- Readable: 명확한 데이터베이스 작업 로직
- Unified: 단일 책임 - 데이터베이스 관리만 담당

### TransactionManager

트랜잭션 관리자

### PreparedInsertManager

Prepared statement 삽입 관리자
