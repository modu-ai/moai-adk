# SPEC-WORKTREE-KEY-WIRING-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-12
tier: M
artifacts: [spec.md, plan.md, acceptance.md, design.md, progress.md]
depends_on: [SPEC-SESSION-WORKTREE-001, SPEC-CONFIG-KEY-HONESTY-001]
code_baseline: 1d150a27d
card: t655
design_decisions_settled_at_plan:
  - integration-window: option-A-acquire (design.md §1)
  - integration-target: reuse git_strategy develop_branch, no new key (design.md §2)
  - auto_create: explicit advisory scope, wording truth-fix (design.md §3)
cycle_type: ddd
plan_audit:
  iteration_1:
    verdict: FAIL
    score: 0.75
    report: .moai/reports/t655/plan-audit.md
    defects: [D1, D2, D3]
    resolved_in: spec v0.2.0 (design decisions 3/3 survived; mechanical fixes only)
  iteration_2:
    verdict: PASS
    score: 1.00
    date: 2026-09-12
    report: .moai/reports/t655/plan-audit.md
    note: >-
      iter-2 revision (spec v0.2.0) cleared all three iter-1 defects D1-D3;
      plan-audit PASS 1.00. Same file carries both iterations; the PASS
      verdict is the final-iteration record.
```

## §E.2 Run-phase Evidence

Run phase executed by manager-develop (card t655, lane-4 Factory, cycle_type
ddd) in worktree `.claude/worktrees/t655`, branch `WT-worktree-keys-wiring`,
base `1d150a27d`. All commands below were run in this tree in this run; each
AC row's judging command is the acceptance.md selector, `-count=1`.

### AC matrix (E1)

| AC | Status | Verification command (this run, this tree) | Actual output |
|----|--------|--------------------------------------------|---------------|
| AC-WKW-001 | PASS | `go test ./internal/cli/ -run '^TestAutoMergeOffBaseline$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 0.792s` |
| AC-WKW-002 | PASS | `go test ./internal/cli/ -run '^TestAutoMergeHappyPath$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 1.839s` (real-git: develop HEAD = two-parent merge commit, 2nd parent == session branch tip; window released; notice names commit) |
| AC-WKW-003 | PASS | `go test ./internal/cli/ -run '^TestAutoMergeNonCleanExit$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 0.758s` |
| AC-WKW-004 | PASS | `go test ./internal/cli/ -run '^TestAutoMergeUnconfiguredTarget$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 0.786s` (3 subcases: develop_branch empty / github-flow / unreadable config — all silent-safe with the config-key notice) |
| AC-WKW-005 | PASS | `go test ./internal/cli/ -run '^TestAutoMergeBusyWindow$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 0.783s` (live holder + stale record both skip with no lock write; own-session hold proceeds via refresh) |
| AC-WKW-006 | PASS | `go test ./internal/cli/ -run '^TestAutoMergeZeroPush$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 0.787s` (seam log: exactly one `merge --no-ff`; zero push/fetch/pull/remote/origin; acquire→release, released at end) |
| AC-WKW-007 | PASS | `go test ./internal/cli/ -run '^TestAutoMergeConflict$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 1.991s` (real-git conflict: MERGE_HEAD gone after abort; session worktree byte-identical; window released) |
| AC-WKW-008 | PASS | `go test ./internal/cli/ -run '^TestAutoMergeSourceDirty$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 0.764s` (no merge; cleanup's own dirty guard preserves; both notice families attributable) |
| AC-WKW-009 | PASS | `go test ./internal/cli/ -run '^TestAutoMergeTargetGuards$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 0.787s` (target absent + target dirty subcases) |
| AC-WKW-010 | PASS | `go test ./internal/cli/ -run '^TestAutoMergeNoOpSilent$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 0.806s` (empty output; zero lock ops; no merge invocation) |
| AC-WKW-011 | PASS | `go test ./internal/cli/ -run '^TestWorktreeAdvisoryTruthful$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 0.827s` (regex match kept; "is auto-creating"/"will be created" absent; false wording + degradation byte-identical) |
| AC-WKW-012 | PASS | `go test ./internal/config/ -run '^TestShippedConfigKeysHaveReaders$' -count=1` then `! grep -q "AutoMerge has no production reader" internal/config/types.go` | `ok github.com/modu-ai/moai-adk/internal/config 1.584s`; `GREP-CLEAN exit 0` (inventory row W/no deprecate_after; template names consumers for all three auto-* keys; make build re-embedded) |
| AC-WKW-013 | PASS | `go test ./internal/cli/ -run '^TestAutoMergeNoticePrefixDistinct$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 0.778s` (constant-level distinctness vs both removal prefixes; every failure-path notice line carries the prefix; merge failure leaves the caller's exit path untouched — no error return by construction) |
| AC-WKW-014 | PASS | `go test ./internal/cli/ -run '^TestAutoMergeToggleIndependence$' -count=1` | `ok github.com/modu-ai/moai-adk/internal/cli 0.901s` (4-combination truth table; merge-before-remove ordering asserted in the both-on case) |

### Verification batch (E2/E4/E5 + plan §E)

- `go build ./...` → exit 0 (darwin/arm64).
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 (`WINDOWS BUILD OK`).
- `go vet ./internal/cli/ ./internal/config/` → clean (`VET OK`).
- `golangci-lint run ./internal/cli/...` → `0 issues.`; `golangci-lint run
  ./internal/config/...` → `0 issues.` — no new warnings or lints introduced.
- `make build` ran after the template edit (//go:embed re-embedded; no
  catalog.yaml delta).
- Affected-package suites: `go test ./internal/config/ -count=1` →
  `ok ... 2.493s`. `go test ./internal/cli/ -count=1 -timeout 45m` →
  see suite record below.
- Regression selectors: `go test ./internal/cli/ -run '^TestSessionWorktree'
  -count=1` → `ok ... 2.204s`; `-run '^TestPRMergeCleanup' -count=1` →
  `ok ... 1.260s`.
- Coverage (E3): scoped profile over the new path —
  `go test ./internal/cli/ -run 'TestAutoMerge|TestWorktreeAdvisoryTruthful|...'
  -coverprofile` → `sessionExitAutoMerge 100.0%`, `autoMergeNoticef 100.0%`,
  `gitMergeNoFFReal 100.0%`, `gitMergeInProgressReal 100.0%`,
  `gitMergeAbortReal 100.0%`, `gitBranchOfWorktreeReal 100.0%`,
  `gitHeadShortReal 100.0%`, `emitWorktreeAdvisory 100.0%`,
  `gitRevListCountReal 85.7%` (one branch: non-numeric rev-list parse error,
  reachable only via a corrupted git output — the strict-parse guard), engine
  decision tree fully covered; ≥ 85% target met on the new merge path.

### Observed reds during the cycle (verification-completeness)

The core assertions were observed failing on known inputs before they were
seen green: `TestAutoMergeHappyPath` failed while a dirty-check bug skipped
the merge (assertion caught it: HEAD not a merge commit);
`TestAutoMergeSourceDirty` failed on a wrong cleanup-notice assertion (EC-13
distinctness violation caught); `TestAutoMergeFailurePaths/abort_also_fails`
failed while the seam mutation was silently dropped (by-value struct) — the
notice assertion caught a successful merge on a path that must have failed.

### Suite record (full affected package)

Two full-package runs of `go test ./internal/cli/ -count=1 -timeout 45m`,
same production code byte-identical to the committed M1–M4 tree:

- **Run 1 — 1339.8s, 1 FAIL** (`TestHookWrapper_LargeStdin_DoesNotExceedTimeout`,
  hook_wrapper_load_test.go:55 "wrapper execution took 1.687s, want < 1s").
  Wall-clock assertion in a pre-existing test OUTSIDE this card's diff
  (untouched at base 1d150a27d). Ran while scoped selector batches + lint +
  `make build` executed concurrently on the same machine.
- **Isolation control**: `-run '^TestHookWrapper_LargeStdin_DoesNotExceedTimeout$'
  -count=1` → 158.9ms, PASS — 6x margin. The run-1 failure measured the
  machine, not the code (load-induced flake of a timing assertion).
- **Run 2 (final) — `ok ... 2392.715s`, exit 0, 0 FAIL** — green even while
  a concurrent lane (t586) ran the same package suite (pgrep-evidenced;
  load 8.8–11.9). This is the recorded package verdict.

Classification: pre-existing load-sensitive test; no new defect; zero
failures attributable to this card's change. CI on origin/develop remains
the integration verdict surface (lane doctrine — lanes do not push).

### Deviations from design.md §5 (with reasons)

1. **force pinned at the seam signature, not just the call site** — §5 says
   "the force argument is a literal false constant on this path, never a
   parameter"; implemented one step stronger: the `autoMergeAcquireLock` /
   `autoMergeReleaseLock` seam signatures do not accept a force parameter at
   all, so no caller on this path can displace a hold even by accident.
2. **Busy-window pre-check before acquire** — §5 pins the lock call shape but
   not the read-before-write; AC-WKW-005 ("the auto path writes no lock
   record" on a held window) requires consulting the record BEFORE acquiring
   (with force=false a stale hold would be taken over silently by the shared
   API). The engine reads, skips on any foreign hold, then acquires.
3. **Target worktree resolved before acquire** (to fill the record's
   `Worktree` field coherently with the acquire verb's record shape); guards
   still run after acquire per plan.md M1 ordering — acquire → guards → merge
   unchanged.
4. No other deviations: no absorb/fetch step, no re-measurement, zero-push,
   reuse of `develop_branch` (no new key), template neutrality kept.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: complete
run_complete_at: 2026-09-12
run_commit_sha: "390d71753" # M4 commit carrying the implementation; M5 evidence commit backfills its own SHA below
run_final_commit_sha: "pending-backfill-m5" # this §E.2-carrying commit cannot cite its own SHA
ac_pass_count: 14
ac_fail_count: 0
preserve_list_post_run_count: 5 # session_name_pattern + tmux_preferred inventory rows, internal/github/pr_merger.go, git-strategy.yaml.tmpl (no develop_branch mirror), git_strategy.worktree_root — all verified untouched in the base..HEAD diff
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: pass
  windows_amd64: pass
total_run_phase_files: 19
m1_to_mN_commit_strategy: "5 commits: plan-artifacts carrier -> M1+M2 (status transition) -> M3 -> M4 -> M5 evidence"
suite:
  internal_config: "ok 2.493s (go test ./internal/config/ -count=1)"
  internal_cli: "ok 2392.715s, exit 0, 0 FAIL (go test ./internal/cli/ -count=1 -timeout 45m; run-1 flake record + isolation control in §E.2)"
l44_pre_commit_fetch: not-applicable # lane does not push (operator doctrine 2026-09-02); lead batch-push owns origin/develop
l44_post_push_fetch: not-applicable
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
