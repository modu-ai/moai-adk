---
id: SPEC-AUDIT-PARTICIPANT-COUNT-001
title: "Convergence participant count — distinguish 'the participants agreed' from 'there were not enough participants to disagree'"
version: "0.2.1"
status: in-progress
created: 2026-09-01
updated: 2026-09-02
author: manager-spec
priority: P1
phase: "v3.1.5 target"
module: internal/cli
lifecycle: spec-anchored
tier: M
tags: "audit, multi-model, convergence, participant-count, disagreement, fail-open"
---

# SPEC-AUDIT-PARTICIPANT-COUNT-001 — Convergence participant count

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-09-01 | Plan-phase artifacts. Successor SPEC for the participant-count axis deferred out of SPEC-CODEX-VERDICT-SYNTH-001 (operator decision 2026-08-26, `.moai/reports/t229/succession.md`). |
| 0.2.0 | 2026-09-01 | Audit-repair revision (plan-audit iter 1, FAIL 0.75, defects D1-D6, `.moai/reports/t284/plan-audit-iter1.md`): REQ-APC-003 gained the intra-backend divergence carve-out (D2, option (a)); AC-APC-005's case enumeration re-partitioned by measured participant count (D1); §A.3 premise restated as derived-field identity (D3); site inventory corrected to 3 files / 12 sites (D4); §A.1/§A.4 coordinates corrected (D5/D6). AC set extended to 8. |
| 0.2.1 | 2026-09-01 | D7 debt discharge + D8 fold-in (plan-audit iter 2, PASS-WITH-DEBT 0.85, `.moai/reports/t284/plan-audit-iter2.md`): AC-APC-002/003 Given clauses scoped to 0-or-1 inputs that produced no intra-backend synthesis divergence (AC-APC-008 owns the divergence input); AC-APC-005's coverage note and §D's option-(b) sentence aligned; the two criteria headings carry the same qualifier; spec.md §C "(existing case #3)" provenance named. Wording/scoping only — no expectation value, coverage-table row, or witness-table change. |

## §A Context

### §A.1 The measured defect

`converge(...)` (`internal/cli/mcp_convergence.go:140`) derives `disagreement_flag` from
three passes and **none of them counts how many backends actually contributed a
comparable verdict**:

| Pass | Location | What it reads |
|---|---|---|
| required-set split | `mcp_convergence.go:168-170` | `len(distinctVerdicts(required, "pass", "fail")) > 1` |
| advisory-vs-required conflict | `mcp_convergence.go:171-185` | an advisory `pass`/`fail` not present in the required set |
| intra-backend synthesis note | `mcp_convergence.go:188-197` | `collectSynthesisNotes(verdicts)` (`:194`); assigns the flag `true` when any note exists (`:195-197`) |

A grep for `participant` across the three convergence surfaces
(`mcp_convergence.go`, `mcp_audit_multi.go`, `multi_review_gate.go`) returns **0 hits**
— there is no participant-count logic to repair, only one to add.

Consequence: `disagreement_flag: false` is emitted identically whether three backends
examined the target and agreed, or one backend examined it and had nobody to disagree
with. The reader of the Verification Matrix cannot tell one from the other, and the
result carries no participant count from which to re-derive it.

### §A.2 Measured evidence

Live `converge(...)` calls at worktree HEAD `64bba61aa`; verbatim JSON at
`.moai/state/verify/t284/premise-probe.log`.

| # | Input | `overall_verdict` | `disagreement_flag` | `fail_open_backends` |
|---|---|---|---|---|
| 1 | claude required=pass; codex advisory=inconclusive; glm advisory=inconclusive | `pass` | `false` | `["codex","glm"]` |
| 2 | claude/codex/glm all pass (genuine 3-way agreement) | `pass` | `false` | `[]` |
| 3 | claude required=pass, no other backend present | `pass` | `false` | `[]` |

### §A.3 Premise refinement (the card's wording is imprecise; this SPEC does not inherit it)

