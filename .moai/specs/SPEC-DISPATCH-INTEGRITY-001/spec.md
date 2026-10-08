---
id: SPEC-DISPATCH-INTEGRITY-001
title: "Factory dispatch and bundle integrity — eight P1 review findings"
version: "0.1.6"
status: completed
created: 2026-10-09
updated: 2026-10-09
author: manager-spec
priority: P1
phase: "v3.2.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "factory, dispatch, bundle, memory-fold, concurrency, review-findings, t1595"
tier: M
---

# SPEC-DISPATCH-INTEGRITY-001 — Factory Dispatch & Bundle Integrity

## History

- 2026-10-07 — the t1498 review round observed 8 P1 defects across the factory
  dispatch/bundle path (`internal/cli/factory_bundle.go`, `factory_card.go`,
  `factory_card_pr.go`) and the memory-fold write path
  (`internal/cli/memory_fold.go`). Observation record:
  `.moai/reports/t1595/overlay/final-results.txt` (local-only).
- 2026-10-08 — card t1595 issued for the 8-defect bundle. The predecessor lane
  (lane-9 gen 5) completed the defect-to-RED-test mapping and relayed the
  overlay test assets (`.moai/reports/t1595/handoff-to-card-session.md`).
- 2026-10-08/09 — plan-phase baseline re-verification on THIS tree
  (81786284e) classified the 8 (plan.md §B): defects (4) and (5) are NOT
  reproduced (already repaired by landed work — both overlay tests PASS);
  defect (8) is RED-live (both fold tests FAIL); defects (1)(2)(3)(6)(7) have
  no overlay test and classify RED-first in M1–M3.
- v0.1.0 (2026-10-09) — initial SPEC authored (card t1595, plan phase,
  Tier M).
- v0.1.1 (2026-10-09) — plan-audit round-1 repairs (verdict FAIL 0.78 vs
  Tier M 0.80; defect delta D1–D8 + codex gate r1 convergence): AC-DI-010
  and its committed test re-authored for serialized semantics (fold B as a
  separate process completing after A releases the lock); AC-DI-009's
  guarantee scoped to the last observable byte comparison with the
  irreducible TOCTOU tail stated as residual risk and the seam call site
  pinned against relocation in both directions; AC-DI-003 extended to the
  multi-hub case (two hubs, independent predecessors, blocked until both
  merge); §2.1 four-element evidence ledger added for the baseline cells;
  lane-probe deferral made two-sided via the tracked `deferred/` archive
  (owner card t1596).
- v0.1.2 (2026-10-09) — plan-audit round-2 repairs (iter-2 delta 0.90,
  D1–D8 verified RESOLVED; D9/D10 blocking): AC-DI-009 Given/Then and plan
  §G/M4 state the consistent post-fix write geometry explicitly (byte
  comparison cmp1 → seam probe → NEW final byte comparison → rename;
  "final byte comparison" is a role the post-probe comparison takes — a
  byte comparison must follow the probe in every acceptable geometry;
  gate option (a), strong AC kept); the evidence ledger gains
  measurement-input binding — the four owned test sources are tracked at
  `owned-tests/` (mirror + drop-in procedure), closing the replay hazard
  (`no tests to run`, exit 0, demonstrated by the codex gate); new ledger
  captures use anchored selectors.
- v0.1.3 (2026-10-09) — plan-audit round-3 repairs (the ceiling policy's
  single automatic delta round; D12–D14 + gate-P2): AC-DI-009's committed
  test body asserts the preserved concurrent-author bytes, not an error
  value alone (the detect-then-overwrite mutant class — D12), mirrored in
  `owned-tests/`; AC-DI-005's guard observes the runFactory error AND the
  card state in a two-arm assertion, with the positive-control test
  `TestReviewFindingNominatedLeasesAfterPredecessorMerges` proving the
  merged-predecessor path actually leases (D13); AC-DI-012's family
  enumeration adds `factory_nominate_test.go` and derives from the
  touched functions' callers at fix time (D14); the AC-DI-010 test
  rewrite + RED measurement move from M4 to M0 — M4 keeps the
  implementation fix + GREEN verification (gate-P2 sequencing).
