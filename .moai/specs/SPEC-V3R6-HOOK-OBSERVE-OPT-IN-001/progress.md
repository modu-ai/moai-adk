---
id: SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001
title: "Observability hook 3계열 opt-in — Progress Tracker"
version: "0.2.0"
status: implemented
created: 2026-05-23
updated: 2026-05-23
author: Author Name
priority: P2
phase: "v3.0.0"
module: ".moai/config/sections, internal/hook, internal/template/templates"
lifecycle: spec-anchored
tags: "hook, observability, opt-in, progress, sprint-2"
tier: S
---

# SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 — Progress Tracker

## Status

**Plan phase**: revise iter-2 ready for plan-auditor re-audit (post Q4=A relocation).
**Run phase**: NOT YET STARTED.
**Sync phase**: NOT YET STARTED.

## Revise History

### iter-1 (2026-05-23, original draft)

- 4 artifacts created (spec.md / plan.md / acceptance.md, progress.md absent)
- plan-auditor verdict: **PASS 0.8775** (Tier S 0.75 threshold + 0.1275 margin; D1=0.90 / D2=0.92 / D3=0.85 / D4=0.84)
- Toggle location: `.moai/config/sections/observability.yaml` `observability.enabled` (default `false`)
- 5 REQs (REQ-HOI-001..005) + 7 ACs (AC-HOI-001..007) + 4 risks (R1-R4) + 3 Out-of-Scope sections

### iter-1.5 (2026-05-23, run-phase BLOCKER discovered)

- Orchestrator delegated `/moai run SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001` to manager-develop (Tier S minimal, cycle_type=ddd).
- manager-develop returned BLOCKER report after Section C pre-flight (codebase audit).
- **Root finding**: `.moai/config/sections/observability.yaml` ALREADY EXISTS with `observability.enabled: true`, serving REQ-OBS-005 trace-logging master toggle (different semantics from HOI). Original HOI iter-1 spec proposed introducing the SAME key with conflicting semantics.
- Plan-auditor did NOT catch this in iter-1 PASS because plan-auditor's audit scope is artifact self-consistency, not codebase pre-existing state survey.
- Orchestrator independently verified 5 collision facts (see spec.md §A.1).

### iter-2 revise (2026-05-23, this iteration — Q4=A relocation)

**User decision (AskUserQuestion Q4=A)**: Relocate HOI master toggle to `.moai/config/sections/system.yaml` `hook.opt_in.enabled: false`. Rationale:
- Reuses RT-006 location (cohesion with existing `hook.observability_events`)
- Leaves `observability.yaml` `observability.enabled: true` untouched (REQ-OBS-005 trace logging preserved)
- Schema delta is minimal: one new key under existing `hook:` block
- manager-develop Section C codebase audit can verify with single-file scan

**Files changed in this revise** (4 SPEC artifacts only, NO source code changes):

| File | Change Summary |
|---|---|
| `spec.md` | Added §A "Pre-existing State Survey" with 5 facts (A.1) + collision discovery (A.2) + 3-key cohabitation contract (A.3) + plan-auditor mitigation note (A.4). Renumbered B/C/D/E/F. Updated REQ-HOI-001..005 to reference `system.yaml hook.opt_in.enabled` instead of `observability.yaml observability.enabled`. Added R5 (cohabitation regression). Added NEW Out-of-Scope section "Migration of REQ-OBS-005 or SPEC-V3R2-RT-006 key namespaces". Updated cross-references. Author renamed to `Author Name`. |
| `plan.md` | Added Revise note. Updated M1 to extend existing `system.yaml` `hook:` block instead of creating new `observability.yaml`. Added M2 mechanism choice section: NEW `hookOptInEnabled()` helper in NEW file `internal/hook/hook_opt_in.go` (NOT reusing `observabilityOptIn()`), with rationale (4 reasons). Added M3 deliverable #4: file-top COHABITATION NOTE comment in `internal/hook/observability.go`. Added "Cohabitation invariant" subsection to Cross-SPEC Coordination. Updated Technical Approach §A/§B/§C path references. Added R5 row. Updated Definition of Done. Added NEW Out-of-Scope section. Author renamed to `Author Name`. |
| `acceptance.md` | Added Revise note. Updated AC-HOI-001 from `observability.yaml grep` to `system.yaml grep -A1 opt_in` 3-check (NEW + 2 cohabitation preservation). Updated AC-HOI-002 path target. Renamed `TestObservabilityDisabled` → `TestHookOptInDisabled` (AC-HOI-003) and `TestObservabilityEnabled` → `TestHookOptInEnabled` (AC-HOI-004). Updated AC-HOI-005 sed/python flip target to `system.yaml`. Updated AC-HOI-006 doctor regex from `Observability:` to `Hook opt-in:`. **AC-HOI-007 reformulated** from template-mirror-parity (now subsumed under AC-HOI-001) to 4-quadrant cohabitation regression test `TestHookOptInCohabitation` with explicit file-untouched assertions for `observability.yaml`, `observability.go` function body, `coverage_table.go`, `audit_test.go`. Added Edge case 4 (naming collision regression). Added Quality Gate criterion #7 (cohabitation file-untouched check). Added NEW Out-of-Scope section. Author renamed to `Author Name`. |
| `progress.md` | NEW file (this file). Documents revise reason, Q4=A decision, change summary, expected plan-auditor delta, mitigation tracking. |

