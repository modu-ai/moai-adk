# SPEC-TRUST-001 Implementation Plan

> **TRUST 원칙 자동 검증 시스템 구현 계획**

---

## 개요

이 문서는 SPEC-TRUST-001 (TRUST 원칙 자동 검증)의 구현 계획을 정의합니다. TDD 방식으로 Red-Green-Refactor 사이클을 따릅니다.

---

## Phase 1: Core 검증 로직 구현 (우선순위: High)

### 목표
- TRUST 5원칙 개별 검증기 구현
- 언어별 도구 자동 선택 메커니즘

### 작업 항목

#### 1.1 테스트 커버리지 검증기
- **파일**: `src/core/quality/coverage-validator.ts`
- **기능**:
  - 언어별 커버리지 도구 실행 (pytest-cov, Vitest, JaCoCo 등)
  - 커버리지 수치 파싱 (85% 기준 검증)
  - 낮은 커버리지 파일 목록 반환
- **의존성**: ripgrep, 언어별 테스트 도구
- **테스트**: `tests/core/quality/test_coverage_validator.py`

#### 1.2 코드 제약 검증기
- **파일**: `src/core/quality/code-constraints-validator.ts`
- **기능**:
  - 파일 ≤300 LOC 검증
  - 함수 ≤50 LOC 검증
  - 매개변수 ≤5개 검증
  - 순환 복잡도 ≤10 검증
- **도구**: AST 파서 (ts-morph, ast-types, Roslyn 등)
- **테스트**: `tests/core/quality/test_code_constraints_validator.py`

#### 1.3 타입 안전성 검증기
- **파일**: `src/core/quality/type-safety-validator.ts`
- **기능**:
  - 언어별 타입 체커 실행 (tsc, mypy, javac 등)
  - 타입 오류 수집 및 보고
- **테스트**: `tests/core/quality/test_type_safety_validator.py`

#### 1.4 보안 스캔 검증기
- **파일**: `src/core/quality/security-validator.ts`
- **기능**:
  - 보안 취약점 스캔 (npm audit, pip-audit, Snyk 등)
  - High/Critical 취약점 필터링
  - CVE ID 및 CVSS 점수 보고
- **테스트**: `tests/core/quality/test_security_validator.py`

#### 1.5 언어 감지 및 도구 선택기
- **파일**: `src/core/quality/language-detector.ts`
- **기능**:
  - `.moai/config.json`에서 `project.language` 읽기
  - 언어별 도구 매핑 (S-002 참조)
  - 도구 설치 여부 확인
- **테스트**: `tests/core/quality/test_language_detector.py`

### 완료 기준
- [ ] 5개 검증기 구현 완료
- [ ] 단위 테스트 커버리지 ≥85%
- [ ] 모든 테스트 통과 (pytest, Vitest)

---

## Phase 2: TAG 시스템 통합 (우선순위: High)

### 목표
- @TAG 체인 검증 로직 구현
- 고아 TAG 탐지 및 순환 참조 감지

### 작업 항목

#### 2.1 TAG 스캔 엔진
- **파일**: `src/core/quality/tag-scanner.ts`
- **기능**:
  - ripgrep으로 `@(SPEC|TEST|CODE|DOC):` 패턴 스캔
  - TAG ID 추출 및 파싱 (`<DOMAIN>-<NUMBER>` 형식)
  - TAG 위치 정보 수집 (파일명, 라인 번호)
- **테스트**: `tests/core/quality/test_tag_scanner.py`

#### 2.2 TAG 체인 검증기
- **파일**: `src/core/quality/tag-chain-validator.ts`
- **기능**:
  - `@SPEC:ID` → `@TEST:ID` → `@CODE:ID` 연결 검증
  - 고아 TAG 탐지:
    - `@CODE:ID` 있으나 `@SPEC:ID` 없음
    - `@SPEC:ID` 있으나 `@CODE:ID` 없음
  - 순환 참조 탐지 (depends_on 필드 분석)
- **알고리즘**: S-003 참조
- **테스트**: `tests/core/quality/test_tag_chain_validator.py`

#### 2.3 TAG 무결성 보고
- **파일**: `src/core/quality/tag-reporter.ts`
- **기능**:
  - TAG 체인 시각화 (그래프 형식)
  - 고아 TAG 목록
  - 끊어진 링크 목록
- **테스트**: `tests/core/quality/test_tag_reporter.py`

### 완료 기준
- [ ] TAG 스캔 정확도 100%
- [ ] 고아 TAG 탐지 100%
- [ ] 순환 참조 감지 100%

---

## Phase 3: 보고서 생성 (우선순위: Medium)

### 목표
- Markdown 및 JSON 형식 보고서 생성
- 오류 메시지 표준화

### 작업 항목

#### 3.1 보고서 생성기
- **파일**: `src/core/quality/report-generator.ts`
- **기능**:
  - Markdown 템플릿 렌더링 (S-004 참조)
  - JSON 구조 생성 (CI/CD 통합용)
  - 파일 저장: `.moai/reports/trust-report-{timestamp}.md`
- **테스트**: `tests/core/quality/test_report_generator.py`

#### 3.2 오류 메시지 포맷터
- **파일**: `src/core/quality/error-formatter.ts`
- **기능**:
  - 심각도별 아이콘 (❌ ⚠️ ℹ️)
  - 메시지 템플릿 (S-005 참조)
  - 파일명/라인 번호 하이퍼링크 생성
