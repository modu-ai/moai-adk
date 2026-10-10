---
id: SPEC-FACTORY-RECORD-REPAIR-001
title: "Factory record repair: lease renewal, one-shot recovery, re-entry edges, and the interim empty-stage guard"
version: "0.1.0"
status: draft
created: 2026-10-10
updated: 2026-10-10
author: manager-spec (card t1621)
priority: P0
phase: "v3.2.0 target"
module: "internal/homestate"
lifecycle: spec-anchored
tags: "factory-record, lease, renew, recovery, transition-table, card-t1621"
tier: M
depends_on: [SPEC-FACTORY-RECORD-001]
related_specs: [SPEC-FACTORY-ATOMIC-LEASE-001, SPEC-FACTORY-SELF-DISPATCH-001]
---

# SPEC: Factory record repair (first batch)

## HISTORY

- 0.1.0 — 2026-10-10 — Plan-phase artifacts authored for card t1621 (Class C, Tier M), on branch `WT-factory-record-fix` from HEAD `2aab5f797`. Scope is the first batch only (card brief R1 to R6). Evidence: lens-4 (factory evidence and leases) defects 1, 7, 9, and 27 and the interim part of defect 8; lens-2 (audit verdict engine) D1, marked P0; leader decision records `d-20261009T235640Z-3653`, `d-20261009T180009Z-ae0f`, and `d-20261009T182157Z-8ccf`. Priority P0 because the batch carries lens-4 defect 1 and lens-2 D1, both marked P0. The brief's defect numbers were checked against the lens-4 table, and the corrections are recorded in `plan.md` §B. Requirement count 15, criterion count 16 (Tier M ceilings 16 / 16).

## Background

The factory record (`factory.db`) keeps one row per card and enforces a version-checked state machine. Five defects in that record block this batch:

- **Lease expiry.** A card lease lasts 15 minutes from acquisition and is extended only when a stage transition succeeds. An expired lease is reclaimed only when a later call reaches the card, and that call commits the reclaim and then returns an error (lens-4 defect 1). The lane watchdog never reads the lease (lens-4 defect 27). The next verb refuses a holder's own expired row as "in flight" (lens-4 defect 9).
- **No re-entry from sync audit.** The state machine has no edge from sync audit back to run. Decision `d-20261009T235640Z-3653` records that the state machine "has no sync to run edge". A card whose own code fails sync audit can only be split or held.
- **No re-stamp edge.** A plan-audit or sync-audit entry whose recorded evidence hash no longer matches the verdict has no exit except a verdict that matches the stored hash (lens-4 defect 7). Decision `d-20261009T180009Z-ae0f` records that "no re-stamp edge exists".
- **Empty-stage bypass.** The edge from an empty stage to run admits any card, including a card that carries a SPEC id, without a plan verdict (lens-2 D1, P0). A new run row is born with an empty stage, so the route also crosses runs (lens-4 §4 and defect 8).
- **Unstamped lease.** A lease does not record the card's worktree path. Decision `d-20261009T182157Z-8ccf` records "a record field the lease did not stamp" on a card whose worktree path was empty.

The verdict file format and the ceiling engine belong to t1618 and are not changed by this SPEC.

## Requirements

The requirements use GEARS notation. Each requirement maps to at least one criterion in `acceptance.md`. The criteria are numbered AC-FRR-001 to AC-FRR-016, and the full mapping is in `plan.md` §E.

### R1: Lease renewal (lens-4 defects 1 and 27)

- **REQ-FRR-001** When the lane session that holds a live card lease runs the factory renew verb for that card, the factory record shall extend the lease to the configured lease duration counted from the renewal, and the verb shall report the new expiry time.
- **REQ-FRR-002** The factory record shall expose the renewal as an MCP tool whose success and refusal results match the factory renew verb for the same card and caller.
- **REQ-FRR-003** The factory record shall take the card lease duration from a configuration key whose default value is 15 minutes.
- **REQ-FRR-004** When a renewal is requested for a lease that has already expired, the factory record shall refuse the renewal and leave the card's state, version, and event history unchanged.
- **REQ-FRR-005** When the lane stall watchdog starts a pass, the watchdog shall renew its own card lease before it measures lane progress.

### R2: One-shot recovery of the holder's own expired lease (lens-4 defect 9)

- **REQ-FRR-006** When the holder of an expired card lease requests a transition through the factory stage verb, or through the factory next verb where that verb is reachable, the factory record shall collect the expired lease, re-lease the card to the same holder, and apply the requested transition in one step, so that a refused transition leaves the expired row unchanged.
- **REQ-FRR-007** When a card lease is refused as expired, the refusal shall name the recovery command that the holder runs to re-lease the card.

