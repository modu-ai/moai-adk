---
id: SPEC-HOOK-EVENT-REGISTRY-001
title: "Hook 이벤트 레지스트리에 3개 공식 observe-only 이벤트 추가 (PostToolBatch / UserPromptExpansion / MessageDisplay)"
version: "0.2.0"
status: completed
created: 2026-06-03
updated: 2026-06-03
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/hook"
lifecycle: spec-anchored
tier: S
tags: "hook, event-registry, observe-only, claude-code-alignment, tdd"
---

## HISTORY

- v0.1.0 (2026-06-03, manager-spec): 최초 plan-phase 작성. SPEC-CC-DOCS-ALIGNMENT-001(33개 문서 결함)에서 분리된 단일 Go-코드 결함 — 공식 Claude Code Hooks 가이드에 존재하나 MoAI Go 레지스트리에 `EventType` 상수/엔트리가 없는 3개 현행 공식 이벤트(`PostToolBatch`, `UserPromptExpansion`, `MessageDisplay`)를 observe-only 패턴으로 추가.
- v0.1.1 (2026-06-03, manager-spec): plan-auditor FAIL 0.71 결함 패치. **D1(BLOCKING)** — `CoverageTable`에 3개 observe-only 행 추가 시 `TestAuditThreeWaySync`(audit_test.go:296)가 결정론적으로 `HOOK_SYNC_DRIFT` 발생(audit_test.go:303-308 루프가 Resolution∉{COMPOSITE,REMOVE}을 Go-registered로 취급 → 3개가 settings.json에도 retiredEventNames(retired_events.go:17)에도 deregisteredButLiveEventNames(audit_test.go:284)에도 없어 :356에서 fail) → REQ-HER-007 + AC-HER-007 신설(`deregisteredButLiveEventNames` audit_test.go:284에 3개 이름 추가, M2/M3 명시 deliverable). **D5(major, cross-package)** — `internal/cli/doctor_hook_test.go`가 테이블에 결합: `:35` eventCount(non-composite) `!= 26` → 29 BREAK, `:125` `summary.RetireObsOnly != 4` → 7 BREAK, `:108` observability filter `len != 4` → 7 BREAK → In-Scope/References에 추가 + REQ-HER-008 신설 + AC-HER-006a 검증을 `./internal/cli/...`로 확장. **D3(major)** — `HandlerFile` placeholder hedge 제거, observe-only 일관성 위해 빈 문자열 `""` 확정(어떤 테스트도 HandlerFile에 os.Stat 안 함 — auditor 확인). **D4(minor)** — AC-HER-004b를 `>= 2` → `== 2`로 정밀화(정확히 line 37 주석 + line 39 ANCHOR 2곳). **D2(major)** — plan.md §B/§F M3의 drift "가능성" 표현을 "결정론적 확정 + deregisteredButLiveEventNames append 확정 remediation"으로 재작성(plan.md에서 처리).
- v0.1.2 (2026-06-03, manager-spec): plan-auditor iter-2 FAIL 0.79 결함 패치 (D1-D5 RESOLVED 검증됨; 신규 단일 must-fix). **D6(must-fix, 7th coupled file)** — orchestrator의 exhaustive grep sweep가 `internal/cli/hook_e2e_test.go` 누락을 적발(이것이 AC-HER-006a의 `go test ./internal/cli/...`를 결정론적으로 fail시킴). 두 breakage: (a) `:202` `len(validTypes) != 26`(에러 메시지 :203) → 29 BREAK, (b) `:286-340` `TestHookValidEventTypes_AllHaveSubcommands`가 `:291` `excludedEvents`(현재 empty map)를 사용 — 3개 observe-only 이벤트는 subcommand 없음(Exclusion "CLI/cobra subcommand 추가 금지")이므로 `:294-321` `eventToSubcmd`가 **아니라** `excludedEvents`에 추가해야 :334 "no expected subcommand mapping" 에러 회피. → REQ-HER-009 + AC-HER-009 신설, In-Scope/References에 `hook_e2e_test.go`(:202/:203/:291) 추가, AC-HER-006a 검증 command를 full-package(`go test ./internal/hook/... ./internal/cli/...`)로 canonical화. **Verified-safe (무변경, plan.md §G 문서화)**: cohabitation_guard_test.go:112(string-presence) / doctor_hook_test.go:23(`len != len(hook.CoverageTable)` relative) / doctor_hook_test.go:132(`Remove want 0` 불변, RETIRE-OBS-ONLY≠REMOVE) / hook_e2e_test.go:212-217(subset membership) / :219-226(dup) / :265-282(determinism). **파일 카운트 정정**: production 2 + test 5 = 7 files. 단 test 5 중 cohabitation_guard_test.go는 verified-no-edit이므로 **edit 받는 test 파일은 4개**(types_test.go / audit_test.go / doctor_hook_test.go / hook_e2e_test.go).

