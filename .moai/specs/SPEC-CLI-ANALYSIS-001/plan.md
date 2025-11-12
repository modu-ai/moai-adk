# @SPEC:CLI-ANALYSIS-001 구현 계획

## 프로젝트 개요

**목표**: CLI 분석 명령어(`moai analyze`) 개발로 프로젝트 상태를 5개 카테고리로 종합 분석

**범위**:
- 신규 `analyze` 명령어 구현
- 분석 엔진 (`analyzer.py`) 개발
- 메트릭 수집 모듈 (`metrics.py`) 개발
- 리포팅 및 내보내기 모듈 (`reporter.py`, `exporters.py`) 개발
- 포괄적 테스트 (90%+ 커버리지)

**의존성**:
- SPEC-CLI-001 (doctor/status/restore 기본 기능) - **완료**
- SPEC-TEST-COVERAGE-001 (테스트 커버리지 85%+) - **완료**

---

## 구현 마일스톤 (우선순위 기반)

### Phase 1: 핵심 분석 엔진 개발 (높음)

**목표**: 5개 분석 카테고리별 메트릭 수집 및 점수화 엔진 구축

#### 1.1 데이터 모델 정의 (`src/moai_adk/core/analysis/models.py`)

```python
# Pydantic 모델 정의
- AnalysisResult (카테고리별 분석 결과)
- ProjectAnalysis (프로젝트 전체 분석)
- RecommendationItem (추천사항)
- CacheMetadata (캐시 정보)
```

**결과물**: `models.py` (~100 LOC)

#### 1.2 메트릭 수집 모듈 (`src/moai_adk/core/analysis/metrics.py`)

```python
class StructureMetrics:
    - analyze_directory_structure()
    - detect_large_files()
    - calculate_module_depth()

class DependencyMetrics:
    - build_dependency_graph()
    - detect_circular_dependencies()
    - calculate_coupling_cohesion()

class QualityMetrics:
    - get_test_coverage()
    - verify_tag_chain_completeness()
    - count_linter_warnings()
    - analyze_docstring_coverage()
    - detect_high_complexity_functions()

class PerformanceMetrics:
    - measure_init_time()
    - measure_doctor_time()
    - measure_test_time()

class CompatibilityMetrics:
    - check_python_version()
    - verify_dependency_compatibility()
    - check_platform_compatibility()
```

**결과물**: `metrics.py` (~400 LOC)

#### 1.3 분석 엔진 (`src/moai_adk/core/analysis/analyzer.py`)

```python
class ProjectAnalyzer:
    - analyze_structure() → AnalysisResult
    - analyze_dependencies() → AnalysisResult
    - analyze_quality() → AnalysisResult
    - analyze_performance() → AnalysisResult
    - analyze_compatibility() → AnalysisResult
    - generate_overall_analysis() → ProjectAnalysis
    - generate_recommendations() → List[str]
```

**결과물**: `analyzer.py` (~250 LOC)

**테스트**: `tests/unit/test_analyzer.py` (~300 테스트)

### Phase 2: CLI 명령어 구현 (높음)

**목표**: `moai analyze` 명령어 및 옵션 구현

#### 2.1 분석 명령어 (`src/moai_adk/cli/commands/analyze.py`)

```python
@click.command()
@click.option('--deep', is_flag=True)
@click.option('--format', type=click.Choice(['table', 'json', 'csv']))
@click.option('--export', type=click.Path())
@click.option('--category', type=click.Choice([...]))
@click.option('--compare', is_flag=True)
@click.option('--threshold', type=int)
@click.option('--ci', is_flag=True)
def analyze(ctx, deep, format, export, category, compare, threshold, ci):
    # 분석 실행
    # 결과 포맷팅 및 표시
    # 파일 내보내기 (선택사항)
```

**결과물**: `analyze.py` (~150 LOC)

**테스트**: `tests/integration/test_analyze_command.py` (~100 테스트)

### Phase 3: 리포팅 및 내보내기 (중간)

**목표**: 다양한 형식으로 분석 결과 제시

#### 3.1 리포팅 모듈 (`src/moai_adk/core/analysis/reporter.py`)

```python
class AnalysisReporter:
    - format_as_table() → str
    - format_as_plain_text() → str
    - add_progress_bar()
    - highlight_critical_issues()
    - display_recommendations()
```

