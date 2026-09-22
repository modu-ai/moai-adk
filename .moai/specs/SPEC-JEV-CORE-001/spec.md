---
id: SPEC-JEV-CORE-001
title: "Jev core capability — package, config gate, credential, fail-open"
version: "0.2.0"
status: completed
created: 2026-09-20
updated: 2026-09-21
author: manager-spec
priority: P2
phase: "v3.1.0 target"
module: "internal/jev + internal/jevcred + internal/config"
lifecycle: spec-anchored
tags: "jev, typesafe, system-one, core, credential, fail-open"
tier: L
---

# SPEC-JEV-CORE-001

## HISTORY

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-09-21 | 0.2.0 | **Clarifying amendment to a `completed` SPEC, card t1066.** REQ-JEVC-011 and REQ-JEVC-012 were reconciled with `SPEC-JEV-CONSUMERS-001` REQ-JEVN-001, which authorises an inert finding append during `moai todo add` and which a plain reading of both clauses forbade. REQ-JEVC-011 gains a bounded inert-sibling-record carve-out with three binding conditions; REQ-JEVC-012 gains a definition of "for" that separates *reaching* a listed act from merely *co-occurring* with one. No requirement was removed, no id renumbered, no AC mapping changed, and the scope this SPEC closed is unchanged. **This amendment postdates the sync commit `c8731f965`** recorded in `progress.md` §E.4, so the SPEC body no longer byte-matches the tree that was synced; `sync_commit_sha` is NOT rewritten, because it records which tree was synced and not which text currently stands. `status` stays `completed`. | manager-spec |
| 2026-09-20 | 0.1.0 | Split from SPEC-JEV-INTEGRATION-001 (card t1020) on the M1 seam, after that SPEC exceeded the Tier L REQ/AC ceilings (68/56 against 25/25). Carries M1: the single implementation, the config gate, the credential, the display-only invariant, fail-open, and the doctor check. Requirements re-numbered `REQ-JEVC-*`; no requirement was dropped. | manager-spec |

## Position in the chain

First of four. Nothing precedes it.

`SPEC-JEV-CORE-001` → `SPEC-JEV-OPTIN-MEASURE-001` → `SPEC-JEV-CONSUMERS-001` → `SPEC-JEV-GOAL-DIST-001`

This SPEC ships a capability with **no consumers**. That is deliberate: the package, the gate, the credential, and the fail-open contract are what every later SPEC rests on, and shipping them alone makes the display-only invariant testable before anything can violate it.

---

## §A. Problem Statement

TypeSafe **Jev** is a System One judgment model: it answers a typed question about a supplied state and returns a typed answer (Choice / Noul / Score) with a probability. It generates no text. Today it reaches this repository only as an uncommitted local script set in the primary checkout's working tree — `scripts/jev/` does not exist on `develop` (verified: `ls scripts/` shows no `jev` entry), and no Go file under `internal/`, `pkg/`, or `cmd/` mentions `typesafe` or `jev` (verified: 0 matches, against a `glmcred` positive control returning 5 files).

A capability living only in one machine's uncommitted working copy cannot be used by anyone else, cannot be tested, and cannot be cited — an uncommitted working copy is not canonical and may not be quoted as one. Every operator other than the one holding that working tree has no Jev at all.

Making it reachable is not enough on its own. A judgment model wired into a workflow without a boundary becomes a decision-maker by default, because a displayed answer is the cheapest thing in view when someone has to act. So the boundary ships in the same SPEC as the capability: display-only, default-off, fail-open, and unreachable from any irreversible judgment.

## §B. Goal

One Go package owning the single call path; one config gate defaulting to off; one credential outside the repository and outside the settings schema; a fail-open contract in which no absence becomes a failure; and a `moai doctor` line that reports readiness without asking the model anything.

---

## §C. Requirements (GEARS)

### C.1 — The single implementation

