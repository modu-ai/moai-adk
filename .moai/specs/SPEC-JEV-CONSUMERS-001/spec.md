---
id: SPEC-JEV-CONSUMERS-001
title: "Jev consumers — near-duplicate marking, lane-question routing, skill suggestion"
version: "0.1.0"
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

Three consumers, each shipped only after beating its own constant-answer baseline, each recording or presenting a signal that no code path acts on.

---

## §C. Requirements (GEARS)

### C.1 — Consumer C: near-duplicate card marking

**REQ-JEVN-001** (Event-driven) **When** a card is admitted to the backlog and the capability is enabled, **the system shall** record its Jev near-duplicate judgment as a `BacklogFinding`, changing no card.

**REQ-JEVN-002** (Ubiquitous) A Jev finding shall carry a **third** source constant, distinct from both `BacklogSourceMechanical` and `BacklogSourceAgent`.

**REQ-JEVN-003** (Unwanted) A Jev finding shall not be written with `Source: agent`. `HasAgentFindingForPair` is the predicate behind the `machine-only` mark, and that mark records the absence of an agent-sourced record for a pair — never that a review took place. Writing a Jev finding as `agent` would make the queue claim a review happened when none did.

**REQ-JEVN-004** (Ubiquitous) `HasAgentFindingForPair` shall continue to return false for a pair whose only finding is Jev-sourced.

**REQ-JEVN-005** (Ubiquitous) The rendering of a Jev finding shall follow the existing source-conditioned rules in `todoFindingLine`: a score is printed only where a measurement was taken, so a Jev finding's probability shall be rendered in a form a reader cannot mistake for the text analyser's similarity score.

**REQ-JEVN-006** (Ubiquitous) The dedup precedence between a Jev finding and a mechanical or agent finding naming the same unordered pair shall be stated explicitly and enforced through the existing `AppendFindingOnce` / `SamePairAs` paths.

**REQ-JEVN-007** (Unwanted) No code path shall write a card field as a consequence of a Jev finding.

### C.2 — Consumer A: lane-question routing

**REQ-JEVN-008** (Event-driven) **When** a lane's blocking question is routed and the capability is enabled, **the system shall** ask one Choice question naming the decision owner (lead / operator / worker) and two Noul questions — whether a new measurement is needed first, and whether the decision is cheap to reverse.

**REQ-JEVN-009** (Ubiquitous) The three questions shall be batched into one request over one state.

**REQ-JEVN-010** (Ubiquitous) The routing answer shall be surfaced to the reader as a signal and shall not itself dispatch, answer, or close anything.

**REQ-JEVN-011** (Capability gate) **Where** the Choice answer is `operator`, or the Choice confidence falls below this consumer's fitted threshold, **the system shall** treat the item as operator-owned.

### C.3 — Consumer B: skill suggestion

**REQ-JEVN-012** (Event-driven) **When** a turn's intent is being routed and the capability is enabled, **the system shall** issue two requests: a wide rank over all skills paired with a Noul asking whether the turn needs a skill at all, then a rerank of the top three under fuller text.

**REQ-JEVN-013** (Ubiquitous) The suggestion shall be presented to the orchestrator as a ranked signal; the `/moai` intent router shall retain its existing selection authority unchanged.

**REQ-JEVN-014** (Capability gate) **Where** the "needs a skill at all" Noul answers negatively above this consumer's fitted threshold, **the system shall** suppress the ranked list rather than present a best-of-nothing.

---

## §D. Exclusions

### Out of Scope — the gate itself

- Building the measurement harness, the constant-answer baseline, the two language arms, or the threshold-fitting procedure. All owned by `SPEC-JEV-OPTIN-MEASURE-001`. This SPEC *passes* that gate; it does not build it.
- Shipping any consumer whose measurement does not beat its baseline. The withheld consumer's absence is recorded as a decision, not an omission.

### Out of Scope — card schema and queue authority

- Changing `BacklogItem`'s fields, JSON tags, or the frozen per-item contract. The third source constant extends `BacklogFinding.Source`; no card field changes.
- Changing how cards are issued, who may issue them, or the queue's producer rules.
- Any folding, reordering, dropping, or editing of a card in response to a finding. The finding record's structural property — that no code path writes a card field as a consequence of one — is inherited and preserved.

### Out of Scope — premise-death judgment

- Using Jev to judge whether a stale card's premise is dead. Measured and rejected; `moai todo triage` stays model-free, and no requirement here routes a model answer into it. The rejection record is owned by `SPEC-JEV-GOAL-DIST-001`.

### Out of Scope — authority transfer

- Giving any of the three consumers the ability to dispatch, answer, close, or select. Routing surfaces a classification; skill suggestion surfaces a ranking; the intent router keeps its selection authority unchanged.
- The `/moai goal --auto` seats that consume routing, owned by `SPEC-JEV-GOAL-DIST-001`.

---

## §E. Open Questions

| # | Question | Status |
|---|---|---|
| N1 | The dedup precedence rule in `plan.md` §B1 | OPEN — carried from the pre-split SPEC as a **proposal**, not a settled decision. Needs a decision before M4 implementation. |
| Q3 | Labelled-set size per consumer | OPEN (owned by `SPEC-JEV-OPTIN-MEASURE-001`) — determines each gate's sample size before this SPEC's consumers can pass it. |
