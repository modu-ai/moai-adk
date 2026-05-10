# SPEC-V3R2-RT-006 Task List

> Task decomposition for **Hook Handler Completeness and 27-Event Coverage**.
> Companion to `spec.md`, `plan.md`, `research.md`, `acceptance.md`.
> All tasks scoped to run-phase (post plan-PR-merge); milestone breakdown follows `plan.md` § 2.

## HISTORY

| Version | Date       | Author                        | Description                                                                                                  |
|---------|------------|-------------------------------|--------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow)  | Initial task decomposition. 35 tasks (T-RT006-01..35) grouped by 5 milestones (M1-M5) per `plan.md` § 2. |

---

## Task ID convention

`T-RT006-NN` where NN is zero-padded two-digit ordinal within milestone, increasing across milestones.

REQ → Task mapping is the canonical traceability matrix in `plan.md` § 1.4. AC → Task mapping is in `acceptance.md` § 4.

---

## Milestone M1: Test scaffolding + setup.go removal (RED) — Priority P0

| ID | Subject | Files | REQ refs | AC refs | Type |
|----|---------|-------|----------|---------|------|
| T-RT006-01 | Delete orphan `internal/hook/setup.go` (0-byte file). Verify no `NewSetupHandler` reference exists. Run `go build ./...` to confirm. | `internal/hook/setup.go` (rm) | REQ-005 | AC-09 | structural |
| T-RT006-02 | Add 4 RED audit sub-tests in `internal/hook/audit_test.go`: TestAuditRegistrationParity, TestAuditPerFileCategoryHeader, TestAuditRetiredEventsNotInSettings, TestAuditObservabilityWhitelist. All should FAIL initially. | `internal/hook/audit_test.go` | REQ-003, -042, -063 | AC-10, -13, -14 | RED test |
| T-RT006-03 | Add 6 RED AC-path tests across 5 critical handler test files. All should FAIL initially. | `internal/hook/{subagent_stop,config_change,instructions_loaded,file_changed,post_tool_failure}_test.go` | REQ-010, -011, -012, -013, -014 | AC-01, -02, -04, -05, -06, -07, -08 | RED test |
| T-RT006-04 | Add `hook.observability_events: []` + `strict_mode: false` schema to `.moai/config/sections/system.yaml` (local) and `internal/template/templates/.moai/config/sections/system.yaml` (template). | `.moai/config/sections/system.yaml`, `internal/template/templates/.moai/config/sections/system.yaml` | REQ-004 | AC-11 | schema |

---

## Milestone M2: SystemHookConfig + RT-005 typed loader integration (GREEN seed) — Priority P0

| ID | Subject | Files | REQ refs | AC refs | Type |
|----|---------|-------|----------|---------|------|
| T-RT006-05 | Add `SystemHookConfig` struct to `internal/config/types.go` with validator/v10 tags. Extend `SystemConfig` with `Hook SystemHookConfig` field. Verify RT-005 audit_test (TestAuditParity) passes. | `internal/config/types.go` | REQ-004 | AC-11 | struct |
| T-RT006-06 | Implement `observabilityOptIn(ctx, eventName) bool` helper in new file `internal/hook/observability.go`. Reads `config.FromContext(ctx).System.Hook.ObservabilityEvents`. Returns false on nil cfg or missing event. | `internal/hook/observability.go` (new) | REQ-040 | AC-11, -16 | helper |
| T-RT006-07 | Apply observability gate to 4 retire handlers: `notification.go`, `elicitation.go` (both ElicitationHandler + ElicitationResultHandler), `task_created.go`. Use Pattern A: silent return if `!observabilityOptIn()`. Never emit SystemMessage / Continue:false. | `internal/hook/{notification,elicitation,task_created}.go` | REQ-040 | AC-11, -16 | logic |
| T-RT006-08 | Verify M1-T02 RED test TestAuditObservabilityWhitelist now GREEN. Run `go test ./internal/hook/ -run TestAuditObservabilityWhitelist`. | (test only) | REQ-040 | AC-11 | verify |

---

## Milestone M3: Settings.json retire + Resolution headers + handler upgrades (GREEN main) — Priority P0

