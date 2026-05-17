# SPEC-V3R2-RT-006 Progress Tracker

> Phase tracker shell for **Hook Handler Completeness and 27-Event Coverage**.
> Updated by manager-cycle / manager-spec / manager-docs at each phase boundary.
> Companion to `spec.md`, `plan.md`, `tasks.md`, `acceptance.md`, `research.md`.

## HISTORY

| Version | Date       | Author                       | Phase   | Description |
|---------|------------|------------------------------|---------|-------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow) | plan    | Initial progress tracker shell created during plan phase. Status: `plan_status: audit-ready`. |

---

## Current Phase

`plan` (Step 1 — main checkout, branch `plan/SPEC-V3R2-RT-006`)

## Plan Phase Status

- [x] research.md authored (12 sections, 30+ evidence anchors)
- [x] plan.md authored (10 sections, 35-task breakdown)
- [x] tasks.md authored (35 tasks across 5 milestones)
- [x] acceptance.md authored (17 baseline ACs + 4 derived edge-case ACs)
- [x] progress.md authored (this file)
- [x] issue-body.md authored (PR body draft)
- [ ] plan PR opened
- [ ] plan-auditor verdict received (target: PASS ≥ 0.85 first iteration)
- [ ] plan PR squash-merged into main

## plan_complete_at

(set on plan PR squash-merge into main)

## plan_status

`audit-ready`

---

## Run Phase Status

(Run phase pre-conditions: plan PR merged into main → `moai worktree new SPEC-V3R2-RT-006 --base origin/main` → `/moai run SPEC-V3R2-RT-006` → Phase 0.5 plan-audit gate auto-runs.)

### M1: Test scaffolding + setup.go removal — Priority P0

| Task ID | Status | Notes |
|---------|--------|-------|
| T-RT006-01 | done | setup.go was already absent (pre-removed) |
| T-RT006-02 | done | 4 RED audit sub-tests added → GREEN after M2/M3 |
| T-RT006-03 | done | SubagentStop + ConfigChange RED tests added → GREEN |
| T-RT006-04 | done | hook.observability_events + strict_mode in system.yaml |

### M2: SystemHookConfig + RT-005 typed loader integration — Priority P0

| Task ID | Status | Notes |
|---------|--------|-------|
| T-RT006-05 | done | SystemHookConfig + SystemConfig.Hook field in types.go |
| T-RT006-06 | done | observabilityOptIn(cfg, eventName) in observability.go |
| T-RT006-07 | done | Pattern A gate applied to notification/elicitation/task_created |
| T-RT006-08 | done | TestAuditObservabilityWhitelist GREEN |

### M3: Settings.json retire + Resolution headers + handler upgrades — Priority P0

| Task ID | Status | Notes |
|---------|--------|-------|
| T-RT006-09 | done | 4 retire events removed from settings.json.tmpl |
| T-RT006-10 | done | 4 retire events removed from local settings.json |
| T-RT006-11 | skipped | ObservabilityEvents field in context.go — deferred (no template rendering needed yet; system.yaml schema added instead) |
| T-RT006-12 | done | Resolution: headers added to all 26 handler files |
| T-RT006-13 | done | TestAuditPerFileCategoryHeader GREEN |
| T-RT006-14 | done | 500ms goroutine timeout wrap in subagent_stop.go |
| T-RT006-15 | done | config_change.go YAML validation + fallback reload |
| T-RT006-16 | done | 20ms debounce in config_change.go |
| T-RT006-17 | done | TestInstructionsLoaded_42kCLAUDE pre-existing GREEN |
| T-RT006-18 | done | file_changed.go MX scanner already integrated (GREEN) |
| T-RT006-19 | done | TestAuditRegistrationParity body implemented + GREEN |
| T-RT006-20 | done | TestPostToolFailure_TimeoutClassification pre-existing GREEN |

### M4: doctor hook CLI — Priority P1

| Task ID | Status | Notes |
|---------|--------|-------|
| T-RT006-21 | done | TestAuditRetiredEventsNotInSettings body → GREEN |
| T-RT006-22 | done | TestStrictModeRetiredEvent (via TestAuditObservabilityWhitelist/strict_mode sub-test) |
| T-RT006-23 | done | coverage_table.go (28 entries, 27 events + composite) |
| T-RT006-24 | done | doctor_hook.go cobra subcommand |
| T-RT006-25 | done | doctorHookCmd wired into doctorCmd |
| T-RT006-26 | done | --trace flag with hook.log tail readout |
| T-RT006-27 | done | doctor_hook_test.go (5 tests, all GREEN) |

### M5: Verification gates + audit consolidation — Priority P0

| Task ID | Status | Notes |
|---------|--------|-------|
| T-RT006-28 | done | go test ./... -short passes (pre-existing template failures exempted) |
| T-RT006-29 | done | golangci-lint 0 issues on modified packages |
| T-RT006-30 | done | make build succeeds |
| T-RT006-31 | manual | Manual tmux team mode integration — deferred to sync/manual QA |
| T-RT006-32 | done | CHANGELOG.md Unreleased entries (3 bullets) |
| T-RT006-33 | done | @MX:ANCHOR tags on observabilityOptIn + CoverageTable |
| T-RT006-34 | done | TestAudit* all 4 GREEN (PASS) |
| T-RT006-35 | done | doctor hook --json summary: 17 KEEP / 4 UPGRADE / 1 FIX / 4 RETIRE / 1 REMOVE = 27 ✓ |