---

## 1. 개요 (Overview)

### 1.1 배경 (WHY)

공식 Claude Code Hooks 가이드와 MoAI-ADK를 대조한 감사 결과, 3개의 **현행 공식 hook 이벤트**가 MoAI hook 레지스트리에 Go `EventType` 상수도 핸들러도 없는 상태로 누락되어 있음이 확인되었다. 동시에 MoAI 문서 레이어(`.claude/rules/moai/core/hooks-system.md`)는 이미 이 3개 이벤트를 문서화하고 있어 — **Go 코드가 문서보다 뒤처진(doc-ahead-of-code) drift** 상태이다.

검증된 문서 레이어 선행 사실 (`.claude/rules/moai/core/hooks-system.md`):

- `PostToolBatch` — line 24, line 71, line 109 ("Runs after a batch of parallel tool calls resolves (v2.1.89+)")
- `UserPromptExpansion` — line 25, line 69, line 110 ("Runs when slash command expands into prompt (v2.1.90+)")
- `MessageDisplay` — line 83 ("Runs while assistant message text is displayed (v2.1.152+) … No MoAI handler registered by default.")

본 SPEC은 이 drift의 **Go-코드 측면만** 해소한다. 문서 33개 결함은 별도 SPEC(SPEC-CC-DOCS-ALIGNMENT-001) 소관이며, 문서는 이미 정확하므로 본 SPEC은 **`hooks-system.md`를 재편집하지 않는다**.

### 1.2 목표 (WHAT)

`PostToolBatch`, `UserPromptExpansion`, `MessageDisplay` 3개 이벤트를 기존 observe-only(RETIRE-OBS-ONLY) 패턴을 모방하여 **non-blocking observe-only 상수**로 hook 레지스트리에 추가한다. 즉:

- `internal/hook/types.go`의 `EventType` 상수 블록에 3개 상수 추가
- `ValidEventTypes()` 슬라이스에 3개 멤버 추가 (26 → 29)
- `internal/hook/coverage_table.go`의 `CoverageTable`에 3개 엔트리 추가 (`HandlerFile: ""`)
- `coverage_table.go:37`의 헤더 주석 + `@MX:ANCHOR`의 이벤트 수를 26 → 29로 갱신
- `internal/hook/audit_test.go:284` `deregisteredButLiveEventNames`에 3개 추가 (D1 — 3-way sync drift 방지)
- `internal/cli/doctor_hook_test.go` 3개 단언 갱신 (D5 — non-composite 26→29 / RetireObsOnly 4→7 / observability filter 4→7)
- `internal/cli/hook_e2e_test.go` 2개 갱신 (D6 — `:202` count 26→29 / `:291` `excludedEvents`에 3개 observe-only 추가)

### 1.3 검증된 현재 상태 (Ground Truth — Read/grep으로 실측)