- **테스트**: `tests/core/quality/test_error_formatter.py`

#### 3.3 진행 상태 표시기
- **파일**: `src/core/quality/progress-indicator.ts`
- **기능**:
  - 검증 단계별 진행률 표시 (예: [3/5] TAG 체인 검증 중...)
  - 예상 완료 시간 (선택적)
- **테스트**: `tests/core/quality/test_progress_indicator.py`

### 완료 기준
- [ ] Markdown 보고서 생성 확인
- [ ] JSON 보고서 스키마 검증
- [ ] 오류 메시지 표준 준수

---

## Phase 4: /alfred:2-build 통합 (우선순위: High)

### 목표
- `/alfred:2-build` 완료 후 자동 실행
- 커밋 차단 메커니즘 구현

### 작업 항목

#### 4.1 trust-checker CLI
- **파일**: `src/core/quality/trust-checker.ts`
- **기능**:
  - 5개 검증기 순차 실행 (또는 병렬)
  - 검증 실패 시 exit code 1 반환
  - 보고서 생성 및 저장
- **CLI 옵션**:
  - `--report-format`: markdown | json | both
  - `--fail-on-warning`: 경고 발생 시도 실패 처리
  - `--skip`: 특정 검증 건너뛰기 (예: --skip=security)
- **테스트**: `tests/core/quality/test_trust_checker.py`

#### 4.2 /alfred:2-build 통합
- **파일**: `src/commands/alfred-2-build.ts`
- **기능**:
  - GREEN 단계 커밋 직후 trust-checker 자동 실행
  - 검증 실패 시 REFACTOR 단계 진입 차단
  - 오류 메시지 표시 및 사용자 안내
- **테스트**: `tests/commands/test_alfred_2_build.py` (통합 테스트)

#### 4.3 커밋 차단 메커니즘
- **파일**: `src/core/git/commit-guard.ts`
- **기능**:
  - trust-checker 실행 결과 확인
  - 실패 시 커밋 차단 (exit code 1)
  - 성공 시 커밋 진행
- **테스트**: `tests/core/git/test_commit_guard.py`

### 완료 기준
- [ ] trust-checker CLI 정상 동작
- [ ] /alfred:2-build 통합 확인
- [ ] 커밋 차단 메커니즘 동작 확인

---

## 기술 스택

### TypeScript 구현 (Primary)
- **런타임**: Node.js 18+
- **테스트**: Vitest
- **린터**: Biome
- **타입 체커**: tsc
- **AST 파서**: ts-morph
- **보안**: npm audit, Snyk

### Python 구현 (Secondary, 언어 감지 시)
- **런타임**: Python 3.10+
- **테스트**: pytest
- **린터**: ruff
- **타입 체커**: mypy
- **커버리지**: coverage.py
- **AST 파서**: ast (내장)

### 공통 도구
- **TAG 스캔**: ripgrep (rg)
- **파일 처리**: fs-extra (Node.js), pathlib (Python)
- **병렬 실행**: Promise.all (TS), asyncio.gather (Py)

---

## 리스크 및 완화 방안

### 리스크 1: 도구 의존성 (High)
- **문제**: 언어별 도구가 설치되지 않은 경우
- **완화**:
  - 도구 설치 여부 자동 감지
  - 미설치 시 설치 명령 안내 메시지
  - Fallback: 기본 검증만 수행 (예: LOC 계산만)

### 리스크 2: 성능 (Medium)
- **문제**: 대규모 프로젝트에서 검증 시간 초과
- **완화**:
  - 병렬 실행 (Promise.all, asyncio.gather)
  - 증분 검증 (Git diff 기반)
  - 캐싱 (이전 검증 결과 재사용)

### 리스크 3: 복잡도 계산 차이 (Medium)
- **문제**: 언어별 복잡도 계산 도구가 다름
- **완화**:
  - 복잡도 기준 명확히 정의 (순환 복잡도)
  - 도구별 차이 허용 범위 설정 (±2)

### 리스크 4: TAG 파싱 오류 (Low)
- **문제**: 잘못된 TAG 형식 (`@CODE:AUTH` 대신 `@CODE AUTH-001`)
- **완화**:
  - 정규식 패턴 강화 (`@[A-Z]+:[A-Z]+-\d{3}`)
  - 파싱 오류 시 경고 메시지

---

## 마일스톤

| 단계    | 목표                          | 완료 기준                     |
| ------- | ----------------------------- | ----------------------------- |
| Phase 1 | Core 검증 로직 구현           | 5개 검증기 완료, 테스트 통과  |
| Phase 2 | TAG 시스템 통합               | TAG 체인 검증 100%            |
| Phase 3 | 보고서 생성                   | Markdown/JSON 보고서 생성     |
| Phase 4 | /alfred:2-build 통합          | 자동 실행 및 커밋 차단 동작   |

---

## 다음 단계

1. `/alfred:2-build TRUST-001` 실행 (TDD 구현 시작)
2. Phase 1 완료 후 Phase 2로 진행
3. Phase 4 완료 후 `/alfred:3-sync` 실행

---

**Last Updated**: 2025-10-16
**Author**: @Goos