- v0.1.4 (2026-10-09) — micro-followup (leader disposition, operator
  decision A): D15 — the nomination negative fixture's nil-error branch
  fails unconditionally ("leased past an unmerged predecessor — dependency
  check did not run"); the positive control is the only place a lease
  outcome is asserted. D16 — every EL cell binds its measurement input to
  the mirror as of e725633e0 (the original bodies the cells executed);
  strengthened-body outcomes remain post-M0 evidence, not retrofitted
  into the ledger.
- v0.1.5 (2026-10-09) — micro-followup closing round (iter-5, leader
  disposition / operator decision A): D16-r — every input-binding pointer
  corrected 5ae7d6ebc → e725633e0, the mirror's first tracked revision
  (verified: `git show e725633e0:…owned_red_tests.go.txt` returns the
  original measured bodies; 5ae7d6ebc does not contain the mirror — the
  off-by-one originated in the iter-4 instruction wording, implemented
  verbatim); D17 — the strengthened nomination body classified by its
  OBSERVED outcome (codex execution at HEAD: PASS — a green-at-adoption
  regression guard, not a RED-now body), AC-DI-005's two-cell text
  aligned; D18 — the methodology control's source tracked at
  `owned-tests/zone_control_test.go.txt` (was gitignored-overlay-only)
  with AC-DI-013's baseline cell bound to that input.
- v0.1.6 (2026-10-09) — final additive round (iter-6, leader
  disposition): D19 — the control's baseline cell added to the evidence
  ledger as EL-005, stated AS pending (the 2026-10-08 PASS observation
  was captured through a piped, tail-bounded command without the exit
  code as its own field and cannot honestly satisfy §2.1's four
  elements; plan §B is the interim pointer; M0 replaces it with the
  measured cell), AC-DI-013 aligned; D20 — the blanket
  "both bodies are RED" sentence removed from the owned-tests README so
  only the per-body observed classification stands.

## Requirements

GEARS notation. Per-defect traceability: REQ-DISPATCH-001 → defect (1) …
REQ-DISPATCH-008 → defect (8), defect numbering per card t1595.

### REQ-DISPATCH-001 (defect 1 — multi-hub bundle predecessor)

**While** a bundle load records members whose files cross one or more hub
paths, **When** a member's hub path is shared with an open, recorded card
outside the bundle, the factory bundle command shall not record that member
with a predecessor chain that leaves the sharer unchecked: the member shall
not become lease-eligible before every such unmerged sharer has reached a
merged state. A single stored hint naming only one hub's sharer does not
satisfy this requirement for a member crossing several hub paths.

### REQ-DISPATCH-002 (defect 2 — bundle first-member hub constraint)

**When** a bundle's first member's files share a hub path with an open,
recorded non-member card, the bundle load shall give the first member a
hub-predecessor constraint against that sharer, or refuse the load; a first
member recorded with an empty after assignment that leaves the conflict
unguarded shall not pass the load.

### REQ-DISPATCH-003 (defect 3 — duplicate member refusal)

**When** a bundle load receives a member list that names the same card id
more than once, the factory bundle command shall refuse the load with a
duplicate-member refusal, and shall record no bundle identity, no after
relation (a self-dependency included), and no lane assignment.

### REQ-DISPATCH-004 (defect 4 — nomination preserves stored after)

**When** `factory next --card` targets a card whose record already carries a
stored after relation, the nomination shall preserve the stored after; a
computed hub hint is a record-creation input only and shall never be written
into an existing row's after field.

### REQ-DISPATCH-005 (defect 5 — merged-pr predecessor releases successor)

**When** a successor card's stored after predecessor sits at the merged-pr
state, selection shall treat that predecessor as merged and the successor
shall be lease-eligible (github-flow delivery parity with merged-local).

### REQ-DISPATCH-006 (defect 6 — no-record arm skips without aborting)

**When** the no-record selection arm evaluates a queue-picked card whose hub
path is shared with an unmerged predecessor, selection shall skip that card
and continue evaluating remaining candidates; the dispatch shall not abort
for the whole pass.

### REQ-DISPATCH-007 (defect 7 — lease validation precedes remote mutation)

**When** a delivery retry runs on a card at the merging state, the
lease-ownership and lease-expiry validation shall run and pass before the
first remote mutation (branch push, pull-request creation, auto-merge
request); no remote mutation shall execute under a lease held by another
lane or held expired.

### REQ-DISPATCH-008 (defect 8 — memory-fold store serialization)

**While** two memory-fold executions act on the same memory store, the fold
shall serialize its index writes through a cross-process lock spanning the
fold's whole write transaction; **When** a concurrent index change is
detected between the fold's last verification and its rename, the fold shall
refuse the write (no silent overwrite), and a completed fold's index line
shall survive every other fold's completion or refusal.

## Constraints

- C1 — Baseline gate: every defect's fix is gated on a same-tree baseline
  observation (command + verbatim output + tree SHA). A defect measuring
  not-reproducible at the baseline SHA closes as a regression guard with the
  observation recorded — no re-implementation (two-cell rule,
  `verification-completeness.md` §2).
- C2 — Lane-local verification: owning-package test families only; the full
  suite is CI's job (`gitflow-lane-protocol.md` §8).
- C3 — TDD mode (`quality.yaml` `constitution.development_mode: tdd`): every
  characterization test is authored RED-first and observed before its fix.
- C4 — Defect (8) acceptance requires the race detector and repeated runs
  (`go test -count=5 -race`) over the owning package's fold family — a single
  green run does not close a concurrency defect.
- C5 — The REQ-DISPATCH-008 lock must be cross-process: two `moai memory
  fold` invocations are separate OS processes; an in-process mutex does not
  satisfy it.
- C6 — No time estimates anywhere in this SPEC's artifacts.

## Acceptance Criteria

Tier M: the criteria live in `acceptance.md` (13 ACs, AC-DI-001..013), each
binary-testable and anchored to a named RED test or a named characterization
test, in the two-cell form (RED-now observed on this tree; the milestone that
flips it green). ACs whose defect measures not-reproducible at the baseline
are classified regression-guard per C1.

## Out of Scope

### Out of Scope — review findings owned by other cards

- `TestReviewFindingWhitespaceLanding` (landing_patch-id) — card t1561, fix
  landed in PR #1795, observed PASS on this tree (81786284e).
- `TestReviewFindingBackgroundReceiptRecycling` (audit_receipt) — card
  t1562, in flight as PR #1803; its RED stays off this card's branch.
- pre_tool:1397 — card t1556, fix landed in PR #1783 (mapping-resolved in
  the card handoff; nothing claims it in scope).