### R3: Sync-audit to run re-entry (decision `d-20261009T235640Z-3653`)

- **REQ-FRR-008** When the leader records a decision, citing the sync-audit verdict file, that the blocking defects named by a sync-audit FAIL are in the card's own code, the factory record shall move the card from sync audit back to run.
- **REQ-FRR-009** When a sync-audit to run transition is requested without a verdict that meets REQ-FRR-008, the factory record shall refuse the transition and leave the card unchanged.

### R4: Evidence re-stamp by the holder (lens-4 defect 7; decision `d-20261009T180009Z-ae0f`)

- **REQ-FRR-010** When the lease holder requests a re-audit entry at plan audit or sync audit with a new evidence hash, the factory record shall update the card's evidence hash and keep the card in the same stage, under the same lease, with its event history intact.

### R5: Worktree path recorded at lease (decision `d-20261009T182157Z-8ccf`)

- **REQ-FRR-011** When a lane acquires the lease on a card, the factory record shall record the card's worktree path, and shall refuse the lease when a different worktree path is already recorded for that card.

### R6: Interim guard on the empty-stage to run edge (lens-2 D1; lens-4 defect 8)

- **REQ-FRR-012** When a card with a non-empty SPEC id is moved from an empty stage to run, the factory record shall refuse the transition unless an admitted plan verdict is recorded for the card. The operator branch is not part of this SPEC, because no operator edge carrier exists (plan.md D2).
- **REQ-FRR-013** The factory record shall leave unchanged the empty-stage to run transition of a card whose SPEC id is empty, and the sync-audit to run transition governed by REQ-FRR-008.

### Cross-cutting

- **REQ-FRR-014** When a caller that does not hold the lease requests a renewal or a recovery of an expired lease, the factory record shall refuse the request and leave the card unchanged.
- **REQ-FRR-015** Where the card lease duration is configured, the factory record shall keep the queue claim lease duration unchanged.

## Exclusions (What NOT to Build)

The following items are out of scope for this SPEC.

### Out of Scope — Named follow-ups

- Run migration (lens-4 defect 8, minimal fix): moving a card's stage, evidence, SPEC id, decisions, and tree into a new run row; this SPEC adds only the interim guard of REQ-FRR-012.
- SPEC id immutability after picked (lens-4 defect 5): changing or setting the SPEC id while a lease is held; cards with an empty SPEC id keep their current behavior under REQ-FRR-013.
- Kickoff audit-ready key format (lens-4 defect 6): accepting the documented audit-ready line form.
- Audit re-entry for non-holders (lens-4 defect 7, non-holder part): re-audit by anyone other than the lease holder; REQ-FRR-010 is holder-only.
- Stage kickoff commit-then-error (lens-4 defect 4): the stage-to-kickoff path that commits the transition and then returns an error.
- Leased-to-T3 bypass through `stage <card> leased` (lens-4 defect 12): the stage verb opening the assignment-to-lease edge outside the hold, serial, and quota gates.
- Needs-decision and failed producers (lens-4 defect 16): verbs that enter the needs-decision or failed states.
- Leader reassign and collect (lens-4 defect 17): leader reassignment and collection of expired cards.
- Abandon lease residue (lens-4 defect 18): lease columns left on abandoned rows.
- Reclassify (lens-4 defect 15): a reclassification verb for queued cards.
- Stage help evidence table (lens-4 §7 case (d), no numbered defect): the stage verb's help text and its artifact-absent message.
- Drift table (lens-4 defect 23): help and description drift across documents and help text.
- Broker defects (lens-4 defects 3, 11, and 21): message handover on a generation change, token rotation with null lists, and the byte-based truncation rule.
- Lens-4 defects 10, 13, 14, 19, 20, 22, 24, 25, and 26 are not addressed by this SPEC and have no follow-up assigned yet.

### Out of Scope — Adjacent cards and remote landing

- t1616 (lane-13): the done-transition opening and the local main landing.
- t1618 (not started): the ceiling engine and the verdict file format; this SPEC does not change the verdict parser.
- Remote push, pull requests, and the remote done edge (lens-4 defect 2): this repository lands locally, with no push and no pull request.

## Residual risk

- A card with an empty SPEC id can still reach run through the empty-stage edge, because REQ-FRR-013 keeps that route unchanged until defect 5 and the run migration are addressed.
- REQ-FRR-008 uses the leader's recorded own-scope decision as its discriminator, so it needs no verdict field (plan.md § 검증된 전제; D3).
- The standing awaken cadence of 20 minutes is longer than the 15-minute lease, so a renewal at each watchdog pass alone does not keep a lease alive between passes (plan.md § 검증된 전제, item (h)).
