---
id: SPEC-AUDIT-PARTICIPANT-COUNT-001
title: "Implementation plan — convergence participant count"
version: "0.2.1"
created: 2026-09-01
---

# Implementation Plan — SPEC-AUDIT-PARTICIPANT-COUNT-001

Sections are ordered by decision-reversibility: the shape decision (§D) leads because it
is the one choice that is expensive to change after it lands. The mechanical work
(tests, documentation) is deferred to the end.

## §A Tier classification

**Tier M.** The two Tier axes disagree and the higher one is taken:

| Axis | Reading | Tier |
|---|---|---|
| LOC | well under 300 — one struct field, one derivation helper, one narrowed field | S |
| Files affected | 11 across the 3-phase lifecycle (§G) | M |

Tier M also buys `acceptance.md`, which this SPEC needs: the per-criterion
mutation-witness record (§F) does not fit comfortably inline in `spec.md` §3, and it is
the artifact the operator asked to be able to read on its own.

REQ/AC budget for Tier M is 16/16; this SPEC uses 5 REQ and 8 AC.

## §B Known state (measured, not assumed)

| Fact | How it was established |
|---|---|
| `converge` never counts participants | `grep -ci participant` over `mcp_convergence.go`, `mcp_audit_multi.go`, `multi_review_gate.go` → `0 0 0` |
| `HandleMultiReviewGate` reads only `OverallVerdict` + `ResidualRiskNote` | read `multi_review_gate.go:60-123`; the only field reads are `:83` and `:97` |
| `audit_multi` declares **no** output schema | read `mcp_server.go:405-425`; the two `WithOutputSchema[ReviewOutput]()` sites (`:259`, `:400`) belong to other tools |
| Gate-`off` backends never enter `PerBackendVerdicts` via the fan-out | `mcp_convergence.go:571-573` (`continue`) + `mcp_convergence_test.go:419` |
| `converge` is nonetheless reachable with a gate-`off` entry | it is a pure function; `mcp_convergence_test.go:30` `pbv()` constructs arbitrary gates |
| `DisagreementFlag` read/write sites in tests | 12 across 3 files: 8 in `mcp_convergence_test.go` (`:75,:94,:113,:131,:163,:180,:201,:457`), 3 in `multi_review_gate_test.go` (`:65,:80,:95`), 1 in `codex_verdict_divergence_test.go:70` |
| Behaviour of the three §A.2 cases | `.moai/state/verify/t284/premise-probe.log` |
| Measured `participant_count` for all 8 existing derivation cases | `.moai/reports/t284/participant-count-probe.log` (plan-audit iter 1 probe) |

## §C Constraints the implementation must respect

1. **C3 — disagreement is information, never a gate.** Nothing added here may become a
   block category. (`mcp_convergence.go:16-21`)
2. **C2 — fail-open in both directions.** A missing optional backend stays
   evidence-of-absence. A low participant count must not push `overall_verdict` toward
   `fail`, and must not push it toward `pass` either.
3. **REQ-AMM-008** — no new `overall_verdict` enum value.
4. **The persisted state file must stay readable.** State files written by the current
   binary under `.moai/state/audit-multi/` must still decode in `loadConvergenceResult`
   (`multi_review_gate.go:110`), or an in-flight session silently fails open at the
   Stop gate.

## §D Shape decision (the reversibility-critical choice)

### Chosen: additive `participant_count int` + narrow `disagreement_flag` to a nullable boolean

```
participant_count : int       // new field, always present
disagreement_flag : bool|null // narrowed: null when participant_count < 2 — except the
                              // REQ-APC-003 carve-out: a divergence the intra-backend
                              // synthesis pass observed keeps the flag true
```

In Go: a new `ParticipantCount int` field, and `DisagreementFlag` becomes `*bool`
carrying `json:"disagreement_flag"` **without** `omitempty` — a nil pointer must
serialize as an explicit `null`, not vanish. An absent field would be indistinguishable
from output produced by an older binary; an explicit `null` says "this producer
considered the question and had no answer".

Why this shape:

- It is the only shape that satisfies the card's item 2 literally. `false` is forbidden
  when participants < 2, and a two-valued type has no third value to move to.
