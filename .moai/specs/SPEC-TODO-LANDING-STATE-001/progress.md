# Progress — SPEC-TODO-LANDING-STATE-001

Card t331. Plan-phase artifacts authored at tree `3de2f85a2` (worktree `.claude/worktrees/t331`,
branch `WT-card-landing-state`).

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M set).
- SPEC ID regex self-check executed: `SPEC-TODO-LANDING-STATE-001` → `PASS`.
- ID uniqueness confirmed against `.moai/specs/SPEC-TODO-*`.
- Root cause measured, not inferred: `LandedRef = "origin/main"` against an integration branch of
  `develop`. Re-measured at 0.2.0 against `origin/main` `48239c7dc` / `origin/develop` `c6aa61346`,
  observed 2026-08-28T13:15Z: `git rev-list --count --left-right origin/main...origin/develop` →
  `0	349` (`0	329` at 0.1.0 — the count moves, the leading zero does not).

### Version 0.2.0 — plan-audit iteration 1 remediation (FAIL 0.74)

- **Scope split by operator ruling.** The landing-evidence half — storage, the recording verb, an
  observed commit on `todo pr`, the live SPEC-status read — moved to card **t359**, which depends on
  this SPEC landing first. What remains is half A, the discriminator.
- Requirements: **REQ-TLS-001..011** (11, Tier M ceiling 16), GEARS notation, traceable to
  **AC-TLS-001..010** (10, ceiling 16) in spec.md §E. Was 26 REQ / 16 AC at 0.1.0.
- The storage deviation that required a gate ruling at 0.1.0 is **withdrawn with the B half**: this
  SPEC persists nothing, so there is no deviation from the dispatch's storage direction to rule on.
- Open clarifications: **none**. Two were ruled from the sources (plan.md §C); the third retired
  with the storage half.
- AC-TLS-008's RED **observed by planting a mutant** in `runTodoPR`, running
  `TestTodoPR_QueueDirUnchanged` (4/4 sub-cases FAIL), and reverting; production tree restored and
  re-verified GREEN. Full record: acceptance.md §D.1.
- Citations: seven corrected, every remaining one re-opened at its address at HEAD `11426a128`.
- Not committed; awaiting lead review.

## §E.2 Run-phase Evidence

Run-phase entered at HEAD `87b16c345` (worktree `.claude/worktrees/t331`, branch
`WT-card-landing-state`, `0 6` ahead of `origin/develop` at entry). Baseline measured at that HEAD
before the first edit: `go build ./...` → exit 0; `go test ./internal/kanban/... ./internal/cli/
-run 'TestTodoPR|TestTodoDone|TestPrlink|TestLanded' -count=1` → `ok internal/kanban 0.939s` /
`ok internal/cli 4.342s`.

### Milestone commits

| M | SHA | Subject |
|---|---|---|
| M1 | `260ea5369` | resolve the landed ref from the integration branch |
| M2 | `9ba33d0a2` | pin the unknown outcome at the todo pr surface |
| M3 | `61424aed0` | say on stdout whether the landing guard ran |
| M4 | `9414374b4` | put the queue state on the todo pr row |
| AC-TLS-008 | `f10827fd3` | draw the read-only assertion at the verb's reach |
| M5 | `5be48b3f8` | state the landed check's limits where they are read |

Six commits rather than five: AC-TLS-008's digest widening is separated from M4 so its four-step
mutant transcript (below) carries its own evidence rather than riding another milestone's.

**Milestone-boundary deviation, stated rather than glossed.** M1 and M2 could not be staged
separately at the implementation level. Replacing `Landed`'s `(bool, error)` with the three-valued
answer is a single compile-forced edit that necessarily drags the resolver's fifth outcome kind
with it, so M1's commit carries M2's behaviour and M2's commit carries only its criteria. The M2
criteria were consequently GREEN on arrival; they are committed as regression-guards, and the
flipped RED for that axis is the reversed `TestResolve_LandedErrorDegradesToUnknown` in M1.