**REQ-JEVC-001** (Ubiquitous) The `internal/jev` package shall be the single implementation of the TypeSafe System One call path: request construction, credential read, HTTP transport to `POST https://api.typesafe.ai/v1/systemone`, response decoding, and usage accounting.

**REQ-JEVC-002** (Ubiquitous) Every consumer — CLI, MCP tool wrapper, and web surface — shall reach the model only through `internal/jev`; no second client implementation shall exist.

**REQ-JEVC-003** (Ubiquitous) The package shall send a **pinned** model id. The alias `jev-latest` shall not appear in any request this package constructs.

**REQ-JEVC-004** (Ubiquitous) The package shall record, per call, the request's input-token count and the pinned model id, so a later cost or provenance question is answerable from the record rather than from recollection.

### C.2 — Fail-open

**REQ-JEVC-005** (Capability gate) **Where** the credential file is absent, unreadable, or empty, **the `internal/jev` package shall** return a typed unavailable result rather than an error that propagates to a caller's exit status.

**REQ-JEVC-006** (Event-detected) **When** the transport returns HTTP 401, 429, or 529, or the network is unreachable, **the `internal/jev` package shall** return a typed unavailable result carrying the observed condition.

**REQ-JEVC-007** (Ubiquitous) Every consumer of an unavailable result shall emit at most one notice line, exit 0, and continue the workflow it was performing. Jev absence is graceful degradation, never failure.

**REQ-JEVC-008** (Capability gate, unwanted) **Where** a call failed non-idempotently or ambiguously, **the `internal/jev` package shall not** retry it. Retry shall be limited to calls whose repetition is provably free of side effects.

### C.3 — Request shape

**REQ-JEVC-009** (Ubiquitous) The package shall keep a request's `state` payload and its longest question within 32k tokens, and the whole request within 64k tokens, refusing to send a request that exceeds either bound rather than truncating it silently.

**REQ-JEVC-010** (Ubiquitous) Where several independent questions are asked over one state, the package shall batch them into a single request.

### C.4 — Display-only invariant

**REQ-JEVC-011** (Ubiquitous) **[AMENDED 2026-09-21 — v0.2.0; see HISTORY]** A Jev answer shall not mutate the backlog queue, a card's state, a card's text, a file, a branch, or a merge. The capability is display-only.

**What "mutate the backlog queue" means.** It means change a card — its existence, its identity, its ordering, its state, or its text — or change what the queue *asserts* about a card's disposition. It does **not** mean "write the queue file at all", and the enumerated "a file" above does not reach the queue file on the single path carved out below.

**The carve-out: an inert sibling record.** A record appended *alongside* the objects it names, rather than *into* them, is outside this prohibition — but **only while all three conditions hold together**:

> **(i) Structural inertness.** No code path reads the record back to write a field of any object it names, and this is a property of the code rather than a convention — the writes in question would each require code that does not exist (`BacklogFinding`, `internal/kanban/backlog_store.go:145-151`).
>
> **(ii) Distinguishability.** The record is distinguishable at every surface from a mechanical measurement and from a human or agent judgement, per REQ-JEVC-013 — its own source constant and its own render form, never a reuse of an existing one.
>
> **(iii) Nothing selects on it.** No decision, gate, verdict, mark, or disposition selects on the record. **Where any predicate or surface begins to select on it, the carve-out lapses and the append is a mutation again** — the carve-out is conditional on a property that a later change can remove, and its removal is what a reviewer watches for.

A consumer relying on this carve-out is authorised **by name** in a downstream SPEC (today: `SPEC-JEV-CONSUMERS-001` REQ-JEVN-001, the card-admission near-duplicate finding). This is not a general permission, and a consumer that is not named does not have it.

