---
id: SPEC-BACKLOG-LOCK-BUDGET-001
title: "Acceptance criteria — queue lock-wait budget (card t354)"
version: "0.1.0"
created: 2026-08-28
updated: 2026-08-28
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: "internal/kanban"
lifecycle: spec-anchored
tags: "kanban, backlog, lock, contention, ci-flake, t354"
tier: S
---

# Acceptance Criteria — SPEC-BACKLOG-LOCK-BUDGET-001

## §D AC Matrix

| AC | Covers | Verified by |
|---|---|---|
| AC-BLB-001 | REQ-BLB-001, REQ-BLB-002 | unit test over the derivation |
| AC-BLB-002 | REQ-BLB-003, REQ-BLB-004 | unit test over the wait sequence |
| AC-BLB-003 | REQ-BLB-005 | unit test over a held lock, bounded |
| AC-BLB-004 | REQ-BLB-006 | source-level single-policy check |
| AC-BLB-005 | the guard itself | `TestConcurrencyStress` under `-race -count=5` |
| AC-BLB-006 | the defect being closed | CI on the PR head — the binding evidence |

**No background load.** No criterion below spawns background load. Where a criterion needs a lock
held while another goroutine contends for it, the holder is released by a `t.Cleanup`-registered
function or the whole command is bounded by an external `timeout` wrapper. A trailing `kill` does
not satisfy this.

**Local scope.** Every local command is `go test ./internal/kanban/...` or narrower. The full suite
is never run locally.

---

### AC-BLB-001 — the budget is derived from named inputs, with headroom

**Given** the queue lock-wait policy in `internal/kanban`,
**When** a unit test reads the policy's named inputs — the supported contender count, the CI-class
per-mutation cost, and the headroom factor — and recomputes the budget from them,
**Then** the recomputed value equals the policy's effective budget, and that budget is at least
`headroom ×` the per-mutation cost × the supported contender count.

Binary: the test asserts an inequality over named constants and fails if the budget is ever lowered
below the derivation, or if a constant is changed without the derivation following it. A bare
numeric literal with no derivable inputs fails this criterion.

**What this criterion does NOT claim.** The product `perMutation × contenders` is a sizing
*heuristic*, not a worst-case bound: the lock is `flock`-non-blocking with sleep-and-retry polling,
so there is no queue and no fairness, and no closed-form worst case exists (`spec.md` §A, §F). This
AC verifies that the figure is derived from visible inputs and carries stated headroom — it does not
verify that the budget is sufficient. Sufficiency is only evidenced by AC-BLB-006.

---

### AC-BLB-002 — the retry wait is not lockstep

**Given** the retry-wait function used by both acquisition paths,
**When** a unit test samples the wait it produces across a run of consecutive attempts, and across
independent contenders at the same attempt index,
**Then** the sampled values are not all identical, and every sampled value lies within the policy's
declared minimum and maximum wait bounds.

Binary: distinct-value count > 1 for the same attempt index across contenders, AND every sample
within `[min, max]`. A constant 25 ms for every sample fails.

---

### AC-BLB-003 — a stuck holder still surfaces a bounded, named error

**Given** the queue lock is held by a handle the test itself owns, released via `t.Cleanup`,
**When** a second acquisition runs against the same lock path and the budget elapses,
**Then** the call returns an error (never blocks past the budget), the error is recognized by
`IsBoardLockHeld`, and the error text names both the queue file path and the lock artifact path.

Binary: the call returns within a wall-clock bound of `budget × 2` measured inside the test; the
error string contains both paths. The test spawns no background process; the holder is a value in
the test's own scope with a `t.Cleanup` release.

---

### AC-BLB-004 — one policy, both call sites

**Given** `acquireBoardLockSerialized` (`internal/kanban/board_store.go`) and
`(*BacklogStore).acquireLock` (`internal/kanban/backlog_store.go`),
**When** the source is inspected after the change,
**Then** both obtain their budget and their per-attempt wait from the same shared policy function or
constant set — neither carries its own literal delay or its own retry ceiling.

Binary: a grep for a numeric sleep literal inside either function returns zero hits, and both call
the same named helper. Cited as command + output.

---

### AC-BLB-005 — the guard passes locally under repetition

**Given** this worktree,
**When** both of these are run —
1. `go test ./internal/kanban/ -run TestConcurrencyStress -race -count=5` (isolated), and
2. `go test ./internal/kanban/ -race -count=1 -v`, three times (whole package — the shape CI runs),
**Then** every run PASSES with zero adds failing under contention, and the observed
`TestConcurrencyStress` duration in (2) is recorded.

Baselines for comparison, both pre-change:
- (1) at HEAD `5e194bba2`: 5/5 PASS, 0.65 / 0.67 / 0.68 / 0.70 / 0.71s.
- (2) pre-change: 0.73 / 0.80 / 0.89s, package wall 19.7s — worst 0.89s = 87% of the 1.025s budget.

Form (2) is the more informative local check: it is measured the way CI measures. This criterion
establishes **no regression plus a margin comparison**, not the fix — see AC-BLB-006 and `spec.md`
§F Gaps.

---

### AC-BLB-006 — the binding evidence: CI on the PR head

**Given** the branch for card t354 pushed and a PR opened,
**When** the CI workflow runs on the PR head,
**Then** both the `Test (ubuntu-latest)` job and the `Race Test` job report success on
`attempt=1`, with no `adds failed under contention` line in either job's log.

Binary: `gh run view <id> --json jobs` shows both jobs `conclusion: success` at `run_attempt=1`, and
a log grep for `failed under contention` returns zero hits. `attempt=1` is required — a green
re-run does not close a contention defect.

**Resolved — 2026-09-02.** Observed by the lead's CI read: run `33564147725`, develop head
`09bf452c0` (a descendant of this card's landing merge `728f91006`), `run_attempt=1`, both jobs
success, on a run whose log proves the kanban package executed. The "PR head" surface became a
develop head under the repo's git-flow transition — the evidence intent is unchanged. Full
record: `progress.md` closure block. This SPEC closed `implemented → completed` on this read,
jointly discharged with SPEC-STRESS-INVARIANT-VERDICT-001's AC-SIV-013.

## §D.1 Definition of Done

- AC-BLB-001 … AC-BLB-005 pass locally, each cited as command + verbatim output.
- AC-BLB-006 passes on the PR head, cited as the run id, head SHA, and `run_attempt`.
- `go vet ./internal/kanban/...` clean.
- No test in `internal/kanban` was weakened, skipped, or had its concurrency lowered (diff review).
