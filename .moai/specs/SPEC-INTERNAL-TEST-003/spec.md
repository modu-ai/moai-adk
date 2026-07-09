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

# SPEC-INTERNAL-TEST-003 — i18n 사전 누락 보충: workflow.agentic_loop.max_iterations

## §A. Context / Problem Statement

`internal/web/` 패키지의 i18n 계약 테스트 2개가 FAIL 상태다. 원인은 기계적이고 단일 지점에 있다: 스키마 필드 `workflow.agentic_loop.max_iterations`가 웹 콘솔 렌더 시 `f.workflow.agentic_loop.max_iterations.title` / `.desc` 라는 `data-i18n` 훅을 방출하지만, i18n 사전(`internal/web/assets/i18n.js`)이 이 키에 대한 항목을 4개 로컬(en/ko/ja/zh) 모두에서 누락하고 있다.

### Verified failure evidence (orchestrator-ran, not inferred)

명령: `go test ./internal/web/ -run 'TestDataI18nKeysSubsetOfDictionary|TestI18nKeySetParity' -count=1`
결과: 2 FAIL. 축어 출력은 `/tmp/moai-verify/web-i18n-test.log`에 보존됨.

두 실패 주장:
1. `TestDataI18nKeysSubsetOfDictionary` (`internal/web/i18n_test.go:267`)
   - `data-i18n key "f.workflow.agentic_loop.max_iterations.title" in the rendered page is absent from the dictionary (R6: would render blank/untranslated)`
   - `data-i18n key "f.workflow.agentic_loop.max_iterations.desc" in the rendered page is absent from the dictionary (R6: would render blank/untranslated)`
2. `TestI18nKeySetParity` (`internal/web/schema_label_test.go:101`)
   - `i18n.js missing key "f.workflow.agentic_loop.max_iterations.title" in all 4 locales (schema field "workflow.agentic_loop.max_iterations")`
   - `i18n.js missing key "f.workflow.agentic_loop.max_iterations.desc" in all 4 locales (schema field "workflow.agentic_loop.max_iterations")`

### Why this matters

- **R6 boundary 위반**: 렌더된 페이지의 data-i18n 키가 사전에 없으면 해당 엘리먼트는 blank/untranslated로 렌더된다. 이는 웹 콘솔 사용자 경험 직결 결함이다.
- **아키텍처 SPEC 선결 조건**: 본 SPEC은 `SPEC-INTERNAL-ARCH-001`의 "sufficient" 선행 조건이다. ARCH-001의 plan-audit D1-D9 재진입은 whole-repo `go test` exit 0을 요구하며, 이 2개 FAIL이 남은 한 그 조건이 충족되지 않는다.
- **계보 잔여 부채**: SPEC-INTERNAL-TEST-002의 M1이 cli golden testdata를 rc6→rc7로 재생성하고 M3가 `pipeline.go` 테스트를 추가했으나, 웹 i18n 사전 동기화는 누락되었다. 본 SPEC이 해당 부채를 청산한다.

## §B. Scope

### In Scope

- `internal/web/assets/i18n.js` 단일 파일에 4-locale × 2-key = 8개 문자열 항목 추가.
- 추가 대상 키 (각 4 locale):
  - `f.workflow.agentic_loop.max_iterations.title`
  - `f.workflow.agentic_loop.max_iterations.desc`
- 번역은 semantic sibling인 `f.workflow.loop_prevention.max_iterations.{title,desc}`의 패턴을 반영하되, `agentic_loop`의 의미(pipeline-level completion-loop iteration ceiling)를 정확히 전달해야 한다.

### Out of Scope

#### Out of Scope — 스키마 변경

- `internal/settings/schema_sections.go:193`의 `workflow.agentic_loop.max_iterations` 필드 정의는 이미 존재하며 올바르다 (TypeInt, I18nKey 자동 파생). 스키마 변경 없음.
- `internal/config/types.go`의 `AgenticLoopConfig` 구조체 및 `internal/config/agentic_loop_distinctness_test.go`의 distinctness 계약도 이미 존재하며 이 SPEC의 범위가 아니다.

#### Out of Scope — 테스트 코드 변경

- 두 계약 테스트(`TestDataI18nKeysSubsetOfDictionary`, `TestI18nKeySetParity`)의 코드는 변경하지 않는다. 이 테스트들은 올바른 계약을 강제하고 있으며, 실패 원인은 사전 누락이지 테스트 결함이 아니다.

#### Out of Scope — 웹 콘솔 위젯/렌더 로직 변경

- `internal/web/schema_label.go`, `internal/web/fieldsets_templ.go`, `internal/web/schemaform.go` 등 렌더링 로직은 이미 `f.<field>.title` / `f.<field>.desc` data-i18n 훅을 정상 방출하고 있다 (실패 메시지가 "rendered page"에 방출된 키를 감지한 것 자체가 증거). 렌더 로직 수정 없음.

#### Out of Scope — 다른 locale/키 동기화

- 본 SPEC은 `agentic_loop.max_iterations` 누락 1건만 다룬다. 다른 schema 필드의 i18n 키 동기화, locale 누락 점검, 번역 품질 개선 등은 별도 범위.

#### Out of Scope — ARCH-001 자체 구현

- `SPEC-INTERNAL-ARCH-001`의 아키텍처 변경은 본 SPEC이 unblock 한 이후 별도 run-phase에서 진행된다.

## §C. Requirements (GEARS)

### REQ-I18N-001 (Ubiquitous)

The `internal/web/assets/i18n.js` dictionary **shall** define both `"f.workflow.agentic_loop.max_iterations.title"` and `"f.workflow.agentic_loop.max_iterations.desc"` entries for each of the 4 locales (`en`, `ko`, `ja`, `zh`).

