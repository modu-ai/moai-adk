---
id: DOC-TAG-004
version: 0.0.1
status: completed
created: 2025-10-29
updated: 2025-10-29
author: "@Goos"
priority: high
category: Quality / Automation / CI-CD
labels: [quality-gates, automation, ci-cd, validation]
depends_on: [DOC-TAG-001, DOC-TAG-002, DOC-TAG-003]
scope: "Phase 4 of 4-phase @DOC TAG automatic generation system - Quality gates & automation"
---

# @SPEC:DOC-TAG-004

**@TAG 자동 검증 및 품질 게이트 (Phase 4)**

---

## HISTORY

### v0.0.1 (2025-10-29)
- **초안 작성**: Phase 4 전체 시스템 설계 완료
- **범위**: Pre-commit Hooks, CI/CD Pipeline, Validation System, Documentation
- **의존성**: Phase 1 (DOC-TAG-001), Phase 2 (DOC-TAG-002), Phase 3 (DOC-TAG-003) 완료 상태
- **상태**: Draft - 구현 대기 중

---

## ENVIRONMENT

### 현재 시스템 상태

**Phase 1-3 완료 상태** (@SPEC:DOC-TAG-001, @SPEC:DOC-TAG-002, @SPEC:DOC-TAG-003):
- ✅ `TAGGenerator` 클래스: 82개 파일 자동 TAG 생성 완료
- ✅ `doc-syncer` 에이전트: Markdown 문서 자동 TAG 삽입 완료
- ✅ `tag-agent` 기반: TAG 스캔 및 동기화 기능 구현
- ✅ 체인 관리 시스템: `TAG-CHAIN-MAP.md` 자동 생성

**현재 TAG 인벤토리**:
- 총 82개 파일에 TAG 적용 완료
- 파일 유형: Python (46), Markdown (25), YAML (7), Shell (4)
- TAG 체인: 12개 주요 체인 확립 (AUTH, DOC, TRUST, LANG 등)
- 검증 상태: 수동 검증만 가능 (자동 검증 부재)

**인프라 준비 상태**:
- Git 저장소: 활성 상태 (feature/DOC-TAG-001-auto-generation)
- Python 환경: 3.9+ 설치됨
- GitHub 연동: 가능 (GitHub Actions 사용 가능)
- Pre-commit 프레임워크: 미설치 (Phase 4에서 설치 예정)

**현재 문제점**:
- ❌ TAG 무결성 자동 검증 부재
- ❌ 커밋 시 TAG 중복/누락 감지 불가
- ❌ CI/CD 파이프라인 미통합
- ❌ TAG 인벤토리 자동 업데이트 부재
- ❌ 품질 게이트 미적용

---

## ASSUMPTIONS

### 시스템 가정

1. **Phase 1-3 완료 상태**:
   - `TAGGenerator` 클래스가 정상 작동하며, 82개 파일에 TAG가 정확히 삽입되어 있다.
   - `doc-syncer` 에이전트가 Markdown 문서의 TAG를 올바르게 관리한다.
   - `TAG-CHAIN-MAP.md`가 최신 상태를 반영한다.

2. **개발 환경**:
   - Python 3.9 이상이 설치되어 있다.
   - Git 2.30 이상 사용 (Git Hooks 지원).
   - GitHub 저장소 접근 권한 보유 (GitHub Actions 사용 가능).
   - `pre-commit` 프레임워크 설치 가능.

3. **팀 협업 환경**:
   - 개발자들이 로컬에서 `pre-commit install` 실행 가능.
   - CI/CD 파이프라인 실행 권한 보유 (GitHub Actions).
   - 품질 게이트 실패 시 PR 머지 차단 정책 수용.

4. **기술 스택**:
   - Bash 스크립트 작성 가능 (Pre-commit Hooks).
   - YAML 작성 가능 (GitHub Actions 워크플로우).
   - Python 검증 로직 작성 가능 (TAG Validator).

5. **비기능 요구사항**:
   - Pre-commit Hook 실행 시간: 3초 이내 (개발자 경험 중요).
   - CI/CD 검증 시간: 5분 이내 (전체 TAG 검증).
   - False Positive 비율: 5% 이하 (잘못된 검증 오류 최소화).

---

## REQUIREMENTS

### Ubiquitous Requirements (언제나 적용)

**REQ-1**: 시스템은 모든 커밋 및 PR에서 TAG 무결성을 자동으로 검증해야 한다.

