---
id: SPEC-JEV-OPTIN-MEASURE-001
title: "Jev opt-in surfaces and the measurement gate"
version: "0.1.0"
status: draft
created: 2026-09-20
updated: 2026-09-20
author: manager-spec
priority: P2
phase: "v3.1.0 target"
module: "internal/cli/wizard + internal/web + internal/settings"
lifecycle: spec-anchored
tags: "jev, typesafe, opt-in, wizard, web-console, measurement, baseline"
tier: L
depends_on: [SPEC-JEV-CORE-001]
---

# SPEC-JEV-OPTIN-MEASURE-001

## HISTORY

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-09-20 | 0.1.0 | Split from SPEC-JEV-INTEGRATION-001 (card t1020) on the M2+M3 seam. Carries the two opt-in entrances, the shared persistence path, the measurement harness, and the question-design rules. Operator decision folded in: the wizard question joins the **init-only** set, so the setting is not reachable through reconfigure — three requirements encode that and its consequence. | manager-spec |

## Position in the chain

Second of four. Depends on `SPEC-JEV-CORE-001` for the package, the config key, and the credential reader.

`SPEC-JEV-CORE-001` → **`SPEC-JEV-OPTIN-MEASURE-001`** → `SPEC-JEV-CONSUMERS-001` → `SPEC-JEV-GOAL-DIST-001`

Two concerns share this SPEC because they share a precondition and nothing else: both need the core capability, and neither needs a consumer. Surfaces let a human turn the thing on; the measurement harness decides whether any consumer is allowed to exist. Shipping them together means the first consumer SPEC arrives with both its switch and its gate already in place.

---

## §A. Problem Statement

`SPEC-JEV-CORE-001` ships a capability nobody can turn on and nobody has measured.

**Nobody can turn it on.** The config key exists and defaults to false. Editing YAML by hand is not an opt-in surface: it is undiscoverable, it has no place to state what enabling costs, and it bypasses the one thing an opt-in must carry — a statement, at the moment of choice, that enabling sends card text and user-request text to a third-party server.

**Nobody has measured it.** The capability's whole value proposition is that a typed judgment beats no signal. That is an empirical claim, and it has already been false once: judging whether a stale card's premise is dead was measured at 58.9% 2-class accuracy against a constant "always alive" answer scoring 75.0%. A consumer shipped without its own measurement is a consumer shipped on the assumption that the last failure will not repeat.

The measurement's hard part is not running it. It is knowing what number to compare against. Raw accuracy is unreadable without the base rate: on the rejected task, a measurement-design repair moved 3-class accuracy from 31.5% to 47.6% — a sixteen-point gain that looked like success while the constant baseline sat at 75.0% the whole time, untouched and unbeaten.

## §B. Goal

Two entrances to one persistence path, each stating the privacy cost where the choice is made; and a measurement harness that produces, per consumer, a labelled answer set scored against a constant-answer baseline in two language arms, under a pinned model id — with the verdict binding.

---

## §C. Requirements (GEARS)

### C.1 — Opt-in surfaces

**REQ-JEVO-001** (Ubiquitous) The `moai init` wizard shall carry exactly one Jev question, and the `moai web` settings screen shall carry one Jev section holding an enable toggle and a credential field.

**REQ-JEVO-002** (Ubiquitous) Both entrances shall write through **one** shared persistence path in the neutral `internal/settings` package; no parallel writer shall exist.

**REQ-JEVO-003** (Ubiquitous) Both entrances shall state, at the point of choice, that enabling the capability sends card text or user-request text to a third-party server.

**REQ-JEVO-004** (Ubiquitous) The wizard question and the web section shall carry strings in all four locales (ko / en / ja / zh).

### C.2 — Init-only placement and its consequence

**REQ-JEVO-005** (Ubiquitous) The Jev question shall join the **init-only** question set (`InitQuestions`), bringing that set to five questions. It shall NOT be added to `DefaultQuestions` and shall NOT appear in `ReconfigureQuestions`.

**REQ-JEVO-006** (Ubiquitous) Because the question is init-only, the Jev setting shall NOT be changeable through `moai update --reconfigure`. The `moai web` settings screen shall be the only path to change the setting after initialization, and the documentation shall state this plainly rather than leaving it to be discovered.

**REQ-JEVO-007** (Ubiquitous) The wizard question text, or the `moai init` completion output, shall name `moai web` as the place the setting can later be changed — so a user who initializes from the terminal and never opens the console can still find the switch.

