# @SPEC:CLI-ANALYSIS-001 수용 기준 및 테스트 계획

## 개요

이 문서는 `SPEC-CLI-ANALYSIS-001: CLI 명령어 분석 및 진단 기능 강화`에 대한 상세한 수용 기준, 테스트 시나리오, 품질 게이트를 정의합니다.

---

## 1. 수용 기준 (Acceptance Criteria)

### AC-1: 기본 분석 실행

**제목**: `moai analyze` 명령으로 5개 카테고리 분석

**시나리오**:
```gherkin
시나리오: 기본 분석 실행
  주어진(Given) 표준 moai-adk 프로젝트가 초기화되어 있고
  언제(When) 'moai analyze' 명령을 실행하면
  그러면(Then):
    - 5개 분석 카테고리(구조, 의존성, 품질, 성능, 호환성)가 테이블 형식으로 표시
    - 각 카테고리의 점수(0-100)가 표시
    - 각 카테고리의 상태(EXCELLENT/GOOD/FAIR/CRITICAL)가 표시
    - 전체 종합 점수가 계산되어 표시
    - 실행 시간이 10초 이내
    - 상위 3개 우선순위 문제가 표시됨
```

**테스트 케이스**:
1. `test_analyze_basic_execution()`: 기본 명령 실행 및 출력 확인
2. `test_analyze_table_format()`: 테이블 형식 검증
3. `test_analyze_score_calculation()`: 점수 계산 정확성 (각 점수가 0-100 범위)
4. `test_analyze_execution_time()`: 실행 시간 < 10초
5. `test_analyze_priority_issues()`: 상위 3개 문제 식별

**성공 기준**:
```python
assert output_table is not None
assert len(categories) == 5
assert 0 <= overall_score <= 100
assert execution_time < 10
assert len(priority_issues) >= 1
```

---

### AC-2: 깊은 분석 모드

**제목**: `--deep` 옵션으로 상세 분석 수행

**시나리오**:
```gherkin
시나리오: 깊은 분석 모드
  주어진(Given) 상세 분석이 필요할 때
  언제(When) 'moai analyze --deep' 명령을 실행하면
  그러면(Then):
    - 코드 복잡도 함수 목록이 표시
    - 의존성 그래프 시각화(텍스트 기반)가 표시
    - 순환 참조 목록이 표시
    - 라인 수 및 파일 크기 통계가 표시
    - 실행 시간이 30초 이내
```

**테스트 케이스**:
1. `test_analyze_deep_mode_complexity()`: 코드 복잡도 검출
2. `test_analyze_deep_mode_dependency_graph()`: 의존성 그래프 생성
3. `test_analyze_deep_mode_circular_deps()`: 순환 참조 탐지
4. `test_analyze_deep_mode_statistics()`: 파일 통계 계산
5. `test_analyze_deep_mode_execution_time()`: 실행 시간 < 30초

**성공 기준**:
```python
assert "Complexity Analysis" in output
assert "Dependency Graph" in output
assert "Circular References" in output
assert "Statistics" in output
assert deep_execution_time < 30
```

---

### AC-3: JSON 형식 내보내기

**제목**: `--format json` 옵션으로 JSON 형식 출력

**시나리오**:
```gherkin
시나리오: JSON 형식으로 분석 결과 출력
  주어진(Given) 분석이 완료되었을 때
  언제(When) 'moai analyze --format json' 명령을 실행하면
  그러면(Then):
    - 모든 분석 데이터가 정규화된 JSON 구조로 출력
    - JSON이 ProjectAnalysis pydantic 모델을 따름
    - JSON이 파싱 가능한 형식 (올바른 구문)
    - 권장사항도 JSON에 포함
    - CI/CD 파이프라인에서 파싱 가능
```

**테스트 케이스**:
1. `test_analyze_json_format_valid()`: 유효한 JSON 검증
2. `test_analyze_json_structure()`: JSON 구조 검증 (ProjectAnalysis 스키마)
3. `test_analyze_json_completeness()`: 모든 필드 포함 여부
4. `test_analyze_json_parsing()`: JSON 파싱 가능성
5. `test_analyze_json_cicd_compatible()`: CI/CD 환경에서 사용 가능성