**REQ-2**: 시스템은 TAG 중복, 누락, 고아 TAG를 자동으로 감지해야 한다.

**REQ-3**: 시스템은 TAG 체인의 완전성을 검증해야 한다 (예: `@SPEC:AUTH-004` → `@TEST:AUTH-004` → `@CODE:AUTH-004` 연결).

**REQ-4**: 시스템은 검증 실패 시 명확한 오류 메시지와 수정 가이드를 제공해야 한다.

**REQ-5**: 시스템은 Phase 1-3의 `TAGGenerator` 및 `doc-syncer` 모듈을 재사용해야 한다.

### Event-Driven Requirements (이벤트 발생 시)

**REQ-6**: **WHEN** 개발자가 로컬에서 커밋을 실행하면, **THEN** Pre-commit Hook가 변경된 파일의 TAG를 검증해야 한다.

**REQ-7**: **WHEN** 개발자가 GitHub에 PR을 생성하면, **THEN** CI/CD 파이프라인이 전체 저장소의 TAG 무결성을 검증해야 한다.

**REQ-8**: **WHEN** TAG 검증이 실패하면, **THEN** 시스템은 커밋/PR을 차단하고 상세한 오류 리포트를 생성해야 한다.

**REQ-9**: **WHEN** TAG 검증이 성공하면, **THEN** 시스템은 TAG 인벤토리를 자동으로 업데이트해야 한다.

### State-Driven Requirements (특정 상태에서)

**REQ-10**: **WHILE** 시스템이 검증 모드에 있는 동안, TAG 중복을 감지하면 첫 번째 발견된 TAG 외의 모든 중복을 리포트해야 한다.

**REQ-11**: **WHILE** 시스템이 검증 모드에 있는 동안, 고아 TAG (체인이 끊긴 TAG)를 발견하면 체인 복구 제안을 제공해야 한다.

**REQ-12**: **WHILE** CI/CD 파이프라인이 실행 중인 동안, 시스템은 진행 상황을 실시간으로 로깅해야 한다.

### Optional Requirements (선택적 기능)

**REQ-13**: 시스템은 CLI 유틸리티를 제공하여 개발자가 수동으로 TAG 검증을 실행할 수 있어야 한다 (예: `moai-adk validate-tags`).

**REQ-14**: 시스템은 TAG 검증 리포트를 JSON 형식으로 출력할 수 있어야 한다 (CI/CD 통합 용도).

**REQ-15**: 시스템은 TAG 인벤토리를 HTML 형식으로 출력하여 웹 대시보드에 통합할 수 있어야 한다.

### Unwanted Behaviors (방지해야 할 동작)

**REQ-16**: 시스템은 TAG 검증 실패 시 커밋/PR을 머지하지 않아야 한다 (강제 차단).

**REQ-17**: 시스템은 고아 TAG를 무시하지 않아야 한다 (모든 TAG는 체인에 연결되어야 함).

**REQ-18**: 시스템은 False Positive를 최소화해야 한다 (정상 TAG를 오류로 판단하지 않음).

**REQ-19**: 시스템은 Pre-commit Hook 실행 시간이 5초를 초과하지 않아야 한다 (개발자 경험 저하 방지).

**REQ-20**: 시스템은 TAG 검증 규칙을 하드코딩하지 않아야 한다 (설정 파일로 관리).

---

## SPECIFICATIONS

### Overview

Phase 4는 4개의 핵심 컴포넌트로 구성됩니다:

1. **Pre-commit Hooks**: 로컬 커밋 시 빠른 TAG 검증
2. **CI/CD Pipeline**: GitHub Actions 통합 및 전체 검증
3. **Validation System**: 복잡한 TAG 체인 및 무결성 검증
4. **Documentation**: TAG 인벤토리 및 매트릭스 자동 생성

### Component 1: Pre-commit Hooks

**목표**: 로컬 커밋 시 변경된 파일의 TAG를 빠르게 검증하여 조기에 문제를 발견한다.

**파일 구조**:
```
.moai/hooks/
├── pre-commit.sh          # Main pre-commit hook
├── tag-validator.sh       # TAG validation logic
└── config/
    └── validation-rules.yml  # Validation rules configuration
```

**기능 명세**:

1. **빠른 검증 (3초 이내)**:
   - Git staged files만 검증 (전체 저장소 스캔 안 함).
   - Python `TAGGenerator`의 `scan_file()` 메서드 재사용.
   - 병렬 처리로 성능 최적화.