**결과물**: `reporter.py` (~200 LOC)

#### 3.2 내보내기 모듈 (`src/moai_adk/core/analysis/exporters.py`)

```python
class JSONExporter:
    - export(analysis: ProjectAnalysis, filepath: str) → None

class CSVExporter:
    - export(analysis: ProjectAnalysis, filepath: str) → None

class TextExporter:
    - export(analysis: ProjectAnalysis, filepath: str) → None
```

**결과물**: `exporters.py` (~150 LOC)

**테스트**: `tests/unit/test_exporters.py` (~80 테스트)

### Phase 4: 캐싱 및 최적화 (중간)

**목표**: 분석 성능 최적화 및 캐싱 구현

#### 4.1 캐싱 모듈 (`src/moai_adk/core/analysis/cache.py`)

```python
class AnalysisCache:
    - get_cached_analysis() → Optional[ProjectAnalysis]
    - cache_analysis(analysis: ProjectAnalysis) → None
    - invalidate_cache() → None
    - is_cache_valid() → bool
```

**결과물**: `cache.py` (~120 LOC)

**테스트**: `tests/unit/test_cache.py` (~50 테스트)

### Phase 5: 테스트 강화 (높음)

**목표**: 90%+ 테스트 커버리지 달성

#### 5.1 단위 테스트 (Unit Tests)

```
tests/unit/
├── test_analyzer.py           (~300 테스트)
├── test_metrics.py            (~200 테스트)
├── test_exporters.py          (~80 테스트)
└── test_cache.py              (~50 테스트)
```

**목표**: 각 모듈별 95%+ 커버리지

#### 5.2 통합 테스트 (Integration Tests)

```
tests/integration/
├── test_analyze_command.py    (~100 테스트)
├── test_analysis_workflow.py  (~50 테스트)
```

**목표**: 전체 워크플로우 검증

**결과물**: 총 ~800개 테스트, 90%+ 커버리지

### Phase 6: 문서화 (낮음)

**목표**: 분석 기능 사용 가이드 및 아키텍처 문서화

#### 6.1 사용 가이드 (`.moai/docs/CLI-ANALYSIS.md`)

```
- 기본 사용법
- 분석 카테고리 설명
- 옵션별 예제
- CI/CD 통합 가이드
- 추천사항 해석 방법
```

#### 6.2 아키텍처 문서 (`.moai/docs/ARCHITECTURE-ANALYSIS.md`)

```
- 분석 엔진 설계
- 메트릭 계산 알고리즘
- 캐싱 전략
- 성능 최적화
```

---

## 기술 접근 방식

### 1. 분석 엔진 설계

#### 1.1 5개 분석기의 독립적 설계
- 각 분석기(`StructureMetrics`, `DependencyMetrics` 등)는 독립적으로 작동
- 병렬 실행 가능 (async 고려)
- 캐싱을 통한 성능 최적화

#### 1.2 점수 산정 알고리즘

```
각 카테고리별 점수 = (달성한 지표 수 / 전체 지표 수) × 100

범위별 상태:
- 90-100: EXCELLENT ✓
- 75-89: GOOD ✓ (주의 필요)
- 60-74: FAIR ⚠ (개선 권고)
- <60: CRITICAL ✗ (즉시 개선)
```

#### 1.3 추천사항 생성 로직

```python
RECOMMENDATIONS = {
    'quality': {
        'coverage': {
            '<85%': 'pytest --cov로 테스트 커버리지 측정 후 부족 부분 테스트 추가',
            '<90%': '테스트 커버리지를 90% 이상으로 향상시키기'
        },
        'tag_chain': {
            'broken': f'고아 TAG 수정: /alfred:1-plan {tag_id}',
            'orphan': '해당 SPEC에 대한 구현 또는 테스트 추가'
        }
    },
    # ... 더 많은 카테고리별 추천사항
}
```

### 2. 성능 최적화 전략

#### 2.1 캐싱 메커니즘

```
캐시 위치: .moai/cache/analysis/
캐시 키: {project_hash}_{analysis_type}
유효 기간: 파일 변경 감지 또는 24시간
```

#### 2.2 병렬 처리

