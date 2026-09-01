---
id: SPEC-AUDIT-PARTICIPANT-COUNT-001
title: "Acceptance criteria — convergence participant count"
version: "0.2.1"
created: 2026-09-01
---

# Acceptance Criteria — SPEC-AUDIT-PARTICIPANT-COUNT-001

Eight criteria, each binary-testable. §D records, per criterion, whether it is a witness
for the representative mutant of `plan.md` §F. A criterion the mutant survives is listed
as a non-witness rather than left ambiguous — it may be a good criterion for another
axis, but it is not evidence for this one.

## §A Criteria

### AC-APC-001 — participant count is exposed and counts the right entries

**Given** a `ConvergenceResult` produced by `converge(...)`,
**when** the input carries a mixture of gates and verdicts,
**then** `participant_count` equals the number of entries whose gate is not `off` and
whose verdict is `pass` or `fail`, for every row of this table:

| # | Input entries | Expected `participant_count` |
|---|---|---|
| a | (none) | 0 |
| b | claude required=pass | 1 |
| c | claude required=pass; codex advisory=inconclusive; glm advisory=inconclusive | 1 |
| d | codex advisory=pass; glm advisory=pass | 2 |
| e | claude required=pass; codex required=fail | 2 |
| f | claude required=pass; codex advisory=pass; glm advisory=pass | 3 |
| g | claude required=pass; codex **off**=pass | 1 |
| h | claude required=inconclusive; codex required=inconclusive | 0 |

Row **g** is the gate-`off` boundary REQ-APC-002 settles; row **h** is the all-fail-open
boundary. Both are required — a table omitting them does not pin the definition.

Traces: REQ-APC-001, REQ-APC-002.

### AC-APC-002 — below two participants without an observed divergence, the in-process flag is absent, not false

