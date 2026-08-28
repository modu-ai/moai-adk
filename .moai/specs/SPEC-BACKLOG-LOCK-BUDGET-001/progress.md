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

_<pending sync-phase>_