### C.3 — Measurement gate

**REQ-JEVO-008** (Ubiquitous) Each consumer shall carry its own measurement: a labelled answer set drawn from this repository's data, scored against a **constant-answer baseline** for that consumer's question shape.

**REQ-JEVO-009** (Unwanted) A consumer whose measured accuracy does not beat its own constant-answer baseline shall not be shipped. The measurement's verdict is binding, not advisory.

**REQ-JEVO-010** (Ubiquitous) Every measurement shall run two arms — the card or request text in its Korean original, and the same text in English translation — and shall record the per-arm result and the delta.

**REQ-JEVO-011** (Unwanted) This SPEC shall not change the card schema, the card issuance path, or the language any card is written in. The English arm is a measurement, not a migration.

**REQ-JEVO-012** (Ubiquitous) A confidence threshold shall be fitted per consumer on this repository's measured data.

**REQ-JEVO-013** (Unwanted) A threshold shall not be copied from vendor documentation — the figures `0.6` and `0.85` appearing there are illustrative — and shall not be transferred between a Noul question and a Choice question.

**REQ-JEVO-014** (Ubiquitous) Every measurement result shall cite the pinned model id it was taken under.

**REQ-JEVO-015** (Ubiquitous) A measurement reporting an absence — a zero-hit, a no-difference, a null result — shall be accompanied by a positive control demonstrating the measuring apparatus fires on that path.

### C.4 — Question design under known model weaknesses

**REQ-JEVO-016** (Ubiquitous) All counting, arithmetic, numeric proximity, date ordering, and SHA comparison shall be performed in Go, and the results passed to the model as named JSON fields. The model shall not be asked to compute them.

**REQ-JEVO-017** (Ubiquitous) Every Choice question shall carry an explicit no-match option.

**REQ-JEVO-018** (Ubiquitous) A request's state shall carry only the fields its questions read, and shall favour identifiers and English-language fields over long prose.

**REQ-JEVO-019** (Unwanted) The system shall not treat text inside a state payload as instruction. State is untrusted data.

**REQ-JEVO-020** (Unwanted) Consumer logic shall not assume `P(yes)` equals `1 − P(no)`; where both are needed, both shall be read.

**REQ-JEVO-021** (Ubiquitous) Questions shall be written to survive literal reading: negation, scoping words, and implied conditions shall be made explicit rather than left to inference, and multi-hop indirection shall be resolved in Go before the question is asked.

---

## §D. Exclusions

### Out of Scope — an init flow inside the console

- Any `moai init` flow inside `moai web`. The console has no init route — its route set is overview / kanban / monitor / todo / settings / events / save / specs / profile CRUD / GLM-key reveal / shutdown — and designing one is a different piece of work.
- A Jev-driven write path anywhere in the console. The Jev section reads and persists configuration; it displays no judgment that mutates state.

### Out of Scope — consumers

- Near-duplicate marking, lane-question routing, and skill suggestion. This SPEC builds the gate each must pass; `SPEC-JEV-CONSUMERS-001` builds the consumers.
- Any call path that asks the model a question in production. The harness asks questions against a labelled set; nothing in this SPEC wires a question into a workflow.

### Out of Scope — card language and schema

- Switching card text to English. The English measurement arm records a delta so a later card can decide; this SPEC decides nothing about it, and changes no card, no schema, and no issuance path.

### Out of Scope — premise-death judgment

- Re-running the rejected premise-death experiment under a different gate, prompt, or language arm. The gate sweep was already flat across 0.30-0.80 and the English arm already measured; a new run needs a new reason, not a new attempt. The rejection record itself is owned by `SPEC-JEV-GOAL-DIST-001`.

### Out of Scope — the core package

- The client, the credential reader, the config key, the fail-open contract, and the doctor check, all owned by `SPEC-JEV-CORE-001`.

---

## §E. Open Questions

| # | Question | Status |
|---|---|---|
| Q1 | Which wizard constructor takes the question | **RESOLVED** — operator decision: init-only (`InitQuestions`, five questions). Encoded as REQ-JEVO-005 with its consequence in REQ-JEVO-006 and REQ-JEVO-007. |
| Q3 | Labelled-set size per consumer | OPEN — depends on each consumer's base rate, which this SPEC's harness measures. The rejected task used 124 cards; that is a precedent, not a target. |