2. **검증 규칙**:
   - TAG 중복 감지: 동일한 TAG ID가 여러 파일에 존재하는지 확인.
   - TAG 형식 검증: `@{TYPE}:{DOMAIN}-{ID}` 패턴 준수 확인.
   - 필수 TAG 존재: SPEC 파일에는 `@SPEC:` TAG 필수.

3. **오류 처리**:
   - 검증 실패 시 커밋 차단.
   - 상세한 오류 메시지 출력 (파일명, 라인 번호, 문제 설명).
   - 수정 가이드 제공 (예: "중복 TAG 제거 방법").

4. **설정 파일** (`.moai/hooks/config/validation-rules.yml`):
```yaml
validation:
  enable_duplicate_check: true
  enable_format_check: true
  enable_chain_check: false  # Pre-commit에서는 경량 검증만
  timeout_seconds: 3
  exclude_patterns:
    - "*.pyc"
    - "__pycache__"
    - ".git/"
```

**구현 시간**: 8 시간 (1주차)

**기대 효과**:
- 조기 문제 발견으로 CI/CD 실패 감소 (예상 70% 감소).
- 개발자 피드백 루프 단축 (커밋 즉시 오류 확인).

---

### Component 2: CI/CD Pipeline

**목표**: GitHub Actions를 통해 PR 시 전체 저장소의 TAG 무결성을 검증하고 품질 게이트를 적용한다.

**파일 구조**:
```
.github/workflows/
├── tag-validation.yml     # Main TAG validation workflow
└── tag-report.yml         # TAG inventory update workflow
```

**기능 명세**:

1. **전체 검증 (5분 이내)**:
   - 전체 저장소 스캔 (82개 파일 + 신규 파일).
   - Python `TAGValidator` 클래스 호출.
   - TAG 체인 완전성 검증 (SPEC → TEST → CODE 연결).

2. **워크플로우 트리거**:
   - PR 생성/업데이트 시 자동 실행.
   - `main` 브랜치로의 머지 시 실행.
   - 수동 트리거 가능 (`workflow_dispatch`).

3. **검증 단계**:
   ```yaml
   jobs:
     validate-tags:
       runs-on: ubuntu-latest
       steps:
         - name: Checkout code
         - name: Setup Python 3.9+
         - name: Install dependencies
         - name: Run TAG validation
         - name: Generate validation report
         - name: Upload report artifact
         - name: Comment on PR (if failed)
   ```

4. **품질 게이트**:
   - 검증 실패 시 PR 머지 차단 (required status check).
   - 고아 TAG 발견 시 경고 (차단 안 함, 리뷰어 판단).
   - TAG 인벤토리 업데이트 자동 커밋.

5. **리포트 생성**:
   - Markdown 형식의 검증 리포트 생성.
   - PR 코멘트로 자동 추가.
   - GitHub Actions Artifact로 저장.

**구현 시간**: 12 시간 (1주차)

**기대 효과**:
- PR 머지 전 TAG 무결성 보장 (100% 검증).
- 자동화된 품질 게이트로 수동 리뷰 부담 감소.

---

### Component 3: Validation System

**목표**: 복잡한 TAG 체인, 중복, 고아 TAG를 정밀하게 검증하는 핵심 검증 로직을 구현한다.

**파일 구조**:
```
src/moai_adk/core/tags/
├── validator.py           # Main TAG validation logic
├── chain_validator.py     # TAG chain validation
├── duplicate_detector.py  # Duplicate TAG detection
├── orphan_detector.py     # Orphan TAG detection
└── models.py              # Validation result models
```

**기능 명세**:

1. **`TAGValidator` 클래스**:
   ```python
   class TAGValidator:
       def validate_repository(self, repo_path: str) -> ValidationReport:
           """전체 저장소 검증"""

       def validate_files(self, file_paths: List[str]) -> ValidationReport:
           """특정 파일 목록 검증"""

       def validate_tag_chain(self, tag_id: str) -> ChainValidationResult:
           """TAG 체인 완전성 검증"""
   ```

2. **중복 검증** (`duplicate_detector.py`):
   - 동일한 TAG ID가 여러 파일에 존재하는지 확인.
   - 예외 케이스: 문서 참조는 중복 허용 (예: `plan.md`에서 `@SPEC:AUTH-004` 참조).
   - 검증 로직:
     ```python
     def detect_duplicates(tags: List[TAGInfo]) -> List[DuplicateIssue]:
         # Group by TAG ID
         # Filter by file type (source files only)
         # Return duplicates with file locations
     ```