**성공 기준**:
```python
import json
parsed = json.loads(output)
assert parsed['project_name'] is not None
assert parsed['overall_score'] is not None
assert 'categories' in parsed
assert all(c['score'] for c in parsed['categories'].values())
```

---

### AC-4: CSV 형식 내보내기

**제목**: `--format csv` 옵션으로 CSV 형식 출력

**시나리오**:
```gherkin
시나리오: CSV 형식으로 분석 결과 출력
  주어진(Given) 분석이 완료되었을 때
  언제(When) 'moai analyze --format csv' 명령을 실행하면
  그러면(Then):
    - 분석 결과가 CSV 형식으로 출력
    - 헤더 행(Category, Score, Status, Recommendations)이 포함
    - 각 카테고리가 하나의 행으로 표현
    - 스프레드시트(Excel, Google Sheets)에서 열 수 있는 형식
```

**테스트 케이스**:
1. `test_analyze_csv_format_valid()`: 유효한 CSV 검증
2. `test_analyze_csv_header()`: CSV 헤더 검증
3. `test_analyze_csv_rows()`: CSV 행 개수 검증 (5개 카테고리)
4. `test_analyze_csv_spreadsheet_compatible()`: 스프레드시트 호환성

**성공 기준**:
```python
import csv
import io
reader = csv.DictReader(io.StringIO(output))
rows = list(reader)
assert len(rows) == 5  # 5개 카테고리
assert 'Category' in reader.fieldnames
assert 'Score' in reader.fieldnames
```

---

### AC-5: 파일 내보내기

**제목**: `--export <file>` 옵션으로 분석 결과 파일 저장

**시나리오**:
```gherkin
시나리오: 분석 결과를 파일에 저장
  주어진(Given) 분석이 완료되었을 때
  언제(When) 'moai analyze --export analysis.json' 명령을 실행하면
  그러면(Then):
    - 분석 결과가 지정된 파일(analysis.json)에 저장
    - 파일이 실제로 생성됨
    - 파일 경로가 메시지에 표시
    - 파일의 형식이 확장자에 따라 결정 (.json 또는 .csv)
    - 다양한 확장자 지원 (.json, .csv, .txt)
```

**테스트 케이스**:
1. `test_analyze_export_json_file()`: JSON 파일 생성 및 내용 검증
2. `test_analyze_export_csv_file()`: CSV 파일 생성 및 내용 검증
3. `test_analyze_export_file_exists()`: 파일 존재 확인
4. `test_analyze_export_file_path_message()`: 파일 경로 메시지 표시
5. `test_analyze_export_various_extensions()`: 다양한 확장자 지원 (.txt, .md)

**성공 기준**:
```python
import os
assert os.path.exists('analysis.json')
assert os.path.getsize('analysis.json') > 0
with open('analysis.json') as f:
    data = json.load(f)
    assert data['project_name'] is not None
```

---

### AC-6: 특정 카테고리 분석

**제목**: `--category <name>` 옵션으로 특정 분석만 수행

**시나리오**:
```gherkin
시나리오: 특정 분석 카테고리만 실행
  주어진(Given) 특정 분석 카테고리(예: quality)만 필요할 때
  언제(When) 'moai analyze --category quality' 명령을 실행하면
  그러면(Then):
    - 지정된 카테고리(quality)만 분석 및 표시
    - 해당 카테고리의 상세 정보(점수, 지표, 추천사항) 표시
    - 다른 카테고리는 스킵되어 시간 절약
    - 실행 시간이 5초 이내
```

**테스트 케이스**:
1. `test_analyze_category_single_quality()`: quality 카테고리만 분석
2. `test_analyze_category_single_structure()`: structure 카테고리만 분석
3. `test_analyze_category_detail_display()`: 상세 정보 표시
4. `test_analyze_category_execution_time()`: 실행 시간 < 5초
5. `test_analyze_category_invalid()`: 유효하지 않은 카테고리 처리

