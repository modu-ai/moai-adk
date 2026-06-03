# SPEC-HOOK-EVENT-REGISTRY-001 — 구현 계획 (plan.md)

> Tier S · development_mode = tdd · module: internal/hook
> 단일 cohesive Go-코드 결함: doc-ahead-of-code drift 해소 (3개 공식 observe-only 이벤트 등록)

---

## §A. Context (배경 / 검증된 현재 상태)

- **결함 본질**: 공식 Claude Code 현행 이벤트 3종(`PostToolBatch` v2.1.89+, `UserPromptExpansion` v2.1.90+, `MessageDisplay` v2.1.152+)이 MoAI hook 레지스트리(Go)에 누락. 문서(`hooks-system.md`)는 이미 정확히 기술 → Go 코드가 문서보다 뒤처짐.
- **검증된 카운트 (Read/grep 실측)**:
  - `EventType` 상수: **26** (types.go lines 20-111) → 목표 **29**
  - `ValidEventTypes()` 멤버: **26** (lines 115-142) → 목표 **29**
  - `CoverageTable` 행: **27** (실제 이벤트 26 + 합성 "AutoUpdate" COMPOSITE 1) → 목표 **30**
  - 헤더 주석/ANCHOR 표기: "26-event" (line 37, 39) → "29-event"
  - 카운트 테스트: `types_test.go:12-13` `want 26` → `want 29`
  - **[D1] 3-way sync allowlist**: `audit_test.go:284` `deregisteredButLiveEventNames` (현재 `{WorktreeCreate, WorktreeRemove}`) → 3개 추가 필요
  - **[D5] cross-package doctor_hook**: `internal/cli/doctor_hook_test.go` non-composite eventCount `:35` 26→29 / `summary.RetireObsOnly` `:125` 4→7 / observability filter `len` `:108` 4→7
  - **[D6] cross-package hook_e2e (7th coupled file)**: `internal/cli/hook_e2e_test.go` `len(validTypes)` `:202`(메시지 :203) 26→29 / `excludedEvents` `:291`(현재 empty map)에 3개 observe-only 추가 — `eventToSubcmd`(:294-321) 아님(Exclusion 정합)
- **observe-only 패턴 모범**: `retired_events.go`(`RetiredEventNames`, line 17 — `internal/migrate` 소비, 불변), `notification.go`/`task_created.go`(Resolution=`ResolutionRetireObsOnly`, `IsActive: false`).
- **RetireObsOnly cascade**: 신규 3개 모두 `ResolutionRetireObsOnly`이므로 `Summarize().RetireObsOnly`가 4(Notification/TaskCreated/Elicitation/ElicitationResult)→7로 증가. 이것이 D5의 `:125`/`:108` 변경 원인.

## §B. Known Issues / 주의점 (off-by-one 함정)

