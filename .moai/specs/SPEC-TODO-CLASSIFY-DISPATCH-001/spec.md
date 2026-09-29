---
id: SPEC-TODO-CLASSIFY-DISPATCH-001
title: "LLM-classified card metadata at creation — priority·blocked·execution-mode on every card, a priority-sorted queue, mode-aware factory lane leases, and default-on auto-dispatch for -f lanes"
version: "0.3.0"
status: draft
created: 2026-09-29
updated: 2026-09-29
author: manager-spec (card t1332)
priority: P2
phase: "v3.2.0 target"
module: "internal/cli, internal/kanban, internal/homestate, internal/hook, internal/config"
lifecycle: spec-anchored
tags: "todo, factory, classification, priority, execution-mode, serial, parallelizable, auto-dispatch, lease, queue-sort, card-t1332"
tier: M
card: t1332
depends_on: [SPEC-FACTORY-SELF-DISPATCH-001]
related_specs: [SPEC-MANAGER-TODO-001, SPEC-TODO-HOLD-STATE-001, SPEC-AUTONOMY-CONTRACT-001, SPEC-KANBAN-TODO-CLI-001, SPEC-TODO-SQLITE-001]
---

# SPEC: LLM-classified card metadata at creation

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-29 | Initial plan-phase authoring (card t1332, operator directive 2026-09-29), measured in worktree `.moai/worktrees/t1332`, branch `WT-card-autodispatch`, at HEAD `145c3d98c` (== local develop tip). The measured surface basis is exported at `.moai/reports/t1332/surface-notes.md`. Card premises verified against the tree: the t1240 self-dispatch surface (`factory next`/`stage`/`complete`) exists on branch `WT-factory-self-dispatch` (tip `d43e50bb3`, 28 commits ahead of develop at plan time — the card said 25; the difference is an absorb of develop, and the surface matches the card's description) and is NOT merged to develop; `internal/cli/contract_decide.go` and the `homestate`/`contract` decider vocabulary are the t1261 layer this SPEC reuses rather than re-plans; t1306's `/moai:todo --auto` serial cycle is merged on develop. No card premise was falsified. |
| 0.2.0 | 2026-09-29 | **Leader-ruling revision** (plan-audit PASS 0.875 iter 1; D1 + OD rulings folded, one repair pass). **D1 (provenance-only)**: the t1240-only `--clear-policy` citations in plan.md §B.5/§F M4 and surface-notes.md items 5/8 now carry the `WT-factory-self-dispatch:` prefix (the flag is absent from this tree's develop — measured 0 `ClearPolicy` hits); no REQ/AC content changed. **OD-1 REVISED (operator ruling 2026-09-29)**: pipeline exclusivity REJECTED; serial-card mutual exclusivity ADOPTED — a serial card blocks only OTHER SERIAL cards, served one at a time in priority order; parallelizable selection is unaffected while a serial card is in flight. Rationale recorded: a full `next` refusal costs half the throughput benefit that classification exists to deliver. REQ-TCD-008 rewritten, REQ-TCD-009 extended, AC-TCD-008 re-derived (three clauses). **OD-3 REVISED (same ruling)**: decider-failure default is now `serial` (fail-safe — a parallelizable failure default could run true-serial cards concurrently and violate ordering; the serial default costs throughput only); REQ-TCD-003 rewritten. The new serial failure-default vs REQ-TCD-014's parallelizable absent-field read default is a recorded TENSION, flagged for the lead in plan.md §C — not silently resolved. Provenance: leader Jev doctrine-fallback rulings, noul below threshold (OD-1 0.31, OD-3 0.36), dated 2026-09-29. REQ and AC counts unchanged (14 / 14). |
| 0.3.0 | 2026-09-29 | **Tension resolved (leader ruling 2026-09-29, OD-3 extension).** REQ-TCD-014's absent-field READ default for the MODE axis flips `parallelizable` → `serial`; the PRIORITY axis default (`normal`) and BLOCKED axis default (`false`) are untouched. Rationale recorded: absent = unclassified = conservative treatment; a parallelizable absence default would silently bypass OD-3's serial failure default (a fail-safe bypassed through silence). plan.md §C FLAGGED item becomes RESOLVED with this provenance. No new read-default AC added (auditor optional D4 stays run-phase discretion); no existing AC asserted the old parallelizable absence default (measured: zero default-value hits in acceptance.md). REQ and AC counts unchanged (14 / 14). |

> **Provenance discipline.** Every `file:line` citation was measured at HEAD `145c3d98c` in this
> worktree. t1240-branch citations are prefixed `WT-factory-self-dispatch:` and were read via
> `git show` — that branch is NEVER merged or checked out by this SPEC's plan phase, and its
> absorption into develop is a run-phase entry precondition (REQ-TCD-013).

## §A Context

### A.1 The gap: the queue is an unordered pile to every machine reader

Today a card carries no judgment about how urgent it is or how it may be executed. The queue item
is five frozen fields plus additive evidence stamps (`backlog_store.go:83-110`); order is insertion
order, and every machine consumer reads that order as priority. `moai factory next`'s
auto-promotion arm takes "the first `queued` item in slice order"
(`WT-factory-self-dispatch:internal/cli/factory_card.go`, arm (c)) — the oldest card, whatever its
urgency, and regardless of whether the lane running it can proceed concurrently with others. The
operator's judgment lives — at best — in card prose, which machines do not read (the same
text-invisibility defect SPEC-TODO-HOLD-STATE-001 §A.1 recorded for holds).

### A.2 What this SPEC adds — three pieces, exactly

Per the operator directive (card t1332):

1. **Creation-time classification.** `moai todo add` classifies every admitted card through a
   decider seam and records the judgment as card metadata: a priority ordinal, a blocked flag
   (urgency and blocked-ness), and an execution mode (`serial | parallelizable`). The queue is
   kept sorted by priority.
2. **Mode-aware factory dispatch.** `factory next` picks up cards in priority order;
   parallelizable cards allow distinct cards to be leased by distinct lanes concurrently; a serial
   card excludes other SERIAL cards — while one is in flight, no other serial card is leased
   (served one at a time in priority order) and parallelizable selection is unaffected; the
   factory record/status shows which lane holds which card.
3. **Auto-dispatch default-on for `-f` lane sessions.** A `-f` lane self-dispatches by default;
   manual specification remains a launcher flag option.

### A.3 The decider seam, in the existing vocabulary

The product already carries a decider-identity vocabulary this SPEC reuses, not duplicates: the
kickoff contract's closed set `human | llm | llm+jev` (`internal/contract/seal.go:55-56`), the
"jev is never a sole decider" refusal (`internal/contract/kickoff/decide.go:111-112`), and the
`homestate.Decider` identity recorded on card decisions (`internal/homestate/card_record.go:45-46`).
This SPEC defines the classification seam (a `Decider` abstraction for card classification) so that
its product implementation is `Decider(llm)` and its absence-of-judgment fallback is deterministic.
`scripts/jev/` is LOCAL-ONLY tooling (AGENTS.local.md §29) and is never wired into any product
path; the local dogfood seam is stated in plan.md §D.3.

## §B Requirements (GEARS)

> Value sets used throughout: `priority ∈ {high, normal, low}`; `blocked ∈ {true, false}`;
> `mode ∈ {serial, parallelizable}`; decider identity reuses the contract vocabulary
> (`llm` for the product implementation; the deterministic fallback records `default`).

- **REQ-TCD-001 (Ubiquitous)** — The `moai todo add` path shall classify every admitted card
  through the card-classification decider seam at creation time, inside the same locked write that
  appends the card, so no card becomes visible to a machine selector without a recorded
  classification.

- **REQ-TCD-002 (Ubiquitous)** — The queue item shall carry the classification as ONE additive
  nullable field on `BacklogItem` (priority, blocked, mode, decider identity, classified-at stamp,
  one-line reason) following the `Landing`/`PickedAt` additive discipline: a pointer, `omitempty`,
  absent-field marshal byte-identical to the pre-SPEC record, and the SQLite mirror added by a
  guarded `ALTER TABLE ... ADD COLUMN` (`backlog_sqlite.go:480-550` precedent).

- **REQ-TCD-003 (Event-driven)** — When the classification decider is unavailable or fails, the
  add path shall admit the card with the fail-safe default (priority `normal`, blocked `false`,
  mode `serial`, decider identity `default`), emit one stderr notice naming the fallback, and
  record the decider identity — the serial default is the fail-safe direction, because a
  wrongly-parallel default could run true-serial cards concurrently and violate the ordering the
  mode exists to protect, while a wrongly-serial default costs throughput only; admission never
  blocks on classification.

- **REQ-TCD-004 (Event-driven)** — When a classification is supplied explicitly to `todo add`
  (judgement file or standard input, mirroring `contract decide --judgement`), the add path shall
  validate it against the closed value sets before the locked write, refuse an out-of-set value as
  a usage error with nothing written, and record the supplying decider's identity; the decider
  identity `jev` shall be refused on every product classification path.

- **REQ-TCD-005 (Ubiquitous)** — The queue shall be kept sorted by the classification key —
  non-blocked before blocked, then priority `high > normal > low`, then insertion order stable
  within a rank — and the sort shall be re-established inside the same locked write as the add
  that may change it; no other verb reorders ranked cards.

- **REQ-TCD-006 (Ubiquitous)** — The queue position `add` prints, the `todo list` rendering, and
  the web queue read shall reflect the priority-sorted order, so every consumer that reads
  position reads priority.

- **REQ-TCD-007 (Event-driven)** — When a lane runs `factory next` and selection reaches the
  auto-promotion arm, the factory shall promote and lease the highest-ranked eligible queued card
  in sorted order, and shall never auto-select a `blocked` card (an operator pick or unblock is
  the only path that dispatches one).

- **REQ-TCD-008 (Ubiquitous)** — A `serial` card shall hold mutual exclusivity against other
  serial cards: while a serial card is recorded in the factory run in a non-terminal state,
  `factory next` shall lease no serial card for any lane — serial cards are served one at a time
  in priority order, and a lane finding only serial candidates blocks on the existing no-card
  exit — while parallelizable candidates remain leasable by other lanes throughout; the terminal
  states that re-admit serial selection shall be enumerated positively, never as a negation.

- **REQ-TCD-009 (Ubiquitous)** — `parallelizable` cards shall allow concurrent multi-lane
  dispatch: distinct parallelizable cards may be held by distinct lanes at the same time —
  including while a serial card is in flight — and the version-checked lease edges
  (`ErrStaleVersion`/`ErrLeaseHolder` race retry) remain the only duplicate-prevention mechanism
  — a card is never leased twice.

- **REQ-TCD-010 (Ubiquitous)** — The factory record and the `factory status` surface shall show
  which lane holds which card together with the card's execution mode and priority (the
  `factoryCardView` extension), so a reader of status can see the mode-aware dispatch state
  without opening the queue.