3. **고아 TAG 검증** (`orphan_detector.py`):
   - TAG 체인이 끊긴 TAG 발견 (예: `@TEST:AUTH-004` 존재, `@SPEC:AUTH-004` 부재).
   - 체인 복구 제안 생성:
     ```python
     def detect_orphans(tags: List[TAGInfo]) -> List[OrphanIssue]:
         # Build TAG chain graph
         # Find disconnected nodes
         # Suggest reconnection strategies
     ```

4. **체인 검증** (`chain_validator.py`):
   - SPEC → TEST → CODE → DOC 체인 완전성 확인.
   - 필수 체인 규칙:
     ```yaml
     required_chains:
       - pattern: "@SPEC:{ID} -> @TEST:{ID}"
         severity: error
       - pattern: "@TEST:{ID} -> @CODE:{ID}"
         severity: warning
       - pattern: "@CODE:{ID} -> @DOC:{ID}"
         severity: info
     ```

5. **검증 리포트 모델** (`models.py`):
   ```python
   @dataclass
   class ValidationReport:
       total_tags: int
       valid_tags: int
       issues: List[ValidationIssue]
       duplicate_count: int
       orphan_count: int
       chain_breaks: int
       execution_time: float

   @dataclass
   class ValidationIssue:
       severity: Literal["error", "warning", "info"]
       issue_type: str
       tag_id: str
       file_path: str
       line_number: int
       message: str
       suggestion: str
   ```

**구현 시간**: 15 시간 (2주차)

**기대 효과**:
- 정밀한 TAG 무결성 검증 (중복/고아/체인 모두 커버).
- 재사용 가능한 검증 라이브러리 (CLI, CI/CD, IDE 플러그인 모두 사용 가능).

---

### Component 4: Documentation & Reporting

**목표**: TAG 인벤토리, 매트릭스, 검증 리포트를 자동으로 생성하여 프로젝트의 TAG 상태를 투명하게 유지한다.

**파일 구조**:
```
docs/
├── tag-inventory.md       # TAG 인벤토리 (자동 생성)
├── tag-matrix.md          # TAG 매트릭스 (자동 생성)
└── validation-reports/
    └── {date}-validation-report.md  # 검증 리포트
```

**기능 명세**:

1. **TAG 인벤토리 자동 생성** (`tag-inventory.md`):
   - 전체 TAG 목록 (파일별, 타입별, 도메인별 그룹핑).
   - 생성 시점: CI/CD 파이프라인 성공 시 자동 업데이트.
   - 포맷:
     ```markdown
     # TAG Inventory

     ## Summary
     - Total TAGs: 82
     - By Type: SPEC (15), TEST (20), CODE (30), DOC (17)
     - By Domain: AUTH (8), DOC (12), TRUST (6), LANG (10), ...

     ## TAG List by File

     ### src/moai_adk/core/tags/generator.py
     - @CODE:DOC-TAG-001 (Line 15)
     - @TEST:DOC-TAG-001 (Line 120)

     ...
     ```

2. **TAG 매트릭스 자동 생성** (`tag-matrix.md`):
   - TAG 체인 매핑 테이블 (SPEC → TEST → CODE → DOC).
   - 생성 시점: CI/CD 파이프라인 성공 시 자동 업데이트.
   - 포맷:
     ```markdown
     # TAG Matrix

     | SPEC | TEST | CODE | DOC | Status |
     |------|------|------|-----|--------|
     | @SPEC:AUTH-004 | @TEST:AUTH-004 | @CODE:AUTH-004 | @DOC:AUTH-001 | ✅ Complete |
     | @SPEC:DOC-TAG-001 | @TEST:DOC-TAG-001 | @CODE:DOC-TAG-001 | - | ⚠️ Missing DOC |

     ...
     ```

3. **검증 리포트 생성** (`validation-reports/{date}-validation-report.md`):
   - 검증 실행 시점 기록.
   - 발견된 이슈 상세 리스트.
   - 수정 제안 및 가이드.
   - 포맷:
     ```markdown
     # TAG Validation Report

     **Date**: 2025-10-29
     **Commit**: abc1234
     **Status**: ❌ Failed

     ## Summary
     - Total TAGs scanned: 82
     - Issues found: 3
     - Errors: 2 (duplicate TAGs)
     - Warnings: 1 (orphan TAG)

     ## Issues

     ### Error: Duplicate TAG detected
     - **TAG**: @CODE:AUTH-004
     - **Files**:
       - src/auth/login.py (Line 45)
       - src/auth/token.py (Line 12)
     - **Suggestion**: Remove duplicate TAG from one file or rename to unique ID.

     ...
     ```

