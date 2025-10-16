---
id: TRUST-001
version: 0.0.1
status: draft
created: 2025-10-16
updated: 2025-10-16
author: @Goos
priority: high
category: feature
labels:
  - quality
  - validation
  - automation
  - trust
depends_on:
  - VALID-001
related_specs:
  - VALID-001
  - INIT-003
scope:
  packages:
    - moai-adk-ts/src/core/quality
  files:
    - trust-checker.ts
    - coverage-validator.ts
    - code-constraints-validator.ts
    - tag-chain-validator.ts
---

# @SPEC:TRUST-001: TRUST 원칙 자동 검증

## HISTORY
### v0.0.1 (2025-10-16)
- **INITIAL**: TRUST 5원칙 자동 검증 시스템 명세 최초 작성
- **AUTHOR**: @Goos
- **SCOPE**: Test First, Readable, Unified, Secured, Trackable 자동 검증
- **CONTEXT**: /alfred:2-build 완료 후 품질 게이트 자동화
- **DEPENDS_ON**: SPEC-VALID-001 (메타데이터 검증)

## Environment (환경)

**실행 컨텍스트**:
- **실행 시점**: `/alfred:2-build` TDD 구현 완료 후 자동 실행
- **실행 에이전트**: `@agent-trust-checker` (Alfred가 호출)
- **실행 위치**: 프로젝트 루트 디렉토리

**지원 언어 및 도구**:
- **TypeScript/JavaScript**: Vitest + Biome + TypeScript compiler
- **Python**: pytest + coverage.py + mypy + ruff
- **Java**: JUnit + JaCoCo + SpotBugs
- **Go**: go test + gofmt + staticcheck
- **Rust**: cargo test + rustfmt + clippy
- **Dart**: flutter test + dart analyze
- **Swift**: XCTest + swiftlint
- **Kotlin**: JUnit + detekt

**필수 도구**:
- `ripgrep` (TAG 스캔)
- 언어별 테스트 프레임워크 (자동 감지)
- Git (커밋 이력 조회)

**성능 목표**:
- 전체 검증 ≤10초 (대규모 프로젝트 ≤30초)
- TAG 스캔 ≤2초
- 커버리지 계산 ≤5초

## Assumptions (가정)

1. **프로젝트 구조**:
   - MoAI-ADK 표준 디렉토리 구조 준수
   - `.moai/config.json` 존재 및 `project.language` 정의

2. **@TAG 시스템**:
   - 코드베이스에 `@SPEC:ID`, `@TEST:ID`, `@CODE:ID` 주석 적용
   - TAG 형식: `@<TYPE>:<DOMAIN>-<NUMBER>`

3. **테스트 환경**:
   - 언어별 테스트 프레임워크가 설치됨
   - 테스트는 `tests/` 디렉토리에 위치
   - 테스트는 독립적으로 실행 가능

4. **코드 품질**:
   - 모든 파일은 UTF-8 인코딩
   - 주석은 언어별 표준 형식 사용
   - 순환 복잡도 계산 도구 사용 가능

## Requirements (요구사항)

### Ubiquitous (필수 기능)

1. **R-001**: 시스템은 TRUST 5원칙 자동 검증 기능을 제공해야 한다
   - Test First (T)
   - Readable (R)
   - Unified (U)
   - Secured (S)
   - Trackable (T)

2. **R-002**: 시스템은 언어별 최적 도구를 자동 선택해야 한다
   - `.moai/config.json`의 `project.language` 기반 선택
   - 도구 미설치 시 설치 안내 메시지 제공

3. **R-003**: 시스템은 검증 결과를 구조화된 보고서로 제공해야 한다
   - Markdown 형식 (`.moai/reports/trust-report-{timestamp}.md`)
   - JSON 형식 (CI/CD 통합용)
   - 오류 발생 시 구체적인 파일명/라인 번호 포함

### Event-driven (이벤트 기반)

4. **R-004**: WHEN `/alfred:2-build` 완료 후, 시스템은 TRUST 검증을 자동 실행해야 한다
   - GREEN 단계 커밋 직후 실행
   - 검증 실패 시 REFACTOR 단계 진입 차단

