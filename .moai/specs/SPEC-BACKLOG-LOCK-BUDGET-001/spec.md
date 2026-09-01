---
id: SPEC-BACKLOG-LOCK-BUDGET-001
title: "Queue lock-wait budget: break the retry lockstep under an unfair lock, and derive the wait budget (card t354)"
version: "0.1.0"
status: implemented
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

# SPEC-BACKLOG-LOCK-BUDGET-001 — Queue Lock-Wait Budget

## §A Context and Problem

`TestConcurrencyStress` (`internal/kanban/backlog_concurrency_test.go`) fails on the develop CI
under contention while passing locally. The failure is not a lost update or an id collision — the
invariants the test was written to guard hold. It is the *lock wait* giving up:

```
backlog_concurrency_test.go:56: 2/48 adds failed under contention; first: mutate backlog
/tmp/.../backlog.json: lock /tmp/.../backlog.lock: kanban board lock held
--- FAIL: TestConcurrencyStress (1.57s)
```

Observed on CI run `33128899299`, head `d566ecc75`, `attempt=1`, in BOTH the
`Test (ubuntu-latest)` and `Race Test` jobs. Package `internal/kanban` total: 12.688s.

The attributed introducing commit is `83a1d492a` (SPEC-TODO-SQLITE-001 M2, card t306). That card
is already closed, which is why this defect is carried by a new card (t354) rather than reopened
there.

### The defect, stated precisely

Both queue-lock acquisition paths bound their wait by a **retry count**, over a fixed, un-jittered
delay, with no fairness underneath:

- `internal/kanban/board_store.go:76-79` — `boardLockRetryDelay = 25ms`, `boardLockRetries = 40`.
- `internal/kanban/board_store.go:83-97` — `acquireBoardLockSerialized`, loop
  `for attempt := 0; attempt <= boardLockRetries`.
- `internal/kanban/backlog_store.go:568-585` — `(*BacklogStore).acquireLock`, the same loop shape
  over the same two constants.

Both loops therefore sleep **41** times: a total wait budget of ~**1.025s**.

A bounded retry count conflates two different situations that demand opposite responses:

1. **"I keep losing races to healthy peers."** The lock is changing hands and real work is
   finishing; the correct response is to keep waiting.
2. **"The holder is stuck."** Nothing is progressing; the correct response is to give up and
   surface a bounded, operator-actionable error.

A retry count cannot distinguish them, and the count chosen is small relative to how long a run of
lost races can last at the concurrency the product supports. The header comment of
`backlog_concurrency_test.go` states the supported figure: **Factory mode runs up to ten lanes
against one queue.** Why "lost races" and not "queue position" is the right frame — and why the
budget alone is not the fix — is established below.

### The measurements

The test uses `writers = 8`, `addsPerWriter = 6`, `wantTotal = 48`
(`backlog_concurrency_test.go:19-94`). All 48 mutations serialize on one lock.

> **Correction of record.** The card text for t354 describes the test as "12 writer". That is
> inaccurate: the source declares 8 writers × 6 adds = 48. Everything below uses the source values.

- **Local, single test, `-race`** (this worktree, HEAD `5e194bba2`,
  `go test ./internal/kanban/ -run TestConcurrencyStress -race -count=5`): 5/5 PASS, per-run
  0.65s / 0.67s / 0.68s / 0.70s / 0.71s ⇒ ~14 ms per mutation.
- **Local, whole package under `-race`** — the shape CI actually runs
  (`go test ./internal/kanban/ -race -count=1 -v`, three times): `TestConcurrencyStress` took
  0.73s / 0.80s / 0.89s, package wall 19.7s. The worst local observation, **0.89s, is 87% of the
  1.025s budget**. No load was spawned to obtain this.
- **CI** (run `33128899299`, head `d566ecc75`, `attempt=1`): the same test takes **1.57s** and
  2/48 adds fail. Package `internal/kanban` total 12.688s. At 48 mutations that is ~**33 ms per
  mutation** — roughly 2.3× the isolated local figure.

