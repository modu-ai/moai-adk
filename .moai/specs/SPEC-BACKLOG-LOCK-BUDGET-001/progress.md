# Progress — SPEC-BACKLOG-LOCK-BUDGET-001 (card t354)

## §E.1 Plan-phase Audit-Ready Signal

- Tier S artifact set authored: `spec.md`, `plan.md`, `acceptance.md` (+ this record).
- SPEC ID regex self-check run as Bash: `PASS`. No directory collision under `.moai/specs/`.
- Requirements REQ-BLB-001 … REQ-BLB-006 in GEARS notation; AC-BLB-001 … AC-BLB-006 in
  Given-When-Then.
- **Causal-account correction applied (coordinator, pre-run).** The first draft explained the
  overrun as FIFO queue depth. That reading was wrong: `acquireBoardLockImpl` uses
  `unix.Flock(fd, LOCK_EX|LOCK_NB)` (`internal/kanban/board_lock_unix.go:41`, verified by direct
  read) with caller-side sleep-and-retry, giving no queueing and no fairness. §A now carries the
  starvation account; REQ-BLB-003/004 are annotated primary and REQ-BLB-001/002 supporting (no ids
  renumbered); the no-closed-form-worst-case Gap is recorded in §F and in `plan.md` §B D1. All
  measurements retained verbatim, plus the package-wide local run (0.73 / 0.80 / 0.89s, worst =
  87% of the 1.025s budget).
- Card-text correction recorded in `spec.md` §A: the test declares 8 writers × 6 adds = 48, not
  "12 writer".
- Gap carried forward (`spec.md` §F): local runs cannot prove the CI failure is gone; AC-BLB-006 on
  the PR head is the binding evidence.
- Status: `draft`. Awaiting plan audit and Implementation Kickoff Approval.

## §E.2 Run-phase Evidence

**Baseline attribution.** Every figure below was measured in THIS worktree
(`.claude/worktrees/t354`, branch `WT-concurrency-stress`), against the tree at HEAD
`e08d5e55c` for the pre-change column and against the working tree carrying this SPEC's diff for
the post-change column. `git rev-list --count --left-right origin/develop...HEAD` read `0	0` at
entry. The SPEC body's original figures (taken at `5e194bba2`) are **superseded** by the
re-measurement below: the branch was fast-forwarded onto current `origin/develop` after the SPEC
was authored.

### Pre-flight (re-confirmed after the fast-forward)

| Check | Command | Observed |
|---|---|---|
| No third consumer of the retired constants | `grep -rn "boardLockRetries\|boardLockRetryDelay" internal/ pkg/ cmd/` | 6 hits: `board_store.go:78,79,86,95` + `backlog_store.go:573,582`. No third consumer. |
| Fairness premise still holds | `grep -n "LOCK_NB" internal/kanban/board_lock_*.go` | `board_lock_unix.go:16`, `board_lock_unix.go:41` — the acquire is still non-blocking, so the starvation account in `spec.md` §A stands. |

### Pre-change baselines (HEAD `e08d5e55c`, this tree, this run)

- `go test ./internal/kanban/ -run TestConcurrencyStress -race -count=5 -v` → 5/5 PASS,
  `0.70 / 0.68 / 0.68 / 0.69 / 0.65 s`.
- `go test ./internal/kanban/ -race -count=1 -v` (whole package — the shape CI runs) → PASS,
  `TestConcurrencyStress` **0.86s**, package wall **20.146s**. That is **84%** of the retired
  1.025s budget.

