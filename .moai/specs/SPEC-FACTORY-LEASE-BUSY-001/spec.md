---
id: SPEC-FACTORY-LEASE-BUSY-001
title: "Cause-accurate lease refusal classification for `factory next` and `factory status` read latency (card t1916)"
version: "0.1.0"
status: draft
created: 2026-10-21
updated: 2026-10-21
author: manager-spec (card t1916)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/homestate"
lifecycle: spec-anchored
tags: "factory-next, refusal-classification, record-busy, raced, exit-code, status-latency, card-t1916"
tier: M
card: t1916
depends_on: [SPEC-FACTORY-ATOMIC-LEASE-001]
related_specs: [SPEC-FACTORY-LEASE-RENEW-001, SPEC-TODO-AUTO-PICK-001]
---

# SPEC: cause-accurate lease refusal classification and factory status read latency

## HISTORY

- 0.1.0 — 2026-10-21 — plan-phase artifacts authored (card t1916, reissue of t1743; Tier M; worktree
  `.moai/worktrees/t1916`, branch `WT-factory-next-raced`, plan-start HEAD `8249bff8c3`, full SHA
  `8249bff8c3933d647b4f99e59048595f9808e83a`). Inputs: the lane-observed misclassified refusal and its
  code anchors (`internal/cli/factory_card.go` 469, 492-510, 1600-1614, 1857-1861; test anchors
  `internal/cli/factory_lease_bound_test.go` 109, 151, 298, 443, 464); the leader dispatch (03:16Z
  reissue) and the leader constraint message (03:26Z); and the measured pre-implementation file-map
  survey `.moai/reports/t1916/file-map-survey.md`. **Classification honesty (leader directive):** the
  reissued card text labels this work `Class B` — a label mismatch; the operative contract is the
  original card scope WITH this SPEC binding: plan → run → sync, no ceremony reduction. **Sequencing
  (leader directive):** run-phase implementation is HELD until PR #1859 lands — both this fix and
  #1859 touch `internal/cli/factory_card.go` (FILE overlap, no hunk collision today; the survey
  records the measured hunk map). The hold is a sequencing precondition, not a scope reduction; the
  leader coordinates the resume. Requirement count 12, criterion count 11 (Tier M ceilings 16 / 16).

## Background

The lease path of `moai factory next` (machinery from SPEC-FACTORY-ATOMIC-LEASE-001) bounds one
claim at `factoryLeaseClaimWaitCap` — a 2 s claim deadline plus a 200 ms record busy timeout carried
in the record connection's DSN (`factory_card.go` 61-76, 387-393, 1510-1516). When a claim write
fails, `factoryClaimFailure` (`factory_card.go` 492-510) marks the error, and the nominated-lease
switch (`factory_card.go` 1600-1614) maps the marked error onto a refusal.

Two distinct failure causes share one label today:

1. **Host load, not a race.** When the claim's own deadline passes while a record write is still
   running, `factoryClaimFailure` wraps the error in `errFactoryRecordBusy` and its detail states the
   truth — "the host is under load, not a lock race" — yet the outcome is still marked as the
   busy-record outcome, and the caller maps every `errFactoryRecordBusy` onto the `raced` refusal
   token (exit 4). The operator and the machine consumers both read "raced": another lane won. The
   truth is that no lane won anything; the host was too slow. The masking is the defect.
2. **Record contention.** A SQLite busy/locked error from a competing writer is also marked
   `errFactoryRecordBusy` ("another writer held the record past every retry") and also surfaces as
   `raced`.

