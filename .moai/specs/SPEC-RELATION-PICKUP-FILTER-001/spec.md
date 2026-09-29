---
id: SPEC-RELATION-PICKUP-FILTER-001
title: "blocks/depends relations consumed as a self-dispatch pickup filter — relation-blocked cards excluded from todo --auto pickup, and a cycle guard on todo relate"
version: "0.1.0"
status: completed
created: 2026-09-29
updated: 2026-09-30
author: manager-spec (card t1343)
priority: P2
phase: "v3.2.0 target"
module: "internal/cli, internal/kanban"
lifecycle: spec-anchored
tags: "todo, relations, blocks, depends, pickup, self-dispatch, cycle-guard, auto-dispatch, card-t1343"
tier: S
card: t1343
related_specs: [SPEC-TODO-ANALYSIS-001, SPEC-TODO-HOLD-STATE-001, SPEC-TODO-CLASSIFY-DISPATCH-001]
---

# SPEC: blocks/depends relations as a self-dispatch pickup filter

## HISTORY

- 0.1.0 — 2026-09-29 — plan-phase artifact set authored (card t1343; worktree
  `.moai/worktrees/t1343`, branch `WT-relation-pickup-filter`, HEAD `113082295`).
  Tier S scope with a dispatcher-mandated `acceptance.md` (card t1343 gate
  requires release-blocking criteria naming command + expected output).

## §A Context

### A.1 Origin

Card t1343 (queued, `[보고서 P3·Medium]`) carries proposal P3 of
`.moai/reports/autonomy-bottleneck-proposal-20260929.html` (lead-session report,
2026-09-29; local-only file in the primary checkout):

> **P3 — 선행 관계를 배차 필터로** — blocks/depends가 남은 카드는 픽업 후보에서
> 제외 · ⑧ 선행 위반 배차 원천 차단 · 범위 S

and the same report's §5.2 candidate 3:

> 선행관계 소비 — 자가 배차 후보 필터에 "미해결 depends/blocks 제외" 한 줄 —
> 순환 관계 가드 포함 (S)

### A.2 Verified findings (this tree, HEAD `113082295`)

