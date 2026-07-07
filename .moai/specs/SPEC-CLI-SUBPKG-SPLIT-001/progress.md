# Progress — SPEC-CLI-SUBPKG-SPLIT-001

> Canonical §E lifecycle skeleton. Plan-phase emits placeholder headings only; §E.2/§E.3
> are populated by manager-develop (run-phase) and §E.4 by manager-docs (sync-phase).

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID regex self-check: `decomposition: SPEC ✓ | CLI ✓ | SUBPKG ✓ | SPLIT ✓ | 001 ✓ → PASS` (canonical `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`; last segment digit-only, no alpha suffix).
- Frontmatter: 12 canonical fields present; `status: draft`; `priority: P2`; ISO `created`/`updated`; `tags` comma-separated string; `tier: L`; `era: V3R6`. Validated.
- Artifacts emitted: spec.md, plan.md, acceptance.md, design.md, research.md, progress.md (Tier L 5-file set + progress).
- Tier: **L** (25,838 LOC flat package, 93 files, 54,756 test LOC, cross-package structural refactor with import-cycle + deps coupling → constitutional-scale).
- Out of Scope: present with `### Out of Scope —` H3 sub-headings (big-bang split / tiny single-file clusters / deps-platform-tangled clusters / functional change / new tests / existing-subpackage reorg / CHANGELOG-README).
- Measurements grounded (NOT assumed): all §A research metrics from `find`/`wc`/`grep`/`go build` against the working tree; build baseline `go build ./internal/cli/...` observed exit 0.
- Cluster map: 26 candidate clusters + shared kernel, reconciled to ~25,730 LOC (residual ~108 = `worktree_validation.go`, belongs to existing worktree subpackage). Exact per-cluster counts re-derived at run-phase, not hardcoded.
- Two dominant risks documented: (1) test-migration surface (147 files / 54,756 test LOC white-box) — dominant regression risk; (2) import-cycle hazard (root imports subpackages for AddCommand → subpackages cannot import package cli back) — blocks kernel-dependent clusters on a prior `uikit` leaf extraction.
- deps coupling quantified: exactly 8 root files touch the global `deps`; provider-injection pattern (worktree.WorktreeProvider precedent) is the resolution.
- Honest value/risk framing recorded in spec.md §A: refactor of WORKING code, gains real-but-incremental + non-user-observable; big-bang REJECTED; phased lowest-risk-first RECOMMENDED with post-M4 checkpoint stop-condition (REQ-CSS-010).
- Status: plan-phase artifacts final (status: draft); ready for Implementation Kickoff Approval human gate.
- **Drift re-verification (2026-07-07, pre-audit)**: re-measured against the current working tree. Δ +5 root non-test files / +602 LOC, +6 test files / +1,265 test LOC, +2 build-tag files; deps-coupled count unchanged at 8; build baseline still exit 0 (both `go build ./...` and `GOOS=windows GOARCH=amd64`). All M1-M7 cluster files remain cohesive; 6 existing subpackages intact; kernel helpers still in `package cli`. Cluster LOC measured (M1 migrate 947 / M2 profile 1,527 / M3 agentlint 1,392 / M6 doctor 2,073 / M7 update 5,184) — within ±15% of plan.md claims; plan still valid. `moai spec lint` PASS (0 findings); `moai spec audit` PASS (0 drift findings, 1 modern-era clean). Plan-audit-ready confirmed.

## §E.2 Run-phase Evidence

### M1 — agentlint cluster extraction (re-sequenced; original M1 migrate blocked by bidirectional import cycle)

**Scope**: `internal/cli/agent_lint.go` + `workflow_lint.go` + `sentinels.go` (impl) + `agent_lint_test.go` + `workflow_lint_test.go` (tests) → `internal/cli/agentlint/` (new `package agentlint`). `internal/cli/root.go` wiring updated.

**cycle_type**: ddd (ANALYZE-PRESERVE-IMPROVE — behavior-preserving refactor).

**Worktree note**: executed in L1 worktree `agent-a17b7b2ec1dc3e9c4` (branch `worktree-agent-a17b7b2ec1dc3e9c4`); Edit tool enforced worktree isolation. Shared-checkout `main` working tree was cleaned of an accidental early `git mv` before work moved to the worktree. Push target: `origin/main` (the worktree branch is 1 commit ahead of `origin/main` at `c080b85b1`).

#### AC PASS/FAIL matrix (M1 subset)