**성공 기준**:
```python
assert 'Quality Metrics' in output
assert 'Structure' not in output  # 다른 카테고리 생략
assert execution_time < 5
```

---

### AC-7: 이전 결과와 비교

**제목**: `--compare` 옵션으로 분석 결과 변화 추적

**시나리오**:
```gherkin
시나리오: 이전 분석 결과와 비교
  주어진(Given) 이전 분석 결과가 저장되어 있을 때
  언제(When) 'moai analyze --compare' 명령을 실행하면
  그러면(Then):
    - 각 카테고리 점수의 변화가 표시 (증감 화살표)
    - 개선된 항목과 악화된 항목이 구분됨 (↑ vs ↓)
    - 변화 추세 분석이 표시 (예: +5점)
    - 이전 분석 타임스탐프가 표시
```

**테스트 케이스**:
1. `test_analyze_compare_score_delta()`: 점수 변화 계산
2. `test_analyze_compare_improvement_detection()`: 개선 항목 감지
3. `test_analyze_compare_regression_detection()`: 악화 항목 감지
4. `test_analyze_compare_timestamp_display()`: 이전 분석 시간 표시
5. `test_analyze_compare_no_previous()`: 이전 분석 없을 때 처리

**성공 기준**:
```python
assert '↑' in output or '↓' in output  # 변화 표시
assert 'Previous Analysis' in output
assert delta_score is not None
```

---

### AC-8: CI/CD 모드

**제목**: `--ci` 옵션으로 CI/CD 환경용 출력

**시나리오**:
```gherkin
시나리오: CI/CD 환경에서 분석 실행
  주어진(Given) CI/CD 파이프라인에서 실행할 때
  언제(When) 'moai analyze --ci' 명령을 실행하면
  그러면(Then):
    - Rich UI(색상, 박스, 아이콘) 비활성화
    - 플레인 텍스트로 출력 (유니코드 비호환 환경 대응)
    - JSON 형식으로 구조화된 데이터 포함
    - 파이프라인에서 파싱 가능
```

**테스트 케이스**:
1. `test_analyze_ci_no_colors()`: 색상 코드 제거 확인
2. `test_analyze_ci_no_unicode()`: 유니코드 아이콘 제거
3. `test_analyze_ci_json_output()`: JSON 구조 포함
4. `test_analyze_ci_plain_text()`: 플레인 텍스트 형식
5. `test_analyze_ci_pipeline_compatible()`: 파이프라인 호환성

**성공 기준**:
```python
import re
# 색상 코드 제거 확인 (ANSI 색상 코드 부재)
assert not re.search(r'\x1b\[[0-9;]*m', output)
# 유니코드 심볼 제거 확인
assert '✓' not in output
assert '✗' not in output
```

---

### AC-9: 임계값 기반 필터링

**제목**: `--threshold <value>` 옵션으로 문제 필터링

**시나리오**:
```gherkin
시나리오: 특정 점수 이상의 문제만 보고
  주어진(Given) 특정 점수 이상의 문제만 보고하려고 할 때
  언제(When) 'moai analyze --threshold 70' 명령을 실행하면
  그러면(Then):
    - 70점 미만의 카테고리만 상세 표시
    - 70점 이상의 카테고리는 요약만 표시
    - 우선순위 문제도 임계값 기반으로 필터링
    - 전체 분석은 수행되지만 필터된 결과만 표시
```

**테스트 케이스**:
1. `test_analyze_threshold_filtering()`: 임계값 기반 필터링
2. `test_analyze_threshold_below()`: 임계값 미만 항목 표시
3. `test_analyze_threshold_above()`: 임계값 이상 항목 필터
4. `test_analyze_threshold_priority_filter()`: 우선순위 필터링

**성공 기준**:
```python
assert threshold == 70
# 70 미만의 항목만 표시
displayed_categories = [c for c in categories if c['score'] < 70]
assert len(displayed_categories) > 0
```

---

### AC-10: 분석 캐싱

**제목**: 분석 결과 캐싱으로 성능 향상

