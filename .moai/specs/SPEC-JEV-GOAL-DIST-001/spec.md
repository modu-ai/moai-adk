---
id: SPEC-JEV-GOAL-DIST-001
title: "Jev goal --auto seats, MCP wrapper, and distribution"
version: "0.1.0"
status: draft
created: 2026-09-20
updated: 2026-09-20
author: manager-spec
priority: P2
phase: "v3.1.0 target"
module: ".claude/skills/moai + .claude/rules/moai/core + internal/template/templates"
lifecycle: spec-anchored
tags: "jev, typesafe, goal-auto, mission-governor, mcp, template, negative-result"
tier: M
depends_on: [SPEC-JEV-CONSUMERS-001]
---

# SPEC-JEV-GOAL-DIST-001

## HISTORY

| Date | Version | Change | Author |
|---|---|---|---|
| 2026-09-20 | 0.1.0 | Split from SPEC-JEV-INTEGRATION-001 (card t1020) on the M7+M8 seam. Carries both `/moai goal --auto` seats, the MCP tool wrapper, the reference skill, the template and documentation surfaces, and the preserved negative result. Requirements re-numbered `REQ-JEVG-*`; none dropped. Tier M: 14 requirements and 12 acceptance criteria against a 16/16 ceiling. | manager-spec |

## Position in the chain

Last of four. Depends on `SPEC-JEV-CONSUMERS-001` twice over: seat (i) consumes Consumer A's routing, and the MCP tool counts cannot be settled until the shipped consumer set is known.

`SPEC-JEV-CORE-001` → `SPEC-JEV-OPTIN-MEASURE-001` → `SPEC-JEV-CONSUMERS-001` → **`SPEC-JEV-GOAL-DIST-001`**

Two concerns share this SPEC because both are downstream of everything else: the autonomous-loop seats can only be wired once routing exists, and the distribution surfaces can only be counted once the tool set is final.

---

## §A. Problem Statement

Three loose ends remain after the capability, its switch, its gate, and its consumers are in place.

**The autonomous loop cannot use what the consumers produce.** Inside `/moai goal --auto`'s sealed-scope loop, a lane that blocks with a question becomes a persisted blocked result and the loop stops. Some of those questions are the lead's own and cheap to reverse — exactly the class Consumer A's routing classifies — and stopping the loop for them costs an operator round-trip that the classification could have avoided. Separately, `mission-governor` evaluates a sealed snapshot and returns one bounded decision; a Noul answer is one more piece of evidence that snapshot could carry.

Both wirings are places where a display-only capability could quietly become a decision-maker, because an autonomous loop has no human in it to notice. The seats therefore ship with their boundaries stated, not implied.

**The capability is unreachable from anything but Go.** An agent that would benefit from a typed judgment has no path to one; the MCP surface is where that path belongs, and it must be a wrapper rather than a second implementation.

**The rejected use has no written home.** Judging whether a stale card's premise is dead was measured and rejected — dead-call precision 14/48 = 29.2% against a 25% base rate, 2-class accuracy 58.9% against 75.0% for a constant "always alive" answer, and an English control on 40 cards reaching 67.5%, still below the constant. That rejection currently lives in prose and in a gitignored local evidence file. Nothing in the product records it, so nothing stops the next person re-litigating it.

## §B. Goal

Two `--auto` seats that move no authority; one thin MCP wrapper over the existing implementation; one reference skill carrying question-design rules only; template and documentation surfaces updated in both copies; and the negative result written down where it survives.

---

## §C. Requirements (GEARS)

### C.1 — `/moai goal --auto` seats

**REQ-JEVG-001** (Event-driven) **When** a lane blocks with a question inside the sealed-scope loop and the capability is enabled, **the lead shall** classify the decision owner through Consumer A's routing, answer items that are both lead-owned and cheap-to-reverse, and leave operator-owned items as a persisted blocked result exactly as today.

**REQ-JEVG-002** (Unwanted) Seat (i) shall not introduce a user question inside the sealed-scope loop. The single approval before the loop remains the only one.

**REQ-JEVG-003** (Ubiquitous) A Jev Noul answer supplied to `mission-governor` shall be an **auxiliary input signal**, recorded as a separate item in the governance receipt.