| 대상 | 파일 | 실측 현재값 | 목표값 |
|------|------|------------|--------|
| `EventType` 상수 개수 | `internal/hook/types.go` (lines 20-111) | **26** | **29** |
| `ValidEventTypes()` 슬라이스 멤버 수 | `internal/hook/types.go` (lines 115-142) | **26** | **29** |
| `CoverageTable` 행 수 (`len()`) | `internal/hook/coverage_table.go` (lines 41-72) | **27** (실제 이벤트 26 + 합성 "AutoUpdate" COMPOSITE 행 1) | **30** (27 + 신규 3) |
| 헤더 주석 + `@MX:ANCHOR` 표기 이벤트 수 | `internal/hook/coverage_table.go` line 37, line 39 | "26-event" / "26-event inventory" | "29-event" / "29-event inventory" |
| 카운트 단언 테스트 | `internal/hook/types_test.go` line 12-13 | `len(events) != 26` … `want 26` | `!= 29` … `want 29` |

> [HARD] 카운트 정밀성 주의 — off-by-one 함정: `ValidEventTypes()`(=EventType 상수)는 **26→29**로 깔끔하게 증가하나, `CoverageTable`의 `len()`은 현재 **27**(합성 COMPOSITE 행 "AutoUpdate (SessionStart composite)" 포함)이므로 3개 추가 시 **30 행**이 된다. 즉 `len(CoverageTable) == 29`로 단언하면 **틀린다**. `CoverageTable`의 "29"는 **행 수가 아니라 이벤트 의미 수**(헤더 주석/ANCHOR 텍스트)에 해당하고, `len()`은 30이다. 구현 시 이 구분을 반드시 지킬 것.

> [HARD] 카운트 cascade 정밀성 — `RetireObsOnly` summary는 **4→7**: 3개 신규 이벤트가 모두 `ResolutionRetireObsOnly`이므로 기존 4개(Notification/TaskCreated/Elicitation/ElicitationResult)에 더해 7이 된다. `internal/cli/doctor_hook_test.go:125`(`summary.RetireObsOnly`)와 `:108`(observability filter `len`) 두 곳이 4→7로 변경되어야 한다. non-composite eventCount(`:35`)는 26→29. 세 cross-package 단언이 함께 변하므로 `./internal/cli/...`를 반드시 같이 테스트할 것.

---

## 2. 범위 (Scope)

### 2.1 In Scope

대상 파일 (production 2 + test 5 = **7 files**; 단 test 5 중 `cohabitation_guard_test.go`는 verified-no-edit이므로 **실제 edit 받는 test 파일은 4개**: types_test.go / audit_test.go / doctor_hook_test.go / hook_e2e_test.go). 7번째 file `hook_e2e_test.go`는 orchestrator exhaustive sweep에서 발견(D6).

**Production code (`internal/hook/`)**:
- `internal/hook/types.go`: `EventPostToolBatch`, `EventUserPromptExpansion`, `EventMessageDisplay` 3개 `EventType` 상수 추가 (observe-only — 각 상수에 도큐먼트 주석 + 가능 버전 표기)
- `internal/hook/types.go`: `ValidEventTypes()` 반환 슬라이스에 3개 멤버 추가 (26 → 29)
- `internal/hook/coverage_table.go`: `CoverageTable`에 3개 `EventCoverageEntry` 추가 (Resolution = `ResolutionRetireObsOnly`, `IsActive: false`, `HandlerFile: ""` — D3 빈 문자열 확정)
- `internal/hook/coverage_table.go`: line 37 헤더 주석 + line 39 `@MX:ANCHOR` 텍스트를 26 → 29로 갱신

