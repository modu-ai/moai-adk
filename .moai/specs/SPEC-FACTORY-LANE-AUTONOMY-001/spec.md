---
id: SPEC-FACTORY-LANE-AUTONOMY-001
title: "Factory lane autonomy completion — messaging-fallback self-service, classified pickup, lane-direct merge, post-push disposal"
version: "0.1.0"
status: draft
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli/factory, internal/cli/worktree, internal/sessionmsg, internal/kanban"
lifecycle: spec-anchored
tags: "factory, lane-autonomy, umbrella, messaging-fallback, todo-auto, classification, merge-window, worktree-disposal, card-t1338"
tier: L
card: t1338
related_specs: [SPEC-FACTORY-SELF-DISPATCH-001, SPEC-FACTORY-CONTROLLER-001, SPEC-MANAGER-TODO-001, SPEC-FACTORY-RECORD-001]
---

# SPEC-FACTORY-LANE-AUTONOMY-001 — Factory lane autonomy completion (card t1338, umbrella)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-29 | manager-spec | Initial plan-phase draft (card t1338, factory lane autonomy completion). UMBRELLA SPEC: 4 uncovered fragments layered on 4 boundary cards (t1240 / t1241 / t1332 / t1330 + landed t1306, t1239) WITHOUT overlapping them. Run-entry predecessor: t1240 develop merge (expressed as plan milestone gate, not `depends_on` — see design.md D8). Baseline: worktree `.moai/worktrees/t1338`, branch `WT-lane-autonomy-umbrella`, base develop `145c3d98c`. 16 REQ, 17 AC. |

## §A. Background and Motivation

The operator directed on 2026-09-29 (card t1338) that factory-lane autonomy be COMPLETED: the four
remaining uncovered fragments of a fully autonomous `-f` lane — messaging-unavailability fallback,
classified self-pickup, lane-direct merge under the autonomous-merge conditions, and post-push
worktree disposal — layered as an umbrella on top of the boundary cards that own the adjacent
surfaces:

- **t1240 / SPEC-FACTORY-SELF-DISPATCH-001** (F2 self-lease) — complete on branch `WT-factory-self-dispatch` (tip `d43e50bb3`, merge-ready, NOT yet in develop). Landed there: `factory next [--wait] [--wait-bound]`, `factory stage`/`complete`. NOT in this tree (verified).
- **t1241 / SPEC-FACTORY-CONTROLLER-001** (F3 controller) — M1a landed-on-branch `WT-factory-controller`; F3's merge automation (M3) is NOT built anywhere. Its card text carries the autonomous-merge condition triple: "sync-audit PASS · 충돌 없음 · HEAD^{tree}=HEAD^2^{tree} → local develop merge". NOT in this tree (verified).
- **t1332** (Jev classification, queued) — owns classification metadata PRODUCTION. Does not exist as a SPEC yet.
- **t1330** (join repair, queued) — owns dispatch-join defects; adjacent to but disjoint from fragment 1.
- **Landed foundations this SPEC builds on**: `todo --auto` foreman (t1306 / SPEC-MANAGER-TODO-001, merge `6af4d4ca3`) and the factory F1 record layer (t1239 / SPEC-FACTORY-RECORD-001, merge `ed506740b`).

The umbrella discipline is the point: each fragment below consumes an interface the boundary card
owns and never re-implements the producer side. The queue remains the delegation channel; operator
authority over card admission never shrinks.

## §B. Requirements

### §B.1 Fragment 1 — Messaging-unavailable fallback to self-service mode

