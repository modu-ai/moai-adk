---
id: SPEC-JEV-GOAL-DIST-001
title: "Jev goal --auto governor seat, MCP wrapper, and distribution"
version: "0.2.0"
status: completed
created: 2026-09-20
updated: 2026-09-22
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
| 2026-09-20 | 0.1.0 | Split from SPEC-JEV-INTEGRATION-001 (card t1020) on the M7+M8 seam. Carried both `/moai goal --auto` seats, the MCP tool wrapper, the reference skill, the template and documentation surfaces, and the preserved negative result. Requirements re-numbered `REQ-JEVG-*`; none dropped. Tier M: 14 requirements and 12 acceptance criteria against a 16/16 ceiling. | manager-spec |
| 2026-09-22 | 0.2.0 | Plan-phase iter-2 revision answering the plan-audit FAIL (0.68 against the Tier M 0.80 threshold; `.moai/reports/SPEC-JEV-GOAL-DIST-001/plan-audit-iter1.md`). **Seat (i) is withdrawn from the deliverable set** (§C.1 disposition): its sole producer `SPEC-JEV-CONSUMERS-001` M5 is recorded blocked with question N2 OPEN, so the seat's carrier is the recorded disposition, not a shipped seat — `REQ-JEVG-001` restated as the unchanged-behaviour guarantee, `AC-JEVG-001`/`002` restated as regression guards, a fourth Out-of-Scope section added. **The MCP wrapper is gate-coupled**: `REQ-JEVG-006`/`007` amended to hold the wrapper inert behind the `workflow.jev.enabled` default-false gate with gated-unavailable presentation while the fitness gate stands unrun; `AC-JEVG-013` added. **Record destinations named** (`docs/jev-negative-results.md`, tracked, non-template): `AC-JEVG-012` fixed and `AC-JEVG-014` added covering `REQ-JEVG-012` (previously a traceability orphan). Baselines pinned (`AC-001`/`004`), methods stated (`AC-005`/`008`), every AC names its REQ ids, §F carries per-question status. REQ 14 with unchanged ids; AC 12→14. Not an amendment — no prior completed version exists. | manager-spec |

## Position in the chain

Last of four. Depends on `SPEC-JEV-CONSUMERS-001` twice over, in two different modes: the MCP tool counts cannot be settled until the shipped consumer set is known (a build dependency), and the seat-(i) disposition recorded in §C.1 is a reaction to that SPEC's own recorded state — M5 blocked, N2 open — rather than a consumption of its output.

`SPEC-JEV-CORE-001` → `SPEC-JEV-OPTIN-MEASURE-001` → `SPEC-JEV-CONSUMERS-001` → **`SPEC-JEV-GOAL-DIST-001`**

Two concerns share this SPEC because both are downstream of everything else: the governor seat can only be wired once the capability's boundaries are settled, and the distribution surfaces can only be counted once the tool set is final.

---

## §A. Problem Statement

**The autonomous loop's handling of a blocking question is where the chain's capability almost reached — and where this SPEC now draws a line instead.** Inside `/moai goal --auto`'s sealed-scope loop, a lane that blocks with a question becomes a persisted blocked result and the loop stops. Consumer A's routing was meant to classify those questions so lead-owned, cheap-to-reverse ones could be answered without an operator round-trip. The measured chain state, recorded by the predecessor itself, is that this routing has no place to land: its producer milestone (CONSUMERS M5) is recorded blocked, and the host-design question N2 — which code path holds lane questions in a routable shape — is OPEN with no owner since CONSUMERS closed `completed`. This SPEC therefore withdraws seat (i) rather than restating it, and carries the unchanged-behaviour guarantee in its place (§C.1). Separately, `mission-governor` evaluates a sealed snapshot and returns one bounded decision; a Noul answer is one more piece of evidence that snapshot could carry — that seat (ii) remains in scope.

**A display-only capability could quietly become a decision-maker in an autonomous loop**, because an autonomous loop has no human in it to notice. The remaining seat ships with its boundaries stated, not implied.

**The capability is unreachable from anything but Go.** An agent that would benefit from a typed judgment has no path to one; the MCP surface is where that path belongs, and it must be a wrapper rather than a second implementation — one that stays inert and unpresented while the chain's fitness gate stands unrun.