*측정 가능한 증거*: `grep -c '"f.workflow.agentic_loop.max_iterations.title":' internal/web/assets/i18n.js` returns `>= 4`; 동일하게 `.desc` 에 대해 `>= 4`.

### REQ-I18N-002 (Event-detected)

**When** the web console renders the schema field `workflow.agentic_loop.max_iterations`, the rendered `data-i18n="f.workflow.agentic_loop.max_iterations.title"` and `data-i18n="f.workflow.agentic_loop.max_iterations.desc"` hooks **shall** resolve to non-empty localized strings sourced from the dictionary, for every locale.

*측정 가능한 증거*: `TestDataI18nKeysSubsetOfDictionary` exits 0 (R6 contract — rendered keys ⊆ dictionary).

### REQ-I18N-003 (State-driven)

**While** the `TestI18nKeySetParity` contract runs, the i18n.js dictionary **shall** satisfy the 4-locale parity contract for every schema field whose widget emits `data-i18n` hooks (현 34-필드 위젯 + PersistSeam/PersistTypedSection 승격 필드). `agentic_loop.max_iterations`는 이 계약의 누락된 35번째 필드로서 동일한 의무를 진다.

*측정 가능한 증거*: `TestI18nKeySetParity` exits 0.

### REQ-I18N-004 (Capability gate / semantic distinctness)

**Where** the sibling key `f.workflow.loop_prevention.max_iterations.{title,desc}` exists (per-locale), the new `f.workflow.agentic_loop.max_iterations.{title,desc}` entries **shall** be semantically distinct from the sibling — `agentic_loop`은 pipeline-level completion-loop ceiling (default 10)을, `loop_prevention`은 per-operation diagnostic fix-loop bound를 의미하므로, 번역문이 구분 가능해야 한다 (copy-paste drift 금지).

*측정 가능한 증거*: 4-locale 각각에 대해 두 키의 값이 동일하지 않음 (특히 KO/JA/ZH에서 "agentic/completion"과 "loop_prevention/diagnostic"의 의미 구분이 번역에 반영됨).

## §D. Constraints (Non-Functional)

- **CON-001 (locale 순서 보존)**: i18n.js의 4-locale block 순서(en → ko → ja → zh) 및 각 block 내 키의 기존 정렬 순서를 훼손하지 않는다. 새 항목은 schema top-down 순서를 따라 `auto_clear.token_threshold.*` 이후, `loop_prevention.*` 블록 이전에 삽입하는 것을 권장한다 (스키마 `schema_sections.go:193` 줄 순서와 일치).
- **CON-002 (인용부호/문법 일치)**: 모든 새 항목은 기존 항목과 동일한 JSON 형식(`"<key>": "<value>",`)을 따른다. trailing comma, 공백, 따옴표 스타일 일관성 유지.
- **CON-003 (번역 품질 최소 bar)**: en 값은 humanized 영문 baseline (`schema_label.go` `humanizeLastSegment` / `fieldDescByType`의 출력과 양립 가능한 자연어). ko/ja/zh 값은 `loop_prevention.max_iterations`의 해당 locale 번역을 semantic anchor로 삼되, "agentic"/"completion-loop"/"pipeline"의 의미가 해당 언어의 자연스러운 표현으로 반영되어야 한다.
- **CON-004 (회귀 없음)**: 본 변경으로 인해 기존 통과하던 테스트가 FAIL로 전락하지 않는다. 특히 `TestI18nSegmentKeysRemovedFromWebDictionary`, `TestNoReviewKeys`, `TestSchemaEmptyLabelParity`가 여전히 exit 0이어야 한다.

## §E. Verification Approach

run-phase에서 다음 3-레벨 검증을 수행한다:

1. **단위 검증 (실패 2건 해소)**:
   ```bash
   go test ./internal/web/ -run 'TestDataI18nKeysSubsetOfDictionary|TestI18nKeySetParity' -count=1 -v
   ```
   기대: exit 0, 두 테스트 모두 PASS.

2. **패키지 회귀 검증**:
   ```bash
   go test ./internal/web/ -count=1
   ```
   기대: exit 0 (패키지 내 다른 테스트의 회귀 없음).

3. **whole-repo 선결 조건 검증 (ARCH-001 unblock)**:
   ```bash
   go test ./...
   ```
   기대: exit 0. 이것이 ARCH-001 plan-audit D1 재진입의 전제 조건이다.

run-phase에서는 테스트 출력을 `/tmp/moai-verify/` 또는 `.moai/state/verify/<session>/`에 redirect하여 `verification-claim-integrity.md §1.1 surface 2` (manager-agent §E self-verification)의 evidence-reachability 의무를 충족한다.

## §F. Open Questions

없음. 원인이 기계적이고 단일 지점이며, sibling 패턴이 명확하다.

## §G. Dependencies

- **선행 (resolved)**: SPEC-INTERNAL-TEST-002 — 계보상 선행 SPEC. 본 SPEC은 TEST-002 M1/M3가 남긴 잔여 부채를 청산.
- **후행 (blocked by this)**: SPEC-INTERNAL-ARCH-001 — plan-audit D1-D9 재진입이 whole-repo `go test` exit 0을 요구. 본 SPEC이 그 마지막 2개 FAIL을 제거한다.
- **외부**: 무관. i18n.js는 순수 정적 asset이며, 다른 패키지와의 런타임 의존성 없음.

## §H. History

- 2026-07-09: 최초 draft 작성 (manager-spec). orchestrator가 `/tmp/moai-verify/web-i18n-test.log`에 축어 실패 출력을 보존한 상태에서 plan-phase 위임.
