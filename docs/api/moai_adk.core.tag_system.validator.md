# moai_adk.core.tag_system.validator

@FEATURE:TAG-VALIDATOR-001 - Primary Chain 검증 로직 최소 구현

 TAG 체인 유효성 검증 및 무결성 검사

## Functions

### __init__

검증기 초기화

```python
__init__(self)
```

### validate_primary_chain

Primary Chain 검증

Args:
    tags: 검증할 TAG 목록

Returns:
    검증 결과

```python
validate_primary_chain(self, tags)
```

### detect_circular_references

순환 참조 검사

Args:
    tags: 검사할 TAG 목록

Returns:
    순환 참조가 발견된 TAG 체인 목록

```python
detect_circular_references(self, tags)
```

### find_orphaned_tags

고아 TAG 검색

Args:
    tags: 검사할 TAG 목록

Returns:
    고아 TAG 목록

```python
find_orphaned_tags(self, tags)
```

### check_naming_consistency

명명 일관성 검사

Args:
    tags: 검사할 TAG 목록

Returns:
    일관성 위반 목록

```python
check_naming_consistency(self, tags)
```

### calculate_tag_coverage

TAG 커버리지 계산

Args:
    tags: 커버리지를 계산할 TAG 목록

Returns:
    카테고리별 커버리지 비율

```python
calculate_tag_coverage(self, tags)
```

### validate_reference_integrity

참조 무결성 검사

Args:
    tags: 검사할 TAG 목록

Returns:
    깨진 참조 목록

```python
validate_reference_integrity(self, tags)
```

### _is_consistent_naming

명명 일관성 검사

```python
_is_consistent_naming(self, identifier)
```

### dfs

```python
dfs(tag_key, path)
```

## Classes

### ChainValidationResult

Chain 검증 결과

### ValidationError

검증 오류 정보

### BrokenReference

깨진 참조 정보

### ConsistencyViolation

일관성 위반 정보

### TagValidator

Primary Chain 검증 최소 구현

TRUST 원칙 적용:
- Test First: 테스트 요구사항만 충족
- Readable: 명확한 검증 로직
- Unified: 검증 책임만 담당
