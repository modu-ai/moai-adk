# SPEC-V3R2-RT-006 Implementation Plan

> Implementation plan for **Hook Handler Completeness and 27-Event Coverage**.
> Companion to `spec.md` v0.1.0, `research.md` v0.1.0, `acceptance.md` v0.1.0, `tasks.md` v0.1.0.
> Authored on branch `plan/SPEC-V3R2-RT-006` (Step 1 plan-in-main; base `origin/main` HEAD `c810b11b7`).
> Run phase will execute on a fresh worktree `feat/SPEC-V3R2-RT-006` per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline Step 2.

## HISTORY

| Version | Date       | Author                          | Description                                                                                                                                                                                                                                                                                                                                                                                                                                            |
|---------|------------|---------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 1B) | Initial implementation plan per `.claude/skills/moai/workflows/plan.md`. Scope: convert spec.md §5 EARS REQs into milestone breakdown, given the **partial-pre-completion** discovered in research.md §2 (5 critical handlers + 4 RETIRE headers + setup.go body removal already merged into `c810b11b7`). Run scope reduced to retire-mechanism (settings.json/template), system.yaml schema (hook.observability_events), audit_test.go v2, doctor hook CLI, per-file category headers, and AC-level test verification. |

---

## 1. Plan Overview

### 1.1 Goal restatement

`spec.md` §1 의 핵심 목표를 milestone 분해:

> Decide each of the 27 Claude Code hook events with an explicit "full logic" or "retire from settings.json" verdict; fix P-H02 tmux pane leak; remove orphan setupHandler; reconcile settings.json registration count with Go handler count.

`research.md` §2 의 inventory 결과 다음 상태를 발견:

- **5 critical handlers** (subagentStop FIX, configChange UPGRADE, instructionsLoaded UPGRADE, fileChanged UPGRADE, postToolUseFailure UPGRADE) 의 **business logic 은 이미 main `c810b11b7` 에 머지**됨 (PR 미상; pre-existing work).
- **4 RETIRE-OBS-ONLY 헤더** (notification.go, elicitation.go, task_created.go) 는 추가됨, 그러나 settings.json + template 등록은 미제거.
- **`setup.go` 본문 (0 bytes)**: 비워졌지만 file 자체는 잔존.
- **`audit_test.go` (152 LOC)**: handler count + retired-not-active 만 검증; settings.json parity / per-file category 헤더 grep / observability whitelist 미구현.
- **`internal/cli/doctor_hook.go`**: 미존재.
- **`system.yaml hook.observability_events`**: schema 정의 없음 + RT-005 SystemHookConfig struct 미존재.

이로 인한 **Plan-phase delta** (research.md §2.1 from 73% baseline → 100% spec coverage):

1. `internal/template/templates/.claude/settings.json.tmpl` + `.claude/settings.json` 에서 4 retire event 제거 (Notification, Elicitation, ElicitationResult, TaskCreated).
2. `system.yaml` 에 `hook:` 섹션 + `observability_events: []` + `strict_mode: false` 추가.
3. `internal/config/types.go` 에 `SystemHookConfig` struct 추가 + `SystemConfig.Hook` field. RT-005 audit_test 가 yaml↔struct parity 강제하므로 동시 PR 필수.
4. 4 retire 핸들러 본문에 `observabilityOptIn(ctx)` runtime gate 추가 (Option A per research.md §3.3).
5. `audit_test.go` v2: 4 신규 sub-test (TestAuditRegistrationParity, TestAuditPerFileCategoryHeader, TestAuditRetiredEventsNotInSettings, TestAuditObservabilityWhitelist).
6. 22개 핸들러 파일에 `// Resolution: <CATEGORY>` 헤더 일괄 추가 (4 RETIRE 는 보유).
7. `setup.go` 파일 자체 제거 (`git rm internal/hook/setup.go`).
8. `internal/cli/doctor_hook.go` 신규 (~150 LOC) + `internal/cli/doctor_hook_test.go`.
9. AC-level integration test 보강: 5 critical handler 의 spec §6 acceptance criteria 기반 path test (AC-01 ~ AC-08).
10. spec.md §5.7 final-count footer 의 KEEP=15 vs 실제 17 reconciliation note (research.md §3.5/§3.6).

핵심 deltas (research.md §2/§3/§6/§7/§8 cross-reference):

- **Settings.json template branching**: `{{- if .ObservabilityEvents.Notification }}` Go template 분기로 opt-in 시 다시 등록.
- **Runtime observability gate**: `notification.go` / `elicitation.go` / `task_created.go` 의 `Handle(ctx, ...)` 가 `cfg.System.Hook.ObservabilityEvents` slice 를 읽고 미옵트인 시 `slog.Info` 까지 차단 (or 옵트인 시에만 호출).
- **Per-file Resolution header**: 22 handler file 일괄 sed-style edit (Plan-phase 에서는 file list 만 enumeration; Run-phase 에서 일괄 적용).
- **doctor hook CLI**: cobra subcommand `moai doctor hook [--trace <event>] [--observability] [--json]`.
- **audit_test.go v2**: settings.json + system.yaml + deps.go 3-way parity.

### 1.2 Implementation Methodology

`.moai/config/sections/quality.yaml`: `development_mode: tdd`. Run phase MUST follow RED → GREEN → REFACTOR per `.claude/rules/moai/workflow/spec-workflow.md` § Run Phase.