- [HARD] **`len(CoverageTable) == 30`, NOT 29**. `ValidEventTypes()`는 26→29(깔끔)이지만 `CoverageTable`은 현재 27행(COMPOSITE 합성 행 포함)이라 +3 = 30행. 헤더 주석/ANCHOR의 "29-event"는 **이벤트 의미 수**이고 `len()`은 30이다. 두 숫자를 혼동하면 AC-HER-003b 실패.
- `@MX:ANCHOR`(line 39)는 **갱신만**, 삭제 금지(mx-tag-protocol.md).
- 3개 이벤트는 **observe-only** — `Handle()` 구현·`settings.json` 등록·blocking 동작 추가 금지.
- `cohabitation_guard_test.go`가 `coverage_table.go`의 `ResolutionRetireObsOnly`/`ObservabilityOptIn` 필드 존재를 가드 → 신규 엔트리가 이 필드 스키마를 깨지 않도록 주의.
- **[D1 BLOCKING — drift는 가능성이 아니라 결정론적 확정]** `audit_test.go::TestAuditThreeWaySync`(line 296)는 `CoverageTable`(line 303-308 루프)을 Go-registered 권위 소스로 사용한다. Go-registered 판정 = `Resolution ∉ {COMPOSITE, REMOVE}` → `ResolutionRetireObsOnly`인 신규 3개는 **Go-registered로 분류된다**. 그리고 skip 집합은 `retiredEventNames`(retired_events.go:17 = {Notification, Elicitation, ElicitationResult, TaskCreated})와 test-local `deregisteredButLiveEventNames`(audit_test.go:284 = {WorktreeCreate, WorktreeRemove}) 둘뿐이다. 신규 3개는 settings.json에도, 두 skip 집합 어디에도 없으므로 line 356 `HOOK_SYNC_DRIFT`가 **3건 결정론적으로 발생**한다(auditor가 패치+테스트 실행으로 실증). 이전 §B 초안의 "settings.json에 없어도 정합(observability whitelist 경로)" 주장은 **거짓**이다 — observability whitelist(`ObservabilityOptIn`)는 3-way sync 판정에 관여하지 않으며, 본 plan은 어떤 allowlist에도 이름을 추가하지 않은 상태였다.
- **[D1 확정 remediation]** `deregisteredButLiveEventNames`(audit_test.go:284)에 `"PostToolBatch"`, `"UserPromptExpansion"`, `"MessageDisplay"` 3개를 추가한다(M2 deliverable). production-export `RetiredEventNames`(retired_events.go:17)가 **아니라** test-local 슬라이스에 추가 — 후자는 `internal/migrate.CleanupUserSettings`가 소비하지 않으므로 user settings.json 정리 동작에 영향 없음. (RetiredEventNames는 migrate가 소비하므로 거기에 추가하면 의미가 달라진다.)
- 3개 이벤트는 **observe-only** — `Handle()` 구현·`settings.json` 등록·blocking 동작 추가 금지. `HandlerFile`은 빈 문자열 `""`로 둔다(D3 — 존재하지 않는 핸들러 파일명을 적으면 안 됨; 어떤 테스트도 HandlerFile에 os.Stat 안 함을 auditor가 확인).

## §C. Pre-flight (착수 전 확인)

```bash
# 현재 카운트 baseline 재확인
grep -cE '^\tEvent[A-Za-z]+ EventType = "' internal/hook/types.go          # 기대 26
grep -cE '^\t\tEvent[A-Za-z]+,' internal/hook/types.go                      # 기대 26
grep -cE '\{EventName:' internal/hook/coverage_table.go                     # 기대 27
grep -n 'want 26\|26-event' internal/hook/types_test.go internal/hook/coverage_table.go
# [D1] 3-way sync allowlist baseline
grep -n 'deregisteredButLiveEventNames' internal/hook/audit_test.go         # line 284 슬라이스 def 확인
# [D5] cross-package doctor_hook baseline
grep -n '!= 26\|RetireObsOnly != 4\|len(entries) != 4' internal/cli/doctor_hook_test.go  # :35 / :125 / :108
# [D6] cross-package hook_e2e baseline (7th coupled file)
grep -n 'len(validTypes); got != 26\|excludedEvents := map' internal/cli/hook_e2e_test.go  # :202 / :291
go test ./internal/hook/... ./internal/cli/... -count=1                      # baseline GREEN (두 패키지)
```

## §D. Constraints

- Go 1.23+, 테이블 주도 테스트, `t.Parallel()` 권장.
- 식별자/코드 영어, 주석 ko(`code_comments`).
- template-first 미적용(internal/ 도구 코드).

## §E. Self-Verification (완료 게이트 — read-only 병렬 배치)

