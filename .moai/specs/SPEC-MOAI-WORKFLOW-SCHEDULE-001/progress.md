# Progress — SPEC-MOAI-WORKFLOW-SCHEDULE-001

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-07-17
- tier: M
- artifacts: spec.md, plan.md, acceptance.md, progress.md
- requirements: 24 (REQ-MWS-001 … REQ-MWS-024)
- acceptance_criteria: 21 (AC-MWS-001 … AC-MWS-021) + 5 edge cases
- open_clarifications: 2 (name-collision policy; session-scoped loop re-arm responsibility) — see plan.md §D
- notes: GEARS format; native-scheduler-only; cadence read-only invariant inherited from cadence-bridge.md.

## §E.2 Run-phase Evidence

Milestones (commits): M1-M2 `6ea2fcfbf` (data model + safety boundary), M3-M5,M7 `577f12f4e` (entry surface + router wiring), M6 `e59702241` (template scaffold + presence guard). cycle_type=tdd (RED→GREEN captured for the M6 scaffold guard).

Primary deliverables:
- `.claude/skills/moai/workflows/workflow.md` (+ template mirror) — the MoAI-Workflow skill body.
- `.claude/commands/moai/workflow.md` (+ template `.md.tmpl`) — `/moai workflow` thin wrapper.
- `.claude/skills/moai/SKILL.md` (+ template mirror) — router registration (Priority-1 match + subcommand definition + description list).
- `internal/template/templates/.moai/workflows/{README.md,example.md}` (+ local `.moai/workflows/` mirror) — template scaffold.
- `internal/template/moai_workflows_scaffold_test.go` — presence + neutrality guard.

### AC PASS/FAIL matrix (21 AC — SSOT acceptance.md)

| AC | Status | Evidence |
|----|--------|----------|
| AC-MWS-001 | PASS | workflow.md §Frontmatter Schema + example.md declares name/description/schedule/safety + non-empty body; `TestMoaiWorkflowsScaffoldPresence` asserts frontmatter fields + non-empty body |
| AC-MWS-002 | PASS | workflow.md §schedule mapping (`expression` + `mechanism` ∈ {cron,loop}) |
| AC-MWS-003 | PASS | workflow.md §safety field (`read-only`\|`write`; absent → `read-only` default) |
| AC-MWS-004 | PASS | workflow.md §Creation — Guided Capture (AskUserQuestion rounds capture name/description/schedule/safety/steps → write `.moai/workflows/<name>.md`); router registration |
| AC-MWS-005 | PASS | workflow.md §Validation (invalid `mechanism` e.g. `daemon` / malformed expression → file NOT written, failure surfaced) |
| AC-MWS-006 | PASS | workflow.md §Entry surface (single creation code-path; NL capture routes to guided path); SKILL.md Priority-1 `workflow` entry |
| AC-MWS-007 | PASS | workflow.md §mechanism: cron (CronCreate + interval + mechanism vocabulary) |
| AC-MWS-008 | PASS | workflow.md §mechanism: loop (session-scoped disclosure + record-only + session-start re-arm reminder, no auto-arm — DECISION-MWS-D2) |
| AC-MWS-009 | PASS | workflow.md §Unregistration (CronDelete before delete, notice names mechanism; loop → session-scoped cancellation notice) |
| AC-MWS-010 | PASS | workflow.md §Cron-unavailable fallback (degrade to `/loop` form) |
| AC-MWS-011 (must-pass) | PASS | workflow.md §Safety Boundary HARD invariant (no commit/push/run-phase for any scheduled run) |
| AC-MWS-012 | PASS | workflow.md §Safety Boundary "Level-1 edits only, left uncommitted" |
| AC-MWS-013 (must-pass) | PASS | workflow.md §`safety: write` governs interactive invocation only (does not relax cadence invariant) |
| AC-MWS-014 | PASS | workflow.md §Human gates are cadence-unsatisfiable (queue + surface next interactive session) |
| AC-MWS-015 | PASS | workflow.md §list (name+description+schedule; schedule-less renders without detail, no error) |
| AC-MWS-016 | PASS | workflow.md §edit (open markdown file; frontmatter is SSOT) |
| AC-MWS-017 | PASS | workflow.md §remove (no-op on absent; existing → unregister then delete) |
| AC-MWS-018 | PASS | `TestMoaiWorkflowsScaffoldPresence` (README.md + exactly one neutral example; README explains schedule/mechanism/safety) |
| AC-MWS-019 (must-pass) | PASS | `TestMoaiWorkflowsScaffoldNeutrality` + whole-tree `internal_content_leak_test` + manual grep (no SPEC-ID/date/SHA) |
| AC-MWS-020 | PASS | Template-First: source under `internal/template/templates/` edited first, embedded via `make build` (catalog.yaml regen), local mirror byte-identical |
| AC-MWS-021 | PASS | workflow.md §Boundary — Non-Duplication (4 sibling assets; vocabulary reused not forked) + spec.md §C |