A third class — a race genuinely lost to another claimant (stale version, holder mismatch surviving
the lease path's retries) — already exists as its own signal and also surfaces as `raced`.

The contrast case shows the format working correctly when classification is honest: the observed
verbatim refusal `factory next: refused serial-slot: another serial card is in flight and holds the
serial slot` (2026-10-21) names its cause in the token. The one-line refusal contract
(`factory next: refused <token>: <detail>`, exit 4 via `factoryNextRefusedExit`) is established and
stays.

The second axis is `moai factory status` read latency: the status read path
(`newFactoryStatusCommand`, `factory_card.go` 3335+; text writer at 3409+) is observed slow in lane
use. The command is read-only and composes several reads — the effective lease configuration, a pure
queue read, the factory record's `ListCards`, and the record-unavailable log. No measurement exists
yet; the run phase measures the baseline first (REQ-FLB-011) and then fixes the dominant cost.

The CLI verb and the `factory_next` MCP handler share one implementation
(`mcp_factory_card.go:123` relays the same classification), so one fix repairs both surfaces.

## Requirements

The requirements use GEARS notation. Each active requirement maps to at least one criterion in
`acceptance.md` (mapping in `plan.md` §E). The active requirements are REQ-FLB-001 to REQ-FLB-012.

### Cause-accurate lease refusal classification

- **REQ-FLB-001** The `factory next` lease path shall attribute every refused lease attempt to
  exactly one declared cause class — race loss, record-write contention, claim-deadline expiry, or
  one of the pre-claim refusal classes that already carries its own token — and shall surface that
  class in the reason token of the one-line refusal.
- **REQ-FLB-002** When a lease claim fails because another claimant won the version-checked record
  edge (a stale-version or holder-mismatch error that survives the lease path's own retries), the
  lease path shall refuse with the `raced` reason token and exit code 4.
- **REQ-FLB-003** When a lease claim fails because a record write stayed busy past every retry (a
  SQLite busy or locked error reported while a competing writer held the record), the lease path
  shall refuse with a record-contention reason token distinct from `raced`, with a detail naming the
  record lock, and with exit code 4.
- **REQ-FLB-004** When a lease claim fails because the claim's own deadline passed while a record
  write was still running, the lease path shall refuse with a claim-deadline reason token distinct
  from `raced`, with a detail naming the observed signals — the expired claim deadline and, as
  recorded, the presence or absence of a concurrent busy error (measured runs show the deadline
  signal joined with `SQLITE_BUSY`) — asserting neither a load-only nor a race-only cause, and with
  exit code 4.
- **REQ-FLB-005** The claim-failure classification of this SPEC's cause table shall not surface the
  `raced` reason token for a claim failure whose underlying cause is not a race lost to another
  claimant; the queue-lock-held refusal (`factory_card.go` 1527-1528, preserved out of scope per
  decision-index Q4) keeps its current token and exit code and sits outside this prohibition.
- **REQ-FLB-006** Every lease-path refusal shall render as a single line matching
  `factory next: refused <token>: <detail>` with no embedded newline.
- **REQ-FLB-007** Every lease-path refusal of a nomination shall exit with code 4, and no
  reclassification under this SPEC shall change the exit codes or refusal texts of the existing
  pre-claim refusal classes (serial-slot, quota-hold, backend-skip, foreign-worktree).
- **REQ-FLB-008** When a nominated lease fails after its promotion for any cause class, the lease
  path shall run the same compensation it runs today — acting only on the promotion this invocation
  made and leaving a state another holder owns alone — so the reclassification changes no store
  state and changes only the refusal surface, where the refusal surface includes the wait-retry
  classification: the new record-contention and claim-deadline tokens shall be wait-retryable
  exactly as the `raced` token is today (the `--wait` retry predicate at `factory_card.go`
  1270-1276), so `--wait` keeps waiting through transient contention instead of stopping on the
  first refusal.
- **REQ-FLB-009** Where the lease path decides on a serial-slot outcome — the slot-hold refusal
  computed from the in-memory slot check (`factorySerialInFlightExcluding`, `factory_card.go` 330,
  refusing at 1368) or the peek's over-approximating serial read (`factory_card.go` 1857-1865,
  whose store-read failures fail open) — the lease path shall attribute no race loss to that
  outcome: the slot-hold refusal keeps its `serial-slot` token and text unchanged, and a read
  failure that fails the peek open produces no race claim in any rendered line.
- **REQ-FLB-010** The lease path shall keep the existing cause-accurate refusal texts unchanged for
  the causes this SPEC does not reclassify — in particular the serial-slot refusal line observed
  verbatim as `factory next: refused serial-slot: another serial card is in flight and holds the
  serial slot`.

### `factory status` read latency

- **REQ-FLB-011** The `factory status` read-path change shall be preceded by a committed latency
  baseline captured with the recipe of `acceptance.md` (sample count, host-load condition,
  percentiles), and the same recipe shall produce the post-change measurement so the two are
  comparable; the baseline artifact shall be committed before the first commit that modifies the
  read path.
- **REQ-FLB-012** When the post-change measurement is taken, the `factory status` read path shall
  satisfy the outcome the baseline selects: when the baseline 95th percentile is below 200 ms, the
  outcome is "defect not reproduced" (recorded with the baseline as evidence, no improvement
  claimed, and the post-change 95th percentile recorded at or below max(2× the baseline 95th
  percentile, 200 ms)); otherwise the outcome is "defect fixed", with the post-change
  95th-percentile latency at or below one half of the baseline 95th percentile.

## Out of Scope

### Out of Scope — the queue-lock-held `raced` refusal

- The queue-lock refusal at `factory_card.go` 1527-1528 (token `raced`, detail naming the queue
  lock's wait budget) keeps its current token and text; the test pinning it
  (`TestFactoryLeaseQueueLockStallBounded`, subtests `nominated` and `bare`) keeps passing
  unchanged. Decision-index Q4 records the disposition (default: preserve current behavior).

### Out of Scope — lease bound sizing and PR #1859's scope

- The claim deadline (2 s), the busy timeout (200 ms), and the wait cap are not resized; card t1696
  and SPEC-FACTORY-ATOMIC-LEASE-001 own those numbers (AC-FAL-007's timing machinery is adjacent and
  untouched).
- PR #1859's changes (self-dispatch gate, SpecID eligibility, bundle predecessor) are not modified;
  this SPEC only rebases its own sites after #1859 lands.
- No new `factory status` fields, verbs, or output shapes: the status report's content is unchanged;
  only its latency is in scope.

## Constraints

- One implementation per verb: the classification lives in the shared claim path, so the CLI verb
  and the `factory_next` MCP handler (`mcp_factory_card.go:123`) classify identically by
  construction — no relay-side re-classification.
- The `raced` token keeps its exit-code pairing with `factoryNextRefusedExit` (4) per REQ-FLB-007;
  decision-index Q1 carries the operator decision if a distinct exit code is ever wanted.
- Where a claim failure carries both the deadline signal and a busy error (a joined error), the
  claim-deadline class takes precedence — the deadline is the outer bound.

## Cross-references

- SPEC-FACTORY-ATOMIC-LEASE-001 — the claim machinery, the wait cap, and AC-FAL-007's timing bounds
  this SPEC must not disturb.
- SPEC-TODO-AUTO-PICK-001 §C.2 — the `raced` definition this SPEC narrows to genuine race loss.
- SPEC-FACTORY-LEASE-RENEW-001 — the `factory status` lease fields (REQ-FLR-009) untouched here.
- `.moai/reports/t1916/file-map-survey.md` — the measured overlap survey behind the sequencing hold
  (machine-local evidence, cited per the leader's directive; not an authority-register artifact).
- `decision-index.md` — Q1 (refusal surface), Q2 (latency budget), Q4 (queue-lock scope).