### AC PASS/FAIL matrix

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-BLB-001 | PASS | `go test ./internal/kanban/ -run TestBoardLockWaitBudgetDerivedFromNamedInputs -race -count=1 -v` | `--- PASS: TestBoardLockWaitBudgetDerivedFromNamedInputs (0.00s)` |
| AC-BLB-002 | PASS | `go test ./internal/kanban/ -run TestBoardLockRetryWaitIsNotLockstep -race -count=1 -v` | `--- PASS: TestBoardLockRetryWaitIsNotLockstep (0.00s)` |
| AC-BLB-003 | PASS | `go test ./internal/kanban/ -run TestBacklogLockStuckHolderSurfacesBoundedNamedError -race -count=1 -v` | `--- PASS: TestBacklogLockStuckHolderSurfacesBoundedNamedError (1.68s)` — returned at 1.68s against a 1.65s budget, inside the `budget x 2` = 3.3s bound |
| AC-BLB-004 | PASS | `grep -n 'time\.Sleep([0-9]\|boardLockRetries\|boardLockRetryDelay' internal/kanban/board_store.go internal/kanban/backlog_store.go` | no output, exit 1 (zero hits). Companion: `grep -n 'boardLockWaitBudget\|boardLockRetryWait' …` shows both loops reading `boardLockWaitBudget` (`board_store.go:158`, `backlog_store.go:575`) and calling `boardLockRetryWait` (`board_store.go:171`, `backlog_store.go:588`). |
| AC-BLB-005 | PASS | see the timing table below | 5/5 + 3/3 PASS, zero adds failed under contention |
| AC-BLB-006 | **PENDING — not claimable here** | CI on the PR head | Not run. This is the binding evidence and it belongs to CI (`spec.md` §F). Nothing below establishes the defect is closed. |

### AC-BLB-005 — the guard under repetition, pre vs post

| Form | Pre-change (HEAD `e08d5e55c`) | Post-change | Budget consumed (worst) |
|---|---|---|---|
| (1) `-run TestConcurrencyStress -race -count=5` | 5/5 PASS — 0.70 / 0.68 / 0.68 / 0.69 / 0.65 s | 5/5 PASS — **0.66 / 0.65 / 0.62 / 0.68 / 0.66 s** | — |
| (2) `-race -count=1 -v` (whole package, x3) | PASS — `TestConcurrencyStress` 0.86s, wall 20.146s (1 run) | PASS x3 — **0.83 / 0.76 / 1.02 s**, wall **21.424 / 20.547 / 21.028 s** | pre 0.86 / 1.025 = **84%** → post 1.02 / 1.65 = **62%** |

Two observations worth recording:

- The worst post-change observation, **1.02s**, would have been **99.5%** of the retired 1.025s
  budget. On this machine, on this tree, the pre-change budget was one scheduling hiccup from
  failing locally — which is the margin CI was already past.
- Package wall grew **20.146s → ~21.0s**. The delta is the new AC-BLB-003 test, which deliberately
  waits the full budget out against a held lock. No other test slowed.

Zero adds failed under contention in any of the 8 runs.

### RED route per test (verification-completeness §1.1 — no check is complete until its red was observed)

| Test | RED route | Observed failure |
|---|---|---|
| AC-BLB-001 `TestBoardLockWaitBudgetDerivedFromNamedInputs` | (a) test-first compile RED, then (b) **mutant probe** — the compile RED only proves the symbols were absent, not that the assertion bites, so the derivation was replaced with the bare literal `1025 * time.Millisecond` | `budget 1.025s is not the product of its named inputs (10 writers x 33ms x 5 headroom = 1.65s) — a bare literal with no derivable inputs fails REQ-BLB-001`. Mutant reverted; `boardLockWaitBudget = boardLockSupportedWriters * boardLockCIMutationCost * boardLockHeadroom` restored. |
| AC-BLB-002 `TestBoardLockRetryWaitIsNotLockstep` | (a) test-first compile RED, then (b) **mutant probe** — `boardLockRetryWait` forced to return the constant `boardLockWaitMin + boardLockWaitStep` | 7 failures: `attempt 0..5: all 32 contenders drew the identical wait — the retry loop is lockstep` + `one contender drew the identical wait across 16 consecutive attempts — fixed delay`. Mutant reverted. |
| AC-BLB-003 `TestBacklogLockStuckHolderSurfacesBoundedNamedError` (named-error half) | **mutant probe** — this is a guard over pre-existing behaviour, so RED had to be planted: the timeout error's `lock %s` / `s.LockPath()` operand was dropped | `error does not name the lock artifact …/backlog.lock: mutate backlog …/backlog.json: kanban board lock held`. Mutant reverted. |
| AC-BLB-003 (bounded-return half) | **mutant probe** — the deadline widened to `3 * boardLockWaitBudget` | `the mutation blocked 4.965028375s, past the bound of 3.3s (budget 1.65s)`. Mutant reverted. |
| AC-BLB-004 | **pre-change tree** — the grep was run at HEAD `e08d5e55c` before any edit | 6 hits (exit 0): the two constant definitions plus two hits in each loop. Post-change: zero hits, exit 1. |