**The reading that was rejected, and why it is recorded rather than deleted.** "Mutate the backlog queue" could be read as "write the queue file at all", and that reading had real textual support before this amendment: the enumeration lists "a file", and the queue file is a file. Under it, `SPEC-JEV-CONSUMERS-001` would be building something this SPEC forbids. It is rejected on three grounds. First, the queue's own governing doctrine already separates the two acts: `.claude/rules/moai/workflow/kanban-dispatch.md` § Entry into the board is an operator act states that analysis "records a relation between two cards" and that "analysis changes exactly one thing on its own authority — it refuses the admission of a card whose normalized text is identical to one already queued or picked" — so recording a relation is, by that doctrine, not a queue change, and the one act that is remains mechanical and untouched. Second, the pre-existing mechanical analyser already appends a finding on `todo add` (`internal/cli/todo_analysis.go:58-76`) and no SPEC in this chain treats that as a queue mutation; the strict reading would retroactively classify the non-Jev baseline as one. Third, the strict reading forbids the act on the basis of *which file bytes changed* rather than *what the change can cause*, which is the axis every other clause in §C.4 is written on. The rejection is recorded here rather than argued away so a later reader can disagree with the choice by reading it.

**REQ-JEVC-012** (Unwanted) **[AMENDED 2026-09-21 — v0.2.0; see HISTORY]** The system shall not consult Jev — not as a decision, and **not as an input** — for any of: a completion verdict, a merge approval, a `moai todo` or `moai gtd` mutation, an operator gate, a user-surface behaviour change, or a CodeRabbit slot-wait adjudication.

**"For" names the role the answer plays, not the moment it is asked.** A consultation is *for* a listed act when its answer **can reach** that act — as the decision, as an input to it, or as something a surface presenting it selects on. A consultation that happens **during a command that also performs a listed act, while its answer cannot reach that act**, is not a consultation *for* it.

**The discriminator is mechanical, so it can be checked rather than argued.** Both must hold:

> **(1) Determined first.** The listed act is fully determined before the consultation is constructed — its inputs read, its branch taken.
> **(2) Unread after.** No value derived from the answer is read on the act's path.

**The one authorised instance today, with its coordinates.** On the card-admission path (`appendAnalyzedCard`, `internal/cli/todo_analysis.go:38`) the admission decision is complete before the consultation exists: the identical-text refusal is evaluated at `:40-44`, the card is appended at `:53`, and the mechanical finding branches run at `:58-76` — all of it decided by `ClassifyCardText` alone. The consultation is constructed at `:90` and **its return value is discarded at the call site**. The queue mutation is therefore not informed by the model answer in either direction, and both limbs of the discriminator hold. The standing guard on this boundary is `TestJevCallPath_UnreachableFromDecisionSurfaces` (`internal/cli/doctor_jev_test.go:272`), which scans the named decision surfaces for same-package references as well as imports and carries exactly one authorised entry, declared by name and by citation.

**The reading that was rejected, and why it is recorded rather than deleted.** "For" could be read as "during" — any consultation occurring inside a command that performs a listed act. That reading also had textual support, since the clause names the act (`a moai todo ... mutation`) rather than the role, and under it `SPEC-JEV-CONSUMERS-001` REQ-JEVN-001 would be forbidden outright. It is rejected on two grounds. First, it forbids on the basis of **co-location in a command**, a property no mechanism can distinguish from the authorised case: a call sequenced after a completed decision and a call feeding that decision sit in the same function and are told apart only by dataflow, which is exactly what the discriminator above reads. Second, every sibling in the enumeration — a completion verdict, a merge approval, an operator gate, a slot-wait adjudication — is a *judgement whose input is being protected*, not a command whose runtime is being fenced; reading one member as a time window and the rest as decision inputs makes the clause inconsistent with itself.

**REQ-JEVC-013** (Ubiquitous) A Jev answer surfaced anywhere shall be labelled as a model-produced signal, distinguishable by a reader from a mechanical measurement and from a human or agent judgement.

**REQ-JEVC-014** (Capability gate) **Where** Jev is disabled, uncredentialed, or unreachable, **every consumer shall** produce output identical to its pre-Jev behaviour apart from at most one notice line.

