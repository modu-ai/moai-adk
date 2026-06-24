# Progress — SPEC-GITSTRATEGY-SAVE-ISOLATION-001

> Tier S regression fix (TDD, reproduction-first). GitHub Issue #1064. The reproduction tests were already RED on the run base (`aeea93872`); the fix flips them GREEN without weakening either test.

## §E.1 Run-phase Milestones

| Milestone | Description | Status |
|-----------|-------------|--------|
| M1 | Add `gitStrategyDirty` tracking to ConfigManager (set in SetSection, reset in Load/LoadRaw/Reload + after successful Save) | DONE |
| M2 | Guard `git-strategy.yaml` write in Save() — write only when dirty OR file absent | DONE |
| M3 | Confirm both reproductions flip RED→GREEN; SAVE-WIRING stays GREEN | DONE |
| M4 | No caller regression (config/web/cli/profile), race-clean, vet-clean, lint-clean, cross-platform build | DONE |
| M5 | EC-3 dirty-flag-reset unit test (D2 debt) + Reload-resets-dirty assertion (D4) | DONE |

## §E.2 Run-phase Evidence

### AC matrix (acceptance.md SSOT)

| AC | Status | Verification Command | Actual Output |
|----|--------|---------------------|---------------|
| AC-GSI-001 (PRIMARY: project-config write does not touch git_strategy) | PASS | `go test ./internal/web/ -run TestWriteProjectConfigSectionIsolation` | `--- PASS: TestWriteProjectConfigSectionIsolation (0.00s)` — was RED before fix |
| AC-GSI-002 (no-regression: set→save→reload round-trip) | PASS | `go test ./internal/config/ -run TestConfigManagerSaveGitStrategyRoundTrip` | `--- PASS: TestConfigManagerSaveGitStrategyRoundTrip` — stays GREEN |
| AC-GSI-003 (no-regression: greenfield Save creates git-strategy.yaml) | PASS | `go test ./internal/config/ -run TestConfigManagerSaveCreatesGitStrategyFile` | `--- PASS: TestConfigManagerSaveCreatesGitStrategyFile` — stays GREEN |
| AC-GSI-004 (no caller regression across all Save() consumers) | PASS | `go test ./internal/config/... ./internal/web/... ./internal/cli/... ./internal/profile/...` | all `ok` (config/web/cli/cli·harness/cli·pr/cli·wizard/cli·worktree/profile) |
| AC-GSI-005 (guard test integrity — not weakened) | PASS | `git diff --stat internal/web/projectconfig_scope_test.go internal/web/integration_test.go` | empty (zero changes to both reproduction tests) |

### Edge-case evidence

| Edge case | Status | Evidence |
|-----------|--------|----------|
| EC-2 (existing file, no git_strategy modification → byte-unchanged) | PASS | covered by AC-GSI-001 generalization + `TestConfigManagerSaveDirtyFlagReset` (sentinel byte-equality) |
| EC-3 (dirty flag reset post-Save: 2nd Save w/o SetSection does NOT rewrite) | PASS | `go test ./internal/config/ -run TestConfigManagerSaveDirtyFlagReset` → `--- PASS` (NEW test, D2 debt resolved) |
| EC-3/D4 (Reload resets dirty flag) | PASS | `go test ./internal/config/ -run TestConfigManagerReloadResetsDirtyFlag` → `--- PASS` (NEW test, D4 folded in) |
| EC-4 (concurrent SetSection/Save under m.mu) | PASS | `go test -race ./internal/config/...` → `ok ... 1.649s` (flag mutated only under existing m.mu lock) |

### Reproduction-first proof (CLAUDE.md §7 Rule 4)

| Test | Before fix (base aeea93872) | After fix |
|------|-----------------------------|-----------|
| `TestWriteProjectConfigSectionIsolation` (web, PRIMARY) | FAIL — sentinel `git_strategy:\n  sentinel: DO_NOT_TOUCH\n` expanded into full compiled-default tree (mode/provider/manual/personal/team/hooks/...) | PASS — sentinel byte-preserved |
| `TestGoldenPath_ReadWriteRoundTrip` (web, 2nd reproduction) | FAIL — same root cause (default-tree expansion) | PASS |

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-06-13
run_commit_sha: "f8872e2be0396f8b5dd36dc0931aed39e18bf304"
run_status: green
ac_pass_count: 5
ac_fail_count: 0
edge_case_pass_count: 4
new_warnings_or_lints_introduced: 0
guard_tests_byte_unchanged: true   # projectconfig_scope_test.go + integration_test.go: git diff empty
cross_platform_build:
  host: pass        # go build ./...
  windows: pass     # GOOS=windows GOARCH=amd64 go build ./internal/config/... ./internal/web/...
coverage_internal_config: "78.0%"   # no decrease vs ~77.9% pre-fix baseline
total_run_phase_files: 2            # internal/config/manager.go + internal/config/manager_save_git_strategy_test.go
m1_to_mN_commit_strategy: single M1 commit (Tier S, bounded fix)
exclusions_honored:
  EX-1_no_rearchitect: true         # gitStrategyFileWrapper type unchanged
  EX-2_loader_untouched: true       # loadGitStrategySection not modified
  EX-3_schema_validators_defaults_templates_untouched: true
  EX-4_no_sec_harden_file: true     # diff = internal/config only
  EX-5_git_strategy_scoped: true    # dirty flag is git_strategy-specific, not a generic Save() redesign
  EX-6_callers_byte_unchanged: true # web ×2, profile_setup, profile/sync call sites unchanged
```

### Full-suite note

`go test ./...` exits 0. The first verbose run surfaced 2 transient FAILs in `internal/hook` (`TestHookWrapper_ValidJSON`, `TestHookWrapper_MoaiBinaryFallback`, both `signal: killed` 5s subprocess timeout under parallel load) — these are pre-existing flaky tests, contain ZERO references to config/git_strategy, are outside this SPEC's scope, and PASS when run in isolation (`ok ... 0.631s`). Not a regression from this change.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-06-13
sync_commit_sha: "f046287c4"
sync_status: green
status_transition: in-progress → implemented
changelog_entry: added           # CHANGELOG.md [Unreleased] § Fixed
plan_audit_verdict: PASS-WITH-DEBT 0.86   # Tier S threshold 0.80
```

### (Migrated from §E.5)

```yaml
mx_complete_at: 2026-06-13
mx_commit_sha: "c45e18f98"
status_transition: implemented → completed
four_phase_close: true            # plan + run + sync + Mx
github_issue: 1064                 # closed by this fix
```
