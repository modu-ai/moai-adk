# Progress — SPEC-CLI-WORKTREE-FLAG-RACE-001

Card `t464` · worktree `WT-worktree-flag-race` · base `d592b0551`.

## §E.1 Plan-phase Audit-Ready Signal

**Claim** — the plan-phase artifact set for SPEC-CLI-WORKTREE-FLAG-RACE-001 is authored at Tier M
(`spec.md` + `plan.md` + `acceptance.md` + `progress.md`), status `draft`, with both repair options
carried and no winner declared.

**Evidence**

- SPEC ID regex self-check, executed:
  `[[ "SPEC-CLI-WORKTREE-FLAG-RACE-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`.
- Sibling-scope correction measured in this tree:
  `grep -n 'func Test\|t.Parallel()\|findProjectRootFn\|launcherWorktreeMaterialize\|t.Cleanup' internal/cli/worktree_branch_flag_test.go`
  → four `TestResolveWorktreeExistingBranch_*` functions at lines 65 / 95 / 125 / 160, each with a
  `t.Parallel()` on the following line and seam assignments plus `t.Cleanup` restores.
- Blast-radius measurement: `grep -rln 'findProjectRootFn = \|launcherWorktreeMaterialize = ' internal/cli/*_test.go`
  → 23 files. Two independent function-body scans across the same set (this lane's, and the
  auditor's `awk` scan resetting at each top-level `}`) returned the four in-file siblings and
  **zero** out-of-file `t.Parallel()`-plus-global-write hits.
- RED evidence pre-existing and non-empty: `.moai/reports/t464/red-race-d592b0551.txt`, 1369 lines
  (`wc -l`).
- Frame attribution:
  `grep -o 'internal/cli/[a-z_]*\.go' .moai/reports/t464/red-race-d592b0551.txt | sort | uniq -c`
  → `46 main_test.go`, `52 worktree_branch_flag_test.go`, `5 worktree_branch_flag.go`. The
  production file's 5 frames are at report lines 291 (`:70`) and 461 / 920 / 1288 / 1361 (`:74`).
- RED truncation: `sed -n '1340,1369p' …/red-race-d592b0551.txt` → `panic: Log in goroutine after
  TestResolveWorktreeExistingBranch_NoFlagIsNoop has completed: materialize must not run without
  --branch` at line 1344, then `FAIL … 1.543s`.
- Tree identity at authoring: `git rev-parse --short HEAD` → `d592b0551`;
  `git branch --show-current` → `WT-worktree-flag-race`.

**Baseline-attribution** — all of the above were run in this worktree, in this run, against
`d592b0551`. The RED figures (exit 1, `≥ 23` warnings — a floor from a truncated run) are
**carried from the lane's own prior measurement** recorded in
`.moai/reports/t464/red-race-d592b0551.txt`; the frame-attribution and truncation greps above were
run against that file in this run and are this agent's own measurements.

**Gaps**

- No `-race` run was executed during plan-phase; the RED is cited from the persisted capture, and
  the frame/truncation facts were read out of it rather than re-produced.
- The linux/amd64 behaviour of the RED command is unmeasured.
- The package-wide scan is syntactic: a shared global written indirectly (through a helper or an
  alias) would not have been matched by either scan.
- The completed-run warning total is unknown, because no completed 20-iteration RED run exists.

**Residual-risk** — the option choice is deliberately open, so the plan cannot pre-verify the
diff shape the repair will produce; AC-WFR-003's fence is stated but unexercised until run-phase.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