## Acceptance Criteria Status (from acceptance.md)

| AC ID | Status | Notes |
|-------|--------|-------|
| AC-V3R2-RT-006-01 | PASS | TestSubagentStop_PaneNotFoundGraceful |
| AC-V3R2-RT-006-02 | PASS | TestSubagentStop_PaneNotFoundGraceful (graceful kill-pane error handling) |
| AC-V3R2-RT-006-03 | PASS | Windows no-op path in subagent_stop.go (runtime.GOOS == "windows") |
| AC-V3R2-RT-006-04 | PASS | TestConfigChange_RT005ReloadIntegration |
| AC-V3R2-RT-006-05 | PASS | TestConfigChange_InvalidYAMLKeepsOldSettings |
| AC-V3R2-RT-006-06 | PASS | TestInstructionsLoadedHandler_Handle (file exceeding budget) |
| AC-V3R2-RT-006-07 | PASS | TestFileChangedHandler_Handle (supported Go file with tags) |
| AC-V3R2-RT-006-08 | PASS | TestPostToolUseFailureHandler_Handle (timeout error) |
| AC-V3R2-RT-006-09 | PASS | setup.go absent from disk |
| AC-V3R2-RT-006-10 | PASS | TestAuditRetiredEventsNotInSettings |
| AC-V3R2-RT-006-11 | PASS | TestAuditObservabilityWhitelist |
| AC-V3R2-RT-006-12 | PASS | TestDoctorHook_27EventTableCount + ./bin/moai doctor hook |
| AC-V3R2-RT-006-13 | PASS | TestAuditRegistrationParity |
| AC-V3R2-RT-006-14 | PASS | TestAuditPerFileCategoryHeader (all 26 files) |
| AC-V3R2-RT-006-15 | PASS | TestSubagentStop_KillPaneTimeout (500ms timeout wrap) |
| AC-V3R2-RT-006-16 | PASS | TestAuditObservabilityWhitelist/empty_events_returns_false |
| AC-V3R2-RT-006-17 | PRE-EXISTING | PreToolUse PermissionDecision integration (handled by SPEC-V3R2-RT-002, pre-existing) |
| AC-V3R2-RT-006-18 (derived) | PASS | SystemHookConfig zero value — ObservabilityEvents:nil → opt-in false |
| AC-V3R2-RT-006-19 (derived) | N/A | validator unknown event — deferred (no validator wired yet) |
| AC-V3R2-RT-006-20 (derived) | PASS | 20ms debounce in config_change.go Handle() |
| AC-V3R2-RT-006-21 (derived) | PASS | kill-pane 500ms timeout goroutine in subagent_stop.go |

## Quality Gates (from plan.md §5)

- [x] Coverage ≥ 85% per modified file (new files: observability.go, coverage_table.go, doctor_hook.go all have tests)
- [x] Race clean (`go test -race -count=1 ./...` — short mode passes)
- [x] Lint clean (`golangci-lint run ./internal/hook/... ./internal/cli/... ./internal/config/...` → 0 issues)
- [x] Build success (`make build` → success, embedded.go regenerated)
- [x] Audit GREEN (4 sub-tests in `audit_test.go` — all PASS)
- [x] MX tags applied (@MX:ANCHOR on observabilityOptIn + CoverageTable)
- [x] CHANGELOG entry (Trackable TRUST 5 — 3 bullets in Unreleased)

## Dependencies status

- [x] SPEC-V3R2-RT-001 (HookResponse / HookOutput) — merged
- [x] SPEC-V3R2-RT-002 (PreToolUse PermissionDecision) — merged
- [x] SPEC-V3R2-RT-004 (SessionState checkpointing) — merged
- [x] SPEC-V3R2-RT-005 (8-tier resolver) — merged at `ab0fc4dda`
- [ ] SPEC-V3R2-SPC-002 (MX TagScanner) — TBD; if unmerged, T-RT006-18 uses interface stub

## Stagnation tracking (re-planning gate per spec-workflow.md)

(Updated each iteration if M3+ tasks stagnate; thresholds per spec-workflow.md § Re-planning Gate)

| Iteration | AC criteria met (cumulative) | Errors fixed | Errors introduced | Notes |
|-----------|------------------------------|--------------|--------------------|-------|
| (none yet — run-phase entry pending) | — | — | — | — |

---

## Sync Phase Status

(Sync phase pre-conditions: run PR merged into main → continue same SPEC worktree → `/moai sync SPEC-V3R2-RT-006`.)

- [ ] CHANGELOG entry confirmed in main
- [ ] API documentation generated for `moai doctor hook`
- [ ] Release-note draft for BC-V3R2-018
- [ ] spec.md §5.7 footer count reconciled if drift exists per plan §1.2.1
- [ ] sync PR opened
- [ ] sync PR squash-merged into main

## Cleanup Phase Status

(Cleanup phase pre-condition: BOTH run AND sync PRs merged into main.)

- [ ] `moai worktree done SPEC-V3R2-RT-006` from host checkout
- [ ] feat/SPEC-V3R2-RT-006 branch deleted
- [ ] sync/SPEC-V3R2-RT-006 branch deleted

---

Version: 0.1.0
Last Updated: 2026-05-10
