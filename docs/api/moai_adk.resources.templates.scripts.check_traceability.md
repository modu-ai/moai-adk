# moai_adk.resources.templates.scripts.check_traceability

MoAI-ADK Traceability Checker
16-Core TAG 시스템의 추적성 검증

## Functions

### print_report

보고서 출력

```python
print_report(report, verbose)
```

### main

```python
main()
```

### __init__

```python
__init__(self, project_root)
```

### scan_all_tags

프로젝트 전체에서 모든 @TAG 수집

```python
scan_all_tags(self)
```

### extract_tags_from_content

텍스트에서 @TAG 추출

```python
extract_tags_from_content(self, content)
```

### validate_tag_naming

태그 네이밍 규칙 검증

```python
validate_tag_naming(self, tags)
```

### check_traceability_chains

추적성 체인 검증

```python
check_traceability_chains(self, tags)
```

### check_consistency

일관성 검증

```python
check_consistency(self, tags)
```

### generate_report

전체 추적성 보고서 생성

```python
generate_report(self)
```

## Classes

### TraceabilityChecker