- It keeps the field **name** every documentation surface already uses (§G), so the doc
  delta is one sentence per surface rather than a rename.
- Old persisted state files decode cleanly: a recorded `false` unmarshals into a `*bool`
  pointing at `false`. Constraint C4 holds with no migration.
- The compile break is the *desirable* failure mode. Every existing reader of
  `r.DisagreementFlag` stops compiling and is forced to confront the third state. A
  shape that left the bool in place would let a future reader consume the ambiguous
  `false` in silence — which is the defect this SPEC exists to remove.

Cost, stated plainly: 12 test sites (§B) stop compiling and must be updated. That work
is in run-phase scope and is mechanical.

### Rejected alternatives

| # | Alternative | Why rejected |
|---|---|---|
| R1 | Add `participant_count` only; leave `disagreement_flag bool` unchanged | Fails REQ-APC-003. The `false` value is still emitted alongside `participant_count: 1`, so the ambiguous value survives in the bytes. This is precisely the representative mutant (§F), not a design. |
| R2 | Add `participant_count` + a new tri-state string `disagreement_state` (`agreed` / `disagreed` / `undetermined`), keep the bool | Two representations of one fact. And unless the bool is *also* changed, `false` still appears at `participant_count < 2` — so R2 either violates REQ-APC-003 or makes the string redundant. |
| R3 | Change `disagreement_flag` to a string enum under the same name | Same name, different type is a silent-misread hazard for every non-Go JSON reader (the string `"false"` is truthy in JavaScript), and it breaks decoding of already-persisted state files (C4). |
| R4 | Keep the Go field a `bool`, suppress it at the marshal layer via a custom `MarshalJSON` | Fixes only the serialized bytes. The in-process field still reads `false`, so `HandleMultiReviewGate` and any future in-process consumer keep seeing the ambiguous value. It hides the defect rather than removing it. |
| R5 | A new `undetermined` value on `overall_verdict`, or a new gate/block category | Forbidden by REQ-AMM-008 and C3. Disagreement — and its absence — is information. |

### Deliberately not decided here

Whether `residual_risk_note` should gain a sentence in the undetermined case. Arguments
both ways: it makes the case legible to a human reader of the Verification Matrix, but
it is a third surface for one fact and it would perturb note expectations in existing
tests. **Run-phase decision, recorded either way in `progress.md` §E.2.** Whichever way
it goes, `blockReason` (`multi_review_gate.go:96`) echoes the note only on the
`overall_verdict == fail` path, so an undetermined-case note can never reach the gate.

## §E Milestones

**M1 — data model + derivation (Priority: High).** Add `ParticipantCount` to
`ConvergenceResult`; narrow `DisagreementFlag` to `*bool`; add the participant-counting
helper next to the existing `filterVerdict` / `distinctVerdicts` family; wire both into
`converge`'s Step 2. Update the one non-`converge` producer that constructs a result
literal: the DQ-2 refusal path (`mcp_convergence.go:526-533`) — a refusal has zero
participants, so it emits a count of `0` and a null flag.

**M2 — repair existing call sites (Priority: High).** The 12 test sites of §B. Expected
values follow the measured participant counts
(`.moai/reports/t284/participant-count-probe.log`): the six cases with 2+ participants
keep their asserted booleans unchanged — `false`: `AllRequiredPass` (3), `AllRequiredFail`
(2), `NoRequiredBackends_VacuousPass` (2); `true`: `RequiredSplit` (3),
`AdvisoryOnlyConflict` (3), `DisagreementAdvisoryNotBlock` (2).
`RequiredFailWithInconclusive` (measured 1 participant: claude required=fail, codex
required=inconclusive) moves from `false` to `null` per REQ-APC-003 — a deliberate move
specified by this SPEC (AC-APC-002/003 cover it), not a repair-time judgment. The
twelfth site — `codex_verdict_divergence_test.go:70`
(`TestConverge_SurfacesSignalDivergence_WithoutBlocking`) — is routed through the
REQ-APC-003 carve-out rather than the mechanical false-preserving pass: its input has
one participant and an intra-backend divergence, so the asserted `true` survives as a
non-nil `true` — the one sub-2 case where `null` is forbidden.