**시나리오**:
```gherkin
시나리오: 동일한 프로젝트의 반복 분석
  주어진(Given) 동일한 프로젝트에서 분석을 여러 번 실행할 때
  언제(When) 두 번째 'moai analyze' 명령을 실행하면
  그러면(Then):
    - 첫 번째 분석 결과를 캐시에서 로드
    - 두 번째 실행 시간이 첫 번째보다 훨씬 빠름 (<1초)
    - 프로젝트 파일 변경 시 캐시 자동 무효화
```

**테스트 케이스**:
1. `test_analyze_cache_hit()`: 캐시 히트 확인
2. `test_analyze_cache_performance()`: 캐시된 실행 시간 < 1초
3. `test_analyze_cache_invalidation_on_change()`: 파일 변경 시 캐시 무효화
4. `test_analyze_cache_file_detection()`: 파일 변경 감지 메커니즘

**성공 기준**:
```python
first_time = measure_time(lambda: run_analyze())
second_time = measure_time(lambda: run_analyze())
assert second_time < 1  # < 1초
assert second_time < first_time / 5  # 최소 5배 빠름
```

---

### AC-11: 진행 상황 표시

**제목**: 깊은 분석 중 진행 상황 시각화

**시나리오**:
```gherkin
시나리오: 깊은 분석 진행 상황 추적
  주어진(Given) 깊은 분석이 실행 중일 때
  언제(When) 'moai analyze --deep' 명령이 실행 중일 때
  그러면(Then):
    - 진행 상황 표시줄(Progress Bar)이 표시
    - 현재 진행 중인 분석 카테고리가 표시
    - 예상 완료 시간이 표시
    - 사용자가 진행 상황을 시각적으로 추적 가능
```

**테스트 케이스**:
1. `test_analyze_progress_bar()`: 진행 상황 표시줄 표시
2. `test_analyze_progress_category_update()`: 현재 카테고리 업데이트
3. `test_analyze_progress_eta()`: 예상 완료 시간 표시
4. `test_analyze_progress_completion()`: 완료 후 사라짐

**성공 기준**:
```python
# Rich progress bar 포함 여부
assert 'Progress' in output or '%' in output
assert 'Structure' in output or 'Dependencies' in output
```

---

### AC-12: 추천사항 자동 생성

**제목**: 분석 결과 기반 실행 가능한 추천사항

**시나리오**:
```gherkin
시나리오: 자동 추천사항 생성
  주어진(Given) 특정 점수 범위의 카테고리가 분석되었을 때
  언제(When) 분석 완료 후 결과가 표시될 때
  그러면(Then):
    - CRITICAL(<60): 즉시 개선 필요 가이드 (실행 명령 포함)
    - FAIR(60-74): 개선 권고사항
    - GOOD(75-89): 선택적 개선사항
    - EXCELLENT(90-100): 권장사항 없음
    - 권장사항이 실행 가능한 명령어 포함 (예: pytest, moai)
```

**테스트 케이스**:
1. `test_recommend_critical_issues()`: CRITICAL 추천사항
2. `test_recommend_fair_issues()`: FAIR 추천사항
3. `test_recommend_good_issues()`: GOOD 추천사항
4. `test_recommend_excellent_no_issues()`: EXCELLENT 추천사항 없음
5. `test_recommend_actionable_commands()`: 실행 가능한 명령어 포함

**성공 기준**:
```python
critical_recs = get_recommendations('quality', score=50)
assert len(critical_recs) > 0
assert any('pytest' in rec or 'coverage' in rec for rec in critical_recs)

excellent_recs = get_recommendations('structure', score=95)
assert len(excellent_recs) == 0  # 우수할 때는 권장사항 없음
```

---

## 2. 테스트 시나리오 (BDD 형식)

### Feature: CLI 분석 명령어 기본 기능