```bash
# 1. 카운트 단언 (구현 후)
grep -cE '^\tEvent(PostToolBatch|UserPromptExpansion|MessageDisplay) EventType = "' internal/hook/types.go   # == 3  (AC-HER-001a)
grep -cE '^\t\tEvent(PostToolBatch|UserPromptExpansion|MessageDisplay),' internal/hook/types.go              # == 3  (AC-HER-002c)
grep -cE '\{EventName: "(PostToolBatch|UserPromptExpansion|MessageDisplay)"' internal/hook/coverage_table.go # == 3  (AC-HER-003a)
# 2. 헤더/ANCHOR 갱신 + HandlerFile (D3) 빈 문자열
grep -c '26-event' internal/hook/coverage_table.go    # == 0  (AC-HER-004a)
grep -c '29-event' internal/hook/coverage_table.go    # == 2  (AC-HER-004b — D4 정밀: 정확히 2곳)
grep -q '@MX:ANCHOR.*29-event' internal/hook/coverage_table.go && echo OK   # (AC-HER-004c)
grep -E '\{EventName: "(PostToolBatch|UserPromptExpansion|MessageDisplay)".*HandlerFile: ""' internal/hook/coverage_table.go | wc -l  # == 3  (AC-HER-003d, D3)
# 3. [D1] 3-way sync allowlist + retiredEventNames 불변
grep -c 'PostToolBatch\|UserPromptExpansion\|MessageDisplay' internal/hook/retired_events.go   # == 0  (AC-HER-007c — production export 불변)
# 4. [D5] cross-package doctor_hook 26/4 잔재 제거
grep -c 'want 26\|!= 26' internal/cli/doctor_hook_test.go    # == 0  (AC-HER-008a)
grep -c 'RetireObsOnly != 4\|want 4' internal/cli/doctor_hook_test.go   # == 0  (AC-HER-008b/c)
# 5. [D6] cross-package hook_e2e 26 잔재 제거 + excludedEvents 3개 등재
grep -c 'want 26\|!= 26' internal/cli/hook_e2e_test.go    # == 0  (AC-HER-009a)
grep -cE 'Event(PostToolBatch|UserPromptExpansion|MessageDisplay):\s+true' internal/cli/hook_e2e_test.go  # == 3  (AC-HER-009b — excludedEvents 등재; eventToSubcmd 아님)
# 6. 테스트 (두 패키지 전체) + lint + 가드
go test ./internal/hook/... ./internal/cli/... -count=1                 # GREEN (AC-HER-006a/007a/008d/009c/005a) — D5+D6: cli 패키지 전체 필수(-run limit 금지)
go test ./internal/hook/... -run TestAuditThreeWaySync -count=1         # GREEN (AC-HER-007a — drift 0)
go vet ./internal/hook/... ./internal/cli/...                           # clean (AC-HER-006b)
golangci-lint run ./internal/hook/... ./internal/cli/... || true        # clean (AC-HER-006b)
grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v _test.go | wc -l   # == 0 (AC-HER-006c)
```

`len()` 단언(AC-HER-002a `==29`, AC-HER-003b `len(CoverageTable)==30`)은 Go 테스트 내부에서 검증(`TestValidEventTypes` 확장 + 신규 `TestCoverageTableLen`).

## §F. Milestones

### M1 — RED: 카운트/멤버십 테스트 작성·확장 (실패 우선)

- **파일**: `internal/hook/types_test.go` (확장)
- **작업**:
  - `TestValidEventTypes`의 `len(events) != 26` / `want 26` → `!= 29` / `want 29`로 갱신
  - `expected` 맵에 `EventPostToolBatch`, `EventUserPromptExpansion`, `EventMessageDisplay: true` 3개 추가
  - `TestIsValidEventType` 케이스에 3개 신규 이벤트 `true` 추가
  - 신규 테스트 `TestCoverageTableLen`: `len(CoverageTable) == 30` 단언 (또는 기존 테스트 확장)
- **RED 확인**: `go test ./internal/hook/ -run 'TestValidEventTypes|TestIsValidEventType' -count=1` → FAIL (상수 미정의로 컴파일 에러 또는 카운트 불일치)
- **참고**: D1(audit_test) / D5(doctor_hook)는 기존 테스트의 **단언 갱신**이므로 M2에서 production 코드와 함께 처리한다(별도 RED 작성 불필요 — 기존 테스트가 production 변경으로 자동 RED가 되고 단언 갱신으로 GREEN). 단 `TestAuditThreeWaySync`는 production만 바꾸고 audit_test를 안 바꾸면 결정론적 FAIL(=의도된 RED 신호)이 되며, M2 파일 3 append로 GREEN 전이한다.
- **AC**: AC-HER-005a, AC-HER-005b, AC-HER-002a(단언 셋업), AC-HER-003b(단언 셋업)