**Test code (count assertion + 3-way sync allowlist + cross-package)**:
- `internal/hook/types_test.go`(또는 신규 테스트): 26 카운트 단언을 29로 갱신 + 3개 신규 이벤트의 `ValidEventTypes()` 멤버십 단언 추가 + `len(CoverageTable) == 30` 단언
- **[D1 BLOCKING] `internal/hook/audit_test.go:284`**: `deregisteredButLiveEventNames` 슬라이스에 `"PostToolBatch"`, `"UserPromptExpansion"`, `"MessageDisplay"` 3개 추가. 이것 없이는 `TestAuditThreeWaySync`(audit_test.go:296)가 결정론적으로 `HOOK_SYNC_DRIFT` 발생(:303-308 루프가 Resolution∉{COMPOSITE,REMOVE}을 Go-registered로 분류 → 3개가 settings.json·retiredEventNames·deregisteredButLiveEventNames 어디에도 없어 :356 fail). 참고: production-export `RetiredEventNames`(retired_events.go:17)가 **아니라** test-local `deregisteredButLiveEventNames`에 추가(후자는 `internal/migrate`가 소비하지 않음 — observer-style 미등록이지만 Go 핸들러 live인 이벤트의 3-way sync 면제 leg). 단 3개는 핸들러조차 없으므로 의미상으로는 "deregistered + no-handler-yet"이나, 3-way sync 통과를 위한 가장 작은 변경이 본 슬라이스 append이다.
- **[D5 cross-package] `internal/cli/doctor_hook_test.go`**: 3개 단언 갱신 — `:35` non-composite eventCount `26 → 29`, `:125` `summary.RetireObsOnly` `4 → 7`, `:108` observability filter `len(entries)` `4 → 7`. CoverageTable이 `internal/cli`에 `hook.CoverageTable`/`hook.Summarize()`로 결합되어 있어 `./internal/hook/...`만 테스트하면 이 회귀를 놓친다(auditor 실증).
- **[D6 cross-package, 7th file] `internal/cli/hook_e2e_test.go`**: 2개 갱신 — (a) `:202` `if got := len(validTypes); got != 26`(에러 메시지 `:203` "want 26") → `!= 29` / "want 29" (`ValidEventTypes()` count). (b) `:286-340` `TestHookValidEventTypes_AllHaveSubcommands`의 `:291` `excludedEvents := map[hook.EventType]bool{}`(현재 empty)에 3개 observe-only 이벤트를 `true`로 추가. 이유: 이 테스트는 `ValidEventTypes()`를 순회하며 `excludedEvents`에 없고 `eventToSubcmd`(:294-321)에도 없는 이벤트에 대해 `:334` `t.Errorf("no expected subcommand mapping for event %q")`를 발생. 3개 신규 이벤트는 observe-only(subcommand 없음, Exclusion "CLI/cobra subcommand 추가 금지")이므로 `eventToSubcmd`가 **아니라** `excludedEvents`에 추가해야 함(Exclusion 정합). `hook_e2e_test.go`는 orchestrator exhaustive sweep로 발견된 7번째 coupled file이며 AC-HER-006a의 `go test ./internal/cli/...`를 결정론적으로 fail시킨다.

### 2.2 Exclusions (What NOT to Build)

- [HARD] **`hooks-system.md` 등 문서 레이어 재편집 금지**: 문서는 이미 3개 이벤트를 정확히 기술하고 있다(검증 완료). doc 결함은 SPEC-CC-DOCS-ALIGNMENT-001 소관이다.
- **blocking/decision 동작 추가 금지**: 3개 이벤트는 observe-only로만 등록한다. `decision: "block"`, exit-code-2, 핸들러 비즈니스 로직(`Handle()` 구현체), `settings.json` 등록(`IsActive: true`)을 추가하지 않는다.
- **무관한 기존 이벤트 변경 금지**: 기존 26개 `EventType` 상수, 핸들러, production-export `RetiredEventNames`(retired_events.go:17 — `internal/migrate` 소비), 합성 COMPOSITE 행을 수정하지 않는다. (단 test-local `deregisteredButLiveEventNames` audit_test.go:284 append은 D1 BLOCKING remediation으로 In Scope — production 코드 아님.)
- **`internal/template/templates/` 미러링 불필요**: 대상 Go 파일은 `internal/` 하위 도구 코드로 template-managed가 아니다(`internal/template/templates/internal/hook/` 부재 실측 확인). template-first 규칙 미적용.
- **`HookInput` 필드 추가 금지**: 이벤트별 페이로드 필드(`batch_id`, `tool_results`, `expansion_type`, `displayContent` 등) 구조체 확장은 본 SPEC 범위 밖이다(핸들러가 없으므로 소비처 없음). 후속 SPEC 권장 사항으로만 남긴다.
- **CLI/cobra subcommand 추가 금지**: `moai hook <event>` 디스패치 핸들러 등록은 범위 밖(observe-only는 상수 + 커버리지 인벤토리 등재까지만).