4. **자동 커밋 로직**:
   - CI/CD 파이프라인이 인벤토리/매트릭스 업데이트 후 자동 커밋.
   - 커밋 메시지: `docs(TAG): Update TAG inventory and matrix [skip ci]`.

**구현 시간**: 8 시간 (3주차)

**기대 효과**:
- TAG 상태의 투명성 확보 (언제든지 최신 인벤토리 확인 가능).
- 검증 리포트로 이슈 추적 및 수정 가능.

---

## TRACEABILITY

### TAG Chains (추적성)

**현재 SPEC**:
- `@SPEC:DOC-TAG-004` (이 문서)

**의존 SPEC**:
- `@SPEC:DOC-TAG-001` (Phase 1 - TAG 생성 라이브러리)
- `@SPEC:DOC-TAG-002` (Phase 2 - SPEC 통합 및 TAG 체인)
- `@SPEC:DOC-TAG-003` (Phase 3 - 대규모 적용 및 doc-syncer 통합)

**생성될 TAG** (Phase 4 구현 완료 시):
- `@TEST:DOC-TAG-004` (Pre-commit Hook 테스트, CI/CD 테스트)
- `@CODE:DOC-TAG-004` (validator.py, CI/CD 워크플로우)
- `@DOC:DOC-TAG-004` (TAG 인벤토리, 매트릭스, 검증 리포트)

**연결될 TAG**:
- `@CODE:DOC-TAG-001` → `TAGGenerator` 클래스 재사용
- `@CODE:DOC-TAG-003` → `doc-syncer` 에이전트 연동
- `@TEST:DOC-TAG-001` → 검증 로직 테스트 재사용

**문서 체인**:
- `.moai/specs/SPEC-DOC-TAG-004/spec.md` (이 문서)
- `.moai/specs/SPEC-DOC-TAG-004/plan.md` (구현 계획)
- `.moai/specs/SPEC-DOC-TAG-004/acceptance.md` (인수 기준)
- `docs/tag-inventory.md` (TAG 인벤토리)
- `docs/tag-matrix.md` (TAG 매트릭스)

---

## NOTES

### 재사용 전략

Phase 4는 Phase 1-3의 성과물을 최대한 재사용합니다:

1. **`TAGGenerator` 클래스** (Phase 1):
   - `scan_file()`: Pre-commit Hook에서 파일 스캔에 사용.
   - `generate_tag()`: 신규 TAG 생성 시 재사용.

2. **`doc-syncer` 에이전트** (Phase 3):
   - TAG 동기화 로직을 검증 후 인벤토리 업데이트에 활용.
   - Markdown 파싱 로직 재사용.

3. **`TAG-CHAIN-MAP.md`** (Phase 2):
   - 체인 검증의 기준 데이터로 사용.
   - 매트릭스 생성 시 참조.

### 위험 요소 및 대응

| 위험 | 영향 | 확률 | 대응 전략 |
|------|------|------|-----------|
| Pre-commit Hook 실행 시간 초과 (>5초) | 개발자 경험 저하 | 중 | 병렬 처리 + 경량 검증만 수행 |
| CI/CD 파이프라인 실패율 증가 | 개발 속도 저하 | 중 | False Positive 최소화 + 명확한 오류 메시지 |
| TAG 검증 규칙 복잡도 증가 | 유지보수 어려움 | 저 | 설정 파일로 규칙 관리 + 문서화 |
| 고아 TAG 자동 복구 실패 | 수동 개입 필요 | 중 | 복구 제안만 제공, 자동 수정 안 함 |

### 성공 기준

Phase 4 구현 완료 시 다음 조건을 모두 충족해야 합니다:

- ✅ Pre-commit Hook 실행 시간 3초 이내
- ✅ CI/CD 파이프라인 실행 시간 5분 이내
- ✅ TAG 중복 감지율 100%
- ✅ 고아 TAG 감지율 95% 이상
- ✅ False Positive 비율 5% 이하
- ✅ TAG 인벤토리 자동 업데이트 성공률 100%
- ✅ 모든 테스트 통과 (단위, 통합, E2E)

---

**문서 종료**