**M3 — new criteria (Priority: High).** The eight `acceptance.md` criteria, including
the two byte-level ones. Follow the existing conventions in `mcp_convergence_test.go`:
`pbv` / `pbvReq` / `pbvAdv` constructors, one `Test<Subject>_<Case>_<AC-ID>` function per
case, table-driven where a table reads better than repetition.

**M4 — documentation (Priority: Low, mechanical).** Deferred to sync phase; scope in §G.

## §F Representative mutant

> An implementation that **does** count participants but still emits the plain `false`
> in the `< 2` case.

Concretely: `ParticipantCount` computed correctly, and `DisagreementFlag` set to a
non-nil pointer to `false` when the count is below 2. This mutant satisfies card item 1
and violates card item 2 — it is R1 of §D wearing item 1's clothes.

It must take at least one criterion red. `acceptance.md` §D records, per criterion,
whether it is a witness for this mutant; `AC-APC-002` (in-process), `AC-APC-003`
(serialized bytes), and `AC-APC-008` (the carve-out case, where the mutant's
pointer-to-`false` would flip an observed divergence's `true`) are the three that are.
Criteria the mutant passes are marked as non-witnesses rather than quietly listed,
because a criterion the mutant survives is not evidence for this axis.

## §G Sync-phase documentation scope (do NOT edit during run phase)

Surfaces that describe `disagreement_flag` and will need the nullable third state,
measured 2026-09-01 by occurrence count:

| Surface | Hits |
|---|---|
| `docs-site/content/ko/advanced/multi-model-audit.md` | 1 |
| `docs-site/content/en/advanced/multi-model-audit.md` | 1 |
| `docs-site/content/ja/advanced/multi-model-audit.md` | 1 |
| `docs-site/content/zh/advanced/multi-model-audit.md` | 1 |
| `docs-site/content/en/advanced/autonomous-loops.md` | 3 |
| `docs-site/content/ja/advanced/autonomous-loops.md` | 3 |
| `docs-site/content/zh/advanced/autonomous-loops.md` | 3 |
| `.claude/skills/moai-ref-cross-model-audit/SKILL.md` | 5 |
| `.claude/skills/moai/workflows/review.md` | 1 |
| `internal/template/templates/.claude/skills/moai-ref-cross-model-audit/SKILL.md` | 5 |
| `internal/template/templates/.claude/skills/moai/workflows/review.md` | 1 |

Two notes for whoever runs sync:

- The last two rows are the **template mirrors** of the two rows above them. Per the
  Template-First rule they must move together, and `make build` must run after.
- `docs-site/content/ko/advanced/autonomous-loops.md` is absent from this table because
  it carries 0 occurrences while its three siblings carry 3 each. That is a pre-existing
  4-locale parity gap (`spec.md` §E) — do **not** silently repair it under this SPEC's
  scope, and do not read its absence here as an oversight.

## §H Risks

| Risk | Mitigation |
|---|---|
| A consumer outside `internal/cli` reads `disagreement_flag` as a strict boolean | Measured: the only Go reader is the engine itself plus tests; `HandleMultiReviewGate` never touches the field, and `audit_multi` declares no output schema (§B). The remaining readers are humans and agents reading JSON, which §G addresses. |
| Semantics drift while repairing the 12 call sites | M2 states the per-case expected values explicitly, from the measured probe. Two movements are deliberate and specified by this SPEC: `RequiredFailWithInconclusive` moves `false`→`null` (REQ-APC-003), and the divergence site's `true` survives as a non-nil `true` (carve-out). Any case whose expected value moves beyond those two is a finding to report, not to adjust. |
| Participant definition disputed at the boundary (a gate-`off` entry carrying a `pass`) | REQ-APC-002 settles it: gate-`off` is never a participant, whatever verdict it carries. `AC-APC-001` pins the boundary with an explicit case. |
| Scope creep into a minimum-participant policy | `spec.md` §E forbids it. This SPEC reports the count; acting on it is a different SPEC. |

## §I Cross-references

- `spec.md` — requirements and the observable contract.
- `acceptance.md` — criteria and the mutation-witness record.
- `.moai/reports/t229/succession.md` — why this axis is a separate SPEC.