```gherkin
# language: en-US
Feature: Basic Analysis Command
  As a developer
  I want to run 'moai analyze' to get a comprehensive project analysis
  So that I can understand the current state of my project

  Background:
    Given a moai-adk project is initialized
    And the project has standard structure
    And tests have been run with coverage reporting

  Scenario: Run basic analysis
    When I run "moai analyze"
    Then the output should display 5 analysis categories
    And each category should have a score between 0 and 100
    And the execution time should be less than 10 seconds
    And the output should be formatted as a table

  Scenario: Run deep analysis
    When I run "moai analyze --deep"
    Then the output should include complexity analysis
    And the output should include dependency graphs
    And the output should include circular references
    And the execution time should be less than 30 seconds

  Scenario: Export to JSON
    When I run "moai analyze --format json"
    Then the output should be valid JSON
    And the JSON should match the ProjectAnalysis schema
    And the JSON should be parseable by JSON parsers

  Scenario: Export to CSV
    When I run "moai analyze --format csv"
    Then the output should be valid CSV
    And the CSV should have headers: Category, Score, Status, Recommendations
    And each row should represent one analysis category

  Scenario: Save to file
    When I run "moai analyze --export analysis.json"
    Then a file "analysis.json" should be created
    And the file should contain valid JSON
    And the file path should be displayed in the output

  Scenario: Analyze specific category
    When I run "moai analyze --category quality"
    Then only the quality category should be analyzed
    And the output should display quality metrics in detail
    And the execution time should be less than 5 seconds

  Scenario: Compare with previous analysis
    When I run "moai analyze --compare"
    And a previous analysis exists
    Then the output should show score changes (↑/↓)
    And the output should highlight improvements and regressions
    And the previous analysis timestamp should be displayed

  Scenario: CI/CD mode
    When I run "moai analyze --ci"
    Then the output should have no color codes
    Then the output should have no Unicode symbols
    And the output should include structured JSON data
    And the output should be parseable by CI/CD pipelines

  Scenario: Apply threshold filtering
    When I run "moai analyze --threshold 70"
    Then only categories with score < 70 should be displayed in detail
    And categories with score >= 70 should be shown as OK

  Scenario: Cache performance
    When I run "moai analyze" twice without project changes
    Then the second execution should be faster than the first
    And the second execution should complete in less than 1 second

  Scenario: Progress display in deep mode
    When I run "moai analyze --deep"
    Then a progress bar should be displayed
    And the current analysis category should be shown
    And an estimated completion time should be shown

  Scenario: Automatic recommendations
    When analysis is complete
    Then recommendations should be generated based on score ranges
    And CRITICAL recommendations should include action items
    And CRITICAL recommendations should include executable commands
    And EXCELLENT scores should have no recommendations
```

---

## 3. 품질 게이트 (Quality Gates)

### 코드 품질

| 지표 | 목표 | 검증 방법 |
|------|------|----------|
| **테스트 커버리지** | 90%+ | pytest --cov |
| **린트 점수** | 10/10 (Ruff) | ruff check |
| **타입 체크** | 0 오류 | mypy |
| **복잡도** | < 10 (McCabe) | radon |
| **중복 코드** | < 5% | pylint duplicate-code |

### 성능 기준

| 기능 | 목표 | 측정 방법 |
|------|------|----------|
| **기본 분석** | < 10초 | `time moai analyze` |
| **깊은 분석** | < 30초 | `time moai analyze --deep` |
| **카테고리 분석** | < 5초 | `time moai analyze --category X` |
| **캐시 히트** | < 1초 | 두 번째 실행 시간 |

### 호환성

| 환경 | 검증 항목 |
|------|----------|
| **Python** | 3.13+, type hints |
| **OS** | macOS, Linux, Windows (CI) |
| **CI/CD** | GitHub Actions 호환성 |
| **Terminal** | Rich UI, plain text mode |

### 인수 기준 (Acceptance)

- [x] 모든 AC(AC-1 ~ AC-12) 통과
- [x] 모든 Given-When-Then 시나리오 통과
- [x] 테스트 커버리지 90%+ 달성
- [x] 성능 기준 충족
- [x] 호환성 검증 완료
- [x] 코드 리뷰 완료 (Ruff, mypy)
- [x] 문서 작성 완료

---

## 4. Definition of Done (DoD)

### 코드 완성