### M2 — GREEN: 상수 + 슬라이스 + CoverageTable + 주석/ANCHOR

- **파일 1 `internal/hook/types.go`**:
  - `EventType` const 블록 끝(line 110 `EventPermissionDenied` 다음)에 3개 상수 추가 (observe-only 도큐먼트 주석 + 가능 버전 표기 ko):
    - `EventPostToolBatch EventType = "PostToolBatch"` (v2.1.89+, observe-only)
    - `EventUserPromptExpansion EventType = "UserPromptExpansion"` (v2.1.90+, observe-only)
    - `EventMessageDisplay EventType = "MessageDisplay"` (v2.1.152+, observe-only)
  - `ValidEventTypes()` 슬라이스(line 141 `EventPermissionDenied,` 다음)에 3개 멤버 추가
  - **AC**: AC-HER-001a, AC-HER-001b, AC-HER-002b, AC-HER-002c
- **파일 2 `internal/hook/coverage_table.go`**:
  - `CoverageTable`에 3개 `EventCoverageEntry` 추가 (`ElicitationResult` 행(line 67) 다음, COMPOSITE 행(line 71) 앞 권장) — **[D3] `HandlerFile: ""` 빈 문자열 확정**:
    - `{EventName: "PostToolBatch", Resolution: ResolutionRetireObsOnly, IsActive: false, HandlerFile: ""}`
    - `{EventName: "UserPromptExpansion", Resolution: ResolutionRetireObsOnly, IsActive: false, HandlerFile: ""}`
    - `{EventName: "MessageDisplay", Resolution: ResolutionRetireObsOnly, IsActive: false, HandlerFile: ""}`
    - [D3 확정] `HandlerFile`은 빈 문자열 `""` — 존재하지 않는 핸들러 파일명(`post_tool_batch.go` 등)을 적지 말 것. auditor가 어떤 테스트도 `HandlerFile` 값에 `os.Stat`을 수행하지 않음을 확인했으므로 `""`가 안전하고 observe-only 의미("핸들러 미존재")에 정합한다. placeholder hedge 제거됨.
  - line 37 헤더 주석 "26-event" → "29-event", line 39 `@MX:ANCHOR` "26-event inventory" → "29-event inventory" (정확히 이 2곳만 — D4)
  - **AC**: AC-HER-003a, AC-HER-003c, AC-HER-003d, AC-HER-004a, AC-HER-004b, AC-HER-004c
- **파일 3 [D1 BLOCKING] `internal/hook/audit_test.go`**:
  - line 284 `deregisteredButLiveEventNames` 슬라이스에 3개 추가:
    ```go
    var deregisteredButLiveEventNames = []string{
        "WorktreeCreate",
        "WorktreeRemove",
        "PostToolBatch",        // observe-only (no handler yet), 3-way sync 면제
        "UserPromptExpansion",  // observe-only (no handler yet)
        "MessageDisplay",       // observe-only (no handler yet)
    }
    ```
  - 슬라이스 도큐먼트 주석을 신규 3개에 맞게 보강(선택) — 단 의미: 기존 WorktreeCreate/Remove는 "Go 핸들러 live, settings 미등록"이고 신규 3개는 "핸들러 미존재 observe-only"이나 둘 다 3-way sync 면제 leg로 동일 처리.
  - **AC**: AC-HER-007a, AC-HER-007b, AC-HER-007c(retired_events.go 불변 — 본 파일 외)
