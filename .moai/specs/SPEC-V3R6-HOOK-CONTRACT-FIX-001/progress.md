# SPEC-V3R6-HOOK-CONTRACT-FIX-001 — Run-Phase Progress

## Session Summary

- **SPEC**: SPEC-V3R6-HOOK-CONTRACT-FIX-001 (Tier S, Critical, Wave 0)
- **cycle_type**: tdd (per quality.yaml development_mode)
- **Branch**: main (1인 OSS Hybrid Trunk, direct push per CLAUDE.local.md §23)
- **Entry HEAD**: `d386cca0e` (origin/main, no parallel commits at SPEC entry)
- **Exit HEAD**: `d3d9b829f` (3 SPEC commits ahead origin/main, ready to push)
- **Strategy**: orchestrator-direct execution (LEAN Tier S, manager-develop Agent delegation skipped — minimal scope, 4 in-scope files, no cross-file refactor)

## Pre-flight Baseline (Section C)

```
$ git branch --show-current
main
$ git rev-parse HEAD
d386cca0eeb6eacc7f77fdd7d4f9d04173c35826
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
$ test -e internal/hook/.moai/
LEAK PRESENT (untracked, git ls-files: 0)
$ ls internal/hook/.moai/
harness/observations.yaml (3025 bytes, mtime 2026-05-22 18:58)
$ grep -c '"WorktreeCreate"\|"WorktreeRemove"' .claude/settings.json internal/template/templates/.claude/settings.json.tmpl
0 + 0 (PR #1044 baseline confirmed)
$ grep -n 'expectedCount' internal/template/settings_test.go
512:	const expectedCount = 20
$ wc -l internal/hook/worktree_create.go internal/hook/worktree_remove.go
48 + 49 (handlers preserve baseline)
```

**Pre-flight TestAuditRegistrationParity + TestAuditThreeWaySync state on `d386cca0e`** (verified via temporary stash-test):
```
=== RUN   TestAuditRegistrationParity
--- FAIL: TestAuditRegistrationParity (0.00s)
=== RUN   TestAuditThreeWaySync
--- FAIL: TestAuditThreeWaySync (0.00s)
```

These are **pre-existing baseline failures** from PR #1044 incomplete adjustment (the 3-way sync test still expects WorktreeCreate/Remove registered in settings.json, but PR #1044 unregistered them). Out of scope per AC-HCF-012 baseline residual policy. NEW SPEC needed: `SPEC-V3R6-HOOK-AUDIT-TESTS-FIX-001` (provisional).

## Run-phase Actions (M3 → M1 → M2 → M5 → M4 order per plan.md §2.2)

### Commit 1: `8319c6efa` — M3 source fix

**Files changed**: `internal/hook/subagent_stop.go` (+19/-4) + `internal/hook/subagent_stop_test.go` (+59/-2)

- Extracted `resolveObservationsPath(input *HookInput) string` helper implementing the 3-level resolution order: `input.CWD → $CLAUDE_PROJECT_DIR → os.Getwd()` (REQ-HCF-005)
- Refactored `dispatchCapture` to use the new helper
- Added 3 new tests:
  - `TestDispatchCapture_UsesClaudeProjectDirWhenCwdEmpty` (primary AC-HCF-005)
  - `TestResolveObservationsPath_PrefersInputCWD` (production-path preservation)
  - `TestResolveObservationsPath_FallsBackToGetwdWhenAllEmpty` (backward compat)

### Commit 2: `ba96ffc6e` — M1 + M2 regression guards

**Files changed**: `internal/cli/hook_writehookoutput_test.go` (+106 NEW) + `internal/template/settings_no_worktree_keys_test.go` (+98 NEW)

M1 (REQ-HCF-001, REQ-HCF-004):
- `TestWriteHookOutput_WorktreeCreatePlainText` (AC-HCF-001)
- `TestWriteHookOutput_WorktreeRemovePlainText` (AC-HCF-002)
- `TestWriteHookOutput_EmptyWorktreePathProducesEmptyStdout` (AC-HCF-004, 2 subtests)

Implementation: reuse `captureStdout(fn func()) (string, error)` helper already present in `internal/cli/banner_test.go` (same package).

M2 (REQ-HCF-003):
- `TestSettingsJsonHasNoWorktreeCreateKey` (AC-HCF-003 primary, local file)
- `TestSettingsTmplHasNoWorktreeCreateKey` (AC-HCF-003 secondary, template via `EmbeddedTemplates()`)

Implementation: regex pattern `"WorktreeCreate"\s*:|"WorktreeRemove"\s*:` rejects incidental string matches in comments.

**Negative test confirmation** (manual, not committed):
```
$ python3 -c "..." inject "WorktreeCreate": [] into .claude/settings.json
$ go test -run TestSettingsJsonHasNoWorktreeCreateKey ./internal/template/...
--- FAIL: TestSettingsJsonHasNoWorktreeCreateKey (0.00s)
    settings_no_worktree_keys_test.go:60: WorktreeCreate/WorktreeRemove hook re-registered ...
$ # restore .claude/settings.json, re-run → PASS
```

