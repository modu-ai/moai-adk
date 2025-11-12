---
id: CLI-ANALYSIS-001
version: 0.1.0
status: pending
created: 2025-11-13
updated: 2025-11-13
author: @alfred-planner
priority: high
category: feature
labels:
  - cli
  - analysis
  - diagnostics
  - enhancement
depends_on:
  - CLI-001
  - TEST-COVERAGE-001
blocks: []
related_specs:
  - CLI-001
  - TEST-COVERAGE-001
scope:
  packages:
    - src/moai_adk/cli/commands
    - src/moai_adk/core/analysis
  files:
    - analyze.py
    - analyzer.py
    - metrics.py
    - reporter.py
    - exporters.py
---


## HISTORY

### v0.1.0 (2025-11-13)
- **INITIAL**: CLI 분석 기능 강화 명세 최초 작성
- **AUTHOR**: @alfred-planner
- **CONTEXT**: SPEC-CLI-001(doctor/status/restore) 기본 기능 완성 후 고급 분석 단계 진입
- **MOTIVATION**:
  - 현재 doctor는 도구 체인 검증만 수행 (깊은 의존성 분석 부재)
  - status는 정적 지표만 표시 (추세 분석, 성능 메트릭 부재)
  - 프로젝트 상태에 대한 고급 분석 및 추천사항 부재
  - 분석 결과를 구조화된 형식(JSON/CSV)으로 내보낼 수 없음
  - 프로젝트 구조, 의존성, 호환성을 통합 분석하는 기능 필요

## Environment (환경)

- **Python 버전**: 3.13+
- **CLI 프레임워크**: Click 8.1.x
- **Rich 라이브러리**: 13.7.x (터미널 UI)
- **기존 CLI 명령어**: doctor, status, restore, init, update, backup
- **의존성 그래프**: networkx 2.6.3+
- **데이터 검증**: pydantic 2.4.2+
- **기존 분석 도구**: ripgrep (rg), pytest-cov

## Assumptions (가정)

- Click 프레임워크의 옵션 확장이 가능하다 (`--format`, `--export`, `--deep`)
- 프로젝트 구조가 표준 moai-adk 레이아웃을 따른다
- 테스트 메트릭 데이터(pytest-cov JSON)가 수집 가능하다
- 분석 실행 시간은 기본 10초 이내, 깊은 분석 30초 이내로 제한한다
- 모든 CLI 개선은 기존 테스트 커버리지 85% 이상을 유지한다
- 분석 기능은 오프라인 환경에서도 작동해야 한다(네트워크 의존성 최소화)

## Requirements (요구사항)

### Ubiquitous (필수 요구사항)

- 시스템은 프로젝트 상태를 5개 카테고리로 분석해야 한다 (구조, 의존성, 품질, 성능, 호환성)
- 시스템은 analyze 명령어를 통해 고급 분석 기능을 제공해야 한다
- 시스템은 분석 결과를 구조화된 형식(JSON, CSV, Table)으로 제공해야 한다
- 시스템은 분석 결과 기반 추천사항을 자동 생성해야 한다
- 시스템은 모든 분석 출력에 Rich 라이브러리를 사용하여 가독성을 보장해야 한다

### Event-driven (이벤트 기반 요구사항)

- WHEN `moai analyze` 실행 시, 시스템은 프로젝트의 5개 분석 카테고리를 테이블 형식으로 표시해야 한다
- WHEN `moai analyze --deep` 실행 시, 시스템은 코드 복잡도, 의존성 깊이, 순환 참조 등 상세 분석을 제시해야 한다
- WHEN `moai analyze --format json` 실행 시, 시스템은 모든 분석 데이터를 JSON 형식으로 출력해야 한다
- WHEN `moai analyze --format csv` 실행 시, 시스템은 분석 데이터를 CSV 형식으로 출력해야 한다
- WHEN `moai analyze --export <file>` 실행 시, 시스템은 분석 결과를 파일로 저장해야 한다
- WHEN 분석 중 경고 또는 문제 발견 시, 시스템은 실행 가능한 해결 가이드를 함께 제시해야 한다
- WHEN 프로젝트에 TAG 체인이 완전하지 않을 때, 시스템은 고아 TAG를 감지하고 보고해야 한다

### State-driven (상태 기반 요구사항)