**The rejected use has no written home.** Judging whether a stale card's premise is dead was measured and rejected — dead-call precision 14/48 = 29.2% against a 25% base rate, 2-class accuracy 58.9% against 75.0% for a constant "always alive" answer, and an English control on 40 cards reaching 67.5%, still below the constant. That rejection currently lives in prose and in a gitignored local evidence file. Nothing in the product records it, so nothing stops the next person re-litigating it.

## §B. Goal

One `--auto` seat (the governor auxiliary input) that moves no authority; the recorded disposition of the withdrawn seat (i); one thin, gate-coupled MCP wrapper over the existing implementation; one reference skill carrying question-design rules only; template and documentation surfaces updated in both copies; and the negative result written down where it survives — `docs/jev-negative-results.md`, a tracked, non-template surface.

---

## §C. Requirements (GEARS)

### C.1 — `/moai goal --auto` seats

**Seat (i) disposition (recorded 2026-09-22, plan iter-2).** Lane-question routing through Consumer A is **withdrawn** from this SPEC's deliverable set. Its sole producer, `SPEC-JEV-CONSUMERS-001` M5, is recorded blocked — CONSUMERS progress.md §E.4: *"M5 and SPEC-JEV-GOAL-DIST-001 seat (i) remain blocked and are forbidden this session"* — with host-design question N2 OPEN (CONSUMERS spec.md §E) and CONSUMERS closed `completed` with the block recorded in its sync status. A three-state conditional ("proceeds only if M5 unblocks") was considered and rejected: the enabling branch has no owner, no host design, and no scheduled path to exist, so the conditional would be a deferred promise rather than scope — the exact silently-promised seat this revision removes. Seat (i)'s future home is a future SPEC that first answers N2; until one is authored, this disposition and plan.md §B1 are the seat's only carriers. What this SPEC still owes is the behaviour guarantee: the loop's blocked-question outcome does not change.

**REQ-JEVG-001** (Event-driven) **When** a lane blocks with a question inside the sealed-scope loop, **the loop shall** produce the pre-SPEC outcome unchanged — a persisted blocked result and a stop — with no routing consumed and no item answered by classification.

**REQ-JEVG-002** (Unwanted) No change in this SPEC shall introduce a user question inside the sealed-scope loop. The single approval before the loop remains the only one.

**REQ-JEVG-003** (Ubiquitous) A Jev Noul answer supplied to `mission-governor` shall be an **auxiliary input signal**, recorded as a separate item in the governance receipt.

**REQ-JEVG-004** (Unwanted) Seat (ii) shall not alter the governor's authority, its read-only scope, its output shape, or the receipt's existing binding fields.

**REQ-JEVG-005** (Unwanted) A Jev answer shall not be an element of a completion predicate, nor of the landed-ancestry or authoritative-readback evidence a completion receipt binds.

### C.2 — MCP wrapper and reference skill

**REQ-JEVG-006** (Ubiquitous) The moai-mcp tool wrapper shall be a thin shell over `internal/jev`, carrying no second implementation of the call path, and shall stay inert behind the `workflow.jev.enabled` gate: while the gate is false — the shipped template default — the wrapper constructs no request, makes no network call, and reports gated-unavailable.

**REQ-JEVG-007** (Ubiquitous) `.claude/rules/moai/core/moai-mcp-tools.md` and its template mirror shall both be updated, including both tool-count figures each carries; and while the chain's fitness gate stands unrun — `SPEC-JEV-OPTIN-MEASURE-001` REQ-JEVO-009 (binding, unrun) with the consumers' gate-unrun state per `SPEC-JEV-CONSUMERS-001` REQ-JEVN-016 condition (iii) — no shipped or template surface shall present the tool as available, and the wrapper's gate-unrun disposition shall be recorded, naming what the gate still needs and who owns it and citing no measurement. "Shipped" carries CONSUMERS' declared reading: reachable at the shipped default.

**REQ-JEVG-008** (Ubiquitous) One reference skill shall carry question-design rules only, and shall not carry the call path.

### C.3 — Template and distribution

**REQ-JEVG-009** (Ubiquitous) Every new or changed file this SPEC introduces under `.claude/` or `.moai/` shall land in `internal/template/templates/` first.

**REQ-JEVG-010** (Unwanted) Template content shall not carry a SPEC ID, a card id, a date, a price, a measurement figure, or a reference to a local-only file.

**REQ-JEVG-011** (Capability gate) **Where** a file under `internal/template/templates/.claude/agents/moai/` changes, **the build shall** require `make agents-emit`; **where** a command source changes, **the build shall** require `make commands-emit`.

