# moai_adk.resources.templates.scripts.repair_tags

MoAI-ADK TAG 자동 리페어 시스템
단절된 링크 탐지, 자동 제안, traceability.json 보정

## Functions

### main

```python
main()
```

### __init__

```python
__init__(self, project_root)
```

### scan_project_tags

프로젝트 전체에서 모든 @TAG 수집

```python
scan_project_tags(self)
```

### extract_tags

텍스트에서 @TAG 추출

```python
extract_tags(self, content)
```

### analyze_tag_integrity

단절된 @TAG 링크 분석

```python
analyze_tag_integrity(self)
```

### has_references

태그가 다른 태그와 연결되어 있는지 확인

```python
has_references(self, tag, all_tags)
```

### generate_repair_preview

수리 미리보기 생성

```python
generate_repair_preview(self, missing_links)
```

### extract_requirements_from_tag

태그에서 요구사항 정보 추출

```python
extract_requirements_from_tag(self, tag)
```

### get_tag_category

태그 타입의 카테고리 반환

```python
get_tag_category(self, tag_type)
```

### estimate_task_count

태그 기반 예상 작업 개수

```python
estimate_task_count(self, source)
```

### create_design_from_template

DESIGN 템플릿으로부터 문서 생성

```python
create_design_from_template(self, item)
```

### create_tasks_from_design

DESIGN으로부터 TASKS 문서 생성

```python
create_tasks_from_design(self, item)
```

### update_traceability_index

traceability.json 갱신

```python
update_traceability_index(self)
```

### auto_repair_tags

자동 수리 실행

```python
auto_repair_tags(self, dry_run)
```

## Classes

### TagRepairer