```python
# 5개 분석기 병렬 실행 가능 (asyncio 또는 concurrent.futures)
# 각 분석기 시간: 1-2초 → 전체: 5-10초 (캐시 시 <1초)
```

### 3. Rich UI 통합

#### 3.1 테이블 형식

```
분석 카테고리   점수    상태          주요 지표
─────────────────────────────────────────────
구조 분석     85      OK ✓          모듈 깊이: 3
의존성 분석   92      EXCELLENT ✓   순환 참조: 0
품질 지표     78      GOOD ✓        커버리지: 85.6%
성능 분석     88      GOOD ✓        doctor: 2.3초
호환성 분석   95      EXCELLENT ✓   호환성: 100%
─────────────────────────────────────────────
전체 점수     87      GOOD
```

#### 3.2 색상 코딩

```
EXCELLENT (90-100): 초록색 + ✓
GOOD (75-89): 파란색 + ✓
FAIR (60-74): 노랑색 + ⚠
CRITICAL (<60): 빨강색 + ✗
```

### 4. CI/CD 모드

```
--ci 플래그 사용 시:
- Rich 색상 비활성화
- 아이콘 제거 (유니코드 비호환 환경 대응)
- JSON 구조화된 데이터 출력
```

---

## 아키텍처 설계 방향

### 모듈 구조

```
src/moai_adk/
├── cli/
│   └── commands/
│       └── analyze.py                    # CLI 진입점
│
└── core/
    └── analysis/                          # 신규 분석 모듈
        ├── __init__.py
        ├── models.py                     # Pydantic 데이터 모델
        ├── analyzer.py                   # 분석 엔진 (5개 분석기 조율)
        ├── metrics.py                    # 메트릭 수집 (StructureMetrics, DependencyMetrics 등)
        ├── reporter.py                   # Rich UI 리포팅
        ├── exporters.py                  # JSON/CSV 내보내기
        └── cache.py                      # 캐싱 메커니즘
```

### 의존성 관계

```
analyze.py (CLI)
    ↓
analyzer.py (분석 엔진)
    ├─→ metrics.py (메트릭 수집)
    ├─→ reporter.py (리포팅)
    ├─→ exporters.py (내보내기)
    └─→ cache.py (캐싱)

모든 모듈:
    ↓
models.py (데이터 모델)
```

### 클래스 다이어그램

```
ProjectAnalyzer
├─ StructureMetrics
├─ DependencyMetrics
├─ QualityMetrics
├─ PerformanceMetrics
└─ CompatibilityMetrics

AnalysisResult
├─ category: str
├─ score: float
├─ status: str
├─ details: Dict
└─ recommendations: List[str]

ProjectAnalysis
├─ project_name: str
├─ analysis_time: float
├─ categories: Dict[str, AnalysisResult]
├─ overall_score: float
└─ priority_issues: List[str]

AnalysisReporter
├─ format_as_table()
├─ format_as_plain_text()
└─ format_as_ci_output()

Exporters (Abstract)
├─ JSONExporter.export()
├─ CSVExporter.export()
└─ TextExporter.export()

AnalysisCache
├─ get_cached_analysis()
├─ cache_analysis()
└─ is_cache_valid()
```

---

## 위험 및 대응 계획

| 위험 항목 | 가능성 | 영향 | 대응 방안 |
|----------|--------|------|----------|
| **분석 시간 초과** (10초+) | 중간 | 높음 | 캐싱, 병렬 처리, 단계적 분석 |
| **의존성 그래프 복잡도** | 중간 | 중간 | networkx 최적화, 캐싱 |
| **CI/CD 환경 호환성** | 낮음 | 높음 | 충분한 --ci 모드 테스트 |
| **캐시 무효화 실패** | 낮음 | 중간 | 파일 변경 감지 메커니즘 강화 |
| **메모리 부하** (대규모 프로젝트) | 낮음 | 중간 | 스트리밍 처리, 청킹 |

---

## 테스트 전략

### RED-GREEN-REFACTOR 사이클

#### RED Phase: 테스트 작성
```python
# tests/unit/test_analyzer.py
def test_analyze_structure_returns_valid_score():
    # Given: 표준 프로젝트 구조
    # When: analyzer.analyze_structure() 호출
    # Then: 0-100 범위의 점수 반환
    pass

def test_analyze_dependencies_detects_circular_refs():
    # Given: 순환 참조가 있는 모듈 구조
    # When: analyzer.analyze_dependencies() 호출
    # Then: 순환 참조 목록 반환
    pass

# ... 총 ~800개 테스트
```