The originating card says "the same bytes hide two facts". Measured, the two JSON
documents are **not** byte-identical: case 2 carries three `per_backend_verdicts` entries
where case 3 carries one (447 vs 235 bytes,
`.moai/state/verify/t284/premise-probe.log`). What **is** identical is every derived
field the reader of a convergence result acts on — `overall_verdict`,
`disagreement_flag`, `residual_risk_note`, `fail_open_backends` — and case 3 is
therefore the strongest instance of the defect: the per-backend detail distinguishes the
bytes, but nothing in the derived summary does. Case 1 vs case 2 are distinguishable
through `fail_open_backends` (`["codex","glm"]` vs `[]`), so the "identical derived
fields" framing does not hold for the single-on-target-participant case with fail-open
siblings either.

The defect is correctly stated as:

> **`disagreement_flag = false` alone does not distinguish "the participants agreed"
> from "there were not enough participants to disagree", and the result carries no
> explicit participant count.**

It is *not* correctly stated as "the whole result is identical in every
single-participant case".

### §A.4 Inherited constraints (all preserved, none amended)

From the `mcp_convergence.go` header block (lines 1-39):

- **C2 — fail-open identity.** A missing / unauthenticated / optional backend returns
  `inconclusive` and convergence continues. Evidence of absence is not evidence of failure.
- **C3 — disagreement is INFORMATION, never a GATE.** `disagreement_flag` never
  hard-blocks on its own.
- **C5 — no `AskUserQuestion` from the engine.** Every condition is a structured result.
- **REQ-AMM-008** — `overall_verdict` stays within `{pass, fail}`; no new enum.

## §B Requirements (GEARS)

**REQ-APC-001 (ubiquitous).** The convergence engine shall expose, in every
`ConvergenceResult`, the number of on-target backends that contributed a comparable
verdict, as a `participant_count` field.

**REQ-APC-002 (ubiquitous — participant definition).** The convergence engine shall
count a backend entry as a participant if and only if its gate is not `off` **and** its
verdict is `pass` or `fail`; an `inconclusive` verdict and an `off`-gated entry shall
not be counted.

**REQ-APC-003 (state-driven + unwanted, with the intra-backend divergence carve-out).**
While `participant_count` is less than 2, the convergence engine shall emit
`disagreement_flag` as the JSON value `null`, and shall not emit it as `false` —
**unless** the intra-backend synthesis divergence pass (§A.1 pass 3) observed a
divergence for that input, in which case the engine shall emit `disagreement_flag` as
the boolean `true`. Below 2 participants the value `false` is forbidden without
exception: `false` is a positive claim that two or more participants were compared and
none disagreed, and a single participant cannot ground it.

*Carve-out rationale (option (a) of plan-audit D2, decided by the orchestrator and
surfaced at the kickoff gate).* (1) The originating card forbids `false` below 2
participants and says nothing about suppressing `true`. (2) The intra-backend synthesis
pass is the landed REQ-CVS-003 behaviour of completed SPEC-CODEX-VERDICT-SYNTH-001,
regression-guarded by `TestConverge_SurfacesSignalDivergence_WithoutBlocking`
(`internal/cli/codex_verdict_divergence_test.go:70`); suppressing its `true` would
discard directly-observed information and turn a completed SPEC's regression guard red.
(3) Inherited constraint C3 — disagreement is information, never a gate. *Rejected
alternative (option (b))* — accept the information loss: below 2 forces `null`
unconditionally. Rejected because it discards a divergence the engine directly observed
and turns the predecessor SPEC's regression guard red (reasons 2 and 3).

**REQ-APC-004 (state-driven).** While `participant_count` is 2 or greater, the
convergence engine shall emit `disagreement_flag` as exactly the boolean the existing
three-pass derivation produces, unchanged.

**REQ-APC-005 (unwanted — C3 preservation).** The undetermined state shall not alter
`overall_verdict`, shall not add a block path to the multi-review Stop gate, and shall
not change which backends appear in `fail_open_backends`.