---

## 3. 요구사항 (Requirements — GEARS) + 수용 기준 (Acceptance Criteria INLINE)

> 표기: GEARS notation. `<subject>` = "the hook event registry" / "the coverage table" / "the test suite" 등 일반화 주어. 모든 AC는 grep/go-test로 검증 가능.

### REQ-HER-001 — 3개 observe-only EventType 상수 추가

**The hook event registry shall** `internal/hook/types.go`에 `EventPostToolBatch`, `EventUserPromptExpansion`, `EventMessageDisplay` 3개 `EventType` 상수를 각각 공식 이벤트 문자열 값(`"PostToolBatch"`, `"UserPromptExpansion"`, `"MessageDisplay"`)으로 정의해야 한다.

- **AC-HER-001a**: `grep -c 'EventPostToolBatch\|EventUserPromptExpansion\|EventMessageDisplay' internal/hook/types.go` 의 결과가 상수 정의 + `ValidEventTypes()` 멤버 참조를 합산하여 `>= 6` (각 식별자 최소 2회: 정의 1 + 슬라이스 멤버 1). 단, 상수 정의부만 한정한 카운트 `grep -cE '^\tEvent(PostToolBatch|UserPromptExpansion|MessageDisplay) EventType = "' internal/hook/types.go == 3`.
- **AC-HER-001b**: 각 상수 값이 공식 문자열과 정확히 일치 — `grep -q 'EventPostToolBatch EventType = "PostToolBatch"' internal/hook/types.go` && 동일 패턴 `UserPromptExpansion`, `MessageDisplay` 각각 exit 0.

### REQ-HER-002 — ValidEventTypes() 멤버십 확장 (26 → 29)

**When** `ValidEventTypes()` 가 호출되면, **the hook event registry shall** 기존 26개에 더해 3개 신규 이벤트를 포함한 총 29개 `EventType`을 반환해야 하며, `IsValidEventType()`은 3개 신규 이벤트에 대해 `true`를 반환해야 한다.

- **AC-HER-002a**: `len(ValidEventTypes()) == 29` (테스트로 단언; `internal/hook/types_test.go`의 `want 26` → `want 29` 갱신).
- **AC-HER-002b**: `IsValidEventType(EventPostToolBatch)`, `IsValidEventType(EventUserPromptExpansion)`, `IsValidEventType(EventMessageDisplay)` 모두 `true` (테스트 케이스로 단언).
- **AC-HER-002c**: `grep -cE '^\t\tEvent(PostToolBatch|UserPromptExpansion|MessageDisplay),' internal/hook/types.go == 3` (슬라이스 멤버 3개 등재).

### REQ-HER-003 — CoverageTable observe-only 엔트리 추가 (27 → 30 행)

**The coverage table shall** `internal/hook/coverage_table.go`의 `CoverageTable`에 3개 신규 이벤트를 각각 observe-only 성격(`IsActive: false`, Resolution = `ResolutionRetireObsOnly`, `HandlerFile: ""`)의 `EventCoverageEntry`로 추가해야 한다.

- **AC-HER-003a**: `grep -cE '\{EventName: "(PostToolBatch|UserPromptExpansion|MessageDisplay)"' internal/hook/coverage_table.go == 3`.
- **AC-HER-003b**: `len(CoverageTable) == 30` (테스트로 단언 — 현재 27 + 신규 3; **29 아님**, off-by-one 주의).
- **AC-HER-003c**: 3개 신규 엔트리 모두 `IsActive: false` (observe-only — `settings.json` 미등록). 신규 행 각각에 대해 `grep`으로 `IsActive: false` 동반 확인.
- **AC-HER-003d** (D3): 3개 신규 엔트리의 `HandlerFile` 필드가 빈 문자열 `""` (observe-only — 존재하지 않는 핸들러 파일을 암시하지 않음). `HandlerFile`에 os.Stat을 수행하는 테스트는 없으므로 `""`는 안전하다 — 어떤 핸들러 파일 경로(예: `post_tool_batch.go`)도 작성하지 말 것. 검증: `grep -E '\{EventName: "(PostToolBatch|UserPromptExpansion|MessageDisplay)".*HandlerFile: ""' internal/hook/coverage_table.go` 가 3행 매치(필드 순서/포맷에 따라 정규식 조정 가능).