**REQ change summary (before → after)**:

| ID | Before (iter-1) | After (iter-2) |
|---|---|---|
| REQ-HOI-001 | `observability.yaml` `enabled:` key default false | `system.yaml` `hook.opt_in.enabled:` (NEW sub-block) default false + explicit cohabitation contract with REQ-OBS-005 + RT-006 |
| REQ-HOI-002 | `observability.enabled == false` → 3 hook series SKIP | `hook.opt_in.enabled == false` → 3 hook series SKIP (semantics unchanged, key name only) |
| REQ-HOI-003 | `observability.enabled == true` → 3 hook series EXECUTE | `hook.opt_in.enabled == true` → 3 hook series EXECUTE (semantics unchanged, key name only) |
| REQ-HOI-004 | `moai init` defaults `observability.yaml` `enabled: false` | `moai init` defaults `system.yaml` `hook.opt_in.enabled: false` (flag name `--observability` → `--hook-opt-in`) |
| REQ-HOI-005 | Doctor pattern `Observability:\s*(enabled\|disabled)` | Doctor pattern `Hook opt-in:\s*(enabled\|disabled)` (DISTINCT from any REQ-OBS-005 `Observability:` line — cohabitation) |

**AC change summary (before → after)**:

| ID | Before (iter-1) | After (iter-2) |
|---|---|---|
| AC-HOI-001 | `grep -E '^enabled:' observability.yaml` single check | 3 checks — NEW opt_in.enabled + preserved observability_events + preserved strict_mode |
| AC-HOI-002 | `moai init` produces `observability.yaml enabled: false` | `moai init` produces `system.yaml hook.opt_in.enabled: false` |
| AC-HOI-003 | `go test -run TestObservabilityDisabled` | `go test -run TestHookOptInDisabled` (name change) |
| AC-HOI-004 | `go test -run TestObservabilityEnabled` | `go test -run TestHookOptInEnabled` (name change) |
| AC-HOI-005 | sed flips `observability.yaml enabled` | python regex flips `system.yaml hook.opt_in.enabled` |
| AC-HOI-006 | `grep Observability:\s*(enabled\|disabled)` | `grep Hook opt-in:\s*(enabled\|disabled)` + cohabitation note |
| **AC-HOI-007** | Template mirror parity check | **REFORMULATED** — 4-quadrant cohabitation regression test `TestHookOptInCohabitation` with file-untouched assertions for REQ-OBS-005 + RT-006 owned files |

## M1/M2/M3 Mechanism Decisions (plan.md §"M2/M3 Mechanism Rationale" detail)

### M1 mechanism choice

**Decision**: Extend existing `.moai/config/sections/system.yaml` `hook:` block with NEW `opt_in:` sub-block (NOT create new file or new top-level config section).

