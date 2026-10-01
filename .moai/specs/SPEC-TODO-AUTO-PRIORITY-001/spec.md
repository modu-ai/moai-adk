---
id: SPEC-TODO-AUTO-PRIORITY-001
title: "Priority-aware card selection for todo --auto — a Jev-ordered or recorded-priority-fallback ranking of queued candidates, a printed selection decision record, and an --auto-scoped doctrine exception"
version: "0.1.1"
status: in-progress
created: 2026-10-02
updated: 2026-10-02
author: manager-spec (card t1400)
priority: P2
phase: "v3.2.0 target"
module: "internal/cli, internal/kanban"
lifecycle: spec-anchored
tags: "todo, auto, priority, ranking, jev, fallback, selection-record, doctrine-amendment, card-t1400"
tier: M
card: t1400
related_specs: [SPEC-MANAGER-TODO-001, SPEC-RELATION-PICKUP-FILTER-001, SPEC-TODO-CLASSIFY-DISPATCH-001, SPEC-TODO-HOLD-STATE-001, SPEC-JEV-CORE-001, SPEC-JEV-OPTIN-MEASURE-001]
---

# SPEC: priority-aware card selection for `todo --auto`

## HISTORY

- 0.1.1 — 2026-10-02 — plan-audit iteration 1 revision (card t1400). Operator
  decisions 5-6 (§B.1) recorded as settled: the doctrine amendment is limited to
  `kanban-dispatch.md`, `workflows/gtd.md` and `manager-todo.md`, with the Jev-side
  surfaces deferred to a follow-up card, and the Jev ordering consumer ships only
  behind the default-off gate with no accuracy claim. REQ-TAP-001, -002, -003,
  -007, -011, -012 and -014 reworded (unchanged-filters clause, Jev order key,
  defect boundary and numeric score range, the `selection order only` wording
  constraint); §B.5 (interim state) added; §D and §G updated. No requirement or
  acceptance criterion was added (14 / 15).
- 0.1.0 — 2026-10-02 — plan-phase artifact set authored (card t1400; worktree
  `.moai/worktrees/t1400`, branch `WT-auto-priority-pick`, HEAD `7d8a9bdbc`).
  Tier M: spec.md + plan.md + acceptance.md. Operator decisions 1-4 (§B.1) are
  binding input and are not re-opened here.

## §A Context

### A.1 Problem

`moai todo --auto` picks its next card in `autoPickTargets`
(`internal/cli/todo_auto.go:142`): dead-owner `picked` cards first, then every
`queued` card in stored queue order. No stage ranks or filters the queued
candidates by readiness, and the file header states the contract as "the cycle
never reorders" (`todo_auto.go:27`). An unworkable card at the front of the
queue — a card whose text carries a `[보류` marker, a card already landed, a card
flagged as a near-duplicate, a card blocked on a lead signal — therefore becomes
the cycle's first target and burns a worker on work nobody can do.

### A.2 Verified basis (this tree, HEAD `7d8a9bdbc`)

Full evidence with `file:line` is in plan.md §Findings. The facts the
requirements stand on:

- **A priority field exists.** Every card carries an additive nullable
  `Classification` (priority `high|normal|low`, blocked flag, mode); an absent
  classification reads as priority `normal`, blocked `false`
  (`internal/kanban/classification.go:74-81`, `:137-142`). The comment
  "there are no priority fields" at `internal/cli/todo_edit_move.go:99` is stale.
- **Queue order is the classification order at add time.** `SortByClassification`
  (`classification.go:198-209`: non-blocked first, then priority, stable on
  insertion order) runs inside the add's locked write only; `todo move` permutes
  without re-sorting.
- **The pickup predicate has one safe insertion point.** `autoPickTargets`
  returns the rescue arm followed by the queued arm; a ranking stage applied to
  the queued suffix inside `runAutoCycle` leaves the rescue arm and the
  relation filter untouched (plan.md §Findings (c)).
