---
id: SPEC-BACKLOG-LOCK-BUDGET-001
title: "Implementation plan — queue lock-wait budget (card t354)"
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

# Implementation Plan — SPEC-BACKLOG-LOCK-BUDGET-001

## §A Context

Card **t354** — "develop CI 봉쇄 해제 — t306 종결분의 잔여". `TestConcurrencyStress` fails on
develop CI (2/48 adds, `kanban board lock held`) and passes locally. Diagnosis, arithmetic, and the
requirement set are in `spec.md`. Attributed introducing commit `83a1d492a` (t306, closed).

Every commit on this branch carries the card id `t354`.

## §B Decisions that are most likely to change — review these first

> **Ordering note.** D2 (breaking the lockstep) is the **primary** fix — it addresses the starvation
> mechanism in `spec.md` §A. D1 (the budget) is **supporting**. D1 is presented first only because
> it is the decision most likely to be argued over; it is not the decision that closes the defect.

### D1 — the shape of the budget (supporting; the most-contested decision)

The budget must stop being a retry count and start being a derived duration. Two shapes are viable
and the choice is a genuine design decision, not a mechanical one:

- **(a) Derived constant.** Named inputs in the source — the contender count, a CI-class
  per-mutation cost, and a headroom factor — combined into one `boardLockWaitBudget`, with the loop
  bounded by elapsed time rather than by attempt count. Simple, statically auditable, and
  AC-BLB-001 reads it directly.
- **(b) Deadline parameter.** The budget becomes an argument (or a field on the store), letting a
  caller widen it for a known-deep queue. More flexible; adds a parameter to two signatures and a
  default that every existing caller must be checked against.

**Recommendation: (a).** The simplicity ladder favours it — nothing in the current call graph wants
a per-call budget, and (b) buys configurability no caller has asked for. (b) is a later change if a
caller ever needs it.

**What the derivation is, and is not, sized for.** It must NOT be presented as covering FIFO queue
depth: `spec.md` §A shows the lock is `flock`-non-blocking with sleep-and-retry polling, so there is
no queue and no fairness, and a `perMutation × supportedWriters` formula is FIFO-shaped reasoning
that does not apply here. What the budget is sized for is **surviving a bounded run of lost races**
under an unfair lock, for the supported contender count, on a CI-class machine.

Concretely: choose the per-mutation cost from the CI-implied ~33 ms (1.57s / 48), not the isolated
local ~14 ms, and state the headroom factor over *that*. The strongest local datum for how thin the
current margin is comes from the package-wide run: worst observation 0.89s = 87% of the 1.025s
budget.

> **Gap (carried into the SPEC, `spec.md` §F).** No closed-form worst case exists for an unfair
> lock — the maximum wait is the tail of a lost-race run, not a queue position, and is unbounded in
> principle. The chosen figure is therefore a **judgement with stated headroom, not a proof**. Say
> so in the source comment. Do not write a derivation whose form implies a guarantee it cannot make.

### D2 — jitter, backoff, or both (primary; this is the decision that closes the defect)

REQ-BLB-003 is satisfied by any per-contender variation. Options:

- **Jitter only** — `base ± rand`. Breaks the lockstep with the least behaviour change; the mean
  wait is unchanged, so a deep queue is polled just as often as today.
- **Backoff only** — deterministic growth. Reduces polling pressure but does **not** break
  lockstep on its own: contenders released together still grow their delays identically.
- **Both** — exponential (or linear) growth with jitter applied to each step.

**Recommendation: jitter, applied to a modest backoff.** Backoff alone fails REQ-BLB-004's intent;
jitter alone leaves a busy poll on a deep queue. Whatever is chosen, the policy must expose its
`[min, max]` bounds so AC-BLB-002 can assert against them rather than against an implementation
detail.

Randomness source: use the standard library's per-call `rand` — no package-level seeding, and no
global generator that a test would have to control. The AC asserts distribution bounds and
distinctness, not specific values, precisely so the test does not depend on a seed.

### D3 — where the shared policy lives

REQ-BLB-006 requires one policy for both call sites. `board_store.go` already owns the constants
and `backlog_store.go` already imports them, so the smallest correct move is a helper beside the
existing constants that both loops call — not a new file, and not a new package. Confirm no third
caller exists before editing (`grep -n boardLockRetries internal/kanban/`).

## §C Pre-flight

- [ ] `grep -n "boardLockRetr\|boardLockRetryDelay" internal/kanban/` — enumerate every consumer;
      the SPEC names two, confirm there is no third.
- [ ] Read `internal/kanban/board_store.go:70-100` and `internal/kanban/backlog_store.go:560-590`
      in full before editing.