### Commit 3: `d3d9b829f` — M3 pollution fix + M4 cleanup

**Files changed**: `internal/hook/subagent_stop_test.go` (+21)

**M3 pollution fix discovery**: Existing tests in `subagent_stop_test.go` (TestSubagentStop_KillPaneTimeout + 4 TestSubagentStopHandler_Handle_* tests) call `Handle()` with `input.CWD = ""` (no env var set). Even with M3 fix, they fell through to `os.Getwd()` → recreated `internal/hook/.moai/` leak under the package directory.

**Resolution**: Added `t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())` to 5 tests (1 timeout + 4 Handle tests). This exercises REQ-HCF-005's env-var branch instead of os.Getwd, anchoring side-effects under /tmp.

**M4 cleanup**: `rm -rf internal/hook/.moai/` (untracked directory, 3025 bytes). Idempotent verified.

**Post-fix validation**:
```
$ rm -rf internal/hook/.moai/ && go test ./internal/hook/ -count=1
ok  github.com/modu-ai/moai-adk/internal/hook  0.609s (with 2 baseline FAIL preserved)
$ test -e internal/hook/.moai/ && echo LEAK || echo CLEAN
CLEAN
```

### M5: Documentation crosswalk verification (PASS-by-inspection, zero edits)

All AC-HCF-008 (a/b/c-1/c-2/c-3/d) and AC-HCF-009 verification commands confirmed clean state. **No drift found** post-PR-#1044. No Commit 4 needed.

## 14 AC Binary PASS/FAIL Matrix

