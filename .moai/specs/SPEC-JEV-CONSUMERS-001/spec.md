---
id: SPEC-JEV-CONSUMERS-001
title: "Jev consumers — near-duplicate marking, lane-question routing, skill suggestion"
version: "0.2.0"
status: draft
created: 2026-09-20
updated: 2026-09-20
author: manager-spec
priority: P2
phase: "v3.1.0 target"
module: "internal/kanban + internal/cli"
lifecycle: spec-anchored
tags: "jev, typesafe, consumers, backlog-finding, routing, skill-suggestion"
tier: L
depends_on: [SPEC-JEV-OPTIN-MEASURE-001]
---

# SPEC-JEV-CONSUMERS-001

## HISTORY

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-09-20 | 0.1.0 | Split from SPEC-JEV-INTEGRATION-001 (card t1020) on the M4+M5+M6 seam. Carries the three consumers, each independently gated on its own measurement from `SPEC-JEV-OPTIN-MEASURE-001`. Requirements re-numbered `REQ-JEVN-*`; none dropped. | manager-spec |
| 2026-09-20 | 0.2.0 | Revision answering the iteration-1 plan-audit FAIL (`.moai/reports/t1020/plan-audit-SPEC-JEV-CONSUMERS-001.md`, 0.64 against the Tier L 0.85 threshold). N1 recorded as a settled decision; `REQ-JEVN-006` restated to name the predicate its settled rule actually needs; `REQ-JEVN-015` added for the gate-not-run disposition; `REQ-JEVN-001` given an explicit re-sweep scope clause; `AC-JEVN-012`…`015` added; every AC now names its REQ ids. REQ 14→15, AC 11→15. | manager-spec |

## Position in the chain

Third of four. Depends on `SPEC-JEV-OPTIN-MEASURE-001` for the measurement gate every consumer here must pass, and transitively on `SPEC-JEV-CORE-001` for the call path.

`SPEC-JEV-CORE-001` → `SPEC-JEV-OPTIN-MEASURE-001` → **`SPEC-JEV-CONSUMERS-001`** → `SPEC-JEV-GOAL-DIST-001`

The three consumers are independent of each other. They share this SPEC because they share one predecessor and one gate shape, not because any of them needs another.

---

## §A. Problem Statement

Three decisions in MoAI are currently made with no signal at all.

**Whether a newly admitted card duplicates an existing one.** The text analyser measures similarity and records a `near-duplicate` finding above a threshold. That catches restatements; it does not catch two cards that describe the same work in different vocabulary, which is the case a reader notices immediately and a string metric cannot see.

**Who owns a lane's blocking question.** When a lane stops and asks, someone has to decide whether the answer is the lead's, the operator's, or the lane's own — and whether a measurement should come first. That classification is made from scratch each time, by whoever reads the question.

**Which skill, if any, a turn needs.** The intent router picks. Nothing ranks the alternatives it did not pick, and nothing asks whether a skill was needed at all.

Each is a typed judgment over a small state — the shape the model answers well. Each is also a place where a *displayed* signal helps and a *decided* answer is a defect, because all three feed work that a human is about to act on.

The near-duplicate consumer carries a specific structural hazard the other two do not, and it is the reason this SPEC exists as its own document rather than as a paragraph. The backlog's finding record distinguishes *who observed* a relation: `mechanical` means the text analyser measured it, `agent` means a reader who understands the cards judged it. `HasAgentFindingForPair` selects on `agent` and drives the `machine-only` mark, whose documented meaning is "nothing agent-sourced was recorded here". Filing a model answer under `agent` would clear that mark from pairs nobody reviewed — the queue would assert a review that never happened, which is exactly the display-becomes-verdict failure the whole capability is bounded against.

## §B. Goal

Three consumers, each shipped only after beating its own constant-answer baseline, each recording or presenting a signal that no code path acts on — and, where a consumer's gate cannot be run at all, a recorded decision in place of that consumer.

---

## §C. Requirements (GEARS)

### C.1 — Consumer C: near-duplicate card marking

**REQ-JEVN-001** (Event-driven) **When** a card is admitted to the backlog and the capability is enabled, **the system shall** record its Jev near-duplicate judgment as a `BacklogFinding`, changing no card. The trigger is **admission only**: the `moai todo analyze` re-sweep (`analyzeQueue`, `internal/cli/todo_analysis.go:140`, whose own mechanical write site is `:179`) is a distinct entry point from the admission path (`appendAnalyzedCard`, `:38`), and Consumer C shall not be invoked from it. Verified by `AC-JEVN-013`.