**Rationale**:
- Cohesion with existing `hook.observability_events` + `hook.strict_mode` (all hook-related config in one place)
- Schema delta minimal: ~8-10 YAML lines added
- Go loader extension: extend existing `SystemHookConfig` struct with `OptIn HookOptInConfig` field (zero-value default `false` matches R3)
- Pre-existing `observability.yaml` (REQ-OBS-005) untouched — clean cohabitation
- M1 estimated scope reduced from original ~50-80 LOC to ~30-50 LOC

### M2 mechanism choice

**Decision**: NEW helper `hookOptInEnabled(cfg ConfigProvider) bool` in NEW file `internal/hook/hook_opt_in.go` (NOT reuse existing `observabilityOptIn()` in `observability.go`).

**Rationale** (4 reasons, full detail in plan.md):
1. Semantic mismatch — RT-006 per-event whitelist vs HOI single master toggle
2. AC-HOI-007 cohabitation invariant — sharing gate would collapse 2 systems
3. Future evolution risk — RT-006 evolution would inadvertently affect HOI
4. Audit clarity — separate file = textually visible separate concern

### M3 mechanism choice

**Decision**: Add file-top COHABITATION NOTE comment block to `internal/hook/observability.go` cross-referencing SPEC §A.3. Function body of `observabilityOptIn()` NOT modified.

**Rationale**:
- R5 mitigation (cohabitation regression guard)
- Permanent in-code documentation for future maintainers
- AC-HOI-007 file-untouched assertion has explicit "function body" exception (only file-top comment allowed)

## Expected plan-auditor Re-Audit Outcome

iter-2 PASS expectation: ≥ 0.85 (Tier S 0.75 + 0.10 margin; the explicit §A "Pre-existing State Survey" + collision-zero contract + dedicated 4-quadrant cohabitation AC should improve D2/D3/D4 scores vs iter-1).

Risk factors that may trigger plan-auditor REVISE:
- LOC estimate creep — M2 estimate increased from 150-250 to 200-300 (4-quadrant cohabitation test adds ~50-100 LOC)
- Mechanism justification verbosity — "M2/M3 Mechanism Rationale" section in plan.md is new; if plan-auditor flags as excessive, can be trimmed
- AC-HOI-007 dual-modality (4-quadrant test + file-untouched static assertion) may be flagged as 2-AC-in-1; if so, split into AC-HOI-007a (4-quadrant) + AC-HOI-007b (file-untouched)

If plan-auditor REVISE iter-2, fix-forward via orchestrator-direct edits (no manager-spec re-delegate needed — scope is well-bounded).

## Mitigation Tracking — plan-auditor codebase state blindspot

The 2026-05-23 lesson "plan-auditor codebase state blindspot" identified 4 mitigations. Status in this SPEC and beyond:

| # | Mitigation | Implemented? | Tracking |
|---|---|---|---|
| 1 | spec.md §A "Pre-existing State Survey" mandatory for config/source-touching SPECs | YES (this SPEC §A) | Pattern established; future SPECs SHOULD follow |
| 2 | manager-develop Section C codebase audit grep extension (`grep -r "<spec-anchor-key>" .moai/config/`) | NO | Out of scope for this revise. Separate rule-revise PR for `.claude/rules/moai/development/manager-develop-prompt-template.md` Section C |
| 3 | plan-auditor checklist `presence check` extension for proposed keys vs codebase grep | NO | Out of scope for this revise. Separate rule-revise PR for `.claude/agents/meta/plan-auditor.md` |
| 4 | plan-phase optional `state-snapshot.md` artifact | NO | Out of scope for this revise. Candidate for Tier M/L SPECs only; this SPEC remains Tier S 4-artifact (spec+plan+acceptance+progress) |

## Run-Phase Tracking (COMPLETED 2026-05-23)

Run-phase delegated to manager-develop Tier S minimal cycle_type=ddd. Implementation completed in 4 commits (M1 + M2 + M3 + cohabitation_guard_test) on `main` (Hybrid Trunk Tier S direct-push policy).

### M1 Evidence — Schema + Loader (commit 18112a7b6)