- **REQ-TCD-011 (Where)** — Where a `-f` lane session is launched, auto-dispatch — the lane
  entering its `factory next` self-dispatch loop without a per-card lead routing step — shall be
  the default behavior, the launcher shall stamp the selection into the lane bootstrap, and a
  launcher opt-out flag shall select the manual mode; the default is recorded in code, and no new
  config key is introduced.

- **REQ-TCD-012 (Ubiquitous)** — The classification backend shall be resolved through the
  Decider abstraction; the product shall ship the deterministic default and the `Decider(llm)`
  implementation, and no product path (`internal/`, `pkg/`, `cmd/`,
  `internal/template/templates/`) shall reference `scripts/jev/` or any local-only tooling path.

- **REQ-TCD-013 (Event-driven)** — When run-phase entry is attempted before local develop carries
  the t1240 self-dispatch surface (SPEC-FACTORY-SELF-DISPATCH-001 merged; `factory next` and
  `complete` present in develop's `internal/cli`), the run shall be blocked — the absorption order
  is a run-phase entry precondition, and the absorption source is branch
  `WT-factory-self-dispatch`.

- **REQ-TCD-014 (Event-driven)** — When a queue recorded before this SPEC is read, every
  classification consumer shall derive the no-judgment default for a card whose classification
  field is absent by positive default derivation at read — priority `normal`, blocked `false`,
  mode `serial` (leader ruling 2026-09-29, OD-3 extension: absent = unclassified = conservative
  treatment; a parallelizable absence default would silently bypass the serial failure default
  through silence) — never by treating absence as a fourth priority or a third mode.

## §C Constraints

- C1. The five-field per-item contract of REQ-TODO-013 keeps its names, types, and JSON tags; the
  classification field is additive after the `Landing` precedent (`backlog_store.go:92-100`).
- C2. The lead-side serial cycle `/moai:todo --auto` (SPEC-MANAGER-TODO-001, merged) is untouched;
  it consumes the queue in queue order and thereby inherits the sorted order without change.
- C3. Hold semantics (SPEC-TODO-HOLD-STATE-001) are untouched: a `hold` card stays invisible to
  every machine selector; sorting positions it, selection filters it.
- C4. Jev boundary: `scripts/jev/` is never referenced by product code (REQ-TCD-012); the existing
  `internal/jev` package stays behind its own `workflow.jev.enabled` default-off gate and gains no
  new consumer from this SPEC.
- C5. No new config key (REQ-TCD-011); the auto-dispatch default is a launcher code default.

## §D Dependencies

- **depends_on: SPEC-FACTORY-SELF-DISPATCH-001** (t1240) — the `factory next`/`stage`/`complete`
  lease surface, the version-checked claim edges, and the no-card exit this SPEC extends. Not yet
  merged to develop; absorption order is REQ-TCD-013.
- **related: SPEC-MANAGER-TODO-001** (t1306, merged) — the lead-side serial cycle that inherits
  the sorted queue. **SPEC-TODO-HOLD-STATE-001** (t1308, merged) — the fourth state and the
  positive-enumeration discipline REQ-TCD-008/014 follow. **SPEC-AUTONOMY-CONTRACT-001** — the
  decider vocabulary REQ-TCD-004/012 reuse. **SPEC-KANBAN-TODO-CLI-001 / SPEC-TODO-SQLITE-001** —
  the queue store and its additive-schema discipline.

## §E Out of Scope

### Out of Scope — lane self-lease verbs (t1240 / SPEC-FACTORY-SELF-DISPATCH-001)

- The `factory next`/`stage`/`complete` verb surface, the duplicate-prevention lease mechanics,
  and the admission/refusal asymmetry are t1240's deliverables; this SPEC only reorders the
  selection input and adds the mode-aware eligibility gate on top. Merging, checking out, or
  re-planning that branch is forbidden.

### Out of Scope — lead-side serial auto-pickup (t1306 / SPEC-MANAGER-TODO-001)

- The `/moai:todo --auto` foreman cycle, its evidence-judged completion, and its serial
  card-processing contract are merged and owned; no change here.

### Out of Scope — the kickoff contract decider layer (t1261)

- `contract decide`, the kickoff receipt, the R5 jev-sole refusal, and the
  `workflow.autonomy.kickoff.decider` config are owned there; this SPEC defines a separate
  classification seam and reuses only the identity vocabulary.

### Out of Scope — Jev product wiring

- `scripts/jev/` stays a local-only dev tool (AGENTS.local.md §29). No template mirror, no
  product-path reference, no new `internal/jev` consumer. The local dogfood seam is a plan.md
  concern (§D.3), not a deliverable.

### Out of Scope — scheduling intelligence

- The blocks/depends relation graph (t1309) stays record-only; no scheduler reads it. Priority is
  the classification's three-level ordinal — not a computed score, not a re-classification on
  edit, not a deadline system.

## §F Cross-references

- Measured surface basis: `.moai/reports/t1332/surface-notes.md`.
- Acceptance criteria: `acceptance.md` (AC-TCD-001..014, Given-When-Then, binary-testable).
- Implementation plan and open decisions: `plan.md`.
