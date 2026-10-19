---
id: SPEC-FACTORY-DEFAULT-MODE-001
title: "Absent card classification defaults to parallelizable, the decider-failure fail-safe stays serial, and the official `moai todo classify` reclassify verb"
version: "0.1.0"
status: in-progress
created: 2026-10-19
updated: 2026-10-19
author: manager-spec (card t1737)
priority: P0
phase: "v3.2.0 target"
module: "internal/cli, internal/factory"
lifecycle: spec-anchored
tags: "factory, classification, default-mode, parallelizable, serial, fail-safe, reclassify, todo-queue, card-t1737"
tier: M
card: t1737
depends_on: [SPEC-TODO-CLASSIFY-DISPATCH-001]
related_specs: [SPEC-FACTORY-CONFLICT-SLOTS-001, SPEC-TCD-LLM-DECIDER-001, SPEC-FACTORY-ATOMIC-LEASE-001, SPEC-TODO-HOLD-STATE-001]
---

# SPEC: Factory default classification mode and the official reclassify verb

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-10-19 | Initial plan-phase authoring (card t1737, operator ruling 2026-10-19, board `d-20261019T015400Z-02f4`), measured in worktree `.moai/worktrees/t1737`, branch `WT-spec-factory-default`, at HEAD `74ad197380`. All card premises verified against the tree: the absent-field derivation seam exists (`internal/factory/classification.go:148-153`, `EffectiveCardClassification`, the sole derivation seam per its MX:ANCHOR) and today derives mode `serial` for absent records (`DefaultCardClassification`, `classification.go:98-105`); the fail-safe branch and the absent-read derivation share that one function ("the two defaults are the same defaults by ruling", `classification.go:94-97`) — this SPEC splits them by operator decision; the factory record (`internal/homestate/card_record.go:100` `homestate.Card`) carries NO mode field, so factory-side mode reads derive from the queue at read time (`internal/cli/factory_card.go:839-842`, `:3093-3098`, `:3155-3160`); no `moai todo classify` subcommand exists (the only `classify` wiring is the add path's decider seam, `internal/cli/todo_classify.go`); the fallback notice reads `mode=serial` (`todo_classify.go:41`). The card's REQ 2 + AC 2 figures are the parent split proposal's C3 minimum (`t1714` split-proposal §1, authored before the operator's scope extension); the dispatched scope adds the reclassify verb and the fallback-path closure, so this SPEC carries 7 REQ / 10 AC, inside the Tier M ceiling of 16/16. Predecessor C1 (t1735, SPEC-FACTORY-CONFLICT-DOMAIN-001) is skipped by operator exception; the safety rationale is recorded in plan.md §A.2. No card premise was falsified. |

> **Provenance discipline.** Every `file:line` citation was measured at HEAD `74ad197380` in this
> worktree. The t1714 split proposal is read-only reference material under another card's tree.

## §A Context

### A.1 The gap: the conservative default wedges the queue it was meant to protect

SPEC-TODO-CLASSIFY-DISPATCH-001 made `serial` the default on BOTH axes that lack an explicit
judgment: a queue item recorded with no classification field derives `serial` on read
(`classification.go:148-153` returning `DefaultCardClassification()`, `:98-105`), and a card
admitted without a supplied judgment records `serial` through `DefaultCardDecider`
(`classification.go:253-260`). The ruling that unified them (parent HISTORY v0.2.0 OD-3 REVISED
and v0.3.0, 2026-09-29) was deliberately conservative: absent = unclassified = serial, so an
unclassified card participates in the serial-vs-serial mutual exclusivity of the dispatch layer.

