# Progress — SPEC-WEB-CONSOLE-003

> S2a of the `web-console-v3` cohort. Tier M. FLAT/SHALLOW project-config parity (development_mode + git_convention.convention).

## Plan-phase signal

plan_complete_at: 2026-06-03
plan_status: audit-ready

## Tier & artifacts

- Tier: M (3 artifacts: spec.md + plan.md + acceptance.md)
- Justification: new project-config persistence path crossing internal/web → internal/config boundary; ~6-7 files; dual-editor (web + TUI) × validate/widget/persist/4-locale/read-on-render → 11 ACs.

## Scope decisions recorded

- Confirmed flat settings: `development_mode` ({ddd,tdd}, exported predicate `models.ValidDevelopmentModes()`), `git_convention.convention` ({auto,conventional-commits,angular,karma,custom}, `pkg/models` oneof SSOT; new exported `IsValidConvention` to be added in M1).
- NARROWED OUT: `llm.mode` (backend-switch toggle, only `""|glm`), `llm.default_model` (legacy enum-less string, no `validate` tag, no canonical enum to reuse). Recorded in spec.md §1.
- KEY design: project-config persistence via config-manager `LoadRaw`→`SetSection`→`Save` (new `app` seams), NOT `ProfilePreferences` (which has no slot for these). Bounded `SyncToProjectConfig`-pattern extension.

## REQ / AC counts

- REQs: 8 (REQ-WC3-001 .. REQ-WC3-008)
- ACs: 11 + closure (AC-WC3-001a/001b/002a/002b/003/004/005/006a/006b/007/008/009) + 6 edge cases (EC-1..EC-6)

## Next

- Phase 0.5 plan-auditor (Tier M PASS threshold 0.80) → GATE-2 → /moai run (cycle_type=tdd).

---

## §E — Phase 0.95 Mode Selection

**Input parameters**
- tier: M
- scope (file count): ~7 (internal/config validation.go + internal/web app.go/validate.go/handlers.go/projectconfig.go/page.html.tmpl + internal/cli profile_setup.go/profile_setup_translations.go)
- domain count: 2 (Go source: internal/web + internal/cli + internal/config; SPEC artifacts)
- file language mix: 100% Go (+ one html/template + SPEC markdown)
- concurrency benefit: LOW (coding-heavy, sequential milestone dependency M1→M2→M3→M4→M5)
- Agent Teams prereqs status: not evaluated (single-agent run-phase delegation)

**Mode evaluation table**

| Mode | Selected? | Rationale |
|------|-----------|-----------|
| 1 trivial | not selected | multi-file semantic feature, not a typo |
| 2 background | not selected | write-heavy (Write/Edit), cannot background |
| 3 agent-team | not selected | single-agent delegation; not multi-domain research-heavy |
| 4 parallel | not selected | coding-heavy (Finding A4 caveat); milestones have sequential deps |
| 5 sub-agent | **selected** | coding-heavy run-phase, single manager-develop, sequential M1→M5 |
| 6 workflow | not selected | < 30 files, not a uniform mechanical transform |

**Decision: sub-agent**

**Justification**: Run-phase implementation of a coding-heavy Tier M feature with strictly sequential milestone dependencies (M1 predicate → M2 seams reuse it → M3 handler wires M2 → M4 reuses M2 write path → M5 verifies). Per Finding A4 (coding tasks involve fewer truly parallelizable subtasks than research), Mode 5 (single sequential sub-agent) is the correct default — orchestrator-decided.

**Phase 0.5 SKIP rationale**: plan-auditor verdict PASS 0.91 ≥ 0.90 AND no plan-PR commit landed since that verdict → Phase 0.5 re-execution skipped per spec-workflow.md Plan Audit Gate skip policy (CONST-V3R5-026). GATE-2 (plan-to-implement HUMAN GATE) was approved by the user independently of the 0.91 score per CLAUDE.local.md §19.1.

## §E.2 Run-phase Evidence

| AC | REQ | Status | Verification command | Actual output |
|----|-----|--------|----------------------|---------------|
| AC-WC3-001a | REQ-WC3-001 | PASS | `go test ./internal/web/ -run 'TestSaveRejectsBogusDevelopmentMode\|TestValidateProjectConfig'` | ok internal/web (bogus dev → 400 + field error + no write) |
| AC-WC3-001b | REQ-WC3-001/005 | PASS | `go test ./internal/web/ -run TestSaveValidProjectConfigPersists` | ok (ddd → 200, write seam receives ddd) |
| AC-WC3-002a | REQ-WC3-002 | PASS | `go test ./internal/web/ -run TestSaveRejectsBogusConvention` + `go test ./internal/config/ -run TestIsValidConvention` | ok both (gitflow → 400 + git_convention error + no write; IsValidConvention 100% cov) |
| AC-WC3-002b | REQ-WC3-002/005 | PASS | `go test ./internal/web/ -run TestSaveValidProjectConfigPersists` | ok (angular → 200, write seam receives angular) |
| AC-WC3-003 | REQ-WC3-003 | PASS | `go test ./internal/web/ -run TestProjectFieldsetRendersSelects` | ok (`<legend>Project</legend>` + 2 `<select>`, no type=text, 7 canonical opts + 2× "(project default)") |
| AC-WC3-004 | REQ-WC3-004 | PASS | `go test ./internal/web/ -run 'TestProjectSelectsPreselectCurrentValues\|TestProjectReadSeamFailureRendersInlineError'` | ok (ddd/karma pre-selected; read-fail → ≥400 inline error, no panic) |
| AC-WC3-005 | REQ-WC3-005/007 | PASS | `go test ./internal/web/ -run 'TestDefaultAppHasProjectConfigSeams\|TestWriteProjectConfig\|TestSaveEmptyProjectConfigPasses'` + no-marshal grep | ok (2 seams non-nil; empty keeps existing; `grep yaml.Marshal\|os.WriteFile internal/web/*.go` non-test = 0) |
| AC-WC3-006a | REQ-WC3-006 | PASS | `go test ./internal/cli/ -run 'TestGetProfileText_AllLanguages\|TestProfileSetupConstructsProjectSelects'` | ok (7 labels × 4 locales non-empty; both selects bound + 7 canonical opt values present) |
| AC-WC3-006b | REQ-WC3-006 | PASS | `go test ./internal/cli/ -run TestPersistProjectConfig` | ok (quality.yaml=ddd + git-convention.yaml=angular; preferences.yaml has neither key) |
| AC-WC3-007 | REQ-WC3-007 | PASS | `go test ./internal/web/ -run TestWriteProjectConfigSectionIsolation` | ok (workflow/harness/git-strategy byte-identical; llm no backend switch; test_coverage_target=85 preserved) |
| AC-WC3-008 | REQ-WC3-008 | PASS | `go test ./internal/web/ -run 'TestGoldenPath\|TestHostCheck\|TestSaveScopeBoundary'` | ok (loopback/no-auth/Host-check/DO_NOT_TOUCH sentinels green; integration_test.go unmodified) |
| AC-WC3-009 | all | PASS | `go test ./internal/web/... ./internal/cli/... ./internal/config/...` | exit 0 (all packages ok) |

