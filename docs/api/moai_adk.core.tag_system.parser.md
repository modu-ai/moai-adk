# moai_adk.core.tag_system.parser

@FEATURE:TAG-PARSER-001 - TAG 파싱 엔진 최소 구현

16-Core TAG 시스템의 TAG 추출 및 분류 엔진

## Functions

### __post_init__

```python
__post_init__(self)
```

### __init__

TAG 파서 초기화

```python
__init__(self)
```

### get_tag_categories

16-Core TAG 카테고리 반환

```python
get_tag_categories(self)
```

### extract_tags

텍스트에서 TAG 추출

Args:
    content: 분석할 텍스트 콘텐츠

Returns:
    추출된 TAG 목록

```python
extract_tags(self, content)
```

### parse_tag_chains

TAG 체인 파싱

Args:
    content: 체인이 포함된 텍스트

Returns:
    파싱된 TAG 체인 목록

```python
parse_tag_chains(self, content)
```

### validate_tag_format

TAG 형식 검증

Args:
    tag_string: 검증할 TAG 문자열

Returns:
    유효한 형식인지 여부

```python
validate_tag_format(self, tag_string)
```

### extract_tags_with_positions

위치 정보와 함께 TAG 추출

Args:
    content: 분석할 텍스트

Returns:
    (TAG, 위치) 튜플 목록

```python
extract_tags_with_positions(self, content)
```

### find_duplicate_tags

중복 TAG 검색

Args:
    content: 검색할 텍스트

Returns:
    중복 TAG 정보 목록

```python
find_duplicate_tags(self, content)
```

### _is_valid_tag_category

TAG 카테고리 유효성 검사

```python
_is_valid_tag_category(self, category)
```

## Classes

### TagCategory

16-Core TAG 카테고리 분류

### TagMatch

TAG 매칭 결과

### TagPosition

TAG 위치 정보

### TagChain

TAG 체인

### DuplicateTagInfo

중복 TAG 정보

### TagParser

16-Core TAG 파싱 엔진 최소 구현

TRUST 원칙 적용:
- Test First: 테스트가 요구하는 최소 기능만 구현
- Readable: 명시적이고 이해하기 쉬운 코드
- Unified: 단일 책임 - TAG 파싱만 담당