Measured on 2026-10-19 (board `d-20261019T014804Z-a672`, operator instruction "무한 갱신 원인
찾아 해결"): 29 cards sat on the serial default, 6 of them waiting on the single serial slot, and
the lanes holding them renewed their leases in a loop — the conservative default had become the
queue's largest source of waiting. The operator ruled (board `d-20261019T015400Z-02f4`): cards
WITHOUT an explicit classification default to `parallelizable`; the decider-failure fail-safe
stays a separate `serial` value; and an official reclassify verb replaces direct database edits
so the leader can pin genuinely-serial cards by name. A companion standing rule (board
`d-20261019T020637Z-0c51`) already moved issuance to explicit `--classification-file`
judgments (decider=human), so the unclassified population is shrinking by policy — the default
flip covers the records that predate that discipline and any record that slips through.

### A.2 What exists today (measured surface)

- **One derivation seam** — `EffectiveCardClassification(it BacklogItem)` (`classification.go:148-153`):
  recorded judgment when present, `DefaultCardClassification()` when absent. Its MX:ANCHOR declares it
  the sole absent-field default derivation seam. `factoryQueueClassification`
  (`internal/cli/factory_card.go:345-360`) extends the same derivation to cards absent from the queue.
- **One default function, two call sites** — `DefaultCardClassification()` (`classification.go:98-105`)
  serves both the absent-read derivation and the fail-safe promotion in `todoClassifyInLock`
  (`todo_classify.go:81-89`), which prints exactly one fallback notice (`:41`, text names `mode=serial`).
- **The default decider** — `DefaultCardDecider` (`classification.go:253-260`) answers the default
  classification (mode `serial`) with reason "deterministic default: no classification judgment
  supplied". It is a judgment, not a failure.
- **Factory-side reads derive, never store** — `homestate.Card` (`card_record.go:100`) carries no mode
  field; `factoryCardView.Mode` is read from the queue snapshot (`factory_card.go:3093-3098`,
  `:3155-3160`), and every dispatch arm (selection `:839-842`, serial-in-flight `:324-341` with the
  mode test at `:335`, nominated validation `:1397`/`:1406`) reads the queue once per pass.
- **The serial slot** — a card holds the slot while `factorySerialSlotHeld` (`factory_card.go:295-297`)
  AND its mode is `serial` (`:335`); a second serial card waits, parallelizable cards are unaffected.
- **The mutation lock** — `BacklogStore.Mutate` (`internal/factory/backlog_store.go:921-924`) holds the
  queue's cross-process lock across the whole read-modify-write; a refused mutation writes nothing and
  the file stays byte-identical (the hold-verb contract, `internal/cli/todo_hold.go`).
- **The command registry** — `internal/cli/todo.go:342` attaches the todo subcommands; no
  `classify` subcommand exists (verified: the only `classify` hits are the add-path seam files).
- **The classification store** — one nullable JSON column (`backlog_sqlite.go:457`, `:546`), written
  only through `CardClassificationValue` (validated on write); value sets are closed
  (`classification.go:33-69`): mode ∈ {serial, parallelizable}, decider ∈ {default, human, llm};
  `human` is already an accepted decider identity.

## §B Requirements

All requirements use GEARS notation. Requirement IDs are stable; no renumbering after publication.

- **REQ-FDM-001** — **While** a queue item carries no classification record — and **while** a
  factory-recorded card names no queue item — the absent-field classification derivation shall
  produce mode `parallelizable` with priority `normal`, blocked `false`, and decider `default`,
  and every machine consumer shall read the card's execution mode through that single derivation
  seam, never by testing the field for absence itself.

- **REQ-FDM-002** — **When** a card is admitted to the queue without a supplied judgment (no
  `--classification-file`), the default decider shall record mode `parallelizable` under decider
  identity `default`, with a reason that names the deterministic default.

- **REQ-FDM-003** — **When** the classification decider is unavailable or fails at admission, the
  add path shall promote the fail-safe classification — mode `serial`, decider `default` — and
  shall print exactly one stderr notice naming the fallback; admission shall never block on
  classification, and the fail-safe value shall remain `serial` regardless of the absent-read
  default. (The two defaults are separate values from this SPEC onward.)

- **REQ-FDM-004** — **When** an operator (or the leader speaking for one) runs
  `moai todo classify <id> --mode <serial|parallelizable> --reason <text>` against a queued or
  picked card, the verb shall record the human judgment — decider `human`, the given mode, the
  given reason verbatim, a fresh classified-at stamp — in exactly one locked queue write
  (one `Mutate` acquisition), leaving the card's priority and blocked fields and every other
  queue item untouched; and **when** the target card is in flight (picked) and the requested
  mode is `serial` while another distinct in-flight card holds the serial slot (the selection
  gate's own read: slot-held ∧ mode `serial`, `factory_card.go:324-340`), the verb shall
  refuse with a recovery-naming message and write nothing — a reclassification shall never
  leave two distinct in-flight cards in serial mode at once (round-2 D2: the unserialized
  overlap reproduced on the plan-close tree is foreclosed at the verb; decision-index Q6).

- **REQ-FDM-005** — **When** a card already recorded in the factory record is reclassified,
  every factory-side mode reader — selection, nominated validation, and the status view — shall
  observe the reclassified mode from its next pass with no second write: the queue is the only
  store carrying the mode, and the single locked write of REQ-FDM-004 is the synchronization point.

- **REQ-FDM-006** — **When** the classify verb is aimed at a card in any state other than queued
  or picked, at an id absent from the queue, with an empty `--reason`, or with a `--mode` outside
  the closed set, the verb shall refuse with a usage error and shall write nothing — the queue
  record stays byte-identical on every refusal.

- **REQ-FDM-007** — **While** a serial-mode card holds the serial slot, the factory shall not
  lease another serial-mode card; parallelizable cards shall remain leasable beside it — the
  exclusivity invariant is unchanged by the default flip, which widens the parallelizable
  population but never weakens serial ordering.

## §C Constraints

- C1 — The value sets stay CLOSED and validated in one place (`classification.go:33-69`); this
  SPEC adds no new mode or decider value. `human` already exists in the accepted decider set.
- C2 — The fail-safe direction is untouched in kind: an unavailable or failed decider promotes
  `serial`, because a wrongly-parallel fail-safe could run true-serial cards concurrently
  (parent HISTORY v0.2.0, OD-3 rationale — preserved).
- C3 — Recorded judgments are never rewritten by the default change: a card carrying an explicit
  record (including the ~29 cards recorded `serial` by the old default decider, identifiable by
  decider `default` ∧ mode `serial`) keeps its recorded judgment until a human reclassifies it
  through the verb. There is no batch migration.
- C4 — The SQLite layout is unchanged: the classification column exists (`backlog_sqlite.go:457`);
  no migration, no schema-version bump.
- C5 — Refused mutations leave the queue file byte-identical (the hold/edit-verb refusal
  contract); the classify verb inherits it.
- C6 — The `jev` decider identities stay refused on every product classification path
  (`classification.go:45-49`); the classify verb records decider `human` unconditionally and
  accepts no decider override.
- C7 — Exactly one fallback notice per fallback admission (`todo_classify.go:38-41`): the count
  is observable, and the notice text continues to name `mode=serial`.

## §D Acceptance Criteria

Acceptance criteria live in `acceptance.md` (Tier M): AC-FDM-001 … AC-FDM-011, each
Given-When-Then, mapped onto REQ-FDM-001 … REQ-FDM-007 with the refusal matrix (including the
serial-slot conflict refusal of REQ-FDM-004, AC-FDM-011) and the factory-side observation
covered explicitly.

## §E Exclusions and Out of Scope

### Out of Scope — classification axes other than mode

- The classify verb does not reclassify priority or blocked; it records mode, decider=human,
  reason, and the stamp only. An operator who needs a priority change keeps using the existing
  edit paths.

### Out of Scope — batch migration of recorded serial-default cards

- No tool rewrites the ~29 existing `decider=default ∧ mode=serial` records in bulk. The
  operator-decided path (board `d-20261019T015400Z-02f4`) is per-card leader reclassification
  with the new verb ("the leader reclassifies parallel-safe serial-default cards (first t1888,
  t1889) with the new verb"), and the standing issuance rule (board `d-20261019T020637Z-0c51`)
  prevents new arrivals on the fail-safe default.

### Out of Scope — conflict-domain slots and record-schema work (predecessor C1)

- Domain slots, the record-schema V7 migration, and domain measurement belong to
  SPEC-FACTORY-CONFLICT-DOMAIN-001 (t1735, skipped by operator exception). This SPEC performs no
  schema change (C4) and precludes nothing C1 may later add.
- SPEC-FACTORY-CONFLICT-SLOTS-001, carried in frontmatter `related_specs` as lineage context
  (parent index), has no committed SPEC on this lineage: the proposed slots split was never
  executed here — the naming artifact is the t1714 split proposal (read-only reference material
  under another card's tree). The reference is disposition-only lineage context, not a
  dependency; this SPEC requires nothing from it (round-2 D3).

### Out of Scope — decider automation changes

- No change to the LLM decider (SPEC-TCD-LLM-DECIDER-001 surface), the `MOAI_TODO_DECIDER`
  operator switch, or the `--classification-file` transport. The verb is a human-judgment
  recording path, not a decider implementation.

## §F Success Criteria (summary)

1. An absent classification reads `parallelizable` on every machine path (AC-FDM-001, -002);
   a fresh no-judgment admission records it (AC-FDM-003).
2. The fail-safe stays `serial` with exactly one notice (AC-FDM-004).
3. The verb records a human judgment in one locked write, refuses everything else
   byte-identically — including an in-flight serial reclassification that would overlap a
   live serial slot holder (AC-FDM-011) — and factory-side readers observe it
   (AC-FDM-005 … -008, -011).
4. Serial exclusivity and recorded-judgment stability are regression-guarded
   (AC-FDM-009, -010).
5. `moai spec lint --strict SPEC-FACTORY-DEFAULT-MODE-001` passes.