- **RED**: 기존 `audit_test.go` 의 152 LOC 위에 4개 새 sub-test 추가 + 5 critical handler 의 AC path test 추가 + `doctor_hook_test.go` 신규 생성. 모든 신규 test 가 처음에 FAIL (system.yaml 미정의, settings.json 미수정, doctor_hook 미구현 상태).
- **GREEN**: §1.1 의 10 deltas 구현. 기존 5 critical handler 의 본문은 GREEN 상태 유지; 추가 작업은 헤더 + observability gate + audit + doctor + retire-mechanism.
- **REFACTOR**: per-file category 헤더 일괄 적용 후 lint check; observability gate 헬퍼를 `internal/hook/internal_helpers.go` 등 공용 위치로 이동; doctor_hook 의 27-event table builder 를 internal/hook 패키지로 export.

### 1.2.1 Acknowledged Discrepancies

본 plan 이 spec.md 와 의도적으로 다르게 처리하는 부분 (research.md §3.5/§3.6 에서 발견):

- **§5.7 footer count "15 KEEP"**: 실제 row count = 17 KEEP (KEEP rows: 1, 2, 3, 4, 6, 7, 8, 9, 10, 13, 14, 15, 16, 17, 19, 20, 22, 33 - 17개; UPGRADE rows: 5, 21, 23, 24 = 4; FIX row: 11 = 1; RETIRE rows: 12, 18, 25, 26 = 4; REMOVE row: 27 = 1; total 17+4+1+4+1=27. → spec footer 의 "15 KEEP + 5 UPGRADE" 는 오타로 보이며 (실제는 17 KEEP + 4 UPGRADE; PostToolUseFailure 가 row 5 UPGRADE 이면 5 UPGRADE 도 가능). 

  **결정** (이 plan 의 binding count): **17 KEEP + 4 UPGRADE (5는 오타) + 1 FIX + 4 RETIRE-OBS-ONLY + 1 REMOVE = 27**. 단, spec.md row 5 (PostToolUseFailure UPGRADE), row 21 (ConfigChange UPGRADE), row 23 (FileChanged UPGRADE), row 24 (InstructionsLoaded UPGRADE) = 4. SubagentStop (row 11) 은 FIX. 따라서 **17 KEEP + 4 UPGRADE + 1 FIX + 4 RETIRE + 1 REMOVE = 27**. ✓
  
  Audit_test 와 doctor hook 의 footer summary 는 이 정정 카운트 사용. spec.md 의 footer 는 sync-phase HISTORY entry 에서 정정.

- **`HookResponse` vs `HookOutput`**: spec.md EARS 절은 `HookResponse` 라고 명명하지만 실제 코드는 `HookOutput` 사용. 본 plan 에서 두 이름은 동일 type 의 alias 로 취급; per-file category 헤더 grep 은 두 이름 모두 검출 가능.

### 1.3 Deliverables

| Deliverable | Path | REQ Coverage |
|-------------|------|--------------|
| Remove 4 retire events from settings.json template | `internal/template/templates/.claude/settings.json.tmpl` (delete 4 keys + Go template branching) | REQ-V3R2-RT-006-004 |
| Remove 4 retire events from local settings.json | `.claude/settings.json` (delete 4 keys) | REQ-V3R2-RT-006-004, AC-V3R2-RT-006-10 |
| Add `hook.observability_events` to system.yaml | `.moai/config/sections/system.yaml` (new section ~5 LOC) + `internal/template/templates/.moai/config/sections/system.yaml` (template parity) | REQ-V3R2-RT-006-004, AC-V3R2-RT-006-11 |
| Add SystemHookConfig struct | `internal/config/types.go` (new struct ~12 LOC) | REQ-V3R2-RT-006-004, RT-005 audit parity |
| Runtime observability gate for 4 retire handlers | `internal/hook/notification.go` (+15 LOC), `internal/hook/elicitation.go` (+30 LOC for 2 handlers), `internal/hook/task_created.go` (+15 LOC) | REQ-V3R2-RT-006-040, AC-V3R2-RT-006-11, -16 |
| audit_test.go v2 with 4 new sub-tests | `internal/hook/audit_test.go` (extend from 152 → ~280 LOC) | REQ-V3R2-RT-006-003, -042, -063, AC-V3R2-RT-006-10, -13, -14 |
| Per-file Resolution category headers (22 handlers) | 22 files in `internal/hook/*.go` (+1 line each) | REQ-V3R2-RT-006-002, AC-V3R2-RT-006-14 |
| Remove orphan setup.go | `internal/hook/setup.go` (`git rm`) | REQ-V3R2-RT-006-005, AC-V3R2-RT-006-09 |
| New `moai doctor hook` CLI | `internal/cli/doctor_hook.go` (new ~150 LOC) + `internal/cli/doctor_hook_test.go` (new ~80 LOC) | REQ-V3R2-RT-006-050, -051, AC-V3R2-RT-006-12 |
| AC-level integration tests for 5 critical handlers | extend existing `subagent_stop_test.go`, `config_change_test.go`, `instructions_loaded_test.go`, `file_changed_test.go`, `post_tool_failure_test.go` | AC-V3R2-RT-006-01, -02, -03, -04, -05, -06, -07, -08, -15 |
| ConfigChange RT-005 reload integration | `internal/hook/config_change.go` (+10 LOC: import RT-005 Manager, call `mgr.Reload(input.ConfigFilePath)` after validation) | REQ-V3R2-RT-006-011, AC-V3R2-RT-006-04 |
| ConfigChange 20 ms debounce | `internal/hook/config_change.go` (+20 LOC: time.Sleep gate or fsnotify integration) | spec §8 risk row 4 |
| SubagentStop 500 ms per-pane timeout | `internal/hook/subagent_stop.go` (+10 LOC: context.WithTimeout wrap of killTmuxPane goroutine) | spec §8 risk row 1 |
| TagScanner integration stub for FileChanged | `internal/hook/file_changed.go` (verify SPC-002 interface usage; if SPC-002 unmerged, stub interface + integration test mock) | REQ-V3R2-RT-006-013 |
| 27-event coverage table renderer | `internal/hook/coverage_table.go` (new ~80 LOC: shared by audit_test + doctor_hook) | REQ-V3R2-RT-006-050 |
| CHANGELOG entry | `CHANGELOG.md` Unreleased section | Trackable (TRUST 5) |
| MX tags per §6 | 6 files (per §6 below) | mx_plan |

Embedded-template parity is **applicable** because settings.json.tmpl + system.yaml.tmpl change. `make build` regeneration required.

### 1.4 Traceability Matrix (REQ → AC → Task)

Per plan-auditor PASS criterion #2 (every REQ maps to at least one AC and at least one Task). Built **after** tasks.md was finalized; each row references actual T-RT006-NN IDs.

| REQ ID | Category | Mapped AC(s) | Mapped Task(s) |
|--------|----------|--------------|----------------|
| REQ-V3R2-RT-006-001 | Ubiquitous (every event has Go handler emitting HookResponse/HookOutput) | (baseline; verified by 27-event table coverage) | T-RT006-19 (audit handler count), T-RT006-25 (doctor hook table) |
| REQ-V3R2-RT-006-002 | Ubiquitous (per-file Resolution header) | AC-14 | T-RT006-12, T-RT006-13 (header insertion + audit grep test) |
| REQ-V3R2-RT-006-003 | Ubiquitous (audit_test.go registration parity) | AC-13 | T-RT006-19 (RegistrationParity test) |
| REQ-V3R2-RT-006-004 | Ubiquitous (4 retire events removed + observability tap) | AC-10, AC-11 | T-RT006-04, T-RT006-05, T-RT006-06, T-RT006-07, T-RT006-21 |
| REQ-V3R2-RT-006-005 | Ubiquitous (setupHandler removed) | AC-09 | T-RT006-09 (git rm setup.go) |
| REQ-V3R2-RT-006-006 | Ubiquitous (subagentStop reads tmuxPaneId + kill-pane + update registry) | AC-01, AC-15 | (baseline ✅; verification only) T-RT006-14 |
| REQ-V3R2-RT-006-007 | Ubiquitous (Windows tmux no-op) | AC-03 | (baseline ✅) T-RT006-14 |
| REQ-V3R2-RT-006-010 | Event-Driven (SubagentStop a-e 5-step) | AC-01, AC-15 | (baseline ✅) T-RT006-14 |
| REQ-V3R2-RT-006-011 | Event-Driven (ConfigChange RT-005 reload) | AC-04, AC-05 | T-RT006-15 (RT-005 Manager.Reload call), T-RT006-16 (debounce + invalid-config test) |
| REQ-V3R2-RT-006-012 | Event-Driven (InstructionsLoaded 40k char check) | AC-06 | (baseline ✅; AC path test) T-RT006-17 |
| REQ-V3R2-RT-006-013 | Event-Driven (FileChanged MX rescan for 16+ extensions) | AC-07 | T-RT006-18 (TagScanner stub integration + AC test) |
| REQ-V3R2-RT-006-014 | Event-Driven (PostToolUseFailure 7-class) | AC-08 | (baseline ✅; AC path test) T-RT006-20 |
| REQ-V3R2-RT-006-020..033 | Event-Driven KEEP semantic (14 REQs) | (semantic reaffirmation; no implementation delta) | T-RT006-25 (doctor hook prints status) |
| REQ-V3R2-RT-006-040 | State-Driven (observability_events live → structured logs only) | AC-11, AC-16 | T-RT006-04 (system.yaml schema), T-RT006-05 (SystemHookConfig struct), T-RT006-06 (runtime gate), T-RT006-07 (gate tests) |
| REQ-V3R2-RT-006-041 | State-Driven (strict_mode + retired event = silent succeed) | (negative test) | T-RT006-22 |
| REQ-V3R2-RT-006-042 | State-Driven (audit_test fail when undocumented handler added) | AC-13 | T-RT006-19 |
| REQ-V3R2-RT-006-050 | Optional (doctor hook prints 27-event table) | AC-12 | T-RT006-23, T-RT006-24, T-RT006-25 |
| REQ-V3R2-RT-006-051 | Optional (doctor hook --trace) | AC-12 | T-RT006-26 |
| REQ-V3R2-RT-006-060 | Unwanted (subagentStop tmuxPaneId not found → DEBUG + return) | AC-02 | (baseline ✅) T-RT006-14 |
| REQ-V3R2-RT-006-061 | Unwanted (kill-pane "pane not found" → success) | AC-02 | (baseline ✅) T-RT006-14 |
| REQ-V3R2-RT-006-062 | Unwanted (configChange reload validation fail → keep old + reject) | AC-05 | T-RT006-16 |
| REQ-V3R2-RT-006-063 | Unwanted (audit_test fail on retired event in settings.json) | AC-10 | T-RT006-19 |

Coverage: **22 unique REQs (001..063, with 020..033 grouped as 14 KEEP semantic) → 16 ACs (AC-01 through AC-17 minus AC-15 partial overlap) → 26 tasks (T-RT006-01..26)**.

→ Spec-driven row "AC-15" maps to REQ-010 (verifies pane teardown ordering). All AC IDs from spec §6 are covered. All REQ IDs (REQ-001 through REQ-063) are mapped except REQ-020..033 which are KEEP-semantic (no implementation delta; verified via doctor hook + existing tests).

---

## 2. Milestone Breakdown (M1-M5)

각 milestone 은 **priority-ordered** (no time estimates per `.claude/rules/moai/core/agent-common-protocol.md` § Time Estimation HARD rule).

### M1: Test scaffolding + setup.go removal (RED phase) — Priority P0

Reference existing tests: `internal/hook/{audit,subagent_stop,config_change,instructions_loaded,file_changed,post_tool_failure}_test.go`.

Owner role: `expert-backend` (Go test) or direct `manager-cycle` execution.

Tasks:
- T-RT006-01: Delete `internal/hook/setup.go` (0-byte orphan). Verify no import / no NewSetupHandler reference exists in `deps.go`. Run `go build ./...` to confirm.
- T-RT006-02: Add 4 RED sub-tests in `internal/hook/audit_test.go`:
  - TestAuditRegistrationParity (RED — currently 27 handlers vs 25 settings.json events ≠ 22 + 1 + 4)
  - TestAuditPerFileCategoryHeader (RED — only 4 RETIRE handlers have header)
  - TestAuditRetiredEventsNotInSettings (RED — settings.json still contains 4 retired)
  - TestAuditObservabilityWhitelist (RED — system.yaml has no `hook:` section)
- T-RT006-03: Add RED AC-path tests:
  - subagent_stop_test.go: TestSubagentStop_PaneNotFoundGraceful (verifies AC-02 via mocked exec.Command)
  - config_change_test.go: TestConfigChange_RT005ReloadIntegration (verifies AC-04 via stub Manager.Reload)
  - config_change_test.go: TestConfigChange_InvalidYAMLKeepsOldSettings (verifies AC-05/REQ-062)
  - instructions_loaded_test.go: TestInstructionsLoaded_42kCLAUDE (verifies AC-06 with synthetic 42k file)
  - file_changed_test.go: TestFileChanged_MXTagDelta (verifies AC-07 with mocked TagScanner)
  - post_tool_failure_test.go: TestPostToolFailure_TimeoutClassification (verifies AC-08 with exit code 124 input)
- T-RT006-04: Add `hook.observability_events: []` schema to local `.moai/config/sections/system.yaml` + template `internal/template/templates/.moai/config/sections/system.yaml`. RED: RT-005 typed loader will fail until step T-RT006-05.

Verification gate: All new tests FAIL on `go test ./internal/hook/ ./internal/config/ -run "TestAudit|TestSubagentStop_Pane|TestConfigChange_RT005|TestInstructionsLoaded_42k|TestFileChanged_MX|TestPostToolFailure_Timeout"`. Existing tests still PASS.

### M2: SystemHookConfig + RT-005 typed loader integration (GREEN seed) — Priority P0

Owner role: `expert-backend`.

Tasks:
- T-RT006-05: Add `SystemHookConfig` struct to `internal/config/types.go`:
  ```go
  type SystemHookConfig struct {
      ObservabilityEvents []string `yaml:"observability_events" validate:"dive,oneof=notification elicitation elicitationResult taskCreated"`
      StrictMode          bool     `yaml:"strict_mode"`
  }
  type SystemConfig struct {
      // existing fields...
      Hook SystemHookConfig `yaml:"hook"`
  }
  ```
  Add validator/v10 tags. Run RT-005 audit_test (TestAuditParity) to confirm yaml↔struct parity.
- T-RT006-06: Implement `observabilityOptIn(ctx context.Context, eventName string) bool` helper in `internal/hook/observability.go` (new file ~30 LOC). Reads `config.FromContext(ctx).System.Hook.ObservabilityEvents`. Returns false on nil config or missing event.
- T-RT006-07: Apply observability gate to 4 retire handlers (notification.go, elicitation.go ElicitationHandler + ElicitationResultHandler, task_created.go). Pattern:
  ```go
  func (h *notificationHandler) Handle(ctx context.Context, input *HookInput) (*HookOutput, error) {
      if !observabilityOptIn(ctx, "notification") {
          return &HookOutput{}, nil  // silent
      }
      slog.Info("notification received", ...)
      return &HookOutput{}, nil  // never SystemMessage / Continue:false
  }
  ```
- T-RT006-08: Verify M1 RED tests now show:
  - TestAuditObservabilityWhitelist → GREEN (observability_events present + valid).
  - TestAuditPerFileCategoryHeader → still RED (headers not yet added).

Verification gate: `go test ./internal/config/ -run TestAuditParity` PASS. `go test ./internal/hook/ -run TestAuditObservabilityWhitelist` PASS.

### M3: Settings.json retire + Resolution headers + handler upgrades (GREEN main) — Priority P0

Owner role: `expert-backend`.

Tasks:
- T-RT006-09: Remove 4 retire entries from `internal/template/templates/.claude/settings.json.tmpl`. Add Go template `{{- if .ObservabilityEvents.<EventName> }}...{{- end }}` branching for opt-in.
- T-RT006-10: Remove 4 retire entries from local `.claude/settings.json`. Verify `moai update` (dry-run) does not recreate them when system.yaml has empty observability_events.
- T-RT006-11: Add template context field `ObservabilityEvents map[string]bool` in `internal/template/context.go`. Render path reads from system.yaml at template-deploy time.
- T-RT006-12: Apply `// Resolution: <CATEGORY>` header to 22 handler files (lookup table per spec §5.7):
  - KEEP (17 files): session_start.go, session_end.go, pretool_*.go (if separate), posttool_*.go, pre_compact.go, post_compact.go, stop.go, stop_failure.go, agent_start.go, prompt_submit.go, permission_request.go, permission_denied.go, teammate_idle.go, task_completed.go, worktree_create.go, worktree_remove.go, cwd_changed.go.
  - UPGRADE (4 files): post_tool_failure.go, config_change.go, file_changed.go, instructions_loaded.go.
  - FIX (1 file): subagent_stop.go.
  - COMPOSITE (1 file): auto_update.go.
  - (already done) RETIRE-OBS-ONLY: notification.go, elicitation.go, task_created.go.
- T-RT006-13: Verify TestAuditPerFileCategoryHeader now GREEN (regex `^// Resolution: (KEEP|UPGRADE|FIX|REMOVE|RETIRE-OBS-ONLY|COMPOSITE)$` matches every non-test, non-aux handler file).
- T-RT006-14: Add subagent_stop.go 500 ms per-pane timeout wrap (spec §8 risk row 1):
  ```go
  ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
  defer cancel()
  errCh := make(chan error, 1)
  go func() { errCh <- h.killTmuxPane(tmuxPaneID) }()
  select {
  case err := <-errCh: /* handle */
  case <-ctx.Done(): slog.Warn("kill-pane timeout"); /* graceful */
  }
  ```
  Verify TestSubagentStop_PaneNotFoundGraceful + new TestSubagentStop_KillPaneTimeout PASS.
- T-RT006-15: Wire RT-005 Manager.Reload into config_change.go. After successful `validateConfig()`, call:
  ```go
  if mgr := config.ManagerFromContext(ctx); mgr != nil {
      if err := mgr.Reload(input.ConfigFilePath); err != nil {
          return &HookOutput{Continue: false, SystemMessage: fmt.Sprintf("Reload rejected: %v; old settings retained", err)}, nil
      }
  }
  ```
  Verify TestConfigChange_RT005ReloadIntegration GREEN.
- T-RT006-16: Add 20 ms debounce to config_change.go (spec §8 risk row 4):
  ```go
  time.Sleep(20 * time.Millisecond)  // debounce mid-write
  data, err := os.ReadFile(input.ConfigFilePath)
  ```
  Verify TestConfigChange_InvalidYAMLKeepsOldSettings GREEN.
- T-RT006-17: Verify TestInstructionsLoaded_42kCLAUDE GREEN (existing 89-LOC handler already covers; AC-06 path test).
- T-RT006-18: Wire SPC-002 TagScanner stub into file_changed.go. If SPC-002 not yet merged at run-time, use interface placeholder + mock in test. Verify TestFileChanged_MXTagDelta GREEN.
- T-RT006-19: Implement TestAuditRegistrationParity in audit_test.go:
  ```go
  // Parse settings.json native event list (excluding 4 retired)
  // Parse deps.go HookRegistry.Register count (excluding setup, including autoUpdate composite)
  // Parse system.yaml observability_events
  // Assert: |deps.go| == |settings.json native| + 1 (autoUpdate) + |observability_events|
  ```
  Initial expected: 26 handlers (after setup.go removal) == 21 native + 1 composite + 4 retire-residual-handlers = 26. ✓
- T-RT006-20: Verify TestPostToolFailure_TimeoutClassification GREEN (existing 134-LOC handler covers; AC-08 path test with stderr "timeout").

Verification gate: All M1 RED tests now GREEN. `go test ./internal/hook/` PASS. `go vet ./...` clean.

### M4: doctor hook CLI + observability list reflection (REFACTOR + new feature) — Priority P1

Owner role: `expert-backend`.

Tasks:
- T-RT006-21: Implement TestAuditRetiredEventsNotInSettings — verify 4 retire events absent from `.claude/settings.json`. After M3-T10 complete, this PASSes.
- T-RT006-22: Implement TestStrictModeRetiredEvent (REQ-V3R2-RT-006-041 negative): With `strict_mode: true` + retire event firing (mocked), assert handler returns `HookOutput{}` silently.
- T-RT006-23: Create `internal/hook/coverage_table.go` (~80 LOC). Build the 27-event table data structure used by both audit_test and doctor_hook. Each entry: `{EventName, Resolution, IsActive, ObservabilityOptIn}`.
- T-RT006-24: Create `internal/cli/doctor_hook.go` (~150 LOC):
  ```go
  var doctorHookCmd = &cobra.Command{
      Use: "hook",
      Short: "27-event hook coverage diagnostic",
      RunE: func(cmd *cobra.Command, args []string) error {
          table := hook.BuildCoverageTable(deps.HookRegistry, cfg)
          if jsonFlag { return json.NewEncoder(os.Stdout).Encode(table) }
          return printCoverageTable(os.Stdout, table)
      },
  }
  doctorHookCmd.Flags().Bool("json", false, "machine-readable output")
  doctorHookCmd.Flags().String("trace", "", "stream last invocation decision path for <event>")
  doctorHookCmd.Flags().Bool("observability", false, "show observability_events status")
  ```
- T-RT006-25: Wire doctor_hook into `internal/cli/doctor.go` parent. Add to `moai doctor` help.
- T-RT006-26: Implement `--trace <event>` reading from `.moai/logs/hook.log` last N lines for the named event. (Out-of-scope simplification: tail-only readout, no real-time stream.)
- T-RT006-27: Add `internal/cli/doctor_hook_test.go` (~80 LOC):
  - Test 27-event table count
  - Test default observability_events = empty
  - Test --json flag output is parseable JSON
  - Test --trace with non-existent event returns no-op message

Verification gate: `moai doctor hook` (manual) prints 27-row table. `go test ./internal/cli/ -run TestDoctorHook` PASS.

### M5: Verification gates + audit consolidation — Priority P0

Owner role: `manager-cycle` + `manager-quality`.

Tasks:
- T-RT006-28: Run full test suite: `go test ./... -race -count=1`. Ensure 0 regressions.
- T-RT006-29: Run linter: `golangci-lint run`. Fix any issues introduced by header insertion or new code.
- T-RT006-30: Verify `make build` succeeds and `internal/template/embedded.go` is regenerated correctly with the modified settings.json.tmpl.
- T-RT006-31: Manual integration test: spawn live tmux team mode, verify SubagentStop pane cleanup works end-to-end (records lessons #12, #13 patterns).
- T-RT006-32: Update `CHANGELOG.md` Unreleased section with bullet entries:
  - "fix(hook/SPEC-V3R2-RT-006): tmux pane leak on SubagentStop (P-H02)"
  - "feat(hook/SPEC-V3R2-RT-006): retire 4 hook events from settings.json with observability opt-in"
  - "feat(cli/SPEC-V3R2-RT-006): add `moai doctor hook` 27-event diagnostic"
- T-RT006-33: Add @MX tags per §6 below.
- T-RT006-34: Re-run `go test ./internal/hook/ -run TestAudit` (all 4 sub-tests). Confirm GREEN.
- T-RT006-35: Verify spec.md §5.7 footer count via `moai doctor hook --json | jq '.summary'` → expect 17 KEEP / 4 UPGRADE / 1 FIX / 4 RETIRE / 1 REMOVE.

Verification gate: All AC-V3R2-RT-006-01 through AC-V3R2-RT-006-17 verified per acceptance.md.

---

## 3. File-Level Modification Map

### 3.1 Files modified (existing)

| File | Lines added/changed | Purpose |
|------|---------------------|---------|
| `internal/hook/notification.go` | +15 LOC | observability gate |
| `internal/hook/elicitation.go` | +30 LOC | observability gate (2 handlers) |
| `internal/hook/task_created.go` | +15 LOC | observability gate |
| `internal/hook/subagent_stop.go` | +10 LOC | 500 ms per-pane timeout + Resolution header |
| `internal/hook/config_change.go` | +30 LOC | RT-005 reload + 20 ms debounce + Resolution header |
| `internal/hook/instructions_loaded.go` | +1 LOC | Resolution header |
| `internal/hook/file_changed.go` | +5 LOC | TagScanner stub + Resolution header |
| `internal/hook/post_tool_failure.go` | +1 LOC | Resolution header |
| 17 KEEP handler files | +1 LOC each | Resolution: KEEP header |
| `internal/hook/auto_update.go` | +1 LOC | Resolution: COMPOSITE header |
| `internal/hook/audit_test.go` | +130 LOC | 4 new sub-tests |
| `internal/hook/subagent_stop_test.go` | +40 LOC | AC-02 path + timeout test |
| `internal/hook/config_change_test.go` | +50 LOC | AC-04 + AC-05 path tests |
| `internal/hook/instructions_loaded_test.go` | +30 LOC | AC-06 path test |
| `internal/hook/file_changed_test.go` | +35 LOC | AC-07 path test |
| `internal/hook/post_tool_failure_test.go` | +25 LOC | AC-08 path test |
| `internal/config/types.go` | +12 LOC | SystemHookConfig + SystemConfig.Hook |
| `internal/template/templates/.claude/settings.json.tmpl` | -4 keys + Go template branches | Retire 4 events + opt-in |
| `.claude/settings.json` | -4 keys | Local mirror |
| `.moai/config/sections/system.yaml` | +5 LOC | hook section |
| `internal/template/templates/.moai/config/sections/system.yaml` | +5 LOC | template parity |
| `internal/template/context.go` | +8 LOC | ObservabilityEvents field |
| `CHANGELOG.md` | +3 lines | Unreleased entry |

### 3.2 Files created (new)

| File | Lines | Purpose |
|------|-------|---------|
| `internal/hook/observability.go` | ~30 | observabilityOptIn helper |
| `internal/hook/coverage_table.go` | ~80 | 27-event table data structure |
| `internal/cli/doctor_hook.go` | ~150 | new CLI subcommand |
| `internal/cli/doctor_hook_test.go` | ~80 | doctor_hook tests |

### 3.3 Files removed

| File | Reason |
|------|--------|
| `internal/hook/setup.go` | Orphan handler (REQ-005, AC-09) |

### 3.4 Files NOT modified (out-of-scope)

- `internal/hook/{post_tool_*.go}` — KEEP semantic (POSTToolUse keeps existing logic)
- `internal/hook/mx/**` — TagScanner integration is SPC-002 owned
- `internal/hook/dbsync/**` — orthogonal to RT-006

---

## 4. Technical Approach

### 4.1 Settings.json template branching

Go template syntax:
```jsonc
{
  "hooks": {
    "SessionStart": [...],
    {{- if .ObservabilityEvents.notification }}
    "Notification": [{
      "hooks": [{"command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-notification.sh\""}]
    }],
    {{- end }}
    {{- if .ObservabilityEvents.elicitation }}
    "Elicitation": [{
      "hooks": [{"command": "\"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/handle-elicitation.sh\""}]
    }],
    {{- end }}
    {{- if .ObservabilityEvents.elicitationResult }}
    "ElicitationResult": [...],
    {{- end }}
    {{- if .ObservabilityEvents.taskCreated }}
    "TaskCreated": [...],
    {{- end }}
    // ... rest of hooks
  }
}
```

Template context (`internal/template/context.go`):
```go
type TemplateContext struct {
    GoBinPath           string
    HomeDir             string
    ObservabilityEvents map[string]bool  // NEW
    // ...
}

func NewTemplateContext(opts ...Option) *TemplateContext {
    // Read system.yaml via RT-005 SettingsResolver if available; fallback to empty map
    obs := loadObservabilityEvents()
    return &TemplateContext{ObservabilityEvents: obs, ...}
}
```

### 4.2 SystemHookConfig validation

`validator/v10` tag enforces:
- `observability_events` slice items MUST be one of `notification`, `elicitation`, `elicitationResult`, `taskCreated`. Other values raise validation error during RT-005 Load.
- `strict_mode` is bool, no constraint.

### 4.3 Concurrency safety

- 4 retire handlers' observability gate is read-only against `config.FromContext(ctx)`. RT-005 resolver is concurrent-safe via sync.RWMutex.
- subagent_stop.go's team config write is a single-writer per teammate; no lock added (consistent with current 185-LOC behavior). Race tests via `go test -race` will catch any regressions.

### 4.4 Backward compatibility

- v2.x users with custom hook scripts targeting Notification / Elicitation / TaskCreated will lose hook routing after upgrade (Claude Code skips hook when settings.json omits the key). Migration: add event name to `system.yaml` `hook.observability_events` list.
- BC-V3R2-018 documented in CHANGELOG breaking-changes section.

### 4.5 Cross-platform behavior

- Windows: `subagent_stop.go:41-45` already short-circuits (no tmux). Other handlers are platform-neutral.
- macOS: full support (primary test platform per CI matrix).
- Linux: full support.

---

## 5. Quality Gates

Per `.moai/config/sections/quality.yaml`:

| Gate | Requirement | Verification command |
|------|-------------|-----------------------|
| Coverage | ≥ 85% per modified file | `go test -cover ./internal/hook/ ./internal/cli/ ./internal/config/` |
| Race | `go test -race ./...` clean | `go test -race -count=1 ./...` |
| Lint | golangci-lint clean | `golangci-lint run` |
| Build | embedded.go regenerated | `make build` |
| Audit | All 4 audit sub-tests PASS | `go test -run TestAudit ./internal/hook/` |
| MX | @MX tags applied per §6 | `moai mx scan internal/hook/` |
| Backward compat | settings.json migration documented | CHANGELOG entry + manual review |

---

## 6. @MX Tag Plan (mx_plan)

Apply per `.claude/rules/moai/workflow/mx-tag-protocol.md`. Language: ko (per `code_comments: ko`).

| File | Tag | Reason |
|------|-----|--------|
| `internal/hook/subagent_stop.go:30-37` | `@MX:ANCHOR @MX:REASON: P-H02 핵심 fix; tmux pane teardown 의 단일 경로 — 모든 teammate shutdown 이 이 함수를 통과` | fan_in ≥ 3 (Handle 함수가 deps.go + handler dispatcher + test 에서 호출) |
| `internal/hook/subagent_stop.go:128-134` | `@MX:WARN @MX:REASON: tmux kill-pane 호출이 외부 프로세스 (tmux) 에 의존하며 1500 ms SessionEnd 한도 내 완료 보장 필요; 500 ms timeout wrap 추가 (M3-T14)` | external blocking call |
| `internal/hook/observability.go:1-30` | `@MX:NOTE: 4 retire 핸들러의 runtime gate; system.yaml observability_events 미설정 시 핸들러 silent return; BC-V3R2-018 의 핵심 메커니즘` | non-obvious business rule |
| `internal/hook/audit_test.go:TestAuditRegistrationParity` | `@MX:ANCHOR @MX:REASON: 27-event coverage 의 CI lock; settings.json + deps.go + system.yaml 3-way parity 강제` | high fan_in (모든 신규 핸들러 추가가 이 test 통과 필수) |
| `internal/cli/doctor_hook.go:RunE` | `@MX:NOTE: 27-event 표 출력; SPEC-V3R2-RT-006 §5.7 의 single source of truth 가 doctor 출력으로 노출` | observability surface |
| `internal/config/types.go:SystemHookConfig` | `@MX:NOTE: hook.observability_events whitelist; validator/v10 tag 가 4 retire event 외 값 거부` | schema invariant |

---

## 7. Risk Mitigation Plan (spec §8 risks → run-phase tasks)

| spec §8 risk | Mitigation in run-phase |
|--------------|--------------------------|
| Row 1 — kill-pane 1500 ms 초과 | T-RT006-14 (500 ms goroutine + WithTimeout) |
| Row 2 — stale binary regression | sync-phase release note (T-RT006-32 CHANGELOG entry) |
| Row 3 — Retired handler 혼란 | T-RT006-12 (per-file Resolution header) + T-RT006-25 (doctor hook 표) |
| Row 4 — ConfigChange race | T-RT006-16 (20 ms debounce) |
| Row 5 — Windows tmux 미지원 | (covered by baseline subagent_stop.go:41-45) |
| Row 6 — 사용자 피로 | (covered by baseline instructions_loaded.go:48-51 warn-only) |
| Row 7 — MX rescan I/O | T-RT006-18 (delegated to mx subdir memoization, SPC-002 owned) |
| Row 8 — orphan setup 사용 | T-RT006-01 (file 자체 git rm) |

---

## 8. Dependencies (status as of `c810b11b7`)

### 8.1 Blocking (consumed)

- ✅ SPEC-V3R2-RT-001 (HookResponse / HookOutput) — merged.
- ✅ SPEC-V3R2-RT-002 (PreToolUse PermissionDecision) — merged. Consumer: REQ-022 semantic reaffirmation only.
- ✅ SPEC-V3R2-RT-004 (SessionState checkpointing) — merged. Consumer: REQ-027/-028 semantic reaffirmation.
- ✅ SPEC-V3R2-RT-005 (8-tier resolver) — merged at `ab0fc4dda`. Consumer: ConfigChange reload + SystemHookConfig.
- ⚠ SPEC-V3R2-SPC-002 (MX TagScanner) — status TBD at run-time. If not merged, T-RT006-18 uses interface stub.

### 8.2 Blocked by (none active)

All blockers cleared.

### 8.3 Blocks (downstream consumers)

- SPEC-V3R2-HRN-002: SubagentStart/SubagentStop semantics for evaluator memory.
- SPEC-V3R2-WF-003: Multi-mode router consumes FileChanged MX rescan output.
- SPEC-V3R2-MIG-002: Hook registration cleanup aligns to 21-native + 1-composite + 4-observability count.

---

## 9. Verification Plan

### 9.1 Pre-merge verification (run-phase end)

- [ ] All 26 tasks (T-RT006-01..26) complete per tasks.md
- [ ] All 17 ACs (AC-V3R2-RT-006-01..17) verified per acceptance.md
- [ ] `go test -race -count=1 ./...` PASS (no regressions)
- [ ] `golangci-lint run` clean
- [ ] `make build` regenerates `internal/template/embedded.go` correctly
- [ ] `moai doctor hook` (manual) prints 27-row table
- [ ] CHANGELOG entry written in Unreleased section
- [ ] @MX tags applied per §6 (verified via `moai mx scan`)

### 9.2 Plan-auditor target

- [ ] All 22 unique REQs mapped in §1.4 traceability matrix
- [ ] All 17 ACs mapped to ≥1 task
- [ ] No orphan tasks (every task supports ≥1 REQ)
- [ ] research.md evidence anchors cited (≥30 per §1 mandate)
- [ ] §1.2.1 explicitly addresses spec.md §5.7 footer reconciliation
- [ ] HookResponse/HookOutput naming alias acknowledged
- [ ] BC-V3R2-018 retire mechanism design (Option A) justified
- [ ] Worktree-base alignment per Step 2 (run-phase) called out
- [ ] §6 mx_plan covers ≥3 of {ANCHOR, WARN, NOTE} types
- [ ] No time estimates anywhere (P0/P1 priority labels only)
- [ ] Parallel SPEC isolation: this plan touches only `internal/hook/`, `internal/cli/`, `internal/config/types.go`, `internal/template/`, `.claude/settings.json`, `.moai/config/sections/system.yaml`, `CHANGELOG.md`.

### 9.3 Plan-auditor risk areas (front-loaded mitigations)

- **Risk: spec.md §5.7 footer count drift (15 vs 17 KEEP)** → addressed in §1.2.1 Acknowledged Discrepancies + research.md §3.5/§3.6.
- **Risk: HookResponse vs HookOutput type naming inconsistency** → addressed in §1.2.1 + audit_test grep covers both names.
- **Risk: SPC-002 TagScanner not yet merged at run-time** → T-RT006-18 uses interface stub + mock test; integration deferred per §8.1 ⚠ note.
- **Risk: settings.json template branching breaks `moai update` rendering for users without observability opt-in** → T-RT006-11 default empty map + Go template `{{- if }}` ensures no rendering when key absent.
- **Risk: 4 retire handlers' Go code paths confuse maintainers** → §3.1 file-level map + per-file Resolution headers + doctor_hook --observability flag.
- **Risk: `c810b11b7` baseline drift if main advances during plan PR review** → run-phase explicitly rebases on `origin/main` (Step 2 `moai worktree new --base origin/main`).

---

## 10. Run-Phase Entry Conditions

After plan PR squash-merged into main:

1. `git checkout main && git pull` (host checkout).
2. `moai worktree new SPEC-V3R2-RT-006 --base origin/main` per Step 2 spec-workflow.
3. `cd ~/.moai/worktrees/moai-adk/SPEC-V3R2-RT-006`.
4. `git rev-parse --show-toplevel` should output the worktree path (Block 0 verification per session-handoff.md).
5. `git rev-parse HEAD` should match plan-merge commit SHA on main.
6. `/moai run SPEC-V3R2-RT-006` invokes Phase 0.5 plan-audit gate, then proceeds to M1.

---

Version: 0.1.0
Status: Plan artifact for SPEC-V3R2-RT-006
Run-phase methodology: TDD (per `.moai/config/sections/quality.yaml` `development_mode: tdd`)
Estimated artifacts: 4 new files + 22 modified handlers + 1 file removed + 1 audit test extension = ~580 LOC delta