- **파일 4 [D5] `internal/cli/doctor_hook_test.go`**:
  - line 35 `if eventCount != 26` → `!= 29` (+ 메시지 `want 26` → `want 29`)
  - line 125 `if summary.RetireObsOnly != 4` → `!= 7` (+ 메시지 `want 4` → `want 7`)
  - line 108 `if len(entries) != 4` → `!= 7` (+ 메시지 `want 4` → `want 7`) — observability filter가 RETIRE-OBS-ONLY 이벤트 7개 반환
  - 주석 정합(선택): `TestDoctorHook_27EventTableCount` 헤더 주석의 "26 events + 1 composite = 27" → "29 events + 1 composite = 30" 갱신 권장(테스트 이름 자체는 변경 불필요 — `len(entries) != len(hook.CoverageTable)` 동적 비교라 무관).
  - **verified-no-edit**: `:23` `len(entries) != len(hook.CoverageTable)`(relative 자기조정) / `:132` `Remove want 0`(RETIRE-OBS-ONLY≠REMOVE, 불변) → 건드리지 말 것.
  - **AC**: AC-HER-008a, AC-HER-008b, AC-HER-008c, AC-HER-008d
- **파일 5 [D6, 7th coupled file] `internal/cli/hook_e2e_test.go`**:
  - line 202 `if got := len(validTypes); got != 26` → `!= 29` (+ 메시지 line 203 `want 26` → `want 29`)
  - line 291 `excludedEvents := map[hook.EventType]bool{}` → 3개 observe-only 이벤트를 `true`로 추가:
    ```go
    excludedEvents := map[hook.EventType]bool{
        hook.EventPostToolBatch:       true, // observe-only, no subcommand
        hook.EventUserPromptExpansion: true,
        hook.EventMessageDisplay:      true,
    }
    ```
  - [D6 핵심] `eventToSubcmd`(:294-321)에는 **추가 금지** — 3개는 subcommand가 없는 observe-only(Exclusion "CLI/cobra subcommand 추가 금지"). `excludedEvents`에 넣어야 `TestHookValidEventTypes_AllHaveSubcommands`의 `:328-340` 순회가 3개를 skip하여 `:334` "no expected subcommand mapping" 에러를 회피한다.
  - **verified-no-edit**: subset membership(`:212-217` "Verify all 7 new events" — 신규 7개와 무관, 통과) / dup(`:219-226`) / determinism(`:265-282`) → 건드리지 말 것.
  - **AC**: AC-HER-009a, AC-HER-009b, AC-HER-009c
- **GREEN 확인**: `go test ./internal/hook/... ./internal/cli/... -count=1` → PASS (두 패키지 전체 — D5+D6; `-run` limit 금지)

### M3 — REFACTOR + 전체 검증 (회귀 0)

- **작업**: hook + cli 두 패키지 전체 테스트 + vet + lint + 가드 grep. **[D2] `TestAuditThreeWaySync` drift는 가능성이 아니라 결정론적 확정**이었으며, M2의 `deregisteredButLiveEventNames` append(파일 3)가 그 확정 remediation이다. M3는 "오탐 여부 확인"이 아니라 "M2 remediation이 drift를 0으로 만들었는지 **검증**"한다 — `go test ./internal/hook/... -run TestAuditThreeWaySync` GREEN 확인.
- **검증**: §E Self-Verification 전체 배치 (hook + cli 두 패키지 **전체** 명시 — `-run` limit 금지)
- **AC**: AC-HER-006a(cli 패키지 전체), AC-HER-006b, AC-HER-006c, AC-HER-006d, AC-HER-007a, AC-HER-008d, AC-HER-009c, AC-HER-005a(최종 GREEN)

## §G. Anti-Patterns (회피)