### Files changed

| File | Change |
|---|---|
| `internal/kanban/board_store.go` | Retired `boardLockRetryDelay`/`boardLockRetries`. Added the shared queue lock-wait policy: `boardLockSupportedWriters` (10), `boardLockCIMutationCost` (33ms), `boardLockHeadroom` (5), the derived `boardLockWaitBudget` (1.65s), the declared bounds `boardLockWaitMin`/`boardLockWaitMax` (5ms / 50ms), `boardLockWaitStep` (10ms), and `boardLockRetryWait(attempt)`. `acquireBoardLockSerialized` now bounds by elapsed time, not attempt count. Import `math/rand/v2` added. |
| `internal/kanban/backlog_store.go` | `(*BacklogStore).acquireLock` routed through the same `boardLockWaitBudget` + `boardLockRetryWait`; no residual literal or local ceiling. Both error-return paths still name `s.path` and `s.LockPath()` and still wrap `ErrBoardLockHeld`. |
| `internal/kanban/board_lock_wait_test.go` | **New.** The three unit guards above. |
| `internal/kanban/backlog_concurrency_test.go` | **Untouched** — `git diff --stat` on it is empty. `writers`, `addsPerWriter`, and every assertion stand as they were. |

### Design notes carried from `plan.md` §B

- **D1 (a) — derived constant**, not a deadline parameter. Nothing in the call graph wants a
  per-call budget.
- **D2 — jitter over a modest linear backoff.** Both halves are load-bearing and the source comment
  says why: backoff alone does not break lockstep (contenders released together grow identically);
  jitter is what makes two contenders at the same attempt index diverge. Randomness is
  `math/rand/v2`'s per-call top-level source — no package-level seeding, no global generator a test
  must control.
- **D3 — the policy lives beside the existing constants** in `board_store.go`. No new file, no new
  package.
- **The §F Gap is written into the source.** The comment above the const block states plainly that
  the product is a *sizing heuristic with stated headroom, NOT a worst-case bound*, and why no
  closed-form worst case exists for an unfair lock.

### Constraints honoured

- No background load was spawned. AC-BLB-003's holder is a value in the test's own scope released
  by a `t.Cleanup`-registered function.
- No full local suite. Every command was `go test ./internal/kanban/…` or narrower.
- The guard was not weakened, skipped, retried, or reduced.
- `internal/kanban/integration_lock.go` and every other lock in the package are untouched.
- No `internal/template/templates/` mirror was created — these are Go sources under `internal/`.

### Gaps (explicitly NOT observed)

- **AC-BLB-006 is unverified and unclaimed.** No push, no PR, no CI run. Whether the CI failure is
  gone is not established by anything above.
- **The defect does not reproduce on this machine.** 5/5 and 3/3 local PASS both pre- and
  post-change. Local evidence establishes the derivation, the jitter, and the stuck-holder bound
  behave as specified — it cannot establish sufficiency.
- **`golangci-lint` was not run.** Only `go vet ./internal/kanban/...` (clean, exit 0).
- **Windows path not exercised.** `board_lock_windows.go` uses atomic-create rather than `flock`;
  the wait policy is platform-neutral and shared, but no Windows run was performed here.
- **Coverage was not measured.** No `-cover` run.

### Residual risk

- **A widened budget surfaces a genuinely stuck holder later**: 1.65s instead of 1.025s. AC-BLB-003
  measures the bound, and the budget stays derived rather than merely large, but the trade is real.
- **More polling.** The mean wait fell from a fixed 25ms to roughly 10-27ms, so a contender makes
  more acquire attempts inside the (larger) window. Each attempt is an open + `flock` + close; the
  added syscall pressure is small but not zero.
- **Jitter can flake in the other direction.** AC-BLB-002 asserts distinctness and bounds only,
  never a sampled value — but a distinctness assertion is probabilistic in principle. With a
  10ms span at nanosecond resolution across 32 draws, the false-failure probability is
  vanishing rather than zero.
