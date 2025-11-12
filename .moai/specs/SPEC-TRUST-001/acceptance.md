# @SPEC:TRUST-001 Acceptance Criteria

> **TRUST 원칙 자동 검증 시스템 수락 기준**

---

## 개요

이 문서는 SPEC-TRUST-001의 상세한 수락 기준과 테스트 시나리오를 정의합니다. 모든 테스트는 Given-When-Then 형식을 따릅니다.

---

## TC-001: 테스트 커버리지 ≥85% 검증

### 정상 케이스

**Given**:
- 프로젝트에 테스트 파일이 작성됨 (`tests/`)
- 테스트 커버리지가 87%

**When**:
- `trust-checker --check=coverage` 실행

**Then**:
- 검증 결과: ✅ PASS
- 출력 메시지:
  ```
  ✅ Test Coverage: 87% (Target: 85%)
  ```
- Exit code: 0

### 오류 케이스

**Given**:
- 프로젝트에 테스트 파일이 작성됨
- 테스트 커버리지가 78%

**When**:
- `trust-checker --check=coverage` 실행

**Then**:
- 검증 결과: ❌ FAIL
- 출력 메시지:
  ```
  ❌ Test Coverage: 78% (Target: 85%)
    → 추가 테스트 케이스 작성 권장
    → 낮은 커버리지 파일:
      - src/utils/helper.ts: 72%
      - src/core/validator.ts: 78%
  ```
- Exit code: 1

---

## TC-002: 파일 ≤300 LOC 검증

### 정상 케이스

**Given**:
- 프로젝트에 소스 파일이 존재
- 모든 파일이 300 LOC 이하

**When**:
- `trust-checker --check=code-constraints` 실행

**Then**:
- 검증 결과: ✅ PASS
- 출력 메시지:
  ```
  ✅ Code Constraints: All files ≤300 LOC
  ```
- Exit code: 0

### 오류 케이스

**Given**:
- 프로젝트에 소스 파일이 존재
- 3개 파일이 300 LOC 초과

**When**:
- `trust-checker --check=code-constraints` 실행

**Then**:
- 검증 결과: ❌ FAIL
- 출력 메시지:
  ```
  ❌ Code Constraints: 3 files exceed 300 LOC
    → 리팩토링 권장
    → 위반 파일:
      - src/installer/template-processor.ts: 342 LOC
      - src/core/git-manager.ts: 315 LOC
      - src/utils/file-helper.ts: 305 LOC
  ```
- Exit code: 1

---

## TC-003: 함수 ≤50 LOC 검증

### 정상 케이스

**Given**:
- 프로젝트에 함수가 정의됨
- 모든 함수가 50 LOC 이하

**When**:
- `trust-checker --check=code-constraints` 실행

**Then**:
- 검증 결과: ✅ PASS
- 출력 메시지:
  ```
  ✅ Code Constraints: All functions ≤50 LOC
  ```
- Exit code: 0

### 오류 케이스

**Given**:
- 프로젝트에 함수가 정의됨
- 2개 함수가 50 LOC 초과

**When**:
- `trust-checker --check=code-constraints` 실행

**Then**:
- 검증 결과: ❌ FAIL
- 출력 메시지:
  ```
  ❌ Code Constraints: 2 functions exceed 50 LOC
    → 리팩토링 권장
    → 위반 함수:
      - src/utils/file-helper.ts:45 - processFiles(): 58 LOC
      - src/core/git-manager.ts:102 - executeTddCycle(): 63 LOC
  ```
- Exit code: 1

---

## TC-004: 매개변수 ≤5개 검증

### 정상 케이스

**Given**:
- 프로젝트에 함수가 정의됨
- 모든 함수의 매개변수가 5개 이하

**When**:
- `trust-checker --check=code-constraints` 실행

**Then**:
- 검증 결과: ✅ PASS
- 출력 메시지:
  ```
  ✅ Code Constraints: All functions have ≤5 parameters
  ```
- Exit code: 0

### 오류 케이스

**Given**:
- 프로젝트에 함수가 정의됨
- 1개 함수가 매개변수 6개

**When**:
- `trust-checker --check=code-constraints` 실행

**Then**:
- 검증 결과: ❌ FAIL
- 출력 메시지:
  ```
  ❌ Code Constraints: 1 function has >5 parameters
    → 리팩토링 권장 (객체로 그룹화)
    → 위반 함수:
      - src/core/installer.ts:78 - createProject(name, path, lang, mode, author, version): 6 parameters
  ```
- Exit code: 1

---

## TC-005: 순환 복잡도 ≤10 검증

### 정상 케이스

**Given**:
- 프로젝트에 조건문/반복문이 있음
- 모든 함수의 복잡도가 10 이하

**When**:
- `trust-checker --check=code-constraints` 실행