### REQ-HER-004 — 헤더 주석 + @MX:ANCHOR 이벤트 수 갱신 (26 → 29)

**The coverage table shall** `coverage_table.go`의 line 37 헤더 주석("the authoritative 26-event table")과 line 39 `@MX:ANCHOR` 텍스트("26-event inventory")의 이벤트 수를 29로 갱신해야 한다.

- **AC-HER-004a**: `grep -c '26-event' internal/hook/coverage_table.go == 0` (모든 "26-event" 표기 제거).
- **AC-HER-004b** (D4): `grep -c '29-event' internal/hook/coverage_table.go == 2` (정확히 2곳: line 37 헤더 주석 + line 39 `@MX:ANCHOR`. `>= 2`가 아니라 `== 2` 정밀 단언 — 신규 "29-event" 표기를 그 2곳 외에 추가하지 않음).
- **AC-HER-004c**: `@MX:ANCHOR` 행이 보존되며 29로 갱신 — `grep -q '@MX:ANCHOR.*29-event' internal/hook/coverage_table.go` exit 0 (ANCHOR는 절대 삭제 금지, mx-tag-protocol.md 준수).

### REQ-HER-005 — RED-우선 카운트 테스트 (TDD)

**While** `development_mode: tdd`인 동안, **the test suite shall** 구현(GREEN) 이전에 3개 신규 이벤트의 레지스트리 인지를 단언하는 테스트가 먼저 실패(RED)하도록 작성/확장되어야 한다.

- **AC-HER-005a**: `go test ./internal/hook/ -run 'TestValidEventTypes' -count=1` 이 GREEN 이후 통과 (RED→GREEN 전이 — 구현 전에는 `want 29` 단언이 실패해야 함).
- **AC-HER-005b**: 신규/확장 테스트가 3개 이벤트의 `ValidEventTypes()` 멤버십 + `len()==29` + `len(CoverageTable)==30`을 모두 단언.

### REQ-HER-006 — observe-only 무해성 + 회귀 0 (REFACTOR/검증)

**The hook event registry shall not** 3개 신규 이벤트에 대해 blocking 동작(decision/block, exit-code-2)이나 핸들러 비즈니스 로직을 도입해서는 안 되며, 기존 26개 이벤트의 동작을 변경해서는 안 된다.

- **AC-HER-006a** (D5+D6 canonical full-package): `go test ./internal/hook/... ./internal/cli/... -count=1` 전체 통과 (두 패키지 모두 — `internal/cli`의 `doctor_hook_test.go`(D5)와 `hook_e2e_test.go`(D6)가 `hook.CoverageTable`/`hook.Summarize()`/`hook.ValidEventTypes()`에 결합되어 있어 `./internal/hook/...` 단독 검증은 cross-package 회귀를 놓침. `-run TestDoctorHook`처럼 부분 limit하면 `hook_e2e_test.go`의 `TestHookValidEventTypes_*` 회귀를 놓치므로 **반드시 패키지 전체** 실행. 기존 테스트 포함, 회귀 0).
- **AC-HER-006b**: `go vet ./internal/hook/...` clean + `golangci-lint run ./internal/hook/...` clean (또는 프로젝트 기준 lint 0).
- **AC-HER-006c**: subagent boundary 가드 보존 — `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go` == 0 matches (신규 코드가 가드 위반 미도입).
- **AC-HER-006d**: 3개 신규 이벤트에 대해 `Handle()` 구현 핸들러 파일이나 `settings.json` 등록이 추가되지 않음 (observe-only 불변식) — production-export `RetiredEventNames`(retired_events.go:17) 슬라이스 및 합성 COMPOSITE 행 불변 확인.