**REQ-JEVG-012** (Ubiquitous) The disposition of the uncommitted `scripts/jev/` working copy shall be stated in `docs/jev-negative-results.md`, with the Go package as the canonical implementation.

### C.4 — Preserved negative result

**REQ-JEVG-013** (Ubiquitous) The system shall record in `docs/jev-negative-results.md` that using Jev to judge whether a stale card's premise is dead was measured and rejected, with the measured figures preserved on that tracked, non-template surface.

**REQ-JEVG-014** (Unwanted) `moai todo triage` shall remain model-free. No requirement in this SPEC chain shall route a model answer into it.

---

## §D. Exclusions

### Out of Scope — governor authority

- Changing `mission-governor`'s scope, tool set, output shape, or the governance receipt's existing binding fields. Seat (ii) adds one recorded item and changes nothing else.
- Any path by which a Jev answer reaches a completion predicate, landed-ancestry evidence, or authoritative-readback evidence. REQ-JEVG-005 forbids it, and those are the irreversible judgments the whole chain is bounded against.
- Adding a user question inside the `--auto` sealed-scope loop. The single approval before the loop stays the only one.

### Out of Scope — seat (i) lane-question routing

- Building any host path for Consumer A's routing inside the goal loop. Host-design question N2 is OPEN (CONSUMERS spec.md §E) with its producer M5 recorded blocked; this SPEC neither answers N2 nor amends CONSUMERS' recorded state.
- Re-admitting seat (i) without a SPEC that first answers N2. The withdrawal recorded in §C.1 is the seat's only carrier until that SPEC is authored.
- A conditional seat-(i) requirement whose enabling branch has no owner. Rejected at iter-2: an unenablable conditional is a deferred promise, not scope.

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

Fourteen criteria (AC-JEVG-001 … AC-JEVG-014), in `acceptance.md`, together with the quality gates and the Definition of Done. Tier M carries acceptance criteria in their own artifact rather than inline.

## §F. Open Questions — status and dispositions

No open question is owned by this SPEC. The chain questions this SPEC consumes or records carry their statuses explicitly, because one of them gates the seat this SPEC withdrew:

| # | Question | Recorded owner | Status | Bearing on this SPEC |
|---|---|---|---|---|
| Q2 | Is the pinned model id compiled, or operator-visible configuration? | `SPEC-JEV-CORE-001` | **OPEN** (CORE §E) | Shapes the `workflow.jev` config block the wrapper's gate reads; CORE owns it, this SPEC records it. |
| Q4 | Is a credential-reveal route wanted? | `SPEC-JEV-CORE-001` | **OPEN** (CORE §E) | None — REQ-JEVC-020 is satisfied without it. |
| Q3 | Labelled-set size per consumer | `SPEC-JEV-OPTIN-MEASURE-001` | **OPEN** (carried to OPTIN's Kickoff gate) | Gates the consumers' fitted thresholds; with seat (i) withdrawn it reaches this SPEC only through a future re-scope. |
| R1 | Which surface carries the `moai web` pointer | `SPEC-JEV-OPTIN-MEASURE-001` | **RESOLVED** (OPTIN progress.md §E.1 — the wizard `Description`, all four locales) | None. |
| N1 | Dedup precedence between a Jev finding and a mechanical or agent finding naming the same unordered pair | `SPEC-JEV-CONSUMERS-001` | **SETTLED** (CONSUMERS §E; operator decision 2026-09-20, recorded as REQ-JEVN-006 halves (a)/(b)) | None — this SPEC introduces no card-deduplication surface. |
| N2 | Which code path hosts Consumer A | Recorded in `SPEC-JEV-CONSUMERS-001` §E; **currently unowned** — CONSUMERS closed `completed` with the M5 block recorded (N2's M6 half — Consumer B's verification surface — was answered at CONSUMERS' Kickoff gate) | **OPEN — blocked CONSUMERS M5 and thereby this SPEC's seat (i)** | Seat (i) is withdrawn (§C.1). A future SPEC must answer N2 before any seat-(i) scope is re-admitted; this SPEC does not resolve N2. |

**Deferred, not owned:** the six-item ops checklist proposed by the 2026-09-22 third-party aitmpl analysis of the skill-suggestion surface (per-decision logging + announce, origin filter, disable-model-invocation filter, SKILL.md session cache, duplicate-injection prevention, id canonicalization) — **DEFERRED** until a caller exists, per that report's own classification. This SPEC claims no such scope; the report itself is absent from this tree and is cited here from the plan-audit iter-1 record.