**Then**:
- 검증 결과: ✅ PASS
- 출력 메시지:
  ```
  ✅ Code Constraints: All functions have cyclomatic complexity ≤10
  ```
- Exit code: 0

### 오류 케이스

**Given**:
- 프로젝트에 조건문/반복문이 많음
- 1개 함수의 복잡도가 14

**When**:
- `trust-checker --check=code-constraints` 실행

**Then**:
- 검증 결과: ❌ FAIL
- 출력 메시지:
  ```
  ❌ Code Constraints: 1 function has cyclomatic complexity >10
    → 리팩토링 권장 (함수 분리)
    → 위반 함수:
      - src/core/validator.ts:120 - validateAllSpecs(): complexity 14
  ```
- Exit code: 1

---

## TC-006: @TAG 체인 완전성 검증

### 정상 케이스

**Given**:
- 코드에 `@SPEC:AUTH-001`, `@TEST:AUTH-001`, `@CODE:AUTH-001`이 모두 존재

**When**:
- `trust-checker --check=tag-chain` 실행

**Then**:
- 검증 결과: ✅ PASS
- 출력 메시지:
  ```
  ✅ TAG Chain: All TAGs are properly linked
    → SPEC: 15 | TEST: 15 | CODE: 15 | DOC: 12
  ```
- Exit code: 0

### 오류 케이스

**Given**:
- 코드에 `@CODE:AUTH-003`은 있으나 `@SPEC:AUTH-003`이 없음

**When**:
- `trust-checker --check=tag-chain` 실行

**Then**:
- 검증 결과: ⚠️ WARNING
- 출력 메시지:
  ```
  ⚠️ TAG Chain: 2 orphan TAGs found
    → /alfred:3-sync 실행 전 TAG 수정 필요
    → 고아 TAG:
      - @CODE:AUTH-003 (no @SPEC:AUTH-003) at src/auth/service.ts:45
      - @SPEC:USER-005 (no @CODE:USER-005) at .moai/specs/SPEC-USER-005/spec.md:1
  ```
- Exit code: 0 (WARNING은 통과, --fail-on-warning 옵션 시 exit 1)

---

## TC-007: 고아 TAG 탐지

### 정상 케이스

**Given**:
- 모든 `@CODE:ID`가 `@SPEC:ID`와 연결됨

**When**:
- `trust-checker --check=tag-chain` 실행

**Then**:
- 검증 결과: ✅ PASS
- 출력 메시지:
  ```
  ✅ TAG Chain: No orphan TAGs
  ```
- Exit code: 0

### 오류 케이스

**Given**:
- `@CODE:PAYMENT-007`은 있으나 `@SPEC:PAYMENT-007`이 없음

**When**:
- `trust-checker --check=tag-chain` 실행

**Then**:
- 검증 결과: ⚠️ WARNING
- 출력 메시지:
  ```
  ⚠️ TAG Chain: 1 orphan TAG found
    → 고아 TAG:
      - @CODE:PAYMENT-007 (no @SPEC:PAYMENT-007) at src/payment/gateway.ts:89
    → 권장 조치:
      1. SPEC-PAYMENT-007 문서 작성
      2. 또는 @CODE:PAYMENT-007 제거
  ```
- Exit code: 0

---

## TC-008: 검증 결과 보고서 생성

### 정상 케이스

**Given**:
- trust-checker 실행 완료

**When**:
- `trust-checker --report-format=markdown` 실행

**Then**:
- 보고서 파일 생성: `.moai/reports/trust-report-2025-10-16-143000.md`
- 보고서 구조:
  ```markdown
  # TRUST Validation Report

  **Generated**: 2025-10-16 14:30:00
  **Project**: MoAI-ADK
  **Language**: TypeScript

  ## Summary
  - ✅ Test Coverage: 87% (Target: 85%)
  - ❌ Code Constraints: 3 violations
  - ✅ Type Safety: 100%
  - ✅ Security Scan: 0 vulnerabilities
  - ⚠️ TAG Chain: 2 orphan TAGs

  ## Details
  ...
  ```

### JSON 형식

**Given**:
- trust-checker 실행 완료

**When**:
- `trust-checker --report-format=json` 실행

**Then**:
- 보고서 파일 생성: `.moai/reports/trust-report-2025-10-16-143000.json`
- JSON 스키마:
  ```json
  {
    "timestamp": "2025-10-16T14:30:00Z",
    "project": "MoAI-ADK",
    "language": "TypeScript",
    "summary": {
      "testCoverage": { "status": "PASS", "value": 87 },
      "codeConstraints": { "status": "FAIL", "violations": 3 },
      "typeSafety": { "status": "PASS", "errors": 0 },
      "securityScan": { "status": "PASS", "vulnerabilities": 0 },
      "tagChain": { "status": "WARNING", "orphans": 2 }
    },
    "details": { ... }
  }
  ```

---

