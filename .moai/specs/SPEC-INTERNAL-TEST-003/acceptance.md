---
id: SPEC-INTERNAL-TEST-003
title: "Add missing i18n dictionary entries for workflow.agentic_loop.max_iterations"
version: "0.1.0"
status: completed
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.x target"
module: "internal/web/assets"
lifecycle: spec-anchored
tags: "i18n, web-console, test-fix, debt-cleanup"
tier: S
depends_on: []
related_specs: [SPEC-INTERNAL-TEST-002, SPEC-INTERNAL-ARCH-001]
---

# SPEC-INTERNAL-TEST-003 — Acceptance Criteria

## §D. AC Matrix

본 SPEC은 Tier S (minimal) 이므로 6개 AC로 계약을 완전히 커버한다. 각 AC는 observable evidence를 강제하며, carry-over 추정 금지 (`verification-claim-integrity.md §1.1 surface 2` 의거).

| AC ID | Requirement ref | Severity | Description |
|-------|-----------------|----------|-------------|
| AC-001 | REQ-I18N-001 | MUST | i18n.js에 `.title` 키가 4 locale 모두에 존재 |
| AC-002 | REQ-I18N-001 | MUST | i18n.js에 `.desc` 키가 4 locale 모두에 존재 |
| AC-003 | REQ-I18N-002 | MUST | `TestDataI18nKeysSubsetOfDictionary` exit 0 |
| AC-004 | REQ-I18N-003 | MUST | `TestI18nKeySetParity` exit 0 |
| AC-005 | REQ-I18N-004 | MUST | agentic_loop vs loop_prevention 4-locale semantic distinctness |
| AC-006 | CON-004 | MUST | whole-repo `go test ./...` exit 0 (ARCH-001 unblock) |

### §D.1 AC 상세 (Given-When-Then)

#### AC-001 — `.title` 키 4-locale 존재

**Given** `internal/web/assets/i18n.js` 파일이 주어졌을 때,
**When** `grep -c '"f.workflow.agentic_loop.max_iterations.title":' internal/web/assets/i18n.js` 명령을 실행하면,
**Then** 결과값이 `4` 이상이어야 한다 (en, ko, ja, zh 각 블록에 1회 이상).

**Evidence**: grep 명령의 축어 출력.

#### AC-002 — `.desc` 키 4-locale 존재

**Given** `internal/web/assets/i18n.js` 파일이 주어졌을 때,
**When** `grep -c '"f.workflow.agentic_loop.max_iterations.desc":' internal/web/assets/i18n.js` 명령을 실행하면,
**Then** 결과값이 `4` 이상이어야 한다.

**Evidence**: grep 명령의 축어 출력.

#### AC-003 — R6 boundary 계약 충족

**Given** `agentic_loop.max_iterations` 스키마 필드가 렌더링되어 `data-i18n="f.workflow.agentic_loop.max_iterations.title"` 및 `.desc` 훅을 방출하는 웹 콘솔 페이지가 주어졌을 때,
**When** `go test ./internal/web/ -run 'TestDataI18nKeysSubsetOfDictionary' -count=1 -v` 명령을 실행하면,
**Then** exit code 0, 테스트 결과 PASS. 축어 출력에 "FAIL" 문자열이 없어야 한다.

**Evidence**: `/tmp/moai-verify/test-003-e1.log` (또는 `.moai/state/verify/<session>/test-003-e1.log`)의 축어 출력.

#### AC-004 — 4-locale parity 계약 충족

**Given** 모든 schema 필드에 대해 `.title` / `.desc` 키가 4 locale에 존재해야 하는 parity 계약이 주어졌을 때,
**When** `go test ./internal/web/ -run 'TestI18nKeySetParity' -count=1 -v` 명령을 실행하면,
**Then** exit code 0, 테스트 결과 PASS.

**Evidence**: AC-003과 동일한 로그 파일의 축어 출력 (또는 별도로 분리 실행한 경우 해당 로그).

#### AC-005 — semantic distinctness (copy-paste drift 방지)

**Given** `f.workflow.agentic_loop.max_iterations.{title,desc}` 및 `f.workflow.loop_prevention.max_iterations.{title,desc}` 두 sibling 키 쌍이 각 locale에 존재할 때,
**When** 각 locale (en/ko/ja/zh)에 대해 두 키의 value를 비교하면,
**Then** 두 value가 동일하지 않아야 한다 (적어도 one token의 의미적 차이 — "agentic" / "completion" / "pipeline" vs "prevention" / "diagnostic" / "per-operation"의 반영).

**Evidence**: locale별 두 키 값을 나란히 출력한 grep diff. 예시 (KO):
```
ko:
  "f.workflow.agentic_loop.max_iterations.title":    "<pipeline completion-loop 의미의 자연어>"
  "f.workflow.loop_prevention.max_iterations.title": "<diagnostic fix-loop 의미의 자연어>"
```

#### AC-006 — whole-repo 회귀 없음 + ARCH-001 unblock

**Given** 본 SPEC의 변경이 `internal/web/assets/i18n.js` 단일 파일에 국한되었을 때,
**When** `go test ./...` 명령을 whole-repo에 대해 실행하면,
**Then** exit code 0 (어떤 패키지에서도 FAIL 없음). 이것이 `SPEC-INTERNAL-ARCH-001` plan-audit D1-D9 재진입의 sufficient 전제 조건이다.

**Evidence**: `/tmp/moai-verify/test-003-e3.log` 또는 `.moai/state/verify/<session>/test-003-e3.log`의 축어 출력. exit code 관측 필수.