- **REQ-FLA-001** (Event-driven): **When** the lane's messaging-availability probe reports the cross-session messaging channel unavailable — no registered lead peer reachable, or the lead's heartbeat age exceeds the configured availability bound — the factory lane shall declare messaging fallback and switch to `/moai:todo --auto` self-service pickup for subsequently eligible cards.
- **REQ-FLA-002** (Event-driven): **When** a directed lead request (a dispatch nudge or a window-grant announcement) goes unanswered past the bounded no-response timer, the lane shall record a no-response observation and treat the channel as unavailable for the remainder of the current bound period.
- **REQ-FLA-003** (Ubiquitous): The lane shall emit exactly one recorded fallback-transition event per mode switch, carrying the lane id, the trigger kind (`channel-unavailable` | `no-response`), a timestamp, and the card in progress at switch time, into a queryable transition log (fallback-transition observability constraint).
- **REQ-FLA-004** (Event-driven): **When** a lane resumes a card already in `picked` state whose previous owner is determined stalled — the stall determination joining t1241's re-alert/stall interface — the resuming lane shall first read the card's `progress.md` and its recorded evidence, then continue from the recorded phase (resume semantics, not a silent restart).
- **REQ-FLA-005** (Unwanted): A resuming lane shall not discard or overwrite the previous owner's recorded progress or evidence; resumption shall append its own records alongside them.

### §B.2 Fragment 2 — Classified self-pickup (t1332 metadata consumption)

- **REQ-FLA-006** (State-driven): **While** a queued card carries t1332's classification metadata (priority and execution axis), the lane's self-service pickup shall consume it: a sequential card shall be picked only while no other lane holds or is actively picking a card of the same sequential group, and a parallel card shall be pickable by multiple lanes concurrently.
- **REQ-FLA-007** (Where): **Where** a queued card carries no classification metadata — t1332 not yet landed, or the card left unclassified — the lane's pickup shall fall back to the current operator-picked single-dispatch behavior and shall not autonomously multi-pick that card.
- **REQ-FLA-008** (Ubiquitous): This SPEC's pickup contract shall define CONSUMPTION only: the classification field schema (names, types, defaults) remains t1332's producer concern, and the consumer shall tolerate absent or unknown metadata fields without error.

### §B.3 Fragment 3 — Lane-direct merge under the autonomous-merge conditions

- **REQ-FLA-009** (Event-driven): **When** a lane seeks to merge its own card branch into local develop, the lane shall verify the t1241 autonomous-merge condition triple — sync-audit PASS, the merge carries no unresolved conflict, and tree identity `HEAD^{tree} == HEAD^2^{tree}` — BEFORE entering the integration window.
- **REQ-FLA-010** (State-driven): **While** the integration window is held by another lane (`moai integration acquire` reports a live holder), a lane shall wait for the recorded hold to be released before its own acquire; the integration-window serial invariant is unchanged.
- **REQ-FLA-011** (Unwanted): A lane shall not merge into develop outside the `moai integration acquire`/`release` window and shall not bypass the condition triple of REQ-FLA-009.

### §B.4 Fragment 4 — Post-push worktree disposal

- **REQ-FLA-012** (Event-driven): **When** card worktree disposal is requested, the disposal path shall run the machine check — `git fetch origin develop` followed by `git rev-list --count --left-right origin/develop...<merge-commit>` confirming the card's merge commit has reached origin — and shall refuse disposal while the check does not confirm the landing.
- **REQ-FLA-013** (Ubiquitous): The disposal path shall preserve both existing guards unchanged — the L1-tier refusal (`isL1WorktreePath`) and the anchored-session guard (`LiveAnchoredSessions` → ANCHORED_SESSIONS_PRESENT).
- **REQ-FLA-014** (Where): **Where** the machine check of REQ-FLA-012 confirms the origin landing, the disposal path shall permit unattended (`--auto`) disposal without a further operator approval round. CI-green is NOT part of the machine precondition (plan decision D4 — CI judgment remains lead-side).
- **REQ-FLA-015** (Unwanted): The disposal path shall not dispose a worktree whose card's merge commit is not confirmed on origin — the pre-merge-disposal-prohibition invariant is unchanged.

### §B.5 Cross-cutting invariants

- **REQ-FLA-016** (Ubiquitous): The queue shall remain the sole delegation channel — the fallback self-service, classified pickup, lane-direct merge, and auto-disposal paths shall not add any messaging-based delegation route, and card admission shall remain operator-owned (operator pick or an operator-authorized standing source only).

## §C. Constraints