- WHILE 기본 분석 실행 중, 시스템은 현재 분석 중인 카테고리를 표시해야 한다
- WHILE 깊은 분석 실행 중, 시스템은 진행 상황 표시줄을 표시해야 한다
- WHILE 프로젝트가 복잡할 때, 시스템은 분석 시간을 초과하지 않도록 캐싱 메커니즘을 사용해야 한다
- WHILE CI/CD 환경에서 실행 시, 시스템은 Rich UI를 비활성화하고 플레인 텍스트로 출력해야 한다

### Optional (선택적 기능)

- WHERE `--category <name>` 옵션이 제공되면, 시스템은 특정 분석 카테고리만 실행할 수 있다
- WHERE `--compare` 옵션이 제공되면, 시스템은 이전 분석 결과와 비교 분석을 제시할 수 있다
- WHERE `--threshold` 옵션이 제공되면, 시스템은 임계값 기반 경고 필터링을 수행할 수 있다
- WHERE 프로젝트가 Team 모드일 때, analyze는 GitHub PR 품질 메트릭을 추가로 표시할 수 있다

### Constraints (제약사항)

- 기본 분석 실행 시간은 10초를 초과하지 않아야 한다
- 깊은 분석(`--deep`) 실행 시간은 30초를 초과하지 않아야 한다
- 모든 새로운 분석 기능은 90% 이상의 테스트 커버리지를 유지해야 한다
- 분석 기능은 오프라인 환경에서도 작동해야 한다
- analyze 명령어 추가는 doctor/status/restore 기존 기능과 호환되어야 한다

## Specifications (명세)

### 1. 분석 명령어 구조

```bash
moai analyze [CATEGORY] [OPTIONS]

옵션:
  --deep                    깊은 분석 수행 (코드 복잡도, 의존성 그래프 등)
  --format {table|json|csv} 출력 형식 (기본값: table)
  --export <file>          분석 결과를 파일로 저장
  --category <name>        특정 카테고리만 분석 (structure|deps|quality|perf|compat)
  --compare                이전 분석 결과와 비교
  --threshold <value>      경고 임계값 설정 (0-100)
  --ci                     CI/CD 모드 (Rich UI 비활성화)
```

### 2. 분석 카테고리 (5개)

#### 2.1 구조 분석 (Structure Analysis)

```python
# 분석 항목:
- 프로젝트 디렉토리 구조 검증
- 모듈 조직 (src/, tests/, docs/) 확인
- 패키지 계층 깊이 분석
- 파일 크기 및 라인 수 통계
- 대형 파일 식별 (>500 LOC)
```

**점수 산정**: 표준 구조 준수도 기반 (0-100)

#### 2.2 의존성 분석 (Dependency Analysis)

```python
# 분석 항목:
- 내부 모듈 간 의존성 맵핑
- 순환 참조(Circular Dependency) 탐지
- 의존성 그래프 깊이 분석
- 고립된 모듈 식별
- 의존성 강도 측정 (응집도)
```

**점수 산정**: 순환 참조 없음 + 적정 응집도 (0-100)

#### 2.3 품질 지표 (Quality Metrics)

```python
# 분석 항목:
- 테스트 커버리지 (pytest-cov)
- 코드 스타일 준수 (린터 경고)
- 문서화율 (docstring 비율)
- 복잡도 함수 개수 (McCabe > 10)
```

**점수 산정**: 목표 대비 달성도 (0-100)

#### 2.4 성능 지표 (Performance Metrics)

```python
# 분석 항목:
- 초기화 시간 (moai init)
- doctor/status 진단 시간
- 테스트 실행 시간
- 빌드/배포 시간 (CI/CD)
- 메모리 사용량 추정
```

**점수 산정**: 성능 목표 대비 달성도 (0-100)

#### 2.5 호환성 분석 (Compatibility Analysis)

```python
# 분석 항목:
- Python 버전 호환성
- 의존 라이브러리 호환성
- 운영 체제 호환성 (Windows/macOS/Linux)
- IDE/에디터 통합 점검
- 배포 환경 호환성
```

**점수 산정**: 호환성 문제 수 기반 (0-100)

### 3. 분석 엔진 아키텍처