5. **R-005**: WHEN 테스트 커버리지 < 85%이면, 시스템은 오류 메시지와 함께 커밋을 차단해야 한다
   - 현재 커버리지 수치 표시
   - 미달 파일 목록 제공 (커버리지 낮은 순 정렬)

6. **R-006**: WHEN 파일 > 300 LOC 또는 함수 > 50 LOC이면, 시스템은 위반 파일 목록을 반환해야 한다
   - 위반 파일 경로 및 실제 LOC 수치
   - 리팩토링 권장 메시지

7. **R-007**: WHEN 고아 TAG가 발견되면, 시스템은 경고 메시지와 함께 TAG 수정을 권장해야 한다
   - 고아 TAG: `@CODE:ID`는 있으나 `@SPEC:ID`가 없는 경우
   - 역방향: `@SPEC:ID`는 있으나 `@CODE:ID`가 없는 경우 (구현 누락)

8. **R-008**: WHEN 매개변수 > 5개인 함수가 발견되면, 시스템은 리팩토링을 요구해야 한다
   - 함수명, 파일명, 라인 번호 표시
   - 매개변수 개수 표시

### State-driven (상태 기반)

9. **R-009**: WHILE 구현 중일 때, 시스템은 실시간 TRUST 피드백을 제공해야 한다
   - 파일 저장 시 lint 오류 표시
   - 테스트 실행 시 커버리지 표시

10. **R-010**: WHILE 검증 실행 중일 때, 시스템은 진행 상태를 표시해야 한다
    - 진행률 (예: [3/5] TAG 체인 검증 중...)
    - 예상 완료 시간 (선택적)

### Constraints (제약사항)

11. **R-011**: IF 테스트 커버리지 < 85%이면, 시스템은 PR merge를 차단해야 한다
    - GitHub Actions에서 자동 실행
    - PR 코멘트에 커버리지 보고서 추가

12. **R-012**: IF 순환 복잡도 > 10이면, 시스템은 리팩토링을 요구해야 한다
    - 복잡도 높은 함수 목록 제공
    - 리팩토링 권장사항 제시

13. **R-013**: IF 보안 취약점(High 이상)이 발견되면, 시스템은 즉시 빌드를 중단해야 한다
    - 취약점 상세 정보 (CVE ID, CVSS 점수)
    - 수정 방법 링크 제공

14. **R-014**: IF TAG 체인이 끊어지면, 시스템은 `/alfred:3-sync` 실행을 차단해야 한다
    - 끊어진 TAG 체인 표시 (예: `@SPEC:AUTH-001` → `@CODE:AUTH-001` 누락)
    - 수정 명령 제안

15. **R-015**: 전체 검증 시간은 10초를 초과하지 않아야 한다 (소규모 프로젝트 기준)
    - 대규모 프로젝트 (1000+ 파일): ≤30초
    - 검증 단계별 소요 시간 측정

### Optional Features (선택적 기능)

16. **R-016**: WHERE Team 모드이면, 시스템은 PR 코멘트로 검증 결과를 자동 추가할 수 있다
    - GitHub Actions 워크플로우 통합
    - 코멘트 템플릿: `.moai/templates/trust-report-comment.md`

17. **R-017**: WHERE CI/CD가 구성되면, 시스템은 GitHub Actions 워크플로우를 자동 생성할 수 있다
    - `.github/workflows/trust-check.yml` 자동 생성
    - PR 이벤트 시 자동 실행

## Specifications (상세 명세)

### S-001: TRUST 검증 흐름

```
/alfred:2-build 완료
    ↓
@agent-trust-checker 호출
    ↓
[1/5] T - Test Coverage 검증 (≥85%)
    ↓
[2/5] R - Code Constraints 검증 (≤300 LOC, ≤50 LOC, ≤5 params)
    ↓
[3/5] U - Type Safety 검증 (언어별)
    ↓
[4/5] S - Security Scan 검증 (취약점)
    ↓
[5/5] T - TAG Chain 검증 (고아 TAG, 순환 참조)
    ↓
검증 통과 → REFACTOR 단계 진행
검증 실패 → 오류 보고서 + 커밋 차단
```

### S-002: 언어별 도구 매핑