## §C Observable contract

The SPEC fixes the observable result shape; the Go-level representation and the
alternatives weighed are `plan.md` §D.

| Case (from §A.2) | `participant_count` | `disagreement_flag` |
|---|---|---|
| 1 — one on-target, two fail-open | `1` | `null` |
| 2 — genuine 3-way agreement | `3` | `false` |
| 3 — claude only | `1` | `null` |
| required split (existing derivation Case #3 of SPEC-AUDIT-MULTI-MODEL-001 — the Step-1 verdict-case numbering at `mcp_convergence.go:152-153`, carried by `TestConverge_RequiredSplit_Case3_AC_AMM_008`) | `2`+ | `true` |
| intra-backend divergence, single participant (`SurfacesSignalDivergence` input) | `1` | `true` — the REQ-APC-003 carve-out; the one sub-2 case where `null` is forbidden |

Case 2 and case 3 become distinguishable in the derived summary: they differ in **both**
new-field positions (`participant_count` 3 vs 1, `disagreement_flag` `false` vs `null`).

## §D Traceability

Requirement-to-criterion mapping, and the per-criterion mutation-witness record, are in
`acceptance.md` §D. Every REQ above has at least one criterion; `AC-APC-002`,
`AC-APC-003`, and `AC-APC-008` are the criteria the representative mutant of
`plan.md` §F must fail.

## §E Exclusions

Everything below is **out of scope** for this SPEC. The three items of the originating
card are the whole contract; adjacent improvements to the convergence engine are not
admitted.

### Out of Scope — verdict-synthesis axis

- Intra-backend verdict-signal synthesis (`adoptConservativeVerdict`,
  `codexScoredVerdict`, `codexVerdictSignalsOf`). That axis landed as
  SPEC-CODEX-VERDICT-SYNTH-001 (#1663) and is complete; this SPEC is the successor for
  the participant-count axis **only** (`.moai/reports/t229/succession.md`).
- Any change to `collectSynthesisNotes` itself — its inputs, when it fires, or what it
  reports. Its observable effect on `disagreement_flag` is preserved, not changed: the
  REQ-APC-003 carve-out exists precisely so that a divergence this pass observes keeps
  the flag `true` even below 2 participants. The intended observable consequence of
  this SPEC for that pass is therefore *nil* — it fires, reads, and surfaces exactly as
  before.

### Out of Scope — gate semantics

- Making a low participant count block, warn, or gate anything. C3 holds: this is
  information, not a gate.
- Any new `overall_verdict` value, any new block category in `HandleMultiReviewGate`,
  and any change to its opt-in default-off toggle.
- A minimum-participant policy, threshold, or configuration key. This SPEC reports the
  count; it does not act on it.

### Out of Scope — backend behaviour

- Changing when a backend is invoked, which gate it defaults to, or how a fail-open
  `inconclusive` verdict is produced.
- Retrying, re-invoking, or substituting a backend that returned `inconclusive` in order
  to raise the participant count.

### Out of Scope — pre-existing documentation gap

- `docs-site/content/ko/advanced/autonomous-loops.md` carries **0** occurrences of
  `disagreement` while its `en` / `ja` / `zh` siblings carry 3 each (measured
  2026-09-01). This 4-locale parity gap predates this SPEC and is not caused by it; it
  is recorded here so its omission is not silent. Repairing it is a separate card.

## §F Cross-references

- `internal/cli/mcp_convergence.go` — the convergence engine (header C1-C5, `converge` at :140).
- `internal/cli/multi_review_gate.go` — the Stop-hook consumer; reads `OverallVerdict` (:83) and `ResidualRiskNote` (:97) only.
- `.moai/reports/t229/succession.md` — the deferral record and the operator decision of 2026-08-26.
- `.moai/state/verify/t284/premise-probe.log` — the §A.2 measurement.
- SPEC-AUDIT-MULTI-MODEL-001 — the parent SPEC that established the `ConvergenceResult` shape.