1. **Queue-is-the-channel doctrine preserved** (card constraint; REQ-FLA-016). The delegation
   channel is the queue on disk; a cross-session message stays a nudge, never the delegation.
2. **Operator-gate reduction scope MUST be explicit** (card constraint). The reduction table lives
   in `plan.md` § Operator-Gate Reduction Table; card admission and Implementation Kickoff
   Approval never shrink.
3. **Fallback-transition observability** (card constraint; REQ-FLA-003): every mode switch is a
   recorded, queryable event.
4. **Run-entry predecessor**: run-phase entry is gated on t1240's develop merge (plan.md §F
   Milestone Gate M0). NOT expressed as `depends_on` — reasoning in design.md D8.
5. **Umbrella non-overlap**: fragments consume boundary-card interfaces (t1240 self-lease verbs,
   t1241 condition triple + stall interface, t1332 metadata, t1306 --auto foreman) and never
   re-implement the producer side (REQ-FLA-008, REQ-FLA-011, §F Out of Scope).
6. **Tier L** artifact set (5 artifacts + progress.md); REQ/AC ceiling 25 each (16 REQ, 17 AC used).

## §D. Interface Boundaries (consumed, not built)

| Interface | Owner | What this SPEC consumes | What it must NOT build |
|-----------|-------|------------------------|------------------------|
| `factory next [--wait]`, `factory stage/complete` | t1240 (F2) | The self-lease pickup verbs self-service mode rides on | The lease model itself |
| Autonomous-merge condition triple + stall/re-alert determination | t1241 (F3) | The triple as a lane-executed check sequence; stall as the resumption trigger | F3 M3 controller machinery (write-ahead events, trial merge) |
| Classification metadata (priority, exec axis) | t1332 | The consumption rules | The field schema or its production |
| `/moai:todo --auto` foreman | t1306 | The foreman cycle as the fallback pickup executor | Evidence-read contract, liveness, --auto-wait semantics |
| `moai integration acquire/release` | F1 record layer (t1239, landed) | The window as the only merge serialization point | A second serialization mechanism |
| `worktree done` guards | `internal/cli/worktree/done.go` | The path as the disposal act | A parallel disposal verb that bypasses the guards |

## §E. Success Criteria Summary

All 17 ACs in `acceptance.md` pass; zero overlap with any boundary card's owned surface (verified
by the §F exclusions + the interface-boundary table above); every quantitative claim in the
artifact set carries a source citation from the carried reconnaissance.

## §F. Exclusions

### Out of Scope — F3 merge-automation controller machinery (t1241's M3)

- No write-ahead start events, no trial merges, no controller tick/hook loop — this SPEC applies the condition triple as a lane-executed check sequence only; the controller itself remains t1241's.

### Out of Scope — classification metadata production (t1332)

- No field-schema definition, no producer code, no queue field migration for priority/exec-axis — the consumer reads t1332's future output and tolerates its absence (REQ-FLA-007/008).

### Out of Scope — `todo --auto` foreman internals (t1306, landed)

- No change to the evidence-read completion contract (`.moai/reports/<card>/evidence.md`), liveness detection (`autoLiveness`), or `--auto-wait` semantics.

### Out of Scope — card admission autonomy

- No auto-admission of operator-unpicked cards; the queue stays the channel and the operator (or an operator-authorized standing source) remains the only producer of admissible cards.

### Out of Scope — CI verdict consumption in disposal

- CI-green is NOT a machine precondition of disposal (plan decision D4); CI judgment remains lead-side per the factory doctrine.

### Out of Scope — `moai todo add` argument parsing

- The `-f lane` body-prefix flag-parse defect discovered during planning (research.md R-D1) is a dispatch/registration surface belonging to t1330's arg-parsing scope, not this SPEC.

## §G. Cross-References

- `plan.md` — milestones, pre-flight, operator-gate reduction table, run-entry gate.
- `acceptance.md` — AC matrix (AC-FLA-001..017), edge cases, quality gates, DoD.
- `design.md` — D1..D8 design decisions with rejected alternatives.
- `research.md` — reconnaissance evidence (R1..R14) and discovery note R-D1.