### C.5 — Configuration gate

**REQ-JEVC-015** (Ubiquitous) The capability shall be gated on `workflow.jev.enabled` inside the existing `.moai/config/sections/workflow.yaml`. No new config section file shall be created.

**REQ-JEVC-016** (Ubiquitous) The shipped template default for `workflow.jev.enabled` shall be `false`, with `internal/config/defaults.go` as the compiled source of truth, following the shape of the existing opt-in switches in that file (`codex.review_gate.enabled`, `multi.review_gate.enabled`, `slot_lease.enabled`).

**REQ-JEVC-017** (State-driven) **While** `workflow.jev.enabled` is false, **the system shall** make no network call to the TypeSafe endpoint and construct no request.

### C.6 — Credential

**REQ-JEVC-018** (Ubiquitous) The API credential shall live at `~/.moai/.env.typesafe` at file mode 0600, outside the repository, read and written through one package modelled on `internal/glmcred` — including that package's tightening of a pre-existing wider mode on write.

**REQ-JEVC-019** (Unwanted) The credential shall not enter `settings.AllFields()`, so no generic schema-walking loop can read, render, or write it.

**REQ-JEVC-020** (Ubiquitous) A view of the credential shall disclose only a `configured` boolean and, for a credential longer than four characters, its final four characters — never the credential itself, and never any part of a credential of four characters or fewer.

**REQ-JEVC-021** (Capability gate) **Where** a request payload would carry text, **the `internal/jev` package shall** apply a secret-screening step before sending, and refuse to send a payload in which a credential-shaped token is detected.

### C.7 — Readiness reporting

**REQ-JEVC-022** (Ubiquitous) `moai doctor` shall carry one Jev check reporting, without sending a judgment request, whether the capability is enabled, whether a credential is present, and whether the endpoint is reachable.

---

## §D. Exclusions

### Out of Scope — consumers

- Any caller of the package. This SPEC ships the capability with zero consumers; near-duplicate marking, lane-question routing, and skill suggestion belong to `SPEC-JEV-CONSUMERS-001`.
- Any measurement of model accuracy. The measurement harness and its gates belong to `SPEC-JEV-OPTIN-MEASURE-001`.

### Out of Scope — user-facing surfaces

- The `moai init` wizard question and the `moai web` settings section. Both belong to `SPEC-JEV-OPTIN-MEASURE-001`, which also owns the shared persistence path.
- The MCP tool wrapper and the reference skill, which belong to `SPEC-JEV-GOAL-DIST-001`.

### Out of Scope — premise-death judgment

- Using Jev to judge whether a stale backlog card's premise is dead. Measured and rejected; the record and the exclusion are owned by `SPEC-JEV-GOAL-DIST-001` so they land with the documentation surfaces. Nothing in this SPEC routes a model answer into `moai todo triage`.

### Out of Scope — authority

- Any path by which a Jev answer reaches a completion verdict, a merge approval, a queue mutation, an operator gate, a user-surface behaviour change, or a CodeRabbit slot-wait adjudication. REQ-JEVC-012 forbids it as an input, not merely as a decision.
- **Still excluded after the v0.2.0 amendment:** a Jev answer that a queue mutation, a verdict, a gate, or a mark **selects on**. The REQ-JEVC-011 carve-out covers an inert sibling record and nothing else, and condition (iii) makes the carve-out lapse the moment anything selects on the record. The amendment narrows the wording; it does not widen the boundary.

---

## §E. Open Questions

| # | Question | Status |
|---|---|---|
| Q2 | Is the pinned model id compiled, or operator-visible configuration? | OPEN — decides whether `workflow.jev` is a bare `enabled` flag or a block. Affects REQ-JEVC-003 and REQ-JEVC-015. |
| Q4 | Is a credential-reveal route wanted? | OPEN — the GLM precedent has one (`glmKeyRevealPath`); REQ-JEVC-020 is satisfied without it. |