**REQ-JEVN-002** (Ubiquitous) A Jev finding shall carry a **third** source constant, distinct from both `BacklogSourceMechanical` (`internal/kanban/backlog_store.go:118`) and `BacklogSourceAgent` (`:120`). Verified by `AC-JEVN-002`.

**REQ-JEVN-003** (Unwanted) A Jev finding shall not be written with `Source: agent`. `HasAgentFindingForPair` is the predicate behind the `machine-only` mark, and that mark records the absence of an agent-sourced record for a pair — never that a review took place. Writing a Jev finding as `agent` would make the queue claim a review happened when none did. The prohibition binds the **write path**, not only the predicate: `AC-JEVN-001` verifies the predicate's behaviour on a Jev-only pair, and `AC-JEVN-012` verifies that no Jev producer in the diff sets `BacklogSourceAgent`.

**REQ-JEVN-004** (Ubiquitous) `HasAgentFindingForPair` (`internal/kanban/backlog_store.go:409`) shall continue to return false for a pair whose only finding is Jev-sourced. Verified by `AC-JEVN-001`.

**REQ-JEVN-005** (Ubiquitous) The rendering of a Jev finding shall follow the existing source-conditioned rules in `todoFindingLine` (`internal/cli/todo_analysis.go:215`): a score is printed only where a measurement was taken, so a Jev finding's probability shall be rendered in a form a reader cannot mistake for the text analyser's similarity score. Verified by `AC-JEVN-004`.

**REQ-JEVN-006** (Event-driven) — the dedup precedence rule, **settled 2026-09-20 by operator decision**. The rule has two halves whose implementation costs are opposite, and they are stated separately for that reason. Verified by `AC-JEVN-003`.

> **(a) Suppression of an arriving Jev finding — requires new code.**
> **When** a Jev finding naming an unordered pair with a given relation is appended, and a finding of **any** source already names that same unordered pair with that same relation, **the system shall not** append the Jev finding.
>
> **(b) Non-suppression of a later mechanical or agent finding — requires no code.**
> **When** a mechanical or agent finding naming a pair that already carries a Jev finding is appended, **the system shall** append it alongside the Jev finding rather than replacing it.

Half (b) is the **unchanged default** of `AppendFindingOnce` (`internal/kanban/backlog_store.go:369`) and shall be implemented by changing nothing. It is stated as a requirement so that a future change to the dedup key is recognised as breaking it rather than as a refactor.

Half (a) **shall be enforced by a new source-agnostic unordered predicate** over `BacklogRecord.Findings` — provisionally named `HasFindingForPairAnySource` — keyed on the unordered pair plus the relation and **not** on `Source`. It shall **not** be enforced by `AppendFindingOnce` or by `SamePairAs` in their present form, because neither can express it: `AppendFindingOnce` delegates solely to `HasFindingTuple` (`internal/kanban/backlog_store.go:357`), whose key is `{SubjectID, RelatedID, Relation, Source}` — ordered **and** source-inclusive — so a Jev finding always differs in `Source` from a mechanical or agent finding and is never suppressed by it; and `AppendFindingOnce` never calls `SamePairAs` (`:159`), whose single non-test caller in this repository is `HasAgentFindingForPair` (`:409`). The new predicate is authorised by this requirement and budgeted in `plan.md` §F M4.

**REQ-JEVN-007** (Unwanted) No code path shall write a card field as a consequence of a Jev finding. Verified by `AC-JEVN-005`.

### C.2 — Consumer A: lane-question routing

**REQ-JEVN-008** (Event-driven) **When** a lane's blocking question is routed and the capability is enabled, **the system shall** ask one Choice question naming the decision owner (lead / operator / worker) and two Noul questions — whether a new measurement is needed first, and whether the decision is cheap to reverse. Verified by `AC-JEVN-006`.

**REQ-JEVN-009** (Ubiquitous) The three questions shall be batched into one request over one state. Verified by `AC-JEVN-006`.

**REQ-JEVN-010** (Ubiquitous) The routing answer shall be surfaced to the reader as a signal and shall not itself dispatch, answer, or close anything. Verified by `AC-JEVN-008`.

**REQ-JEVN-011** (Capability gate) **Where** this consumer's fitted threshold exists, **When** the Choice answer is `operator` or the Choice confidence falls below that threshold, **the system shall** treat the item as operator-owned — that is, the routing surface labels the item `operator` and the consumer offers no other disposition for it. Where no fitted threshold exists, the confidence trigger is unsatisfiable and the consumer falls under `REQ-JEVN-015`. Verified by `AC-JEVN-007`.