#### GREEN Phase: 최소 구현

#### REFACTOR Phase: 코드 개선
- 중복 제거
- 성능 최적화
- 캐싱 추가

### 커버리지 목표

```
- Unit Tests: 95%+ (models, metrics, exporters, cache)
- Integration Tests: 90%+ (analyze_command, workflows)
- 전체 커버리지: 90%+
```

---

## 구현 체크리스트

### 1단계: 데이터 모델 및 메트릭
- [ ] models.py 구현 및 테스트 (RED-GREEN-REFACTOR)
- [ ] metrics.py 구현 및 테스트 (각 분석기별)
- [ ] 메트릭 정확성 검증

### 2단계: 분석 엔진
- [ ] analyzer.py 구현 및 테스트
- [ ] 5개 분석기 통합 테스트
- [ ] 점수 산정 알고리즘 검증
- [ ] 추천사항 생성 로직 검증

### 3단계: CLI 명령어
- [ ] analyze.py 구현
- [ ] 옵션 파싱 및 검증
- [ ] 기본 분석 실행 테스트
- [ ] --deep 모드 테스트

### 4단계: 리포팅 및 내보내기
- [ ] reporter.py 구현
- [ ] exporters.py (JSON, CSV) 구현
- [ ] 포맷 검증 테스트
- [ ] 파일 저장 테스트

### 5단계: 캐싱 및 최적화
- [ ] cache.py 구현
- [ ] 캐시 유효성 검증
- [ ] 성능 측정 및 최적화
- [ ] 병렬 처리 구현 (선택사항)

### 6단계: 테스트 강화
- [ ] 단위 테스트 ~600개
- [ ] 통합 테스트 ~150개
- [ ] 커버리지 90%+ 달성
- [ ] 성능 테스트 (분석 시간 < 10초)

### 7단계: 문서화
- [ ] `.moai/docs/CLI-ANALYSIS.md` 작성
- [ ] `.moai/docs/ARCHITECTURE-ANALYSIS.md` 작성
- [ ] README.md analyze 명령어 추가
- [ ] CHANGELOG.md 업데이트

### 8단계: 검수 및 배포
- [ ] 코드 리뷰 (Ruff, mypy, pylint)
- [ ] 전체 워크플로우 검증
- [ ] CI/CD 파이프라인 통과
- [ ] 배포 및 릴리스 노트 작성

---

## 예상 결과물 요약

| 항목 | 수량 | 라인 수 |
|------|------|--------|
| Python 모듈 | 7개 | ~1,250 LOC |
| 테스트 파일 | 6개 | ~850 테스트 |
| 테스트 코드 | - | ~2,000 LOC |
| 문서 | 2개 | ~500 라인 |
| **합계** | - | **~3,750 LOC** |

### 테스트 커버리지
- analyzer.py: 95%+
- metrics.py: 95%+
- exporters.py: 95%+
- analyze.py (CLI): 90%+
- **전체: 90%+**

---

## 일정 추정치 (참고용, 시간 미포함)

```
1. 데이터 모델 정의           → 완료
2. 메트릭 수집 모듈          → 완료
3. 분석 엔진 개발           → 완료
4. CLI 명령어 구현          → 완료
5. 리포팅 및 내보내기       → 완료
6. 캐싱 및 최적화          → 완료
7. 포괄적 테스트           → 완료
8. 문서화                 → 완료
9. 코드 리뷰 & 최적화      → 완료
10. 배포 및 릴리스         → 완료
```

---

## 성공 기준

1. ✅ analyze 명령어 정상 작동 (`moai analyze`)
2. ✅ 5개 분석 카테고리 완전 구현 및 검증
3. ✅ 90%+ 테스트 커버리지 달성
4. ✅ 분석 시간 < 10초 (캐시 시 < 1초)
5. ✅ JSON/CSV 내보내기 정상 작동
6. ✅ CI/CD 모드 호환성 검증
7. ✅ 모든 Accept

 Criteria 통과
8. ✅ 문서 완성 및 업데이트