```
src/moai_adk/core/analysis/
├── __init__.py
├── analyzer.py          # 핵심 분석 엔진 (5개 분석기 조율)
├── metrics.py           # 메트릭 수집 및 계산
├── reporter.py          # 결과 리포팅 및 포맷팅
├── exporters.py         # JSON/CSV 내보내기
├── models.py            # Pydantic 데이터 모델
└── cache.py             # 분석 결과 캐싱

src/moai_adk/cli/commands/
└── analyze.py           # analyze 명령어 구현
```

### 4. 데이터 모델 (Pydantic)

```python
class AnalysisResult(BaseModel):
    timestamp: datetime
    category: str
    score: float  # 0-100
    status: str   # OK, WARNING, CRITICAL
    details: Dict[str, Any]
    recommendations: List[str]

class ProjectAnalysis(BaseModel):
    project_name: str
    analysis_time: float
    categories: Dict[str, AnalysisResult]
    overall_score: float
    priority_issues: List[str]
```

### 5. 추천사항 생성 규칙

```
Score Ranges:
- 90-100: EXCELLENT (권장사항 없음)
- 75-89: GOOD (선택적 개선사항)
- 60-74: FAIR (개선 권고)
- <60: CRITICAL (즉시 개선 필요)

예시:
- 테스트 커버리지 < 85% → "coverage run으로 레포트 생성" 안내
- 순환 참조 발견 → "의존성 리팩토링 권고"
```

## Acceptance Criteria (인수 기준)

### AC-1: 기본 분석 실행

- **Given**: 표준 moai-adk 프로젝트가 초기화되어 있을 때
- **When**: `moai analyze` 명령 실행
- **Then**:
  - 5개 분석 카테고리(구조, 의존성, 품질, 성능, 호환성)가 테이블로 표시
  - 각 카테고리의 점수(0-100)와 상태(OK/WARNING/CRITICAL)가 표시
  - 전체 종합 점수가 표시
  - 실행 시간이 10초 이내
  - 상위 3개 우선순위 문제가 표시

### AC-2: 깊은 분석 모드

- **Given**: 상세 분석이 필요할 때
- **When**: `moai analyze --deep` 명령 실행
- **Then**:
  - 코드 복잡도 함수 목록 표시
  - 의존성 그래프 시각화 (텍스트 기반)
  - 순환 참조 목록 표시
  - 라인 수 및 파일 크기 통계 표시
  - 실행 시간이 30초 이내

### AC-3: JSON 형식 내보내기

- **Given**: 분석 결과를 JSON으로 내보내야 할 때
- **When**: `moai analyze --format json` 명령 실행
- **Then**:
  - 모든 분석 데이터가 정규화된 JSON 구조로 출력
  - JSON이 pydantic 모델 구조를 따름
  - CI/CD 파이프라인에서 파싱 가능
  - 권장사항도 JSON에 포함

### AC-4: CSV 형식 내보내기

- **Given**: 분석 결과를 CSV로 내보내야 할 때
- **When**: `moai analyze --format csv` 명령 실행
- **Then**:
  - 분석 결과가 CSV 형식으로 출력 (카테고리, 점수, 상태, 권장사항)
  - 스프레드시트에서 열 수 있는 형식
  - 각 행이 하나의 분석 결과

### AC-5: 파일 내보내기

- **Given**: 분석 결과를 파일로 저장해야 할 때
- **When**: `moai analyze --export analysis.json` 명령 실행
- **Then**:
  - 분석 결과가 지정된 파일에 저장
  - 파일 경로가 메시지에 표시
  - 파일이 읽을 수 있는 형식(JSON 또는 CSV)으로 저장
  - 다양한 확장자 지원 (.json, .csv, .txt)

### AC-6: 특정 카테고리 분석

- **Given**: 특정 분석 카테고리만 필요할 때
- **When**: `moai analyze --category quality` 명령 실행
- **Then**:
  - 지정된 카테고리(quality)만 분석 및 표시
  - 해당 카테고리의 상세 정보 표시
  - 다른 카테고리는 스킵
  - 실행 시간이 5초 이내

### AC-7: 이전 결과와 비교

- **Given**: 이전 분석 결과가 저장되어 있을 때
- **When**: `moai analyze --compare` 명령 실행
- **Then**:
  - 각 카테고리 점수의 변화 표시 (증감 표시)
  - 개선된 항목과 악화된 항목 구분
  - 변화 추세 분석 표시
  - 이전 분석 타임스탐프 표시

