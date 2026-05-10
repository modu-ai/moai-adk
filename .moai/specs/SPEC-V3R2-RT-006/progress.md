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
| T-RT006-01 | pending | git rm internal/hook/setup.go |
| T-RT006-02 | pending | 4 RED audit sub-tests |
| T-RT006-03 | pending | 6 RED AC-path tests |
| T-RT006-04 | pending | system.yaml hook section schema |

### M2: SystemHookConfig + RT-005 typed loader integration — Priority P0

| Task ID | Status | Notes |
|---------|--------|-------|
| T-RT006-05 | pending | SystemHookConfig struct |
| T-RT006-06 | pending | observabilityOptIn helper |
| T-RT006-07 | pending | gate apply to 4 retire handlers |
| T-RT006-08 | pending | TestAuditObservabilityWhitelist GREEN |

### M3: Settings.json retire + Resolution headers + handler upgrades — Priority P0

| Task ID | Status | Notes |
|---------|--------|-------|
| T-RT006-09 | pending | template settings.json.tmpl retire 4 events |
| T-RT006-10 | pending | local settings.json retire 4 events |
| T-RT006-11 | pending | template context ObservabilityEvents field |
| T-RT006-12 | pending | 22 handler files Resolution: header |
| T-RT006-13 | pending | TestAuditPerFileCategoryHeader GREEN |
| T-RT006-14 | pending | subagent_stop 500ms timeout wrap |
| T-RT006-15 | pending | config_change RT-005 reload integration |
| T-RT006-16 | pending | config_change 20ms debounce |
| T-RT006-17 | pending | TestInstructionsLoaded_42kCLAUDE GREEN |
| T-RT006-18 | pending | file_changed TagScanner stub integration |
| T-RT006-19 | pending | TestAuditRegistrationParity body |
| T-RT006-20 | pending | TestPostToolFailure_TimeoutClassification GREEN |

### M4: doctor hook CLI — Priority P1

| Task ID | Status | Notes |
|---------|--------|-------|
| T-RT006-21 | pending | TestAuditRetiredEventsNotInSettings body |
| T-RT006-22 | pending | TestStrictModeRetiredEvent |
| T-RT006-23 | pending | coverage_table.go shared data |
| T-RT006-24 | pending | doctor_hook.go CLI |
| T-RT006-25 | pending | doctor.go parent wire-up |
| T-RT006-26 | pending | --trace flag readout |
| T-RT006-27 | pending | doctor_hook_test.go |

### M5: Verification gates + audit consolidation — Priority P0

| Task ID | Status | Notes |
|---------|--------|-------|
| T-RT006-28 | pending | go test ./... -race -count=1 |
| T-RT006-29 | pending | golangci-lint run clean |
| T-RT006-30 | pending | make build embedded.go regen |
| T-RT006-31 | pending | manual tmux team mode integration |
| T-RT006-32 | pending | CHANGELOG Unreleased entries |
| T-RT006-33 | pending | @MX tags per plan.md §6 |
| T-RT006-34 | pending | TestAudit* re-verify GREEN |
| T-RT006-35 | pending | doctor hook footer reconcile |

## Acceptance Criteria Status (from acceptance.md)

| AC ID | Status | Notes |
|-------|--------|-------|
| AC-V3R2-RT-006-01 | pending | SubagentStop full pipeline |
| AC-V3R2-RT-006-02 | pending | pane-not-found graceful |
| AC-V3R2-RT-006-03 | pending | Windows no-op |
| AC-V3R2-RT-006-04 | pending | RT-005 reload integration |
| AC-V3R2-RT-006-05 | pending | invalid YAML keeps old |
| AC-V3R2-RT-006-06 | pending | 42k char overage |
| AC-V3R2-RT-006-07 | pending | MX tag delta |
| AC-V3R2-RT-006-08 | pending | timeout classification |
| AC-V3R2-RT-006-09 | pending | setup.go removed |
| AC-V3R2-RT-006-10 | pending | retire-event audit |
| AC-V3R2-RT-006-11 | pending | observability opt-in |
| AC-V3R2-RT-006-12 | pending | doctor hook 27-event |
| AC-V3R2-RT-006-13 | pending | undocumented handler audit |
| AC-V3R2-RT-006-14 | pending | per-file Resolution header |
| AC-V3R2-RT-006-15 | pending | pane teardown ordering |
| AC-V3R2-RT-006-16 | pending | empty observability silent |
| AC-V3R2-RT-006-17 | pending | PreToolUse PermissionDecision |
| AC-V3R2-RT-006-18 (derived) | pending | SystemHookConfig zero value |
| AC-V3R2-RT-006-19 (derived) | pending | validator unknown event reject |
| AC-V3R2-RT-006-20 (derived) | pending | ConfigChange debounce |
| AC-V3R2-RT-006-21 (derived) | pending | kill-pane 500ms timeout |

## Quality Gates (from plan.md §5)

- [ ] Coverage ≥ 85% per modified file (`go test -cover ./internal/hook/ ./internal/cli/ ./internal/config/`)
- [ ] Race clean (`go test -race -count=1 ./...`)
- [ ] Lint clean (`golangci-lint run`)
- [ ] Build success (`make build`)
- [ ] Audit GREEN (4 sub-tests in `audit_test.go`)
- [ ] MX tags applied (`moai mx scan internal/hook/`)
- [ ] CHANGELOG entry (Trackable TRUST 5)

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