| 언어       | 테스트 도구   | 커버리지 도구 | 린터        | 타입 체커      |
| ---------- | ------------- | ------------- | ----------- | -------------- |
| TypeScript | Vitest        | Vitest        | Biome       | tsc            |
| Python     | pytest        | coverage.py   | ruff        | mypy           |
| Java       | JUnit         | JaCoCo        | SpotBugs    | javac          |
| Go         | go test       | go test       | staticcheck | go build       |
| Rust       | cargo test    | cargo tarpaulin | clippy    | rustc          |

### S-003: TAG 체인 검증 알고리즘

1. **TAG 스캔**:
   ```bash
   rg '@(SPEC|TEST|CODE|DOC):' -n .moai/specs/ tests/ src/ docs/
   ```

2. **TAG 쌍 매칭**:
   - `@SPEC:ID` → `@TEST:ID` 존재 확인
   - `@TEST:ID` → `@CODE:ID` 존재 확인
   - `@CODE:ID` → `@SPEC:ID` 역참조 확인

3. **고아 TAG 탐지**:
   - `@CODE:ID`는 있으나 `@SPEC:ID` 없음
   - `@SPEC:ID`는 있으나 `@CODE:ID` 없음 (구현 누락)

4. **순환 참조 탐지**:
   - `@SPEC:A` depends_on `@SPEC:B`
   - `@SPEC:B` depends_on `@SPEC:A`

### S-004: 보고서 형식

**Markdown 형식** (`.moai/reports/trust-report-{timestamp}.md`):
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

### [T] Test Coverage
- **Status**: PASS
- **Coverage**: 87%
- **Files with low coverage**:
  - src/utils/helper.ts: 72%
  - src/core/validator.ts: 78%

### [R] Code Constraints
- **Status**: FAIL
- **Violations**:
  1. src/installer/template-processor.ts: 342 LOC (Limit: 300)
  2. src/core/git-manager.ts: 315 LOC (Limit: 300)
  3. src/utils/file-helper.ts:45 - processFiles(): 58 LOC (Limit: 50)

### [U] Type Safety
- **Status**: PASS
- **Errors**: 0

### [S] Security Scan
- **Status**: PASS
- **Vulnerabilities**: 0

### [T] TAG Chain
- **Status**: WARNING
- **Orphan TAGs**:
  - @CODE:AUTH-003 (no @SPEC:AUTH-003)
  - @SPEC:USER-005 (no @CODE:USER-005)
```

**JSON 형식** (CI/CD 통합용):
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

### S-005: 오류 메시지 표준

**심각도별 아이콘**:
- ❌ Critical: 작업 중단, 즉시 조치 필요
- ⚠️ Warning: 주의 필요, 계속 진행 가능
- ℹ️ Info: 정보성 메시지, 참고용

**메시지 형식**:
```
[아이콘] [검증 항목]: [문제 설명]
  → [권장 조치]
  → [관련 파일/라인]
```

**예시**:
```
❌ Test Coverage: 현재 78% (목표 85%)
  → 추가 테스트 케이스 작성 권장
  → 낮은 커버리지 파일:
    - src/utils/helper.ts: 72%
    - src/core/validator.ts: 78%

⚠️ Code Constraints: 3개 파일이 300 LOC 초과
  → 리팩토링 권장
  → 위반 파일:
    - src/installer/template-processor.ts: 342 LOC
    - src/core/git-manager.ts: 315 LOC

ℹ️ TAG Chain: 2개 고아 TAG 발견
  → /alfred:3-sync 실행 전 TAG 수정 필요
  → 고아 TAG:
    - @CODE:AUTH-003 (no @SPEC:AUTH-003)