### §D.2 Indirect Verification (허용되는 보조 증거)

직접 AC 외에 다음 간접 증거가 회귀 없음을 보강한다 (MUST가 아닌 SHOULD):

- `go test ./internal/web/ -count=1` 전체 패키지 exit 0 — §E.2 (AC-006보다 좁은 범위).
- `go vet ./internal/web/...` exit 0 — 단순 형식/문법 검증 (i18n.js는 JS라 직접 vet 대상은 아니지만, 패키지 빌드 정상의 보조 지표로 `go build ./internal/web/...`).

### §D.3 Edge Cases

| Edge case | Expected behavior |
|-----------|-------------------|
| i18n.js의 trailing comma 관례 | 새 항목도 기존과 동일하게 trailing comma (`",..."` 형식). 마지막 항목 여부와 무관하게 comma 추가 (JS object literal 허용) |
| 새 항목의 물리적 위치 | 테스트는 키 존재 여부만 검사하므로 위치는 자유이나, 가독성을 위해 schema 순서(`auto_clear.token_threshold` 이후, `loop_prevention.*` 이전)를 따를 것을 권장 |
| 번역 미정 locale (예: ko의 "agentic loop" 자연어 표현 부재) | 기존 i18n.js의 ko/ja/zh 번역 스타일을 따름; 필요시 "에이전트 완성 루프 최대 반복" (KO) / "最大完了ループ反復" (JA) / "最大完成循环迭代" (ZH) 정도의 자연어 후보 (확정은 run-phase에서 기존 번역 톤과 비교 후 결정) |
| 4-locale block 외 실수로 다른 변수/구조체 수정 | 절대 금지 — `git diff internal/web/assets/i18n.js`가 오직 8개 신규 라인만 추가함을 확인 |

### §D.4 Closure Gates (Definition of Done)

본 SPEC이 "completed"로 전환되기 위해 다음 4개 게이트가 모두 충족되어야 한다:

- [ ] **Gate G1 (AC-001, AC-002)**: `grep -c` 결과 각각 `>= 4`. 축어 출력 보존.
- [ ] **Gate G2 (AC-003, AC-004)**: 두 테스트 exit 0. `/tmp/moai-verify/test-003-e1.log` 보존.
- [ ] **Gate G3 (AC-005)**: 4-locale grep diff에서 두 sibling 키 값이 모두 상이함.
- [ ] **Gate G4 (AC-006)**: `go test ./...` exit 0. `/tmp/moai-verify/test-003-e3.log` 보존. — 이것이 `SPEC-INTERNAL-ARCH-001` unblock의 직접 증거.

### §D.5 Forward-Looking Checks (run-phase 종료 시 확인)

- [ ] `internal/web/` 패키지에 새로운 lint warning/error 없음 (`golangci-lint run ./internal/web/...` 단, i18n.js는 JS라 직접 대상 아님 — `go vet ./internal/web/...`로 대체).
- [ ] `git diff --stat`이 `internal/web/assets/i18n.js` 단일 파일만 표시 (unrelated changes는 본 SPEC과 분리됨).
- [ ] commit message가 Conventional Commits 형식 (`fix(i18n): add missing agentic_loop.max_iterations dictionary entries for 4 locales` 정도).

### §D.6 Severity / Traceability Matrix

| AC | REQ | Test/mechanism | Tier S 적합성 |
|----|-----|----------------|----------------|
| AC-001 | REQ-I18N-001 | grep count | minimal |
| AC-002 | REQ-I18N-001 | grep count | minimal |
| AC-003 | REQ-I18N-002 | `TestDataI18nKeysSubsetOfDictionary` | 기존 테스트 (신규 작성 불필요) |
| AC-004 | REQ-I18N-003 | `TestI18nKeySetParity` | 기존 테스트 (신규 작성 불필요) |
| AC-005 | REQ-I18N-004 | grep diff (수동 검증) | minimal |
| AC-006 | CON-004 + ARCH-001 선행 | `go test ./...` | minimal |

### §D.7 Quality Gate Criteria

Tier S (minimal) 이므로 표준 게이트를 최소화:

- **Testability**: 기존 2개 계약 테스트가 PASS (AC-003, AC-004). 신규 테스트 작성 불필요 — 이 테스트들이 이미 올바른 계약을 강제 중.
- **Readability**: 새 i18n 항목의 번역이 locale별로 자연스러움 (native review는 선택).
- **Unified**: JSON 형식/들여쓰기 기존 항목과 일관.
- **Secured**: N/A (정적 번역 문자열, 보안 민감 정보 없음).
- **Trackable**: commit message가 변경을 명확히 서술.

## §E. Out of Scope (acceptance verification 제외)

본 acceptance.md는 본 SPEC의 in-scope 범위만 검증한다. 다음은 verification 제외:

- `SPEC-INTERNAL-ARCH-001` 자체의 AC (별도 SPEC).
- i18n.js의 다른 locale/키 동기화 상태.
- 번역의 문학적 품질 (min bar: REQ-I18N-004 distinctness + 기존 locale 톤 일관성).

## §F. Evidence Persistence

모든 AC evidence는 다음 중 한 곳에 보존되어야 audit-time reachability (`verification-claim-integrity.md §1.1 surface 2`)를 충족한다:

- 1차: `/tmp/moai-verify/test-003-*.log` (run-phase working location)
- 영속 (post-run, `/tmp` clear 대비): `.moai/state/verify/<session>/test-003-*.log` (copy step 권장)

manager-develop의 `§E` self-verification 블록은 위 경로를 인용하여 claim이 audit 가능함을 보증.