The package-wide local number is the strongest available local evidence: measured the way CI
measures, the budget is already 87% consumed on a machine where the test passes.

### Why the queue-depth reading does not hold

An earlier reading of this defect attributed the overrun to queue depth: 48 mutations serialize, so
the last writer waits behind all 47 ahead of it (~0.66s local, over budget on CI). **That reasoning
silently assumes the lock hands out turns in FIFO order, and it does not.**

`acquireBoardLockImpl` applies `unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB)`
(`internal/kanban/board_lock_unix.go:41`) and returns `ErrBoardLockHeld` on failure; the callers
then sleep and retry (`backlog_store.go:573-583`, `board_store.go:83-97`). A non-blocking acquire
plus a sleep-and-retry loop provides **no queueing and no fairness**: a contender that loses does
not gain priority for the next round, and `flock` itself guarantees no ordering between waiters.

Under a fair lock, a writer's own wait would be bounded by the *other* contenders' in-flight
mutations, not by the whole queue: 8 writers ⇒ at most 7 mutations ahead of any single attempt. At
the CI-implied ~33 ms per mutation that is ~0.23s — comfortably inside the 1.025s budget. The
queue-depth account therefore does not predict the observed failure at all, and must not be used to
justify the fix.

### The account that holds: starvation under an unfair lock

Starvation is the mechanism. The thin budget is the condition that makes it visible.

- On CI a mutation takes ~33 ms, which is **longer than the 25 ms retry delay**. A contender waking
  every 25 ms therefore almost always finds the lock still held, and among the writers awake at the
  moment of release the winner is effectively arbitrary.
- With no fairness and 8 contenders, one writer losing **41 consecutive races** is an ordinary tail
  event — which is exactly the observed shape: **2 of 48**, scattered, not a systematic tail of the
  queue. A budget shortfall driven by depth would fail the *last* writers, uniformly.
- The fixed, un-jittered 25 ms sleep compounds it: contenders released together stay phase-locked,
  so an unlucky writer keeps arriving at the same bad moment relative to the holder's release.

This is why the anti-starvation requirements are the primary fix (§B) and the budget is the
supporting one: the budget is not sized to cover a FIFO queue's depth — it is sized to survive a
bounded run of lost races under an unfair lock.

## §B Requirements (GEARS)

Ordered by role in the fix, not by id. **REQ-BLB-003 and REQ-BLB-004 are primary** — they address
the starvation mechanism identified in §A. **REQ-BLB-001 and REQ-BLB-002 are supporting** — they
remove the thin, undeclared budget that makes the mechanism visible. Ids are unchanged from the
first draft of this SPEC; only their ordering and annotation moved.

### Primary — the mechanism

**REQ-BLB-003** (While — contention). *Primary.*
**While** the queue lock is contended, the wait between retries shall vary per contender — by
jitter, by backoff, or by both — so that no contender is systematically beaten by the same peers.

**REQ-BLB-004** (Unwanted). *Primary.*
The retry loop shall not use a fixed delay identical across all contenders.

### Supporting — the condition

**REQ-BLB-001** (Ubiquitous — derivation). *Supporting.*
The queue lock-wait budget shall be derived, in the source, from named inputs with a stated headroom
factor, rather than declared as a bare constant with no derivation. The derivation shall be sized
for a bounded run of lost races under an unfair lock — **not** for FIFO queue depth, which §A shows
does not describe this lock.