- **The headroom factor 5 is a judgement.** `spec.md` §F says why it cannot be a proof.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-28
run_commit_sha: pending-backfill-run
run_status: ac-pass-with-ci-pending
ac_pass_count: 5          # AC-BLB-001 … AC-BLB-005
ac_fail_count: 0
ac_pending_count: 1       # AC-BLB-006 — CI on the PR head; NOT claimable locally
preserve_list_post_run_count: 1   # backlog_concurrency_test.go, diff empty
l44_pre_commit_fetch: not-run     # lane-local worktree; branch entered at 0/0 vs origin/develop
l44_post_push_fetch: not-run      # no push performed — the lead's call
new_warnings_or_lints_introduced: none-observed  # go vet clean; golangci-lint NOT run (Gap)
cross_platform_build:
  darwin_arm64: pass              # go vet ./internal/kanban/... exit 0
  windows_amd64: not-run          # Gap
total_run_phase_files: 3          # 2 modified + 1 new test
m1_to_mN_commit_strategy: single-commit   # M1-M5 land together; the change is one policy across two call sites
```

## §E.4 Sync-phase Audit-Ready Signal

**Baseline attribution.** Every figure below was measured by the orchestrator in THIS worktree
(`.claude/worktrees/t354`), on branch `WT-concurrency-stress`, against the **rebased** tree at HEAD
`a680ea6e8` — the implementation commit `da6b3b438` replayed onto `origin/develop` at `77b2bcae6`.
`git rev-list --count --left-right origin/develop...HEAD` read `0	1` (one commit ahead, clean).
The §E.2 run-phase figures were taken on the pre-rebase tree at `e08d5e55c` and are **not**
re-cited here as sync-phase evidence; the table below is a fresh measurement on the tree that will
be integrated.

### Claim / Evidence / Baseline-attribution

| Claim | Command | Observed output |
|---|---|---|
| The package builds and vets clean on the rebased tree | `go vet ./internal/kanban/...` | exit 0, no output |
| The whole package passes under `-race` | `go test ./internal/kanban/ -race -count=1 -v` | `ok github.com/modu-ai/moai-adk/internal/kanban 20.725s`, zero `FAIL` lines |
| The guard passes under repetition | `go test ./internal/kanban/ -run TestConcurrencyStress -race -count=5` | 5/5 PASS — `0.66 / 0.69 / 0.64 / 0.68 / 0.67 s` |
| `TestConcurrencyStress` cost on the rebased tree | (from the whole-package `-v` run above) | `0.79 s` |
| The three new guards pass | (from the whole-package `-v` run above) | `TestBoardLockWaitBudgetDerivedFromNamedInputs` PASS, `TestBoardLockRetryWaitIsNotLockstep` PASS, `TestBacklogLockStuckHolderSurfacesBoundedNamedError` PASS (1.69 s — it waits the 1.65 s budget out against a `t.Cleanup`-released holder; that duration is the design, not a hang) |
| AC-BLB-004 holds by source read, not only by grep | direct read of `acquireBoardLockSerialized` (`board_store.go:155-173`) and `(*BacklogStore).acquireLock` (`backlog_store.go:719-745`) | Neither function carries a literal delay or a local retry ceiling; both bound by `boardLockWaitBudget` (`board_store.go:158`, `backlog_store.go:730`) and both call `boardLockRetryWait(attempt)` (`board_store.go:171`, `backlog_store.go:743`) |
| The guard was not weakened | `git diff HEAD~1 HEAD -- internal/kanban/backlog_concurrency_test.go` | 0 lines — `writers`, `addsPerWriter`, and every assertion stand as they were |
| No residual retired constant or literal sleep | `grep -n 'time\.Sleep([0-9]\|boardLockRetries\|boardLockRetryDelay' internal/kanban/board_store.go internal/kanban/backlog_store.go` | no output, exit 1 (zero hits) |

**Margin comparison, pre vs post (worst observation of the whole-package `-race` form).**

| | Worst `TestConcurrencyStress` | Budget | Consumed |
|---|---|---|---|
| Pre-change (§E.2, HEAD `e08d5e55c`) | 0.86 s | 1.025 s (retired) | **84%** |
| Post-change (§E.2, HEAD `e08d5e55c` + diff) | 1.02 s | 1.65 s (derived) | **62%** |

**What this establishes, stated exactly.** No regression, plus a margin improvement from 84% to
62% of the wait budget on the machine that measured it. It does **not** establish that the CI
defect is closed. The failure does not reproduce on this machine — it passed 5/5 and 3/3 both
before and after the change — so no local run can distinguish "fixed" from "still latent".

### AC PASS/FAIL matrix at sync

| AC | Status | Basis |
|---|---|---|
| AC-BLB-001 | PASS | `TestBoardLockWaitBudgetDerivedFromNamedInputs` PASS on the rebased tree |
| AC-BLB-002 | PASS | `TestBoardLockRetryWaitIsNotLockstep` PASS on the rebased tree |
| AC-BLB-003 | PASS | `TestBacklogLockStuckHolderSurfacesBoundedNamedError` PASS, returned at 1.69 s inside the `budget × 2` = 3.3 s bound |
| AC-BLB-004 | PASS | zero-hit grep + direct read of both functions (row above) |
| AC-BLB-005 | PASS | 5/5 isolated + whole-package PASS, zero adds failed under contention |
| AC-BLB-006 | **PENDING — not claimable in sync-phase** | CI on the PR head is the binding evidence (`spec.md` §F, `acceptance.md` AC-BLB-006). Nothing in this section, and nothing that can be run in this worktree, discharges it. It is **not** softened and **not** marked PASS. |

**AC-BLB-006 remains PENDING, in as many words.** The criterion requires both the
`Test (ubuntu-latest)` job and the `Race Test` job green at `run_attempt=1` on the PR head, with a
log grep for `failed under contention` returning zero hits. No push, no PR, and no CI run has
occurred for this card at the time of this commit. Because the defect is a contention tail event, a
green re-run would not close it either — `attempt=1` is required by the criterion. The SPEC
therefore closes at `implemented`, **not** `completed`: the terminal transition waits on evidence
this session cannot read.

### Gaps (explicitly NOT observed in sync-phase)

- **CI on the PR head** — the AC-BLB-006 binding evidence. Not run.
- **`golangci-lint`** — not run in sync-phase either. Only `go vet ./internal/kanban/...`.
- **Windows** — `board_lock_windows.go` uses atomic-create rather than `flock`. The wait policy is
  platform-neutral and shared, but no Windows run was performed.
- **Coverage** — no `-cover` run at sync.
- **Full suite** — never run locally, by constraint. The full-suite verdict belongs to CI.

### Residual risk

- **A widened budget surfaces a genuinely stuck holder later** — 1.65 s instead of 1.025 s.
  AC-BLB-003 measures the bound and the budget stays derived rather than merely large, but the
  trade is real.
- **The margin improvement is a single-machine observation.** 84% → 62% was measured here, on a
  machine where the guard never failed. It predicts nothing about the CI machine's tail.
- **The headroom factor 5 is a judgement, not a proof** — `spec.md` §F says why no closed-form
  worst case exists for an unfair lock.
- **Jitter is probabilistic in principle.** AC-BLB-002 asserts distinctness and bounds only; with a
  10 ms span at nanosecond resolution across 32 draws the false-failure probability is vanishing
  rather than zero.

```yaml
sync_complete_at: 2026-08-28
sync_commit_sha: pending-backfill-sync
sync_status: implemented-ci-pending
b12_self_test_a: pass   # grep -c 'SPEC-BACKLOG-LOCK-BUDGET-001' CHANGELOG.md -> 0 (no duplicate entry)
b12_self_test_b: pass   # AC ids in acceptance.md -> 6 distinct (AC-BLB-001..006); CHANGELOG entry states 6
b12_self_test_c: pass   # every path claimed in the CHANGELOG entry verified present via ls
changelog_entry_position: "[Unreleased] -> ### Fixed (top of section)"
frontmatter_status_transitions:
  spec_md: in-progress -> implemented
  plan_md: in-progress -> implemented
  acceptance_md: in-progress -> implemented
  progress_md: n/a           # carries no frontmatter block
  completed_transition: withheld   # AC-BLB-006 (CI on the PR head) is unread; the terminal close waits on it
canary_compliance_check: n/a       # this SPEC defines no forward-looking policy that its own sync tests
mx_tag_validation: not-applicable  # no new exported surface; the policy constants are unexported and the
                                   # existing @MX:ANCHOR on openEngine (backlog_store.go) is untouched
docs_surface: none                 # internal/kanban only; no CLI flag, no template mirror, no docs-site page
```