### Out of Scope — lane-probe finding (deferred to t1596)

- `TestReviewFindingLaneProbe` (productionLaneFilesProbe: github-flow empty
  list + whitespace path splitting) maps to no defect in this card's 8.
  Deferred to t1596 per `decision-index.md` Q1 (DEFAULT-APPLIED). The RED
  evidence (all 3 subtests FAIL on this tree, 2026-10-08) and the test file
  remain archived for t1596's intake — the deferral records ownership, not
  the defect's existence.
- Inheritance is durable on THIS branch: the deferred test source and its
  mapping live in the tracked `deferred/` directory beside this spec
  (`deferred/lane-probe_test.go.txt`, `deferred/README.md` — owner card
  t1596, intake path documented there). The deferring card cannot edit
  t1596's queue body (queue mutation is prohibited for factory lanes and
  their agents), so this tracked record is the two-sided side this card can
  land.

### Out of Scope — t1596-class issuance findings

- `TestReviewFindingExplicitCandidateFiles` and
  `TestReviewFindingDryRunFactoryMigration` are t1596 scope (both observed
  RED on this tree; evidence archived in the overlay record).

### Out of Scope — full-suite verification and lock-mechanism design

- Running the full `go test ./...` suite is CI's job; verification here is
  lane-local (C2).
- The concrete cross-process lock mechanism for defect (8) (flock vs lockfile
  vs write-protocol shape) is run-phase design; this SPEC pins the behavior
  contract only.

## References

- `.moai/reports/t1595/handoff-to-card-session.md` — defect→RED-test mapping
  (§3), card defect summary (§4) (local-only).
- `.moai/reports/t1595/lane-wait-20261008.md` — dispatch-mismatch
  adjudication t1589→t1595, lease history (local-only).
- `.moai/reports/t1595/overlay/` — overlay test assets + observation records
  relayed from `/tmp/t1498-review-overlay/` (local-only; `.moai/reports/*` is
  gitignored by operator directive 2026-09-14).
- Cards: t1498 (review round), t1533 (landed repairs the baseline confirmed),
  t1561, t1562, t1596 (deferred owners).
- Rules: `verification-claim-integrity.md`, `verification-completeness.md`,
  `gitflow-lane-protocol.md` §8.
