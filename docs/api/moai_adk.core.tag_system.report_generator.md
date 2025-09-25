# moai_adk.core.tag_system.report_generator

@FEATURE:REPORT-GENERATOR-001 - TAG 추적성 리포트 생성 최소 구현

Jinja2 템플릿 기반 Markdown/HTML 리포트 생성

## Functions

### __init__

리포트 생성기 초기화

Args:
    output_dir: 출력 디렉토리
    template_dir: 템플릿 디렉토리 (None이면 내장 템플릿 사용)

```python
__init__(self, output_dir, template_dir)
```

### generate_chain_matrix

TAG 체인 매트릭스 생성

Args:
    tags: TAG 목록

Returns:
    체인 매트릭스 딕셔너리

```python
generate_chain_matrix(self, tags)
```

### analyze_missing_connections

누락된 연결 분석

Args:
    tags: 분석할 TAG 목록

Returns:
    누락 분석 결과

```python
analyze_missing_connections(self, tags)
```

### generate_report

리포트 생성

Args:
    tags: TAG 목록
    format: 출력 형식
    title: 리포트 제목
    template_name: 사용할 템플릿 이름

Returns:
    생성된 리포트 내용

```python
generate_report(self, tags, format, title, template_name)
```

### calculate_implementation_coverage

구현 완료율 계산

Args:
    tags: TAG 목록

Returns:
    커버리지 정보

```python
calculate_implementation_coverage(self, tags)
```

### export_to_file

리포트를 파일로 내보내기

Args:
    tags: TAG 목록
    output_path: 출력 파일 경로
    format: 출력 형식

```python
export_to_file(self, tags, output_path, format)
```

### generate_summary_statistics

요약 통계 생성

Args:
    tags: TAG 목록

Returns:
    요약 통계

```python
generate_summary_statistics(self, tags)
```

### _setup_default_templates

기본 템플릿 설정

```python
_setup_default_templates(self)
```

### _get_category_group

카테고리 그룹 결정

```python
_get_category_group(self, category)
```

## Classes

### ReportFormat

리포트 출력 형식

### TraceabilityReport

추적성 리포트 데이터

### TagReportGenerator

TAG 추적성 리포트 생성기 최소 구현

TRUST 원칙 적용:
- Test First: 테스트 요구사항만 구현
- Readable: 명확한 리포트 생성 로직
- Unified: 리포트 생성 책임만 담당