- ❌ `len(CoverageTable)`를 29로 단언 → 실제 30 (off-by-one). COMPOSITE 합성 행 망각 금지.
- ❌ **[D1] `CoverageTable`만 바꾸고 `deregisteredButLiveEventNames`(audit_test.go:284)를 안 바꿈** → `TestAuditThreeWaySync` 결정론적 `HOOK_SYNC_DRIFT` 3건. drift는 "가능성"이 아니라 확정.
- ❌ **[D1] `RetiredEventNames`(retired_events.go:17, production export)에 3개 추가** → `internal/migrate.CleanupUserSettings` 동작 의미 변경. test-local `deregisteredButLiveEventNames`에만 추가할 것.
- ❌ **[D5] `./internal/hook/...`만 테스트** → `internal/cli/doctor_hook_test.go`의 `:35`/`:125`/`:108` cross-package 회귀를 놓침. `./internal/cli/...` 반드시 함께.
- ❌ **[D6] `./internal/cli/...`를 `-run TestDoctorHook`로 제한** → `hook_e2e_test.go`의 `TestHookValidEventTypes_*`(count :202 + subcommand-mapping :286) 회귀 누락. 패키지 전체(`go test ./internal/hook/... ./internal/cli/... -count=1`) 실행 필수.
- ❌ **[D6] 3개 신규 이벤트를 `eventToSubcmd`(:294-321)에 추가** → observe-only인데 subcommand가 등록되지 않아 `:337` "expects subcommand … not registered" 에러. `excludedEvents`(:291)에 넣어야 함(Exclusion 정합).
- ❌ **[D3] `HandlerFile`에 `post_tool_batch.go` 등 존재하지 않는 파일명** → 빈 문자열 `""`로 둘 것(observe-only, 핸들러 미존재 의미).
- ❌ `hooks-system.md` 재편집 → 문서는 이미 정확, doc SPEC 소관.
- ❌ 3개 이벤트에 `Handle()` 핸들러/`settings.json` 등록/blocking 추가 → observe-only 위반.
- ❌ `@MX:ANCHOR` 삭제 → 갱신만 허용.
- ❌ 기존 26개 이벤트·production `RetiredEventNames`·COMPOSITE 행 변경.
- ❌ GREEN 직행 (RED 생략) → TDD 위반, M1 먼저.

## §G.1 Verified-Safe — 확인했으나 무변경 (Exhaustive Sweep 문서화)

orchestrator의 exhaustive grep sweep로 count/coverage-결합 가능성을 전수 점검한 결과, 다음 site들은 **변경 불필요**(자기조정 또는 불변)임을 확인했다. 향후 audit가 재-flag하지 않도록 명시 기록한다. **이것이 완전한 coupled surface이며 8번째 site는 없다.**

| Site | 이유 (무변경) |
|------|--------------|
| `internal/hook/cohabitation_guard_test.go:112` | 문자열 `ResolutionRetireObsOnly`의 **존재**만 단언(string-presence) — count-coupled 아님 |
| `internal/cli/doctor_hook_test.go:23` | `len(entries) != len(hook.CoverageTable)` — **relative** 비교(테이블 길이에 자기조정) |
| `internal/cli/doctor_hook_test.go:132` | `Remove count … want 0` — 신규 이벤트는 RETIRE-OBS-ONLY이지 REMOVE 아니므로 `Remove`는 계속 0 |
| `internal/cli/hook_e2e_test.go:212-217` | "Verify all 7 new events" — **subset membership** 체크(신규 7개와 무관, 통과) |
| `internal/cli/hook_e2e_test.go:219-226` | 중복 없음 단언(dup) — 통과 |
| `internal/cli/hook_e2e_test.go:265-278` | `TestHookValidEventTypes_Deterministic` 결정성 — 통과 |

## §H. Cross-References

- spec.md §3 (REQ-HER-001~009 + 인라인 AC)
- `internal/hook/types.go`, `coverage_table.go`, `types_test.go`, `audit_test.go`(:284/:296/:303-308/:356), `retired_events.go`(:17)
- `internal/cli/doctor_hook_test.go`(:19/:35/:108/:125 — D5 cross-package; :23/:132 verified-no-edit)
- `internal/cli/hook_e2e_test.go`(:202/:203/:291 — D6 7th cross-package; :212-217/:219-226/:265-282 verified-no-edit)
- `internal/hook/cohabitation_guard_test.go`(:112 verified-no-edit, string-presence)
- `.claude/rules/moai/core/hooks-system.md` (문서 선행, 재편집 금지)
- `.claude/rules/moai/workflow/mx-tag-protocol.md` (ANCHOR 보존)
- 자매 SPEC: SPEC-CC-DOCS-ALIGNMENT-001 (문서 33 결함)