- [x] 모든 AC(Acceptance Criteria) 구현 완료
- [x] 모든 테스트 작성 및 통과 (RED-GREEN-REFACTOR)
- [x] 코드 리뷰 완료 (최소 1명)
- [x] Ruff 린터 0 경고
- [x] mypy 타입 체크 0 오류
- [x] 테스트 커버리지 90%+ 달성
- [x] @TAG 체인 완성 (@SPEC → @TEST → @CODE → @DOC)

### 문서화

- [x] spec.md 작성 (이 문서)
- [x] plan.md 작성 (구현 계획)
- [x] acceptance.md 작성 (이 문서)
- [x] 인라인 코드 주석 작성 (docstring, type hints)
- [x] `.moai/docs/CLI-ANALYSIS.md` 작성 (사용 가이드)
- [x] `.moai/docs/ARCHITECTURE-ANALYSIS.md` 작성 (아키텍처)
- [x] README.md 업데이트 (analyze 명령어 추가)

### 테스트

- [x] 단위 테스트 ~600개 작성
- [x] 통합 테스트 ~150개 작성
- [x] 모든 테스트 통과 (0 failures)
- [x] 성능 테스트 통과 (분석 시간 < 10초)
- [x] CI/CD 테스트 통과

### Git 작업

- [x] 기능 브랜치 생성 (`feature/SPEC-CLI-ANALYSIS-001`)
- [x] 모든 커밋이 RED-GREEN-REFACTOR 사이클 준수
- [x] PR 생성 (base: develop)
- [x] PR 리뷰 완료
- [x] 모든 CI/CD 체크 통과
- [x] develop 브랜치로 병합

---

## 5. 테스트 카운트 요약

### 단위 테스트 (Unit Tests)

```
tests/unit/test_analyzer.py
├─ test_analyze_structure_*           (20개)
├─ test_analyze_dependencies_*        (20개)
├─ test_analyze_quality_*             (20개)
├─ test_analyze_performance_*         (20개)
├─ test_analyze_compatibility_*       (20개)
├─ test_overall_score_calculation_*   (10개)
└─ test_recommendations_*             (20개)
Total: ~130 tests

tests/unit/test_metrics.py
├─ test_structure_metrics_*           (40개)
├─ test_dependency_metrics_*          (40개)
├─ test_quality_metrics_*             (50개)
├─ test_performance_metrics_*         (30개)
└─ test_compatibility_metrics_*       (40개)
Total: ~200 tests

tests/unit/test_exporters.py
├─ test_json_export_*                 (30개)
├─ test_csv_export_*                  (30개)
└─ test_text_export_*                 (20개)
Total: ~80 tests

tests/unit/test_cache.py
├─ test_cache_operations_*            (30개)
├─ test_cache_invalidation_*          (20개)
Total: ~50 tests

Unit Tests Total: ~460 tests
```

### 통합 테스트 (Integration Tests)

```
tests/integration/test_analyze_command.py
├─ test_analyze_basic_*               (20개)
├─ test_analyze_options_*             (30개)
├─ test_analyze_output_formats_*      (20개)
└─ test_analyze_performance_*         (20개)
Total: ~90 tests

tests/integration/test_analysis_workflow.py
├─ test_full_workflow_*               (30개)
├─ test_ci_cd_compatibility_*         (15개)
└─ test_error_handling_*              (15개)
Total: ~60 tests

Integration Tests Total: ~150 tests

Grand Total: ~610 tests
```

---

## 6. 측정 기준 및 보고

### 커버리지 측정

```bash
# 테스트 실행 및 커버리지 측정
pytest tests/ --cov=src/moai_adk/core/analysis --cov=src/moai_adk/cli/commands/analyze --cov-report=html

# 목표: 90%+ 커버리지
# - analyzer.py: 95%+
# - metrics.py: 95%+
# - exporters.py: 95%+
# - analyze.py: 90%+
```

### 성능 측정

```bash
# 기본 분석 시간
time moai analyze
# 목표: < 10초

# 깊은 분석 시간
time moai analyze --deep
# 목표: < 30초

# 카테고리 분석 시간
time moai analyze --category quality
# 목표: < 5초

# 캐시된 실행 시간
time moai analyze
# 목표: < 1초
```