**REQ-JEVG-004** (Unwanted) Seat (ii) shall not alter the governor's authority, its read-only scope, its output shape, or the receipt's existing binding fields.

**REQ-JEVG-005** (Unwanted) A Jev answer shall not be an element of a completion predicate, nor of the landed-ancestry or authoritative-readback evidence a completion receipt binds.

### C.2 — MCP wrapper and reference skill

**REQ-JEVG-006** (Ubiquitous) The moai-mcp tool wrapper shall be a thin shell over `internal/jev`, carrying no second implementation of the call path.

**REQ-JEVG-007** (Ubiquitous) `.claude/rules/moai/core/moai-mcp-tools.md` and its template mirror shall both be updated, including both tool-count figures each carries.

**REQ-JEVG-008** (Ubiquitous) One reference skill shall carry question-design rules only, and shall not carry the call path.

### C.3 — Template and distribution

**REQ-JEVG-009** (Ubiquitous) Every new file under `.claude/` or `.moai/` shall land in `internal/template/templates/` first.

**REQ-JEVG-010** (Unwanted) Template content shall not carry a SPEC ID, a card id, a date, a price, a measurement figure, or a reference to a local-only file.

**REQ-JEVG-011** (Capability gate) **Where** a file under `internal/template/templates/.claude/agents/moai/` changes, **the build shall** require `make agents-emit`; **where** a command source changes, **the build shall** require `make commands-emit`.

**REQ-JEVG-012** (Ubiquitous) The disposition of the uncommitted `scripts/jev/` working copy shall be stated, with the Go package as the canonical implementation.

### C.4 — Preserved negative result

**REQ-JEVG-013** (Ubiquitous) The system shall record that using Jev to judge whether a stale card's premise is dead was measured and rejected, with the measured figures preserved.

**REQ-JEVG-014** (Unwanted) `moai todo triage` shall remain model-free. No requirement in this SPEC chain shall route a model answer into it.

---

## §D. Exclusions

### Out of Scope — governor authority

- Changing `mission-governor`'s scope, tool set, output shape, or the governance receipt's existing binding fields. Seat (ii) adds one recorded item and changes nothing else.
- Any path by which a Jev answer reaches a completion predicate, landed-ancestry evidence, or authoritative-readback evidence. REQ-JEVG-005 forbids it, and those are the irreversible judgments the whole chain is bounded against.
- Adding a user question inside the `--auto` sealed-scope loop. The single approval before the loop stays the only one.

### Out of Scope — re-litigating the negative result

- Re-running the premise-death experiment under a different gate, prompt, or language arm. The gate sweep was already flat across 0.30-0.80 and the English arm already measured; a new run needs a new reason, not a new attempt.
- Any change to `moai todo triage`, which stays model-free.
- Claiming the figures were re-measured here. `.moai/reports/t943/verdict.md` is gitignored local evidence in the primary checkout, possibly absent from any given tree; this SPEC cites it as the origin and re-measures nothing.

### Out of Scope — earlier chain members

- The client, credential, config gate, fail-open contract, and doctor check, owned by `SPEC-JEV-CORE-001`.
- The wizard question, the web settings section, the shared persistence path, and the measurement harness, owned by `SPEC-JEV-OPTIN-MEASURE-001`.
- The three consumers themselves, owned by `SPEC-JEV-CONSUMERS-001`.

### Out of Scope — script distribution

- Distributing `scripts/jev/` to user projects. The scripts have no template mirror, reach no user project, and are superseded by the Go package. Nothing in this SPEC deletes anything from anyone's working tree.

---

## §E. Acceptance Criteria

Twelve criteria (AC-JEVG-001 … AC-JEVG-012), in `acceptance.md`, together with the quality gates and the Definition of Done. Tier M carries acceptance criteria in their own artifact rather than inline.

## §F. Open Questions

None owned by this SPEC. Q2 and Q4 are owned by `SPEC-JEV-CORE-001`; Q3 and R1 by `SPEC-JEV-OPTIN-MEASURE-001`; N1 and N2 by `SPEC-JEV-CONSUMERS-001`.