### REQ-HER-007 — 3-way sync invariant 유지 (D1 BLOCKING)

**When** `CoverageTable`에 3개 observe-only 행이 추가되면, **the test suite shall** `TestAuditThreeWaySync`(audit_test.go:296)가 계속 통과하도록 `deregisteredButLiveEventNames`(audit_test.go:284)에 3개 이벤트 이름을 추가해야 한다. 이 추가가 없으면 audit_test.go:303-308 루프가 3개를 Go-registered(Resolution∉{COMPOSITE,REMOVE})로 분류하나 settings.json·retiredEventNames·deregisteredButLiveEventNames 어디에도 없어 :356에서 `HOOK_SYNC_DRIFT`가 **결정론적으로** 발생한다(auditor가 패치+실행으로 실증).

- **AC-HER-007a**: `go test ./internal/hook/... -run TestAuditThreeWaySync -count=1` 통과 (3개 이벤트에 대한 `HOOK_SYNC_DRIFT` 0건).
- **AC-HER-007b**: `grep -cE '"(PostToolBatch|UserPromptExpansion|MessageDisplay)"' internal/hook/audit_test.go` 의 결과가 `deregisteredButLiveEventNames` 슬라이스 내 3개 이름 등재를 포함하여 `>= 3` (단, 슬라이스 한정 카운트: `deregisteredButLiveEventNames` 블록 내 3개 매치 확인).
- **AC-HER-007c**: production-export `RetiredEventNames`(retired_events.go:17)는 **변경 없음** — `grep -c 'PostToolBatch\|UserPromptExpansion\|MessageDisplay' internal/hook/retired_events.go == 0` (3개는 test-local allowlist에만 추가, `internal/migrate` 소비 슬라이스 불변).

### REQ-HER-008 — cross-package doctor_hook 테스트 정합 (D5)

**The test suite shall** `internal/cli/doctor_hook_test.go`의 `CoverageTable`-결합 카운트 단언을 신규 이벤트 수에 맞게 갱신해야 한다 (3개가 `ResolutionRetireObsOnly`이므로 non-composite event count와 RetireObsOnly summary count가 동시에 증가).

- **AC-HER-008a**: `internal/cli/doctor_hook_test.go:35`의 non-composite eventCount 단언이 `!= 29`로 갱신 — `grep -c 'want 26\|!= 26' internal/cli/doctor_hook_test.go == 0` (26 잔재 제거) 및 29 단언 존재 확인.
- **AC-HER-008b**: `internal/cli/doctor_hook_test.go:125`의 `summary.RetireObsOnly` 단언이 `!= 7`로 갱신(기존 4 + 신규 3) — `want 4` 잔재 제거.
- **AC-HER-008c**: `internal/cli/doctor_hook_test.go:108`의 observability filter `len(entries)` 단언이 `!= 7`로 갱신(RETIRE-OBS-ONLY 이벤트 4 → 7) — `want 4` 잔재 제거.
- **AC-HER-008d**: `go test ./internal/cli/... -run 'TestDoctorHook' -count=1` 통과 (3개 doctor_hook 테스트 모두 GREEN).

### REQ-HER-009 — cross-package hook_e2e 테스트 정합 (D6 — 7th coupled file)

**The test suite shall** `internal/cli/hook_e2e_test.go`의 `ValidEventTypes()`-결합 카운트 단언과 subcommand-매핑 테스트를 신규 observe-only 이벤트에 맞게 갱신해야 한다.

