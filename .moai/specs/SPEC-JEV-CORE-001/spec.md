---
id: SPEC-JEV-CORE-001
title: "Jev core capability — package, config gate, credential, fail-open"
version: "0.1.0"
status: completed
created: 2026-09-20
updated: 2026-09-20
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

**REQ-JEVC-011** (Ubiquitous) A Jev answer shall not mutate the backlog queue, a card's state, a card's text, a file, a branch, or a merge. The capability is display-only.

**REQ-JEVC-012** (Unwanted) The system shall not consult Jev — not as a decision, and **not as an input** — for any of: a completion verdict, a merge approval, a `moai todo` or `moai gtd` mutation, an operator gate, a user-surface behaviour change, or a CodeRabbit slot-wait adjudication.

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

---

## §E. Open Questions

| # | Question | Status |
|---|---|---|
| Q2 | Is the pinned model id compiled, or operator-visible configuration? | OPEN — decides whether `workflow.jev` is a bare `enabled` flag or a block. Affects REQ-JEVC-003 and REQ-JEVC-015. |
| Q4 | Is a credential-reveal route wanted? | OPEN — the GLM precedent has one (`glmKeyRevealPath`); REQ-JEVC-020 is satisfied without it. |
