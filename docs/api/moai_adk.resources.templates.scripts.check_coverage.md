# moai_adk.resources.templates.scripts.check_coverage

MoAI-ADK Test Coverage Checker v0.1.12  
테스트 커버리지 측정 및 임계값 검증

이 스크립트는 프로젝트의 테스트 커버리지를:
- pytest-cov를 사용하여 정확히 측정
- 최소 80% 임계값 검증 
- 미커버 코드 위치 상세 리포트
- HTML 커버리지 리포트 생성
- 개발 가이드 5원칙 중 Testing 원칙 준수 확인

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

### load_coverage_config

커버리지 설정 로드

```python
load_coverage_config(self)
```

### detect_test_framework

사용 중인 테스트 프레임워크 감지

```python
detect_test_framework(self)
```

### has_pytest

pytest 사용 가능 여부 확인

```python
has_pytest(self)
```

### has_pytest_cov

pytest-cov 사용 가능 여부 확인

```python
has_pytest_cov(self)
```

### run_pytest_coverage

pytest-cov로 커버리지 측정

```python
run_pytest_coverage(self)
```

### parse_coverage_json

coverage.json 파일 파싱

```python
parse_coverage_json(self, coverage_data)
```

### parse_coverage_output

pytest 출력에서 커버리지 정보 추출

```python
parse_coverage_output(self, output)
```

### parse_line_ranges

라인 범위 문자열을 라인 번호 리스트로 변환

```python
parse_line_ranges(self, line_ranges)
```

### run_unittest_coverage

unittest + coverage.py로 커버리지 측정

```python
run_unittest_coverage(self)
```

### analyze_coverage_quality

커버리지 품질 분석

```python
analyze_coverage_quality(self, result)
```

### generate_report

커버리지 리포트 생성

```python
generate_report(self, result, analysis)
```

### run_coverage_check

전체 커버리지 검사 실행

```python
run_coverage_check(self)
```

## Classes

### CoverageResult

커버리지 결과 구조

### FileCoverage

파일별 커버리지 정보

### CoverageChecker

테스트 커버리지 검사기