| ID | Subject | Files | REQ refs | AC refs | Type |
|----|---------|-------|----------|---------|------|
| T-RT006-09 | Remove 4 retire entries (Notification, Elicitation, ElicitationResult, TaskCreated) from `internal/template/templates/.claude/settings.json.tmpl`. Add Go template `{{- if .ObservabilityEvents.<EventName> }}...{{- end }}` branching for opt-in render. | `internal/template/templates/.claude/settings.json.tmpl` | REQ-004 | AC-10 | template |
| T-RT006-10 | Remove 4 retire entries from local `.claude/settings.json`. Verify `moai update --dry-run` does not recreate them when system.yaml has empty observability_events. | `.claude/settings.json` | REQ-004 | AC-10 | local |
| T-RT006-11 | Add template context field `ObservabilityEvents map[string]bool` in `internal/template/context.go`. Render path reads from system.yaml at `moai update` deploy time. Default empty map (no opt-in). | `internal/template/context.go` | REQ-004 | AC-10, -11 | template |
| T-RT006-12 | Apply `// Resolution: <CATEGORY>` header to 22 handler files (KEEP=17, UPGRADE=4, FIX=1, COMPOSITE=1; RETIRE-OBS-ONLY=4 already done). Use lookup table per spec §5.7. | 22 files in `internal/hook/*.go` | REQ-002 | AC-14 | annotation |
| T-RT006-13 | Verify TestAuditPerFileCategoryHeader now GREEN. Regex `^// Resolution: (KEEP\|UPGRADE\|FIX\|REMOVE\|RETIRE-OBS-ONLY\|COMPOSITE)` matches every non-test, non-aux handler file. | (test only) | REQ-002 | AC-14 | verify |
| T-RT006-14 | Add 500 ms per-pane timeout wrap to `subagent_stop.go` killTmuxPane call. Use goroutine + `context.WithTimeout`. Verify TestSubagentStop_PaneNotFoundGraceful + new TestSubagentStop_KillPaneTimeout PASS. | `internal/hook/subagent_stop.go` | REQ-006, -007, -010, -060, -061 | AC-01, -02, -03, -15 | logic |
| T-RT006-15 | Wire RT-005 `Manager.Reload(path)` call into `config_change.go` after successful YAML validation. On reload error, return `Continue:false + SystemMessage` "Reload rejected; old settings retained". Verify TestConfigChange_RT005ReloadIntegration GREEN. | `internal/hook/config_change.go` | REQ-011 | AC-04 | logic |
| T-RT006-16 | Add 20 ms debounce to `config_change.go` (sleep before ReadFile to mitigate fsnotify mid-write race). Verify TestConfigChange_InvalidYAMLKeepsOldSettings GREEN. | `internal/hook/config_change.go` | REQ-011, -062 | AC-05 | logic |
| T-RT006-17 | Verify TestInstructionsLoaded_42kCLAUDE GREEN (existing 89-LOC handler covers; AC-06 path test only). | (test only) | REQ-012 | AC-06 | verify |
| T-RT006-18 | Wire SPC-002 TagScanner stub into `file_changed.go`. If SPC-002 not yet merged, use interface placeholder + mock test. Verify TestFileChanged_MXTagDelta GREEN. | `internal/hook/file_changed.go` | REQ-013 | AC-07 | integration |
| T-RT006-19 | Implement TestAuditRegistrationParity body in `audit_test.go`. Parse settings.json, deps.go, system.yaml. Assert: 26 deps == 21 native + 1 composite + 4 obs-residual. | `internal/hook/audit_test.go` | REQ-003, -042 | AC-13 | test impl |
| T-RT006-20 | Verify TestPostToolFailure_TimeoutClassification GREEN (existing 134-LOC handler covers; AC-08 path test). | (test only) | REQ-014 | AC-08 | verify |

---

## Milestone M4: doctor hook CLI + observability list reflection — Priority P1

| ID | Subject | Files | REQ refs | AC refs | Type |
|----|---------|-------|----------|---------|------|
| T-RT006-21 | Implement TestAuditRetiredEventsNotInSettings body. Parse `.claude/settings.json`. Assert 4 retire keys absent. After M3-T10 complete, this PASSes. | `internal/hook/audit_test.go` | REQ-063 | AC-10 | test impl |
| T-RT006-22 | Implement TestStrictModeRetiredEvent (REQ-041 negative). With `strict_mode: true` + retire event firing (mocked), assert handler returns `HookOutput{}` silently. | `internal/hook/audit_test.go` | REQ-041 | (negative) | test impl |
| T-RT006-23 | Create `internal/hook/coverage_table.go` (~80 LOC). Build 27-event table data structure with fields: EventName, Resolution, IsActive, ObservabilityOptIn. Used by audit_test + doctor_hook. | `internal/hook/coverage_table.go` (new) | REQ-050 | AC-12 | data |
| T-RT006-24 | Create `internal/cli/doctor_hook.go` (~150 LOC). cobra subcommand `moai doctor hook`. Flags: `--json`, `--trace <event>`, `--observability`. RunE builds coverage table from HookRegistry + cfg, prints text or JSON. | `internal/cli/doctor_hook.go` (new) | REQ-050 | AC-12 | CLI |
| T-RT006-25 | Wire `doctor_hook` into `internal/cli/doctor.go` parent. Add to `moai doctor` help. | `internal/cli/doctor.go` | REQ-050 | AC-12 | CLI wire |
| T-RT006-26 | Implement `--trace <event>` reading from `.moai/logs/hook.log` last N lines for the named event. (Out-of-scope simplification: tail-only readout, no real-time stream.) | `internal/cli/doctor_hook.go` | REQ-051 | AC-12 | CLI |
| T-RT006-27 | Add `internal/cli/doctor_hook_test.go` (~80 LOC). Test 27-event table count, default observability_events empty, --json output JSON-parseable, --trace non-existent event no-op. | `internal/cli/doctor_hook_test.go` (new) | REQ-050, -051 | AC-12 | test |