### AC-8: CI/CD 모드

- **Given**: CI/CD 환경에서 실행할 때
- **When**: `moai analyze --ci` 명령 실행
- **Then**:
  - Rich UI(색상, 박스, 아이콘) 비활성화
  - 플레인 텍스트로 출력
  - JSON 형식으로 구조화된 데이터 포함
  - 파이프라인에서 파싱 가능

### AC-9: 임계값 기반 필터링

- **Given**: 특정 점수 이상의 문제만 보고하려고 할 때
- **When**: `moai analyze --threshold 70` 명령 실행
- **Then**:
  - 70점 미만의 카테고리만 상세 표시
  - 우선순위 문제도 임계값 기반으로 필터링
  - 전체 분석은 수행되지만 필터된 결과만 표시

### AC-10: 분석 캐싱

- **Given**: 동일한 프로젝트에서 분석을 여러 번 실행할 때
- **When**: 두 번째 `moai analyze` 명령 실행
- **Then**:
  - 첫 번째 분석 결과를 캐시에서 로드 (프로젝트 변경 없을 시)
  - 두 번째 실행 시간이 첫 번째보다 훨씬 빠름 (<1초)
  - 파일 변경 감지 시 캐시 무효화

### AC-11: 진행 상황 표시

- **Given**: 깊은 분석이 실행 중일 때
- **When**: `moai analyze --deep` 명령이 실행 중일 때
- **Then**:
  - 진행 상황 표시줄(Progress Bar) 표시
  - 현재 진행 중인 분석 카테고리 표시
  - 예상 완료 시간 표시
  - 사용자는 진행 상황을 시각적으로 추적 가능

### AC-12: 추천사항 자동 생성

- **Given**: 특정 점수 범위의 카테고리가 분석되었을 때
- **When**: 분석 완료 후 결과 표시
- **Then**:
  - 각 카테고리별 점수에 따른 추천사항이 자동 생성
  - CRITICAL(<60): 즉시 개선 필요 가이드 (실행 명령 포함)
  - FAIR(60-74): 개선 권고사항
  - GOOD(75-89): 선택적 개선사항
  - EXCELLENT(90-100): 권장사항 없음
  - 권장사항이 실행 가능한 명령어 포함


- **TEST**:
  - `tests/unit/test_analyzer.py`
  - `tests/unit/test_metrics.py`
  - `tests/unit/test_exporters.py`
  - `tests/integration/test_analyze_command.py`
- **CODE**:
  - `src/moai_adk/cli/commands/analyze.py`
  - `src/moai_adk/core/analysis/analyzer.py`
  - `src/moai_adk/core/analysis/metrics.py`
  - `src/moai_adk/core/analysis/reporter.py`
  - `src/moai_adk/core/analysis/exporters.py`
  - `src/moai_adk/core/analysis/models.py`
  - `src/moai_adk/core/analysis/cache.py`
- **DOC**: `.moai/docs/CLI-ANALYSIS.md`

## References

- **TRUST 원칙**: `.moai/memory/development-guide.md#trust-5원칙`
- **Click 프레임워크**: https://click.palletsprojects.com/
- **Rich 라이브러리**: https://rich.readthedocs.io/
- **networkx**: https://networkx.org/
- **Pydantic**: https://docs.pydantic.dev/
- **관련 SPEC**: `SPEC-CLI-001` (기본 doctor/status/restore)

## Non-Goals (범위 외)

- **자동 수정**: 분석 기능은 진단과 추천만 수행하며, 자동 수정은 향후 고려
- **GUI 인터페이스**: CLI만 지원, GUI는 향후 고려
- **원격 분석**: 로컬 프로젝트만 분석, 원격 서버 분석은 별도 SPEC
- **머신러닝 기반 예측**: 현재는 규칙 기반 분석, ML 기반 예측은 향후 고려

## Migration Plan (기존 사용자 영향)

**기존 명령어 호환성**: 100% 유지
- `moai doctor`, `moai status`, `moai restore`: 변경 없음
- 모든 기존 옵션 유지

**새 명령어**: 선택적 사용
- `moai analyze`: 새로운 명령어로 완전히 독립적
- 기존 사용자 워크플로우에 영향 없음
- 새로운 기능을 원하는 사용자만 사용

**문서 업데이트**: README.md와 `.moai/docs/`에 analyze 명령어 추가
