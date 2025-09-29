# moai_adk.resources.templates.scripts.validate_tags

MoAI-ADK Tag System Validator v0.1.12
 @TAG 무결성 검사 및 추적성 매트릭스 검증

이 스크립트는 프로젝트 전체의 @TAG 시스템을:
-  태그 체계 준수 검증
- 고아 태그 및 연결 끊김 감지  
- 태그 인덱스 일관성 확인
- 추적성 매트릭스 업데이트
- 태그 품질 점수 계산

## Functions

### main

메인 실행 함수

```python
main()
```

### __init__

```python
__init__(self, project_root)
```

### scan_project_files

프로젝트 파일에서 모든 태그 스캔

```python
scan_project_files(self)
```

### build_tag_index

태그 인덱스 구축

```python
build_tag_index(self, tags)
```

### validate_tag_format

태그 형식 검증

```python
validate_tag_format(self, tag)
```

### find_orphan_tags

고아 태그 (참조되지 않는 태그) 찾기

```python
find_orphan_tags(self, tag_index)
```

### find_broken_links

깨진 링크 (존재하지 않는 태그 참조) 찾기

```python
find_broken_links(self, tag_index)
```

### validate_traceability_chains

추적성 체인 검증

```python
validate_traceability_chains(self, tag_index)
```

### calculate_tag_quality_score

태그 품질 점수 계산

```python
calculate_tag_quality_score(self, total_tags, valid_tags, orphan_tags, broken_links)
```

### update_tag_indexes

태그 인덱스 파일 업데이트

```python
update_tag_indexes(self, tag_index)
```

### generate_health_report

태그 시스템 건강도 리포트 생성

```python
generate_health_report(self, total_tags, valid_tags, invalid_tags, orphan_tags, broken_links, chain_violations)
```

### run_validation

전체 태그 검증 실행

```python
run_validation(self)
```

## Classes

### TagReference

태그 참조 정보

### TagHealthReport

태그 건강도 리포트

### TagValidator

 TAG 시스템 검증기