Must-pass ACs (AC-MWS-011, 013, 019): all PASS.

### Verification commands (verbatim outputs redirected)

- `go test ./...` → exit 1, but the ONLY failure is `internal/cli TestRunHarnessObserveStop_AutoClassifyChain` — a PRE-EXISTING, unrelated harness-learning-subsystem failure (confirmed failing at pre-work baseline `56ef3fb9a` in a clean worktree). My changed scope (`internal/template`) is green: `ok internal/template 1.925s`. Log: `.moai/state/verify/SPEC-MOAI-WORKFLOW-SCHEDULE-001/1-gotest.log`.
- `go test ./internal/template/ -run TestMoaiWorkflowsScaffold` → PASS (RED captured before scaffold creation; GREEN after).
- `go vet ./internal/template/` → exit 0.
- `golangci-lint run ./internal/template/... --timeout=3m` → 0 issues.
- `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- `make build` → exit 0 (catalog.yaml hash regen for the `moai` skill).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-17
run_commit_sha: 96d498ac6f99f0f3500cfe0f4d4518f1ef9d7a10
run_status: audit-ready
ac_pass_count: 21
ac_fail_count: 0
must_pass_ac: [AC-MWS-011, AC-MWS-013, AC-MWS-019]  # all PASS
preserve_list_post_run_count: 0  # no PRESERVE-list files modified beyond declared scope
l44_pre_commit_fetch: not-applicable  # Route A main-direct discipline; orchestrator owns push/fetch
l44_post_push_fetch: not-applicable  # manager-develop does not push (orchestrator pushes)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  linux_amd64: pass
  windows_amd64: pass
total_run_phase_files: 8  # workflow.md(local+template), workflow.md(cmd local)+.md.tmpl, SKILL.md(local+template), scaffold README+example(template+local x2)=4, scaffold test, catalog.yaml, spec.md, progress.md
m1_to_mN_commit_strategy: per-milestone-group (M1-M2 / M3-M5,M7 / M6); not pushed — orchestrator pushes
pre_existing_unrelated_failure: internal/cli TestRunHarnessObserveStop_AutoClassifyChain (baseline 56ef3fb9a also fails; out of scope)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-17
sync_commit_sha: 469d5023c
sync_status: audit-ready
changelog_entry_position: "[Unreleased] > ### Added (SPEC-MOAI-WORKFLOW-SCHEDULE-001 entry, inserted before ### Fixed)"
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed (merged 3-phase close, this commit)"
  updated_field_refreshed: 2026-07-17
b12_self_test_a: "grep -c 'SPEC-MOAI-WORKFLOW-SCHEDULE-001' CHANGELOG.md (pre-emission) == 0 -> emission proceeded"
b12_self_test_b: "acceptance.md SSOT AC count 21 matches CHANGELOG entry claim (21/21 AC PASS)"
b12_self_test_c: "ls verified: .claude/skills/moai/workflows/workflow.md, .claude/commands/moai/workflow.md, internal/template/templates/.claude/commands/moai/workflow.md.tmpl, internal/template/templates/.moai/workflows/{README.md,example.md}, internal/template/moai_workflows_scaffold_test.go — all present"
canary_compliance_check:
  spec_body_untouched: true   # spec.md/plan.md/acceptance.md body content NOT modified; only frontmatter status+updated
route: "A (Hybrid Trunk main-direct, no PR)"
```

## §F Phase 4 Mode Selection

- Inputs: tier=M, scope≈10-15 files (skill workflow md + template scaffold + docs), domains=2 (workflow skills, template), language mix=markdown-heavy + minimal Go test guard, concurrency benefit=LOW (coding/authoring-heavy)
- Evaluation: trivial=not selected (multi-file feature) / background=not selected (write work) / agent-team=RETIRED / parallel=not selected (coding-heavy, <3 domains) / workflow=not selected (<30 files, non-mechanical) / sub-agent=SELECTED
- Decision: sub-agent
- Justification: Coding/authoring-heavy Tier M work with sequential milestone dependencies; per Anthropic's coding-task parallelism caveat, Mode 5 sequential sub-agent is the default and correct envelope. Implementation Kickoff Approval passed 2026-07-17 (AskUserQuestion, run-phase 진입 selected); all preferences collected (engine/format/safety + D1/D2 decisions).