### AC PASS/FAIL matrix (against acceptance.md §B, counted at HEAD `5be48b3f8`)

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-TLS-001 | PASS | `go test ./internal/kanban/ -run 'TestLandedRefFor\|TestLandedGrepArgs_CarriesTheResolvedRef\|TestLandedQuerier_AnswersAboutTheConfiguredRef' -count=1 -v` | `--- PASS: TestLandedRefFor (0.01s)` (4 sub-cases) / `--- PASS: TestLandedGrepArgs_CarriesTheResolvedRef (0.00s)` / `--- PASS: TestLandedQuerier_AnswersAboutTheConfiguredRef (0.62s)` / `ok internal/kanban 0.957s` |
| AC-TLS-002 | PASS (regression-guard) | same run, sub-case `TestLandedRefFor/empty_configuration_keeps_the_default` | PASS; `DefaultLandedRef == "origin/main"` asserted in the same test |
| AC-TLS-003 | PASS | `go test ./internal/cli/ -run TestTodoDone -count=1 -v` | `--- PASS: TestTodoDone_RequireLandedRefusesWhenNotLanded (0.27s)` — the assertion now reads `todoLandedRef()` rather than a constant |
| AC-TLS-004 | PASS | `go test ./internal/kanban/ -count=1` | `ok internal/kanban 13.993s`, incl. `TestResolve_LandedErrorDegradesToUnknown` (reversed from the former `…DegradesToNoLink`) and its answered-empty control |
| AC-TLS-005 | PASS | `go test ./internal/cli/ -run 'TestTodoPR_UnanswerableRendersUnknownNotNoLink\|TestTodoPR_UnknownReachesJSON' -count=1 -v` | `--- PASS: TestTodoPR_UnanswerableRendersUnknownNotNoLink (0.20s)` / `--- PASS: TestTodoPR_UnknownReachesJSON (0.17s)` / `ok internal/cli 1.407s` |
| AC-TLS-006 | PASS | `go test ./internal/cli/ -run TestTodoDone -count=1 -v` | `--- PASS: TestTodoDone_StdoutCarriesTheLandingVerdict (0.70s)` (3 sub-cases). RED before M3, verbatim: `stdout = "done t1\n" carries 0 landing verdicts, want exactly 1` and `a satisfied guard and an unanswerable one print the same bytes ("done t1")` |
| AC-TLS-007 | PASS (regression-guard) | same run | `--- PASS: TestTodoDone_UnknownStillProceeds (0.24s)` and `--- PASS: TestTodoDone_RequireLandedProceedsWhenInconclusive (0.24s)` — archive + exit code unchanged |
| AC-TLS-008 | PASS | `go test ./internal/cli/ -run TestTodoPR_QueueDirUnchanged -count=1 -v` | `--- PASS: TestTodoPR_QueueDirUnchanged (0.98s)`, 6/6 sub-cases. Mutant transcript below |
| AC-TLS-009 | PASS | `grep -n 'LAST step'` + `grep -c -i 'unknown'` on both `todo.md` copies; `diff`; `strings bin/moai \| grep -c` | `LAST step` → lines `51` (the `todo pr` outcome row) and `60`; `unknown` → `3`, rc=0; `diff` of the two copies empty; embedded-bundle grep → `2` |
| AC-TLS-010 | PASS | `go test ./internal/cli/ -run TestTodoPR_RowCarriesQueueState -count=1 -v` | `--- PASS: TestTodoPR_RowCarriesQueueState (0.16s)`. RED before M4, verbatim: `row [t1 no-link   picked but no commits] has 5 columns, want 6` |

10 PASS, 0 FAIL, 0 PASS-WITH-DEBT. Eight release-blocking, two regression-guards (AC-TLS-002,
AC-TLS-007) — classification per acceptance.md §C, unchanged.

### AC-TLS-008 — the operator's kickoff condition, four steps in order

Plan-phase proved the widened assertion with a throwaway probe; the kickoff condition was that the
gap be closed by observing the SHIPPED test (`TestTodoPR_QueueDirUnchanged`, its `queueDirDigest`
helper widened from `kanban.StateDirForRoot(root)` to `root` with `.git` skipped). None of the four
steps substitutes for an earlier one, and the plan-phase probe transcript is NOT reused as evidence.

1. **Before planting** — `go test ./internal/cli/ -run TestTodoPR_QueueDirUnchanged -count=1 -v` →
   `--- PASS: TestTodoPR_QueueDirUnchanged (1.03s)`, all 6 sub-cases. This is what excludes a
   vacuously-red criterion.