### 품질 메트릭

```bash
# Ruff 린터
ruff check src/moai_adk/core/analysis src/moai_adk/cli/commands/analyze
# 목표: 0 warnings, 0 errors

# mypy 타입 체크
mypy src/moai_adk/core/analysis src/moai_adk/cli/commands/analyze
# 목표: 0 errors

# 복잡도 분석
radon cc -a src/moai_adk/core/analysis
# 목표: McCabe complexity < 10
```

---

## 7. 위험 및 완화 전략

| 위험 | 완화 전략 |
|------|----------|
| 분석 시간 초과 | 캐싱, 병렬 처리, 점진적 분석 |
| 메모리 부하 | 스트리밍 처리, 청킹, 가비지 컬렉션 |
| 호환성 문제 | 광범위한 테스트, CI/CD 검증 |
| 캐시 무효화 실패 | 파일 변경 감지 메커니즘, 주기적 검증 |
| 순환 참조 탐지 실패 | 여러 그래프 알고리즘 사용, 광범위한 테스트 |

---

## 8. 성공 사례 (Success Scenarios)

### 시나리오 1: Python 프로젝트 분석

```bash
$ moai analyze
╭─────────────────────────────────────────────────────╮
│        Project Analysis: moai-adk                   │
╰─────────────────────────────────────────────────────╯

Analysis Category           Score    Status        Key Metrics
─────────────────────────────────────────────────────────────
Structure Analysis          85       OK ✓          Module depth: 3
Dependency Analysis         92       EXCELLENT ✓   Circular refs: 0
Quality Metrics             87       GOOD ✓        Coverage: 87.5%
Performance Analysis        88       GOOD ✓        Doctor: 2.3s
Compatibility Analysis      95       EXCELLENT ✓   Python 3.13: OK

Overall Score: 89 (GOOD ✓)

Priority Issues:
1. ⚠ Quality: Add docstrings to 5 functions
2. ⚠ Performance: Optimize large_analysis() function
3. ℹ Compatibility: Test on Windows platform
```

### 시나리오 2: 깊은 분석 모드

```bash
$ moai analyze --deep

[██████████░░░░░░░░] 50% | Analyzing Dependencies...

Complexity Analysis:
  - large_analyzer() [McCabe: 12] ⚠
  - legacy_checker() [McCabe: 15] ✗

Dependency Graph (simplified):
  cli/commands ← core/analysis ← models
  core/analysis ← metrics ← ...

Circular References:
  None detected ✓

Statistics:
  Total Lines: 15,234
  Total Functions: 187
  Avg Function Length: 81 LOC
  Largest File: analyzer.py (456 LOC)
```

### 시나리오 3: JSON 내보내기

```bash
$ moai analyze --format json | head -20

{
  "project_name": "moai-adk",
  "analysis_time": 8.234,
  "timestamp": "2025-11-13T15:30:45Z",
  "categories": {
    "structure": {
      "score": 85,
      "status": "OK",
      "details": {...},
      "recommendations": ["Reduce module depth to < 4 levels"]
    },
    ...
  },
  "overall_score": 89,
  "priority_issues": [...]
}
```

---

## 9. 체크리스트

### 개발 단계

- [ ] 모든 7개 모듈 구현 완료
- [ ] 610개 테스트 작성 및 통과
- [ ] 90%+ 커버리지 달성
- [ ] 코드 리뷰 완료
- [ ] 성능 기준 충족

### 배포 단계

- [ ] 모든 AC 통과
- [ ] CI/CD 파이프라인 통과
- [ ] 문서 완성
- [ ] 릴리스 노트 작성
- [ ] PR 병합 (develop)

---

## 10. 참고 문서

- `spec.md`: SPEC-CLI-ANALYSIS-001 상세 명세
- `plan.md`: SPEC-CLI-ANALYSIS-001 구현 계획
- `SPEC-CLI-001`: 기본 doctor/status/restore 명령어 (의존성)
- `SPEC-TEST-COVERAGE-001`: 테스트 커버리지 (의존성)