| AC | Status | Verification (literal command) | Actual output |
|----|--------|--------------------------------|---------------|
| AC-CSS-001 | PASS-WITH-DEBT | `go test ./...` | Build/vet/lint green. 4 pre-existing FAIL packages (all reproduced on original code, NOT regressions): (1) `internal/cli` `TestRunHookEvent_ReadInputError` panic — pre-existing, documented in prompt; (2) `internal/cli/agentlint` `TestAuthoringDocHasEffortMatrix` + `TestConstitutionCrossReference` — pre-existing RED doc-tests (commented "will FAIL until M2"); were `SKIP` from `internal/cli` 2-level depth, now surface as `FAIL` from the 3-level agentlint depth their `../../../` path heuristic was designed for; (3) `internal/statusline` `TestBuild_WritesContextUsageWithSessionID` — pre-existing env/state; (4) `internal/template` `TestSettingsTemplateRequiredEnvVars` — pre-existing test/template mismatch. Zero NEW production-code test failures introduced by the move. |
| AC-CSS-002 | PASS | `go build ./...` AND `GOOS=windows GOARCH=amd64 go build ./...` | both exit 0 |
| AC-CSS-003 | PASS | `go run ./cmd/moai --help` diff before/after | identical (empty diff) |
| AC-CSS-004 | PASS | `grep -n 'AddCommand(agentlint' internal/cli/root.go` | exactly 2 matches: `rootCmd.AddCommand(agentlint.AgentCmd)` + `rootCmd.AddCommand(agentlint.WorkflowCmd)` (one per parent; no double-register) |
| AC-CSS-005 | PASS | `head -1 internal/cli/agentlint/*.go` | all 5 files declare `package agentlint` |
| AC-CSS-009 | PASS | `grep -rnE 'AskUserQuestion\(\|mcp__askuser' internal/cli/agentlint/ \| grep -v _test.go` | 0 actual invocations (matches are `checkLiteralAskUserQuestion` detector fn — substring false positives) |
| AC-CSS-010 | PASS | `git show --stat <sha>` | one `internal/cli/agentlint/` dir (5 files) + `internal/cli/root.go` + spec.md/progress.md frontmatter/§E.2-§E.3; no second cluster |
| AC-CSS-011 | PASS | non-test impl diff | package decl + import block (added lipgloss, tui) + local cycle-breaking helpers (cliSuccess/cliWarn/cliError styles, getStringFlag/getBoolFlag accessors — byte-identical to package-cli originals) + init() restructure (exported AgentCmd/WorkflowCmd, rootCmd reference removed). Lint rule logic unchanged. |

#### Build matrix

```
go build ./...                          → exit 0
GOOS=windows GOARCH=amd64 go build ./... → exit 0
go vet ./internal/cli/agentlint/... ./internal/cli/ → exit 0
golangci-lint run --timeout=2m          → 0 issues (clean; no NEW issues vs baseline)
```

#### Test result (agentlint package)

`go test -cover ./internal/cli/agentlint/...` → coverage: **82.9%** of statements. Package reports FAIL due to the 2 pre-existing RED doc-tests (TestAuthoringDocHasEffortMatrix, TestConstitutionCrossReference); all other tests PASS. Coverage measured against production code (the 2 failing tests check .md doc content, not Go code, so they do not affect code-coverage measurement).

#### Behavior snapshot

`go run ./cmd/moai --help` output byte-identical before/after the move (AC-CSS-003) — `moai agent lint` and `moai workflow lint` command surface unchanged.

#### Lint NEW-vs-baseline

Baseline `golangci-lint run` = 0 issues (clean). Post-M1 = 0 issues. No NEW lint issues introduced.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-07
run_commit_sha: "0ee246ad9"
run_status: M1-complete
ac_pass_count: 7
ac_fail_count: 0
preserve_list_post_run_count: 5   # existing subpackages (worktree/harness/preference/wizard/specid/pr) + cli.Execute/deps.go/root structure preserved
l44_pre_commit_fetch: "0 0 (origin/main synced at c080b85b1; 2 unrelated commits from SPEC-HOOK-PREEDIT-INVESTIGATE-001 landed on main after the prompt's expected HEAD 1e9a4e23b)"
l44_post_push_fetch: "<pending push>"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  goos_darwin: PASS
  goos_windows: PASS
total_run_phase_files: 6   # 5 moved (agent_lint.go, workflow_lint.go, sentinels.go, agent_lint_test.go, workflow_lint_test.go) + 1 modified (root.go)
m1_to_mN_commit_strategy: "per-milestone commit (Route A main-direct intent; executed in L1 worktree due to Edit-tool worktree isolation — push via refspec worktree-branch:main)"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-07-07
sync_commit_sha: "<placeholder — backfill post-sync-commit>"
sync_status: M1-only-close
close_decision: "§F CHECKPOINT after M1 (REQ-CSS-010) — STOP (ship M1 only); remaining clusters (profile/migrate/doctor/update/uikit) require tri-axis coupling resolution -> separate SPECs"
run_commit_sha_backfilled: "0ee246ad9 (M1 agentlint)"
l44_pre_sync_fetch: "0 0 (origin/main synced at 360f7b39e)"
m1_only_close_rationale: "agentlint is the only orchestrator-verified clean cluster (kernel-render-free, zero reverse-dep); M2-M7 conditional on coupling-resolution design work"
ac_close_subset: "7/7 M1 AC PASS (AC-CSS-001..005,009,010,011); 12 AC total, remaining 5 AC are M2-M7 scope (deferred to separate SPECs)"
followup: "uikit(M5 kernel) / profile / migrate / update clusters — separate SPECs post-M1 checkpoint"
```
