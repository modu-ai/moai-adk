# moai_adk.core.tag_system.adapter

@FEATURE:SPEC-009-TAG-ADAPTER-001 - SQLite 백엔드와 JSON API 호환성 어댑터

GREEN 단계: 기존 JSON API와 100% 호환되는 SQLite 어댑터 구현

## Functions

### __init__

어댑터 초기화

```python
__init__(self, database_path, json_fallback_path, performance_monitor)
```

### is_watching

파일 감시 상태 반환 (기존 API 호환)

```python
is_watching(self)
```

### initialize

초기화 (기존 initialize_index와 호환)

```python
initialize(self)
```

### initialize_index

기존 API 호환을 위한 별칭

```python
initialize_index(self)
```

### load_index

기존 JSON API와 완전 호환되는 인덱스 로드

Returns:
    기존 JSON 형식과 동일한 구조

```python
load_index(self)
```

### _load_from_sqlite

SQLite에서 기존 JSON 형식으로 데이터 변환

```python
_load_from_sqlite(self)
```

### save_index

인덱스 저장 (기존 API 호환)

```python
save_index(self, index_data)
```

### _save_to_sqlite

JSON 형식 데이터를 SQLite에 저장

```python
_save_to_sqlite(self, index_data)
```

### process_file_change

파일 변경 처리 (기존 API와 동일한 시그니처)

Args:
    file_path: 변경된 파일 경로
    event_type: 이벤트 타입 (created, modified, deleted)

```python
process_file_change(self, file_path, event_type)
```

### _process_file_change_sqlite

SQLite 백엔드에서 파일 변경 처리

```python
_process_file_change_sqlite(self, file_path, event_type)
```

### validate_index_schema

인덱스 스키마 검증 (기존 API와 동일)

```python
validate_index_schema(self, index_data)
```

### start_watching

파일 감시 시작 (기존 API 호환)

```python
start_watching(self)
```

### stop_watching

파일 감시 중지 (기존 API 호환)

```python
stop_watching(self)
```

### migrate_from_json

JSON에서 SQLite로 데이터 마이그레이션

```python
migrate_from_json(self, json_path)
```

### export_to_json

SQLite에서 JSON으로 데이터 내보내기

```python
export_to_json(self, json_path)
```

### get_configuration_info

설정 정보 반환 (디버깅용)

```python
get_configuration_info(self)
```

### _create_empty_index

빈 인덱스 구조 생성

```python
_create_empty_index(self)
```

### _get_category_group

카테고리 그룹 결정 (기존 로직과 동일)

```python
_get_category_group(self, category)
```

### search_by_category

카테고리별 TAG 검색 (JSON API 호환)

TRUST 원칙 적용:
- Test First: 실패 테스트로 시작하여 최소 구현
- Readable: 명확한 매개변수와 반환값 문서화
- Unified: 기존 JSON API와 완전 호환
- Secured: 입력 검증과 구조화 로깅
- Trackable: 성능 메트릭 수집

Args:
    category: TAG 카테고리 (REQ, DESIGN, TASK, TEST 등)
             빈 문자열이나 None은 빈 결과 반환
    **filters: 추가 필터 (향후 확장용)
             file_pattern: 파일 경로 패턴 필터
             description_pattern: 설명 텍스트 필터

Returns:
    JSON API 형식의 TAG 목록:
    [
        {
            "category": "REQ",
            "identifier": "USER-AUTH-001",
            "description": "사용자 인증 요구사항",
            "file_path": "specs/auth.md",
            "line_number": 25
        }
    ]

Raises:
    ValueError: 유효하지 않은 카테고리 이름

```python
search_by_category(self, category)
```

### get_traceability_chain

TAG 추적성 체인 구축 (16-Core TAG 시스템 지원)

TRUST 원칙 적용:
- Test First: 실패 테스트 기반 구현
- Readable: 명확한 체인 구조와 예제
- Unified: 표준 그래프 구조 사용
- Secured: 순환 참조 방지와 깊이 제한
- Trackable: 체인 구축 성능 모니터링

Args:
    tag_identifier: TAG 식별자
                  형식: "CATEGORY:IDENTIFIER" (예: "REQ:USER-AUTH-001")
                  또는 "IDENTIFIER" (카테고리 자동 추론)
    direction: 참조 방향
              "forward": 순방향 (REQ → DESIGN → TASK → TEST)
              "backward": 역방향 (TEST → TASK → DESIGN → REQ)
              "both": 양방향 (전체 연결 그래프)
    max_depth: 최대 탐색 깊이 (1-50, 기본값 10)
              순환 참조 방지와 성능 보장

Returns:
    추적성 체인 그래프 구조:
    {
        "nodes": [
            {
                "id": 123,
                "identifier": "REQ:USER-AUTH-001",
                "category": "REQ",
                "description": "사용자 인증 요구사항",
                "file_path": "specs/auth.md"
            }
        ],
        "edges": [
            {
                "source": 123,
                "target": 456,
                "type": "chain",
                "direction": "forward"
            }
        ],
        "direction": "forward",
        "max_depth": 10,
        "truncated": false,
        "metadata": {
            "total_nodes": 4,
            "total_edges": 3,
            "chain_depth": 4
        }
    }

Raises:
    ValueError: 잘못된 매개변수 값
    ApiCompatibilityError: 백엔드 연결 실패

```python
get_traceability_chain(self, tag_identifier, direction, max_depth)
```

### close

리소스 정리

```python
close(self)
```

## Classes

### ApiCompatibilityError

API 호환성 관련 오류

### AdapterConfiguration

어댑터 설정

### TagIndexAdapter

SQLite 백엔드와 기존 JSON API 호환성 어댑터

TRUST 원칙 적용:
- Test First: 기존 API 호환성 테스트 통과
- Readable: 명확한 어댑터 패턴 구현
- Unified: API 변환 책임만 담당