## TC-009: 검증 실패 시 구체적 오류 메시지

### 테스트 시나리오

**Given**:
- 테스트 커버리지가 78% (목표 85%)

**When**:
- `trust-checker` 실행

**Then**:
- 오류 메시지에 다음 정보 포함:
  1. 현재 커버리지 수치 (78%)
  2. 목표 커버리지 수치 (85%)
  3. 미달 파일 목록 (커버리지 낮은 순 정렬)
  4. 권장 조치 (추가 테스트 케이스 작성)
- 출력 메시지:
  ```
  ❌ Test Coverage: 78% (Target: 85%)
    → 추가 테스트 케이스 작성 권장
    → 낮은 커버리지 파일:
      - src/utils/helper.ts: 72% (23/32 lines covered)
      - src/core/validator.ts: 78% (45/58 lines covered)
      - src/installer/processor.ts: 80% (120/150 lines covered)
  ```

---

## TC-010: 언어별 도구 자동 선택

### TypeScript 프로젝트

**Given**:
- `.moai/config.json`에 `project.language: "typescript"` 정의

**When**:
- `trust-checker` 실행

**Then**:
- 다음 도구 자동 선택:
  - 테스트: Vitest
  - 커버리지: Vitest (--coverage)
  - 린터: Biome
  - 타입 체커: tsc
- 출력 메시지:
  ```
  ℹ️ Detected language: TypeScript
  ℹ️ Using tools: Vitest, Biome, tsc
  ```

### Python 프로젝트

**Given**:
- `.moai/config.json`에 `project.language: "python"` 정의

**When**:
- `trust-checker` 실행

**Then**:
- 다음 도구 자동 선택:
  - 테스트: pytest
  - 커버리지: coverage.py
  - 린터: ruff
  - 타입 체커: mypy
- 출력 메시지:
  ```
  ℹ️ Detected language: Python
  ℹ️ Using tools: pytest, coverage.py, ruff, mypy
  ```

---

## 통합 시나리오

### IS-001: /alfred:2-run 완료 후 자동 실행

**Given**:
- `/alfred:2-run AUTH-001` 실행
- GREEN 단계 커밋 완료

**When**:
- trust-checker 자동 실행

**Then**:
- 5개 검증 항목 순차 실행:
  1. [1/5] Test Coverage 검증
  2. [2/5] Code Constraints 검증
  3. [3/5] Type Safety 검증
  4. [4/5] Security Scan 검증
  5. [5/5] TAG Chain 검증
- 검증 통과 시: REFACTOR 단계 진행
- 검증 실패 시: 오류 메시지 표시 + 커밋 차단

### IS-002: SPEC-VALID-001과 연계

**Given**:
- SPEC-VALID-001 (메타데이터 검증) 구현 완료

**When**:
- `trust-checker` 실행

**Then**:
- SPEC 메타데이터 검증 포함:
  - YAML Front Matter 필수 필드 확인
  - HISTORY 섹션 존재 확인
  - 버전 체계 검증 (Semantic Versioning)
- 출력 메시지:
  ```
  ℹ️ Running SPEC metadata validation (SPEC-VALID-001)
  ✅ SPEC Metadata: 15/15 specs valid
  ```

---

## 성능 벤치마크

### 소규모 프로젝트 (<100 파일)

**목표**: 전체 검증 ≤10초

**측정 항목**:
- TAG 스캔: ≤2초
- 테스트 커버리지: ≤3초
- 코드 제약: ≤2초
- 타입 검증: ≤2초
- 보안 스캔: ≤1초

### 대규모 프로젝트 (1000+ 파일)

**목표**: 전체 검증 ≤30초

**최적화 전략**:
- 병렬 실행 (5개 검증 동시)
- 증분 검증 (Git diff 기반)
- 캐싱 (이전 검증 결과 재사용)

---

## 품질 게이트

### 필수 기준 (모두 통과해야 PASS)
- [ ] 테스트 커버리지 ≥85%
- [ ] 파일 ≤300 LOC
- [ ] 함수 ≤50 LOC
- [ ] 매개변수 ≤5개
- [ ] 순환 복잡도 ≤10
- [ ] 보안 취약점 0개 (High/Critical)

### 경고 기준 (통과 가능, 개선 권장)
- [ ] TAG 체인 완전성 (고아 TAG 있을 시 WARNING)
- [ ] 타입 오류 (TypeScript strict 모드)

---

## Definition of Done

- [ ] 10개 테스트 케이스 모두 통과
- [ ] 통합 시나리오 2개 통과
- [ ] 성능 벤치마크 충족 (≤10초)
- [ ] 보고서 생성 확인 (Markdown + JSON)
- [ ] /alfred:2-run 통합 확인
- [ ] 커밋 차단 메커니즘 동작 확인

---

**Last Updated**: 2025-10-16
**Author**: @Goos