2. **Mutant planted** — `_ = writeLandingSweepCacheMUTANT(rec)` at the head of `runTodoPR`, writing
   `<root>/.moai/cache/landing-sweep.cache` on every invocation. Same command, verbatim:

   ```
   --- FAIL: TestTodoPR_QueueDirUnchanged (1.60s)
       --- FAIL: TestTodoPR_QueueDirUnchanged/linked_and_ambiguous (0.60s)
       --- FAIL: TestTodoPR_QueueDirUnchanged/json_form (0.21s)
       --- FAIL: TestTodoPR_QueueDirUnchanged/landed_path (0.20s)
       --- FAIL: TestTodoPR_QueueDirUnchanged/fail-open_path (0.19s)
       --- FAIL: TestTodoPR_QueueDirUnchanged/landed_and_picked
       --- FAIL: TestTodoPR_QueueDirUnchanged/unanswerable_path
       todo_pr_test.go:226: queue directory changed across the invocation
           before:
           .moai/state/todo/backlog.db cf843d07b190559d597e5c13c84f09765caead2f874f080f175273ab3accf7ec
           .moai/state/todo/backlog.lock 014fab0b1564fdf0b0fe024369fc58c24d7b699da130ca027b3f4d056efe8ecb
           after:
           .moai/cache/landing-sweep.cache 9da572e0bfc0ee22af846684a2a9825959b38a5560fea2ae4a326d2663a68f84
           .moai/state/todo/backlog.db cf843d07b190559d597e5c13c84f09765caead2f874f080f175273ab3accf7ec
           .moai/state/todo/backlog.lock 014fab0b1564fdf0b0fe024369fc58c24d7b699da130ca027b3f4d056efe8ecb
   ```

3. **Reverted** — `shasum internal/cli/todo_pr.go` reads
   `a3292e44fe1b37df1cb7c2eb58db537f6754f615` both before the plant and after the revert, and
   `git status --short internal/cli/todo_pr.go` is empty.
4. **Re-run** — `--- PASS: TestTodoPR_QueueDirUnchanged (0.98s)`, 6/6.

**Approved residual, kept accurate and NOT closed.** A mutant writing OUTSIDE the fixture root
(`$HOME`, `/tmp`, a global cache) still evades the criterion. The criterion is drawn at *the verb's
reach within the project*, NOT at *this verb writes nowhere*; the helper's own comment says so.

### Contract changes for the sync-phase notes

Two machine-readable surfaces changed and both are named here so the sync phase can carry them:

1. **`todo pr` gains a fifth outcome kind, `unknown`** — reaches `--json` as
   `"outcome":"unknown"`. The in-repo consumer count is zero (`internal/cli/todo_pr.go` formats
   `o.Kind` with `%s` rather than switching on it; `grep -rn 'PRLink' internal/web` → rc 1);
   external consumers of `todo pr --json` are the open residual.
2. **The `todo pr` row goes from five tab-separated columns to six** — `state` inserted between
   `Confidence` and the card text. A consumer doing `cut -f5` now gets the state; one reading the
   LAST field still gets the text.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-29
run_commit_sha: 5be48b3f8
run_status: audit-ready
ac_pass_count: 10
ac_fail_count: 0
ac_pass_with_debt_count: 0
preserve_list_post_run_count: 0        # nothing under internal/kanban/backlog_*.go modified;
                                       # `git diff --name-only 260ea5369^..HEAD -- internal/kanban/backlog_`
                                       # returned empty (acceptance.md §E gate 5)
l44_pre_commit_fetch: not-run          # no fetch performed this run; the card worktree was entered
                                       # at HEAD 87b16c345 with origin/develop already absorbed
                                       # (`0 6`), and no push was made. Reported as a gap, not a pass.
l44_post_push_fetch: not-applicable    # nothing pushed — the operator pushes; the lead names the
                                       # integration window
new_warnings_or_lints_introduced: 0    # `golangci-lint run --timeout=5m ./internal/kanban/...
                                       # ./internal/cli/...` → "0 issues.", exit 0
cross_platform_build:
  host: pass                           # `go build ./...` → exit 0
  windows_amd64: pass                  # `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
  windows_test_compile: not-run        # `GOOS=windows go vet` not executed this run — gap
coverage:
  internal_kanban: 87.1                # `go test -cover ./internal/kanban/` → coverage: 87.1%
  internal_cli: 79.9                   # `go test -cover ./internal/cli/` → coverage: 79.9%;
                                       # below the 85% target, and NO pre-change baseline was
                                       # measured at 87b16c345, so the delta is unattributed — a
                                       # gap, not a claim that this change caused or avoided it
full_suite: ci-pending                 # local run scoped to affected packages per the standing
                                       # contract; CI on the integration branch owns the verdict
total_run_phase_files: 9               # internal/kanban/{prlink.go,prlink_landed.go,prlink_test.go,
                                       # prlink_landed_test.go,prlink_landedref_test.go},
                                       # internal/cli/{todo.go,todo_pr.go,todo_pr_test.go,
                                       # todo_undone_test.go,todo_landing_test.go} + both todo.md
                                       # copies + spec.md frontmatter
m1_to_mN_commit_strategy: per-milestone commits, no push (operator pushes; lead names the window)
milestone_shas: [260ea5369, 9ba33d0a2, 61424aed0, 9414374b4, f10827fd3, 5be48b3f8]
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_