**Given** any input whose `participant_count` is 0 or 1 **and which produced no
intra-backend synthesis divergence** (the divergence case is AC-APC-008's, not this
criterion's),
**when** `converge(...)` returns,
**then** the returned `DisagreementFlag` is nil — and specifically **not** a non-nil
pointer whose value is `false`.

The assertion must distinguish those two states explicitly. A test written as
`if r.DisagreementFlag != nil && *r.DisagreementFlag { ... }` does not satisfy this
criterion: it passes for both nil and a pointer to `false`, which is exactly the
distinction under test.

Covers at minimum the three measured cases of `spec.md` §A.2 (rows 1 and 3 are the
undetermined ones), the DQ-2 refusal path (zero participants), and
`RequiredFailWithInconclusive` — measured `participant_count` 1 (claude required=fail;
codex required=inconclusive, the inconclusive entry not being a participant under
REQ-APC-002), the one existing derivation case whose asserted `false` becomes `null`
under REQ-APC-003. That move is deliberate and specified; AC-APC-005's ≥2 enumeration
excludes the case for exactly this reason.

Traces: REQ-APC-003.

### AC-APC-003 — below two participants without an observed divergence, the serialized value is JSON null

**Given** a `ConvergenceResult` whose `participant_count` is 0 or 1 **and which
produced no intra-backend synthesis divergence** (the divergence case is AC-APC-008's,
not this criterion's),
**when** the result is marshalled to JSON (the form written to the per-session state
file and returned by the `audit_multi` tool),
**then** the document contains a `disagreement_flag` member, and that member's value is
JSON `null` — not `false`, and not absent.

Binary form: unmarshal the produced bytes into a `map[string]any`; the key
`disagreement_flag` must be present, and its value must be nil rather than the boolean
`false`. Asserting on the raw bytes that `"disagreement_flag":false` does not occur is an
acceptable equivalent, provided the presence of the member is also asserted — absence
must fail this criterion just as `false` does.

This criterion is deliberately separate from AC-APC-002: they fail independently. A
`omitempty` tag added to the pointer would satisfy AC-APC-002 and fail this one.

Traces: REQ-APC-003.

### AC-APC-004 — the two measured results stop sharing an identical derived summary

**Given** the two inputs measured at `.moai/state/verify/t284/premise-probe.log`:
genuine three-way agreement (claude/codex/glm all `pass`) and claude-only (`claude`
required=`pass`, no other entry),
**when** both results are marshalled to JSON,
**then** the two documents differ in **both** of the positions this SPEC adds or
narrows: `participant_count` (3 vs 1) and `disagreement_flag` (`false` vs `null`).

This is the direct regression witness for the defect recorded in `spec.md` §A.3 — the
one case where the current engine's *derived summary* is identical for two different
facts. (The full documents were never byte-identical — they differ in
`per_backend_verdicts`, 447 vs 235 bytes — which is why the criterion asserts the two
derived positions rather than byte inequality: byte inequality already holds at HEAD
and would assert nothing.)

Traces: REQ-APC-001, REQ-APC-003.

### AC-APC-005 — at two or more participants, nothing about the boolean changes

**Given** any input whose `participant_count` is 2 or greater,
**when** `converge(...)` returns,
**then** `DisagreementFlag` is non-nil and its value equals the value the pre-change
three-pass derivation produced for the same input.

Concretely, the six existing derivation cases with measured `participant_count` ≥ 2
keep their asserted values (counts measured in
`.moai/reports/t284/participant-count-probe.log`): the three `false` cases
(`AllRequiredPass` 3, `AllRequiredFail` 2, `NoRequiredBackends_VacuousPass` 2) and the
three `true` cases (`RequiredSplit` 3, `AdvisoryOnlyConflict` 3,
`DisagreementAdvisoryNotBlock` 2).

The two existing cases this enumeration excludes are excluded deliberately, and each is
owned elsewhere: `RequiredFailWithInconclusive` — measured `participant_count` 1 (claude
required=fail; codex required=inconclusive) — moves from `false` to `null` as
REQ-APC-003 requires; AC-APC-002 and AC-APC-003 cover it (both bind every 0-or-1 input
that produced no intra-backend divergence; AC-APC-008 owns the divergence input, and
AC-APC-002's coverage list names this case). `SurfacesSignalDivergence` — measured
`participant_count` 1 — keeps a non-nil `true` under the REQ-APC-003 carve-out;
AC-APC-008 covers it. Any case whose expected value would move under the new derivation
*beyond those two specified moves* is a finding to report, not a value to adjust.

Traces: REQ-APC-004.

### AC-APC-006 — the undetermined state gates nothing

**Given** an undetermined result (`participant_count` < 2),
**when** it is passed through the full path — `converge(...)`, persistence, and
`HandleMultiReviewGate` with the gate enabled and a change detected,
**then** all of the following hold:

- `overall_verdict` is the same value the pre-change engine produced for that input;
- `fail_open_backends` contains the same backends as before;
- the gate returns ALLOW, and its only BLOCK path remains `overall_verdict == fail`.

Traces: REQ-APC-005.

### AC-APC-007 — a previously written state file still decodes

**Given** a state file in the format the current binary writes — carrying
`disagreement_flag` as a boolean and no `participant_count` member —
**when** `loadConvergenceResult` reads it,
**then** the load succeeds (`ok == true`), the flag decodes to a non-nil pointer holding
the recorded boolean, `participant_count` reads as its zero value, and the gate's
decision for that result is unchanged.

A decode failure here would fail the gate open silently for any session in flight across
the upgrade, which is why this is a criterion and not a note.

Traces: REQ-APC-005 (constraint C4 of `plan.md` §C).

### AC-APC-008 — the carve-out: a single participant's intra-backend divergence stays `true`

**Given** an input whose only participant observed an intra-backend divergence — the
`SurfacesSignalDivergence` input (`TestConverge_SurfacesSignalDivergence_WithoutBlocking`,
`internal/cli/codex_verdict_divergence_test.go:55-79`: claude required=`pass`; codex
required=`inconclusive` carrying a synthesis note; measured `participant_count` = 1),
**when** `converge(...)` returns and the result is marshalled,
**then** `participant_count` is 1; `disagreement_flag` is a non-nil `true`, in process
and serialized as the boolean `true` — **not** `null`, this being the one below-2 case
where `null` is forbidden; and `overall_verdict`, `fail_open_backends`, and the gate
decision are unchanged from the pre-change engine's output for the same input.

The regression guard already exists and must stay green:
`TestConverge_SurfacesSignalDivergence_WithoutBlocking` asserts this `true` today and
guards the landed REQ-CVS-003 behaviour of completed SPEC-CODEX-VERDICT-SYNTH-001.

Traces: REQ-APC-003 (carve-out clause).

## §B Edge cases

| Case | Expected |
|---|---|
| Empty verdict slice | `participant_count: 0`, flag null, `overall_verdict` unchanged (vacuous pass) |
| DQ-2 refusal (claude anchor missing) | `participant_count: 0`, flag null, `overall_verdict: fail`, existing note preserved |
| All entries `inconclusive` | `participant_count: 0`, flag null, fail-open to claude preserved |
| Exactly 2 participants that agree | `participant_count: 2`, flag non-nil `false` — the boundary is inclusive at 2 |
| Gate-`off` entry carrying `pass` | not counted (AC-APC-001 row g) |

## §C Definition of Done

- [ ] All eight criteria pass.
- [ ] `go test ./internal/cli/...` green for the affected package.
- [ ] `go vet ./...` clean; `golangci-lint run` clean.
- [ ] The representative mutant of `plan.md` §F, applied by hand, turns AC-APC-002,
      AC-APC-003, and AC-APC-008 red — observed and recorded in `progress.md` §E.2 with
      the failing output, not asserted from reasoning.
- [ ] No file under `docs-site/`, `.claude/skills/`, or
      `internal/template/templates/` modified during run phase (sync-phase scope,
      `plan.md` §G).
- [ ] The `residual_risk_note` decision left open in `plan.md` §D is recorded either way
      in `progress.md` §E.2.

## §D Mutation-witness record

Representative mutant (`plan.md` §F): *counts participants correctly, but still emits a
non-nil pointer to `false` when the count is below 2.*

| Criterion | Mutant outcome | Witness? | Why |
|---|---|---|---|
| AC-APC-001 | passes | **No** | the mutant counts correctly; this criterion tests only the count |
| AC-APC-002 | **fails** | **Yes** | it asserts nil specifically, and the mutant returns a pointer to `false` |
| AC-APC-003 | **fails** | **Yes** | the marshalled member is `false`, not `null` |
| AC-APC-004 | passes | No | the two documents already differ via `participant_count` (3 vs 1) under the mutant; this criterion witnesses the original defect, not this mutant |
| AC-APC-005 | passes | No | the mutant leaves the 2+ path untouched |
| AC-APC-006 | passes | No | the mutant gates nothing either; C3 is preserved by both |
| AC-APC-007 | passes | No | decoding is unaffected by the mutant |
| AC-APC-008 | **fails** | **Yes** | the mutant emits a non-nil pointer to `false` for *every* sub-2 case, this one included — it would flip the divergence case's `true` to `false`, discarding exactly the directly-observed information the carve-out exists to keep |

Three witnesses: AC-APC-002 in process, AC-APC-003 in the serialized bytes, AC-APC-008
on the carve-out's information preservation. They are independent — a mutation that
satisfies one can still fail another: an `omitempty` tag satisfies AC-APC-002 and fails
AC-APC-003; an option-(b) implementation (unconditional `null` below 2) satisfies
AC-APC-002 and AC-APC-003 — both of which bind only 0-or-1 inputs without an observed
intra-backend divergence — and fails AC-APC-008.

A second mutant worth killing, recorded for run phase rather than as a separate
criterion: *counts `inconclusive` entries as participants*. It turns `spec.md` §A.2 case
1 into three participants and re-emits `false`. AC-APC-001 row c is its witness.