---

## Milestone M5: Verification gates + audit consolidation — Priority P0

| ID | Subject | Files | REQ refs | AC refs | Type |
|----|---------|-------|----------|---------|------|
| T-RT006-28 | Run full test suite: `go test ./... -race -count=1`. Ensure 0 regressions. Fix any flaky tests introduced. | (all) | (all) | (all) | gate |
| T-RT006-29 | Run linter: `golangci-lint run`. Fix any issues introduced by header insertion or new code. | (all) | (all) | (all) | gate |
| T-RT006-30 | Verify `make build` succeeds and `internal/template/embedded.go` is regenerated correctly with the modified settings.json.tmpl + system.yaml. | `internal/template/embedded.go` | REQ-004 | (build) | gate |
| T-RT006-31 | Manual integration test: spawn live tmux team mode (3+ teammates), verify SubagentStop pane cleanup works end-to-end. Record session in `.moai/specs/SPEC-V3R2-RT-006/progress.md`. Failure here triggers escalation per spec §8 row 1. | (manual) | REQ-006, -010 | AC-01, -15 | gate |
| T-RT006-32 | Update `CHANGELOG.md` Unreleased section with bullets: (a) "fix(hook/SPEC-V3R2-RT-006): tmux pane leak on SubagentStop (P-H02)"; (b) "feat(hook/SPEC-V3R2-RT-006): retire 4 hook events from settings.json with observability opt-in (BC-V3R2-018)"; (c) "feat(cli/SPEC-V3R2-RT-006): add `moai doctor hook` 27-event diagnostic". | `CHANGELOG.md` | (Trackable) | (Trackable) | docs |
| T-RT006-33 | Add @MX tags per `plan.md` § 6 mx_plan (6 tags across 5 files). Verify with `moai mx scan internal/hook/`. | 6 files | REQ-002 | AC-14 | mx |
| T-RT006-34 | Re-run `go test ./internal/hook/ -run TestAudit`. All 4 sub-tests GREEN. | (verify) | REQ-003, -042, -063 | AC-10, -13, -14 | gate |
| T-RT006-35 | Verify spec.md §5.7 footer count via `moai doctor hook --json | jq '.summary'`. Expect 17 KEEP / 4 UPGRADE / 1 FIX / 4 RETIRE / 1 REMOVE = 27. If mismatch, update sync-phase HISTORY of spec.md (handled in /moai sync). | (verify) | REQ-050 | AC-12 | gate |

---

## Task summary

- **Total**: 35 tasks (T-RT006-01..35)
- **By milestone**: M1=4, M2=4, M3=12, M4=7, M5=8
- **By type**: structural=2, RED test=2, GREEN logic=8, GREEN test impl=4, CLI=4, schema=4, gate=8, docs=1, mx=1, verify=4, integration=1
- **By priority**: P0=28 (M1, M2, M3, M5), P1=7 (M4 doctor hook)

## Ownership

- **Lead role profile**: `expert-backend` (Go file edits, RT-005 integration, audit_test) for M1-M3, M5.
- **CLI role profile**: `expert-backend` for M4 doctor_hook.
- **Manual gate**: `manager-quality` for M5-T31 live tmux integration test.
- **Sync phase only**: `manager-docs` for CHANGELOG + spec.md HISTORY update.

## Run-phase entry condition

After plan PR squash-merged into main, all tasks executed inside `feat/SPEC-V3R2-RT-006` worktree per `.claude/rules/moai/workflow/spec-workflow.md` § Phase Discipline Step 2.

---

Version: 0.1.0
Last Updated: 2026-05-10