- **Jev has two distinct surfaces and only one ships.** The Go capability
  (`internal/jev`, gate `workflow.jev.enabled`, default `false`) is tracked and
  distributed. The local scripts the cycle's existing display line reads
  (`scripts/jev/route.sh`) are NOT tracked in this tree — `git ls-files
  scripts/jev` lists `triage.py` and `test_triage.py` only, and `route.sh` is
  absent — and `route.sh` emits a five-value decision grade, not an ordering.

## §B Decisions

### B.1 Operator decisions (binding input — do not re-open)

Decisions 1-4 were binding at plan start. Decisions 5-6 were made on 2026-10-02
in answer to the two open items the first plan draft carried; they are settled,
not open.

1. **Ranking source.** When a Jev signal is available it decides the order of the
   queued candidates. When it is not (capability gate off — the shipped default —,
   call failed or timed out), fall back to: recorded priority (`high > normal >
   low`), exclusion of blocked cards, and demotion — not dropping — of
   readiness-poor cards (PR state `landed`, a near-duplicate finding, a card text
   starting with the `[보류` marker). The fallback reason is printed in the cycle
   output; the fallback is never silent.
2. **Doctrine amendment, limited to `--auto`.** The `[HARD]` clause in
   `kanban-dispatch.md` (the leader never reorders by inferred priority) and the
   `--auto` section of `workflows/gtd.md` ("never reorders", "Jev consultation is
   display-only") become an `--auto`-scoped exception. The queue is NOT changed:
   no move, no reorder, no new writes beyond the existing pick / unpick / done.
   Only the order in which the cycle chooses changes.
3. **Kept.** Dead-owner `picked` cards still come first; hold-state cards stay
   excluded; the relation pickup filter is unchanged; Jev stays a selection-order
   signal only — never the basis of a completion verdict or a queue mutation; the
   `--auto` invocation remains the operator's batch approval.
4. **Template-First.** Live files and their `internal/template/templates/`
   mirrors are both in scope.
5. **Doctrine scope (2026-10-02).** Amend ONLY `kanban-dispatch.md`,
   `workflows/gtd.md` and `manager-todo.md`, live files and their template
   mirrors where they exist. The Jev-side surfaces are NOT touched by this card:
   a follow-up card, to be issued by the leader, carries that linked amendment.
   The surfaces are listed exactly in §D and in plan.md §B (Deferred follow-up);
   the interim state they create is disclosed in §B.5.
6. **Jev ordering consumer and accuracy (2026-10-02).** The Jev ordering consumer
   ships ONLY behind the default-off gate `workflow.jev.enabled`. Enabling it is
   the operator's act. This SPEC's requirements and acceptance criteria claim no
   ordering accuracy.

### B.2 "Jev available" is defined as the Go capability, not the scripts

"Jev available" means ALL of: (a) `workflow.jev.enabled` is `true`; (b) the
credential at `~/.moai/.env.typesafe` loads; (c) one bounded request returns
`Available` within the call bound; (d) the answer set passes the validation of
REQ-TAP-011. Any other outcome is "unavailable" and routes to the fallback with
a named reason.

The local scripts are deliberately NOT part of the definition: they are
untracked, absent from this tree and from the template, and emit a decision grade
that carries no per-card order. The existing display-only script line printed by
`runAutoCycle` (`todo_auto.go:226`) is left exactly as it is and stays
independent of the ranking. This resolves the operator's "scripts absent" fallback
condition onto the surface that actually ships; plan.md §B A-5 records it as a
flagged interpretation, to be corrected before kickoff if the operator meant the
scripts.

### B.3 Eligibility versus order

Exclusion is an eligibility rule applied on BOTH sources: a card whose effective
classification is `blocked` is never an `--auto` target (the same direction the
factory lease path already takes). Order is what differs: the Jev path takes
Jev's order; the fallback path applies priority plus demotion. Readiness signals
are printed on both sources; they are applied (demotion) on the fallback only.

Applying the blocked exclusion on the Jev source is an extension of the
operator's wording, which lists it under the fallback only; it is a flagged
assumption (plan.md §B A-1) to be corrected before kickoff if wrong.

### B.4 Ranking is computed once per invocation

The ranking is taken once, over the record loaded at cycle start, and the cycle
processes that order for the whole invocation. Re-ranking per card is not
planned (plan.md §G R-3).

### B.5 Interim state and the wording constraint

This card amends three documents only (decision 5). Until the follow-up card
lands, the Jev-side surfaces keep saying that a Jev answer is display-only, while
the three amended documents say that the `--auto` cycle takes Jev's order for its
own candidate selection. A reader of the Jev-side text alone is not told about the
exception.

The amended clauses are therefore constrained so that the two statements stay as
compatible as the text can make them. Each amended passage shall (a) name the
exception with the literal `auto-scoped ranking exception`, (b) bound it with the
literal `selection order only` in the same paragraph, (c) keep every prohibition
on every other surface, and (d) not state or imply that the Jev principle is
amended anywhere outside the `--auto` cycle. REQ-TAP-012 and REQ-TAP-014 carry
the constraint as requirements; the wording check is literal and
paragraph-scoped (acceptance.md AC-TAP-011, AC-TAP-015).

## §C Requirements

Verification layer: `acceptance.md`. The requirement layer below is GEARS.

- **REQ-TAP-001** (Ubiquitous) — The `todo --auto` cycle shall order its queued
  pickup candidates through one ranking stage that runs once per invocation,
  after the existing eligibility filters and before the first `accept` line, and
  shall process the resulting order for the whole invocation; the existing
  eligibility filters (the relation filter and the hold-state exclusion) and the
  existing `jev:` display line shall stay exactly as they are.

- **REQ-TAP-002** (Where) — **Where** the Jev ordering signal is available (§B.2),
  the cycle shall ask the Jev capability one bounded request over the candidate
  set, one typed score question per candidate, and shall order the candidates by
  the returned score, highest score first, breaking a tie by the answer's
  confidence (highest first) and then by fallback order, and shall record
  `source=jev`; a candidate set larger than the request bound shall place the
  surplus after the Jev-ordered candidates in fallback order, and the record
  shall name that.

- **REQ-TAP-003** (When) — **When** the Jev signal is not available — gate off,
  credential absent or refused, rate-limited, overloaded, unreachable or timed
  out, request refused for size or content, or answers defective (REQ-TAP-011) —
  the cycle shall rank by the fallback keys, shall record `source=fallback`, and
  shall not exit non-zero because of the unavailability.

- **REQ-TAP-004** (Ubiquitous) — The fallback ranking shall order candidates by
  effective recorded priority `high > normal > low` (an absent classification
  reads as `normal`), keeping queue order within one priority, and shall place
  every readiness-poor candidate after every readiness-clean candidate while
  keeping each group in that same order.

- **REQ-TAP-005** (Ubiquitous) — A queued candidate shall be classified
  readiness-poor when at least one holds: (a) `moai todo pr` reports its PR state
  as `landed`; (b) a live near-duplicate finding names the card; (c) its text,
  with surrounding whitespace trimmed, begins with the marker `[보류`. A signal
  that cannot be measured — the landed lookup answering `unknown` or failing —
  shall never be read as poor, and the record shall name the unmeasured signal.

- **REQ-TAP-006** (Ubiquitous) — The cycle shall exclude from its targets every
  queued card whose effective classification is `blocked`, on both ranking
  sources, and shall print one labelled non-finding per excluded card; exclusion
  shall leave the card's state unchanged.

- **REQ-TAP-007** (Ubiquitous) — While the source is `fallback`, the cycle shall
  print exactly one reason on the source line, drawn from the closed vocabulary
  `jev-disabled`, `jev-no-credential`, `jev-unauthorized`, `jev-rate-limited`,
  `jev-overloaded`, `jev-unreachable`, `jev-oversize`, `jev-secret-detected`,
  `jev-malformed`, `jev-incomplete-answer`. The two lookalike reasons are
  distinct: `jev-malformed` names a result whose availability is `Malformed` (a
  request that carried no question — reachable only through the test seam,
  because the stage never sends an empty question list), while an answer set that
  fails REQ-TAP-011 is `jev-incomplete-answer`; a response that cannot be read is
  `jev-unreachable`.

- **REQ-TAP-008** (Ubiquitous) — Before the first `accept` line the cycle shall
  print a selection decision record: a `selection: source=…` line, a
  `selection: ranked …` line listing the ranked candidate ids in processing
  order (omitted when every candidate was excluded), one
  `selection: flagged <id> (<signals>)` line per readiness-poor card, one
  `selection: excluded <id> (blocked)` line per excluded card, and any
  `selection: note …` lines; the record shall be printed whenever the queued
  candidate set that survived the relation filter is non-empty.

- **REQ-TAP-009** (Ubiquitous) — The cycle shall place every dead-owner `picked`
  rescue target before every ranked queued target regardless of the ranking
  source, and the ranking shall neither reorder, filter, nor demote the rescue
  arm.

- **REQ-TAP-010** (Unwanted) — The ranking stage shall not write the queue: no
  move, reorder, edit, hold, drop, new field, or new record; the only queue
  writes the cycle performs remain the existing pick, unpick, and done
  transitions, and the cycle's authority shall continue to derive solely from the
  `--auto` invocation, which grants no admission and no pick outside the cycle.

- **REQ-TAP-011** (Unwanted) — The cycle shall not use a Jev answer for anything
  other than the selection-order key — never as a completion verdict, a queue
  mutation, a merge approval, or an operator gate — and an answer set that names
  a card that was not sent, leaves a sent card unanswered, or carries a score
  that is not a finite number in the closed range 0 to 4 (the five declared
  levels, zero-based; a higher score marks a card Jev judges more ready to be
  picked now) shall make the whole Jev result unavailable
  (`jev-incomplete-answer`) instead of being partially applied.

- **REQ-TAP-012** (Ubiquitous) — The `[HARD]` clause in `kanban-dispatch.md` and
  the `--auto` section of `workflows/gtd.md` shall state the `--auto`-scoped
  ranking exception — naming it with the literal `auto-scoped ranking exception`
  and bounding it with the literal `selection order only` in the same paragraph
  (§B.5) — and shall keep every prohibition on every other surface (the leader
  pick, `gtd next`, the analyser) unchanged; the live files and their template
  mirrors shall carry the same amended wording, and the template mirrors shall
  carry no SPEC id, requirement token, date, or commit hash.

- **REQ-TAP-013** (Ubiquitous) — The `--auto` section of `workflows/gtd.md` (live
  and mirror) and the `--auto` flag help text shall disclose that only a card
  whose text begins with the `[보류` marker is demoted, that a hold stated in
  prose without the marker is not, and that the structural hold is `moai todo
  hold`.

- **REQ-TAP-014** (Ubiquitous) — The `manager-todo` agent definition (live and
  template mirror) shall carry the same `--auto`-scoped exception — both
  literals of REQ-TAP-012 in the same paragraph — in its serial-cycle contract
  and its Jev decision boundary, so that no agent text keeps asserting that the
  cycle consumes queue order only.

### C.1 Traceability

REQ-TAP-001 → AC-TAP-001 + AC-TAP-012 · REQ-TAP-002 → AC-TAP-002 · REQ-TAP-003 →
AC-TAP-003 · REQ-TAP-004 → AC-TAP-004 · REQ-TAP-005 → AC-TAP-005 + AC-TAP-006 ·
REQ-TAP-006 → AC-TAP-007 · REQ-TAP-007 → AC-TAP-003 · REQ-TAP-008 → AC-TAP-001 ·
REQ-TAP-009 → AC-TAP-008 + AC-TAP-012 · REQ-TAP-010 → AC-TAP-009 + AC-TAP-012 ·
REQ-TAP-011 → AC-TAP-010 · REQ-TAP-012 → AC-TAP-011 + AC-TAP-013 · REQ-TAP-013 →
AC-TAP-014 · REQ-TAP-014 → AC-TAP-015 (bodies and `**Covers**` clauses in
`acceptance.md`).

## §D Out of Scope

### Out of Scope — queue mutation and operator surfaces

- No change to the queue's stored order, to `todo move`, `todo hold`, `todo add`,
  or to the classification sort (`SortByClassification`); no new card field and
  no schema change. The `--auto` flag keeps its meaning and takes no new argument.

### Out of Scope — other Jev surfaces and the Jev doctrine linkage

- Decided 2026-10-02 (operator decision 5): this SPEC amends only the three
  documents named in §B.1 item 5. It does not touch the other surfaces that state
  the Jev display-only principle: `internal/jev/jev.go` (package comment, `:24`),
  the `workflow.jev` comment of `workflow.yaml` (live `:227`, template `:229`),
  `internal/cli/mcp_jev.go` (`:8`), the MCP tools catalogue
  (`moai-mcp-tools-catalogue.md` § Judgment, live `:135` and `:139`, with its
  template mirror), `SPEC-JEV-CORE-001`, and `SPEC-MANAGER-TODO-001`
  (REQ-MT-014/015). The contract-mode Kickoff exception to that principle was made
  as a linked multi-surface amendment guarded by `TestJevAmendmentLinkage`; a
  follow-up card, to be issued by the leader, carries the corresponding linked
  amendment for the `--auto` exception (plan.md §B, Deferred follow-up).
- Interim state (known limitation): until that follow-up lands, those surfaces
  still say "display-only" while the three amended documents describe the
  `--auto` exception. The amended wording is constrained to keep the two
  statements compatible (§B.5, REQ-TAP-012); this SPEC discloses the
  contradiction and does not resolve it (§G R-6).

### Out of Scope — Jev accuracy and the shipping gate

- Decided 2026-10-02 (operator decision 6): `SPEC-JEV-OPTIN-MEASURE-001`
  REQ-JEVO-009 forbids shipping a consumer whose measured accuracy does not beat
  its constant-answer baseline. This SPEC builds the ordering consumer only
  behind the default-off `workflow.jev.enabled` gate (code default `false`,
  `internal/config/defaults.go:1180-1182`), does not measure or certify its
  ordering accuracy, and claims none in any requirement or acceptance criterion;
  enabling it in a project is the operator's act under that gate, not a
  deliverable here.

### Out of Scope — factory lease and other selectors

- `moai factory next`, `moai todo next`, `moai todo claim`, and `todo auto-done`
  keep their own selection predicates; the ranking applies only to the `--auto`
  cycle.

### Out of Scope — per-card re-ranking and the stale comment sweep

- Re-ranking after each completed card, a per-card Jev call, and a sweep of every
  historical comment that restates the old contract beyond the passages named in
  plan.md §E are not part of this SPEC.

## §G Gaps and Residual Risks

- **R-1 — Locale-specific marker.** The demotion marker `[보류` is a Korean
  literal chosen by the operator. A project that records holds in another
  language or in prose is not demoted (REQ-TAP-013 discloses this).
- **R-2 — Jev may rank a readiness-poor card first.** On the Jev source the code
  applies no demotion by design (operator decision 1); the flagged line makes
  the disagreement visible but does not prevent it.
- **R-3 — Cycle-start ranking.** Landed state, findings, and Jev answers are
  measured once per invocation; a card that becomes poor mid-invocation is not
  re-ranked until the next invocation.
- **R-4 — Operator `move` intent versus priority.** The fallback orders by
  priority first, so a `todo move --top` of a `normal` card above a `high` card
  is not honoured by `--auto`; within one priority the queue order, and so every
  `move`, is kept.
- **R-5 — `landed` is attribution, not completion.** `moai todo pr` reports
  `landed` when something naming the card landed on the integration ref; a
  partially landed card is demoted, not dropped, which is the safe direction.
- **R-6 — Interim doctrine contradiction.** Until the follow-up card lands, the
  Jev-side surfaces (§D) still say a Jev answer is display-only while the amended
  documents grant the `--auto`-scoped exception. The wording constraint (§B.5)
  narrows the contradiction; it does not remove it, and the literal check cannot
  detect a sentence that generalises the exception in other words.
- **R-7 — Score range base unobserved.** The range 0 to 4 in REQ-TAP-011 assumes
  zero-based level indices for a five-level score question; the base was not
  observed against the live capability (plan.md §B A-6). An answer that omits its
  score decodes to 0, which is in range, so the unanswered-card check is by
  presence per card id, not by score value.