Edge cases: EC-1 (both empty → no clobber) PASS; EC-2 (one bogus, one valid → atomic reject, neither persisted) PASS `TestSaveEC2AtomicReject`; EC-3 (`custom` enum accepted) PASS `TestValidateProjectConfig`; EC-4 (uppercase TDD/Angular non-canonical) PASS; EC-5 (absent config → LoadRaw defaults, no panic) PASS `TestReadProjectConfig_AbsentFiles` + `absent config creates section`; EC-6 (whitespace value non-canonical) PASS.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-03
run_commit_sha: e83864047
run_status: implemented
ac_pass_count: 12
ac_fail_count: 0
preserve_list_post_run_count: 5
new_warnings_or_lints_introduced: 0
cross_platform_build:
  native: "go build ./... exit 0"
  windows: "GOOS=windows GOARCH=amd64 go build ./... exit 0"
coverage:
  internal_web: "90.9%"
  internal_cli: "71.5% (package-level; new SPEC fns readCurrentProjectConfig 80%, persistProjectConfig 81%)"
  internal_config: "77.7% (package-level; new SPEC fns IsValidConvention 100%, ValidConventions 100%)"
  new_spec_code: "M1 predicates 100%, validateProjectConfig 100%, projectView 100%, developmentModesFromModels 100%, read/write seams 80-81% (error branches need fault injection)"
subagent_boundary_c_hra_008: "0 matches (AskUserQuestion in new internal/web + internal/cli files)"
no_direct_marshal_guard: "0 matches (yaml.Marshal/os.WriteFile of config section in internal/web non-test)"
status_transition: "draft → in-progress (M1 commit 3294eb5b3, manager-develop); run-phase complete (implemented) on M-final"
m1_to_mN_commit_strategy: "M1 3294eb5b3 (predicate + frontmatter transition) → M2 d3b40997b (validator+seams) → M3 162889960 (handler+widget) → M4 e861790ce (TUI parity) → M5 (scope test + progress)"
total_run_phase_files: 15
integration_test_do_not_touch: "unmodified (last touched by SPEC-WEB-CONSOLE-001 b1ab60454)"
```

## §E.4 Notes for sync-phase (manager-docs)

- Implementation matches plan.md §F M1-M5 exactly. No scope expansion. `llm.mode`/`llm.default_model` NARROWED OUT (not implemented, per spec.md §1).
- No template mirroring / `make build` (B-NEW): web assets embed via `internal/web/assets.go` `go:embed`, NOT the template-deploy system.
- Commits NOT pushed by manager-develop (B9 exception (a): active parallel session race present — orchestrator pushes after Trust-but-verify).
- L1 worktree: this run executed in isolated worktree `worktree-agent-aa4e5e1f0817715c5` (branch == origin/main base a090f29ac). Orchestrator integrates via cherry-pick of the 4-5 SPEC commits.

## Sync-phase Audit-Ready Signal

```yaml
sync_commit_sha: d2db9701d
sync_status: complete
changelog_entry_added: true
status_transition: "in-progress → implemented"
version_bump: "0.1.0 → 0.2.0"
sync_executor: orchestrator-direct
sync_rationale: >
  Tier M bounded close-tail (run M1-M5 already on origin/main via cherry-pick
  e83864047). Active parallel sessions (GO-DEPS / WEB-CONSOLE-004) → orchestrator-direct
  sync/Mx to avoid manager-docs L1-worktree + race overhead. Authored-By-Agent trailer
  omitted (legacy silent SKIP) to avoid OwnershipTransitionInvalid on the
  orchestrator-direct in-progress→implemented transition.
```

## §E.5 Mx-phase Audit-Ready Signal

```yaml
mx_commit_sha: (this commit — backfilled by follow-up)
status_transition: "implemented → completed"
four_phase_close: true
close_subject_full_id: SPEC-WEB-CONSOLE-003
mx_executor: orchestrator-direct
audit_ready: true
notes: >
  4-phase close (plan a090f29ac / run 7a5b1698c..e83864047 / sync d2db9701d / Mx this).
  Era H-4 (§E.2 + §E.5 + sync_commit_sha + mx_commit_sha) → V3R6 modern, drift-aligned.
```