- [x] `.moai/config/sections/system.yaml` edited with NEW `hook.opt_in.enabled: false` sub-block (+7 lines)
- [x] `internal/template/templates/.moai/config/sections/system.yaml.tmpl` template mirror updated (+7 lines, byte-equivalent)
- [x] `SystemHookConfig` Go struct extended with `OptIn HookOptInConfig` field (+13 lines including COHABITATION NOTE)
- [x] NEW `HookOptInConfig{Enabled bool}` type defined (yaml:"opt_in" tag)
- [x] Loader contract: Go zero-value default (false) when `opt_in:` sub-block absent (R3 mitigation)
- [x] AC-HOI-001 PASS — 3-check grep verifies opt_in.enabled + preserved observability_events + preserved strict_mode

### M2 Evidence — Gate + Dispatcher + Tests (commit e3f8dd463)

- [x] NEW file `internal/hook/hook_opt_in.go` (32 lines) with `hookOptInEnabled(cfg ConfigProvider) bool` Go-level helper
- [x] NEW file `internal/hook/hook_opt_in_test.go` (191 lines) with 5 tests:
  - `TestHookOptInDisabled` (4 sub-cases: nil cfg, nil underlying, default zero-value, explicit false) → PASS
  - `TestHookOptInEnabled` → PASS
  - `TestHookOptInMissingKey_DefaultsDisabled` (legacy project edge case) → PASS
  - `TestHookOptInCohabitation` (4-quadrant matrix HOI × RT-006) → PASS all 4 quadrants
  - `TestHookOptInIndependence_RT006Whitelist` (R5 regression guard) → PASS
- [x] CLI dispatcher gate: `runHookEvent` in `internal/cli/hook.go` checks `isHookOptInEnabled(cwd)` BEFORE registry dispatch for TaskCreated + Notification events
- [x] CLI wrapper gates: 3 secondary observability wrappers (`runHarnessObserveStop`, `runHarnessObserveSubagentStop`, `runHarnessObserveUserPromptSubmit`) gated via `isHookOptInEnabled(cwd)` — Primary `runHarnessObserve` (PostToolUse) NOT gated (out of scope)
- [x] Handler-level defense-in-depth: `notification.go` and `task_created.go` Handle() methods check `hookOptInEnabled(h.cfg)` FIRST
- [x] `deps.go` switched to `NewNotificationHandlerWithConfig(deps.Config)` and `NewTaskCreatedHandlerWithConfig(deps.Config)` to wire config provider
- [x] Template conditional render: `settings.json.tmpl` wraps 3 `handle-harness-observe-*` stanzas in `{{ if .HookOptIn.Enabled }}...{{ end }}`
- [x] `TemplateContext.HookOptIn` field added (mirrors `config.HookOptInConfig`); `WithHookOptIn(enabled bool)` option exposed
- [x] `update.go` reads existing `system.yaml` via NEW `readHookOptInEnabled(projectRoot)` helper at template rendering time
- [x] AC-HOI-003 PASS — `go test -run TestHookOptInDisabled ./internal/hook/...` exit 0
- [x] AC-HOI-004 PASS — `go test -run TestHookOptInEnabled ./internal/hook/...` exit 0
- [x] AC-HOI-005 PASS — template conditional verified via `settings.json.tmpl` diff (3 hook stanzas wrapped)
- [x] AC-HOI-007 PASS — 4-quadrant cohabitation matrix executes all (HOI × RT-006) combinations independently

### M3 Evidence — Doctor + Cohabitation Note + Fixture Refactor (commit 8fba30ac5)