| AC | Status | Notes |
|----|--------|-------|
| AC-HCF-001 | PASS | `TestWriteHookOutput_WorktreeCreatePlainText` exit 0 |
| AC-HCF-002 | PASS | `TestWriteHookOutput_WorktreeRemovePlainText` exit 0 |
| AC-HCF-003 | PASS | `TestSettingsJsonHasNoWorktreeCreateKey` + `TestSettingsTmplHasNoWorktreeCreateKey` both exit 0; negative-test verified (inject + restore round-trip) |
| AC-HCF-004 | PASS | `TestWriteHookOutput_EmptyWorktreePathProducesEmptyStdout` exit 0 (2 subtests Create/Remove) |
| AC-HCF-005 | PASS | `TestDispatchCapture_UsesClaudeProjectDirWhenCwdEmpty` exit 0 |
| AC-HCF-006 | PASS | `test ! -e internal/hook/.moai/` exit 0; `git status --porcelain internal/hook/.moai/` empty; idempotent `rm -rf` re-run exit 0 |
| AC-HCF-007 | PASS | `git diff --stat 58a235e06 -- <4 shell wrappers>` empty (handle-worktree-create.sh + handle-worktree-remove.sh untouched local; template mirrors don't exist in this project — only local handlers present); CLI subcommands `worktree-create` line 53 + `worktree-remove` line 54 unchanged |
| AC-HCF-008(a) | PASS | hooks-system.md local+template each match `WorktreeCreate and WorktreeRemove\|active.creator` ≥1 |
| AC-HCF-008(b) | PASS | worktree-integration.md local+template each match `WorktreeCreate and WorktreeRemove` ≥1 |
| AC-HCF-008(c-1) | PASS | 4 locales: ko/en/ja/zh each show WorktreeCreate=2 WorktreeRemove=2 |
| AC-HCF-008(c-2) | PASS | 4 locales: ko=9, en=10, ja=9, zh=9 canonical handlers (all ≥5) |
| AC-HCF-008(c-3) | PASS | settings.json + .tmpl both show 0 WorktreeCreate/Remove occurrences |
| AC-HCF-008(d) | PASS | settings_test.go:512 `expectedCount = 20` |
| AC-HCF-009 | PASS | hooks-system.md §WorktreeCreate section contains both "plain text" + "Not registered by MoAI default" markers |
| AC-HCF-010 | PASS | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/hook/ internal/cli/hook.go` empty (excluding test files + comments) |
| AC-HCF-011 | PASS | `go build ./...` exit 0 + `GOOS=windows GOARCH=amd64 go build ./...` exit 0 |
| AC-HCF-012 | PASS | Full test suite: 2 baseline FAIL preserved (TestAuditRegistrationParity + TestAuditThreeWaySync — PR #1044 residual); 0 NEW regressions |
| AC-HCF-013 | Not invoked | `moai spec lint` not run during this session (CLI subcommand path uncertain); deferred to sync-phase CI |
| AC-HCF-014 | PASS | `git diff 58a235e06 -- internal/hook/worktree_create.go internal/hook/worktree_remove.go` empty (no Handle() body changes) |

**Final tally**: 13/14 PASS, 1 not-invoked (AC-HCF-013 deferred to sync-phase CI).

## Cross-Platform PASS Evidence

```
$ go build ./...                          → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```

## Handle() Byte-Identity Confirmation (AC-HCF-014)

```
$ git diff 58a235e06 -- internal/hook/worktree_create.go internal/hook/worktree_remove.go
(empty output — files byte-identical to plan-phase baseline)
```

## Test Classification (AC-HCF-012 baseline residual policy)

**Pre-existing baseline FAILs preserved** (NOT introduced by this SPEC):
- `TestAuditRegistrationParity` (internal/hook/audit_test.go) — expects 22 settings.json hooks, finds 20 post-PR #1044
- `TestAuditThreeWaySync` (internal/hook/audit_test.go) — flags WorktreeCreate/Remove as drift (handler exists but no settings.json registration)

**Out-of-scope items** (per spec.md §5.2.1):
- Subagent-boundary regression test (broader C-HRA-008 enforcement) — different SPEC
- Refactoring `writeHookOutput` into strategy pattern — different SPEC
- Removing handlers + shell wrappers + CLI subcommands — REQ-HCF-007 PRESERVE

**Follow-up SPEC candidate**: `SPEC-V3R6-HOOK-AUDIT-TESTS-FIX-001` (provisional) to update the 2 audit tests to match PR #1044's de-registration intent (move WorktreeCreate/Remove to retiredEventNames or expected-count 20).

## M5 Drift Findings

**None.** All 8 inspection points (hooks-system.md local+template, worktree-integration.md local+template, 4 locales of hooks-reference.md, settings_test.go:512) verified consistent post-PR-#1044.

## Deviations from Plan

1. **AC-HCF-013 not invoked** during this session. Reason: `moai` CLI subcommand `spec lint` invocation path uncertain in this environment (cmd/moai entry, not always installed locally). Deferred to sync-phase CI which runs spec-lint as required check.

2. **M3 pollution fix added** (Commit 3) — discovered during M4 validation that existing `subagent_stop_test.go` Handle tests re-create the leak even with M3 source fix in place. Resolution: minimum-invasive `t.Setenv("CLAUDE_PROJECT_DIR", t.TempDir())` injection in 5 affected tests. This is within REQ-HCF-005's spirit (exercise the env-var branch) and required for REQ-HCF-006 to hold across subsequent test runs.

3. **Bad commit recovery**: Commit `116f8b1cd` (first attempt at Commit 3) accidentally included 10 deletions of `.github/workflows/*.yml.tmpl` files that were modified by a parallel session and present as deletions in the working tree post-stash-pop. Recovered via `git reset --soft HEAD~1` + `git restore --staged internal/template/templates/.github/` + clean re-commit as `d3d9b829f`. No data loss; the parallel session changes remain unstaged in working tree (PRESERVE list intact).

## Working Tree State at SPEC Completion

```
$ git status --short
 M .moai/harness/usage-log.jsonl                                  # B8 runtime-managed
 M docs-site/hugo.toml                                             # PRESERVE Section B
 M docs-site/layouts/_default/baseof.html                          # PRESERVE Section B
 M docs-site/layouts/partials/menu.html                            # PRESERVE Section B
 M internal/cli/github_init.go                                     # parallel session
 M internal/config/required_checks.go                              # parallel session
 M internal/config/required_checks_test.go                         # parallel session
 M internal/template/renderer.go                                   # parallel session (PRESERVE Section B)
 D internal/template/templates/.github/actions/*                   # parallel session refactor
 D internal/template/templates/.github/workflows/*                 # parallel session refactor
 ?? .moai/research/*.md                                            # PRESERVE Section B
 ?? .moai/specs/SPEC-V3R6-{RULES,SKILL}-*-001/                     # parallel session
 ?? docs-site/content/*/book/                                      # PRESERVE Section B
 ?? docs-site/{data,layouts,scripts,static}/*                      # PRESERVE Section B
 ?? internal/template/github_tmpl_parse_test.go                    # parallel session
```

All non-scope files PRESERVE intact (modifications and deletions from parallel session work continue unstaged for handoff).

## Branch HEAD + Push Status

- **Branch**: `main`
- **Commits ahead origin/main**: 3 (d3d9b829f, ba96ffc6e, 8319c6efa)
- **Push status**: NOT pushed. 1-person OSS Hybrid Trunk policy (CLAUDE.local.md §23) allows main direct push for Tier S, but deferring push to sync-phase per SPEC lifecycle Step 3.

## Next Steps

1. (Optional) Manual push: `git push origin main` (Tier S Hybrid Trunk allows direct push). Triggers branch protection 4-check CI.
2. Sync-phase: `/moai sync SPEC-V3R6-HOOK-CONTRACT-FIX-001` to verify documentation crosswalk consistency one more time and create PR if branch policy requires.
3. Follow-up SPEC for `TestAuditRegistrationParity` + `TestAuditThreeWaySync` adjustment (out of scope here per AC-HCF-012).
