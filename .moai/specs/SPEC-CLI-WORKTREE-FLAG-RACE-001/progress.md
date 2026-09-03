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

**Resumed by a different lane at HEAD `6a56c96bd`** (the original lane left no report; plan-phase
artifacts re-read directly, `plan-audit-iter1.md` + `red-race-d592b0551.txt` both present). Local
`develop` at `4e91bf6a9` was absorbed first per the lead's dispatch (merge commit `6a56c96bd`,
conflict-free; one transient `fatal: Unable to write index.` on the first attempt — no path in the
message, fsmonitor unset, lock count 0 → lead's cause-B disposition, one retry, succeeded).

### AC-WFR-001a — fresh RED on the moved head (recorded before any code edit)

- Command: `go test ./internal/cli/ -run TestResolveWorktreeExistingBranch -count=20 -race`
- Tree: `6a56c96bd` (absorbed tip), this run
- Output: `.moai/reports/t464/red-race-6a56c96bd.txt`, 1757 lines
- Exit code: `1` (background task bbusrtk32 exit status, read directly)
- `WARNING: DATA RACE` count: **30** (`grep -c`, count read as printed value)
- `Log in goroutine after` count: **0** — this run panicked nowhere and **completed** all
  20 iterations, unlike the `d592b0551` capture which truncated at a panic (spec §A.2).
  This upgrades the RED: the panic manifestation did not reproduce this run, the warning
  manifestation did, at 30 > 23, on a completed run. RED satisfied (AC-WFR-001a: exit 1
  AND ≥1 warning or the panic — the warnings limb holds).

### AC-WFR-006 — option choice and its criteria (written before the first code edit)

**Choice: Option A — remove `t.Parallel()` from the four siblings (lines 66/96/126/161).**

Evaluation against `spec.md` §D.3, criterion by criterion:

| Criterion | Verdict | Ground |
|---|---|---|
| Blast radius | **A** | A is 4 lines in 1 file — the REQ-WFR-005 fence holds mechanically. B changes a production signature and its fence-compliance would depend on a consumer sweep of `resolveWorktreeExistingBranch` before a single line could be written. |
| What is lost | **A** | The SPEC itself states these four tests are sub-millisecond stubs whose parallelism "buys no measurable wall-time" (`spec.md` §D.3) — Option A loses nothing measurable. |
| Recurrence | **B** (only B wins this one) | A re-adds the defect if a later author re-adds `t.Parallel()`. But the standing regression guard (REQ-WFR-006, the `-race` repetition command) exists under either option, and B's structural fix is itself partial (spec §F4: the default-to-global fallback preserves the seam; the package-wide injection conversion is a separate card per §C). |
| Consistency with the package | **A** | The package's dominant idiom is global-seam assignment (23 files); B introduces a second idiom into one file (`spec.md` §D.3). |
| Reviewability | **A** | A four-line deletion vs a signature change plus four test rewrites. |

Four of five criteria favour A; the one criterion favouring B is mitigated by the standing
regression guard and is a partial fix even under B. This is not "A because it is smaller" —
the deciding grounds are the SPEC's own statement that the lost parallelism is unmeasurable
plus the fence holding mechanically without a consumer sweep.

### Repair

Applied after the entries above were written: the four `t.Parallel()` calls at lines 66/96/126/161
of `internal/cli/worktree_branch_flag_test.go` removed; no other change. The `t.Parallel()` calls
at lines 25 and 44 belong to other tests in the file and are untouched.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