```

## Acceptance Criteria

### AC-001: 테스트 커버리지 ≥85% 검증
- Given: 프로젝트에 테스트가 작성됨
- When: trust-checker 실행
- Then: 커버리지 87% 이상이면 PASS, 미만이면 FAIL

### AC-002: 파일 ≤300 LOC 검증
- Given: 소스 파일이 존재
- When: trust-checker 실행
- Then: 모든 파일이 300 LOC 이하면 PASS, 초과 파일이 있으면 FAIL + 파일 목록

### AC-003: 함수 ≤50 LOC 검증
- Given: 소스 코드에 함수가 정의됨
- When: trust-checker 실행
- Then: 모든 함수가 50 LOC 이하면 PASS, 초과 함수가 있으면 FAIL + 함수 목록

### AC-004: 매개변수 ≤5개 검증
- Given: 함수에 매개변수가 정의됨
- When: trust-checker 실행
- Then: 모든 함수의 매개변수가 5개 이하면 PASS, 초과 함수가 있으면 FAIL

### AC-005: 순환 복잡도 ≤10 검증
- Given: 소스 코드에 조건문/반복문이 있음
- When: trust-checker 실행
- Then: 모든 함수의 복잡도가 10 이하면 PASS, 초과 함수가 있으면 FAIL

### AC-006: @TAG 체인 완전성 검증
- Given: 코드에 @TAG가 주석으로 작성됨
- When: trust-checker 실행
- Then: 모든 @CODE:ID가 @SPEC:ID와 연결되면 PASS, 끊어진 체인이 있으면 WARNING

### AC-007: 고아 TAG 탐지
- Given: @CODE:ID는 있으나 @SPEC:ID가 없음
- When: trust-checker 실행
- Then: 고아 TAG 목록 반환 + WARNING

### AC-008: 검증 결과 보고서 생성
- Given: trust-checker 실행 완료
- When: 보고서 생성 요청
- Then: Markdown 및 JSON 형식으로 `.moai/reports/` 디렉토리에 저장

### AC-009: 검증 실패 시 구체적 오류 메시지
- Given: 커버리지 < 85%
- When: trust-checker 실행
- Then: 오류 메시지에 현재 커버리지, 미달 파일 목록, 권장 조치 포함

### AC-010: 언어별 도구 자동 선택
- Given: `.moai/config.json`에 `project.language: "python"` 정의
- When: trust-checker 실행
- Then: pytest, coverage.py, mypy, ruff 자동 선택 및 실행

## TRUST Mapping

| TRUST 원칙 | 관련 AC              | 검증 내용                  |
| ---------- | -------------------- | -------------------------- |
| **T** est  | AC-001               | 테스트 커버리지 ≥85%       |
| **R** ead  | AC-002, AC-003, AC-004, AC-005 | 코드 제약 (LOC, 복잡도) |
| **U** nify | AC-010               | 언어별 도구 통합           |
| **S** ecure| AC-013 (R-013)       | 보안 취약점 스캔           |
| **T** race | AC-006, AC-007       | TAG 체인 검증, 고아 TAG    |

## Technical Approach

### 도구 선택

1. **TAG 스캔**: ripgrep (고성능 정규식 검색)
2. **커버리지**: 언어별 표준 도구 (coverage.py, Vitest, JaCoCo 등)
3. **코드 분석**: AST 파싱 (언어별 파서 라이브러리)
4. **보안 스캔**: Snyk, npm audit, pip-audit 등

### 성능 최적화

1. **병렬 실행**: 5개 검증 항목을 병렬로 실행 (Promise.all, asyncio.gather)
2. **증분 검증**: 변경된 파일만 검증 (Git diff 기반)
3. **캐싱**: 이전 검증 결과 캐싱 (`.moai/cache/trust-cache.json`)
4. **조기 종료**: Critical 오류 발견 시 즉시 중단

### 오류 처리

1. **도구 미설치**: 설치 안내 메시지 + 설치 명령 제공
2. **권한 오류**: 파일 접근 권한 확인 메시지
3. **타임아웃**: 검증 시간 초과 시 경고 메시지 (10초 초과)
4. **파싱 오류**: 소스 코드 파싱 실패 시 건너뛰기 + 경고

## Traceability (@TAG)

- **SPEC**: `@SPEC:TRUST-001` (본 문서)
- **TEST**: `tests/core/quality/test_trust_checker.py`
- **CODE**: `src/core/quality/trust_checker.py`
- **DOC**: `docs/quality/trust-validation.md`

## Dependencies

- **SPEC-VALID-001**: SPEC 메타데이터 검증 (필수)
- **SPEC-INIT-003**: 프로젝트 초기화 (선택적)

## Related Issues

- GitHub Issue: (생성 예정)
- PR: (구현 후 생성)

---

**Last Updated**: 2025-10-16
**Next Steps**: `/alfred:2-build TRUST-001`로 TDD 구현 시작