### C.3 — Consumer B: skill suggestion

**REQ-JEVN-012** (Event-driven) **When** a turn's intent is being routed and the capability is enabled, **the system shall** issue two requests: a wide rank over all skills paired with a Noul asking whether the turn needs a skill at all, then a rerank of the top three under fuller text. Verified by `AC-JEVN-009`.

**REQ-JEVN-013** (Ubiquitous) The suggestion shall be presented to the orchestrator as a ranked signal carrying each candidate's rank position, and the `/moai` intent router shall retain its existing selection authority unchanged. The presentation half is verified by `AC-JEVN-014`; the authority half by `AC-JEVN-011`.

**REQ-JEVN-014** (Capability gate) **Where** this consumer's fitted threshold exists, **When** the "needs a skill at all" Noul answers negatively above that threshold, **the system shall** suppress the ranked list rather than present a best-of-nothing. Where no fitted threshold exists, this requirement is unsatisfiable and the consumer falls under `REQ-JEVN-015`. Verified by `AC-JEVN-010`.

### C.4 — Cross-consumer: disposition when the gate cannot be run

**REQ-JEVN-015** (Event-driven) **When** a consumer's measurement gate **cannot be run** — a state materially distinct from the gate having been run and failed — **the system shall** produce, as that consumer's run-phase output, a **recorded decision citing the un-runnable gate**: the consumer's call path is absent from the shipped build, and the record names why the gate could not be run and the mechanism that withheld it (`Report.Verdict()`, `internal/jevmeasure/measure.go:193`, which returns `VerdictWithhold` for any `Source` other than `SourceLive`). No measurement artifact shall be required on this branch, because in this state none exists. Verified by `AC-JEVN-015`.

---

## §D. Exclusions

### Out of Scope — the gate itself

- Building the measurement harness, the constant-answer baseline, the two language arms, or the threshold-fitting procedure. All owned by `SPEC-JEV-OPTIN-MEASURE-001`. This SPEC *passes* that gate; it does not build it.
- Shipping any consumer whose measurement does not beat its baseline. The withheld consumer's absence is recorded as a decision, not an omission.
- Fitting any threshold. `REQ-JEVN-011` and `REQ-JEVN-014` consume a threshold whose provenance is fixed by `AC-JEVO-013` in the predecessor; where no such threshold exists, both capability gates are unsatisfied and their consumers fall under `REQ-JEVN-015`.

### Out of Scope — card schema and queue authority

- Changing `BacklogItem`'s fields, JSON tags, or the frozen per-item contract. The third source constant extends `BacklogFinding.Source`; no card field changes.
- Changing how cards are issued, who may issue them, or the queue's producer rules.
- Any folding, reordering, dropping, or editing of a card in response to a finding. The finding record's structural property — that no code path writes a card field as a consequence of one — is inherited and preserved.
- Changing the `moai todo analyze` re-sweep. `REQ-JEVN-001` scopes Consumer C to admission; the re-sweep's existing mechanical write site (`internal/cli/todo_analysis.go:179`) is read for scope purposes and is not modified.

### Out of Scope — premise-death judgment

- Using Jev to judge whether a stale card's premise is dead. Measured and rejected; `moai todo triage` stays model-free, and no requirement here routes a model answer into it. The rejection record is owned by `SPEC-JEV-GOAL-DIST-001`.

### Out of Scope — authority transfer

- Giving any of the three consumers the ability to dispatch, answer, close, or select. Routing surfaces a classification; skill suggestion surfaces a ranking; the intent router keeps its selection authority unchanged.
- The `/moai goal --auto` seats that consume routing, owned by `SPEC-JEV-GOAL-DIST-001`.

---

## §E. Open Questions

| # | Question | Status |
|---|---|---|
| N1 | Dedup precedence between a Jev finding and a mechanical or agent finding naming the same unordered pair | **SETTLED 2026-09-20 (operator decision).** Recorded as a decision in `plan.md` §B1 and `design.md` §2; stated as `REQ-JEVN-006` halves (a) and (b). No longer carried to the Kickoff gate. |
| N2 | Which code path hosts Consumer A (`research.md` §5) | **OPEN — blocking M5 and M6.** Not covered by the 2026-09-20 operator decisions. M5 and M6 are declared blocked on it in `plan.md` §F; carried to the Implementation Kickoff Approval gate. |
| Q3 | Labelled-set size per consumer | OPEN (owned by `SPEC-JEV-OPTIN-MEASURE-001`) — determines each gate's sample size before this SPEC's consumers can pass it. |