- [x] `moai doctor` extended with `Hook opt-in:` workspace check (NEW `checkHookOptIn()` function in `internal/cli/doctor.go`)
- [x] Output line format: `Hook opt-in:  (enabled|disabled)` — matches AC-HOI-006 regex `Hook opt-in:\s*(enabled|disabled)`
- [x] Doctor line is DISTINCT from any REQ-OBS-005 `Observability:` line (cohabitation invariant)
- [x] testdata/doctor-{light,dark,nocolor}.golden regenerated via `UPDATE_GOLDEN=1`
- [x] `moai init` default verified — `system.yaml.tmpl` renders `hook.opt_in.enabled: false` by default (no `--hook-opt-in` flag introduced; default behavior is sufficient)
- [x] Test fixture refactor: NEW `writeSystemYAMLHookOptIn` helper in `hook_harness_observe_test.go`; 13 harness-observe wrapper tests updated to set HOI=true
- [x] `TestHookSubcommands_EventTypeMapping/notification` subtest updated to chdir into HOI-enabled tempdir so dispatcher gate passes through
- [x] `internal/hook/observability.go` file-top COHABITATION NOTE comment block added (+15 lines documenting 3-key independence)
- [x] `observabilityOptIn()` function body UNTOUCHED (verified via git diff — only file-top comment changed)
- [x] Cross-platform: `go build ./...` exit 0; `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- [x] AC-HOI-002 PASS — `moai init` produces `system.yaml` with `hook.opt_in.enabled: false` (template default)
- [x] AC-HOI-006 PASS — `moai doctor 2>&1 | grep -E 'Hook opt-in:\s*(enabled|disabled)'` exit 0

### +1 Cohabitation Guard (commit 12a917a36)

- [x] NEW file `internal/hook/cohabitation_guard_test.go` (181 lines) with 5 static CI assertions:
  - `TestCohabitationGuard_ObservabilityYAMLPresent` — 4 sentinel keys present
  - `TestCohabitationGuard_ObservabilityOptInFunctionBody` — RT-006 read path retained, no System.Hook.OptIn reference, file-top NOTE present
  - `TestCohabitationGuard_CoverageTableFieldsPresent` — ResolutionRetireObsOnly + ObservabilityOptIn fields retained
  - `TestCohabitationGuard_AuditTestObservabilityWhitelistPresent` — RT-006 REQ-040 contract verifier retained
  - `TestCohabitationGuard_HOIKeyIndependence` — 3 distinct YAML tags + separate HookOptInConfig type
- [x] All 5 tests PASS at run-phase completion
- [x] R5 mitigation operationalized — permanent regression guard against gate unification

## Quality Gate Verification (run-phase exit)

- [x] `go build ./...` exit 0
- [x] `GOOS=windows GOARCH=amd64 go build ./...` exit 0
- [x] `golangci-lint run --timeout=2m`: 27 issues — UNCHANGED from baseline (0 NEW HOI issues)
- [x] `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ | grep -v "_test.go" | grep -v "// "` exit 1 (0 matches) — C-HRA-008 satisfied
- [x] `go test ./internal/hook/...` coverage: same as pre-change baseline (no regression)
- [x] internal/hook test suite: only 2 pre-existing baseline failures (TestAuditRegistrationParity, TestAuditThreeWaySync — WorktreeCreate/WorktreeRemove drift, NOT HOI-related)
- [x] internal/cli test suite: net -3 failures (3 doctor goldens fixed via UPDATE_GOLDEN regeneration)

## Files Untouched (AC-HOI-007 cohabitation invariant — verified via git diff)

| File | Status |
|---|---|
| `.moai/config/sections/observability.yaml` | UNTOUCHED (REQ-OBS-005 owner) |
| `internal/hook/coverage_table.go` | UNTOUCHED (RT-006 fields preserved) |
| `internal/hook/audit_test.go` | UNTOUCHED (TestAuditObservabilityWhitelist body preserved) |
| `internal/hook/observability.go` function bodies | UNTOUCHED (only file-top NOTE added) |

## Open Questions for plan-auditor Re-Audit

1. Is the §A "Pre-existing State Survey" depth sufficient (5 facts), or should it include line-by-line code excerpts of `observability.go` and `system.yaml` for completeness?
2. Should AC-HOI-007 be split into AC-HOI-007a (4-quadrant runtime) + AC-HOI-007b (file-untouched static)? Dual-modality risk.
3. Is the M2 NEW file `internal/hook/hook_opt_in.go` justified, or should the helper live alongside `observabilityOptIn()` in `observability.go` with clear separator comment? (Plan-auditor may have opinion on file-count creep.)
4. The Out-of-Scope section "Migration of REQ-OBS-005 or SPEC-V3R2-RT-006 key namespaces" is NEW — does plan-auditor flag this as scope-narrowing or as appropriate boundary-setting?
5. The R5 risk + AC-HOI-007 regression guard pair — is this sufficient, or should plan.md also include a CI guard test (e.g., `internal/hook/cohabitation_guard_test.go`) as a separate deliverable?