**REQ-BLB-002** (Where — supported concurrency). *Supporting.*
**Where** the product supports up to ten concurrent lane writers against one queue (the figure of
record in `backlog_concurrency_test.go`'s header comment and the Factory-mode dispatch doctrine),
the derived budget shall account for that contender count and shall exceed the per-mutation cost
observed on a CI-class machine by at least the headroom factor REQ-BLB-001 states.

### Preserved properties

**REQ-BLB-005** (When — budget elapsed; a preserved property).
**When** the wait budget elapses without the lock being acquired, the acquiring path shall return a
bounded error naming the lock artifact path, and shall not block indefinitely. This property exists
today (`backlog_store.go:583` wraps the failure with `s.path` and `s.LockPath()`); the change shall
not regress it.

**REQ-BLB-006** (Ubiquitous — one policy, both call sites).
`acquireBoardLockSerialized` (`board_store.go:83`) and `(*BacklogStore).acquireLock`
(`backlog_store.go:568`) shall consume the same budget-and-backoff policy, so a change to the
policy applies to both without a second edit.

## §C Acceptance Criteria

Enumerated in `acceptance.md` (AC-BLB-001 … AC-BLB-006).

## §D Exclusions — what this SPEC does NOT build

### Out of Scope — replacing the locking substrate

- Swapping the advisory file lock for a different mechanism (SQLite `BEGIN IMMEDIATE` as the sole
  serializer, a lease server, an OS semaphore). The outer advisory lock stays; only its wait policy
  changes.
- Removing the outer lock in favour of the engine's `UNIQUE(id)` backstop. The lock is what makes
  the whole read-modify-write atomic; the backstop is not a substitute.

### Out of Scope — the test's own invariants

- Weakening `TestConcurrencyStress` to make it pass: lowering `writers`, lowering `addsPerWriter`,
  adding a retry around `store.Add`, or marking the test as flaky/skipped. The test is the guard;
  the budget is the defect.
- Changing the test's zero-lost-update or unique-id assertions.

### Out of Scope — general CI flakiness

- Any other failing or flaky test on develop. This SPEC is scoped to the queue lock-wait budget and
  its two call sites.
- CI runner sizing, job parallelism, or `-race` configuration changes.

### Out of Scope — reopening t306

- Amending SPEC-TODO-SQLITE-001 or its closed card. The attribution to `83a1d492a` is recorded here
  as provenance only.

## §E Constraints

- **No background load in verification.** Verification MUST NOT spawn background load processes.
  Any process a test needs is bounded either by `t.Cleanup` or by an external `timeout` wrapper; a
  trailing `kill` is not cleanup.
- **Local verification is package-scoped.** `go test ./internal/kanban/...` only — never the full
  suite locally. The full-suite verdict belongs to CI.
- **Windows/macOS parity.** The constants are platform-neutral; the change must not introduce a
  platform-conditional wait policy.

## §F Gaps

- **A local run cannot prove the CI failure is gone.** The failure reproduces on CI and not on this
  machine (5/5 local PASS at HEAD `5e194bba2`). Local evidence can therefore only establish that
  the budget derivation, the jitter, and the stuck-holder error behave as specified. The **binding**
  evidence that the defect is closed is the CI job on the PR head — AC-BLB-006.
- **The ~14 ms/mutation figure is a local, `-race`, single-machine measurement.** It grounds the
  derivation's order of magnitude; it is not a portable constant, which is why REQ-BLB-001 requires
  the derivation to be stated in the source with its headroom rather than tuned to this number.
- **No closed-form worst case exists for an unfair lock.** Because `flock` + non-blocking polling
  gives no ordering guarantee (§A), a contender's maximum wait is unbounded in principle — it is the
  tail of a lost-race run, not a queue position. Any budget is therefore a **judgement with stated
  headroom, not a proof**. REQ-BLB-001 requires the judgement and its inputs to be visible in the
  source; it cannot require a derivation that does not exist. This is precisely why REQ-BLB-003/004
  are primary: only breaking the lockstep attacks the tail itself.

## §G Cross-references

- `internal/kanban/board_store.go:76-97` — the constants and the first call site.
- `internal/kanban/backlog_store.go:568-585` — the second call site.
- `internal/kanban/backlog_concurrency_test.go:1-94` — the guard, and the ten-lane figure of record.
- SPEC-TODO-SQLITE-001 (card t306, commit `83a1d492a`) — provenance of the current constants.