- **AC-HER-009a**: `internal/cli/hook_e2e_test.go:202`의 `len(validTypes)` 단언이 `!= 29`로 갱신(에러 메시지 :203 "want 26" → "want 29") — `grep -c 'want 26\|!= 26' internal/cli/hook_e2e_test.go == 0`.
- **AC-HER-009b**: `internal/cli/hook_e2e_test.go:291`의 `excludedEvents` 맵에 `EventPostToolBatch`, `EventUserPromptExpansion`, `EventMessageDisplay` 3개가 `true`로 등재 (observe-only ⇒ subcommand 없음). `eventToSubcmd`(:294-321)에는 **추가 금지**(Exclusion 정합) — `grep -cE 'Event(PostToolBatch|UserPromptExpansion|MessageDisplay):\s+true' internal/cli/hook_e2e_test.go == 3` (또는 excludedEvents 블록 한정 카운트 3).
- **AC-HER-009c**: `TestHookValidEventTypes_AllHaveSubcommands`가 `:334` "no expected subcommand mapping for event" 에러 없이 통과 — `go test ./internal/cli/... -run 'TestHookValidEventTypes' -count=1` GREEN (count + subcommand-mapping 테스트 모두).

---

## 4. 비기능 제약 (Non-Functional Constraints)

- **언어 정책**: 식별자/경로/코드/REQ-AC 토큰은 영어, 산문 주석/문서는 `code_comments` 설정(ko) 준수.
- **@MX 규약**: `coverage_table.go`의 `@MX:ANCHOR`는 갱신만 하고 삭제 금지(mx-tag-protocol.md — ANCHOR는 NEVER auto-delete).
- **테스트 격리**: 신규 테스트는 `t.Parallel()` 가능 시 사용, 임시 디렉터리 필요 시 `t.TempDir()`.
- **template 중립성 무관**: 대상 파일이 template 밖이므로 16-언어 중립성/template-first 미적용.

---

## 5. 참조 (References)

- `.claude/rules/moai/core/hooks-system.md` (line 24-25, 83, 109-110) — 3개 이벤트 문서 선행(이미 정확, 재편집 금지)
- `internal/hook/types.go` — `EventType` 상수 블록(lines 20-111) + `ValidEventTypes()`(lines 115-148)
- `internal/hook/coverage_table.go` — `CoverageTable`(lines 41-72), 헤더 주석(line 37), `@MX:ANCHOR`(line 39)
- `internal/hook/types_test.go` — `TestValidEventTypes`(line 8-50, `want 26` 단언 line 12-13)
- `internal/hook/audit_test.go` — `deregisteredButLiveEventNames`(line 284, D1 append 대상) + `TestAuditThreeWaySync`(line 296, drift 검출) + Go-registered 분류 루프(line 303-308) + `HOOK_SYNC_DRIFT` emit(line 356)
- `internal/cli/doctor_hook_test.go` — `TestDoctorHook_27EventTableCount`(line 19, eventCount `!= 26` line 35) + `TestDoctorHook_ObservabilityFilter`(line 100, `len != 4` line 108) + `TestDoctorHook_SummaryCountsConsistent`(line 115, `RetireObsOnly != 4` line 125; `Remove want 0` line 132 = verified-no-edit) — D5 cross-package 결합
- `internal/cli/hook_e2e_test.go` — `len(validTypes) != 26`(line 202, 에러 메시지 line 203) + `TestHookValidEventTypes_AllHaveSubcommands`(line 286-340: `excludedEvents` line 291, `eventToSubcmd` line 294-321, "no expected subcommand mapping" 에러 line 334) — D6 7th cross-package 결합. verified-no-edit: subset membership(line 212-217) / dup(line 219-226) / determinism(line 265-282)
- `internal/hook/retired_events.go` (line 17 `RetiredEventNames`, `internal/migrate` 소비 — 불변) + `internal/hook/notification.go` / `task_created.go` — observe-only(RETIRE-OBS-ONLY) 기존 패턴 모범
- `.claude/rules/moai/workflow/mx-tag-protocol.md` — `@MX:ANCHOR` 보존 규약
- 분리 모태: SPEC-CC-DOCS-ALIGNMENT-001 (문서 33 결함, 본 SPEC의 자매 — 문서 측면 소관)