**F-1 — the relations are record-only by design.**
`internal/kanban/backlog_store.go:139-148` (card t1309 landed them at this
range; the card text's `:135-138` citation predates a comment growth):

```go
// BacklogRelationBlocks records that the subject must land before the
// related card can proceed (card t1309). Record-only, like every agent
// relation: no scheduler, dispatcher, or self-dispatch path reads it —
// whether factory self-dispatch ever consults the blocks graph is a
// separate adjudication (t1240), and until that lands the record exists
// for the operator to read.
BacklogRelationBlocks = "blocks"
// BacklogRelationDepends is the inverse spelling of blocks: the subject
// waits on the related card. Same record-only posture.
BacklogRelationDepends = "depends"
```

`internal/cli/todo_relate.go:15-16` states the same posture from the verb side:
"Recording one causes nothing: no scheduler, dispatcher, or self-dispatch path
reads these findings".

**F-2 — the pickup path consults no relation.** The self-dispatch pickup
selection is `autoPickTargets` (`internal/cli/todo_auto.go:142`): a first loop
rescues `picked` cards whose owner measures dead, then a second loop appends
EVERY `queued` card unconditionally (`todo_auto.go:162-166`):

```go
for _, it := range rec.Items {
    if it.State == kanban.BacklogStateQueued {
        targets = append(targets, it)
    }
}
```

A non-test grep for `BacklogRelationBlocks|BacklogRelationDepends` hits only the
constant definitions and the `BacklogSemanticRelations` vocabulary list — no
dispatch, scheduler, or pickup consumer exists. **A queued card blocked by an
unresolved relation is a pickup candidate today.** Confirmed.

**F-3 — the relate write path has no cycle check.**
`runTodoRelate` (`internal/cli/todo_relate.go:60-94`) validates four things:
the relation is semantic, subject ≠ related, both cards exist, and the exact
{subject, related, relation, source} tuple is not already recorded
(`AppendFindingOnce`). Nothing walks the graph — `A depends B` followed by
`B depends A` records cleanly today.

**F-4 — resolution already exists in the archive path.**
`ArchiveCard` (`internal/kanban/backlog_store.go:308-341`) is what `done`
performs: the card AND every finding naming it move into the archive entry and
out of `rec.Findings`. When a predecessor lands, its sequencing findings leave
the live record with it — a blocked successor becomes eligible with **no
relation bookkeeping**. Drop does NOT move findings (a dropped card stays in
`Items` with `state: dropped`), and `RemoveFindingsNaming` has no production
caller.

**F-5 — a cycle guard already exists in a sibling subsystem.**
`ValidateGTDRelation` (`internal/kanban/gtd_relation.go:48-86`) refuses a
`depends_on` edge that can reach its subject (`gtd relation: dependency_cycle`)
over the GTD `gtd_relations` tables. That is the `moai gtd` capture/organize
subsystem — a DIFFERENT relation store from the `todo relate` findings. It is a
pattern precedent, not a consumer to couple to.

### A.3 The two relation systems (do not conflate)

| | `todo relate` findings | `moai gtd` relations |
|---|---|---|
| Store | `BacklogFinding` in the queue record | `gtd_relations` SQL tables |
| Sequencing values | `blocks`, `depends` | `depends_on` |
| Cycle guard | **none — this SPEC adds it** | exists (`ValidateGTDRelation`) |
| Pickup consumer | **none — this SPEC adds it** | none |

### A.4 Merge-vs-separate decision: t1343 stands alone (dispatch-mandated)

Investigated: (1) no `SPEC-*` directory or file under `.moai/specs/` names
t1338; (2) no `.moai/reports/` hit except coincidental substrings inside test
temp-dir names in `t60-review-diff.txt` / `t60/web-preexisting-failures.txt`;
(3) the originating proposal attaches no card id to P3 (it names t1332, t1339,
t1341 elsewhere); (4) the live queue (`~/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db`,
read-only query) holds t1338 as `queued` — "팩토리 레인 자율 운영 완성 카드",
Class C, Tier L candidate, precondition "선행: t1240 병합", whose FOUR enumerated
pieces are messaging-failure fallback → `todo --auto` self-service mode,
t1332 classification-meta consumption, t1241 autonomous-merge conditions, and
origin-push-confirmed WT disposal. The blocks/depends relation filter is NOT one
of them.

**Decision: t1343 stands alone.** Rationale:

1. t1338 has no landed SPEC or design artifact — there is nothing to merge
   into, and its own plan has not run (queued behind the t1240 merge).
2. Absorption couples a Class S quick-rotation card to a Tier L candidate
   gated on another card's merge — the opposite of the dispatch's "small
   scope, quick rotation" mandate.
3. The convergence point is recorded, not lost: t1338 piece 1 routes fallback
   lanes INTO `todo --auto` — the exact path this SPEC makes relation-aware.
   When t1338's plan runs, the filter is an existing pickup-predicate surface
   it consumes for free.
4. Factory-lease-side consumption (`moai factory next`) stays the separate
   adjudication the code already names (t1240 — see F-1's comment and §D).

### A.5 Coordination frame

Card t1343, Class S, procedure `plan → run → sync`, cycle_type=tdd
(`quality.yaml` `development_mode: tdd`). Plan-phase authored in worktree
`.moai/worktrees/t1343` on `WT-relation-pickup-filter`. No code changes at
plan-phase; the live operator queue is never mutated by this card.

## §B Decisions

### B.1 "Unresolved" = the sequencing finding still exists in the live record

The filter's predicate is the finding's EXISTENCE, not a card-state scan:

- predecessor `done` → `ArchiveCard` moved the finding out of `rec.Findings`
  (F-4) → resolved automatically;
- predecessor dropped or held → the finding persists → the card stays excluded
  (conservative stall; escape hatch is `todo unrelate <index>`, and `undrop`
  of a dropped predecessor keeps the record honest either way).

This reuses existing data with zero schema change and makes "미해결" exactly
what the proposal wrote: the record is still there.

### B.2 The filter governs the queued-pickup arm only

`autoPickTargets`' dead-owner rescue arm (a `picked` card whose owner measured
dead) is a takeover of already-started work, not a pickup. Gating it on
relations would strand in-flight SPECs. The rescue arm is pinned unchanged
(REQ-RPF-006).

### B.3 Cycle guard = waits-on reachability, mirroring the GTD precedent

Normalize both sequencing values to one directed "waits-on" edge:
`depends {S,R}` → S waits on R (edge S→R); `blocks {S,R}` → R waits on S
(edge R→S). A candidate is refused when its wait target already reaches its
waiter through recorded waits-on edges. The check runs in the relate write path
BEFORE `AppendFindingOnce`. Same-pair opposite spellings (`A depends B` then
`B blocks A`) encode the same edge, form no cycle, and are not refused.

### B.4 Skipped cards are labelled non-findings, exit stays 0

A relation-blocked skip prints one labelled line per skipped card (card id,
relation, blocking predecessor id) — the same labelled non-finding vocabulary
the cycle already uses for degraded liveness and live-owner skips. An empty
pickup set remains a labelled outcome, never an error (the no-eligible-card
path at `todo_auto.go:218-219` keeps its exit-0 contract).

### B.5 The record-only doctrine comments are updated in the same change

F-1's comments (`backlog_store.go:139-148`, `todo_relate.go:15-18`) assert
"no scheduler, dispatcher, or self-dispatch path reads it". After this SPEC
that sentence is false and must not survive as doctrine text: the run updates
both comment blocks to name `autoPickTargets` as the consuming path and the
cycle guard as the write-time check. Stale doctrine is the defect this card
must not ship.

## §C Requirements

Verification layer: `acceptance.md` (dispatcher-mandated for card t1343 —
release-blocking criteria name command + expected output). Requirement layer
below is GEARS.

- **REQ-RPF-001** (Ubiquitous) — The `todo --auto` pickup selection shall treat
  a queued card as ineligible for pickup while a recorded `blocks` or `depends`
  finding names that card as the blocked side in the live queue record.

- **REQ-RPF-002** (Ubiquitous) — The pickup selection shall read the blocked
  side of a sequencing finding as the finding's `related_id` when the relation
  is `blocks` and as the finding's `subject_id` when the relation is `depends`.

- **REQ-RPF-003** (When) — **When** a predecessor card named by a sequencing
  finding is archived through `done`, the finding shall have left the live
  record with the archive entry, and the previously blocked card shall be a
  pickup candidate again with no relation bookkeeping.

- **REQ-RPF-004** (When) — **When** the `todo --auto` cycle skips a queued card
  because of an unresolved sequencing relation, the cycle shall print one
  labelled non-finding per skipped card naming the card id, the relation, and
  the blocking predecessor id, and the cycle's exit status shall remain 0.

- **REQ-RPF-005** (When) — **When** a candidate `blocks` or `depends` relation
  would close a directed cycle in the waits-on graph formed by the recorded
  sequencing findings, `todo relate` shall refuse the write with an error
  naming both endpoints, and the queue record shall remain unchanged.

- **REQ-RPF-006** (Unwanted) — The relation filter shall not exclude a `picked`
  card from the dead-owner rescue arm, and shall not consume any relation value
  other than `blocks` and `depends`.

- **REQ-RPF-007** (Ubiquitous) — The filter and the cycle guard shall consume
  and write only the existing `BacklogFinding` records through the existing
  `todo relate` / `todo unrelate` verbs — no new card field, no schema change,
  and no read of the GTD relation tables (`gtd_relations`).

### C.1 Traceability

REQ-RPF-001 → AC-RPF-001 · REQ-RPF-002 → AC-RPF-002 · REQ-RPF-003 → AC-RPF-003
· REQ-RPF-004 → AC-RPF-004 · REQ-RPF-005 → AC-RPF-005 + AC-RPF-006 ·
REQ-RPF-006 → AC-RPF-007 · REQ-RPF-007 → AC-RPF-007 (AC bodies and
`**Covers**: maps …` clauses in `acceptance.md`).

## §D Out of Scope

### Out of Scope — factory-lease consumption (`moai factory next`)

- The factory-record lease path (`factoryNextSelectAndLease`) is a separate
  selection surface over the factory record; whether IT consults the blocks
  graph is the adjudication already named in the code (card t1240). This SPEC
  does not touch `factory next`, `stage`, or `complete`.

### Out of Scope — t1338 absorption

- No piece of t1343 is implemented inside t1338's future scope; the
  convergence is the recorded §A.4 note. t1338's own four pieces (fallback
  mode switch, classification consumption, merge conditions, WT disposal) are
  untouched here.

### Out of Scope — display surfaces

- `todo list`, `todo why`, the web relations rendering, and Jev output are
  display-only and unchanged. P3's scope is the pickup predicate, not the
  render.

### Out of Scope — relation vocabulary and dedup policy

- Cross-vocabulary duplicate orderings (same pair, opposite spelling), finding
  merge/split, `RemoveFindingsNaming`'s uncalled status, and any `contains`/
  `absorbs`/`replaces`/`conflicts` semantics stay exactly as they are.

## §G Gaps and Residual Risks

- **R-1** — A dropped or held predecessor wedges its successor by design
  (B.1). The wedge is visible (labelled non-findings every cycle) and has an
  operator escape (`todo unrelate`), but nothing AUTOMATIC surfaces "your
  predecessor was dropped, consider unrelating". Accepted for Tier S; a
  future card may add hint text.
- **R-2** — The filter sees only `rec.Findings` of the live record. A record
  that lost findings through legacy merge paths would silently unblock — the
  same trust every other findings consumer already extends.
- **R-3** — `factory next` remains relation-blind until t1240's adjudication;
  a blocked card CAN still be leased through that path today. This is the
  documented boundary, not an oversight.