- [ ] Confirm the fairness premise still holds in the tree being edited:
      `grep -n "LOCK_NB" internal/kanban/board_lock_*.go` — the non-blocking acquire is what makes
      the starvation account correct (`spec.md` §A). If a blocking acquire has appeared, the
      diagnosis must be re-derived before any edit.
- [ ] Record the pre-change local baselines (both forms AC-BLB-005 compares against — neither is a
      fix signal):
      `go test ./internal/kanban/ -run TestConcurrencyStress -race -count=5` (expected 5/5 PASS) and
      `go test ./internal/kanban/ -race -count=1 -v` ×3 (expected PASS; record the
      `TestConcurrencyStress` duration — pre-change worst was 0.89s).

## §D Constraints

- **No background load.** No verification step spawns background load. Any process a test needs is
  bounded by `t.Cleanup` or an external `timeout`; a trailing `kill` is not cleanup.
- **Package-scoped local testing.** `go test ./internal/kanban/...` only. Never `go test ./...`
  locally — the full-suite verdict is CI's.
- **Do not weaken the guard.** `writers`, `addsPerWriter`, and the test's assertions stay as they
  are. Making the test smaller would make the symptom disappear without touching the defect.
- **Template-First does not apply.** These are Go source files under `internal/`, with no mirror in
  `internal/template/templates/`.

## §E Milestones

Primary fix first, then the supporting budget, then the mechanical convergence.

### M1 — break the lockstep (D2, REQ-BLB-003/004) — the primary fix

Introduce the per-contender wait with its declared `[min, max]` bounds, plus the AC-BLB-002 test.
This is the milestone that attacks the starvation mechanism; the budget work below only widens the
window in which the mechanism has to be survived.

### M2 — the derived budget (D1, REQ-BLB-001/002) — supporting

Replace the retry-count bound with a derived, time-bounded budget in the shared policy. Land the
named inputs and the derivation, and the unit test for AC-BLB-001, together — the test is what makes
the derivation binding rather than decorative. Carry the §B Gap into the source comment: the figure
is a judgement with stated headroom, not a proof.

### M3 — preserve the stuck-holder error (REQ-BLB-005)

Add the AC-BLB-003 test against a `t.Cleanup`-released held lock, asserting the bounded return and
both paths in the error text. If this test passes before M1/M2 as well, that is the point: it is a
regression guard on an existing property.

### M4 — converge both call sites (D3, REQ-BLB-006)

Route `acquireBoardLockSerialized` and `(*BacklogStore).acquireLock` through the one policy; remove
any residual literal. Verify with the AC-BLB-004 grep.

### M5 — verify and land

`go vet ./internal/kanban/...`; `go test ./internal/kanban/... -race`;
`go test ./internal/kanban/ -run TestConcurrencyStress -race -count=5`. Push, open the PR, and read
CI for AC-BLB-006 — both `Test (ubuntu-latest)` and `Race Test` green at `attempt=1`.

## §F Risks

- **The fix cannot be proven locally.** The failure does not reproduce on this machine. M1-M4 are
  verifiable locally; *closure* is only verifiable on CI. Do not report the card fixed on local
  evidence (`spec.md` §F Gaps).
- **A widened budget can mask a real hang.** If `worstCasePerMutation × supportedWriters × headroom`
  is set generously, a genuinely stuck holder is surfaced later than today. REQ-BLB-005's bound is
  the mitigation, and AC-BLB-003 measures it — keep the budget derived, not merely large.
- **Jitter can make a test flaky in the other direction.** AC-BLB-002 asserts distinctness and
  bounds, never specific values; resist any assertion that pins a sampled duration.

## §G Anti-patterns

- Bumping `boardLockRetries` from 40 to some larger number and calling it done. That is the same
  undeclared constant, one notch up; REQ-BLB-001 exists to forbid it — and it leaves the starvation
  mechanism (REQ-BLB-003/004, the primary fix) entirely untouched.
- Justifying the budget as covering the serialized queue depth. That reasoning is FIFO-shaped and
  does not describe this lock (`spec.md` §A "Why the queue-depth reading does not hold").
- Retrying `store.Add` inside the test.
- Marking `TestConcurrencyStress` as flaky, skipping it on CI, or dropping `-race`.
- Running the full local suite to "be sure" (§D).

## §H Cross-references

- `spec.md` §A (arithmetic, the 8×6=48 correction), §B (REQ-BLB-001…006), §F (Gaps).
- `acceptance.md` (AC-BLB-001…006).
- `internal/kanban/board_store.go`, `internal/kanban/backlog_store.go`,
  `internal/kanban/backlog_concurrency_test.go`.
